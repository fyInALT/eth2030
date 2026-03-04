> [← Back to Sprint Index](README.md)

# EPIC 7 — EIP Specification Compliance

**Goal**: Close the gaps found during cross-referencing of user stories against the six source EIP documents (EIP-8141, EIP-7732, EIP-7805, EIP-7928, EIP-7706, EIP-7864). Every story in this epic corresponds to a spec requirement that was absent from Epics 1–6.

---

## US-SPEC-1: EIP-8141 Frame TX Full Compliance

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As an **EVM engineer**, I want frame transaction receipts to use the correct 3-layer structure, TSTORE/TLOAD discarded between frames, and all 16 TXPARAM* parameter indices correctly implemented, so that ETH2030 is fully EIP-8141 spec-compliant.

**Priority**: P0 | **Story Points**: 13 | **Sprint Target**: Sprint 2

### Acceptance Criteria
- Frame tx receipt RLP encodes as `[cumulative_gas_used, payer, [[status, gas_used, logs], ...]]`
- `TSTORE` written in frame N is not visible to frame N+1 (`TLOAD` returns zero)
- `ORIGIN` opcode inside a frame returns the frame caller address, not the traditional tx sender
- Frame mode SENDER (0x02) reverts immediately if `sender_approved == false` at entry
- All 16 parameter indices from the EIP-8141 spec table are implemented and testable
- `TXPARAMLOAD(0x08, 0)` returns `compute_sig_hash(tx)` — 32-byte signing hash
- `TXPARAMLOAD(0x09, 0)` returns `len(tx.frames)`
- `TXPARAMLOAD(0x10, 0)` returns currently executing frame index
- `TXPARAMLOAD(0x15, frame_index)` returns frame execution status (0=fail, 1=success)
- `TXPARAMSIZE` returns correct byte length for each parameter

### Tasks

#### Task SPEC-1.1 — Implement frame tx receipt 3-layer encoding
- **Description**: In `pkg/core/types/receipt.go`, add `FrameReceipt` struct `{Status uint64, GasUsed uint64, Logs []*Log}` and update `Receipt` for frame txs to contain `Payer common.Address` and `FrameReceipts []FrameReceipt`. RLP encoder must produce `[cumulative_gas_used, payer, [frame_receipt, ...]]` exactly per EIP-8141 spec §receipt.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (core/types)
- **Testing Method**: Unit test: execute frame tx with 3 frames → receipt has 3 `FrameReceipts` entries, each with correct status/gas/logs. RLP round-trip test. `go test ./core/types/... -run TestFrameTxReceipt`.
- **Definition of Done**: `FrameReceipt` type defined. RLP encoding correct. Round-trip passes. ≥ 80% coverage. EF state tests unaffected.

#### Task SPEC-1.2 — Enforce TSTORE/TLOAD cross-frame discard
- **Description**: In `pkg/core/vm/evm.go` frame execution loop, after each frame completes (success or revert), call `stateDB.ClearTransientStorage()` before starting the next frame. Per EIP-8141 spec: warm/cold state journal is shared across frames, but transient storage (`TSTORE`/`TLOAD`) is discarded at frame boundary.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Test: frame 0 does `TSTORE(key, 42)`, frame 1 does `TLOAD(key)` → returns 0. Contrast: `SLOAD` set in frame 0 is visible to frame 1 (warm journal shared). `go test ./core/vm/... -run TestFrameTSTORECrossFrame`.
- **Definition of Done**: `ClearTransientStorage()` called at each frame boundary. Test passes. EF state tests (36,126) unaffected.

#### Task SPEC-1.3 — ORIGIN opcode returns frame caller in frame context
- **Description**: In `pkg/core/vm/instructions.go` `opOrigin`, when executing inside a FRAME context (detect via `FrameContext.IsActive`), return `FrameContext.Caller` instead of `tx.From`. Outside frame context, ORIGIN continues to return `tx.From` as usual.
- **Estimated Effort**: 1 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Unit test: ORIGIN inside a frame → returns frame caller. ORIGIN outside frame → returns `tx.From`. `go test ./core/vm/... -run TestOriginInFrameContext`.
- **Definition of Done**: ORIGIN correct in both contexts. Test green. EF state tests unaffected.

