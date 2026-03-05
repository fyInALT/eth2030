package epbs

// BuilderPendingPayment represents a pending payment from a builder.
type BuilderPendingPayment struct {
	BuilderIdx BuilderIndex
	Amount     uint64
}

// BuilderEpochState holds per-epoch builder state for payment processing.
type BuilderEpochState struct {
	balances        map[BuilderIndex]uint64
	pendingPayments []BuilderPendingPayment
}

// NewBuilderEpochState creates a new epoch state.
func NewBuilderEpochState() *BuilderEpochState {
	return &BuilderEpochState{
		balances: make(map[BuilderIndex]uint64),
	}
}

// SetBuilderBalance sets the balance for a builder.
func (s *BuilderEpochState) SetBuilderBalance(idx BuilderIndex, balance uint64) {
	s.balances[idx] = balance
}

// GetBuilderBalance returns the balance for a builder.
func (s *BuilderEpochState) GetBuilderBalance(idx BuilderIndex) uint64 {
	return s.balances[idx]
}

// AddPendingPayment adds a pending payment for a builder.
func (s *BuilderEpochState) AddPendingPayment(idx BuilderIndex, amount uint64) {
	s.pendingPayments = append(s.pendingPayments, BuilderPendingPayment{idx, amount})
}

// PendingPayments returns the current pending payments.
func (s *BuilderEpochState) PendingPayments() []BuilderPendingPayment {
	return s.pendingPayments
}

// ProcessBuilderPendingPayments deducts pending payment amounts from builder
// balances and clears the pending payments list. Per EIP-7732 §epoch-processing.
// If the payment exceeds the builder balance, the payment is capped at balance.
func ProcessBuilderPendingPayments(state *BuilderEpochState) {
	for _, p := range state.pendingPayments {
		bal := state.balances[p.BuilderIdx]
		if p.Amount >= bal {
			state.balances[p.BuilderIdx] = 0
		} else {
			state.balances[p.BuilderIdx] = bal - p.Amount
		}
	}
	state.pendingPayments = nil
}
