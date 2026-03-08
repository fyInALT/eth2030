# core/vops — Validity-Only Partial Statelessness (VOPS)

[← core](../README.md)

## Overview

Package `vops` implements Validity-Only Partial Statelessness, which allows nodes to validate Ethereum state transitions using only a partial subset of the world state and cryptographic validity proofs, without storing or accessing the full state trie. This is a key building block for the I+ roadmap item "validity-only partial state".

A `PartialExecutor` applies transactions against a `PartialState` (accounts + storage slots + contract code needed for those transactions). A `VOPSValidator` verifies that a pre→post state transition is attested by a valid `ValidityProof` without re-executing. The `WitnessAccumulator` collects accessed keys during execution for proof construction. `ProofChecker` provides Merkle storage proof verification.

## Functionality

### Types

```go
type PartialState struct {
    Accounts map[types.Address]*AccountState
    Storage  map[types.Address]map[types.Hash]types.Hash
    Code     map[types.Address][]byte
}
func NewPartialState() *PartialState
func (ps *PartialState) GetAccount(addr) *AccountState
func (ps *PartialState) SetAccount(addr, acct)
func (ps *PartialState) GetStorage(addr, key) types.Hash
func (ps *PartialState) SetStorage(addr, key, value)

type ValidityProof struct {
    PreStateRoot, PostStateRoot types.Hash
    AccessedKeys [][]byte
    ProofData    []byte
}

type ExecutionResult struct {
    GasUsed      uint64
    Success      bool
    AccessedKeys [][]byte
    PostState    *PartialState
}

type VOPSConfig struct {
    MaxStateSize int // default 10000
}
func DefaultVOPSConfig() VOPSConfig
```

### PartialExecutor

```go
func NewPartialExecutor(config VOPSConfig) *PartialExecutor
func (pe *PartialExecutor) Execute(tx *types.Transaction, state *PartialState, header *types.Header) (*ExecutionResult, error)
```

Errors: `ErrStateTooLarge`, `ErrMissingSender`, `ErrInsufficientBal`, `ErrNonceMismatch`

### VOPSValidator (complete stateless verification)

```go
func NewVOPSValidator() *VOPSValidator
func (v *VOPSValidator) AddWitness(stateRoot types.Hash, witness []byte) error
func (v *VOPSValidator) AddToAccessList(addr types.Address)
func (v *VOPSValidator) AddStorageProof(stateRoot types.Hash, proof [][]byte) error
func (v *VOPSValidator) ValidateBlock(blockData []byte, preRoot, postRoot types.Hash) (bool, error)
```

Errors: `ErrWitnessNotFound`, `ErrEmptyWitness`, `ErrEmptyBlock`

### Proof construction and verification

```go
// Build a proof from pre/post roots and the accessed key set.
func BuildValidityProof(preRoot, postRoot types.Hash, accessedKeys [][]byte) *ValidityProof

// Verify a proof (uses Keccak-256 binding commitment; production would use SNARK).
func ValidateTransition(preRoot, postRoot types.Hash, proof *ValidityProof) bool
```

### WitnessAccumulator

```go
func NewWitnessAccumulator() *WitnessAccumulator
func (wa *WitnessAccumulator) RecordAccess(key []byte)
func (wa *WitnessAccumulator) Keys() [][]byte
func (wa *WitnessAccumulator) BuildProof(pre, post types.Hash) *ValidityProof
```

## Usage

```go
config := vops.DefaultVOPSConfig()
executor := vops.NewPartialExecutor(config)

state := vops.NewPartialState()
state.SetAccount(sender, &vops.AccountState{Nonce: 0, Balance: big.NewInt(1e18)})

result, err := executor.Execute(tx, state, header)
if err != nil { ... }

proof := vops.BuildValidityProof(preRoot, result.PostState..., result.AccessedKeys)
valid := vops.ValidateTransition(preRoot, postRoot, proof)
```
