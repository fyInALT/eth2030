package rpc

// filter_extended.go re-exports extended filter manager types from rpc/filter
// for backward compatibility.

import (
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
)

// Re-export extended filter manager errors.
var (
	ErrFilterNotFound      = rpcfilter.ErrFilterNotFound
	ErrFilterWrongType     = rpcfilter.ErrFilterWrongType
	ErrFilterLimitReached  = rpcfilter.ErrFilterLimitReached
	ErrFilterExpired       = rpcfilter.ErrFilterExpired
	ErrFilterTopicMismatch = rpcfilter.ErrFilterTopicMismatch
	ErrFilterBlockRange    = rpcfilter.ErrFilterBlockRange
	ErrFilterLogOverflow   = rpcfilter.ErrFilterLogOverflow
)

// Re-export extended filter types.
type (
	ExtFilterType    = rpcfilter.ExtFilterType
	ExtFilterConfig  = rpcfilter.ExtFilterConfig
	ExtFilter        = rpcfilter.ExtFilter
	ExtFilterManager = rpcfilter.ExtFilterManager
)

const MaxTopicPositions = rpcfilter.MaxTopicPositions

// Re-export extended filter constants.
const (
	ExtLogFilter       = rpcfilter.ExtLogFilter
	ExtBlockFilter     = rpcfilter.ExtBlockFilter
	ExtPendingTxFilter = rpcfilter.ExtPendingTxFilter
)

// Re-export constructors and functions.
var (
	DefaultExtFilterConfig = rpcfilter.DefaultExtFilterConfig
	NewExtFilterManager    = rpcfilter.NewExtFilterManager
	MatchesExtFilter       = rpcfilter.MatchesExtFilter
)
