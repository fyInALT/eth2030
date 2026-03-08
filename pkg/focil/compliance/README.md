# focil/compliance — FOCIL validator compliance tracking, MEV filtering, and builder inclusion auditing

## Overview

Package `compliance` provides three enforcement layers for Fork-Choice Enforced Inclusion Lists (EIP-7805). `ComplianceTracker` monitors individual validator duty fulfillment — recording violations by kind, maintaining a 0–1000 compliance score, and issuing slash recommendations when a validator accrues excessive misses after a grace period expires. `MEVFilter` identifies MEV transactions by known DEX and liquidation contract addresses and verifies that block builders include all required FOCIL transactions. `BuilderInclusionTracker` provides longitudinal per-builder compliance history across many slots, computing per-slot and aggregate inclusion rates.

All components are thread-safe.

## Functionality

**ComplianceTracker** (`compliance_tracker.go`)
- `ComplianceViolationKind`: `MissedSubmission=1`, `LateSubmission`, `Conflicting`, `InvalidContent`
- `ValidatorComplianceState{Score uint64 (0-1000), TotalDuties, DutiesFulfilled, ConsecutiveMisses, GraceRemaining uint64, SlashRecommended bool}`
- `ComplianceReport{Slot, ValidatorCount, DutiesFulfilled, TotalViolations uint64, Violations []..., ReportHash Hash}`
- `NewComplianceTracker(gracePeriod uint64) *ComplianceTracker`
- `RegisterValidator(validatorIndex uint64)`
- `RecordDutyFulfilled(validatorIndex uint64) error`
- `RecordViolation(validatorIndex uint64, kind ComplianceViolationKind, details string) error`
- `GetValidatorState(validatorIndex uint64) (*ValidatorComplianceState, bool)`
- `GetComplianceScore(validatorIndex uint64) (uint64, error)`
- `IsInGracePeriod(validatorIndex uint64) (bool, error)`
- `GenerateReport(slot uint64) (*ComplianceReport, error)` — Keccak-256 `ReportHash` binds all fields
- `SlashCandidates() []uint64` — validators with `SlashRecommended == true`
- `ResetSlashRecommendation(validatorIndex uint64) error`

**MEVFilter** (`mev_filter.go`)
- `MEVFilterConfig{KnownDEXContracts, KnownLiquidationContracts []Address, MinGasPriceMultiplier=5}`
- `BuilderComplianceResult{Compliant bool, MissingFOCILTxs []Hash, NonMEVTxs []Hash}`
- `NewMEVFilter(cfg MEVFilterConfig) *MEVFilter`
- `IsMEVTransaction(tx *types.Transaction) bool`
- `FilterMEVOnly(txs []*types.Transaction) []*types.Transaction`
- `FilterNonMEV(txs []*types.Transaction) []*types.Transaction`
- `AddDEXContract(addr Address)` / `AddLiquidationContract(addr Address)`
- `ValidateBuilderCompliance(blockTxs, focilTxs []*types.Transaction, mevOnly bool) BuilderComplianceResult`

**BuilderInclusionTracker** (`builder_inclusion_tracker.go`)
- `SlotComplianceResult{Slot uint64, BuilderID string, IncludedCount, RequiredCount int, MissingTxs []Hash, CompliancePercent float64}`
- `NewBuilderInclusionTracker(maxHistory int) *BuilderInclusionTracker`
- `RecordSlot(slot uint64, builderID string, required, included []Hash)`
- `GetComplianceRate(builderID string) float64` — fraction of slots meeting 100% inclusion
- `GetSlotCompliance(slot uint64) (*SlotComplianceResult, bool)`
- `IsCompliant(builderID string, threshold float64) bool`
- `GetNonCompliantBuilders(threshold float64) []string`
- `GetMissingTransactions(builderID string) []Hash` — union of all missing tx hashes
- `PruneHistory(keepSlots int)`
- `GetBuilderComplianceHistory(builderID string) []SlotComplianceResult`
- `GetAllSlotCompliance() []SlotComplianceResult`

Parent package: [`focil`](../README.md)
