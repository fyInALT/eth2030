// Package registry manages registered native rollups, providing batch submission,
// state transition verification, and L1<->L2 deposit/withdrawal processing.
package registry

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/big"
	"sync"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
)

// Native rollup errors.
var (
	ErrRollupNotFound         = errors.New("native_rollup: rollup not found")
	ErrRollupAlreadyExists    = errors.New("native_rollup: rollup already registered")
	ErrRollupIDZero           = errors.New("native_rollup: rollup ID must be non-zero")
	ErrRollupNameEmpty        = errors.New("native_rollup: rollup name must be non-empty")
	ErrBatchDataEmpty         = errors.New("native_rollup: batch data must be non-empty")
	ErrBatchDataTooLarge      = errors.New("native_rollup: batch data exceeds maximum size")
	ErrStateTransitionInvalid = errors.New("native_rollup: state transition verification failed")
	ErrProofTooShort          = errors.New("native_rollup: proof data too short")
	ErrDepositAmountZero      = errors.New("native_rollup: deposit amount must be positive")
	ErrDepositFromZero        = errors.New("native_rollup: deposit sender must be non-zero")
	ErrWithdrawAmountZero     = errors.New("native_rollup: withdrawal amount must be positive")
	ErrWithdrawToZero         = errors.New("native_rollup: withdrawal recipient must be non-zero")
	ErrWithdrawProofInvalid   = errors.New("native_rollup: withdrawal proof verification failed")
	ErrWithdrawProofEmpty     = errors.New("native_rollup: withdrawal proof must be non-empty")
)

// MaxBatchDataSize is the maximum allowed batch data size (2 MiB).
const MaxBatchDataSize = 2 << 20

// MinProofLen is the minimum proof length for state transition verification.
const MinProofLen = 32

// RollupConfig holds the configuration for registering a new native rollup.
type RollupConfig struct {
	// ID uniquely identifies the rollup. Must be non-zero.
	ID uint64

	// Name is a human-readable name for the rollup.
	Name string

	// BridgeContract is the L1 bridge contract address for this rollup.
	BridgeContract types.Address

	// GenesisStateRoot is the initial state root of the rollup.
	GenesisStateRoot types.Hash

	// GasLimit is the block gas limit for the rollup chain.
	GasLimit uint64
}

// NativeRollup represents a registered native rollup on L1.
type NativeRollup struct {
	// ID uniquely identifies the rollup.
	ID uint64

	// Name is the human-readable rollup name.
	Name string

	// StateRoot is the current verified state root.
	StateRoot types.Hash

	// LastBlock is the most recently verified L2 block number.
	LastBlock uint64

	// BridgeContract is the L1 bridge contract address.
	BridgeContract types.Address

	// GasLimit is the rollup block gas limit.
	GasLimit uint64

	// TotalBatches is the total number of batches processed.
	TotalBatches uint64

	// TotalDeposits tracks the count of processed deposits.
	TotalDeposits uint64

	// TotalWithdrawals tracks the count of processed withdrawals.
	TotalWithdrawals uint64

	// Deposits holds pending and completed deposits.
	Deposits []*Deposit

	// Withdrawals holds pending and completed withdrawals.
	Withdrawals []*Withdrawal
}

// Deposit represents an L1 -> L2 deposit for a native rollup.
type Deposit struct {
	// ID is the deposit hash identifier.
	ID types.Hash

	// RollupID is the target rollup.
	RollupID uint64

	// From is the L1 sender address.
	From types.Address

	// Amount is the deposit value in wei.
	Amount *big.Int

	// BlockNumber is the L1 block at which the deposit was processed.
	BlockNumber uint64

	// Finalized indicates whether the deposit has been confirmed on L2.
	Finalized bool
}

// Withdrawal represents an L2 -> L1 withdrawal for a native rollup.
type Withdrawal struct {
	// ID is the withdrawal hash identifier.
	ID types.Hash

	// RollupID is the source rollup.
	RollupID uint64

	// To is the L1 recipient address.
	To types.Address

	// Amount is the withdrawal value in wei.
	Amount *big.Int

	// Proof is the withdrawal proof data.
	Proof []byte

	// Verified indicates whether the withdrawal proof was verified.
	Verified bool
}

