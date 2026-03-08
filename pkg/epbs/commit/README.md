# epbs/commit — EIP-7732 builder commitment-reveal protocol

## Overview

Package `commit` implements the commitment-reveal protocol specified in EIP-7732. Builders commit to an execution payload by submitting a `BuilderCommitment` containing the slot, builder index, bid amount, and a Keccak-256 commitment hash. After winning the auction, the builder must reveal the matching `PayloadEnvelope` within a configurable slot window; failure triggers a non-reveal penalty, and revealing a mismatched payload triggers a mismatch penalty.

The `CommitmentManager` orchestrates all protocol phases: recording commitments in a per-slot linked-list chain for auditability, verifying reveals via `RevealVerifier`, enforcing deadlines via `RevealWindow`, and delegating penalty accounting to `CRPenaltyEngine`. All components are thread-safe.

## Functionality

**Types**
- `BuilderCommitment{Slot, BuilderIndex, BuilderAddr, BidAmount, CommitmentHash, BlockRoot Hash, Revealed bool, RevealedAt uint64}`
- `RevealWindow{DeadlineSlots uint64}` — `DefaultRevealWindow()` uses 1 slot
  - `IsExpired(commitSlot, currentSlot uint64) bool`
  - `IsWithinWindow(commitSlot, currentSlot uint64) bool`
  - `Deadline(commitSlot uint64) uint64`
- `CommitmentNode{Commitment *BuilderCommitment, Next *CommitmentNode}` — linked-list node
- `CRPenaltyConfig{NonRevealBasisPoints=20000, MismatchBasisPoints=30000}` — `DefaultCRPenaltyConfig()`
- `CRPenaltyRecord{Slot, BuilderIndex, BuilderAddr, BidAmount, PenaltyGwei, Reason}`

**CommitmentChain** (per-slot audit trail)
- `NewCommitmentChain() *CommitmentChain`
- `Append(c *BuilderCommitment)`
- `ForSlot(slot uint64) []*BuilderCommitment`
- `Len(slot uint64) int`
- `PruneSlot(slot uint64)`

**RevealVerifier**
- `NewRevealVerifier() *RevealVerifier`
- `Verify(commitment *BuilderCommitment, payload *epbs.PayloadEnvelope) error` — checks slot, builder index, and `BlockRoot == PayloadRoot`

**CRPenaltyEngine**
- `NewCRPenaltyEngine(config CRPenaltyConfig) *CRPenaltyEngine`
- `PenalizeNonReveal(commitment *BuilderCommitment) (*CRPenaltyRecord, error)`
- `PenalizeMismatch(commitment *BuilderCommitment, reason string) (*CRPenaltyRecord, error)`
- `Records() []*CRPenaltyRecord`
- `TotalPenaltyForBuilder(addr Address) uint64`

**CommitmentManager** (top-level orchestrator)
- `NewCommitmentManager(window RevealWindow, penaltyConfig CRPenaltyConfig) *CommitmentManager`
- `Commit(c *BuilderCommitment) error` — computes commitment hash; rejects duplicates
- `Reveal(payload *epbs.PayloadEnvelope, currentSlot uint64) error` — verifies and marks revealed; penalizes mismatch
- `CheckDeadlines(currentSlot uint64) []*CRPenaltyRecord` — penalizes all expired unrevealed commitments
- `GetCommitment(slot uint64, builder BuilderIndex) (*BuilderCommitment, bool)`
- `CommitmentCount() int`
- `PenaltyRecords() []*CRPenaltyRecord`

**Internal**
- `computeCommitmentHash(c *BuilderCommitment) Hash` — Keccak-256(blockRoot || builderAddr || slot:builderIndex:bidAmount)
- `computeBasisPointsPenalty(amount, basisPoints uint64) uint64`

Parent package: [`epbs`](../README.md)