#### Task SPEC-1.4 — SENDER frame mode enforces sender_approved precondition
- **Description**: In `pkg/core/vm/aa_executor.go` frame execution dispatch, when frame mode = `0x02` (SENDER), check `FrameContext.SenderApproved == true` before executing. If false, immediately revert with error `"frame: SENDER mode requires prior sender_approved"`. This is a stateful precondition not validated at tx admission.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Test: frame tx with SENDER frame before a VERIFY+APPROVE frame → SENDER frame reverts. After VERIFY+APPROVE, SENDER frame executes. `go test ./core/vm/... -run TestFrameSenderModePrecondition`.
- **Definition of Done**: Precondition enforced. Error message matches spec text. Test green.

#### Task SPEC-1.5 — Audit and implement missing TXPARAM indices
- **Description**: In `pkg/core/vm/eip8141_opcodes.go`, compare the existing `TXPARAMLOAD` switch against the 16-entry spec table. Add any missing cases: particularly `in1=0x06` (max cost), `in1=0x07` (blob hash count), `in1=0x08` (`compute_sig_hash`), `in1=0x09` (`len(frames)`), `in1=0x10` (current frame index), `in1=0x15` (frame status). Implement `compute_sig_hash(tx)` as `keccak256(RLP(chain_id, nonce, sender, frames, ...))`.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Table-driven test in `pkg/core/vm/eip8141_txparam_test.go` covering all 16 parameter indices. Each test case: construct a frame tx with known fields, execute `TXPARAMLOAD(in1, in2)`, assert expected value. `go test ./core/vm/... -run TestTXPARAMAllIndices`.
- **Definition of Done**: All 16 parameter indices covered. Table-driven test passes. EF state tests unaffected.

#### Task SPEC-1.6 — TXPARAMCOPY and TXPARAMSIZE completeness
- **Description**: In `pkg/core/vm/eip8141_opcodes.go`, verify `TXPARAMCOPY` (0xb2) correctly handles variable-size parameters (frame data via `in1=0x12`). Verify `TXPARAMSIZE` (0xb1) returns 32 for all fixed-size params and the correct dynamic size for `in1=0x12` (frame data) and blob hashes. Add tests for dynamic-size copy.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Unit test: `TXPARAMCOPY` copies frame data into memory at correct offset; memory matches frame data exactly. `TXPARAMSIZE` for fixed params returns 32; for dynamic returns actual length. `go test ./core/vm/... -run TestTXPARAMCOPY`.
- **Definition of Done**: Copy and size ops correct for all param types. Tests green.

---

## US-SPEC-3: EIP-7732 ePBS Builder Withdrawal & Epoch Processing

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **builder node operator**, I want the builder withdrawal mechanism (64-epoch delay, withdrawal prefix `0x03`, batch sweep of 16,384 builders/epoch) and `process_builder_pending_payments` epoch processing correctly implemented, so that builder balances are managed safely and predictably across epoch boundaries.

**Priority**: P1 | **Story Points**: 8 | **Sprint Target**: Sprint 2

### Acceptance Criteria
- `MIN_BUILDER_WITHDRAWABILITY_DELAY = 64` epochs enforced; withdrawal request before delay → rejected
- `MAX_BUILDERS_PER_WITHDRAWALS_SWEEP = 16,384` per epoch sweep
- `process_builder_pending_payments()` runs in epoch processing and correctly deducts from beacon chain
- `ProposerPreferences` P2P gossip topic (`DOMAIN_PROPOSER_PREFERENCES = 0x0D000000`) registered and handled
- Builder self-build flag (`BUILDER_INDEX_SELF_BUILD = UINT64_MAX`) accepted without auction

### Tasks

#### Task SPEC-3.1 — Implement `process_builder_pending_payments` in epoch processing
- **Description**: In `pkg/consensus/epoch_processing.go` (or equivalent), add `processBuilderPendingPayments(state)` that iterates `state.builder_pending_payments`, deducts amounts from the beacon chain, and updates builder balances. Must run after `processWithdrawals` and before `processFinalUpdates`. Per EIP-7732 §epoch-processing.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: add 3 pending payments, run epoch processing, verify all 3 deducted correctly. Edge case: payment amount > builder balance → capped at balance. `go test ./consensus/... -run TestBuilderPendingPayments`.
- **Definition of Done**: Epoch processing function runs. Payments deducted. Edge cases handled. Tests green.

