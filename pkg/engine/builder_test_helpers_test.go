package engine

import (
	"github.com/eth2030/eth2030/core/types"
)

func newTestPubkey(b byte) BLSPubkey {
	var pk BLSPubkey
	pk[0] = b
	return pk
}

func newTestRegistration(pubkey BLSPubkey) *BuilderRegistrationV1 {
	return &BuilderRegistrationV1{
		FeeRecipient: types.HexToAddress("0xfee"),
		GasLimit:     30_000_000,
		Timestamp:    1700000000,
		Pubkey:       pubkey,
	}
}

func newTestBid(builderIdx BuilderIndex, slot uint64, value uint64) *SignedExecutionPayloadBid {
	return &SignedExecutionPayloadBid{
		Message: ExecutionPayloadBid{
			ParentBlockHash: types.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			BlockHash:       types.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
			Slot:            slot,
			Value:           value,
			GasLimit:        30_000_000,
			BuilderIndex:    builderIdx,
			FeeRecipient:    types.HexToAddress("0xfeefeefeefeefeefeefe"),
		},
	}
}
