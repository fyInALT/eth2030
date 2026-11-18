## Canonical execution system transition qualification (2026-11-19)

Correct history address/8191 ring and shared actual-code pre-block callbacks now produce identical complete roots to unmodified SDK in deployed/missing/custom/revert scenarios. Native Node HTTP build/import/query/wrap/restart and exact runtime/genesis consumers pass. Seven owned profiles preserve existing allocations while preloading canonical code before roots. All 212 Go packages, 27 tagged Core/EIP packages, 11 affected race packages, vet/build and seven compiling detected mutations qualify this source. [Evidence](docs/verification/glamsterdam-system-transitions-20261119.md). Gas draft composition, live independent-client/devnet/resource and final release gates remain open.


## Moving pivot proof binding composed regression (2026-11-19)

Source `c0df6062` composes PR 81's unchanged-account exact-proof repair with reviews 65–82. All 212 Go packages, 27 tagged Core/EIP packages, affected snap/scheduler/Node race, vet and build pass. Seven encrypted Node pivot scenarios now cover unchanged pending storage and its damaged proof. The Gas draft remains separate; final cross-client/resource/release gates stay open. [Evidence](docs/verification/glamsterdam-snap2-binding-integration-20261119.md).

## Unchanged storage account pivot proof repair (2026-11-19)

Reproduced `storage account not verified` over real encrypted Node sync when an unchanged account still has partial storage after a pivot move. The verified storage consumer now obtains an exact account inclusion proof against the fresh pivot before trusting its storage root. Seven encrypted Node scenarios pass, including damaged-proof rejection for the unchanged account; two added compiling mutations are detected. Full Go, tagged Core/EIP, affected race/vet and build pass. This closes the reproduced cache-binding gap without waiving root authentication. Independent-client/resource and final release gates remain open.


## Snap version 2 socket lifecycle follow up (2026-11-19)

Explicitly stop the test-owned P2P servers before Node/database restart and on every scenario exit. Five encrypted Node cases pass again under race, with affected vet and build. This makes the test resource ownership match the manually started component setup; production protocol behavior is unchanged.

## snap version 2 state catch up (2026-11-19)

The production node negotiates snap/2 and snap/1, preserving range/code schemas and rejecting removed trie-node messages on version 2. Raw positional BAL responses distinguish missing 0x80 from available c0; the actual sender and receiver use the custom codec. Header commitments and canonical BAL structure are verified before staged state rebasing.

CL-selected moving pivots reuse partial account/storage progress. Reorg recovery restores old-only entries through exact inclusion/absence proofs and applies ordered new-fork diffs; missing orphan BALs discard partial state and restart bulk download. Final roots and CL authority gate adoption. Five real encrypted Node scenarios and an unmodified Geth SDK packet consumer qualify cursor preservation, rollback/deletion, restart and invalid-proof rejection. Seven compiling guard mutations are detected. Full Go, tagged Core/EIP, affected race/vet and build pass. [Evidence](docs/verification/glamsterdam-snap2-20261119.md). Full independent-client cold sync, mainnet resource limits, peer reliability policy, eth/72, final gas and release gates remain open.

## EIP-8282 native builder requests (2026-11-18)

Canonical contracts and mandatory Amsterdam type-3/type-4 extraction now compose with the existing snapshot and Engine commitment path. Explicit owned genesis allocation precedes roots; missing production predeploys invalidate blocks. Actual HTTP and separate real-BLS Lighthouse component consumers qualify request bytes and registration/exit gates. Full Go, tagged Core/EIP, targeted race/vet/build and failing guard mutations pass. See [verification](docs/verification/glamsterdam-builder-requests-20261118.md). Final gas, live full-client/devnet and release gates remain open.

## Negotiated eth/69-71 and historical consumers (2026-11-18)

Branch `feat/glamsterdam-eth71` registers eth/68/69/70/71 on the production node. Implements seven-field range status without TD for 69+, version-correct bloom-free receipts, nonblocking queued range updates and canonical raw-list BAL wire encoding. Independent golden bytes expose the old double-wrapped BAL response.

Correlated historical fetchers paginate receipts, bound decoding by requested block transaction/gas limits, verify the complete receipt root and rebuild persistent metadata only for validated blocks. BAL replies are matched to peer/request/hash, checked against headers before any restore, and reject tamper/reordering/replay/unsolicited data. Unavailable markers preserve body availability and are not trusted as arbitrary BALs.

Independent Geth RLPx/codec tests negotiate all four versions over real encrypted sockets and check status plus actual receipt/BAL formats. Real-chain eiptests validate restored receipt metadata and tampered BAL preservation; a disabled consumer hash-check mutation fails the intended tests. Affected eth/P2P/chain/node/eiptest suites, targeted race, full build and vet passed. Full independent-client/devnet, eth/72/snap/2 and complete sync/recovery qualification remain separate program gates. [Evidence](docs/verification/glamsterdam-eth71-20261118.md).

## Sparse relay composed regression (2026-11-19)

At source `20b1c1f`, submitted feature reviews 65–82 compose: all 212 Go packages, 27 tagged Core/EIP packages, 97 affected race packages, vet and build pass. Real eth/72 and snap/2 consumers pass alongside prior Engine/blob/configuration/request paths. [Evidence](docs/verification/glamsterdam-eth72-integration-20261119.md). Final gas, complete CL/multiclient/resource and release qualification remain open.

## Sparse blob relay qualification (2026-11-19)

