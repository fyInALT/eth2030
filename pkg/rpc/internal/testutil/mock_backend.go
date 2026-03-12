// Package testutil provides shared test infrastructure for rpc sub-packages.
// It is only importable by packages under github.com/eth2030/eth2030/rpc/*.
package testutil

import (
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
	rpcbackend "github.com/eth2030/eth2030/rpc/backend"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
	"github.com/eth2030/eth2030/trie"
)

// Ensure MockBackend satisfies the Backend interface at compile time.
var _ rpcbackend.Backend = (*MockBackend)(nil)

// MockTxInfo holds per-transaction data for the mock backend.
type MockTxInfo struct {
	Tx       *types.Transaction
	BlockNum uint64
	Index    uint64
}

// MockBackend implements rpcbackend.Backend for use in sub-package tests.
// All fields are exported so individual tests can customise behaviour.
type MockBackend struct {
	ChainIDVal    *big.Int
	Headers       map[uint64]*types.Header
	Blocks        map[uint64]*types.Block
	Statedb       *state.MemoryStateDB
	Transactions  map[types.Hash]*MockTxInfo
	Receipts      map[types.Hash][]*types.Receipt
	Logs          map[types.Hash][]*types.Log
	SentTxs       []*types.Transaction
	CallResult    []byte
	CallGasUsed   uint64
	CallErr       error
	HistoryOldest uint64
}

// NewMockBackend returns a MockBackend pre-populated with a canonical header
// at block 42 and a funded account at 0xaaaa…:
//
//   - balance  = 1e18 wei
//   - nonce    = 5
//   - code     = 0x6000
//   - chainID  = 1337
func NewMockBackend() *MockBackend {
	sdb := state.NewMemoryStateDB()
	addr := types.HexToAddress("0xaaaa")
	sdb.AddBalance(addr, big.NewInt(1e18))
	sdb.SetNonce(addr, 5)
	sdb.SetCode(addr, []byte{0x60, 0x00})

	header := &types.Header{
		Number:   big.NewInt(42),
		GasLimit: 30_000_000,
		GasUsed:  15_000_000,
		Time:     1_700_000_000,
		BaseFee:  big.NewInt(1_000_000_000),
	}

	return &MockBackend{
		ChainIDVal:   big.NewInt(1337),
		Headers:      map[uint64]*types.Header{42: header},
		Blocks:       make(map[uint64]*types.Block),
		Statedb:      sdb,
		Transactions: make(map[types.Hash]*MockTxInfo),
		Receipts:     make(map[types.Hash][]*types.Receipt),
		Logs:         make(map[types.Hash][]*types.Log),
	}
}

// --- rpcbackend.Backend implementation ---

func (b *MockBackend) HeaderByNumber(number rpctypes.BlockNumber) *types.Header {
	if number == rpctypes.LatestBlockNumber ||
		number == rpctypes.SafeBlockNumber ||
		number == rpctypes.FinalizedBlockNumber {
		return b.Headers[42]
	}
	return b.Headers[uint64(number)]
}

func (b *MockBackend) HeaderByHash(hash types.Hash) *types.Header {
	for _, h := range b.Headers {
		if h.Hash() == hash {
			return h
		}
	}
	return nil
}

func (b *MockBackend) BlockByNumber(number rpctypes.BlockNumber) *types.Block {
	if number == rpctypes.LatestBlockNumber ||
		number == rpctypes.SafeBlockNumber ||
		number == rpctypes.FinalizedBlockNumber {
		return b.Blocks[42]
	}
	return b.Blocks[uint64(number)]
}

func (b *MockBackend) BlockByHash(hash types.Hash) *types.Block {
	for _, block := range b.Blocks {
		if block != nil && block.Hash() == hash {
			return block
		}
	}
	return nil
}

func (b *MockBackend) CurrentHeader() *types.Header {
	return b.Headers[42]
}

func (b *MockBackend) ChainID() *big.Int {
	return b.ChainIDVal
}

func (b *MockBackend) StateAt(_ types.Hash) (state.StateDB, error) {
	return b.Statedb, nil
}

func (b *MockBackend) SendTransaction(tx *types.Transaction) error {
	b.SentTxs = append(b.SentTxs, tx)
	return nil
}

func (b *MockBackend) GetTransaction(hash types.Hash) (*types.Transaction, uint64, uint64) {
	if info, ok := b.Transactions[hash]; ok {
		return info.Tx, info.BlockNum, info.Index
	}
	return nil, 0, 0
}

func (b *MockBackend) SuggestGasPrice() *big.Int {
	return big.NewInt(1_000_000_000)
}

func (b *MockBackend) GetReceipts(blockHash types.Hash) []*types.Receipt {
	return b.Receipts[blockHash]
}

func (b *MockBackend) GetLogs(blockHash types.Hash) []*types.Log {
	return b.Logs[blockHash]
}

func (b *MockBackend) GetBlockReceipts(number uint64) []*types.Receipt {
	header := b.Headers[number]
	if header == nil {
		return nil
	}
	return b.Receipts[header.Hash()]
}

func (b *MockBackend) EVMCall(_ types.Address, _ *types.Address, _ []byte, _ uint64, _ *big.Int, _ rpctypes.BlockNumber) ([]byte, uint64, error) {
	return b.CallResult, b.CallGasUsed, b.CallErr
}

func (b *MockBackend) GetProof(addr types.Address, storageKeys []types.Hash, _ rpctypes.BlockNumber) (*trie.AccountProof, error) {
	stateTrie := b.Statedb.BuildStateTrie()
	storageTrie := b.Statedb.BuildStorageTrie(addr)
	return trie.ProveAccountWithStorage(stateTrie, addr, storageTrie, storageKeys)
}

func (b *MockBackend) TraceTransaction(_ types.Hash) (*vm.StructLogTracer, error) {
	return vm.NewStructLogTracer(), nil
}

func (b *MockBackend) HistoryOldestBlock() uint64 {
	return b.HistoryOldest
}

// BlobSchedule returns Cancun blob parameters as the default for tests.
func (b *MockBackend) BlobSchedule(_ uint64) (target, max, updateFraction uint64) {
	return 3, 6, 3338477 // Cancun/Dencun defaults (EIP-4844)
}

// MinTestRawTxHex returns a minimal RLP-encoded legacy transaction as a
// 0x-prefixed hex string for use in eth_sendRawTransaction tests.
func MinTestRawTxHex() (string, error) {
	to := types.HexToAddress("0x1111111111111111111111111111111111111111")
	inner := &types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1_000_000_000),
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(0),
		V:        big.NewInt(27),
		R:        big.NewInt(1),
		S:        big.NewInt(1),
	}
	tx := types.NewTransaction(inner)
	raw, err := tx.EncodeRLP()
	if err != nil {
		return "", fmt.Errorf("MinTestRawTxHex: EncodeRLP: %w", err)
	}
	return fmt.Sprintf("0x%x", raw), nil
}
