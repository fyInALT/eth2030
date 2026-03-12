# Engine Block Processor Refactor: Channel-Based Architecture

## 1. Problem Statement

The current `EngineBackend` processes blocks and builds payloads synchronously inside
HTTP handler goroutines, holding `b.mu.Lock()` for the entire duration of state
execution. This causes several problems:

### 1.1 Lock Contention During State Execution

**Current flow for `engine_newPayload`:**
```
HTTP goroutine
  → ProcessBlockV5()
      → b.mu.Lock()              ← held here
      → statedb.Dup()
      → processor.ProcessWithBAL(blk, stateCopy)  ← CAN TAKE 100ms+, LOCK HELD
      → b.blocks[hash] = blk
      → b.mu.Unlock()
```

While `ProcessWithBAL` runs, every other Engine API call that needs even a read
lock (`GetHeadTimestamp`, `GetBlockTimestamp`, `GetPayloadByID`) is blocked. Under
the CL's strict timing windows (12-second slots) this can cause missed deadlines.

### 1.2 Synchronous Payload Building in ForkchoiceUpdated

**Current flow for `engine_forkchoiceUpdated`:**
```
HTTP goroutine
  → ForkchoiceUpdated()
      → b.mu.Lock()
      → update headHash/safeHash/finalHash
      → if attrs != nil:
          → builder.BuildBlock(...)       ← synchronous
          → processor.ProcessWithBAL(...) ← synchronous, LOCK HELD
      → b.mu.Unlock()
```

The Engine API spec allows the CL to call `engine_getPayload` separately to pick
up the built payload. The `ForkchoiceUpdated` response need only carry the
`payloadID`; the actual payload can be built asynchronously.

### 1.3 Structural Coupling

All Engine API handler goroutines (one per HTTP request) directly call into
`EngineBackend` methods that hold the same write lock. There is no separation
between "request intake" and "block processing"; everything runs inline on the
HTTP goroutine.

---

## 2. Proposed Architecture

### 2.1 Core Idea

Introduce a single **processor goroutine** that owns all mutable backend state.
API handlers submit typed requests through an unbuffered channel and receive
results through a per-request reply channel. This is the standard Go actor pattern.

```
HTTP goroutine A (newPayload)      HTTP goroutine B (forkchoiceUpdated)
        │                                    │
        │ send blockReq{..., replyCh}        │ send fcuReq{..., replyCh}
        │                                    │
        ▼                                    ▼
    ┌───────────────────────────────────────────────┐
    │          b.reqCh  (chan processorReq)          │
    └─────────────────────┬─────────────────────────┘
                          │  read one at a time
                          ▼
              ┌──────────────────────┐
              │   processor loop     │   ← single goroutine, owns all mutable state
              │   (processLoop)      │     blocks, statedb, headHash, payloads, ils
              └──────────┬───────────┘
                         │ brief b.mu.Lock() only during final state commit
                         ▼
                   send result to req.replyCh

HTTP goroutine A ← receives PayloadStatusV1
HTTP goroutine B ← receives ForkchoiceUpdatedResult
```

### 2.2 Separation of Concerns

| Operation | Who runs it | When |
|-----------|-------------|------|
| Decode payload / validate block hash | HTTP goroutine | before enqueue |
| State execution (ProcessWithBAL) | processor goroutine | owns statedb |
| Write to b.blocks / b.statedb | processor goroutine | under brief b.mu.Lock() |
| Read b.payloads / b.blocks | any goroutine | under b.mu.RLock() |
| Payload build (async, post-FCU) | processor goroutine | after FCU reply |

**Key**: State execution (`ProcessWithBAL`) is **no longer under `b.mu`**.
The write lock is only held for a short final commit (map insert + pointer swap),
not for the entire execution duration.

---

## 3. Implementation Plan

### Step 1 — Define request/response types (`engine/processor_types.go`)

New file introducing typed messages for the processor channel:

```go
// processorReqType identifies the kind of request.
type processorReqType int

const (
    reqProcessBlock   processorReqType = iota // newPayload V1/V3
    reqProcessBlockV4                         // newPayload V4 (Prague)
    reqProcessBlockV5                         // newPayload V5 (Amsterdam)
    reqForkchoiceUpdated                      // FCU V1/V3
    reqForkchoiceUpdatedV4                    // FCU V4 (Amsterdam)
)

// processorReq is a single work item enqueued to the processor loop.
type processorReq struct {
    kind processorReqType

    // newPayload fields
    payloadV3               *ExecutionPayloadV3
    payloadV5               *ExecutionPayloadV5
    blobVersionedHashes     []types.Hash
    parentBeaconBlockRoot   types.Hash
    executionRequests       [][]byte

    // forkchoice fields
    fcState   ForkchoiceStateV1
    attrsV3   *PayloadAttributesV3
    attrsV4   *PayloadAttributesV4

    // reply channel — processor writes the result here; caller blocks on it.
    replyCh chan processorResp
}

// processorResp carries the result of a processed request.
type processorResp struct {
    payloadStatus PayloadStatusV1
    fcuResult     ForkchoiceUpdatedResult
    err           error
}
```

