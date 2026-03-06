package pq

// LeanConfig holds the PQ-relevant subset of the consensus configuration.
// Callers from the root consensus package convert *ConsensusConfig to *LeanConfig.
type LeanConfig struct {
	// LeanAvailableChainMode enables lean available-chain PQ attestor selection.
	LeanAvailableChainMode bool

	// LeanAvailableChainValidators is the number of validators to select per slot
	// in lean mode.
	LeanAvailableChainValidators int
}
