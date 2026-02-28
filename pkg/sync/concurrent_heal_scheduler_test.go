package sync

import (
	"testing"
	"time"
)

func TestConcurrentHealerScheduleAndProcess(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	cfg.MaxWorkers = 2
	ch := NewConcurrentHealer(cfg)

	task := SchedulerHealTask{
		NodeHash:      [32]byte{0x01},
		Path:          []byte{0x0a},
		Priority:      PriorityNormal,
		EstimatedCost: 100,
		CreatedAt:     time.Now(),
	}
	if err := ch.ScheduleTask(task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.PendingCount() != 1 {
		t.Fatalf("expected 1 pending, got %d", ch.PendingCount())
	}

	batch := ch.ProcessBatch()
	if len(batch) != 1 {
		t.Fatalf("expected batch of 1, got %d", len(batch))
	}
	if ch.ActiveCount() != 1 {
		t.Fatalf("expected 1 active, got %d", ch.ActiveCount())
	}
}

func TestConcurrentHealerPriorityOrdering(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	cfg.MaxWorkers = 4
	ch := NewConcurrentHealer(cfg)

	now := time.Now()
	tasks := []SchedulerHealTask{
		{NodeHash: [32]byte{0x01}, Priority: PriorityBackground, CreatedAt: now, EstimatedCost: 10},
		{NodeHash: [32]byte{0x02}, Priority: PriorityCritical, CreatedAt: now, EstimatedCost: 10},
		{NodeHash: [32]byte{0x03}, Priority: PriorityNormal, CreatedAt: now, EstimatedCost: 10},
		{NodeHash: [32]byte{0x04}, Priority: PriorityUrgent, CreatedAt: now, EstimatedCost: 10},
	}
	for _, task := range tasks {
		if err := ch.ScheduleTask(task); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	batch := ch.ProcessBatch()
	if len(batch) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(batch))
	}
	// First task should be Critical priority.
	if batch[0].Priority != PriorityCritical {
		t.Fatalf("expected first task to be Critical, got %d", batch[0].Priority)
	}
	// Second should be Urgent.
	if batch[1].Priority != PriorityUrgent {
		t.Fatalf("expected second task to be Urgent, got %d", batch[1].Priority)
	}
}

func TestConcurrentHealerCompleteTask(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	cfg.MaxWorkers = 1
	ch := NewConcurrentHealer(cfg)

	hash := [32]byte{0xaa}
	ch.ScheduleTask(SchedulerHealTask{
		NodeHash: hash, Priority: PriorityNormal, EstimatedCost: 50, CreatedAt: time.Now(),
	})
	ch.ProcessBatch()

	if ch.ActiveCount() != 1 {
		t.Fatalf("expected 1 active, got %d", ch.ActiveCount())
	}

	ch.CompleteTask(hash)
	if ch.ActiveCount() != 0 {
		t.Fatalf("expected 0 active after complete, got %d", ch.ActiveCount())
	}
	if ch.CompletedCount() != 1 {
		t.Fatalf("expected 1 completed, got %d", ch.CompletedCount())
	}
}

func TestConcurrentHealerFailTaskWithRetry(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	cfg.RetryLimit = 2
	cfg.MaxWorkers = 1
	ch := NewConcurrentHealer(cfg)

	task := SchedulerHealTask{
		NodeHash: [32]byte{0xbb}, Priority: PriorityNormal, EstimatedCost: 50, CreatedAt: time.Now(),
	}
	ch.ScheduleTask(task)

	batch := ch.ProcessBatch()
	if len(batch) != 1 {
		t.Fatalf("expected 1 task in batch")
	}

	// First failure should re-queue.
	requeued := ch.FailTaskWithRetry(batch[0])
	if !requeued {
		t.Fatal("expected task to be requeued on first failure")
	}
	if ch.PendingCount() != 1 {
		t.Fatalf("expected 1 pending after requeue, got %d", ch.PendingCount())
	}

	// Process and fail again — should exceed retry limit.
	batch = ch.ProcessBatch()
	requeued = ch.FailTaskWithRetry(batch[0])
	if requeued {
		t.Fatal("expected task to be dropped after exceeding retry limit")
	}
}

