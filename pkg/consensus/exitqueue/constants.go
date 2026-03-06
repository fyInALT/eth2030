package exitqueue

// Constants copied from consensus package to avoid circular import.
const (
	MaxSeedLookahead                 = 4   // from consensus/beacon_state_v2.go
	MinValidatorWithdrawabilityDelay = 256 // from consensus/validator_lifecycle.go
)
