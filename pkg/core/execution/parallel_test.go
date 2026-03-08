package execution

import (
	"errors"
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/bal"
	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
)

// makeTransferTx creates a legacy transfer transaction.
func makeTransferTx(nonce uint64, to types.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int) *types.Transaction {
	toAddr := to
	return types.NewTransaction(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddr,
		Value:    value,
	})
}

// setupState funds accounts and returns a ready MemoryStateDB.
func setupState(accounts map[types.Address]*big.Int) *state.MemoryStateDB {
	statedb := state.NewMemoryStateDB()
	for addr, balance := range accounts {
		statedb.AddBalance(addr, balance)
	}
	return statedb
}

// buildBALForIndependentTransfers builds a BlockAccessList where each transfer
// has its own sender/recipient pair and no entries conflict.
func buildBALForIndependentTransfers(senders, recipients []types.Address) *bal.BlockAccessList {
	b := bal.NewBlockAccessList()
	for i := range senders {
		b.AddEntry(bal.AccessEntry{
			Address:     senders[i],
			AccessIndex: uint64(i + 1),
			BalanceChange: &bal.BalanceChange{
				OldValue: big.NewInt(0),
				NewValue: big.NewInt(0),
			},
			NonceChange: &bal.NonceChange{
				OldValue: 0,
				NewValue: 1,
			},
		})
		b.AddEntry(bal.AccessEntry{
			Address:     recipients[i],
			AccessIndex: uint64(i + 1),
			BalanceChange: &bal.BalanceChange{
				OldValue: big.NewInt(0),
				NewValue: big.NewInt(0),
			},
		})
	}
	return b
}

// buildBALForConflictingTransfers builds a BlockAccessList where all transfers
// touch a shared address (making them conflict and serialize into one-per-group).
func buildBALForConflictingTransfers(senders []types.Address, sharedRecipient types.Address, n int) *bal.BlockAccessList {
	b := bal.NewBlockAccessList()
	for i := 0; i < n; i++ {
		b.AddEntry(bal.AccessEntry{
			Address:     senders[i],
			AccessIndex: uint64(i + 1),
			BalanceChange: &bal.BalanceChange{
				OldValue: big.NewInt(0),
				NewValue: big.NewInt(0),
			},
			NonceChange: &bal.NonceChange{
				OldValue: 0,
				NewValue: 1,
			},
		})
		b.AddEntry(bal.AccessEntry{
			Address:     sharedRecipient,
			AccessIndex: uint64(i + 1),
			BalanceChange: &bal.BalanceChange{
				OldValue: big.NewInt(0),
				NewValue: big.NewInt(0),
			},
		})
	}
	return b
}

func TestParallelProcessIndependentTransactions(t *testing.T) {
	senders := []types.Address{
		types.HexToAddress("0xA1"),
		types.HexToAddress("0xC1"),
		types.HexToAddress("0xE1"),
	}
	recipients := []types.Address{
		types.HexToAddress("0xB1"),
		types.HexToAddress("0xD1"),
		types.HexToAddress("0xF1"),
	}

	funding := new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(1e18))
	accounts := make(map[types.Address]*big.Int)
	for _, s := range senders {
		accounts[s] = new(big.Int).Set(funding)
	}
	statedb := setupState(accounts)

	transferAmt := new(big.Int).SetUint64(1e18)
	gasPrice := big.NewInt(1)
	gasLimit := uint64(50000)

	txs := make([]*types.Transaction, 3)
	for i := range senders {
		tx := makeTransferTx(0, recipients[i], transferAmt, gasLimit, gasPrice)
		tx.SetSender(senders[i])
		txs[i] = tx
	}

	header := newTestHeader()
	block := types.NewBlock(header, &types.Body{Transactions: txs})

	accessList := buildBALForIndependentTransfers(senders, recipients)

	proc := NewParallelProcessor(config.TestConfig)
	receipts, err := proc.ProcessParallel(statedb, block, accessList)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(receipts) != 3 {
		t.Fatalf("expected 3 receipts, got %d", len(receipts))
	}

	for i, r := range receipts {
		if r == nil {
			t.Fatalf("receipt %d is nil", i)
		}
		if r.Status != types.ReceiptStatusSuccessful {
			t.Fatalf("receipt %d: expected successful status, got %d", i, r.Status)
		}
		if r.TransactionIndex != uint(i) {
			t.Fatalf("receipt %d: expected TransactionIndex %d, got %d", i, i, r.TransactionIndex)
		}
	}

	for i, r := range recipients {
		b := statedb.GetBalance(r)
		if b.Cmp(transferAmt) != 0 {
			t.Fatalf("recipient %d balance: want %v, got %v", i, transferAmt, b)
		}
	}
}

