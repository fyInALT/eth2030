// Package debugapi implements the Ethereum debug namespace JSON-RPC methods,
// providing block tracing, storage inspection, and chain management utilities.
package debugapi

import (
	"encoding/json"
	"time"

	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// RPCBlock is a re-export of rpctypes.RPCBlock for use within this package.
type RPCBlock = rpctypes.RPCBlock

// StructLog is a single step in an EVM execution trace.
type StructLog struct {
	PC      uint64            `json:"pc"`
	Op      string            `json:"op"`
	Gas     uint64            `json:"gas"`
	GasCost uint64            `json:"gasCost"`
	Depth   int               `json:"depth"`
	Stack   []string          `json:"stack"`
	Memory  []string          `json:"memory,omitempty"`
	Storage map[string]string `json:"storage,omitempty"`
}

// TraceResult is the response for debug_traceTransaction.
type TraceResult struct {
	Gas         uint64      `json:"gas"`
	Failed      bool        `json:"failed"`
	ReturnValue string      `json:"returnValue"`
	StructLogs  []StructLog `json:"structLogs"`
}

// BlockTraceResult is a single transaction trace within a block trace.
type BlockTraceResult struct {
	TxHash string       `json:"txHash"`
	Result *TraceResult `json:"result"`
}

// DebugBlockTraceEntry is a single transaction trace in a block trace response.
type DebugBlockTraceEntry struct {
	TxHash string       `json:"txHash"`
	Result *TraceResult `json:"result"`
}

// DebugMemStats returns runtime memory statistics for diagnostics.
type DebugMemStats struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"totalAlloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"numGC"`
}

// DbgTraceConfig configures how a debug trace should be executed.
// Mirrors the tracerConfig parameter from debug_traceBlockByNumber et al.
type DbgTraceConfig struct {
	// Tracer selects the tracer type: "callTracer", "prestateTracer",
	// or empty string for the default struct logger.
	Tracer string `json:"tracer,omitempty"`

	// Timeout is the maximum duration for the trace. "5s", "1m", etc.
	Timeout string `json:"timeout,omitempty"`

	// Reexec specifies the number of blocks to re-execute when the
	// requested state is not directly available. Default 128.
	Reexec *uint64 `json:"reexec,omitempty"`

	// TracerConfig is an opaque configuration object forwarded to the
	// selected tracer (e.g. {"onlyTopCall": true} for callTracer).
	TracerConfig json.RawMessage `json:"tracerConfig,omitempty"`

	// DisableStorage disables storage capture in the struct logger.
	DisableStorage bool `json:"disableStorage,omitempty"`

	// DisableStack disables stack capture in the struct logger.
	DisableStack bool `json:"disableStack,omitempty"`

	// DisableMemory disables memory capture in the struct logger.
	DisableMemory bool `json:"disableMemory,omitempty"`

	// EnableReturnData enables return data capture in the struct logger.
	EnableReturnData bool `json:"enableReturnData,omitempty"`
}

// DefaultDbgTraceConfig returns a DbgTraceConfig with reasonable defaults.
func DefaultDbgTraceConfig() DbgTraceConfig {
	reexec := uint64(128)
	return DbgTraceConfig{
		Reexec:  &reexec,
		Timeout: "5s",
	}
}

// parseDbgTimeout parses the timeout string from the trace config.
// Returns 5 seconds as default when the string is empty or invalid.
func parseDbgTimeout(s string) time.Duration {
	if s == "" {
		return 5 * time.Second
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 5 * time.Second
	}
	if d <= 0 {
		return 5 * time.Second
	}
	return d
}

// DbgCallFrame represents a single call in a callTracer output.
type DbgCallFrame struct {
	Type    string          `json:"type"`
	From    string          `json:"from"`
	To      string          `json:"to,omitempty"`
	Value   string          `json:"value,omitempty"`
	Gas     string          `json:"gas"`
	GasUsed string          `json:"gasUsed"`
	Input   string          `json:"input"`
	Output  string          `json:"output,omitempty"`
	Error   string          `json:"error,omitempty"`
	Calls   []*DbgCallFrame `json:"calls,omitempty"`
}

// DbgStorageRangeResult is the response for debug_storageRangeAt.
type DbgStorageRangeResult struct {
	Storage map[string]DbgStorageEntry `json:"storage"`
	NextKey *string                    `json:"nextKey"`
}

// DbgStorageEntry is a single entry in the storage range output.
type DbgStorageEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DbgBadBlock represents a rejected block returned by debug_getBadBlocks.
type DbgBadBlock struct {
	Hash   string    `json:"hash"`
	Block  *RPCBlock `json:"block"`
	Reason string    `json:"reason,omitempty"`
}

// DbgStateDump is the response for debug_dumpBlock.
type DbgStateDump struct {
	Root     string                     `json:"root"`
	Accounts map[string]*DbgDumpAccount `json:"accounts"`
}

// DbgDumpAccount is a single account in the state dump.
type DbgDumpAccount struct {
	Balance  string            `json:"balance"`
	Nonce    uint64            `json:"nonce"`
	CodeHash string            `json:"codeHash"`
	Code     string            `json:"code,omitempty"`
	Storage  map[string]string `json:"storage,omitempty"`
}

// DbgBlockTraceEntry wraps a per-transaction trace result in a block trace.
type DbgBlockTraceEntry struct {
	TxHash string       `json:"txHash"`
	Result *TraceResult `json:"result"`
}

// StorageRangeResult is the response for debug_storageRangeAt (ext API).
type StorageRangeResult struct {
	Storage map[string]StorageEntry `json:"storage"`
	NextKey *string                 `json:"nextKey"`
}

// StorageEntry represents a single storage slot in the debug response.
type StorageEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AccountRangeResult is the response for debug_accountRange.
type AccountRangeResult struct {
	Accounts map[string]AccountEntry `json:"accounts"`
	NextKey  string                  `json:"next"`
}

// AccountEntry represents a single account in the debug_accountRange response.
type AccountEntry struct {
	Balance string `json:"balance"`
	Nonce   uint64 `json:"nonce"`
	Code    string `json:"code"`
	Root    string `json:"root"`
	HasCode bool   `json:"hasCode"`
}

// DumpBlockResult contains the dumped state of a block.
type DumpBlockResult struct {
	Root     string                 `json:"root"`
	Accounts map[string]DumpAccount `json:"accounts"`
}

// DumpAccount is a single account in a block dump.
type DumpAccount struct {
	Balance  string            `json:"balance"`
	Nonce    uint64            `json:"nonce"`
	Root     string            `json:"root"`
	CodeHash string            `json:"codeHash"`
	Code     string            `json:"code"`
	Storage  map[string]string `json:"storage"`
}

// ModifiedAccountsResult lists accounts modified between two blocks.
type ModifiedAccountsResult struct {
	Accounts []string `json:"accounts"`
}
