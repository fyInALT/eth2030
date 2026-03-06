package zkvm

// poseidon_compat.go re-exports types from zkvm/poseidon for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/zkvm/poseidon"
)

// bn254ScalarField is used by circuit_builder.go and constraint_compiler.go
// in this package. Re-initialized here since the original was in poseidon.go.
var bn254ScalarField = poseidon.Bn254ScalarField()

// Poseidon type aliases.
type (
	PoseidonParams  = poseidon.PoseidonParams
	PoseidonSponge  = poseidon.PoseidonSponge
	Poseidon2Params = poseidon.Poseidon2Params
	Poseidon2Sponge = poseidon.Poseidon2Sponge
)

// Poseidon function wrappers.
func DefaultPoseidonParams() *PoseidonParams  { return poseidon.DefaultPoseidonParams() }
func SBox(x, field *big.Int) *big.Int         { return poseidon.SBox(x, field) }
func MDSMul(state []*big.Int, mds [][]*big.Int, field *big.Int) []*big.Int {
	return poseidon.MDSMul(state, mds, field)
}
func PoseidonHash(params *PoseidonParams, inputs ...*big.Int) *big.Int {
	return poseidon.PoseidonHash(params, inputs...)
}
func NewPoseidonSponge(params *PoseidonParams) *PoseidonSponge {
	return poseidon.NewPoseidonSponge(params)
}
func DefaultPoseidon2Params() *Poseidon2Params { return poseidon.DefaultPoseidon2Params() }
func Poseidon2Hash(params *Poseidon2Params, inputs ...*big.Int) *big.Int {
	return poseidon.Poseidon2Hash(params, inputs...)
}
func NewPoseidon2Sponge(params *Poseidon2Params) *Poseidon2Sponge {
	return poseidon.NewPoseidon2Sponge(params)
}
func Poseidon2HashBytes(data []byte) [32]byte { return poseidon.Poseidon2HashBytes(data) }
