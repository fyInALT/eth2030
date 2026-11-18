# Lean4 Verification Check-list

## Canonical execution system transitions (2026-11-19)

- [x] Read complete pinned EIPs 2935/4788 and the selected 8037 system-call margin; compare canonical SDK assets.
- [x] Reproduce wrong history address/window and missing/custom-runtime behavior through actual block processing.
- [x] Share pre-block EVM system-call execution between builder/importer; preserve caller exclusion, ordering, isolation, actual BAL changes and no block/user fee debit.
- [x] Supply explicit canonical assets before owned genesis roots without mutating supplied network state or fork-time injection.
- [x] Qualify independent runtime/storage/root consumers, wrap/missing-slot/query bounds and native Node build/import/restart; preserve 8282/6110/7002/7251 coverage.
- [x] Run full/tagged/affected race/vet/build and sensitive mutations; prepare a separate feature PR.

- [x] Publish the separate review as PR 84; composed qualification is recorded independently.

## eth version 72 sparse blob relay (2026-11-19)

- [x] Pin full EIP-8070/7594 and devp2p schemas plus Geth merged flat request-id fix.
- [x] Add actual negotiated version 72 announcements, elided pooled metadata and correlated cell exchange.
- [x] Stage signed sparse metadata/cells under resource bounds; authenticate every cell and reconstruct before ordinary pool promotion.
- [x] Wire provider/sampler selection, two-provider threshold, custody plus random extra, custody expansion and eager building.
- [x] Consume actual FCU custody updates and implement getBlobsV4 cell/proof readback.
- [x] Verify encrypted SDK/Node consumers, invalid/unrequested data, recovery and independent reconstruction; build/test/mutate and prepare a separate PR.

- [x] Publish the independent feature review as PR 82; composed qualification is tracked separately.

## snap version 2 unchanged storage account proof binding (2026-11-19)

- [x] Reproduce a moving pivot with partial storage whose account body/storage root is unchanged.
- [x] Authenticate that account against the fresh pivot before using its storage root.
- [x] Reject a damaged proof for that unchanged account without checkpoint/adoption.
- [x] Qualify all encrypted Node scenarios, affected race/vet/build and update PR 81.

## snap/2 BAL state catch-up (2026-11-18)

- [x] Pin EIP-8189/7928 and trace real snap peer, state scheduler/writer and CL target adoption.
- [x] Negotiate snap/2 separately, retain unchanged range/code packets and reject removed trie-node messages.
- [x] Serve and authenticate ordered raw BAL responses with missing entries, budgets and cancellation.
- [x] Rebase downloaded state and outstanding ranges through verified BALs, including moving pivots and reorg handling.
- [x] Qualify actual Node sync/restart and independent wire/root consumers; run builds/tests and publish the separate PR.

- [x] Read EIP-7975/8159 and current devp2p status/message schemas; inspect the actual node registration.
- [x] Add independent canonical BAL-wire, versioned handshake and correlated tamper tests.
- [x] Register negotiated versions, implement status/range and version-aware receipt serving.
- [x] Correlate and validate received BALs before restoring verified historical data.
- [x] Verify affected network/chain/node/consumer paths and race/build; prepare the separate feature review.


## EIP-8282 builder execution requests (2026-11-18)

- [x] Read the complete pinned EIP, canonical bytecode assets and selected CL processing.
- [x] Add exact runtime/constructor assets and request types/SSZ encoding with provenance.
- [x] Wire mandatory Amsterdam system queues, rollback and commitment validation.
- [x] Supply explicit pre-fork deployment/genesis fixtures without fork-time code injection.
- [x] Qualify fee/queue/endian/cap/failure and actual Engine/CL consumer paths.
- [x] Run affected builds/tests and publish a dedicated feature PR.

## Complete ordinary blob sidecar validation (2026-11-18)

- [x] Read pinned EIP-4844/7594 networking and proof requirements.
- [x] Reproduce invalid ordinary-cell-proof admission through real node HTTP.
- [x] Validate complete sidecar binding/proofs and preserve Frame regression behavior.
- [x] Require full data for pool-based builders and return genuine legacy blob proofs.
- [x] Qualify actual wire/admission/Engine consumers, affected checks/build and regression mutations.
- [x] Publish the isolated blob-validation PR (PR 76).

## Fork-aware blob capacity, fees and configuration (2026-11-18)

- [x] Base on PR 79 configuration authority and compose PR 76 full-data/proof consumers; retain the 21-blob race timeout as evidence until the current wrapper is qualified.

- [x] Read EIPs 7840/7892/7691/7918 and trace actual builder/import/pool/VM/RPC consumers.
- [x] Reproduce hardcoded capacity/fee and ignored configuration defects against independent reference inputs.
- [x] Wire one selected schedule through canonical block admission, fee calculation, payloads and RPC.
- [x] Qualify actual signed blob traffic, fork transitions, invalid blocks and configuration/restart boundaries.
- [x] Run affected/full builds/tests and publish the isolated feature review.

## Persisted fork configuration and restart authority (2026-11-18)

- [x] Trace actual Node startup before history replay and reuse the existing persisted config key.
- [x] Detect incompatible active schedules and immutable parameter changes before state/metadata adoption.
- [x] Qualify future updates, either-side activation, missing/corrupt metadata and failed persistence.
- [x] Verify actual node restart preserves heads/state/configuration on rejection and does not mutate network globals.
- [x] Run affected/full tests, build and race/vet; publish a dedicated PR.

## Canonical fork schedule and Engine version boundaries (2026-11-18)

- [x] Read the current Engine Amsterdam/Osaka/Prague/Cancun rules and trace actual startup/HTTP consumers.
- [x] Reject incoherent canonical fork schedules before genesis/state adoption; preserve explicit private extensions.
- [x] Enforce canonical Engine version/required-field boundaries with exact activation and presence cases.
- [x] Qualify actual HTTP failures and valid forkchoice adoption independently from attribute rejection.
- [x] Run affected/full tests, build, race/vet and publish the isolated review.

