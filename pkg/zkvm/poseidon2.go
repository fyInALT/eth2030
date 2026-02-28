// poseidon2.go implements the Poseidon2 hash function, an improved variant
// of the Poseidon hash designed for better ZK circuit efficiency. Poseidon2
// differs from Poseidon1 by using diagonal MDS matrices (cheaper in circuits)
// and splitting rounds into external (full S-box) and internal (partial S-box)
// categories. Operates over the BN254 scalar field.
package zkvm

import (
	"encoding/binary"
	"math/big"
)

// Poseidon2Params holds parameters for the Poseidon2 hash function.
type Poseidon2Params struct {
	// T is the state width (rate + capacity).
	T int

	// ExternalRounds is the number of external (full S-box) rounds.
	ExternalRounds int

	// InternalRounds is the number of internal (partial S-box on element 0) rounds.
	InternalRounds int

	// RoundConstants are the additive round constants.
	// Length = T * ExternalRounds + InternalRounds (one per internal round for element 0).
	RoundConstants []*big.Int

	// DiagMDS holds the diagonal MDS matrix entries (T elements).
	// The linear layer computes y[i] = state[i] + diag[i] * sum(state).
	DiagMDS []*big.Int

	// Field is the prime field modulus.
	Field *big.Int
}

// DefaultPoseidon2Params returns Poseidon2 parameters for BN254 scalar field
// with t=3, external rounds=8, internal rounds=56.
func DefaultPoseidon2Params() *Poseidon2Params {
	t := 3
	externalRounds := 8
	internalRounds := 56
	field := new(big.Int).Set(bn254ScalarField)

	// Total round constants: T constants per external round + 1 per internal round.
	totalRC := t*externalRounds + internalRounds
	rcs := generatePoseidon2RoundConstants(t, totalRC, field)
	diag := generateDiagonalMDS(t, field)

	return &Poseidon2Params{
		T:              t,
		ExternalRounds: externalRounds,
		InternalRounds: internalRounds,
		RoundConstants: rcs,
		DiagMDS:        diag,
		Field:          field,
	}
}

// generatePoseidon2RoundConstants produces deterministic round constants
// for Poseidon2 using a Grain LFSR with a different seed than Poseidon1.
// This follows the same Grain LFSR specification from the Poseidon paper.
func generatePoseidon2RoundConstants(t, totalConstants int, field *big.Int) []*big.Int {
	constants := make([]*big.Int, totalConstants)

	// For Poseidon2: externalRounds=8, internalRounds derived from totalConstants.
	// totalConstants = T * externalRounds + internalRounds
	externalRounds := 8
	internalRounds := totalConstants - t*externalRounds
	if internalRounds < 0 {
		internalRounds = 0
	}

	g := newGrainLFSR(field.BitLen(), t, externalRounds, internalRounds, []byte("Poseidon2BN254"))

	for i := 0; i < totalConstants; i++ {
		constants[i] = g.nextFieldElement(field)
	}
	return constants
}

// generateDiagonalMDS produces a diagonal MDS matrix for Poseidon2.
// Each entry is d_i = 1 + (i+1)^2 mod field. The linear layer uses:
// y[i] = state[i] + diag[i] * sum(state).
func generateDiagonalMDS(t int, field *big.Int) []*big.Int {
	diag := make([]*big.Int, t)
	for i := 0; i < t; i++ {
		// d_i = (i+1)^2
		iPlusOne := big.NewInt(int64(i + 1))
		sq := new(big.Int).Mul(iPlusOne, iPlusOne)
		// 1 + d_i
		val := new(big.Int).Add(big.NewInt(1), sq)
		val.Mod(val, field)
		diag[i] = val
	}
	return diag
}

// poseidon2ExternalLinear applies the external round linear layer.
// Computes: sum = sum(state), then y[i] = state[i] + diag[i] * sum.
func poseidon2ExternalLinear(state []*big.Int, diag []*big.Int, field *big.Int) []*big.Int {
	t := len(state)
	sum := new(big.Int)
	for i := 0; i < t; i++ {
		sum.Add(sum, state[i])
	}
	sum.Mod(sum, field)

	result := make([]*big.Int, t)
	for i := 0; i < t; i++ {
		// diag[i] * sum
		prod := new(big.Int).Mul(diag[i], sum)
		prod.Mod(prod, field)
		// state[i] + prod
		result[i] = new(big.Int).Add(state[i], prod)
		result[i].Mod(result[i], field)
	}
	return result
}

// poseidon2InternalLinear applies the internal round linear layer.
// Uses the same diagonal structure as the external layer for simplicity,
// consistent with the simplified Poseidon2 specification.
func poseidon2InternalLinear(state []*big.Int, diag []*big.Int, field *big.Int) []*big.Int {
	return poseidon2ExternalLinear(state, diag, field)
}