[PR 82](https://github.com/alt-research/eth2030/pull/82) connects production eth/72 flat correlated cell exchange, elided signed metadata, authenticated bounded staging, provider/sampler custody requests, 64-cell reconstruction and full-data-only pool promotion. Real encrypted native-pool peers qualify the two-provider threshold, random extra, custody expansion and eager mode. Actual Node HTTP/Engine/P2P/restart consumers and independent SDK packet consumers pass. All 212 Go packages, 27 tagged Core/EIP packages, 36 affected race packages, vet/build and eight failing behavioral mutations qualify this source. [Evidence](docs/verification/glamsterdam-eth72-20261119.md). Full independent-client gossip, live CL custody reconnect, resource/fairness/WAL campaigns, selected gas and final release gates remain open.


## Snap version 2 composed regression (2026-11-19)

At source `13969eaf`, feature reviews 65–81 compose successfully: all 211 Go packages, 27 tagged Core/EIP packages, affected race/vet and build pass, including five encrypted Node snap/2 scenarios. Repeated identical tracker sections are removed. [Evidence](docs/verification/glamsterdam-snap2-integration-20261119.md). Final gas, eth/72 and complete CL/multiclient/release qualification remain open.

## Qualified blob schedule and settlement (2026-11-18)

The current version-1 signed HTTP scenario passes at 7/10/21 blobs through producer, independent receiver, receipts/RPC and fresh-node replay. Canonical per-fork capacity, pricing, fee burn/affordability, opcode context and persisted parameter authority agree with independent Geth formulas. Six behavioral mutations fail; full Go/tagged, affected race/vet/build pass after composing PR 76 and PR 79. The earlier legacy-wrapper race timeout is retained as failed evidence, not erased. See [verification](docs/verification/glamsterdam-blob-schedule-20261118.md). Final gas, sparse relay/state sync, CL/devnet and release gates remain open.

## Ordinary blob proof validation and Engine publication (2026-11-18)

Reproduced and repaired invalid ordinary cell-proof admission through real HTTP. Preserve wrapper versions and pooled sidecars, authenticate KZG data, wait for complete pool data, derive genuine legacy proofs and expose canonical BlobAndProofV1 DATA. Shared payload collectors no longer fabricate zero proofs or publish missing data. Journal insert/rotation and fresh-node replay preserve authenticated sidecars. Frame checks remain qualified. Affected suites, targeted race/vet and build pass; four controlled mutations fail as intended. See [verification](docs/verification/glamsterdam-blob-validation-20261118.md). Complete eth/72 sparse relay and release gates remain open.

## Blob schedule implementation in progress (2026-11-18)

Fork-aware block capacity, pricing, ordinary fee settlement, opcode/receipt/RPC consumers and metadata history guards are implemented. Signed HTTP producer/receiver controls pass at 7/10/21 blobs. The full module and tagged suites pass; the 21-blob race run hits the unchanged Engine timeout in the legacy proof-generation route. Compose the submitted PR 76 proof path and qualify current version-1 wrappers before completion. Final gas/release gates remain open.

## Persisted configuration recovery (2026-11-18)

Actual Node startup now checks persisted runtime configuration before history replay, rejects active or immutable changes and retains compatible future updates. Runtime overrides do not mutate network globals. Real signed HTTP/restart evidence preserves head/state/config bytes on rejection and verifies successful future updates; missing/corrupt authority is explicit. Full Go, affected race/vet/build and four failing behavioral mutations pass. See [verification](docs/verification/glamsterdam-config-recovery-20261118.md). Final gas, blob schedules, unknown legacy migration and release gates remain open.

## Canonical schedule and Engine wire boundaries (2026-11-18)

Reject incoherent canonical genesis schedules in actual startup, add Osaka to fork identity and fix asymmetric activated-fork comparisons. Production Engine versions now enforce their timestamp windows; Osaka getPayloadV5 has its own correct shape. Actual HTTP tests cover required fields, DATA/QUANTITY bounds, zero values, malformed BAL, capabilities and forkchoice adoption before attribute errors. Full Go/tagged tests, affected race/vet/build and four failing behavioral mutations qualify this scope. See [verification](docs/verification/glamsterdam-fork-schema-20261118.md). Persisted compatibility and broader program/release gates remain open.

## Builder-request feature composition (2026-11-18)

PR 77 composes with the previous feature set. Isolated blob readiness, factory non-injection and memory slot tests now use an explicit owned genesis with mandatory builder contracts. Existing feature obligations are preserved. At source `d97be4c3`, all 211 tested Go packages, tagged Core/EIP, affected vet and local build pass. Final gas and independent live release gates remain open.

## Glamsterdam signed candidate-gas integration checkpoint (2026-11-18)

The isolated feature integration passes the full Go suite, including the new differential command. The actual 44-case signed corpus preserves 19/22 Amsterdam candidate mismatches and separately classifies Osaka transition-versus-outer-validator checks; it does not claim release parity. See [integrated observations](docs/audits/glamsterdam-gas-2026-11-18/INTEGRATED.md). All remaining gas/CL/network/release gates remain open.

## Glamsterdam independent gas corpus (2026-11-18)

Added an offline signed-transaction differential command with 22 scenarios across Osaka controls and a pinned Amsterdam reference (44 comparisons). Reports preserve complete inputs, gas dimensions, validity versus included failure, logs and final state. The independent context explicitly includes CPSB; a targeted mutation proves its sanity check detects missing state creation pricing. Scoped tests/race/vet and local build pass. This is mismatch evidence, not release parity; see [verification](docs/verification/glamsterdam-gas-corpus-20261118.md).

## Glamsterdam slot commitments across Engine backends (2026-11-18)

Preserved explicit zero slots through production HTTP Engine build/import, and supplied the missing slots to memory Engine construction/import. Independent go-ethereum header hashes now agree with actual payloads, SLOTNUM execution, public block RPC and committed historical BAL bytes. Slot zero, nondefault and uint64 maximum values are covered. See [verification](docs/verification/glamsterdam-slot-commitment-20261118.md); full independent-client/schema/CL and final gas gates remain open.

## Historical empty BAL cache recovery (2026-11-18)

The bounded native Engine backend now regenerates the canonical empty BAL only when its header commitment matches, even without a cached BAL object. Both historical body APIs retain the execution body; nonempty/unknown commitments stay unavailable. This extends PR 66's slot-based retention qualification.

## Glamsterdam BAL availability window (2026-11-18)

Branch `feat/glamsterdam-bal-retention` updates EIP-7928 availability to 14,299 epochs. The bounded Engine backend uses actual header slots when available, with a conservative block-number fallback for legacy headers; both ByHash and ByRange preserve the body when BAL is unavailable. The production chain backend already persists validated canonical BALs on mainline and retains them beyond the minimum window.

Tests exercise the prior-window gap, exact cutoff, outside cutoff and missed slots through both Engine body APIs. A targeted mutation restoring 3,533 epochs fails the intended availability assertions. Affected storage/Engine/node/eiptest suites, build and targeted vet passed. This closes the obsolete-window subtask of GL-040; physical pruning, complete independent CL/recovery and full network qualification remain separate gates. [Evidence](docs/verification/glamsterdam-bal-retention-20261118.md).

## Glamsterdam EIP-8024 bytecode and activation (2026-11-18)

Corrected native DUPN/SWAPN/EXCHANGE immediate permutations and reserved ranges, and activated the instructions at canonical Amsterdam. Published bytecode vectors, all valid depths/pairs, stack and jump boundaries, and native transactions/fork transitions pass. Affected suites, targeted race/vet and the local build pass. See [verification evidence](docs/verification/glamsterdam-stack-opcodes-20261118.md). Independent-client and full release gates remain open.

## Glamsterdam EIP-8246 execution and BAL (2026-11-18)

Implemented balance-preserving destruction, transaction-end account clearing, journaled native creations and rollback, correct CREATE2 creator nonce/prechecks, and final-state BAL reconciliation. Native and canonical Frame tests cover late credits, CREATE2 reuse, existing contracts, fork boundaries and four state backends. Affected suites, targeted race/vet and the local build pass. See [verification evidence](docs/verification/glamsterdam-selfdestruct-20261118.md). Independent-client and final gas qualification remain open.

## Glamsterdam protocol ETH transfer logs (2026-11-18)

Branch `feat/glamsterdam-transfer-logs` activates 7708 at canonical Amsterdam independently of the private gas timestamp and removes extra fee-burn and self-beneficiary burn events. Specified transfers keep the ERC20-compatible system event; same-address/zero-value transfers are suppressed. Existing CREATE/CREATE2 endowment emission is retained and verified before constructor logs, with failed-call rollback preserved.

Native regressions reproduce baseline extra fees, missing Amsterdam activation and incorrect creation log counts; after correction, fork-boundary, fees/self/zero-value, ordering and rollback cases pass. The new logs exposed the modern P2P decoder zero-bloom bug, fixed in shared review #70 and included here. Affected core/P2P/Engine/node/RPC/eiptest suites, full build and targeted vet passed. Full 8246 destruction/finalization and the remaining gas model are separate gates; use a fresh experimental chain to qualify corrected historical fork semantics.

## RPC access-list discovery (2026-11-18)

Branch `feat/glamsterdam-access-list` replaces the eth_createAccessList empty placeholder with production backend discovery. Independent state copies execute native transactions until the access list converges; returned gas is measured with the returned list, without selecting an empty list merely because it is cheaper. Reverted child touches are retained, lists/keys are deterministic, and malformed addresses/keys/quantities return invalid-params errors. Unsupported backends no longer report fictitious discovery.

The actual HTTP/node-backend regression first failed on the placeholder and now passes. Additional tests cover reverted inner calls, contract-creation accesses, historical versus current state, no canonical nonce/storage mutation and malformed input. RPC/node/eiptest suites, targeted race, full build and vet passed. This delivers GL-044 access discovery for current CallArgs; final gas-model/API semantic and broader authorization/blob argument qualification remain GL-020/025/045 work. [Evidence](docs/verification/glamsterdam-access-list-20261118.md).

## Glamsterdam canonical deterministic factory (2026-11-18)

Branch `feat/glamsterdam-factory` uses the EIP-7997 factory at 0x4e59b44847b379578588920cA78FbF26c0B4956C with its exact 69-byte runtime and nonce 1 for explicit genesis/test state. Removed block-processor and both builder paths that injected a different contract at a fork. Private chains must include the canonical alloc or deploy it with an ordinary transaction before activation; no fork existence check or new switch is added.

Canonical address/runtime/nonce and actual CREATE2/value/return-byte/collision tests plus a no-fork-injection regression first failed for the intended reasons and now pass. Preserved legacy private BLS 0x12 without changing the canonical 0x0b..0x11 map. Core/Engine/node/eiptest suites and full build passed. This implements the canonical factory path; live network/genesis tooling qualification remains in the broader integration gates.

## Glamsterdam EIP-7954 limits (2026-11-18)

Branch `feat/glamsterdam-contract-limits` updates deployed/initcode limits to the published 64/128 KiB EIP-7954 revision, retaining pre-fork limits. Activation follows canonical Amsterdam independently of the private gas-repricing timestamp. Oversized Shanghai+ creation transactions are rejected before fees/nonce/gas-pool mutation, and post-fork pool envelope policy leaves space for a valid maximum-size initcode transaction.

New CREATE/CREATE2 exact-limit tests and transaction/pool boundary regressions first failed on the baseline for the intended limits/admission reasons, then passed. Affected core/txpool/Engine/node/RPC/eiptest suites, local full build and targeted vet passed. This closes the size/admission portion of GL-030; the coherent gas model, complete 8024 qualification and independent-client release gates remain separate work. [Evidence](docs/verification/glamsterdam-7954-20261118.md).

## Modern receipt bloom restoration (2026-11-18)

Branch `fix/eth-receipt-bloom` repairs a receipt-root interoperability defect: eth/69+ omits bloom on the wire, but the decoder OR-ed the newly constructed receipt's zero bloom instead of computing it from decoded logs. It now uses LogsBloom. A nonzero-log typed receipt roundtrip first failed on baseline and now preserves exact consensus bytes. p2p/ethproto, eth and eiptest suites plus full build passed. The fix is also required by the negotiated protocol and transfer-log feature reviews.

## Glamsterdam feature review batch (2026-11-18)

Published independent reviews from current master: [#65 contract/initcode limits](https://github.com/alt-research/eth2030/pull/65), [#66 BAL availability window](https://github.com/alt-research/eth2030/pull/66), [#67 canonical factory](https://github.com/alt-research/eth2030/pull/67), and [#68 native RPC access discovery](https://github.com/alt-research/eth2030/pull/68). Each branch retains its affected test/build evidence and explicit remaining qualification boundaries. The negotiated eth/69-71/historical-consumer feature is prepared on its own branch and awaiting review publication.

No review is marked merged or release-qualified here. The [implementation tracker](docs/plans/glamsterdam/IMPLEMENTATION.md) distinguishes merged pre-existing header/history fixes from newly submitted work. Full gas/state, remaining protocols, CL/builder and combined independent-client/recovery gates remain open; the original dirty checkout is untouched.

## Glamsterdam implementation baseline (2026-11-18)

Started the 33-task completion plan on isolated branch `feat/glamsterdam-baseline`, based on freshly fetched `origin/master` at `38ea9259`. Imported the dated plan/audits without replacing newer mainline work. [Implementation tracker](docs/plans/glamsterdam/IMPLEMENTATION.md).

Revalidation finds BAL/slot canonical headers, execution-slot propagation, canonical persisted Engine bodies and the old tagged expectation already repaired on mainline. Those defects must not be reimplemented from the stale audit branch. The latest targeted core/Engine/node/eiptest suite is green; source/spec decisions and remaining size/factory/retention/RPC/gas/network work remain under qualification. Large functional changes will use separate branches and PRs.

## Merged roadmap and homepage progress (2026-11-18)

- Refreshed the active roadmap, progress report and next-stage/CL iteration
  sequence after PR 61 merged at `911afb0c`. Replaced obsolete aggregate completion
  claims with qualified results and independent-client/service acceptance gates.
- Homepage and `/roadmap` share a dated development panel: sponsored slot 66,
  blob slot 165, fresh CL join at slot 256 and receiver restart. The panel links
  the merged PR and immutable run record and states the synchronized-EL boundary.
- Updated Frame, FOCIL and historical BAL evidence paths, scoped gaps and the
  maintenance guide. Historical tracker labels remain distinct from current-spec
  conformance; no fork completion percentage was inferred from the paired control.
- Site production build, 386 EIP implementation/test paths, retained-record export
  assertions and the local Go build passed. Browser checks covered homepage and
  roadmap at 1440px/390px, milestone visibility, phase/detail navigation, layout
  overflow and JavaScript errors. Pages deployment now checks the exported claims.
- No devnet was started for this documentation/site refresh. The temporary static
  preview was stopped after verification. Publication follows review/merge through
  the existing Pages workflow.

## Qualified paymaster, blob and fresh CL history (2026-11-18)

- Native EL `e0451092` and Lighthouse `a4ceb86e` passed the complete capped
  control. Coded sponsorship finalized through signed FOCIL at slot
  66; the blob Frame finalized at slot 165.
- A validator-free CL with a new database joined at slot 256,
  checkpoint-synced and returned both original finalized historical envelopes.
  Raw transaction identities, Beacon roots, execution hashes and ordered blob
  commitments matched the source. It shares the synchronized second EL.
- Actual receiver restart preserved sponsored readback and the original blob
  receipt, parent/post application storage and both paymaster balances. The
  [final record](docs/verification/frame-8141-20261118-paymaster-history-final.json)
  retains runtime/image identities, source versions, resource caps and cleanup.
- The owned enclave, containers and volumes were removed and its engine stopped.
  Seven owned runtime images and temporary binaries/contexts were also removed.
  Native Go tests/build/vet/race, real consensus SSZ/block replay, twelve Python
  controls, Heze regressions, Git-version cache tests, cargo check and full lint
  passed. The CI matrix has seven required jobs; the final review head is
  checked after publishing the retained evidence.
- Remaining promotion gates: independent-client/runtime ABI agreement,
  independent EL late join/state sync, Dora's production epoch consumer and
  service restart, external-builder/PTC and other separately planned profiles.
  Earlier entries below retain intermediate runs and their explicit failures.

## Qualified CL runtime packaging (2026-11-18)

- The new runtime reports `a4ceb86`, including the checkpoint and worktree-version
  fixes. The first image accidentally added an entrypoint despite the pinned
  package's explicit VC executable, so validator startup failed before traffic.
  The entire failed enclave and its volumes were removed and its engine stopped.
- The [runtime Dockerfile](pkg/devnet/kurtosis/scripts/lighthouse-runtime.Dockerfile)
  now records the package's command contract. Beacon/observer entrypoints remain
  supplied by their Starlark services; validator commands run directly.

## Heze checkpoint parent follow-up (2026-11-18)

- EL `e0451092` now serves reconstructed finalized sponsored, blob and checkpoint
  envelopes through Lighthouse's historical API. The same run preserved blob
  receipt, original storage and sponsor balances across a real receiver restart.
- Fresh CL range sync exposed the next CL-specific failure: the Heze notifier
  resolves the trusted checkpoint's IL view through its pre-anchor parent,
  which a checkpoint-synced database does not contain. A debug observer at
  checkpoint slot 224 repeatedly failed on the missing slot-223 parent.
- Lighthouse fix `d8a064bf` on `fix/heze-checkpoint-history` passes five Heze
  regressions, cargo check, full lint-fix and the release build. The Heze spec
  initializes the anchor's IL satisfaction to true; the fix must preserve that
  trusted bootstrap and keep unresolved live-parent views strict. A regression
  uses real signed blocks and an empty checkpoint store with Engine notification
  forced on, rather than manually marking the anchor payload received.
- The failed enclave, temporary debug observer and labeled volumes were removed
  and its engine stopped. The [record](docs/verification/frame-8141-20261118-checkpoint-parent-failed.json)
  distinguishes repaired source history from the still-open fresh CL gate.

## Historical Engine payload bodies (2026-11-18)

- The fresh CL control exposed actual historical serving failures after the
  source CL pruned its payload cache: base64 transaction DATA and absent BALs
  prevented envelope reconstruction, so range sync stalled at the checkpoint.
- Payload-body V1/V2 now use hex DATA, correct pre-Shanghai null withdrawals,
  an explicit nullable V2 BAL and hex range quantities. Ranges reject invalid
  bounds with the specified errors, preserve unavailable internal positions
  and trim heights past the canonical head without uint64 overflow.
- Native chain validation stores the reconstructed RLP BAL atomically with the
  validated block. Historical lookup checks its header commitment and retains
  the execution body when BAL data is absent/pruned. Native hash/range V1/V2
  HTTP regressions compare original transaction and BAL bytes on both nodes and
  after FileDB reopen, at nonce zero and the final valid scalar nonce.
- The separate Engine backend now returns RLP BALs and preserves historical
  transaction bodies beyond its BAL-retention window. Corrupt stored BAL data
  is rejected by a real build/import/storage regression.
- The rebuilt live receiver also reopened the original blob receipt, parent/post
  application storage and both sponsor balances after a real EL restart. This
  check now runs in the repeatable controller, with negative readback controls.
- Removing BAL persistence and removing the V2 DATA codec were each detected
  by the actual HTTP/restart integration using isolated Go build overlays.
- The failed observer run cleaned all enclave resources and stopped its engine;
  a rebuilt runtime must pass the same fresh-CL control before closing history.

## Frame upper nonce boundary (2026-11-18)

- Reproduced public admission rejecting the last valid scalar nonce
  (`2^64 - 2`) because `stateNonce + MaxNonceGap` overflowed. The lower-nonce
  guard now precedes a subtraction-based gap comparison.
- The failing-before-fix pool regression now passes. The actual HTTP/Engine
  integration also runs at this nonce: admission, FOCIL omission verdict,
  construction, both imports, receipt/storage readback and FileDB reopen all
  preserve the final post-approval nonce (`2^64 - 1`).
- The running live image remains `38bbcff5`; this boundary correction is covered
  by local node tests and is identified separately from that live image.

## Consensus service codec replay (2026-11-18)

- Added a separate [consensus replay module](tools/consensus-replay/README.md)
  against released ethpandaops Go client `v0.1.11`. The actual finalized Heze
  state is retained as compressed SSZ with its Beacon header and spec response.
- The published Heze decoder rejects the current CL layout by 92 fixed bytes
  because EIP-8015 removed four Eth1 fields. A minimal upstream-shaped replay
  type matches byte-for-byte SSZ and the original state root, every validator
  key and balance, and the 96-slot PTC window. Changed state, root, epoch JSON,
  PTC preset and execution-requests-root controls are rejected. Removing the
  state-root comparison was detected by two behavioral regressions.
- Original sponsored and blob BeaconBlock SSZ also roundtrip exactly; their
  message roots match the finalized state's block-root window. Signatures,
  execution identities, FOCIL bits and commitments match captured evidence.
  Changed bits, commitments, signatures and unrelated EL identities fail.
- CI now runs the standalone replay, its race regressions and all twelve
  Python control tests when either protocol code or the replay module changes.
- The March local Dora codec also failed actual Gloas state decoding. It lacks
  the relocated latest block hash, execution requests root and PTC state layout.
  Successful block JSON display does not qualify epoch statistics. The replay
  candidate does not yet wire the Dora consumer or qualify its restart path.

## Frame paymaster completion (2026-11-18)

- Frozen ETH2030's [paymaster v1 candidate](pkg/core/framepay/README.md), retaining
  the EIP reference owner and delayed-withdrawal storage layout. Its Yul runtime
  is reproducible with the recorded compiler and settings; the witness uses the
  protocol-validated canonical-hash signature at index 1.
- Prefix execution, exact-code recognition, multiple sponsored senders,
  aggregate max-fee exposure, pending/matured withdrawal reservations and code
  change revalidation pass behavioral tests. Actual EVM tests cover deployment,
  unauthorized requests, exact timelock boundary, transfer and rollback.
- Four targeted mutations were detected: relaxed runtime identity, omitted
  withdrawal reservation, removed owner check and removed timelock. The last
  two recompiled the actual Yul runtime and failed behavioral state tests.
- The workload CLI now supports the frozen coded paymaster. A bounded control
  covers sponsored FOCIL, a blob Frame and a validator-free checkpoint-synced
  CL observer reading older execution envelopes. The observer shares an already
  synchronized EL; independent EL late join remains a separate acceptance gate.
- The first coded-paymaster control finalized the sponsored FOCIL transaction
  on both peers. Its subsequent blob submission exposed the real HTTP path's
  use of the plain transaction decoder; both public RPC entrypoints now decode
  pooled wrappers. A real HTTP regression rejects a valid KZG proof at the
  wrong cell index, admits the corrected wrapper and retrieves all 128 ordered
  proofs through Engine API. The test also exposed and fixed `BlobAndProofV2`
  JSON (hex DATA and lower-case field names, per Engine Osaka).
- Replacements now test rejected payer changes preserving the old exposure and
  successful payer changes releasing it for another sender.
- The next live run finalized both coded sponsorship (slot 66) and a blob
  Frame (slot 162), with identical receipts and exact sponsor settlement.
  The observer launch rejected incompatible checkpoint/genesis CLI options;
  its failure path removed the entire enclave and stopped the engine. The
  observer now uses checkpoint sync exclusively.
- This is an explicitly versioned experimental implementation; upstream bytecode
  and ABI agreement, live coded sponsorship and fresh CL history are next.

## Canonical Frame execution (2026-11-18)

- Branch `feat/frame-8141-execution` extends merged master `cc567f93` in an
  isolated worktree. The [execution record](docs/frame-8141-execution.md)
  maps the pinned EIP to transition, journal, state-gas, admission, network,
  receipt, Engine and workload paths.
- Behavioral tests pass for sponsored settlement, sender creation, code
  deposit, introspection, atomic rollback and cross-frame refills. Real KZG
  sidecar validation and eth/70 receipt-page tests pass.
- Build/import, same-height reorg, FileDB reopen and production startup
  replay tests pass for sponsored and blob-bearing Frames. Native HTTP
  tests pass for RPC admission, list construction, V6 import, FCU V5 and
  receiver restart. The tests exposed and fixed the trusted-genesis Bogotá
  FCU and execution-sync guards.
- The full Go suite, targeted race checks, local build and vet checks pass.
  Three behavioral mutations were each detected by their regression. Nine
  Python controller/readback tests pass. The resource-capped sponsored Frame
  control passed on Lighthouse `a05a56de` with three native EL revisions.
  Controlled tests do not establish independent-client interoperability.
- The first live launch used Heze at genesis and failed while decoding the
  generated beacon SSZ (`OffsetOutOfBounds`) on Lighthouse `a05a56de`.
  Cleanup removed its labeled containers and volumes and stopped its engine.
  The Frame control now uses the previously qualified Fulu genesis with
  scheduled Gloas/Heze transitions; the generator compatibility gate remains open.
- The compatible live pair passed with native EL `88d9c118` and Lighthouse
  `a05a56de`: a signed list at slot 66 delivered the sponsored Frame to the
  isolated proposing EL at slot 67. Both pairs finalized it. The sponsor paid
  exactly 137946955735149 wei; the sender balance stayed unchanged. The receiver
  reopened the same receipt and historical storage after restart. Cleanup
  reported zero remaining labeled containers and volumes and stopped the engine.
- Follow-up review wires the production head's fee and gas-limit context into
  pool admission, handles zero-valued Go fee metadata safely, and supplies the
  including block's blob base fee to RPC transaction formatting. Targeted
  regressions pass, including binary-overlay receipt parity. KZG admission
  now uses the production batch verifier, with a valid proof deliberately
  assigned to the wrong final cell as a rejection case. The updated images
  and retained signed-list traces passed the paired control.
- EL `6d10283d` repeated the finalized sponsored Frame control with the raw
  transaction, original signed list, dependent root and payload bid retained.
  Both peers agreed on settlement and restart readback. Its enclave, labeled
  containers and volumes were removed and the engine stopped.
- A real CREATE deploy-prefix regression exposed an approval nonce rewrite:
  payment approval must increment the current account nonce after deployment,
  rather than reset it from the signed transaction nonce. Deployment and
  overflow rollback regressions now pass. The final consensus corner-case
  fix was qualified on the retained `1c20e3c7` runtime.
- EL `1c20e3c7` passed the final paired control with Lighthouse `a05a56de`:
  slot 65's original signed list delivered the Frame at slot 66, both pairs
  finalized it, and the receiver returned identical historical data after
  restart. Raw bytes, signed list, payload bid, receipts, balances, image IDs
  and memory limits are retained in the [final run evidence](docs/verification/frame-8141-20261118-final.json).
  The sponsor paid exactly 137949520800762 wei; sender balance was unchanged.
- The four-account deploy-prefix post-state matches an independently encoded
  go-ethereum StateAccount/StackTrie oracle. Clipped-budget tests reconcile
  actual payer debit and proposer tip even when GASPRICE is only a lower bound.
  Public self-verify now requires the actual combined APPROVE scope, rather
  than treating separate successful scopes as a combined call.
- Non-FCU head changes now supply fee/gas context during the locked pool
  state reset, covering P2P insertion, asynchronous execution sync and native
  fork choice. The stale-head admission regression passes without RecordBlock.
- Draft [PR 61](https://github.com/alt-research/eth2030/pull/61) retains the
  implementation and evidence. All owned runtime images and temporary test
  binaries were removed after verification; Docker reports no running test
  containers, enclave networks or engine containers.
- The later paymaster/history follow-up qualifies the local runtime, withdrawal
  accounting and fresh CL control. Formal cross-client agreement and service
  cutover remain separate promotion gates.

## ETH2030 PR readiness review (2026-11-17)

- Upstream master is `ad35748f` with zero open upstream PRs at the
  review snapshot. The [PR readiness review](https://github.com/alt-research/eth2030/blob/911afb0c42a2e33c9e65881b31ed1b1c446390b8/docs/next-stage-work-plan.md#pr-readiness-on-the-current-eth2030-master)
  identifies the Bogotá branch for direct review, scopes the devnet
  controls to a 14-file delta on the Bogotá base, and records the
  documentation candidate and Frame integration dependency.

## Next-stage planning snapshot (2026-11-17)

- The [next-stage work plan](docs/next-stage-work-plan.md) records upstream
  ETH2030 master at `ad35748f`, the pushed integration branches, and the
  exact binary revisions used for past live evidence.
- Read-only merge simulation of Lighthouse `413541ba` and upstream
  `3777cfd7` found a conflict in the FCU verdict area of
  `beacon_chain.rs`. The plan scopes that reconciliation and the next
  Gloas column regression.
- P0 covers package review, patched-pair qualification and pinned Frame
  execution. Independent Track 3 late join, moving-pivot recovery and
  ordinary service adoption have parallel gates.

## Follow-up review of the upstream integration

- ETH2030 `db91e7ef` closes the malformed-attributes and background
  execution-sync bypasses in Bogotá fork choice. A deferred payload with
  missing parent data now returns SYNCING and can be retried when the
  parent arrives. Native HTTP, race and package regressions passed.
- Lighthouse `413541ba9` keeps Req/Resp requests distinct across dependent
  roots and committee bitfields at the same slot. Both release network
  tests and the Lighthouse build check passed.
- Devnet controls `2239039e` verify process-group termination, retained
  cleanup ownership, Fulu transaction readback after finality and signed
  Gloas self-build lineage. All six Python tests passed. The preceding
  live control evidence remains pinned to its original binaries.

## Upstream adoption implementation (2026-11-17)

- ETH2030 branch `feat/devnet-upstream-controls` at `9a52e0ae` records
  the pinned Fulu control. Both pairs finalized the same execution block at
  beacon slot 64. EL 2 reopened the original type-7 receipt and historical
  storage. The controller removed its enclave and stopped its Kurtosis
  engine, with zero remaining labeled containers and volumes.
- Lighthouse staging branch `integrate/eth2030-upstream-1bcdcbfa` at
  `bc56927f` passed `cargo check`, the Fulu-to-Gloas light-client regression,
  and three Heze integration tests covering the proposer list snapshot,
  the empty-head FCU ancestor, and verdict persistence. Its release build
  produced `v8.4.0-rc.0-bc56927`, and the container reported that revision.
- The ETH2030 native deferred-validation branch at `4b10a445` covers
  missing-data SYNCING, off-canonical ACCEPTED, FileDB reopen before and
  after FCU, and replacement of the submitted view. The native CLI built,
  and the Amsterdam and Bogotá HTTP tests passed with the race detector.
  Integration branch `b7944602` also passed the pinned Frame codec tests.
- The paired `bc56927f` and `b7944602` images verified the same self-built
  Gloas-at-genesis envelope, execution state, slot and BAL at slot 4.
  The Heze control later finalized a transaction delivered from EL 1's
  isolated mempool through its slot 64 list and EL 2's slot 65 proposal.
  The Fulu, Gloas and Heze control evidence is on `929870c9`. Each
  enclave was removed before starting the next. The independent deferred
  EL gate remains open.
- Upstream Spamoor `f0f83695` passes its Frame codec tests and matches the
  ETH2030 base-envelopes' RLP and signing hash. Its scalar nonce upper
  boundary still differs. The upstream Go codec's inclusion-list JSON
  cases pass on its declared Go 1.25.2. Its released SSZ path decoded live
  Gloas slot 32 and rejected the retained ETH2030 Heze slot 64 state.
  The old Heze schema remains pinned for Dora. Service cutover,
  pinned Frame execution and the newer package remain open.

## Devnet upstream adoption implementation plan (2026-11-17)

- The [implementation plan](docs/devnet-upstream-implementation-plan.md)
  maps the audit to eight PR groups with source pins, code entry points,
  verification and per-profile promotion conditions.
- The native Bogotá work distinguishes the missing-data `SYNCING` path
  from staging an eligible off-canonical `ACCEPTED` payload and completing
  its verdict through FCU. The plan also separates upstream image swaps from
  Heze/Frame execution, Dora migration and package qualification.

## Devnet fork implementation ownership (2026-11-17)

- The [feature-level disposition](docs/devnet-fork-upstream-audit.md#feature-level-disposition)
  assigns overlapping released and maintained-mainline features to upstream
  implementations. For experimental features it specifies the ETH2030 owner,
  matching upstream architecture, local behavior to carry and retirement gate.
- Lighthouse unstable moved to `1bcdcbfa0`. Its builder configuration and
  metric changes leave the Heze implementation decisions in place. The audit
  also records the two independent inclusion-list reads in the experimental
  producer and the shared snapshot required by the ETH2030 self-build path.

## Devnet fork and upstream adoption audit (2026-11-17)

- The [upstream adoption audit](docs/devnet-fork-upstream-audit.md) pins Lighthouse
  stable, unstable and the upstream Heze devnet branch, plus upstream Dora, Go client,
  Spamoor and ethereum-package candidates. It records the retirement gates for each
  maintained fork.
- The upstream Lighthouse Heze branch covers VC duties, immediate V6 verdicts and
  Req/Resp serving. Deferred verdicts, embedded-bid recovery and local Frame blob
  extraction still require the ETH2030 path for the audited profile.
- Upstream Dora already ingests and persists inclusion lists. Its Heze codec has a
  different SSZ layout from the current ETH2030 CL. Upstream Spamoor covers the
  Bogotá Frame format, with scalar-nonce and receipt-failover differences to check.

## Devnet third-party inventory (2026-11-17)

- The [devnet service inventory](docs/devnet-third-party-inventory.md) records the
  maintained forks, nested source pins, upstream image services, profile choices and
  prior run evidence. The checks covered 51 tracked profiles and the companion Heze
  fixture.
- Compared the Dora and Spamoor fork refs with the images selected by older profiles.
  The two `v0.0.1` local tags lack a registry digest and a source revision label.
- Documented the Assertoor quick-run result boundary and the dependency on the
  fork-specific genesis generator. The verification used existing checkouts, images and
  recorded results.

## Lighthouse CL gap audit (2026-11-17)

- The [CL gap audit](docs/lighthouse-cl-gap-audit.md) records the published Lighthouse
  `1bf89a4fc` baseline, upstream `e723ca55f`, and the companion EL and Dora revisions.
- The EL still needs to retain Bogotá lists for deferred FCU validation. The pinned
  Frame transaction format has encoding coverage but is rejected at mempool admission
  and execution. The audit records the real-client acceptance tests for both paths.
- Targeted EL Frame and Bogotá V6 tests passed during the audit. Prior capped devnet
  runs provide the live CL evidence.

## Heze consensus validation (2026-11-16)

- Lighthouse `1bf89a4fc` includes upstream `ca9dff457`. The beacon store and slasher redb tests reopened a copy of an uncleanly closed database. The regular Lighthouse release image built and started its version command.
- The deferred Engine regression sends `engine_newPayloadV6`, receives `ACCEPTED`, and then sends `engine_forkchoiceUpdatedV5` from a beacon head on its empty node. The wire server returns `VALID,false` for the executed ancestor. Its execution validity and inclusion-list failure survive fork-choice reload. The test uses a controlled Engine HTTP server. Independent EL interoperability remains open.
- In `eth2030-heze-cl-gossip-loss`, both peers verified Gloas slot 107 at root `0xdd454aca869b6709ac6610265dfc5ae68187170a20b9e4bab06363188a4c097c`. The same peers agreed on the finalized Fulu checkpoint at slot 64 and root `0x8498839b333daa3f611d309b8a8904b162110093aab9b5fe6c113d3b5d76866d`. The initial monitor required both conditions in the Gloas window. Separate queries captured the verified Gloas block and the finalized Fulu checkpoint.
- The receiver's test image discarded the gossip for validator 98 at slot 133. The embedded self-build bid in slot 134 triggered Req/Resp and the receiver recovered the same signed list, including transaction `0x4e9a16ac2b5279ed7891f8cf73bddccab05c7bb3cb323849e53f092080c69800` and committee positions 1, 9. Recovery was late and did not make the list timely. Reopening the receiver preserved block `0x265a3e70bc3855c132807514c1d3613477f24b9e3a1806c0920ef02c87e80e0e` and the prior finalized checkpoint. The six-node suite used 2048 MB per beacon node or EL and 1024 MB per validator. It stopped after verification.
- The CL historical envelope test accepts a Frame transaction referencing two distinct KZG commitments in order. It rejects reversed commitments and a changed payload hash. Live sponsored and blob-bearing Frame execution still depends on the EL.

## Pinned Frame encoding and inclusion history (2026-11-16)

- ETH2030 branch `feat/frame-8141` at `67aef736` decodes the pinned seven-field EIP-8141 envelope and preserves the eight-field historical route. Independent vectors cover the transaction and signing hashes, secp256k1 and P-256 signatures, fee budgets and split-gas receipts. Typed receipts retain both gas dimensions and payer when reopened from FileDB and projected through RPC. Execution and mempool admission still reject the pinned frame type until its state transition is implemented.
- Heze fixture commit `e628c770` adds 2 GiB budgets for each EL and beacon node and 1 GiB for each validator.
- Lighthouse `5d4500270` serves the retained previous-slot view keyed by the proposal's block root and shuffling-dependent root. Its HTTP integration test verifies ordered committee positions and the first signed observation of an equivocating validator, ignores a list on another dependent root, and returns 404 after pruning.
- Dora `12a0198` stores views by block root and displays the bid bits by committee position. Its Go Beacon client at `79989d2` reads both bid message fields and falls back to the JSON API when the older Gloas or Heze block decoder fails on current SSZ. Dora retains the source JSON for historical blocks. Epoch statistics still need the current BeaconState SSZ layout.
- The capped two-pair devnet `eth2030-heze-dora-capped-r5` finalized a compulsory transaction from slot 32 in the slot 33 block `0x3de7e15dc9873301bd04dc1cf0f00dbb0ba382aee4ac9cfa2395805609e26c40`. Lighthouse returned 404 after pruning the list. A restarted Dora rendered the bytes of transaction `0xf01e49efb5461bc3307110ee9ca0f38f619ba43606d071481c603df986c8fbcf` under committee positions 0 and 8. All six node memory caps were verified and the enclave stopped after verification.

## Devnet host memory incident (2026-11-16)

- Before the 19:05 CST reboot, 16 ETH devnet enclaves had 92 EL, beacon and validator containers running without per-container memory limits. At 19:01 CST, the kernel OOM process table attributed about 31.2 GiB of anonymous resident memory to 60 Lighthouse, 24 ETH2030 and eight Geth processes. System activity records showed the 8 GiB swap essentially full from 18:40 onward.
- The kernel repeatedly invoked the global OOM killer starting at 18:52 CST. At 18:57 and 18:59 the allocations that triggered it came from validator containers in two ETH devnets. Several terminated processes were developer tools. The next boot recorded a reset-pin event and unclean filesystem recovery. This establishes devnet memory pressure as a major contributor and leaves the immediate cause of the reset undetermined.
- Docker showed zero running EL, beacon or validator containers after reboot. The Docker and kernel records remain available for inspection.

## Bogotá Engine and Heze inclusion recovery (2026-11-16)

- Published Lighthouse `eth2030` at `dd48a965e` over upstream unstable `7a9d14115`. Heze sends the parent-dependent timely inclusion-list view in `engine_newPayloadV6`. An execution-valid envelope with unmet inclusion requirements remains execution-valid in fork choice, and the rejection of its payload extension survives database reload. The linked ETH2030 branch is `fyInALT/eth2030:feat/bogota-inclusion-list-wire` at `42e7ada9`. The upstream Gloas database-invariants test passed on the runtime revision `dd7263111`. The test-only follow-up applies both V6 outcomes to the same payload hash on separate harnesses and asserts the proposer decision before and after fork-choice reload.
- Implemented Heze `InclusionListsByIndices` request and response, scoped by slot, dependent root, and committee bitfield. Signed responses pass the inclusion-list validation path. The tests verify SSZ/Snappy framing with fork context and previous-slot import, storage, and serving with termination. Lists first learned in the previous-slot recovery test are excluded from the timely view.
- Two native ELs returned `VALID,true`, `VALID,false`, and `VALID,true` for the same Bogotá execution block as its list input changed. The latest paired run accepted V5 and rejected V6 at Gloas slot 63, then accepted V6 and rejected V5 at Bogotá slot 64. Both wrong-fork calls returned `-38005`. The V6 release image also passed the Fulu Track 4 regression on both pairs.
- In `eth2030-bogota-isolated`, both ELs reported zero execution peers at submission and inclusion. The transaction `0xc6b8cfb159f79ae202333abe9d405a9bdd1c9dbaa452d6a424d7964193085842` was submitted to EL 2 at slot 74 and absent from EL 1's RPC. EL 1's Lighthouse saw the signed list, and validator 115 included the transaction at slot 75. Its receipt matched the finalized block `0xc92be1b65abbb0827e68bf25e244cfe7970f78440ad874f44fb04a3c3ecd0361`. In a separate same-parent Engine control, the transaction appeared only in the build with list delivery.
- Amsterdam headers now commit to the EIP-7843 slot. The genesis hash matches Geth v1.17.10. The node integration rejects a slot-only mutation and confirms the canonical slot and `SLOTNUM` value after restarting the verifying EL. The separate `eth2030-amsterdam-direct` devnet started at Gloas genesis on two pairs and agreed at slot 7 on the execution hash, state root and BAL hash.
- The updated Fulu control finalized on both pairs. After an EL restart, RPC reopened on a new Docker host port. Querying the new binding returned the same finalized history, canonical Track 4 receipt and state root as the other EL.
- The Heze runbook includes an isolated-mempool fixture and a verifier that requires the proposer mempool negative check, the matching signed list, canonical inclusion in the next slot, and matching finalized observations from both pairs. With Lighthouse runtime `dd7263111` and native EL `db9f384b`, `eth2030-heze-noelpeer` delivered transaction `0x769c08305e1626c98ef18cbd7551dc899d97618c16791f9740830721ebe019f5` only to EL 1, recovered its signed slot 66 list on the other beacon node through gossip, and included it at slot 67. Both pairs finalized execution block `0x76ef95ef0fbd718cb917e8191975390157d048fe573b99be08f9f260e73fb901` with matching receipts and zero EL peers at the verifier checkpoints. The profile clears Kurtosis-injected EL bootnodes and keeps distinct EL network IDs after a control using network IDs alone observed intermittent execution peers.
- The pinned EIP-8141 EL migration, sponsored and blob-bearing Frame interoperability, Dora's durable inclusion-list ingestion, and independent late-join acceptance are still open. The native decoder rejected an independently encoded seven-field Frame payload at RLP decode, before reaching execution. The Req/Resp regression exercises the protocol codec and service pipeline. Live recovery under deliberate gossip loss and an independent EL that exercises deferred satisfaction through forkchoice remain open.

## Lighthouse migration implementation (2026-11-16)

- Integrated upstream unstable `8104a2467` and fast-forwarded the fork's `eth2030` branch to `4e1167314`. Ported the Heze gossip check for missing timely list bits and added EIP-8141 local hash verification for historical and optimistically imported envelopes. Lighthouse compilation, lint and targeted Engine, gossip, historical envelope, signer and Heze HTTP tests passed.
- Wired Heze validator duties, candidate retrieval from the EL, signing and publication. The first live run exposed a missing BN duties endpoint. After adding the route, both BNs returned all 256 epoch 2 duties with the same dependent root and emitted the same nonempty signed list at slot 64. Its transaction bytes hashed to the submitted transaction `0x41ce93275b4289fb67a7af188e6f1db1cbb5de74950b06332b34d40d1687751c`. Both BNs returned nonempty bid bits at slot 65 and both ELs returned the same successful receipt. The two-node mempools were connected in this run.
- The paired run also passed the Fulu Track 4 verifier. The earlier migration image served matching Gloas envelopes and EL state roots on both pairs at slot 34. The EL companion commit `e1df4b17` returns the Bogotá transaction array and excludes EIP-4844 transactions from committee selection. Its Engine, EIP and node backend tests and build passed.

## Lighthouse legacy branch integration audit (2026-11-16)

- The [legacy branch integration plan](docs/lighthouse-legacy-branch-merge-plan.md) maps all 68 old Lighthouse commits outside `eth2030` to their current call sites and work packages. A whole-branch trial merge reported 76 conflicts. Both historical EL feature tips are already ancestors of `origin/master`.
- Traced the live Engine from node startup to `pkg/node/backend`. That backend already forwards canonical inclusion-list attributes to the builder, while the legacy Engine backend has a separate global list and status enum. Documented the actual V6, gas target, and Frame blob changes and the corresponding node tests.
- The existing focused Engine, FOCIL and node backend tests passed. The update records the migration plan and source findings.

## Lighthouse and Hegotá implementation plan (2026-11-16)

- Recorded the published Lighthouse `eth2030` branch at `3e3978d4e` and the verified local-build inclusion-list snapshot change in `6e5317104`.
- The [implementation plan](docs/lighthouse-hegota-implementation-plan.md) pins the initial Frame, Heze and Bogotá references, specifies CL/EL ownership, and lists the separate Frame, ordinary FOCIL and combined sponsored-Frame acceptance cases. The current update records planning and source verification.

## Lighthouse baseline update and historical patch audit (2026-11-16)

- Merged upstream unstable `1a704e58f` into Lighthouse branch `integrate/unstable-20261116` at `f85fa05b6`. It includes the `v8.3.3` stable release and has one documentation symlink beyond upstream. Formatting, compile checks, a release container build, 40 Engine API tests, 14 blob-fetch tests, and 30 inclusion list tests passed.
- Official stable, pinned official unstable, and the integration build each validated a live Track 4 PQ transaction with two ETH2030 ELs and two Lighthouse beacon nodes at Fulu. Both official images finalized epoch 2 with matching execution hashes across both ELs and both CLs. The [baseline verification](docs/cl-iteration-roadmap.md#lighthouse-baseline-verification) records the image IDs, source revisions and matching execution blocks.
- The integration build crossed from Fulu to Gloas and served the same signed envelope and execution block on both pairs at slot 34. At Gloas genesis, generator `6.2.8` resolved the SSZ decode error seen with `5.3.5`. The next blocker is the native EL genesis header. Adding the empty BAL hash and slot zero to the EL header reproduced the exact execution hash both CLs expect.
- The Heze path still needs VC list signing and publication, proposer bid bits, the V6 envelope import, and forkchoice consumption of inclusion satisfaction. The EL needs the matching Bogotá RPCs, and Dora needs to move from the March inclusion-list endpoint to the current event and committee model.

## Consensus client iteration plan (2026-11-16)

- Audited EL master after PR #55 and compared the local Lighthouse integration with upstream stable and unstable. The [iteration plan](docs/cl-iteration-roadmap.md) prioritizes Gloas/Amsterdam regression evidence and the missing Heze/Bogotá Engine contract, including nonempty inclusion lists and a sponsored Frame transaction.
- The plan keeps upstream Lighthouse as the release CL and specifies separate acceptance for late join, timing, proof providers and native consensus research. This update records the plan and source audit.

## Track 4 integration with master (2026-11-16)

- Track 4 block staging and canonical publication now use the shared execution validation and Engine forkchoice path. The block batch retains the binary overlay nodes and conversion cursor. Startup rebuilds missing validation markers for legacy canonical overlay history after replay and root checks.
- The optional native consensus verifier resolves ML-DSA keys from its configured validator registry. Full Go tests, local build, vet, and targeted race tests passed during merge reconciliation.

## Bogotá follow-up review

- Malformed FCU V5 attributes now use the Bogotá fork-choice path.
  Asynchronous execution sync also waits for an evaluated Bogotá list
  view before publishing its head. Deferred FCU requests sync when an
  accepted block's execution ancestry is unavailable.
- Authenticated HTTP regressions cover malformed attributes for resolved
  and unresolved heads and resumption after the staged block's parent
  arrives. The affected Go packages, targeted race tests and native build passed.

## Native Bogotá deferred validation (2026-11-17)

- The native Engine retains known off-canonical V6 blocks and their list
  input before returning ACCEPTED. FCU runs the existing execution path,
  evaluates the retained view, and stores its independent verdict.
- An authenticated HTTP integration test covers missing-parent SYNCING,
  an altered header, two views of the same payload, restart between stage
  and FCU, and restart after a negative verdict. A state-synced block with
  no submitted view requires V6 reconciliation before selection.
- Engine, backend and chain package tests passed. The deferred Bogotá and
  Amsterdam HTTP restart tests passed with the race detector. The native
  CLI compiled. The Lighthouse paired and independent-EL gates remain open.

## Track 4 integration and devnet verification (2026-11-16)

- Added a two-node HTTP integration test for the private profile. It checks Geth-compatible genesis hashing, Engine authentication, pre-fork gating, both EIP-8051 variants and state-backed EIP-7932 recovery, and rejects wrong-chain and corrupted signatures. ML-DSA, FN-DSA and SLH-DSA transactions enter real payloads and are validated by an independent EL. The test selects competing branches, checks canonical receipts, and reopens FileDB to read historical and current state.
- Fixed empty genesis byte strings, PQ transaction hash consistency with the signed wire envelope, pool admission before I+ and for wrong chain IDs, and sender and effective gas price reconstruction for persisted PQ receipts. Block-scoped RPC reads now resolve overlay state by block hash, including for historical blocks after restart.
- Added a two-pair Kurtosis profile and an executable verifier. The `eth2030-track4-89cfff9` run used EL revision `89cfff9`. Both ELs and both Lighthouse Beacon APIs verified the PQ payload at slot 5 with block hash `0x2eb74b1df29b3e729ee174772f334fcb5d0ef592a75fc8e9f513771e029cf53d`. After restarting EL 2, historical slots at execution blocks 1 and 5 returned 44 and zero, the old receipt retained its sender and gas price, and the two ELs agreed again at block 11.
- A preceding run of the same private profile reached finalized beacon epoch 2. Both ELs returned the finalized execution block 59, which matched the Lighthouse finalized payload.
- Full Go tests, local build, targeted race tests and no-CGO integration tests passed. The private profile exercises pinned EIPs and keeps the production proof, committee and witness-gas work open.

## Track 4 overlay checkpoints and witness accounting (2026-11-14)

- The private overlay atomically persists changed binary nodes, the frozen root, and the conversion cursor with each imported block. Restart and forkchoice validate an available checkpoint against the reconstructed state. Existing replay-only databases can still open.
- BLAKE3 branches share immutable trie paths and memoized node hashes. Checkpoint publication skips previously committed subtrees. The logical execution state and frozen conversion data remain memory-backed, and startup still replays blocks.
- Witness gas helpers distinguish an explicit zero storage leaf from an absent leaf, drop tentative events for unaffordable opcodes, and clamp code ranges without overflow. Witness gas is still separate from the private overlay's active fork rules.
- Startup rejects private PQ draft addresses that conflict with the current or a later configured precompile set. The published EIP-7932 and EIP-8051 still share address 0x12.
- The full Go test suite, local build, and targeted race tests passed. The regression suite checks corrupted cursors and nodes, batch rollback, legacy replay, changed-path writes, exact witness gas, and a future PQ address collision.
- Open work includes the published binary key layout, complete EIP-4762 transaction and opcode integration, scalable frozen-state access and direct state recovery, node committee provenance, and production STARK, lattice and recursive aggregation protocols.

## Track 4 correctness review (2026-11-14)

- Private binary overlay imports and reorgs publish database batches before changing canonical pointers. Side branches retain receipts, canonical transaction lookup follows forkchoice, and startup rejects missing bodies or a mismatched genesis. FileDB commits its WAL before materializing data files and can replay a full batch after a materialization failure.
- Overlay copies preserve pending writes, snapshots, original slot values, and independent log data. The legacy SHA-256 trie proof API verifies the entire key and suffix proof. Legacy proofs that omit required suffix siblings must be rebuilt for stems with multiple leaves.
- Engine payload IDs commit to all build attributes and the requested slot. The RPC layer keeps valid forkchoice updates when payload attributes fail validation and maps wrapped errors to their Engine API codes.
- PQ attestation verification requires a trusted committee-key lookup for individual and batched ML-DSA signatures. Merkle batches verify their input signatures. Execution proof batches own submitted bytes, enforce size limits and deadlines, and clear stale verification flags.
- Regressions exercise signed transactions on both branches, rejected batch writes and retries, FileDB WAL recovery, distinct Engine roots for the same parent and timestamp, suffix-proof verification, real ML-DSA committee authorization, and proof-provider deadlines.
- Validation passed with `go test ./...`, `go build ./...`, targeted race tests, and no-CGO PQ signer and precompile tests in `pkg`.
- The production EIP-7864 layout and full witness gas, canonical PQ precompile assignments, node committee provider, and production STARK, lattice and recursive aggregation protocols remain open.

## Track 4 private-chain binary overlay execution (2026-11-14)

- A private genesis can set `binaryOverlayTime`, `binaryConversionTime`, and `binaryConversionStride`. At activation the execution path starts with an empty BLAKE3 overlay and a frozen snapshot of the parent MPT state. Conversion uses the pinned EIP-7748 order before system calls and transactions. EVM journal writes determine the binary root, including explicit zero storage leaves and transaction-scoped original storage values.
- The node Engine backend builds and imports overlay payloads and applies forkchoice to the selected branch. Header roots are verified during execution and replay. Restart restores the canonical state and conversion cursor by replaying live blocks. The experimental profile retains full state in memory and retains live block bodies for replay.
- Tests compare two Engine payload roots with separately constructed trie roots. A signed-transaction FileDB test exercises storage clears, branch-specific BLOCKHASH, forged roots, reorg, rewind, restart and missing ancestry. State tests cover snapshot rollback, code shrink, account deletion and atomic in-memory commit failure. Mutation runs confirmed that dropping explicit zero leaves or resolving BLOCKHASH on the canonical branch is detected.
- Validation passed in `pkg` with `go test ./...`, `go build ./...`, and targeted `go test -race` for the state, chain, and node overlay and forkchoice tests.
- The EIP-4762 opcode helper now uses the full storage key for access and edit events. Bytecode tests assert exact gas for repeated slots, adjacent slots in one stem, and slots that differ above bit 63.
- The published consensus layout, EIP-4762 witness gas, durable per-block state checkpoints and scalable state access remain open. Existing opcode-level witness gas helpers are separate from this profile. The PQ committee source and application-specific STARK, lattice and aggregation protocols remain open.

## Track 4 consensus attestation verification (2026-11-14)

- Replaced the PQ attestation hash-trace shortcut with individual ML-DSA verification for all batch sizes and lean modes. The batch enforces a shared signing statement, unique indices and keys, authenticated metadata and committee commitment, and copies the original signed data. Gas savings are zero for this path. Legacy trace-only proofs are rejected.
- XMSS bundles now return an explicit unavailable-verifier error instead of being relabeled as ML-DSA. Propagation verifies the supplied batch and requires an authorized committee key lookup for network publication. Classic fallback recovers the ECDSA signer and requires a trusted slot/committee/validator key lookup.
- Tests use actual ML-DSA and ECDSA signatures. They reject an invalid sixth signature, duplicate signer identity, inconsistent statements, wrong batch metadata, and an unauthorized classic signer. Temporarily removing signature verification caused the invalid-sixth-signature test to fail. Full `go test ./...`, targeted consensus and conversion race tests, and builds with and without CGO passed. A real recursive proof and production consensus-state lookup wiring remain open.

## Track 4 ordered binary overlay conversion (2026-11-14)

- Added a pinned EIP-7748 conversion step for a frozen memory-state snapshot and pinned EIP-7864 BLAKE3 overlay. It iterates in MPT hash-key order, converts storage before account headers and code, and charges a unit for each slot or account. Existing overlay values, including explicit zero, take precedence. Cursor copies support independent branches and checkpoint recovery verifies the frozen MPT root and reopens the overlay.
- Tests cover ordered conversion, exact stride, wrong roots and cursors, branch isolation, explicit zero and live value preservation, and FileDB reopen inside a partially converted account. Reversing either sort or overwriting zero makes the corresponding test fail.
- The node still needs a binary-overlay StateDB, execution and access gas wiring, durable block-to-checkpoint publication, and per-block conversion scheduling.

## Track 4 private-chain PQ execution profile (2026-11-14)

- Wired explicitly addressed EIP-7932/8051 draft precompiles into transaction execution, system calls, and RPC simulation at a configured genesis timestamp. The private profile reads expanded keys from STOP-prefixed code at their key-hash address, charges cold or warm access plus copy gas, and records the account in the BAL. Direct verification keeps the 4500-gas draft cost.
- Both independent SHAKE and Keccak signatures pass through the transaction execution path. Tests exercise activation, malformed and absent key state, exact access gas, reserved and duplicate addresses, RPC simulation, and genesis address parsing. Replacing the cold access cost with the warm cost caused both positive-path gas tests to fail.
- Full `go test ./...`, the targeted execution and node race tests, and local builds with and without CGO passed.
- The addresses, key account format and lookup surcharge are private-chain choices. Canonical fork assignments and registry semantics remain open. Binary overlay execution, EIP-4762 accounting, EIP-7748 conversion, and production proof provider adoption remain open.

## Track 4 historical signers and migration command (2026-11-14)

- Added explicit Round 3 Falcon-padded-512 and Round 3.1 SPHINCS+-SHA2-128f-simple signing APIs using the pinned PQClean clean implementations. Existing transaction IDs continue to select ML-DSA-65, FN-DSA and SLH-DSA. The historical Falcon and SPHINCS+ APIs require CGO and report unavailability without it.
- A standalone PQClean C executable produced the historical wire vectors. Tests accept them, sign using their serialized private keys, reject changed messages, altered signatures, truncated fields and cross-format verification, and exercise key generation. Falcon also verifies case 0 from the pinned EIP-8052 assets and rejects nonzero padding.
- Added `eth2030 binary-migrate` for an existing committed flat-state FileDB, with a trusted MPT source root and durable checkpoints. The command integration test checks source locking, partial conversion, a second invocation that resumes, and balance, storage and code readback from the reopened destination.
- Track 4 review still finds two consensus integration areas and one provider migration open. EIP-7932/8051 need a fork address assignment compatible with EIP-7997 and a state-backed key lookup. EIP-7864 delegates activation to EIP-7612 and access events to EIP-4762, with later conversion under EIP-7748. The offline command implements snapshot conversion. STARK frame and aggregate verification still need a real provider, a concrete public statement and a transport format. Experimental blob openings are authenticated Merkle paths, while the local deterministic lattice scheme remains experimental.

- Validation: full `go test ./...` and `go build ./...` passed in `pkg`. Targeted migration and historical signature race tests passed. A no-CGO build and the tests for the unavailable providers, Dilithium3 and offline migration also passed.

## Track 4 expanded-key verification and reference roots (2026-11-14)

- Implemented the pinned EIP-8051 expanded-key SHAKE and Keccak verifiers. They enforce canonical field encodings, hint ordering and padding, the strict coefficient norm bound, and the challenge hash. Explicit constructors bind the verifiers to the precompile and EIP-7932 registry adapter. EIP-7932, EIP-8051, and EIP-7997 assign overlapping uses to address 0x12, so the tests configure the two draft precompile maps separately.
- The expanded-key known answers use the EIP's 32-byte tr and raw message with independent ZKNOX polynomial and signing code. The generator and reference revisions are in the VM testdata directory. Both variants pass through EVM STATICCALL with exact gas. The tests reject modified messages, signatures, keys, noncanonical coefficients, wrong variants, and out-of-order hints. Temporary mutations confirmed that removing the field range checks, accepting reordered hints, or relaxing the positive norm boundary causes the corresponding tests to fail.
- EIP-7864 tests pin insertion roots from the linked Python reference and a full state embedding generated with the pinned draft's storage offset. The state vector includes inline and main storage, the largest storage slot, and PUSH32 data across chunks. The persisted tree reopens at the independent root and a modified code-chunk opening is rejected.
- Replaced experimental blob witnesses with ordered Merkle paths, binding original byte length, chunk position, and all siblings. The signed wrapper checks the entire digest, the serialized path, and the configured algorithm, and uses SLH-DSA for its SPHINCS selector. This changes the experimental commitment format and requires regeneration. EIP-4844 and PeerDAS retain KZG.
- The experimental deterministic lattice verifier now checks the full commitment, opening, and data for each batch entry, and compares the serialized vector on the blob path. The prior linear batch equation ignored the supplied blob. These checks preserve the local commitment encoding.
- Added an explicit round-3 Dilithium3 signer using CIRCL mode3. Tests cover both directions of wire interoperability, exact deterministic signing, invalid key and signature lengths, and rejection by the ML-DSA-65 transaction provider. Historical Falcon and SPHINCS+ formats, node binary-state activation, and real STARK/lattice/aggregation migration remain open.

- Validation: full `go test ./...`, targeted race tests for the draft verifiers, reference roots, signing and blob verification, and `go build ./...` passed in `pkg`.

## Track 4 validation-frame proof boundary (2026-11-14)

- The former validation-frame prover committed to calldata and a constant; it did not prove EVM execution, state, or EIP-8141 `APPROVE`. The old verifier accepted a proof without binding it to a block, and the node and payload builder discarded the returned proof while keeping the original block. Those paths did not establish execution validity.
- The legacy STARK validation-frame prover now returns `ErrFrameProofUnavailable`. The VM requires an explicit block-bound execution-statement provider to generate and verify a frame proof and identifies only `VERIFY` frames. No real provider is installed; node startup explicitly rejects `StarkValidationFrames=true` instead of advertising unavailable proof generation. Frame calldata is not stripped.
- Tests exercise missing providers, disabled node configuration, no-VERIFY blocks, rejection of an unbound legacy proof, cross-block proof substitution, tampering and provider failure. `go test ./...`, targeted prover/VM/node/payload race tests, and `go build ./...` passed in `pkg`, with Go test cache reused where applicable. The block-bound test provider is a simulated test double, not a STARK or a proof of EVM execution. A real provider, statement format, proof transport and consensus adoption remain open.

## Track 4 binary materialization and state rollback (2026-11-13)

- Added an explicit-error conversion of a complete `MemoryStateDB` including dirty storage to an EIP-7864 draft BLAKE3 tree. It checks code hashes and header constraints and does not mutate the source. Tests compare its root to committed-state migration, exercise snapshot/revert, and materialize, persist and read back an actual transaction execution result. This is a full in-memory rebuild, not a binary `StateDB` or fork activation.
- Fixed implicit account creation so reverting a memory-state snapshot removes the newly created account. Fixed `TrieStateDB` cached-root invalidation on account/storage mutation and rollback, and snapshot tracking of account recreation. Recreated accounts no longer expose or hash old slots, and stale slot deletion now precedes new writes. On a freshly reopened database, storage-root computation scans untouched slots until the complete persisted index is initialized at commit. Tests reopen committed state and validate it through binary migration.
- Memory-state code reads and writes now copy byte slices, preventing external mutation from making stored code disagree with its hash. Existing valid-state roots and transaction encodings do not intentionally change; the rollback and recreated-account fixes correct previously inconsistent state.
- Validation: `go test ./...`, targeted state/execution race tests, and `go build ./...` passed in `pkg`; the final full-suite run reused Go test cache where applicable. No cross-client binary-state vectors or fork integration were exercised. Node adoption, production-scale state, crash recovery and the remaining cryptographic provider migrations remain open.

## Track 4 PQ provider and signing hardening (2026-11-13)

- Added cross-provider tests for the ML-DSA-65 and FN-DSA-512 wrappers, including tampered signatures. The historical Dilithium3 identifier does not accept a genuine round-3 Dilithium3 signature. The SLH-DSA-SHA2-128f signer reproduces a public NIST ACVP deterministic signature, and rejects a modified signature. These results do not establish compatibility with historical Falcon or SPHINCS+ encodings.
- PQ signer registration and lookup now copy descriptors; the configurable transaction validator uses copy-on-write when replacing a verifier. This prevents external mutation of installed entries and races between registration and verification. The consensus transaction verifier continues to use its fixed providers.
- PQ transaction signing and encoding now reject negative unsigned integer fields rather than authenticating or serializing their absolute value. Both signer paths and the transaction validators reject invalid signing payloads. Existing valid nonnegative transaction encodings are unchanged.
- Validation: `go test ./core/types ./crypto/pqc ./core/execution ./txpool/... ./core/vm`, targeted PQ race tests, and `go build ./...` passed in `pkg`. Expanded-key EIP-8051 verification, legacy algorithm interoperability and activation remain open; these tests do not certify those implementations.

## Track 4 checkpointed binary migration (2026-11-13)

- Added an experimental account-batched migration checkpoint keyed by the trusted source MPT root. Each invocation reconstructs the source account root, validates copied account storage and code, opens the previous authenticated binary root, writes the updated nodes, and only then advances the checkpoint. The intermediate root is not published as canonical.
- A file-database test covers zero-address cursor handling, a simulated checkpoint-write failure, clean close/reopen between batches, retry, completion and equality with the independently run one-shot migration. Tests also reject changed source state, malformed checkpoints and missing checkpoint nodes; empty state and repeated completion are covered.
- Validation: `go test ./core/state/... ./core/rawdb ./trie/... ./core/execution`, targeted migration race tests, and `go build ./...` passed in `pkg`. The caller must stop concurrent source and destination writers and use a database with suitable durability. No process-kill or power-loss injection was run. Each batch rescans the complete source account trie, eagerly reloads and rewrites the target tree, and retains that tree in memory. This is not a scalable migration, fork transition, or node activation.

## Track 4 binary trie state operations (2026-11-13)

- Draft BLAKE3 storage deletion now removes absent leaves and restores minimal stem paths; zero-valued storage writes remove the slot. Legacy SHA-256 behavior is retained. Empty-root exclusion proofs are accepted for an empty path.
- Added atomic-in-memory replacement of an existing account's code, updating code size and hash and removing obsolete chunks. Code readback reconstructs and checks the hash and chunk metadata. `UpdateContractCode` remains a lower-level chunk writer, not a full account update. `GetAccount` no longer exposes its stored code-hash slice.
- Tests check delete/reinsert roots against freshly built trees, zero storage, code shrinking across stems, persisted code readback, corrupted chunk metadata and migration code reconstruction. `go test ./trie/... ./core/state/... ./core/execution`, targeted trie/state race tests, and `go build ./...` passed in `pkg`. These operations are not a complete `StateDB` and are not activated in the node.

## Track 4 EIP-8051 draft envelope and registry tests (2026-11-13)

- Added isolated draft precompile gas and input-envelope handling for both proposed ML-DSA variants: 4500 gas, a 32-byte message, 2420-byte signature and 20512-byte expanded key. A valid-length call requires an explicitly supplied verifier; no cryptographic provider or fork address is configured by default.
- Added optional EIP-7932 registry adapters for the draft `0xd1` and `0xd2` signature-info layouts. They require injected key lookup and verifier, validate the returned key hash and 32-byte payload hash, and use the draft's provisional 3000 gas penalty. Tests exercise dispatch, 6000 total gas, input boundaries, missing providers and tampering with simulated callbacks; they do not prove ML-DSA or ML-DSA-ETH conformance.
- The pinned EIP asset case 0 uses a 1312-byte compact key and 33-byte message. Despite its `Dilithium2` file heading, the NIST case verifies with the current FIPS ML-DSA-44 provider and not its round-3 Dilithium2 provider; the EVM case verifies with neither. These cases are not direct vectors for the specified expanded-key precompile ABI. The EVM variant and expanded-key cryptographic checks remain unimplemented.
- Validation: `go test ./core/vm ./core/types ./crypto/pqc ./core/execution`, a targeted VM race test, and `go build ./...` passed in `pkg`. Replaced two tests that copied gas constants into a fake lookup with a comparison against the actual `vm.GasPQVerify` function. That function currently has no production call sites and its local table is not EIP-8051's precompile schedule. Draft address and revision conflicts, actual precompile activation and end-to-end PQ verification remain open.

## Track 4 EIP-7932 pinned-draft SIGRECOVER (2026-11-13)

- Added an isolated `EIP7932DraftSigRecover` precompile and immutable-at-construction algorithm registry for the pinned `refs/EIPs` draft. It implements the draft's secp256k1 selector `0xff`, signature validation and recovery, address compatibility rule, base gas and signing-data penalty. Additional algorithms require explicit registration; none of the project-local PQ identifiers are silently treated as EIP algorithms.
- Tests reproduce the pinned `assets/eip-7932/precompile_test_cases.py` cases, exercise EVM precompile gas dispatch and insufficient gas, reject high-S signatures and invalid registry entries, and check a simulated additional algorithm. `go test ./core/vm ./core/types ./crypto/pqc ./core/execution`, a targeted VM race test, and `go build ./...` passed in `pkg`. This implementation is not registered in a fork or on the network, and is not the later EIP-7932 revision used by the existing SSZ codec. EIP-8051 expanded-key verification, gas, activation and cross-version/address reconciliation remain open.

## Track 4 committed-state binary migration progress (2026-11-13)

- Added an offline `TrieStateDB` flat-table migration to the draft EIP-7864 BLAKE3 trie. Address and storage-slot preimages are present in these project-specific tables. Before writing binary nodes, the migration reconstructs the account MPT against a caller-provided trusted root, validates each account storage root and code hash, and rejects malformed or out-of-range data.
- File-database tests commit accounts with code and storage, migrate, close and reopen the binary database, then read accounts, inline and non-inline storage, code chunks and BLAKE3 inclusion proofs. Wrong source root, altered storage and altered code are rejected before binary node writes.
- Validation: `go test ./core/state/... ./trie/... ./core/rawdb`, `go test -race ./core/state -run TestMigrateCommittedToBinary`, and `go build ./...` passed in `pkg`. This path requires a quiescent source and keeps the converted tree in memory. It does not publish the root, provide a durable checkpoint, run during fork transition, or wire binary state to the node. Generic MPTs with only hashed keys do not provide address and slot preimages. The older `trie/migrate` utilities target a different legacy tree.

## Track 4 aggregation verification boundaries (2026-11-13)

- The previous `STARKAggregator` regenerated a trace commitment from state root, block hash and type; it neither bound the input proof bytes nor retained or verified the outer proof. It could therefore report validity without proving an execution statement. It now requires an explicit `STARKAggregateProvider`, retains its proof bytes, binds all ordered input fields with length delimiters, and asks the provider to verify the result. No production provider or concrete recursive statement is configured.
- `BatchAggregator.VerifyBatch` now requires an individual execution proof verifier before setting `Verified`; a Merkle commitment alone is not proof validation. Verification state resets on each call. The legacy recursive STARK interface does not include complete public statements and now returns an error instead of asserting that digest aggregation proves inner validity.
- Execution-proof commitment hashing is domain-separated and length-delimited, changing prior aggregate roots; old roots must be recomputed. Regression tests use a deliberately simulated provider to check dispatch, artifact and input tampering, missing providers and invalid input; they do not demonstrate a real STARK or recursive verification.
- Validation: `go test ./proofs/... ./txpool/stark ./node`, targeted `go test -race ./proofs`, and `go build ./...` passed in `pkg`. Mempool STARK gossip remains experimental and advisory; its local trace does not prove transaction validity. The real STARK, lattice commitment and aggregation protocol migrations remain open.

## Track 4 PQ signing pipeline progress (2026-11-13)

- Replaced the pipeline's fabricated public key with provider-derived ML-DSA-65 and SLH-DSA public keys. FN-DSA callers must now supply the matching public key; the provider does not expose public-key derivation from the serialized private key. A supplied key is checked against the signature.
- Default pipeline signatures use the same algorithm-bound transaction signing hash and sender derivation as native validation. The experimental direct mode retains its legacy message format and is not accepted by native transaction validation. Previously generated pipeline signatures use the retired default message format and require re-signing.
- Tests sign, encode, decode and validate all three provider formats, reject changed fields and algorithms, and execute transfers through the existing PQ pool validator and state processor. The pinned EIP-7932 draft address derivation test covers its 0xff compatibility case; this does not implement its registry or precompile. Local PQ gas estimates are not EIP-8051 precompile gas.
- Validation: `go test ./crypto/pqc ./core/types ./txpool/... ./core/execution`, targeted race tests for the pipeline and processor, and `go build ./...` passed in `pkg`.
- The pinned EIP-7932 and EIP-8051 drafts both assign address 0x12; EIP-8051 also assigns 0x13, while the active Prague profile reserves 0x0b–0x11 for BLS and uses the existing native profile. The versioned EIP-6404 codec follows a different published EIP-7932 revision with selector 0x00, whereas the pinned draft uses 0xff. These conflicts require an explicit fork/spec decision before consensus activation; EIP-7932/8051 conformance remains open.

## Track 4 binary trie persistence progress (2026-11-13)

- Added experimental content-addressed node persistence for SHA-256 and BLAKE3 trees. `Persist` returns a root only after node writes succeed; root publication and durability are controlled by the caller. `OpenPersisted` eagerly authenticates node hashes and stem paths, rejects corrupt or missing data, and restores a mutable trie.
- A file-database test closes and reopens the database, checks values and inclusion proofs, persists an update, reads the original root, and rejects node corruption. A failed node write does not return a root.
- Validation: `go test ./trie/... ./core/rawdb ./core/state`, `go test -race ./trie/bintrie`, and `go build ./...` passed in `pkg`. The node format is local, eager loading is not suitable for production-scale state, and no node lifecycle, canonical root publication, atomic fork migration or crash-injection test has been added. Existing migration tooling targets a different experimental trie, not EIP-7864 state. Track 4 remains open.

## Track 4 BLAKE3 binary trie progress (2026-11-13)

- `NewWithHashFunc(HashFunctionBlake3)` now computes BLAKE3 roots using the pinned EIP-7864 zero-pair rule. Copies retain the hash selection, and ordinary reads and writes do not expose caller-owned value buffers.
- The structured inclusion and exclusion proof builder supports SHA-256 and BLAKE3, including multiple values and absent suffixes in a stem. Callers must select `NewBlake3ProofVerifier` for BLAKE3 roots. The older `Prove` API rejects BLAKE3; it remains a separate legacy proof format.
- BLAKE3 account, storage and code-chunk operations use the EIP-7864 tree-key helpers. Account updates retain existing storage and reject out-of-range header values. Tests compare the root to a separately computed BLAKE3 construction, exercise state key readback, and reject tampered proofs.
- This is experimental and does not establish persistence, restart recovery, migration or node/consensus adoption. Existing SHA-256 roots and BLAKE3 roots are different formats. The pinned draft says its BLAKE3 choice is provisional.
- EIP-7932/8051 conformance and the STARK, lattice commitment and aggregation migrations remain open; see the Track 4 checklist. Validation: `go test ./trie/... ./crypto/pqc ./core/state` and `go build ./...` in `pkg` passed; full-module tests and external vectors were not run.

## Track 4 proof verification progress (2026-11-13)

- Authenticated divergent stems in the standalone SHA-256 binary trie exclusion verifier and added suffix-tree siblings to inclusion proofs, allowing proofs for stems with multiple values. The proof representation changes: divergent proofs now require the value subtree root; multi-value inclusion proofs require eight sibling hashes.
- Added regression coverage for a forged divergent stem, modified suffix sibling, multi-value inclusion, and unsupported same-stem exclusion. At that commit, proof generation rejected non-SHA-256 tries and same-stem suffix exclusion; the follow-up above adds both.
- The historical Dilithium3, Falcon-512, and SPHINCS+ signer identifiers already use ML-DSA-65, FN-DSA-512, and SLH-DSA-SHA2-128f providers respectively. These names do not establish compatibility with historical scheme encodings. The pinned EIP-8051 specifies ML-DSA-44 precompiles, not the project PQ transaction gas table; EIP-7932 registry/precompile and EIP-8051 conformance remain to be tested and implemented as applicable.
- Track 4 remains open: EIP-7864 uses BLAKE3 in the pinned draft, and proof, persistence, migration, and state integration are not validated end-to-end. STARK, lattice commitment, and aggregation protocol migrations remain open; no interchangeable provider or statement was established in the audit.
- Validation: `go test ./trie/... ./crypto/pqc` and `go build ./...` from `pkg` passed. The full module test suite and cross-client vectors were not run.

# Project Status

## README Link and Quick Start Audit (2026-11-13)

- Updated the GitHub Pages URL, clone commands, badges, and release links. Fixed the fixture and devnet setup commands, documented the Go version required by the module, and directed readers to local devnet setup while the public endpoints are unavailable.

## Homepage Hero Actions (2026-11-13)

- Redesigned the Roadmap, GitHub and Docs actions as equal-sized responsive cards with consistent icons, descriptions, focus states and reduced-motion support.
- Validation: production site build and export checks passed. Browser checks confirmed equal dimensions and valid destinations at 1440px, 390px and 320px, without horizontal overflow; the Roadmap action has a visible keyboard focus state. Changes remain local and unpublished.

## Homepage Roadmap Entry (2026-11-13)

- Added direct Roadmap links to the homepage hero and footer. The site build and export checks passed; browser verification confirmed the hero link is visible at mobile width and opens `/roadmap/`.

## Roadmap Homepage Integration (2026-11-13)

- Promoted the six-card roadmap prototype to `/roadmap/` and integrated a compact six-card summary into the homepage. The homepage displays non-complete tracker entries only, with direct links to the corresponding detailed card; the detailed page retains the full node and status display.
- Both views use the same fork mapping and tracker-count calculation. Percentages are the share of mapped entries historically labelled complete, not independent conformance or readiness measurements.
- Validation: site build and export assertions passed, including all six links, filtering of complete entries and matching homepage/detail percentages. Tracker verification checked 371 paths without missing files or blocking discrepancies. Browser checks covered 1440px/390px layouts, card sizes, anchors and console errors; local Go build passed. This branch builds on the unmerged website refresh PR #52 and is not deployed.

## EF Roadmap Prototype Card Alignment (2026-11-13)

- Made all six roadmap cards equal in height and width within each responsive layout, preserved chronological reading order, aligned card footers, and allowed content to expand the shared grid height instead of being clipped.
- Validation: production site build and export checks passed. Browser checks at 1440px, 1024px, 390px and 320px confirmed matching card dimensions, reading order, and no card or page overflow. This remains a local prototype.

## EF Roadmap Prototype Density and Percentage (2026-11-13)

- Made the six local prototype cards more compact, with responsive three-column layout, visible status ratios, animated percent rings and condensed EIP rows. Each percentage is the count of mapped entries labelled `complete` divided by all mapped tracker entries, rounded to the nearest whole percent; the unmapped EF node is excluded. Historical labels are not independent conformance or production-readiness measurements.
- Validation: site production build and export checks passed; 371 tracker paths checked; browser checks recomputed all six ratios from displayed entries, including zero-complete cases, verified no detail links or console errors, and checked 1440px/390px layouts and reduced motion. This remains a local prototype.

## EF Roadmap Prototype Visual Revision (2026-11-12)

- Revised the local-only six-card prototype to use the existing dark, purple/cyan, glitch and matrix visual language. Cards now display all mapped nodes, individual tracker labels and status distributions inline, with no EIP detail-page navigation. An EF node without a tracker mapping is marked explicitly; the EIP-8077 association remains tentative.
- Status distribution bars count historical tracker labels and are not implementation-completion percentages. The prototype retains this limitation and the EF draft-roadmap attribution.
- Validation: production site build and export checks passed; 371 tracker paths verified; browser checks confirmed six cards, all 51 mapped entries and their counts, no detail links, no console errors, 1440px/390px layout, and reduced-motion behavior. Not published.

## EF Roadmap Six-Card Prototype (2026-11-12)

- Added a local-only `/roadmap-prototype/` view using the EF-maintained November 2 Strawmap as its source. Six interactive fork cards open mapped tracker entries grouped by roadmap node, with source, code/test references and explicit historical-status limits.
- The original drawing places leanVM PQ proofs in I* and ethp2p unification in L*, while the older tracker classifies them differently; the prototype follows the drawing and calls out the discrepancy. Longer-term items remain outside these six cards.
- Validation: static site build and export checks, tracker path verification (371 paths, no missing), local Go build, and browser checks of all six cards, grouped entries and 390px/1440px layouts passed. This is a review prototype, not a deployed page or a new full conformance audit.

## Website Account and Content Refresh (2026-11-12)

- Updated site repository links, clone instructions and issue contacts to `alt-research/eth2030`; aligned legal pages and deployment CNAME with the configured `eth2030.altllm.ai` domain. Updated the Go minimum to match `pkg/go.mod`.
- Refreshed selected EIP tracker gaps and evidence against the current source and records; updated the traceability matrix for the checked items, retained the historical roadmap date, and described the remaining snap-sync and native-observer limitations. Tracker labels and percentages are explicitly historical estimates, not conformance claims.
- Validation: site production build and Go module build passed; exported-page checks passed; tracker verification found zero missing paths across 371 references and no blocking discrepancies; desktop and mobile local browser checks found no horizontal overflow. The Google Analytics measurement ID was not changed because a replacement was not supplied. The live site will update after the change is deployed.

## Track 3 Remaining Acceptance Design (2026-11-12)

- Recorded specific crash boundaries, proof-binding and reuse evidence, independent state/proof readback, memory scaling, real-client interoperability, and mutation gates in [the remaining-work acceptance design](docs/track3-remaining-acceptance.md). These gates are requirements, not completed test results.
- Re-ran the current node, state-sync, beacon-sync and snap package race suites and the Go module build successfully. Docker and Kurtosis executables, the local ETH2030 and Lighthouse images, and an `ethereum/client-go` image are available; their presence is not an interoperability test.

## Track 3 Completed Snapshot Recovery (2026-11-12)

- A CL-selected pivot can reuse a completely downloaded snapshot persisted before a crash or interrupted adoption. The node revalidates the account/storage roots and code, heals trie nodes, and checks the current finalized target before adoption. Offline restart succeeds when the required ancestry and trie nodes are available; corrupted leaves and missing trie nodes without peers are rejected.
- This recovers complete snapshots only. Interrupted range progress is not durable and moving-pivot continuation remains open.
- Downloaded bytecode is also saved independently in hash-verified batches and reused after restart or pivot changes. Corrupt cached code is rejected and fetched again; this does not persist account or storage range progress.

## Track 3 Snapshot Assembly Follow-up (2026-11-12)

- Snapshot validation and trie-node reconstruction now use a streaming trie; node and snapshot writes flush in bounded batches, and the completion marker remains last. Tests compare trie roots with the native MPT and inject failures after partial writes.
- The affected-package race run exposed a separate speculative-state cache race. Cache appends are now serialized and cache readers receive a slice copy; a concurrent regression checks that no results are lost.
- The full downloaded snapshot, sorted hashes, execution overlays, serving snapshot and FileDB index still scale with state size. This reduces peak intermediate allocation but does not complete bounded-memory snap sync.

## Track 3 Snap Peer Failover (2026-11-12)

- One fixed-pivot scheduler can now continue across connected snap peers without throwing away ranges already downloaded in the same attempt. Peers share storage roots only after account-range verification against the same pivot; a root mismatch is rejected even if concurrent verification races with pivot initialization.
- Tests cover proof-verified account roots consumed by a second peer, request-preserving transport failover, cancellation, and the existing two-node snap bootstrap. This is not durable resume or moving-pivot range reuse.
- Auditing a possible bounded-memory trie builder revealed that the existing `StackTrie` retains inserted nodes until finalization and discarded node-writer errors. Its documentation now states the actual memory behavior, and `Commit` propagates write failures; it is not yet a bounded-memory replacement for snap assembly.

## Track 3 Sync Scalability Follow-up (2026-11-12)

- A completed snap pivot can now be adopted after finalized advances only after verifying hash-linked, height-contiguous ancestry against the current CL target and checking that the target did not change during verification. Conflicting branches and malformed ancestry fail closed.
- Storage pagination retries preserve the last verified origin, and completed storage tasks are traversed linearly instead of rescanning the entire task list. Regression tests cover interrupted pages, advancing finality, conflicting finality, concurrent target replacement, and invalid heights.
- This is partial progress, not production snap completion: moving-pivot range reuse, bounded-memory state assembly and serving, bounded-memory database storage, and real-client/mainnet-scale validation remain open.

## Forkchoice Missing-Ancestor Follow-up (2026-11-12)

- Rechecked `0084028` and found that old-canonical ancestry traversal could terminate at a missing block body without finding a common ancestor, then publish a reorg without fully removing old transaction indices. `ApplyForkchoice` now rejects that case before committing state or metadata.
- `TestForkchoiceRejectsMissingOldBranchBody` reproduces a missing intermediate canonical body and verifies the proposed reorg fails without changing the head. The test passed with the race detector; affected-package race and local build checks were run separately.
- This follow-up does not close the production snap limitations recorded below: moving-pivot continuation and bounded-memory assembly remain open.

## Track 3 Corrective Review (2026-11-12)

Reviewed `fc1f70f` through specification, production callers, wire messages, persisted state, restart and adversarial tests. The previous Track 3 completion claim was too broad. Detailed findings and regressions: [Track 3 review](docs/track3-review.md).

- Fixed synchronous FCU validation/publication, candidate staging, validation provenance, missing-body handling, canonical index cleanup, branch-specific `BLOCKHASH`, failed-recovery target preservation and genesis mismatch handling. Backend shutdown now joins actual work instead of unused queue loops; partially constructed nodes release their database locks.
- Fixed FileDB's mixed-transaction recovery window by committing WAL before materializing files and fencing failed writes. Added failure injection that previously recovered a new head with an old finalized checkpoint.
- Fixed per-transaction original storage, reversible account creation/deletion/recreation, adopted committed-storage persistence, mutation-invalidated trie roots, and snap/trie RPC proofs. Added a complete multi-transaction comparison with Geth, including the missing per-transaction `Finalise` in the Geth adapter itself. State/proof comparisons include independent verification and values requiring multi-byte RLP encoding.
- Replaced the native node's private plaintext handshake with standard RLPx, canonical Hello, capability negotiation, reserved message offsets, persistent identity and timestamp-aware fork IDs. Corrected eth/68 pooled-transaction/receipt envelopes and eth/70 bloom-free receipts. Added direct interoperability checks against Geth's RLPx, RLP and ForkID implementations.
- Corrected the design overreach in the native lane: `--native-consensus` now enables a research observer and never replaces external CL authority. Votes use a chain/timing/registry domain, per-source-target quorums, duplicate handling, retryable callback failures and CL re-anchoring. Normal future/unavailable observations do not disconnect EL sync peers.
- Broader race checks exposed and fixed gossip/anonymous-transport counter races and discovery test-counter races. Five temporary behavior mutations were detected by their intended regressions.
- Validation passed on Go 1.26.10: full uncached Go suite; affected-package and complete P2P subtree race suites; integration, blst and goethkzg variants; vet; full build; native executable build with goethkzg. Final changed packages were rechecked after the last edits. Pinned Geth-backed EF vectors passed with 36,264 valid outcomes, zero failures/skips and 20 existing legacy outcomes. Five final-source behavior mutations were each detected by their intended regression.
- **Still open:** production moving-pivot continuation and bounded-memory snap assembly. The original real-snap Track 3 checkbox is reopened. The observer is not an autonomous consensus client. No mainnet-scale qualification or complete native-consensus claim is made.


## Track 3 Node Sync and Native Consensus (2026-11-12)

> Historical implementation record for `fc1f70f`; the corrective review above supersedes its authority and completion claims.

Implemented on `feat/track3-node`, from `master` at `9fde8aa` (PR #50 merged). See [architecture, operation and scope](docs/track3-node.md).

- Replaced the node's no-op snap writers and local-only download path with real correlated `eth` requests and `snap/1` clients. Account/storage range proofs are checked against the pivot and storage roots; bytecodes and reconstructed state are hash-verified before a persistent completion marker permits adoption. Hashed state supports ordinary address/slot reads without preimages. The healer follows authenticated account/storage trie references and fails exhausted work.
- Authenticated Engine forkchoice updates now schedule a cancellable CL-driven execution sync worker. Snap bootstrap uses only the CL's finalized hash, retains the 256-parent `BLOCKHASH` window, stages subsequent execution, and publishes the complete forkchoice atomically. Advancing heads do not repeatedly cancel the finalized snapshot. `eth_syncing` reflects pending/completed work. Full and snap startup restore the persisted chain by validated replay; node shutdown joins workers.
- Added explicit `--native-consensus` configuration for the research 3SF/PQ lane: quick-slot scheduling, real ML-DSA-65 verification bound to configured validator identities and chain/source/target fields, quorum justification and source finalization, opt-in `ncs/1` TCP vote delivery, execution checkpoint updates, and lifecycle/restart integration. External Engine forkchoice is rejected in native mode; the default remains external CL control. Signing keys and automated proposal/signing duties remain outside this runtime.
- End-to-end tests use real TCP nodes and JWT-authenticated Engine HTTP calls, wait for peers after SYNCING, assert the RPC sync-status transition, compare state roots/code/storage and post-pivot execution, and reopen databases. Native-consensus tests deliver real signed votes between nodes, verify finalized execution checkpoints and restart state, and verify shutdown. Negative tests cover incomplete/gapped state, wrong roots, tampered trie nodes, exhausted retries, cancellation, superseded targets, signature identity substitution and cross-chain replay.
- These tests exposed and fixed pre-existing canonical header/body wire encoding errors, missing EIP-7928 BAL header commitment, and loss of the storage-root cache when copying dirty state. The BAL execution-rejection regression now recomputes the advertised block hash after tampering so it still reaches execution validation. Amsterdam hashes now include the BAL; old development-chain data using the previous hash encoding requires a fresh chain or resynchronization.
- Verification: Go 1.26.10 full uncached suite, affected race suites (node/backend, consensus, chain/state/types, eth, snap/ethproto, beacon/healer/statesync, trie and RPC), integration/blst/goethkzg CI variants, vet and local build passed. The native executable also built with the Docker image's `goethkzg` tag. Pinned EF vectors at `81862e4` passed: 36,264 valid outcomes, zero failures/skips, 20 existing legacy outcomes. The EF run used a temporary Go overlay only to point the test fixture lookup at the existing read-only checkout; production sources and reference repositories were unchanged.
- Boundaries: fixed-pivot snapshot assembly/serving materializes state in memory; moving-pivot range reuse, mainnet-scale resource qualification and snap/2 are not implemented. Native consensus is an opt-in research lane, not a mainnet fork activation. No existing worktree or running devnet was modified.

## PR #50 CI Repair and Master Merge (2026-11-12)

- Merged `origin/master` at `c8fd3a1` (PR #49 merge, including fix `4903084`), importing the EIP-7928 fork-boundary regression fix and formatting corrections. Resolved the `STATUS.md` conflict by retaining both Track 2 and PR #49 records.
- Applied `gofmt` to the seven remaining Track 2 files reported by the CI formatting check. No protocol behavior changed.
- Validation: CI-matching Go 1.26.10 passed formatting, `go build -p 4 ./...`, `go vet -p 4 ./...`, `go test -p 4 -count=1 -timeout=15m ./...`, and `go test -p 4 -count=1 -timeout=15m -tags integration ./core ./p2p/...`. External EF fixtures are absent in the isolated worktree, so fixture-dependent tests skip locally; the dedicated remote CI job validates the pinned vectors.

## Hegotá Track 2 Closure (2026-11-12)

Completed the Track 2 items on `feat/track2-hegota` (base: `feat/track1-glamsterdam`, PR #49).

- **EIP-7805 FOCIL**: the six CL P2P validation rules implemented in spec order in `focil/gossip_validator.go` (current-or-previous slot, IL committee root match, two-ILs-per-member with equivocation evidence, BLS signature over the message root with DOMAIN_IL_COMMITTEE 0x0C000000, 16-member committee check, 8 KiB byte cap) with an optional-hook GossipContext for EL-side use. `ILStore` gained gossip-grade semantics: identical rebroadcasts dedup, the second distinct IL records `ILEquivocationEvidence` and withdraws trust, further ILs rejected; accessors GetIL/ILs/Evidence/PruneBefore. `list_validator.go` slot window corrected from current-or-next to current-or-previous per the spec, byte cap added. `inclusion_list` gossip topic added and the node subscribes, funnelling through the validator into the IL store.
- **EIP-8141**: the sponsored-frame-tx integration test surfaced three real gaps, all fixed: `opApprove` set the approve flag before validation (a failed APPROVE still counted for VERIFY frames — flags now set only after all checks pass); the CALLER==tx.sender checks were unreachable for the canonical entry-point-caller flow (entry-point case now passes, sender authorization stays with frame.target == tx.sender); and the gas payer was never actually charged (frame txs skip upfront payment, the payer transfer excluded FrameTxType while refunds routed to the payer — the total gas cost is now collected from the payer on APPROVE(1)/(2) with refunds routed back). Receipts carry payer and per-frame results as derived fields per the receipt format. New scenario contracts FrameApproveExec (scope 0) and FrameApprovePay (scope 1); the two-VERIFY-frame sponsored flow is covered end-to-end (sender pays nothing, sponsor charged effectiveGasPrice*gasUsed).
- **EIP-7495**: StableContainer HashTreeRoot aligned with the spec form mix_in_active_fields(merkleize_progressive([field roots]), active_fields) — progressive merkleization for stable generalized indices, packed-bits mix-in; capacity clamped at 256. Conformance tests pin the spec's Square/Circle stable-gindex example, absent-vs-zero distinction, progressive growth stability, and illegal configurations.
- **EIP-8025**: `GenerateExecutionProofWithBAL` anchors execution proofs to the block's real state-access record (EIP-7928 BAL grouped by access index) and commits to the BAL hash, so a stateless verifier can check the proof against the header commitment without re-executing; tampered access sets rejected.
- **SSZ fork activation (EIP-6404/7807/7916)**: `SSZBlockTime` chain-config fork (fork schedule, ordering, genesis loader); header-only SSZ projection `SSZBlockHash`; `engine_getPayloadV7` returns the SSZ payload encoding plus its hash tree root when active; `engine_newPayloadV7` validates the SSZ binding (decode, hash match, field cross-check) before executing through the Amsterdam path; capabilities advertised. eiptest drives the full V7 surface on an SSZ-active sim: SSZ binding accepted, tampered binding INVALID.
- Verification: full uncached `go test ./...` green (207 packages), `go vet` clean.
- **Devnet node verification** (`full-feature-eth2030`, 3x EL + 3x Lighthouse 8.4.0-rc.0, rebuilt with the Track 2 code): fulu@slot 32, gloas@slot 64 with Amsterdam BAL enforcement, and heze@slot 96 all activated cleanly — blocks progressed past 115 with spamoor transfers landing; every post-Amsterdam block processed with `amsterdamActive:true` and nonempty `balEntries` (43); zero WARN/ERROR across all six nodes. The FOCIL gossip subscription, 8141 payer billing, 7495/8025 code, and SSZ fork plumbing (inert without a scheduled SSZBlockTime) run in production without interop regressions.

## PR #49 CI Repair (2026-11-12)

- Reproduced both failures from CI run `43086199969`: nine Go files failed `gofmt`, and `TestBAL_HashValidationAcrossFork` still expected post-Amsterdam headers without a BAL commitment to pass.
- Applied `gofmt` to the nine reported files. Aligned the integration test with EIP-7928's mandatory header commitment, covering absent/present hashes immediately before, exactly at, and immediately after activation. Invalid cases assert `ErrInvalidBlockAccessList`; production validation remains unchanged.
- Validation: Go 1.26.10 (matching CI) and Go 1.27.3 passed formatting, `go build -p 4 ./...`, `go vet -p 4 ./...`, `go test -p 4 -count=1 -timeout=15m ./...`, and `go test -p 4 -count=1 -timeout=15m -tags integration ./core ./p2p/...`. The isolated worktree has no external EF fixtures, so those fixture-dependent tests skip locally; the dedicated remote CI job checks the pinned vectors.

## Full-Feature Devnet Interop After BAL Enforcement (2026-11-12)

Ran `full-feature-eth2030` kurtosis devnet (3x eth2030:local EL + 3x eth2030cl:local Lighthouse 8.4.0-rc.0, 15s slots) against the BAL-enforced Track 1 branch.

- Fork schedule confirmed from the devnet genesis (`refs/ethereum-package/el_cl_genesis_data/genesis.json`): prague@0, osaka@+480s, **amsterdam@+960s — coinciding with CL gloas_fork_epoch 2 (slot 64)**, fulu at slot 32, heze at slot 96.
- All three forks activated cleanly: blocks progressed through slot 32 (Fulu/PeerDAS), slot 64 (Gloas ePBS + Amsterdam BAL enforcement) and slot 96 (Heze/FOCIL) with no stall; head passed 131 during monitoring with spamoor EOA transfers landing continuously.
- BAL enforcement interoperates with Lighthouse: every block after Amsterdam processed with `amsterdamActive:true` and nonempty `balEntries` (40-43 per block); `engine_newPayloadV5` responses all `error:false`; two EL nodes independently processed block 131 with identical hash/gasUsed/balEntries (43), confirming deterministic BAL derivation under the strict header-hash rule.
- Lighthouse's ePBS bid flow is live on the devnet ("Bid has no blobs, data is available" debug logs) with no CL ERROR/CRIT and no EL WARN/ERROR (beyond filtered noise) on any of the six nodes.
- Finding recorded as follow-up: the JSON-RPC block object does not expose the `blockAccessListHash` header field (the field exists and is enforced; only the eth_getBlockByNumber serializer omits it). Explorers/CL tooling reading the field from RPC will not see it.
- The devnet was left running (enclave `eth2030-devnet`, dora at http://127.0.0.1:32809, assertoor at http://127.0.0.1:32811); stop with `pkg/devnet/kurtosis/scripts/cleanup.sh eth2030-devnet`.

## EIP Integration Test Coverage for Track 1 (2026-11-12)

Added EIP-standard integration tests for the Glamsterdam Track 1 work on `feat/track1-glamsterdam`.

- **EIP-7732**: `TestEIP7732_BuilderPaymentLifecycle` walks the payment state machine over the consensus state with real BLS-signed bids (bid → pending payment → payload envelope → withdrawal queue; attest-backed quorum flush across two epoch boundaries). `TestEIP7732_PTCAttestationBlockFlow` aggregates 257 real PTC member signatures over a 16384-validator state, validates through `process_payload_attestation`, includes the attestation in a produced block body (MAX_PAYLOAD_ATTESTATIONS cap), and confirms fork-choice timeliness. `TestEIP7732_ForkChoiceHeadRespectsPayloadStatus` drives head eligibility between payload-building and empty-building children.
- **Integration test found a real bug**: `GetIndexedPayloadAttestation` dropped the aggregated `Signature` field, so every aggregated payload-attestation verification checked a zero signature. Fixed in `consensus/epbs_state.go`; the new tests pin the behavior.
- **EIP-7975**: `TestEIP7975_ReceiptsPaginationOverRealChain` paginates real Sim-chain receipts through the eth handler over a message pipe (`firstBlockReceiptIndex` slicing, request-order pages, no false incomplete flag); `TestEIP7975_ValidatePartialReceipts` covers the three spec sanity checks against a real block (count match, receipt size vs gasLimit/8, unknown block) plus the block-gas-limit total check; `TestEIP7975_LastBlockIncompleteTruncation` drives serving-side truncation across the 10 MiB budget with `lastBlockIncomplete` set.
- **EIP-8159**: `TestEIP8159_BALEndToEnd` ties the Sim's BAL RLP encoding to the header `block-access-list-hash` commitment and validates served wire bytes against real headers, with 0xc0 for unavailable blocks; `TestEIP8159_RejectsTamperedBAL` enforces the client-side tamper rejection and the empty/non-empty rules for headers with and without the field.
- **EIP-7594**: canonical `data_column_sidecar_<subnet>` topic round-trips for all 128 columns, custody-manager assignments validated against canonical subnets, sidecar gossip REJECT/IGNORE rules (subnet, future slot, finalized slot) on builder-built sidecars with signed headers, and reconstruction from one column per subnet meeting the 64-column threshold.
- **Registry**: EIP-7975 and EIP-8159 added to the coverage registry (TierB/Encoding) and to the IFLOW.md implementation list (Substantial); EIP-7594 upgraded from the TierD exemption to TierC with in-process coverage. The inverted `integration`-tagged BAL test (`TestValidator_RejectsMissingBALHashPostAmsterdam`) plus a block-gas-limit total-size case keep the tagged CI suite aligned with the strict post-Amsterdam rule.
- Verification: full uncached `go test ./...` green (207 packages), `go vet ./...` clean, integration-tagged core suite green, eiptest suite 68+ tests green in ~2.5 s.

## Glamsterdam Track 1: ePBS, eth/70-71, BAL, PeerDAS (2026-11-12)

Implemented the Track 1 closure items on `feat/track1-glamsterdam` (base: `master` at `137006e`, PR #48 merged; merged branch `feat/eiptest-framework` deleted locally).

- **EIP-7732 ePBS (Gloas)** per `refs/consensus-specs/specs/gloas`: new `consensus` epbs layer with the Builder registry container, `ExecutionPayloadBid`, payload envelope, and PTC attestation types; `get_ptc` with balance-weighted selection over concatenated slot committees; `process_execution_payload_bid` (self-build infinity-signature rule, active/funded builder checks, parent/randao/slot commitments, pending-payment recording), `process_payload_attestation`, `process_builder_pending_payments` (quorum flush plus window slide), the consensus side of `process_execution_payload` (payment queueing, payload availability, latest block hash), builder withdrawal plumbing (`get_builder_withdrawals`, builders sweep, apply/cursor updates), attestation weight and slashed-proposal payment clearing. `BeaconBlockBody` now carries `SignedExecutionPayloadBid` and `PayloadAttestations` (Gloas shape, replacing the inline payload header). `EPBSForkChoice` wraps ForkChoiceV2 with PTC vote vectors, payload states, FULL/EMPTY/PENDING classification, the payload-status tiebreaker, and child eligibility filtering. `engine_submitBuilderBidV1` now feeds the node's auction engine through a bid-observer hook.
- **EIP-7975 (eth/70)**: receipts pagination moved to the spec encodings — `GetReceipts` (0x0f) with `firstBlockReceiptIndex`, `Receipts` (0x10) with `lastBlockIncomplete` — replacing the invented 0x11/0x12 messages. Serving paginates under a 10 MiB byte budget. `ValidatePartialReceipts` implements the spec sanity checks: delivered count equals the transaction count, each receipt fits `gasLimit/8` of its transaction, and the total stays within the block gas limit.
- **EIP-8159 (eth/71)**: `GetBlockAccessLists`/`BlockAccessLists` moved to message IDs 0x12/0x13 and carry the canonical EIP-7928 RLP BAL encoding. Unavailable BALs are served as the empty list (0xc0). `ValidateBlockAccessListForHeader` enforces `keccak256(rlp(bal)) == block-access-list-hash` and rejects non-empty BALs for headers without the field.
- **EIP-7928 (BAL)**: post-Amsterdam blocks now MUST include `block-access-list-hash` (the CL-compatibility leniency predated mandatory BAL in V5 payloads and lighthouse >= 8.4). The earlier "probe only logs a BAL mismatch" note was stale — both engine backends already rejected mismatches.
- **EIP-7594 (PeerDAS/Fulu)**: canonical `data_column_sidecar_<subnet_id>` gossip topics with `compute_subnet_for_data_column_sidecar` and topic parsing in `p2p/gossip`; `DataColumnSidecar` now carries `SignedBeaconBlockHeader` per the Fulu container; gossip-rule validators implement the subnet REJECT rule and the slot IGNORE rules (future slot, not newer than finalized); sidecar builders gained header-carrying variants.
- Verification: `go vet ./...` clean, `go build ./...` clean, full uncached `go test ./...` green across 207 packages including the pinned EF fixtures, plus focused suites for `consensus`, `engine`, `epbs/...`, `node`, `eth`, `p2p/...`, `das/...`, `bal`, `eiptest`. New unit tests cover PTC election determinism, bid validation paths, payment quorum/slide, envelope processing, PTC wire-message validation, fork-choice payload status filtering, eth/70 pagination and partial-receipt validation, eth/71 BAL round-trip and header validation, and sidecar subnet/slot rules.

## Native Execution Requests and SSZ Integration (2026-11-11)

- Merged upstream master at `677bbb8` for PR preparation and retained both sets of status and checklist entries.
- Replaced the 32-byte storage fixture path with canonical native execution. EIP-6110 deposits come from matching receipt logs with the exact ABI layout and produce 192-byte records in log order. The EIP-7002 and EIP-7251 system calls execute the canonical queue contracts and return 76-byte withdrawal and 116-byte consolidation records. The native public-network configurations require queue code. `WithRequestContracts` and `ExecutionRequestContractAlloc` provide the deployed runtimes for new private and test chains. The deposit runtime comes from execution-specs at `c708ab6075973b6b928d17ec073dd43a7fe161b6`.
- Block builders, replayers, and validators share post-transaction finalization. Header commitments use the requests derived from execution, and the post-execution BAL records queue writes. Engine payload builders return the canonical request groups, including through the asynchronous builder. Prague and Amsterdam payload validation rejects malformed wire ordering and mismatches between supplied and executed requests.
- Verified two real deposits in one block and all three request types through the native Amsterdam Engine HTTP path, an independent receiver, block replay, and the EIP-7807 projection. A payload with altered requests and a matching advertised header hash was rejected by execution validation. The valid payload produced the expected typed consensus requests root, and the next block emitted no repeated requests.
- The default Go suite with required EF fixtures passed in 130.82 seconds. Affected race checks, the tagged integration suite, formatting, vet and build passed. The native request integration tests passed 20 repetitions. Four isolated mutations that dropped deposits, withdrawals or async build output, or skipped request-commitment validation, each failed its intended regression. Local verification used Go 1.27.3.

## Execution SSZ and Framework Test Effectiveness Review (2026-11-11)

- Reviewed the pushed `b3972f0` branch with the CI-shaped default Go suite and required EF fixtures. After the test updates, 19,424 top-level Go tests passed and seven existing top-level skips were reported. All 68 framework tests passed with zero skips. The pinned EF results were 36,264 passes, zero failures or skips, and 20 known legacy outcomes. The default command took 130.74 seconds using Go 1.27.3.
- Ran 18 isolated behavior mutations using Go overlays. Initially, 11 failed a targeted assertion and seven passed the selected tests. After adding assertions, the same 18 mutations each failed the intended tests. This covered EIP-6404 signature encoding and decoding, full-width chain IDs, compatible-union roots, the empty progressive list, canonical offsets, EIP-7807 payload construction and roots, Engine BAL and hash validation, pool selection, Geth receipt comparison, and the registry gate.
- Added specific rejection assertions for high-S encoding and decoding and for a container with padding between otherwise valid fields. Regenerated the independent EIP-7807 fixture with two transactions and two receipts, then verified that the block projection reproduces its bytes and root with nonempty requests and withdrawals and distinct blob gas limit and usage. The Engine BAL case now verifies the execution validation reason and the unchanged receiver head.
- Ran the targeted suites 20 times. Formatting, vet, build, the integration, blst and goethkzg tagged suites, CI and extended race coverage, and the EEST generation-to-execution smoke test passed. CI race coverage now includes `ssz`, `core/types` and `eiptest`.
- The EIP-7807 projection regression uses valid nonempty reference requests. A separate probe confirmed that the current native EIP-6110 storage queue fixture emits 32-byte deposit request data, which the consensus request root rejects. The test coverage therefore establishes the projection contract for valid requests and the current queue behavior. Native production of canonical nonempty requests and SSZ fork activation remain open.

## Execution SSZ Specification Review (2026-11-11)

- Replaced the experimental transaction byte layout with the ten EIP-6404 compatible-union variants and typed Merkleization. The current EIP-7932 draft uses secp256k1 selector `0x00`. EIP-155 chain IDs and signature parity round-trip independently of the size of the original RLP `v`, preserving the original bytes and signing hash.
- Replaced the block serializer with the published EIP-7807 payload and header summary. It retains EIP-2718 transaction and receipt envelopes and RLP withdrawals, and derives the typed Electra requests root. Both SSZ objects commit to the same root. The codec is available to callers with explicit execution inputs. Consensus activation, Engine API versioning and RPC migration remain separate fork work.
- Corrected the empty EIP-7916 progressive basic list root and tightened canonical container offsets. Reference vectors use the published EIP schemas at `bd93d240b72702450920eeeadc2ce593626bfd8a`, pinned Electra definitions and `ethereum/remerkleable` at `2f0baeef0082d4278acaef7d822deb7009d7db7e`. They cover byte-for-byte encoding and roots for all ten transaction variants, an EIP-155 signed example, a large chain ID, both authorization variants, and an EIP-7807 payload with nonempty transaction, withdrawal and execution request data.
- Verified deterministic vector regeneration, the uncached default Go suite (138.46 seconds), race checks for `ssz`, `core/types` and `eiptest`, the tagged integration suite, formatting, vet and build. Local verification used Go 1.27.3.

## EIP Test Framework Completion (2026-11-11)

- The simulator returns the execution results from canonical block insertion. Regressions cover five-block state continuity, scheduled genesis JSON, exact fork timestamps and stable phase views, receipt and log assertions, state diagnostics, historical calls, and negative validation that preserves the head. Pool tests exercise admission, replacement, eviction, nonce ordering, branch rollback, and finalization.
- The Engine probe uses the node's HTTP endpoint and an independent validating chain. Tests cover consecutive signed payloads, hash and head consistency, inclusion-list attributes, canonical EIP-7928 BAL encoding and rejection, and ePBS bid ordering. The RPC probe checks typed transactions and receipts. The standalone Engine backend also checks V5 payload hash and BAL consistency.
- The differential probe compares state roots and full encoded receipts across a sequence. Native and Geth now agree on the previously divergent contract creation at 75,572 gas. The calldata floor determines the minimum transaction gas limit and clamps final gas usage. Intrinsic gas alone is charged before the EVM runs. The committed calldata, beacon root, blob, Frame and DELEGATECALL regressions assert observable execution outcomes.
- The transaction matrix verifies RLP round trips and recovery for signed envelopes. The pinned EIP-155 vector is checked independently. EIP-7825 is exercised through pool admission, block building and block validation at Osaka. The EIP-7934 scaffold demonstrated a red/green cycle: the original block path reached an intrinsic-gas error, then the Osaka RLP-size check rejected the same candidate.
- The coverage gate detects missing implementation entries and assertion-free test files. Both implementation list paths trigger CI. Verified the scaffold in an isolated directory from generation to compilation, expected gate failure, and passing behavior assertion. It also generated a shell-valid devnet config and verifier pair. Optional legacy test migration remains incremental.
- `make contracts` generates seven embedded runtimes from Solidity and Yul. `make eest` uses the pinned EELS workspace to fill a fresh Amsterdam fixture and immediately executes it through Geth. The smoke fixture tests calldata-derived storage with zero gas price across the two pinned revisions. `make bin/evm` builds the older pinned Geth tool for forks it supports.
- Local verification uses Go 1.27.3. The framework ran 68 top-level tests with no skips in 1.57 seconds of package time. The uncached default suite passed in 130.84 seconds wall time, including the pinned EF fixtures. Build, vet, formatting, integration and goethkzg tagged suites, and race checks for the framework, Engine and node backend passed. Contract generation was repeatable with a clean artifact comparison.
- Compatibility boundaries remain in the underlying protocols. The current SSZ legacy signature field is one byte, so large EIP-155 `v` values produce an explicit encoding error. EIP-8079's draft address and ABI are unspecified; coverage exercises the rollup component. The Amsterdam fixture verifies a shared state-transition case across pinned EELS and Geth revisions. The optional devnet scaffold was verified locally without starting a network.

## EIP Test Framework Acceptance Review (2026-11-11)

Reviewed `feat/eiptest-framework` at `4610eac` against the [framework plan](docs/plans/eiptest-framework-plan.md). The framework runs in the default suite. The v1 acceptance criteria remain open.

- Local Go 1.27.3 verification passed for `go test -count=1 ./eiptest/...` (57 top-level passes and two EIP-8141 skips, 0.795 s package time), `go test -count=1 ./...` (127.22 s wall time, including the EF suite), `go test -race -count=1 ./eiptest/...`, `go test -count=1 -tags integration ./core ./p2p/...`, `go vet ./...`, `go build ./...`, and the CI formatting check. The Go 1.26.8 toolchain download timed out locally. `make contracts` regenerated three contracts with a clean diff. The EIP-7702 selection took 0.711 s of package time and 1.59 s wall time.
- T0.2 state continuity passed an independent five-block test that incremented the same storage slot across three blocks. T0.1/T2.2 still need the genesis JSON round trip and the planned fork transition. With activation at second 18, the probe built its boundary block at second 24 and the retained before-phase view moved to the new head.
- T1.1/T1.3 and T2.1 are incomplete. `BlockResult` lacks block logs, BAL and requests, and the diagnostic dump is unused. `CommitInvalid` returned the same unrelated gas-limit validation error with and without a nonce-invalid transaction. The STF call uses `STATICCALL` and rejects a storage write that succeeds in a simulated RPC `eth_call`. T2.6 wires ordinary insert feedback, but an empty `Commit()` consumes pending pool transactions and the probe has no finalize/reorg wiring.
- T1.2/T2.3 are incomplete. An EIP-155 legacy transaction signed for Alice recovered another address after RLP decode, while the helper passed because it skipped sender comparison on a missing sender cache. Several planned signed builders and the EOF builder are absent. The transaction SSZ helper checks repeated encodings and never decodes them.
- T1.4/T2.8/T2.9 need stronger enforcement and a working walkthrough. A synthetic implemented EIP failed the gate while absent from the registry. Adding its generated TODO-only test and a manual registry row made the gate pass. A missing `IFLOW.md` skipped the gate. The CI path filters omit both `IFLOW.md` and its tracked target `CLAUDE.md`. The documented scaffold command fails from the repository root and the generator prints the registry row for manual entry. T3.1/T3.2 have no demonstrated test-first red/green workflow for a new feature or resulting feedback round.
- T2.4 verifies live block RPC responses, with typed-transaction and receipt assertions still due. T2.5 passes the first injected transaction, but two successive Engine cycles both returned block 1 with `VALID`; the Engine head remained at genesis and the latest-valid hash differed from the advertised payload hash. The probe uses the standalone Engine backend and leaves inclusion-list attributes and node-level HTTP dispatch uncovered. Its BAL mismatch path only logs the mismatch, which conflicts with the EIP-7928 validation requirement. T2.7 compares one transfer's receipt status and gas. It lacks multi-block root parity and storage snapshots. A direct comparison reproduced contract creation at 75,930 native gas and 75,570 Geth gas. The committed contract-parity test logs a fixed Geth value without running Geth.
- The EIP-8141 skip is explained by the existing STOP-only frames omitting payer approval. A separately run funded sender with an `APPROVE` contract and Glamsterdam enabled successfully executed a VERIFY frame and advanced its nonce. The calldata exclusion at 1,500 and 2,000 bytes was explained by the EIP-7976 floor: both were included at their respective minimum limits of 100,500 and 132,500 gas.
- T4.1 still lacks the delegate caller, AA entrypoint, rollup anchor and frame approval scenario contracts. T4.2 has two distinct blockers. `make eest` invokes `go build` for Geth from outside its module. Building the pinned Geth binary inside its module succeeded, but the direct EEST fill exited 5 without running any tests or producing Amsterdam vectors. The default test executes the committed Cancun fixture at a fixed path. T4.3 depends on the first test-first feature. T5.2 emits a devnet stub with no feature validation. T5.3 is already covered by the tagged CI job.

## eiptest Framework Implementation (2026-11-11)

- Implemented the full `eiptest` framework on `feat/eiptest-framework` per the [task plan](docs/plans/eiptest-framework-plan.md): `Sim` chain simulator over the production blockchain/builder/processor path with scheduled-fork configs and genesis-alloc funding; ten-type tx builder DSL; assertion library; seven surface probes (STF, pool, fork, encoding, RPC over the production RPCBackend, Engine payload flow with transaction injection, geth differential); EIP coverage registry gate wired into the default suite; scaffold generator with optional devnet config emission (`tools/eiptest-scaffold`).
- Shipped 45 EIP integration test files (60 test functions) covering the implementation list (tiers A–C; network/CL tiers documented as exempt). Full suite runs in ~0.75s with no build tags; the registry gate fails the default suite when an implemented EIP lacks a test file.
- Contract pipeline (`make contracts`): three solc scenario contracts compiled to runtime bytecode with a Go generator (`tools/gen-contracts`) and an end-to-end deploy/ABI test. Vector pipeline (`make eest`): wired uv + geth `evm` fill; the pinned EEST checkout is mid weld-migration (pytest cannot collect), so consumption is verified with a real GeneralStateTests fixture through `eftest.RunSingleFixture`.
- Production fixes the framework surfaced during development: EIP-7702 delegated-EOA calls now resolve the 0xef0100 designator in `core/vm` (Call + call handlers — previously calls into delegated EOAs executed the designator bytes and reverted); Engine API `ProcessBlockV5` now decodes the lighthouse ≥ 8.4 hex-string BAL form it itself emits; `SetTxPool` now feeds payload building (async builder + the sync FCU path built with a nil pool) instead of only inclusion-list generation.
- Recorded tracked findings for follow-up: ~360-gas contract-creation divergence between native and geth backends; post-Glamsterdan builder drops calldata-carrying txs between 1–2 KB; EIP-8141 frame execution reverts across probed variants while envelope/pool surfaces work; geth differential adapter applies no post-block system contracts (EIP-4788/2935), so the Diff probe compares receipts not roots.
- Corrected the plan's record: the `integration` build tag is not orphaned — CI's `test-variants` matrix runs it and the suite is green locally (core tagged suite verified; `core/eftest` under the tag runs the full EF pass at ~115s).
- Verification: `go vet ./...` clean, `go build ./...` clean, `go test ./eiptest/...` green (58 tests, ~0.75s), `make contracts` idempotent, scaffold generate→compile→test round-trip verified with a throwaway EIP.

## EIP Integration Test Framework Plan (2026-11-11)

- Authored the `eiptest` framework design and phased task plan at [docs/plans/eiptest-framework-plan.md](docs/plans/eiptest-framework-plan.md): an in-process integration test base for future EIP development, with seven surface probes (STF, pool, fork gating, encoding, RPC, engine payload flow, geth differential), an EIP coverage registry gate, and a test scaffold generator.
- Plan targets: single-EIP test iteration under 1 second, full framework suite under 2 minutes, no build tags, default `go test ./...` inclusion. Devnet stays responsible for multi-node surfaces only.
- Baselines recorded on the dev host: `-short` full suite 37.3s, `go vet` 5.7s, single-client devnet boot to block 10 ~50s; the git history review (288 fix commits) shows engine/RPC/txpool surfaces carry the most recurring fixes, which sets the probe priorities.
## PR #47 Merge Resolution

- Merged upstream `master` at `d610385` into the audit follow-up branch and retained both sets of status entries.
- Preserved the refreshed Strawmap and linked its seven milestone cards to the corresponding EIP tracker sections.
- Passed the Go module build and focused BAL, ePBS, FOCIL, core EIP, PQ, and rollup tests. The site build, static export assertions, and EIP tracker file checks also passed.

## Cryptography Regression Fixes (2026-11-10)

- Authenticated native PQ execution and transaction admission with a shared typed signing digest and the existing transaction wire format.
- Matched the pinned Geth Prague activation and EIP-2537 contract layout. Historical native BLS replay uses an explicit chain configuration cutoff.
- Corrected RLPx identity handling, PeerDAS cell-size consistency, empty ECIES messages, and proof batch deadlines. STARK gossip now transmits the complete proof and remains advisory.
- The clean staged snapshot passed 206 Go test packages, builds, vet, tagged tests, Go-only tests, and targeted race checks. The BLS tests passed 914 success and failure vectors. The pinned EF gate reported 36,264 passes and 20 known outcomes with no unexpected failures or skips.

## Ethereum Client Cryptography Follow-Up (2026-11-10)

- Traced Geth's secp256k1, EVM curves, KZG, Engine JWT, and keystore providers and the separate Lighthouse BLS and Verkle paths.
- Switched IPA openings, Poseidon, VDF, VRF, and prime-field share reconstruction to public libraries. Retired unused local signers and the old threshold prototype.
- Added real opening verification for IPA and KZG batches and required registered verifiers for SNARK and STARK statements. The PQ attestation codec carries the full ML-DSA key and signature.
- The clean staged source passed all 206 test-bearing Go packages, build, vet, tagged tests, Go-only tests, and targeted race checks. The pinned EF gate recorded 36,264 passes and 20 known outcomes with zero unexpected failures or skips.
- A Geth provider probe selected Go KZG and C KZG in turn, generated and verified an opening with each, and compared their commitment bytes.
- Recorded the format changes and outstanding STARK, lattice and aggregation work in the [codebase audit](docs/CODEBASE_AUDIT.md#vii-ethereum-client-references-and-follow-up-migration).

## Public Cryptography Library Migration (2026-11-10)

- Moved secp256k1, BN254 precompiles, BLAKE2b compression, ECIES, keystore and RLPx transport to go-ethereum. Engine JWT authentication uses golang-jwt.
- Moved BLS12-381 arithmetic and encodings, BN254 point operations, Bandersnatch operations, EVM transforms, and Groth16 verification to gnark libraries. Blob and PeerDAS operations share the Ethereum KZG context.
- Replaced the project-local PQ signing and KEM implementations with CIRCL and go-fn-dsa. Earlier PQ and ECIES artifacts require regeneration in the public library formats.
- The staged Go snapshot passed builds, vet, the full fixture-backed suite, and tagged tests. The geth-backed EF gate recorded 36,264 passes and 20 known outcomes with zero unexpected failures or skips.
- Recorded the remaining protocol migration work and compatibility boundaries in the [codebase audit](docs/CODEBASE_AUDIT.md#vi-public-cryptography-library-migration).

## Unused Package Cleanup (2026-11-10)

- Removed six packages with no in-repository importers in the [first cleanup batch](docs/CODEBASE_AUDIT.md#v-follow-up-results).
- Verified the clean snapshot with Go builds, vet, module tidy, default and tagged tests, and the website path check. Both executables retained the same production dependency sets in the compared configurations.

## Codebase Audit Follow-Up (2026-11-10)

- Completed the [audit follow-up](docs/CODEBASE_AUDIT.md#v-follow-up-results) on `fix/codebase-audit-follow-up`.
- Native point evaluation verifies the tracked upstream vector using the shared Ethereum KZG ceremony context. Docker and Makefile binary builds use `goethkzg`.
- The geth-backed EF gate executed 36,284 pinned vectors locally. It reported 36,264 passes and 20 exact legacy outcomes, with zero unexpected failures or skips. The missing-fixture probe failed as intended.
- Gnark v0.16.5 passed BN254 AA proving and verification. A compatibility probe also loaded the previous serialized R1CS and keys and verified a prior proof.
- The fresh source snapshot passed formatting, vet, build, and all 213 test-bearing Go packages. The tagged and race suites passed locally. The container started both binary entry points and reported source commit `7e05845`. The site build and EIP path check passed.
- Removed unused internal wrappers and moved ignored artifacts out of the worktree. Tor/Nym live probes require configured services, and the workflow results recorded here are local.

## Codebase Audit Evidence Review (2026-11-10)

- Updated [the codebase audit](docs/CODEBASE_AUDIT.md) with verified production import paths, the native KZG fixture result, tagged test outcomes, EF vector totals and exit behavior, and the gnark advisory.
- Revised the follow-up to match the evidence and added commands for reproducing the results at the audited revisions.
- Verified report links, the website verifier on the documented source snapshot, the Go build, and the website build.

## English Documentation Review (2026-11-10)

- Manually translated `docs/CODEBASE_AUDIT.md` into English, retaining the reported findings and reproduction commands.
- Added the English documentation requirement to `AGENTS.md`.
- Found no remaining Chinese text in the project-owned text files. Extracted text from the five tracked PDFs and visually checked pages without a text layer.
- Verified the Go module with `go build ./...` and the website with `npm run build`.

## Update Plan Phases 2+3: Dependency Upgrades + Lighthouse Sync (2026-11-06)

### Current Status: COMPLETE (with recorded Gloas/Heze devnet dependency)

**Scope:** Execute UPDATE_PLAN.md phases 2 (Go dependency upgrades) and 3 (lighthouse upstream sync), per the "upgrade as far as possible" directive.

**Completed Work:**

Phase 2 — eth2030 Go deps (each group committed separately):
- go-ethereum v1.17.1 -> v1.17.10: GasPool AddGas -> NewGasPool, StateDB.Commit(rules, block), SetTxContext blockAccessIndex, Message gas fields -> *uint256.Int, Holesky removed upstream (mapped to Hoodi 560048, added hoodi network). EF state tests 36,126 passing. (f20cf72)
- gnark v0.14.0 -> v0.16.4, gnark-crypto v0.19.2 -> v0.22.0: 65 packages green (eb20de1)
- x/crypto v0.48.0 -> v0.58.0 (+net/sys/text/sync): 43 packages green (732e844)
- blst v0.3.16 -> v0.3.17: 55 packages green (1934f0b)
- Indirect refresh via go get -u: pebble v2 2.1.7, otel 1.47, grpc 1.85 etc.; full suite 213 packages, 0 cached (da9c207)
- Dockerfile builder -> golang:1.27-alpine (go directive bumped to 1.26.8 by -u) (02a8789)
- CLAUDE.md geth version refs updated (e711263)

Phase 3 — lighthouse:
- Verified all 69 fork commits (feat/fix-native-aa) are superseded by upstream unstable (ForkName::Heze + full IL infra, payload envelopes, PTC). Backup: backup/feat-fix-native-aa-20261106.
- New branch integrate/unstable-202611 = sigp/lighthouse unstable 25b9b754 (8.4.0-rc.0, rust 1.91 image) + IFLOW symlink. (4df5a7bcf)
- cargo update batch reverted: parking_lot trio now requires rustc >= 1.95, breaking the rust:1.91 Docker build; upstream lockfile is the MSRV-compatible set. (2c468be9c)
- cargo check + make lint (clippy) clean; eth2030cl:local rebuilt.

Phase 3e — devnet cross-validation (eth2030cl:local + forks enabled):
- Sanity (forks disabled, upgraded deps): blocks advancing, CL verified, peers=2.
- Fulu transition (PeerDAS): passed on both rounds.
- Engine API compat fixes for lighthouse >= 8.4 (in pkg/engine, with tests):
  - fcuV4/V5 third custodyColumns param accepted; attributes shape sniffing (lighthouse snake_case/hex vs fork-era camelCase); engine_forkchoiceUpdatedV5 routed (1d56d58)
  - fcuV5 advertised in exchangeCapabilities; getBlobsV3 custody bitarray param (d673210)
  - Gloas getPayloadV5/V6 responses carry blockAccessList (0x-hex, RLP bytes; EIP-7928 SSZ wire encoding TODO) + slotNumber from payload attributes; e2e test drives real backend (77287d4, 1991580)

**Recorded dependency (blocks Gloas/Heze devnet acceptance):**
The devnet EL runs eth2030-geth (ethereum-package's geth launcher hardcodes the `geth` command), so the Engine API is go-ethereum's built-in server — go-ethereum v1.17.10 does not emit the Gloas payload shape (blockAccessList/slotNumber), and the CL correctly rejects it (NoViablePayloadBid -> missing field blockAccessList). The pkg/engine fixes serve the self-built binary path (verified by e2e test against the real backend). Full Gloas/Heze devnet bring-up requires either: (a) devnet EL switched to the self-built binary (needs an ethereum-package launcher override or fork — refs are read-only), or (b) Gloas support upstream in go-ethereum. Local manual verification confirmed fcuV5 + lighthouse-shaped attributes parse and build correctly on the self-built EngineAPI.

**Verification:**
- `cd pkg && go build ./... && go test -timeout 240m ./...` (213 packages green incl. EF state tests)
- `cd ../lighthouse && cargo check --workspace && make lint` (clean)
- `cd pkg && go test -run TestGetPayloadV6ResponseCarriesGloasFields ./engine/` (Gloas wire shape)
- Devnet sanity + fulu: `./scripts/run-devnet.sh full-feature-eth2030` -> fulu boundary passed (blocks past slot 32)

## Update Plan Phase 0/1 Baseline (2026-11-06)

### Current Status: COMPLETE

**Scope:** Baseline re-validation after ~6.5 months of maintenance downtime, plus devnet infrastructure pinning per workspace docs/UPDATE_PLAN.md (phases 0 and 1).

**Completed Work:**
- Toolchain: Go 1.25.2 -> 1.27.3 (gvm default), rust/cargo 1.99.1
- spamoor: pushed 3 unpushed commits (4d4b439..4908a05 on feat/support-native-aa)
- eth2030: `go build ./...`, `go vet ./...`, `go test ./...` all green on go1.27.3 (213 packages, 0 cached results, 0 FAIL; includes core/eftest EF state tests, 156s)
- lighthouse: `cargo check --workspace` clean (2m17s, 0 errors, 22 upstream future-incompat warnings)
- Devnet infra pinned: run-devnet.sh / run-feature-tests.sh / run-geth-testnet.sh default KURTOSIS_PACKAGE to local refs/ethereum-package instead of remote master; submodule advanced 234fb546 (2026-02-12) -> 6.1.0 c4a8e15d (2026-04-02) because configs use heze_fork_epoch (upstream renamed eip7805_fork_epoch -> heze on 2026-02-26, eb8c590a)
- Devnet smoke (full-feature, pinned package): blocks advancing, CL "Synced (verified)" with peers=2, EL net_peerCount=2, no WARN/ERROR in EL logs

**Verification:**
- `cd pkg && go build ./... && go vet ./... && go test -timeout 240m ./...`
- `cd ../lighthouse && cargo check --workspace`
- `cd pkg/devnet/kurtosis && ./scripts/cleanup.sh && ./scripts/run-devnet.sh full-feature`
- `cast bn -r http://$(kurtosis port print eth2030-devnet el-1-geth-lighthouse rpc)` -> advancing

## Website Strawmap Refresh (2026-11-05)

### Current Status: LOCAL UPDATE COMPLETE; FURTHER AUDIT AND DEPLOYMENT PENDING

Updated the site against the EF Strawmap drawing dated 2026-11-02, separating proposed Ethereum milestones from selected ETH2030 code evidence. Replaced unsupported completion claims with links to code, tests, specifications, and known limits. Reviewed selected EIPs and implementation paths at repository revision `14497b5`; this is not a full EIP-by-EIP conformance or cross-client interoperability audit. In particular, PQ registration verification uses a stub and ePBS bid validation has a zero-signature testing path.

Verification: `cd site && npm run build && node scripts/check-export.mjs`; desktop and mobile browser previews; `cd pkg && env -u GOROOT go test ./bal ./epbs ./focil ./core/eips ./crypto/pqc ./rollup/...` (run as two commands). Publishing remains pending; the deployment workflow still names `eth2030.com`, while GitHub Pages currently binds `eth2030.altllm.ai`.


## Website Domain Recovery (2026-11-05)

### Current Status: BLOCKED ON REPOSITORY ADMIN

GitHub Pages is configured for GitHub Actions but currently binds `eth2030.altllm.ai`, while the website and deployment workflow specify `eth2030.com`. The latter resolves to GitHub Pages and returns 404. A repository administrator must change the Pages custom domain to `eth2030.com` and verify DNS and HTTPS; this replaces the existing custom-domain binding. The privacy page now identifies GitHub Pages as the hosting provider.

## Full-Feature Devnet Bring-Up (2026-03-25)

### Current Status: IN PROGRESS

**Scope:** Build the local Lighthouse and ETH2030 images as needed, launch the Kurtosis `full-feature` devnet in the `eth2030-devnet` enclave, and verify block production plus basic EL/CL health.

**Planned Verification:**
- `cd ../lighthouse && docker build -t eth2030cl:local .`
- `cd pkg && docker build -t eth2030:local .`
- `cd pkg/devnet/kurtosis && ./scripts/cleanup.sh eth2030-devnet`
- `cd pkg/devnet/kurtosis && ./scripts/run-devnet.sh full-feature`
- `cast bn -r http://$(kurtosis port print eth2030-devnet el-1-geth-lighthouse rpc)`
- `kurtosis service logs eth2030-devnet cl-1-lighthouse-geth 2>&1 | grep -E "verified|ERROR" | tail -20`
- `kurtosis service logs eth2030-devnet el-1-geth-lighthouse 2>&1 | grep -E "WARN|ERROR" | tail -20`

## Inclusion List JSON Refactor (2026-03-21)

### Current Status: COMPLETE

**Scope:** Simplify the local inclusion-list JSON parsing helpers in `pkg/engine/inclusion_list.go` without changing the wire format or validation behavior.

**Completed Work:**
- Extract small helper functions for repeated hex-prefix stripping and byte-slice conversion.
- Reuse the helpers from the custom JSON marshal/unmarshal paths.
- Add focused parsing tests and rerun the relevant engine test suites.

**Verification:**
- `cd pkg && go test ./engine/...`

## Blob Bundle Refactor (2026-03-21)

### Current Status: COMPLETE

**Scope:** Deduplicate the cell-proof expansion logic shared by the V2 and V6 blob-bundle builders in `pkg/engine/backend.go` without changing output formats.

**Completed Work:**
- Extract shared blob sidecar collection and cell-proof expansion logic.
- Reuse the helper from the V2 and V6 blob-bundle builders.
- Verify the relevant engine test suites after the refactor.

**Verification:**
- `cd pkg && go test ./engine/...`

## Payload Retrieval Refactor (2026-03-21)

### Current Status: COMPLETE

**Scope:** Deduplicate the repeated payload retrieval paths in `pkg/engine/backend.go` without changing API-specific response contents or async eviction behavior.

**Completed Work:**
- Extract shared helpers for async payload waiting and mutex-protected payload lookup.
- Reuse the helpers from `GetPayloadByID`, `GetPayloadV4ByID`, `GetPayloadV5`, and `GetPayloadV6ByID`.
- Verify the relevant engine test suites after the refactor.

**Verification:**
- `cd pkg && go test ./engine/...`

## Inclusion List Selection Refactor (2026-03-21)

### Current Status: COMPLETE

**Scope:** Deduplicate inclusion-list transaction selection across engine backends and make oversized pending transactions skippable instead of aborting the whole selection pass.

**Completed Work:**
- Extract shared inclusion-list transaction selection into a helper in `pkg/core/eips/inclusion_list.go`.
- Use the helper from both `pkg/engine/backend.go` and `pkg/node/backend/engine_backend.go`.
- Add a focused test for oversized-transaction skipping.
- Verify the relevant Go test suites after the refactor.

**Verification:**
- `cd pkg && go test ./core/eips ./engine/... ./node/backend/...`

## Engine Review Follow-Up (2026-03-21)

### Current Status: COMPLETE

**Scope:** Recheck branch review findings, confirm whether `engine_getPayloadV5` JSON encoding is already fixed, and address the remaining safe bugs in inclusion-list handling.

**Completed Work:**
- Verified that the `engine_getPayloadV5` response encoding issue was already fixed on this branch by `fdd0820`.
- Reject oversized hex input in `flexibleUint64` instead of truncating it.
- Make the node backend return an explicit error for `ProcessInclusionList` until real storage/wiring exists.
- Add tests for the inclusion-list fixes and rerun the relevant engine/node test suites.

**Verification:**
- `cd pkg && go test ./engine/... ./node/backend/...`

## Engine Lock Wrapper Refactor (2026-03-19)

### Current Status: COMPLETE

**Scope:** Refactor `pkg/engine/backend.go` so accesses guarded by `stateMu`, `blocksMu`, `payloadMu`, and `ilMu` go through helper functions instead of open-coded lock/unlock pairs.

**Completed Work:**
- Add backend accessor methods that encapsulate `stateMu`, `blocksMu`, `payloadMu`, and `ilMu` reads/writes.
- Migrate existing call sites in `pkg/engine/backend.go` to the accessors without changing lock ordering or behavior.
- Follow up on the accessor refactor by removing remaining unsynchronized forkchoice field reads in `ForkchoiceUpdated`.
- Extend the accessor refactor into `pkg/engine/backend_bodies.go` to remove direct mutex usage there as well.
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

## Scoped devnet controls

- The controller, verifiers and fixtures are packaged directly against the
  Track 4 master baseline. The shell syntax checks, six Python regressions,
  profile and evidence checks, and independent Track 4 HTTP test passed.
- The live fixture records identify the revisions used in each run. The
  Gloas and Heze cases use the paired EL and CL revisions recorded there.

## Pinned Frame codec review

- The seven-field Frame codec and split-gas receipt encodings are packaged
  against the current Bogotá master. Full Go tests, targeted codec and RPC
  race tests, and the build passed.
- The native transaction pool and executor still reject the pinned format.
  Activation, execution, blob sidecars and historical import remain separate
  acceptance gates. The Heze memory limits are reviewed independently.
