# RLP Encode/Decode Audit

This document catalogs all types that use RLP encoding/decoding in eth2030,
compares each with go-ethereum's equivalent, and records any differences found.

## Types Inventory

| Type | File | Geth Equivalent | Status |
|------|------|-----------------|--------|
| `Header` | `core/types/header_rlp.go` | `core/types/gen_header_rlp.go` | BUG (nil gap) |
| `Block` | `core/types/block_rlp.go` | `core/types/block.go` | BUG (withdrawals + tx wrapping) |
| `Transaction` (legacy) | `core/types/transaction_rlp.go` | `core/types/transaction.go` | OK |
| `Transaction` (AccessList) | `core/types/transaction_rlp.go` | `core/types/tx_access_list.go` | OK |
| `Transaction` (DynamicFee) | `core/types/transaction_rlp.go` | `core/types/tx_dynamic_fee.go` | OK |
| `Transaction` (Blob) | `core/types/transaction_rlp.go` | `core/types/tx_blob.go` | OK |
| `Transaction` (SetCode) | `core/types/transaction_rlp.go` | `core/types/tx_setcode.go` | OK |
| `Transaction` (Frame/PQ/LocalTx/MultiDim) | `core/types/transaction_rlp.go` | N/A (eth2030-specific) | N/A |
| `Receipt` | `core/types/receipt_rlp.go` | `core/types/receipt.go` | OK (Byzantium+) |
| `Log` | `core/types/receipt_rlp.go` | `core/types/gen_log_rlp.go` | OK |
| `Withdrawal` | `core/types/withdrawal.go` | `core/types/gen_withdrawal_rlp.go` | OK |

---

## Detailed Findings

### 1. Block — Missing Withdrawals (CRITICAL BUG)

**File:** `pkg/core/types/block_rlp.go`

**Geth encoding** (`extblock` struct with `rlp:"optional"`):
```
Block = [Header, [tx...], [uncle...], [withdrawal...]]  // withdrawals optional
```

**Eth2030 encoding** (manual):
```
Block = [Header, [tx...], [uncle...]]  // withdrawals MISSING
```

Geth's `Block.EncodeRLP` uses an `extblock` struct:
```go
type extblock struct {
    Header      *Header
    Txs         []*Transaction
    Uncles      []*Header
    Withdrawals []*Withdrawal `rlp:"optional"`
}
```

Eth2030's `Block.EncodeRLP` only encodes `[header, txs, uncles]`. Withdrawals are
stored in `b.body.Withdrawals` but never written into the RLP output.

**Impact:** Any post-Shanghai block (with withdrawals) encoded by eth2030 is
missing the withdrawals list. This breaks P2P block propagation, DB storage/retrieval,
and cross-node compatibility.

**Fix:** Add withdrawals to both `EncodeRLP` and `DecodeBlockRLP`.

---

### 2. Block — Legacy Transaction Double-Wrapping (CRITICAL BUG)

**File:** `pkg/core/types/block_rlp.go` (lines 17–29)

**Geth behavior:**
- Legacy txs in block body → RLP list (direct, not byte-string-wrapped)
- Typed txs in block body → RLP byte string containing `type_byte || RLP_list`

`Transaction.EncodeRLP(w io.Writer)` in geth:
```go
if tx.Type() == LegacyTxType {
    return rlp.Encode(w, tx.inner)  // produces RLP list
}
// typed:
buf.Reset(); tx.encodeTyped(buf)
return rlp.Encode(w, buf.Bytes())  // produces RLP byte string
```

**Eth2030 behavior:**
```go
txEnc, _ := tx.EncodeRLP()                    // returns raw []byte
wrapped, _ := rlp.EncodeToBytes(txEnc)        // ALWAYS wraps as byte string
```

`rlp.EncodeToBytes([]byte)` encodes the byte slice as an RLP byte string.
So for a legacy tx whose `EncodeRLP()` already returns an RLP list (`0xc8...`),
`rlp.EncodeToBytes` double-wraps it: the block body contains
`[byte_string_prefix][list_prefix][nonce][gas]...` instead of `[list_prefix][nonce][gas]...`.

