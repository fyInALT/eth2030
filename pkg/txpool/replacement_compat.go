package txpool

// replacement_compat.go re-exports types from txpool/replacement for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/txpool/replacement"
)

// Replacement type aliases.
type (
	ReplacementChainEntry = replacement.ReplacementChainEntry
	RBFPolicyConfig       = replacement.RBFPolicyConfig
	RBFPolicyEngine       = replacement.RBFPolicyEngine
	RBFStats              = replacement.RBFStats
	ReplacementPolicy     = replacement.ReplacementPolicy
	AccountPending        = replacement.AccountPending
)

// Replacement constants.
const (
	RBFMinFeeBump      = replacement.RBFMinFeeBump
	RBFMinTipBump      = replacement.RBFMinTipBump
	RBFBlobFeeBump     = replacement.RBFBlobFeeBump
	RBFBlobGasFeeBump  = replacement.RBFBlobGasFeeBump
	RBFMaxReplacements = replacement.RBFMaxReplacements
	RBFMaxChainDepth   = replacement.RBFMaxChainDepth
	DefaultMinPriceBump = replacement.DefaultMinPriceBump
	DefaultMaxPoolSize  = replacement.DefaultMaxPoolSize
	DefaultAccountSlots = replacement.DefaultAccountSlots
)

// Replacement error variables.
var (
	ErrRBFNilTx                = replacement.ErrRBFNilTx
	ErrRBFNonceMismatch        = replacement.ErrRBFNonceMismatch
	ErrRBFInsufficientFeeBump  = replacement.ErrRBFInsufficientFeeBump
	ErrRBFInsufficientTipBump  = replacement.ErrRBFInsufficientTipBump
	ErrRBFInsufficientBlobBump = replacement.ErrRBFInsufficientBlobBump
	ErrRBFMaxReplacements      = replacement.ErrRBFMaxReplacements
	ErrRBFMaxChainDepth        = replacement.ErrRBFMaxChainDepth
	ErrRBFTypeMismatch         = replacement.ErrRBFTypeMismatch
	ErrNilTransaction          = replacement.ErrNilTransaction
	ErrNonceMismatch           = replacement.ErrNonceMismatch
	ErrInsufficientBump        = replacement.ErrInsufficientBump
	ErrInsufficientTipBump     = replacement.ErrInsufficientTipBump
	ErrPoolCapacity            = replacement.ErrPoolCapacity
	ErrAccountFull             = replacement.ErrAccountFull
)

// Replacement function wrappers.
func DefaultRBFPolicyConfig() RBFPolicyConfig { return replacement.DefaultRBFPolicyConfig() }
func NewRBFPolicyEngine(config RBFPolicyConfig) *RBFPolicyEngine {
	return replacement.NewRBFPolicyEngine(config)
}
func DefaultReplacementPolicy() *ReplacementPolicy { return replacement.DefaultReplacementPolicy() }
func NewReplacementPolicy(priceBump, poolSize, accountSlots int) *ReplacementPolicy {
	return replacement.NewReplacementPolicy(priceBump, poolSize, accountSlots)
}
func ComputePriceBump(existing, newTx *types.Transaction) int {
	return replacement.ComputePriceBump(existing, newTx)
}
func CompareEffectiveGasPrice(a, b *types.Transaction, baseFee *big.Int) int {
	return replacement.CompareEffectiveGasPrice(a, b, baseFee)
}
func EffectiveGasPriceCapped(tx *types.Transaction, baseFee *big.Int) *big.Int {
	return replacement.EffectiveGasPriceCapped(tx, baseFee)
}
func EffectiveTip(tx *types.Transaction, baseFee *big.Int) *big.Int {
	return replacement.EffectiveTip(tx, baseFee)
}
func SortByPrice(txs []*types.Transaction, baseFee *big.Int) []*types.Transaction {
	return replacement.SortByPrice(txs, baseFee)
}
func GetPromotable(pending map[[20]byte]*AccountPending, baseFee *big.Int) []*types.Transaction {
	return replacement.GetPromotable(pending, baseFee)
}
func FilterByMinTip(txs []*types.Transaction, baseFee, minTip *big.Int) []*types.Transaction {
	return replacement.FilterByMinTip(txs, baseFee, minTip)
}
func GroupByNonce(txs []*types.Transaction) map[uint64][]*types.Transaction {
	return replacement.GroupByNonce(txs)
}
func BestByNonce(txs []*types.Transaction, baseFee *big.Int) []*types.Transaction {
	return replacement.BestByNonce(txs, baseFee)
}
