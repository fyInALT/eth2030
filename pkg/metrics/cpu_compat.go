package metrics

// cpu_compat.go re-exports types from metrics/cpu for backward compatibility.

import "github.com/eth2030/eth2030/metrics/cpu"

// CPU type aliases.
type (
	CPUStats   = cpu.CPUStats
	CPUTracker = cpu.CPUTracker
)

// CPU function wrappers.
func ReadCPUStats() *CPUStats    { return cpu.ReadCPUStats() }
func NewCPUTracker() *CPUTracker { return cpu.NewCPUTracker() }
