package execution

import (
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
)

// TxExecutor processes blocks of transactions against a state database.
type TxExecutor interface {
	Process(block *types.Block, statedb state.StateDB) ([]*types.Receipt, error)
	ProcessWithBAL(block *types.Block, statedb state.StateDB) (*ProcessResult, error)
	SetGetHash(fn vm.GetHashFunc)
}

// Ensure StateProcessor satisfies TxExecutor at compile time.
var _ TxExecutor = (*StateProcessor)(nil)
