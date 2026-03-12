// Package backendapi defines backend interfaces for the Engine API.
// Sub-packages (engine/api, engine/blocks, etc.) import from here
// to avoid circular dependencies with the main engine package.
package backendapi

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/payload"
)

// Backend defines the interface that the execution layer must implement
// for the Engine API to interact with it.
type Backend interface {
	// ProcessBlock validates and executes a new payload from the consensus layer.
	ProcessBlock(p *payload.ExecutionPayloadV3, expectedBlobVersionedHashes []types.Hash, parentBeaconBlockRoot types.Hash) (payload.PayloadStatusV1, error)

	// ProcessBlockV4 validates and executes a Prague payload with execution requests.
	ProcessBlockV4(p *payload.ExecutionPayloadV3, expectedBlobVersionedHashes []types.Hash, parentBeaconBlockRoot types.Hash, executionRequests [][]byte) (payload.PayloadStatusV1, error)

	// ProcessBlockV5 validates and executes a new Amsterdam payload with BAL.
	ProcessBlockV5(p *payload.ExecutionPayloadV5, expectedBlobVersionedHashes []types.Hash, parentBeaconBlockRoot types.Hash, executionRequests [][]byte) (payload.PayloadStatusV1, error)

	// ForkchoiceUpdated processes a forkchoice state update from the consensus layer.
	ForkchoiceUpdated(state payload.ForkchoiceStateV1, attrs *payload.PayloadAttributesV3) (payload.ForkchoiceUpdatedResult, error)

	// ForkchoiceUpdatedV4 processes a forkchoice update with V4 payload attributes (Amsterdam).
	ForkchoiceUpdatedV4(state payload.ForkchoiceStateV1, attrs *payload.PayloadAttributesV4) (payload.ForkchoiceUpdatedResult, error)

	// GetPayloadByID retrieves a previously requested payload by its ID.
	GetPayloadByID(id payload.PayloadID) (*payload.GetPayloadResponse, error)

	// GetPayloadV4ByID retrieves a previously built payload for getPayloadV4 (Prague).
	GetPayloadV4ByID(id payload.PayloadID) (*payload.GetPayloadV4Response, error)

	// GetPayloadV6ByID retrieves a previously built payload for getPayloadV6 (Amsterdam).
	GetPayloadV6ByID(id payload.PayloadID) (*payload.GetPayloadV6Response, error)

	// GetHeadTimestamp returns the timestamp of the current head block.
	GetHeadTimestamp() uint64

	// GetBlockTimestamp returns the timestamp of the block with the given hash,
	// or 0 if the block is not known.
	GetBlockTimestamp(hash types.Hash) uint64

	// IsCancun returns true if the given timestamp falls within the Cancun fork.
	IsCancun(timestamp uint64) bool

	// IsPrague returns true if the given timestamp falls within the Prague fork.
	IsPrague(timestamp uint64) bool

	// IsAmsterdam returns true if the given timestamp falls within the Amsterdam fork.
	IsAmsterdam(timestamp uint64) bool

	// GetHeadHash returns the current canonical head block hash.
	GetHeadHash() types.Hash

	// GetSafeHash returns the current safe (justified) block hash.
	GetSafeHash() types.Hash

	// GetFinalizedHash returns the current finalized block hash.
	GetFinalizedHash() types.Hash
}

// InclusionListBackend extends Backend with inclusion list support (EIP-7805 FOCIL).
type InclusionListBackend interface {
	// ProcessInclusionList validates and stores a new inclusion list from the CL.
	ProcessInclusionList(il *types.InclusionList) error

	// GetInclusionList generates an inclusion list from the mempool.
	GetInclusionList() *types.InclusionList
}

// V4Backend defines the minimal backend interface required by EngV4.
type V4Backend interface {
	// GetPayloadV4ByID retrieves a previously built payload for getPayloadV4 (Prague).
	GetPayloadV4ByID(id payload.PayloadID) (*payload.GetPayloadV4Response, error)
	// IsPrague returns true if the given timestamp falls within the Prague fork.
	IsPrague(timestamp uint64) bool
}

// GlamsterdamBackend defines the backend interface for post-Glamsterdam Engine API.
type GlamsterdamBackend interface {
	// NewPayloadV5 validates and executes a post-Glamsterdam payload.
	NewPayloadV5(p *payload.ExecutionPayloadV5,
		expectedBlobVersionedHashes []types.Hash,
		parentBeaconBlockRoot types.Hash,
		executionRequests [][]byte) (*payload.PayloadStatusV1, error)

	// ForkchoiceUpdatedV4G processes a forkchoice update with V4 attributes.
	ForkchoiceUpdatedV4G(state *payload.ForkchoiceStateV1, attrs *payload.GlamsterdamPayloadAttributes) (*payload.ForkchoiceUpdatedResult, error)

	// GetPayloadV5 retrieves a previously built payload by ID.
	GetPayloadV5(id payload.PayloadID) (*payload.GetPayloadV5Response, error)

	// GetBlobsV2 retrieves blobs by versioned hashes from the blob pool.
	GetBlobsV2(versionedHashes []types.Hash) ([]*payload.BlobAndProofV2, error)
}

// EngineV7Backend defines the backend interface for Engine API V7.
type EngineV7Backend interface {
	// NewPayloadV7 validates and executes a 2028-era payload.
	NewPayloadV7(p *payload.ExecutionPayloadV7) (*payload.PayloadStatusV1, error)

	// ForkchoiceUpdatedV7 processes a forkchoice update with V7 attributes.
	ForkchoiceUpdatedV7(state *payload.ForkchoiceStateV1, attrs *payload.PayloadAttributesV7) (*payload.ForkchoiceUpdatedResult, error)

	// GetPayloadV7 retrieves a previously built V7 payload by ID.
	GetPayloadV7(id payload.PayloadID) (*payload.ExecutionPayloadV7, error)
}

// PayloadBodiesBackend exposes block body retrieval for engine_getPayloadBodies*.
// Backends that support payload body queries implement this interface.
type PayloadBodiesBackend interface {
	// GetPayloadBodiesByHash returns payload bodies for the given block hashes.
	// Entries for unknown or out-of-retention-window blocks are nil.
	GetPayloadBodiesByHash(hashes []types.Hash) ([]*payload.ExecutionPayloadBodyV2, error)

	// GetPayloadBodiesByRange returns payload bodies for a range of block numbers
	// starting at start for count blocks. Entries outside the retention window are nil.
	GetPayloadBodiesByRange(start, count uint64) ([]*payload.ExecutionPayloadBodyV2, error)
}

// BlobsV1Backend provides blob retrieval by versioned hash for engine_getBlobsV1.
// Backends that store blob sidecar data implement this interface.
type BlobsV1Backend interface {
	// GetBlobsByVersionedHashes returns blob data for each requested versioned hash.
	// Returns nil for each entry not found in the txpool (EIP-4844 behaviour).
	GetBlobsByVersionedHashes(hashes []types.Hash) []*payload.BlobAndProofV1
}

// UncoupledBackend is the backend interface for EIP-7898 uncoupled execution payloads.
// Implementations provide uncoupled payload handling; left minimal for extensibility.
type UncoupledBackend interface{} //nolint:revive
