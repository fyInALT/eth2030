package queue

import (
	"errors"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

// PriceBump is the minimum price bump percentage required for a replacement.
const PriceBump = 10

// Queue errors.
var (
	ErrTxPoolFull             = errors.New("transaction pool is full")
	ErrReplacementUnderpriced = errors.New("replacement transaction underpriced")
	ErrSenderLimitExceeded    = errors.New("per-sender transaction limit exceeded")
)

// EffectiveGasPrice returns the effective gas price of a transaction
// given the current base fee. For EIP-1559 transactions it is
// min(feeCap, baseFee + tipCap); for legacy transactions it is gasPrice.
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

// isDynamic returns true if the transaction is an EIP-1559 dynamic fee tx.
func isDynamic(tx *types.Transaction) bool {
	return tx.GasFeeCap() != nil
}
