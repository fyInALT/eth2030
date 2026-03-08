package sync

// snap_test_helpers_test.go provides mock implementations of SnapPeer and
// StateWriter for use by root sync package integration tests.

import (
	"bytes"
	"math/big"
	"sort"
	gosync "sync"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/sync/snap"
)

// mockSnapPeer implements snap.SnapPeer for testing.
type mockSnapPeer struct {
	mu       gosync.Mutex
	id       string
	accounts []snap.AccountData
	storage  map[types.Hash][]snap.StorageData
	codes    map[types.Hash][]byte
	healData map[string][]byte

	accountErr  error
	storageErr  error
	bytecodeErr error
	healErr     error
}

func newMockSnapPeer(id string) *mockSnapPeer {
	return &mockSnapPeer{
		id:       id,
		storage:  make(map[types.Hash][]snap.StorageData),
		codes:    make(map[types.Hash][]byte),
		healData: make(map[string][]byte),
	}
}

func (m *mockSnapPeer) ID() string { return m.id }

func (m *mockSnapPeer) RequestAccountRange(req snap.AccountRangeRequest) (*snap.AccountRangeResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.accountErr != nil {
		return nil, m.accountErr
	}
	var result []snap.AccountData
	for _, acct := range m.accounts {
		if bytes.Compare(acct.Hash[:], req.Origin[:]) >= 0 &&
			bytes.Compare(acct.Hash[:], req.Limit[:]) <= 0 {
			result = append(result, acct)
		}
		if len(result) >= snap.MaxAccountRange {
			break
		}
	}
	return &snap.AccountRangeResponse{ID: req.ID, Accounts: result}, nil
}

func (m *mockSnapPeer) RequestStorageRange(req snap.StorageRangeRequest) (*snap.StorageRangeResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.storageErr != nil {
		return nil, m.storageErr
	}
	var result []snap.StorageData
	for _, acctHash := range req.Accounts {
		for _, slot := range m.storage[acctHash] {
			result = append(result, slot)
		}
	}
	return &snap.StorageRangeResponse{ID: req.ID, Slots: result}, nil
}

func (m *mockSnapPeer) RequestBytecodes(req snap.BytecodeRequest) (*snap.BytecodeResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.bytecodeErr != nil {
		return nil, m.bytecodeErr
	}
	var result []snap.BytecodeData
	for _, hash := range req.Hashes {
		if code, ok := m.codes[hash]; ok {
			result = append(result, snap.BytecodeData{Hash: hash, Code: code})
		}
	}
	return &snap.BytecodeResponse{ID: req.ID, Codes: result}, nil
}

func (m *mockSnapPeer) RequestTrieNodes(root types.Hash, paths [][]byte) ([][]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.healErr != nil {
		return nil, m.healErr
	}
	result := make([][]byte, len(paths))
	for i, path := range paths {
		result[i] = m.healData[string(path)]
	}
	return result, nil
}

// mockStateWriter implements snap.StateWriter for testing.
type mockStateWriter struct {
	mu       gosync.Mutex
	accounts map[types.Hash]snap.AccountData
	storage  map[string][]byte
	codes    map[types.Hash][]byte
	nodes    map[string][]byte

	missingPaths [][]byte
	healRounds   int
	healCalls    int
}

func newMockStateWriter() *mockStateWriter {
	return &mockStateWriter{
		accounts: make(map[types.Hash]snap.AccountData),
		storage:  make(map[string][]byte),
		codes:    make(map[types.Hash][]byte),
		nodes:    make(map[string][]byte),
	}
}

func (w *mockStateWriter) WriteAccount(hash types.Hash, data snap.AccountData) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.accounts[hash] = data
	return nil
}

func (w *mockStateWriter) WriteStorage(accountHash, slotHash types.Hash, value []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.storage[string(accountHash[:])+string(slotHash[:])] = value
	return nil
}

func (w *mockStateWriter) WriteBytecode(hash types.Hash, code []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.codes[hash] = code
	return nil
}

func (w *mockStateWriter) WriteTrieNode(path []byte, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.nodes[string(path)] = data
	return nil
}

func (w *mockStateWriter) HasBytecode(hash types.Hash) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, ok := w.codes[hash]
	return ok
}

func (w *mockStateWriter) HasTrieNode(path []byte) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, ok := w.nodes[string(path)]
	return ok
}

func (w *mockStateWriter) MissingTrieNodes(_ types.Hash, limit int) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.healCalls++
	if w.healCalls > w.healRounds || len(w.missingPaths) == 0 {
		return nil
	}
	count := limit
	if count > len(w.missingPaths) {
		count = len(w.missingPaths)
	}
	return w.missingPaths[:count]
}

// blockingSnapPeer is a snap.SnapPeer that blocks until unblocked via blockCh.
type blockingSnapPeer struct {
	inner   *mockSnapPeer
	blockCh chan struct{}
}

func (p *blockingSnapPeer) ID() string { return p.inner.ID() }
func (p *blockingSnapPeer) RequestAccountRange(req snap.AccountRangeRequest) (*snap.AccountRangeResponse, error) {
	<-p.blockCh
	return p.inner.RequestAccountRange(req)
}
func (p *blockingSnapPeer) RequestStorageRange(req snap.StorageRangeRequest) (*snap.StorageRangeResponse, error) {
	return p.inner.RequestStorageRange(req)
}
func (p *blockingSnapPeer) RequestBytecodes(req snap.BytecodeRequest) (*snap.BytecodeResponse, error) {
	return p.inner.RequestBytecodes(req)
}
func (p *blockingSnapPeer) RequestTrieNodes(root types.Hash, paths [][]byte) ([][]byte, error) {
	return p.inner.RequestTrieNodes(root, paths)
}

// makeTestAccounts creates n test accounts sorted by hash.
func makeTestAccounts(n int) []snap.AccountData { return makeSnapTestAccounts(n) }

// makeSnapTestAccounts creates n test accounts sorted by hash.
func makeSnapTestAccounts(n int) []snap.AccountData {
	accounts := make([]snap.AccountData, n)
	for i := 0; i < n; i++ {
		hashInput := []byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)}
		hash := types.BytesToHash(crypto.Keccak256(hashInput))
		accounts[i] = snap.AccountData{
			Hash:     hash,
			Nonce:    uint64(i),
			Balance:  big.NewInt(int64(1000 * (i + 1))),
			Root:     types.EmptyRootHash,
			CodeHash: types.EmptyCodeHash,
		}
	}
	sort.Slice(accounts, func(i, j int) bool {
		return bytes.Compare(accounts[i].Hash[:], accounts[j].Hash[:]) < 0
	})
	return accounts
}
