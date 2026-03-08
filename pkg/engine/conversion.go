package engine

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/payload"
)

// PayloadToHeader converts an ExecutionPayloadV4 to a block Header.
func PayloadToHeader(p *ExecutionPayloadV4) *types.Header {
	return payload.PayloadToHeader(p)
}

// HeaderToPayloadFields extracts common payload fields from a Header.
func HeaderToPayloadFields(header *types.Header) ExecutionPayloadV1 {
	return payload.HeaderToPayloadFields(header)
}

// WithdrawalsToEngine converts core Withdrawal types to engine Withdrawal types.
func WithdrawalsToEngine(ws []*types.Withdrawal) []*Withdrawal {
	return payload.WithdrawalsToEngine(ws)
}

// WithdrawalsToCore converts engine Withdrawal types to core Withdrawal types.
func WithdrawalsToCore(ws []*Withdrawal) []*types.Withdrawal {
	return payload.WithdrawalsToCore(ws)
}
