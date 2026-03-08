// server.go re-exports Server from rpc/server for backward compatibility,
// with a convenience wrapper for AdminBackend wiring.
package rpc

import (
	rpcserver "github.com/eth2030/eth2030/rpc/server"
)

// Server is a JSON-RPC HTTP server that dispatches requests to the EthAPI.
// It wraps rpcserver.Server and adds AdminBackend convenience wiring.
type Server struct {
	*rpcserver.Server
}

// NewServer creates a new JSON-RPC server for the given backend.
func NewServer(backend Backend) *Server {
	api := NewEthAPI(backend)
	return &Server{rpcserver.NewServer(api)}
}

// SetAdminBackend wires an AdminBackend so admin_* methods are served.
func (s *Server) SetAdminBackend(b AdminBackend) {
	s.SetAdminHandler(NewAdminDispatchAPI(b))
}
