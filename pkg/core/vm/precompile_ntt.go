package vm

import (
	"errors"
	"fmt"
	"math/big"
	"math/bits"

	"github.com/eth2030/eth2030/core/types"
)

// NTT precompile over BN254 scalar field.
// Split precompiles at addresses 0x0f-0x14 (registered for I+ fork).
//
// EIP-7885 NTT Precompile Alignment
//
// This precompile implements the Number Theoretic Transform (NTT) as proposed
// in EIP-7885 by ZKNoxHQ. NTT is the core operation for:
//   - Lattice-based cryptography (Falcon, Dilithium, Kyber)
//   - STARK proof verification (polynomial evaluation over finite fields)
//   - Efficient polynomial multiplication in ZK circuits
//
// Reference implementations:
//   - ZKNoxHQ/NTT (Solidity + Python): refs/ntt-eip/
//   - ZKNoxHQ/ETHFALCON: refs/ethfalcon/
//
// Supported fields:
//   - BN254 scalar field (existing): p = 21888242871839275222246405745257275088548364400416034343698204186575808495617
//   - Goldilocks field (new): p = 2^64 - 2^32 + 1 = 18446744069414584321 (STARK-friendly)

// BN254 scalar field modulus.
var bn254ScalarField, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

// Known primitive root of the BN254 scalar field (a generator of the multiplicative group).
// 5 is a primitive root mod bn254ScalarField.
var bn254PrimitiveRoot = big.NewInt(5)

// Goldilocks field modulus: p = 2^64 - 2^32 + 1
// Used by STARK-friendly systems (Plonky2, Polygon zkEVM).
var goldilocksField = new(big.Int).SetUint64(18446744069414584321)

// Goldilocks primitive root: 7 is a primitive root mod p.
var goldilocksPrimitiveRoot = big.NewInt(7)

// NTT operation types.
const (
	NTTOpForward           = 0 // Forward NTT (evaluation) over BN254
	NTTOpInverse           = 1 // Inverse NTT (interpolation) over BN254
	NTTOpGoldilocks        = 2 // Forward NTT over Goldilocks field
	NTTOpGoldilocksInverse = 3 // Inverse NTT over Goldilocks field
)

// NTT gas cost constants.
const (
	NTTBaseCost       uint64 = 1000
	NTTPerElementCost uint64 = 10
	NTTMaxDegree             = 1 << 16 // 65536
)

// NTT errors.
var (
	ErrNTTInvalidInput  = errors.New("ntt: invalid input")
	ErrNTTNotPowerOfTwo = errors.New("ntt: size must be a power of two")
	ErrNTTTooLarge      = errors.New("ntt: exceeds maximum degree")
	ErrNTTInvalidOpType = errors.New("ntt: invalid operation type")
	ErrNTTNoRootOfUnity = errors.New("ntt: no root of unity for given size")
	ErrNTTZeroModulus   = errors.New("ntt: modulus must be non-zero")
	ErrNTTOutputRange   = errors.New("ntt: output element out of field range")
)

type nttPrecompile struct{}

func (c *nttPrecompile) RequiredGas(input []byte) uint64 {
	if len(input) < 1 {
		return 0
	}
	nBytes := len(input) - 1
	n := uint64(nBytes / 32)
	if n == 0 {
		return NTTBaseCost
	}
	// gas = base + n * log2(n) * perElement
	log2n := uint64(bits.Len(uint(n)) - 1)
	if log2n == 0 {
		log2n = 1
	}
	// Overflow check: compute n * log2n first, then multiply by perElement.
	product := n * log2n
	if n != 0 && product/n != log2n {
		return ^uint64(0) // return max gas on overflow
	}
	gasAdditional := product * NTTPerElementCost
	if product != 0 && gasAdditional/product != NTTPerElementCost {
		return ^uint64(0)
	}
	total := NTTBaseCost + gasAdditional
	if total < NTTBaseCost {
		return ^uint64(0)
	}
	return total
}

