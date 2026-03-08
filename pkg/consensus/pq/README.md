# pq

Post-quantum attestation signing, verification, and PQ chain security (L+ era).

## Overview

Package `pq` provides two related post-quantum capabilities:

1. **PQ attestations** (`pq_attestation.go`): `PQAttestation` carries both a
   Dilithium3 PQ signature and an optional ECDSA classic signature for the
   transition period. `PQAttestationVerifier` tries PQ first, falls back to
   classic if configured. `CreatePQAttestation` signs using a `pqc.DilithiumKeyPair`.
   STARK-based batch verification (`STARKAggregateVerify`) is more efficient for
   batches larger than 4 attestations.

2. **PQ chain security** (`pq_chain_security.go`): `PQChainValidator` enforces
   SHA-3-based block hashing and tracks per-epoch PQ key registration ratios.
   `PQForkChoice` applies a 10% weight bonus to PQ-signed attestations.
   `PQHistoryAccumulator` maintains a SHA-3 Merkle commitment over block hashes.

`LeanConfig` (in `config.go`) allows the consensus package to pass PQ-relevant
config fields without a circular import.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `PQAttestationConfig` | `UsePQSignatures`, `FallbackToClassic`, `MinPQValidators` |
| `PQAttestationVerifier` | Verifies PQ+classic signatures; tracks verified/failed counts |
| `PQAttestation` | Attestation with Dilithium PQ sig + optional classic ECDSA sig |
| `PQChainConfig` | `SecurityLevel`, `PQThresholdPercent`, `TransitionEpoch`, `SlotsPerEpoch` |
| `PQChainValidator` | Epoch-based PQ enforcement; SHA-3 block hash validation |
| `PQForkChoice` | Fork choice with PQ-weighted attestations |
| `PQHistoryAccumulator` | Append-only SHA-3 Merkle tree of block hashes |
| `PQSecurityLevel` | `PQSecurityOptional`, `PQSecurityPreferred`, `PQSecurityRequired` |
| `LeanConfig` | PQ subset of `ConsensusConfig` (avoids circular import) |

### Key functions

| Name | Description |
|------|-------------|
| `CreatePQAttestation(slot, committeeIndex, blockRoot, sourceEpoch, targetEpoch, validatorIndex, pqKey)` | Sign a new PQ attestation |
| `(*PQAttestationVerifier).VerifyAttestation(att) (bool, error)` | Verify PQ then classic |
| `(*PQAttestationVerifier).STARKAggregateVerify(attestations) (*STARKSignatureAggregation, error)` | Batch STARK verification |
| `SelectLeanPQAttestors(validators, count, slot, epochSeed)` | Deterministic subset selection for lean mode |
| `PQBlockHash(header) (Hash, error)` | SHA-3-256 quantum-resistant block hash |
| `(*PQChainValidator).IsPQEnforced(epoch) bool` | True when PQ is mandatory for this epoch |
| `(*PQChainValidator).ValidateChainPQSecurity(headers) (*PQChainAuditResult, error)` | Audit a chain segment |
| `IntegratePQForkChoice(fc, pqFC) int` | Merge PQ weight bonuses into main fork choice |

## Usage

```go
import "github.com/eth2030/eth2030/consensus/pq"

cfg := pq.DefaultPQAttestationConfig()
verifier := pq.NewPQAttestationVerifier(cfg)

att, _ := pq.CreatePQAttestation(slot, committeeIdx, blockRoot,
    sourceEpoch, targetEpoch, validatorIdx, dilithiumKey)
ok, err := verifier.VerifyAttestation(att)
```

[← consensus](../README.md)
