package vm

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// mockBALTracker records BAL events for testing.
type mockBALTracker struct {
	storageReads   []storageReadEvent
	storageChanges []storageChangeEvent
	balanceChanges []balanceChangeEvent
	addressTouches []types.Address
}

type storageReadEvent struct {
	Addr  types.Address
	Slot  types.Hash
	Value types.Hash
}

type storageChangeEvent struct {
	Addr   types.Address
	Slot   types.Hash
	OldVal types.Hash
	NewVal types.Hash
}

type balanceChangeEvent struct {
	Addr   types.Address
	OldBal *big.Int
	NewBal *big.Int
}

func (m *mockBALTracker) RecordStorageRead(addr types.Address, slot, value types.Hash) {
	m.storageReads = append(m.storageReads, storageReadEvent{addr, slot, value})
}

func (m *mockBALTracker) RecordStorageChange(addr types.Address, slot, oldVal, newVal types.Hash) {
	m.storageChanges = append(m.storageChanges, storageChangeEvent{addr, slot, oldVal, newVal})
}

func (m *mockBALTracker) RecordBalanceChange(addr types.Address, oldBal, newBal *big.Int) {
	m.balanceChanges = append(m.balanceChanges, balanceChangeEvent{addr, new(big.Int).Set(oldBal), new(big.Int).Set(newBal)})
}

func (m *mockBALTracker) RecordAddressTouch(addr types.Address) {
	m.addressTouches = append(m.addressTouches, addr)
}

func newTestEVMWithBAL() (*EVM, *mockBALTracker) {
	statedb := NewMockStateDB()
	evm := NewEVMWithState(
		BlockContext{BlockNumber: big.NewInt(1)},
		TxContext{Origin: types.BytesToAddress([]byte{0x01})},
		Config{},
		statedb,
	)
	tracker := &mockBALTracker{}
	evm.SetBALTracker(tracker, 1)
	return evm, tracker
}

func TestBALHook_SloadEmitsStorageRead(t *testing.T) {
	evm, tracker := newTestEVMWithBAL()
	addr := types.BytesToAddress([]byte{0x42})
	slot := types.BytesToHash([]byte{0x01})
	val := types.BytesToHash([]byte{0xaa})
	evm.StateDB.SetState(addr, slot, val)

	contract := &Contract{Address: addr, Gas: 100000}
	stack := NewStack()
	stack.Push(new(big.Int).SetBytes(slot[:]))

	_, err := opSload(new(uint64), evm, contract, NewMemory(), stack)
	if err != nil {
		t.Fatalf("opSload: %v", err)
	}

	if len(tracker.storageReads) != 1 {
		t.Fatalf("expected 1 storage read, got %d", len(tracker.storageReads))
	}
	if tracker.storageReads[0].Addr != addr {
		t.Errorf("addr = %v, want %v", tracker.storageReads[0].Addr, addr)
	}
	if tracker.storageReads[0].Slot != slot {
		t.Errorf("slot = %v, want %v", tracker.storageReads[0].Slot, slot)
	}
	if tracker.storageReads[0].Value != val {
		t.Errorf("value = %v, want %v", tracker.storageReads[0].Value, val)
	}
}

func TestBALHook_SstoreEmitsStorageChange(t *testing.T) {
	evm, tracker := newTestEVMWithBAL()
	addr := types.BytesToAddress([]byte{0x42})
	slot := types.BytesToHash([]byte{0x01})
	oldVal := types.BytesToHash([]byte{0xaa})
	newVal := types.BytesToHash([]byte{0xbb})
	evm.StateDB.SetState(addr, slot, oldVal)

	contract := &Contract{Address: addr, Gas: 100000}
	stack := NewStack()
	stack.Push(new(big.Int).SetBytes(newVal[:]))
	stack.Push(new(big.Int).SetBytes(slot[:]))

	_, err := opSstore(new(uint64), evm, contract, NewMemory(), stack)
	if err != nil {
		t.Fatalf("opSstore: %v", err)
	}

	if len(tracker.storageChanges) != 1 {
		t.Fatalf("expected 1 storage change, got %d", len(tracker.storageChanges))
	}
	if tracker.storageChanges[0].OldVal != oldVal {
		t.Errorf("oldVal = %v, want %v", tracker.storageChanges[0].OldVal, oldVal)
	}
	if tracker.storageChanges[0].NewVal != newVal {
		t.Errorf("newVal = %v, want %v", tracker.storageChanges[0].NewVal, newVal)
	}
}