### Step 2 — Add processor loop to `EngineBackend` (`engine/backend.go`)

Add fields to `EngineBackend`:
```go
type EngineBackend struct {
    mu      sync.RWMutex   // protects reads (payloads, blocks, hashes)
    reqCh   chan processorReq  // NEW: unbuffered channel to processor
    stopCh  chan struct{}      // NEW: stops the goroutine on Close()
    // ... existing fields unchanged ...
}
```

Add `Start()` and `Close()` methods:
```go
// Start launches the processor goroutine. Must be called once before use.
func (b *EngineBackend) Start() {
    go b.processLoop()
}

// Close shuts down the processor goroutine.
func (b *EngineBackend) Close() {
    close(b.stopCh)
}
```

Add the processor loop — the goroutine that exclusively runs state transitions:
```go
func (b *EngineBackend) processLoop() {
    for {
        select {
        case <-b.stopCh:
            return
        case req := <-b.reqCh:
            var resp processorResp
            switch req.kind {
            case reqProcessBlock:
                resp.payloadStatus, resp.err = b.execProcessBlock(req.payloadV3, req.blobVersionedHashes, req.parentBeaconBlockRoot)
            case reqProcessBlockV4:
                resp.payloadStatus, resp.err = b.execProcessBlockV4(req)
            case reqProcessBlockV5:
                resp.payloadStatus, resp.err = b.execProcessBlockV5(req)
            case reqForkchoiceUpdated:
                resp.fcuResult, resp.err = b.execForkchoiceUpdated(req.fcState, req.attrsV3)
            case reqForkchoiceUpdatedV4:
                resp.fcuResult, resp.err = b.execForkchoiceUpdatedV4(req.fcState, req.attrsV4)
            }
            req.replyCh <- resp
        }
    }
}
```

### Step 3 — Replace public methods with channel dispatch

Replace each blocking public method (currently holding `b.mu.Lock()` for execution)
with a channel submission wrapper:

```go
// ProcessBlockV5 submits the payload to the processor goroutine and waits.
func (b *EngineBackend) ProcessBlockV5(
    payload *ExecutionPayloadV5,
    expectedBlobVersionedHashes []types.Hash,
    parentBeaconBlockRoot types.Hash,
    executionRequests [][]byte,
) (PayloadStatusV1, error) {
    // Decode and hash-check before enqueueing (cheap, safe on any goroutine).
    blk, err := payloadToBlock(&payload.ExecutionPayloadV3)
    if err != nil { ... return StatusInvalid }
    if blockHashMismatch(blk, payload.BlockHash) { ... return StatusInvalidBlockHash }

    replyCh := make(chan processorResp, 1)
    b.reqCh <- processorReq{
        kind:                  reqProcessBlockV5,
        payloadV5:             payload,
        blobVersionedHashes:   expectedBlobVersionedHashes,
        parentBeaconBlockRoot: parentBeaconBlockRoot,
        executionRequests:     executionRequests,
        replyCh:               replyCh,
    }
    resp := <-replyCh
    return resp.payloadStatus, resp.err
}
```

The actual execution moves to `execProcessBlockV5` (private, only called by the loop):

```go
// execProcessBlockV5 runs inside the processor goroutine — no external lock needed.
func (b *EngineBackend) execProcessBlockV5(req processorReq) (PayloadStatusV1, error) {
    payload := req.payloadV5
    blk, _ := payloadToBlock(&payload.ExecutionPayloadV3)   // already validated by caller

    // Parent check — goroutine owns b.blocks, no lock needed.
    parentHash := blk.ParentHash()
    if _, ok := b.blocks[parentHash]; !ok {
        return PayloadStatusV1{Status: StatusSyncing}, nil
    }

    // Restore calldata fields.
    if b.config.IsGlamsterdan(blk.Header().Time) {
        blk = restoreCalldataGasFields(blk, b.blocks[parentHash], payload.BlockHash)
    }

    // State execution — NOT under any lock.
    stateCopy := b.statedb.Dup()
    result, err := b.processor.ProcessWithBAL(blk, stateCopy)
    if err != nil { ... return StatusInvalid }

    // BAL and IL checks (unchanged logic) ...

    // Commit: brief lock only for map/pointer update.
    blockHash := blk.Hash()
    b.mu.Lock()
    b.blocks[blockHash] = blk
    if result.BlockAccessList != nil {
        b.bals[blockHash] = result.BlockAccessList
    }
    b.evictOldBlocks()
    b.statedb = stateCopy
    b.ils = b.ils[:0]
    b.mu.Unlock()

    return PayloadStatusV1{Status: StatusValid, LatestValidHash: &blockHash}, nil
}
```

### Step 4 — Async payload building in ForkchoiceUpdated

`ForkchoiceUpdated` with payload attributes:
1. Update `headHash/safeHash/finalHash` (under write lock, fast).
2. Generate a `payloadID` and return immediately with it.
3. Enqueue payload building to happen **after** the reply is sent back.

