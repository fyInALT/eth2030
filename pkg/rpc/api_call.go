package rpc

// api_call.go re-exports call-related types from rpc/ethapi.

import "github.com/eth2030/eth2030/rpc/ethapi"

// CallRequest is re-exported from ethapi.
type CallRequest = ethapi.CallRequest

// StateOverride is re-exported from ethapi.
type StateOverride = ethapi.StateOverride

// AccountOverride is re-exported from ethapi.
type AccountOverride = ethapi.AccountOverride

// RevertError is re-exported from ethapi.
type RevertError = ethapi.RevertError

// BlockNumberOrHashParam is re-exported from ethapi.
type BlockNumberOrHashParam = ethapi.BlockNumberOrHashParam

// StateOverrideApplier is re-exported from ethapi.
type StateOverrideApplier = ethapi.StateOverrideApplier

// ErrCodeExecution is re-exported from ethapi.
const ErrCodeExecution = ethapi.ErrCodeExecution

// NewStateOverrideApplier is re-exported from ethapi.
var NewStateOverrideApplier = ethapi.NewStateOverrideApplier
