// Package adminapi provides the admin namespace JSON-RPC backend interface,
// data types, and core logic for node administration.
package adminapi

import (
	"errors"
	"fmt"
)

// Backend provides access to node administration data.
type Backend interface {
	NodeInfo() NodeInfoData
	Peers() []PeerInfoData
	AddPeer(url string) error
	RemovePeer(url string) error
	ChainID() uint64
	DataDir() string
}

// NodeInfoData contains information about the running node.
// Fields match the geth admin_nodeInfo response so Kurtosis extractors work.
type NodeInfoData struct {
	Name       string                 `json:"name"`
	ID         string                 `json:"id"`
	ENR        string                 `json:"enr"`
	Enode      string                 `json:"enode"`
	IP         string                 `json:"ip"`
	ListenAddr string                 `json:"listenAddr"`
	Ports      NodePorts              `json:"ports"`
	Protocols  map[string]interface{} `json:"protocols"`
}

// NodePorts holds the discovery and listener port numbers.
type NodePorts struct {
	Discovery int `json:"discovery"`
	Listener  int `json:"listener"`
}

// PeerInfoData contains information about a connected peer.
type PeerInfoData struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	RemoteAddr string   `json:"remoteAddr"`
	Caps       []string `json:"caps"`
	Static     bool     `json:"static"`
	Trusted    bool     `json:"trusted"`
}

// API implements the admin namespace JSON-RPC methods.
type API struct {
	backend Backend
}

// NewAPI creates a new admin API service.
func NewAPI(backend Backend) *API {
	return &API{backend: backend}
}

// NodeInfo returns information about the running node.
func (api *API) NodeInfo() (*NodeInfoData, error) {
	if api.backend == nil {
		return nil, errors.New("admin backend not available")
	}
	info := api.backend.NodeInfo()
	return &info, nil
}

// Peers returns information about connected peers.
func (api *API) Peers() ([]PeerInfoData, error) {
	if api.backend == nil {
		return nil, errors.New("admin backend not available")
	}
	peers := api.backend.Peers()
	if peers == nil {
		return []PeerInfoData{}, nil
	}
	return peers, nil
}

// AddPeer requests adding a new remote peer.
func (api *API) AddPeer(url string) (bool, error) {
	if api.backend == nil {
		return false, errors.New("admin backend not available")
	}
	if url == "" {
		return false, errors.New("empty peer URL")
	}
	if err := api.backend.AddPeer(url); err != nil {
		return false, err
	}
	return true, nil
}

// RemovePeer requests disconnection from a remote peer.
func (api *API) RemovePeer(url string) (bool, error) {
	if api.backend == nil {
		return false, errors.New("admin backend not available")
	}
	if url == "" {
		return false, errors.New("empty peer URL")
	}
	if err := api.backend.RemovePeer(url); err != nil {
		return false, err
	}
	return true, nil
}

// DataDir returns the data directory of the node.
func (api *API) DataDir() (string, error) {
	if api.backend == nil {
		return "", errors.New("admin backend not available")
	}
	return api.backend.DataDir(), nil
}

// StartRPC starts the HTTP RPC listener (stub).
func (api *API) StartRPC(host string, port int) (bool, error) {
	if api.backend == nil {
		return false, errors.New("admin backend not available")
	}
	if host == "" {
		return false, errors.New("empty host")
	}
	if port <= 0 || port > 65535 {
		return false, errors.New("invalid port number")
	}
	// Stub: in a full implementation this would start the HTTP listener.
	return true, nil
}

// StopRPC stops the HTTP RPC listener (stub).
func (api *API) StopRPC() (bool, error) {
	if api.backend == nil {
		return false, errors.New("admin backend not available")
	}
	// Stub: in a full implementation this would stop the HTTP listener.
	return true, nil
}

// ChainID returns the chain ID as a hex string.
func (api *API) ChainID() (string, error) {
	if api.backend == nil {
		return "", errors.New("admin backend not available")
	}
	id := api.backend.ChainID()
	return fmt.Sprintf("0x%x", id), nil
}

// AdminNodeInfo is an alias for NodeInfo for backward compatibility.
func (api *API) AdminNodeInfo() (*NodeInfoData, error) { return api.NodeInfo() }

// AdminPeers is an alias for Peers for backward compatibility.
func (api *API) AdminPeers() ([]PeerInfoData, error) { return api.Peers() }

// AdminAddPeer is an alias for AddPeer for backward compatibility.
func (api *API) AdminAddPeer(url string) (bool, error) { return api.AddPeer(url) }

// AdminRemovePeer is an alias for RemovePeer for backward compatibility.
func (api *API) AdminRemovePeer(url string) (bool, error) { return api.RemovePeer(url) }

// AdminDataDir is an alias for DataDir for backward compatibility.
func (api *API) AdminDataDir() (string, error) { return api.DataDir() }

// AdminStartRPC is an alias for StartRPC for backward compatibility.
func (api *API) AdminStartRPC(host string, port int) (bool, error) { return api.StartRPC(host, port) }

// AdminStopRPC is an alias for StopRPC for backward compatibility.
func (api *API) AdminStopRPC() (bool, error) { return api.StopRPC() }

// AdminChainID is an alias for ChainID for backward compatibility.
func (api *API) AdminChainID() (string, error) { return api.ChainID() }

// GetBackend returns the underlying Backend for testing/inspection.
func (api *API) GetBackend() Backend { return api.backend }
