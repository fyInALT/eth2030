package rpc

// filter_system.go re-exports filter system types from rpc/filter for
// backward compatibility.

import (
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
)

// Re-export filter config and types.
type (
	FilterConfig  = rpcfilter.FilterConfig
	FSLogFilter   = rpcfilter.FSLogFilter
	FSBlockFilter = rpcfilter.FSBlockFilter
	FilterSystem  = rpcfilter.FilterSystem
)

// Re-export constructors.
var (
	DefaultFilterConfig = rpcfilter.DefaultFilterConfig
	NewFilterSystem     = rpcfilter.NewFilterSystem
)
