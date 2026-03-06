package node

// lifecycle_compat.go re-exports types from node/lifecycle for backward compatibility.

import "github.com/eth2030/eth2030/node/lifecycle"

// Type aliases.
type (
	ServiceState     = lifecycle.ServiceState
	Service          = lifecycle.Service
	LifecycleConfig  = lifecycle.LifecycleConfig
	ServiceEntry     = lifecycle.ServiceEntry
	LifecycleManager = lifecycle.LifecycleManager
)

// ServiceState constants.
const (
	StateCreated  = lifecycle.StateCreated
	StateStarting = lifecycle.StateStarting
	StateRunning  = lifecycle.StateRunning
	StateStopping = lifecycle.StateStopping
	StateStopped  = lifecycle.StateStopped
	StateFailed   = lifecycle.StateFailed
)

// Function wrappers.
func DefaultLifecycleConfig() LifecycleConfig { return lifecycle.DefaultLifecycleConfig() }
func NewLifecycleManager(config LifecycleConfig) *LifecycleManager {
	return lifecycle.NewLifecycleManager(config)
}
