package clconfig

import "fmt"

// ConsensusConfig holds consensus-layer parameters.
type ConsensusConfig struct {
	SecondsPerSlot    uint64 // slot duration in seconds
	SlotsPerEpoch     uint64 // number of slots per epoch
	MinGenesisTime    uint64 // minimum genesis timestamp
	EpochsForFinality uint64 // epochs required for finalization (2 = Casper FFG, 1 = single-epoch)

	LeanAvailableChainMode       bool // enables lean available chain mode (PQ subset attestation)
	LeanAvailableChainValidators int  // number of PQ attestors per slot; range [256,1024], default 512
}

// DefaultConfig returns the standard Ethereum mainnet consensus config.
func DefaultConfig() *ConsensusConfig {
	return &ConsensusConfig{
		SecondsPerSlot:               12,
		SlotsPerEpoch:                32,
		MinGenesisTime:               0,
		EpochsForFinality:            2,
		LeanAvailableChainValidators: 512,
	}
}

// QuickSlotsConfig returns the config for the quick-slots + 1-epoch finality upgrade.
func QuickSlotsConfig() *ConsensusConfig {
	return &ConsensusConfig{
		SecondsPerSlot:    6,
		SlotsPerEpoch:     4,
		MinGenesisTime:    0,
		EpochsForFinality: 1,
	}
}

// Validate checks config constraints and returns an error if invalid.
func (c *ConsensusConfig) Validate() error {
	if c.SecondsPerSlot == 0 {
		return fmt.Errorf("consensus: SecondsPerSlot must be > 0")
	}
	if c.SlotsPerEpoch == 0 {
		return fmt.Errorf("consensus: SlotsPerEpoch must be > 0")
	}
	if c.EpochsForFinality == 0 {
		return fmt.Errorf("consensus: EpochsForFinality must be > 0")
	}
	if c.LeanAvailableChainMode && c.LeanAvailableChainValidators != 0 {
		if c.LeanAvailableChainValidators < 256 || c.LeanAvailableChainValidators > 1024 {
			return fmt.Errorf("consensus: LeanAvailableChainValidators must be in [256, 1024], got %d", c.LeanAvailableChainValidators)
		}
	}
	return nil
}

// EpochDuration returns the total duration of one epoch in seconds.
func (c *ConsensusConfig) EpochDuration() uint64 {
	return c.SecondsPerSlot * c.SlotsPerEpoch
}

// IsSingleEpochFinality returns true if the config uses 1-epoch finality.
func (c *ConsensusConfig) IsSingleEpochFinality() bool {
	return c.EpochsForFinality == 1
}
