package sync

// snap_test_helpers_test.go provides mock implementations of SnapPeer and
// StateWriter for use by root sync package integration tests.

import (
	"bytes"
	gosync "sync"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"

	"math/big"
	"sort"
)

// mockSnapPeer implements SnapPeer for testing.
type mockSnapPeer struct {
	mu       gosync.Mutex
	id       string
	accounts []AccountData
	storage  map[types.Hash][]StorageData
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
		storage:  make(map[types.Hash][]StorageData),
		codes:    make(map[types.Hash][]byte),
		healData: make(map[string][]byte),
	}
}

func (m *mockSnapPeer) ID() string { return m.id }

func (m *mockSnapPeer) RequestAccountRange(req AccountRangeRequest) (*AccountRangeResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.accountErr != nil {
		return nil, m.accountErr
	}
	var result []AccountData
	for _, acct := range m.accounts {
		if bytes.Compare(acct.Hash[:], req.Origin[:]) >= 0 &&
			bytes.Compare(acct.Hash[:], req.Limit[:]) <= 0 {
			result = append(result, acct)
		}
		if len(result) >= MaxAccountRange {
			break
		}
	}
	return &AccountRangeResponse{ID: req.ID, Accounts: result}, nil
}

func (m *mockSnapPeer) RequestStorageRange(req StorageRangeRequest) (*StorageRangeResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.storageErr != nil {
		return nil, m.storageErr
	}
	var result []StorageData
	for _, acctHash := range req.Accounts {
		for _, slot := range m.storage[acctHash] {
			result = append(result, slot)
		}
	}
	return &StorageRangeResponse{ID: req.ID, Slots: result}, nil
}

func (m *mockSnapPeer) RequestBytecodes(req BytecodeRequest) (*BytecodeResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.bytecodeErr != nil {
		return nil, m.bytecodeErr
	}
	var result []BytecodeData
	for _, hash := range req.Hashes {
		if code, ok := m.codes[hash]; ok {
			result = append(result, BytecodeData{Hash: hash, Code: code})
		}
	}
	return &BytecodeResponse{ID: req.ID, Codes: result}, nil
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

// mockStateWriter implements StateWriter for testing.
type mockStateWriter struct {
	mu       gosync.Mutex
	accounts map[types.Hash]AccountData
	storage  map[string][]byte
	codes    map[types.Hash][]byte
	nodes    map[string][]byte

	missingPaths [][]byte
	healRounds   int
	healCalls    int
}

func newMockStateWriter() *mockStateWriter {
	return &mockStateWriter{
		accounts: make(map[types.Hash]AccountData),
		storage:  make(map[string][]byte),
		codes:    make(map[types.Hash][]byte),
		nodes:    make(map[string][]byte),
	}
}

func (w *mockStateWriter) WriteAccount(hash types.Hash, data AccountData) error {
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

// blockingSnapPeer is a SnapPeer that blocks until unblocked via blockCh.
type blockingSnapPeer struct {
	inner   *mockSnapPeer
	blockCh chan struct{}
}

func (p *blockingSnapPeer) ID() string { return p.inner.ID() }
func (p *blockingSnapPeer) RequestAccountRange(req AccountRangeRequest) (*AccountRangeResponse, error) {
	<-p.blockCh
	return p.inner.RequestAccountRange(req)
}
func (p *blockingSnapPeer) RequestStorageRange(req StorageRangeRequest) (*StorageRangeResponse, error) {
	return p.inner.RequestStorageRange(req)
}
func (p *blockingSnapPeer) RequestBytecodes(req BytecodeRequest) (*BytecodeResponse, error) {
	return p.inner.RequestBytecodes(req)
}
func (p *blockingSnapPeer) RequestTrieNodes(root types.Hash, paths [][]byte) ([][]byte, error) {
	return p.inner.RequestTrieNodes(root, paths)
}

// makeTestAccounts creates n test accounts sorted by hash.
func makeTestAccounts(n int) []AccountData { return makeSnapTestAccounts(n) }

// makeSnapTestAccounts creates n test accounts sorted by hash.
func makeSnapTestAccounts(n int) []AccountData {
	accounts := make([]AccountData, n)
	for i := 0; i < n; i++ {
		hashInput := []byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)}
		hash := types.BytesToHash(crypto.Keccak256(hashInput))
		accounts[i] = AccountData{
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
