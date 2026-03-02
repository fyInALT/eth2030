// json.go provides custom JSON marshal/unmarshal for Engine API types.
// Per the Engine API spec, all integer quantities are hex-encoded strings
// (e.g. "0x1" not 1), and byte arrays are 0x-prefixed hex strings.
// This mirrors what go-ethereum achieves via hexutil.Uint64 / hexutil.Bytes.
package engine

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/eth2030/eth2030/core/types"
)

// ── primitive hex types ──────────────────────────────────────────────────────

// hexUint64 marshals/unmarshals as a 0x-prefixed hex string.
type hexUint64 uint64

func (h hexUint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("0x%x", uint64(h)))
}

func (h *hexUint64) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// Accept plain decimal number for backward compatibility with tests.
		var n uint64
		if err2 := json.Unmarshal(data, &n); err2 != nil {
			return fmt.Errorf("hexUint64: %w", err)
		}
		*h = hexUint64(n)
		return nil
	}
	n, err := parseHexUint64(s)
	if err != nil {
		return err
	}
	*h = hexUint64(n)
	return nil
}

func parseHexUint64(s string) (uint64, error) {
	orig := s
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	if s == "" {
		return 0, nil
	}
	n := new(big.Int)
	if _, ok := n.SetString(s, 16); !ok {
		return 0, fmt.Errorf("hexUint64: invalid hex %q", orig)
	}
	return n.Uint64(), nil
}

// hexBytes marshals/unmarshals as a 0x-prefixed hex string.
type hexBytes []byte

func (h hexBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal("0x" + hex.EncodeToString(h))
}

func (h *hexBytes) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("hexBytes: %w", err)
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
		return fmt.Errorf("hexBytes: %w", err)
	}
	*h = b
	return nil
}

// hexBig marshals/unmarshals *big.Int as a 0x-prefixed hex string.
type hexBig struct{ *big.Int }

func newHexBig(v *big.Int) *hexBig {
	if v == nil {
		return nil
	}
	return &hexBig{v}
}

func (h hexBig) MarshalJSON() ([]byte, error) {
	if h.Int == nil {
		return []byte("null"), nil
	}
	return json.Marshal("0x" + h.Int.Text(16))
}

func (h *hexBig) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// Accept plain decimal number.
		n := new(big.Int)
		if err2 := json.Unmarshal(data, n); err2 != nil {
			return fmt.Errorf("hexBig: %w", err)
		}
		h.Int = n
		return nil
	}
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	n := new(big.Int)
	if _, ok := n.SetString(s, 16); !ok {
		return fmt.Errorf("hexBig: invalid hex")
	}
	h.Int = n
	return nil
}

// toBigInt safely extracts *big.Int from a *hexBig.
func toBigInt(h *hexBig) *big.Int {
	if h == nil {
		return nil
	}
	return h.Int
}

// toHexBytesSlice converts [][]byte → []hexBytes.
func toHexBytesSlice(in [][]byte) []hexBytes {
	if in == nil {
		return nil
	}
	out := make([]hexBytes, len(in))
	for i, b := range in {
		out[i] = hexBytes(b)
	}
	return out
}

// fromHexBytesSlice converts []hexBytes → [][]byte.
func fromHexBytesSlice(in []hexBytes) [][]byte {
	if in == nil {
		return nil
	}
	out := make([][]byte, len(in))
	for i, b := range in {
		out[i] = []byte(b)
	}
	return out
}

// ── Withdrawal ───────────────────────────────────────────────────────────────

type withdrawalJSON struct {
	Index          hexUint64     `json:"index"`
	ValidatorIndex hexUint64     `json:"validatorIndex"`
	Address        types.Address `json:"address"`
	Amount         hexUint64     `json:"amount"`
}

func (w Withdrawal) MarshalJSON() ([]byte, error) {
	return json.Marshal(withdrawalJSON{
		Index:          hexUint64(w.Index),
		ValidatorIndex: hexUint64(w.ValidatorIndex),
		Address:        w.Address,
		Amount:         hexUint64(w.Amount),
	})
}

func (w *Withdrawal) UnmarshalJSON(data []byte) error {
	var j withdrawalJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	w.Index = uint64(j.Index)
	w.ValidatorIndex = uint64(j.ValidatorIndex)
	w.Address = j.Address
	w.Amount = uint64(j.Amount)
	return nil
}

