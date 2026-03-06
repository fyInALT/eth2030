package exitqueue

// MinslackExitQueue implements the "minslack" exit queue flexibility model
// from the L+ roadmap. Exit delay adapts to current queue pressure: empty
// queues get the full withdrawal delay while a saturated queue fast-tracks
// exits to MaxSeedLookahead epochs. This prevents queue bottlenecks when
// many validators exit simultaneously.

// MinslackExitQueue computes adaptive exit delays based on queue fill level.
// All methods are stateless; callers supply epoch, queueSize, and maxChurn.
type MinslackExitQueue struct{}

// ComputeExitDelay returns the exit delay (in epochs) for a validator joining
// the exit queue at the given epoch when the queue currently has queueSize
// entries and the per-epoch churn cap is maxChurn.
//
// Delay schedule:
//   - Empty queue (queueSize == 0):       MinValidatorWithdrawabilityDelay (256)
//   - Full queue (queueSize >= maxChurn): MaxSeedLookahead (4)
//   - Half-full (queueSize < maxChurn/2): MinValidatorWithdrawabilityDelay / 2 (128)
//   - Otherwise (proportional):           MinValidatorWithdrawabilityDelay/2 * (maxChurn - queueSize) / maxChurn
func (m *MinslackExitQueue) ComputeExitDelay(epoch, queueSize, maxChurn uint64) uint64 {
	if maxChurn == 0 {
		// Degenerate case: treat as full queue to avoid division by zero.
		return MaxSeedLookahead
	}

	switch {
	case queueSize == 0:
		// Empty queue: use the standard full withdrawal delay.
		return MinValidatorWithdrawabilityDelay

	case queueSize >= maxChurn:
		// Saturated queue: fast-track exits to avoid indefinite blocking.
		return MaxSeedLookahead

	case queueSize < maxChurn/2:
		// Lightly loaded queue: halved delay encourages orderly exits.
		return MinValidatorWithdrawabilityDelay / 2

	default:
		// Proportional delay: linearly interpolate between MaxSeedLookahead
		// and MinValidatorWithdrawabilityDelay/2 as the queue fills from
		// half-capacity to full capacity.
		return MinValidatorWithdrawabilityDelay / 2 * (maxChurn - queueSize) / maxChurn
	}
}
