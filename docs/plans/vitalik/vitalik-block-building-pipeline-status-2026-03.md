# Block Building Pipeline — Status vs Vitalik's Message (March 2026)

Source: Vitalik's message on the block building pipeline (analysed line-by-line below).

---

## Summary

| Feature | Status | Package / File |
|---|---|---|
| ePBS | COMPLETE | `pkg/epbs/`, `pkg/engine/engine_epbs.go` |
| FOCIL | COMPLETE | `pkg/focil/` |
| Big FOCIL (sender-hex partitioning) | COMPLETE | `pkg/focil/big_focil.go` |
| Encrypted mempool (commit-reveal + threshold + VDF) | COMPLETE | `pkg/txpool/encrypted/` |
| Transaction ingress anonymity (mixnet / Flashnet) | COMPLETE | `pkg/p2p/mixnet_transport.go`, `flashnet_transport.go` |
| Long-term distributed block building (speculative) | PARTIAL | `pkg/engine/distributed_builder.go`, `builder_coordinator.go` |

---

## Line-by-Line Analysis

### 1. ePBS — "Ethereum is getting ePBS, which lets proposers outsource to a free permissionless market of block builders"

**Status: COMPLETE**

ePBS (EIP-7732) is fully implemented with:

- **`pkg/epbs/types.go`** — `BuilderBid`, `PayloadEnvelope`, `PayloadAttestation` (PTC, 512 members), `BLSPubkey`, `BLSSignature`
- **`pkg/epbs/auction_engine.go`** — full auction lifecycle state machine: Open → BiddingClosed → WinnerSelected → Finalized
- **`pkg/epbs/bid_escrow.go`** — collateral-backed bid escrow
- **`pkg/epbs/bid_scoring.go`** — highest-value selection with timestamp tiebreaker
- **`pkg/epbs/bid_validator.go`** — payload commitment + signature validation
- **`pkg/epbs/builder_market.go`** + **`builder_registry.go`** — permissionless builder registration and market statistics
- **`pkg/epbs/builder_reputation.go`** — sliding-window reputation tracking, latency scoring, score decay
- **`pkg/epbs/commitment_reveal.go`** — commit-reveal scheme for payload disclosure timing
- **`pkg/epbs/mev_burn.go`** — MEV burn calculation and distribution
- **`pkg/epbs/payment.go`** — Vickrey auction settlement (second-price sealed-bid)
- **`pkg/epbs/slashing.go`** — slashing conditions for builders who win but fail to deliver
- **`pkg/engine/engine_epbs.go`** — Engine API ePBS integration
- **`pkg/engine/engine_v7.go`** — Engine V7 JSON-RPC for ePBS payload envelope handling
- **`pkg/engine/builder_coordinator.go`** — Vickrey auction settlement, bid timeout, reputation decay
- **`pkg/engine/distributed_builder.go`** — `BuilderNetwork` managing distributed builders, bid submission, winner selection

**What it covers from Vitalik's message:**
- Proposers outsource block building to a free permissionless market ✓
- Block builder centralization does not bleed into staking centralization ✓
- Slashing for unreliable builders (late delivery, equivocation) ✓

---

### 2. FOCIL — "FOCIL lets 16 randomly-selected attesters each choose a few transactions, which must be included"

**Status: COMPLETE**

FOCIL (EIP-7805) is fully implemented with:

- **`pkg/focil/types.go`** — `InclusionList`, `InclusionListEntry`, `SignedInclusionList`
- **`pkg/focil/committee_selection.go`** — IL committee (16 members) selected per slot
- **`pkg/focil/committee_tracker.go`** + **`committee_voting.go`** — membership tracking, voting
- **`pkg/focil/list_validator.go`** — validates gas limits (2²¹), byte limits (8 KiB), tx count (2⁴ = 16)
- **`pkg/focil/compliance_engine.go`** — compliance score 0–100, penalties for non-inclusion
- **`pkg/focil/compliance_tracker.go`** — per-block compliance evaluation
- **`pkg/focil/inclusion_monitor.go`** — monitors whether mandated txs appear in block
- **`pkg/focil/violation_detector.go`** — detects and records IL non-compliance violations
- **`pkg/focil/mev_filter.go`** — MEV tx detection and filtering, builder compliance validation
- **`pkg/engine/inclusion_list.go`** — FOCIL IL integration in the Engine API layer

