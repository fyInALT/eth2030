package node

// health_checker_compat.go re-exports types from node/healthcheck for backward compatibility.

import "github.com/eth2030/eth2030/node/healthcheck"

// Type aliases.
type (
	SubsystemChecker = healthcheck.SubsystemChecker
	SubsystemHealth  = healthcheck.SubsystemHealth
	HealthReport     = healthcheck.HealthReport
	HealthChecker    = healthcheck.HealthChecker
)

// Status constants.
const (
	StatusHealthy   = healthcheck.StatusHealthy
	StatusDegraded  = healthcheck.StatusDegraded
	StatusUnhealthy = healthcheck.StatusUnhealthy
)

// NewHealthChecker creates a new HealthChecker with no registered subsystems.
func NewHealthChecker() *HealthChecker { return healthcheck.NewHealthChecker() }
