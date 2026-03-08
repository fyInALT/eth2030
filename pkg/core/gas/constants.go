package gas

import (
	"errors"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

// Transaction gas constants (from EIP-2028, EIP-3860, etc.)
const (
	// TxGas is the base gas cost of a transaction (21000).
	TxGas uint64 = 21000

	// TxDataZeroGas is the gas cost per zero byte of transaction data.
	TxDataZeroGas uint64 = 4

	// TxDataNonZeroGas is the gas cost per non-zero byte of transaction data.
	TxDataNonZeroGas uint64 = 16

	// TxCreateGas is the extra gas for contract creation transactions.
	TxCreateGas uint64 = 32000

	// TotalCostFloorPerToken is the floor gas cost per calldata token (EIP-7623).
	TotalCostFloorPerToken uint64 = 10

	// TotalCostFloorPerTokenGlamst is the floor gas cost per calldata token
	// under Glamsterdam (EIP-7976).
	TotalCostFloorPerTokenGlamst uint64 = 16

	// StandardTokenCost is the standard EIP-2028 calldata cost for non-zero bytes.
	StandardTokenCost uint64 = 16
)

// Block gas limit constants.
const (
	// GasLimitBoundDivisor is the divisor for max gas limit change per block.
	GasLimitBoundDivisor uint64 = 1024

	// MinGasLimit is the minimum gas limit.
	MinGasLimit uint64 = 5000

	// ElasticityMultiplier is the EIP-1559 elasticity multiplier.
	ElasticityMultiplier uint64 = 2

	// BaseFeeChangeDenominator is the EIP-1559 base fee change denominator.
	BaseFeeChangeDenominator uint64 = 8
)

// EIP-4844 blob transaction constants.
const (
	// MaxBlobGasPerBlock is the maximum blob gas allowed in a single block (Cancun).
	MaxBlobGasPerBlock = 786432

	// TargetBlobGasPerBlock is the target blob gas per block for the
	// EIP-4844 blob base fee adjustment mechanism.
	TargetBlobGasPerBlock = 393216

	// GasPerBlob is the gas consumed by each blob (2^17).
	GasPerBlob = 131072

	// BlobTxHashVersion is the required first byte of each versioned blob hash.
	BlobTxHashVersion = 0x01

	// MaxBlobsPerBlock is the maximum number of blobs per block (Cancun default).
	MaxBlobsPerBlock = 6
)

// calcBlobBaseFee computes the blob base fee from the excess blob gas.
// Uses the EIP-4844 fake exponential with MIN_BASE_FEE=1 and update fraction 3338477.
func calcBlobBaseFee(excessBlobGas uint64) *big.Int {
	return fakeExponential(big.NewInt(1), new(big.Int).SetUint64(excessBlobGas), big.NewInt(3338477))
}

// fakeExponential approximates factor * e^(numerator / denominator) using Taylor expansion.
func fakeExponential(factor, numerator, denominator *big.Int) *big.Int {
	i := new(big.Int).SetUint64(1)
	output := new(big.Int)
	numeratorAccum := new(big.Int).Mul(factor, denominator)
	tmp := new(big.Int)
	denom := new(big.Int)
	for numeratorAccum.Sign() > 0 {
		output.Add(output, numeratorAccum)
		tmp.Mul(numeratorAccum, numerator)
		denom.Mul(denominator, i)
		numeratorAccum.Div(tmp, denom)
		i.Add(i, big.NewInt(1))
	}
	output.Div(output, denominator)
	return output
}

// Shared gas errors.
var (
	// ErrGasLimitExceeded is returned when a transaction's gas exceeds the block gas limit.
	ErrGasLimitExceeded = errors.New("gas limit exceeded")

	// ErrIntrinsicGasTooLow is returned when a transaction's gas limit is below intrinsic cost.
	ErrIntrinsicGasTooLow = errors.New("intrinsic gas too low")
)

// calldataTokens computes calldata tokens for the standard EIP-7623 path.
// tokens = zero_bytes * 1 + nonzero_bytes * 4
func calldataTokens(data []byte) uint64 {
	var tokens uint64
	for _, b := range data {
		if b == 0 {
			tokens++
		} else {
			tokens += 4
		}
	}
	return tokens
}

// accessListDataTokens computes data tokens for access list entries per EIP-7981.
// tokens = zero_bytes + nonzero_bytes * 4 for all addresses and storage keys.
func accessListDataTokens(accessList types.AccessList) uint64 {
	var zero, nonzero uint64
	for _, tuple := range accessList {
		// Count bytes in address (20 bytes).
		for _, b := range tuple.Address {
			if b == 0 {
				zero++
			} else {
				nonzero++
			}
		}
		// Count bytes in each storage key (32 bytes).
		for _, key := range tuple.StorageKeys {
			for _, b := range key {
				if b == 0 {
					zero++
				} else {
					nonzero++
				}
			}
		}
	}
	return zero + nonzero*4
}