// ── BlobsBundleV1 ────────────────────────────────────────────────────────────

type blobsBundleV1JSON struct {
	Commitments []hexBytes `json:"commitments"`
	Proofs      []hexBytes `json:"proofs"`
	Blobs       []hexBytes `json:"blobs"`
}

func (b BlobsBundleV1) MarshalJSON() ([]byte, error) {
	return json.Marshal(blobsBundleV1JSON{
		Commitments: toHexBytesSlice(b.Commitments),
		Proofs:      toHexBytesSlice(b.Proofs),
		Blobs:       toHexBytesSlice(b.Blobs),
	})
}

func (b *BlobsBundleV1) UnmarshalJSON(data []byte) error {
	var j blobsBundleV1JSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	b.Commitments = fromHexBytesSlice(j.Commitments)
	b.Proofs = fromHexBytesSlice(j.Proofs)
	b.Blobs = fromHexBytesSlice(j.Blobs)
	return nil
}

// ── ExecutionPayloadV3 (covers V1 + V2 + V3 fields) ─────────────────────────

type executionPayloadV3JSON struct {
	ParentHash    types.Hash    `json:"parentHash"`
	FeeRecipient  types.Address `json:"feeRecipient"`
	StateRoot     types.Hash    `json:"stateRoot"`
	ReceiptsRoot  types.Hash    `json:"receiptsRoot"`
	LogsBloom     types.Bloom   `json:"logsBloom"`
	PrevRandao    types.Hash    `json:"prevRandao"`
	BlockNumber   hexUint64     `json:"blockNumber"`
	GasLimit      hexUint64     `json:"gasLimit"`
	GasUsed       hexUint64     `json:"gasUsed"`
	Timestamp     hexUint64     `json:"timestamp"`
	ExtraData     hexBytes      `json:"extraData"`
	BaseFeePerGas *hexBig       `json:"baseFeePerGas"`
	BlockHash     types.Hash    `json:"blockHash"`
	Transactions  []hexBytes    `json:"transactions"`
	Withdrawals   []*Withdrawal `json:"withdrawals"`
	BlobGasUsed   hexUint64     `json:"blobGasUsed"`
	ExcessBlobGas hexUint64     `json:"excessBlobGas"`
}

// toV3JSON builds the shadow struct for a V3 payload. Used by V4/V5 too.
func (p *ExecutionPayloadV3) toV3JSON() executionPayloadV3JSON {
	return executionPayloadV3JSON{
		ParentHash:    p.ParentHash,
		FeeRecipient:  p.FeeRecipient,
		StateRoot:     p.StateRoot,
		ReceiptsRoot:  p.ReceiptsRoot,
		LogsBloom:     p.LogsBloom,
		PrevRandao:    p.PrevRandao,
		BlockNumber:   hexUint64(p.BlockNumber),
		GasLimit:      hexUint64(p.GasLimit),
		GasUsed:       hexUint64(p.GasUsed),
		Timestamp:     hexUint64(p.Timestamp),
		ExtraData:     hexBytes(p.ExtraData),
		BaseFeePerGas: newHexBig(p.BaseFeePerGas),
		BlockHash:     p.BlockHash,
		Transactions:  toHexBytesSlice(p.Transactions),
		Withdrawals:   p.Withdrawals,
		BlobGasUsed:   hexUint64(p.BlobGasUsed),
		ExcessBlobGas: hexUint64(p.ExcessBlobGas),
	}
}

// applyV3JSON populates a V3 payload from its shadow struct.
func (p *ExecutionPayloadV3) applyV3JSON(j executionPayloadV3JSON) {
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
	p.BaseFeePerGas = toBigInt(j.BaseFeePerGas)
	p.BlockHash = j.BlockHash
	p.Transactions = fromHexBytesSlice(j.Transactions)
	p.Withdrawals = j.Withdrawals
	p.BlobGasUsed = uint64(j.BlobGasUsed)
	p.ExcessBlobGas = uint64(j.ExcessBlobGas)
}

func (p ExecutionPayloadV3) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.toV3JSON())
}

func (p *ExecutionPayloadV3) UnmarshalJSON(data []byte) error {
	var j executionPayloadV3JSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.applyV3JSON(j)
	return nil
}

// ── ExecutionPayloadV4 (V3 + executionRequests) ──────────────────────────────

type executionPayloadV4JSON struct {
	executionPayloadV3JSON
	ExecutionRequests []hexBytes `json:"executionRequests"`
}