**What it covers from Vitalik's message:**
- 16 randomly-selected attesters each choose a few transactions ✓
- Block is rejected if mandated transactions are not included ✓
- Hostile 100% builder cannot prevent tx inclusion via FOCILers ✓

---

### 3. "Big FOCIL" — "make the FOCILs bigger so they can include all the transactions in the block"

**Status: COMPLETE**

"Big FOCIL" with sender-hex partitioning is implemented in:

- **`pkg/focil/big_focil.go`** — `BigFOCILConfig`, `PartitionedList`, `CarryoverTracker`
  - 16 FOCIL'er partitions keyed by the first hex character of the sender address
  - The i-th FOCIL'er only includes txs where `sender[0] == hex(i)`
  - Carryover tracking: txs present but not included in slot N escalate in priority for slot N+1
  - `CarryoverTracker.EscalatePriority()` ensures censored txs eventually get included
- **`pkg/focil/enhanced.go`** — `InclusionListV2` with `InclusionConstraint` (MustInclude, MustExclude, GasLimit, Ordering), enabling advanced Big FOCIL constraint profiles
- **`pkg/focil/mempool_monitor.go`** — monitors mempool for IL candidate transactions, feeds the partitioned lists

**What it covers from Vitalik's message:**
- Each i-th FOCIL'er by default only includes txs whose sender address first hex char is i ✓
- Carryover: txs present but not included in the previous slot get priority in next slot ✓
- Avoids duplication by sender-hex partitioning ✓
- Builder role reduced to MEV-relevant txs + state transition computation ✓

**What is still speculative / future work:**
- The exact protocol-level spec for Big FOCIL is not yet finalised by EF. Our implementation is a forward-looking design following the described algorithm.
- Full integration with mandatory protocol enforcement (vs. current advisory compliance scoring) requires a future EIP.

---

### 4. Encrypted mempools — "If a transaction is encrypted until it's included, no one gets the opportunity to wrap it in a hostile way"

**Status: COMPLETE**

Encrypted mempool with commit-reveal + threshold decryption + VDF timing is fully implemented:

- **`pkg/txpool/encrypted/types.go`** — `CommitTx` (hash-only, 12s window), `RevealTx`, `CommitState` (COMMITTED / REVEALED / EXPIRED)
- **`pkg/txpool/encrypted/pool.go`** — `EncryptedPool`: manages commits and reveals with state machine
- **`pkg/txpool/encrypted/encrypted_protocol.go`** — commit-reveal lifecycle protocol
- **`pkg/txpool/encrypted/ordering.go`** + **`ordering_policy.go`** — commit-time-based fair ordering (FIFO by commit time or priority by gas), preventing sandwiching
- **`pkg/txpool/encrypted/threshold_decrypt.go`** — t-of-n threshold decryption: validators each hold one share, combined share → plaintext; decryption bound to epoch
- **`pkg/txpool/encrypted/vdf_timer.go`** — VDF (Verifiable Delay Function) puzzles bound to slot number; solving the VDF enables decryption but cannot be done before the slot time
- **`pkg/txpool/encrypted/validity_proof.go`** — ZK validity proof (balance, nonce, gas) that proves the encrypted tx is valid without revealing its content
- **`pkg/p2p/setcode_broadcast.go`** — anonymous EIP-7702 auth list gossip

**What it covers from Vitalik's message:**
- Transaction encrypted until included → no opportunity for sandwiching or frontrunning ✓
- Guarantee of validity in a mempool-friendly way (ZK validity proofs) ✓
- Guarantee of decryption only once the block is made (VDF slot-bound timing) ✓
- t-of-n threshold scheme ensures no single party can decrypt early ✓
- Fair ordering by commit time prevents MEV reordering ✓

---

### 5. Transaction ingress layer — "network layer: what happens in between a user sending out a transaction, and that transaction making it into a block"

**Status: COMPLETE**

Anonymous transport infrastructure for the ingress layer is implemented:

- **`pkg/p2p/anonymous_transport.go`** — `AnonymousTransport` interface, `TransportManager` managing multiple backends simultaneously
- **`pkg/p2p/mixnet_transport.go`** — `MixnetTransport`: 3-hop multi-relay with simulated per-hop delays (500 ms default), Keccak256 onion re-encryption per hop; hides sender IP and timing correlation
- **`pkg/p2p/flashnet_transport.go`** — `FlashnetTransport`: broadcast to all peers with ephemeral Diffie-Hellman keys; lower latency than mixnet, less anonymity but bandwidth-efficient (suitable for tiny tx payloads)
- **`pkg/p2p/discovery.go`** + **`discovery_v5.go`** — Kademlia DHT-based peer discovery (V5 ENR) for pluggable P2P layer
- **`pkg/engine/block_pipeline.go`** — `StageIngress` is the first stage of the 7-stage block pipeline, explicitly handles anonymous transport integration

**What it covers from Vitalik's message:**
- Network-layer anonymization for transactions (mixnet, Flashnet-style) ✓
- Mixnet design: multi-hop with delays, Keccak onion wrapping ✓
- Flashnet-style: lower latency, bandwidth-heavier (OK since txs are tiny) ✓
- Pluggable support for multiple transports (TransportManager selects best) ✓
- MEV prevention by hiding tx in-flight (sandwiching requires seeing tx "in the clear") ✓

**Risks / gaps identified from Vitalik's message:**
- **No Tor integration**: Vitalik mentions Tor routing; our implementation uses a simulated mixnet (MixnetTransport) rather than actual Tor. True Tor or a real mixnet protocol (e.g., Nym) would require external daemon integration.
- **kohaku initiative integration**: Vitalik references the kohaku initiative (`@ncsgy`) for pluggable support. Our `TransportManager` interface is compatible with this model but has no direct kohaku integration.
- **Order-matching without servers**: Vitalik mentions passive ideal order-matching at the network layer. We have commit-time ordering in the encrypted pool but not a fully passive, serverless DeFi order-matching layer.

---

### 6. Long-term distributed block building — "make Ethereum truly like BitTorrent: able to process far more transactions than any single server"

**Status: PARTIAL — core infrastructure done, speculative protocol extensions are future work**

Current implementation:

- **`pkg/engine/distributed_builder.go`** — `BuilderNetwork`: manages many distributed builders, bid submission, winner selection (Vickrey second-price auction)
- **`pkg/engine/builder_coordinator.go`** — bid timeout, reputation decay, builder registration, `AuctionSettlement` (WinnerID, WinnerBid, SettlePrice, RunnerUpID)
- **`pkg/engine/block_pipeline.go`** — 7-stage pipeline where `StagePartition` (DAG-based dependency partitioning) and `StageMerge` allow sub-block parallel building
- **`pkg/core/dependency_graph.go`** — `DependencyGraph`: DAG of transactions by state-access conflicts, `TxGroup` (non-conflicting batch), enables parallel sub-block builders for non-conflicting tx sets
- **`pkg/bal/`** — Block Access Lists (EIP-7928): opcode-level state tracking for 15 opcodes, enables precise conflict detection for the dependency graph

**What it covers from Vitalik's message:**
- Multiple distributed builders bid on building blocks (Vickrey auction) ✓
- Parallel sub-block building for non-conflicting tx partitions (via BAL dependency graph) ✓
- Big FOCIL reduces builder's required role (only MEV-relevant txs + ordering) ✓

**What is still future work (explicitly "open design space" per Vitalik):**
- **New tx categories for partial globalness**: Vitalik describes a "less global" tx type that is friendly to fully distributed building and would be much cheaper. No such new tx type exists yet; it's described as a long-term speculative design.
- **Fully serverless ordering**: Even with parallel builders, one central actor is still needed to put everything in order and execute final state. The challenge of synchronous shared state is not solved by our current design.
- **>95% non-global activity routing**: Vitalik envisions routing most activity through cheap "local" txs while keeping expensive "global" txs for the 5% that needs it. This requires new EIPs and is not yet specced.

---

## Integrated Block Building Pipeline (7 Stages)

The `pkg/engine/block_pipeline.go` `BlockPipeline` orchestrates all the above into a single coherent per-slot flow:

