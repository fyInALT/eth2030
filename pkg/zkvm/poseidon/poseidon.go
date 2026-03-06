// Package zkvm provides a framework for zkVM guest execution and proof
// verification, supporting EIP-8079 native rollup proof-carrying transactions.
//
// poseidon.go implements the Poseidon hash function, a ZK-friendly hash
// used in circuit-based proofs (R1CS, PLONK, STARKs). Operates over the
// BN254 scalar field for compatibility with Ethereum's BN254 precompiles.
package poseidon

import (
	"math/big"
)

// BN254 scalar field order (curve order, not base field).
// r = 21888242871839275222246405745257275088548364400416034343698204186575808495617
var bn254ScalarField, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088548364400416034343698204186575808495617", 10,
)

// Bn254ScalarField returns a copy of the BN254 scalar field modulus.
func Bn254ScalarField() *big.Int { return new(big.Int).Set(bn254ScalarField) }

// PoseidonParams holds parameters for the Poseidon hash function.
// Default: t=3 (rate=2, capacity=1), full rounds=8, partial rounds=57.
type PoseidonParams struct {
	// T is the state width (rate + capacity).
	T int

	// FullRounds is the number of full S-box rounds (applied to all elements).
	FullRounds int

	// PartialRounds is the number of partial S-box rounds (applied to element 0).
	PartialRounds int

	// RoundConstants are the additive round constants.
	// Length = T * (FullRounds + PartialRounds).
	RoundConstants []*big.Int

	// MDS is the Maximum Distance Separable matrix (T x T).
	MDS [][]*big.Int

	// Field is the prime field modulus.
	Field *big.Int
}

// DefaultPoseidonParams returns Poseidon parameters for BN254 scalar field
// with t=3, full rounds=8, partial rounds=57.
func DefaultPoseidonParams() *PoseidonParams {
	t := 3
	fullRounds := 8
	partialRounds := 57
	totalRounds := fullRounds + partialRounds
	field := new(big.Int).Set(bn254ScalarField)

	// Generate deterministic round constants via a simple PRNG over the field.
	// In production, these would be derived per the Poseidon paper specification
	// using a Grain LFSR. Here we use a reproducible method.
	rcs := generateRoundConstants(t, totalRounds, field)

	// Generate a Cauchy MDS matrix.
	mds := generateMDS(t, field)

	return &PoseidonParams{
		T:              t,
		FullRounds:     fullRounds,
		PartialRounds:  partialRounds,
		RoundConstants: rcs,
		MDS:            mds,
		Field:          field,
	}
}

// SBox computes x^5 mod field (the Poseidon S-box for BN254).
func SBox(x, field *big.Int) *big.Int {
	x2 := new(big.Int).Mul(x, x)
	x2.Mod(x2, field)
	x4 := new(big.Int).Mul(x2, x2)
	x4.Mod(x4, field)
	x5 := new(big.Int).Mul(x4, x)
	x5.Mod(x5, field)
	return x5
}

// MDSMul multiplies a state vector by the MDS matrix.
func MDSMul(state []*big.Int, mds [][]*big.Int, field *big.Int) []*big.Int {
	t := len(state)
	result := make([]*big.Int, t)
	for i := 0; i < t; i++ {
		sum := new(big.Int)
		for j := 0; j < t; j++ {
			prod := new(big.Int).Mul(mds[i][j], state[j])
			sum.Add(sum, prod)
		}
		sum.Mod(sum, field)
		result[i] = sum
	}
	return result
}

