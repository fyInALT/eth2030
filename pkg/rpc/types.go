// Package rpc provides JSON-RPC 2.0 types and the standard Ethereum
// JSON-RPC API (eth_ namespace) for the ETH2030 execution client.
package rpc

// types.go re-exports types from rpc/types for backward compatibility.

import (
	"math/big"

	coretypes "github.com/eth2030/eth2030/core/types"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// Re-export all types as type aliases so existing code in this package
// and external callers continue to compile without changes.
type (
	BlockNumber      = rpctypes.BlockNumber
	Request          = rpctypes.Request
	Response         = rpctypes.Response
	RPCError         = rpctypes.RPCError
	RPCBlock         = rpctypes.RPCBlock
	RPCBlockWithTxs  = rpctypes.RPCBlockWithTxs
	RPCAccessTuple   = rpctypes.RPCAccessTuple
	RPCAuthorization = rpctypes.RPCAuthorization
	RPCTransaction   = rpctypes.RPCTransaction
	RPCReceipt       = rpctypes.RPCReceipt
	RPCLog           = rpctypes.RPCLog
	RPCWithdrawal    = rpctypes.RPCWithdrawal
	CallArgs         = rpctypes.CallArgs
	FilterCriteria   = rpctypes.FilterCriteria
)

// Re-export constants.
const (
	LatestBlockNumber    = rpctypes.LatestBlockNumber
	PendingBlockNumber   = rpctypes.PendingBlockNumber
	EarliestBlockNumber  = rpctypes.EarliestBlockNumber
	SafeBlockNumber      = rpctypes.SafeBlockNumber
	FinalizedBlockNumber = rpctypes.FinalizedBlockNumber

	ErrCodeParse          = rpctypes.ErrCodeParse
	ErrCodeInvalidRequest = rpctypes.ErrCodeInvalidRequest
	ErrCodeMethodNotFound = rpctypes.ErrCodeMethodNotFound
	ErrCodeInvalidParams  = rpctypes.ErrCodeInvalidParams
	ErrCodeInternal       = rpctypes.ErrCodeInternal
	ErrCodeHistoryPruned  = rpctypes.ErrCodeHistoryPruned
)

// Re-export formatting functions.
var (
	FormatBlock       = rpctypes.FormatBlock
	FormatHeader      = rpctypes.FormatHeader
	FormatTransaction = rpctypes.FormatTransaction
	FormatReceipt     = rpctypes.FormatReceipt
	FormatLog         = rpctypes.FormatLog
)

// Unexported wrappers for internal helpers used throughout the rpc package.
// These delegate to the exported versions in rpctypes.

func encodeHash(h coretypes.Hash) string       { return rpctypes.EncodeHash(h) }
func encodeAddress(a coretypes.Address) string { return rpctypes.EncodeAddress(a) }
func encodeBytes(b []byte) string              { return rpctypes.EncodeBytes(b) }
func encodeBloom(b coretypes.Bloom) string     { return rpctypes.EncodeBloom(b) }
func encodeUint64(n uint64) string             { return rpctypes.EncodeUint64(n) }
func encodeBigInt(n *big.Int) string           { return rpctypes.EncodeBigInt(n) }
func fromHexBytes(s string) []byte             { return rpctypes.FromHexBytes(s) }
func unhex(c byte) byte                        { return rpctypes.Unhex(c) }
func parseHexUint64(s string) uint64           { return rpctypes.ParseHexUint64(s) }
func parseHexBigInt(s string) *big.Int         { return rpctypes.ParseHexBigInt(s) }
func formatUncleHashes(u []*coretypes.Header) []string {
	return rpctypes.FormatUncleHashes(u)
}
func formatAccessList(al coretypes.AccessList) []RPCAccessTuple {
	return rpctypes.FormatAccessList(al)
}
func formatAuthorizationList(auths []coretypes.Authorization) []RPCAuthorization {
	return rpctypes.FormatAuthorizationList(auths)
}