func TestParallelProcessFallbackSequential(t *testing.T) {
	sender := types.HexToAddress("0xA1")
	recipient := types.HexToAddress("0xB1")

	funding := new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(1e18))
	statedb := setupState(map[types.Address]*big.Int{sender: funding})

	transferAmt := new(big.Int).SetUint64(1e18)
	gasPrice := big.NewInt(1)
	gasLimit := uint64(50000)

	tx := makeTransferTx(0, recipient, transferAmt, gasLimit, gasPrice)
	tx.SetSender(sender)

	header := newTestHeader()
	block := types.NewBlock(header, &types.Body{Transactions: []*types.Transaction{tx}})

	proc := NewParallelProcessor(config.TestConfig)
	receipts, err := proc.ProcessParallel(statedb, block, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(receipts) != 1 {
		t.Fatalf("expected 1 receipt, got %d", len(receipts))
	}
	if receipts[0].Status != types.ReceiptStatusSuccessful {
		t.Fatalf("expected successful receipt, got status %d", receipts[0].Status)
	}

	recipientBal := statedb.GetBalance(recipient)
	if recipientBal.Cmp(transferAmt) != 0 {
		t.Fatalf("recipient balance: want %v, got %v", transferAmt, recipientBal)
	}
}

func TestParallelProcessEmptyBlock(t *testing.T) {
	statedb := state.NewMemoryStateDB()
	header := newTestHeader()
	block := types.NewBlock(header, &types.Body{})

	proc := NewParallelProcessor(config.TestConfig)
	receipts, err := proc.ProcessParallel(statedb, block, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(receipts) != 0 {
		t.Fatalf("expected 0 receipts, got %d", len(receipts))
	}
}

func TestValidateBAL(t *testing.T) {
	senders := []types.Address{types.HexToAddress("0xA1")}
	recipients := []types.Address{types.HexToAddress("0xB1")}
	accessList := buildBALForIndependentTransfers(senders, recipients)

	correctHash := accessList.Hash()

	t.Run("valid BAL", func(t *testing.T) {
		header := &types.Header{
			Number:              big.NewInt(1),
			BlockAccessListHash: &correctHash,
		}
		if err := ValidateBAL(header, accessList); err != nil {
			t.Fatalf("expected valid BAL, got error: %v", err)
		}
	})

	t.Run("hash mismatch", func(t *testing.T) {
		wrongHash := types.HexToHash("0xdeadbeef")
		header := &types.Header{
			Number:              big.NewInt(1),
			BlockAccessListHash: &wrongHash,
		}
		err := ValidateBAL(header, accessList)
		if err == nil {
			t.Fatal("expected hash mismatch error")
		}
		if !errors.Is(err, ErrBALHashMismatch) {
			t.Fatalf("expected ErrBALHashMismatch, got: %v", err)
		}
	})

	t.Run("nil header hash", func(t *testing.T) {
		header := &types.Header{
			Number:              big.NewInt(1),
			BlockAccessListHash: nil,
		}
		err := ValidateBAL(header, accessList)
		if err == nil {
			t.Fatal("expected error for nil header hash")
		}
	})

	t.Run("nil access list", func(t *testing.T) {
		header := &types.Header{
			Number:              big.NewInt(1),
			BlockAccessListHash: &correctHash,
		}
		err := ValidateBAL(header, nil)
		if err == nil {
			t.Fatal("expected error for nil access list")
		}
	})
}

func TestParallelProcessReceiptOrdering(t *testing.T) {
	n := 5
	senders := make([]types.Address, n)
	recipients := make([]types.Address, n)

	for i := 0; i < n; i++ {
		senders[i] = types.BytesToAddress([]byte{byte(0x10 + i)})
		recipients[i] = types.BytesToAddress([]byte{byte(0x50 + i)})
	}

	funding := new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(1e18))
	accounts := make(map[types.Address]*big.Int)
	for _, s := range senders {
		accounts[s] = new(big.Int).Set(funding)
	}
	statedb := setupState(accounts)

	gasPrice := big.NewInt(1)
	gasLimit := uint64(50000)

	txs := make([]*types.Transaction, n)
	for i := 0; i < n; i++ {
		amount := new(big.Int).Mul(big.NewInt(int64(i+1)), new(big.Int).SetUint64(1e17))
		tx := makeTransferTx(0, recipients[i], amount, gasLimit, gasPrice)
		tx.SetSender(senders[i])
		txs[i] = tx
	}

	header := newTestHeader()
	block := types.NewBlock(header, &types.Body{Transactions: txs})

	accessList := buildBALForIndependentTransfers(senders, recipients)

	proc := NewParallelProcessor(config.TestConfig)
	receipts, err := proc.ProcessParallel(statedb, block, accessList)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(receipts) != n {
		t.Fatalf("expected %d receipts, got %d", n, len(receipts))
	}

	for i, r := range receipts {
		if r == nil {
			t.Fatalf("receipt %d is nil", i)
		}
		if r.TransactionIndex != uint(i) {
			t.Fatalf("receipt %d has TransactionIndex %d, expected %d", i, r.TransactionIndex, i)
		}
		if r.Status != types.ReceiptStatusSuccessful {
			t.Fatalf("receipt %d: expected successful, got %d", i, r.Status)
		}
	}

	for i, r := range recipients {
		expected := new(big.Int).Mul(big.NewInt(int64(i+1)), new(big.Int).SetUint64(1e17))
		actual := statedb.GetBalance(r)
		if actual.Cmp(expected) != 0 {
			t.Fatalf("recipient %d balance: want %v, got %v", i, expected, actual)
		}
	}
}

