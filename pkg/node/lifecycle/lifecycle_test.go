package lifecycle

import (
	"errors"
	"testing"
)

type mockSvc struct {
	name    string
	startFn func() error
	stopFn  func() error
}

func (m *mockSvc) Start() error {
	if m.startFn != nil {
		return m.startFn()
	}
	return nil
}

func (m *mockSvc) Stop() error {
	if m.stopFn != nil {
		return m.stopFn()
	}
	return nil
}

func (m *mockSvc) Name() string { return m.name }

func TestRegisterAndStart(t *testing.T) {
	lm := NewLifecycleManager(DefaultLifecycleConfig())

	svc := &mockSvc{name: "svc1"}
	if err := lm.Register(svc, 0); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if lm.ServiceCount() != 1 {
		t.Errorf("ServiceCount = %d, want 1", lm.ServiceCount())
	}

	errs := lm.StartAll()
	if len(errs) != 0 {
		t.Fatalf("StartAll errors: %v", errs)
	}

	if lm.GetState("svc1") != StateRunning {
		t.Errorf("state = %v, want Running", lm.GetState("svc1"))
	}
}

func TestStartError(t *testing.T) {
	lm := NewLifecycleManager(DefaultLifecycleConfig())
	startErr := errors.New("start failed")
	svc := &mockSvc{name: "broken", startFn: func() error { return startErr }}

	if err := lm.Register(svc, 0); err != nil {
		t.Fatalf("Register: %v", err)
	}

	errs := lm.StartAll()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if lm.GetState("broken") != StateFailed {
		t.Errorf("state = %v, want Failed", lm.GetState("broken"))
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
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("state %d String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestDefaultLifecycleConfig(t *testing.T) {
	cfg := DefaultLifecycleConfig()
	if cfg.MaxServices != 32 {
		t.Errorf("MaxServices = %d, want 32", cfg.MaxServices)
	}
}
