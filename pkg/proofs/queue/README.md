# proofs/queue — Async proof validation queue with mandatory 3-of-5 tracking

## Overview

This package implements a bounded, concurrent proof validation queue that
supports the mandatory 3-of-5 proof requirement from the K+ roadmap. Callers
submit proofs for a block asynchronously; a configurable pool of worker
goroutines validates each proof and records the result in a `MandatoryProofTracker`
that determines when a block has reached the required threshold of distinct
proof types.

The five recognized proof types are `StateProof`, `StorageProof`,
`ExecutionTrace`, `WitnessProof`, and `ReceiptProof`. A block is considered
mandatory-proof-complete when at least three distinct types are validated
(`MandatoryThreshold = 3`). Validation metrics are exported via the `metrics`
package.

## Functionality

**Types**

- `QueueProofType` — enum of 5 proof categories; `String()` returns the name
- `ProofResult` — `BlockHash`, `ProofType`, `IsValid`, `Duration`, `Error`
- `ProofQueueConfig` — `Workers`, `QueueSize`, `DefaultDeadline`
- `ProofQueue` — created with `NewProofQueue(config)` and closed with `Close()`
- `MandatoryProofTracker` — per-block proof type registry
- `ProofDeadline` — deadline tracker with `SetDeadline`, `IsExpired`, `Prune`

**`ProofQueue` methods**

- `Submit(blockHash, proofType, proof) (<-chan ProofResult, error)` — enqueue
  and receive a buffered result channel
- `Tracker() *MandatoryProofTracker`
- `Metrics() (validated, failed, timedOut int64)`

**`MandatoryProofTracker` methods**

- `RecordProof(blockHash, proofType)`
- `HasMandatoryProofs(blockHash) bool` — true if >= 3 distinct types validated
- `ProofCount(blockHash) int`
- `ValidatedTypes(blockHash) []QueueProofType`
- `MissingTypes(blockHash) []QueueProofType`

**Test helper** — `MakeValidProof(blockHash, proofType) []byte`

**Parent package:** [proofs](../)
