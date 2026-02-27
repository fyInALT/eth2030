# US-1.1 — Frame Transaction RLP Serialization

**Epic:** EP-1 Transaction Type & RLP Encoding
**Total Story Points:** 7
**Sprint:** 1 (Foundations)

> **As a** protocol engineer,
> **I want** the `FrameTx` struct to serialize to and deserialize from canonical RLP as defined in EIP-8141,
> **so that** frame transactions can be included in blocks, signed, and transmitted over the wire.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Tasks

### Task 1.1.1 — Finalize `FrameTx` RLP Encode/Decode

| Field | Detail |
|-------|--------|
| **Description** | Implement `EncodeRLP` and `DecodeRLP` for `FrameTx` following the canonical layout: `[chain_id, nonce, sender, frames, max_priority_fee_per_gas, max_fee_per_gas, max_fee_per_blob_gas, blob_versioned_hashes]`. Each frame encodes as `[mode, target, gas_limit, data]`. Null target must encode as empty bytes (`0x80`) and decode back to `nil`. Ensure `blob_versioned_hashes` is an empty list (not nil RLP) and `max_fee_per_blob_gas` is `0` when no blobs are present. Integrate with the existing `FrameTx` in `pkg/core/types/tx_frame.go`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Unit tests in `pkg/core/types/tx_frame_test.go`: (1) round-trip encode→decode for transactions with 0, 1, and `MAX_FRAMES` frames; (2) assert null target encodes as `0x80` and decodes to `nil`; (3) assert blob fields zero when empty; (4) fuzz test with random byte inputs to `DecodeRLP`. |
| **Definition of Done** | All unit tests pass; `go fmt ./...` clean; `go vet ./...` clean; no regression in existing `tx_frame_test.go`; code reviewed and merged; coverage ≥ 80% on new encode/decode paths. |

### Task 1.1.2 — EIP-2718 Transaction Envelope Integration

| Field | Detail |
|-------|--------|
| **Description** | Register `FrameTxType = 0x06` in the EIP-2718 typed transaction envelope dispatcher in `pkg/core/types/transaction.go` (or equivalent dispatch file). Ensure `TypedTxData` switch statements handle `0x06` for `Hash()`, `SigningHash()`, `RawSigningHash()`, `Cost()`, `Gas()`. Confirm that `Transaction.Type()` returns `0x06` for `FrameTx` instances. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Unit test: create a `Transaction` wrapping a `FrameTx`, call `tx.Type()` and assert `0x06`; call `tx.Hash()` and assert deterministic output; serialize to bytes, deserialize, compare hash. |
| **Definition of Done** | `tx.Type() == 0x06` passes; round-trip hash equality holds; no panic in switch fallthrough; code reviewed; no regressions in existing typed-tx tests. |

### Task 1.1.3 — `CalcFrameTxGas` Total Gas Computation

| Field | Detail |
|-------|--------|
| **Description** | Implement or complete `CalcFrameTxGas(tx *FrameTx) uint64` in `pkg/core/types/tx_frame.go`. The formula is: `FRAME_TX_INTRINSIC_COST (15000) + calldata_cost(rlp(tx.frames)) + sum(frame.gas_limit for all frames)`. `calldata_cost` uses standard EVM rules: 4 gas per zero byte, 16 gas per non-zero byte, applied to the RLP encoding of the frames list only (not the full transaction). Overflow must be detected and return `math.MaxUint64` or panic with a sentinel error. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Unit tests: (1) single frame with all-zero data yields `15000 + calldata_bytes*4 + gas_limit`; (2) frame with mixed zero/non-zero data; (3) `MAX_FRAMES` frames with large gas limits to test overflow guard; (4) compare output with hand-computed values from the EIP data-efficiency tables. |
| **Definition of Done** | All tests pass; overflow handled without panic; `go vet` clean; reviewed. |

---

## EIP-8141 Reference Excerpts

### Specification → Constants

| Name                      | Value                                   |
| ------------------------- | --------------------------------------- |
| `FRAME_TX_TYPE`           | `0x06`                                  |
| `FRAME_TX_INTRINSIC_COST` | `15000`                                 |
| `ENTRY_POINT`             | `address(0xaa)`                         |
| `MAX_FRAMES`              | `10^3`                                  |

### Specification → New Transaction Type

> A new [EIP-2718](./eip-2718.md) transaction with type `FRAME_TX_TYPE` is introduced. Transactions of this type are referred to as "Frame transactions".
>
> The payload is defined as the RLP serialization of the following:
>
> ```
> [chain_id, nonce, sender, frames, max_priority_fee_per_gas, max_fee_per_gas, max_fee_per_blob_gas, blob_versioned_hashes]
>
> frames = [[mode, target, gas_limit, data], ...]
> ```
>
> If no blobs are included, `blob_versioned_hashes` must be an empty list and `max_fee_per_blob_gas` must be `0`.

### Specification → Gas Accounting (gas limit formula)

> The total gas limit of the transaction is:
>
> ```
> tx_gas_limit = FRAME_TX_INTRINSIC_COST + calldata_cost(rlp(tx.frames)) + sum(frame.gas_limit for all frames)
> ```
>
> Where `calldata_cost` is calculated per standard EVM rules (4 gas per zero byte, 16 gas per non-zero byte).

### Rationale → Data Efficiency

> **Basic transaction sending ETH from a smart account:**
>
> | Field                             | Bytes |
> | --------------------------------- | ----- |
> | Tx wrapper                        | 1     |
> | Chain ID                          | 1     |
> | Nonce                             | 2     |
> | Sender                            | 20    |
> | Max priority fee                  | 5     |
> | Max fee                           | 5     |
> | Max fee per blob gas              | 1     |
> | Blob versioned hashes (empty)     | 1     |
> | Frames wrapper                    | 1     |
> | Sender validation frame: target   | 1     |
> | Sender validation frame: gas      | 2     |
> | Sender validation frame: data     | 65    |
> | Sender validation frame: mode     | 1     |
> | Execution frame: target           | 1     |
> | Execution frame: gas              | 1     |
> | Execution frame: data             | 20+5  |
> | Execution frame: mode             | 1     |
> | **Total**                         | 134   |
