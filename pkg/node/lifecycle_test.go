package node

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// mockService implements the Service interface for testing.
type mockService struct {
	name     string
	started  bool
	stopped  bool
	startErr error
	stopErr  error

	mu       sync.Mutex
	startSeq int // records start order globally
	stopSeq  int // records stop order globally
}

func (m *mockService) Start() error {
	if m.startErr != nil {
		return m.startErr
	}
	m.mu.Lock()
	m.started = true
	m.mu.Unlock()
	return nil
}

func (m *mockService) Stop() error {
	if m.stopErr != nil {
		return m.stopErr
	}
	m.mu.Lock()
	m.stopped = true
	m.mu.Unlock()
	return nil
}

func (m *mockService) Name() string {
	return m.name
}

// seqCounter is a global counter for tracking start/stop ordering in tests.
var (
	seqMu      sync.Mutex
	seqCounter int
)

func nextSeq() int {
	seqMu.Lock()
	defer seqMu.Unlock()
	seqCounter++
	return seqCounter
}

// orderedMockService records its start/stop order.
type orderedMockService struct {
	name     string
	startSeq int
	stopSeq  int
}

func (m *orderedMockService) Start() error {
	m.startSeq = nextSeq()
	return nil
}

func (m *orderedMockService) Stop() error {
	m.stopSeq = nextSeq()
	return nil
}

func (m *orderedMockService) Name() string {
	return m.name
}

func TestRegisterService(t *testing.T) {
	lm := NewLifecycleManager(DefaultLifecycleConfig())

	svc := &mockService{name: "test-svc"}
	if err := lm.Register(svc, 1); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if lm.ServiceCount() != 1 {
		t.Fatalf("want 1 service, got %d", lm.ServiceCount())
	}

	// Registering duplicate name should fail.
	err := lm.Register(&mockService{name: "test-svc"}, 2)
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
}

func TestRegisterMaxServices(t *testing.T) {
	config := DefaultLifecycleConfig()
	config.MaxServices = 2
	lm := NewLifecycleManager(config)

	lm.Register(&mockService{name: "svc1"}, 1)
	lm.Register(&mockService{name: "svc2"}, 2)

	err := lm.Register(&mockService{name: "svc3"}, 3)
	if err == nil {
		t.Fatal("expected error when max services reached")
	}
}

func TestStartAll(t *testing.T) {
	lm := NewLifecycleManager(DefaultLifecycleConfig())

	svc1 := &mockService{name: "svc1"}
	svc2 := &mockService{name: "svc2"}
	lm.Register(svc1, 1)
	lm.Register(svc2, 2)

	errs := lm.StartAll()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !svc1.started || !svc2.started {
		t.Fatal("both services should be started")
	}
	if lm.RunningCount() != 2 {
		t.Fatalf("want 2 running, got %d", lm.RunningCount())
	}
}

func TestStopAll(t *testing.T) {
	lm := NewLifecycleManager(DefaultLifecycleConfig())

	// Reset global counter for this test.
	seqMu.Lock()
	seqCounter = 0
	seqMu.Unlock()

	svc1 := &orderedMockService{name: "svc1"}
	svc2 := &orderedMockService{name: "svc2"}
	svc3 := &orderedMockService{name: "svc3"}

	// Register with priorities: svc1=1, svc2=2, svc3=3.
	lm.Register(svc1, 1)
	lm.Register(svc2, 2)
	lm.Register(svc3, 3)

	lm.StartAll()

	errs := lm.StopAll()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	// Stop should be in reverse priority order: svc3, svc2, svc1.
	if svc3.stopSeq > svc2.stopSeq || svc2.stopSeq > svc1.stopSeq {
		t.Fatalf("stop order wrong: svc3=%d, svc2=%d, svc1=%d",
			svc3.stopSeq, svc2.stopSeq, svc1.stopSeq)
	}
}

func TestGetState(t *testing.T) {
	lm := NewLifecycleManager(DefaultLifecycleConfig())

	svc := &mockService{name: "myservice"}
	lm.Register(svc, 1)

	if lm.GetState("myservice") != StateCreated {
		t.Fatalf("want StateCreated, got %v", lm.GetState("myservice"))
	}

	lm.StartAll()
	if lm.GetState("myservice") != StateRunning {
		t.Fatalf("want StateRunning, got %v", lm.GetState("myservice"))
	}

	lm.StopAll()
	if lm.GetState("myservice") != StateStopped {
		t.Fatalf("want StateStopped, got %v", lm.GetState("myservice"))
	}

	// Unknown service returns StateFailed.
	if lm.GetState("nonexistent") != StateFailed {
		t.Fatalf("want StateFailed for unknown service, got %v", lm.GetState("nonexistent"))
	}
}

