// filter_sys.go re-exports FilterSys types from rpc/filter for backward
// compatibility.
package rpc

import (
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
)

// Re-export FilterSys errors.
var (
	ErrSysFilterNotFound    = rpcfilter.ErrSysFilterNotFound
	ErrSysFilterWrongKind   = rpcfilter.ErrSysFilterWrongKind
	ErrSysFilterCapacity    = rpcfilter.ErrSysFilterCapacity
	ErrSysFilterClosed      = rpcfilter.ErrSysFilterClosed
	ErrSysInvalidBlockRange = rpcfilter.ErrSysInvalidBlockRange
	ErrSysTopicOverflow     = rpcfilter.ErrSysTopicOverflow
)

// Re-export FilterSys types.
type (
	SysFilterKind   = rpcfilter.SysFilterKind
	FilterSysConfig = rpcfilter.FilterSysConfig
	SysLogQuery     = rpcfilter.SysLogQuery
	FilterSys       = rpcfilter.FilterSys
)

// Re-export FilterSys constants.
const (
	SysLogFilter       = rpcfilter.SysLogFilter
	SysBlockFilter     = rpcfilter.SysBlockFilter
	SysPendingTxFilter = rpcfilter.SysPendingTxFilter
)

// Re-export constructors and functions.
var (
	DefaultFilterSysConfig = rpcfilter.DefaultFilterSysConfig
	NewFilterSys           = rpcfilter.NewFilterSys
	FilterLogsByBloom      = rpcfilter.FilterLogsByBloom
	BloomMatchesQuery      = rpcfilter.BloomMatchesQuery
	SysLogMatches          = rpcfilter.SysLogMatches
)
