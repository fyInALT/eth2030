# Complete ordinary blob validation and proof consumers

[PR 76](https://github.com/alt-research/eth2030/pull/76), branch `fix/glamsterdam-blob-validation`, based on `origin/master` `38ea9259`.

## Normative input and reproduced failures

Read [EIP-4844](https://github.com/ethereum/EIPs/blob/3db38a3725105cb67b8804ac3e147967113bc3ca/EIPS/eip-4844.md) (blob `de2e4a46ad93fc04e6fe3174dc6e90a3307bdb5f`) and [EIP-7594](https://github.com/ethereum/EIPs/blob/3db38a3725105cb67b8804ac3e147967113bc3ca/EIPS/eip-7594.md) (blob `fd5dde8f8925fc318027722dfa049dcc39d3bee3`) in full. Legacy wrappers need matching blob/commitment/proof counts and authentic KZG proofs. PeerDAS version 1 needs all 128 ordered cell proofs per blob, commitment/versioned-hash binding and full verification. Type-3 transactions are limited to six blobs.

The real HTTP node accepted an ordinary signed type-3 wrapper whose last cell proof was replaced by the first. This compiled baseline fails `TestOrdinaryBlobCellProofsThroughActualHTTP` with incorrect-proof admission (`/tmp/glamsterdam-blob-http-baseline.log`). The Frame path already authenticated these proofs; ordinary admission did not.

The actual legacy Engine readback also exposed capitalized/base64 fields instead of the two lowercase DATA fields required by [BlobAndProofV1](https://github.com/ethereum/execution-apis/blob/main/src/engine/cancun.md#blobandproofv1) (observed file blob `7339531899f8f2eca2197077d350c400c77ea9a6`). A cell proof cannot substitute for a blob proof.

## Implementation and actual consumers

- Preserve ordinary wrapper versions through decode; reject unsupported versions. Pooled encoding retains both legacy and PeerDAS sidecars without changing the signed body/hash. The production pooled packet codec already uses this encoder.
- Ordinary pool admission checks the six-blob bound, hash prefix, complete supplied sidecar shape, commitment hashes and legacy or complete cell KZG proofs. It uses the existing batch verifier, shared with Frame admission. Expensive proof checks follow ordinary cheap validation and affordability checks. The current full-data pool rejects missing sidecars; a future sparse manager must stage authenticated partial data separately.
- The production pool-based builder requires full data shape, including an allocated-but-empty sidecar case. `BuildBlockLegacy` remains the explicit-transaction offline/fixture constructor: it is not called by production payload builders and is not made dependent on historical sidecar retention.
- The production transaction journal writes and reloads pooled wrappers, including rotation, so fresh node replay retains the sidecar and revalidates it. Historical consensus block encoding remains unchanged.
- Legacy blob retrieval derives a genuine blob proof from a cell-proof sidecar. BlobAndProofV1 HTTP encoding uses exactly `blob` and `proof`, both hex DATA; internal commitments remain backend-only.
- Shared Engine payload collectors preserve authenticated ordered cell proofs, derive legacy blob proofs when needed, and extend legacy full data when necessary. Missing data and proof-generation errors propagate to payload publication; zero-filled substitute proofs are removed from both backend paths. Empty bundles use arrays. This unifies demonstrated divergent proof handling, without changing consensus transaction encoding.

No new runtime switch or environment flag is introduced.

## Verification and regression effectiveness

The real node HTTP test uses an independently signed Geth type-3 wrapper and real KZG material. It rejects a wrong final proof and unknown wrapper version before pool adoption, accepts the valid transaction with the same hash, checks native and SDK pooled encodings byte-for-byte for both wrapper versions, decodes the actual pooled response with independent RLP, and verifies legacy Engine proof authenticity and exact HTTP DATA fields. It then rotates the real journal, stops the node, constructs a fresh node on the same data directory, and reads the replayed sidecar through Engine again. Invalid hash prefixes and more than six blobs are also rejected.

Additional tests cover mismatched data/commitments/counts/proofs/version, genuine legacy proof validation, metadata/empty/full builder availability, per-blob versus per-cell proof cardinality, retained order and proof-generation failure. The Frame HTTP regression still passes. Existing fee-policy fixtures now carry authentic sidecars, and the former sparse-admission placeholder explicitly qualifies the full-data boundary rather than claiming sparse relay.

Controlled mutations each fail their intended test; sources were restored afterward:

- Ignoring the batch KZG result admits the bad last proof.
- Restoring the old legacy-proof selection serves a cell proof that fails blob verification.
- Restoring the builder's non-nil-only test includes an allocated empty sidecar.
- Restoring plain consensus encoding in journal insert/rotation loses the sidecar at fresh-node replay.

Logs: `/tmp/glamsterdam-blob-mutation-{kzg_result,legacy_proof_kind,builder_full_data,journal}.log`.

The full Go suite passes (210 tested packages on the independent branch). Affected suites, targeted race including journal/restart, vet and local build pass:

```sh
cd pkg
go test -count=1 ./...
go test -count=1 ./core/... ./txpool/... ./p2p/... ./engine/... ./node/... ./rpc/... ./eiptest
go test -race -count=1 ./core/block ./txpool ./engine/payload ./node -run 'TestPoolBuilderWaits|TestOrdinaryBlob|TestBlobBundlesRetain|TestFrameBlobSidecarThroughActualHTTPAdmission'
go vet ./core/types ./core/block ./txpool ./engine/... ./node/... ./p2p/ethproto
go build ./...
```

This is the full-data validation and consumer foundation for GL-042, not completed eth/72 sparse relay. Previously stored bare blob entries have no recoverable data in their signed body; this change cannot recreate lost data. Custody updates, provider/sampler roles, GetCells/Cells negotiation and reconstruction, full peer/devnet recovery, inclusion-list-specific availability and final gas qualification remain separate open tasks. Local acceptance of legacy wrappers does not qualify a complete post-Fusaka gossip policy.
