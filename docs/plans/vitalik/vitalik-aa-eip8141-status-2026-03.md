# Account Abstraction (EIP-8141) — Line-by-Line Status Analysis

> Source: Vitalik's message on Account Abstraction via EIP-8141 Frame Transactions
> Analyzed: 2026-03-04
> Prior gap analysis: `docs/plans/gap-analysis-eip8141-eip7701-vitalik.md` (2026-02-28, 7 findings)

---

## Executive Summary

| Area | Status | Notes |
|------|--------|-------|
| EIP-8141 core (frame tx type, opcodes, execution) | **COMPLETE** | All 4 critical gaps from Feb-28 analysis fixed |
| 2D nonces (parallel nonce spaces) | **COMPLETE** | 256-bit nonce in FrameTx |
| Paymaster gas settlement | **COMPLETE** | Payer charged/refunded in processor.go |
| VERIFY opcode restrictions | **COMPLETE** | NewFrameVerifyJumpTable() blocks dangerous opcodes |
| APPROVE target check | **COMPLETE** | Direct ADDRESS == frame.Target comparison |
| EIP-7997 deterministic factory | **COMPLETE** | Predeploy at 0x12, deployed at Glamsterdam |
| FOCIL complementarity | **COMPLETE** | FrameTx implements TxData; FOCIL wired |
| ZK-SNARK paymaster circuits | **COMPLETE** | Groth16 AA circuits in proofs/ |
| PQ signature support | **COMPLETE** | ML-DSA-65, Falcon512, SPHINCS+ in crypto/pqc/ |
| Encrypted mempool (MEV/privacy) | **COMPLETE** | Commit-reveal + threshold decryption |
| Txpool VERIFY simulation (EVM) | **PARTIAL** | Structural checks only; EVM execution deferred |
| EOA codeless VERIFY error | **PARTIAL** | Generic caller error, no "target has no code" message |
| Paymaster staking registry | **MISSING** | No DoS stake registry for mempool |
| Conservative / aggressive mempool tiers | **MISSING** | Single mempool rule set |
| EOA compatibility in EIP-8141 | **MISSING** | Not yet implemented; Vitalik says "in principle possible" |

**Previous 7 findings (Feb-28):** 4 CRITICAL all fixed; 2 PARTIAL remain; 1 LOW remains.
**New gaps from this analysis:** 3 MISSING items identified.

---

## Line-by-Line Analysis

### 1. "A transaction is N calls, which can read each other's calldata, and which have the ability to authorize a sender and authorize a gas payer. At the protocol layer, that's it."

**Status: COMPLETE**

The EIP-8141 frame transaction type is fully implemented.

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/core/types/tx_frame.go` | 1–280 | `FrameTx` struct: N frames, 256-bit nonce, sender, fee fields, blob hashes |
| `pkg/core/types/tx_frame.go` | 27–45 | `Frame` struct: Mode (Default/Verify/Sender), Target, GasLimit, Data |
| `pkg/core/vm/eip8141_opcodes.go` | 69–180 | `opApprove` (0xaa): authorizes sender (scope 0), payer (scope 1), or both (scope 2) |
| `pkg/core/frame_execution.go` | 1–220 | `ExecuteFrameTx()`: sequential frame execution, approval tracking, receipt building |
| `pkg/core/types/tx_frame.go` | 135–175 | `TXPARAM*` indices 0x00–0x15 for cross-frame calldata access |

---

### 2. "First, a 'normal transaction from a normal account' (e.g. a multisig, or an account with changeable keys, or with a quantum-resistant signature scheme). This would have two frames: Validation + Execution."

**Status: COMPLETE**

The two-frame validation+execution flow works natively. The VERIFY frame calls APPROVE with scope 0 (sender only) or scope 2 (sender+payer), then the DEFAULT/SENDER frame executes.

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/core/vm/eip8141_opcodes.go` | 93–130 | Scope 0: execution approval; scope 2: combined sender+payer approval |
| `pkg/core/frame_execution.go` | 65–100 | Frame loop: executes VERIFY then SENDER frames in order |
| `pkg/core/vm/jump_table.go` | 703–750 | `NewFrameVerifyJumpTable()`: restricted opcodes for VERIFY frames (BLOCKHASH, COINBASE, TIMESTAMP, NUMBER, DIFFICULTY, GASLIMIT, CHAINID, BASEFEE, BLOBBASEFEE, ORIGIN, GASPRICE, CREATE, CREATE2, SELFDESTRUCT, SSTORE) |