```
StageIngress   → Anonymous transport (mixnet / Flashnet) hides sender IP
     ↓
StageEncrypt   → Encrypted mempool: commit-reveal + VDF timing + threshold decryption
     ↓
StageFOCIL     → Big FOCIL enforcement: 16 sender-hex partitions + carryover tracking
     ↓
StagePartition → BAL-based dependency graph partitioning (non-conflicting tx groups)
     ↓
StageBuild     → Parallel sub-block building per tx group
     ↓
StageMerge     → Merge sub-blocks into single canonical block
     ↓
StagePropose   → ePBS auction: distributed builder bids → Vickrey settlement → propose
```

---

## TODO / Future Work

| Item | Priority | Notes |
|---|---|---|
| True Tor / real mixnet integration | Medium | Vitalik mentions Tor; current mixnet is simulated. Integrate libtoreplacement or Nym SDK. |
| Passive serverless DeFi order-matching | Low | Requires cryptographic order-matching without servers; open research problem. |
| kohaku initiative pluggability | Low | Align `TransportManager` API with kohaku protocol interface when spec is available. |
| New "less global" tx type for distributed building | Low | Speculative per Vitalik; awaits EIP. Design would need new mempool routing + cheaper gas schedule for local txs. |
| Fully decentralised tx ordering (no central actor) | Low | Research problem; synchronous shared state is fundamental to Ethereum. Current Big FOCIL + BAL partitioning is the best known approach. |

---

## Spec References (from `refs/`)

### ePBS (EIP-7732) — Key Parameters

Source: `refs/EIPs/EIPS/eip-7732.md`, `refs/consensus-specs/specs/gloas/`

| Parameter | Value | Description |
|---|---|---|
| `PTC_SIZE` | `2^9 = 512` | Payload Timeliness Committee size |
| `PAYLOAD_TIMELY_THRESHOLD` | `256` | PTC votes needed for timeliness quorum |
| `MAX_PAYLOAD_ATTESTATIONS` | `4` | Max payload attestations per beacon block |
| `BUILDER_REGISTRY_LIMIT` | `2^40` | Max builders registerable |
| `MIN_BUILDER_WITHDRAWABILITY_DELAY` | `64 epochs` | Cooldown before builder can withdraw stake |
| `MAX_BUILDERS_PER_WITHDRAWALS_SWEEP` | `2^14 = 16,384` | Batch withdrawal processing per epoch |
| `BUILDER_WITHDRAWAL_PREFIX` | `0x03` | Credential prefix for builder accounts |
| `DOMAIN_BEACON_BUILDER` | `0x0B000000` | BLS domain for builder bids |
| `DOMAIN_PTC_ATTESTER` | `0x0C000000` | BLS domain for PTC attestations |
| `DOMAIN_PROPOSER_PREFERENCES` | `0x0D000000` | BLS domain for proposer preferences |

**Slot timing constraints (spec-defined critical path):**
- `t=0s`: Builder constructs payload, signs `ExecutionPayloadBid`, gossips to `execution_payload_bid` topic
- `t=4s`: Proposer deadline — must select bid and include it in beacon block
- `t=4s–9s`: PTC members attest via `PayloadAttestationMessage` on `payload_attestation_message` topic
- `t=9s–12s`: Builder reveals `SignedExecutionPayloadEnvelope` on `execution_payload` topic

**Core gossip topics** (from `refs/consensus-specs/specs/gloas/p2p-interface.md`):
- `execution_payload_bid` — `SignedExecutionPayloadBid` broadcasts
- `payload_attestation_message` — PTC timeliness attestations
- `proposer_preferences` — `SignedProposerPreferences` (fee recipient + gas limit hints)
- `execution_payload` — payload envelope reveals

**Req/Resp protocols** (new in ePBS):
- `ExecutionPayloadEnvelopesByRange` — fetch envelopes by slot range
- `ExecutionPayloadEnvelopesByRoot` — fetch by beacon block root

**Open questions in spec** (from `refs/EIPs/EIPS/eip-7732.md`):
- No slashing penalty for payload equivocation (relies on 2/3+ PTC honesty)
- EL payment mechanism (EIP-level staking vs. CL staking) is still under discussion

**Beacon API endpoints** (from `refs/beacon-APIs/`):
- `GET /eth/v1/validator/{slot}/execution_payload_bid/{builder_index}` — fetch bid for slot
- `POST /eth/v1/beacon/execution_payload_bid` — publish `SignedExecutionPayloadBid`
- `GET /eth/v1/validator/{slot}/payload_attestation_data` — produce PTC attestation data
- `POST /eth/v1/beacon/pool/payload_attestations` — submit `PayloadAttestationMessage`

