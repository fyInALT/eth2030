// Block-level erasure coding: k-of-n encoding/decoding of execution blocks
// using Reed-Solomon over GF(2^8). This enables block recovery from any k
// of n pieces, improving availability under partial network partitions.
//
// Encoder splits a block into n erasure-coded pieces (k data + m parity).
// Decoder reconstructs the original block from any k pieces.
package das

import (
	"errors"
	"fmt"
	"sync"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/das/erasure"
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
