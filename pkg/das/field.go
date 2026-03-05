package das

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

// blsModulus is the BLS12-381 scalar field order, used in tests.
var blsModulus = fr.Modulus()

// FieldElement represents an element of the BLS12-381 scalar field.
// Internally backed by gnark-crypto fr.Element (Montgomery form, 4×uint64),
// which is ~10-50x faster than big.Int modular arithmetic.
type FieldElement struct {
	v fr.Element
}

// NewFieldElement creates a FieldElement from a big.Int, reducing mod p.
func NewFieldElement(v *big.Int) FieldElement {
	var e fr.Element
	e.SetBigInt(v)
	return FieldElement{v: e}
}

// NewFieldElementFromUint64 creates a FieldElement from a uint64.
func NewFieldElementFromUint64(v uint64) FieldElement {
	return FieldElement{v: fr.NewElement(v)}
}

// FieldZero returns the additive identity.
func FieldZero() FieldElement {
	return FieldElement{}
}

// FieldOne returns the multiplicative identity.
func FieldOne() FieldElement {
	return FieldElement{v: fr.One()}
}

// IsZero returns true if the element is zero.
func (a FieldElement) IsZero() bool {
	return a.v.IsZero()
}

// Equal returns true if two field elements are equal.
func (a FieldElement) Equal(b FieldElement) bool {
	return a.v.Equal(&b.v)
}

// BigInt returns the value as a big.Int.
func (a FieldElement) BigInt() *big.Int {
	return a.v.BigInt(new(big.Int))
}

// Add returns a + b mod p.
func (a FieldElement) Add(b FieldElement) FieldElement {
	var r fr.Element
	r.Add(&a.v, &b.v)
	return FieldElement{v: r}
}

// Sub returns a - b mod p.
func (a FieldElement) Sub(b FieldElement) FieldElement {
	var r fr.Element
	r.Sub(&a.v, &b.v)
	return FieldElement{v: r}
}

// Mul returns a * b mod p.
func (a FieldElement) Mul(b FieldElement) FieldElement {
	var r fr.Element
	r.Mul(&a.v, &b.v)
	return FieldElement{v: r}
}

// Neg returns -a mod p.
func (a FieldElement) Neg() FieldElement {
	var r fr.Element
	r.Neg(&a.v)
	return FieldElement{v: r}
}

// Inv returns the multiplicative inverse a^{-1} mod p.
// Returns zero if a is zero.
func (a FieldElement) Inv() FieldElement {
	if a.v.IsZero() {
		return FieldZero()
	}
	var r fr.Element
	r.Inverse(&a.v)
	return FieldElement{v: r}
}

// Exp returns a^exp mod p.
func (a FieldElement) Exp(exp *big.Int) FieldElement {
	var r fr.Element
	r.Exp(a.v, exp)
	return FieldElement{v: r}
}

// Div returns a / b mod p (i.e., a * b^{-1}).
func (a FieldElement) Div(b FieldElement) FieldElement {
	var r fr.Element
	r.Div(&a.v, &b.v)
	return FieldElement{v: r}
}

// rootOfUnity computes a primitive n-th root of unity in the BLS12-381 scalar field.
// n must be a power of 2 and must divide (p-1).
func rootOfUnity(n uint64) FieldElement {
	if n == 0 || n&(n-1) != 0 {
		panic("das: rootOfUnity: n must be a power of 2")
	}
	// Use gnark-crypto's built-in Generator which returns the 2^s-th root of unity.
	// BLS12-381 Fr has 2-adicity of 32 (p-1 = 2^32 * q).
	// Generator(m) returns a primitive m-th root of unity for m dividing 2^32.
	g, err := fr.Generator(n)
	if err != nil {
		panic("das: rootOfUnity: " + err.Error())
	}
	return FieldElement{v: g}
}

// computeRootsOfUnity returns the n-th roots of unity [w^0, w^1, ..., w^{n-1}]
// where w is a primitive n-th root of unity.
func computeRootsOfUnity(n uint64) []FieldElement {
	w := rootOfUnity(n)
	roots := make([]FieldElement, n)
	roots[0] = FieldOne()
	for i := uint64(1); i < n; i++ {
		roots[i] = roots[i-1].Mul(w)
	}
	return roots
}

// FFT computes the Number Theoretic Transform (forward FFT) of vals
// over the BLS12-381 scalar field. len(vals) must be a power of 2.
func FFT(vals []FieldElement) []FieldElement {
	n := len(vals)
	if n <= 1 {
		out := make([]FieldElement, n)
		copy(out, vals)
		return out
	}
	if n&(n-1) != 0 {
		panic("das: FFT: length must be a power of 2")
	}
	roots := computeRootsOfUnity(uint64(n))
	return fftInner(vals, roots)
}

// InverseFFT computes the inverse NTT (inverse FFT) of vals.
func InverseFFT(vals []FieldElement) []FieldElement {
	n := len(vals)
	if n <= 1 {
		out := make([]FieldElement, n)
		copy(out, vals)
		return out
	}
	if n&(n-1) != 0 {
		panic("das: InverseFFT: length must be a power of 2")
	}
	roots := computeRootsOfUnity(uint64(n))

	// Inverse roots: reverse the root array (except index 0).
	invRoots := make([]FieldElement, n)
	invRoots[0] = roots[0]
	for i := 1; i < n; i++ {
		invRoots[i] = roots[n-i]
	}

	result := fftInner(vals, invRoots)

	// Divide by n.
	nInv := NewFieldElementFromUint64(uint64(n)).Inv()
	for i := range result {
		result[i] = result[i].Mul(nInv)
	}
	return result
}

// fftInner performs the Cooley-Tukey butterfly FFT using precomputed roots.
func fftInner(vals []FieldElement, roots []FieldElement) []FieldElement {
	n := len(vals)
	if n == 1 {
		return []FieldElement{vals[0]}
	}

	half := n / 2
	even := make([]FieldElement, half)
	odd := make([]FieldElement, half)
	evenRoots := make([]FieldElement, half)
	for i := 0; i < half; i++ {
		even[i] = vals[2*i]
		odd[i] = vals[2*i+1]
		evenRoots[i] = roots[2*i]
	}

	yEven := fftInner(even, evenRoots)
	yOdd := fftInner(odd, evenRoots)

	result := make([]FieldElement, n)
	for i := 0; i < half; i++ {
		t := roots[i].Mul(yOdd[i])
		result[i] = yEven[i].Add(t)
		result[i+half] = yEven[i].Sub(t)
	}
	return result
}
