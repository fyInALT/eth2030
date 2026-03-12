// Package state provides state management for Ethereum accounts and storage.
package state

import (
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/core/rawdb"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/rlp"
	"github.com/eth2030/eth2030/trie"
)

// DB key prefixes — chosen to avoid collision with block/receipt/tx rawdb keys.
var (
	dbPrefixAccount = []byte("sa") // "sa" + addr[20]   → RLP-encoded account
	dbPrefixStorage = []byte("ss") // "ss" + addr[20] + slot[32] → val[32]
	dbPrefixCode    = []byte("sc") // "sc" + codeHash[32] → code bytes
)

func accountDBKey(addr types.Address) []byte {
	key := make([]byte, 2+20)
	copy(key, dbPrefixAccount)
	copy(key[2:], addr[:])
	return key
}

func storageDBKey(addr types.Address, slot types.Hash) []byte {
	key := make([]byte, 2+20+32)
	copy(key, dbPrefixStorage)
	copy(key[2:], addr[:])
	copy(key[22:], slot[:])
	return key
}

func storageDBPrefix(addr types.Address) []byte {
	key := make([]byte, 2+20)
	copy(key, dbPrefixStorage)
	copy(key[2:], addr[:])
	return key
}

func codeDBKey(codeHash []byte) []byte {
	key := make([]byte, 2+32)
	copy(key, dbPrefixCode)
	copy(key[2:], codeHash)
	return key
}

// GCModeArchive and GCModeFull are the two supported garbage-collection modes.
// GCModeArchive (default) retains all historical state — no state keys are
// ever deleted beyond explicit zeroing or self-destruct. GCModeFull will in
// future prune overwritten state keys after each Commit to save disk space.
const (
	GCModeArchive = "archive"
	GCModeFull    = "full"
)

// TrieStateDB is a disk-backed StateDB implementation. It keeps a
// MemoryStateDB as a dirty write buffer for the current block and persists
// committed state to a rawdb.Database.
//
// Memory model: after each Commit() the dirty buffer is reset. Subsequent
// reads populate the buffer on demand from the DB. Memory usage is therefore
// proportional to the working set of a single block rather than growing with
// chain history.
type TrieStateDB struct {
	mem       *MemoryStateDB             // dirty write buffer for the current block
	db        rawdb.Database             // persistent backing store
	gcMode    string                     // GCModeArchive or GCModeFull
	recreated map[types.Address]struct{} // accounts reset by CreateAccount this block

	// frozenAccounts is a read-only snapshot of every account RLP in the DB
	// as it was at the end of the most recent Commit(). buildStateTrie() uses
	// this map instead of iterating the live DB so that concurrent commits from
	// other goroutines (e.g. a newPayload running in parallel with a block
	// builder) cannot corrupt the state-root computation.
	// The map is shared across Dup() copies; callers must not mutate it.
	frozenAccounts map[types.Address][]byte
}

// GetMem returns a reference to the internal MemoryStateDB dirty buffer.
// Callers must not modify the returned value; use it only for snapshotting
// the in-memory state (e.g. at genesis, before any Commit has been called).
func (t *TrieStateDB) GetMem() *MemoryStateDB {
	return t.mem
}

// GCMode returns the garbage-collection mode string ("archive" or "full").
func (t *TrieStateDB) GCMode() string {
	return t.gcMode
}

// DB returns the underlying database backing this TrieStateDB.
func (t *TrieStateDB) DB() rawdb.Database {
	return t.db
}

// NewTrieStateDB creates a TrieStateDB backed by db using archive GC mode.
func NewTrieStateDB(db rawdb.Database) *TrieStateDB {
	return NewTrieStateDBWithGCMode(db, GCModeArchive)
}

// NewTrieStateDBWithGCMode creates a TrieStateDB with the specified GC mode.
// Use GCModeArchive to retain all history or GCModeFull to prune overwritten
// state after each Commit (pruning is a no-op in the current implementation;
// the field is plumbed for future use).
func NewTrieStateDBWithGCMode(db rawdb.Database, gcMode string) *TrieStateDB {
	if gcMode != GCModeFull {
		gcMode = GCModeArchive
	}
	return &TrieStateDB{
		mem:    NewMemoryStateDB(),
		db:     db,
		gcMode: gcMode,
	}
}

