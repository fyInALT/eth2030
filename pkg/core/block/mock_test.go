package block

import (
	"errors"
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
)

// mockValidator implements Validator for testing.
type mockValidator struct {
	validateHeaderFn          func(header, parent *types.Header) error
	validateBodyFn            func(block *types.Block) error
	validateRequestsFn        func(header *types.Header, requests types.Requests) error
	validateBlockAccessListFn func(header *types.Header, computedBALHash *types.Hash) error
}

func (m *mockValidator) ValidateHeader(header, parent *types.Header) error {
	if m.validateHeaderFn != nil {
		return m.validateHeaderFn(header, parent)
	}
	return nil
}

func (m *mockValidator) ValidateBody(block *types.Block) error {
	if m.validateBodyFn != nil {
		return m.validateBodyFn(block)
	}
	return nil
}

func (m *mockValidator) ValidateRequests(header *types.Header, requests types.Requests) error {
	if m.validateRequestsFn != nil {
		return m.validateRequestsFn(header, requests)
	}
	return nil
}

func (m *mockValidator) ValidateBlockAccessList(header *types.Header, computedBALHash *types.Hash) error {
	if m.validateBlockAccessListFn != nil {
		return m.validateBlockAccessListFn(header, computedBALHash)
	}
	return nil
}

// mockChain implements BlockchainReader for testing.
type mockChain struct {
	config       *config.ChainConfig
	currentBlock *types.Block
	genesis      *types.Block
	blocks       map[types.Hash]*types.Block
	stateAt      func(*types.Block) (state.StateDB, error)
}

func (m *mockChain) Config() *config.ChainConfig { return m.config }

func (m *mockChain) CurrentBlock() *types.Block { return m.currentBlock }

func (m *mockChain) Genesis() *types.Block { return m.genesis }

func (m *mockChain) GetBlock(hash types.Hash) *types.Block {
	if m.blocks != nil {
		return m.blocks[hash]
	}
	return nil
}

func (m *mockChain) StateAtBlock(block *types.Block) (state.StateDB, error) {
	if m.stateAt != nil {
		return m.stateAt(block)
	}
	return state.NewMemoryStateDB(), nil
}

func (m *mockChain) GetHashFn() func(uint64) types.Hash {
	return func(uint64) types.Hash { return types.Hash{} }
}

// mockPool implements TxPoolReader for testing.
type mockPool struct{ txs []*types.Transaction }

func (p *mockPool) Pending() []*types.Transaction { return p.txs }

// Compile-time interface checks.
var _ Validator = (*mockValidator)(nil)
var _ BlockchainReader = (*mockChain)(nil)
var _ TxPoolReader = (*mockPool)(nil)

// --- Interface tests ---

func TestValidator_Interface(t *testing.T) {
	var v Validator = &mockValidator{
		validateHeaderFn: func(h, p *types.Header) error {
			return errors.New("bad header")
		},
	}
	err := v.ValidateHeader(&types.Header{}, &types.Header{})
	if err == nil || err.Error() != "bad header" {
		t.Fatal("expected 'bad header' error")
	}
	// Body validation returns nil when fn not set.
	if err := v.ValidateBody(&types.Block{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidator_ValidateBody_Interface(t *testing.T) {
	wantErr := errors.New("bad body")
	var v Validator = &mockValidator{
		validateBodyFn: func(block *types.Block) error { return wantErr },
	}
	if err := v.ValidateBody(&types.Block{}); err != wantErr {
		t.Fatalf("got %v, want %v", err, wantErr)
	}
}

func TestValidator_ValidateRequests_Interface(t *testing.T) {
	called := false
	var v Validator = &mockValidator{
		validateRequestsFn: func(h *types.Header, reqs types.Requests) error {
			called = true
			return nil
		},
	}
	if err := v.ValidateRequests(&types.Header{}, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("validateRequestsFn not called")
	}
}

func TestValidator_ValidateBlockAccessList_Interface(t *testing.T) {
	h := types.Hash{0x01}
	var got *types.Hash
	var v Validator = &mockValidator{
		validateBlockAccessListFn: func(header *types.Header, computedBALHash *types.Hash) error {
			got = computedBALHash
			return nil
		},
	}
	if err := v.ValidateBlockAccessList(&types.Header{}, &h); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || *got != h {
		t.Fatal("expected BAL hash to be forwarded")
	}
}

func TestBlockchainReader_Interface(t *testing.T) {
	genesis := makeGenesis(15_000_000, big.NewInt(1e9))
	var chain BlockchainReader = &mockChain{
		config:       config.TestConfig,
		currentBlock: genesis,
		genesis:      genesis,
		blocks:       map[types.Hash]*types.Block{genesis.Hash(): genesis},
	}
	if got := chain.Config(); got != config.TestConfig {
		t.Fatal("wrong config")
	}
	if got := chain.CurrentBlock(); got.Hash() != genesis.Hash() {
		t.Fatal("wrong current block")
	}
	if got := chain.Genesis(); got.Hash() != genesis.Hash() {
		t.Fatal("wrong genesis")
	}
	if got := chain.GetBlock(genesis.Hash()); got == nil || got.Hash() != genesis.Hash() {
		t.Fatal("GetBlock returned wrong block")
	}
	if got := chain.GetBlock(types.Hash{0xff}); got != nil {
		t.Fatal("expected nil for unknown hash")
	}
	statedb, err := chain.StateAtBlock(genesis)
	if err != nil {
		t.Fatalf("StateAtBlock: %v", err)
	}
	if statedb == nil {
		t.Fatal("expected non-nil statedb")
	}
}

func TestBlockchainReader_StateAtBlock_Error(t *testing.T) {
	wantErr := errors.New("state unavailable")
	var chain BlockchainReader = &mockChain{
		stateAt: func(*types.Block) (state.StateDB, error) { return nil, wantErr },
	}
	_, err := chain.StateAtBlock(&types.Block{})
	if err != wantErr {
		t.Fatalf("got %v, want %v", err, wantErr)
	}
}

func TestBlockBuilder_WithMockChain(t *testing.T) {
	genesis := makeGenesis(15_000_000, big.NewInt(1e9))
	chain := &mockChain{
		config:       config.TestConfig,
		currentBlock: genesis,
		genesis:      genesis,
		blocks:       map[types.Hash]*types.Block{genesis.Hash(): genesis},
	}
	pool := &mockPool{}
	// NewBlockBuilder accepts BlockchainReader — pass the interface, not a concrete chain type.
	var chainIface BlockchainReader = chain
	builder := NewBlockBuilder(config.TestConfig, chainIface, pool)
	if builder == nil {
		t.Fatal("expected non-nil builder")
	}
}