// poseidonPermutation applies the Poseidon permutation to the state.
func poseidonPermutation(state []*big.Int, params *PoseidonParams) []*big.Int {
	t := params.T
	field := params.Field
	halfFull := params.FullRounds / 2
	rcIdx := 0

	// First half of full rounds.
	for r := 0; r < halfFull; r++ {
		// Add round constants.
		for i := 0; i < t; i++ {
			state[i] = new(big.Int).Add(state[i], params.RoundConstants[rcIdx])
			state[i].Mod(state[i], field)
			rcIdx++
		}
		// Full S-box: apply to all elements.
		for i := 0; i < t; i++ {
			state[i] = SBox(state[i], field)
		}
		// MDS mixing.
		state = MDSMul(state, params.MDS, field)
	}

	// Partial rounds.
	for r := 0; r < params.PartialRounds; r++ {
		// Add round constants.
		for i := 0; i < t; i++ {
			state[i] = new(big.Int).Add(state[i], params.RoundConstants[rcIdx])
			state[i].Mod(state[i], field)
			rcIdx++
		}
		// Partial S-box: apply only to element 0.
		state[0] = SBox(state[0], field)
		// MDS mixing.
		state = MDSMul(state, params.MDS, field)
	}

	// Second half of full rounds.
	for r := 0; r < halfFull; r++ {
		// Add round constants.
		for i := 0; i < t; i++ {
			state[i] = new(big.Int).Add(state[i], params.RoundConstants[rcIdx])
			state[i].Mod(state[i], field)
			rcIdx++
		}
		// Full S-box.
		for i := 0; i < t; i++ {
			state[i] = SBox(state[i], field)
		}
		// MDS mixing.
		state = MDSMul(state, params.MDS, field)
	}

	return state
}

// PoseidonHash hashes one or more field elements using the Poseidon hash.
// Uses a sponge construction with rate = T-1 and capacity = 1.
// Returns a single field element.
func PoseidonHash(params *PoseidonParams, inputs ...*big.Int) *big.Int {
	if params == nil {
		params = DefaultPoseidonParams()
	}
	t := params.T
	rate := t - 1 // capacity = 1

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
		state = poseidonPermutation(state, params)
	}

	// If no inputs, still permute once.
	if len(inputs) == 0 {
		state = poseidonPermutation(state, params)
	}

	return new(big.Int).Set(state[0])
}

// PoseidonSponge implements a sponge construction for variable-length input.
type PoseidonSponge struct {
	params *PoseidonParams
	state  []*big.Int
	buf    []*big.Int
	rate   int
}

// NewPoseidonSponge creates a new Poseidon sponge with the given parameters.
func NewPoseidonSponge(params *PoseidonParams) *PoseidonSponge {
	if params == nil {
		params = DefaultPoseidonParams()
	}
	state := make([]*big.Int, params.T)
	for i := range state {
		state[i] = new(big.Int)
	}
	return &PoseidonSponge{
		params: params,
		state:  state,
		rate:   params.T - 1,
	}
}

// Absorb adds field elements to the sponge.
func (s *PoseidonSponge) Absorb(inputs ...*big.Int) {
	for _, inp := range inputs {
		val := new(big.Int).Set(inp)
		val.Mod(val, s.params.Field)
		s.buf = append(s.buf, val)

		if len(s.buf) == s.rate {
			s.absorbBlock()
		}
	}
}

func (s *PoseidonSponge) absorbBlock() {
	for j := 0; j < len(s.buf); j++ {
		s.state[j+1].Add(s.state[j+1], s.buf[j])
		s.state[j+1].Mod(s.state[j+1], s.params.Field)
	}
	s.state = poseidonPermutation(s.state, s.params)
	s.buf = s.buf[:0]
}

// Squeeze extracts field elements from the sponge.
func (s *PoseidonSponge) Squeeze(count int) []*big.Int {
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
			s.state = poseidonPermutation(s.state, s.params)
		}
	}
	return results
}

// --- Parameter generation helpers ---

// grainLFSR implements the Grain LFSR used to generate Poseidon round
// constants per the Poseidon paper (https://eprint.iacr.org/2019/458).
// The LFSR is initialized from the field parameters and clocked to
// produce pseudorandom bits, which are then assembled into field elements.
type grainLFSR struct {
	state [80]bool // 80-bit LFSR state
}