func (c *nttPrecompile) Run(input []byte) ([]byte, error) {
	if len(input) < 1 {
		return nil, ErrNTTInvalidInput
	}

	opType := input[0]
	if opType > 3 {
		return nil, ErrNTTInvalidOpType
	}

	coeffData := input[1:]
	if len(coeffData)%32 != 0 {
		return nil, ErrNTTInvalidInput
	}
	n := len(coeffData) / 32
	if n == 0 {
		return nil, ErrNTTInvalidInput
	}
	if n > NTTMaxDegree {
		return nil, ErrNTTTooLarge
	}
	if n&(n-1) != 0 {
		return nil, ErrNTTNotPowerOfTwo
	}

	// Select field parameters based on operation type.
	var fieldMod, primRoot *big.Int
	switch {
	case opType <= NTTOpInverse:
		fieldMod = bn254ScalarField
		primRoot = bn254PrimitiveRoot
	default:
		fieldMod = goldilocksField
		primRoot = goldilocksPrimitiveRoot
	}

	// Parse coefficients as big-endian 32-byte big.Ints, reduced mod p.
	coeffs := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		val := new(big.Int).SetBytes(coeffData[i*32 : (i+1)*32])
		val.Mod(val, fieldMod)
		coeffs[i] = val
	}

	inverse := opType == NTTOpInverse || opType == NTTOpGoldilocksInverse
	result, err := computeNTT(coeffs, fieldMod, primRoot, inverse)
	if err != nil {
		return nil, err
	}

	// Validate output field elements are within the field.
	for i, val := range result {
		if val.Sign() < 0 || val.Cmp(fieldMod) >= 0 {
			return nil, fmt.Errorf("%w: element %d = %s", ErrNTTOutputRange, i, val.String())
		}
	}

	// Encode output: n * 32-byte big-endian values.
	out := make([]byte, n*32)
	for i, val := range result {
		b := val.Bytes()
		copy(out[i*32+(32-len(b)):], b)
	}
	return out, nil
}

// computeNTT performs a full NTT (forward or inverse) over the given field.
func computeNTT(coefficients []*big.Int, fieldMod, primitiveRoot *big.Int, inverse bool) ([]*big.Int, error) {
	n := len(coefficients)

	omega, err := findRootOfUnityWithRoot(n, fieldMod, primitiveRoot)
	if err != nil {
		return nil, err
	}

	if inverse {
		return nttInverse(coefficients, omega, fieldMod), nil
	}
	return nttForward(coefficients, omega, fieldMod), nil
}

// findRootOfUnity finds a primitive n-th root of unity mod p using the BN254
// primitive root. This is a backward-compatible wrapper.
func findRootOfUnity(n int, p *big.Int) (*big.Int, error) {
	return findRootOfUnityWithRoot(n, p, bn254PrimitiveRoot)
}

// findRootOfUnityWithRoot finds a primitive n-th root of unity mod p using the
// given primitive root g. omega = g^((p-1)/n) where g is a primitive root.
func findRootOfUnityWithRoot(n int, p *big.Int, g *big.Int) (*big.Int, error) {
	if p == nil || p.Sign() == 0 {
		return nil, ErrNTTZeroModulus
	}
	if n <= 0 || n&(n-1) != 0 {
		return nil, ErrNTTNotPowerOfTwo
	}

	// Check that n divides p-1.
	pMinus1 := new(big.Int).Sub(p, big.NewInt(1))
	nBig := big.NewInt(int64(n))

	if new(big.Int).Mod(pMinus1, nBig).Sign() != 0 {
		return nil, ErrNTTNoRootOfUnity
	}

	// omega = g^((p-1)/n) mod p
	exp := new(big.Int).Div(pMinus1, nBig)
	omega := new(big.Int).Exp(g, exp, p)

	// Verify: omega^n == 1 mod p
	check := new(big.Int).Exp(omega, nBig, p)
	if check.Cmp(big.NewInt(1)) != 0 {
		return nil, ErrNTTNoRootOfUnity
	}

	return omega, nil
}

