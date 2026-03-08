package rpc

// api_proof.go re-exports proof and trace types from rpc/ethapi and rpc/debugapi.

import (
	"github.com/eth2030/eth2030/rpc/debugapi"
	"github.com/eth2030/eth2030/rpc/ethapi"
)

// Re-export trace types from rpc/debugapi.
type (
	// StructLog is re-exported from rpc/debugapi.
	StructLog = debugapi.StructLog
	// TraceResult is re-exported from rpc/debugapi.
	TraceResult = debugapi.TraceResult
)

// AccountProof is re-exported from rpc/ethapi.
type AccountProof = ethapi.AccountProof

// StorageProof is re-exported from rpc/ethapi.
type StorageProof = ethapi.StorageProof