```go
func (b *EngineBackend) execForkchoiceUpdated(
    fcState ForkchoiceStateV1,
    attrs *PayloadAttributesV3,
) (ForkchoiceUpdatedResult, error) {
    if fcState.HeadBlockHash != (types.Hash{}) {
        if _, ok := b.blocks[fcState.HeadBlockHash]; !ok {
            return ForkchoiceUpdatedResult{PayloadStatus: PayloadStatusV1{Status: StatusSyncing}}, nil
        }
    }

    b.mu.Lock()
    b.headHash = fcState.HeadBlockHash
    b.safeHash = fcState.SafeBlockHash
    b.finalHash = fcState.FinalizedBlockHash
    b.mu.Unlock()

    result := ForkchoiceUpdatedResult{
        PayloadStatus: PayloadStatusV1{
            Status:          StatusValid,
            LatestValidHash: &fcState.HeadBlockHash,
        },
    }

    if attrs != nil {
        // Validate basic preconditions (fast, no state access).
        id, err := b.validateAndReservePayload(fcState, attrs.Timestamp)
        if err != nil {
            return ForkchoiceUpdatedResult{}, err
        }
        result.PayloadID = &id

        // Build payload asynchronously AFTER returning the ID.
        // The processor loop will pick this up on the next iteration.
        go b.buildPayloadAsync(id, fcState.HeadBlockHash, attrs)
    }

    return result, nil
}
```

**Note on async build**: the spec says `getPayload` may be called up to ~4 seconds
after `forkchoiceUpdated`. `buildPayloadAsync` runs as a goroutine that writes to
`b.payloads` under a brief write lock when complete. `getPayload` callers use
`RLock` and will either find the payload ready or receive `ErrUnknownPayload` (which
the CL should retry).

Alternatively, for simplicity in this first step, keep payload building synchronous
(inside the processor goroutine) but just decouple it from the `b.mu` held during
state execution. Mark async build as a follow-up.

### Step 5 — Read-only methods unchanged (keep RLock)

These methods do not route through the channel; they use `b.mu.RLock()` directly:

- `GetPayloadByID`, `GetPayloadV4ByID`, `GetPayloadV6ByID`
- `GetHeadTimestamp`, `GetBlockTimestamp`
- `GetHeadHash`, `GetSafeHash`, `GetFinalizedHash`
- `IsCancun`, `IsPrague`, `IsAmsterdam`
- `ProcessInclusionList` (fast, just appends — can stay with mutex)
- `GetInclusionList` (returns empty stub)

Since the processor goroutine only holds `b.mu.Lock()` during the brief final
commit (not during execution), read operations are no longer blocked for 100ms+.

### Step 6 — Wire Start/Close in node.go

In `pkg/node/node.go`, call `backend.Start()` after construction and `backend.Close()`
during node shutdown.

---

## 4. Files Changed

| File | Change |
|------|--------|
| `pkg/engine/processor_types.go` | **NEW**: request/response types for the channel |
| `pkg/engine/backend.go` | Add `reqCh`, `stopCh` fields; replace `ProcessBlock*` / `ForkchoiceUpdated*` with channel wrappers; add `processLoop`, `execProcess*`, `execForkchoice*`, `Start`, `Close` |
| `pkg/node/node.go` | Call `backend.Start()` / `backend.Close()` |
| `pkg/engine/backend_test.go` | Update tests to call `Start()` before use; add concurrency tests |

No changes to `backendapi/interfaces.go` (public API unchanged).
No changes to `handler.go`, `server.go`, or any API sub-packages.

---

## 5. Invariants Preserved

- **Serialized writes**: Only the processor goroutine writes to `b.blocks`,
  `b.statedb`, `b.payloads`, `b.headHash/safeHash/finalHash`. No concurrent writers.
- **Safe reads**: Read-only methods continue to use `b.mu.RLock()`. The write
  lock is now held only for O(1) map/pointer updates, not for O(N) state execution.
- **Sequential block processing**: `reqCh` is unbuffered; only one block processes
  at a time (consistent with CL's sequential `newPayload` calls per slot).
- **Backpressure**: If the processor is busy, the HTTP handler goroutine blocks
  on the channel send. This is correct — the CL should not send a second
  `newPayload` before receiving the response to the first.
- **No goroutine leak**: `stopCh` is closed on `Close()`, which terminates the loop.

---

## 6. Testing Strategy

1. **Unit tests** (`backend_test.go`): existing tests call `backend.Start()` and
   verify correct results — no behavioral change expected.
2. **Concurrency test**: launch 10 goroutines calling `GetHeadTimestamp` while
   `ProcessBlock` is running; verify no deadlock and correct results.
3. **Race detector**: run `go test -race ./engine/...` — must pass clean.
4. **Devnet**: boot Kurtosis devnet, verify blocks advance past block 10.

---

## 7. Follow-Up Work (not in this PR)

- **Async payload building**: `ForkchoiceUpdated` returns immediately; payload
  built in background goroutine (go-ethereum approach).
- **Txpool integration**: fill built blocks with pending transactions.
- **Timeout handling**: add context deadline to channel sends for safety.
