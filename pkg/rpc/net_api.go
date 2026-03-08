// net_api.go provides a standalone Net namespace API with its own backend
// interface for network information. The dispatch layer lives here (top-level
// rpc package) so it can access successResponse/errorResponse helpers.
// Core logic is in rpc/netapi.
package rpc

import (
	"fmt"

	"github.com/eth2030/eth2030/rpc/netapi"
)

// NetBackend re-exports the net Backend interface.
type NetBackend = netapi.Backend

// NetAPI implements the net_ namespace JSON-RPC methods.
type NetAPI struct {
	inner *netapi.API
}

// NewNetAPI creates a new net API service.
func NewNetAPI(backend NetBackend) *NetAPI {
	return &NetAPI{inner: netapi.NewAPI(backend)}
}

// HandleNetRequest dispatches a net_ namespace JSON-RPC request.
func (n *NetAPI) HandleNetRequest(req *Request) *Response {
	switch req.Method {
	case "net_version":
		return n.netVersionFull(req)
	case "net_listening":
		return n.netListeningFull(req)
	case "net_peerCount":
		return n.netPeerCountFull(req)
	case "net_maxPeers":
		return n.netMaxPeers(req)
	default:
		return errorResponse(req.ID, ErrCodeMethodNotFound,
			fmt.Sprintf("method %q not found in net namespace", req.Method))
	}
}

// netVersionFull returns the network ID as a decimal string.
func (n *NetAPI) netVersionFull(req *Request) *Response {
	v, err := n.inner.Version()
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}
	return successResponse(req.ID, v)
}

// netListeningFull returns whether the node is listening for connections.
func (n *NetAPI) netListeningFull(req *Request) *Response {
	l, err := n.inner.Listening()
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}
	return successResponse(req.ID, l)
}

// netPeerCountFull returns the connected peer count as a hex string.
func (n *NetAPI) netPeerCountFull(req *Request) *Response {
	count, err := n.inner.PeerCount()
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}
	return successResponse(req.ID, encodeUint64(uint64(count)))
}

// netMaxPeers returns the max peer count as a hex string.
func (n *NetAPI) netMaxPeers(req *Request) *Response {
	max, err := n.inner.MaxPeers()
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}
	return successResponse(req.ID, encodeUint64(uint64(max)))
}

// --- Direct Go-typed API methods (for programmatic / internal use) ---

// ErrNetBackendNil is returned when the net backend is nil.
var ErrNetBackendNil = netapi.ErrBackendNil

// Version returns the network ID as a decimal string.
func (n *NetAPI) Version() (string, error) {
	return n.inner.Version()
}

// Listening returns whether the node is accepting connections.
func (n *NetAPI) Listening() (bool, error) {
	return n.inner.Listening()
}

// PeerCount returns the connected peer count.
func (n *NetAPI) PeerCount() (int, error) {
	return n.inner.PeerCount()
}

// GetBackend returns the underlying backend for testing/inspection.
func (n *NetAPI) GetBackend() NetBackend {
	return n.inner.GetBackend()
}
