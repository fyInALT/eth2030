package vm

// abi_compat.go re-exports types from core/vm/abi for backward compatibility.

import "github.com/eth2030/eth2030/core/vm/abi"

// ABI type aliases.
type (
	ABITypeKind = abi.ABITypeKind
	ABIType     = abi.ABIType
	ABIValue    = abi.ABIValue
)

// ABI kind constants.
const (
	ABIUint256      = abi.ABIUint256
	ABIAddress      = abi.ABIAddress
	ABIBool         = abi.ABIBool
	ABIBytes        = abi.ABIBytes
	ABIString       = abi.ABIString
	ABIFixedArray   = abi.ABIFixedArray
	ABIDynamicArray = abi.ABIDynamicArray
	ABITuple        = abi.ABITuple
	ABIFixedBytes   = abi.ABIFixedBytes
)

// ABI error variables.
var (
	ErrABIShortData      = abi.ErrABIShortData
	ErrABIInvalidBool    = abi.ErrABIInvalidBool
	ErrABIInvalidType    = abi.ErrABIInvalidType
	ErrABIOffsetOverflow = abi.ErrABIOffsetOverflow
)

// ABI function wrappers.
func ComputeSelector(signature string) [4]byte { return abi.ComputeSelector(signature) }
func EncodeFunctionCall(selector [4]byte, args []ABIValue) []byte {
	return abi.EncodeFunctionCall(selector, args)
}
func DecodeFunctionResult(data []byte, abiTypes []ABIType) ([]ABIValue, error) {
	return abi.DecodeFunctionResult(data, abiTypes)
}
func Uint256ToBytes(v uint64) []byte { return abi.Uint256ToBytes(v) }
