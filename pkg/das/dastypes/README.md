# das/dastypes — Core PeerDAS types and constants (EIP-7594)

[← das](../README.md)

## Overview

This package defines the fundamental data types and consensus constants for PeerDAS (Peer Data Availability Sampling) as specified by EIP-7594 and the Fulu consensus spec. It is the shared type vocabulary used by all other `das/` sub-packages, avoiding import cycles while keeping numeric constants in a single authoritative location.

Constants are derived directly from the Fulu DAS spec (`consensus-specs/specs/fulu/das-core.md`): 128 columns, 128 custody groups, 4 minimum custody requirement, 64 gossip subnets, 4096 field elements per blob, and 2048 bytes per cell.

## Functionality

**Constants**
- `NumberOfColumns = 128` — columns in the extended data matrix
- `NumberOfCustodyGroups = 128`
- `CustodyRequirement = 4` — minimum custody groups per honest node
- `SamplesPerSlot = 8`
- `DataColumnSidecarSubnetCount = 64`
- `FieldElementsPerBlob = 4096`, `FieldElementsPerExtBlob = 8192`
- `FieldElementsPerCell = 64`, `BytesPerFieldElement = 32`
- `BytesPerCell = 2048`, `CellsPerExtBlob = 128`
- `MaxBlobCommitmentsPerBlock = 9` (EIP-7691)
- `ReconstructionThreshold = 64` (50% of columns)

**Named types**
- `SubnetID uint64`, `CustodyGroup uint64`, `ColumnIndex uint64`, `RowIndex uint64`

**Data structures**
- `Cell [2048]byte` — smallest independently provable unit of blob data
- `KZGCommitment [48]byte` — compressed BLS12-381 G1 commitment
- `KZGProof [48]byte` — compressed BLS12-381 G1 proof
- `DataColumn` — `Index ColumnIndex`, `Cells []Cell`, `KZGProofs []KZGProof`
- `DataColumnSidecar` — network gossip container: `Index`, `Column []Cell`, `KZGCommitments`, `KZGProofs`, `InclusionProof [][32]byte`
- `MatrixEntry` — `Cell`, `KZGProof`, `ColumnIndex`, `RowIndex`
- `STARKCommitment` — post-quantum DA commitment with `Root [32]byte`, `ProofSize`, `BlowupFactor`

## Usage

```go
// Check column count at compile time
var _ [dastypes.NumberOfColumns]dastypes.Cell

// Construct a data column sidecar
sidecar := &dastypes.DataColumnSidecar{
    Index:          colIdx,
    Column:         cells,
    KZGCommitments: commitments,
    KZGProofs:      proofs,
}
```
