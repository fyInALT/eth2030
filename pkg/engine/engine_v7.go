// engine_v7.go re-exports Engine API V7 symbols from engine/api sub-package
// and provides backward-compatible unexported wrappers for package-internal tests.
package engine

import (
	engapi "github.com/eth2030/eth2030/engine/api"
)

// EngineV7Backend defines the backend interface for Engine API V7.
// This is a re-declaration of the interface in engine/api so callers in
// package engine can implement it without importing engine/api directly.
type EngineV7Backend interface {
	NewPayloadV7(payload *ExecutionPayloadV7) (*PayloadStatusV1, error)
	ForkchoiceUpdatedV7(state *ForkchoiceStateV1, attrs *PayloadAttributesV7) (*ForkchoiceUpdatedResult, error)
	GetPayloadV7(id PayloadID) (*ExecutionPayloadV7, error)
}

// engineV7Wrapper wraps engapi.EngineV7 and exposes the backend field
// to package-internal tests via the backend accessor pattern.
type engineV7Wrapper struct {
	inner   *engapi.EngineV7
	backend EngineV7Backend
}

// HandleNewPayloadV7 delegates to the inner engapi.EngineV7.
func (e *engineV7Wrapper) HandleNewPayloadV7(p *ExecutionPayloadV7) (*PayloadStatusV1, error) {
	return e.inner.HandleNewPayloadV7(p)
}

// HandleForkchoiceUpdatedV7 delegates to the inner engapi.EngineV7.
func (e *engineV7Wrapper) HandleForkchoiceUpdatedV7(state *ForkchoiceStateV1, attrs *PayloadAttributesV7) (*ForkchoiceUpdatedResult, error) {
	return e.inner.HandleForkchoiceUpdatedV7(state, attrs)
}

// HandleGetPayloadV7 delegates to the inner engapi.EngineV7.
func (e *engineV7Wrapper) HandleGetPayloadV7(payloadID PayloadID) (*ExecutionPayloadV7, error) {
	return e.inner.HandleGetPayloadV7(payloadID)
}

// NewEngineV7 creates a new Engine API V7 handler.
func NewEngineV7(backend EngineV7Backend) *engineV7Wrapper {
	bridge := &engineV7Bridge{inner: backend}
	return &engineV7Wrapper{
		inner:   engapi.NewEngineV7(bridge),
		backend: backend,
	}
}

// engineV7Bridge adapts the engine-package EngineV7Backend to the
// engapi.EngineV7Backend interface required by EngineV7.
type engineV7Bridge struct {
	inner EngineV7Backend
}

func (b *engineV7Bridge) NewPayloadV7(p *ExecutionPayloadV7) (*PayloadStatusV1, error) {
	return b.inner.NewPayloadV7(p)
}

func (b *engineV7Bridge) ForkchoiceUpdatedV7(state *ForkchoiceStateV1, attrs *PayloadAttributesV7) (*ForkchoiceUpdatedResult, error) {
	return b.inner.ForkchoiceUpdatedV7(state, attrs)
}

func (b *engineV7Bridge) GetPayloadV7(id PayloadID) (*ExecutionPayloadV7, error) {
	return b.inner.GetPayloadV7(id)
}
