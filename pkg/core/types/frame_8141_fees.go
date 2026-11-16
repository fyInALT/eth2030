package types

import (
	"errors"
	"math/big"
)

// FramePayment is the upfront payment calculated for a particular block.
type FramePayment struct {
	maxGas  uint64
	escrow  *big.Int
	price   *big.Int
	baseFee *big.Int
	blobFee *big.Int
	tipCap  *big.Int
}

func (payment *FramePayment) Escrow() *big.Int   { return copyBigInt(payment.escrow) }
func (payment *FramePayment) GasPrice() *big.Int { return copyBigInt(payment.price) }

// FrameSettlement separates the burned fee, proposer tip and payer refund.
type FrameSettlement struct {
	Burned *big.Int
	Tip    *big.Int
	Refund *big.Int
}

// Payment computes the maximum charge and effective GASPRICE under EIP-8141.
// The same max_fee budget covers execution, state and blob gas.
func (tx *FrameTx8141) Payment(baseFee, blobBaseFee *big.Int) (*FramePayment, error) {
	if !frameUint256(tx.Fees.MaxFee) || !frameUint256(tx.Fees.MaxPriorityFeePerGas) ||
		!frameUint256(baseFee) || !frameUint256(blobBaseFee) {
		return nil, errors.New("EIP-8141 payment requires unsigned 256-bit fees")
	}
	budget, err := tx.GasBudget()
	if err != nil {
		return nil, err
	}
	gas := new(big.Int).SetUint64(budget.MaxGas)
	blobGas := new(big.Int).Mul(new(big.Int).SetUint64(uint64(len(tx.BlobVersionedHashes))), big.NewInt(131072))
	blobFee := new(big.Int).Mul(blobGas, bigOrZero(blobBaseFee))
	required := new(big.Int).Mul(gas, bigOrZero(baseFee))
	required.Add(required, blobFee)
	headroom := new(big.Int).Sub(bigOrZero(tx.Fees.MaxFee), required)
	if headroom.Sign() < 0 {
		return nil, ErrFrame8141FeeBudget
	}
	tipCap := bigOrZero(tx.Fees.MaxPriorityFeePerGas)
	maxTip := new(big.Int).Mul(gas, tipCap)
	if maxTip.Cmp(headroom) > 0 {
		maxTip.Set(headroom)
	}
	price := new(big.Int).Div(new(big.Int).Set(maxTip), gas)
	price.Add(price, bigOrZero(baseFee))
	return &FramePayment{
		maxGas: budget.MaxGas, escrow: new(big.Int).Add(required, maxTip),
		price: price, baseFee: new(big.Int).Set(bigOrZero(baseFee)),
		blobFee: blobFee, tipCap: new(big.Int).Set(tipCap),
	}, nil
}

func (payment *FramePayment) Settle(gasUsed uint64) (FrameSettlement, error) {
	if gasUsed > payment.maxGas {
		return FrameSettlement{}, errors.New("EIP-8141 settlement exceeds maximum gas")
	}
	gas := new(big.Int).SetUint64(gasUsed)
	burned := new(big.Int).Mul(gas, payment.baseFee)
	burned.Add(burned, payment.blobFee)
	refund := new(big.Int).Sub(payment.escrow, burned)
	tip := new(big.Int).Mul(gas, payment.tipCap)
	if tip.Cmp(refund) > 0 {
		tip.Set(refund)
	}
	return FrameSettlement{Burned: burned, Tip: tip, Refund: refund.Sub(refund, tip)}, nil
}
