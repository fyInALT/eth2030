# Canonical execution system transitions verification

Date: 2026-11-19. Branch: `fix/glamsterdam-system-transitions`. Tested production source: `3355156e45b16f11eef2acea20931fd77ab70ab9`, independently based on qualified composition `95382598`; the Gas draft is separate.

## Protocol source and repaired behavior

The complete selected [EIP 2935](https://github.com/ethereum/EIPs/blob/3db38a3725105cb67b8804ac3e147967113bc3ca/EIPS/eip-2935.md) and [EIP 4788](https://github.com/ethereum/EIPs/blob/3db38a3725105cb67b8804ac3e147967113bc3ca/EIPS/eip-4788.md) were read before implementation. Their exact runtime bytecode is checked against unmodified Geth v1.17.10 assets. EIP-2935 specifies history address `0x0000F90827F1C53a10cb7A02335B175320002935` and an 8191-block ring; the old native helper used an obsolete address and 8192.

Both EIPs allow direct-storage optimization, so direct writes alone were not treated as a bug. The actual defect was inconsistent canonical storage plus fabricated writes for missing/custom/reverting runtime allocations. Four real StateProcessor cases with identical frozen allocations reproduced complete state-root differences against unmodified SDK system transitions. All four roots now match exactly; [the result record](glamsterdam-system-transitions-results-20261119.json) retains them.

Builder, legacy builder and importer share `ApplyPreExecutionSystems`. It calls deployed code as SYSTEM_ADDRESS with zero value/price, before user transactions, in beacon-then-history order. Missing code produces no synthetic storage, custom code executes and reverting code rolls back without invalidating the block, matching the SDK. Genesis performs no system invocation. The callback supplies the complete block/hash/slot/blob-fee context, separates storage originals per system operation, clears access/transient/refund state and reconciles actual final BAL changes. Untouched system caller state is excluded.

Request queues share the same execution boundary while retaining their fatal failure/output validation. System operations produce no user receipts and debit no block/user gas pool. The [selected EIP-8037](https://github.com/ethereum/EIPs/blob/3db38a3725105cb67b8804ac3e147967113bc3ca/EIPS/eip-8037.md) 16-slot state margin is supplied separately from 30M execution gas; final metering composition is requalified with the separate Gas draft rather than inferred from this branch's pre-existing gas implementation.

The differential bridge no longer copies native ring writes into its Geth pre-state or disables EIP-161. It invokes the SDK's own system transitions with normal rules. Its existing complete-state tests pass after explicit owned simulation allocations and complete system-account snapshots. Native custom-precompile injection remains a separate bridge feature; the four golden root tests instantiate unmodified SDK EVM/system functions directly.

## Explicit genesis and actual consumers

Canonical history/beacon runtime accessors return fresh bytes. The explicit owned-genesis allocator supplies nonce-one contracts and preserves existing custom code/nonce. No runtime/fork callback deploys code or rewrites supplied genesis. Seven existing owned Kurtosis profiles now supply both canonical assets before EL/CL genesis roots are generated; their pre-existing allocations remain byte-for-byte equivalent as parsed objects.

The actual Node HTTP test verifies the independently computed genesis hash, then builds and imports two empty blocks in separate native Nodes. Header gas remains zero. Ordinary eth_call reads the canonical history and beacon contracts; current/future/short history queries and an overwritten beacon timestamp are rejected. A gap of 8191 slots exercises timestamp-ring overwrite without linear catch-up. Complete header hashing agrees with SDK decoding, and fresh receiver restart preserves head/root and contract readback.

Fixtures that required functioning system storage now deploy the contracts explicitly. The binary-overlay test has intentionally absent system code: its manual sparse-root expectation now contains no fabricated storage, and competing payload independence is tested with a real withdrawal difference. Existing protocol obligations are retained.

## Tests and mutation sensitivity

All 212 Go packages, 27 tagged Core/EIP packages and 11 affected race-tested packages pass. Affected vet and local build pass, and affected suites are repeated after restoring mutation edits. Run from `pkg`:

```sh
go test -count=1 ./...
go test -tags integration -count=1 ./core/... ./eiptest
go test -race -count=1 ./core/eips ./core/execution ./core/block ./core/chain ./geth ./node/... ./eiptest
go vet ./core/eips ./core/execution ./core/block ./core/chain ./geth ./node/... ./eiptest
go build ./...
```

Seven compiling mutations fail intended consumer assertions: wrong canonical history address; wrong history readback window; bypassed actual runtime; wrong system caller; omitted reconciliation of reverted BAL writes; runtime code injection for missing contracts; bypassed actual builder pre-call. All edits are restored. Local logs: `/tmp/glamsterdam-system-transitions-{red-runtime,first-green,node-first,all-third,tagged-third,race-third,final-affected}.log`, plus the seven mutation logs and genesis-profile record.

## Compatibility and remaining gates

The corrected history address/window and absent-code semantics change roots compared with experimental histories produced by the old implementation. No old database or history was rewritten. Such experimental chains require a fresh genesis/chain or verified resync under an explicitly compatible rule set.

This qualifies system-transition/root/ABI/BAL consumers, explicit owned configuration and native Node recovery. It does not qualify a live Kurtosis run, an independent full-client process campaign, final Gas/authentication choices, mainnet resource limits or complete release acceptance. Those program gates remain open.