#### Task SPEC-3.2 — Implement builder withdrawal with 64-epoch delay
- **Description**: In `pkg/epbs/builder_registry.go`, implement `RequestBuilderWithdrawal(builderIdx, amount)` that sets `builder.withdrawable_epoch = current_epoch + MIN_BUILDER_WITHDRAWABILITY_DELAY` (64 epochs). In `pkg/consensus/epoch_processing.go` withdrawal sweep, process up to `MAX_BUILDERS_PER_WITHDRAWALS_SWEEP = 16,384` builders per epoch, advancing `state.next_withdrawal_builder_index`. Enforce `BUILDER_WITHDRAWAL_PREFIX = 0x03` on execution addresses.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: request withdrawal at epoch 0 → not processed until epoch 64. Batch test: 20,000 builders requesting withdrawal → only 16,384 processed per epoch sweep. `go test ./epbs/... -run TestBuilderWithdrawal`.
- **Definition of Done**: 64-epoch delay enforced. Batch sweep limit correct. Prefix validated. Tests green.

#### Task SPEC-3.3 — ProposerPreferences P2P topic and self-build support
- **Description**: In `pkg/p2p/gossip_topics.go`, register gossip topic for `ProposerPreferences` messages using domain `DOMAIN_PROPOSER_PREFERENCES = 0x0D000000`. In `pkg/epbs/auction_engine.go`, when a bid has `builder_index = UINT64_MAX` (`BUILDER_INDEX_SELF_BUILD`), skip the auction and set the proposer as payload builder directly.
- **Estimated Effort**: 2 SP
- **Assignee**: P2P Engineer + Consensus Engineer
- **Testing Method**: P2P test: publish `ProposerPreferences` → topic subscriber receives it. Self-build test: proposer submits bid with `BUILDER_INDEX_SELF_BUILD` → auction skipped, proposer builds payload. `go test ./epbs/... -run TestBuilderSelfBuild`.
- **Definition of Done**: Topic registered. Self-build works. Tests green.

---

## US-SPEC-4: EIP-7805 FOCIL IL Equivocation Detection & Satisfaction Algorithm

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **validator**, I want IL equivocation detection (rejecting members who publish two conflicting ILs) and the correct O(n) IL satisfaction check (validating nonce + balance against post-execution state), so that the FOCIL protocol cannot be gamed by malicious IL committee members.

**Priority**: P1 | **Story Points**: 8 | **Sprint Target**: Sprint 2

### Acceptance Criteria
- A validator who sends 2 different ILs for the same slot is marked as equivocator; subsequent ILs from them are ignored
- The IL satisfaction algorithm correctly evaluates each tx in ILs against the post-execution state (nonce + balance)
- `engine_getInclusionListV1` Engine API endpoint is implemented and returns the current IL for the given slot
- `INCLUSION_LIST_UNSATISFIED` is returned by `engine_newPayload` when a tx in the IL is valid but absent from the block

### Tasks

#### Task SPEC-4.1 — Implement IL equivocation detection
- **Description**: In `pkg/focil/il_store.go`, maintain per-validator-per-slot IL store. On receiving a second `SignedInclusionList` from the same validator for the same slot: if the ILs differ, mark validator as equivocator via `il_store.MarkEquivocator(validatorIdx, slot)`. Subsequent ILs from the equivocator are silently dropped. Per EIP-7805 spec §equivocation.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: receive 2 identical ILs → no equivocation. Receive 2 different ILs → equivocator flagged, 3rd IL dropped. Verify count: `il_store.EquivocatorCount(slot) == 1`. `go test ./focil/... -run TestILEquivocationDetection`.
- **Definition of Done**: Equivocation detection correct. Equivocator's ILs dropped. Tests green. ≥ 80% coverage.

#### Task SPEC-4.2 — Implement EIP-7805 O(n) IL satisfaction algorithm
- **Description**: In `pkg/focil/il_validator.go`, implement `CheckILSatisfaction(block, ils, postState) bool` per EIP-7805 spec §satisfaction: for each tx T in ILs, if T is in block → skip. If gas remaining < T's gas limit → skip (insufficient gas is not a violation). Else validate T's nonce and balance against `postState` (state after all prior txs). If nonce/balance valid but T is absent → return `INCLUSION_LIST_UNSATISFIED`. Replace any ad-hoc current check with this canonical algorithm.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test cases: (1) all ILs txs in block → satisfied. (2) IL tx absent, gas available, nonce/balance valid → unsatisfied. (3) IL tx absent, insufficient gas → satisfied (gas exemption). (4) IL tx absent, invalid nonce → satisfied (state-invalid exemption). `go test ./focil/... -run TestILSatisfactionAlgorithm`.
- **Definition of Done**: Algorithm matches EIP-7805 spec text exactly. All 4 test cases pass.

