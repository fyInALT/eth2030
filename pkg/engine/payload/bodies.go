package payload

import (
	"github.com/eth2030/eth2030/core/types"
)

// ExecutionPayloadBodyV2 is the response body for engine_getPayloadBodiesByHash/RangeV2.
// It extends V1 (transactions + withdrawals) with a blockAccessList field (EIP-7928).
type ExecutionPayloadBodyV2 struct {
	Transactions    [][]byte      `json:"transactions"`
	Withdrawals     []*Withdrawal `json:"withdrawals"`
	BlockAccessList []byte        `json:"blockAccessList,omitempty"`
}

// BlockToPayloadBodyV2 converts a types.Block to an ExecutionPayloadBodyV2.
func BlockToPayloadBodyV2(block *types.Block) *ExecutionPayloadBodyV2 {
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
	return &ExecutionPayloadBodyV2{
		Transactions: txs,
		Withdrawals:  ws,
	}
}
