package txpool

// stark_compat.go re-exports types from txpool/stark for backward compatibility.

import (
	"time"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/txpool/stark"
)

// Stark type aliases.
type (
	ValidatedTx            = stark.ValidatedTx
	MempoolAggregationTick = stark.MempoolAggregationTick
	P2PBroadcaster         = stark.P2PBroadcaster
	STARKAggregator        = stark.STARKAggregator
)

// Stark constants.
const (
	DefaultTickInterval = stark.DefaultTickInterval
	MaxTickTransactions = stark.MaxTickTransactions
	MaxTickSize         = stark.MaxTickSize
)

// Stark error variables.
var (
	ErrAggNotRunning     = stark.ErrAggNotRunning
	ErrAggAlreadyRunning = stark.ErrAggAlreadyRunning
	ErrAggNoTransactions = stark.ErrAggNoTransactions
	ErrAggTickFailed     = stark.ErrAggTickFailed
	ErrAggInvalidTick    = stark.ErrAggInvalidTick
	ErrAggMergeFailed    = stark.ErrAggMergeFailed
	ErrAggTickTooLarge   = stark.ErrAggTickTooLarge
)

// Stark function wrappers.
func NewSTARKAggregator(peerID string) *STARKAggregator {
	return stark.NewSTARKAggregator(peerID)
}
func NewSTARKAggregatorWithInterval(peerID string, interval time.Duration) *STARKAggregator {
	return stark.NewSTARKAggregatorWithInterval(peerID, interval)
}
func TickHash(tick *MempoolAggregationTick) types.Hash {
	return stark.TickHash(tick)
}