Multisig and changeable keys are supported via standard CALL semantics within a VERIFY frame. A wallet contract verifies signatures and calls APPROVE(0).

For **quantum-resistant keys**, the VERIFY frame can call PQ signature verification:

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/crypto/pqc/` | — | ML-DSA-65 (FIPS 204), Dilithium3, Falcon512, SPHINCS+ real lattice ops |
| `pkg/core/types/pq_tx_validator.go` | 1–180 | `PQTxValidatorReal`: dispatches PQ verify per algorithm |
| `pkg/core/types/pq_transaction.go` | 1–200 | `PQTransaction` type with algorithm selector and PQ public key |

---

### 3. "If the account does not exist yet, then you prepend another frame, 'Deployment', which calls a proxy to create the contract. EIP-7997 (deterministic factory predeploy) is good for this, as it would also let the contract address reliably be consistent across chains."

**Status: COMPLETE**

EIP-7997 is fully implemented. The factory is predeploy at address `0x0000000000000000000000000000000000000012`.

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/core/eip7997.go` | 10–50 | `FactoryAddress = 0x12`, 73-byte minimal CREATE2 factory bytecode |
| `pkg/core/eip7997.go` | 35–50 | `ApplyEIP7997()`: deploys factory at Glamsterdam fork activation |
| `pkg/core/processor.go` | 145–150 | Calls `ApplyEIP7997()` at Glamsterdam block 0 |
| `pkg/core/block_builder.go` | 248, 586 | `ApplyEIP7997()` called in genesis and block-building paths |

A Deployment frame sends `salt(32 bytes) || initcode` to `0x12`, which CREATE2-deploys the wallet contract at a deterministic address that is consistent across all chains using the same factory.

---

### 4. "Paymaster for gas in RAI: Deployment [if needed], Validation (ACCEPT approves sender only, not gas payment), Paymaster validation, Send RAI to paymaster, Execution, Paymaster refunds unused RAI and converts to ETH."

**Status: MOSTLY COMPLETE** — gas settlement wired; no protocol-layer paymaster registry

The six-frame paymaster pattern works as described. The critical gaps from Feb-28 (payer charged the wrong account) are **fixed**.

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/core/vm/eip8141_opcodes.go` | 101–130 | Scope 0: sender-only approval; scope 1: payer-only (paymaster calls this) |
| `pkg/core/frame_execution.go` | 20–30 | `FrameExecutionContext.Payer` tracks the address that called APPROVE(1) |
| `pkg/core/processor.go` | 1178–1183 | Payer gas deduction: if `frameCtx.Payer != sender`, deduct from payer |
| `pkg/core/processor.go` | 1234–1240 | Refund unused gas to payer |
| `pkg/core/types/frame_receipt.go` | 1–60 | `FrameTxReceipt.Payer` records the settled payer address |

The token-to-ETH conversion logic (RAI → ETH) is application-level (paymaster contract), not protocol-layer, which is correct per the EIP.

**Remaining gap (no blocking):** There is no protocol-enforced paymaster staking registry. Vitalik notes a staking mechanism is needed for general DoS safety in mempools — see §8 below.

---

### 5. "Basically the same thing that is done in existing sponsored transactions mechanisms, but with no intermediaries required. Intermediary minimization is a core principle of non-ugly cypherpunk ethereum."

**Status: COMPLETE** (by design of EIP-8141)

No off-chain relayers, bundlers, or entry-point contracts are needed for the core frame transaction mechanism. The transaction itself is the bundle.

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/core/types/tx_frame.go` | 24 | `EntryPointAddress = 0xaa` is the canonical caller, not a separate contract |
| `pkg/core/frame_execution.go` | 40–65 | Frame executor is built into the EVM; no external bundler needed |

---

### 6. "Privacy protocols. First, a paymaster contract which checks for a valid ZK-SNARK and pays for gas if it sees one."

