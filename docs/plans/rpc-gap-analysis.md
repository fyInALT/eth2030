# RPC Gap Analysis: ETH2030 vs Geth v1.17.2

Reference: `docs/plans/geth-rpc-reference.md` (real Geth responses captured 2026-03-12).

## Summary

| Category | Count |
|---|---|
| Bugs (wrong output) | 6 |
| Missing fields (Geth extensions) | 3 |
| Intentional differences (our EIP upgrades) | 3 |
| Methods not in Geth but in our impl | 4 |
| OK (matches Geth) | 45+ |

---

## Bugs — Wrong Output

### BUG-1: `eth_blobBaseFee` returns raw excessBlobGas, not blob fee

**File**: `pkg/rpc/ethapi/eth_api.go:1398`

```go
// WRONG: returns raw excess gas (could be 262144)
return successResponse(req.ID, encodeBigInt(new(big.Int).SetUint64(*header.ExcessBlobGas)))
```

**Expected** (Geth): `"0x1"` (1 wei minimum when excessBlobGas=0)

**Fix**: Use `gas.CalcBlobBaseFeeWithSchedule(excessBlobGas, schedule)` to compute actual blob base fee.

---

### BUG-2: `eth_feeHistory` missing `baseFeePerBlobGas` and `blobGasUsedRatio`

**File**: `pkg/rpc/ethapi/eth_api.go:428-433` (FeeHistoryResult struct)

```go
// CURRENT — missing blob fields
type FeeHistoryResult struct {
    OldestBlock   string     `json:"oldestBlock"`
    BaseFeePerGas []string   `json:"baseFeePerGas"`
    GasUsedRatio  []float64  `json:"gasUsedRatio"`
    Reward        [][]string `json:"reward,omitempty"`
}
```

**Expected** (Geth also returns):
- `baseFeePerBlobGas`: []string — N+1 entries (hex)
- `blobGasUsedRatio`: []float64 — N entries

**Fix**: Add fields to struct and populate from block headers (use `CalcBlobBaseFeeWithSchedule`).

---

### BUG-3: `FormatTransaction` — wrong `gasPrice` for EIP-1559 txs (type 2/3/4)

**File**: `pkg/rpc/types/types.go:596`

```go
// WRONG: for type-2, tx.GasPrice() returns GasFeeCap (maxFeePerGas = 20 Gwei)
GasPrice: EncodeBigInt(tx.GasPrice()),
```

**Expected** (Geth): `gasPrice` = `min(maxFeePerGas, baseFee + maxPriorityFeePerGas)` = effective gas price.

Example from Geth: `gasPrice=0x780fac05` (≈2.01 Gwei) with `baseFee=0xda1805` (14.3 Mwei) + `tip=0x77359400` (2 Gwei).

**Fix**: Pass `baseFee *big.Int` to `FormatTransaction`. For type ≥ 2, compute:
`effectiveGasPrice = min(tx.GasFeeCap(), baseFee + tx.GasTipCap())`

---

### BUG-4: `RPCTransaction.AccessList` omitempty suppresses empty access lists

**File**: `pkg/rpc/types/types.go:170`

```go
// WRONG: empty access list for type-2 tx is omitted by omitempty (Go treats len==0 slice as "empty")
AccessList []RPCAccessTuple `json:"accessList,omitempty"`
```

**Expected** (Geth): `"accessList": []` is always present for type ≥ 1 txs.

**Fix**: Change to `AccessList *[]RPCAccessTuple json:"accessList,omitempty"`. A nil pointer is omitted (for legacy txs), a non-nil pointer to empty slice is included as `[]`.

---

### BUG-5: `RPCTransaction` missing `yParity` field

**File**: `pkg/rpc/types/types.go:150-174`

Our `RPCTransaction` has no `yParity` field. Geth includes `yParity` for type ≥ 1 (EIP-2718 txs).

**Expected** (Geth): `"yParity": "0x1"` alongside `"v": "0x1"` for EIP-1559 txs.

**Fix**: Add `YParity *string json:"yParity,omitempty"` to `RPCTransaction`, set it to same value as `V` for type ≥ 1.

---

### BUG-6: `RPCTransaction` and `RPCLog` missing `blockTimestamp` field (Geth extension)

Geth adds `blockTimestamp` to both transaction objects and log objects:
- Tx: `"blockTimestamp": "0x69b22a25"`
- Log: `"blockTimestamp": "0x69b22a61"`

Our structs have no such field.

