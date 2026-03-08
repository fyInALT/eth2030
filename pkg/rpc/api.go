package rpc

// api.go provides backward-compatible re-exports from rpc/ethapi.

import (
	"github.com/eth2030/eth2030/rpc/ethapi"
)

// EthAPI is a type alias for ethapi.EthAPI.
type EthAPI = ethapi.EthAPI

// NewEthAPI creates a new EthAPI with an embedded SubscriptionManager.
// This is the primary constructor used by the top-level rpc package.
func NewEthAPI(backend Backend) *EthAPI {
	subs := NewSubscriptionManager(backend)
	return ethapi.NewEthAPI(backend, subs)
}

// FeeHistoryResult is re-exported from ethapi.
type FeeHistoryResult = ethapi.FeeHistoryResult

// SyncStatus is re-exported from ethapi.
type SyncStatus = ethapi.SyncStatus

// AccessListResult is re-exported from ethapi.
type AccessListResult = ethapi.AccessListResult

// AccessListEntry is re-exported from ethapi.
type AccessListEntry = ethapi.AccessListEntry