// NewTrieStateDBFromMemory creates a TrieStateDB that adopts an existing
// MemoryStateDB as its initial dirty layer. Call Commit() immediately after
// to persist the in-memory state to the DB and free the dirty layer. This
// is typically used to convert a genesis MemoryStateDB into a TrieStateDB.
func NewTrieStateDBFromMemory(db rawdb.Database, mem *MemoryStateDB) *TrieStateDB {
	return NewTrieStateDBFromMemoryWithGCMode(db, mem, GCModeArchive)
}

// NewTrieStateDBFromMemoryWithGCMode is like NewTrieStateDBFromMemory but
// accepts an explicit GC mode.
func NewTrieStateDBFromMemoryWithGCMode(db rawdb.Database, mem *MemoryStateDB, gcMode string) *TrieStateDB {
	if gcMode != GCModeFull {
		gcMode = GCModeArchive
	}
	return &TrieStateDB{
		mem:    mem,
		db:     db,
		gcMode: gcMode,
	}
}

// loadFromDB loads account state from the persistent store into the dirty
// buffer. It is a no-op if the address is already buffered.
func (t *TrieStateDB) loadFromDB(addr types.Address) {
	if t.mem.stateObjects[addr] != nil {
		return
	}
	data, err := t.db.Get(accountDBKey(addr))
	if err != nil {
		return // not in DB → account does not exist
	}
	var acc rlpAccount
	if err := rlp.DecodeBytes(data, &acc); err != nil {
		return
	}
	if acc.Balance == nil {
		acc.Balance = new(big.Int)
	}
	obj := &stateObject{
		account: types.Account{
			Nonce:    acc.Nonce,
			Balance:  new(big.Int).Set(acc.Balance),
			CodeHash: make([]byte, len(acc.CodeHash)),
		},
		dirtyStorage:     make(map[types.Hash]types.Hash),
		committedStorage: make(map[types.Hash]types.Hash),
	}
	copy(obj.account.CodeHash, acc.CodeHash)

	// Storage slots are NOT loaded here — they are loaded lazily on demand
	// by GetState / GetCommittedState / SetState.  Eager bulk loading was the
	// root cause of O(N²) memory growth with storagespam workloads.

	// Load contract code if non-empty.
	if len(acc.CodeHash) > 0 && types.BytesToHash(acc.CodeHash) != types.EmptyCodeHash {
		if code, err := t.db.Get(codeDBKey(acc.CodeHash)); err == nil {
			obj.code = code
		}
	}

	t.mem.stateObjects[addr] = obj
}

// loadSlotFromDB loads a single storage slot from the persistent store into
// obj.committedStorage.  Returns the slot value (zero hash if not in DB).
func (t *TrieStateDB) loadSlotFromDB(addr types.Address, obj *stateObject, slot types.Hash) types.Hash {
	data, err := t.db.Get(storageDBKey(addr, slot))
	if err != nil {
		return types.Hash{} // slot not in DB
	}
	var val types.Hash
	copy(val[:], data)
	if val != (types.Hash{}) {
		obj.committedStorage[slot] = val
	}
	return val
}

// computeStorageRootFromDB computes the storage Merkle root for addr by
// merging all DB-persisted slots with the in-memory committed and dirty
// overlays (dirty takes priority over committed which takes priority over DB).
// This is used instead of computeStorageRoot when committedStorage may be
// sparse due to lazy slot loading.
func (t *TrieStateDB) computeStorageRootFromDB(addr types.Address, obj *stateObject) types.Hash {
	// Build overlay: committed (sparse cache + recently-flushed dirty) then dirty.
	overlay := make(map[types.Hash]types.Hash, len(obj.committedStorage)+len(obj.dirtyStorage))
	for slot, val := range obj.committedStorage {
		overlay[slot] = val
	}
	for slot, val := range obj.dirtyStorage {
		overlay[slot] = val
	}

	storageTrie := trie.New()
	seen := make(map[types.Hash]bool, len(overlay))

	// Walk DB slots as the base layer, applying overlay overrides.
	iter := t.db.NewIterator(storageDBPrefix(addr))
	for iter.Next() {
		key := iter.Key()
		if len(key) != 2+20+32 {
			continue
		}
		var slot types.Hash
		copy(slot[:], key[22:])
		seen[slot] = true

		val, overridden := overlay[slot]
		if !overridden {
			copy(val[:], iter.Value())
		}
		if val != (types.Hash{}) {
			hashedSlot := crypto.Keccak256(slot[:])
			trimmed := trimLeadingZeros(val[:])
			encoded, err := rlp.EncodeToBytes(trimmed)
			if err == nil {
				storageTrie.Put(hashedSlot, encoded)
			}
		}
	}
	iter.Release()

	// Add overlay slots that are not in the DB (new slots written this block).
	for slot, val := range overlay {
		if seen[slot] {
			continue
		}
		if val != (types.Hash{}) {
			hashedSlot := crypto.Keccak256(slot[:])
			trimmed := trimLeadingZeros(val[:])
			encoded, err := rlp.EncodeToBytes(trimmed)
			if err == nil {
				storageTrie.Put(hashedSlot, encoded)
			}
		}
	}

	return storageTrie.Hash()
}

