package rpc

// api_eth_extended.go re-exports EthExtendedAPI from rpc/ethapi.

import "github.com/eth2030/eth2030/rpc/ethapi"

// EthExtendedAPI is re-exported from rpc/ethapi.
type EthExtendedAPI = ethapi.EthExtendedAPI

// NewEthExtendedAPI is re-exported from rpc/ethapi.
var NewEthExtendedAPI = ethapi.NewEthExtendedAPI
