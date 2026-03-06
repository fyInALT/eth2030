package rpc

// middleware_compat.go re-exports types from rpc/middleware for backward compatibility.

import "github.com/eth2030/eth2030/rpc/middleware"

// Middleware type aliases.
type (
	RPCRateLimitConfig    = middleware.RPCRateLimitConfig
	ClientRateStats       = middleware.ClientRateStats
	MethodRateStats       = middleware.MethodRateStats
	GlobalRateStats       = middleware.GlobalRateStats
	RPCRateLimiter        = middleware.RPCRateLimiter
	LifecycleState        = middleware.LifecycleState
	RPCEndpoint           = middleware.RPCEndpoint
	LifecycleEvent        = middleware.LifecycleEvent
	LifecycleManager      = middleware.LifecycleManager
	LifecycleManagerConfig = middleware.LifecycleManagerConfig
)

// Lifecycle constants.
const (
	LCStateIdle     = middleware.LCStateIdle
	LCStateStarting = middleware.LCStateStarting
	LCStateRunning  = middleware.LCStateRunning
	LCStateStopping = middleware.LCStateStopping
	LCStateStopped  = middleware.LCStateStopped
)

// Middleware function wrappers.
func DefaultRPCRateLimitConfig() *RPCRateLimitConfig { return middleware.DefaultRPCRateLimitConfig() }
func NewRPCRateLimiter(config *RPCRateLimitConfig) *RPCRateLimiter {
	return middleware.NewRPCRateLimiter(config)
}
func DefaultLifecycleManagerConfig() LifecycleManagerConfig {
	return middleware.DefaultLifecycleManagerConfig()
}
func NewLifecycleManager(config LifecycleManagerConfig) *LifecycleManager {
	return middleware.NewLifecycleManager(config)
}