func TestStartError(t *testing.T) {
	lm := NewLifecycleManager(DefaultLifecycleConfig())

	good := &mockService{name: "good"}
	bad := &mockService{name: "bad", startErr: errors.New("startup failure")}
	lm.Register(good, 1)
	lm.Register(bad, 2)

	errs := lm.StartAll()
	if len(errs) != 1 {
		t.Fatalf("want 1 error, got %d", len(errs))
	}

	if lm.GetState("good") != StateRunning {
		t.Fatal("good service should be running")
	}
	if lm.GetState("bad") != StateFailed {
		t.Fatal("bad service should be in failed state")
	}
	if lm.RunningCount() != 1 {
		t.Fatalf("want 1 running, got %d", lm.RunningCount())
	}
}

func TestHealthCheck(t *testing.T) {
	lm := NewLifecycleManager(DefaultLifecycleConfig())

	svc1 := &mockService{name: "svc1"}
	svc2 := &mockService{name: "svc2"}
	lm.Register(svc1, 1)
	lm.Register(svc2, 2)

	lm.StartAll()

	health := lm.HealthCheck()
	if len(health) != 2 {
		t.Fatalf("want 2 entries, got %d", len(health))
	}
	if !health["svc1"] || !health["svc2"] {
		t.Fatalf("all services should be healthy: %v", health)
	}
}

func TestPriorityOrder(t *testing.T) {
	lm := NewLifecycleManager(DefaultLifecycleConfig())

	// Reset global counter for this test.
	seqMu.Lock()
	seqCounter = 0
	seqMu.Unlock()

	low := &orderedMockService{name: "low"}   // priority 10
	mid := &orderedMockService{name: "mid"}   // priority 5
	high := &orderedMockService{name: "high"} // priority 1

	// Register in non-sorted order.
	lm.Register(low, 10)
	lm.Register(high, 1)
	lm.Register(mid, 5)

	lm.StartAll()

	// Lower priority value should start first.
	if high.startSeq > mid.startSeq || mid.startSeq > low.startSeq {
		t.Fatalf("start order wrong: high=%d, mid=%d, low=%d",
			high.startSeq, mid.startSeq, low.startSeq)
	}
}

func TestStopError(t *testing.T) {
	lm := NewLifecycleManager(DefaultLifecycleConfig())

	svc := &mockService{name: "broken", stopErr: errors.New("stop failure")}
	lm.Register(svc, 1)
	lm.StartAll()

	errs := lm.StopAll()
	if len(errs) != 1 {
		t.Fatalf("want 1 error, got %d", len(errs))
	}
	if lm.GetState("broken") != StateFailed {
		t.Fatal("service should be in failed state after stop error")
	}
}

func TestStopAll_SkipsNonRunning(t *testing.T) {
	lm := NewLifecycleManager(DefaultLifecycleConfig())

	// Register two services but only start one.
	svc1 := &mockService{name: "running"}
	svc2 := &mockService{name: "not-running", startErr: errors.New("start failed")}
	lm.Register(svc1, 1)
	lm.Register(svc2, 2)

	lm.StartAll() // svc2 will be in StateFailed after this

	errs := lm.StopAll()
	if len(errs) != 0 {
		t.Fatalf("unexpected stop errors: %v", errs)
	}
	// svc1 should be stopped, svc2 stays failed (was never running).
	if lm.GetState("running") != StateStopped {
		t.Fatalf("want StateStopped for running svc, got %v", lm.GetState("running"))
	}
	if lm.GetState("not-running") != StateFailed {
		t.Fatalf("want StateFailed for non-started svc, got %v", lm.GetState("not-running"))
	}
}

func TestServiceStateString(t *testing.T) {
	tests := []struct {
		state ServiceState
		want  string
	}{
		{StateCreated, "created"},
		{StateStarting, "starting"},
		{StateRunning, "running"},
		{StateStopping, "stopping"},
		{StateStopped, "stopped"},
		{StateFailed, "failed"},
		{ServiceState(99), "unknown"},
	}
	for _, tc := range tests {
		got := tc.state.String()
		if got != tc.want {
			t.Errorf("state %d: want %q, got %q", tc.state, tc.want, got)
		}
	}
}

func TestDefaultLifecycleConfig(t *testing.T) {
	cfg := DefaultLifecycleConfig()
	if cfg.ShutdownTimeout != 30*time.Second {
		t.Errorf("want ShutdownTimeout=30s, got %v", cfg.ShutdownTimeout)
	}
	if cfg.GracePeriod != 5*time.Second {
		t.Errorf("want GracePeriod=5s, got %v", cfg.GracePeriod)
	}
	if cfg.MaxServices != 32 {
		t.Errorf("want MaxServices=32, got %d", cfg.MaxServices)
	}
}
