package block

import (
	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
)

// Validator validates block headers, bodies, and state.
type Validator interface {
	ValidateHeader(header, parent *types.Header) error
	ValidateBody(block *types.Block) error
	ValidateRequests(header *types.Header, requests types.Requests) error
	ValidateBlockAccessList(header *types.Header, computedBALHash *types.Hash) error
}

// BlockchainReader provides read-only access to the blockchain needed
// by block building. It avoids a circular dependency between block/ and chain/.
type BlockchainReader interface {
	Config() *config.ChainConfig
	CurrentBlock() *types.Block
	Genesis() *types.Block
	GetBlock(hash types.Hash) *types.Block
	StateAtBlock(block *types.Block) (state.StateDB, error)
	// GetHashFn returns a function that maps block numbers to canonical hashes,
	// used by the BLOCKHASH opcode so the builder matches the verifier.
	GetHashFn() func(uint64) types.Hash
}

// Compile-time checks.
var _ Validator = (*BlockValidator)(nil)
