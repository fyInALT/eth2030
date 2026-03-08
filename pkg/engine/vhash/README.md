# engine/vhash — EIP-4844 KZG versioned hash utilities

## Overview

Package `vhash` implements the EIP-4844 versioned hash scheme for KZG polynomial commitments carried in blob transactions. A versioned hash is a 32-byte value whose first byte encodes the hash version (currently `0x01` for KZG) and whose remaining 31 bytes are the low-order bytes of the SHA-256 digest of the 48-byte KZG commitment.

The package provides both single-commitment and batch operations, transaction-level collection of blob hashes, and full verification that a set of versioned hashes is consistent with a corresponding set of KZG commitments. These utilities are used by the Engine API newPayload handlers to validate blob transactions arriving from the consensus layer.

## Functionality

**Constants**
- `VersionKZG byte = 0x01` — current KZG version byte
- `KZGCommitmentSize = 48` — size of a serialized KZG commitment

**Single-commitment functions**
- `ComputeVersionedHash(commitment []byte, version byte) (Hash, error)` — SHA-256 then overwrite byte[0] with version
- `ComputeVersionedHashKZG(commitment []byte) (Hash, error)` — shorthand using `VersionKZG`

**Batch and collection functions**
- `BatchComputeVersionedHashes(commitments [][]byte) ([]Hash, error)`
- `CollectBlobHashesFromTransactions(txs [][]byte) ([]Hash, error)` — decodes blob txs and extracts their versioned hashes
- `BuildCommitmentHashMap(commitments [][]byte) (map[Hash][]byte, error)` — index commitments by their versioned hash

**Validation functions**
- `VerifyVersionedHashesAgainstCommitments(hashes []Hash, commitments [][]byte) error` — checks 1-to-1 correspondence
- `ValidateBlobTxVersionedHashes(tx []byte, expectedHashes []Hash) error` — validates a single blob tx
- `VerifyAllBlobVersionBytes(hashes []Hash) error` — ensures every hash has version byte `0x01`
- `ValidateVersionByte(hash Hash) error`

**Inspection helpers**
- `ExtractVersionByte(hash Hash) byte`
- `IsKZGVersionedHash(hash Hash) bool`

Parent package: [`engine`](../README.md)
