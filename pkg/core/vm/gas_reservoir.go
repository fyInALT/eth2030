package vm

// ReservoirConfig holds parameters for the state creation gas reservoir.
// The reservoir accumulates gas earmarked for state-creating operations
// (SSTORE zero->nonzero, CREATE), separate from the main execution gas.
// This prevents state growth from consuming regular execution gas.
type ReservoirConfig struct {
	// Enabled indicates whether reservoir accounting is active.
	Enabled bool
	// ReservoirFraction is the fraction of intrinsic gas allocated to the reservoir (0.0-1.0).
	// Default: 0.25 (25% of intrinsic gas goes to state creation reservoir).
	ReservoirFraction float64
	// MinReservoir is the minimum reservoir amount (floor).
	MinReservoir uint64
	// MaxReservoir is the maximum reservoir amount (cap).
	MaxReservoir uint64
}

// DefaultReservoirConfig returns production defaults for the gas reservoir.
func DefaultReservoirConfig() *ReservoirConfig {
	return &ReservoirConfig{
		Enabled:           true,
		ReservoirFraction: 0.25,
		MinReservoir:      5000,
		MaxReservoir:      500_000,
	}
}

// InitReservoir splits intrinsic gas into execution gas and a state creation
// reservoir. The reservoir is sized as a fraction of intrinsicGas, clamped to
// [MinReservoir, MaxReservoir].
func InitReservoir(intrinsicGas uint64, config *ReservoirConfig) (execGas, reservoir uint64) {
	if config == nil || !config.Enabled {
		return intrinsicGas, 0
	}

	// Calculate reservoir as a fraction of intrinsic gas.
	reservoir = uint64(float64(intrinsicGas) * config.ReservoirFraction)

	// Clamp to [MinReservoir, MaxReservoir].
	if reservoir < config.MinReservoir {
		reservoir = config.MinReservoir
	}
	if reservoir > config.MaxReservoir {
		reservoir = config.MaxReservoir
	}

	// Reservoir cannot exceed intrinsic gas.
	if reservoir > intrinsicGas {
		reservoir = intrinsicGas
	}

	execGas = intrinsicGas - reservoir
	return execGas, reservoir
}

// DrawReservoir attempts to draw amount gas from the reservoir. If the
// reservoir has sufficient balance, it deducts the amount and returns true.
// Otherwise the reservoir is unchanged and false is returned, signaling that
// the caller should fall back to regular gas.
func DrawReservoir(reservoir *uint64, amount uint64) bool {
	if amount == 0 {
		return true
	}
	if *reservoir >= amount {
		*reservoir -= amount
		return true
	}
	return false
}

// ForwardReservoir transfers the entire parent reservoir to a child call
// context. The parent reservoir is zeroed during the child call.
func ForwardReservoir(parent, child *uint64) {
	*child = *parent
	*parent = 0
}

// ReturnReservoir returns unused child reservoir gas back to the parent
// after a child call completes.
func ReturnReservoir(parent, child *uint64) {
	*parent += *child
	*child = 0
}

// ReservoirGasCost computes SSTORE gas with reservoir awareness. For
// zero->nonzero transitions (state creation), it first attempts to draw the
// state creation cost (GasSstoreSetGlamsterdam = 24084) from the reservoir.
// If the draw succeeds, the caller only pays WarmStorageReadGlamst plus any
// cold access penalty. For all other cases, normal gas calculation applies.
//
// Parameters:
//   - original: committed value of the slot
//   - current:  current (possibly dirty) value
//   - newVal:   value being written
//   - cold:     true if the slot is cold (not yet in access list)
//   - reservoir: pointer to the reservoir balance
//
// Returns:
//   - gas: the gas to charge the caller
//   - drewFromReservoir: true if state creation cost was paid from reservoir
func ReservoirGasCost(original, current, newVal [32]byte, cold bool, reservoir *uint64) (gas uint64, drewFromReservoir bool) {
	var coldGas uint64
	if cold {
		coldGas = ColdSloadGlamst
	}

	// Only the zero->nonzero case (state creation) is reservoir-eligible.
	// This matches: original is zero, newVal is non-zero, and original == current
	// (i.e., the slot has not been dirtied in this transaction).
	if isZero(original) && !isZero(newVal) && original == current {
		// Try to draw the state creation cost from the reservoir.
		if DrawReservoir(reservoir, GasSstoreSetGlamsterdam) {
			// State creation cost came from reservoir. The caller pays only
			// the warm read cost plus any cold access penalty.
			return WarmStorageReadGlamst + coldGas, true
		}
		// Reservoir insufficient -- fall back to normal Glamsterdam cost.
		return GasSstoreSetGlamsterdam + coldGas, false
	}

	// Non-creation cases: delegate to standard Glamsterdam gas logic.
	if current == newVal {
		// No-op: current value equals new value.
		return WarmStorageReadGlamst + coldGas, false
	}

	if original == current {
		// original == current but newVal differs.
		if !isZero(original) {
			// Non-zero to different non-zero (update).
			return GasSstoreReset + coldGas, false
		}
		// isZero(original) && isZero(newVal) is impossible here because
		// we already handled current == newVal above (both zero).
	}

	// Dirty slot: original != current. Charge warm read only.
	return WarmStorageReadGlamst + coldGas, false
}
