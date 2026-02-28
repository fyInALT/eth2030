# US-2.2 — Intermediate 8s Slot Step

**Epic:** EP-2 Fast Slots Infrastructure
**Total Story Points:** 5
**Sprint:** 2

> **As a** protocol designer,
> **I want** the slot timing system to support an intermediate 8-second slot time between 12s and 6s,
> **so that** the sqrt(2) progressive slot reduction schedule (12→8→6→4→3→2s) can be followed.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Vitalik's Proposal

> Use a sqrt(2) progression for slot time reduction: 12s → 8s → 6s → 4s → 3s → 2s. Each step divides by approximately sqrt(2) (≈1.414). The 8s step is between the current 12s and the K+ era 6s. This requires the slot timer to support 8s slots with a 3-phase structure: ~2.67s proposal + ~2.67s attestation + ~2.67s aggregation.

---

## Tasks

### Task 2.2.1 — QuickSlotConfig for 8s Slots

| Field | Detail |
|-------|--------|
| **Description** | Add a pre-K+ slot configuration for 8-second slots: 4 slots per epoch, 8s per slot = 32s epochs. Update `DefaultQuickSlotConfig()` to support a `SlotDurationSeconds` parameter and add `EightSecondSlotConfig()` factory. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) Config produces 8s slots. (2) Epoch duration = 32s. (3) Slot numbering correct. (4) Duties assigned per 8s cadence. |
| **Definition of Done** | Tests pass; 8s config works alongside 6s and 12s; reviewed. |

### Task 2.2.2 — PhaseTimer for 8s Slots

| Field | Detail |
|-------|--------|
| **Description** | Configure `PhaseTimer` for 8s slots with even 3-phase split: ~2667ms proposal + ~2667ms attestation + ~2666ms aggregation. Handle the 1ms rounding by assigning it to the last phase. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) Three phases sum to 8000ms. (2) Phase boundaries at correct timestamps. (3) `CurrentPhase()` returns correct phase at any time within slot. (4) Phase events emitted at correct boundaries. |
| **Definition of Done** | Tests pass; 8s phase timing is accurate; reviewed. |

### Task 2.2.3 — Progressive Slot Schedule

| Field | Detail |
|-------|--------|
| **Description** | Add `SlotSchedule` type that maps fork epochs to slot durations, implementing the sqrt(2) progression: `{Genesis: 12s, PreK: 8s, K: 6s, PostK: 4s, L: 3s, M: 2s}`. The scheduler picks the correct slot duration based on the current epoch. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) Genesis epoch → 12s slots. (2) PreK epoch → 8s slots. (3) K epoch → 6s slots. (4) Transition across fork boundary handled correctly. (5) Epoch numbering continuous across slot duration changes. |
| **Definition of Done** | Tests pass; slot schedule transitions work; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/consensus/quick_slots.go:25-32` | `QuickSlotConfig` struct — `SlotDuration time.Duration` (default 6s), `SlotsPerEpoch uint64` (default 4). Must support 8s configuration. |
| `pkg/consensus/quick_slots.go:34-40` | `DefaultQuickSlotConfig()` — returns 6s/4-slot config. Add `EightSecondSlotConfig()` variant. |
| `pkg/consensus/quick_slots.go:42-45` | `EpochDuration()` — `SlotDuration * SlotsPerEpoch`. Works for any slot duration. |
| `pkg/consensus/quick_slots.go:102-116` | `CurrentSlot()` and `SlotAt()` — wall-clock slot calculation. Works for any slot duration (divides elapsed time by `SlotDuration`). |
| `pkg/consensus/phase_timer.go:36-44` | `PhaseTimerConfig` — phase durations currently hardcoded at 2s each. Must support 2667ms phases for 8s slots. |
| `pkg/consensus/phase_timer.go:97-129` | `NewPhaseTimer()` — validates phase durations sum to slot duration. Already supports variable phase lengths. |
| `pkg/consensus/phase_timer.go:131-153` | `CurrentPhase()` — millisecond-resolution phase tracking. Works for any phase duration. |

---

## Implementation Status

**❌ Not Implemented (but infrastructure supports it)**

### What Exists
- ✅ `QuickSlotConfig` with configurable `SlotDuration` — supports any duration
- ✅ `PhaseTimer` with configurable phase durations — already validates sum equals slot duration
- ✅ `CurrentSlot()` / `CurrentPhase()` — works for any slot/phase duration (arithmetic-based)
- ✅ 6s slot configuration as default (`quick_slots.go:36-38`)
- ✅ 12s slot configuration in SSF (`ssf.go:47-48`)

### What's Missing
- ❌ No `EightSecondSlotConfig()` factory function
- ❌ No `SlotSchedule` type for progressive slot reduction across forks
- ❌ No fork-based slot duration transitions
- ❌ Phase durations for 8s slots (2667/2667/2666ms) not defined

### Proposed Solution

The existing infrastructure is designed for variable slot durations. Implementation is straightforward:

1. Add `EightSecondSlotConfig()` returning `QuickSlotConfig{SlotDuration: 8*time.Second, SlotsPerEpoch: 4}`
2. Add `SlotSchedule` as `map[uint64]time.Duration` (epoch → slot duration)
3. Update `QuickSlotScheduler` to consult `SlotSchedule` in `CurrentSlot()` and `SlotAt()`
4. Phase durations auto-computed as `slotDuration / phaseCount` with rounding

### Effort Assessment

This is a **LOW** effort gap — the slot/phase infrastructure already supports variable durations. The main work is adding the schedule type and fork-transition logic.
