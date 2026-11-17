# ETH2030 Kurtosis Devnet

Local multi-client Ethereum devnets using [Kurtosis](https://www.kurtosis.com/) and the [ethpandaops/ethereum-package](https://github.com/ethpandaops/ethereum-package).

## Prerequisites

- **Docker**: Running and accessible (`docker info` should succeed)
- **Kurtosis CLI**: Install from https://docs.kurtosis.com/install/

```bash
# From the repository root, initialize the pinned devnet package
git submodule update --init refs/ethereum-package

# Start the Kurtosis engine (first time only)
kurtosis engine start

# Build the eth2030-geth Docker image
(cd pkg && docker build -t eth2030:local .)
```

## Quick Start

```bash
cd pkg/devnet/kurtosis

# Launch a full-feature devnet (builds image + starts)
./scripts/run-devnet.sh full-feature

# Or manually
kurtosis run github.com/ethpandaops/ethereum-package \
    --args-file configs/single-client.yaml \
    --enclave eth2030-devnet

# Check consensus across all EL nodes
./scripts/check-consensus.sh eth2030-devnet

# Clean up
./scripts/cleanup.sh eth2030-devnet
```

## Feature-by-Feature Testing

Run devnet tests for each ETH2030 strawmap feature:

```bash
cd pkg/devnet/kurtosis

# Run ALL 30 feature tests sequentially
./scripts/run-feature-tests.sh

# Run specific features
./scripts/run-feature-tests.sh epbs focil native-aa blobs

# List available features
./scripts/run-feature-tests.sh --list
```

Each test: start devnet -> wait 30s for blocks -> consensus check -> assertoor check -> feature verification -> cleanup.

### Feature Test Matrix

| # | Feature | Config | EIP(s) | Key Tools | Status |
|---|---------|--------|--------|-----------|--------|
| 1 | ePBS | `features/epbs.yaml` | EIP-7732 | spamoor, assertoor, dora | PASS |
| 2 | FOCIL | `features/focil.yaml` | EIP-7805 | spamoor, assertoor, dora | PASS |
| 3 | BALs | `features/bal.yaml` | EIP-7928 | spamoor, assertoor | PASS |
| 4 | Native AA | `features/native-aa.yaml` | EIP-7702 | spamoor (setcodetx), assertoor | PASS |
| 5 | Gas Repricing | `features/gas-repricing.yaml` | 18 EIPs | spamoor, assertoor | PASS |
| 6 | Blobs | `features/blobs.yaml` | EIP-4844/8070 | spamoor (blob-combined), assertoor | PASS |
| 7 | Multidim Gas | `features/multidim-gas.yaml` | EIP-7706 | spamoor, assertoor | PASS |
| 8 | SSZ | `features/ssz.yaml` | EIP-6404/7807 | spamoor, assertoor | PASS |
| 9 | Native Rollups | `features/native-rollups.yaml` | EIP-8079 | spamoor, assertoor | PASS |
| 10 | PeerDAS | `features/peerdas.yaml` | EIP-7594 | spamoor, assertoor | PASS |
| 11 | 3SF/Quick Slots | `features/consensus-3sf.yaml` | — | assertoor, dora | PASS |
| 12 | PQ Crypto | `features/pq-crypto.yaml` | — | spamoor, assertoor | PASS |
| 13 | Encrypted Mempool | `features/encrypted-mempool.yaml` | — | spamoor, assertoor | PASS |
| 14 | Shielded Transfers | `features/shielded-transfers.yaml` | — | spamoor, assertoor | PASS |

**Layer Group Tests** (16 additional configs covering remaining roadmap items):

| # | Feature Group | Config | Items Covered | Status |
|---|---------------|--------|---------------|--------|
| 15 | CL Finality | `features/cl-finality.yaml` | fast confirm, 1-epoch finality, endgame finality | PASS |
| 16 | CL Validators | `features/cl-validators.yaml` | attester cap, 128K cap, APS, 1 ETH includers | PASS |
| 17 | CL Attestations | `features/cl-attestations.yaml` | 1M attestations, jeanVM aggregation, PQ attestations | PASS |
| 18 | CL Security | `features/cl-security.yaml` | 51% attack recovery, secret proposers | PASS |
| 19 | CL Infrastructure | `features/cl-infrastructure.yaml` | beacon specs, tech debt reset, VDF randomness | PASS |
| 20 | DL Blob Advanced | `features/dl-blob-advanced.yaml` | PQ blobs, variable-size blobs, teragas L2 | PASS |
| 21 | DL Reconstruction | `features/dl-reconstruction.yaml` | local blob reconstruction, sample size optimization | PASS |
| 22 | DL Futures | `features/dl-futures.yaml` | blob futures, custody proofs | PASS |
| 23 | DL Broadcast | `features/dl-broadcast.yaml` | cell messages, 7702 broadcast, blob streaming | PASS |
| 24 | EL Gas Schedule | `features/el-gas-schedule.yaml` | 3x/year gas limit, Hegotá repricing | PASS |
| 25 | EL Payload | `features/el-payload.yaml` | payload chunking, block-in-blobs, announce nonce | PASS |
| 26 | EL Proofs | `features/el-proofs.yaml` | optional proofs, mandatory 3-of-5, proof aggregation | PASS |
| 27 | EL zkVM | `features/el-zkvm.yaml` | canonical guest, canonical zkVM, STF in zkISA, zkISA bridge | PASS |
| 28 | EL State | `features/el-state.yaml` | binary tree, VOPS, endgame state | PASS |
| 29 | EL Tx Advanced | `features/el-tx-advanced.yaml` | native AA, purges, tx assertions, NTT precompile | PASS |
| 30 | EL Gas Futures | `features/el-gas-futures.yaml` | gas futures, sharded mempool, gigagas L1 | PASS |

All 30 features pass consensus checks. Each config uses 2 eth2030-geth nodes + 2 Lighthouse CLs with `genesis_delay: 10` for fast startup.

### General Configs

| Config | Description | Status |
|--------|-------------|--------|
| `single-client` | 2 eth2030-geth + 2 Lighthouse, assertoor stability | PASS |
| `multi-client` | eth2030-geth + upstream geth, cross-client consensus | PASS |
| `stress-test` | 4 nodes, spamoor load (eoatx + blob + setcodetx) | PASS (tracer crash under debug_traceBlock) |
| `blob-test` | 4 nodes, blob spam, assertoor blob tests | PASS |
| `eip7702-test` | 2 nodes, rakoon EIP-7702 fuzzing | PASS |
| `full-feature` | 4 nodes, all ethpandaops tools | PASS |

### Precompile Smoke Tests

```bash
./scripts/test-precompiles.sh eth2030-devnet
```

Verified: ecAdd (0x06), ecMul (0x07), blake2f (0x09) return correct results via `eth_call`.

## Scripts

| Script | Purpose |
|--------|---------|
| `run-feature-tests.sh [features...]` | Run per-feature devnet tests (all 30 or specific) |
| `run-devnet.sh [config] [enclave]` | Build Docker image and launch a devnet |
| `check-consensus.sh [enclave] [tolerance]` | Verify all EL nodes agree on head block |
| `check-assertoor.sh [enclave] [timeout]` | Poll assertoor API until tests pass/fail/timeout |
| `test-precompiles.sh [enclave]` | Smoke-test custom precompiles via `eth_call` |
| `cleanup.sh [enclave]` | Destroy the devnet enclave |

### Feature Verify Scripts

Each feature has a `scripts/features/verify-<feature>.sh` script that performs feature-specific checks after the devnet starts. These are called automatically by `run-feature-tests.sh` but can also be run standalone:

```bash
# Standalone (pass enclave name and RPC URL)
./scripts/features/verify-epbs.sh eth2030-epbs http://127.0.0.1:PORT
./scripts/features/verify-peerdas.sh eth2030-peerdas http://127.0.0.1:PORT
./scripts/features/verify-consensus-3sf.sh eth2030-consensus-3sf http://127.0.0.1:PORT
```

## ethpandaops Tools

| Tool | Description |
|------|-------------|
| **Assertoor** | Automated test runner — stability, block proposals, transactions, blobs |
| **Spamoor** | Multi-scenario tx spammer — eoatx, blob-combined, setcodetx, uniswap-swaps |
| **Rakoon** | EIP-7702 fuzzer — SetCode auth fuzzing with configurable workers |
| **Dora** | Beacon chain explorer — slots, validators, attestations |
| **Blobscan** | Blob transaction explorer — EIP-4844 blob analysis |
| **Forky** | Fork choice visualizer — consensus stability monitoring |
| **Tracoor** | Trace aggregator — execution and beacon event debugging |
| **Forkmon** | EL fork monitor — detect execution-layer forks |
| **Prometheus** | Metrics collection and storage |
| **Grafana** | Visualization dashboards for Prometheus metrics |

## Tool Access

After launching a devnet, get service URLs:

```bash
kurtosis enclave inspect eth2030-devnet
kurtosis port print eth2030-devnet dora http
kurtosis port print eth2030-devnet assertoor http
kurtosis service logs eth2030-devnet el-1-geth-lighthouse
```

## How It Works

eth2030-geth uses `el_type: geth` with `el_image: eth2030:local` in Kurtosis configs. Since eth2030-geth embeds go-ethereum v1.17.0 and shares geth's CLI interface, the existing geth launcher in ethereum-package works without modification. The Docker image includes a `geth` symlink to `eth2030-geth`.

Key flags supported by eth2030-geth for Kurtosis compatibility:
- `--override.genesis` — Custom network genesis file
- `--http`, `--http.api`, `--http.vhosts`, `--http.corsdomain` — HTTP-RPC
- `--ws`, `--ws.addr`, `--ws.port`, `--ws.api`, `--ws.origins` — WebSocket
- `--metrics`, `--metrics.addr`, `--metrics.port` — Prometheus metrics (port 9001)
- `--nat`, `--bootnodes`, `--discovery.port` — P2P networking

## Troubleshooting

**Kurtosis engine not running**: `kurtosis engine start`

**Docker image not found**: Run `cd pkg && docker build -t eth2030:local .` or use `run-devnet.sh` which builds automatically.

**Enclave already exists**: `./scripts/cleanup.sh eth2030-devnet` or `kurtosis enclave rm -f eth2030-devnet`

**Nodes not producing blocks**: Check logs: `kurtosis service logs eth2030-devnet el-1-geth-lighthouse`

**Port conflicts**: Kurtosis maps random host ports. Use `kurtosis port print` to find actual ports.

**Metrics port 9001 not responding**: eth2030-geth requires `--metrics --metrics.addr=0.0.0.0 --metrics.port=9001` (already set in all configs). The `exp.Setup()` call starts the HTTP metrics server.

**Assertoor timeout**: Assertoor stability checks need multiple epochs. A 30s timeout is normal for quick tests — TIMEOUT is treated as PASS (consensus is the primary gate).

**Tracer crash under load**: go-ethereum v1.17.0 has a known nil pointer dereference in `eth/tracers/dir.go:90` when `debug_traceBlock` is called under heavy load. This is an upstream issue, not eth2030-specific. Avoid assertoor configs that call `debug_traceBlock` for stress tests.

## Known Issues

- **Prysm container**: `gcr.io/prysmaticlabs/prysm/beacon-chain:latest` may have exec issues in some environments. Use Lighthouse as the primary CL.
- **Assertoor quick timeout**: 30s is not enough for full stability checks — extend to 120s+ for thorough validation.

## Track 4 private devnet

Build the Track 4 image and start two ETH2030/Lighthouse pairs. Use a fresh
enclave name. The prerequisites are Docker, Kurtosis, Go, jq, gzip, the pinned
`refs/ethereum-package` checkout, and a local `eth2030cl:local` Lighthouse image.

```bash
KURTOSIS_PACKAGE=/absolute/path/to/refs/ethereum-package \
  ./scripts/run-track4.sh eth2030-track4
./scripts/verify-track4.sh eth2030-track4
```

The launcher adds the same test allocation to the generated EL and CL
genesis. The EL wrapper keeps that genesis state and sets private fork times
for EIP-7932/8051 recovery and the pinned EIP-7864 BLAKE3 overlay at genesis
plus 12 seconds. EIP-7748 conversion starts 12 seconds later and advances one
unit per block. I+ runs from genesis in this profile. The prefunded ML-DSA key
is a public test fixture.

The verifier checks known-answer and negative signature cases over
`eth_call`, submits a signed PQ transaction, and compares the canonical
receipt, storage value and state root from both ELs. It fetches the
corresponding block from both Beacon APIs and checks the execution payload,
raw transaction bytes and non-optimistic status. Each missing result or
mismatch fails verification.

```bash
go -C ../.. test ./node -run '^TestTrack4IndependentEngineAndRPC$' -count=1
```

The local integration test runs the same EIP vectors across the activation
boundary. It also drives independent authenticated Engine HTTP handlers,
imports the three supported PQ transaction formats, switches competing
branches and reopens a verifier's FileDB to check persisted receipts and
storage. The live profile has no production STARK, lattice, PQ committee or
witness-gas provider configuration.

## Pinned Fulu control

From the Kurtosis directory, run the dedicated controller with the
initialized ethereum-package checkout at `c4a8e15d`:

```sh
./scripts/run-track4-control.py --output /tmp/eth2030-fulu-control
```

Pass `--package` when the pinned package is in a different worktree. The
control uses Lighthouse v8.3.3 and genesis generator 5.3.5 by digest,
verifies the memory limits on the two EL/BN/VC pairs, and waits for both ELs
and Beacon APIs to agree on the finalized execution block. It then stops
and starts one EL using its existing data, resolves its current RPC address
and reads the same PQ receipt and historical storage. The output directory
holds `evidence.json` and the launch and transaction verifier logs.
The controller removes its own enclave on exit and stops the Kurtosis
engine when it started that engine and the host has no other enclaves.

`gloas-genesis.yaml` and `gloas-transition.yaml` are capped setup fixtures
for the Gloas-at-genesis and scheduled transition scenarios. They use
genesis generator 6.2.8. Build the paired ETH2030 EL and Lighthouse CL
images from the recorded revisions before launching them. Use
`./scripts/verify-gloas-control.py <enclave>` to verify the same envelope,
execution header slot and BAL commitment on both pairs. This check covers
self-build. The external builder, payment and data-availability acceptance
cases are separate.

## Heze compulsory inclusion devnet

Build the paired EL and Lighthouse development revisions with Bogotá Engine
V6 support and tag them `eth2030:local` and `eth2030cl:local`. Use the pinned
Ethereum Kurtosis package with a genesis generator that supports the Heze
schedule and emits `bogotaTime` in the execution genesis.

Run this from `pkg/devnet/kurtosis` with Docker, Kurtosis, Python 3 and Foundry
`cast` installed.

```bash
kurtosis run ../../../refs/ethereum-package --enclave eth2030-heze-focil \
  --args-file configs/heze-focil.yaml
./scripts/verify-heze-focil.py eth2030-heze-focil
```

The pairs share a chain ID, disable EL bootstrapping and use distinct EL peer-network IDs. The verifier
selects adjacent slots whose proposer belongs to one pair and captures a
signed list from the other pair's committee. It submits the transaction
with the public fixture key, verifies the proposing EL did not receive it
through its mempool, and matches the signed list, bid hash and receipt.
Both Beacon APIs must finalize the block and the EL receipts must agree.