func TestParallelProcessConflictingTransactions(t *testing.T) {
	n := 3
	senders := make([]types.Address, n)
	for i := 0; i < n; i++ {
		senders[i] = types.BytesToAddress([]byte{byte(0x10 + i)})
	}
	sharedRecipient := types.HexToAddress("0xAAAA")

	funding := new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(1e18))
	accounts := make(map[types.Address]*big.Int)
	for _, s := range senders {
		accounts[s] = new(big.Int).Set(funding)
	}
	statedb := setupState(accounts)

	gasPrice := big.NewInt(1)
	gasLimit := uint64(50000)
	transferAmt := new(big.Int).SetUint64(1e18)

	txs := make([]*types.Transaction, n)
	for i := 0; i < n; i++ {
		tx := makeTransferTx(0, sharedRecipient, transferAmt, gasLimit, gasPrice)
		tx.SetSender(senders[i])
		txs[i] = tx
	}

	header := newTestHeader()
	block := types.NewBlock(header, &types.Body{Transactions: txs})

	accessList := buildBALForConflictingTransfers(senders, sharedRecipient, n)

	proc := NewParallelProcessor(config.TestConfig)
	receipts, err := proc.ProcessParallel(statedb, block, accessList)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(receipts) != n {
		t.Fatalf("expected %d receipts, got %d", n, len(receipts))
	}

	for i, r := range receipts {
		if r == nil {
			t.Fatalf("receipt %d is nil", i)
		}
		if r.Status != types.ReceiptStatusSuccessful {
			t.Fatalf("receipt %d: expected successful, got %d", i, r.Status)
		}
	}

	expectedBal := new(big.Int).Mul(transferAmt, big.NewInt(int64(n)))
	actualBal := statedb.GetBalance(sharedRecipient)
	if actualBal.Cmp(expectedBal) != 0 {
		t.Fatalf("shared recipient balance: want %v, got %v", expectedBal, actualBal)
	}
}

func TestParallelProcessWithEmptyBAL(t *testing.T) {
	sender := types.HexToAddress("0xA1")
	recipient := types.HexToAddress("0xB1")

	funding := new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(1e18))
	statedb := setupState(map[types.Address]*big.Int{sender: funding})

	transferAmt := new(big.Int).SetUint64(1e18)
	gasPrice := big.NewInt(1)
	gasLimit := uint64(50000)

	tx := makeTransferTx(0, recipient, transferAmt, gasLimit, gasPrice)
	tx.SetSender(sender)

	header := newTestHeader()
	block := types.NewBlock(header, &types.Body{Transactions: []*types.Transaction{tx}})

	emptyBAL := bal.NewBlockAccessList()

	proc := NewParallelProcessor(config.TestConfig)
	receipts, err := proc.ProcessParallel(statedb, block, emptyBAL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(receipts) != 1 {
		t.Fatalf("expected 1 receipt, got %d", len(receipts))
	}
	if receipts[0].Status != types.ReceiptStatusSuccessful {
		t.Fatalf("expected successful receipt")
	}
}

func TestMemoryStateDBCopy(t *testing.T) {
	original := state.NewMemoryStateDB()
	addr := types.HexToAddress("0x1234")
	original.AddBalance(addr, big.NewInt(1000))
	original.SetNonce(addr, 5)

	copied := original.Copy()

	copied.AddBalance(addr, big.NewInt(500))
	copied.SetNonce(addr, 10)

	if original.GetBalance(addr).Cmp(big.NewInt(1000)) != 0 {
		t.Fatalf("original balance changed after copy modification: %v", original.GetBalance(addr))
	}
	if original.GetNonce(addr) != 5 {
		t.Fatalf("original nonce changed after copy modification: %d", original.GetNonce(addr))
	}

	if copied.GetBalance(addr).Cmp(big.NewInt(1500)) != 0 {
		t.Fatalf("copy balance: want 1500, got %v", copied.GetBalance(addr))
	}
	if copied.GetNonce(addr) != 10 {
		t.Fatalf("copy nonce: want 10, got %d", copied.GetNonce(addr))
	}
}

func TestMemoryStateDBMerge(t *testing.T) {
	dst := state.NewMemoryStateDB()
	addr := types.HexToAddress("0x1234")
	dst.AddBalance(addr, big.NewInt(1000))

	src := state.NewMemoryStateDB()
	src.AddBalance(addr, big.NewInt(2000))
	src.SetNonce(addr, 3)

	dst.Merge(src)

	if dst.GetBalance(addr).Cmp(big.NewInt(2000)) != 0 {
		t.Fatalf("merged balance: want 2000, got %v", dst.GetBalance(addr))
	}
	if dst.GetNonce(addr) != 3 {
		t.Fatalf("merged nonce: want 3, got %d", dst.GetNonce(addr))
	}
}
