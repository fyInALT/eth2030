package rollup

// anchor_chain_tracker_compat.go re-exports types from rollup/anchortx for backward compatibility.

import "github.com/eth2030/eth2030/rollup/anchortx"

// Type aliases.
type (
	AnchorChainConfig  = anchortx.AnchorChainConfig
	AnchorPoint        = anchortx.AnchorPoint
	AnchorMetrics      = anchortx.AnchorMetrics
	AnchorChainTracker = anchortx.AnchorChainTracker
)

// Errors.
var (
	ErrChainAlreadyRegistered = anchortx.ErrChainAlreadyRegistered
	ErrChainNotRegistered     = anchortx.ErrChainNotRegistered
	ErrChainMaxReached        = anchortx.ErrChainMaxReached
	ErrChainIDZero            = anchortx.ErrChainIDZero
	ErrAnchorBlockRegression  = anchortx.ErrAnchorBlockRegression
	ErrAnchorAlreadyConfirmed = anchortx.ErrAnchorAlreadyConfirmed
	ErrAnchorBlockNotFound    = anchortx.ErrAnchorBlockNotFound
)

// NewAnchorChainTracker creates a new tracker with the given maximum number of chains.
func NewAnchorChainTracker(maxChains int) *AnchorChainTracker {
	return anchortx.NewAnchorChainTracker(maxChains)
}
