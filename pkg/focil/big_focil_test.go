package focil

import (
	"math/big"
	"sync"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func testTx(sender types.Address, gas uint64) *types.Transaction {
	to := types.HexToAddress("0xdeaddeaddeaddeaddeaddeaddeaddeaddeaddead")
	tx := types.NewTransaction(&types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1000),
		Gas:      gas,
		To:       &to,
		Value:    big.NewInt(0),
		V:        big.NewInt(27),
		R:        big.NewInt(1),
		S:        big.NewInt(1),
	})
	tx.SetSender(sender)
	return tx
}

func testTxWithPrice(sender types.Address, gas uint64, price int64) *types.Transaction {
	to := types.HexToAddress("0xdeaddeaddeaddeaddeaddeaddeaddeaddeaddead")
	tx := types.NewTransaction(&types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(price),
		Gas:      gas,
		To:       &to,
		Value:    big.NewInt(0),
		V:        big.NewInt(27),
		R:        big.NewInt(1),
		S:        big.NewInt(1),
	})
	tx.SetSender(sender)
	return tx
}

// --- SenderHexPartition tests ---

func TestSenderHexPartition(t *testing.T) {
	tests := []struct {
		name      string
		addr      types.Address
		wantPart  uint8
	}{
		{"0x0_ prefix", addrWithPrefix(0x00), 0},
		{"0x1_ prefix", addrWithPrefix(0x10), 1},
		{"0x2_ prefix", addrWithPrefix(0x20), 2},
		{"0x3_ prefix", addrWithPrefix(0x30), 3},
		{"0x4_ prefix", addrWithPrefix(0x40), 4},
		{"0x5_ prefix", addrWithPrefix(0x50), 5},
		{"0x6_ prefix", addrWithPrefix(0x60), 6},
		{"0x7_ prefix", addrWithPrefix(0x70), 7},
		{"0x8_ prefix", addrWithPrefix(0x80), 8},
		{"0x9_ prefix", addrWithPrefix(0x90), 9},
		{"0xA_ prefix", addrWithPrefix(0xA0), 10},
		{"0xB_ prefix", addrWithPrefix(0xB0), 11},
		{"0xC_ prefix", addrWithPrefix(0xC0), 12},
		{"0xD_ prefix", addrWithPrefix(0xD0), 13},
		{"0xE_ prefix", addrWithPrefix(0xE0), 14},
		{"0xF_ prefix", addrWithPrefix(0xF0), 15},
		{"low nibble ignored", addrWithPrefix(0x0F), 0},
		{"both nibbles", addrWithPrefix(0xAB), 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SenderHexPartition(tt.addr)
			if got != tt.wantPart {
				t.Errorf("SenderHexPartition(%x) = %d, want %d", tt.addr[0], got, tt.wantPart)
			}
		})
	}
}

func addrWithPrefix(firstByte byte) types.Address {
	var addr types.Address
	addr[0] = firstByte
	return addr
}

// --- AssignPartitions tests ---

func TestAssignPartitions(t *testing.T) {
	t.Run("16 members 1:1", func(t *testing.T) {
		m := AssignPartitions(16)
		if len(m) != NumPartitions {
			t.Fatalf("len = %d, want %d", len(m), NumPartitions)
		}
		for i := uint8(0); i < NumPartitions; i++ {
			if m[i] != int(i) {
				t.Errorf("partition %d -> member %d, want %d", i, m[i], i)
			}
		}
	})

	t.Run("fewer members wrapping", func(t *testing.T) {
		m := AssignPartitions(4)
		if len(m) != NumPartitions {
			t.Fatalf("len = %d, want %d", len(m), NumPartitions)
		}
		for i := uint8(0); i < NumPartitions; i++ {
			want := int(i) % 4
			if m[i] != want {
				t.Errorf("partition %d -> member %d, want %d", i, m[i], want)
			}
		}
	})

	t.Run("more than 16 members", func(t *testing.T) {
		m := AssignPartitions(32)
		if len(m) != NumPartitions {
			t.Fatalf("len = %d, want %d", len(m), NumPartitions)
		}
		// Each partition still maps to 0-15 since i % 32 = i for i < 16.
		for i := uint8(0); i < NumPartitions; i++ {
			if m[i] != int(i) {
				t.Errorf("partition %d -> member %d, want %d", i, m[i], i)
			}
		}
	})

	t.Run("single member", func(t *testing.T) {
		m := AssignPartitions(1)
		for i := uint8(0); i < NumPartitions; i++ {
			if m[i] != 0 {
				t.Errorf("partition %d -> member %d, want 0", i, m[i])
			}
		}
	})
}

