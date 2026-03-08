// Package convert provides payload conversion utilities for the Engine API.
// It converts between execution payload versions and block headers.
package convert

import (
	"crypto/sha256"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	engblobs "github.com/eth2030/eth2030/engine/blobsbundle"
	engpayload "github.com/eth2030/eth2030/engine/payload"
)

// PayloadVersion indicates which execution payload format to use.
type PayloadVersion int

const (
	PayloadV1 PayloadVersion = 1
	PayloadV2 PayloadVersion = 2
	PayloadV3 PayloadVersion = 3
	PayloadV4 PayloadVersion = 4
	PayloadV5 PayloadVersion = 5
)

// PayloadToHeaderV1 converts an ExecutionPayloadV1 to a block Header.
func PayloadToHeaderV1(p *engpayload.ExecutionPayloadV1) *types.Header {
	header := &types.Header{
		ParentHash:  p.ParentHash,
		Coinbase:    p.FeeRecipient,
		Root:        p.StateRoot,
		ReceiptHash: p.ReceiptsRoot,
		Bloom:       p.LogsBloom,
		MixDigest:   p.PrevRandao,
		Number:      new(big.Int).SetUint64(p.BlockNumber),
		GasLimit:    p.GasLimit,
		GasUsed:     p.GasUsed,
		Time:        p.Timestamp,
		Extra:       p.ExtraData,
		BaseFee:     p.BaseFeePerGas,
	}
	header.Difficulty = new(big.Int)
	header.UncleHash = types.EmptyUncleHash
	return header
}

// PayloadToHeaderV2 converts an ExecutionPayloadV2 (Shanghai) to a block Header.
func PayloadToHeaderV2(p *engpayload.ExecutionPayloadV2) *types.Header {
	header := PayloadToHeaderV1(&p.ExecutionPayloadV1)
	// V2 adds withdrawals, but the withdrawals root goes into the header
	// through a separate mechanism (block body).
	return header
}

// PayloadToHeaderV3 converts an ExecutionPayloadV3 (Cancun) to a block Header.
func PayloadToHeaderV3(p *engpayload.ExecutionPayloadV3) *types.Header {
	header := PayloadToHeaderV2(&p.ExecutionPayloadV2)
	header.BlobGasUsed = &p.BlobGasUsed
	header.ExcessBlobGas = &p.ExcessBlobGas
	return header
}

// PayloadToHeaderV5 converts an ExecutionPayloadV5 (Amsterdam) to a block Header.
func PayloadToHeaderV5(p *engpayload.ExecutionPayloadV5) *types.Header {
	return PayloadToHeaderV3(&p.ExecutionPayloadV3)
}

// HeaderToPayloadV2 extracts V2 payload fields from a Header and withdrawals.
func HeaderToPayloadV2(header *types.Header, withdrawals []*engpayload.Withdrawal) engpayload.ExecutionPayloadV2 {
	v1 := engpayload.HeaderToPayloadFields(header)
	return engpayload.ExecutionPayloadV2{
		ExecutionPayloadV1: v1,
		Withdrawals:        withdrawals,
	}
}

// HeaderToPayloadV3 extracts V3 payload fields from a Header and withdrawals.
func HeaderToPayloadV3(header *types.Header, withdrawals []*engpayload.Withdrawal) engpayload.ExecutionPayloadV3 {
	v2 := HeaderToPayloadV2(header, withdrawals)
	v3 := engpayload.ExecutionPayloadV3{
		ExecutionPayloadV2: v2,
	}
	if header.BlobGasUsed != nil {
		v3.BlobGasUsed = *header.BlobGasUsed
	}
	if header.ExcessBlobGas != nil {
		v3.ExcessBlobGas = *header.ExcessBlobGas
	}
	return v3
}

// ExtractVersionedHashes extracts EIP-4844 versioned hashes from encoded
// transactions. Each blob transaction's blob hashes are collected in order.
func ExtractVersionedHashes(txBytes [][]byte) []types.Hash {
	var hashes []types.Hash
	for _, raw := range txBytes {
		tx, err := types.DecodeTxRLP(raw)
		if err != nil {
			continue
		}
		blobHashes := tx.BlobHashes()
		if len(blobHashes) > 0 {
			hashes = append(hashes, blobHashes...)
		}
	}
	return hashes
}

// VersionedHashFromCommitment computes the EIP-4844 versioned hash from a
// KZG commitment. SHA-256 hash with byte 0 replaced by version byte 0x01.
func VersionedHashFromCommitment(commitment []byte) types.Hash {
	h := sha256.Sum256(commitment)
	h[0] = engblobs.VersionedHashVersion
	return types.Hash(h)
}