func TestBALHook_SstoreNoopEmitsStorageRead(t *testing.T) {
	evm, tracker := newTestEVMWithBAL()
	addr := types.BytesToAddress([]byte{0x42})
	slot := types.BytesToHash([]byte{0x01})
	val := types.BytesToHash([]byte{0xaa})
	evm.StateDB.SetState(addr, slot, val)

	contract := &Contract{Address: addr, Gas: 100000}
	stack := NewStack()
	stack.Push(new(big.Int).SetBytes(val[:])) // same value
	stack.Push(new(big.Int).SetBytes(slot[:]))

	_, err := opSstore(new(uint64), evm, contract, NewMemory(), stack)
	if err != nil {
		t.Fatalf("opSstore: %v", err)
	}

	// No-op write should emit StorageRead, not StorageChange.
	if len(tracker.storageChanges) != 0 {
		t.Errorf("expected 0 storage changes for no-op SSTORE, got %d", len(tracker.storageChanges))
	}
	if len(tracker.storageReads) != 1 {
		t.Fatalf("expected 1 storage read for no-op SSTORE, got %d", len(tracker.storageReads))
	}
	if tracker.storageReads[0].Slot != slot {
		t.Errorf("slot = %v, want %v", tracker.storageReads[0].Slot, slot)
	}
}

func TestBALHook_SstoreGlamstNoopEmitsStorageRead(t *testing.T) {
	evm, tracker := newTestEVMWithBAL()
	addr := types.BytesToAddress([]byte{0x42})
	slot := types.BytesToHash([]byte{0x01})
	val := types.BytesToHash([]byte{0xaa})
	evm.StateDB.SetState(addr, slot, val)

	contract := &Contract{Address: addr, Gas: 100000}
	stack := NewStack()
	stack.Push(new(big.Int).SetBytes(val[:])) // same value
	stack.Push(new(big.Int).SetBytes(slot[:]))

	_, err := opSstoreGlamst(new(uint64), evm, contract, NewMemory(), stack)
	if err != nil {
		t.Fatalf("opSstoreGlamst: %v", err)
	}

	if len(tracker.storageChanges) != 0 {
		t.Errorf("expected 0 storage changes for no-op, got %d", len(tracker.storageChanges))
	}
	if len(tracker.storageReads) != 1 {
		t.Fatalf("expected 1 storage read for no-op, got %d", len(tracker.storageReads))
	}
}

func TestBALHook_BalanceEmitsAddressTouch(t *testing.T) {
	evm, tracker := newTestEVMWithBAL()
	addr := types.BytesToAddress([]byte{0x42})
	evm.StateDB.AddBalance(addr, big.NewInt(1000))

	contract := &Contract{Address: types.BytesToAddress([]byte{0x01}), Gas: 100000}
	stack := NewStack()
	stack.Push(new(big.Int).SetBytes(addr[:]))

	_, err := opBalance(new(uint64), evm, contract, NewMemory(), stack)
	if err != nil {
		t.Fatalf("opBalance: %v", err)
	}

	if len(tracker.addressTouches) != 1 {
		t.Fatalf("expected 1 address touch, got %d", len(tracker.addressTouches))
	}
	if tracker.addressTouches[0] != addr {
		t.Errorf("touched addr = %v, want %v", tracker.addressTouches[0], addr)
	}
}

func TestBALHook_SelfBalanceEmitsAddressTouch(t *testing.T) {
	evm, tracker := newTestEVMWithBAL()
	addr := types.BytesToAddress([]byte{0x42})
	evm.StateDB.AddBalance(addr, big.NewInt(1000))

	contract := &Contract{Address: addr, Gas: 100000}
	stack := NewStack()

	_, err := opSelfBalance(new(uint64), evm, contract, NewMemory(), stack)
	if err != nil {
		t.Fatalf("opSelfBalance: %v", err)
	}

	if len(tracker.addressTouches) != 1 {
		t.Fatalf("expected 1 address touch, got %d", len(tracker.addressTouches))
	}
	if tracker.addressTouches[0] != addr {
		t.Errorf("touched addr = %v, want %v", tracker.addressTouches[0], addr)
	}
}

func TestBALHook_CallEmitsAddressTouch(t *testing.T) {
	evm, tracker := newTestEVMWithBAL()
	caller := types.BytesToAddress([]byte{0x01})
	target := types.BytesToAddress([]byte{0x42})
	evm.StateDB.AddBalance(caller, big.NewInt(1000000))
	evm.StateDB.CreateAccount(target)

	contract := &Contract{Address: caller, CallerAddress: caller, Gas: 100000}
	evm.callGasTemp = 50000

	stack := NewStack()
	stack.Push(new(big.Int))                     // retLength
	stack.Push(new(big.Int))                     // retOffset
	stack.Push(new(big.Int))                     // argsLength
	stack.Push(new(big.Int))                     // argsOffset
	stack.Push(new(big.Int))                     // value = 0
	stack.Push(new(big.Int).SetBytes(target[:])) // addr
	stack.Push(new(big.Int).SetUint64(50000))    // gas

	_, err := opCall(new(uint64), evm, contract, NewMemory(), stack)
	if err != nil {
		t.Fatalf("opCall: %v", err)
	}

	found := false
	for _, a := range tracker.addressTouches {
		if a == target {
			found = true
			break
		}
	}
	if !found {
		t.Error("CALL should emit RecordAddressTouch for the target address")
	}
}

