// json.go provides custom JSON marshal/unmarshal for Engine API types that
// remain in the engine package (V7 and Glamsterdam extension types).
// Core payload types (V3/V4/V5, attributes, withdrawals) are marshaled in
// engine/payload/json.go.
package engine

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/eth2030/eth2030/core/types"
)

// ── primitive hex types (local to engine for V7/Glamsterdam types) ───────────

// jsonHexUint64 marshals/unmarshals as a 0x-prefixed hex string.
type jsonHexUint64 uint64

func (h jsonHexUint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("0x%x", uint64(h)))
}

func (h *jsonHexUint64) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		var n uint64
		if err2 := json.Unmarshal(data, &n); err2 != nil {
			return fmt.Errorf("jsonHexUint64: %w", err)
		}
		*h = jsonHexUint64(n)
		return nil
	}
	orig := s
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	if s == "" {
		*h = 0
		return nil
	}
	n := new(big.Int)
	if _, ok := n.SetString(s, 16); !ok {
		return fmt.Errorf("jsonHexUint64: invalid hex %q", orig)
	}
	*h = jsonHexUint64(n.Uint64())
	return nil
}

// jsonHexBytes marshals/unmarshals as a 0x-prefixed hex string.
type jsonHexBytes []byte

func (h jsonHexBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal("0x" + hex.EncodeToString(h))
}

func (h *jsonHexBytes) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("jsonHexBytes: %w", err)
	}
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	if s == "" {
		*h = nil
		return nil
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return fmt.Errorf("jsonHexBytes: %w", err)
	}
	*h = b
	return nil
}

// jsonHexBig marshals/unmarshals *big.Int as a 0x-prefixed hex string.
type jsonHexBig struct{ *big.Int }

func newJSONHexBig(v *big.Int) *jsonHexBig {
	if v == nil {
		return nil
	}
	return &jsonHexBig{v}
}

func (h jsonHexBig) MarshalJSON() ([]byte, error) {
	if h.Int == nil {
		return []byte("null"), nil
	}
	return json.Marshal("0x" + h.Int.Text(16))
}

func (h *jsonHexBig) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		n := new(big.Int)
		if err2 := json.Unmarshal(data, n); err2 != nil {
			return fmt.Errorf("jsonHexBig: %w", err)
		}
		h.Int = n
		return nil
	}
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	n := new(big.Int)
	if _, ok := n.SetString(s, 16); !ok {
		return fmt.Errorf("jsonHexBig: invalid hex")
	}
	h.Int = n
	return nil
}

func toJSONBigInt(h *jsonHexBig) *big.Int {
	if h == nil {
		return nil
	}
	return h.Int
}

func toHexBytesSliceLocal(in [][]byte) []jsonHexBytes {
	out := make([]jsonHexBytes, len(in))
	for i, b := range in {
		out[i] = jsonHexBytes(b)
	}
	return out
}

func fromHexBytesSliceLocal(in []jsonHexBytes) [][]byte {
	if in == nil {
		return nil
	}
	out := make([][]byte, len(in))
	for i, b := range in {
		out[i] = []byte(b)
	}
	return out
}

// ── ExecutionPayloadV7 (V3 + 2030 roadmap fields) ────────────────────────────

type executionPayloadV7JSON struct {
	ParentHash       types.Hash     `json:"parentHash"`
	FeeRecipient     types.Address  `json:"feeRecipient"`
	StateRoot        types.Hash     `json:"stateRoot"`
	ReceiptsRoot     types.Hash     `json:"receiptsRoot"`
	LogsBloom        types.Bloom    `json:"logsBloom"`
	PrevRandao       types.Hash     `json:"prevRandao"`
	BlockNumber      jsonHexUint64  `json:"blockNumber"`
	GasLimit         jsonHexUint64  `json:"gasLimit"`
	GasUsed          jsonHexUint64  `json:"gasUsed"`
	Timestamp        jsonHexUint64  `json:"timestamp"`
	ExtraData        jsonHexBytes   `json:"extraData"`
	BaseFeePerGas    *jsonHexBig    `json:"baseFeePerGas"`
	BlockHash        types.Hash     `json:"blockHash"`
	Transactions     []jsonHexBytes `json:"transactions"`
	Withdrawals      []*Withdrawal  `json:"withdrawals"`
	BlobGasUsed      jsonHexUint64  `json:"blobGasUsed"`
	ExcessBlobGas    jsonHexUint64  `json:"excessBlobGas"`
	BlobCommitments  []types.Hash   `json:"blobCommitments"`
	ProofSubmissions []jsonHexBytes `json:"proofSubmissions"`
	ShieldedResults  []types.Hash   `json:"shieldedResults"`
}

func (p ExecutionPayloadV7) MarshalJSON() ([]byte, error) {
	withdrawals := p.Withdrawals
	if withdrawals == nil {
		withdrawals = []*Withdrawal{}
	}
	return json.Marshal(executionPayloadV7JSON{
		ParentHash:       p.ParentHash,
		FeeRecipient:     p.FeeRecipient,
		StateRoot:        p.StateRoot,
		ReceiptsRoot:     p.ReceiptsRoot,
		LogsBloom:        p.LogsBloom,
		PrevRandao:       p.PrevRandao,
		BlockNumber:      jsonHexUint64(p.BlockNumber),
		GasLimit:         jsonHexUint64(p.GasLimit),
		GasUsed:          jsonHexUint64(p.GasUsed),
		Timestamp:        jsonHexUint64(p.Timestamp),
		ExtraData:        jsonHexBytes(p.ExtraData),
		BaseFeePerGas:    newJSONHexBig(p.BaseFeePerGas),
		BlockHash:        p.BlockHash,
		Transactions:     toHexBytesSliceLocal(p.Transactions),
		Withdrawals:      withdrawals,
		BlobGasUsed:      jsonHexUint64(p.BlobGasUsed),
		ExcessBlobGas:    jsonHexUint64(p.ExcessBlobGas),
		BlobCommitments:  p.BlobCommitments,
		ProofSubmissions: toHexBytesSliceLocal(p.ProofSubmissions),
		ShieldedResults:  p.ShieldedResults,
	})
}

