package state

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/rawdb"
	"github.com/eth2030/eth2030/core/types"
)

// newTestTrieStateDB creates a TrieStateDB backed by an in-memory DB.
func newTestTrieStateDB(t *testing.T) *TrieStateDB {
	t.Helper()
	db := rawdb.NewMemoryDB()
	return NewTrieStateDB(db)
}

// TestTrieStateDB_BasicAccountOps tests CreateAccount, GetBalance, AddBalance,
// SubBalance, GetNonce, SetNonce.
func TestTrieStateDB_BasicAccountOps(t *testing.T) {
	s := newTestTrieStateDB(t)
	addr := types.Address{0x01}

	s.CreateAccount(addr)
	if bal := s.GetBalance(addr); bal.Sign() != 0 {
		t.Fatalf("expected zero balance, got %v", bal)
	}

	s.AddBalance(addr, big.NewInt(1000))
	if bal := s.GetBalance(addr); bal.Cmp(big.NewInt(1000)) != 0 {
		t.Fatalf("expected 1000, got %v", bal)
	}

	s.SubBalance(addr, big.NewInt(300))
	if bal := s.GetBalance(addr); bal.Cmp(big.NewInt(700)) != 0 {
		t.Fatalf("expected 700, got %v", bal)
	}

	s.SetNonce(addr, 7)
	if n := s.GetNonce(addr); n != 7 {
		t.Fatalf("expected nonce 7, got %d", n)
	}
}

// TestTrieStateDB_Code tests SetCode and GetCode.
func TestTrieStateDB_Code(t *testing.T) {
	s := newTestTrieStateDB(t)
	addr := types.Address{0x02}
	code := []byte{0x60, 0x01, 0x60, 0x00}

	s.CreateAccount(addr)
	s.SetCode(addr, code)

	if got := s.GetCode(addr); string(got) != string(code) {
		t.Fatalf("expected code %x, got %x", code, got)
	}
	if sz := s.GetCodeSize(addr); sz != len(code) {
		t.Fatalf("expected code size %d, got %d", len(code), sz)
	}
	codeHash := s.GetCodeHash(addr)
	if codeHash == (types.Hash{}) {
		t.Fatal("code hash should not be empty after SetCode")
	}
}

// TestTrieStateDB_Storage tests SetState, GetState, GetCommittedState.
func TestTrieStateDB_Storage(t *testing.T) {
	s := newTestTrieStateDB(t)
	addr := types.Address{0x03}
	slot := types.Hash{0xAB}
	val := types.Hash{0xCD}

	s.CreateAccount(addr)
	s.SetState(addr, slot, val)

	if got := s.GetState(addr, slot); got != val {
		t.Fatalf("expected %v, got %v", val, got)
	}

	// Before commit, committed state is zero.
	if got := s.GetCommittedState(addr, slot); got != (types.Hash{}) {
		t.Fatalf("committed state should be zero before commit, got %v", got)
	}
}

// TestTrieStateDB_SelfDestruct tests SelfDestruct and HasSelfDestructed.
func TestTrieStateDB_SelfDestruct(t *testing.T) {
	s := newTestTrieStateDB(t)
	addr := types.Address{0x04}

	s.CreateAccount(addr)
	s.AddBalance(addr, big.NewInt(500))

	if s.HasSelfDestructed(addr) {
		t.Fatal("account should not be self-destructed yet")
	}

	s.SelfDestruct(addr)
	if !s.HasSelfDestructed(addr) {
		t.Fatal("account should be self-destructed")
	}
	// Balance zeroed on self-destruct.
	if bal := s.GetBalance(addr); bal.Sign() != 0 {
		t.Fatalf("balance should be 0 after self-destruct, got %v", bal)
	}
}

// TestTrieStateDB_SnapshotRevert tests Snapshot/RevertToSnapshot.
func TestTrieStateDB_SnapshotRevert(t *testing.T) {
	s := newTestTrieStateDB(t)
	addr := types.Address{0x05}

	s.CreateAccount(addr)
	s.AddBalance(addr, big.NewInt(100))

	snap := s.Snapshot()
	s.AddBalance(addr, big.NewInt(50))
	if bal := s.GetBalance(addr); bal.Cmp(big.NewInt(150)) != 0 {
		t.Fatalf("expected 150, got %v", bal)
	}

	s.RevertToSnapshot(snap)
	if bal := s.GetBalance(addr); bal.Cmp(big.NewInt(100)) != 0 {
		t.Fatalf("expected 100 after revert, got %v", bal)
	}
}

