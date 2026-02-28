// slot_schedule.go implements a progressive slot schedule for the ETH2030
// consensus layer. It maps fork epochs to slot durations and slots-per-epoch
// counts, enabling the transition from 12s genesis slots through 8s fast
// slots to 6s quick slots across successive upgrades.
package consensus

import (
	"errors"
	"math"
	"sort"
	"time"
)

// Progressive slot schedule errors.
var (
	ErrSSNoEntries    = errors.New("slot_schedule: no schedule entries")
	ErrSSInvalidEpoch = errors.New("slot_schedule: invalid epoch")
	ErrSSOverlapping  = errors.New("slot_schedule: overlapping epoch ranges")
)

// ProgressiveSlotEntry maps a fork epoch to a slot duration.
type ProgressiveSlotEntry struct {
	ForkEpoch     uint64
	SlotDuration  time.Duration
	SlotsPerEpoch uint64
	Name          string // human-readable fork name
}

// ProgressiveSlotSchedule holds a progression of slot durations across forks.
// Unlike SlotSchedule (in slots.go) which is timestamp-based, this is
// epoch-based and supports the full three-phase progressive schedule.
type ProgressiveSlotSchedule struct {
	entries []ProgressiveSlotEntry
}

// DefaultProgressiveSlotSchedule returns the progressive schedule with three
// forks: genesis (12s, 32 slots/epoch), fast-slots (8s, 8 slots/epoch),
// quick-slots (6s, 4 slots/epoch).
func DefaultProgressiveSlotSchedule() *ProgressiveSlotSchedule {
	s, _ := NewProgressiveSlotSchedule([]ProgressiveSlotEntry{
		{ForkEpoch: 0, SlotDuration: 12 * time.Second, SlotsPerEpoch: 32, Name: "genesis"},
		{ForkEpoch: 100000, SlotDuration: 8 * time.Second, SlotsPerEpoch: 8, Name: "fast-slots"},
		{ForkEpoch: 200000, SlotDuration: 6 * time.Second, SlotsPerEpoch: 4, Name: "quick-slots"},
	})
	return s
}

// NewProgressiveSlotSchedule creates a slot schedule from the given entries.
// Entries are sorted by ForkEpoch. Returns an error if the list is empty or
// contains duplicate fork epochs.
func NewProgressiveSlotSchedule(entries []ProgressiveSlotEntry) (*ProgressiveSlotSchedule, error) {
	if len(entries) == 0 {
		return nil, ErrSSNoEntries
	}

	// Sort by ForkEpoch.
	sorted := make([]ProgressiveSlotEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ForkEpoch < sorted[j].ForkEpoch
	})

	// Check for duplicate epochs.
	for i := 1; i < len(sorted); i++ {
		if sorted[i].ForkEpoch == sorted[i-1].ForkEpoch {
			return nil, ErrSSOverlapping
		}
	}

	return &ProgressiveSlotSchedule{entries: sorted}, nil
}

// GetSlotDuration returns the slot duration for the given epoch.
// It finds the latest entry whose ForkEpoch <= epoch.
func (s *ProgressiveSlotSchedule) GetSlotDuration(epoch uint64) time.Duration {
	entry := s.GetEntry(epoch)
	if entry == nil {
		return 12 * time.Second // fallback
	}
	return entry.SlotDuration
}

// GetSlotsPerEpoch returns the number of slots per epoch for the given epoch.
func (s *ProgressiveSlotSchedule) GetSlotsPerEpoch(epoch uint64) uint64 {
	entry := s.GetEntry(epoch)
	if entry == nil {
		return 32 // fallback
	}
	return entry.SlotsPerEpoch
}

// GetEntry returns the active schedule entry for the given epoch.
// Returns the latest entry where ForkEpoch <= epoch.
func (s *ProgressiveSlotSchedule) GetEntry(epoch uint64) *ProgressiveSlotEntry {
	if len(s.entries) == 0 {
		return nil
	}

	// Find the latest entry where ForkEpoch <= epoch.
	var result *ProgressiveSlotEntry
	for i := range s.entries {
		if s.entries[i].ForkEpoch <= epoch {
			result = &s.entries[i]
		} else {
			break
		}
	}
	return result
}