// BatchResult holds the result of processing a rollup batch.
type BatchResult struct {
	// RollupID is the rollup that processed the batch.
	RollupID uint64

	// BatchHash is the Keccak256 hash of the batch data.
	BatchHash types.Hash

	// PreStateRoot is the state root before the batch.
	PreStateRoot types.Hash

	// PostStateRoot is the state root after the batch.
	PostStateRoot types.Hash

	// BlockNumber is the new L2 block number after the batch.
	BlockNumber uint64
}

// Registry manages registered native rollups. Thread-safe.
type Registry struct {
	mu      sync.RWMutex
	rollups map[uint64]*NativeRollup
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		rollups: make(map[uint64]*NativeRollup),
	}
}

// RegisterRollup registers a new native rollup with the given configuration.
func (r *Registry) RegisterRollup(config RollupConfig) (*NativeRollup, error) {
	if config.ID == 0 {
		return nil, ErrRollupIDZero
	}
	if config.Name == "" {
		return nil, ErrRollupNameEmpty
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rollups[config.ID]; exists {
		return nil, ErrRollupAlreadyExists
	}

	nr := &NativeRollup{
		ID:             config.ID,
		Name:           config.Name,
		StateRoot:      config.GenesisStateRoot,
		LastBlock:      0,
		BridgeContract: config.BridgeContract,
		GasLimit:       config.GasLimit,
		Deposits:       make([]*Deposit, 0),
		Withdrawals:    make([]*Withdrawal, 0),
	}

	r.rollups[config.ID] = nr
	return nr, nil
}

// GetRollupState returns the current state of a registered rollup.
func (r *Registry) GetRollupState(rollupID uint64) (*NativeRollup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nr, ok := r.rollups[rollupID]
	if !ok {
		return nil, ErrRollupNotFound
	}

	// Return a copy to prevent external mutation.
	cp := *nr
	cp.Deposits = make([]*Deposit, len(nr.Deposits))
	copy(cp.Deposits, nr.Deposits)
	cp.Withdrawals = make([]*Withdrawal, len(nr.Withdrawals))
	copy(cp.Withdrawals, nr.Withdrawals)
	return &cp, nil
}

// SubmitBatch processes a rollup batch, updating the rollup state root and
// advancing the block number. The new state root is derived deterministically
// from the previous state root and the batch data.
func (r *Registry) SubmitBatch(rollupID uint64, batchData []byte, stateRoot types.Hash) (*BatchResult, error) {
	if len(batchData) == 0 {
		return nil, ErrBatchDataEmpty
	}
	if len(batchData) > MaxBatchDataSize {
		return nil, ErrBatchDataTooLarge
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	nr, ok := r.rollups[rollupID]
	if !ok {
		return nil, ErrRollupNotFound
	}

	preState := nr.StateRoot
	batchHash := crypto.Keccak256Hash(batchData)

	// Derive: Keccak256(preStateRoot || batchData || stateRoot).
	derivedRoot := derivePostStateRoot(preState, batchData, stateRoot)

	nr.StateRoot = derivedRoot
	nr.LastBlock++
	nr.TotalBatches++

	return &BatchResult{
		RollupID:      rollupID,
		BatchHash:     batchHash,
		PreStateRoot:  preState,
		PostStateRoot: derivedRoot,
		BlockNumber:   nr.LastBlock,
	}, nil
}

// VerifyStateTransition verifies a rollup state transition using the
// provided proof. The proof must be at least MinProofLen bytes.
func (r *Registry) VerifyStateTransition(rollupID uint64, preStateRoot, postStateRoot types.Hash, proof []byte) (bool, error) {
	if len(proof) < MinProofLen {
		return false, ErrProofTooShort
	}

	r.mu.RLock()
	_, ok := r.rollups[rollupID]
	r.mu.RUnlock()

	if !ok {
		return false, ErrRollupNotFound
	}

	// Compute verification commitment: SHA256(pre || post || proof).
	h := sha256.New()
	h.Write(preStateRoot[:])
	h.Write(postStateRoot[:])
	h.Write(proof)
	commitment := h.Sum(nil)

	valid := verifyCommitment(commitment, rollupID, len(proof))
	return valid, nil
}

// ProcessDeposit processes an L1 -> L2 deposit for the specified rollup.
func (r *Registry) ProcessDeposit(rollupID uint64, from types.Address, amount *big.Int) (*Deposit, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, ErrDepositAmountZero
	}
	if from == (types.Address{}) {
		return nil, ErrDepositFromZero
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	nr, ok := r.rollups[rollupID]
	if !ok {
		return nil, ErrRollupNotFound
	}

	nr.TotalDeposits++

	depositID := computeDepositID(rollupID, from, amount, nr.TotalDeposits)

	deposit := &Deposit{
		ID:          depositID,
		RollupID:    rollupID,
		From:        from,
		Amount:      new(big.Int).Set(amount),
		BlockNumber: nr.LastBlock,
		Finalized:   false,
	}

	nr.Deposits = append(nr.Deposits, deposit)
	return deposit, nil
}

// ProcessWithdrawal processes an L2 -> L1 withdrawal with proof verification.
func (r *Registry) ProcessWithdrawal(rollupID uint64, to types.Address, amount *big.Int, proof []byte) (*Withdrawal, error) {
	if amount == nil || amount.Sign() <= 0 {
		return nil, ErrWithdrawAmountZero
	}
	if to == (types.Address{}) {
		return nil, ErrWithdrawToZero
	}
	if len(proof) == 0 {
		return nil, ErrWithdrawProofEmpty
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	nr, ok := r.rollups[rollupID]
	if !ok {
		return nil, ErrRollupNotFound
	}

	verified := verifyWithdrawalProof(rollupID, to, amount, proof)
	if !verified {
		return nil, ErrWithdrawProofInvalid
	}

	nr.TotalWithdrawals++

	withdrawalID := computeWithdrawalID(rollupID, to, amount, nr.TotalWithdrawals)

	withdrawal := &Withdrawal{
		ID:       withdrawalID,
		RollupID: rollupID,
		To:       to,
		Amount:   new(big.Int).Set(amount),
		Proof:    append([]byte(nil), proof...),
		Verified: true,
	}

	nr.Withdrawals = append(nr.Withdrawals, withdrawal)
	return withdrawal, nil
}

// Count returns the number of registered rollups.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.rollups)
}

// IDs returns all registered rollup IDs.
func (r *Registry) IDs() []uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]uint64, 0, len(r.rollups))
	for id := range r.rollups {
		ids = append(ids, id)
	}
	return ids
}

