package txpool

// pricing_compat.go re-exports types from txpool/pricing for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/txpool/pricing"
)

// Pricing type aliases.
type (
	BumperConfig            = pricing.BumperConfig
	BumperBlockFeeData      = pricing.BumperBlockFeeData
	FeeSuggestion           = pricing.FeeSuggestion
	TieredSuggestion        = pricing.TieredSuggestion
	PriceBumper             = pricing.PriceBumper
	PriceHeap               = pricing.PriceHeap
	PriorityPoolConfig      = pricing.PriorityPoolConfig
	PriorityEntry           = pricing.PriorityEntry
	PriorityPool            = pricing.PriorityPool
	TxPriorityQueueConfig   = pricing.TxPriorityQueueConfig
	QueueEntry              = pricing.QueueEntry
	EffectiveTipCalculator  = pricing.EffectiveTipCalculator
	TxPriorityQueue         = pricing.TxPriorityQueue
)

// Pricing constants.
const (
	TierUrgent              = pricing.TierUrgent
	TierFast                = pricing.TierFast
	TierStandard            = pricing.TierStandard
	TierSlow                = pricing.TierSlow
	DefaultFeeHistoryDepth  = pricing.DefaultFeeHistoryDepth
	BumperMinSuggestedTip   = pricing.BumperMinSuggestedTip
	DefaultBaseFeeMultiplier = pricing.DefaultBaseFeeMultiplier
)

// Pricing error variables.
var (
	ErrPriorityBelowMin  = pricing.ErrPriorityBelowMin
	ErrPriorityDuplicate = pricing.ErrPriorityDuplicate
	ErrPQFull            = pricing.ErrPQFull
	ErrPQDuplicate       = pricing.ErrPQDuplicate
	ErrPQNotFound        = pricing.ErrPQNotFound
	ErrPQNilTx           = pricing.ErrPQNilTx
	ErrPQNonceGap        = pricing.ErrPQNonceGap
)

// Pricing function wrappers.
func DefaultBumperConfig() BumperConfig { return pricing.DefaultBumperConfig() }
func NewPriceBumper(config BumperConfig) *PriceBumper {
	return pricing.NewPriceBumper(config)
}
func NewPriceHeap(baseFee *big.Int) *PriceHeap { return pricing.NewPriceHeap(baseFee) }
func NewPriorityPool(config PriorityPoolConfig) *PriorityPool {
	return pricing.NewPriorityPool(config)
}
func NewEffectiveTipCalculator(baseFee, blobBaseFee *big.Int) *EffectiveTipCalculator {
	return pricing.NewEffectiveTipCalculator(baseFee, blobBaseFee)
}
func NewTxPriorityQueue(config TxPriorityQueueConfig) *TxPriorityQueue {
	return pricing.NewTxPriorityQueue(config)
}

// Ensure types.Transaction is used to avoid import cycle.
var _ = (*types.Transaction)(nil)