// --- FilterByPartition tests ---

func TestFilterByPartition(t *testing.T) {
	addr0 := addrWithPrefix(0x00) // partition 0
	addr5 := addrWithPrefix(0x50) // partition 5
	addrF := addrWithPrefix(0xF0) // partition 15

	tx0 := testTx(addr0, 21000)
	tx5 := testTx(addr5, 21000)
	txF := testTx(addrF, 21000)

	// tx with nil sender (no SetSender called)
	txNil := types.NewTransaction(&types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1),
		Gas:      21000,
		V:        big.NewInt(27),
		R:        big.NewInt(1),
		S:        big.NewInt(1),
	})

	all := []*types.Transaction{tx0, tx5, txF, txNil}

	t.Run("partition 0", func(t *testing.T) {
		got := FilterByPartition(all, 0)
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
	})

	t.Run("partition 5", func(t *testing.T) {
		got := FilterByPartition(all, 5)
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
	})

	t.Run("partition 15", func(t *testing.T) {
		got := FilterByPartition(all, 15)
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
	})

	t.Run("nil sender skipped", func(t *testing.T) {
		// Partition 0 should only get tx0, not txNil.
		got := FilterByPartition(all, 0)
		if len(got) != 1 {
			t.Errorf("nil sender should be skipped, got %d txs", len(got))
		}
	})

	t.Run("empty partition", func(t *testing.T) {
		got := FilterByPartition(all, 7)
		if len(got) != 0 {
			t.Errorf("expected 0 for unused partition, got %d", len(got))
		}
	})
}

// --- BuildPartitionedList tests ---

func TestBuildPartitionedList(t *testing.T) {
	addr0 := addrWithPrefix(0x00)
	addr5 := addrWithPrefix(0x50)

	tx1 := testTxWithPrice(addr0, 21000, 500)
	tx2 := testTxWithPrice(addr0, 21000, 1000)
	tx3 := testTxWithPrice(addr5, 21000, 2000) // different partition

	pending := []*types.Transaction{tx1, tx2, tx3}

	t.Run("basic build without carryover", func(t *testing.T) {
		pl := BuildPartitionedList(pending, 0, 3, 100, nil)
		if pl.Partition != 0 {
			t.Errorf("Partition = %d, want 0", pl.Partition)
		}
		if pl.MemberIndex != 3 {
			t.Errorf("MemberIndex = %d, want 3", pl.MemberIndex)
		}
		if pl.SlotNumber != 100 {
			t.Errorf("SlotNumber = %d, want 100", pl.SlotNumber)
		}
		// tx1 and tx2 are partition 0; tx3 is partition 5.
		if len(pl.Transactions) != 2 {
			t.Errorf("Transactions = %d, want 2", len(pl.Transactions))
		}
		if len(pl.CarryoverTxs) != 0 {
			t.Errorf("CarryoverTxs = %d, want 0", len(pl.CarryoverTxs))
		}
	})

	t.Run("build with carryover", func(t *testing.T) {
		carryHash := types.Hash{0xAA}
		carryover := []CarryoverEntry{
			{TxHash: carryHash, Partition: 0, FirstSlot: 98, Priority: 2},
		}
		pl := BuildPartitionedList(pending, 0, 3, 100, carryover)
		if len(pl.CarryoverTxs) != 1 {
			t.Fatalf("CarryoverTxs = %d, want 1", len(pl.CarryoverTxs))
		}
		if pl.CarryoverTxs[0] != carryHash {
			t.Errorf("CarryoverTxs[0] = %x, want %x", pl.CarryoverTxs[0], carryHash)
		}
	})

	t.Run("TotalTransactions includes carryover", func(t *testing.T) {
		carryover := []CarryoverEntry{
			{TxHash: types.Hash{0xBB}, Partition: 0, FirstSlot: 98, Priority: 1},
		}
		pl := BuildPartitionedList(pending, 0, 0, 100, carryover)
		total := pl.TotalTransactions()
		if total != len(pl.Transactions)+1 {
			t.Errorf("TotalTransactions = %d, want %d", total, len(pl.Transactions)+1)
		}
	})
}

