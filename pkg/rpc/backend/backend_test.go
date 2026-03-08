package rpcbackend

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
	"github.com/eth2030/eth2030/trie"
)

// localMock is a minimal Backend implementation for compile-time and
// interface-conformance tests within this package.  It cannot import
// rpc/internal/testutil (which imports rpcbackend) due to import cycles.
type localMock struct {
	chainID       *big.Int
	headers       map[uint64]*types.Header
	blocks        map[uint64]*types.Block
	statedb       *state.MemoryStateDB
	transactions  map[types.Hash]*txInfo
	receipts      map[types.Hash][]*types.Receipt
	logs          map[types.Hash][]*types.Log
	sentTxs       []*types.Transaction
	callResult    []byte
	callGasUsed   uint64
	callErr       error
	historyOldest uint64
}

type txInfo struct {
	tx       *types.Transaction
	blockNum uint64
	index    uint64
}

var _ Backend = (*localMock)(nil) // compile-time interface check

func newLocalMock() *localMock {
	sdb := state.NewMemoryStateDB()
	addr := types.HexToAddress("0xaaaa")
	sdb.AddBalance(addr, big.NewInt(1e18))
	sdb.SetNonce(addr, 5)
	sdb.SetCode(addr, []byte{0x60, 0x00})

	hdr := &types.Header{
		Number:   big.NewInt(42),
		GasLimit: 30_000_000,
		GasUsed:  15_000_000,
		Time:     1_700_000_000,
		BaseFee:  big.NewInt(1_000_000_000),
	}
	return &localMock{
		chainID:      big.NewInt(1337),
		headers:      map[uint64]*types.Header{42: hdr},
		blocks:       make(map[uint64]*types.Block),
		statedb:      sdb,
		transactions: make(map[types.Hash]*txInfo),
		receipts:     make(map[types.Hash][]*types.Receipt),
		logs:         make(map[types.Hash][]*types.Log),
	}
}

func (m *localMock) HeaderByNumber(n rpctypes.BlockNumber) *types.Header {
	if n == rpctypes.LatestBlockNumber || n == rpctypes.SafeBlockNumber || n == rpctypes.FinalizedBlockNumber {
		return m.headers[42]
	}
	return m.headers[uint64(n)]
}
func (m *localMock) HeaderByHash(h types.Hash) *types.Header {
	for _, hdr := range m.headers {
		if hdr.Hash() == h {
			return hdr
		}
	}
	return nil
}
func (m *localMock) BlockByNumber(n rpctypes.BlockNumber) *types.Block {
	if n == rpctypes.LatestBlockNumber || n == rpctypes.SafeBlockNumber || n == rpctypes.FinalizedBlockNumber {
		return m.blocks[42]
	}
	return m.blocks[uint64(n)]
}
func (m *localMock) BlockByHash(h types.Hash) *types.Block {
	for _, b := range m.blocks {
		if b != nil && b.Hash() == h {
			return b
		}
	}
	return nil
}
func (m *localMock) CurrentHeader() *types.Header { return m.headers[42] }
func (m *localMock) ChainID() *big.Int            { return m.chainID }
func (m *localMock) StateAt(_ types.Hash) (state.StateDB, error) {
	return m.statedb, nil
}
func (m *localMock) SendTransaction(tx *types.Transaction) error {
	m.sentTxs = append(m.sentTxs, tx)
	return nil
}
func (m *localMock) GetTransaction(hash types.Hash) (*types.Transaction, uint64, uint64) {
	if info, ok := m.transactions[hash]; ok {
		return info.tx, info.blockNum, info.index
	}
	return nil, 0, 0
}
func (m *localMock) SuggestGasPrice() *big.Int { return big.NewInt(1_000_000_000) }
func (m *localMock) GetReceipts(blockHash types.Hash) []*types.Receipt {
	return m.receipts[blockHash]
}
func (m *localMock) GetLogs(blockHash types.Hash) []*types.Log { return m.logs[blockHash] }
func (m *localMock) GetBlockReceipts(number uint64) []*types.Receipt {
	hdr := m.headers[number]
	if hdr == nil {
		return nil
	}
	return m.receipts[hdr.Hash()]
}
func (m *localMock) EVMCall(_ types.Address, _ *types.Address, _ []byte, _ uint64, _ *big.Int, _ rpctypes.BlockNumber) ([]byte, uint64, error) {
	return m.callResult, m.callGasUsed, m.callErr
}
func (m *localMock) GetProof(addr types.Address, keys []types.Hash, _ rpctypes.BlockNumber) (*trie.AccountProof, error) {
	st := m.statedb.BuildStateTrie()
	ss := m.statedb.BuildStorageTrie(addr)
	return trie.ProveAccountWithStorage(st, addr, ss, keys)
}
func (m *localMock) TraceTransaction(_ types.Hash) (*vm.StructLogTracer, error) {
	return vm.NewStructLogTracer(), nil
}
func (m *localMock) HistoryOldestBlock() uint64 { return m.historyOldest }

// --- tests ---

func TestBackendInterface_ChainID(t *testing.T) {
	mb := newLocalMock()
	var b Backend = mb
	id := b.ChainID()
	if id == nil {
		t.Fatal("ChainID returned nil")
	}
	if id.Cmp(big.NewInt(1337)) != 0 {
		t.Fatalf("want 1337, got %s", id.String())
	}
}

