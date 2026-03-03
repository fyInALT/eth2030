package core

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/bal"
	"github.com/eth2030/eth2030/core/types"
)

func makeTestTx(nonce uint64) *types.Transaction {
	to := types.Address{0x01}
	return types.NewTransaction(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(1),
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(0),
	})
}

func TestNewDependencyGraph_NilBAL(t *testing.T) {
	txs := []*types.Transaction{makeTestTx(0), makeTestTx(1)}
	dg := NewDependencyGraph(txs, nil)
	if dg == nil {
		t.Fatal("expected non-nil graph")
	}
	if dg.ConflictCount() != 0 {
		t.Fatalf("expected 0 conflicts with nil BAL, got %d", dg.ConflictCount())
	}
}

func TestNewDependencyGraph_Empty(t *testing.T) {
	dg := NewDependencyGraph(nil, nil)
	if dg == nil {
		t.Fatal("expected non-nil graph")
	}
	groups := dg.Partition(0)
	if groups != nil {
		t.Fatal("expected nil groups for empty graph")
	}
}

func TestDependencyGraph_NoConflicts(t *testing.T) {
	txs := []*types.Transaction{makeTestTx(0), makeTestTx(1), makeTestTx(2)}

	accessList := bal.NewBlockAccessList()
	// Tx 0 (AccessIndex=1) accesses addr 0x01, slot 0x10
	accessList.AddEntry(bal.AccessEntry{
		Address:     types.Address{0x01},
		AccessIndex: 1,
		StorageChanges: []bal.StorageChange{
			{Slot: types.Hash{0x10}},
		},
	})
	// Tx 1 (AccessIndex=2) accesses addr 0x02, slot 0x20
	accessList.AddEntry(bal.AccessEntry{
		Address:     types.Address{0x02},
		AccessIndex: 2,
		StorageChanges: []bal.StorageChange{
			{Slot: types.Hash{0x20}},
		},
	})
	// Tx 2 (AccessIndex=3) accesses addr 0x03, slot 0x30
	accessList.AddEntry(bal.AccessEntry{
		Address:     types.Address{0x03},
		AccessIndex: 3,
		StorageChanges: []bal.StorageChange{
			{Slot: types.Hash{0x30}},
		},
	})

	dg := NewDependencyGraph(txs, accessList)
	if dg.ConflictCount() != 0 {
		t.Fatalf("expected 0 conflicts, got %d", dg.ConflictCount())
	}

	groups := dg.Partition(0)
	// All txs should be in a single group since there are no conflicts.
	if len(groups) != 1 {
		t.Fatalf("expected 1 group (all independent), got %d", len(groups))
	}
	if len(groups[0].Transactions) != 3 {
		t.Fatalf("expected 3 txs in group, got %d", len(groups[0].Transactions))
	}
}

func TestDependencyGraph_WriteWriteConflict(t *testing.T) {
	txs := []*types.Transaction{makeTestTx(0), makeTestTx(1)}
	addr := types.Address{0x01}
	slot := types.Hash{0x10}

	accessList := bal.NewBlockAccessList()
	// Both tx 0 and tx 1 write to the same slot.
	accessList.AddEntry(bal.AccessEntry{
		Address:     addr,
		AccessIndex: 1,
		StorageChanges: []bal.StorageChange{
			{Slot: slot},
		},
	})
	accessList.AddEntry(bal.AccessEntry{
		Address:     addr,
		AccessIndex: 2,
		StorageChanges: []bal.StorageChange{
			{Slot: slot},
		},
	})

	dg := NewDependencyGraph(txs, accessList)
	if dg.ConflictCount() != 1 {
		t.Fatalf("expected 1 conflict, got %d", dg.ConflictCount())
	}
}

func TestDependencyGraph_ReadWriteConflict(t *testing.T) {
	txs := []*types.Transaction{makeTestTx(0), makeTestTx(1)}
	addr := types.Address{0x01}
	slot := types.Hash{0x10}

	accessList := bal.NewBlockAccessList()
	// Tx 0 reads from slot.
	accessList.AddEntry(bal.AccessEntry{
		Address:     addr,
		AccessIndex: 1,
		StorageReads: []bal.StorageAccess{
			{Slot: slot},
		},
	})
	// Tx 1 writes to the same slot.
	accessList.AddEntry(bal.AccessEntry{
		Address:     addr,
		AccessIndex: 2,
		StorageChanges: []bal.StorageChange{
			{Slot: slot},
		},
	})

	dg := NewDependencyGraph(txs, accessList)
	if dg.ConflictCount() != 1 {
		t.Fatalf("expected 1 conflict, got %d", dg.ConflictCount())
	}
}

