package txpool

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/proofs"
)

// STARK mempool aggregation errors.
var (
	ErrAggNotRunning     = errors.New("stark_aggregation: aggregator not running")
	ErrAggAlreadyRunning = errors.New("stark_aggregation: aggregator already running")
	ErrAggNoTransactions = errors.New("stark_aggregation: no validated transactions")
	ErrAggTickFailed     = errors.New("stark_aggregation: tick generation failed")
	ErrAggInvalidTick    = errors.New("stark_aggregation: invalid tick data")
	ErrAggMergeFailed    = errors.New("stark_aggregation: merge failed")
	ErrAggTickTooLarge   = errors.New("stark_aggregation: tick exceeds 128KB bandwidth limit")
)

// Default aggregation parameters.
const (
	DefaultTickInterval = 500 * time.Millisecond
	MaxTickTransactions = 10000
	// MaxTickSize is the maximum serialized size of a mempool tick (128KB per ethresear.ch).
	MaxTickSize = 128 * 1024
)

// ValidatedTx represents a transaction that has been validated with a proof.
type ValidatedTx struct {
	TxHash          types.Hash
	ValidationProof []byte // proof of tx validity
	ValidatedAt     time.Time
	GasUsed         uint64
	RemoteProven    bool // true if proven by a remote peer's STARK tick
}

// MempoolAggregationTick represents a single aggregation cycle result.
type MempoolAggregationTick struct {
	// Timestamp is when this tick was generated.
	Timestamp time.Time
	// ValidTxHashes are the transaction hashes included in this tick.
	ValidTxHashes []types.Hash
	// AggregateProof is the STARK proving all tx validations are valid.
	AggregateProof *proofs.STARKProofData
	// DiscardList contains txs invalidated since the last tick.
	DiscardList []types.Hash
	// TickInterval is the duration between ticks.
	TickInterval time.Duration
	// PeerID identifies the node that generated this tick.
	PeerID string
	// TickNumber is the sequential tick counter.
	TickNumber uint64
	// ValidBitfield is a compact bitfield where bit i indicates tx i is valid.
	ValidBitfield []byte
	// TxMerkleRoot is the Merkle root of the valid transaction hashes.
	TxMerkleRoot types.Hash
}

// STARKAggregator implements Vitalik's recursive STARK mempool proposal.
// Every tick interval (default 500ms), it creates a STARK proving validity
// of all known validated transactions.
type STARKAggregator struct {
	mu           sync.RWMutex
	validTxs     map[types.Hash]*ValidatedTx
	discardList  []types.Hash
	prover       *proofs.STARKProver
	tickInterval time.Duration
	peerID       string
	tickNumber   uint64
	running      bool
	stopCh       chan struct{}
	tickCh       chan *MempoolAggregationTick
}

// NewSTARKAggregator creates a new STARK mempool aggregator.
func NewSTARKAggregator(peerID string) *STARKAggregator {
	return &STARKAggregator{
		validTxs:     make(map[types.Hash]*ValidatedTx),
		prover:       proofs.NewSTARKProver(),
		tickInterval: DefaultTickInterval,
		peerID:       peerID,
		stopCh:       make(chan struct{}),
		tickCh:       make(chan *MempoolAggregationTick, 16),
	}
}

// NewSTARKAggregatorWithInterval creates a new aggregator with a custom tick interval.
func NewSTARKAggregatorWithInterval(peerID string, interval time.Duration) *STARKAggregator {
	agg := NewSTARKAggregator(peerID)
	if interval > 0 {
		agg.tickInterval = interval
	}
	return agg
}

// Start begins the periodic aggregation tick loop.
func (sa *STARKAggregator) Start() error {
	sa.mu.Lock()
	if sa.running {
		sa.mu.Unlock()
		return ErrAggAlreadyRunning
	}
	sa.running = true
	sa.mu.Unlock()

	go sa.tickLoop()
	return nil
}

// Stop halts the aggregation tick loop.
func (sa *STARKAggregator) Stop() {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	if !sa.running {
		return
	}
	sa.running = false
	close(sa.stopCh)
}

// IsRunning returns whether the aggregator is currently running.
func (sa *STARKAggregator) IsRunning() bool {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	return sa.running
}

// TickChannel returns the channel that receives completed aggregation ticks.
func (sa *STARKAggregator) TickChannel() <-chan *MempoolAggregationTick {
	return sa.tickCh
}

// AddValidatedTx adds a validated transaction to the aggregation set.
func (sa *STARKAggregator) AddValidatedTx(txHash types.Hash, validationProof []byte, gasUsed uint64) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	sa.validTxs[txHash] = &ValidatedTx{
		TxHash:          txHash,
		ValidationProof: append([]byte(nil), validationProof...),
		ValidatedAt:     time.Now(),
		GasUsed:         gasUsed,
	}
}

// RemoveTx removes a transaction from the aggregation set and adds it to the discard list.
func (sa *STARKAggregator) RemoveTx(txHash types.Hash) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if _, exists := sa.validTxs[txHash]; exists {
		delete(sa.validTxs, txHash)
		sa.discardList = append(sa.discardList, txHash)
	}
}

// PendingCount returns the number of validated transactions pending aggregation.
func (sa *STARKAggregator) PendingCount() int {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	return len(sa.validTxs)
}

