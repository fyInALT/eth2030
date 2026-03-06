package engine

// util_compat.go re-exports types from engine/util for backward compatibility.

import "github.com/eth2030/eth2030/engine/util"

// Util type aliases.
type (
	CoordinatorConfig  = util.CoordinatorConfig
	BuilderReputation  = util.BuilderReputation
	AuctionSettlement  = util.AuctionSettlement
	BuilderCoordinator = util.BuilderCoordinator
	MEVBurnCalculator  = util.MEVBurnCalculator
	ShrinkConfig       = util.ShrinkConfig
	ShrinkStats        = util.ShrinkStats
	ShrinkEstimate     = util.ShrinkEstimate
	PayloadShrinker    = util.PayloadShrinker
)

// Util constants.
const (
	StrategyCompress   = util.StrategyCompress
	StrategyPruneZeros = util.StrategyPruneZeros
	StrategyDedup      = util.StrategyDedup
	StrategyCombined   = util.StrategyCombined
)

// Util error variables.
var (
	ErrCoordNoBids           = util.ErrCoordNoBids
	ErrCoordBuilderNotFound  = util.ErrCoordBuilderNotFound
	ErrCoordBuilderExists    = util.ErrCoordBuilderExists
	ErrCoordMaxBuilders      = util.ErrCoordMaxBuilders
	ErrCoordBidDeadline      = util.ErrCoordBidDeadline
	ErrCoordBidIncrement     = util.ErrCoordBidIncrement
	ErrCoordZeroStake        = util.ErrCoordZeroStake
	ErrCoordDuplicateBid     = util.ErrCoordDuplicateBid
	ErrCoordSlotZero         = util.ErrCoordSlotZero
	ErrShrinkEmptyPayload    = util.ErrShrinkEmptyPayload
	ErrShrinkMaxSizeExceeded = util.ErrShrinkMaxSizeExceeded
	ErrShrinkUnknownStrategy = util.ErrShrinkUnknownStrategy
	ErrShrinkCompressFailed  = util.ErrShrinkCompressFailed
)

// Util function wrappers.
func DefaultCoordinatorConfig() CoordinatorConfig { return util.DefaultCoordinatorConfig() }
func NewBuilderCoordinator(cfg CoordinatorConfig) *BuilderCoordinator {
	return util.NewBuilderCoordinator(cfg)
}
func NewMEVBurnCalculator(burnRate float64) *MEVBurnCalculator {
	return util.NewMEVBurnCalculator(burnRate)
}
func DefaultShrinkConfig() ShrinkConfig { return util.DefaultShrinkConfig() }
func NewPayloadShrinker(config ShrinkConfig) *PayloadShrinker {
	return util.NewPayloadShrinker(config)
}
