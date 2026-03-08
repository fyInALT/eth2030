# Package light

Light client for the Ethereum beacon chain with sync committee verification and CL proof generation.

## Overview

The `light` package implements a stateless-friendly light client that tracks the beacon chain's finalized head without downloading the full chain state. It processes `LightClientUpdate` messages signed by beacon sync committees, validates BLS aggregate signatures, enforces the 2/3 supermajority threshold, and advances the finalized header on each successful update.

The package also implements the real-time CL proof system from the I+ upgrade. `CLProofGenerator` produces Merkle proofs for state root, validator existence, and balance queries. Proofs are cached by (type, slot, index) with a configurable TTL and evicted lazily on overflow. The `cache` subpackage provides a general-purpose LRU proof cache keyed by block number, address, and storage slot.

Header storage is abstracted behind the `LightStore` interface; an in-memory implementation (`MemoryLightStore`) is provided for testing. Production deployments should supply a persistent store. The `bls` subpackage provides a standalone BLS verifier for sync committee signatures.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Core Types

- `LightClientUpdate` — an update carrying attested header, finalized header, sync committee participation bits, aggregate BLS signature, and optional next sync committee
- `SyncCommittee` — a beacon chain sync committee (512 validators, one ~27-hour period)
- `LightBlock` — a block header bundled with state and transaction proofs
- `LightClientState` — current finalized header, slot, and active committee

`LightClientUpdate.SignerCount()` counts participation bits; `SupermajoritySigned(committeeSize)` checks the 2/3 threshold.

### Light Client

`LightClient` is the top-level client managing lifecycle and operations:

```
NewLightClient() *LightClient
NewLightClientWithStore(store LightStore) *LightClient
```

Key methods:
- `Start() / Stop() / IsRunning()` — lifecycle
- `ProcessUpdate(update)` — validates and applies a light client update
- `GetFinalizedHeader()` — returns the latest finalized header
- `IsSynced()` — whether a finalized header has been received
- `GetHeader(hash) / GetHeaderByNumber(num)` — header lookup
- `VerifyStateProof(header, key, proof)` — verifies a Keccak256-bound state proof
- `BuildStateProof(root, key, value)` — constructs a verifiable state proof

### Light Syncer

`LightSyncer` is the core update processor:

- Validates that finalized header number does not exceed the attested header
- Enforces the 2/3 supermajority on `SyncCommitteeBits`
- Verifies BLS aggregate signature via `crypto.DefaultBLSBackend()`
- Rotates the active `SyncCommittee` when `NextSyncCommittee` is provided
- Stores both the attested and finalized headers in `LightStore`

### CL Proof Generator

`CLProofGenerator` produces Merkle inclusion proofs for consensus layer state (I+ upgrade):

```
NewCLProofGenerator(config CLProofConfig) *CLProofGenerator
```

Proof types (constants `CLProofTypeStateRoot`, `CLProofTypeValidator`, `CLProofTypeBalance`, `CLProofTypeCommittee`):
- `GenerateStateRootProof(slot, stateRoot)` — leaf = Keccak256(slot || stateRoot)
- `GenerateValidatorProof(slot, validatorIndex, pubkey, balance)` — leaf = Keccak256(index || pubkey || balance)
- `GenerateBalanceProof(slot, validatorIndex, balance)` — leaf = Keccak256(index || balance)
- `VerifyProof(proof)` — recomputes the root from the Merkle branch

`DefaultCLProofConfig()` sets `MaxProofDepth=40`, `CacheSize=1000`, `ProofTTL=12s`.

### Light Store

```go
type LightStore interface {
    StoreHeader(header *types.Header) error
    GetHeader(hash types.Hash) *types.Header
    GetLatest() *types.Header
    GetByNumber(num uint64) *types.Header
}
```

`MemoryLightStore` provides an in-memory implementation with hash and number indexes.

### Sync Helpers

- `SignUpdate(header, committeeBits)` — signs an update with the test BLS secret key
- `MakeCommitteeBits(signers)` — creates a participation bitfield for `signers` out of 512
- `SyncCommitteeSize = 512`

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`bls/`](./bls/) | BLS signature verifier for sync committee aggregate signatures |
| [`cache/`](./cache/) | LRU proof cache keyed by (block number, address, storage key, proof type) with TTL eviction |

## Usage

```go
// Create a light client
lc := light.NewLightClient()
lc.Start()

// Build and process an update
bits := light.MakeCommitteeBits(350) // 350 of 512 signers
sig  := light.SignUpdate(attestedHeader, bits)
update := &light.LightClientUpdate{
    AttestedHeader:    attestedHeader,
    FinalizedHeader:   finalizedHeader,
    SyncCommitteeBits: bits,
    Signature:         sig,
}
if err := lc.ProcessUpdate(update); err != nil {
    // handle error
}

// Get finalized header
hdr := lc.GetFinalizedHeader()

// Generate a CL state root proof
gen := light.NewCLProofGenerator(light.DefaultCLProofConfig())
proof, err := gen.GenerateStateRootProof(slot, stateRoot)
if err == nil && light.VerifyProof(proof) {
    // proof is valid
}
```

## Documentation References

- [Ethereum Light Client Spec](https://github.com/ethereum/consensus-specs/blob/dev/specs/altair/light-client/)
- [L1 Strawmap: real-time CL proofs (I+ upgrade)](https://strawmap.org/)
- ETH2030 consensus layer: `pkg/consensus/`
- ETH2030 crypto (BLS): `pkg/crypto/`
