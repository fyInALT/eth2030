// sstore_dim_gas.go routes SSTORE zero→nonzero state-creation cost to the
// DimStorage gas dimension (GAP-2). The base SSTORE cost (WarmStorageReadGlamst)
// still goes to DimCompute; only the state-creation premium goes to DimStorage.
//
// TxDimGasUsage tracks per-dimension gas consumption for one transaction. It
// is used by block builders to enforce per-dimension block caps.
package vm

// StateCreationGasPremium is the per-SSTORE zero→nonzero state-creation cost
// routed to the DimStorage dimension. This equals GasSstoreSetGlamsterdam
// (the reservoir cost). Block builders enforce a 4M cap on this dimension.
const StateCreationGasPremium = GasSstoreSetGlamsterdam // 24084

// DimStorageBlockGasCap is the maximum DimStorage gas allowed per block.
// This is separate from the 30M DimCompute block limit (GAP-2.2).
const DimStorageBlockGasCap uint64 = 4_000_000

// TxDimGasUsage tracks per-dimension gas usage for one EVM transaction.
// DimCompute receives the base SSTORE cost and all non-creation opcodes.
// DimStorage receives the state-creation premium for SSTORE zero→nonzero.
type TxDimGasUsage struct {
	DimCompute   uint64 // execution/compute gas (base SSTORE + all other ops)
	DimStorage   uint64 // state-creation gas (SSTORE zero→nonzero premium)
	DimBandwidth uint64 // calldata bandwidth
	DimBlob      uint64 // blob gas
	DimWitness   uint64 // stateless witness gas
}

// AccountSSTOREGas updates dim gas counters for one SSTORE operation. For
// zero→nonzero (state creation), the warm read base cost goes to DimCompute
// and the creation premium goes to DimStorage. All other cases charge only
// DimCompute.
//
// Parameters:
//   - isStateCreation: true when SSTORE writes zero→nonzero
//   - coldPenalty: extra gas for cold access (charged to DimCompute)
//   - usage: per-dimension counters to update
func AccountSSTOREGas(isStateCreation bool, coldPenalty uint64, usage *TxDimGasUsage) {
	if usage == nil {
		return
	}
	if isStateCreation {
		// Base SSTORE cost (warm read) goes to DimCompute.
		usage.DimCompute += WarmStorageReadGlamst + coldPenalty
		// Creation premium goes to DimStorage.
		usage.DimStorage += StateCreationGasPremium
	} else {
		// All other SSTORE variants only affect DimCompute.
		usage.DimCompute += WarmStorageReadGlamst + coldPenalty
	}
}

// CheckDimStorageCap returns true if adding txStorageGas to the current block
// DimStorage usage does NOT exceed DimStorageBlockGasCap. Block builders
// should reject transactions for which this returns false (GAP-2.2).
func CheckDimStorageCap(blockStorageGasUsed, txStorageGas uint64) bool {
	return blockStorageGasUsed+txStorageGas <= DimStorageBlockGasCap
}
