// Block-level erasure coding: k-of-n encoding/decoding of execution blocks
// using Reed-Solomon over GF(2^8). This enables block recovery from any k
// of n pieces, improving availability under partial network partitions.
//
// Encoder splits a block into n erasure-coded pieces (k data + m parity).
// Decoder reconstructs the original block from any k pieces.
package blockerasure

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/das/erasure"
)

// Block erasure parameters for the k=8, n=16 standard config (GAP-4.1).
const (
	// StandardDataShards is the number of data shards for block erasure (k=8).
	StandardDataShards = 8
	// StandardParityShards is the number of parity shards (m=8), giving n=16 total.
	StandardParityShards = 8
)

// Block erasure errors.
var (
	ErrBlockErasureEmpty      = errors.New("block_erasure: empty block data")
	ErrBlockErasureTooLarge   = errors.New("block_erasure: block data exceeds max size")
	ErrBlockErasureNilEncoder = errors.New("block_erasure: nil RS encoder")
	ErrBlockPieceInvalid      = errors.New("block_erasure: invalid piece index")
	ErrBlockPieceHashMismatch = errors.New("block_erasure: piece hash mismatch")
	ErrBlockPieceDuplicate    = errors.New("block_erasure: duplicate piece")
	ErrInsufficientPieces     = errors.New("block_erasure: insufficient pieces for reconstruction")
	ErrBlockReconstructFailed = errors.New("block_erasure: reconstruction failed")
)

// DefaultMaxBlockSize is 10 MiB, matching the P2P maximum payload.
const DefaultMaxBlockSize = 10 * 1024 * 1024

// BlockErasureConfig configures block-level erasure coding.
type BlockErasureConfig struct {
	// DataShards is k: the minimum number of pieces needed for reconstruction.
	DataShards int
	// ParityShards is m: the number of additional redundancy pieces.
	ParityShards int
	// MaxBlockSize is the maximum block size in bytes.
	MaxBlockSize uint64
}

// DefaultBlockErasureConfig returns the default configuration: k=4, m=4, 10 MB max.
func DefaultBlockErasureConfig() BlockErasureConfig {
	return BlockErasureConfig{
		DataShards:   4,
		ParityShards: 4,
		MaxBlockSize: DefaultMaxBlockSize,
	}
}

// StandardBlockErasureConfig returns the k=8, n=16 block erasure configuration
// as specified in GAP-4.1. Any 8 of the 16 pieces suffice for reconstruction.
func StandardBlockErasureConfig() BlockErasureConfig {
	return BlockErasureConfig{
		DataShards:   StandardDataShards,
		ParityShards: StandardParityShards,
		MaxBlockSize: DefaultMaxBlockSize,
	}
}

// BlockPiece is a single piece of an erasure-coded block.
type BlockPiece struct {
	// Index is the piece index in [0, TotalPieces).
	Index int
	// Data is the erasure-coded shard bytes.
	Data []byte
	// BlockHash is the Keccak-256 hash of the original block data.
	BlockHash types.Hash
	// BlockSize is the original block size in bytes.
	BlockSize uint64
	// TotalPieces is the total number of pieces (n = k + m).
	TotalPieces int
	// PieceHash is the Keccak-256 hash of this piece's Data.
	PieceHash types.Hash
}

// BlockErasureEncoder encodes blocks into erasure-coded pieces.
type BlockErasureEncoder struct {
	mu     sync.RWMutex
	config BlockErasureConfig
	enc    *erasure.RSEncoderGF256
}

// NewBlockErasureEncoder creates a new encoder with the given configuration.
// Returns an error if the RS encoder cannot be initialised.
func NewBlockErasureEncoder(config BlockErasureConfig) (*BlockErasureEncoder, error) {
	enc, err := erasure.NewRSEncoderGF256(config.DataShards, config.ParityShards)
	if err != nil {
		return nil, fmt.Errorf("block_erasure: %w", err)
	}
	return &BlockErasureEncoder{
		config: config,
		enc:    enc,
	}, nil
}

// Encode splits blockData into n erasure-coded BlockPieces.
// Any k of the returned pieces suffice for reconstruction.
func (e *BlockErasureEncoder) Encode(blockData []byte) ([]*BlockPiece, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(blockData) == 0 {
		return nil, ErrBlockErasureEmpty
	}
	if uint64(len(blockData)) > e.config.MaxBlockSize {
		return nil, fmt.Errorf("%w: %d > %d",
			ErrBlockErasureTooLarge, len(blockData), e.config.MaxBlockSize)
	}

	blockHash := crypto.Keccak256Hash(blockData)
	blockSize := uint64(len(blockData))

	shards, err := e.enc.Encode(blockData)
	if err != nil {
		return nil, fmt.Errorf("block_erasure: encode failed: %w", err)
	}

	totalPieces := e.enc.TotalShards()
	pieces := make([]*BlockPiece, totalPieces)
	for i := 0; i < totalPieces; i++ {
		pieceHash := crypto.Keccak256Hash(shards[i])
		pieces[i] = &BlockPiece{
			Index:       i,
			Data:        shards[i],
			BlockHash:   blockHash,
			BlockSize:   blockSize,
			TotalPieces: totalPieces,
			PieceHash:   pieceHash,
		}
	}

	return pieces, nil
}

