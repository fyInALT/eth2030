# Sprint 4, Story 4.1 — Recursive Tick Merging

**Sprint goal:** Make STARK mempool ticks recursive across peers.
**Files modified:** `pkg/txpool/stark_aggregation.go`
**Files tested:** `pkg/txpool/stark_recursion_test.go`

## Overview

The ethresear.ch recursive STARK mempool proposal requires that each peer's tick proof covers all previously-seen valid transactions — both locally validated and received from remote peers. Without recursive merging, each peer only proves its own local subset, defeating the bandwidth efficiency goal.

## Gap (GAP-STARK1 + GAP-STARK6)

**Severity:** CRITICAL
**File:** `pkg/txpool/stark_aggregation.go` — `MergeTick()` at line 377 and `GenerateTick()` at line 315

**Evidence:** `MergeTick()` verified the remote STARK proof but did NOT add the remote peer's transactions to the local valid set. `GenerateTick()` only proved locally-validated transactions.

**Impact:** STARK proofs didn't grow across peers — each peer reproved only its own subset. The proposal requires recursive accumulation.

## Implement

### Step 1: Add RemoteProven flag to ValidatedTx

```go
// pkg/txpool/stark_aggregation.go
type ValidatedTx struct {
    Hash     types.Hash
    GasPrice *big.Int
    // RemoteProven indicates this tx was verified via a remote STARK proof
    // rather than local validation. It is included in subsequent local ticks
    // to achieve recursive accumulation across the peer network.
    RemoteProven bool
}
```

### Step 2: Merge remote txs in MergeTick

```go
func (sa *STARKAggregator) MergeTick(remote *MempoolAggregationTick) error {
    // ... STARK proof verification ...

    // Merge remote transactions into local valid set for recursive accumulation.
    sa.mu.Lock()
    defer sa.mu.Unlock()
    for _, txHash := range remote.ValidTxHashes {
        exists := false
        for _, local := range sa.validTxs {
            if local.Hash == txHash {
                exists = true
                break
            }
        }
        if !exists {
            sa.validTxs = append(sa.validTxs, ValidatedTx{
                Hash:         txHash,
                RemoteProven: true,
            })
        }
    }
    return nil
}
```

### Step 3: GenerateTick now includes remote txs

After the merge, `GenerateTick()` naturally includes remote-proven txs because they're in `sa.validTxs`. The STARK proof now attests to all known valid transactions (local + remote).

## ethresear.ch Spec Reference

> Each node runs a mempool aggregation tick every 500ms. The tick produces a STARK proof that all known valid transactions (local + received from peers) satisfy the validation rules. Upon receiving a remote tick, a node verifies the STARK proof and merges the remote transactions into its local set. The next local tick then recursively proves the combined set.

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/txpool/stark_aggregation.go` | 315 | GenerateTick — includes all validTxs (local + remote) |
| `pkg/txpool/stark_aggregation.go` | 377 | MergeTick — verifies remote STARK, merges txs |
| `pkg/txpool/stark_recursion_test.go` | — | TestMergeTick_MergesRemoteTransactions |