// TestTrieStateDB_CommitAndReload tests that Commit persists state to the DB
// and a fresh TrieStateDB reading from the same DB sees the committed state.
func TestTrieStateDB_CommitAndReload(t *testing.T) {
	db := rawdb.NewMemoryDB()
	s := NewTrieStateDB(db)

	addr1 := types.Address{0x10}
	addr2 := types.Address{0x11}
	slot := types.Hash{0x01}
	val := types.Hash{0x42}
	code := []byte{0x60, 0x00}

	s.CreateAccount(addr1)
	s.AddBalance(addr1, big.NewInt(9999))
	s.SetNonce(addr1, 3)

	s.CreateAccount(addr2)
	s.SetCode(addr2, code)
	s.SetState(addr2, slot, val)

	root1, err := s.Commit()
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if root1 == (types.Hash{}) {
		t.Fatal("root should not be zero after commit")
	}

	// After Commit, dirty layer is empty. A second TrieStateDB on same DB should see persisted state.
	s2 := NewTrieStateDB(db)

	if bal := s2.GetBalance(addr1); bal.Cmp(big.NewInt(9999)) != 0 {
		t.Fatalf("expected 9999 after reload, got %v", bal)
	}
	if n := s2.GetNonce(addr1); n != 3 {
		t.Fatalf("expected nonce 3 after reload, got %d", n)
	}
	if got := s2.GetCode(addr2); string(got) != string(code) {
		t.Fatalf("code mismatch after reload: got %x", got)
	}
	if got := s2.GetState(addr2, slot); got != val {
		t.Fatalf("storage mismatch after reload: got %v", got)
	}
}

// TestTrieStateDB_GetRoot tests that GetRoot is deterministic and matches
// after a full commit.
func TestTrieStateDB_GetRoot(t *testing.T) {
	s := newTestTrieStateDB(t)
	addr := types.Address{0x20}
	s.CreateAccount(addr)
	s.AddBalance(addr, big.NewInt(1))

	root1 := s.GetRoot()
	root2 := s.GetRoot()
	if root1 != root2 {
		t.Fatal("GetRoot should be deterministic")
	}
	if root1 == (types.Hash{}) {
		t.Fatal("root should not be zero")
	}

	root3, err := s.Commit()
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	// Root returned by Commit equals the root computed before commit.
	if root3 != root1 {
		t.Fatalf("Commit root %v != pre-commit root %v", root3, root1)
	}
}

// TestTrieStateDB_RootAfterCommitAndNewBlock tests that GetRoot after a commit
// and subsequent writes (simulating the next block) reflects the new state.
func TestTrieStateDB_RootAfterCommitAndNewBlock(t *testing.T) {
	db := rawdb.NewMemoryDB()
	s := NewTrieStateDB(db)

	addr := types.Address{0x30}
	s.CreateAccount(addr)
	s.AddBalance(addr, big.NewInt(500))

	root1, err := s.Commit()
	if err != nil {
		t.Fatalf("block1 Commit: %v", err)
	}

	// Simulate block 2: add more balance to the same account.
	s.AddBalance(addr, big.NewInt(100)) // loadFromDB should load the committed account
	root2, err := s.Commit()
	if err != nil {
		t.Fatalf("block2 Commit: %v", err)
	}

	if root2 == root1 {
		t.Fatal("root should change when state changes")
	}

	// Verify state from a fresh reader.
	s3 := NewTrieStateDB(db)
	if bal := s3.GetBalance(addr); bal.Cmp(big.NewInt(600)) != 0 {
		t.Fatalf("expected balance 600 after two blocks, got %v", bal)
	}
}

// TestTrieStateDB_Dup tests that Dup produces an independent copy.
func TestTrieStateDB_Dup(t *testing.T) {
	s := newTestTrieStateDB(t)
	addr := types.Address{0x40}
	s.CreateAccount(addr)
	s.AddBalance(addr, big.NewInt(200))

	dup := s.Dup().(*TrieStateDB)

	// Mutate original; copy should be unaffected.
	s.AddBalance(addr, big.NewInt(800))

	if bal := dup.GetBalance(addr); bal.Cmp(big.NewInt(200)) != 0 {
		t.Fatalf("dup balance should be 200, got %v", bal)
	}
	if bal := s.GetBalance(addr); bal.Cmp(big.NewInt(1000)) != 0 {
		t.Fatalf("original balance should be 1000, got %v", bal)
	}
}

// TestTrieStateDB_DupAfterCommit tests Dup on a committed TrieStateDB is cheap
// (empty dirty layer) and reads from DB correctly.
func TestTrieStateDB_DupAfterCommit(t *testing.T) {
	db := rawdb.NewMemoryDB()
	s := NewTrieStateDB(db)

	addr := types.Address{0x50}
	s.CreateAccount(addr)
	s.AddBalance(addr, big.NewInt(333))

	if _, err := s.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	// After commit, dirty layer is empty.
	dup := s.Dup().(*TrieStateDB)
	if len(dup.mem.stateObjects) != 0 {
		t.Fatalf("dup dirty layer should be empty after commit, got %d entries", len(dup.mem.stateObjects))
	}

	// Dup should still read from DB correctly.
	if bal := dup.GetBalance(addr); bal.Cmp(big.NewInt(333)) != 0 {
		t.Fatalf("expected 333 from dup, got %v", bal)
	}
}

