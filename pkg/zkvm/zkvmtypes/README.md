# zkvm/zkvmtypes — Shared types for the zkVM framework

## Overview

`zkvm/zkvmtypes` defines the core data structures shared across the zkVM subsystem. It provides a common vocabulary for guest programs, verification keys, proofs, execution results, and the `ProverBackend` interface — decoupling the zkVM framework from any specific proving system (SP1, RISC Zero, mock, etc.).

These types are referenced by the broader `zkvm` package (STF executor, zkISA bridge, canonical guest registry) and by `proofs` and `rollup` packages that consume zkVM outputs.

## Functionality

**Types**

| Type | Description |
|---|---|
| `GuestProgram` | Compiled guest bytecode (`Code []byte`), entry point name, and STF `Version uint32`. |
| `VerificationKey` | Serialized verification key (`Data []byte`) and `ProgramHash types.Hash`. |
| `Proof` | Serialized proof (`Data []byte`) and `PublicInputs []byte` (pre/post state roots for rollups). |
| `ExecutionResult` | `PreStateRoot`, `PostStateRoot`, `ReceiptsRoot` (`types.Hash`), `GasUsed uint64`, `Success bool`. |
| `GuestInput` | Input to the guest program: `ChainID uint64`, `BlockData []byte` (RLP block), `WitnessData []byte` (RLP witness). |

**Interface**

```go
type ProverBackend interface {
    Name() string
    Prove(program *GuestProgram, input []byte) (*Proof, error)
    Verify(vk *VerificationKey, proof *Proof) (bool, error)
}
```

Implement `ProverBackend` to wire a concrete proof system (e.g., gnark Groth16, SP1, RISC Zero) into the zkVM pipeline without changing call sites.

## Usage

```go
// Construct guest input for a block proof.
input := &zkvmtypes.GuestInput{
    ChainID:     1,
    BlockData:   rlpBlock,
    WitnessData: rlpWitness,
}

// Call a ProverBackend to generate a proof.
proof, err := backend.Prove(guestProgram, mustMarshal(input))

// Verify the returned proof.
ok, err := backend.Verify(vk, proof)

// Inspect execution outcome embedded in PublicInputs.
var result zkvmtypes.ExecutionResult
_ = rlp.Decode(bytes.NewReader(proof.PublicInputs), &result)
```

---

Parent package: [`zkvm`](../)