func (p ExecutionPayloadV4) MarshalJSON() ([]byte, error) {
	return json.Marshal(executionPayloadV4JSON{
		executionPayloadV3JSON: p.ExecutionPayloadV3.toV3JSON(),
		ExecutionRequests:      toHexBytesSlice(p.ExecutionRequests),
	})
}

func (p *ExecutionPayloadV4) UnmarshalJSON(data []byte) error {
	var j executionPayloadV4JSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.ExecutionPayloadV3.applyV3JSON(j.executionPayloadV3JSON)
	p.ExecutionRequests = fromHexBytesSlice(j.ExecutionRequests)
	return nil
}

// ── ExecutionPayloadV5 (V4 + blockAccessList) ────────────────────────────────

type executionPayloadV5JSON struct {
	executionPayloadV3JSON
	ExecutionRequests []hexBytes      `json:"executionRequests"`
	BlockAccessList   json.RawMessage `json:"blockAccessList,omitempty"`
}

func (p ExecutionPayloadV5) MarshalJSON() ([]byte, error) {
	return json.Marshal(executionPayloadV5JSON{
		executionPayloadV3JSON: p.ExecutionPayloadV3.toV3JSON(),
		ExecutionRequests:      toHexBytesSlice(p.ExecutionRequests),
		BlockAccessList:        p.BlockAccessList,
	})
}

func (p *ExecutionPayloadV5) UnmarshalJSON(data []byte) error {
	var j executionPayloadV5JSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.ExecutionPayloadV3.applyV3JSON(j.executionPayloadV3JSON)
	p.ExecutionRequests = fromHexBytesSlice(j.ExecutionRequests)
	p.BlockAccessList = j.BlockAccessList
	return nil
}

// ── ExecutionPayloadV7 (V3 + 2030 roadmap fields) ────────────────────────────

type executionPayloadV7JSON struct {
	executionPayloadV3JSON
	BlobCommitments  []types.Hash `json:"blobCommitments"`
	ProofSubmissions []hexBytes   `json:"proofSubmissions"`
	ShieldedResults  []types.Hash `json:"shieldedResults"`
}

func (p ExecutionPayloadV7) MarshalJSON() ([]byte, error) {
	return json.Marshal(executionPayloadV7JSON{
		executionPayloadV3JSON: p.ExecutionPayloadV3.toV3JSON(),
		BlobCommitments:        p.BlobCommitments,
		ProofSubmissions:       toHexBytesSlice(p.ProofSubmissions),
		ShieldedResults:        p.ShieldedResults,
	})
}

func (p *ExecutionPayloadV7) UnmarshalJSON(data []byte) error {
	var j executionPayloadV7JSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.ExecutionPayloadV3.applyV3JSON(j.executionPayloadV3JSON)
	p.BlobCommitments = j.BlobCommitments
	p.ProofSubmissions = fromHexBytesSlice(j.ProofSubmissions)
	p.ShieldedResults = j.ShieldedResults
	return nil
}

// ── PayloadAttributesV3 (covers V1 + V2 + V3 fields) ────────────────────────

type payloadAttributesV3JSON struct {
	Timestamp             hexUint64     `json:"timestamp"`
	PrevRandao            types.Hash    `json:"prevRandao"`
	SuggestedFeeRecipient types.Address `json:"suggestedFeeRecipient"`
	Withdrawals           []*Withdrawal `json:"withdrawals"`
	ParentBeaconBlockRoot types.Hash    `json:"parentBeaconBlockRoot"`
}

// toV3AttrsJSON builds the shadow struct for V3 attributes. Used by V4/V7 too.
func (p *PayloadAttributesV3) toV3AttrsJSON() payloadAttributesV3JSON {
	return payloadAttributesV3JSON{
		Timestamp:             hexUint64(p.Timestamp),
		PrevRandao:            p.PrevRandao,
		SuggestedFeeRecipient: p.SuggestedFeeRecipient,
		Withdrawals:           p.Withdrawals,
		ParentBeaconBlockRoot: p.ParentBeaconBlockRoot,
	}
}