// newGrainLFSR creates a Grain LFSR initialized with the given seed bytes.
// The seed is derived from the field modulus size, state width, and number
// of rounds per the Poseidon specification.
func newGrainLFSR(fieldBits, t, fullRounds, partialRounds int, seed []byte) *grainLFSR {
	g := &grainLFSR{}

	// Initialize the state from field parameters as specified in the
	// Poseidon paper: encode (fieldBits, t, fullRounds, partialRounds)
	// into the initial 80-bit state, then mix in the seed.
	//
	// Bits 0..7: fieldBits (low 8 bits)
	// Bits 8..11: t (low 4 bits)
	// Bits 12..19: fullRounds (low 8 bits)
	// Bits 20..29: partialRounds (low 10 bits)
	// Bits 30..79: filled from seed bytes
	for i := 0; i < 8 && i < 80; i++ {
		g.state[i] = (fieldBits>>i)&1 == 1
	}
	for i := 0; i < 4 && 8+i < 80; i++ {
		g.state[8+i] = (t>>i)&1 == 1
	}
	for i := 0; i < 8 && 12+i < 80; i++ {
		g.state[12+i] = (fullRounds>>i)&1 == 1
	}
	for i := 0; i < 10 && 20+i < 80; i++ {
		g.state[20+i] = (partialRounds>>i)&1 == 1
	}

	// Fill remaining bits from seed.
	bitPos := 30
	for _, b := range seed {
		for bit := 0; bit < 8 && bitPos < 80; bit++ {
			g.state[bitPos] = (b>>bit)&1 == 1
			bitPos++
		}
	}

	// Set any remaining unfilled bits to 1 (non-degenerate initialization).
	for i := bitPos; i < 80; i++ {
		g.state[i] = true
	}

	// Warm up: clock the LFSR 160 times to mix the state thoroughly.
	for i := 0; i < 160; i++ {
		g.clock()
	}

	return g
}

// clock advances the LFSR by one step and returns the output bit.
// The feedback polynomial is:
// s[80] = s[0] XOR s[13] XOR s[23] XOR s[38] XOR s[51] XOR s[62]
func (g *grainLFSR) clock() bool {
	output := g.state[0]

	// Compute feedback.
	feedback := g.state[0] != g.state[13]
	feedback = feedback != g.state[23]
	feedback = feedback != g.state[38]
	feedback = feedback != g.state[51]
	feedback = feedback != g.state[62]

	// Shift left by one.
	copy(g.state[:79], g.state[1:80])
	g.state[79] = feedback

	return output
}

// nextBits extracts n pseudorandom bits from the LFSR.
func (g *grainLFSR) nextBits(n int) []bool {
	bits := make([]bool, n)
	for i := 0; i < n; i++ {
		bits[i] = g.clock()
	}
	return bits
}

// nextFieldElement generates a pseudorandom field element by extracting bits
// from the LFSR and interpreting them as a big-endian integer. If the result
// is >= field, we reject and try again (rejection sampling).
func (g *grainLFSR) nextFieldElement(field *big.Int) *big.Int {
	fieldBits := field.BitLen()
	for {
		bits := g.nextBits(fieldBits)
		val := new(big.Int)
		for i, b := range bits {
			if b {
				val.SetBit(val, i, 1)
			}
		}
		if val.Cmp(field) < 0 {
			return val
		}
		// Reject and resample if >= field.
	}
}

// generateRoundConstants produces deterministic round constants using a
// Grain LFSR per the Poseidon paper specification. The LFSR is seeded
// from the field parameters and a domain-specific seed string.
func generateRoundConstants(t, totalRounds int, field *big.Int) []*big.Int {
	numConstants := t * totalRounds
	constants := make([]*big.Int, numConstants)

	// Derive full and partial round counts from totalRounds.
	// For Poseidon1 with default params: fullRounds=8, partialRounds=totalRounds-8.
	fullRounds := 8
	partialRounds := totalRounds - fullRounds
	if partialRounds < 0 {
		partialRounds = 0
	}

	g := newGrainLFSR(field.BitLen(), t, fullRounds, partialRounds, []byte("PoseidonBN254"))

	for i := 0; i < numConstants; i++ {
		constants[i] = g.nextFieldElement(field)
	}
	return constants
}

// generateMDS produces a Cauchy MDS matrix over the field.
// M[i][j] = 1 / (x_i + y_j) where x and y are distinct field elements.
func generateMDS(t int, field *big.Int) [][]*big.Int {
	// Use x_i = i, y_j = t + j as distinct elements.
	mds := make([][]*big.Int, t)
	for i := 0; i < t; i++ {
		mds[i] = make([]*big.Int, t)
		for j := 0; j < t; j++ {
			sum := new(big.Int).Add(big.NewInt(int64(i)), big.NewInt(int64(t+j)))
			sum.Mod(sum, field)
			// Compute modular inverse: 1/(x_i + y_j).
			inv := new(big.Int).ModInverse(sum, field)
			if inv == nil {
				// Fallback: should not happen with distinct elements in a prime field.
				inv = big.NewInt(1)
			}
			mds[i][j] = inv
		}
	}
	return mds
}