**Status: COMPLETE** — ZK paymaster circuits implemented

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/proofs/aa_proof_circuits.go` | 1–280 | `AAValidationCircuit`: Groth16 circuit proving nonce validity, signature validity, gas payment |
| `pkg/proofs/aa_proofs.go` | 1–200 | `AAProofGenerator`, `AAProof` struct, proof compression |
| `pkg/core/vm/precompile_aa_proof.go` | 1–120 | Precompile at `0x0205`: verifies type 0x01 (code hash), 0x02 (storage), 0x03 (validation result) |

A privacy paymaster can: accept a ZK proof in its calldata, call the `0x0205` precompile to verify it, and call APPROVE(1) to pay gas if valid.

---

### 7. "Second, we could add 2D nonces (RIP-7712), which allow an individual account to function as a privacy protocol, and receive txs in parallel from many users."

**Status: COMPLETE** — 2D nonces fully implemented in FrameTx

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/core/types/tx_frame.go` | 38 | `Nonce *big.Int` (256-bit): upper 192 bits = key, lower 64 bits = sequence |
| `pkg/core/types/tx_frame.go` | 65–80 | `NonceKey()` extracts upper 192 bits; `NonceSeq()` extracts lower 64 bits |
| `pkg/core/aa_entrypoint.go` | 200–230 | `EncodeNonce2D()`, `DecodeNonce2D()` for (key, seq) packing |

Different nonce keys create independent nonce spaces, enabling an account to receive parallel transactions from many users without key correlation. This enables privacy protocols where users send transactions to the same contract (privacy pool) using different nonce keys, making them non-linkable via sequential nonce analysis.

---

### 8. "There are specific rulesets that are known to be safe. For paymasters, there has been deep thought about a staking mechanism to limit DoS attacks in a very general-purpose way."

**Status: PARTIAL** — opcode restrictions COMPLETE; staking registry MISSING

#### VERIFY Opcode Restrictions (COMPLETE)

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/core/vm/jump_table.go` | 703–750 | `NewFrameVerifyJumpTable()`: nils all 15 block-state-dependent opcodes |
| `pkg/core/vm/eip8141_opcodes_test.go` | — | Tests for each restricted opcode returning nil in verify table |
| `pkg/txpool/txpool.go` | 407–430 | Structural VERIFY validation: SENDER-without-VERIFY rejected early |

#### Txpool VERIFY simulation (PARTIAL)

Full EVM simulation of VERIFY frames in the txpool is not yet implemented. The txpool performs structural validation (mode ordering, SENDER requires VERIFY) but does not execute the VERIFY frame in a read-only EVM before admission.

```
// pkg/txpool/txpool.go:407-411
// EIP-8141: VERIFY frame structural validation (PARTIAL-5).
// Codeless VERIFY target rejection requires StateDB.GetCodeSize, which
// is not available via the txpool's StateReader. Full VERIFY simulation
// (checking APPROVE is called) is deferred to block processing in
// processor.go where the EVM is available.
```

**TODO:** Wire a StateDB-aware `simulateVerifyFrame()` in the txpool that:
- Executes the first VERIFY frame in a read-only EVM snapshot
- Rejects if APPROVE is never called
- Rejects if the VERIFY frame reverts

#### Paymaster Staking Registry (MISSING)

There is no protocol-level paymaster staking registry. Without staking:
- Any contract can call APPROVE(1) (paymaster approval)
- A malicious paymaster could claim to pay then fail, causing mempool spam with transactions that look valid but never settle
- The staking mechanism should require paymasters to post a bond that can be slashed for such behavior

**TODO:** Implement `pkg/core/paymaster_registry.go`:
- Staking deposit: minimum stake to register as a mempool-approved paymaster
- Reputation tracker: slashing counter per paymaster address
- Txpool integration: `isApprovedPaymaster(addr)` check before accepting sponsored frame tx
- See ERC-4337 EntryPoint staking semantics as reference

---

### 9. "Realistically, when 8141 is rolled out, the mempool rules will be very conservative, and there will be a second optional more aggressive mempool. The former will expand over time."

**Status: MISSING** — only one rule set, no dual-tier mempool

Current txpool has a single rule set. The EIP-8141 specification calls for:
1. **Conservative mempool**: Only transactions matching known-safe patterns (e.g., VERIFY frame must come before execution frames, cannot call out to outside contracts) are accepted by default
2. **Aggressive optional mempool**: A second tier that accepts broader patterns (e.g., paymaster contracts with registered stake, complex multi-contract VERIFY patterns)

**TODO:** Add mempool tier configuration:
- `pkg/txpool/frame_rules.go`: Conservative rule set (VERIFY-first ordering, no external calls in VERIFY, max gas for VERIFY frame)
- `pkg/txpool/frame_rules_aggressive.go`: Extended rules with paymaster stake checks
- `node.go`: Config flag `--frame-mempool=conservative|aggressive` (default: conservative)

---

### 10. "For privacy protocol users, this means that we can completely remove 'public broadcasters' that are the source of massive UX pain in railgun/PP/TC, and replace them with a general-purpose public mempool."

**Status: PARTIAL** — components exist; end-to-end privacy flow not wired

The building blocks are present:

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/txpool/encrypted/` | — | Commit-reveal ordering, threshold decryption, MEV protection |
| `pkg/core/types/tx_frame.go` | 38 | 2D nonces enable non-linkable parallel nonce spaces |
| `pkg/proofs/aa_proof_circuits.go` | — | ZK proofs for private validation |