// --- StateDB interface: account operations ---

func (t *TrieStateDB) CreateAccount(addr types.Address) {
	// If the address has existing DB state, mark it so Commit() can purge
	// stale storage slots that the new account does not overwrite.
	if _, err := t.db.Get(accountDBKey(addr)); err == nil {
		if t.recreated == nil {
			t.recreated = make(map[types.Address]struct{})
		}
		t.recreated[addr] = struct{}{}
	}
	t.mem.CreateAccount(addr)
}

func (t *TrieStateDB) SubBalance(addr types.Address, amount *big.Int) {
	t.loadFromDB(addr)
	t.mem.SubBalance(addr, amount)
}

func (t *TrieStateDB) AddBalance(addr types.Address, amount *big.Int) {
	t.loadFromDB(addr)
	t.mem.AddBalance(addr, amount)
}

func (t *TrieStateDB) GetBalance(addr types.Address) *big.Int {
	t.loadFromDB(addr)
	return t.mem.GetBalance(addr)
}

func (t *TrieStateDB) GetNonce(addr types.Address) uint64 {
	t.loadFromDB(addr)
	return t.mem.GetNonce(addr)
}

func (t *TrieStateDB) SetNonce(addr types.Address, nonce uint64) {
	t.loadFromDB(addr)
	t.mem.SetNonce(addr, nonce)
}

func (t *TrieStateDB) GetCode(addr types.Address) []byte {
	t.loadFromDB(addr)
	return t.mem.GetCode(addr)
}

func (t *TrieStateDB) SetCode(addr types.Address, code []byte) {
	t.loadFromDB(addr)
	t.mem.SetCode(addr, code)
}

func (t *TrieStateDB) GetCodeHash(addr types.Address) types.Hash {
	t.loadFromDB(addr)
	return t.mem.GetCodeHash(addr)
}

func (t *TrieStateDB) GetCodeSize(addr types.Address) int {
	t.loadFromDB(addr)
	return t.mem.GetCodeSize(addr)
}

// --- StateDB interface: self-destruct ---

func (t *TrieStateDB) SelfDestruct(addr types.Address) {
	t.loadFromDB(addr)
	t.mem.SelfDestruct(addr)
}

func (t *TrieStateDB) HasSelfDestructed(addr types.Address) bool {
	t.loadFromDB(addr)
	return t.mem.HasSelfDestructed(addr)
}

// --- StateDB interface: storage (lazy per-slot loading) ---

func (t *TrieStateDB) GetState(addr types.Address, key types.Hash) types.Hash {
	t.loadFromDB(addr)
	obj := t.mem.stateObjects[addr]
	if obj == nil {
		return types.Hash{}
	}
	if val, ok := obj.dirtyStorage[key]; ok {
		return val
	}
	if val, ok := obj.committedStorage[key]; ok {
		return val
	}
	return t.loadSlotFromDB(addr, obj, key)
}

func (t *TrieStateDB) GetCommittedState(addr types.Address, key types.Hash) types.Hash {
	t.loadFromDB(addr)
	obj := t.mem.stateObjects[addr]
	if obj == nil {
		return types.Hash{}
	}
	if val, ok := obj.committedStorage[key]; ok {
		return val
	}
	return t.loadSlotFromDB(addr, obj, key)
}

func (t *TrieStateDB) SetState(addr types.Address, key types.Hash, value types.Hash) {
	t.loadFromDB(addr)
	obj := t.mem.stateObjects[addr]
	if obj != nil {
		// Ensure committed value is populated before MemoryStateDB.SetState reads
		// obj.committedStorage[key] to record the journal prev value.
		if _, ok := obj.committedStorage[key]; !ok {
			if _, ok2 := obj.dirtyStorage[key]; !ok2 {
				t.loadSlotFromDB(addr, obj, key)
			}
		}
	}
	t.mem.SetState(addr, key, value)
}

// --- StateDB interface: account existence ---

