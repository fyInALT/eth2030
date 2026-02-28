package witness

import (
	"math/big"
	"sync"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// makeTestWitness creates a test BlockExecutionWitness with the given addresses
// and storage keys per address.
func makeTestWitness(blockNum uint64, addrs []types.Address, storageKeys int) *BlockExecutionWitness {
	var parentHash, stateRoot types.Hash
	parentHash[0] = byte(blockNum)
	stateRoot[0] = byte(blockNum + 100)

	bw := NewBlockExecutionWitness(parentHash, stateRoot, blockNum)
	for _, addr := range addrs {
		psa := &PreStateAccount{
			Nonce:    blockNum,
			Balance:  big.NewInt(int64(blockNum * 1000)).Bytes(),
			CodeHash: types.Hash{0x01},
			Storage:  make(map[types.Hash]types.Hash),
			Exists:   true,
		}
		for i := 0; i < storageKeys; i++ {
			var k, v types.Hash
			k[0] = byte(i)
			v[0] = byte(i + 1)
			v[31] = byte(blockNum)
			psa.Storage[k] = v
		}
		bw.PreState[addr] = psa
	}
	return bw
}

func TestWitnessAggregatorSingleBlock(t *testing.T) {
	wa := NewWitnessAggregator()

	addr := types.Address{0x01}
	bw := makeTestWitness(1, []types.Address{addr}, 2)

	if err := wa.AddBlockWitness(1, bw); err != nil {
		t.Fatalf("add block: %v", err)
	}

	wr, err := wa.BuildRangeWitness(1, 1)
	if err != nil {
		t.Fatalf("build range: %v", err)
	}
	if len(wr.Witnesses) != 1 {
		t.Errorf("expected 1 witness, got %d", len(wr.Witnesses))
	}
	if _, ok := wr.MergedPreState[addr]; !ok {
		t.Error("expected address in merged pre-state")
	}
}

func TestWitnessAggregatorMultipleBlocks(t *testing.T) {
	wa := NewWitnessAggregator()
	addr1 := types.Address{0x01}
	addr2 := types.Address{0x02}

	wa.AddBlockWitness(1, makeTestWitness(1, []types.Address{addr1}, 1))
	wa.AddBlockWitness(2, makeTestWitness(2, []types.Address{addr2}, 1))
	wa.AddBlockWitness(3, makeTestWitness(3, []types.Address{addr1, addr2}, 2))

	wr, err := wa.BuildRangeWitness(1, 3)
	if err != nil {
		t.Fatalf("build range: %v", err)
	}
	if len(wr.Witnesses) != 3 {
		t.Errorf("expected 3 witnesses, got %d", len(wr.Witnesses))
	}
	if len(wr.MergedPreState) != 2 {
		t.Errorf("expected 2 merged accounts, got %d", len(wr.MergedPreState))
	}
}

func TestWitnessAggregatorMinimizeDeduplicatesAccounts(t *testing.T) {
	wa := NewWitnessAggregator()
	addr := types.Address{0xAA}

	// Same account appears in both blocks.
	wa.AddBlockWitness(1, makeTestWitness(1, []types.Address{addr}, 1))
	wa.AddBlockWitness(2, makeTestWitness(2, []types.Address{addr}, 1))

	wr, _ := wa.BuildRangeWitness(1, 2)
	mw := wa.Minimize(wr)

	// Only one entry for the account.
	if len(mw.Accounts) != 1 {
		t.Errorf("expected 1 deduplicated account, got %d", len(mw.Accounts))
	}
}

func TestWitnessAggregatorMinimizeDeduplicatesStorage(t *testing.T) {
	wa := NewWitnessAggregator()
	addr := types.Address{0xBB}

	bw1 := makeTestWitness(1, []types.Address{addr}, 3)
	bw2 := makeTestWitness(2, []types.Address{addr}, 3)
	// Add a storage diff to bw2 that modifies key 0.
	var diffAddr [20]byte
	copy(diffAddr[:], addr[:])
	bw2.StateDiffs = append(bw2.StateDiffs, StateDiff{
		Address: diffAddr,
		StorageChanges: []StorageChange{
			{Key: [32]byte{0x00}, OldValue: [32]byte{0x01}, NewValue: [32]byte{0xFF}},
		},
	})

	wa.AddBlockWitness(1, bw1)
	wa.AddBlockWitness(2, bw2)

	wr, _ := wa.BuildRangeWitness(1, 2)
	mw := wa.Minimize(wr)

	// FinalStorage should have the latest value for key 0.
	finalVal, ok := mw.FinalStorage[addr][types.Hash{0x00}]
	if !ok {
		t.Fatal("expected final storage entry for key 0")
	}
	if finalVal != (types.Hash{0xFF}) {
		t.Errorf("expected final value 0xFF, got %x", finalVal)
	}
}

func TestWitnessAggregatorGapError(t *testing.T) {
	wa := NewWitnessAggregator()
	addr := types.Address{0x01}

	wa.AddBlockWitness(1, makeTestWitness(1, []types.Address{addr}, 1))
	wa.AddBlockWitness(3, makeTestWitness(3, []types.Address{addr}, 1))
	// Block 2 is missing.

	_, err := wa.BuildRangeWitness(1, 3)
	if err == nil {
		t.Fatal("expected gap error")
	}
}

func TestWitnessAggregatorOutOfOrder(t *testing.T) {
	wa := NewWitnessAggregator()
	addr := types.Address{0x01}

	// Add blocks out of order.
	wa.AddBlockWitness(3, makeTestWitness(3, []types.Address{addr}, 1))
	wa.AddBlockWitness(1, makeTestWitness(1, []types.Address{addr}, 1))
	wa.AddBlockWitness(2, makeTestWitness(2, []types.Address{addr}, 1))

	wr, err := wa.BuildRangeWitness(1, 3)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(wr.Witnesses) != 3 {
		t.Errorf("expected 3 witnesses, got %d", len(wr.Witnesses))
	}
}

func TestWitnessAggregatorEmptyError(t *testing.T) {
	wa := NewWitnessAggregator()
	_, err := wa.BuildRangeWitness(1, 5)
	if err != ErrAggregatorEmpty {
		t.Errorf("expected ErrAggregatorEmpty, got %v", err)
	}
}

func TestWitnessAggregatorPrune(t *testing.T) {
	wa := NewWitnessAggregator()
	addr := types.Address{0x01}

	for i := uint64(1); i <= 10; i++ {
		wa.AddBlockWitness(i, makeTestWitness(i, []types.Address{addr}, 1))
	}

	pruned := wa.PruneOlderThan(6)
	if pruned != 5 {
		t.Errorf("expected 5 pruned, got %d", pruned)
	}
	if wa.BlockCount() != 5 {
		t.Errorf("expected 5 remaining, got %d", wa.BlockCount())
	}

	// Blocks 1-5 should be gone.
	_, err := wa.BuildRangeWitness(1, 5)
	if err == nil {
		t.Error("expected error for pruned blocks")
	}

	// Blocks 6-10 should still be available.
	wr, err := wa.BuildRangeWitness(6, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wr.Witnesses) != 5 {
		t.Errorf("expected 5 witnesses, got %d", len(wr.Witnesses))
	}
}

func TestWitnessAggregatorBlockCount(t *testing.T) {
	wa := NewWitnessAggregator()
	if wa.BlockCount() != 0 {
		t.Errorf("expected 0, got %d", wa.BlockCount())
	}

	addr := types.Address{0x01}
	wa.AddBlockWitness(1, makeTestWitness(1, []types.Address{addr}, 0))
	wa.AddBlockWitness(2, makeTestWitness(2, []types.Address{addr}, 0))

	if wa.BlockCount() != 2 {
		t.Errorf("expected 2, got %d", wa.BlockCount())
	}
}

func TestWitnessAggregatorSizeEstimation(t *testing.T) {
	wa := NewWitnessAggregator()
	addr := types.Address{0x01}
	bw := makeTestWitness(1, []types.Address{addr}, 5)
	wa.AddBlockWitness(1, bw)

	wr, _ := wa.BuildRangeWitness(1, 1)
	size := wa.Size(wr)
	if size <= 0 {
		t.Errorf("expected positive size, got %d", size)
	}
}

func TestAccessSetMergeDisjoint(t *testing.T) {
	a := map[string]bool{"x": true, "y": true}
	b := map[string]bool{"z": true}
	merged := MergeAccessSets(a, b)
	if len(merged) != 3 {
		t.Errorf("expected 3 keys, got %d", len(merged))
	}
}

func TestAccessSetMergeOverlapping(t *testing.T) {
	a := map[string]bool{"x": true, "y": true}
	b := map[string]bool{"y": true, "z": true}
	merged := MergeAccessSets(a, b)
	if len(merged) != 3 {
		t.Errorf("expected 3 keys, got %d", len(merged))
	}
}

func TestAccessSetMergeEmpty(t *testing.T) {
	merged := MergeAccessSets()
	if len(merged) != 0 {
		t.Errorf("expected 0 keys, got %d", len(merged))
	}

	merged2 := MergeAccessSets(map[string]bool{}, map[string]bool{})
	if len(merged2) != 0 {
		t.Errorf("expected 0 keys, got %d", len(merged2))
	}
}

func TestWitnessAggregatorAccountCreatedThenModified(t *testing.T) {
	wa := NewWitnessAggregator()
	addr := types.Address{0xCC}

	// Block 1: account is created.
	bw1 := makeTestWitness(1, []types.Address{addr}, 0)
	bw1.PreState[addr].Exists = false // did not exist before

	// Block 2: account is modified.
	bw2 := makeTestWitness(2, []types.Address{addr}, 1)
	bw2.PreState[addr].Exists = true

	wa.AddBlockWitness(1, bw1)
	wa.AddBlockWitness(2, bw2)

	wr, _ := wa.BuildRangeWitness(1, 2)
	// Merged pre-state should show the account as not-existing (earliest state).
	if wr.MergedPreState[addr].Exists {
		t.Error("expected merged pre-state to show account as not-existing")
	}
}

func TestWitnessAggregatorAccountDeletedThenRecreated(t *testing.T) {
	wa := NewWitnessAggregator()
	addr := types.Address{0xDD}

	// Block 1: account exists.
	bw1 := makeTestWitness(1, []types.Address{addr}, 1)
	// Block 2: account doesn't exist (was deleted).
	bw2 := makeTestWitness(2, []types.Address{}, 0)
	// Block 3: account recreated.
	bw3 := makeTestWitness(3, []types.Address{addr}, 2)

	wa.AddBlockWitness(1, bw1)
	wa.AddBlockWitness(2, bw2)
	wa.AddBlockWitness(3, bw3)

	wr, err := wa.BuildRangeWitness(1, 3)
	if err != nil {
		t.Fatalf("build range: %v", err)
	}
	// Account should be in pre-state from block 1.
	if _, ok := wr.MergedPreState[addr]; !ok {
		t.Error("expected address in merged pre-state")
	}
}

func TestWitnessAggregatorStorageConflictResolution(t *testing.T) {
	wa := NewWitnessAggregator()
	addr := types.Address{0xEE}

	bw1 := makeTestWitness(1, []types.Address{addr}, 1)
	bw2 := makeTestWitness(2, []types.Address{addr}, 1)

	var diffAddr [20]byte
	copy(diffAddr[:], addr[:])

	// Block 1: key 0 goes from 0x01 to 0x02.
	bw1.StateDiffs = append(bw1.StateDiffs, StateDiff{
		Address: diffAddr,
		StorageChanges: []StorageChange{
			{Key: [32]byte{0x00}, OldValue: [32]byte{0x01}, NewValue: [32]byte{0x02}},
		},
	})
	// Block 2: key 0 goes from 0x02 to 0x03 (latest wins).
	bw2.StateDiffs = append(bw2.StateDiffs, StateDiff{
		Address: diffAddr,
		StorageChanges: []StorageChange{
			{Key: [32]byte{0x00}, OldValue: [32]byte{0x02}, NewValue: [32]byte{0x03}},
		},
	})

	wa.AddBlockWitness(1, bw1)
	wa.AddBlockWitness(2, bw2)

	wr, _ := wa.BuildRangeWitness(1, 2)
	mw := wa.Minimize(wr)

	finalVal := mw.FinalStorage[addr][types.Hash{0x00}]
	if finalVal != (types.Hash{0x03}) {
		t.Errorf("expected final value 0x03, got %x", finalVal)
	}
}

func TestWitnessAggregatorConcurrentAddition(t *testing.T) {
	wa := NewWitnessAggregator()
	addr := types.Address{0x01}

	var wg sync.WaitGroup
	for i := uint64(1); i <= 50; i++ {
		wg.Add(1)
		go func(num uint64) {
			defer wg.Done()
			wa.AddBlockWitness(num, makeTestWitness(num, []types.Address{addr}, 1))
		}(i)
	}
	wg.Wait()

	if wa.BlockCount() != 50 {
		t.Errorf("expected 50 blocks, got %d", wa.BlockCount())
	}

	wr, err := wa.BuildRangeWitness(1, 50)
	if err != nil {
		t.Fatalf("build range: %v", err)
	}
	if len(wr.Witnesses) != 50 {
		t.Errorf("expected 50 witnesses, got %d", len(wr.Witnesses))
	}
}

func TestWitnessAggregatorLargeRange(t *testing.T) {
	wa := NewWitnessAggregator()
	addr := types.Address{0x01}

	for i := uint64(1); i <= 100; i++ {
		wa.AddBlockWitness(i, makeTestWitness(i, []types.Address{addr}, 1))
	}

	wr, err := wa.BuildRangeWitness(1, 100)
	if err != nil {
		t.Fatalf("build range: %v", err)
	}
	if len(wr.Witnesses) != 100 {
		t.Errorf("expected 100 witnesses, got %d", len(wr.Witnesses))
	}

	mw := wa.Minimize(wr)
	if len(mw.Accounts) != 1 {
		t.Errorf("expected 1 deduplicated account, got %d", len(mw.Accounts))
	}
}

func TestWitnessAggregatorNilWitness(t *testing.T) {
	wa := NewWitnessAggregator()
	err := wa.AddBlockWitness(1, nil)
	if err != ErrAggregatorNilBlock {
		t.Errorf("expected ErrAggregatorNilBlock, got %v", err)
	}
}

func TestWitnessAggregatorInvalidRange(t *testing.T) {
	wa := NewWitnessAggregator()
	wa.AddBlockWitness(1, makeTestWitness(1, nil, 0))

	_, err := wa.BuildRangeWitness(5, 1)
	if err != ErrAggregatorRange {
		t.Errorf("expected ErrAggregatorRange, got %v", err)
	}
}

func TestWitnessRangeSortedAddresses(t *testing.T) {
	wa := NewWitnessAggregator()
	addr1 := types.Address{0x03}
	addr2 := types.Address{0x01}
	addr3 := types.Address{0x02}

	wa.AddBlockWitness(1, makeTestWitness(1, []types.Address{addr1, addr2, addr3}, 0))

	wr, _ := wa.BuildRangeWitness(1, 1)
	sorted := wr.SortedAddresses()

	if len(sorted) != 3 {
		t.Fatalf("expected 3 addresses, got %d", len(sorted))
	}
	if sorted[0] != addr2 || sorted[1] != addr3 || sorted[2] != addr1 {
		t.Errorf("addresses not properly sorted: %v", sorted)
	}
}

func TestWitnessAggregatorCodeMerge(t *testing.T) {
	wa := NewWitnessAggregator()
	addr := types.Address{0x01}

	bw1 := makeTestWitness(1, []types.Address{addr}, 0)
	codeHash1 := types.Hash{0xAA}
	bw1.Codes[codeHash1] = []byte{0x60, 0x00}

	bw2 := makeTestWitness(2, []types.Address{addr}, 0)
	codeHash2 := types.Hash{0xBB}
	bw2.Codes[codeHash2] = []byte{0x60, 0x01}

	wa.AddBlockWitness(1, bw1)
	wa.AddBlockWitness(2, bw2)

	wr, _ := wa.BuildRangeWitness(1, 2)
	if len(wr.MergedCodes) != 2 {
		t.Errorf("expected 2 merged codes, got %d", len(wr.MergedCodes))
	}
}