// nttForward performs the forward NTT using the Cooley-Tukey butterfly algorithm.
func nttForward(coeffs []*big.Int, omega *big.Int, p *big.Int) []*big.Int {
	n := len(coeffs)
	if n == 1 {
		result := make([]*big.Int, 1)
		result[0] = new(big.Int).Set(coeffs[0])
		return result
	}

	// Bit-reversal permutation.
	result := make([]*big.Int, n)
	logN := bits.Len(uint(n)) - 1
	for i := range result {
		result[i] = new(big.Int).Set(coeffs[bitReverse(i, logN)])
	}

	// Cooley-Tukey butterfly.
	for size := 2; size <= n; size *= 2 {
		halfSize := size / 2
		step := new(big.Int).Exp(omega, big.NewInt(int64(n/size)), p)
		w := big.NewInt(1)
		for j := 0; j < halfSize; j++ {
			for k := j; k < n; k += size {
				u := new(big.Int).Set(result[k])
				v := new(big.Int).Mul(result[k+halfSize], w)
				v.Mod(v, p)
				result[k] = new(big.Int).Add(u, v)
				result[k].Mod(result[k], p)
				result[k+halfSize] = new(big.Int).Sub(u, v)
				result[k+halfSize].Mod(result[k+halfSize], p)
				if result[k+halfSize].Sign() < 0 {
					result[k+halfSize].Add(result[k+halfSize], p)
				}
			}
			w = new(big.Int).Mul(w, step)
			w.Mod(w, p)
		}
	}

	return result
}

// nttInverse performs the inverse NTT.
func nttInverse(evals []*big.Int, omega *big.Int, p *big.Int) []*big.Int {
	// Compute inverse of omega: omega^(-1) = omega^(p-2) mod p (Fermat's little theorem).
	omegaInv := new(big.Int).Exp(omega, new(big.Int).Sub(p, big.NewInt(2)), p)

	result := nttForward(evals, omegaInv, p)

	// Divide each element by n.
	n := len(evals)
	nBig := big.NewInt(int64(n))
	nInv := new(big.Int).Exp(nBig, new(big.Int).Sub(p, big.NewInt(2)), p)
	for i := range result {
		result[i].Mul(result[i], nInv)
		result[i].Mod(result[i], p)
	}

	return result
}

// bitReverse reverses the lower numBits bits of v.
func bitReverse(v, numBits int) int {
	result := 0
	for i := 0; i < numBits; i++ {
		result = (result << 1) | (v & 1)
		v >>= 1
	}
	return result
}

// ValidateNTTInput checks that raw NTT precompile input is well-formed before execution:
//   - Input must have at least 1 byte (op type) + 32 bytes (at least 1 element)
//   - Op type must be 0-3 (forward/inverse for BN254 or Goldilocks)
//   - Element count must be a power of two and at most NTTMaxDegree
func ValidateNTTInput(input []byte) error {
	if len(input) < 1 {
		return ErrNTTInvalidInput
	}
	if input[0] > 3 {
		return ErrNTTInvalidOpType
	}
	coeffData := input[1:]
	if len(coeffData) == 0 || len(coeffData)%32 != 0 {
		return ErrNTTInvalidInput
	}
	n := len(coeffData) / 32
	if n > NTTMaxDegree {
		return ErrNTTTooLarge
	}
	if n&(n-1) != 0 {
		return ErrNTTNotPowerOfTwo
	}
	return nil
}

// GoldilocksNTT performs NTT over the Goldilocks field.
// This is the STARK-friendly variant for use with FRI-based proof systems.
func GoldilocksNTT(coefficients []*big.Int) ([]*big.Int, error) {
	return computeNTT(coefficients, goldilocksField, goldilocksPrimitiveRoot, false)
}

// GoldilocksINTT performs inverse NTT over the Goldilocks field.
func GoldilocksINTT(coefficients []*big.Int) ([]*big.Int, error) {
	return computeNTT(coefficients, goldilocksField, goldilocksPrimitiveRoot, true)
}

// Split NTT precompile structs (EIP-7885 aligned addresses 0x0f-0x14).

// nttFWPrecompile performs forward NTT over BN254 scalar field.
// Address: 0x0f
type nttFWPrecompile struct{}

func (c *nttFWPrecompile) RequiredGas(input []byte) uint64 {
	n := uint64(len(input) / 32)
	if n == 0 {
		return 0
	}
	log2n := uint64(bits.Len(uint(n)) - 1)
	if log2n == 0 {
		log2n = 1
	}
	gas := n * log2n / 8
	if gas < 600 {
		return 600
	}
	return gas
}

func (c *nttFWPrecompile) Run(input []byte) ([]byte, error) {
	return runNTTSplit(input, false)
}

// nttINVPrecompile performs inverse NTT over BN254 scalar field.
// Address: 0x10
type nttINVPrecompile struct{}

