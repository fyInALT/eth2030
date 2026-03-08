// debug_api_ext.go re-exports extended debug namespace types from rpc/debugapi.
package rpc

import "github.com/eth2030/eth2030/rpc/debugapi"

type (
	// DebugExtAPI is re-exported from rpc/debugapi.
	DebugExtAPI = debugapi.DebugExtAPI
	// StorageRangeResult is re-exported from rpc/debugapi.
	StorageRangeResult = debugapi.StorageRangeResult
	// StorageEntry is re-exported from rpc/debugapi.
	StorageEntry = debugapi.StorageEntry
	// AccountRangeResult is re-exported from rpc/debugapi.
	AccountRangeResult = debugapi.AccountRangeResult
	// AccountEntry is re-exported from rpc/debugapi.
	AccountEntry = debugapi.AccountEntry
	// DumpBlockResult is re-exported from rpc/debugapi.
	DumpBlockResult = debugapi.DumpBlockResult
	// DumpAccount is re-exported from rpc/debugapi.
	DumpAccount = debugapi.DumpAccount
	// ModifiedAccountsResult is re-exported from rpc/debugapi.
	ModifiedAccountsResult = debugapi.ModifiedAccountsResult
)

// NewDebugExtAPI is re-exported from rpc/debugapi.
var NewDebugExtAPI = debugapi.NewDebugExtAPI