// applyV3AttrsJSON populates V3 attributes from its shadow struct.
func (p *PayloadAttributesV3) applyV3AttrsJSON(j payloadAttributesV3JSON) {
	p.Timestamp = uint64(j.Timestamp)
	p.PrevRandao = j.PrevRandao
	p.SuggestedFeeRecipient = j.SuggestedFeeRecipient
	p.Withdrawals = j.Withdrawals
	p.ParentBeaconBlockRoot = j.ParentBeaconBlockRoot
}

func (p PayloadAttributesV3) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.toV3AttrsJSON())
}

func (p *PayloadAttributesV3) UnmarshalJSON(data []byte) error {
	var j payloadAttributesV3JSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.applyV3AttrsJSON(j)
	return nil
}

// ── PayloadAttributesV4 (V3 + slotNumber + inclusionList) ───────────────────

type payloadAttributesV4JSON struct {
	payloadAttributesV3JSON
	SlotNumber                hexUint64  `json:"slotNumber"`
	InclusionListTransactions []hexBytes `json:"inclusionListTransactions,omitempty"`
}

func (p PayloadAttributesV4) MarshalJSON() ([]byte, error) {
	return json.Marshal(payloadAttributesV4JSON{
		payloadAttributesV3JSON:   p.PayloadAttributesV3.toV3AttrsJSON(),
		SlotNumber:                hexUint64(p.SlotNumber),
		InclusionListTransactions: toHexBytesSlice(p.InclusionListTransactions),
	})
}

func (p *PayloadAttributesV4) UnmarshalJSON(data []byte) error {
	var j payloadAttributesV4JSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.PayloadAttributesV3.applyV3AttrsJSON(j.payloadAttributesV3JSON)
	p.SlotNumber = uint64(j.SlotNumber)
	p.InclusionListTransactions = fromHexBytesSlice(j.InclusionListTransactions)
	return nil
}

// ── PayloadAttributesV7 (V3 + DA/proof/shielded fields) ─────────────────────

type payloadAttributesV7JSON struct {
	payloadAttributesV3JSON
	DALayerConfig     *DALayerConfig     `json:"daLayerConfig,omitempty"`
	ProofRequirements *ProofRequirements `json:"proofRequirements,omitempty"`
	ShieldedTxs       []hexBytes         `json:"shieldedTxs,omitempty"`
}

func (p PayloadAttributesV7) MarshalJSON() ([]byte, error) {
	return json.Marshal(payloadAttributesV7JSON{
		payloadAttributesV3JSON: p.PayloadAttributesV3.toV3AttrsJSON(),
		DALayerConfig:           p.DALayerConfig,
		ProofRequirements:       p.ProofRequirements,
		ShieldedTxs:             toHexBytesSlice(p.ShieldedTxs),
	})
}

func (p *PayloadAttributesV7) UnmarshalJSON(data []byte) error {
	var j payloadAttributesV7JSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	p.PayloadAttributesV3.applyV3AttrsJSON(j.payloadAttributesV3JSON)
	p.DALayerConfig = j.DALayerConfig
	p.ProofRequirements = j.ProofRequirements
	p.ShieldedTxs = fromHexBytesSlice(j.ShieldedTxs)
	return nil
}

// ── GlamsterdamPayloadAttributes ─────────────────────────────────────────────

type glamsterdamPayloadAttributesJSON struct {
	Timestamp             hexUint64     `json:"timestamp"`
	PrevRandao            types.Hash    `json:"prevRandao"`
	SuggestedFeeRecipient types.Address `json:"suggestedFeeRecipient"`
	Withdrawals           []*Withdrawal `json:"withdrawals"`
	ParentBeaconBlockRoot types.Hash    `json:"parentBeaconBlockRoot"`
	TargetBlobCount       hexUint64     `json:"targetBlobCount"`
	SlotNumber            hexUint64     `json:"slotNumber"`
}

