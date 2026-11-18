package execution

import (
	"math/big"

	"github.com/eth2030/eth2030/bal"
	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/eips"
	"github.com/eth2030/eth2030/core/gas"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
)

const systemExecutionGas = 30_000_000

// EIP-8037's fixed 16-slot margin uses the selected 1530 CPSB and 64-byte slot.
// The reservoir is separate from the execution budget and never debits a block.
const systemStateGas = 16 * 64 * 1530

func newSystemEVM(cfg *config.ChainConfig, getHash vm.GetHashFunc, db state.StateDB, h *types.Header) (*vm.EVM, error) {
	context := vm.BlockContext{GetHash: getHash, BlockNumber: h.Number, Time: h.Time, Coinbase: h.Coinbase, GasLimit: h.GasLimit, BaseFee: h.BaseFee, PrevRandao: h.MixDigest, BlobBaseFee: gas.BlobBaseFeeForHeader(cfg, h)}
	evm := vm.NewEVMWithState(context, vm.TxContext{Origin: types.SystemAddress, GasPrice: new(big.Int)}, vm.Config{}, db)
	if _, err := ConfigureBlockEVM(cfg, h, evm); err != nil {
		return nil, err
	}
	if cfg != nil && cfg.IsMerge() && cfg.IsAmsterdam(h.Time) {
		evm.SetInitialReservoir(systemStateGas)
	}
	return evm, nil
}
func executeSystemCall(evm *vm.EVM, db state.StateDB, address types.Address, input []byte, tracker *bal.AccessTracker) ([]byte, error) {
	db.ClearAccessList()
	db.ClearTransientStorage()
	db.SubRefund(db.GetRefund())
	if original, ok := db.(interface{ FinalizePreState() }); ok {
		original.FinalizePreState()
	}
	db.SetTxContext(types.Hash{}, 0)
	db.AddAddressToAccessList(address)
	if tracker != nil {
		tracker.RecordAddressTouch(address)
		evm.SetBALTracker(tracker, 0)
	}
	output, _, err := evm.Call(types.SystemAddress, address, input, systemExecutionGas, new(big.Int))
	evm.FinalizeSelfDestructs()
	if transaction, ok := db.(interface{ FinalizeTransaction() }); ok {
		transaction.FinalizeTransaction()
	}
	if tracker != nil {
		tracker.Reconcile(db)
	}
	db.ClearAccessList()
	db.ClearTransientStorage()
	db.SubRefund(db.GetRefund())
	return output, err
}

// ApplyPreExecutionSystems is shared by construction and import. It executes
// actual code, so absent or custom allocations cannot be silently replaced by
// the mainnet direct-storage optimization. Neither operation runs at genesis.
func ApplyPreExecutionSystems(cfg *config.ChainConfig, getHash vm.GetHashFunc, db state.StateDB, h *types.Header, tracker *bal.AccessTracker) error {
	if cfg == nil || h.Number == nil || h.Number.Sign() == 0 {
		return nil
	}
	type operation struct {
		address types.Address
		input   []byte
	}
	var operations []operation
	if cfg.IsCancun(h.Time) && h.ParentBeaconRoot != nil {
		operations = append(operations, operation{eips.BeaconRootAddress, h.ParentBeaconRoot.Bytes()})
	}
	if cfg.IsPrague(h.Time) {
		operations = append(operations, operation{eips.HistoryStorageAddress, h.ParentHash.Bytes()})
	}
	for _, operation := range operations {
		evm, err := newSystemEVM(cfg, getHash, db, h)
		if err != nil {
			return err
		}
		_, _ = executeSystemCall(evm, db, operation.address, operation.input, tracker)
	}
	return nil
}