## Mandatory builder-request feature composition (2026-11-18)

- [x] Allocate canonical contracts in blob/factory/slot owned fixtures before root construction.
- [x] Qualify all combined Go tests, tagged Core/EIP, vet and local build.

## Complete ordinary blob sidecar validation (2026-11-18)

- [x] Read pinned EIP-4844/7594 networking and proof requirements.
- [x] Reproduce invalid ordinary-cell-proof admission through real node HTTP.
- [x] Validate complete sidecar binding/proofs and preserve Frame regression behavior.
- [x] Require full data for pool-based builders and return genuine legacy blob proofs.
- [x] Qualify actual wire/admission/Engine consumers, affected checks/build and regression mutations.
- [ ] Publish the isolated blob-validation PR.

## Glamsterdam independent current-text gas corpus (2026-11-18)

- [x] Freeze signed transactions, complete pre-state, fork/config and reference-client version.
- [x] Execute identical inputs through native transaction-level execution and the unmodified Geth Amsterdam transition.
- [x] Record invalid versus included failure, gas dimensions, logs and exact final account/state results.
- [x] Verify pre-Amsterdam controls, archive candidate mismatches and build locally.
- [x] Publish the separate gas-corpus review (PR 75).

## Glamsterdam header and slot independent qualification (2026-11-18)

- [x] Compare native canonical header RLP/hash against go-ethereum v1.17.10.
- [x] Drive nondefault slots through HTTP FCU/build/import and observe SLOTNUM state and block RPC.
- [x] Match historical Engine BAL bytes to the independent decoded header commitment.
- [ ] Run affected tests and local build; submit the isolated qualification PR.

- [x] Revalidate merged canonical body wire/storage and read the current EIP-7928 retention requirement.
- [x] Reproduce the obsolete availability window through the real Engine body consumer.
- [x] Use the 14,299-epoch slot window, preserve conservative fallback and body availability.
- [x] Verify affected consumers, storage regressions, build and prepare a separate PR.


## Empty committed historical BAL recovery (2026-11-18)

- [x] Reproduce an empty BAL cache miss without hiding its execution body.
- [x] Regenerate only the header-committed empty list inside retention, for hash and range APIs.
- [x] Verify affected suites and local build, then update PR 66.

## Glamsterdam EIP-8024 current byte encoding (2026-11-18)

- [x] Pin the current EIP and reproduce published execution vectors against the native VM.
- [x] Correct legacy-bytecode immediate decoding and its invalid ranges.
- [x] Verify every valid depth/pair, stack boundaries, code-end immediates, jumps and native transactions.
- [ ] Run affected tests, local build and vet; publish an isolated implementation PR.

## Glamsterdam EIP-8246 transaction lifecycle (2026-11-18)

- [x] Pin EIP-8246 and read EIP-6780/161; reproduce missing native destruction cleanup.
- [x] Journal same-transaction creations with nested rollback and preserve balances through finalization.
- [x] Verify native CREATE/CREATE2, pre-funded targets, late credits, old contracts and reverted creation.
- [x] Qualify BAL final-state changes and persistence; run affected tests and local build.
- [x] Publish the isolated implementation branch and update program evidence (PR 72).

- [x] Read both local legacy and pinned current EIP-7708, including fee/withdrawal exclusions and 8246 dependency.
- [x] Add failing native fork-boundary, fees/self/zero-value, creation ordering and rollback regressions.
- [x] Correct consensus emission and canonical Amsterdam activation without a new toggle.
- [x] Verify affected consumers/roots and build; prepare a separate PR with the 8246 lifecycle gate explicit.


- [x] Read EIP-7975/8159 and current devp2p status/message schemas; inspect the actual node registration.
- [x] Add independent canonical BAL-wire, versioned handshake and correlated tamper tests.
- [x] Register negotiated versions, implement status/range and version-aware receipt serving.
- [x] Correlate and validate received BALs before restoring verified historical data.
- [x] Verify affected network/chain/node/consumer paths and race/build; prepare the separate feature review.


- [x] Inspect the production HTTP/backend route and merged execution-apis #950 semantics.
- [x] Add failing real-backend HTTP storage-discovery and non-mutation regressions.
- [x] Replace the empty placeholder with bounded discovery/convergence and gas readback.
- [x] Verify returned access items, reverted touches, creation, historical state and affected packages/build.
- [x] Prepare this major RPC feature for its separate PR.


- [x] Read old local and pinned current EIP-7997, including deployment and no-fork-check rules.
- [x] Add canonical address/runtime/nonce and actual CREATE2/value tests plus a no-fork-injection regression.
- [x] Replace the obsolete factory fixture and remove execution/builder fork-time state mutation.
- [x] Qualify affected paths and prepare the separate feature review.


- [x] Revalidate merged canonical body wire/storage and read the current EIP-7928 retention requirement.
- [x] Reproduce the obsolete availability window through the real Engine body consumer.
- [x] Use the 14,299-epoch slot window, preserve conservative fallback and body availability.
- [x] Verify affected consumers, storage regressions, build and prepare a separate PR.


- [x] Read local draft and pinned current EIP-7954; record the 64/128 KiB revision difference.
- [x] Add failing CREATE/CREATE2 and transaction admission boundary tests.
- [x] Correct fork-gated limits and reject oversized creation transactions before mutation.
- [x] Verify affected packages, build and prepare the separate feature review.


- [x] Reproduce nonzero-log consensus receipt roundtrip failure.
- [x] Recompute omitted wire blooms from decoded logs, not from an empty receipt bloom.
- [x] Verify codecs/consumers/build and prepare the scoped fix review.


- [x] Isolate the current-mainline program worktree and preserve the original dirty checkout.
- [ ] Revalidate the 33-task plan against origin/master and freeze specification/decision inputs.
- [ ] Publish the baseline/acceptance program in its own review branch.
- [ ] Deliver each major feature in a separate scoped branch and PR, retaining test evidence.
- [ ] Close native, consumer, independent-client, recovery and release gates with observed evidence.


