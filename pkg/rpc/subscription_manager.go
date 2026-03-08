// subscription_manager.go re-exports SubRegistry types from rpc/subscription
// for backward compatibility.
package rpc

import (
	rpcsub "github.com/eth2030/eth2030/rpc/subscription"
)

// Re-export subscription manager errors.
var (
	ErrSubManagerClosed     = rpcsub.ErrSubManagerClosed
	ErrSubManagerCapacity   = rpcsub.ErrSubManagerCapacity
	ErrSubManagerNotFound   = rpcsub.ErrSubManagerNotFound
	ErrSubManagerRateLimit  = rpcsub.ErrSubManagerRateLimit
	ErrSubManagerInvalidTyp = rpcsub.ErrSubManagerInvalidTyp
)

// Re-export SubKind and SubEntry types.
type (
	SubKind            = rpcsub.SubKind
	SubEntry           = rpcsub.SubEntry
	SubRateLimitConfig = rpcsub.SubRateLimitConfig
	SubRegistry        = rpcsub.SubRegistry
)

// Re-export SubKind constants.
const (
	SubKindNewHeads   = rpcsub.SubKindNewHeads
	SubKindLogs       = rpcsub.SubKindLogs
	SubKindPendingTx  = rpcsub.SubKindPendingTx
	SubKindSyncStatus = rpcsub.SubKindSyncStatus
)

// Re-export constructors and functions.
var (
	DefaultSubRateLimitConfig = rpcsub.DefaultSubRateLimitConfig
	NewSubRegistry            = rpcsub.NewSubRegistry
	ParseSubKind              = rpcsub.ParseSubKind
)