func (c *nttINVPrecompile) RequiredGas(input []byte) uint64 {
	return (&nttFWPrecompile{}).RequiredGas(input)
}

func (c *nttINVPrecompile) Run(input []byte) ([]byte, error) {
	return runNTTSplit(input, true)
}

// runNTTSplit is the common implementation for forward/inverse NTT precompiles.
func runNTTSplit(input []byte, inverse bool) ([]byte, error) {
	if len(input) == 0 || len(input)%32 != 0 {
		return nil, ErrNTTInvalidInput
	}
	n := len(input) / 32
	if n == 0 {
		return nil, ErrNTTInvalidInput
	}
	if n > NTTMaxDegree {
		return nil, ErrNTTTooLarge
	}
	if n&(n-1) != 0 {
		return nil, ErrNTTNotPowerOfTwo
	}

	coeffs := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		val := new(big.Int).SetBytes(input[i*32 : (i+1)*32])
		val.Mod(val, bn254ScalarField)
		coeffs[i] = val
	}

	result, err := computeNTT(coeffs, bn254ScalarField, bn254PrimitiveRoot, inverse)
	if err != nil {
		return nil, err
	}

	out := make([]byte, n*32)
	for i, val := range result {
		b := val.Bytes()
		copy(out[i*32+(32-len(b)):], b)
	}
	return out, nil
}

// nttVecMulModPrecompile performs element-wise multiplication mod BN254.
// Address: 0x11
// Input: 2N 32-byte elements (A[0..N-1] || B[0..N-1]).
// Output: N 32-byte elements where result[i] = A[i]*B[i] mod BN254.
type nttVecMulModPrecompile struct{}

func (c *nttVecMulModPrecompile) RequiredGas(input []byte) uint64 {
	n := uint64(len(input) / 64) // half-input element count
	if n == 0 {
		return 0
	}
	log2n := uint64(bits.Len(uint(n)) - 1)
	if log2n == 0 {
		log2n = 1
	}
	gas := n * log2n / 8
	if gas < 600 {
		return 600
	}
	return gas
}

func (c *nttVecMulModPrecompile) Run(input []byte) ([]byte, error) {
	if len(input) == 0 || len(input)%32 != 0 {
		return nil, ErrNTTInvalidInput
	}
	totalElems := len(input) / 32
	if totalElems%2 != 0 {
		return nil, ErrNTTInvalidInput
	}
	n := totalElems / 2

	out := make([]byte, n*32)
	for i := 0; i < n; i++ {
		a := new(big.Int).SetBytes(input[i*32 : (i+1)*32])
		b := new(big.Int).SetBytes(input[(n+i)*32 : (n+i+1)*32])
		r := new(big.Int).Mul(a, b)
		r.Mod(r, bn254ScalarField)
		rb := r.Bytes()
		copy(out[i*32+(32-len(rb)):], rb)
	}
	return out, nil
}

// nttVecAddModPrecompile performs element-wise addition mod BN254.
// Address: 0x12
// Input: 2N 32-byte elements (A[0..N-1] || B[0..N-1]).
// Output: N 32-byte elements where result[i] = A[i]+B[i] mod BN254.
type nttVecAddModPrecompile struct{}

func (c *nttVecAddModPrecompile) RequiredGas(input []byte) uint64 {
	n := uint64(len(input) / 64)
	if n == 0 {
		return 0
	}
	log2n := uint64(bits.Len(uint(n)) - 1)
	if log2n == 0 {
		log2n = 1
	}
	gas := n * log2n / 32
	if gas < 100 {
		return 100
	}
	return gas
}

func (c *nttVecAddModPrecompile) Run(input []byte) ([]byte, error) {
	if len(input) == 0 || len(input)%32 != 0 {
		return nil, ErrNTTInvalidInput
	}
	totalElems := len(input) / 32
	if totalElems%2 != 0 {
		return nil, ErrNTTInvalidInput
	}
	n := totalElems / 2

	out := make([]byte, n*32)
	for i := 0; i < n; i++ {
		a := new(big.Int).SetBytes(input[i*32 : (i+1)*32])
		b := new(big.Int).SetBytes(input[(n+i)*32 : (n+i+1)*32])
		r := new(big.Int).Add(a, b)
		r.Mod(r, bn254ScalarField)
		rb := r.Bytes()
		copy(out[i*32+(32-len(rb)):], rb)
	}
	return out, nil
}