func (p GlamsterdamPayloadAttributes) MarshalJSON() ([]byte, error) {
	return json.Marshal(glamsterdamPayloadAttributesJSON{
		Timestamp:             hexUint64(p.Timestamp),
		PrevRandao:            p.PrevRandao,
		SuggestedFeeRecipient: p.SuggestedFeeRecipient,
		Withdrawals:           p.Withdrawals,
		ParentBeaconBlockRoot: p.ParentBeaconBlockRoot,
		TargetBlobCount:       hexUint64(p.TargetBlobCount),
		SlotNumber:            hexUint64(p.SlotNumber),
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

// ── GetPayload responses (blockValue as hex, executionRequests as hex) ───────

type getPayloadV3ResponseJSON struct {
	ExecutionPayload *ExecutionPayloadV3 `json:"executionPayload"`
	BlockValue       *hexBig             `json:"blockValue"`
	BlobsBundle      *BlobsBundleV1      `json:"blobsBundle"`
	Override         bool                `json:"shouldOverrideBuilder"`
}

func (r GetPayloadV3Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(getPayloadV3ResponseJSON{
		ExecutionPayload: r.ExecutionPayload,
		BlockValue:       newHexBig(r.BlockValue),
		BlobsBundle:      r.BlobsBundle,
		Override:         r.Override,
	})
}

func (r *GetPayloadV3Response) UnmarshalJSON(data []byte) error {
	var j getPayloadV3ResponseJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	r.ExecutionPayload = j.ExecutionPayload
	r.BlockValue = toBigInt(j.BlockValue)
	r.BlobsBundle = j.BlobsBundle
	r.Override = j.Override
	return nil
}

type getPayloadV4ResponseJSON struct {
	ExecutionPayload  *ExecutionPayloadV3 `json:"executionPayload"`
	BlockValue        *hexBig             `json:"blockValue"`
	BlobsBundle       *BlobsBundleV1      `json:"blobsBundle"`
	Override          bool                `json:"shouldOverrideBuilder"`
	ExecutionRequests []hexBytes          `json:"executionRequests"`
}

func (r GetPayloadV4Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(getPayloadV4ResponseJSON{
		ExecutionPayload:  r.ExecutionPayload,
		BlockValue:        newHexBig(r.BlockValue),
		BlobsBundle:       r.BlobsBundle,
		Override:          r.Override,
		ExecutionRequests: toHexBytesSlice(r.ExecutionRequests),
	})
}

func (r *GetPayloadV4Response) UnmarshalJSON(data []byte) error {
	var j getPayloadV4ResponseJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	r.ExecutionPayload = j.ExecutionPayload
	r.BlockValue = toBigInt(j.BlockValue)
	r.BlobsBundle = j.BlobsBundle
	r.Override = j.Override
	r.ExecutionRequests = fromHexBytesSlice(j.ExecutionRequests)
	return nil
}

type getPayloadV6ResponseJSON struct {
	ExecutionPayload  *ExecutionPayloadV5 `json:"executionPayload"`
	BlockValue        *hexBig             `json:"blockValue"`
	BlobsBundle       *BlobsBundleV1      `json:"blobsBundle"`
	Override          bool                `json:"shouldOverrideBuilder"`
	ExecutionRequests []hexBytes          `json:"executionRequests"`
}

func (r GetPayloadV6Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(getPayloadV6ResponseJSON{
		ExecutionPayload:  r.ExecutionPayload,
		BlockValue:        newHexBig(r.BlockValue),
		BlobsBundle:       r.BlobsBundle,
		Override:          r.Override,
		ExecutionRequests: toHexBytesSlice(r.ExecutionRequests),
	})
}

func (r *GetPayloadV6Response) UnmarshalJSON(data []byte) error {
	var j getPayloadV6ResponseJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	r.ExecutionPayload = j.ExecutionPayload
	r.BlockValue = toBigInt(j.BlockValue)
	r.BlobsBundle = j.BlobsBundle
	r.Override = j.Override
	r.ExecutionRequests = fromHexBytesSlice(j.ExecutionRequests)
	return nil
}

// GetPayloadResponse (internal combined response) also needs hex BlockValue.
type getPayloadResponseJSON struct {
	ExecutionPayload *ExecutionPayloadV4 `json:"executionPayload"`
	BlockValue       *hexBig             `json:"blockValue"`
	BlobsBundle      *BlobsBundleV1      `json:"blobsBundle,omitempty"`
	Override         bool                `json:"shouldOverrideBuilder"`
}

func (r GetPayloadResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(getPayloadResponseJSON{
		ExecutionPayload: r.ExecutionPayload,
		BlockValue:       newHexBig(r.BlockValue),
		BlobsBundle:      r.BlobsBundle,
		Override:         r.Override,
	})
}

func (r *GetPayloadResponse) UnmarshalJSON(data []byte) error {
	var j getPayloadResponseJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	r.ExecutionPayload = j.ExecutionPayload
	r.BlockValue = toBigInt(j.BlockValue)
	r.BlobsBundle = j.BlobsBundle
	r.Override = j.Override
	return nil
}
