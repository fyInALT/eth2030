# slot

Slot timing, phase tracking, and duty scheduling for 6-second quick slots (K+).

## Overview

Package `slot` provides all time-related consensus machinery for the K+
upgrade (2028): 6-second slots and 4-slot epochs. It is organized around three
complementary components:

1. **`QuickSlotScheduler`** (`quick_slots.go`): wall-clock-based slot/epoch
   calculations. Computes `CurrentSlot`, `SlotToEpoch`, `EpochStartSlot`, and
   `GetDuties` (RANDAO-based proposer + committee assignment per slot) using
   `DefaultQuickSlotConfig` (6 s, 4 slots/epoch).

2. **`PhaseTimer`** (`phase_timer.go`): millisecond-resolution sub-slot phase
   tracking. Divides each 6 s slot into `PhaseProposal` (2 s), `PhaseAttestation`
   (2 s), and `PhaseAggregation` (2 s). Supports event subscriptions via `Subscribe()`.

3. **`ProgressiveSlotSchedule`** (`slot_schedule.go`): epoch-keyed fork schedule
   mapping genesis (12 s, 32/epoch) → fast-slots (8 s, 8/epoch) → quick-slots
   (6 s, 4/epoch). `SlotToTime` converts absolute slot numbers to wall-clock times
   accounting for the changing durations across fork boundaries.

## Functionality

### QuickSlotConfig / QuickSlotScheduler

| Name | Description |
|------|-------------|
| `DefaultQuickSlotConfig() *QuickSlotConfig` | 6 s slots, 4 slots/epoch |
| `QuickSlot4sConfig() *QuickSlotConfig` | Experimental 4 s slots, 4 slots/epoch |
| `NewQuickSlotScheduler(config, genesisTime) *QuickSlotScheduler` | Create scheduler |
| `(*QuickSlotScheduler).CurrentSlot() uint64` | Wall-clock slot |
| `(*QuickSlotScheduler).CurrentEpoch() uint64` | Wall-clock epoch |
| `(*QuickSlotScheduler).GetDuties(slot, validatorCount) *ValidatorDuties` | Proposer + committee assignment |
| `(*QuickSlotScheduler).SlotStartTime(slot) time.Time` | Absolute start time of a slot |
| `ValidateConfig(config) error` | Config field validation |

### PhaseTimer

| Name | Description |
|------|-------------|
| `DefaultPhaseTimerConfig() *PhaseTimerConfig` | 6000 ms slot, equal 2000 ms phases, 4-slot epochs |
| `NewPhaseTimer(config) *PhaseTimer` | Create timer |
| `(*PhaseTimer).CurrentSlot() uint64` | Current slot |
| `(*PhaseTimer).CurrentPhase() SlotPhase` | Current sub-slot phase |
| `(*PhaseTimer).PhaseStartTime(slot, phase) time.Time` | Absolute start of a phase |
| `(*PhaseTimer).TimeToNextPhase() time.Duration` | Duration to next phase boundary |
| `(*PhaseTimer).Subscribe() <-chan SlotEvent` | Event channel for slot/phase boundaries |
| `(*PhaseTimer).Unsubscribe(ch)` | Stop receiving events |
| `SlotPhase` | `PhaseProposal`, `PhaseAttestation`, `PhaseAggregation` |

### ProgressiveSlotSchedule

| Name | Description |
|------|-------------|
| `DefaultProgressiveSlotSchedule() *ProgressiveSlotSchedule` | Genesis→fast→quick three-phase schedule |
| `(*ProgressiveSlotSchedule).GetSlotDuration(epoch) time.Duration` | Active slot duration for epoch |
| `(*ProgressiveSlotSchedule).GetSlotsPerEpoch(epoch) uint64` | Active slots-per-epoch for epoch |
| `(*ProgressiveSlotSchedule).SlotToTime(slot, genesisTime) time.Time` | Slot to wall-clock accounting for fork changes |
| `EightSecondSlotConfig() *QuickSlotConfig` | 8 s fast-slots config |
| `ComputeProgressiveDuration(base, step) time.Duration` | `base / sqrt(2)^step` progressive reduction |

## Usage

```go
import "github.com/eth2030/eth2030/consensus/slot"

sched := slot.NewQuickSlotScheduler(slot.DefaultQuickSlotConfig(), genesisTime)
duties := sched.GetDuties(sched.CurrentSlot(), validatorCount)

pt := slot.NewPhaseTimer(slot.DefaultPhaseTimerConfig())
ch := pt.Subscribe()
// Receive SlotEvent on each slot/phase boundary.
```

[← consensus](../README.md)
