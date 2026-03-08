package epbs

import (
	"testing"
)

func TestProcessBuilderPendingPayments(t *testing.T) {
	state := NewBuilderEpochState()

	// Add 3 pending payments.
	state.AddPendingPayment(BuilderIndex(1), 1000)
	state.AddPendingPayment(BuilderIndex(2), 2000)
	state.AddPendingPayment(BuilderIndex(3), 500)

	// Set balances.
	state.SetBuilderBalance(BuilderIndex(1), 5000)
	state.SetBuilderBalance(BuilderIndex(2), 1500) // less than payment (2000) → capped
	state.SetBuilderBalance(BuilderIndex(3), 5000)

	ProcessBuilderPendingPayments(state)

	// Builder 1: 5000 - 1000 = 4000.
	if bal := state.GetBuilderBalance(BuilderIndex(1)); bal != 4000 {
		t.Errorf("builder 1 balance: got %d, want 4000", bal)
	}
	// Builder 2: capped at balance (1500).
	if bal := state.GetBuilderBalance(BuilderIndex(2)); bal != 0 {
		t.Errorf("builder 2 balance: got %d, want 0 (payment capped)", bal)
	}
	// Builder 3: 5000 - 500 = 4500.
	if bal := state.GetBuilderBalance(BuilderIndex(3)); bal != 4500 {
		t.Errorf("builder 3 balance: got %d, want 4500", bal)
	}

	// Payments cleared after processing.
	if len(state.PendingPayments()) != 0 {
		t.Error("pending payments should be cleared after processing")
	}
}

func TestProcessBuilderPendingPaymentsEmpty(t *testing.T) {
	state := NewBuilderEpochState()
	// No payments — should not panic.
	ProcessBuilderPendingPayments(state)
}

func TestBuilderPendingPayments_SameBuilderTwice(t *testing.T) {
	state := NewBuilderEpochState()
	state.SetBuilderBalance(BuilderIndex(1), 5000)
	state.AddPendingPayment(BuilderIndex(1), 1000)
	state.AddPendingPayment(BuilderIndex(1), 2000)

	ProcessBuilderPendingPayments(state)
	// 5000 - 1000 - 2000 = 2000
	if bal := state.GetBuilderBalance(BuilderIndex(1)); bal != 2000 {
		t.Errorf("balance after two payments: got %d, want 2000", bal)
	}
}

func TestBuilderPendingPayments_ZeroPayment(t *testing.T) {
	state := NewBuilderEpochState()
	state.SetBuilderBalance(BuilderIndex(1), 5000)
	state.AddPendingPayment(BuilderIndex(1), 0)
	ProcessBuilderPendingPayments(state)
	if bal := state.GetBuilderBalance(BuilderIndex(1)); bal != 5000 {
		t.Errorf("zero payment changed balance: got %d", bal)
	}
}

func TestBuilderPendingPayments_ExactBalance(t *testing.T) {
	state := NewBuilderEpochState()
	state.SetBuilderBalance(BuilderIndex(1), 1000)
	state.AddPendingPayment(BuilderIndex(1), 1000)
	ProcessBuilderPendingPayments(state)
	if bal := state.GetBuilderBalance(BuilderIndex(1)); bal != 0 {
		t.Errorf("exact payment: expected 0, got %d", bal)
	}
}
