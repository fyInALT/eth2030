package sync

import (
	"testing"
	"time"
)

// TestConcurrentHealerBudget verifies that Budget() returns the internal
// resource budget (non-nil and functional).
func TestConcurrentHealerBudget(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	ch := NewConcurrentHealer(cfg)

	b := ch.Budget()
	if b == nil {
		t.Fatal("expected non-nil budget")
	}

	// Budget should reflect the configured memory limit.
	if b.memLimit != cfg.MemoryLimit {
		t.Fatalf("budget memLimit: got %d, want %d", b.memLimit, cfg.MemoryLimit)
	}
	if b.maxPending != cfg.MaxPending {
		t.Fatalf("budget maxPending: got %d, want %d", b.maxPending, cfg.MaxPending)
	}
}

// TestConcurrentHealerBudget_TracksCost verifies that Budget() reflects
// memory usage changes as tasks are scheduled and completed.
func TestConcurrentHealerBudget_TracksCost(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	cfg.MaxWorkers = 1
	ch := NewConcurrentHealer(cfg)

	hash := [32]byte{0x42}
	task := SchedulerHealTask{
		NodeHash:      hash,
		Priority:      PriorityNormal,
		EstimatedCost: 1024,
		CreatedAt:     time.Now(),
	}
	if err := ch.ScheduleTask(task); err != nil {
		t.Fatalf("ScheduleTask: %v", err)
	}

	b := ch.Budget()
	if b.MemoryUsed() != 1024 {
		t.Fatalf("expected 1024 memory used after schedule, got %d", b.MemoryUsed())
	}

	ch.ProcessBatch()
	ch.CompleteTaskWithCost(hash, 1024)

	if b.MemoryUsed() != 0 {
		t.Fatalf("expected 0 memory used after complete, got %d", b.MemoryUsed())
	}
}

// TestConcurrentHealerBudget_Utilization verifies Utilization via Budget().
func TestConcurrentHealerBudget_Utilization(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	cfg.MemoryLimit = 1000
	ch := NewConcurrentHealer(cfg)

	ch.ScheduleTask(SchedulerHealTask{
		NodeHash:      [32]byte{0x01},
		Priority:      PriorityNormal,
		EstimatedCost: 500,
		CreatedAt:     time.Now(),
	})

	util := ch.Budget().Utilization()
	if util < 0.49 || util > 0.51 {
		t.Fatalf("expected utilization ~0.5, got %f", util)
	}
}

// TestConcurrentHealerFailTask verifies FailTask removes the node from
// the active set without panicking.
func TestConcurrentHealerFailTask(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	cfg.MaxWorkers = 1
	ch := NewConcurrentHealer(cfg)

	hash := [32]byte{0x77}
	ch.ScheduleTask(SchedulerHealTask{
		NodeHash:      hash,
		Priority:      PriorityNormal,
		EstimatedCost: 10,
		CreatedAt:     time.Now(),
	})
	ch.ProcessBatch()

	if ch.ActiveCount() != 1 {
		t.Fatalf("expected 1 active before FailTask, got %d", ch.ActiveCount())
	}

	ch.FailTask(hash)

	if ch.ActiveCount() != 0 {
		t.Fatalf("expected 0 active after FailTask, got %d", ch.ActiveCount())
	}
}

// TestConcurrentHealerFailTask_NonActive verifies FailTask is safe when called
// with a hash that is not in the active set.
func TestConcurrentHealerFailTask_NonActive(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	ch := NewConcurrentHealer(cfg)

	// Should not panic.
	ch.FailTask([32]byte{0xDE, 0xAD})
}

// TestResourceBudgetCanSchedule_PendingLimit verifies CanSchedule returns
// false when the pending limit is hit, exercising the branch uncovered
// by the existing test.
func TestResourceBudgetCanSchedule_PendingLimit(t *testing.T) {
	rb := NewResourceBudget(0, 0, 1)

	// Fill the pending slot.
	if err := rb.Reserve(0); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	// Now pending is full; CanSchedule should be false.
	if rb.CanSchedule(0) {
		t.Fatal("expected CanSchedule to return false when pending limit hit")
	}
}

// TestResourceBudgetUtilization_ZeroLimit verifies Utilization with zero
// memory limit returns 0 (the uncovered branch).
func TestResourceBudgetUtilization_ZeroLimit(t *testing.T) {
	rb := NewResourceBudget(0, 0, 0)
	if rb.Utilization() != 0 {
		t.Fatalf("expected 0 utilization with zero memLimit, got %f", rb.Utilization())
	}
}

// TestConcurrentHealerScheduleTask_BudgetFull exercises the ErrSchedulerFull
// path when CanSchedule returns false due to a full memory budget.
func TestConcurrentHealerScheduleTask_BudgetFull(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	cfg.MemoryLimit = 50
	cfg.MaxPending = 100
	ch := NewConcurrentHealer(cfg)

	// Fill memory budget.
	ch.ScheduleTask(SchedulerHealTask{
		NodeHash:      [32]byte{0x01},
		Priority:      PriorityNormal,
		EstimatedCost: 50,
		CreatedAt:     time.Now(),
	})

	// Next task exceeds memory limit.
	err := ch.ScheduleTask(SchedulerHealTask{
		NodeHash:      [32]byte{0x02},
		Priority:      PriorityNormal,
		EstimatedCost: 10,
		CreatedAt:     time.Now(),
	})
	if err != ErrSchedulerFull {
		t.Fatalf("expected ErrSchedulerFull, got %v", err)
	}
}

// TestConcurrentHealerScheduleTask_DuplicateActive verifies that scheduling
// a task whose hash is already in the active set returns ErrDuplicateTask.
func TestConcurrentHealerScheduleTask_DuplicateActive(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	cfg.MaxWorkers = 1
	ch := NewConcurrentHealer(cfg)

	hash := [32]byte{0x99}
	ch.ScheduleTask(SchedulerHealTask{
		NodeHash:      hash,
		Priority:      PriorityNormal,
		EstimatedCost: 10,
		CreatedAt:     time.Now(),
	})
	ch.ProcessBatch() // moves to active

	err := ch.ScheduleTask(SchedulerHealTask{
		NodeHash:      hash,
		Priority:      PriorityNormal,
		EstimatedCost: 10,
		CreatedAt:     time.Now(),
	})
	if err != ErrDuplicateTask {
		t.Fatalf("expected ErrDuplicateTask for active hash, got %v", err)
	}
}

// TestConcurrentHealerScheduleTask_DuplicateDone verifies that scheduling
// a task whose hash is already in the done set returns ErrDuplicateTask.
func TestConcurrentHealerScheduleTask_DuplicateDone(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	cfg.MaxWorkers = 1
	ch := NewConcurrentHealer(cfg)

	hash := [32]byte{0x88}
	ch.ScheduleTask(SchedulerHealTask{
		NodeHash:      hash,
		Priority:      PriorityNormal,
		EstimatedCost: 10,
		CreatedAt:     time.Now(),
	})
	ch.ProcessBatch()
	ch.CompleteTask(hash) // moves to done

	err := ch.ScheduleTask(SchedulerHealTask{
		NodeHash:      hash,
		Priority:      PriorityNormal,
		EstimatedCost: 10,
		CreatedAt:     time.Now(),
	})
	if err != ErrDuplicateTask {
		t.Fatalf("expected ErrDuplicateTask for done hash, got %v", err)
	}
}
