package engine

import (
	"github.com/eth2030/eth2030/core/types"
	engapi "github.com/eth2030/eth2030/engine/api"
)

// GlamsterdamBackend defines the backend interface for post-Glamsterdam Engine API.
// This is an alias of the interface defined in engine/api — redeclared here
// so callers in package engine can implement it without importing engine/api.
type GlamsterdamBackend interface {
	NewPayloadV5(payload *ExecutionPayloadV5,
		expectedBlobVersionedHashes []types.Hash,
		parentBeaconBlockRoot types.Hash,
		executionRequests [][]byte) (*engapi.PayloadStatusV1, error)

	ForkchoiceUpdatedV4G(state *engapi.ForkchoiceStateV1, attrs *GlamsterdamPayloadAttributes) (*engapi.ForkchoiceUpdatedResult, error)

	GetPayloadV5(id PayloadID) (*GetPayloadV5Response, error)

	GetBlobsV2(versionedHashes []types.Hash) ([]*BlobAndProofV2, error)
}

// glamsterdamBridge adapts the engine-package GlamsterdamBackend to the
// engapi.GlamsterdamBackend interface required by EngineGlamsterdam.
type glamsterdamBridge struct {
	inner GlamsterdamBackend
}

func (b *glamsterdamBridge) NewPayloadV5(
	p *ExecutionPayloadV5,
	hashes []types.Hash,
	root types.Hash,
	reqs [][]byte,
) (*engapi.PayloadStatusV1, error) {
	return b.inner.NewPayloadV5(p, hashes, root, reqs)
}

func (b *glamsterdamBridge) ForkchoiceUpdatedV4G(
	state *engapi.ForkchoiceStateV1,
	attrs *GlamsterdamPayloadAttributes,
) (*engapi.ForkchoiceUpdatedResult, error) {
	return b.inner.ForkchoiceUpdatedV4G(state, attrs)
}

func (b *glamsterdamBridge) GetPayloadV5(id PayloadID) (*GetPayloadV5Response, error) {
	return b.inner.GetPayloadV5(id)
}

func (b *glamsterdamBridge) GetBlobsV2(hashes []types.Hash) ([]*BlobAndProofV2, error) {
	return b.inner.GetBlobsV2(hashes)
}

// NewEngineGlamsterdam creates a new post-Glamsterdam Engine API handler.
func NewEngineGlamsterdam(backend GlamsterdamBackend) *EngineGlamsterdam {
	return engapi.NewEngineGlamsterdam(&glamsterdamBridge{inner: backend})
}
