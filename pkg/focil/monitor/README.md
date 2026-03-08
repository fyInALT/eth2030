# focil/monitor — FOCIL inclusion monitoring, mempool tracking, and censorship detection

## Overview

Package `monitor` provides runtime observability for Fork-Choice Enforced Inclusion Lists (EIP-7805). `InclusionMonitor` tracks per-slot required vs. included items across all registered builders, accumulates per-builder penalty points for missed inclusions, and identifies the most-compliant builders. `MempoolMonitor` watches pending transactions, records their inclusion or exclusion outcomes, and drives `FairnessAnalyzer` and `CensorshipIndicator` to surface statistical evidence of sender-targeted censorship.

Both monitors are safe for concurrent use.

## Functionality

**InclusionMonitor** (`inclusion_monitor.go`)
- `MonitorConfig{MaxTrackedSlots=256, ComplianceThreshold=0.90, PenaltyPerMiss=1000}`
- `InclusionItem{TxHash [32]byte, Sender [20]byte, GasLimit uint64, Priority uint64}`
- `SlotComplianceReport{RequiredCount, IncludedCount, MissedCount int, ComplianceRate float64, MissedItems []InclusionItem}`
- `NewInclusionMonitor(cfg MonitorConfig) *InclusionMonitor`
- `RecordSlot(slot uint64, required, included []InclusionItem)`
- `SlotCompliance(slot uint64) (*SlotComplianceReport, bool)`
- `RegisterBuilder(builderID string)`
- `BuilderCompliance(builderID string) float64` — fraction of slots at or above threshold
- `MostCompliant(n int) []string` — top-n builders by compliance rate
- `PenaltyAccrued(builderID string) uint64` — cumulative missed-inclusion penalties
- `PruneOldSlots(keepSlots int) int`

**MempoolMonitor** (`mempool_monitor.go`)
- `ComplianceMetrics{TotalPending, IncludedCount, ExcludedCount, UnresolvedCount int, ComplianceRate float64}`
- `NewMempoolMonitor() *MempoolMonitor`
- `TrackPending(txHash [32]byte, sender [20]byte)`
- `RecordOutcome(txHash [32]byte, included bool)`
- `Metrics() ComplianceMetrics`

**FairnessAnalyzer**
- `FairnessScore float64` — 1 minus Gini coefficient of per-sender inclusion rates; 1.0 = perfect fairness
- `SuspectedCensored []Address` — senders with inclusion rate < 0.25 and at least 3 exclusion events
- `NewFairnessAnalyzer() *FairnessAnalyzer`
- `Analyze(monitor *MempoolMonitor) FairnessAnalysis`
- `FairnessAnalysis{Score float64, SuspectedCensored []Address, SenderRates map[Address]float64}`

**CensorshipIndicator**
- `NewCensorshipIndicator(threshold float64) *CensorshipIndicator`
- `IsCensored(addr Address) bool` — inclusion rate below threshold
- `CensoredSenders() []Address`
- `RecordOutcome(addr Address, included bool)`

Parent package: [`focil`](../README.md)
