# das/reconstruction — Reed-Solomon blob reconstruction pipeline

Recovers full blobs from partial cell samples using Lagrange interpolation over
the BLS12-381 scalar field. Implements the PeerDAS local blob reconstruction
requirement from the consensus spec (`fulu/das-core.md`).

## Overview

The package operates at two levels. The low-level layer (`reconstruction.go`,
`blob_reconstruct.go`) implements `ReconstructBlob` and `RecoverCellsAndProofs`
using Lagrange polynomial interpolation: given ≥ 64 of the 128 extended-blob
cells, it recovers all 4096 field elements of the original blob. The matrix-level
`RecoverMatrix` helper processes multiple blobs (rows) independently.

The high-level layer (`blob_reconstruct.go`) provides `BlobReconstructor`, which
manages sample collection across multiple blobs with deduplication, parallel
reconstruction, threshold checking, and atomic metrics. A complete orchestration
layer (`reconstruction_pipeline.go`) adds a `CellCollector`, a priority-based
`ReconstructionScheduler`, a `ValidationStep`, and a `ReconstructionPipeline`
that ties all stages together.

## Functionality

**Types**
- `Sample` — a single cell with `BlobIndex`, `CellIndex`, and `Data`
- `BlobReconstructor` — manages pending samples per blob, thread-safe
- `ReconstructionMetrics` — atomic counters for success, failure, latency
- `CellCollector` — per-(slot, blobIndex) cell accumulator
- `ReconstructionScheduler` — priority-ordered list of blobs ready to reconstruct
- `ValidationStep` — verifies reconstructed blob size against commitment
- `ReconstructionPipeline` — full collect → schedule → decode → validate pipeline
- `ReconPipelineMetrics` — pipeline-level counters and latency tracking

**Functions**
- `CanReconstruct(receivedCount int) bool` — threshold check (≥ 64 cells)
- `ReconstructPolynomial(xs, ys, k)` — Lagrange interpolation
- `ReconstructBlob(cells, cellIndices)` — full blob recovery
- `RecoverCellsAndProofs(cells, cellIndices)` — full cell set recovery
- `RecoverMatrix(entries, blobCount)` — multi-blob matrix recovery
- `(*BlobReconstructor).AddSample`, `.Reconstruct`, `.ReconstructBlobs`, `.ReconstructPending`
- `ReconstructWithErasure(br, blobIndex, samples)` — end-to-end helper
- `(*ReconstructionPipeline).InitBlob`, `.AddCell`, `.Reconstruct`, `.RunScheduled`

## Usage

```go
br := reconstruction.NewBlobReconstructor(0)
for _, s := range receivedSamples {
    br.AddSample(s)
}
if br.CanReconstructBlob(0) {
    data, err := br.Reconstruct(receivedSamples, dastypes.CellsPerExtBlob)
}

// Or use the full pipeline:
pipe := reconstruction.NewReconstructionPipeline()
pipe.InitBlob(slot, blobIdx, commitment, reconstruction.PriorityHigh)
pipe.AddCell(slot, blobIdx, cellIdx, cell)
data, err := pipe.Reconstruct(slot, blobIdx)
```

[← das](../README.md)