func (p *ExecutionPayloadV7) UnmarshalJSON(data []byte) error {
	var j executionPayloadV7JSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.ParentHash = j.ParentHash
	p.FeeRecipient = j.FeeRecipient
	p.StateRoot = j.StateRoot
	p.ReceiptsRoot = j.ReceiptsRoot
	p.LogsBloom = j.LogsBloom
	p.PrevRandao = j.PrevRandao
	p.BlockNumber = uint64(j.BlockNumber)
	p.GasLimit = uint64(j.GasLimit)
	p.GasUsed = uint64(j.GasUsed)
	p.Timestamp = uint64(j.Timestamp)
	p.ExtraData = []byte(j.ExtraData)
	p.BaseFeePerGas = toJSONBigInt(j.BaseFeePerGas)
	p.BlockHash = j.BlockHash
	p.Transactions = fromHexBytesSliceLocal(j.Transactions)
	p.Withdrawals = j.Withdrawals
	p.BlobGasUsed = uint64(j.BlobGasUsed)
	p.ExcessBlobGas = uint64(j.ExcessBlobGas)
	p.BlobCommitments = j.BlobCommitments
	p.ProofSubmissions = fromHexBytesSliceLocal(j.ProofSubmissions)
	p.ShieldedResults = j.ShieldedResults
	return nil
}

// ── PayloadAttributesV7 (V3 + DA/proof/shielded fields) ─────────────────────

type payloadAttributesV7JSON struct {
	Timestamp             jsonHexUint64      `json:"timestamp"`
	PrevRandao            types.Hash         `json:"prevRandao"`
	SuggestedFeeRecipient types.Address      `json:"suggestedFeeRecipient"`
	Withdrawals           []*Withdrawal      `json:"withdrawals"`
	ParentBeaconBlockRoot types.Hash         `json:"parentBeaconBlockRoot"`
	DALayerConfig         *DALayerConfig     `json:"daLayerConfig,omitempty"`
	ProofRequirements     *ProofRequirements `json:"proofRequirements,omitempty"`
	ShieldedTxs           []jsonHexBytes     `json:"shieldedTxs,omitempty"`
}

func (p PayloadAttributesV7) MarshalJSON() ([]byte, error) {
	return json.Marshal(payloadAttributesV7JSON{
		Timestamp:             jsonHexUint64(p.Timestamp),
		PrevRandao:            p.PrevRandao,
		SuggestedFeeRecipient: p.SuggestedFeeRecipient,
		Withdrawals:           p.Withdrawals,
		ParentBeaconBlockRoot: p.ParentBeaconBlockRoot,
		DALayerConfig:         p.DALayerConfig,
		ProofRequirements:     p.ProofRequirements,
		ShieldedTxs:           toHexBytesSliceLocal(p.ShieldedTxs),
	})
}

func (p *PayloadAttributesV7) UnmarshalJSON(data []byte) error {
	var j payloadAttributesV7JSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.Timestamp = uint64(j.Timestamp)
	p.PrevRandao = j.PrevRandao
	p.SuggestedFeeRecipient = j.SuggestedFeeRecipient
	p.Withdrawals = j.Withdrawals
	p.ParentBeaconBlockRoot = j.ParentBeaconBlockRoot
	p.DALayerConfig = j.DALayerConfig
	p.ProofRequirements = j.ProofRequirements
	p.ShieldedTxs = fromHexBytesSliceLocal(j.ShieldedTxs)
	return nil
}

// ── GlamsterdamPayloadAttributes ─────────────────────────────────────────────

type glamsterdamPayloadAttributesJSON struct {
	Timestamp             jsonHexUint64 `json:"timestamp"`
	PrevRandao            types.Hash    `json:"prevRandao"`
	SuggestedFeeRecipient types.Address `json:"suggestedFeeRecipient"`
	Withdrawals           []*Withdrawal `json:"withdrawals"`
	ParentBeaconBlockRoot types.Hash    `json:"parentBeaconBlockRoot"`
	TargetBlobCount       jsonHexUint64 `json:"targetBlobCount"`
	SlotNumber            jsonHexUint64 `json:"slotNumber"`
}

func (p GlamsterdamPayloadAttributes) MarshalJSON() ([]byte, error) {
	return json.Marshal(glamsterdamPayloadAttributesJSON{
		Timestamp:             jsonHexUint64(p.Timestamp),
		PrevRandao:            p.PrevRandao,
		SuggestedFeeRecipient: p.SuggestedFeeRecipient,
		Withdrawals:           p.Withdrawals,
		ParentBeaconBlockRoot: p.ParentBeaconBlockRoot,
		TargetBlobCount:       jsonHexUint64(p.TargetBlobCount),
		SlotNumber:            jsonHexUint64(p.SlotNumber),
	})
}

func (p *GlamsterdamPayloadAttributes) UnmarshalJSON(data []byte) error {
	var j glamsterdamPayloadAttributesJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.Timestamp = uint64(j.Timestamp)
	p.PrevRandao = j.PrevRandao
	p.SuggestedFeeRecipient = j.SuggestedFeeRecipient
	p.Withdrawals = j.Withdrawals
	p.ParentBeaconBlockRoot = j.ParentBeaconBlockRoot
	p.TargetBlobCount = uint64(j.TargetBlobCount)
	p.SlotNumber = uint64(j.SlotNumber)
	return nil
}
