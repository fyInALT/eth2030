# proofs/groth16 — Groth16 ZK proof circuit for AA validation

## Overview

This package provides production-grade Groth16 proof generation and verification
for account abstraction (AA) operation validation using the
[gnark](https://github.com/consensys/gnark) library over the BN254 curve.

It implements `AAValidationGnarkCircuit`, an R1CS circuit that enforces three
constraints: nonce is nonzero, gas limit is nonzero, and the nonce is a strict
sequential increment over the previous nonce. Proving and verifying keys are
generated via a Groth16 trusted-setup phase and reused across all AA validations.

## Functionality

**Types**

- `AAValidationGnarkCircuit` — gnark circuit with public inputs `Nonce`,
  `GasLimit` and private witness `PrevNonce`; implements `frontend.Circuit`
- `GnarkAACircuitCS` — holds the compiled R1CS constraint system
- `GnarkAAProverKeys` — holds `groth16.ProvingKey` and `groth16.VerifyingKey`

**Functions**

- `CompileAACircuitGnark() (*GnarkAACircuitCS, error)` — compiles the circuit
  to R1CS via `frontend.Compile` on BN254's scalar field
- `SetupGnarkAAKeys(circuit) (*GnarkAAProverKeys, error)` — runs Groth16 setup
  to produce proving/verifying key pair
- `ProveGnarkAA(circuit, keys, nonce, gasLimit, prevNonce) (groth16.Proof, error)` —
  generates a Groth16 proof for the given witness
- `VerifyGnarkAAProof(keys, proof, nonce, gasLimit) (bool, error)` — verifies
  a proof against its public inputs using `groth16.Verify`
- `GnarkIntegrationStatus() string` — returns the backend identifier string

**Parent package:** [proofs](../)
