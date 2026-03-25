package bal

import (
	"sync"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// mockStateReader implements StateReader for testing.
type mockStateReader struct {
	mu       sync.Mutex
	balances map[types.Address]uint64
	nonces   map[types.Address]uint64
	codes    map[types.Address][]byte
	storage  map[types.Address]map[types.Hash]types.Hash
	accesses int // count of all read calls
}

func newMockStateReader() *mockStateReader {
	return &mockStateReader{
		balances: make(map[types.Address]uint64),
		nonces:   make(map[types.Address]uint64),
		codes:    make(map[types.Address][]byte),
		storage:  make(map[types.Address]map[types.Hash]types.Hash),
	}
}

func (m *mockStateReader) inc() {
	m.mu.Lock()
	m.accesses++
	m.mu.Unlock()
}

func (m *mockStateReader) Accesses() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.accesses
}

func (m *mockStateReader) GetBalance(addr types.Address) uint64 {
	m.inc()
	return m.balances[addr]
}
func (m *mockStateReader) GetNonce(addr types.Address) uint64 {
	m.inc()
	return m.nonces[addr]
}
func (m *mockStateReader) GetCode(addr types.Address) []byte {
	m.inc()
	return m.codes[addr]
}
func (m *mockStateReader) GetCodeSize(addr types.Address) int {
	m.inc()
	return len(m.codes[addr])
}
func (m *mockStateReader) GetState(addr types.Address, key types.Hash) types.Hash {
	m.inc()
	if s, ok := m.storage[addr]; ok {
		return s[key]
	}
	return types.Hash{}
}
func (m *mockStateReader) Exist(addr types.Address) bool {
	m.inc()
	_, b := m.balances[addr]
	_, n := m.nonces[addr]
	_, c := m.codes[addr]
	return b || n || c
}

func TestMutationOverlay_BalanceAndNonce(t *testing.T) {
	base := newMockStateReader()
	addr := types.Address{0x01}
	base.balances[addr] = 1000
	base.nonces[addr] = 5

	mutations := []TxMutation{
		{TxIndex: 0, BalanceDeltas: map[types.Address]int64{addr: -100}, NonceDeltas: map[types.Address]int64{addr: 1}},
		{TxIndex: 1, BalanceDeltas: map[types.Address]int64{addr: -200}, NonceDeltas: map[types.Address]int64{addr: 1}},
	}

	// Apply 0 mutations: see base state.
	o0 := NewMutationOverlay(base, mutations, 0)
	if o0.GetBalance(addr) != 1000 {
		t.Errorf("upToIdx=0: balance = %d, want 1000", o0.GetBalance(addr))
	}
	if o0.GetNonce(addr) != 5 {
		t.Errorf("upToIdx=0: nonce = %d, want 5", o0.GetNonce(addr))
	}

	// Apply 1 mutation: -100 balance, +1 nonce.
	o1 := NewMutationOverlay(base, mutations, 1)
	if o1.GetBalance(addr) != 900 {
		t.Errorf("upToIdx=1: balance = %d, want 900", o1.GetBalance(addr))
	}
	if o1.GetNonce(addr) != 6 {
		t.Errorf("upToIdx=1: nonce = %d, want 6", o1.GetNonce(addr))
	}

	// Apply 2 mutations: -300 balance total, +2 nonce.
	o2 := NewMutationOverlay(base, mutations, 2)
	if o2.GetBalance(addr) != 700 {
		t.Errorf("upToIdx=2: balance = %d, want 700", o2.GetBalance(addr))
	}
	if o2.GetNonce(addr) != 7 {
		t.Errorf("upToIdx=2: nonce = %d, want 7", o2.GetNonce(addr))
	}
}

func TestMutationOverlay_StorageLatestWins(t *testing.T) {
	base := newMockStateReader()
	addr := types.Address{0x01}
	key := types.Hash{0xAA}

	base.storage[addr] = map[types.Hash]types.Hash{key: {0x01}}

	mutations := []TxMutation{
		{TxIndex: 0, StorageWrites: map[types.Address]map[types.Hash]types.Hash{
			addr: {key: {0x02}},
		}},
		{TxIndex: 1, StorageWrites: map[types.Address]map[types.Hash]types.Hash{
			addr: {key: {0x03}},
		}},
	}

	o1 := NewMutationOverlay(base, mutations, 1)
	if got := o1.GetState(addr, key); got != (types.Hash{0x02}) {
		t.Errorf("upToIdx=1: got %x, want 0x02", got)
	}

	o2 := NewMutationOverlay(base, mutations, 2)
	if got := o2.GetState(addr, key); got != (types.Hash{0x03}) {
		t.Errorf("upToIdx=2: got %x, want 0x03", got)
	}
}

func TestReadTracker(t *testing.T) {
	base := newMockStateReader()
	addr1 := types.Address{0x01}
	addr2 := types.Address{0x02}
	key := types.Hash{0xAA}

	base.balances[addr1] = 100
	base.storage[addr2] = map[types.Hash]types.Hash{key: {0x01}}

	tracker := NewReadTracker(base)

	tracker.GetBalance(addr1)
	tracker.GetState(addr2, key)
	tracker.GetNonce(addr1)

	accounts := tracker.AccessedAccounts()
	if len(accounts) != 2 {
		t.Errorf("accessed %d accounts, want 2", len(accounts))
	}

	storage := tracker.AccessedStorage()
	if keys, ok := storage[addr2]; !ok || len(keys) != 1 {
		t.Errorf("expected 1 storage key for addr2, got %v", storage[addr2])
	}
}

func TestBALPrefetcher_StartsAndStops(t *testing.T) {
	base := newMockStateReader()
	p := NewBALPrefetcher(base, 2)
	p.Stop()
	// Should not panic on double stop.
	p.Stop()
}
