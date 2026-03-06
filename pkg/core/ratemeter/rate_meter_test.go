package ratemeter

import "testing"

func TestDefaultRateMeterConfig(t *testing.T) {
	cfg := DefaultRateMeterConfig()
	if cfg.WindowSize != 64 {
		t.Errorf("WindowSize = %d, want 64", cfg.WindowSize)
	}
	if cfg.TargetGasPerSec != 1_000_000_000 {
		t.Errorf("TargetGasPerSec = %g, want 1e9", cfg.TargetGasPerSec)
	}
}

func TestRateMeter_RecordAndRate(t *testing.T) {
	rm := NewRateMeter(DefaultRateMeterConfig())

	// No records yet.
	if rm.RecordCount() != 0 {
		t.Errorf("RecordCount = %d, want 0", rm.RecordCount())
	}
	if rm.CurrentRate() != 0 {
		t.Errorf("CurrentRate = %g, want 0", rm.CurrentRate())
	}

	rm.RecordBlock(1, 1_000_000, 1000)
	rm.RecordBlock(2, 2_000_000, 1002)

	if rm.RecordCount() != 2 {
		t.Errorf("RecordCount = %d, want 2", rm.RecordCount())
	}
	if rm.CurrentRate() == 0 {
		t.Error("CurrentRate should be non-zero after 2 records")
	}
}

func TestRateMeter_Reset(t *testing.T) {
	rm := NewRateMeter(DefaultRateMeterConfig())
	rm.RecordBlock(1, 1_000_000, 1000)
	rm.RecordBlock(2, 1_000_000, 1001)
	rm.Reset()

	if rm.RecordCount() != 0 {
		t.Errorf("RecordCount after Reset = %d, want 0", rm.RecordCount())
	}
	if rm.CurrentRate() != 0 {
		t.Errorf("CurrentRate after Reset = %g, want 0", rm.CurrentRate())
	}
}

func TestRateMeter_WindowSize(t *testing.T) {
	cfg := DefaultRateMeterConfig()
	cfg.WindowSize = 4
	rm := NewRateMeter(cfg)

	for i := uint64(0); i < 10; i++ {
		rm.RecordBlock(i, 1_000_000, i*1000)
	}
	if rm.RecordCount() > 4 {
		t.Errorf("RecordCount = %d, should be capped at 4", rm.RecordCount())
	}
}