// BlobSidecarFromBundle extracts a single blob sidecar from a blobs bundle
// at the given index. Includes the block hash for association.
func BlobSidecarFromBundle(bundle *engpayload.BlobsBundleV1, index int, blockHash types.Hash) (*engblobs.BlobSidecar, error) {
	if bundle == nil {
		return nil, engblobs.ErrBlobBundleEmpty
	}
	if index < 0 || index >= len(bundle.Blobs) {
		return nil, engblobs.ErrBlobBundleSidecarIndex
	}
	return &engblobs.BlobSidecar{
		Index:             uint64(index),
		Blob:              bundle.Blobs[index],
		KZGCommitment:     bundle.Commitments[index],
		KZGProof:          bundle.Proofs[index],
		SignedBlockHeader: blockHash,
	}, nil
}

// ProcessWithdrawalsExt processes engine withdrawals and returns the total
// withdrawal amount in Gwei along with per-validator amounts.
func ProcessWithdrawalsExt(withdrawals []*engpayload.Withdrawal) (totalGwei uint64, byValidator map[uint64]uint64) {
	byValidator = make(map[uint64]uint64, len(withdrawals))
	for _, w := range withdrawals {
		totalGwei += w.Amount
		byValidator[w.ValidatorIndex] += w.Amount
	}
	return totalGwei, byValidator
}

// CoreWithdrawalsFromPayload extracts core Withdrawal types from an
// ExecutionPayloadV2 (or higher version through embedding).
func CoreWithdrawalsFromPayload(p *engpayload.ExecutionPayloadV2) []*types.Withdrawal {
	if p == nil || p.Withdrawals == nil {
		return nil
	}
	return engpayload.WithdrawalsToCore(p.Withdrawals)
}

// ForkTimestamps holds fork activation timestamps for payload version selection.
type ForkTimestamps struct {
	Shanghai  uint64
	Cancun    uint64
	Prague    uint64
	Amsterdam uint64
}

// DeterminePayloadVersion returns the highest applicable payload version
// for a given block timestamp.
func DeterminePayloadVersion(timestamp uint64, forks *ForkTimestamps) PayloadVersion {
	if forks == nil {
		return PayloadV1
	}
	if forks.Amsterdam > 0 && timestamp >= forks.Amsterdam {
		return PayloadV5
	}
	if forks.Prague > 0 && timestamp >= forks.Prague {
		return PayloadV4
	}
	if forks.Cancun > 0 && timestamp >= forks.Cancun {
		return PayloadV3
	}
	if forks.Shanghai > 0 && timestamp >= forks.Shanghai {
		return PayloadV2
	}
	return PayloadV1
}

// ConvertV1ToV2 upgrades a V1 payload to V2 by adding empty withdrawals.
func ConvertV1ToV2(v1 *engpayload.ExecutionPayloadV1) *engpayload.ExecutionPayloadV2 {
	return &engpayload.ExecutionPayloadV2{
		ExecutionPayloadV1: *v1,
		Withdrawals:        []*engpayload.Withdrawal{},
	}
}

// ConvertV2ToV3 upgrades a V2 payload to V3 with initial blob gas fields.
func ConvertV2ToV3(v2 *engpayload.ExecutionPayloadV2) *engpayload.ExecutionPayloadV3 {
	return &engpayload.ExecutionPayloadV3{
		ExecutionPayloadV2: *v2,
		BlobGasUsed:        0,
		ExcessBlobGas:      0,
	}
}

// ConvertV3ToV4 upgrades a V3 payload to V4 with empty execution requests.
func ConvertV3ToV4(v3 *engpayload.ExecutionPayloadV3) *engpayload.ExecutionPayloadV4 {
	return &engpayload.ExecutionPayloadV4{
		ExecutionPayloadV3: *v3,
		ExecutionRequests:  [][]byte{},
	}
}

// ConvertV4ToV5 upgrades a V4 payload to V5 with empty block access list.
func ConvertV4ToV5(v4 *engpayload.ExecutionPayloadV4) *engpayload.ExecutionPayloadV5 {
	return &engpayload.ExecutionPayloadV5{
		ExecutionPayloadV4: *v4,
		BlockAccessList:    nil,
	}
}

// ValidatePayloadConsistency checks that a payload's block hash field
// matches the hash computed from the header derived from the payload.
func ValidatePayloadConsistency(p *engpayload.ExecutionPayloadV3) bool {
	header := PayloadToHeaderV3(p)
	computed := header.Hash()
	return computed == p.BlockHash
}

// WithdrawalsSummary provides summary statistics for a set of withdrawals.
type WithdrawalsSummary struct {
	Count            int
	TotalAmountGwei  uint64
	UniqueValidators int
	UniqueAddresses  int
}

// SummarizeWithdrawals computes summary statistics for withdrawals.
func SummarizeWithdrawals(withdrawals []*engpayload.Withdrawal) WithdrawalsSummary {
	validators := make(map[uint64]struct{})
	addresses := make(map[types.Address]struct{})
	var totalGwei uint64

	for _, w := range withdrawals {
		totalGwei += w.Amount
		validators[w.ValidatorIndex] = struct{}{}
		addresses[w.Address] = struct{}{}
	}

	return WithdrawalsSummary{
		Count:            len(withdrawals),
		TotalAmountGwei:  totalGwei,
		UniqueValidators: len(validators),
		UniqueAddresses:  len(addresses),
	}
}
