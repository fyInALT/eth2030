# Project Status

## Engine Lock Wrapper Refactor (2026-03-19)

### Current Status: COMPLETE

**Scope:** Refactor `pkg/engine/backend.go` so accesses guarded by `stateMu`, `blocksMu`, `payloadMu`, and `ilMu` go through helper functions instead of open-coded lock/unlock pairs.

**Completed Work:**
- Add backend helper methods that wrap read/write access for each fine-grained mutex.
- Migrate existing call sites in `pkg/engine/backend.go` to the new helpers without changing lock ordering or behavior.
- Verify the `pkg/engine` build and tests after the refactor.

**Verification:**
- `cd pkg && go test ./engine/...`

## Go Test Repair (2026-03-19)

### Current Status: COMPLETE

**Scope:** Reproduce and fix failing or hanging `go test ./...` runs from the Go module root at `pkg/`.

**Resolved Issues:**
- `pkg/core/vm` JSON-vector and precompile fixture tests no longer fail when the optional external fixture corpus is absent from the checkout; they now skip cleanly on missing files.
- `pkg/engine` no longer deadlocks while storing processed blocks: `ProcessBlock` and `ProcessBlockV5` were calling `evictOldBlocks` with `blocksMu` already held, while the helper tried to lock `blocksMu` again.

**Verification:**
- `cd pkg && go test ./...`

## Devnet Testing (2026-03-16)

### Current Status: WORKING

**Devnet Configuration:** `full-feature-prysm.yaml` with PeerDAS/Fulu

**Test Results:**
- Slot: 367+ (continuing to advance)
- Finalized Epoch: 9+ (continuing to finalize)
- Block Processing Time: 15-200ms (normal)
- Data Availability Check: ~5 microseconds (instant)

### Issues Fixed

1. **engine_getPayloadV5 "build timed out" (CRITICAL)**
   - Root Cause: `buildMu` mutex held during `StateAtBlock` call which could take tens of seconds on cache miss
   - Death Spiral: state cache miss → state rebuild blocks buildMu → all builds timeout → chain stalls
   - Fix: Pre-fetch state BEFORE acquiring `buildMu`, use singleflight for deduplication
   - File: `pkg/node/backend.go`
   - Status: ✅ RESOLVED

2. **Race condition in computeBlobsV2**
   - Root Cause: `len(b.blobCache)` accessed after `blobCacheMu.RUnlock()`
   - Fix: Capture cache size before releasing lock
   - Status: ✅ RESOLVED

### Known Non-Critical Issues

1. **engine_getBlobsV2 "JSON-RPC response has no result"**
   - This is expected behavior in PeerDAS
   - Blobs are propagated via gossip (data column sidecars), not through Engine API
   - Prysm CL attempts to reconstruct data columns from EL as a fallback
   - EL doesn't have blobs for blocks built by other nodes
   - **Impact: NONE** - Chain continues to advance and finalize normally
   - These errors are just log noise from Prysm's gossip message handling

### Architecture Notes

**PeerDAS Data Flow:**
1. Block proposer (CL-2) broadcasts data column sidecars (128 columns per block)
2. Other CL nodes receive data columns via gossip
3. Data availability is verified from gossip-received columns
4. CL may attempt to reconstruct from EL as fallback (fails but non-fatal)
5. Block processing succeeds regardless

**State Cache Behavior:**
- `StateAtBlock` uses singleflight for deduplication
- Hot cache (`sc`) and permanent cache (`memSC`) provide state caching
- Re-execution needed when cache misses (can take tens of seconds)

---

## Formal Verification

- Formal Lean workspace is present under `formal/lean`.
- Initial EVM/VM model (`Lean2030/VM`) and richer `Lean2030/EVM` model are now linked in `Lean2030/Lean2030.lean`.
- EVM model now includes:
  - arithmetic and bitwise ops,
  - DUP/SWAP,
  - POP/JUMP/JUMPI/JUMPDEST,
  - compiler and theorem suites.
- Current status: still a semantic subset only (toy control-flow, no memory/tracing, no Go cross-checking).
- Added bytecode-offset execution module `formal/lean/Lean2030/EVM/Bytecode.lean` with `decodeAt`/`decodePush`/`runBytecode`.
- Added executable mismatch lemmas showing current op-index `run` diverges from byte-offset EVM behavior for `PUSH`-preceded jumps.
- Lean 4 test driver is now configured in `formal/lean/lakefile.lean` with `Lean2030/Tests.lean`.
- Added `.gitignore` entries for Lean outputs (`.lake`, `*.olean`, `*.ilean`) and linked the Lean VM strategy doc in README.
- Next major deliverable: full `compile`→`runBytecode` refinement theorem and stronger instruction/stack invariants.
