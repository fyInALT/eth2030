# vm/abi — Solidity ABI encoder/decoder

[← vm](../README.md)

## Overview

Package `abi` implements Solidity's contract ABI encoding and decoding as specified in the ABI specification. It supports all static and dynamic types: `uint256`, `address`, `bool`, `bytes`, `string`, fixed and dynamic arrays, tuples, and `bytesN`. The encoder handles head/tail layout for dynamic types automatically.

This package is used internally by precompile helpers and the EVM call handler to encode/decode calldata and return values without relying on external code generation.

## Functionality

### Types

```go
type ABITypeKind uint8
const (
    ABIUint256      ABITypeKind = iota
    ABIAddress
    ABIBool
    ABIBytes
    ABIString
    ABIFixedArray
    ABIDynamicArray
    ABITuple
    ABIFixedBytes   // bytesN (1..32)
)

type ABIType struct {
    Kind   ABITypeKind
    Size   int       // fixed array length or bytesN width
    Elem   *ABIType  // for arrays
    Fields []ABIType // for tuples
}

type ABIValue struct {
    Type    ABIType
    Uint256 *big.Int
    Addr    types.Address
    Bool    bool
    Bytes   []byte
    String  string
    Array   []ABIValue
    Tuple   []ABIValue
}
```

### Functions

```go
// Encode packs a list of ABIValues into ABI-encoded bytes.
func Encode(values []ABIValue) ([]byte, error)

// Decode unpacks ABI-encoded bytes into ABIValues according to types.
func Decode(types []ABIType, data []byte) ([]ABIValue, error)

// EncodeWithSelector prepends a 4-byte function selector derived from the
// Keccak-256 of the signature string.
func EncodeWithSelector(sig string, values []ABIValue) ([]byte, error)

// Selector returns the 4-byte selector for a function signature.
func Selector(sig string) [4]byte
```

## Usage

```go
// Encode: transfer(address, uint256)
values := []abi.ABIValue{
    {Type: abi.ABIType{Kind: abi.ABIAddress}, Addr: recipient},
    {Type: abi.ABIType{Kind: abi.ABIUint256}, Uint256: amount},
}
calldata, err := abi.EncodeWithSelector("transfer(address,uint256)", values)

// Decode return value (uint256)
types := []abi.ABIType{{Kind: abi.ABIUint256}}
results, err := abi.Decode(types, returnData)
fmt.Println(results[0].Uint256)
```
