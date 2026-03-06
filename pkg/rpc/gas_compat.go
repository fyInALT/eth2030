package rpc

// gas_compat.go re-exports types from rpc/gas for backward compatibility.

import "github.com/eth2030/eth2030/rpc/gas"

// Gas type aliases.
type (
	GasOracleConfig  = gas.GasOracleConfig
	BlockFeeData     = gas.BlockFeeData
	GasOracle        = gas.GasOracle
	GasTrackerConfig = gas.GasTrackerConfig
	GasBlockRecord   = gas.GasBlockRecord
	GasFeeEstimate   = gas.GasFeeEstimate
	GasTracker       = gas.GasTracker
)

// Gas error variables.
var (
	ErrGasTrackerEmpty             = gas.ErrGasTrackerEmpty
	ErrGasTrackerInvalidPercentile = gas.ErrGasTrackerInvalidPercentile
)

// Gas function wrappers.
func DefaultGasOracleConfig() GasOracleConfig           { return gas.DefaultGasOracleConfig() }
func NewGasOracle(config GasOracleConfig) *GasOracle    { return gas.NewGasOracle(config) }
func NewGasTracker(config GasTrackerConfig) *GasTracker { return gas.NewGasTracker(config) }
