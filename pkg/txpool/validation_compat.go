package txpool

// validation_compat.go re-exports types from txpool/validation for backward compatibility.

import (
	"math/big"
	"time"

	"github.com/eth2030/eth2030/txpool/validation"
)

// Validation type aliases.
type (
	TxValidationConfig       = validation.TxValidationConfig
	TxValidationResult       = validation.TxValidationResult
	TxValidator              = validation.TxValidator
	ValidationErrorCode      = validation.ValidationErrorCode
	ValidationResult         = validation.ValidationResult
	ValidationPipelineConfig = validation.ValidationPipelineConfig
	StateProvider            = validation.StateProvider
	SyntaxCheck              = validation.SyntaxCheck
	SignatureVerify          = validation.SignatureVerify
	StateCheck               = validation.StateCheck
	BlobCheck                = validation.BlobCheck
	RateLimiter              = validation.RateLimiter
	ValidationPipeline       = validation.ValidationPipeline
)

// Validation error variables.
var (
	ErrTxGasTooLow    = validation.ErrTxGasTooLow
	ErrTxGasTooHigh   = validation.ErrTxGasTooHigh
	ErrTxDataTooLarge = validation.ErrTxDataTooLarge
	ErrTxValueTooHigh = validation.ErrTxValueTooHigh
	ErrTxNoSignature  = validation.ErrTxNoSignature
	ErrTxBadChainID   = validation.ErrTxBadChainID

	ErrVPNilTx            = validation.ErrVPNilTx
	ErrVPGasZero          = validation.ErrVPGasZero
	ErrVPGasExceedsMax    = validation.ErrVPGasExceedsMax
	ErrVPNegativeValue    = validation.ErrVPNegativeValue
	ErrVPNegativeGasPrice = validation.ErrVPNegativeGasPrice
	ErrVPFeeBelowTip      = validation.ErrVPFeeBelowTip
	ErrVPDataTooLarge     = validation.ErrVPDataTooLarge
	ErrVPNoSignature      = validation.ErrVPNoSignature
	ErrVPInvalidSignature = validation.ErrVPInvalidSignature
	ErrVPNonceTooLow      = validation.ErrVPNonceTooLow
	ErrVPNonceTooHigh     = validation.ErrVPNonceTooHigh
	ErrVPInsufficientBal  = validation.ErrVPInsufficientBal
	ErrVPBlobMissingHash  = validation.ErrVPBlobMissingHash
	ErrVPBlobFeeTooLow    = validation.ErrVPBlobFeeTooLow
	ErrVPRateLimited      = validation.ErrVPRateLimited
)

// Validation constants.
const (
	ValidationOK           = validation.ValidationOK
	ValidationSyntaxErr    = validation.ValidationSyntaxErr
	ValidationSignatureErr = validation.ValidationSignatureErr
	ValidationStateErr     = validation.ValidationStateErr
	ValidationBlobErr      = validation.ValidationBlobErr
	ValidationRateLimitErr = validation.ValidationRateLimitErr
)

// Validation function wrappers.
func DefaultTxValidationConfig() TxValidationConfig {
	return validation.DefaultTxValidationConfig()
}
func NewTxValidator(config TxValidationConfig) *TxValidator {
	return validation.NewTxValidator(config)
}
func DefaultValidationPipelineConfig() ValidationPipelineConfig {
	return validation.DefaultValidationPipelineConfig()
}
func NewSyntaxCheck(maxGasLimit uint64, maxDataSize int) *SyntaxCheck {
	return validation.NewSyntaxCheck(maxGasLimit, maxDataSize)
}
func NewSignatureVerify() *SignatureVerify { return validation.NewSignatureVerify() }
func NewStateCheck(state StateProvider, maxNonceGap uint64) *StateCheck {
	return validation.NewStateCheck(state, maxNonceGap)
}
func NewBlobCheck(blobBaseFee *big.Int) *BlobCheck { return validation.NewBlobCheck(blobBaseFee) }
func NewRateLimiter(maxPerPeer int, window time.Duration) *RateLimiter {
	return validation.NewRateLimiter(maxPerPeer, window)
}
func NewValidationPipeline(config ValidationPipelineConfig, state StateProvider) *ValidationPipeline {
	return validation.NewValidationPipeline(config, state)
}

// vpMakeTx is not exported; only used in tests. No wrapper needed.
// The root-package tests that used these types will use the aliases above.