func TestDependencyGraph_Partition(t *testing.T) {
	txs := []*types.Transaction{makeTestTx(0), makeTestTx(1), makeTestTx(2)}
	addr := types.Address{0x01}
	slot := types.Hash{0x10}

	accessList := bal.NewBlockAccessList()
	// Tx 0 and tx 1 conflict (write-write on same slot).
	accessList.AddEntry(bal.AccessEntry{
		Address:     addr,
		AccessIndex: 1,
		StorageChanges: []bal.StorageChange{
			{Slot: slot},
		},
	})
	accessList.AddEntry(bal.AccessEntry{
		Address:     addr,
		AccessIndex: 2,
		StorageChanges: []bal.StorageChange{
			{Slot: slot},
		},
	})
	// Tx 2 accesses a different address, no conflicts.
	accessList.AddEntry(bal.AccessEntry{
		Address:     types.Address{0x02},
		AccessIndex: 3,
		StorageChanges: []bal.StorageChange{
			{Slot: types.Hash{0x20}},
		},
	})

	dg := NewDependencyGraph(txs, accessList)
	groups := dg.Partition(0)

	if len(groups) < 2 {
		t.Fatalf("expected at least 2 groups, got %d", len(groups))
	}

	// Verify that total txs across groups equals original count.
	totalTxs := 0
	for _, g := range groups {
		totalTxs += len(g.Transactions)
	}
	if totalTxs != 3 {
		t.Fatalf("expected 3 total txs across groups, got %d", totalTxs)
	}
}

func TestDependencyGraph_Partition_MaxGroups(t *testing.T) {
	txs := []*types.Transaction{makeTestTx(0), makeTestTx(1), makeTestTx(2)}
	addr := types.Address{0x01}

	accessList := bal.NewBlockAccessList()
	// All three conflict pairwise: each writes to a different slot on the same address
	// but tx0-tx1 conflict on slot 0x10, tx1-tx2 conflict on slot 0x20, tx0-tx2 conflict on slot 0x30.
	accessList.AddEntry(bal.AccessEntry{
		Address:     addr,
		AccessIndex: 1,
		StorageChanges: []bal.StorageChange{
			{Slot: types.Hash{0x10}},
			{Slot: types.Hash{0x30}},
		},
	})
	accessList.AddEntry(bal.AccessEntry{
		Address:     addr,
		AccessIndex: 2,
		StorageChanges: []bal.StorageChange{
			{Slot: types.Hash{0x10}},
			{Slot: types.Hash{0x20}},
		},
	})
	accessList.AddEntry(bal.AccessEntry{
		Address:     addr,
		AccessIndex: 3,
		StorageChanges: []bal.StorageChange{
			{Slot: types.Hash{0x20}},
			{Slot: types.Hash{0x30}},
		},
	})

	dg := NewDependencyGraph(txs, accessList)

	// Without limit, should be 3 groups (complete conflict graph).
	unlimitedGroups := dg.Partition(0)
	if len(unlimitedGroups) < 2 {
		t.Fatalf("expected >= 2 groups without limit, got %d", len(unlimitedGroups))
	}

	// With maxGroups=2, should constrain to at most 2 groups.
	limitedGroups := dg.Partition(2)
	if len(limitedGroups) > 2 {
		t.Fatalf("expected at most 2 groups with limit, got %d", len(limitedGroups))
	}
}

func TestConflictCount(t *testing.T) {
	txs := []*types.Transaction{makeTestTx(0), makeTestTx(1), makeTestTx(2)}
	addr := types.Address{0x01}
	slot := types.Hash{0x10}

	accessList := bal.NewBlockAccessList()
	// Tx 0 and 1 conflict, tx 0 and 2 conflict, tx 1 and 2 do NOT conflict.
	accessList.AddEntry(bal.AccessEntry{
		Address:     addr,
		AccessIndex: 1,
		StorageChanges: []bal.StorageChange{
			{Slot: slot},
		},
	})
	accessList.AddEntry(bal.AccessEntry{
		Address:     addr,
		AccessIndex: 2,
		StorageChanges: []bal.StorageChange{
			{Slot: slot},
		},
	})
	accessList.AddEntry(bal.AccessEntry{
		Address:     types.Address{0x02},
		AccessIndex: 3,
		StorageReads: []bal.StorageAccess{
			{Slot: slot},
		},
	})

	dg := NewDependencyGraph(txs, accessList)
	count := dg.ConflictCount()
	if count != 1 {
		t.Fatalf("expected 1 conflict edge, got %d", count)
	}
}

func TestClassifyTransactions(t *testing.T) {
	to := types.Address{0x01}

	localTx := types.NewLocalTx(big.NewInt(1), 0, &to, big.NewInt(0),
		21000, big.NewInt(1), big.NewInt(1), nil, []byte{0x0a})
	globalTx := types.NewTransaction(&types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1),
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(0),
	})

	txs := []*types.Transaction{localTx, globalTx, localTx}
	local, global := ClassifyTransactions(txs)

	if len(local) != 2 {
		t.Fatalf("expected 2 local txs, got %d", len(local))
	}
	if len(global) != 1 {
		t.Fatalf("expected 1 global tx, got %d", len(global))
	}
}
