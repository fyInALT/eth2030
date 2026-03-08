package rpc

// api.go provides backward-compatible re-exports from rpc/ethapi.

import (
	"github.com/eth2030/eth2030/rpc/ethapi"
)

// NewEthAPI creates a new EthAPI with an embedded SubscriptionManager.
// This is the primary constructor used by the top-level rpc package.
func NewEthAPI(backend Backend) *EthAPI {
	subs := NewSubscriptionManager(backend)
	return ethapi.NewEthAPI(backend, subs)
}
