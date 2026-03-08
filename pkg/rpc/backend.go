package rpc

// backend.go re-exports types from rpc/backend for backward compatibility.

import (
	rpcbackend "github.com/eth2030/eth2030/rpc/backend"
)

// Backend re-exports the Backend interface.
type Backend = rpcbackend.Backend