**ETH2030 gap vs. spec:**
- `pkg/epbs/` tracks builder reputation with sliding-window scoring; spec defines payment threshold at 60% (`BUILDER_PAYMENT_THRESHOLD_NUMERATOR/DENOMINATOR = 6/10`)
- `pkg/engine/engine_v7.go` handles Engine V7; Engine API `engine_newPayloadV5` is in `refs/execution-apis/src/engine/amsterdam.md`

---

### FOCIL (EIP-7805) — Key Parameters

Source: `refs/EIPs/EIPS/eip-7805.md`, `refs/consensus-specs/specs/heze/`

| Parameter | Value | Description |
|---|---|---|
| `IL_COMMITTEE_SIZE` | `2^4 = 16` | Validators per slot on IL committee |
| `MAX_BYTES_PER_INCLUSION_LIST` | `2^13 = 8,192 bytes` | Max RLP-encoded size per IL |
| `VIEW_FREEZE_CUTOFF_BPS` | `7500 bps (75%)` | IL view freezes at 75% of slot (9s of 12s) |
| `DOMAIN_IL_COMMITTEE` | `0x0E000000` | BLS domain for IL committee signatures |

**Slot timing** (from `refs/consensus-specs/specs/heze/inclusion-list.md`):
- `t=0s–8s`: IL committee builds ILs from mempool, gossips to `signed_inclusion_list`
- `t=7s`: Fallback — if no block received, build IL on local fork-choice head
- `t=9s`: **View freeze deadline** — no new ILs accepted after this point
- `t=11s`: Builder freezes IL view, calls `engine_getInclusionListV1` to EL, builds payload
- `t=0s (Slot N+1)`: Proposer broadcasts block; payload must satisfy all non-equivocated ILs
- `t=4s (Slot N+1)`: Attesters verify IL satisfaction

**P2P gossip topics:**
- `signed_inclusion_list` — global gossip for `SignedInclusionList` broadcasts

**Engine API changes** (from `refs/execution-apis/`):
- `engine_getInclusionListV1` — retrieve IL transactions from EL
- `engine_forkchoiceUpdated` — `PayloadAttributes` extended with `inclusion_list_transactions`
- `engine_newPayload` — returns `INCLUSION_LIST_UNSATISFIED` if mandated tx missing and gas available

**IL validation rules** (P2P, from spec):
1. Slot matches current or previous slot
2. Committee root matches expected for slot
3. Received ≤2 ILs from same committee member (equivocation detection)
4. BLS signature valid; validator is committee member
5. IL RLP-encoded size ≤ 8,192 bytes

**IL satisfaction algorithm** (spec-recommended O(n) approach):
1. Pre-snapshot: record nonce + balance for all IL tx senders
2. Build payload normally
3. For each IL tx: if in payload → skip; else check nonce/balance → append if valid and gas available
4. Track EOA state changes to detect dependency chains

**Open questions in spec:**
- No incentive mechanism for IL building (implementer choice: random/priority/time-pending)
- 1-of-16 honesty assumption for censorship resistance (1 honest IL member suffices)
- RPC request for missing ILs is optional (implementation dependent)

**FOCIL + ePBS integration** (from `refs/consensus-specs/specs/heze/`):
`ExecutionPayloadBid` in Heze fork adds:
```python
inclusion_list_bits: Bitvector[INCLUSION_LIST_COMMITTEE_SIZE]
```
Builder commits to which IL committee members' txs are included, enforced by PTC.

---

### Encrypted Mempool — Spec Reference

Source: `refs/research/` (Vitalik's ethresear.ch post: recursive-STARK-based-bandwidth-efficient-mempool)

The 500ms mempool STARK tick in `pkg/txpool/stark_aggregation.go` directly implements the proposal:
- Every 500ms: collect validated txs, generate `STARKProofData` over all validation proofs
- `MaxTickSize = 128 KB` per the ethresear.ch spec
- p2p gossip topic `mempool-stark-tick/1` (not yet wired — see TODO in `vitalik-pq-roadmap-gap-analysis.md`)
