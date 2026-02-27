package vm

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

// BALTracker records EVM state accesses for Block Access List (EIP-7928)
// construction. This interface is defined in the vm package to avoid
// circular imports with bal/. The bal.AccessTracker satisfies this
// interface via Go structural typing.
type BALTracker interface {
	RecordStorageRead(addr types.Address, slot types.Hash, value types.Hash)
	RecordStorageChange(addr types.Address, slot, oldVal, newVal types.Hash)
	RecordBalanceChange(addr types.Address, oldBal, newBal *big.Int)
	RecordAddressTouch(addr types.Address)
}