func (t *TrieStateDB) Exist(addr types.Address) bool {
	t.loadFromDB(addr)
	return t.mem.Exist(addr)
}

func (t *TrieStateDB) Empty(addr types.Address) bool {
	t.loadFromDB(addr)
	return t.mem.Empty(addr)
}

// --- StateDB interface: snapshot/revert ---

func (t *TrieStateDB) Snapshot() int {
	return t.mem.Snapshot()
}

func (t *TrieStateDB) RevertToSnapshot(id int) {
	t.mem.RevertToSnapshot(id)
}

// --- StateDB interface: logs ---

func (t *TrieStateDB) AddLog(log *types.Log) {
	t.mem.AddLog(log)
}

func (t *TrieStateDB) GetLogs(txHash types.Hash) []*types.Log {
	return t.mem.GetLogs(txHash)
}

func (t *TrieStateDB) SetTxContext(txHash types.Hash, txIndex int) {
	t.mem.SetTxContext(txHash, txIndex)
}

// --- StateDB interface: refund counter ---

func (t *TrieStateDB) AddRefund(gas uint64) {
	t.mem.AddRefund(gas)
}

func (t *TrieStateDB) SubRefund(gas uint64) {
	t.mem.SubRefund(gas)
}

func (t *TrieStateDB) GetRefund() uint64 {
	return t.mem.GetRefund()
}

// --- StateDB interface: access list (EIP-2929) ---

func (t *TrieStateDB) AddAddressToAccessList(addr types.Address) {
	t.mem.AddAddressToAccessList(addr)
}

func (t *TrieStateDB) AddSlotToAccessList(addr types.Address, slot types.Hash) {
	t.mem.AddSlotToAccessList(addr, slot)
}

func (t *TrieStateDB) AddressInAccessList(addr types.Address) bool {
	return t.mem.AddressInAccessList(addr)
}

func (t *TrieStateDB) SlotInAccessList(addr types.Address, slot types.Hash) (bool, bool) {
	return t.mem.SlotInAccessList(addr, slot)
}

// --- StateDB interface: transient storage (EIP-1153) ---

func (t *TrieStateDB) GetTransientState(addr types.Address, key types.Hash) types.Hash {
	return t.mem.GetTransientState(addr, key)
}

func (t *TrieStateDB) SetTransientState(addr types.Address, key types.Hash, value types.Hash) {
	t.mem.SetTransientState(addr, key, value)
}

func (t *TrieStateDB) ClearTransientStorage() {
	t.mem.ClearTransientStorage()
}

// --- StateDB interface: root computation ---

func (t *TrieStateDB) StorageRoot(addr types.Address) types.Hash {
	t.loadFromDB(addr)
	obj := t.mem.stateObjects[addr]
	if obj == nil {
		return types.EmptyRootHash
	}
	return t.computeStorageRootFromDB(addr, obj)
}

// GetRoot computes the full state root by merging DB-persisted accounts with
// the dirty mem buffer.
func (t *TrieStateDB) GetRoot() types.Hash {
	return t.buildStateTrie().Hash()
}

// buildStateTrie constructs the full MPT over all accounts: dirty mem accounts
// override (or delete) DB accounts. DB accounts not in mem are included as-is
// using their stored RLP (which embeds the storage root from last Commit).
func (t *TrieStateDB) buildStateTrie() *trie.Trie {
	stateTrie := trie.New()

	// Build a set of dirty addresses to detect overrides.
	memAddrs := make(map[types.Address]bool, len(t.mem.stateObjects))
	for addr := range t.mem.stateObjects {
		memAddrs[addr] = true
	}

	// Insert dirty mem accounts (may override or logically delete DB entries).
	for addr, obj := range t.mem.stateObjects {
		if obj.selfDestructed {
			continue // logically deleted; skip
		}
		storageRoot := t.computeStorageRootFromDB(addr, obj)
		codeHash := obj.account.CodeHash
		if len(codeHash) == 0 {
			codeHash = types.EmptyCodeHash.Bytes()
		}
		acc := rlpAccount{
			Nonce:    obj.account.Nonce,
			Balance:  obj.account.Balance,
			Root:     storageRoot[:],
			CodeHash: codeHash,
		}
		encoded, err := rlp.EncodeToBytes(acc)
		if err != nil {
			continue
		}
		stateTrie.Put(crypto.Keccak256(addr[:]), encoded)
	}

	// Insert committed accounts not present in the dirty buffer.
	//
	// Prefer the frozen snapshot (captured at the last Commit) over a live DB
	// scan. The frozen snapshot is a stable, read-only copy that cannot be
	// corrupted by concurrent Commit calls from other goroutines (e.g. a
	// newPayload handler committing while a block builder is computing a root).
	//
	// Fall back to a live DB scan for TrieStateDB instances that have never
	// been committed (frozenAccounts == nil), such as a freshly-constructed
	// reader or the genesis-initialisation path.
	if t.frozenAccounts != nil {
		for addr, rlpBytes := range t.frozenAccounts {
			if memAddrs[addr] {
				continue // dirty mem overrides frozen snapshot
			}
			stateTrie.Put(crypto.Keccak256(addr[:]), rlpBytes)
		}
	} else {
		iter := t.db.NewIterator(dbPrefixAccount)
		for iter.Next() {
			key := iter.Key()
			if len(key) != 2+20 {
				continue
			}
			var addr types.Address
			copy(addr[:], key[2:])
			if memAddrs[addr] {
				continue
			}
			stateTrie.Put(crypto.Keccak256(addr[:]), iter.Value())
		}
		iter.Release()
	}

	return stateTrie
}

