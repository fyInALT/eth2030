// Package rpc provides JSON-RPC 2.0 types and the standard Ethereum
// JSON-RPC API (eth_ namespace) for the ETH2030 execution client.
package rpc

// types.go re-exports types from rpc/types for backward compatibility.

import (
	"github.com/eth2030/eth2030/rpc/adminapi"
	rpcbackend "github.com/eth2030/eth2030/rpc/backend"
	rpcbatch "github.com/eth2030/eth2030/rpc/batch"
	"github.com/eth2030/eth2030/rpc/debugapi"
	"github.com/eth2030/eth2030/rpc/ethapi"
	rpcserver "github.com/eth2030/eth2030/rpc/server"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// Re-export sub-package types as aliases so the top-level rpc package
// compiles without changes.
type (
	// Core backend and API types.
	Backend          = rpcbackend.Backend
	EthAPI           = ethapi.EthAPI
	BatchHandler     = rpcbatch.BatchHandler
	BatchResponse    = rpcbatch.BatchResponse
	AdminBackend     = adminapi.Backend
	AdminDispatchAPI = adminapi.DispatchAPI
	NodeInfoData     = adminapi.NodeInfoData
	NodePorts        = adminapi.NodePorts
	PeerInfoData     = adminapi.PeerInfoData
	ServerConfig     = rpcserver.ServerConfig
)

// NewBatchHandler wraps rpcbatch.NewBatchHandler for use within this package.
var NewBatchHandler = rpcbatch.NewBatchHandler

// NewAdminDispatchAPI wraps adminapi.NewDispatchAPI.
var NewAdminDispatchAPI = adminapi.NewDispatchAPI

// IsBatchRequest reports whether body is a JSON array (batch request).
var IsBatchRequest = rpcbatch.IsBatchRequest

// MarshalBatchResponse encodes a slice of BatchResponse as JSON.
var MarshalBatchResponse = rpcbatch.MarshalBatchResponse

// Re-export all RPC types as type aliases so existing code in this package
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
	FeeHistoryResult = ethapi.FeeHistoryResult
	AccessListResult = ethapi.AccessListResult
	AccountProof     = ethapi.AccountProof
	TraceResult      = debugapi.TraceResult
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