// --- CarryoverTracker tests ---

func TestCarryoverTracker_AddUnincluded(t *testing.T) {
	ct := NewCarryoverTracker(4)

	hash1 := types.Hash{0x01}
	ct.AddUnincluded(hash1, 0, 10)

	if ct.Len() != 1 {
		t.Fatalf("Len = %d, want 1", ct.Len())
	}

	// Adding same hash again should bump priority, not add new entry.
	ct.AddUnincluded(hash1, 0, 10)
	if ct.Len() != 1 {
		t.Errorf("Len = %d after re-add, want 1", ct.Len())
	}

	entries := ct.GetCarryover(0, 11)
	if len(entries) != 1 {
		t.Fatalf("GetCarryover len = %d, want 1", len(entries))
	}
	if entries[0].Priority != 2 {
		t.Errorf("Priority = %d, want 2 (incremented)", entries[0].Priority)
	}
}

func TestCarryoverTracker_AddUnincluded_PriorityCap(t *testing.T) {
	ct := NewCarryoverTracker(1000)
	hash := types.Hash{0x01}
	ct.AddUnincluded(hash, 0, 10)

	// Bump priority to 255.
	for i := 0; i < 300; i++ {
		ct.AddUnincluded(hash, 0, 10)
	}

	entries := ct.GetCarryover(0, 11)
	if len(entries) != 1 {
		t.Fatalf("len = %d, want 1", len(entries))
	}
	if entries[0].Priority != 255 {
		t.Errorf("Priority = %d, want 255 (capped)", entries[0].Priority)
	}
}

func TestCarryoverTracker_MarkIncluded(t *testing.T) {
	ct := NewCarryoverTracker(4)

	hash1 := types.Hash{0x01}
	hash2 := types.Hash{0x02}
	ct.AddUnincluded(hash1, 0, 10)
	ct.AddUnincluded(hash2, 0, 10)

	ct.MarkIncluded(hash1)
	if ct.Len() != 1 {
		t.Errorf("Len = %d after MarkIncluded, want 1", ct.Len())
	}

	// Mark non-existent -- should not panic.
	ct.MarkIncluded(types.Hash{0xFF})
	if ct.Len() != 1 {
		t.Errorf("Len = %d, want 1 (no change)", ct.Len())
	}
}

