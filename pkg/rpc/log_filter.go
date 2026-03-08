package rpc

// log_filter.go re-exports log filter types from rpc/filter for backward
// compatibility.

import (
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
)

// Re-export log filter types.
type (
	LogFilterConfig = rpcfilter.LogFilterConfig
	LogFilterSpec   = rpcfilter.LogFilterSpec
	FilteredLog     = rpcfilter.FilteredLog
	LogFilterEngine = rpcfilter.LogFilterEngine
)

// Re-export constructors and functions.
var (
	DefaultLogFilterConfig = rpcfilter.DefaultLogFilterConfig
	NewLogFilterEngine     = rpcfilter.NewLogFilterEngine
	MatchesFilter          = rpcfilter.MatchesFilter
)
