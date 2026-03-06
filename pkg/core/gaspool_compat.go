package core

// gaspool_compat.go re-exports types from core/gaspool for backward compatibility.

import "github.com/eth2030/eth2030/core/gaspool"

// GasPool type alias.
type GasPool = gaspool.GasPool

// GasPool error variable.
var ErrGasPoolExhausted = gaspool.ErrGasPoolExhausted
