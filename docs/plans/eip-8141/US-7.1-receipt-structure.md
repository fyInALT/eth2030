# US-7.1 — Frame Transaction Receipt

**Epic:** EP-7 Receipt Structure
**Total Story Points:** 7
**Sprint:** 3

> **As a** block explorer developer,
> **I want** frame transaction receipts to include `payer`, `cumulative_gas_used`, and per-frame receipts,
> **so that** users and tools can trace which account paid fees and inspect per-frame outcomes.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Tasks

### Task 7.1.1 — FrameReceipt Struct and RLP Encoding

| Field | Detail |
|-------|--------|
| **Description** | Define `FrameReceipt` struct: `[status uint64, gas_used uint64, logs []*Log]`. Top-level receipt: `[cumulative_gas_used, payer Address, frame_receipts []FrameReceipt]`. Implement `EncodeRLP`/`DecodeRLP`. Integrate with EIP-2718 receipt dispatch for type `0x06`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Round-trip encode/decode with 3 frame receipts. (2) `payer` = address from APPROVE. (3) `status` = 0 or 1. (4) Logs assigned to correct frames. (5) `cumulative_gas_used` accumulates correctly. |
| **Definition of Done** | Round-trip tests pass; payer populated; logs per-frame; coverage ≥ 80%; reviewed. |

### Task 7.1.2 — Frame Status Tracking During Execution

| Field | Detail |
|-------|--------|
| **Description** | Track each frame's outcome: `status = 1` (success) or `0` (revert/exception). Record `gas_used` (consumed, not limit). Record `logs` per frame. Store in `FrameContext.Frames[i]` and populate into `FrameReceipt`. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Successful frame → status=1, gas_used < gas_limit. (2) Reverted frame → status=0. (3) Log in frame 1 → only in frame_receipts[1].logs. (4) TXPARAM `0x15` matches receipt status. |
| **Definition of Done** | Tests pass; status/gas/logs isolated per frame; reviewed. |

### Task 7.1.3 — JSON-RPC Receipt Serialization

| Field | Detail |
|-------|--------|
| **Description** | Extend `eth_getTransactionReceipt` for type-`0x06`: include `"payer"` and `"frameReceipts": [{"status", "gasUsed", "logs"}]`. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | RPC Engineer |
| **Testing Method** | Integration test: submit frame tx, fetch receipt via RPC, assert `payer` and `frameReceipts` present and correct. |
| **Definition of Done** | JSON includes new fields; values correct; existing tests unaffected; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/types/frame_receipt.go` | `FrameResult` (`Status`, `GasUsed`, `Logs`), `FrameTxReceipt` (`CumulativeGasUsed`, `Payer`, `FrameResults`) |
| `pkg/core/types/frame_receipt_extended.go` | Extended receipt functionality (RLP encode/decode may be here) |

## Implementation Status

**✅ Mostly Implemented**

- ✅ `FrameResult` struct: `{Status, GasUsed, Logs}`
- ✅ `FrameTxReceipt` struct: `{CumulativeGasUsed, Payer, FrameResults}`
- ✅ `TotalGasUsed()` and `AllLogs()` helpers
- ⚠️ **Gap:** RLP encode/decode for `FrameTxReceipt` not visible in base file (may be in extended)
- ⚠️ **Gap:** JSON-RPC serialization for `eth_getTransactionReceipt` not implemented

---

## EIP-8141 Reference Excerpts

### Specification → Receipt

> The `ReceiptPayload` is defined as:
>
> ```
> [cumulative_gas_used, payer, [frame_receipt, ...]]
> frame_receipt = [status, gas_used, logs]
> ```
>
> `payer` is the address of the account that paid the fees for the transaction. `status` is the return code of the top-level call.

### Rationale → Payer in receipt

> The payer cannot be determined statically from a frame transaction and is relevant to users. The only way to provide this information safely and efficiently over the JSON-RPC is to record this data in the receipt object.
