# recovery

51% attack detection and auto-recovery pipeline (M+ roadmap).

## Overview

Package `recovery` implements the CL Accessibility roadmap item "51% attack
auto-recovery". `AttackDetector` monitors chain reorganisations measured in
epochs. Reorgs deeper than `ReorgThresholdLow` (2 epochs) trigger graduated
severity levels up to `SeverityCritical` (16+ epochs). Any reorg reaching back
into finalized territory is immediately classified as at least `SeverityHigh`.

`BuildRecoveryPlan` maps a severity to a `RecoveryPlan` that may combine:
peer isolation, fallback to the last finalized checkpoint, or a social
consensus override flag. `AttackDetector.ExecuteRecovery` applies the plan
and records the resulting `RecoveryStatus`.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `AttackDetector` | Thread-safe detector and recovery coordinator |
| `AttackReport` | Detected flag, severity, reorg depth, affected epochs, recommended action |
| `RecoveryPlan` | `IsolationMode`, `FallbackToFinalized`, `SocialConsensusOverride`, target epoch |
| `RecoveryStatus` | Current recovery state: peers isolated, fell back, social override set |

### Constants — severity levels and thresholds

| Constant | Value | Meaning |
|----------|-------|---------|
| `ReorgThresholdLow` | 2 | Low severity: monitor |
| `ReorgThresholdMedium` | 4 | Medium: isolate peers |
| `ReorgThresholdHigh` | 8 | High: fallback to finalized |
| `ReorgThresholdCritical` | 16 | Critical: social override required |

### Functions / methods

| Name | Description |
|------|-------------|
| `NewAttackDetector() *AttackDetector` | Create detector |
| `(*AttackDetector).DetectAttack(reorgDepth, finalizedEpoch, currentEpoch) *AttackReport` | Classify and store the attack report |
| `(*AttackDetector).IsUnderAttack() bool` | True if last detection found an attack |
| `(*AttackDetector).LastReport() *AttackReport` | Most recent attack report |
| `BuildRecoveryPlan(report) (*RecoveryPlan, error)` | Derive recovery actions from severity |
| `(*AttackDetector).ExecuteRecovery(plan) error` | Apply the plan and update status |
| `(*AttackDetector).ClearRecovery()` | Reset after network stabilisation |
| `SeverityLevel(reorgDepth) string` | Classify depth into severity string |
| `ValidateRecoveryPlan(plan) error` | Consistency check on plan fields |
| `ValidateAttackReport(report) error` | Consistency check on report fields |

## Usage

```go
import "github.com/eth2030/eth2030/consensus/recovery"

detector := recovery.NewAttackDetector()
report := detector.DetectAttack(reorgDepth, finalizedEpoch, currentEpoch)
if report.Detected {
    plan, _ := recovery.BuildRecoveryPlan(report)
    detector.ExecuteRecovery(plan)
}
```

[← consensus](../README.md)