#### Task SPEC-4.3 — Add `engine_getInclusionListV1` and `INCLUSION_LIST_UNSATISFIED` status
- **Description**: In `pkg/engine/`, implement `engine_getInclusionListV1(slot, committee_index) -> SignedInclusionList`. In `engine_newPayload` handler, call `CheckILSatisfaction()` and return `{status: "INCLUSION_LIST_UNSATISFIED"}` if the check fails. Add `INCLUSION_LIST_UNSATISFIED = "INCLUSION_LIST_UNSATISFIED"` constant per EIP-7805 spec §engine-api.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (engine)
- **Testing Method**: API test: call `engine_getInclusionListV1` with valid slot → returns IL. `engine_newPayload` with block missing a required IL tx → returns `INCLUSION_LIST_UNSATISFIED`. `go test ./engine/... -run TestEngineILSatisfied` and `TestEngineILUnsatisfied`.
- **Definition of Done**: Both endpoints implemented. Status constant defined. Tests green.

---

## US-SPEC-5: EIP-7928 BAL Ordering, Sizing Constraint & Retention Policy

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As an **EL client developer**, I want the BAL to enforce correct account ordering (lexicographic by address), `ITEM_COST=2000` sizing constraint, correct `BlockAccessIndex` assignment (0 for pre-tx system calls, 1..n for txs, n+1 for post-tx), early rejection of malicious oversized BALs, and a retention period of ≥ 3,533 epochs, so that ETH2030's BAL is fully EIP-7928-compliant and interoperable with other clients.

**Priority**: P1 | **Story Points**: 13 | **Sprint Target**: Sprint 2

### Acceptance Criteria
- Block validation rejects a BAL whose accounts are not in lexicographic order by address
- Block building rejects txs that would push `bal_items > block_gas_limit // ITEM_COST` (ITEM_COST=2000)
- Pre-execution system contract calls assigned `BlockAccessIndex=0`; txs `1..n`; post-execution `n+1`
- `G_remaining >= R_remaining * 2000` feasibility check runs every 8 txs during execution
- BAL storage layer retains BALs for at least 3,533 epochs before pruning
- `engine_getPayloadBodiesByHashV2` and `engine_getPayloadBodiesByRangeV2` return BAL data

### Tasks

#### Task SPEC-5.1 — BAL account ordering validation in block validation path
- **Description**: In `pkg/bal/validator.go` (create if missing), add `ValidateBALOrdering(bal BlockAccessList) error` that iterates all `AccountChanges` entries and verifies: (a) accounts in strict ascending lexicographic order by address, (b) storage_changes within each account in ascending order by key, (c) changes within each key in ascending order by `BlockAccessIndex`. Return error with first violation found.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (bal)
- **Testing Method**: Unit test: correctly ordered BAL → valid. Out-of-order address → error returned with offending address. Out-of-order storage key → error. `go test ./bal/... -run TestBALOrdering`.
- **Definition of Done**: Ordering validation integrated into block validation (`pkg/core/processor.go`). Tests green. EF state tests unaffected.

#### Task SPEC-5.2 — ITEM_COST=2000 BAL sizing constraint
- **Description**: In `pkg/bal/tracker.go`, add running counter `ItemCount`. After each transaction is tracked, check `ItemCount > block_gas_limit // ITEM_COST` (ITEM_COST=2000 per EIP-7928 §constants). If exceeded, return `ErrBALSizeExceeded` and exclude the offending tx from the block. In block building (`pkg/engine/`), enforce this constraint before including a tx.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (bal + engine)
- **Testing Method**: Unit test: block gas limit 30M → max 15,000 BAL items (30M/2000). Add 15,001st item → `ErrBALSizeExceeded`. `go test ./bal/... -run TestBALItemCostLimit`.
- **Definition of Done**: Constraint enforced. Block builder respects it. Tests green.