// SlotToTime converts a slot number to wall clock time, accounting for
// changing durations across fork boundaries. It iterates through schedule
// entries, accumulating time for each fork epoch range.
func (s *ProgressiveSlotSchedule) SlotToTime(slot uint64, genesisTime time.Time) time.Time {
	if len(s.entries) == 0 {
		return genesisTime
	}

	elapsed := time.Duration(0)
	remainingSlots := slot
	currentSlot := uint64(0)

	for i, entry := range s.entries {
		// Compute the first slot of this fork's epoch range.
		forkStartSlot := s.epochToSlot(entry.ForkEpoch, i)
		if forkStartSlot > currentSlot {
			// We should not be behind the fork start.
			forkStartSlot = currentSlot
		}

		// Determine how many slots are covered by the previous entry
		// before we reach this fork.
		if i > 0 && currentSlot < forkStartSlot {
			prevEntry := s.entries[i-1]
			slotsInPrev := forkStartSlot - currentSlot
			if slotsInPrev > remainingSlots {
				slotsInPrev = remainingSlots
			}
			elapsed += time.Duration(slotsInPrev) * prevEntry.SlotDuration
			remainingSlots -= slotsInPrev
			currentSlot += slotsInPrev
			if remainingSlots == 0 {
				return genesisTime.Add(elapsed)
			}
		}

		// Determine how many slots this fork entry covers.
		var slotsInFork uint64
		if i+1 < len(s.entries) {
			nextForkStartSlot := s.epochToSlot(s.entries[i+1].ForkEpoch, i+1)
			if nextForkStartSlot > currentSlot {
				slotsInFork = nextForkStartSlot - currentSlot
			}
		} else {
			// Last entry covers all remaining slots.
			slotsInFork = remainingSlots
		}

		if slotsInFork > remainingSlots {
			slotsInFork = remainingSlots
		}

		elapsed += time.Duration(slotsInFork) * entry.SlotDuration
		remainingSlots -= slotsInFork
		currentSlot += slotsInFork

		if remainingSlots == 0 {
			return genesisTime.Add(elapsed)
		}
	}

	// If we still have remaining slots, use the last entry's duration.
	if remainingSlots > 0 {
		lastEntry := s.entries[len(s.entries)-1]
		elapsed += time.Duration(remainingSlots) * lastEntry.SlotDuration
	}

	return genesisTime.Add(elapsed)
}

// epochToSlot computes the absolute slot number at the start of a fork epoch,
// by summing slots from all prior entries. entryIdx is the index into entries
// for context (entries before entryIdx are used).
func (s *ProgressiveSlotSchedule) epochToSlot(epoch uint64, entryIdx int) uint64 {
	if entryIdx == 0 || epoch == 0 {
		return 0
	}

	totalSlots := uint64(0)
	for i := 0; i < entryIdx; i++ {
		entry := s.entries[i]
		var epochsInFork uint64
		if i+1 < len(s.entries) && s.entries[i+1].ForkEpoch <= epoch {
			epochsInFork = s.entries[i+1].ForkEpoch - entry.ForkEpoch
		} else {
			epochsInFork = epoch - entry.ForkEpoch
		}
		totalSlots += epochsInFork * entry.SlotsPerEpoch
	}
	return totalSlots
}

// Entries returns a copy of the schedule entries.
func (s *ProgressiveSlotSchedule) Entries() []ProgressiveSlotEntry {
	result := make([]ProgressiveSlotEntry, len(s.entries))
	copy(result, s.entries)
	return result
}

// EightSecondSlotConfig returns a QuickSlotConfig with 8s duration and
// 8 slots per epoch, for the fast-slots upgrade.
func EightSecondSlotConfig() *QuickSlotConfig {
	return &QuickSlotConfig{
		SlotDuration:  8 * time.Second,
		SlotsPerEpoch: 8,
	}
}

// EightSecondPhaseTimerConfig returns a PhaseTimerConfig for 8-second slots:
// 3000ms proposal, 3000ms attestation, 2000ms aggregation.
func EightSecondPhaseTimerConfig() *PhaseTimerConfig {
	return &PhaseTimerConfig{
		SlotDurationMs:     8000,
		ProposalPhaseMs:    3000,
		AttestationPhaseMs: 3000,
		AggregationPhaseMs: 2000,
		GenesisTime:        0,
		SlotsPerEpoch:      8,
	}
}

// ComputeProgressiveDuration computes a progressive slot duration using
// sqrt(2) reduction: baseDuration / sqrt(2)^step. Step 0 returns the
// base duration, step 1 returns base/sqrt(2), step 2 returns base/2, etc.
func ComputeProgressiveDuration(baseDuration time.Duration, step int) time.Duration {
	if step <= 0 {
		return baseDuration
	}
	divisor := math.Pow(math.Sqrt2, float64(step))
	ns := float64(baseDuration.Nanoseconds()) / divisor
	return time.Duration(int64(ns))
}