// TestTrieStateDB_SelfDestructAndCommit tests that self-destructed accounts
// are removed from the DB after Commit.
func TestTrieStateDB_SelfDestructAndCommit(t *testing.T) {
	db := rawdb.NewMemoryDB()
	s := NewTrieStateDB(db)

	addr := types.Address{0x60}
	s.CreateAccount(addr)
	s.AddBalance(addr, big.NewInt(100))
	if _, err := s.Commit(); err != nil {
		t.Fatalf("first Commit: %v", err)
	}

	// Self-destruct in next block.
	s.SelfDestruct(addr)
	if _, err := s.Commit(); err != nil {
		t.Fatalf("second Commit: %v", err)
	}

	// Fresh reader should not find the account.
	s2 := NewTrieStateDB(db)
	if s2.Exist(addr) {
		t.Fatal("self-destructed account should not exist in DB after commit")
	}
}

// TestTrieStateDB_Logs tests AddLog and GetLogs.
func TestTrieStateDB_Logs(t *testing.T) {
	s := newTestTrieStateDB(t)
	txHash := types.Hash{0xAA}
	s.SetTxContext(txHash, 0)
	log1 := &types.Log{Address: types.Address{0x01}}
	s.AddLog(log1)

	logs := s.GetLogs(txHash)
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
}

// TestTrieStateDB_RefundCounter tests AddRefund, SubRefund, GetRefund.
func TestTrieStateDB_RefundCounter(t *testing.T) {
	s := newTestTrieStateDB(t)
	s.AddRefund(100)
	if r := s.GetRefund(); r != 100 {
		t.Fatalf("expected 100, got %d", r)
	}
	s.SubRefund(40)
	if r := s.GetRefund(); r != 60 {
		t.Fatalf("expected 60, got %d", r)
	}
}

// TestTrieStateDB_AccessList tests AddAddressToAccessList, AddressInAccessList,
// AddSlotToAccessList, SlotInAccessList.
func TestTrieStateDB_AccessList(t *testing.T) {
	s := newTestTrieStateDB(t)
	addr := types.Address{0x70}
	slot := types.Hash{0x01}

	if s.AddressInAccessList(addr) {
		t.Fatal("address should not be in access list")
	}

	s.AddAddressToAccessList(addr)
	if !s.AddressInAccessList(addr) {
		t.Fatal("address should be in access list")
	}

	addrOk, slotOk := s.SlotInAccessList(addr, slot)
	if !addrOk || slotOk {
		t.Fatal("slot should not be in access list yet")
	}

	s.AddSlotToAccessList(addr, slot)
	addrOk, slotOk = s.SlotInAccessList(addr, slot)
	if !addrOk || !slotOk {
		t.Fatal("slot should be in access list")
	}
}

// TestTrieStateDB_TransientStorage tests EIP-1153 transient storage.
func TestTrieStateDB_TransientStorage(t *testing.T) {
	s := newTestTrieStateDB(t)
	addr := types.Address{0x80}
	key := types.Hash{0x01}
	val := types.Hash{0x42}

	s.SetTransientState(addr, key, val)
	if got := s.GetTransientState(addr, key); got != val {
		t.Fatalf("expected %v, got %v", val, got)
	}

	s.ClearTransientStorage()
	if got := s.GetTransientState(addr, key); got != (types.Hash{}) {
		t.Fatalf("expected zero after clear, got %v", got)
	}
}

// TestTrieStateDB_InterfaceCompliance verifies that TrieStateDB satisfies StateDB at compile time.
var _ StateDB = (*TrieStateDB)(nil)

// TestTrieStateDB_NewFromMemory tests NewTrieStateDBFromMemory.
func TestTrieStateDB_NewFromMemory(t *testing.T) {
	db := rawdb.NewMemoryDB()
	mem := NewMemoryStateDB()

	addr := types.Address{0x90}
	mem.CreateAccount(addr)
	mem.AddBalance(addr, big.NewInt(42))

	ts := NewTrieStateDBFromMemory(db, mem)
	if _, err := ts.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	// Fresh TrieStateDB on same DB should see the committed state.
	ts2 := NewTrieStateDB(db)
	if bal := ts2.GetBalance(addr); bal.Cmp(big.NewInt(42)) != 0 {
		t.Fatalf("expected 42 after commit from memory, got %v", bal)
	}
}