#### Task SPEC-5.3 — BlockAccessIndex 0 / 1..n / n+1 assignment
- **Description**: In `pkg/bal/tracker.go`, update `BlockAccessIndex` assignment: before any user tx executes, set index to 0 for system contract calls (EIP-6110 deposits, EIP-7002 withdrawals, EIP-7685 requests). For user txs, assign `1..n` in execution order. After all user txs, post-execution system calls get `n+1`. Wire index counter into `pkg/core/processor.go` `Process()` loop.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (bal + core)
- **Testing Method**: Integration test: block with 3 txs + pre/post system calls → verify BAL entries have correct indices: 0 for pre, 1/2/3 for txs, 4 for post. `go test ./bal/... -run TestBlockAccessIndexAssignment`.
- **Definition of Done**: Index assignment correct for all 3 categories. Tests green. Existing parallel execution tests unaffected.

#### Task SPEC-5.4 — Early rejection of malicious oversized BALs
- **Description**: In `pkg/core/processor.go`, implement the EIP-7928 §early-rejection feasibility check every 8 txs: `G_remaining >= R_remaining * 2000` where `R_remaining` is the number of undeclared storage reads not yet accessed and `G_remaining` is remaining block gas. If check fails, return `ErrBALFeasibilityViolated` and reject the block.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (core + bal)
- **Testing Method**: Test: construct adversarial block that declares 10,000 storage reads but has only 1M gas remaining → feasibility check fires and block rejected. Normal block → check passes every 8 txs without rejection. `go test ./core/... -run TestBALEarlyRejection`.
- **Definition of Done**: Feasibility check runs every 8 txs. Adversarial block rejected. Normal blocks unaffected.

#### Task SPEC-5.5 — BAL retention policy and `engine_getPayloadBodies` V2
- **Description**: In `pkg/core/rawdb/`, implement `RetainBALFor(epochs uint64)` that prevents BAL pruning before `3533` epochs (the weak subjectivity period). Add `engine_getPayloadBodiesByHashV2` and `engine_getPayloadBodiesByRangeV2` endpoints in `pkg/engine/` that return `ExecutionPayloadBodyV2` including the `blockAccessList` field.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (engine + rawdb)
- **Testing Method**: Retention test: store BAL, advance 3532 epochs → BAL still present. At 3533 epochs → eligible for pruning. Engine API test: `engine_getPayloadBodiesByHashV2` returns BAL in response body. `go test ./engine/... -run TestPayloadBodiesV2`.
- **Definition of Done**: Retention policy enforced. Both engine API methods return BAL. Tests green.

---

## US-SPEC-6: EIP-7706 Multidimensional Fee Vector Transaction Type

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As an **EVM developer**, I want a new EIP-7706 transaction type with 3-element fee vectors `[execution, blob, calldata]` for `max_fees_per_gas` and `priority_fees_per_gas`, a calldata gas calculation function, and updated block headers with 3D `gas_limits/gas_used/excess_gas` vectors, so that calldata is priced independently from execution gas (preventing calldata from crowding out computation within a block).

**Priority**: P1 | **Story Points**: 13 | **Sprint Target**: Sprint 3

### Acceptance Criteria
- New tx type (EIP-7706) accepted by txpool and included in blocks with `max_fees_per_gas` as 3-element vector
- `get_calldata_gas(calldata)` correctly computes: `tokens = zero_bytes + non_zero_bytes * 4; return tokens * 4` (CALLDATA_GAS_PER_TOKEN=4, TOKENS_PER_NONZERO_BYTE=4)
- Block header fields `gas_limits`, `gas_used`, `excess_gas` are 3-element vectors; `gas_limits[2] = gas_limits[0] // 4` (CALLDATA_GAS_LIMIT_RATIO=4)
- Per-dimension base fee updates via `fake_exponential(MIN_BASE_FEE=1, excess, target * 8)` (BASE_FEE_UPDATE_FRACTION=8)
- Calldata gas cap: tx rejected if calldata gas > `block_gas_limits[2]`

### Tasks