However, the end-to-end flow that replaces public broadcasters is not explicitly wired:
- Privacy pool users need to submit transactions to the public mempool with a ZK-SNARK in the paymaster calldata
- The 2D nonce space must be tied to the privacy pool contract, not the user's EOA
- The encrypted mempool commit-reveal must be compatible with frame tx privacy requirements

**TODO:** Document and test a reference privacy pool flow using:
1. Frame tx with nonce key = privacy pool address
2. Paymaster validation frame verifies ZK-SNARK (using 0x0205 precompile)
3. No dependency on external relayer/broadcaster

---

### 11. "For quantum-resistant signatures, we also have to solve one more problem: efficiency."

**Status: PARTIAL** — PQ algorithms complete; efficiency optimization not yet addressed

The PQ signature algorithms are implemented:

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/crypto/pqc/` | — | ML-DSA-65 (FIPS 204), Dilithium3, Falcon512 (897-byte key, 690-byte sig), SPHINCS+ |
| `pkg/core/types/pq_transaction.go` | 1–200 | Algorithm-specific gas costs: Dilithium=8000, Falcon=12000, SPHINCS+=45000 |
| `pkg/crypto/pqc/hybrid_signer.go` | — | Hybrid ECDSA + PQ signer |

**Known efficiency challenges** (referenced in Vitalik's linked posts):
- SPHINCS+ signatures are 49,216 bytes — 70x larger than ECDSA (65 bytes)
- Falcon-512 signatures are 690 bytes — 10x larger
- The NTT precompile (`pkg/core/vm/` NTT precompile) helps reduce on-chain verification cost

**TODO:**
- Evaluate NTT-accelerated Falcon verification as the default QR algorithm for frame transactions
- Add `pkg/core/vm/precompile_falcon.go` that uses NTT precompile internally for VERIFY frame efficiency
- Consider signature aggregation (like BLS aggregate) for PQ to amortize cost across bundles

---

### 12. "AA is also highly complementary with FOCIL: FOCIL ensures rapid inclusion guarantees for transactions, and AA ensures that all of the more complex operations people want to make actually can be made directly as first-class transactions."

**Status: COMPLETE**

Both FOCIL and Frame Transactions are implemented. They are complementary by design:

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/focil/` | — | EIP-7805 inclusion lists: BuildInclusionList(), ValidateInclusionList() |
| `pkg/core/types/tx_frame.go` | 46–52 | FrameTx implements `TxData` interface (txType, gas, gasPrice, etc.) |
| `pkg/focil/types.go` | — | `InclusionList.Transactions` holds transaction hashes generically |

FOCIL operates on transaction hashes; it is agnostic to transaction type. Frame transactions are valid FOCIL inclusion targets. A validator can include a frame tx hash in its inclusion list to guarantee timely inclusion of complex AA operations.

---

### 13. "Another interesting topic is EOA compatibility in 8141. This is being discussed, in principle it is possible, so all accounts incl existing ones can be put into the same framework and gain the ability to do batch operations, transaction sponsorship, etc, all as first-class transactions that fully benefit from FOCIL."

**Status: MISSING** — Vitalik notes "in principle possible", not yet specified or implemented

Current state: An EOA can already sponsor a frame transaction (the EOA is the sender, and the execution frame contains whatever calldata). But an EOA cannot use 8141 frame syntax to do batch calls or use a paymaster without first deploying code.

EIP-7702 (SetCode) partially bridges this: an EOA can temporarily delegate to a smart contract and then issue frame transactions. But this is two separate transactions.

The "pure EOA in 8141" path — where an existing EOA with no code can directly issue a frame transaction using a default validation logic baked into the protocol — is not yet specified in EIP-8141 and not yet implemented.

