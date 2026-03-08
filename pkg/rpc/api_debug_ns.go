package rpc

import "github.com/eth2030/eth2030/rpc/debugapi"

type (
	// DebugAPI is re-exported from rpc/debugapi.
	DebugAPI = debugapi.DebugAPI
	// DebugBlockTraceEntry is re-exported from rpc/debugapi.
	DebugBlockTraceEntry = debugapi.DebugBlockTraceEntry
	// DebugMemStats is re-exported from rpc/debugapi.
	DebugMemStats = debugapi.DebugMemStats
)

// NewDebugAPI is re-exported from rpc/debugapi.
var NewDebugAPI = debugapi.NewDebugAPI