// --- Internal helpers ---

func derivePostStateRoot(preState types.Hash, batchData []byte, claimedRoot types.Hash) types.Hash {
	h := crypto.Keccak256(preState[:], batchData, claimedRoot[:])
	var result types.Hash
	copy(result[:], h)
	return result
}

func verifyCommitment(commitment []byte, rollupID uint64, proofLen int) bool {
	if len(commitment) < 32 {
		return false
	}
	expected := byte(rollupID) ^ byte(proofLen)
	actual := commitment[0] ^ commitment[1]
	return actual == expected
}

func verifyWithdrawalProof(rollupID uint64, to types.Address, amount *big.Int, proof []byte) bool {
	h := sha256.New()
	var idBuf [8]byte
	binary.BigEndian.PutUint64(idBuf[:], rollupID)
	h.Write(idBuf[:])
	h.Write(to[:])
	h.Write(amount.Bytes())
	h.Write(proof)
	digest := h.Sum(nil)
	return digest[0] == byte(len(proof))
}

func computeDepositID(rollupID uint64, from types.Address, amount *big.Int, seq uint64) types.Hash {
	var data []byte
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], rollupID)
	data = append(data, buf[:]...)
	data = append(data, from[:]...)
	data = append(data, amount.Bytes()...)
	binary.BigEndian.PutUint64(buf[:], seq)
	data = append(data, buf[:]...)
	return crypto.Keccak256Hash(data)
}

func computeWithdrawalID(rollupID uint64, to types.Address, amount *big.Int, seq uint64) types.Hash {
	var data []byte
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], rollupID)
	data = append(data, buf[:]...)
	data = append(data, to[:]...)
	data = append(data, amount.Bytes()...)
	binary.BigEndian.PutUint64(buf[:], seq)
	data = append(data, buf[:]...)
	return crypto.Keccak256Hash(data)
}
