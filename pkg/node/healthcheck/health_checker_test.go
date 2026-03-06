package healthcheck

import "testing"

type alwaysHealthy struct{}

func (a *alwaysHealthy) Check() *SubsystemHealth {
	return &SubsystemHealth{Status: StatusHealthy, Message: "ok"}
}

type alwaysDegraded struct{}

func (a *alwaysDegraded) Check() *SubsystemHealth {
	return &SubsystemHealth{Status: StatusDegraded, Message: "degraded"}
}

func TestHealthChecker_AllHealthy(t *testing.T) {
	hc := NewHealthChecker()
	hc.RegisterSubsystem("svc1", &alwaysHealthy{})
	hc.RegisterSubsystem("svc2", &alwaysHealthy{})

	report := hc.CheckAll()
	if report.OverallStatus != StatusHealthy {
		t.Errorf("overall = %q, want %q", report.OverallStatus, StatusHealthy)
	}
	if len(report.Subsystems) != 2 {
		t.Errorf("subsystems = %d, want 2", len(report.Subsystems))
	}
}

func TestHealthChecker_Degraded(t *testing.T) {
	hc := NewHealthChecker()
	hc.RegisterSubsystem("healthy", &alwaysHealthy{})
	hc.RegisterSubsystem("degraded", &alwaysDegraded{})

	report := hc.CheckAll()
	if report.OverallStatus != StatusDegraded {
		t.Errorf("overall = %q, want %q", report.OverallStatus, StatusDegraded)
	}
}

func TestHealthChecker_IsHealthy(t *testing.T) {
	hc := NewHealthChecker()
	hc.RegisterSubsystem("ok", &alwaysHealthy{})
	if !hc.IsHealthy() {
		t.Error("expected IsHealthy() to return true")
	}
}

func TestHealthChecker_SubsystemCount(t *testing.T) {
	hc := NewHealthChecker()
	if hc.SubsystemCount() != 0 {
		t.Errorf("initial count = %d, want 0", hc.SubsystemCount())
	}
	hc.RegisterSubsystem("a", &alwaysHealthy{})
	hc.RegisterSubsystem("b", &alwaysHealthy{})
	if hc.SubsystemCount() != 2 {
		t.Errorf("count = %d, want 2", hc.SubsystemCount())
	}
}