// Config returns the encoder's configuration.
func (e *BlockErasureEncoder) Config() BlockErasureConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config
}

// BlockErasureDecoder reassembles blocks from erasure-coded pieces.
type BlockErasureDecoder struct {
	mu     sync.RWMutex
	config BlockErasureConfig
	enc    *erasure.RSEncoderGF256
}

// NewBlockErasureDecoder creates a new decoder with the given configuration.
func NewBlockErasureDecoder(config BlockErasureConfig) (*BlockErasureDecoder, error) {
	enc, err := erasure.NewRSEncoderGF256(config.DataShards, config.ParityShards)
	if err != nil {
		return nil, fmt.Errorf("block_erasure: %w", err)
	}
	return &BlockErasureDecoder{
		config: config,
		enc:    enc,
	}, nil
}

// Decode reconstructs the original block data from a set of pieces.
// At least k (DataShards) valid pieces with consistent metadata are required.
func (d *BlockErasureDecoder) Decode(pieces []*BlockPiece) ([]byte, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(pieces) < d.config.DataShards {
		return nil, fmt.Errorf("%w: have %d, need %d",
			ErrInsufficientPieces, len(pieces), d.config.DataShards)
	}

	// Validate consistency: all pieces must share the same BlockHash,
	// TotalPieces, and BlockSize.
	refHash := pieces[0].BlockHash
	refTotal := pieces[0].TotalPieces
	refSize := pieces[0].BlockSize

	seen := make(map[int]bool)
	for _, p := range pieces {
		if p.BlockHash != refHash {
			return nil, fmt.Errorf("block_erasure: mismatched block hash across pieces")
		}
		if p.TotalPieces != refTotal {
			return nil, fmt.Errorf("block_erasure: mismatched total pieces across pieces")
		}
		if p.BlockSize != refSize {
			return nil, fmt.Errorf("block_erasure: mismatched block size across pieces")
		}
		if p.Index < 0 || p.Index >= refTotal {
			return nil, fmt.Errorf("%w: %d not in [0, %d)",
				ErrBlockPieceInvalid, p.Index, refTotal)
		}
		// Verify piece hash integrity.
		computed := crypto.Keccak256Hash(p.Data)
		if computed != p.PieceHash {
			return nil, fmt.Errorf("%w: piece %d",
				ErrBlockPieceHashMismatch, p.Index)
		}
		if seen[p.Index] {
			return nil, fmt.Errorf("%w: piece %d",
				ErrBlockPieceDuplicate, p.Index)
		}
		seen[p.Index] = true
	}

	// Build the shards slice for the RS decoder. Missing shards are nil.
	shards := make([][]byte, refTotal)
	for _, p := range pieces {
		shards[p.Index] = p.Data
	}

	data, err := d.enc.ReconstructData(shards)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBlockReconstructFailed, err)
	}

	// Trim to original block size.
	if uint64(len(data)) < refSize {
		return nil, fmt.Errorf("%w: reconstructed %d bytes, expected %d",
			ErrBlockReconstructFailed, len(data), refSize)
	}
	data = data[:refSize]

	// Verify the reconstructed block hash matches.
	computed := crypto.Keccak256Hash(data)
	if computed != refHash {
		return nil, fmt.Errorf("%w: hash mismatch after reconstruction",
			ErrBlockReconstructFailed)
	}

	return data, nil
}

// CanDecode returns true if there are at least k valid, non-duplicate pieces
// with consistent metadata.
func (d *BlockErasureDecoder) CanDecode(pieces []*BlockPiece) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(pieces) < d.config.DataShards {
		return false
	}

	refHash := pieces[0].BlockHash
	refTotal := pieces[0].TotalPieces
	refSize := pieces[0].BlockSize

	seen := make(map[int]bool)
	validCount := 0
	for _, p := range pieces {
		if p.BlockHash != refHash || p.TotalPieces != refTotal || p.BlockSize != refSize {
			continue
		}
		if p.Index < 0 || p.Index >= refTotal {
			continue
		}
		if seen[p.Index] {
			continue
		}
		computed := crypto.Keccak256Hash(p.Data)
		if computed != p.PieceHash {
			continue
		}
		seen[p.Index] = true
		validCount++
	}

	return validCount >= d.config.DataShards
}