**TODO (pending EIP finalization):**
- Track EIP-8141 EOA compatibility discussions
- If specified: implement default ECDSA validation as a protocol-level VERIFY frame behavior for codeless accounts
- Test: existing EOA sends frame tx with batch execution, no deployment needed

---

### 14. "After over a decade of research and refinement of these techniques, this all looks possible to make happen within a year (Hegota fork)."

**Implementation target: Hegotá** (2026–2027)

Per CLAUDE.md, Hegotá includes EIP-8141 as a target. Our current implementation covers the core protocol pieces. The remaining items are:
1. Paymaster staking registry (txpool DoS protection)
2. Dual-tier mempool (conservative + aggressive)
3. EOA compatibility (pending EIP spec)
4. PQ signature efficiency (NTT-accelerated Falcon in AA context)
5. Full VERIFY simulation in txpool

---

## Complete Items Reference Table

All implemented items with their primary source files:

| Item | File(s) | Key types/functions |
|------|---------|---------------------|
| FrameTx type (type 0x06) | `pkg/core/types/tx_frame.go` | `FrameTx`, `Frame`, `EntryPointAddress` |
| 256-bit 2D nonce | `pkg/core/types/tx_frame.go:38` | `Nonce *big.Int`, `NonceKey()`, `NonceSeq()` |
| 2D nonce encode/decode | `pkg/core/aa_entrypoint.go` | `EncodeNonce2D()`, `DecodeNonce2D()` |
| APPROVE opcode (0xaa) | `pkg/core/vm/eip8141_opcodes.go:69` | `opApprove`, scopes 0/1/2 |
| TXPARAM* opcodes | `pkg/core/vm/eip8141_opcodes.go:135` | `opTxParamLoad`, `opTxParamSize`, `opTxParamCopy` |
| APPROVE target check | `pkg/core/vm/eip8141_opcodes.go:97` | `contract.Address == fc.Frames[idx].Target` |
| VERIFY jump table | `pkg/core/vm/jump_table.go:703` | `NewFrameVerifyJumpTable()` |
| Frame execution engine | `pkg/core/frame_execution.go` | `ExecuteFrameTx()`, `FrameExecutionContext` |
| Payer gas settlement | `pkg/core/processor.go:1178,1239` | `frameCtx.Payer` deduction + refund |
| Frame receipts | `pkg/core/types/frame_receipt.go` | `FrameTxReceipt`, `FrameResult` |
| Transient storage isolation | `pkg/core/frame_execution.go:57` | `ClearTransientStorage()` between frames |
| EIP-7997 factory | `pkg/core/eip7997.go` | `FactoryAddress`, `ApplyEIP7997()` |
| EIP-7702 SetCode | `pkg/core/eip7702.go` | `ProcessAuthorizations()` |
| CURRENT_ROLE / ACCEPT_ROLE | `pkg/core/vm/eip7701_opcodes.go` | `opCurrentRole`, `opAcceptRole` |
| AA executor phases | `pkg/core/vm/aa_executor.go` | `ValidatePhase()`, `ExecutionPhase()`, `PostOpPhase()` |
| AA proof circuits (Groth16) | `pkg/proofs/aa_proof_circuits.go` | `AAValidationCircuit`, nonce/sig/gas constraints |
| AA proof precompile | `pkg/core/vm/precompile_aa_proof.go` | Address 0x0205, types 0x01/0x02/0x03 |
| PQ crypto (ML-DSA-65, etc.) | `pkg/crypto/pqc/` | Real lattice ops, FIPS 204, hybrid signer |
| PQ transaction type | `pkg/core/types/pq_transaction.go` | `PQTransaction`, algorithm gas costs |
| Encrypted mempool | `pkg/txpool/encrypted/` | Commit-reveal, threshold decryption |
| SetCode P2P gossip | `pkg/p2p/setcode_broadcast.go` | `SetCodeMessage`, bloom dedup |
| FOCIL (EIP-7805) | `pkg/focil/` | `BuildInclusionList()`, `ValidateInclusionList()` |
| Txpool structural VERIFY | `pkg/txpool/txpool.go:407` | SENDER-without-VERIFY rejection |

---

## TODO List

