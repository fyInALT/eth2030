// epbs.go implements ePBS (EIP-7732) Engine API types.
// The actual method logic lives in the main engine package as methods on EngineAPI,
// since those methods require access to builderRegistry on EngineAPI.
// This file defines only the exported types used by those methods.
package api

import "github.com/eth2030/eth2030/epbs"

// GetPayloadHeaderV1Response is the response for engine_getPayloadHeaderV1.
type GetPayloadHeaderV1Response struct {
	Bid *epbs.SignedBuilderBid `json:"bid"`
}

// SubmitBlindedBlockV1Request contains a blinded block submission from the proposer.
type SubmitBlindedBlockV1Request struct {
	Slot            uint64   `json:"slot"`
	BuilderIndex    uint64   `json:"builderIndex"`
	BidHash         [32]byte `json:"bidHash"`
	BeaconBlockRoot [32]byte `json:"beaconBlockRoot"`
}

// SubmitBlindedBlockV1Response is the response to engine_submitBlindedBlockV1.
type SubmitBlindedBlockV1Response struct {
	Status string `json:"status"`
}
