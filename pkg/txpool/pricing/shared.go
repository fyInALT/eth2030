package pricing

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

// PriceBump is the minimum price bump percentage required for a replacement.
const PriceBump = 10

// EffectiveGasPrice returns the effective gas price of a transaction given the
// current base fee. For EIP-1559 transactions: min(feeCap, baseFee+tipCap);
// for legacy transactions: gasPrice.
func EffectiveGasPrice(tx *types.Transaction, baseFee *big.Int) *big.Int {
	if feeCap := tx.GasFeeCap(); feeCap != nil && baseFee != nil {
		tip := tx.GasTipCap()
		if tip == nil {
			tip = new(big.Int)
		}
		effective := new(big.Int).Add(baseFee, tip)
		if effective.Cmp(feeCap) > 0 {
			effective.Set(feeCap)
		}
		return effective
	}
	if gp := tx.GasPrice(); gp != nil {
		return new(big.Int).Set(gp)
	}
	return new(big.Int)
}

// EffectiveTip computes the miner tip for a transaction given the current
// baseFee. For legacy transactions, tip = GasPrice - baseFee (clamped to 0).
// For EIP-1559 transactions: min(GasTipCap, GasFeeCap - baseFee).
// If baseFee is nil, returns GasTipCap (or GasPrice for legacy).
func EffectiveTip(tx *types.Transaction, baseFee *big.Int) *big.Int {
	if tx == nil {
		return new(big.Int)
	}
	if baseFee == nil {
		tip := tx.GasTipCap()
		if tip == nil {
			return new(big.Int)
		}
		return new(big.Int).Set(tip)
	}
	feeCap := tx.GasFeeCap()
	tipCap := tx.GasTipCap()
	if feeCap == nil {
		feeCap = tx.GasPrice()
	}
	if tipCap == nil {
		tipCap = tx.GasPrice()
	}
	if feeCap == nil {
		return new(big.Int)
	}
	if tipCap == nil {
		tipCap = new(big.Int)
	}
	availableTip := new(big.Int).Sub(feeCap, baseFee)
	if availableTip.Sign() < 0 {
		return new(big.Int)
	}
	if tipCap.Cmp(availableTip) < 0 {
		return new(big.Int).Set(tipCap)
	}
	return availableTip
}

// cloneBigInt returns a copy of v, or nil if v is nil.
func cloneBigInt(v *big.Int) *big.Int {
	if v == nil {
		return nil
	}
	return new(big.Int).Set(v)
}