#### Task SPEC-6.1 — Implement EIP-7706 3D fee vector transaction type
- **Description**: In `pkg/core/types/`, add `MultiDimFeeTx` implementing `TypedTransaction` with type byte `EIP7706TxType`. Fields: `chain_id, nonce, gas_limit, to, value, data, access_list, blob_versioned_hashes, max_fees_per_gas [3]uint64, priority_fees_per_gas [3]uint64, y_parity, r, s`. Implement `RLP` encode/decode, `Cost()`, `EffectiveGasTip()` for 3D fees. Register in tx type switch.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (core/types)
- **Testing Method**: Round-trip RLP test. Txpool admission test: valid 3D fee tx → admitted; fee vector length ≠ 3 → rejected. `go test ./core/types/... -run TestMultiDimFeeTx`. EF state tests unaffected.
- **Definition of Done**: Tx type defined, encoded, decoded. Txpool admits it. Tests green. ≥ 80% coverage.

#### Task SPEC-6.2 — Implement `get_calldata_gas()` and calldata cap
- **Description**: In `pkg/core/gas_utils.go` (create if missing), implement `GetCalldataGas(calldata []byte) uint64`: count zero bytes, multiply non-zero by `TOKENS_PER_NONZERO_BYTE=4`, multiply total tokens by `CALLDATA_GAS_PER_TOKEN=4`. In tx admission (`pkg/txpool/txpool.go` and block building), enforce: `GetCalldataGas(tx.Data) <= block.GasLimits[2]`.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (core + txpool)
- **Testing Method**: Unit test: `GetCalldataGas([]byte{0x00, 0xff, 0x00})` = `(1 zero + 1 nonzero*4) * 4` + `(1 zero) * 4` ... let me compute: tokens = 2 zeros * 1 + 1 nonzero * 4 = 2+4=6; gas = 6*4=24. Test this. Block cap test: tx with calldata gas > limit → rejected. `go test ./core/... -run TestCalldataGas`.
- **Definition of Done**: `GetCalldataGas()` correct for all byte patterns (zero, nonzero, mixed). Cap enforced. Tests green.

#### Task SPEC-6.3 — Update block header with 3D gas vector fields
- **Description**: In `pkg/core/types/block.go`, add 3-element vector fields `GasLimits [3]uint64`, `GasUsed [3]uint64`, `ExcessGas [3]uint64` to `Header`. Implement `SetCallDataGasLimit()` to enforce `GasLimits[2] = GasLimits[0] / CALLDATA_GAS_LIMIT_RATIO` (ratio=4). Add fork check: before EIP-7706 fork, fields absent (use existing scalar `GasLimit/GasUsed`); after fork, vectors present.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (core/types)
- **Testing Method**: Unit test: header with `GasLimits[0]=30_000_000` → `GasLimits[2]=7_500_000`. Header RLP round-trip. Fork check: pre-fork header has no vector fields; post-fork has them. `go test ./core/types/... -run TestHeaderGasVectors`.
- **Definition of Done**: Header fields defined. Ratio constraint enforced. Fork gate correct. EF state tests unaffected.

#### Task SPEC-6.4 — 3D base fee update formula
- **Description**: In `pkg/core/multidim_gas.go`, extend the per-dimension EIP-1559 base fee update to include `DimCalldata` as the third dimension: `get_base_fee[i] = fake_exponential(MIN_BASE_FEE_PER_GAS=1, excess_gas[i], target_gas[i] * BASE_FEE_UPDATE_FRACTION=8)`. Wire calldata gas tracking into per-tx accounting: deduct `GetCalldataGas(tx.data)` from `GasUsed[2]` for each tx. Update `pkg/core/processor.go` to track all 3 dimensions.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (core)
- **Testing Method**: Unit test: block at target calldata usage → base fee unchanged. Over target → base fee increases. Under target → base fee decreases. `go test ./core/... -run TestCalldata3DBaseFee`.
- **Definition of Done**: `DimCalldata` tracked. Base fee updates correctly for all 3 dimensions. Tests green. No regression.

---

## US-SPEC-7: EIP-7864 Binary Trie Key Generation & Data Layout

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **state transition engineer**, I want the binary trie's key generation functions (`get_tree_key`, `get_tree_key_for_basic_data`, `get_tree_key_for_code_chunk`, `get_tree_key_for_storage_slot`) and the `BASIC_DATA_LEAF_KEY` 32-byte header packing verified against the EIP-7864 spec, so that all tooling (block explorers, light clients, provers) that reads the binary trie gets the correct layout.