// GenerateTick creates an aggregate STARK proof for the current validated tx set.
func (sa *STARKAggregator) GenerateTick() (*MempoolAggregationTick, error) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if len(sa.validTxs) == 0 {
		return nil, ErrAggNoTransactions
	}

	// Collect tx hashes and build execution trace.
	txHashes := make([]types.Hash, 0, len(sa.validTxs))
	trace := make([][]proofs.FieldElement, 0, len(sa.validTxs))

	for hash, vtx := range sa.validTxs {
		txHashes = append(txHashes, hash)
		// Each tx becomes a trace row: [hash_hi, hash_lo, gas_used]
		hi := new(big.Int).SetBytes(hash[:16])
		lo := new(big.Int).SetBytes(hash[16:])
		trace = append(trace, []proofs.FieldElement{
			{Value: hi},
			{Value: lo},
			proofs.NewFieldElement(int64(vtx.GasUsed)),
		})
	}

	constraints := []proofs.STARKConstraint{
		{Degree: 1, Coefficients: []proofs.FieldElement{proofs.NewFieldElement(1)}},
	}

	starkProof, err := sa.prover.GenerateSTARKProof(trace, constraints)
	if err != nil {
		return nil, ErrAggTickFailed
	}

	// Build bitfield: all transactions in the tick are valid, so all bits are set.
	bitfieldLen := (len(txHashes) + 7) / 8
	bitfield := make([]byte, bitfieldLen)
	for i := range txHashes {
		bitfield[i/8] |= 1 << uint(i%8)
	}

	// Compute Merkle root of valid tx hashes.
	txMerkleRoot := computeTxMerkleRoot(txHashes)

	// Capture and reset discard list.
	discards := sa.discardList
	sa.discardList = nil
	sa.tickNumber++

	return &MempoolAggregationTick{
		Timestamp:      time.Now(),
		ValidTxHashes:  txHashes,
		AggregateProof: starkProof,
		DiscardList:    discards,
		TickInterval:   sa.tickInterval,
		PeerID:         sa.peerID,
		TickNumber:     sa.tickNumber,
		ValidBitfield:  bitfield,
		TxMerkleRoot:   txMerkleRoot,
	}, nil
}

// MergeTick merges a remote peer's aggregation tick into the local state.
func (sa *STARKAggregator) MergeTick(remote *MempoolAggregationTick) error {
	if remote == nil {
		return ErrAggInvalidTick
	}
	if remote.AggregateProof == nil {
		return ErrAggInvalidTick
	}

	// Check approximate tick size (each tx hash = 32 bytes, proof overhead ~1KB).
	approxSize := len(remote.ValidTxHashes)*32 + 1024
	if approxSize > MaxTickSize {
		return ErrAggTickTooLarge
	}

	// Verify the remote STARK proof.
	valid, err := sa.prover.VerifySTARKProof(remote.AggregateProof, nil)
	if err != nil || !valid {
		return ErrAggMergeFailed
	}

	sa.mu.Lock()
	defer sa.mu.Unlock()

	// Remove discarded txs.
	for _, hash := range remote.DiscardList {
		delete(sa.validTxs, hash)
	}

	// Merge remote-proven transactions into the local valid set.
	for _, txHash := range remote.ValidTxHashes {
		if _, exists := sa.validTxs[txHash]; !exists {
			sa.validTxs[txHash] = &ValidatedTx{
				TxHash:       txHash,
				ValidatedAt:  remote.Timestamp,
				RemoteProven: true,
			}
		}
	}

	return nil
}

// TickHash computes a deterministic hash of a tick for comparison.
func TickHash(tick *MempoolAggregationTick) types.Hash {
	h := sha256.New()
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], tick.TickNumber)
	h.Write(buf[:])
	for _, txHash := range tick.ValidTxHashes {
		h.Write(txHash[:])
	}
	if tick.AggregateProof != nil {
		h.Write(tick.AggregateProof.TraceCommitment[:])
	}
	var result types.Hash
	copy(result[:], h.Sum(nil))
	return result
}

// computeTxMerkleRoot computes a simple binary Merkle root of transaction hashes.
func computeTxMerkleRoot(hashes []types.Hash) types.Hash {
	if len(hashes) == 0 {
		return types.Hash{}
	}
	if len(hashes) == 1 {
		return hashes[0]
	}

	// Simple binary Merkle tree using SHA-256.
	layer := make([]types.Hash, len(hashes))
	copy(layer, hashes)

	for len(layer) > 1 {
		var next []types.Hash
		for i := 0; i < len(layer); i += 2 {
			h := sha256.New()
			h.Write(layer[i][:])
			if i+1 < len(layer) {
				h.Write(layer[i+1][:])
			} else {
				h.Write(layer[i][:]) // duplicate last if odd
			}
			var hash types.Hash
			copy(hash[:], h.Sum(nil))
			next = append(next, hash)
		}
		layer = next
	}
	return layer[0]
}

// tickLoop runs the periodic aggregation.
func (sa *STARKAggregator) tickLoop() {
	ticker := time.NewTicker(sa.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sa.stopCh:
			return
		case <-ticker.C:
			tick, err := sa.GenerateTick()
			if err != nil {
				continue // skip empty ticks
			}
			select {
			case sa.tickCh <- tick:
			default:
				// channel full, drop oldest
			}
		}
	}
}