// BlockAssemblyManager tracks erasure-coded block pieces per slot.
// When k=8 pieces arrive for the same slot, it triggers reconstruction.
// If only k-1 pieces arrive within the timeout, a fallback flag is set to
// signal that the full block should be downloaded instead (GAP-4.3).
type BlockAssemblyManager struct {
	mu         sync.Mutex
	dataShards int
	timeout    time.Duration
	slots      map[uint64]*slotAssembly
}

// slotAssembly tracks pieces for a single slot.
type slotAssembly struct {
	pieces    map[int]*BlockPiece
	blockHash types.Hash
	complete  bool
	timeout   bool
	createdAt time.Time
}

// BlockAssemblyManagerConfig configures the BlockAssemblyManager.
type BlockAssemblyManagerConfig struct {
	// DataShards is k: minimum pieces needed for reconstruction.
	DataShards int
	// PieceTimeout is the deadline for collecting k pieces before fallback.
	PieceTimeout time.Duration
}

// DefaultBlockAssemblyManagerConfig returns the standard k=8, 2s timeout config.
func DefaultBlockAssemblyManagerConfig() BlockAssemblyManagerConfig {
	return BlockAssemblyManagerConfig{
		DataShards:   StandardDataShards,
		PieceTimeout: 2 * time.Second,
	}
}

// NewBlockAssemblyManager creates a BlockAssemblyManager for slot-based tracking.
func NewBlockAssemblyManager(cfg BlockAssemblyManagerConfig) *BlockAssemblyManager {
	if cfg.DataShards <= 0 {
		cfg.DataShards = StandardDataShards
	}
	if cfg.PieceTimeout == 0 {
		cfg.PieceTimeout = 2 * time.Second
	}
	return &BlockAssemblyManager{
		dataShards: cfg.DataShards,
		timeout:    cfg.PieceTimeout,
		slots:      make(map[uint64]*slotAssembly),
	}
}

// AddPiece records a received block piece for the given slot. Returns (true,
// nil) when k pieces have arrived and reconstruction is possible. Returns
// (false, nil) when more pieces are needed. Returns (false, err) on invalid
// input (nil piece, duplicate piece index).
func (m *BlockAssemblyManager) AddPiece(slot uint64, piece *BlockPiece) (ready bool, err error) {
	if piece == nil {
		return false, ErrBlockErasureNilEncoder
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	asm, exists := m.slots[slot]
	if !exists {
		asm = &slotAssembly{
			pieces:    make(map[int]*BlockPiece),
			blockHash: piece.BlockHash,
			createdAt: time.Now(),
		}
		m.slots[slot] = asm
	}

	if asm.complete {
		return true, nil
	}

	if _, dup := asm.pieces[piece.Index]; dup {
		return false, ErrBlockPieceDuplicate
	}
	asm.pieces[piece.Index] = piece

	if len(asm.pieces) >= m.dataShards {
		asm.complete = true
		return true, nil
	}

	return false, nil
}

// IsReady returns true if k pieces have been collected for the slot.
func (m *BlockAssemblyManager) IsReady(slot uint64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	asm, ok := m.slots[slot]
	return ok && asm.complete
}

// ShouldFallback returns true if the assembly timeout has elapsed without
// reaching k pieces, indicating a full block download is needed (GAP-4.3).
func (m *BlockAssemblyManager) ShouldFallback(slot uint64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	asm, ok := m.slots[slot]
	if !ok {
		return false
	}
	if asm.complete {
		return false
	}
	return time.Since(asm.createdAt) >= m.timeout
}

// Pieces returns the collected pieces for a slot, or nil if the slot is unknown.
func (m *BlockAssemblyManager) Pieces(slot uint64) []*BlockPiece {
	m.mu.Lock()
	defer m.mu.Unlock()
	asm, ok := m.slots[slot]
	if !ok {
		return nil
	}
	result := make([]*BlockPiece, 0, len(asm.pieces))
	for _, p := range asm.pieces {
		result = append(result, p)
	}
	return result
}

// ClearSlot removes all piece data for the given slot once it has been
// assembled or timed out (GAP-4.3).
func (m *BlockAssemblyManager) ClearSlot(slot uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.slots, slot)
}

// PieceCount returns the number of pieces received so far for a slot.
func (m *BlockAssemblyManager) PieceCount(slot uint64) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if asm, ok := m.slots[slot]; ok {
		return len(asm.pieces)
	}
	return 0
}
