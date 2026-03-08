# consensus

Ethereum Consensus Layer (CL) implementation targeting the L1 Strawmap roadmap.

## Overview

Package `consensus` implements the full Ethereum beacon chain consensus layer,
from current Casper FFG finality through the entire L1 Strawmap roadmap: 3-slot
finality (3SF), quick 6-second slots, 1-epoch finality, endgame sub-slot finality,
post-quantum attestations, and 1M attestations per slot.

The package is organized around `FullBeaconState` — a thread-safe beacon state holding
validators, balances, justification bits, finality checkpoints, and historical roots.
`SSFState` implements single-slot finality vote accumulation with 2/3+ supermajority
detection using the 3SF-mini backoff algorithm for justifiable slot selection.
`ForkChoiceStore` implements LMD-GHOST, the greedy heaviest-observed subtree fork
choice rule used by the beacon chain.

Many advanced features live in dedicated subpackages (`jeanvm`, `secretproposer`,
`vdf`, `pq`, etc.) and are re-exported at the top level via compat files for backward
compatibility.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Validator Registry (EIP-7251)

`ValidatorBalance` represents a validator with the EIP-7251 increased maximum
effective balance of 2048 ETH (up from 32 ETH). The 32 ETH minimum activation
balance (`MinActivationBalance`) is preserved. `ComputeEffectiveBalance` applies
hysteresis to avoid oscillation. `ValidatorSet` provides thread-safe add/remove/get
and `ActiveCount(epoch)`.

### Attestations (EIP-7549)

`AttestationData` omits the committee index from the signed payload per EIP-7549,
enabling cross-committee aggregation. `MaxAttestationsElectra = 8` per block (reduced
from 128 because each attestation now represents a full committee). `AttestationPool`
and `AttestationPoolV2` manage pending attestations; `AttestationAggregator` and
`AttestationScaler` handle aggregation up to 1M attestations per slot via parallel BLS
verification with 64 subnets (`parallel_bls.go`, `batch_verifier.go`).

### Single-Slot Finality (3SF / SSF)

`SSFState` accumulates per-slot validator votes (`SSFVote`) and reports finality when
accumulated stake for a block root exceeds 2/3 of `TotalStake`. `IsJustifiableSlot`
implements the 3SF-mini backoff: a slot is justifiable when the delta from the
finalized slot is ≤5, a perfect square, or an oblong number — preventing premature
votes under network partitions while ensuring progress.

`SSFRoundEngine` and `SSFVoteTracker` manage the round-trip: collecting, aggregating,
and verifying votes. The `SSFEngine` coordinates the full SSF pipeline.

### Finality and Epoch Processing

`FinalityTracker` tracks justification bits and runs both standard Casper FFG
dual-epoch finality and single-epoch finality (`trySingleEpochFinality`).
`EpochProcessor` handles end-of-epoch rewards, penalties, and registry updates.
`EpochManager` coordinates epoch transitions across finality, validator lifecycle,
and committee rotation components.

`EndgameFinalityTracker` extends `FinalityTracker` for sub-slot finality (M+ era):
it partitions each slot into sub-slot intervals (`SubSlotCount = 3`) and finalizes
within a single slot when 2/3+ stake attests during the slot.

`BFTFinalityPipeline` provides a Byzantine Fault Tolerant finality pipeline with
explicit view-change support.

### Fork Choice

`ForkChoiceStore` implements LMD-GHOST. It maintains a tree of `BlockNode` values
with accumulated attestation weights, applies proposer boost, and exports `Head()` for
the canonical head computation. `ForkChoiceV2` extends it with additional filtering
for the K+ era.

### Committee Selection and Rotation

`CommitteeAssignment` maps validators to beacon committees for a given epoch.
`CommitteeRotation` manages epoch-boundary committee shuffling. `CommitteeSubnet`
tracks subnet assignments for gossip. `ShufShuffling` implements the beacon chain
shuffle algorithm. APS (Attester Participation Score, L+ era) committee selection is
provided via the `aps` subpackage.

### Validator Lifecycle

`ValidatorLifecycle` manages activation, exit, slashing, and withdrawal queue
processing. `ConsolidationManager` handles EIP-7251 validator consolidations.
`ExitQueue` (via `exitqueue` subpackage) enforces churn limits.

### Proposer Election and Slashing

`ProposerElection` and `ProposerRotation` handle per-slot proposer selection from
RANDAO. `ProposerSlashing` detects and processes equivocating proposers.
`AttesterSlashing` and `EquivocationDetector` handle attester slashing.

### Deposits and Withdrawals

`Deposits` processes EIP-6110 in-protocol deposit receipts. `Withdrawals` implements
EIP-4895 validator withdrawals. `DepProcDepositProcessing` provides the full deposit
processing pipeline.

### RANDAO and Randomness

`Randao` manages RANDAO reveal accumulation for each slot. `RandomAttester`
provides randomized committee sampling for sub-sampled attestation (sampled
attestation scheme). `SampledAttestation` implements the full sampling pipeline.

### Sync Committee

`SyncCommitteeAltair` implements Altair sync committees. `SyncCommitteeRotation`
manages 256-epoch sync committee periods.

### Distributed Block Building

`DistBuilder` and `DistCoordinator` implement the distributed block builder
(L+ roadmap item): multiple builders can register bids, and the coordinator runs
auctions and selects the winning block via an ePBS-compatible mechanism.

### jeanVM Aggregation (Groth16 ZK-circuit BLS)

`JeanVMAggregator` uses Groth16 ZK circuits (`AggregationCircuit`,
`BatchAggregationCircuit`) to aggregate BLS attestation proofs in zero-knowledge,
enabling 1M attestations per slot with compact proof representation (L+ era).

### Secret Proposers (VRF Election)

