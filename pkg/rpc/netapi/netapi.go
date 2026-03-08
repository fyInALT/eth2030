// Package netapi provides the net namespace JSON-RPC backend interface and
// core logic for network status information.
package netapi

import (
	"errors"
	"fmt"

	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// Backend provides access to network status information.
type Backend interface {
	// NetworkID returns the network identifier (e.g. 1 for mainnet).
	NetworkID() uint64

	// IsListening returns whether the node is currently accepting
	// inbound connections.
	IsListening() bool

	// PeerCount returns the number of currently connected peers.
	PeerCount() int

	// MaxPeers returns the configured maximum number of peers.
	MaxPeers() int
}

// API implements the net_ namespace methods.
type API struct {
	backend Backend
}

// NewAPI creates a new net API service.
func NewAPI(backend Backend) *API {
	return &API{backend: backend}
}

// ErrBackendNil is returned when the net backend is nil.
var ErrBackendNil = errors.New("net backend not available")

// Version returns the network ID as a decimal string.
func (n *API) Version() (string, error) {
	if n.backend == nil {
		return "", ErrBackendNil
	}
	return fmt.Sprintf("%d", n.backend.NetworkID()), nil
}

// Listening returns whether the node is accepting connections.
func (n *API) Listening() (bool, error) {
	if n.backend == nil {
		return false, ErrBackendNil
	}
	return n.backend.IsListening(), nil
}

// PeerCount returns the connected peer count.
func (n *API) PeerCount() (int, error) {
	if n.backend == nil {
		return 0, ErrBackendNil
	}
	return n.backend.PeerCount(), nil
}

// MaxPeers returns the max peers count.
func (n *API) MaxPeers() (int, error) {
	if n.backend == nil {
		return 0, ErrBackendNil
	}
	return n.backend.MaxPeers(), nil
}

// NetworkID returns the raw network identifier.
func (n *API) NetworkID() (uint64, error) {
	if n.backend == nil {
		return 0, ErrBackendNil
	}
	return n.backend.NetworkID(), nil
}

// GetBackend returns the underlying Backend for testing/inspection.
func (n *API) GetBackend() Backend { return n.backend }

// HandleNetRequest dispatches a net_ namespace JSON-RPC request.
func (n *API) HandleNetRequest(req *rpctypes.Request) *rpctypes.Response {
	switch req.Method {
	case "net_version":
		v, err := n.Version()
		if err != nil {
			return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, err.Error())
		}
		return rpctypes.NewSuccessResponse(req.ID, v)
	case "net_listening":
		l, err := n.Listening()
		if err != nil {
			return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, err.Error())
		}
		return rpctypes.NewSuccessResponse(req.ID, l)
	case "net_peerCount":
		count, err := n.PeerCount()
		if err != nil {
			return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, err.Error())
		}
		return rpctypes.NewSuccessResponse(req.ID, rpctypes.EncodeUint64(uint64(count)))
	case "net_maxPeers":
		max, err := n.MaxPeers()
		if err != nil {
			return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, err.Error())
		}
		return rpctypes.NewSuccessResponse(req.ID, rpctypes.EncodeUint64(uint64(max)))
	default:
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeMethodNotFound,
			fmt.Sprintf("method %q not found in net namespace", req.Method))
	}
}
