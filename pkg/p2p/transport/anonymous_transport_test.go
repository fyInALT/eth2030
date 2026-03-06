package transport

import (
	"math/big"
	"sync"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// mockTransport implements AnonymousTransport for testing.
type mockTransport struct {
	name    string
	ch      chan *types.Transaction
	started bool
	stopped bool
}

func newMockTransport(name string) *mockTransport {
	return &mockTransport{
		name: name,
		ch:   make(chan *types.Transaction, 16),
	}
}

func (m *mockTransport) Name() string { return m.name }

func (m *mockTransport) Submit(tx *types.Transaction) error {
	if tx == nil {
		return ErrAnonTransportNilTx
	}
	select {
	case m.ch <- tx:
	default:
	}
	return nil
}

func (m *mockTransport) Receive() <-chan *types.Transaction { return m.ch }

func (m *mockTransport) Start() error {
	m.started = true
	return nil
}

func (m *mockTransport) Stop() error {
	m.stopped = true
	return nil
}

func testTx() *types.Transaction {
	to := types.Address{}
	return types.NewTransaction(&types.LegacyTx{
		Nonce:    0,
		To:       &to,
		Value:    big.NewInt(0),
		Gas:      21000,
		GasPrice: big.NewInt(1),
	})
}

func TestTransportManager_RegisterTransport(t *testing.T) {
	tm := NewTransportManager()
	mt := newMockTransport("test")

	if err := tm.RegisterTransport(mt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tm.TransportCount() != 1 {
		t.Fatalf("expected 1 transport, got %d", tm.TransportCount())
	}
}

func TestTransportManager_RegisterTransport_Duplicate(t *testing.T) {
	tm := NewTransportManager()
	mt := newMockTransport("test")

	if err := tm.RegisterTransport(mt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tm.RegisterTransport(mt); err != ErrAnonTransportExists {
		t.Fatalf("expected ErrAnonTransportExists, got %v", err)
	}
}

func TestTransportManager_UnregisterTransport(t *testing.T) {
	tm := NewTransportManager()
	mt := newMockTransport("test")

	if err := tm.RegisterTransport(mt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tm.UnregisterTransport("test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tm.TransportCount() != 0 {
		t.Fatalf("expected 0 transports, got %d", tm.TransportCount())
	}
	if !mt.stopped {
		t.Fatal("expected transport to be stopped after unregister")
	}
}

func TestTransportManager_UnregisterTransport_NotFound(t *testing.T) {
	tm := NewTransportManager()

	if err := tm.UnregisterTransport("nonexistent"); err != ErrAnonTransportNotFound {
		t.Fatalf("expected ErrAnonTransportNotFound, got %v", err)
	}
}

func TestTransportManager_SubmitAnonymous(t *testing.T) {
	tm := NewTransportManager()
	mt1 := newMockTransport("t1")
	mt2 := newMockTransport("t2")

	_ = tm.RegisterTransport(mt1)
	_ = tm.RegisterTransport(mt2)

	tx := testTx()
	n, errs := tm.SubmitAnonymous(tx)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if n != 2 {
		t.Fatalf("expected 2 submissions, got %d", n)
	}

	// Verify stats.
	stats := tm.GetStats()
	for _, s := range stats {
		if s.Submitted != 1 {
			t.Fatalf("expected 1 submitted for %s, got %d", s.Name, s.Submitted)
		}
	}
}

func TestTransportManager_SubmitAnonymous_NilTx(t *testing.T) {
	tm := NewTransportManager()

	n, errs := tm.SubmitAnonymous(nil)
	if n != 0 {
		t.Fatalf("expected 0 submissions, got %d", n)
	}
	if len(errs) != 1 || errs[0] != ErrAnonTransportNilTx {
		t.Fatalf("expected ErrAnonTransportNilTx, got %v", errs)
	}
}

func TestTransportManager_StartAll(t *testing.T) {
	tm := NewTransportManager()
	mt1 := newMockTransport("t1")
	mt2 := newMockTransport("t2")

	_ = tm.RegisterTransport(mt1)
	_ = tm.RegisterTransport(mt2)

	errs := tm.StartAll()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !mt1.started || !mt2.started {
		t.Fatal("expected all transports to be started")
	}

	stats := tm.GetStats()
	for _, s := range stats {
		if !s.Running {
			t.Fatalf("expected %s to be running", s.Name)
		}
	}
}

func TestTransportManager_StopAll(t *testing.T) {
	tm := NewTransportManager()
	mt1 := newMockTransport("t1")
	mt2 := newMockTransport("t2")

	_ = tm.RegisterTransport(mt1)
	_ = tm.RegisterTransport(mt2)
	_ = tm.StartAll()

	errs := tm.StopAll()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !mt1.stopped || !mt2.stopped {
		t.Fatal("expected all transports to be stopped")
	}

	// Manager should be closed; new registrations fail.
	if err := tm.RegisterTransport(newMockTransport("t3")); err != ErrAnonTransportClosed {
		t.Fatalf("expected ErrAnonTransportClosed, got %v", err)
	}
}

func TestTransportManager_GetStats(t *testing.T) {
	tm := NewTransportManager()
	mt := newMockTransport("test")

	_ = tm.RegisterTransport(mt)
	_ = tm.StartAll()

	tx := testTx()
	tm.SubmitAnonymous(tx)
	tm.SubmitAnonymous(tx)

	stats := tm.GetStats()
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat entry, got %d", len(stats))
	}
	if stats[0].Name != "test" {
		t.Fatalf("expected name 'test', got %q", stats[0].Name)
	}
	if stats[0].Submitted != 2 {
		t.Fatalf("expected 2 submitted, got %d", stats[0].Submitted)
	}
	if !stats[0].Running {
		t.Fatal("expected running=true")
	}
}

func TestTransportManager_Concurrent(t *testing.T) {
	tm := NewTransportManager()
	for i := 0; i < 5; i++ {
		mt := newMockTransport("t" + string(rune('0'+i)))
		_ = tm.RegisterTransport(mt)
	}
	_ = tm.StartAll()

	var wg sync.WaitGroup
	// Concurrent submissions.
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tx := testTx()
			tm.SubmitAnonymous(tx)
		}()
	}

	// Concurrent stats reads.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tm.GetStats()
			tm.TransportCount()
		}()
	}

	wg.Wait()

	errs := tm.StopAll()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors on stop: %v", errs)
	}
}