func TestBALHook_NilTrackerNoPanic(t *testing.T) {
	// EVM with no BAL tracker should not panic.
	statedb := NewMockStateDB()
	evm := NewEVMWithState(
		BlockContext{BlockNumber: big.NewInt(1)},
		TxContext{},
		Config{},
		statedb,
	)
	addr := types.BytesToAddress([]byte{0x42})
	slot := types.BytesToHash([]byte{0x01})
	evm.StateDB.SetState(addr, slot, types.BytesToHash([]byte{0xaa}))

	contract := &Contract{Address: addr, Gas: 100000}

	// SLOAD with nil tracker
	stack := NewStack()
	stack.Push(new(big.Int).SetBytes(slot[:]))
	if _, err := opSload(new(uint64), evm, contract, NewMemory(), stack); err != nil {
		t.Fatalf("opSload with nil tracker: %v", err)
	}

	// SSTORE with nil tracker
	stack = NewStack()
	newHash := types.BytesToHash([]byte{0xbb})
	stack.Push(new(big.Int).SetBytes(newHash[:]))
	stack.Push(new(big.Int).SetBytes(slot[:]))
	if _, err := opSstore(new(uint64), evm, contract, NewMemory(), stack); err != nil {
		t.Fatalf("opSstore with nil tracker: %v", err)
	}

	// BALANCE with nil tracker
	stack = NewStack()
	stack.Push(new(big.Int).SetBytes(addr[:]))
	if _, err := opBalance(new(uint64), evm, contract, NewMemory(), stack); err != nil {
		t.Fatalf("opBalance with nil tracker: %v", err)
	}
}

func TestBALHook_CallValueTransferRecordsBalanceChange(t *testing.T) {
	statedb := NewMockStateDB()
	evm := NewEVMWithState(
		BlockContext{BlockNumber: big.NewInt(1)},
		TxContext{},
		Config{},
		statedb,
	)
	tracker := &mockBALTracker{}
	evm.SetBALTracker(tracker, 1)

	caller := types.BytesToAddress([]byte{0x01})
	target := types.BytesToAddress([]byte{0x42})
	statedb.AddBalance(caller, big.NewInt(1000))
	statedb.CreateAccount(target)

	value := big.NewInt(100)
	_, _, err := evm.Call(caller, target, nil, 100000, value)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}

	if len(tracker.balanceChanges) < 2 {
		t.Fatalf("expected at least 2 balance changes, got %d", len(tracker.balanceChanges))
	}

	// Verify caller balance decreased.
	var callerChange *balanceChangeEvent
	for i := range tracker.balanceChanges {
		if tracker.balanceChanges[i].Addr == caller {
			callerChange = &tracker.balanceChanges[i]
			break
		}
	}
	if callerChange == nil {
		t.Fatal("expected balance change for caller")
	}
	if callerChange.OldBal.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("caller old balance = %v, want 1000", callerChange.OldBal)
	}
	if callerChange.NewBal.Cmp(big.NewInt(900)) != 0 {
		t.Errorf("caller new balance = %v, want 900", callerChange.NewBal)
	}
}

func TestBALHook_ExtcodehashEmitsAddressTouch(t *testing.T) {
	evm, tracker := newTestEVMWithBAL()
	addr := types.BytesToAddress([]byte{0x42})
	evm.StateDB.CreateAccount(addr)
	evm.StateDB.SetCode(addr, []byte{0x60, 0x00})

	contract := &Contract{Address: types.BytesToAddress([]byte{0x01}), Gas: 100000}
	stack := NewStack()
	stack.Push(new(big.Int).SetBytes(addr[:]))

	_, err := opExtcodehash(new(uint64), evm, contract, NewMemory(), stack)
	if err != nil {
		t.Fatalf("opExtcodehash: %v", err)
	}

	if len(tracker.addressTouches) != 1 {
		t.Fatalf("expected 1 address touch, got %d", len(tracker.addressTouches))
	}
	if tracker.addressTouches[0] != addr {
		t.Errorf("touched addr = %v, want %v", tracker.addressTouches[0], addr)
	}
}