| Priority | Item | File to create/modify | Notes |
|----------|------|----------------------|-------|
| P0 | Full VERIFY simulation in txpool | `pkg/txpool/txpool.go` | Need StateDB injection into txpool |
| P0 | Paymaster staking registry | `pkg/core/paymaster_registry.go` | Min stake, slashing, txpool integration |
| P1 | Dual-tier mempool (conservative/aggressive) | `pkg/txpool/frame_rules.go` | Config flag, conservative default |
| P1 | EOA codeless VERIFY error message | `pkg/core/frame_execution.go` | `"frame tx: VERIFY target 0x... has no code (EOA)"` |
| P2 | Privacy pool reference flow | `docs/` | Document end-to-end: 2D nonce + ZK paymaster + no broadcaster |
| P2 | NTT-accelerated Falcon for AA | `pkg/core/vm/` | Efficiency for PQ in VERIFY frames |
| P3 | EOA compatibility in EIP-8141 | TBD | Pending EIP spec finalization |

---

## Spec References (from `refs/`)

### EIP-8141 — Frame Transaction Exact Parameters

Source: `refs/EIPs/EIPS/eip-8141.md`

| Constant | Value | Notes |
|---|---|---|
| `FRAME_TX_TYPE` | `0x06` | Transaction type byte |
| `FRAME_TX_INTRINSIC_COST` | `15,000 gas` | Base cost per frame tx |
| `ENTRY_POINT` | `address(0xaa)` | Canonical entry-point address |
| `MAX_FRAMES` | `1,000` | Max frames per frame transaction |

**TXPARAM* exact index mapping** (from spec):

| Index | Field | Width | Notes |
|---|---|---|---|
| `0x00` | transaction type | 32 bytes | |
| `0x01` | nonce | 32 bytes | |
| `0x02` | sender | 32 bytes | 20-byte address, zero-padded |
| `0x03` | max_priority_fee_per_gas | 32 bytes | |
| `0x04` | max_fee_per_gas | 32 bytes | |
| `0x05` | max_fee_per_blob_gas | 32 bytes | |
| `0x06` | max cost (base_fee=max, all gas used) | 32 bytes | Includes blob cost + intrinsic |
| `0x07` | len(blob_versioned_hashes) | 32 bytes | |
| `0x08` | compute_sig_hash(tx) | 32 bytes | Canonical signature hash |
| `0x09` | len(frames) | 32 bytes | |
| `0x10` | currently executing frame index | 32 bytes | |
| `0x11` | frame[index].target | 32 bytes | |
| `0x12` | frame[index].data | dynamic | 0 size if VERIFY mode |
| `0x13` | frame[index].gas_limit | 32 bytes | |
| `0x14` | frame[index].mode | 32 bytes | 0=DEFAULT, 1=VERIFY, 2=SENDER |
| `0x15` | frame[index].status | 32 bytes | 0=fail, 1=success |

**Gas accounting formula** (from spec):
```
tx_gas_limit = 15000 + calldata_cost(rlp(tx.frames)) + sum(frame.gas_limit)
tx_fee       = tx_gas_limit * effective_gas_price + len(blobs) * GAS_PER_BLOB * blob_base_fee
refund       = sum(frame.gas_limit) - total_gas_used   // per-frame unused gas
```

**ETH2030 alignment**: `pkg/core/types/tx_frame.go` uses `TXPARAM*` indices 0x00–0x15 (`pkg/core/types/tx_frame.go:135–175`). `pkg/core/processor.go:1178–1183` implements payer gas deduction matching the spec gas accounting formula.

---

### EIP-7701 — Native AA Exact Parameters

Source: `refs/EIPs/EIPS/eip-7701.md`

| Constant | Value |
|---|---|
| `AA_ENTRY_POINT` | `address(0x7701)` |
| `AA_BASE_GAS_COST` | `15,000 gas` |
| `ROLE_SENDER_DEPLOYMENT` | `0xA0` |
| `ROLE_SENDER_VALIDATION` | `0xA1` |
| `ROLE_PAYMASTER_VALIDATION` | `0xA2` |
| `ROLE_SENDER_EXECUTION` | `0xA3` |
| `ROLE_PAYMASTER_POST_OP` | `0xA4` |

**CURRENT_ROLE** opcode returns the role constant above. **ACCEPT_ROLE** is like RETURN but validates the role matches expected.

**ETH2030 alignment**: `pkg/core/vm/eip7701_opcodes.go` implements `opCurrentRole` and `opAcceptRole`. These match the spec role constants.

---

