# Local Transaction Design (2026-03)

**Status**: Research spike — not yet an EIP
**Author**: ETH2030 Protocol Research
**Date**: 2026-03-05
**Related**: EP-6 US-BB-2, `pkg/core/types/tx_local.go` (type 0x08)

---

## 1. Motivation

Vitalik's distributed block building proposal notes that a "less global" transaction type
would dramatically reduce ordering coordination costs. Today every transaction is global:
any builder must consider it against every other pending transaction, limiting parallel
execution and making distributed block building hard.

A *local transaction* is one that pre-declares its full state access surface. Builders can
partition the pending pool by declared scope and execute non-overlapping subsets in parallel,
with no cross-shard locking required.

---

## 2. What Makes a Transaction "Less Global"

A transaction is considered local when:

1. **Declared BAL** — the transaction includes a Block Access List (BAL) in its body,
   enumerating every storage slot and account it will read or write.
2. **Bounded state access** — the EVM enforces that the transaction accesses only declared
   entries. Any undeclared access reverts (or pays a heavy dynamic-access surcharge).
3. **Scope hint** — a compact 1-byte address-prefix set indicates which portion of the
   state is touched (e.g., `[0x0a, 0x0b]` = accounts starting with `0x0a` or `0x0b`).
4. **No cross-scope calls** — external calls into other declared accounts are allowed, but
   calls into accounts outside the declared scope are prohibited (or penalised).

Contrast with global transactions: they can call any address, read any slot, and require
full ordering coordination to avoid state conflicts.

---

## 3. Distributed Building Without Ordering Coordination

With local transactions:

- **Partition**: the builder pool shards pending txs by `ScopeHint` prefix.
- **Assign**: each sub-builder receives only the scope shards whose prefixes do not overlap
  with other sub-builders' shards.
- **Execute**: sub-builders execute their shards in parallel. Because state ranges do not
  overlap, there are no RAW/WAW hazards.
- **Merge**: sub-builders submit partial execution witnesses and the proposer merges them.
  The merge is O(n) in number of sub-builders rather than O(n²) in tx count.

This maps directly onto the BAL-based parallel execution framework in `pkg/bal/`.

**Coordination cost reduction**: With k non-overlapping sub-builders, the ordering problem
decomposes into k independent sub-problems. For uniformly distributed scopes, expected
ordering cost per sub-builder is O(n/k) instead of O(n).

---

## 4. Gas Discount Model

Two rationales support a gas discount for local transactions:

### 4.1 Reduced validation cost

Pre-declared BALs allow the EVM to:
- Skip dynamic storage-slot hashing (pre-loaded cache hit guaranteed).
- Skip access-list bloom filter checks for declared entries.
- Pre-warm declared slots before execution begins.

Estimated savings: 15–25% of base gas for storage-heavy contracts.

### 4.2 Builder incentive alignment

A discount incentivises users to declare scope, improving block building parallelism.
Without a discount, rational users would use global transactions to avoid the constraint.

### 4.3 Proposed discount schedule

| Declared scope coverage | Gas discount |
|------------------------|--------------|
| Full BAL declared      | 50%          |
| Partial BAL declared   | 25%          |
| No BAL                 | 0% (global tx)|

The 50% cap is chosen to be below the expected execution-time saving of parallel building
(estimated 60–80% for large blocks with k≥8 sub-builders).

**Implementation note**: `pkg/core/types/tx_local.go` includes a `ScopeHint` field but
does not yet enforce the gas discount. The discount logic belongs in `core/state_transition.go`
behind an `--experimental-local-tx` flag.

---

## 5. Mempool Routing

### 5.1 Sharded mempool integration

The sharded mempool in `pkg/txpool/` (consistent hashing) can use `ScopeHint` as the
shard key:

```
shard_id = consistent_hash(ScopeHint[0])  // 1-byte prefix → shard bucket
```

Each shard serves one sub-builder. Transactions without a `ScopeHint` (global) go into a
separate "global" shard served by the full proposer.

### 5.2 Anti-fragmentation

Users choosing scope hints from a uniform distribution avoid hot-shard hotspots. A
`ScopeHint` advisory service could suggest under-loaded prefix ranges at submission time,
similar to EIP-8077's nonce announce mechanism.

### 5.3 Propagation

Local txs are propagated via the existing `ETH/72` wire protocol with no wire-format change
(they are just `TxType=0x08` entries). Receiving nodes route them to the appropriate shard
bucket.

---

## 6. Key Trade-offs

| Concern | Assessment |
|---------|-----------|
| Declared BAL accuracy | Users must over-declare (worst-case access). Under-declaration causes revert. Wallets need tooling to estimate BAL from simulation. |
| Re-entrancy / dynamic dispatch | If `CALL` to an out-of-scope address is prohibited, contracts that use factory patterns or upgradeable proxies cannot use local txs. Applies to ~30% of contract calls (estimate). |
| Scope hint granularity | 1-byte prefix = 256 buckets. Fine enough for ETH2030 at Glamsterdam throughput; may need 2-byte prefix at Gigagas L1. |
| MEV extraction | Local txs weaken cross-slot MEV since searchers cannot assume ordering across scopes. This is a feature for users, a cost for searchers. |
| Backwards compatibility | Type 0x08 is an additive change; existing wallets using Types 0–4 are unaffected. |

---

## 7. EIP Sketch

**Title**: Local Transactions with Predeclared State Access (type 0x08)

**Abstract**: Introduce transaction type `0x08` carrying a `ScopeHint` (1-byte address-
prefix set) and an optional `AccessList`. Transactions may only access storage slots within
declared entries; violations revert. Local transactions receive a gas discount proportional
to coverage. Enables distributed block building by decomposing the ordering problem into
independent non-overlapping subproblems.

**Parameters**:
- `LOCAL_TX_TYPE = 0x08`
- `LOCAL_TX_DISCOUNT_BPS = 5000` (50% of base gas, configurable)
- `LOCAL_TX_VIOLATION_PENALTY = 2x undeclared access gas cost`

**Required EIPs**: EIP-7928 (BAL), EIP-8077 (announce nonce), sharded mempool

**Status**: Pre-EIP, pending tooling for wallet-side BAL estimation.

---

## 8. References

- Vitalik Buterin, "Distributed block building" (strawmap.org L1 roadmap, 2026)
- EIP-7928 Block Access Lists — `pkg/bal/`
- `pkg/core/types/tx_local.go` — prototype implementation (type 0x08)
- EIP-8077 announce nonce — `pkg/eth/announce_nonce.go`
- Sharded mempool — `pkg/txpool/sharded_mempool.go`