**Eth2030 decoder** uses `s.Bytes()` which reads byte strings. This means:
- Eth2030 → eth2030 round-trip works (both sides use byte strings for all tx types)
- Eth2030 cannot decode geth-encoded blocks (legacy txs are lists, not byte strings)
- Geth cannot decode eth2030-encoded blocks (legacy txs appear as byte strings, not lists)

**Impact:** Full interop failure for blocks containing legacy transactions.

**Fix:** In `EncodeRLP`, special-case legacy vs typed txs:
```go
if tx.Type() == LegacyTxType {
    txPayload = txEnc  // append directly (it is already an RLP list)
} else {
    txPayload = rlp.EncodeToBytes(txEnc)  // wrap typed tx as byte string
}
```
Correspondingly fix `DecodeBlockRLP` to use `s.RawItem()` for legacy (list) and
`s.Bytes()` for typed (byte string), or use a `Kind()` check.

---

### 3. Header — Optional Field Nil-Gap Encoding (MODERATE BUG)

**File:** `pkg/core/types/header_rlp.go`

**Geth behavior** (from `gen_header_rlp.go`):
When any later optional field is non-nil, all preceding nil optional fields
are encoded as `0x80` (empty string) to preserve positional ordering:
```go
_tmp1 := obj.BaseFee != nil
_tmp2 := obj.WithdrawalsHash != nil
...
if _tmp1 || _tmp2 || _tmp3 || _tmp4 || _tmp5 || _tmp6 {
    if obj.BaseFee == nil {
        w.Write(rlp.EmptyString)  // 0x80 placeholder
    } else {
        w.WriteBigInt(obj.BaseFee)
    }
}
```

**Eth2030 behavior:**
```go
if h.BaseFee != nil {
    items = append(items, h.BaseFee)
}
if h.WithdrawalsHash != nil {
    items = append(items, *h.WithdrawalsHash)
}
```

If BaseFee is nil but a later field (e.g., CalldataGasUsed) is non-nil,
eth2030 simply omits BaseFee and the decoder will misinterpret the subsequent
field as BaseFee.

**Impact:** In normal Ethereum fork progression, each fork adds all fields
cumulatively, so BaseFee is always present if WithdrawalsHash is present.
However, for eth2030-specific extra fields (CalldataGasUsed after RequestsHash),
this is a real risk: if RequestsHash is nil but CalldataGasUsed is set,
the decoder reads CalldataGasUsed in the RequestsHash slot.

**Fix:** Mirror geth's approach: when encoding, if any field at position N is
non-nil, write `0x80` for all nil fields at positions < N.

---

### 4. Receipt — PostState Field (LOW, pre-Byzantium only)

**File:** `pkg/core/types/receipt_rlp.go`

**Geth encoding** uses `PostStateOrStatus []byte`:
- Pre-Byzantium: `PostState` (32-byte state root hash)
- Post-Byzantium: `[]byte{}` (failed) or `[]byte{0x01}` (succeeded)

**Eth2030 encoding** uses `r.Status uint64`:
- Status 0: encoded as `0x80` (same as geth `[]byte{}` → `0x80`)
- Status 1: encoded as `0x01` (same as geth `[]byte{0x01}` → `0x01`)

**Verdict:** For Byzantium+ receipts the encoding is byte-for-byte identical.
Only pre-Byzantium receipts (with 32-byte PostState) differ. Since eth2030
targets modern forks, this is low priority.

---

### 5. eth/codec.go — Transaction double-wrapping (CRITICAL BUG)

**File:** `pkg/eth/codec.go` (lines 17–27, 44–49)

