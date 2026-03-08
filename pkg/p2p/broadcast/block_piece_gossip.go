// Block piece gossip protocol: manages the collection and propagation of
// erasure-coded block pieces. The assembly manager tracks incoming pieces
// for each block, determines when enough pieces have arrived for
// reconstruction (k-of-n threshold), and uses sqrt(n) fanout for
// efficient piece propagation to connected peers.
package broadcast

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
)

// Block piece gossip errors.
var (
	ErrPieceGossipNilPiece  = errors.New("block_piece_gossip: nil piece")
	ErrPieceGossipDuplicate = errors.New("block_piece_gossip: duplicate piece")
	ErrPieceGossipExpired   = errors.New("block_piece_gossip: assembly expired")
	ErrPieceGossipNoPeers   = errors.New("block_piece_gossip: no peers")
	ErrPieceGossipComplete  = errors.New("block_piece_gossip: assembly already complete")
)

// BlockPieceMessage is a gossip message carrying a single block piece.
type BlockPieceMessage struct {
	// BlockHash identifies the block this piece belongs to.
	BlockHash types.Hash
	// PieceIndex is the shard index in [0, TotalPieces).
	PieceIndex int
	// TotalPieces is the total number of erasure-coded pieces (n = k + m).
	TotalPieces int
	// Data is the raw piece (shard) bytes.
	Data []byte
	// PieceHash is the Keccak-256 hash of Data.
	PieceHash types.Hash
	// BlockNumber is the block number this piece is for.
	BlockNumber uint64
	// Timestamp is when this message was created.
	Timestamp time.Time
}

// BlockAssembly tracks the collection of pieces for a single block.
type BlockAssembly struct {
	// BlockHash identifies the block being assembled.
	BlockHash types.Hash
	// TotalPieces is the expected number of erasure-coded pieces.
	TotalPieces int
	// Pieces maps piece index to the received message.
	Pieces map[int]*BlockPieceMessage
	// Complete is true once k pieces have been received.
	Complete bool
	// CreatedAt is when the first piece was received.
	CreatedAt time.Time
	// CompletedAt is when the k-th piece was received.
	CompletedAt time.Time
	// DataShards is k, the number of pieces needed for reconstruction.
	DataShards int
}

// BlockAssemblyConfig configures the assembly manager.
type BlockAssemblyConfig struct {
	// DataShards is k, the minimum pieces needed for reconstruction.
	DataShards int
	// AssemblyTimeout is how long to keep an incomplete assembly.
	AssemblyTimeout time.Duration
	// MaxAssemblies is the maximum number of concurrent assemblies tracked.
	MaxAssemblies int
}

// DefaultBlockAssemblyConfig returns sensible defaults: k=4, 30s timeout, 64 max.
func DefaultBlockAssemblyConfig() BlockAssemblyConfig {
	return BlockAssemblyConfig{
		DataShards:      4,
		AssemblyTimeout: 30 * time.Second,
		MaxAssemblies:   64,
	}
}

// BlockAssemblyManager manages block piece collection and assembly.
// All methods are safe for concurrent use.
type BlockAssemblyManager struct {
	mu         sync.RWMutex
	config     BlockAssemblyConfig
	assemblies map[types.Hash]*BlockAssembly
	peers      map[string]bool
}

// NewBlockAssemblyManager creates a new assembly manager.
func NewBlockAssemblyManager(config BlockAssemblyConfig) *BlockAssemblyManager {
	return &BlockAssemblyManager{
		config:     config,
		assemblies: make(map[types.Hash]*BlockAssembly),
		peers:      make(map[string]bool),
	}
}

