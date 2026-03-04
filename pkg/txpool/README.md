# txpool

Transaction pool for ETH2030, implementing standard Ethereum mempool semantics plus EIP-8141 frame transaction support.

## EIP-8141 Frame Transaction Support

### StateReader Interface (AA-3.1: StateDB Injection Pattern)

The txpool uses a minimal `StateReader` interface for account state lookups:

```go
type StateReader interface {
    GetNonce(addr types.Address) uint64
    GetBalance(addr types.Address) *big.Int
}
```

**Why a minimal interface?** Tight coupling between the txpool and a concrete StateDB implementation would create circular dependencies (txpool → core/state → txpool) and make the pool difficult to test in isolation. The interface is intentionally narrow: only the fields needed for admission checks are exposed.

### FrameStateReader (VERIFY Frame Code Check)

For VERIFY frame pre-flight simulation, a separate extended interface is defined:

```go
type FrameStateReader interface {
    StateReader
    GetCodeSize(addr types.Address) int
}
```

This is kept separate from `StateReader` so existing implementations (`mockState` in tests, adapters in node integration) do not need to implement `GetCodeSize` unless they opt into VERIFY simulation. Wire it via `pool.SetCodeReader(r)` at node startup.

**Why not extend StateReader directly?** Three existing implementations (`txpool_test.go`, `fuzz_test.go`, `e2e_test.go`) only implement `GetNonce` and `GetBalance`. Adding `GetCodeSize` to `StateReader` would break all three without providing value to non-frame workloads.

### Dual-Tier Frame Rules (AA-2.2)

Two rulesets are available, selected via `--frame-mempool`:

| Tier | Gas cap (VERIFY) | External calls in VERIFY |
|------|-----------------|--------------------------|
| `conservative` (default) | 50,000 | Not allowed |
| `aggressive` | 200,000 | Allowed if paymaster is staked |

Use `ValidateFrameTxConservative(tx)` or `ValidateFrameTxAggressive(tx, registry)` directly, or set `pool.config.FrameMempoolTier` (future: the node wires this based on the `--frame-mempool` flag).

### Paymaster Registry Check (AA-1.2)

When `Config.PaymasterRegistry` is non-nil and `Config.PaymasterStrict` is true, frame transactions whose VERIFY frame targets an external address (different from `tx.Sender`) must be staked in the registry. Unstaked paymasters trigger `ErrUnstakedPaymaster`.

Set `Config.PaymasterStrict = false` (or `--paymaster-registry=off`) to disable this check for testing.

### SimulateVerifyFrame (AA-3.2)

`SimulateVerifyFrame(tx, reader)` performs a lightweight structural pre-check:
1. At least one VERIFY frame must exist.
2. The VERIFY frame's target must have deployed code (not an EOA).

Full EVM simulation (confirming that APPROVE is called inside the VERIFY frame) requires a running EVM instance and is deferred to block processing in `pkg/core/processor.go`. The txpool check is intentionally cheap: a `GetCodeSize` call per transaction, adding < 1μs overhead.
