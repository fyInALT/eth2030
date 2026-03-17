package node

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

// extractBlockTips returns the effective priority fee (tip) for each
// transaction in the block, given the block's base fee.
func extractBlockTips(txs []*types.Transaction, baseFee *big.Int) []*big.Int {
	tips := make([]*big.Int, 0, len(txs))
	if baseFee == nil {
		baseFee = new(big.Int)
	}
	for _, tx := range txs {
		var tip *big.Int
		switch tx.Type() {
		case types.DynamicFeeTxType:
			tipCap := tx.GasTipCap()
			feeCap := tx.GasFeeCap()
			if tipCap == nil || feeCap == nil {
				continue
			}
			effective := new(big.Int).Sub(feeCap, baseFee)
			if effective.Sign() < 0 {
				continue
			}
			tip = tipCap
			if effective.Cmp(tipCap) < 0 {
				tip = effective
			}
		default:
			gp := tx.GasPrice()
			if gp == nil {
				continue
			}
			tip = new(big.Int).Sub(gp, baseFee)
			if tip.Sign() < 0 {
				continue
			}
		}
		tips = append(tips, new(big.Int).Set(tip))
	}
	return tips
}