package consensus

// config_compat.go re-exports types from consensus/clconfig for backward compatibility.

import "github.com/eth2030/eth2030/consensus/clconfig"

// ConsensusConfig type alias.
type ConsensusConfig = clconfig.ConsensusConfig

// ConsensusConfig function wrappers.
func DefaultConfig() *ConsensusConfig    { return clconfig.DefaultConfig() }
func QuickSlotsConfig() *ConsensusConfig { return clconfig.QuickSlotsConfig() }
