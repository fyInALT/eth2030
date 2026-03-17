package payload

import (
	"math/big"
	"testing"
	"time"

	"github.com/eth2030/eth2030/core/types"
)

// TestPendingPayload_StatusTransition tests the status transitions.
func TestPendingPayload_StatusTransition(t *testing.T) {
	id := PayloadID{0x01}
	pending := NewPendingPayload(id, types.Hash{}, 1234, types.Address{}, types.Hash{}, nil)

	// Initial status should be Pending
	if pending.Status() != BuildStatusPending {
		t.Fatalf("want Pending, got %v", pending.Status())
	}

	// Transition to Building
	pending.SetBuilding()
	if pending.Status() != BuildStatusBuilding {
		t.Fatalf("want Building, got %v", pending.Status())
	}

	// Cannot transition back to Pending (status is terminal after Completed/Failed)
}

// TestPendingPayload_SetCompleted tests marking a payload as completed.
func TestPendingPayload_SetCompleted(t *testing.T) {
	id := PayloadID{0x01}
	pending := NewPendingPayload(id, types.Hash{}, 1234, types.Address{}, types.Hash{}, nil)

	block := types.NewBlock(&types.Header{Number: big.NewInt(1)}, nil)
	receipts := []*types.Receipt{{}}

	pending.SetCompleted(block, receipts, nil, big.NewInt(100))

	if pending.Status() != BuildStatusCompleted {
		t.Fatalf("want Completed, got %v", pending.Status())
	}

	// Wait should return immediately
	result, err := pending.Wait(1 * time.Second)
	if err != nil {
		t.Fatalf("Wait error: %v", err)
	}
	if result.Status != BuildStatusCompleted {
		t.Fatalf("want Completed, got %v", result.Status)
	}
	if result.Block == nil {
		t.Fatal("Block should not be nil")
	}
}

// TestPendingPayload_SetFailed tests marking a payload as failed.
func TestPendingPayload_SetFailed(t *testing.T) {
	id := PayloadID{0x01}
	pending := NewPendingPayload(id, types.Hash{}, 1234, types.Address{}, types.Hash{}, nil)

	testErr := "test error"
	pending.SetFailed(nil)

	if pending.Status() != BuildStatusFailed {
		t.Fatalf("want Failed, got %v", pending.Status())
	}

	result, err := pending.Wait(1 * time.Second)
	if err != nil {
		t.Fatalf("Wait error: %v", err)
	}
	if result.Status != BuildStatusFailed {
		t.Fatalf("want Failed, got %v", result.Status)
	}
	_ = testErr // avoid unused variable warning
}

// TestPendingPayload_WaitTimeout tests that Wait times out for pending payloads.
func TestPendingPayload_WaitTimeout(t *testing.T) {
	id := PayloadID{0x01}
	pending := NewPendingPayload(id, types.Hash{}, 1234, types.Address{}, types.Hash{}, nil)

	start := time.Now()
	_, err := pending.Wait(100 * time.Millisecond)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("want timeout error, got nil")
	}
	if elapsed < 90*time.Millisecond {
		t.Fatalf("Wait returned too quickly: %v", elapsed)
	}
}

// TestPendingPayload_GetResult tests GetResult for various statuses.
func TestPendingPayload_GetResult(t *testing.T) {
	t.Run("pending_returns_error", func(t *testing.T) {
		pending := NewPendingPayload(PayloadID{0x01}, types.Hash{}, 1234, types.Address{}, types.Hash{}, nil)
		_, err := pending.GetResult()
		if err == nil {
			t.Fatal("want error for pending status, got nil")
		}
	})

	t.Run("completed_returns_result", func(t *testing.T) {
		pending := NewPendingPayload(PayloadID{0x01}, types.Hash{}, 1234, types.Address{}, types.Hash{}, nil)
		pending.SetCompleted(nil, nil, nil, nil)
		result, err := pending.GetResult()
		if err != nil {
			t.Fatalf("GetResult error: %v", err)
		}
		if result.Status != BuildStatusCompleted {
			t.Fatalf("want Completed, got %v", result.Status)
		}
	})
}

// TestPendingPayload_ReferenceCount tests the reference counting mechanism.
func TestPendingPayload_ReferenceCount(t *testing.T) {
	id := PayloadID{0x01}
	pending := NewPendingPayload(id, types.Hash{}, 1234, types.Address{}, types.Hash{}, nil)

	// Initial ref count should be 1 (from NewPendingPayload)
	if !pending.InUse() {
		t.Fatal("pending should be in use initially")
	}

	// Acquire should succeed
	if !pending.Acquire() {
		t.Fatal("Acquire should succeed")
	}

	// Now ref count is 2
	// Release once
	pending.Release()
	if !pending.InUse() {
		t.Fatal("pending should still be in use after one release")
	}

	// Release again
	pending.Release()
	if pending.InUse() {
		t.Fatal("pending should not be in use after all releases")
	}

	// Acquire should fail now (ref count is 0)
	if pending.Acquire() {
		t.Fatal("Acquire should fail after all references released")
	}
}

// TestAsyncBuilder_New tests creating an AsyncBuilder with defaults.
func TestAsyncBuilder_New(t *testing.T) {
	builder := NewAsyncBuilder(nil, nil, AsyncBuilderConfig{})
	if builder == nil {
		t.Fatal("NewAsyncBuilder returned nil")
	}
	if builder.workers != defaultWorkers {
		t.Fatalf("want workers=%d, got %d", defaultWorkers, builder.workers)
	}
	if builder.timeout != defaultTimeout {
		t.Fatalf("want timeout=%v, got %v", defaultTimeout, builder.timeout)
	}
}

// TestAsyncBuilder_GetPayload tests retrieving a payload.
func TestAsyncBuilder_GetPayload(t *testing.T) {
	builder := NewAsyncBuilder(nil, nil, AsyncBuilderConfig{Workers: 1})
	builder.Start()
	defer builder.Stop()

	id := PayloadID{0x01}

	// Non-existent payload
	_, ok := builder.GetPayload(id)
	if ok {
		t.Fatal("GetPayload should return false for non-existent payload")
	}
}

// TestAsyncBuilder_EvictOldest tests eviction of old payloads.
func TestAsyncBuilder_EvictOldest(t *testing.T) {
	builder := NewAsyncBuilder(nil, nil, AsyncBuilderConfig{
		Workers:    1,
		MaxPending: 2,
	})

	// Add 3 payloads manually
	for i := 0; i < 3; i++ {
		id := PayloadID{byte(i)}
		pending := NewPendingPayload(id, types.Hash{}, 1234, types.Address{}, types.Hash{}, nil)
		builder.mu.Lock()
		builder.pending[id] = pending
		builder.pendingOrder = append(builder.pendingOrder, id)
		builder.mu.Unlock()
	}

	builder.EvictOldest()

	builder.mu.RLock()
	count := len(builder.pending)
	builder.mu.RUnlock()

	if count > 2 {
		t.Fatalf("want at most 2 pending, got %d", count)
	}
}