// --- StateDB interface: commit ---

// Commit flushes the dirty buffer to the DB, resets the buffer, and returns
// the new state root. After a successful Commit, Dup() is O(1).
func (t *TrieStateDB) Commit() (types.Hash, error) {
	// Capture ALL dirty storage slots BEFORE the dirty→committed flush so we
	// know exactly which slots to write/delete in the DB.  We only write dirty
	// slots — never the full committedStorage — to avoid the O(N_total) write
	// amplification that caused 700%+ CPU with storagespam workloads.
	type dirtySlot struct {
		addr types.Address
		slot types.Hash
		val  types.Hash
	}
	var dirtySlots []dirtySlot
	for addr, obj := range t.mem.stateObjects {
		if obj.selfDestructed {
			continue // entire storage will be deleted via prefix scan below
		}
		for slot, val := range obj.dirtyStorage {
			dirtySlots = append(dirtySlots, dirtySlot{addr, slot, val})
		}
	}

	// Flush dirty → committed storage.
	// Unlike MemoryStateDB we keep zero values in committedStorage as explicit
	// "slot was zeroed this block" markers so computeStorageRootFromDB can
	// override the DB value with zero (i.e. exclude the slot from the trie).
	for _, obj := range t.mem.stateObjects {
		for key, val := range obj.dirtyStorage {
			obj.committedStorage[key] = val // keep zeros as explicit overrides
		}
		obj.dirtyStorage = make(map[types.Hash]types.Hash)
	}

	// Compute state root over the full merged state (DB + in-memory overlay).
	root := t.buildStateTrie().Hash()

	// Persist dirty buffer to DB atomically.
	batch := t.db.NewBatch()
	for addr, obj := range t.mem.stateObjects {
		if obj.selfDestructed {
			// Remove account and all its storage from the DB.
			if err := batch.Delete(accountDBKey(addr)); err != nil {
				return types.Hash{}, fmt.Errorf("delete account %s: %w", addr.Hex(), err)
			}
			iter := t.db.NewIterator(storageDBPrefix(addr))
			for iter.Next() {
				if err := batch.Delete(iter.Key()); err != nil {
					iter.Release()
					return types.Hash{}, fmt.Errorf("delete storage %s: %w", addr.Hex(), err)
				}
			}
			iter.Release()
			continue
		}

		// Write account record (includes baked-in storage root for fast reload).
		// committedStorage contains the freshly-flushed dirty values at this point,
		// so computeStorageRootFromDB gives the correct post-block storage root.
		storageRoot := t.computeStorageRootFromDB(addr, obj)
		codeHash := obj.account.CodeHash
		if len(codeHash) == 0 {
			codeHash = types.EmptyCodeHash.Bytes()
		}
		acc := rlpAccount{
			Nonce:    obj.account.Nonce,
			Balance:  obj.account.Balance,
			Root:     storageRoot[:],
			CodeHash: codeHash,
		}
		encoded, err := rlp.EncodeToBytes(acc)
		if err != nil {
			return types.Hash{}, fmt.Errorf("encode account %s: %w", addr.Hex(), err)
		}
		if err := batch.Put(accountDBKey(addr), encoded); err != nil {
			return types.Hash{}, fmt.Errorf("put account %s: %w", addr.Hex(), err)
		}

		// Write contract code (idempotent: code is immutable once deployed).
		if len(obj.code) > 0 {
			if err := batch.Put(codeDBKey(obj.account.CodeHash), obj.code); err != nil {
				return types.Hash{}, fmt.Errorf("put code %s: %w", addr.Hex(), err)
			}
		}
	}

	// Write only the dirty storage slots captured before the flush.
	// Zero-value slots are deleted from DB; non-zero are upserted.
	for _, ds := range dirtySlots {
		if ds.val == (types.Hash{}) {
			if err := batch.Delete(storageDBKey(ds.addr, ds.slot)); err != nil {
				return types.Hash{}, fmt.Errorf("delete slot %s[%s]: %w", ds.addr.Hex(), ds.slot.Hex(), err)
			}
		} else {
			if err := batch.Put(storageDBKey(ds.addr, ds.slot), ds.val[:]); err != nil {
				return types.Hash{}, fmt.Errorf("put slot %s[%s]: %w", ds.addr.Hex(), ds.slot.Hex(), err)
			}
		}
	}

	// For accounts reset by CreateAccount, purge ALL stale DB storage.
	// The new account's dirty slots were captured above and will be written;
	// any remaining DB slots belong to the old incarnation and must be removed.
	for addr := range t.recreated {
		obj, ok := t.mem.stateObjects[addr]
		if !ok || obj.selfDestructed {
			continue // self-destructed: already handled via prefix delete above
		}
		iter := t.db.NewIterator(storageDBPrefix(addr))
		for iter.Next() {
			if err := batch.Delete(iter.Key()); err != nil {
				iter.Release()
				return types.Hash{}, fmt.Errorf("delete stale slot %s: %w", addr.Hex(), err)
			}
		}
		iter.Release()
	}

	if err := batch.Write(); err != nil {
		return types.Hash{}, fmt.Errorf("flush state to db: %w", err)
	}

	// Capture a frozen snapshot of all committed account RLPs.
	// This snapshot is used by buildStateTrie() on subsequent Dup() copies so
	// that concurrent DB writes from other goroutines cannot corrupt the
	// state-root computation. We iterate the DB immediately after the batch
	// write so the snapshot is consistent with the just-committed state.
	frozen := make(map[types.Address][]byte)
	frozenIter := t.db.NewIterator(dbPrefixAccount)
	for frozenIter.Next() {
		key := frozenIter.Key()
		if len(key) != 2+20 {
			continue
		}
		var addr types.Address
		copy(addr[:], key[2:])
		val := make([]byte, len(frozenIter.Value()))
		copy(val, frozenIter.Value())
		frozen[addr] = val
	}
	frozenIter.Release()
	t.frozenAccounts = frozen

	// Reset dirty buffer and recreated tracking: memory drops to near-zero.
	t.mem = NewMemoryStateDB()
	t.recreated = nil

	return root, nil
}

