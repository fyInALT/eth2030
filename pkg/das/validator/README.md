# das/validator — PeerDAS data availability validator

Validates blob data availability using PeerDAS column custody and sampling.
Wraps core DAS primitives into a single validator that verifies column sidecars,
checks custody assignments, and determines per-slot data availability.

## Overview

`DAValidator` is the main entry point. It combines `sampling.GetCustodyColumns`
for custody assignment validation, a deterministic per-(slot, nodeID) column
sampler for random sampling selection, and a hash-based column proof verifier.

Specialised sub-validators handle individual concerns:
- `blob_validator.go` — blob-level KZG commitment structure
- `cell_validator.go` — individual cell verification
- `l2_data_validator.go` — L2 data payload validation
- `async_validator.go` — non-blocking validation with result channels

## Functionality

**Types**
- `DAValidatorConfig` — `MinCustodyGroups`, `ColumnCount`, `SamplesPerSlot`, `MaxBlobsPerBlock`
- `DAValidator` — main availability validator

**Functions**
- `DefaultDAValidatorConfig() DAValidatorConfig`
- `NewDAValidator(config) *DAValidator`
- `(*DAValidator).ValidateColumnSidecar(slot, columnIndex, data, proof) error`
- `ComputeColumnProof(slot, columnIndex, data) []byte` — exported proof helper for tests
- `(*DAValidator).ValidateCustodyAssignment(nodeID, columns) error`
- `(*DAValidator).ComputeCustodyColumns(nodeID, custodyGroups) []uint64`
- `(*DAValidator).IsDataAvailable(slot, availableColumns, requiredColumns) bool`
- `(*DAValidator).SampleColumns(slot, nodeID) []uint64`

## Usage

```go
v := validator.NewDAValidator(validator.DefaultDAValidatorConfig())

proof := validator.ComputeColumnProof(slot, colIdx, data)
err := v.ValidateColumnSidecar(slot, colIdx, data, proof)

required := v.SampleColumns(slot, nodeID)
available := map[uint64]bool{5: true, 42: true /* ... */}
ok := v.IsDataAvailable(slot, available, required)
```

[← das](../README.md)