## Glamsterdam combined feature qualification (2026-11-18)

- [x] Merge only the isolated Glamsterdam review branches into a local integration branch.
- [x] Resolve owned overlapping changes and retain every verification boundary.
- [x] Run the combined Go suite, tagged integration tests, build and affected race/vet checks.
- [x] Record exact branch heads, review links and remaining program gates without claiming release completion.

## Merged Frame roadmap and homepage progress (2026-11-18)

- [x] Reconcile PR 61's merged code and runtime evidence with roadmap/tracker claims.
- [x] Share the qualified development snapshot between the homepage and roadmap.
- [x] Refresh Frame, FOCIL and historical BAL references; retain independent-client and service gates.
- [x] Replace obsolete active plans and progress totals with current evidence and next acceptance work.
- [x] Verify the static export, progress consistency, source paths, local build and responsive rendering.
- [x] Commit and publish only the scoped changes for review.

## Frame paymaster and history follow-up (2026-11-18)

- [x] Freeze a reproducible ETH2030 paymaster runtime with an acyclic protocol-signature witness convention and explicit upstream-reference differences.
- [x] Qualify exact-code recognition, prefix execution, aggregate exposure, delayed-withdrawal reservations and head/reorg revalidation.
- [x] Test actual authorization, timelock enforcement, storage changes and value transfers, including negative and mutation cases.
- [x] Repair historical Engine body DATA encoding and validated BAL persistence, including both wire versions and restart readback.
- [x] Exercise coded sponsorship and fresh CL Frame history with bounded live resources, then remove all owned runtime resources.
- [x] Retain a root-checked replay of the real Heze state using the released Go client and a minimal profile-specific compatibility type.
- [x] Reconcile service-consumer and independent-client gates against current sources; keep unsupported boundaries explicit.

## Canonical Frame execution

- [x] Implement the pinned execution, state gas, approval and atomic rollback paths.
- [x] Wire activation, generic public admission, payload import and receipt persistence.
- [x] Pin the canonical-paymaster runtime and witness convention and qualify withdrawal reservations.
- [x] Verify signatures, sponsorship, native blob order and EL restart.
- [x] Qualify fresh CL Frame backfill against the paired native EL.
- [ ] Qualify independent-client Frame history.
- [ ] Qualify updated services and independent-client and state-sync gates.

## Next-stage promotion (2026-11-17)

See the [next-stage work plan](docs/next-stage-work-plan.md) for verified branch
state, review scope and independent-client gates.

- [x] Prepare and merge the native Bogotá, scoped devnet controls, Lighthouse and documentation reviews.
- [x] Integrate Lighthouse upstream `3777cfd7` and preserve the local FCU verdict and list recovery regressions.
- [x] Rebuild and qualify the patched EL/CL pair on finalized sponsored Frame traffic.
- [x] Finish the pinned Frame transition and run its live receipt and history gates.
- [ ] Run the Track 3 independent late-join control and continue authenticated moving-pivot recovery.

## Consensus Client Follow-up (2026-11-16)

See the [Lighthouse and Hegotá implementation plan](docs/lighthouse-hegota-implementation-plan.md)
for work ownership, dependencies and acceptance criteria. The
[consensus client roadmap](docs/cl-iteration-roadmap.md) covers later milestones.
The [CL gap audit](docs/lighthouse-cl-gap-audit.md) records the audited revisions,
implementation gaps and acceptance criteria.

