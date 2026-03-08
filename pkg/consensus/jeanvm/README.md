# jeanvm

Groth16 ZK-circuit BLS signature aggregation enabling 1M+ attestations per slot (L+ roadmap).

## Overview

Package `jeanvm` implements jeanVM aggregation: a ZK-circuit-based BLS
signature aggregation scheme that compresses up to `jeanVMMaxCommitteeSize`
(2048) attestation signatures per committee into a single 192-byte Groth16-style
proof. Up to `jeanVMMaxBatchSize` (64) committees can be further folded into
a batch proof via `BatchAggregateWithProof`, enabling 1M+ attestations per
slot with constant proof size.

Signatures are aggregated by XOR-ing BLS G2 points via `crypto.BLS12G2Add`
with a Keccak256 fallback for non-standard inputs. Proof generation uses a
deterministic Keccak256 hash chain (A–B–C points, 48+96+48 bytes) with domain
separation per the Groth16 structure.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `JeanVMAggregator` | Thread-safe aggregator with proof generation and verification |
| `JeanVMAttestationInput` | Single attestation: validator index, public key, signature, slot, committee index |
| `AggregationCircuit` | R1CS constraint model for BLS pairing verification |
| `JeanVMAggregationProof` | 192-byte proof + committee root + aggregate signature |
| `BatchAggregationCircuit` | Multi-committee batch circuit |
| `JeanVMBatchProof` | Batch proof over up to 64 committees |

### Constants

| Name | Value | Description |
|------|-------|-------------|
| `jeanVMProofSize` | 192 | Groth16 proof: A(48)+B(96)+C(48) bytes |
| `jeanVMMaxCommitteeSize` | 2048 | Max attestations per single proof |
| `jeanVMMaxBatchSize` | 64 | Max committees per batch proof |

### Functions / methods

| Name | Description |
|------|-------------|
| `NewJeanVMAggregator() *JeanVMAggregator` | Create aggregator |
| `(*JeanVMAggregator).AggregateWithProof(attestations, message) (*JeanVMAggregationProof, error)` | Aggregate and prove a single committee |
| `(*JeanVMAggregator).VerifyAggregationProof(proof, committeePubkeys) (bool, error)` | Verify a committee proof |
| `(*JeanVMAggregator).BatchAggregateWithProof(committees, messages) (*JeanVMBatchProof, error)` | Batch-aggregate up to 64 committees |
| `(*JeanVMAggregator).VerifyBatchProof(proof) (bool, error)` | Verify a batch proof |
| `(*JeanVMAggregator).EstimateGas(numSigs) uint64` | Gas cost estimate |
| `(*JeanVMAggregator).EstimateBatchGas(committees, totalSigs) uint64` | Batch gas with 20% discount |
| `(*JeanVMAggregator).Stats() (generated, verified, aggregated, batches uint64)` | Lifetime counters |
| `ValidateAggregationProof(proof) error` | Structural validation |
| `ValidateBatchAggregationProof(proof) error` | Batch structural validation |

## Usage

```go
import "github.com/eth2030/eth2030/consensus/jeanvm"

agg := jeanvm.NewJeanVMAggregator()
proof, err := agg.AggregateWithProof(attestations, slotMessage)
ok, err := agg.VerifyAggregationProof(proof, committeePubkeys)
```

[← consensus](../README.md)
