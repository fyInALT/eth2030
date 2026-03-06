package consensus

// headval_compat.go re-exports types from consensus/headval for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/consensus/headval"
)

// Header validation type aliases.
type HeaderValidator = headval.HeaderValidator

// Header validation constants.
const (
	MaxExtraDataBytes    = headval.MaxExtraDataBytes
	GasLimitBoundDivisor = headval.GasLimitBoundDivisor
)

// Header validation error variables.
var (
	ErrInvalidParentHash   = headval.ErrInvalidParentHash
	ErrInvalidNumber       = headval.ErrInvalidNumber
	ErrInvalidTimestamp    = headval.ErrInvalidTimestamp
	ErrInvalidGasLimit     = headval.ErrInvalidGasLimit
	ErrGasUsedExceedsLimit = headval.ErrGasUsedExceedsLimit
	ErrExtraDataTooLong    = headval.ErrExtraDataTooLong
	ErrNilHeader           = headval.ErrNilHeader
	ErrNilParent           = headval.ErrNilParent
)

// Header validation function wrappers.
func NewHeaderValidator() *HeaderValidator { return headval.NewHeaderValidator() }
func ValidateGasLimit(parentLimit, headerLimit uint64) bool {
	return headval.ValidateGasLimit(parentLimit, headerLimit)
}
func ValidateTimestamp(parentTime, headerTime uint64) bool {
	return headval.ValidateTimestamp(parentTime, headerTime)
}
func CalcDifficulty(parentDifficulty *big.Int, parentTimestamp, currentTimestamp uint64) *big.Int {
	return headval.CalcDifficulty(parentDifficulty, parentTimestamp, currentTimestamp)
}