- [x] Record the published Lighthouse development baseline, initial Frame and FOCIL draft references, and cross-repository implementation and acceptance plan.
- [x] Inventory the explorer, workload, assertion and monitoring services with fork revisions, profile coverage and historical run evidence.
- [x] Compare the maintained forks with their upstream release, default and devnet branches in the [upstream adoption audit](docs/devnet-fork-upstream-audit.md).
- [x] Assign duplicate mainline features and experimental ETH2030 features to their implementation owners in the [feature-level dispositions](docs/devnet-fork-upstream-audit.md#feature-level-disposition).
- [x] Record the [implementation PR sequence](docs/devnet-upstream-implementation-plan.md#pr-sequence), control pins, verification gates and per-profile promotion.
- [x] Freeze the capped Fulu control and verify finality, EL restart and enclave removal in [PR 1](docs/devnet-upstream-implementation-plan.md#1-freeze-the-control-and-make-failures-observable).
- [x] Verify Gloas-at-genesis self-build and ordinary Heze compulsory inclusion on the merged clients, then remove both enclaves.
- [ ] Run the scheduled Gloas, external builder, PTC and blob cases and the independent-client gates.
- [ ] Qualify upstream Spamoor for ordinary traffic and verify receipt failover and the pinned Frame vectors before retiring the workload fork.
- [ ] Replay the Heze Dora retention and epoch-data scenarios with an upstream Dora and Go client pair on a compatible CL schema.
- [ ] Qualify the newer ethereum-package pin with generated genesis checks and a combined Heze/Bogotá activation run.
- [x] Verify the historical CL and EL branches against the live code paths and assign all 68 unmatched Lighthouse commits to the [legacy branch integration plan](docs/lighthouse-legacy-branch-merge-plan.md).
- [x] Audit the merged EL, Lighthouse branch delta, and upstream fork scope.
- [x] Merge the current Lighthouse unstable revision in an isolated integration worktree and review the source delta.
- [x] Verify official Lighthouse v8.3.3 against two ETH2030 ELs on the private Fulu profile.
- [x] Verify the pinned upstream Lighthouse unstable image against two ETH2030 ELs on the private Fulu profile.
- [x] Audit the March Lighthouse patches against current source, specifications, and Dora consumers.
- [x] Verify a self-built Gloas envelope and its matching execution block on both node pairs after a live fork transition.
- [x] Publish the Lighthouse `eth2030` branch and test the Heze self-build inclusion-list snapshot.
- [x] Integrate the upstream update at `8104a2467` and port Heze proposer-side gossip checks and local Frame hash verification.
- [x] Verify validator-driven Heze inclusion-list production against the native EL's Bogotá list response in the two-pair devnet.
- [x] Implement native Bogotá V6 validation and persist the execution-valid inclusion-unsatisfied decision in Lighthouse fork choice.
- [x] Implement Heze inclusion-list peer recovery and test request/response framing, authenticated previous-slot import, and stream termination.
- [x] Prove ordinary transaction inclusion on two unpeered ELs and verify the same-parent Engine build with and without list delivery.
- [x] Commit EIP-7843 slot numbers to Amsterdam execution headers and match Geth genesis hashing, native payload import, and restart readback.
- [x] Merge upstream Lighthouse `7a9d14115`, run its Gloas database-invariants test, and verify the published fork ref.
- [x] Verify finalized Fulu Track 4 history and the same PQ transaction receipt after restarting one live EL.
- [x] Add the repeatable Heze devnet profile and verifier for a finalized ordinary transaction delivered through FOCIL.
- [x] Reproduce a finalized Fulu checkpoint and the Gloas transition with a published Lighthouse source, the current receiver fixture, and a pinned ETH2030 image.
- [ ] Freeze cross-client Frame and Bogotá wire fixtures.
- [x] Align Amsterdam-at-genesis execution header hashing and confirm block production with Gloas at genesis.
- [ ] Pin a reproducible Lighthouse and ETH2030 reference pair and verify the Gloas/Amsterdam builder, BAL and blob scenarios.
- [ ] Align the EL Frame implementation with the pinned draft and verify Frame payload import, blob commitments and historical backfill in the CL.
- [x] Verify the V5 to V6 fork boundary on both ELs, the same-payload inclusion verdicts, and finalized compulsory inclusion on the published Lighthouse runtime with isolated mempools.
- [x] Force live inclusion-list gossip loss and recover the original signed list through Req/Resp following an embedded self-build bid.
- [x] Implement and test pending Bogotá inclusion-list retention and deferred FCU validation through the native EL HTTP API.
- [ ] Verify deferred inclusion satisfaction through forkchoice against an independent execution client.
- [x] Finalize a sponsored Frame transaction delivered through FOCIL, reconcile payer balances and receipts on both ELs, and verify EL restart.
- [x] Verify the same sponsored Frame through fresh CL late join with a synchronized EL.
- [ ] Verify the same Frame through independent EL late join.
- [x] Verify Dora inclusion-list positions and historical display after Lighthouse pruning and Dora restart on a memory-bounded Heze devnet.
- [ ] Update Dora epoch statistics to decode current Gloas and Heze BeaconState SSZ.
- [ ] Verify full and snap late join and restart against independent source clients.
- [ ] Exercise scheduled Quick Slots and optional execution proofs on separate profiles with independent evidence.
- [ ] Run decoupled-consensus differential tests and PQ validator interop against a pinned Lean specification.

## Heze CL follow-up

- [x] Merge upstream Lighthouse `ca9dff457` and verify both database restart regressions and the local release build.
- [x] Bind deferred Bogotá Engine verdicts to the executed beacon ancestor, including an empty head and fork-choice reload.
- [x] Recover signed lists advertised by embedded self-build bids with gossip delivery deliberately dropped in a capped devnet.
- [x] Read the imported Heze block and its prior finalized checkpoint after restarting the receiver.
- [x] Check Frame blob commitment ordering through the CL historical payload verification path.
- [x] Run historical backfill with Frame transactions executed by a compatible EL, including sponsored Frame settlement.
- [x] Integrate upstream Lighthouse `1bcdcbfa0` on the staging branch and pass the Gloas light-client fork-boundary regression.
- [ ] Pin Heze SSZ and proof vectors and implement Heze light-client headers, bootstrap and updates.
- [ ] Verify VC restart, remote signing and redundant BN publication without signing conflicting lists for a duty.
- [ ] Verify Req/Resp recovery through alternate peers, partial responses and competing dependent roots.

## Pinned EIP-8141 Migration

- [x] Match the seven-field transaction and signing encodings and validate protocol signatures against independent fixtures.
- [x] Match canonical receipt encoding, split gas readback after a FileDB reopen, and the fee-budget arithmetic.
- [x] Implement frame execution, state-gas ownership and refills, atomic rollback, and escrow and settlement on live state.
- [x] Wire Bogotá activation, generic public-mempool policy, blob sidecars, Engine validation, and historical replay.
- [x] Freeze the canonical-paymaster exception and delayed-withdrawal reservation rules against a deployable runtime.
- [x] Run finalized sponsored FOCIL and EL restart scenarios with reconciled receipts and balances.
- [x] Run native blob-bearing build/import, reorg and disk replay with real KZG data.
- [x] Run live blob-bearing Frame history and receiver restart scenarios.
- [ ] Run independent EL late-join scenarios.

## Roadmap Homepage Integration (2026-11-13)

- [x] Promote the approved prototype to a detailed roadmap page and share its progress model with the homepage.
- [x] Replace the homepage roadmap with six compact cards showing progress and only non-complete tracker entries.
- [x] Verify homepage filtering, matching percentages, links, responsive layout, and the site build.
- [x] Update status and commit the integration without publishing.

## Native Bogotá deferred validation

- [x] Retain the exact V6 list and off-canonical block before returning ACCEPTED.
- [x] Complete staged validation on FCU and keep the verdict across FileDB reopen.
- [x] Check missing ancestry, hash validation, list replacement and unresolved synced views through Engine HTTP.
- [ ] Verify the native path with Lighthouse and a compatible independent EL.

## Roadmap Prototype Card Alignment (2026-11-13)

- [x] Keep the six roadmap phases in chronological reading order and equalize card dimensions without clipping content.
- [x] Verify all card sizes, order, visible content, and responsive layouts in the browser.
- [x] Build, record the result, and commit the local prototype without publishing.

## Roadmap Prototype Density and Percentage (2026-11-13)

- [x] Make all six cards denser and emphasize tracker progress with clearly defined percentages and animation.
- [x] Check percentage math against displayed entries, zero-complete cases, responsive layouts and reduced motion.
- [x] Build, update status, and commit the local prototype without publishing.

## Roadmap Prototype Visual Revision (2026-11-12)

- [x] Match the existing site's dark visual language and add accessible animation.
- [x] Put all mapped roadmap nodes and tracker progress directly inside six cards without detail-page navigation.
- [x] Verify every mapped entry, status counts, animations, responsive layout, and the production build.
- [x] Record the prototype revision and commit without publishing.

## EF Roadmap Six-Card Prototype (2026-11-12)

- [x] Check the current EF Strawmap and map its six fork columns to the existing repository tracker.
- [x] Build a responsive, interactive prototype with source attribution and clear evidence limitations.
- [x] Build locally, inspect desktop and mobile, and verify card navigation and source links.
- [x] Record the prototype scope and commit the prototype without publishing it.

## Website Account and Content Refresh (2026-11-12)

- [x] Compare the deployed site and GitHub Pages configuration with current repository ownership and source.
- [x] Update repository links, clone instructions, canonical domain, dependency version, and verified outdated feature claims.
- [x] Distinguish historical roadmap/tracker data from current verification and document remaining sync limitations.
- [x] Build the site and local Go module, verify exported links and evidence paths, update status, and commit task-owned changes.

## Track 3 Completion Follow-Up (2026-11-12)

- [x] Trace bounded-memory snapshot download, execution and serving against the snap protocol and database implementations; the active FileDB also retains all values in memory.
- [x] Preserve storage pagination on transient failure and avoid quadratic scans of completed storage tasks.
- [x] Verify finalized ancestry before reusing a completed snapshot after a CL finality advance.
- [x] Continue a fixed-pivot download across peer failures, sharing only account roots authenticated against that pivot.
- [x] Audit the candidate stack-trie builder: propagate node-writer failures and correct its bounded-memory claim.
- [x] Stream snapshot validation and trie-node generation; write trie nodes and snapshot leaves in bounded batches with failure-before-publication tests.
- [x] Revalidate and recover a completely downloaded, not-yet-adopted snapshot after restart, including offline recovery and corruption/missing-node rejection.
- [x] Persist downloaded bytecode by verified content hash and reuse it across attempts and restarts without publishing an incomplete state.
- [x] Specify observable restart, moving-pivot, resource, independent-client, and mutation-test acceptance gates in `docs/track3-remaining-acceptance.md`.
- [ ] Implement authenticated moving-pivot continuation and safe restart behavior with regression coverage.
- [ ] Remove remaining sync scalability bottlenecks without weakening ancestry validation.
- [ ] Validate real-client interoperability and record the actual scope and resource limits.
- [ ] Run affected tests, race checks and builds; update status and commit task-owned changes.

## Track 3 Adversarial Review (2026-11-12)

- [x] Re-read specification and design contracts; audit every Track 3 production change and its callers.
- [x] Reproduce and fix forkchoice authority, ancestry, durable publication, restart and reorg defects.
- [x] Reproduce and fix native/snap state execution, rollback, storage, and proof-validation defects.
- [x] Audit protocol encodings, peer cancellation, pagination, lifecycle shutdown and native consensus invariants.
- [x] Add behavioral regressions and targeted mutation checks; run affected race, full, integration, tag, vet and build checks.
- [x] Correct completion claims and document remaining scope accurately; commit and push scoped fixes.

## Track 3 Implementation (2026-11-12)

Base: `origin/master` at `9fde8aa` (PR #50 merged), branch `feat/track3-node`.

- [x] Trace current snap serving, download, proof verification, persistence, pivot adoption, and restart paths against the pinned specifications.
- [x] Replace node snap-sync no-op writers with verified persistent state and trie healing; test sync from an empty database and execution after pivot adoption.
- [x] Connect unknown Engine forkchoice heads to a cancellable CL-driven beacon sync loop with peer download, validated ancestry, retries, and shutdown coverage.
- [x] Connect 3SF, quick-slot scheduling, and PQ attestation components to an explicitly configured node lifecycle and verify their event consumers and shutdown.
- [x] Run meaningful negative-path tests, affected race tests, integration tests, full Go tests, vet, formatting, and local builds.
- [x] Update `STATUS.md` and the Track 3 completion record with implementation and verification evidence; commit scoped changes.

Functional scope and evidence: [Track 3 node integration](docs/track3-node.md).
Track 3 is **not fully complete**. Production snap synchronization still requires
moving-pivot range reuse and bounded-memory assembly. Snap/2 and standalone CL
validator/proposer tooling are additional work; the native observer has no
execution forkchoice authority.

## Next Phase Planning (2026-11-11)

Baseline: `master` at `137006e` (PR #48, `feat/eiptest-framework` merged). Re-verify each
gap against current code before starting — parts of the 2026-03 gap analysis are stale
(BAL header hash and `engine_getInclusionListV1` already exist).

### Track 0 — Housekeeping

- [x] Fast-forward local `master` to `origin/master`.
- [x] Delete merged local branch `feat/eiptest-framework` (remote already deleted).

### Track 1 — Glamsterdam / ePBS closure (P0)

- [x] EIP-7732 ePBS (85%): PTC election/voting wiring into fork-choice (types exist in `pkg/epbs/types.go`); BuilderPendingPayment/Withdrawal payment mechanism; builder registry (balance/deposit/withdrawable epoch); CL integration (payload → SignedBuilderBid in BeaconBlockBody); wire `epbs/bid SubmitBid` to a real builder entry point.
- [x] EIP-7975 (ETH/70 partial block receipts) and EIP-8159 (ETH/71 BAL exchange): move from referenced-only (`pkg/eth/protocol.go`, `pkg/p2p/ethproto`) to implemented and tested.
- [x] EIP-7928 BAL residual gaps: Engine BAL mismatch must reject (probe currently only logs); verify remaining 10% gap items against the 2026-03 analysis. (Rejection already enforced; nil BAL hash now rejected post-Amsterdam.)
- [x] EIP-7594 PeerDAS (80%): column subnet gossip, custody, <50% column reconstruction, txpool `cell_proofs`. (Canonical `data_column_sidecar_<subnet>` topics + SignedBlockHeader added; custody/reconstruction/cell_proofs were already in place.)

### Track 2 — Hegotá closure (P1)

- [x] EIP-7805 FOCIL (85%): CL P2P inclusion-list gossip validation rules; equivocation (duplicate IL per committee member) evidence handling.
- [x] EIP-8141 FrameTx residual: re-verify APPROVE scope semantics (0x0/0x1/0x2) against the current draft; payer gas approval path; frame receipt format. (Payer charging was missing entirely — now collected on APPROVE(1)/(2); receipts carry the payer.)
- [x] EIP-7495 SSZ StableContainer: add tests (currently referenced-only). (HTR aligned with the EIP-7495 ProgressiveContainer spec form; conformance tests added.)
- [x] EIP-8025 Execution Proofs: close out to complete. (Proofs now anchored to the real EIP-7928 BAL with header-commitment verification.)
- [x] SSZ fork activation (EIP-6404/7807/7916): consensus activation, Engine API versioning, RPC migration (explicitly deferred fork work). (SSZBlockTime fork switch; engine_getPayloadV7/newPayloadV7 with SSZ hash binding; transitional header-form block hash.)

### Track 3 — Node fundamentals (cross-cutting)

- [ ] Real snap sync: fixed-pivot proof-verified bootstrap, persistent hashed state, authenticated healing and recovery are implemented and tested. Moving-pivot continuation and bounded-memory assembly remain open; do not equate the small-chain end-to-end test with production snap-sync completion.
- [x] CL-driven beacon sync loop (`pkg/sync/beacon`): authenticated Engine request → hash-bound `eth` download → staged execution → atomic forkchoice publication; retries, target replacement, cancellation and `eth_syncing` are wired.
- [x] Wire own consensus components (3SF, quick slots, PQ attestations) into node lifecycle as an explicit research observer: configuration-bound signatures, source-target quorums, RLPx vote delivery, CL reorg re-anchoring and joined shutdown. The external CL retains execution forkchoice authority; this is not a standalone consensus client.

### Track 4 — I* / PQ direction

- [x] PQC suite: real implementations for Dilithium3, padded Falcon-512, and SPHINCS+-SHA2-128f-simple, alongside ML-DSA-65. Historical Falcon and SPHINCS+ require CGO.
- [x] EIP-7932 draft PQ registry and EIP-8051 draft gas and expanded-key validation tests through explicit EVM registration.
- [ ] EIP-7864 Binary Merkle Trie: continue toward complete.
- [ ] Remaining STARK / lattice / aggregation protocol migrations (CODEBASE_AUDIT §VII).

#### Track 4 implementation checklist (2026-11-13)

- [x] Expose the historical signature formats explicitly and test interoperability, malformed inputs, and the no-CGO boundary.
- [x] Test the pinned EIP-7932 registry and both expanded-key EIP-8051 precompile variants through explicitly registered EVM calls.
- [x] Correct binary trie proof authentication and test multi-value stems and exclusion.
- [x] Match EIP-7864 Python reference roots for basic insertion and a state embedding with storage and code boundaries.
- [x] Expose resumable offline migration through the client command and read its output back from a reopened FileDB.
- [x] Exercise EIP-7932/8051 through an opt-in private-chain fork with distinct addresses, state-backed public keys, and account access charging.
- [ ] Confirm canonical fork addresses and key registry rules for EIP-7932/8051. The pinned drafts both assign 0x12, which the pinned EIP-7997 factory uses.
- [x] Implement the EIP-7748 ordered conversion step, zero-override protection, and resumable cursor for the pinned BLAKE3 overlay.
- [x] Exercise the pinned BLAKE3 overlay on an opt-in chain through block construction, import, Engine API payload construction, reorg, rewind, and restart by replay. Apply ordered conversion before transactions and preserve explicit zero overrides.
- [x] Preserve all 256 bits of EVM storage keys when tracking EIP-4762 storage access and edit events. Test branch and leaf gas for repeated, adjacent, and high-bit-distinct slots.
- [ ] Implement the production EIP-7864 path with the agreed key layout and complete EIP-4762 witness gas. Add a scalable frozen-state reader and restore the execution state directly from durable checkpoints. See the completion audit in docs/CODEBASE_AUDIT.md §VII.
- [ ] Connect the PQ attestation verifier to the node's authenticated committee and key source.
- [x] Replace unauthenticated experimental blob witnesses with ordered openings and verify complete deterministic lattice commitments in batches.
- [x] Verify every PQ attestation in the consensus batch, bind publication and classic fallback to authorized keys, and reject XMSS relabeling.
- [ ] Specify and migrate application-specific STARK, lattice and aggregation protocols with real providers and consumer verification. (STARK aggregate and block-bound frame provider boundaries are fail-closed; no production STARK provider is configured.)
- [x] Run affected suites and local build; record completed work and remaining gaps in STATUS.md.

#### Track 4 overlay checkpoint follow-up

- [x] Persist the binary root and conversion cursor with each imported overlay block.
- [x] Validate checkpoints across forkchoice, restart, write failure, and older replay-only databases.
- [x] Share immutable BLAKE3 nodes across branches and write only changed paths.
- [x] Distinguish zero storage leaves from absent leaves for witness gas and discard events for unpaid opcodes.
- [x] Reject private PQ address collisions across scheduled precompile forks before opening a chain.
- [x] Cross-check the public drafts and record outstanding protocol and adoption requirements.

#### Track 4 integration and devnet verification

- [x] Exercise signed PQ transactions through JSON-RPC, block construction, independent Engine import, and persisted state readback.
- [x] Check both EIP-8051 variants and the EIP-7932 registry through the activated private profile.
- [x] Run a fresh Kurtosis devnet with the Track 4 image and compare execution and beacon-chain observations.

#### Track 4 bug review

- [x] Verify overlay state copying, canonical lookup, replay retention, and database error handling.
- [x] Verify Engine payload identity and forkchoice behavior with the binary overlay.
- [x] Verify both binary trie proof APIs bind the complete key and leaf.
- [x] Verify PQ attestation authorization and aggregation inputs.
- [x] Enforce execution proof batch ownership, size limits, and verification deadlines.
- [x] Run regressions, affected race tests, and the local build.

### Track 5 — eiptest framework (incremental)

- [ ] T5.1: opportunistic migration of `pkg/`-root E2E scenarios as files are touched.
- [ ] T4.3: post-first-feature ergonomics notes in the framework README.

### Track 6 — Devnet and site

- [x] Gloas devnet dependency (recorded at `a75a67a`); full-feature devnet run validating ePBS + FOCIL + BAL together. (full-feature-eth2030: fulu@32, gloas@64+Amsterdam BAL enforcement, heze@96 all clean; newPayloadV5 with BAL accepted from lighthouse 8.4.0-rc.0; follow-up: expose blockAccessListHash in the eth_getBlockByNumber serializer.)
- [ ] Site publishing: deployment workflow names `eth2030.com` while GitHub Pages binds `eth2030.altllm.ai`.

Recommended start order: Track 1 (ePBS PTC + fork-choice is the largest single P0), Track 3 snap-sync in parallel, Track 6 devnet as the integration gate.

## Native Execution Requests and SSZ Integration

- [x] Trace native request collection, block commitment, replay, and Engine API handling against the EIPs.
- [x] Collect canonical deposit, withdrawal, and consolidation requests through the native execution paths.
- [x] Verify nonempty requests through block building, validation, Engine API, and the SSZ projection.
- [x] Run affected and default suites and record the results.

## Execution SSZ and Framework Test Effectiveness Review

- [x] Verify the branch and run the default, tagged, and race suites.
- [x] Measure whether targeted behavior changes fail their intended tests.
- [x] Close demonstrated assertion gaps and repeat the targeted checks.
- [x] Record the results in `STATUS.md`.

## Execution SSZ Specification Review

- [x] Compare transaction and block encodings with the pinned and published EIP drafts.
- [x] Normalize EIP-6404 transaction bytes, signatures, and Merkle roots.
- [x] Implement the published EIP-7807 payload and header summary codecs.
- [x] Verify independent vectors and affected suites.

## EIP Test Framework Completion

- [x] Repair simulator results, fork timing, transaction signing, and failure assertions.
- [x] Complete STF, pool, encoding, RPC, Engine, and differential probes with regression tests.
- [x] Enforce the coverage gate and verify generated scaffolds and the README walkthrough.
- [x] Develop and verify a new EIP test-first through the scaffold.
- [x] Complete scenario contracts and run Amsterdam fixture generation and consumption.
- [x] Run the default and variant suites, verify the build, and reconcile the plan and status.

## EIP Test Framework Acceptance Review (2026-11-11)

- [x] Compare the implementation and examples with the plan's acceptance criteria.
- [x] Run the framework suite, default Go suite, and local CI checks.
- [x] Verify failure detection and exercise the documented tooling entry points.
- [x] Record verified results and remaining gaps in `STATUS.md`.
## PR #47 Merge Resolution

- [x] Confirm the pull request head and fetch the current upstream base.
- [x] Create an isolated worktree for the merge.
- [x] Reconcile the status entries and preserve the roadmap and tracker navigation.
- [x] Build the site and Go module, and verify the static export and selected Go packages.
- [x] Scope all merge and staging operations to the isolated worktree.

## Cryptography Regression Fixes

- [x] Wire PQ signature verification through transaction admission and native execution.
- [x] Activate canonical BLS precompiles at Prague and preserve historical native execution rules.
- [x] Correct RLPx identity, PeerDAS cell validation, and ECIES empty messages.
- [x] Repair proof batch deadlines and STARK gossip round-tripping.
- [x] Verify the regression probes, upstream vectors, and the full Go suite.

## Public Cryptography Library Migration (2026-11-10)

- [x] Replace secp256k1 signing and recovery, BN254 precompile arithmetic, and BLAKE2b compression.
- [x] Replace BLS12-381 precompile arithmetic, mappings and compression, and consolidate KZG providers.
- [x] Replace ECIES, keystore, RLPx, and Engine JWT implementations and document protocol migration boundaries.
- [x] Integrate public PQ signature and KEM providers with explicit encoding compatibility.
- [x] Move Groth16 verification, AA circuit startup, Banderwagon group arithmetic, and EVM transforms to gnark.
- [x] Remove superseded implementations and standalone placeholder backends after checking callers.
- [x] Verify a clean snapshot with Go builds, vet, fixture-backed tests, and the tagged suites.
- [ ] Complete experimental proof protocols in the follow-up below.

## Ethereum Client Cryptography Follow-Up (2026-11-10)

- [x] Trace the libraries used by the pinned Geth and consensus references to their actual callers.
- [x] Migrate experimental IPA openings, Poseidon, VDF and VRF to public providers.
- [x] Migrate encrypted-mempool share reconstruction and retire unused standalone schemes.
- [x] Verify actual IPA and KZG openings in batch and require configured SNARK and STARK providers.
- [x] Carry complete ML-DSA keys and signatures through the gossip codec.
- [x] Verify the full source snapshot with pinned EF fixtures, build, vet, tagged, Go-only, and targeted race checks.
- [ ] Migrate the remaining application-specific STARK, lattice commitment and aggregation protocols.

## Unused Package Cleanup (2026-11-10)

- [x] Remove the six packages selected in the first cleanup batch and update package indexes.
- [x] Verify builds, tests, and the executable dependency graphs from a clean snapshot.

## Codebase Audit Follow-Up (2026-11-10)

- [x] Verify native EIP-4844 point evaluation with the Ethereum ceremony setup and align binary build tags.
- [x] Provision pinned EF fixtures in CI and enforce the geth-backed vector results.
- [x] Upgrade gnark and exercise BN254 circuit compilation, setup, proving, and verification.
- [x] Resolve permanent skips, placeholders, weak checks, tagged suites, and race coverage.
- [x] Verify dead-code candidates and clean ignored local audit artifacts.
- [x] Run the supported build and test matrix and record each outcome.

## Codebase Audit Evidence Review (2026-11-10)

- [x] Reconcile reachability findings with production, test, and tagged builds.
- [x] Record the fixture, test coverage, and dependency scan results with their limits.
- [x] Revise the proposed follow-up and verify the documentation and local builds.

## English Documentation Review (2026-11-10)

- [x] Scan project-owned documentation for Chinese text.
- [x] Manually translate `docs/CODEBASE_AUDIT.md` and any other Chinese documentation.
- [x] Record the English documentation requirement in `AGENTS.md`.
- [x] Verify the translation, repeat the language scan, and run local builds.

## Inclusion List JSON Refactor (2026-03-21)
- [x] Extract small inclusion-list JSON parsing helpers in `pkg/engine/inclusion_list.go`.
- [x] Reuse the helpers from the custom marshal/unmarshal code.
- [x] Add or update focused parsing tests.
- [x] Run the relevant Go tests successfully.

## Blob Bundle Refactor (2026-03-21)
- [x] Extract shared blob sidecar / cell-proof expansion helper in `pkg/engine/backend.go`.
- [x] Reuse the helper from the V2 and V6 blob-bundle builders.
- [x] Run the relevant Go tests successfully.

## Payload Retrieval Refactor (2026-03-21)
- [x] Extract shared helpers for async and stored payload lookup in `pkg/engine/backend.go`.
- [x] Reuse the helpers from the V3/V4/V5/V6 payload retrieval methods.
- [x] Run the relevant Go tests successfully.

## Inclusion List Selection Refactor (2026-03-21)
- [x] Extract shared inclusion-list transaction selection helper.
- [x] Switch engine and node backends to the shared helper.
- [x] Add or update tests for oversized transaction skipping.
- [x] Run the relevant Go tests successfully.

## Engine Review Follow-Up (2026-03-21)
- [x] Verify whether `engine_getPayloadV5` JSON encoding has already been fixed on this branch.
- [x] Reject overflowed hex values in `pkg/engine/inclusion_list.go` `flexibleUint64`.
- [x] Make node backend `ProcessInclusionList` fail explicitly when inclusion-list storage is unsupported.
- [x] Add or update tests for the inclusion-list fixes.
- [x] Run the relevant Go tests successfully.

## Engine Lock Wrapper Refactor (2026-03-19)
- [x] Add mutex-backed accessor helpers for `stateMu`, `blocksMu`, `payloadMu`, and `ilMu` in `pkg/engine/backend.go`.
- [x] Migrate direct lock/unlock call sites in `pkg/engine/backend.go` to the accessors.
- [x] Remove remaining direct forkchoice field reads from `ForkchoiceUpdated`.
- [x] Refactor `pkg/engine/backend_bodies.go` to use backend accessors instead of direct mutex operations.
- [x] Run the relevant `pkg/engine` Go tests successfully.

## Go Test Repair (2026-03-19)
- [x] Reproduce the failing or hanging `go test ./...` behavior from `pkg/`.
- [x] Isolate the package or test causing the issue.
- [x] Implement the minimal fix and add or update unit tests.
- [x] Run the affected tests and the full module test suite successfully.

## Ongoing
- [ ] Build the local Lighthouse image `eth2030cl:local`.
- [ ] Ensure the local EL image `eth2030:local` is available.
- [ ] Clean up any existing `eth2030-devnet` enclave.
- [ ] Launch `pkg/devnet/kurtosis` with the `full-feature` config.
- [ ] Verify block production and inspect CL/EL error logs.
- [x] Extend EVM formal core with jump-family opcodes (`JUMP`, `JUMPI`, `JUMPDEST`, `POP`).
- [x] Expand compiler decoding for jump-family opcodes.
- [x] Add execution and run-level correctness lemmas for the expanded opcodes.
- [x] Add bytecode-offset interpreter (`runBytecode`) and decoding helpers (`decodeAt`, `decodeNoImmediate`, `decodePush`).
- [x] Document offset-vs-index jump mismatch using an executable counterexample theorem pair (`run` vs `runBytecode`).
- [x] Configure Lean workspace test driver (`lake test`) and add a unified Lean test module.
- [x] Add `.gitignore` entries for Lean artifacts (`.lake`, `*.olean`, `*.ilean`).
- [ ] Add a full refinement theorem from `compile` to bytecode-offset execution (`compile`↔`runBytecode`).
- [ ] Add machine-state fields for `pc`/memory/calldata/accounting as separate invariants.
- [ ] Prove compilation soundness for a larger instruction subset with controlled jump targets.

## Backlog
- `goals`:
  - define a bytecode operational semantics that respects byte-offset jumps,
  - prove compile-then-run agreement against bytecode-offset semantics,
  - then lift jump semantics from op-index simulation to byte-offset simulation,
  - prove end-to-end `compile`→`run` correctness against bytecode semantics,
  - begin RISC-V interpreter/simulation proof chain once VM core stabilizes.

## Website Strawmap refresh (2026-11-05)

- [x] Inspect the latest EF Strawmap and selected EIPs; distinguish draft milestones from network activation.
- [x] Inspect selected implementation and test evidence; record gaps and limits.
- [x] Replace outdated roadmap and unsupported promotional claims with sourced, reviewable content.
- [x] Verify site build, evidence paths and responsive presentation; update STATUS.md.
- [x] Commit only files changed for this task.
- [ ] Complete a full current-spec conformance and cross-client audit before claiming broader coverage.

## Scoped devnet control review

- [x] Prepare the controller and fixtures independently from the Frame codec.
- [x] Verify the process lifecycle and evidence assertions on this review branch.
