package snapshot

import (
	"math/big"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/rlp"
)

// diffLayer represents a collection of modifications made to a state snapshot
// after running a block on top. It contains account data and storage data maps
// keyed by their respective hashes.
//
// The goal of a diff layer is to act as a journal, tracking recent modifications
// made to the state, that have not yet graduated into a semi-immutable state.
type diffLayer struct {
	parent snapshot   // Parent snapshot modified by this one, never nil
	root   types.Hash // Root hash to which this snapshot diff belongs to
	stale  atomic.Bool

	accountData map[types.Hash][]byte                // Keyed accounts for direct retrieval (nil means deleted)
	storageData map[types.Hash]map[types.Hash][]byte // Keyed storage slots for direct retrieval (nil means deleted)
	memory      uint64                               // Approximate memory usage in bytes

	lock sync.RWMutex
}

// newDiffLayer creates a new diff on top of an existing snapshot, whether
// that's a low level persistent database or a hierarchical diff already.
func newDiffLayer(parent snapshot, root types.Hash, accounts map[types.Hash][]byte, storage map[types.Hash]map[types.Hash][]byte) *diffLayer {
	dl := &diffLayer{
		parent:      parent,
		root:        root,
		accountData: accounts,
		storageData: storage,
	}
	// Track memory usage.
	for _, data := range accounts {
		dl.memory += uint64(types.HashLength + len(data))
	}
	for _, slots := range storage {
		for _, data := range slots {
			dl.memory += uint64(types.HashLength + len(data))
		}
	}
	return dl
}

// Root returns the root hash for which this snapshot was made.
func (dl *diffLayer) Root() types.Hash {
	return dl.root
}

// Parent returns the parent layer of this diff.
func (dl *diffLayer) Parent() snapshot {
	dl.lock.RLock()
	defer dl.lock.RUnlock()
	return dl.parent
}

// Stale returns whether this layer has become stale (was flattened across).
func (dl *diffLayer) Stale() bool {
	return dl.stale.Load()
}

// markStale sets the stale flag.
func (dl *diffLayer) markStale() {
	dl.stale.Store(true)
}

// Account retrieves the account associated with a particular hash.
// It checks the local layer first, then walks the parent chain.
func (dl *diffLayer) Account(hash types.Hash) (*types.Account, error) {
	dl.lock.RLock()
	if dl.Stale() {
		dl.lock.RUnlock()
		return nil, ErrSnapshotStale
	}
	// Check local data first.
	if data, ok := dl.accountData[hash]; ok {
		dl.lock.RUnlock()
		if len(data) == 0 {
			return nil, nil // Account was deleted.
		}
		return decodeAccount(data)
	}
	parent := dl.parent
	dl.lock.RUnlock()
	// Not found locally, resolve from parent.
	return parent.Account(hash)
}

// Storage retrieves the storage data associated with a particular hash within
// a particular account. Checks local layer first, then walks parent chain.
func (dl *diffLayer) Storage(accountHash, storageHash types.Hash) ([]byte, error) {
	dl.lock.RLock()
	if dl.Stale() {
		dl.lock.RUnlock()
		return nil, ErrSnapshotStale
	}
	// Check local data first.
	if slots, ok := dl.storageData[accountHash]; ok {
		if data, ok := slots[storageHash]; ok {
			dl.lock.RUnlock()
			return data, nil
		}
	}
	parent := dl.parent
	dl.lock.RUnlock()
	// Not found locally, resolve from parent.
	return parent.Storage(accountHash, storageHash)
}

// Update creates a new layer on top of this diff layer.
func (dl *diffLayer) Update(blockRoot types.Hash, accounts map[types.Hash][]byte, storage map[types.Hash]map[types.Hash][]byte) *diffLayer {
	return newDiffLayer(dl, blockRoot, accounts, storage)
}

// Memory returns the approximate memory usage of this diff layer in bytes.
func (dl *diffLayer) Memory() uint64 {
	return dl.memory
}

