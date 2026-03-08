package execution

import (
	"testing"

	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
)

// mockTxExecutor implements TxExecutor for testing.
type mockTxExecutor struct {
	processFn    func(block *types.Block, statedb state.StateDB) ([]*types.Receipt, error)
	processBALFn func(block *types.Block, statedb state.StateDB) (*ProcessResult, error)
	getHashFn    vm.GetHashFunc
}

func (m *mockTxExecutor) Process(block *types.Block, statedb state.StateDB) ([]*types.Receipt, error) {
	if m.processFn != nil {
		return m.processFn(block, statedb)
	}
	return nil, nil
}

func (m *mockTxExecutor) ProcessWithBAL(block *types.Block, statedb state.StateDB) (*ProcessResult, error) {
	if m.processBALFn != nil {
		return m.processBALFn(block, statedb)
	}
	return &ProcessResult{}, nil
}

func (m *mockTxExecutor) SetGetHash(fn vm.GetHashFunc) {
	m.getHashFn = fn
}

// Compile-time interface check.
var _ TxExecutor = (*mockTxExecutor)(nil)

// --- Interface tests ---

func TestTxExecutor_Interface(t *testing.T) {
	called := false
	var exec TxExecutor = &mockTxExecutor{
		processFn: func(block *types.Block, statedb state.StateDB) ([]*types.Receipt, error) {
			called = true
			return []*types.Receipt{}, nil
		},
	}
	receipts, err := exec.Process(&types.Block{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("processFn not called")
	}
	if receipts == nil {
		t.Fatal("expected empty (non-nil) receipts slice")
	}
}

func TestTxExecutor_ProcessReturnsNilWhenFnNotSet(t *testing.T) {
	var exec TxExecutor = &mockTxExecutor{}
	receipts, err := exec.Process(&types.Block{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receipts != nil {
		t.Fatalf("expected nil receipts, got %v", receipts)
	}
}

func TestTxExecutor_ProcessWithBAL_Interface(t *testing.T) {
	wantResult := &ProcessResult{Receipts: []*types.Receipt{{}}}
	var exec TxExecutor = &mockTxExecutor{
		processBALFn: func(block *types.Block, statedb state.StateDB) (*ProcessResult, error) {
			return wantResult, nil
		},
	}
	result, err := exec.ProcessWithBAL(&types.Block{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != wantResult {
		t.Fatal("unexpected ProcessResult returned")
	}
}

func TestTxExecutor_SetGetHash_Interface(t *testing.T) {
	mock := &mockTxExecutor{}
	var exec TxExecutor = mock
	fn := func(n uint64) types.Hash { return types.Hash{byte(n)} }
	exec.SetGetHash(fn)
	// Verify the function field was stored on the concrete mock.
	if mock.getHashFn == nil {
		t.Fatal("expected getHashFn to be stored")
	}
	got := mock.getHashFn(5)
	if got != (types.Hash{5}) {
		t.Fatalf("unexpected hash: %v", got)
	}
}