// poseidon2Permutation applies the full Poseidon2 permutation to the state.
//
// Structure:
//  1. First half of external rounds: add T constants -> full S-box -> external linear
//  2. Internal rounds: add 1 constant (element 0) -> partial S-box -> internal linear
//  3. Second half of external rounds: add T constants -> full S-box -> external linear
func poseidon2Permutation(state []*big.Int, params *Poseidon2Params) []*big.Int {
	t := params.T
	field := params.Field
	halfExternal := params.ExternalRounds / 2
	rcIdx := 0

	// First half of external rounds.
	for r := 0; r < halfExternal; r++ {
		// Add round constants (T per external round).
		for i := 0; i < t; i++ {
			state[i] = new(big.Int).Add(state[i], params.RoundConstants[rcIdx])
			state[i].Mod(state[i], field)
			rcIdx++
		}
		// Full S-box: apply to all elements.
		for i := 0; i < t; i++ {
			state[i] = SBox(state[i], field)
		}
		// External linear layer.
		state = poseidon2ExternalLinear(state, params.DiagMDS, field)
	}

	// Internal rounds.
	for r := 0; r < params.InternalRounds; r++ {
		// Add round constant only to element 0.
		state[0] = new(big.Int).Add(state[0], params.RoundConstants[rcIdx])
		state[0].Mod(state[0], field)
		rcIdx++
		// Partial S-box: apply only to element 0.
		state[0] = SBox(state[0], field)
		// Internal linear layer.
		state = poseidon2InternalLinear(state, params.DiagMDS, field)
	}

	// Second half of external rounds.
	for r := 0; r < halfExternal; r++ {
		// Add round constants (T per external round).
		for i := 0; i < t; i++ {
			state[i] = new(big.Int).Add(state[i], params.RoundConstants[rcIdx])
			state[i].Mod(state[i], field)
			rcIdx++
		}
		// Full S-box.
		for i := 0; i < t; i++ {
			state[i] = SBox(state[i], field)
		}
		// External linear layer.
		state = poseidon2ExternalLinear(state, params.DiagMDS, field)
	}

	return state
}

// Poseidon2Hash hashes one or more field elements using the Poseidon2 hash.
// Uses a sponge construction with rate = T-1 and capacity = 1.
// Returns a single field element.
func Poseidon2Hash(params *Poseidon2Params, inputs ...*big.Int) *big.Int {
	if params == nil {
		params = DefaultPoseidon2Params()
	}
	t := params.T
	rate := t - 1

	// Initialize state to zeros.
	state := make([]*big.Int, t)
	for i := range state {
		state[i] = new(big.Int)
	}

	// Absorb inputs into rate portion of state.
	for i := 0; i < len(inputs); i += rate {
		for j := 0; j < rate && i+j < len(inputs); j++ {
			val := new(big.Int).Set(inputs[i+j])
			val.Mod(val, params.Field)
			state[j+1].Add(state[j+1], val)
			state[j+1].Mod(state[j+1], params.Field)
		}
		state = poseidon2Permutation(state, params)
	}

	// If no inputs, still permute once.
	if len(inputs) == 0 {
		state = poseidon2Permutation(state, params)
	}

	return new(big.Int).Set(state[0])
}

// Poseidon2Sponge implements a sponge construction for variable-length input
// using the Poseidon2 permutation.
type Poseidon2Sponge struct {
	params *Poseidon2Params
	state  []*big.Int
	buf    []*big.Int
	rate   int
}

// NewPoseidon2Sponge creates a new Poseidon2 sponge with the given parameters.
func NewPoseidon2Sponge(params *Poseidon2Params) *Poseidon2Sponge {
	if params == nil {
		params = DefaultPoseidon2Params()
	}
	state := make([]*big.Int, params.T)
	for i := range state {
		state[i] = new(big.Int)
	}
	return &Poseidon2Sponge{
		params: params,
		state:  state,
		rate:   params.T - 1,
	}
}

// Absorb adds field elements to the sponge.
func (s *Poseidon2Sponge) Absorb(inputs ...*big.Int) {
	for _, inp := range inputs {
		val := new(big.Int).Set(inp)
		val.Mod(val, s.params.Field)
		s.buf = append(s.buf, val)

		if len(s.buf) == s.rate {
			s.absorbBlock()
		}
	}
}

func (s *Poseidon2Sponge) absorbBlock() {
	for j := 0; j < len(s.buf); j++ {
		s.state[j+1].Add(s.state[j+1], s.buf[j])
		s.state[j+1].Mod(s.state[j+1], s.params.Field)
	}
	s.state = poseidon2Permutation(s.state, s.params)
	s.buf = s.buf[:0]
}

// Squeeze extracts field elements from the sponge.
func (s *Poseidon2Sponge) Squeeze(count int) []*big.Int {
	// Flush remaining buffer.
	if len(s.buf) > 0 {
		s.absorbBlock()
	}

	results := make([]*big.Int, 0, count)
	for len(results) < count {
		// Extract from rate portion.
		for j := 1; j <= s.rate && len(results) < count; j++ {
			results = append(results, new(big.Int).Set(s.state[j]))
		}
		if len(results) < count {
			s.state = poseidon2Permutation(s.state, s.params)
		}
	}
	return results
}

// Poseidon2HashBytes is a convenience function that converts arbitrary bytes
// to field elements (8 bytes per element), hashes with default Poseidon2
// params, and returns a 32-byte result.
func Poseidon2HashBytes(data []byte) [32]byte {
	params := DefaultPoseidon2Params()

	// Convert bytes to field elements, 8 bytes per element.
	var elements []*big.Int
	for i := 0; i < len(data); i += 8 {
		chunk := make([]byte, 8)
		end := i + 8
		if end > len(data) {
			end = len(data)
		}
		copy(chunk, data[i:end])
		elem := new(big.Int).SetUint64(binary.LittleEndian.Uint64(chunk))
		elements = append(elements, elem)
	}

	// If no elements, hash empty.
	h := Poseidon2Hash(params, elements...)

	// Convert result to 32-byte array.
	var result [32]byte
	hBytes := h.Bytes()
	// Right-align in 32 bytes.
	if len(hBytes) > 32 {
		hBytes = hBytes[len(hBytes)-32:]
	}
	copy(result[32-len(hBytes):], hBytes)
	return result
}
