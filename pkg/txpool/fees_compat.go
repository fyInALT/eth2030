package txpool

// fees_compat.go re-exports types from txpool/fees for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/txpool/fees"
)

// Fees type aliases.
type (
	BlobFeeTrackerConfig   = fees.BlobFeeTrackerConfig
	BlobFeeRecord          = fees.BlobFeeRecord
	BlobFeeSuggestion      = fees.BlobFeeSuggestion
	BlobFeeSpike           = fees.BlobFeeSpike
	BlobFeeTracker         = fees.BlobFeeTracker
	BlockFeeData           = fees.BlockFeeData
	FeeEstimatorConfig     = fees.FeeEstimatorConfig
	FeeEstimator           = fees.FeeEstimator
	PriceOracleConfig      = fees.PriceOracleConfig
	BlockFeeRecord         = fees.BlockFeeRecord
	FeeRecommendation      = fees.FeeRecommendation
	FeeHistoryEntry        = fees.FeeHistoryEntry
	PriceOracle            = fees.PriceOracle
	ValidationCacheConfig  = fees.ValidationCacheConfig
	CacheEntry             = fees.CacheEntry
	ValidationCache        = fees.ValidationCache
	BatchSignatureVerifier = fees.BatchSignatureVerifier
)

// Fees constants.
const (
	BlobFeeDefaultWindow             = fees.BlobFeeDefaultWindow
	BlobFeeDefaultFloor              = fees.BlobFeeDefaultFloor
	BlobFeeSpikeThresholdPct         = fees.BlobFeeSpikeThresholdPct
	BlobFeeSlowPercentile            = fees.BlobFeeSlowPercentile
	BlobFeeMedPercentile             = fees.BlobFeeMedPercentile
	BlobFeeFastPercentile            = fees.BlobFeeFastPercentile
	BlobFeeBufferNum                 = fees.BlobFeeBufferNum
	BlobFeeBufferDenom               = fees.BlobFeeBufferDenom
	BlobTargetGasPerBlock            = fees.BlobTargetGasPerBlock
	BlobMaxGasPerBlock               = fees.BlobMaxGasPerBlock
	BlobBaseFeeUpdateFractionTracker = fees.BlobBaseFeeUpdateFractionTracker
	FeeHistorySize                   = fees.FeeHistorySize
	DefaultSuggestedTipMultiplier    = fees.DefaultSuggestedTipMultiplier
	DefaultMinSuggestedGasPrice      = fees.DefaultMinSuggestedGasPrice
	DefaultMinSuggestedTip           = fees.DefaultMinSuggestedTip
	FeeEstPercentileLow              = fees.FeeEstPercentileLow
	FeeEstPercentileMed              = fees.FeeEstPercentileMed
	FeeEstPercentileHigh             = fees.FeeEstPercentileHigh
	PriceOracleDefaultWindow         = fees.PriceOracleDefaultWindow
	PriceOracleSlowPercentile        = fees.PriceOracleSlowPercentile
	PriceOracleMediumPercentile      = fees.PriceOracleMediumPercentile
	PriceOracleFastPercentile        = fees.PriceOracleFastPercentile
	PriceOracleMinBaseFee            = fees.PriceOracleMinBaseFee
	PriceOracleMinTip                = fees.PriceOracleMinTip
	PriceOracleBaseFeeMarginNum      = fees.PriceOracleBaseFeeMarginNum
	PriceOracleBaseFeeMarginDenom    = fees.PriceOracleBaseFeeMarginDenom
)

// Fees function wrappers.
func DefaultBlobFeeTrackerConfig() BlobFeeTrackerConfig {
	return fees.DefaultBlobFeeTrackerConfig()
}
func NewBlobFeeTracker(config BlobFeeTrackerConfig) *BlobFeeTracker {
	return fees.NewBlobFeeTracker(config)
}
func DefaultFeeEstimatorConfig() FeeEstimatorConfig { return fees.DefaultFeeEstimatorConfig() }
func NewFeeEstimator(config FeeEstimatorConfig) *FeeEstimator {
	return fees.NewFeeEstimator(config)
}
func DefaultPriceOracleConfig() PriceOracleConfig          { return fees.DefaultPriceOracleConfig() }
func NewPriceOracle(config PriceOracleConfig) *PriceOracle { return fees.NewPriceOracle(config) }
func DefaultValidationCacheConfig() ValidationCacheConfig {
	return fees.DefaultValidationCacheConfig()
}
func NewValidationCache(cfg ValidationCacheConfig) *ValidationCache {
	return fees.NewValidationCache(cfg)
}
func NewBatchSignatureVerifier(workers int) *BatchSignatureVerifier {
	return fees.NewBatchSignatureVerifier(workers)
}

// MinBlobBaseFee re-exports the minimum blob base fee constant.
func SuggestMinBlobFee() *big.Int { return big.NewInt(fees.MinBlobBaseFee) }