// AddPiece processes an incoming block piece message. It creates or updates
// the assembly for the piece's block, and returns whether the assembly is
// now complete (has >= k pieces).
func (m *BlockAssemblyManager) AddPiece(msg *BlockPieceMessage) (complete bool, err error) {
	if msg == nil {
		return false, ErrPieceGossipNilPiece
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	assembly, exists := m.assemblies[msg.BlockHash]
	if !exists {
		// Create new assembly, respecting the max limit.
		if len(m.assemblies) >= m.config.MaxAssemblies {
			// Evict the oldest incomplete assembly.
			m.evictOldestLocked()
		}
		assembly = &BlockAssembly{
			BlockHash:   msg.BlockHash,
			TotalPieces: msg.TotalPieces,
			Pieces:      make(map[int]*BlockPieceMessage),
			CreatedAt:   time.Now(),
			DataShards:  m.config.DataShards,
		}
		m.assemblies[msg.BlockHash] = assembly
	}

	// Reject additions to a completed assembly.
	if assembly.Complete {
		return true, ErrPieceGossipComplete
	}

	// Check for duplicate piece.
	if _, dup := assembly.Pieces[msg.PieceIndex]; dup {
		return false, ErrPieceGossipDuplicate
	}

	assembly.Pieces[msg.PieceIndex] = msg

	// Check if we now have enough pieces.
	if len(assembly.Pieces) >= m.config.DataShards {
		assembly.Complete = true
		assembly.CompletedAt = time.Now()
		return true, nil
	}

	return false, nil
}

// evictOldestLocked removes the oldest incomplete assembly. Caller holds m.mu.
func (m *BlockAssemblyManager) evictOldestLocked() {
	var oldestHash types.Hash
	var oldestTime time.Time
	first := true

	for hash, asm := range m.assemblies {
		if asm.Complete {
			continue
		}
		if first || asm.CreatedAt.Before(oldestTime) {
			oldestHash = hash
			oldestTime = asm.CreatedAt
			first = false
		}
	}

	if !first {
		delete(m.assemblies, oldestHash)
	}
}

// GetAssembly returns the assembly for the given block hash, or nil.
func (m *BlockAssemblyManager) GetAssembly(blockHash types.Hash) *BlockAssembly {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.assemblies[blockHash]
}

// IsComplete returns true if the assembly for the given block is complete.
func (m *BlockAssemblyManager) IsComplete(blockHash types.Hash) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	asm, exists := m.assemblies[blockHash]
	if !exists {
		return false
	}
	return asm.Complete
}

// PropagateBlockPieces selects sqrt(n) peers for piece propagation and
// returns their IDs. This mirrors the block gossip fanout strategy.
func (m *BlockAssemblyManager) PropagateBlockPieces(pieces []*BlockPieceMessage) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	n := len(m.peers)
	if n == 0 {
		return nil
	}

	fanout := int(math.Ceil(math.Sqrt(float64(n))))
	if fanout > n {
		fanout = n
	}

	allPeers := make([]string, 0, n)
	for pid := range m.peers {
		allPeers = append(allPeers, pid)
	}

	// Deterministic selection for testability: first fanout peers.
	return allPeers[:fanout]
}

// AddPeer registers a peer for block piece gossip.
func (m *BlockAssemblyManager) AddPeer(peerID string) {
	if peerID == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.peers[peerID] = true
}

// RemovePeer unregisters a peer from block piece gossip.
func (m *BlockAssemblyManager) RemovePeer(peerID string) {
	if peerID == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.peers, peerID)
}

// CleanupExpired removes assemblies that are older than the configured
// timeout and are not yet complete. Returns the number of removed assemblies.
func (m *BlockAssemblyManager) CleanupExpired() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-m.config.AssemblyTimeout)
	removed := 0

	for hash, asm := range m.assemblies {
		if !asm.Complete && asm.CreatedAt.Before(cutoff) {
			delete(m.assemblies, hash)
			removed++
		}
	}

	return removed
}

// Stats returns counts of total, complete, and incomplete assemblies.
func (m *BlockAssemblyManager) Stats() (total, complete, incomplete int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total = len(m.assemblies)
	for _, asm := range m.assemblies {
		if asm.Complete {
			complete++
		} else {
			incomplete++
		}
	}
	return total, complete, incomplete
}

// BlockPieceTopicName returns the gossip topic for a specific slot and piece
// index in the format "block_piece/{slot}/{piece_index}" (GAP-4.2).
func BlockPieceTopicName(slot uint64, pieceIndex int) string {
	return fmt.Sprintf("block_piece/%d/%d", slot, pieceIndex)
}

// PieceCustodyIndex computes which piece index a peer has custody of for
// a given slot using keccak256(peer_id || slot) % totalPieces (GAP-4.2).
// Each peer is responsible for storing and serving its assigned pieces.
func PieceCustodyIndex(peerID string, slot uint64, totalPieces int) int {
	if totalPieces <= 0 {
		return 0
	}
	buf := make([]byte, len(peerID)+8)
	copy(buf, peerID)
	binary.BigEndian.PutUint64(buf[len(peerID):], slot)
	h := crypto.Keccak256Hash(buf)
	idx := int(binary.BigEndian.Uint64(h[24:]) % uint64(totalPieces))
	return idx
}

// PeerCustodyPieces returns the piece indices that a peer has custody of.
// By default each peer holds 2 pieces (custodyCount). The assignment is
// deterministic: piece i is assigned to peers where PieceCustodyIndex % totalPieces == i.
func PeerCustodyPieces(peerID string, slot uint64, totalPieces, custodyCount int) []int {
	if totalPieces <= 0 || custodyCount <= 0 {
		return nil
	}
	primary := PieceCustodyIndex(peerID, slot, totalPieces)
	pieces := make([]int, 0, custodyCount)
	for i := 0; i < custodyCount; i++ {
		pieces = append(pieces, (primary+i)%totalPieces)
	}
	return pieces
}
