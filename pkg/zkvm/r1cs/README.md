# zkvm/r1cs — R1CS constraint system solver

## Overview

`zkvm/r1cs` implements a Rank-1 Constraint System (R1CS) solver used by the zkVM framework. R1CS is the standard arithmetic constraint representation consumed by ZK-SNARK backends such as Groth16 and PLONK. The package provides sparse constraint encoding, forward-propagation witness solving, and full constraint satisfaction verification — supporting the K+/M+ roadmap items for canonical guest verification and mandatory 3-of-5 proofs.

Constraints are stored sparsely as `(coefficient, variable_index)` pairs. Verification supports both plain `int64` arithmetic and modular `*big.Int` arithmetic over a configurable prime field (BN254 scalar field when used with the Poseidon/zkVM pipeline).

## Functionality

**Types**

- `SparseTerm` — a `(Index int, Coefficient int64)` pair in a linear combination.
- `SparseConstraint` — one R1CS constraint `<A,w> * <B,w> = <C,w>` expressed as three `[]SparseTerm` vectors.
- `R1CSSystem` — the constraint system; holds `Constraints`, `NumVariables`, `NumPublic`, and optional `field *big.Int`.
- `R1CSStats` — summary statistics: constraint count, variable count, public input count, private wire count, total term count.

**Construction**

| Function | Description |
|---|---|
| `NewR1CSSystem(numVars, numPublic int)` | Integer arithmetic (no field reduction). |
| `NewR1CSSystemWithField(numVars, numPublic int, field *big.Int)` | Modular arithmetic over `field`. |

**Building constraints**

| Method | Description |
|---|---|
| `AddConstraint(a, b, c []SparseTerm)` | Raw R1CS constraint. |
| `AddMultiplicationGate(left, right, output int)` | `w[left] * w[right] = w[output]`. |
| `AddAdditionGate(a, b, result int)` | `w[a] + w[b] = w[result]`. |
| `AddConstantGate(variable int, value int64)` | `w[variable] = value`. |

**Solving and verification**

| Method | Description |
|---|---|
| `Solve(publicInputs []int64) ([]int64, error)` | Forward-propagate from public inputs to compute a full witness. |
| `Verify(witness []int64) bool` | Check all constraints are satisfied. |
| `EvalLinearCombination(terms, witness) int64` | Evaluate a sparse linear combination. |
| `Stats() R1CSStats` | Aggregate statistics. |

Variable 0 is always the constant `1`. Public inputs occupy indices `1..NumPublic`.

## Usage

```go
// 4 variables (0=const, 1=pub, 2=priv, 3=out), 1 public input.
sys, _ := r1cs.NewR1CSSystem(4, 1)

// Encode: w[1] * w[2] = w[3]
_ = sys.AddMultiplicationGate(1, 2, 3)

// Solve given public input w[1]=6; solver finds w[2] and w[3] if constrained.
witness, err := sys.Solve([]int64{6})

// Or verify a complete witness directly.
ok := sys.Verify([]int64{1, 6, 7, 42})
```

---

Parent package: [`zkvm`](../)