func TestConcurrentHealerDuplicateTask(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	ch := NewConcurrentHealer(cfg)

	task := SchedulerHealTask{
		NodeHash: [32]byte{0xcc}, Priority: PriorityNormal, EstimatedCost: 10, CreatedAt: time.Now(),
	}
	if err := ch.ScheduleTask(task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err := ch.ScheduleTask(task)
	if err != ErrDuplicateTask {
		t.Fatalf("expected ErrDuplicateTask, got %v", err)
	}
}

func TestConcurrentHealerClosedSchedule(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	ch := NewConcurrentHealer(cfg)
	ch.Close()

	err := ch.ScheduleTask(SchedulerHealTask{
		NodeHash: [32]byte{0xdd}, EstimatedCost: 10, CreatedAt: time.Now(),
	})
	if err != ErrSchedulerClosed {
		t.Fatalf("expected ErrSchedulerClosed, got %v", err)
	}
}

func TestConcurrentHealerClosedProcessBatch(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	ch := NewConcurrentHealer(cfg)
	ch.Close()

	batch := ch.ProcessBatch()
	if batch != nil {
		t.Fatal("expected nil batch from closed scheduler")
	}
}

func TestResourceBudgetCanSchedule(t *testing.T) {
	rb := NewResourceBudget(100, 0, 10)

	if !rb.CanSchedule(50) {
		t.Fatal("expected CanSchedule to return true for 50/100")
	}
	if rb.CanSchedule(200) {
		t.Fatal("expected CanSchedule to return false for 200/100")
	}
}

func TestResourceBudgetReserveRelease(t *testing.T) {
	rb := NewResourceBudget(100, 0, 5)

	if err := rb.Reserve(40); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rb.MemoryUsed() != 40 {
		t.Fatalf("expected 40 memory used, got %d", rb.MemoryUsed())
	}

	rb.Release(40)
	if rb.MemoryUsed() != 0 {
		t.Fatalf("expected 0 memory used after release, got %d", rb.MemoryUsed())
	}
}

func TestResourceBudgetExceedMemory(t *testing.T) {
	rb := NewResourceBudget(50, 0, 100)

	if err := rb.Reserve(30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err := rb.Reserve(30) // 30 + 30 = 60 > 50
	if err == nil {
		t.Fatal("expected error for exceeding memory budget")
	}
}

func TestResourceBudgetExceedPending(t *testing.T) {
	rb := NewResourceBudget(1000, 0, 2)

	rb.Reserve(1)
	rb.Reserve(1)
	err := rb.Reserve(1) // 3rd task, limit is 2
	if err == nil {
		t.Fatal("expected error for exceeding pending limit")
	}
}

func TestResourceBudgetBandwidth(t *testing.T) {
	rb := NewResourceBudget(1000, 100, 100)

	if err := rb.ReserveBandwidth(50); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err := rb.ReserveBandwidth(60) // 50 + 60 = 110 > 100
	if err == nil {
		t.Fatal("expected error for exceeding bandwidth budget")
	}
	rb.ReleaseBandwidth(50)
	if err := rb.ReserveBandwidth(60); err != nil {
		t.Fatalf("unexpected error after release: %v", err)
	}
}

func TestResourceBudgetUtilization(t *testing.T) {
	rb := NewResourceBudget(200, 0, 100)
	rb.Reserve(100)

	util := rb.Utilization()
	if util < 0.49 || util > 0.51 {
		t.Fatalf("expected utilization ~0.5, got %f", util)
	}
}

func TestResourceBudgetZeroLimit(t *testing.T) {
	rb := NewResourceBudget(0, 0, 0) // unlimited

	if !rb.CanSchedule(999999) {
		t.Fatal("expected unlimited budget to allow any cost")
	}
	if err := rb.Reserve(999999); err != nil {
		t.Fatalf("unexpected error with unlimited budget: %v", err)
	}
}

func TestConcurrentHealerExpiredTask(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	cfg.MaxWorkers = 2
	ch := NewConcurrentHealer(cfg)

	// Add an expired task and a valid task.
	expired := SchedulerHealTask{
		NodeHash:      [32]byte{0x01},
		Priority:      PriorityCritical,
		Deadline:      time.Now().Add(-time.Hour),
		EstimatedCost: 10,
		CreatedAt:     time.Now(),
	}
	valid := SchedulerHealTask{
		NodeHash:      [32]byte{0x02},
		Priority:      PriorityNormal,
		EstimatedCost: 10,
		CreatedAt:     time.Now(),
	}
	ch.ScheduleTask(expired)
	ch.ScheduleTask(valid)

	batch := ch.ProcessBatch()
	// Expired task should be dropped.
	if len(batch) != 1 {
		t.Fatalf("expected 1 task (expired dropped), got %d", len(batch))
	}
	if batch[0].NodeHash != [32]byte{0x02} {
		t.Fatal("expected the valid task to remain")
	}
}

func TestSchedulerHealTaskIsExpired(t *testing.T) {
	task := SchedulerHealTask{Deadline: time.Now().Add(-time.Second)}
	if !task.IsExpired() {
		t.Fatal("expected task to be expired")
	}

	task2 := SchedulerHealTask{Deadline: time.Now().Add(time.Hour)}
	if task2.IsExpired() {
		t.Fatal("expected task to not be expired")
	}

	task3 := SchedulerHealTask{} // zero deadline
	if task3.IsExpired() {
		t.Fatal("expected zero-deadline task to not be expired")
	}
}

func TestConcurrentHealerIsClosed(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	ch := NewConcurrentHealer(cfg)

	if ch.IsClosed() {
		t.Fatal("expected not closed initially")
	}
	ch.Close()
	if !ch.IsClosed() {
		t.Fatal("expected closed after Close()")
	}
}

func TestConcurrentHealerEmptyProcessBatch(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	ch := NewConcurrentHealer(cfg)

	batch := ch.ProcessBatch()
	if batch != nil {
		t.Fatal("expected nil batch from empty scheduler")
	}
}

func TestConcurrentHealerCompleteTaskWithCost(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	cfg.MaxWorkers = 1
	ch := NewConcurrentHealer(cfg)

	hash := [32]byte{0xee}
	ch.ScheduleTask(SchedulerHealTask{
		NodeHash: hash, Priority: PriorityNormal, EstimatedCost: 200, CreatedAt: time.Now(),
	})
	ch.ProcessBatch()

	ch.CompleteTaskWithCost(hash, 200)
	if ch.CompletedCount() != 1 {
		t.Fatalf("expected 1 completed, got %d", ch.CompletedCount())
	}
	if ch.ActiveCount() != 0 {
		t.Fatalf("expected 0 active, got %d", ch.ActiveCount())
	}
}

func TestDefaultHealSchedulerConfig(t *testing.T) {
	cfg := DefaultHealSchedulerConfig()
	if cfg.MaxWorkers <= 0 {
		t.Fatal("expected positive MaxWorkers")
	}
	if cfg.MaxPending <= 0 {
		t.Fatal("expected positive MaxPending")
	}
	if cfg.MemoryLimit <= 0 {
		t.Fatal("expected positive MemoryLimit")
	}
	if cfg.RetryLimit <= 0 {
		t.Fatal("expected positive RetryLimit")
	}
}

func TestResourceBudgetReleaseUnderflow(t *testing.T) {
	rb := NewResourceBudget(100, 100, 100)

	// Release without reservation should clamp to 0, not go negative.
	rb.Release(50)
	if rb.MemoryUsed() != 0 {
		t.Fatalf("expected 0 memory after underflow release, got %d", rb.MemoryUsed())
	}
	if rb.PendingUsed() != 0 {
		t.Fatalf("expected 0 pending after underflow release, got %d", rb.PendingUsed())
	}

	rb.ReleaseBandwidth(50)
	// Should not panic, bandwidth should clamp to 0.
}