Same double-wrapping bug as Block (finding #2), in the P2P tx encoding path:
- `encodeTransactions`: always wraps all txs with `rlp.EncodeToBytes(txEnc)` — legacy txs become byte strings instead of lists.
- `decodeTransactions`: uses `s.Bytes()` for all txs — can only decode byte strings, fails on legacy list txs.

**Fix:** Same pattern as block_rlp.go: check `tx.Type() == LegacyTxType` before wrapping; use `s.Kind()` in decoder.

---

### 6. eth/messages.go — Transactions/Headers use reflection (CRITICAL BUG)

**File:** `pkg/eth/messages.go`

`EncodeMsg`/`DecodeMsg` used `rlp.EncodeToBytes(tm.Transactions)` and `rlp.EncodeToBytes(bm.Headers)`.
eth2030's rlp encoder uses pure reflection and only encodes exported fields. Both `Transaction`
and `Header` have no exported fields (all state is in unexported `inner`/`hash`/etc.) so:
- Each transaction encodes as `0xc0` (empty struct).
- Each header encodes as `0xc0` (empty struct).

Same broken pattern for `MsgPooledTransactions` and `MsgBlockBodies`.

**Fix:** Use proper `tx.EncodeRLP()` / `h.EncodeRLP()` with same legacy/typed dispatch.
Added `encodeTxsToRLP`, `decodeTxsFromRLP`, `encodeHeadersToRLP`, `decodeHeadersFromRLP`,
`encodeBodyListToRLP`, `decodeBodyListFromRLP` helpers.

---

### 7. eth/messages.go — BlockBodyData missing Withdrawals (MODERATE BUG)

**File:** `pkg/eth/messages.go` (BlockBodyData struct)

```go
type BlockBodyData struct {
    Transactions []*types.Transaction
    Uncles       []*types.Header
    // Withdrawals missing!
}
```

Post-Shanghai block bodies include `[txs, uncles, withdrawals]`. Without the Withdrawals field,
any post-Shanghai block body received or sent over P2P truncates the withdrawals list.

**Fix:** Added `Withdrawals []*types.Withdrawal` field; `encodeBodyListToRLP` / `decodeBodyListFromRLP`
handle optional withdrawals encoding matching geth's `extblock` pattern.

---

### 8. core/rawdb/chaindb.go — tx double-wrapping + missing withdrawals (CRITICAL)

**File:** `pkg/core/rawdb/chaindb.go` (`encodeBlockBody` / `decodeBlockBody`)

Same double-wrapping bug: `rlp.EncodeToBytes(txEnc)` for all txs regardless of type, and
`s.Bytes()` in decoder for all txs. Additionally, `encodeBlockBody` never wrote the
withdrawals list, and `decodeBlockBody` didn't check for it — post-Shanghai blocks stored
in the chain DB would silently lose their withdrawals.

**Fix:** Same tx type dispatch as block_rlp.go; added withdrawals encoding/decoding matching
the block_rlp.go pattern.

---

### 9. core/chain/blockchain.go — tx double-wrapping (CRITICAL)

**File:** `pkg/core/chain/blockchain.go` (`encodeBlockBody` / `decodeBlockBody`)

Another copy of the same double-wrapping bug. The decoder (`s.Bytes()`) was the mirror of
the broken encoder so the local roundtrip worked, but blocks stored here wouldn't be
P2P-compatible and couldn't be decoded by the now-fixed `block_rlp.go` decoder.

**Fix:** Same tx type dispatch.

---

## Summary

| # | Severity | Bug | Fixed |
|---|----------|-----|-------|
| 1 | CRITICAL | Block missing Withdrawals in RLP | Yes |
| 2 | CRITICAL | Block legacy tx double-wrapping | Yes |
| 3 | MODERATE | Header nil-gap in optional fields | Yes |
| 4 | LOW | Receipt no pre-Byzantium PostState support | N/A (modern forks) |
| 5 | CRITICAL | eth/codec.go tx double-wrapping + wrong decoder | Yes |
| 6 | CRITICAL | eth/messages.go tx/header reflection encoding (empty) | Yes |
| 7 | MODERATE | eth/messages.go BlockBodyData missing Withdrawals | Yes |
| 8 | CRITICAL | core/rawdb/chaindb.go tx double-wrapping + no withdrawals | Yes |
| 9 | CRITICAL | core/chain/blockchain.go tx double-wrapping | Yes |