**Priority**: P1 | **Story Points**: 8 | **Sprint Target**: Sprint 2

### Acceptance Criteria
- `get_tree_key(address32, tree_index, sub_index)` returns `blake3(address32 || tree_index_le32)[:31] || sub_index_byte`
- `BASIC_DATA_LEAF_KEY` 32-byte value packs: `version(1B) | reserved(4B) | code_size(3B) | nonce(8B) | balance(16B)` at exact byte offsets
- Code chunks use 31-byte chunks with leading PUSHDATA-count byte; chunk boundaries are tracked across PUSH data ranges
- `MAIN_STORAGE_OFFSET = 256^31` — main storage slots are keyed at tree_index ≥ `MAIN_STORAGE_OFFSET // 256`
- Empty leaf hash: if `value == [0x00]*64`, `hash = [0x00]*32`

### Tasks

#### Task SPEC-7.1 — Implement and test all `get_tree_key*` functions
- **Description**: In `pkg/trie/bintrie/keys.go` (create if missing), implement the four key generation functions from EIP-7864 spec §key-generation: `GetTreeKey(addr Address32, treeIndex int, subIndex int)`, `GetTreeKeyForBasicData(addr)`, `GetTreeKeyForCodeChunk(addr, chunkID)`, `GetTreeKeyForStorageSlot(addr, storageKey)`. Use BLAKE3 (from US-EL-1 / US-SPEC dependency on `lukechampine.com/blake3`). Inline storage: slots 0–63 at subindex 64–127; code at 128–255; main storage at `MAIN_STORAGE_OFFSET + slot`.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (trie)
- **Testing Method**: Table-driven tests against EIP-7864 spec examples. Key uniqueness test: 1000 random (address, slot) pairs → all distinct keys. `get_tree_key_for_storage_slot(addr, 0)` vs `get_tree_key_for_storage_slot(addr, 64)` → different subindices (64 vs 128). `go test ./trie/bintrie/... -run TestTreeKeyGeneration`.
- **Definition of Done**: All 4 key functions implemented. Spec example vectors pass. Key uniqueness verified. ≥ 80% coverage.

#### Task SPEC-7.2 — Verify BASIC_DATA_LEAF_KEY header packing
- **Description**: In `pkg/trie/bintrie/account.go` (or equivalent), implement `PackBasicDataLeaf(version uint8, codeSize uint32, nonce uint64, balance *big.Int) [32]byte` and `UnpackBasicDataLeaf([32]byte) (version, codeSize, nonce, balance)` following EIP-7864 spec §header-layout: offset 0: version (1B), offsets 1-4: reserved (4B, zero), offsets 5-7: code_size (3B big-endian), offsets 8-15: nonce (8B big-endian), offsets 16-31: balance (16B big-endian).
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (trie)
- **Testing Method**: Round-trip test: pack known (version=1, codeSize=100, nonce=5, balance=1 ETH) → unpack → identical values. Offset test: verify each field is at the exact byte offset specified. `go test ./trie/bintrie/... -run TestBasicDataLeafPacking`.
- **Definition of Done**: Pack/unpack correct. Byte offsets verified. Round-trip passes.

#### Task SPEC-7.3 — Code chunking: 31-byte chunks with PUSHDATA boundary tracking
- **Description**: In `pkg/trie/bintrie/code_chunker.go` (create if missing), implement `ChunkifyCode(code []byte) [][32]byte`: split code into 31-byte chunks, prepend each chunk with a 1-byte `leadingPUSHDATABytes` count. The leading byte counts how many bytes at the start of the chunk are PUSHDATA (not opcodes) — tracking PUSH1–PUSH32 instruction ranges across chunk boundaries. This follows EIP-7864 §code-chunking exactly.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (trie)
- **Testing Method**: Test 1: `PUSH1 0x60 ADD` → chunk 0 has `leadingPUSHDATABytes=0` (PUSH1 is an opcode). Test 2: code with `PUSH32` spanning a chunk boundary → next chunk has `leadingPUSHDATABytes=N`. Regression test: re-chunking the same code always produces identical output. `go test ./trie/bintrie/... -run TestCodeChunker`.
- **Definition of Done**: PUSH boundary tracking correct for PUSH1–PUSH32. Re-chunking is deterministic. `go test ./trie/bintrie/...` fully green.
