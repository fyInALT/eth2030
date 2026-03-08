package das

// field_compat.go re-exports types from das/field for backward compatibility.

import "github.com/eth2030/eth2030/das/field"

// FieldElement is a BLS12-381 scalar field element.
type FieldElement = field.FieldElement

// Field element constructors.
var (
	NewFieldElement            = field.NewFieldElement
	NewFieldElementFromUint64  = field.NewFieldElementFromUint64
	FieldZero                  = field.FieldZero
	FieldOne                   = field.FieldOne
	FFT                        = field.FFT
	InverseFFT                 = field.InverseFFT
	ComputeRootsOfUnity        = field.ComputeRootsOfUnity
)

// blsModulus is the BLS12-381 scalar field order, kept for package-level tests.
var blsModulus = field.BLSModulus
