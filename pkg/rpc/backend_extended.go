package rpc

// backend_extended.go re-exports extended backend types from rpc/backend
// for backward compatibility.

import (
	rpcbackend "github.com/eth2030/eth2030/rpc/backend"
)

// Re-export extended backend errors.
var (
	ErrBackendBlockNotFound   = rpcbackend.ErrBackendBlockNotFound
	ErrBackendStateUnavail    = rpcbackend.ErrBackendStateUnavail
	ErrBackendTxNotFound      = rpcbackend.ErrBackendTxNotFound
	ErrBackendGasCapExceeded  = rpcbackend.ErrBackendGasCapExceeded
	ErrBackendHistoryPruned   = rpcbackend.ErrBackendHistoryPruned
	ErrBackendNoEstimate      = rpcbackend.ErrBackendNoEstimate
	ErrBackendReceiptNotFound = rpcbackend.ErrBackendReceiptNotFound
)

// Re-export service types.
type (
	GasEstimationConfig = rpcbackend.GasEstimationConfig
	FeeHistoryEntry     = rpcbackend.FeeHistoryEntry
	AccountInfo         = rpcbackend.AccountInfo
	ChainStateAccessor  = rpcbackend.ChainStateAccessor
	GasEstimator        = rpcbackend.GasEstimator
	FeeHistoryCollector = rpcbackend.FeeHistoryCollector
	ChainIDAccessor     = rpcbackend.ChainIDAccessor
	ReceiptAccessor     = rpcbackend.ReceiptAccessor
	BackendServices     = rpcbackend.BackendServices
)

// Re-export constructor functions.
var (
	DefaultGasEstimationConfig = rpcbackend.DefaultGasEstimationConfig
	NewChainStateAccessor      = rpcbackend.NewChainStateAccessor
	NewGasEstimator            = rpcbackend.NewGasEstimator
	NewFeeHistoryCollector     = rpcbackend.NewFeeHistoryCollector
	NewChainIDAccessor         = rpcbackend.NewChainIDAccessor
	NewReceiptAccessor         = rpcbackend.NewReceiptAccessor
	NewBackendServices         = rpcbackend.NewBackendServices
)
