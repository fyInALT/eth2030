// eth_api_debug.go re-exports debug namespace types from rpc/debugapi.
package rpc

import "github.com/eth2030/eth2030/rpc/debugapi"

type (
	// DbgTraceConfig is re-exported from rpc/debugapi.
	DbgTraceConfig = debugapi.DbgTraceConfig
	// DbgCallFrame is re-exported from rpc/debugapi.
	DbgCallFrame = debugapi.DbgCallFrame
	// DbgStorageRangeResult is re-exported from rpc/debugapi.
	DbgStorageRangeResult = debugapi.DbgStorageRangeResult
	// DbgStorageEntry is re-exported from rpc/debugapi.
	DbgStorageEntry = debugapi.DbgStorageEntry
	// DbgBadBlock is re-exported from rpc/debugapi.
	DbgBadBlock = debugapi.DbgBadBlock
	// DbgStateDump is re-exported from rpc/debugapi.
	DbgStateDump = debugapi.DbgStateDump
	// DbgDumpAccount is re-exported from rpc/debugapi.
	DbgDumpAccount = debugapi.DbgDumpAccount
	// DbgBlockTraceEntry is re-exported from rpc/debugapi.
	DbgBlockTraceEntry = debugapi.DbgBlockTraceEntry
	// DbgAPI is re-exported from rpc/debugapi.
	DbgAPI = debugapi.DbgAPI
)

var (
	// NewDbgAPI is re-exported from rpc/debugapi.
	NewDbgAPI = debugapi.NewDbgAPI
	// DefaultDbgTraceConfig is re-exported from rpc/debugapi.
	DefaultDbgTraceConfig = debugapi.DefaultDbgTraceConfig
)
