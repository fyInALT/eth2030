package rpc

// eth_api_state.go re-exports StateAPI from rpc/ethapi.

import "github.com/eth2030/eth2030/rpc/ethapi"

// StateAPI is re-exported from rpc/ethapi.
type StateAPI = ethapi.StateAPI

// StateAccountProof is re-exported from rpc/ethapi.
type StateAccountProof = ethapi.StateAccountProof

// StateStorageProof is re-exported from rpc/ethapi.
type StateStorageProof = ethapi.StateStorageProof

// NewStateAPI is re-exported from rpc/ethapi.
var NewStateAPI = ethapi.NewStateAPI