`SecretProposerSelector` implements commit-reveal VRF-based proposer election: the
proposer commits to a secret before the slot and reveals it at proposal time.
`ValidateCommitReveal` enforces the protocol.

### Post-Quantum Attestations

The `pq` subpackage implements Dilithium post-quantum attestation signing and
verification. `PQFinalityTest` validates PQ-secured finality round-trips. PQ chain
security (SHA-3 fork choice) is also wired via compat exports.

### Attester Stake Cap and APS

`AttesterCap` and `AttesterCapExtended` enforce per-attester stake caps to limit
individual validator influence. `AttestorCap` provides the APS-compatible variant.

### Checkpoint Store and Lean Chain

`CheckpointStore` maintains finality and justification checkpoints for the light
client interface. `LeanChain` provides a minimal chain representation used by
interop and lean-spec tests.

### Unified Beacon State

`UnifiedBeaconState` merges v1, v2, and modern beacon state formats into a single
structure with conversion helpers, used by the Engine API and ETH/72 protocol.

### Rich Data and Rewards

`RichData` provides block-level metadata including inclusion delays and attestation
scores. `RewardCalculator` and `RewardCalculatorV2` compute per-validator rewards and
penalties for attestation participation, sync committee contributions, and proposer
inclusion rewards.

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`aps/`](./aps/) | Attester Participation Score committee selection (L+ APS roadmap item) |
| [`clconfig/`](./clconfig/) | Consensus layer configuration constants and fork parameters |
| [`cltypes/`](./cltypes/) | Shared CL wire types (BeaconBlock, SignedBeaconBlock, etc.) |
| [`endgamepipe/`](./endgamepipe/) | Sub-slot finality pipeline for endgame finality (M+ era) |
| [`exitqueue/`](./exitqueue/) | Validator exit queue with epoch churn limit enforcement |
| [`fastconfirm/`](./fastconfirm/) | Fast confirmation pre-finality signal (Glamsterdam) |
| [`headval/`](./headval/) | Head block validation helpers |
| [`includer/`](./includer/) | 1 ETH includer selection logic (L+ roadmap) |
| [`jeanvm/`](./jeanvm/) | Groth16 ZK-circuit BLS aggregation (jeanVM, L+ roadmap) |
| [`kps/`](./kps/) | K+ era protocol structures and slot timing |
| [`minimmit/`](./minimmit/) | Minimal commit message for fast confirmation |
| [`pq/`](./pq/) | Post-quantum attestation signing (Dilithium, L+ era) |
| [`prequorum/`](./prequorum/) | Pre-quorum voting phase management |
| [`propauction/`](./propauction/) | Proposer auction (ePBS-compatible, distributed block building) |
| [`queue/`](./queue/) | Generic queue utilities used by finality and exit processing |
| [`recovery/`](./recovery/) | 51% attack auto-recovery pipeline (M+ roadmap) |
| [`rewards/`](./rewards/) | Attestation and sync committee reward tables |
| [`richdata/`](./richdata/) | Block-level rich metadata (inclusion delays, scores) |
| [`secretproposer/`](./secretproposer/) | VRF-based secret proposer election with commit-reveal |
| [`slashdetect/`](./slashdetect/) | Slash detection for proposers and attesters |
| [`slot/`](./slot/) | Slot duty scheduler and timing utilities |
| [`techdebt/`](./techdebt/) | Technical debt reset helpers (beacon spec modernization) |
| [`vdf/`](./vdf/) | VDF randomness beacon (Wesolowski scheme, M+ roadmap) |
| [`voting/`](./voting/) | Voting round management and aggregation |
| [`vrf/`](./vrf/) | VRF election utilities for secret proposer selection |

## Usage

```go
import "github.com/eth2030/eth2030/consensus"

// Create a beacon state with EIP-7251 validator support.
cfg := consensus.DefaultConfig()
state := consensus.NewFullBeaconState(cfg)

// Add a validator (max effective balance 2048 ETH per EIP-7251).
v := &consensus.ValidatorBalance{
    Pubkey:           pubkey,
    EffectiveBalance: 32 * consensus.GweiPerETH,
    ActivationEpoch:  0,
    ExitEpoch:        consensus.FarFutureEpoch,
}
state.AddValidator(v, 32*consensus.GweiPerETH)

// Single-slot finality: accumulate votes and check finality.
ssfState := consensus.NewSSFState(consensus.DefaultSSFConfig())
if err := ssfState.CastVote(vote); err == nil {
    if ok, _ := ssfState.CheckFinality(slot, blockRoot); ok {
        ssfState.FinalizeSlot(slot, blockRoot)
    }
}

// LMD-GHOST fork choice.
store := consensus.NewForkChoiceStore(consensus.ForkChoiceConfig{
    FinalizedEpoch: 0,
})
store.OnBlock(block)
store.OnAttestation(attestation)
head := store.Head()

// Secret proposer: commit-reveal VRF election.
selector := consensus.NewSecretProposerSelector(
    consensus.DefaultSecretProposerConfig(), seed,
)
proposerIdx := consensus.DetermineProposer(slot, validatorCount, randaoMix)
```

## Documentation References

- [Roadmap](../../docs/ROADMAP.md)
- [Design Doc](../../docs/DESIGN.md)
- [Roadmap Deep-Dive](../../docs/ROADMAP-DEEP-DIVE.md)
- [EIP Spec Implementation](../../docs/EIP_SPEC_IMPL.md)
- [Gap Analysis](../../docs/GAP_ANALYSIS.md)
- [EIP-7251: Increase MAX_EFFECTIVE_BALANCE](https://eips.ethereum.org/EIPS/eip-7251)
- [EIP-7549: Move Committee Index Outside Attestation](https://eips.ethereum.org/EIPS/eip-7549)
