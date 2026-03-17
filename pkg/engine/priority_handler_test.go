package engine

import (
	"testing"
)

// TestMethodPriority tests the MethodPriority function.
func TestMethodPriority(t *testing.T) {
	tests := []struct {
		method   string
		expected RequestPriority
	}{
		// High priority methods
		{"engine_forkchoiceUpdatedV1", PriorityHigh},
		{"engine_forkchoiceUpdatedV2", PriorityHigh},
		{"engine_forkchoiceUpdatedV3", PriorityHigh},
		{"engine_getPayloadV1", PriorityHigh},
		{"engine_getPayloadV3", PriorityHigh},
		{"engine_getPayloadV7", PriorityHigh},

		// Normal priority methods
		{"engine_newPayloadV1", PriorityNormal},
		{"engine_newPayloadV3", PriorityNormal},
		{"engine_newPayloadV5", PriorityNormal},

		// Low priority methods
		{"engine_getPayloadBodiesByHashV1", PriorityLow},
		{"engine_getPayloadBodiesByRangeV1", PriorityLow},
		{"engine_getPayloadBodiesByHashV2", PriorityLow},

		// Unknown methods default to Normal
		{"engine_unknownMethod", PriorityNormal},
		{"eth_getBlockByNumber", PriorityNormal},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			got := MethodPriority(tt.method)
			if got != tt.expected {
				t.Errorf("MethodPriority(%q) = %v, want %v", tt.method, got, tt.expected)
			}
		})
	}
}

// TestRequestPriority_String tests the String method.
func TestRequestPriority_String(t *testing.T) {
	tests := []struct {
		priority RequestPriority
		expected string
	}{
		{PriorityHigh, "high"},
		{PriorityNormal, "normal"},
		{PriorityLow, "low"},
		{RequestPriority(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.priority.String()
			if got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestPriorityHandler_StartEndRequest tests StartRequest and EndRequest.
func TestPriorityHandler_StartEndRequest(t *testing.T) {
	h := NewPriorityHandler()

	// Start a high priority request
	h.StartRequest(PriorityHigh)

	stats := h.GetStats()
	if stats.HighTotal != 1 {
		t.Errorf("HighTotal = %d, want 1", stats.HighTotal)
	}
	if stats.HighActive != 1 {
		t.Errorf("HighActive = %d, want 1", stats.HighActive)
	}

	// End the request
	h.EndRequest(PriorityHigh)

	stats = h.GetStats()
	if stats.HighTotal != 1 {
		t.Errorf("HighTotal = %d, want 1 (should not change)", stats.HighTotal)
	}
	if stats.HighActive != 0 {
		t.Errorf("HighActive = %d, want 0", stats.HighActive)
	}
}

// TestPriorityHandler_MultipleRequests tests handling multiple concurrent requests.
func TestPriorityHandler_MultipleRequests(t *testing.T) {
	h := NewPriorityHandler()

	// Start multiple requests of different priorities
	h.StartRequest(PriorityHigh)
	h.StartRequest(PriorityHigh)
	h.StartRequest(PriorityNormal)
	h.StartRequest(PriorityLow)

	stats := h.GetStats()
	if stats.HighActive != 2 {
		t.Errorf("HighActive = %d, want 2", stats.HighActive)
	}
	if stats.NormalActive != 1 {
		t.Errorf("NormalActive = %d, want 1", stats.NormalActive)
	}
	if stats.LowActive != 1 {
		t.Errorf("LowActive = %d, want 1", stats.LowActive)
	}

	// End all requests
	h.EndRequest(PriorityHigh)
	h.EndRequest(PriorityHigh)
	h.EndRequest(PriorityNormal)
	h.EndRequest(PriorityLow)

	stats = h.GetStats()
	if stats.HighActive != 0 || stats.NormalActive != 0 || stats.LowActive != 0 {
		t.Errorf("all active counts should be 0, got high=%d normal=%d low=%d",
			stats.HighActive, stats.NormalActive, stats.LowActive)
	}
}

// TestPriorityHandler_ShouldPreempt tests the preemption threshold.
func TestPriorityHandler_ShouldPreempt(t *testing.T) {
	h := NewPriorityHandler()

	// Below threshold
	h.StartRequest(PriorityHigh)
	if h.ShouldPreempt() {
		t.Error("ShouldPreempt should be false with 1 high request")
	}

	// At threshold (2)
	h.StartRequest(PriorityHigh)
	if !h.ShouldPreempt() {
		t.Error("ShouldPreempt should be true with 2 high requests")
	}

	// Above threshold
	h.StartRequest(PriorityHigh)
	if !h.ShouldPreempt() {
		t.Error("ShouldPreempt should be true with 3 high requests")
	}

	// Clean up
	h.EndRequest(PriorityHigh)
	h.EndRequest(PriorityHigh)
	h.EndRequest(PriorityHigh)
}

// TestPriorityHandler_NegativeActive tests that active counts don't go negative.
func TestPriorityHandler_NegativeActive(t *testing.T) {
	h := NewPriorityHandler()

	// End without start (should be idempotent)
	h.EndRequest(PriorityHigh)

	stats := h.GetStats()
	if stats.HighActive < 0 {
		t.Errorf("HighActive = %d, should not be negative", stats.HighActive)
	}
}