// nttDotProductPrecompile computes the dot product of two vectors mod BN254.
// Address: 0x13
// Input: 2N 32-byte elements (A[0..N-1] || B[0..N-1]).
// Output: single 32-byte result = sum(A[i]*B[i]) mod BN254.
type nttDotProductPrecompile struct{}

func (c *nttDotProductPrecompile) RequiredGas(input []byte) uint64 {
	n := uint64(len(input) / 64)
	if n == 0 {
		return 0
	}
	if n < 600 {
		return 600
	}
	return n
}

func (c *nttDotProductPrecompile) Run(input []byte) ([]byte, error) {
	if len(input) == 0 || len(input)%32 != 0 {
		return nil, ErrNTTInvalidInput
	}
	totalElems := len(input) / 32
	if totalElems%2 != 0 {
		return nil, ErrNTTInvalidInput
	}
	n := totalElems / 2

	sum := new(big.Int)
	for i := 0; i < n; i++ {
		a := new(big.Int).SetBytes(input[i*32 : (i+1)*32])
		b := new(big.Int).SetBytes(input[(n+i)*32 : (n+i+1)*32])
		prod := new(big.Int).Mul(a, b)
		prod.Mod(prod, bn254ScalarField)
		sum.Add(sum, prod)
		sum.Mod(sum, bn254ScalarField)
	}

	out := make([]byte, 32)
	sb := sum.Bytes()
	copy(out[32-len(sb):], sb)
	return out, nil
}

// nttButterflyPrecompile applies bit-reversal permutation to input elements.
// Address: 0x14
// Input: N 32-byte elements (power-of-two N).
// Output: N 32-byte elements with bit-reversal permutation applied.
type nttButterflyPrecompile struct{}

func (c *nttButterflyPrecompile) RequiredGas(input []byte) uint64 {
	n := uint64(len(input) / 32)
	if n == 0 {
		return 0
	}
	if n < 300 {
		return 300
	}
	return n
}

func (c *nttButterflyPrecompile) Run(input []byte) ([]byte, error) {
	if len(input) == 0 || len(input)%32 != 0 {
		return nil, ErrNTTInvalidInput
	}
	n := len(input) / 32
	if n == 0 {
		return nil, ErrNTTInvalidInput
	}
	if n&(n-1) != 0 {
		return nil, ErrNTTNotPowerOfTwo
	}

	logN := bits.Len(uint(n)) - 1
	out := make([]byte, n*32)
	for i := 0; i < n; i++ {
		j := bitReverse(i, logN)
		copy(out[i*32:(i+1)*32], input[j*32:(j+1)*32])
	}
	return out, nil
}

// PrecompiledContractsIPlus extends Glamsterdan with I+ fork precompiles:
// EIP-7885 NTT precompiles (0x0f-0x14) and NII precompiles (0x0201-0x0204).
var PrecompiledContractsIPlus = func() map[types.Address]PrecompiledContract {
	m := make(map[types.Address]PrecompiledContract, len(PrecompiledContractsGlamsterdan)+10)
	for addr, c := range PrecompiledContractsGlamsterdan {
		m[addr] = c
	}
	// EIP-7885: Split NTT precompiles.
	m[types.BytesToAddress([]byte{0x0f})] = &nttFWPrecompile{}
	m[types.BytesToAddress([]byte{0x10})] = &nttINVPrecompile{}
	m[types.BytesToAddress([]byte{0x11})] = &nttVecMulModPrecompile{}
	m[types.BytesToAddress([]byte{0x12})] = &nttVecAddModPrecompile{}
	m[types.BytesToAddress([]byte{0x13})] = &nttDotProductPrecompile{}
	m[types.BytesToAddress([]byte{0x14})] = &nttButterflyPrecompile{}
	// NII precompiles.
	m[NiiModExpAddr] = &NiiModExpPrecompile{}
	m[NiiFieldMulAddr] = &NiiFieldMulPrecompile{}
	m[NiiFieldInvAddr] = &NiiFieldInvPrecompile{}
	m[NiiBatchVerifyAddr] = &NiiBatchVerifyPrecompile{}
	return m
}()
