# rollup/sequencer — Transaction batch sequencer for L1 submission

## Overview

This package collects raw transactions into sealed batches for submission to L1
by a native rollup sequencer. Transactions are accumulated into a mutable pending
set until either the batch is manually sealed or the size limit is reached.
Sealed batches carry a Keccak256 batch ID computed over all individual transaction
hashes, an optional zlib-compressed payload, and a sealed timestamp.

The `Sequencer` is safe for concurrent use and maintains a complete history of
sealed batches accessible via `BatchHistory`. Batch integrity can be re-checked
at any time with `VerifyBatch`.

## Functionality

**Types**

- `Config` — `MaxBatchSize`, `BatchTimeout`, `L1SubmissionInterval`,
  `CompressPayload`; `DefaultConfig()` returns 1000-tx batches, 2s timeout,
  6s submission interval
- `Batch` — `ID`, `Transactions`, `L1BlockNumber`, `Timestamp`, `Compressed`,
  `CompressedData`
- `Sequencer` — created with `NewSequencer(config)`

**`Sequencer` methods**

- `AddTransaction(tx []byte) error` — appends tx to pending batch; returns
  `ErrBatchFull` when `MaxBatchSize` is reached
- `SealBatch() (*Batch, error)` — finalizes pending txs into a `Batch`; clears
  pending; optionally zlib-compresses the payload
- `PendingCount() int`
- `BatchHistory() []*Batch` — returns a copy of all sealed batches in order
- `VerifyBatch(batch) bool` — recomputes the batch ID and compares

**Parent package:** [rollup](../)
