package payload

import (
	"github.com/eth2030/eth2030/core/types"
)

// ExecutionPayloadBodyV1 is the response body for engine_getPayloadBodiesByHash/RangeV1.
// Per Shanghai spec, it contains transactions and withdrawals only.
type ExecutionPayloadBodyV1 struct {
	Transactions [][]byte      `json:"transactions"`
	Withdrawals  []*Withdrawal `json:"withdrawals"`
}

// ExecutionPayloadBodyV2 is the response body for engine_getPayloadBodiesByHash/RangeV2.
// It extends V1 (transactions + withdrawals) with a blockAccessList field (EIP-7928).
type ExecutionPayloadBodyV2 struct {
	Transactions    [][]byte      `json:"transactions"`
	Withdrawals     []*Withdrawal `json:"withdrawals"`
	BlockAccessList []byte        `json:"blockAccessList,omitempty"`
}

// BlockToPayloadBodyV1 converts a types.Block to an ExecutionPayloadBodyV1.
func BlockToPayloadBodyV1(block *types.Block) *ExecutionPayloadBodyV1 {
	txs := make([][]byte, 0, len(block.Transactions()))
	for _, tx := range block.Transactions() {
		enc, err := tx.EncodeRLP()
		if err == nil {
			txs = append(txs, enc)
		}
	}
	ws := make([]*Withdrawal, 0, len(block.Withdrawals()))
	for _, w := range block.Withdrawals() {
		if w != nil {
			ws = append(ws, &Withdrawal{
				Index:          w.Index,
				ValidatorIndex: w.ValidatorIndex,
				Address:        w.Address,
				Amount:         w.Amount,
			})
		}
	}
	return &ExecutionPayloadBodyV1{
		Transactions: txs,
		Withdrawals:  ws,
	}
}

// BlockToPayloadBodyV2 converts a types.Block to an ExecutionPayloadBodyV2.
func BlockToPayloadBodyV2(block *types.Block) *ExecutionPayloadBodyV2 {
	v1 := BlockToPayloadBodyV1(block)
	return &ExecutionPayloadBodyV2{
		Transactions: v1.Transactions,
		Withdrawals:  v1.Withdrawals,
	}
}

// V2ToV1 converts an ExecutionPayloadBodyV2 to ExecutionPayloadBodyV1
// by stripping the blockAccessList field.
func V2ToV1(v2 *ExecutionPayloadBodyV2) *ExecutionPayloadBodyV1 {
	if v2 == nil {
		return nil
	}
	return &ExecutionPayloadBodyV1{
		Transactions: v2.Transactions,
		Withdrawals:  v2.Withdrawals,
	}
}

// V2SliceToV1 converts a slice of ExecutionPayloadBodyV2 to ExecutionPayloadBodyV1.
func V2SliceToV1(v2s []*ExecutionPayloadBodyV2) []*ExecutionPayloadBodyV1 {
	result := make([]*ExecutionPayloadBodyV1, len(v2s))
	for i, v2 := range v2s {
		result[i] = V2ToV1(v2)
	}
	return result
}
