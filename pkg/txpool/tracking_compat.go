package txpool

// tracking_compat.go re-exports types from txpool/tracking for backward compatibility.

import "github.com/eth2030/eth2030/txpool/tracking"

// Tracking type aliases.
type (
	NonceGap           = tracking.NonceGap
	NonceTrackerConfig = tracking.NonceTrackerConfig
	NonceTracker       = tracking.NonceTracker
	AcctInfo           = tracking.AcctInfo
	AcctTrack          = tracking.AcctTrack
)

// Tracking error variables.
var (
	ErrAcctNotTracked      = tracking.ErrAcctNotTracked
	ErrAcctInsufficientBal = tracking.ErrAcctInsufficientBal
	ErrAcctNonceGap        = tracking.ErrAcctNonceGap
)

// Tracking function wrappers.
func DefaultNonceTrackerConfig() NonceTrackerConfig {
	return tracking.DefaultNonceTrackerConfig()
}

// NewNonceTracker creates a NonceTracker from a StateReader.
// StateReader satisfies tracking.NonceStateReader (has GetNonce).
func NewNonceTracker(config NonceTrackerConfig, state StateReader) *NonceTracker {
	return tracking.NewNonceTracker(config, state)
}

// NewAcctTrack creates an AcctTrack from a StateReader.
// StateReader satisfies tracking.AccountStateReader (has GetNonce + GetBalance).
func NewAcctTrack(state StateReader) *AcctTrack {
	return tracking.NewAcctTrack(state)
}