// ClearAllState deletes every account, storage-slot, and code entry from the
// underlying DB. Call this before committing a full canonical snapshot (e.g.
// during a reorg) so that stale entries from reverted blocks are removed.
func (t *TrieStateDB) ClearAllState() error {
	prefixes := [][]byte{dbPrefixAccount, dbPrefixStorage, dbPrefixCode}
	for _, pfx := range prefixes {
		iter := t.db.NewIterator(pfx)
		var keys [][]byte
		for iter.Next() {
			k := make([]byte, len(iter.Key()))
			copy(k, iter.Key())
			keys = append(keys, k)
		}
		iter.Release()
		for _, k := range keys {
			if err := t.db.Delete(k); err != nil {
				return fmt.Errorf("clear state key: %w", err)
			}
		}
	}
	return nil
}

// --- StateDB interface: copy ---

// Dup returns an independent copy of this TrieStateDB. Both copies share the
// same DB reference (reads/writes are serialized through the DB's own lock).
// After Commit(), mem is nearly empty so this is cheap.
func (t *TrieStateDB) Dup() StateDB {
	var recreatedCopy map[types.Address]struct{}
	if len(t.recreated) > 0 {
		recreatedCopy = make(map[types.Address]struct{}, len(t.recreated))
		for k := range t.recreated {
			recreatedCopy[k] = struct{}{}
		}
	}
	return &TrieStateDB{
		mem:            t.mem.Copy(),
		db:             t.db,
		gcMode:         t.gcMode,
		recreated:      recreatedCopy,
		frozenAccounts: t.frozenAccounts, // shared read-only; never mutated by callers
	}
}

// Verify interface compliance at compile time.
var _ StateDB = (*TrieStateDB)(nil)