**Fix**:
- Add `BlockTimestamp *string json:"blockTimestamp,omitempty"` to `RPCTransaction`
- Add `BlockTimestamp *string json:"blockTimestamp,omitempty"` to `RPCLog`
- Add `BlockTimestamp uint64` to `types.Log` (matching Geth's Log struct, `rlp:"-"`)
- Populate in `FormatTransaction` (pass `blockTimestamp uint64`) and `FormatLog`
- In `getLogs`, copy log + set `BlockTimestamp = header.Time` before formatting

---

## Intentional Differences (Our EIP upgrades — OK to differ)

### DIFF-1: `eth_protocolVersion` returns `"0x44"` instead of `-32601`

Geth removed this method (returns -32601). We return `"0x44"` (ETH/68) as informational.
**Decision**: Keep our behavior. This is more useful and correct.

### DIFF-2: `eth_coinbase` returns fee recipient address instead of `-32601`

Geth removed this method. We return the coinbase from the current block header (fee recipient in PoS).
**Decision**: Keep our behavior. More useful for tooling.

### DIFF-3: `eth_mining` returns `false`, `eth_hashrate` returns `"0x0"` instead of `-32601`

Geth removed these methods. We return the correct PoS answer.
**Decision**: Keep. Returning false/0x0 is semantically correct and backward-compatible.

---

## Methods in Our Implementation Not in Geth Standard

These are our extensions (not in geth-rpc-reference):

- `eth_getHeaderByNumber` — our addition, useful for CL integration
- `eth_getHeaderByHash` — our addition
- `beacon_*` (10 methods) — our beacon API extensions (EIP-4788, CL queries)
- `debug_getBlockRlp`, `debug_printBlock`, `debug_chaindbProperty`, `debug_chaindbCompact`, `debug_setHead`, `debug_freeOSMemory` — useful debug utilities

---

## Methods Confirmed OK (matches Geth)

| Method | Status |
|---|---|
| `web3_clientVersion` | ✓ (our version string, correct format) |
| `web3_sha3` | ✓ |
| `net_version` | ✓ (decimal string, e.g. "3151908") |
| `net_listening` | ✓ |
| `net_peerCount` | ✓ |
| `eth_chainId` | ✓ |
| `eth_syncing` | ✓ (false when synced) |
| `eth_accounts` | ✓ (empty array []) |
| `eth_gasPrice` | ✓ |
| `eth_maxPriorityFeePerGas` | ✓ |
| `eth_blockNumber` | ✓ |
| `eth_getBalance` | ✓ |
| `eth_getStorageAt` | ✓ (32-byte padded) |
| `eth_getTransactionCount` | ✓ |
| `eth_getCode` | ✓ |
| `eth_getBlockByNumber` | ✓ (all post-Prague fields present) |
| `eth_getBlockByHash` | ✓ |
| `eth_getBlockTransactionCountByHash` | ✓ |
| `eth_getBlockTransactionCountByNumber` | ✓ |
| `eth_getUncleCountByBlockHash` | ✓ (0x0) |
| `eth_getUncleCountByBlockNumber` | ✓ (0x0) |
| `eth_getUncleByBlockHashAndIndex` | ✓ (null) |
| `eth_getUncleByBlockNumberAndIndex` | ✓ (null) |
| `eth_getTransactionByHash` | ~ (bugs 3,5,6 apply) |
| `eth_getTransactionByBlockHashAndIndex` | ~ (bugs 3,5,6 apply) |
| `eth_getTransactionByBlockNumberAndIndex` | ~ (bugs 3,5,6 apply) |
| `eth_getTransactionReceipt` | ✓ (from/to populated) |
| `eth_getBlockReceipts` | ✓ |
| `eth_call` | ✓ |
| `eth_estimateGas` | ✓ |
| `eth_sendRawTransaction` | ✓ |
| `eth_createAccessList` | ✓ |
| `eth_getProof` | ✓ (delegated to backend) |
| `eth_getLogs` | ~ (bug 6 applies: no blockTimestamp) |
| `eth_newFilter` | ✓ |
| `eth_newBlockFilter` | ✓ |
| `eth_newPendingTransactionFilter` | ✓ |
| `eth_getFilterChanges` | ✓ (all 3 filter types) |
| `eth_getFilterLogs` | ✓ |
| `eth_uninstallFilter` | ✓ |
| `eth_feeHistory` | ~ (bug 2: missing blob fields) |
| `eth_blobBaseFee` | ✗ (bug 1: wrong value) |
| `eth_subscribe` | ✓ (newHeads, logs, newPendingTransactions) |
| `eth_unsubscribe` | ✓ |
| `admin_nodeInfo` | ✓ |
| `admin_peers` | ✓ |
| `admin_addPeer` | ✓ |
| `admin_removePeer` | ✓ |
| `txpool_status` | ✓ |
| `txpool_content` | ✓ |
| `txpool_inspect` | ✓ |
| `debug_traceTransaction` | ✓ |
| `debug_traceCall` | ✓ |
| `debug_traceBlockByNumber` | ✓ |
| `debug_traceBlockByHash` | ✓ |

---

## Fix Plan

**Priority order** (most impactful first):

1. BUG-4: `AccessList` always present for type ≥ 1 (`*[]RPCAccessTuple`)
2. BUG-5: Add `yParity` to `RPCTransaction`
3. BUG-3: Effective `gasPrice` for EIP-1559 (pass `baseFee` to `FormatTransaction`)
4. BUG-1: `eth_blobBaseFee` correct value
5. BUG-2: `eth_feeHistory` blob fields
6. BUG-6: `blockTimestamp` on txs and logs

**Files to change**:
- `pkg/core/types/common.go` — add `BlockTimestamp` to Log struct
- `pkg/rpc/types/types.go` — RPCTransaction fields + FormatTransaction signature
- `pkg/rpc/ethapi/eth_api.go` — feeHistory struct, blobBaseFee fix, FormatTransaction call sites
