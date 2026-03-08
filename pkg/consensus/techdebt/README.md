# techdebt

Beacon spec technical debt tracking and automated state migration.

## Overview

Package `techdebt` implements the CL Accessibility roadmap item "tech debt
reset". `TechDebtTracker` maintains a registry of deprecated beacon state
fields — recording when each field was deprecated, what replaces it, and when
it will be fully removed. `MigrateState` operates on a `map[string]interface{}`
beacon state representation: it copies deprecated field values to their
replacements and, when `AutoMigrate` is enabled, deletes fields that have
passed their `RemovalEpoch`.

`DefaultTechDebtConfig` pre-loads well-known Altair deprecations:
`previous_epoch_attestations`, `current_epoch_attestations` (both replaced by
participation flag bytes at epoch 74240), and Phase 1 shard fields
(`compact_committees_root`, `shard_states`, `latest_crosslinks`) that were
never activated on mainnet.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `TechDebtTracker` | Thread-safe registry of deprecated fields |
| `TechDebtConfig` | `KnownDeprecations []*DeprecatedField`, `AutoMigrate bool` |
| `DeprecatedField` | `FieldName`, `DeprecatedSinceEpoch`, `ReplacedBy []string`, `RemovalEpoch` |
| `MigrationReport` | `FieldsMigrated`, `FieldsRemoved`, `Errors []string` |

### Constants

| Name | Value |
|------|-------|
| `AltairEpoch` | 74240 |
| `Phase1RemovalEpoch` | 74240 |

### Functions / methods

| Name | Description |
|------|-------------|
| `DefaultTechDebtConfig() *TechDebtConfig` | Config with Altair + Phase 1 deprecations; AutoMigrate=true |
| `NewTechDebtTracker(config) *TechDebtTracker` | Create tracker and pre-load known deprecations |
| `(*TechDebtTracker).RegisterDeprecation(field) error` | Add a new deprecation record |
| `(*TechDebtTracker).IsDeprecated(fieldName, currentEpoch) bool` | True if field is deprecated at this epoch |
| `(*TechDebtTracker).IsRemoved(fieldName, currentEpoch) bool` | True if field is past its removal epoch |
| `(*TechDebtTracker).GetReplacements(fieldName) []string` | Return replacement field names |
| `(*TechDebtTracker).MigrateState(state, currentEpoch) (map, *MigrationReport, error)` | Non-mutating migration; returns new state map |
| `(*TechDebtTracker).DeprecationReport(currentEpoch) []DeprecatedField` | Active deprecations sorted by epoch |
| `(*TechDebtTracker).CleanupRemovedFields(state, currentEpoch) int` | In-place removal of past-deadline fields |
| `(*TechDebtTracker).ValidateMigrationReadiness(state, currentEpoch) error` | Pre-migration consistency check |

## Usage

```go
import "github.com/eth2030/eth2030/consensus/techdebt"

tracker := techdebt.NewTechDebtTracker(nil) // uses DefaultTechDebtConfig
migratedState, report, err := tracker.MigrateState(beaconStateMap, currentEpoch)
fmt.Printf("migrated %d fields, removed %d\n", report.FieldsMigrated, report.FieldsRemoved)
```

[← consensus](../README.md)
