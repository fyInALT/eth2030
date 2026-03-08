// server_extended.go re-exports ExtServer from rpc/server for backward
// compatibility, with a convenience wrapper for AdminBackend wiring.
package rpc

import (
	rpcserver "github.com/eth2030/eth2030/rpc/server"
)

// ExtServer is a full-featured JSON-RPC server with middleware, CORS,
// auth, rate limiting, batch handling, and graceful shutdown.
// It wraps rpcserver.ExtServer and adds AdminBackend convenience wiring.
type ExtServer struct {
	*rpcserver.ExtServer
}

// NewExtServer creates a new extended JSON-RPC server for the given backend.
func NewExtServer(backend Backend, config ServerConfig) *ExtServer {
	api := NewEthAPI(backend)
	return &ExtServer{rpcserver.NewExtServer(api, config)}
}

// SetAdminBackend wires an AdminBackend so admin_* methods are served.
func (s *ExtServer) SetAdminBackend(b AdminBackend) {
	s.SetAdminHandler(NewAdminDispatchAPI(b))
}

// Re-export middleware constructors.
var (
	ExtCORSMiddleware      = rpcserver.CORSMiddleware
	ExtAuthMiddleware      = rpcserver.AuthMiddleware
	ExtRateLimitMiddleware = rpcserver.RateLimitMiddleware
)