func TestCarryoverTracker_GetCarryover(t *testing.T) {
	ct := NewCarryoverTracker(4)

	hash0 := types.Hash{0x01}
	hash5 := types.Hash{0x02}
	ct.AddUnincluded(hash0, 0, 10)
	ct.AddUnincluded(hash5, 5, 10)

	t.Run("partition filtering", func(t *testing.T) {
		entries := ct.GetCarryover(0, 11)
		if len(entries) != 1 {
			t.Errorf("partition 0: len = %d, want 1", len(entries))
		}
		entries = ct.GetCarryover(5, 11)
		if len(entries) != 1 {
			t.Errorf("partition 5: len = %d, want 1", len(entries))
		}
		entries = ct.GetCarryover(7, 11)
		if len(entries) != 0 {
			t.Errorf("partition 7: len = %d, want 0", len(entries))
		}
	})

	t.Run("priority sort", func(t *testing.T) {
		ct2 := NewCarryoverTracker(10)
		hashA := types.Hash{0xAA}
		hashB := types.Hash{0xBB}
		hashC := types.Hash{0xCC}

		ct2.AddUnincluded(hashA, 0, 10) // priority 1
		ct2.AddUnincluded(hashB, 0, 10) // priority 1
		ct2.AddUnincluded(hashC, 0, 10) // priority 1

		// Bump hashC priority to 3.
		ct2.AddUnincluded(hashC, 0, 10)
		ct2.AddUnincluded(hashC, 0, 10)
		// Bump hashA priority to 2.
		ct2.AddUnincluded(hashA, 0, 10)

		entries := ct2.GetCarryover(0, 11)
		if len(entries) != 3 {
			t.Fatalf("len = %d, want 3", len(entries))
		}
		if entries[0].Priority < entries[1].Priority || entries[1].Priority < entries[2].Priority {
			t.Errorf("not sorted by priority desc: %d, %d, %d",
				entries[0].Priority, entries[1].Priority, entries[2].Priority)
		}
	})

	t.Run("expiry", func(t *testing.T) {
		ct3 := NewCarryoverTracker(4)
		hash := types.Hash{0xDD}
		ct3.AddUnincluded(hash, 0, 10)

		// Within carryover window.
		entries := ct3.GetCarryover(0, 14)
		if len(entries) != 1 {
			t.Errorf("within window: len = %d, want 1", len(entries))
		}

		// Expired (10 + 4 < 15).
		entries = ct3.GetCarryover(0, 15)
		if len(entries) != 0 {
			t.Errorf("expired: len = %d, want 0", len(entries))
		}
	})
}

func TestCarryoverTracker_Prune(t *testing.T) {
	ct := NewCarryoverTracker(4)

	ct.AddUnincluded(types.Hash{0x01}, 0, 10)
	ct.AddUnincluded(types.Hash{0x02}, 0, 12)
	ct.AddUnincluded(types.Hash{0x03}, 0, 20)

	// At slot 16: hash 0x01 (first=10, age=6 > 4) should be pruned.
	// hash 0x02 (first=12, age=4, not > 4) should remain.
	// hash 0x03 (first=20, age < 0) should remain.
	pruned := ct.Prune(16)
	if pruned != 1 {
		t.Errorf("pruned = %d, want 1", pruned)
	}
	if ct.Len() != 2 {
		t.Errorf("Len = %d after prune, want 2", ct.Len())
	}
}

func TestCarryoverTracker_Concurrent(t *testing.T) {
	ct := NewCarryoverTracker(100)
	var wg sync.WaitGroup

	// Concurrent adds.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var h types.Hash
			h[0] = byte(i)
			ct.AddUnincluded(h, uint8(i%16), uint64(i))
		}(i)
	}

	// Concurrent marks.
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var h types.Hash
			h[0] = byte(i)
			ct.MarkIncluded(h)
		}(i)
	}

	// Concurrent gets.
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ct.GetCarryover(uint8(i), 100)
		}(i)
	}

	// Concurrent prune.
	wg.Add(1)
	go func() {
		defer wg.Done()
		ct.Prune(200)
	}()

	wg.Wait()

	// No panics or races -- that is the test.
	_ = ct.Len()
}

func TestCarryoverTracker_DefaultMaxSlots(t *testing.T) {
	ct := NewCarryoverTracker(0)
	// maxSlots should default to DefaultCarryoverSlots.
	hash := types.Hash{0x01}
	ct.AddUnincluded(hash, 0, 10)

	entries := ct.GetCarryover(0, 10+DefaultCarryoverSlots)
	if len(entries) != 1 {
		t.Errorf("within default window: len = %d, want 1", len(entries))
	}

	entries = ct.GetCarryover(0, 10+DefaultCarryoverSlots+1)
	if len(entries) != 0 {
		t.Errorf("expired with default: len = %d, want 0", len(entries))
	}
}