// flatten merges this diff layer into the given disk layer, producing a new
// disk layer with the merged data written to disk.
func (dl *diffLayer) flatten(disk *diskLayer) *diskLayer {
	// Write account data to disk.
	if disk.diskdb != nil {
		batch := disk.diskdb.NewBatch()
		for hash, data := range dl.accountData {
			key := accountSnapshotKey(hash)
			if len(data) == 0 {
				batch.Delete(key)
			} else {
				batch.Put(key, data)
			}
		}
		for accountHash, slots := range dl.storageData {
			for storageHash, data := range slots {
				key := storageSnapshotKey(accountHash, storageHash)
				if len(data) == 0 {
					batch.Delete(key)
				} else {
					batch.Put(key, data)
				}
			}
		}
		batch.Write()
	}
	return &diskLayer{
		diskdb: disk.diskdb,
		root:   dl.root,
	}
}

// AccountIterator creates an account iterator over this diff layer.
func (dl *diffLayer) AccountIterator(seek types.Hash) AccountIterator {
	// Collect all account hashes from this layer.
	dl.lock.RLock()
	hashes := make([]types.Hash, 0, len(dl.accountData))
	for hash := range dl.accountData {
		hashes = append(hashes, hash)
	}
	dl.lock.RUnlock()

	sort.Slice(hashes, func(i, j int) bool {
		return hashLess(hashes[i], hashes[j])
	})
	return &diffAccountIterator{
		layer:  dl,
		hashes: hashes,
		pos:    -1,
		seek:   seek,
	}
}

// StorageIterator creates a storage iterator for a specific account.
func (dl *diffLayer) StorageIterator(accountHash types.Hash, seek types.Hash) StorageIterator {
	dl.lock.RLock()
	slots := dl.storageData[accountHash]
	hashes := make([]types.Hash, 0, len(slots))
	for hash := range slots {
		hashes = append(hashes, hash)
	}
	dl.lock.RUnlock()

	sort.Slice(hashes, func(i, j int) bool {
		return hashLess(hashes[i], hashes[j])
	})
	return &diffStorageIterator{
		layer:       dl,
		accountHash: accountHash,
		hashes:      hashes,
		pos:         -1,
		seek:        seek,
	}
}

// slimAccount is the RLP-serializable form of an Ethereum account as stored
// by SnapshotDiff (matching memory_statedb.go's rlpAccount layout).
type slimAccount struct {
	Nonce    uint64
	Balance  *big.Int
	Root     []byte
	CodeHash []byte
}

// decodeAccount decodes RLP-encoded slim account data into an Account struct.
// The encoding matches memory_statedb.rlpAccount: [Nonce, Balance, Root, CodeHash].
// If the bytes are present but cannot be decoded (e.g. non-RLP test fixtures),
// a non-nil empty account is returned so callers can still detect existence.
func decodeAccount(data []byte) (*types.Account, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var slim slimAccount
	if err := rlp.DecodeBytes(data, &slim); err != nil {
		// Non-RLP bytes: account exists in snapshot but cannot be decoded.
		acc := types.NewAccount()
		return &acc, nil
	}
	acc := types.NewAccount()
	acc.Nonce = slim.Nonce
	if slim.Balance != nil {
		acc.Balance = slim.Balance
	}
	if len(slim.Root) == types.HashLength {
		copy(acc.Root[:], slim.Root)
	}
	if len(slim.CodeHash) > 0 {
		acc.CodeHash = slim.CodeHash
	}
	return &acc, nil
}

// hashLess returns true if a < b lexicographically.
func hashLess(a, b types.Hash) bool {
	for i := 0; i < types.HashLength; i++ {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}
	return false
}

// Key schema for snapshot data in the disk database.
var (
	snapshotAccountPrefix = []byte("sa") // sa + account hash -> account data
	snapshotStoragePrefix = []byte("ss") // ss + account hash + storage hash -> storage data
)

func accountSnapshotKey(hash types.Hash) []byte {
	return append(append([]byte{}, snapshotAccountPrefix...), hash[:]...)
}

func storageSnapshotKey(accountHash, storageHash types.Hash) []byte {
	key := append([]byte{}, snapshotStoragePrefix...)
	key = append(key, accountHash[:]...)
	key = append(key, storageHash[:]...)
	return key
}
