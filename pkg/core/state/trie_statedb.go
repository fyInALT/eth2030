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

// TrieStateDB is a disk-backed StateDB implementation. It keeps a
// MemoryStateDB as a dirty write buffer for the current block and persists
// committed state to a rawdb.Database.
//
// Memory model: after each Commit() the dirty buffer is reset. Subsequent
// reads populate the buffer on demand from the DB. Memory usage is therefore
// proportional to the working set of a single block rather than growing with
// chain history.
type TrieStateDB struct {
	mem *MemoryStateDB // dirty write buffer for the current block
	db  rawdb.Database // persistent backing store
}

// NewTrieStateDB creates a TrieStateDB backed by db.
func NewTrieStateDB(db rawdb.Database) *TrieStateDB {
	return &TrieStateDB{
		mem: NewMemoryStateDB(),
		db:  db,
	}
}

// NewTrieStateDBFromMemory creates a TrieStateDB that adopts an existing
// MemoryStateDB as its initial dirty layer. Call Commit() immediately after
// to persist the in-memory state to the DB and free the dirty layer. This
// is typically used to convert a genesis MemoryStateDB into a TrieStateDB.
func NewTrieStateDBFromMemory(db rawdb.Database, mem *MemoryStateDB) *TrieStateDB {
	return &TrieStateDB{
		mem: mem,
		db:  db,
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

	// Load all storage slots for this account.
	iter := t.db.NewIterator(storageDBPrefix(addr))
	for iter.Next() {
		key := iter.Key()
		if len(key) != 2+20+32 {
			continue
		}
		var slot types.Hash
		copy(slot[:], key[22:])
		var val types.Hash
		copy(val[:], iter.Value())
		if val != (types.Hash{}) {
			obj.committedStorage[slot] = val
		}
	}
	iter.Release()

	// Load contract code if non-empty.
	if len(acc.CodeHash) > 0 && types.BytesToHash(acc.CodeHash) != types.EmptyCodeHash {
		if code, err := t.db.Get(codeDBKey(acc.CodeHash)); err == nil {
			obj.code = code
		}
	}

	t.mem.stateObjects[addr] = obj
}

// --- StateDB interface: account operations ---

func (t *TrieStateDB) CreateAccount(addr types.Address) {
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

// --- StateDB interface: storage ---

func (t *TrieStateDB) GetState(addr types.Address, key types.Hash) types.Hash {
	t.loadFromDB(addr)
	return t.mem.GetState(addr, key)
}

func (t *TrieStateDB) SetState(addr types.Address, key types.Hash, value types.Hash) {
	t.loadFromDB(addr)
	t.mem.SetState(addr, key, value)
}

func (t *TrieStateDB) GetCommittedState(addr types.Address, key types.Hash) types.Hash {
	t.loadFromDB(addr)
	return t.mem.GetCommittedState(addr, key)
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
	return t.mem.StorageRoot(addr)
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
		storageRoot := computeStorageRoot(obj)
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

	// Insert DB accounts not present in the dirty buffer. The stored RLP
	// already contains the correct storage root from last Commit, so we
	// insert it verbatim — no need to reload all storage slots.
	iter := t.db.NewIterator(dbPrefixAccount)
	for iter.Next() {
		key := iter.Key()
		if len(key) != 2+20 {
			continue
		}
		var addr types.Address
		copy(addr[:], key[2:])
		if memAddrs[addr] {
			continue // handled above (dirty or selfDestructed)
		}
		stateTrie.Put(crypto.Keccak256(addr[:]), iter.Value())
	}
	iter.Release()

	return stateTrie
}

// --- StateDB interface: commit ---

// Commit flushes the dirty buffer to the DB, resets the buffer, and returns
// the new state root. After a successful Commit, Dup() is O(1).
func (t *TrieStateDB) Commit() (types.Hash, error) {
	// Flush dirty → committed storage in the mem buffer (mirrors MemoryStateDB.Commit).
	for _, obj := range t.mem.stateObjects {
		for key, val := range obj.dirtyStorage {
			if val == (types.Hash{}) {
				delete(obj.committedStorage, key)
			} else {
				obj.committedStorage[key] = val
			}
		}
		obj.dirtyStorage = make(map[types.Hash]types.Hash)
	}

	// Compute state root over the full merged state.
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

		// Write account record (includes storage root for unchanged-account fast path).
		storageRoot := computeStorageRoot(obj)
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

		// Write committed storage slots (dirty was flushed to committed above).
		for slot, val := range obj.committedStorage {
			dbKey := storageDBKey(addr, slot)
			if val == (types.Hash{}) {
				if err := batch.Delete(dbKey); err != nil {
					return types.Hash{}, fmt.Errorf("delete slot %s[%s]: %w", addr.Hex(), slot.Hex(), err)
				}
			} else {
				if err := batch.Put(dbKey, val[:]); err != nil {
					return types.Hash{}, fmt.Errorf("put slot %s[%s]: %w", addr.Hex(), slot.Hex(), err)
				}
			}
		}
	}

	if err := batch.Write(); err != nil {
		return types.Hash{}, fmt.Errorf("flush state to db: %w", err)
	}

	// Reset dirty buffer: memory usage drops to near-zero until next block.
	t.mem = NewMemoryStateDB()

	return root, nil
}

// --- StateDB interface: copy ---

// Dup returns an independent copy of this TrieStateDB. Both copies share the
// same DB reference (reads/writes are serialized through the DB's own lock).
// After Commit(), mem is nearly empty so this is cheap.
func (t *TrieStateDB) Dup() StateDB {
	return &TrieStateDB{
		mem: t.mem.Copy(),
		db:  t.db,
	}
}

// Verify interface compliance at compile time.
var _ StateDB = (*TrieStateDB)(nil)