### EIP-7702 — SetCode Exact Parameters

Source: `refs/EIPs/EIPS/eip-7702.md`

| Constant | Value |
|---|---|
| `SET_CODE_TX_TYPE` | `0x04` |
| `MAGIC` | `0x05` |
| `PER_AUTH_BASE_COST` | `12,500 gas` |
| `PER_EMPTY_ACCOUNT_COST` | `25,000 gas` |

**Delegation indicator format**: `0xef0100 || address` (23 bytes total)
**Auth message**: `keccak(0x05 || rlp([chain_id, address, nonce]))`
**Chain ID**: 0 = valid on all chains; any other value = chain-specific

---

### ML-DSA-65 (FIPS 204) — Exact Parameters

Source: `refs/circl/sign/mldsa/mldsa65/internal/params.go`

| Parameter | Value | Description |
|---|---|---|
| `SeedSize` | `32 bytes` | Key generation seed |
| `PublicKeySize` | `1,952 bytes` | Serialized public key |
| `PrivateKeySize` | `4,000 bytes` | Serialized private key |
| `SignatureSize` | `3,309 bytes` | Serialized signature |
| `K` | `6` | Vector dimension for A, s2 |
| `L` | `5` | Vector dimension for s1 |
| `Eta` | `4` | Compression parameter |
| `Tau` | `49` | Number of 1s in challenge c |
| `Omega` | `55` | Max ones in sparse vector |
| `Gamma1Bits` | `19` | Bit width for γ₁ |
| `Gamma2` | `261,888` | `(q-1)/88` for hint packing |

**Go API** (`refs/circl/sign/mldsa/mldsa65/dilithium.go`):
```go
func GenerateKey(rand io.Reader) (*PublicKey, *PrivateKey, error)
func NewKeyFromSeed(seed *[32]byte) (*PublicKey, *PrivateKey)
func SignTo(sk *PrivateKey, msg, ctx []byte, randomized bool, sig []byte) error
func Verify(pk *PublicKey, msg, ctx []byte, sig []byte) bool
```

Supports optional **context string** (up to 255 bytes) and **randomized mode** (additional randomness in signing).

**ETH2030 alignment**: `pkg/crypto/pqc/mldsa_signer.go` uses this API via `circl`. Check that `SignTo` context is empty (consistent with EVM sig hash as message).

---

### NTT Precompile — Exact Addresses and Gas

Source: `refs/ntt-eip/EIP/EIPNTT.md`

| Precompile | Address | Gas Formula | Operation |
|---|---|---|---|
| `NTT_FW` | `0x0f` | `600 gas flat` | Forward NTT (Negative Wrap Convolution) |
| `NTT_INV` | `0x10` | `600 gas flat` | Inverse NTT |
| `NTT_VECMULMOD` | `0x11` | `k * log₂(n) / 8` | Element-wise modular multiplication |
| `NTT_VECADDMOD` | `0x12` | `k * log₂(n) / 32` | Element-wise modular addition |

Where `k = smallest power of 2 > log₂(q)` (bits needed to represent the modulus).

**Supported moduli** (relevant to AA PQ efficiency):
- Falcon: `q = 12,289` (3·2¹²+1), n=512
- Dilithium: `q = 8,380,417` (2²³−2¹³+1), n=256
- Goldilocks: `q = 2⁶⁴−2³²+1`

**Gas benchmark: Falcon-512 on-chain verification**:
- Pure Solidity: ~1.8M gas
- With NTT precompile (EPERVIER, Yul + hint): **~1.5M gas**
- NTT calls only: ~1,500 gas (4 × NTT_FW + 1 × NTT_INV at 600 each + VECMULMOD)

**ETH2030 gap**: `pkg/core/vm/precompile_ntt.go` registers the NTT at address `0x15` but the spec defines `0x0f`–`0x12`. The ETH2030 implementation also combines forward/inverse into one precompile rather than separate addresses. Need to align with the ntt-eip spec before Hegotá.

**Falcon CVEs** (from `refs/ethfalcon/`): 3 CVEs filed for on-chain Falcon verification:
- `CVETH-2025-080201` (CRITICAL): Salt size not checked in verification
- `CVETH-2025-080202` (MEDIUM): Signature malleability on coefficient signs
- `CVETH-2025-080203` (LOW): Missing domain separation in XOF

These must be addressed if Falcon is used as the default PQ algorithm in VERIFY frames.