func TestBackendInterface_CurrentHeader(t *testing.T) {
	mb := newLocalMock()
	var b Backend = mb
	header := b.CurrentHeader()
	if header == nil {
		t.Fatal("CurrentHeader returned nil")
	}
	if header.Number.Uint64() != 42 {
		t.Fatalf("want block 42, got %d", header.Number.Uint64())
	}
}

func TestBackendInterface_HeaderByNumber(t *testing.T) {
	mb := newLocalMock()
	var b Backend = mb

	h := b.HeaderByNumber(42)
	if h == nil {
		t.Fatal("HeaderByNumber(42) returned nil")
	}

	h = b.HeaderByNumber(rpctypes.LatestBlockNumber)
	if h == nil || h.Number.Uint64() != 42 {
		t.Fatalf("latest: want 42, got %v", h)
	}

	h = b.HeaderByNumber(999)
	if h != nil {
		t.Fatal("HeaderByNumber(999) should return nil")
	}
}

func TestBackendInterface_HeaderByHash(t *testing.T) {
	mb := newLocalMock()
	var b Backend = mb

	header := b.HeaderByNumber(42)
	found := b.HeaderByHash(header.Hash())
	if found == nil || found.Number.Uint64() != 42 {
		t.Fatal("HeaderByHash returned wrong header")
	}

	if b.HeaderByHash(types.Hash{}) != nil {
		t.Fatal("HeaderByHash should return nil for unknown hash")
	}
}

func TestBackendInterface_StateAt(t *testing.T) {
	mb := newLocalMock()
	var b Backend = mb

	st, err := b.StateAt(types.Hash{})
	if err != nil {
		t.Fatalf("StateAt error: %v", err)
	}
	addr := types.HexToAddress("0xaaaa")
	balance := st.GetBalance(addr)
	if balance == nil || balance.Cmp(big.NewInt(1e18)) != 0 {
		t.Fatalf("want balance 1e18, got %v", balance)
	}
}

func TestBackendInterface_SendTransaction(t *testing.T) {
	mb := newLocalMock()
	var b Backend = mb

	tx := types.NewTransaction(&types.LegacyTx{Nonce: 1, Gas: 21000})
	if err := b.SendTransaction(tx); err != nil {
		t.Fatalf("SendTransaction error: %v", err)
	}
	if len(mb.sentTxs) != 1 {
		t.Fatalf("want 1 sent tx, got %d", len(mb.sentTxs))
	}
}

func TestBackendInterface_GetTransaction(t *testing.T) {
	mb := newLocalMock()
	var b Backend = mb

	tx := types.NewTransaction(&types.LegacyTx{Nonce: 10})
	hash := tx.Hash()
	mb.transactions[hash] = &txInfo{tx: tx, blockNum: 42, index: 0}

	found, blockNum, index := b.GetTransaction(hash)
	if found == nil || blockNum != 42 || index != 0 {
		t.Fatalf("GetTransaction: got (%v, %d, %d)", found, blockNum, index)
	}
	if tx, _, _ := b.GetTransaction(types.Hash{}); tx != nil {
		t.Fatal("GetTransaction should return nil for unknown hash")
	}
}

func TestBackendInterface_SuggestGasPrice(t *testing.T) {
	mb := newLocalMock()
	price := (Backend(mb)).SuggestGasPrice()
	if price == nil || price.Cmp(big.NewInt(1_000_000_000)) != 0 {
		t.Fatalf("want 1 Gwei, got %v", price)
	}
}

func TestBackendInterface_GetReceipts(t *testing.T) {
	mb := newLocalMock()
	var b Backend = mb

	blockHash := mb.headers[42].Hash()
	receipt := &types.Receipt{Status: types.ReceiptStatusSuccessful, GasUsed: 21000, Logs: []*types.Log{}}
	mb.receipts[blockHash] = []*types.Receipt{receipt}

	receipts := b.GetReceipts(blockHash)
	if len(receipts) != 1 || receipts[0].GasUsed != 21000 {
		t.Fatalf("unexpected receipts: %v", receipts)
	}
}

func TestBackendInterface_GetLogs(t *testing.T) {
	mb := newLocalMock()
	var b Backend = mb

	blockHash := mb.headers[42].Hash()
	mb.logs[blockHash] = []*types.Log{{Address: types.HexToAddress("0xcccc"), BlockNumber: 42}}

	logs := b.GetLogs(blockHash)
	if len(logs) != 1 {
		t.Fatalf("want 1 log, got %d", len(logs))
	}
}

func TestBackendInterface_EVMCall(t *testing.T) {
	mb := newLocalMock()
	mb.callResult = []byte{0x01, 0x02}
	mb.callGasUsed = 5000
	var b Backend = mb

	result, gasUsed, err := b.EVMCall(types.Address{}, nil, nil, 100000, nil, rpctypes.LatestBlockNumber)
	if err != nil || len(result) != 2 || result[0] != 0x01 || gasUsed != 5000 {
		t.Fatalf("EVMCall unexpected: result=%x gasUsed=%d err=%v", result, gasUsed, err)
	}
}

func TestBackendInterface_HistoryOldestBlock(t *testing.T) {
	mb := newLocalMock()
	var b Backend = mb

	if b.HistoryOldestBlock() != 0 {
		t.Fatal("want 0 initially")
	}
	mb.historyOldest = 100
	if b.HistoryOldestBlock() != 100 {
		t.Fatal("want 100 after set")
	}
}
