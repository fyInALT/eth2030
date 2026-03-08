package dispatch

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/eth2030/eth2030/p2p/peermgr"
)

// Protocol manager errors.
var (
	ErrPeerAlreadyConnected = errors.New("p2p: peer already connected")
	ErrPeerNotConnected     = errors.New("p2p: peer not connected")
	ErrTooManyPeers         = errors.New("p2p: too many peers")
	ErrProtocolExists       = errors.New("p2p: protocol already registered")
	ErrNoSharedCaps         = errors.New("p2p: no shared capabilities")
	// ErrProtocolNotFound is returned when no handler is registered for a protocol.
	ErrProtocolNotFound = errors.New("p2p: protocol not found")
)

// Capability represents a sub-protocol name and version pair.
type Capability struct {
	Name    string
	Version uint
}

// String returns "name/version" representation.
func (c Capability) String() string {
	return fmt.Sprintf("%s/%d", c.Name, c.Version)
}

// PeerMgrInfo holds information about a connected peer.
type PeerMgrInfo struct {
	NodeID         string
	Capabilities   []Capability
	Latency        time.Duration
	ConnectedSince time.Time
	LastSeen       time.Time
	Inbound        bool
}

// ProtocolHandler handles messages for a registered protocol.
type ProtocolHandler func(peerID string, code uint64, payload []byte) error

// registeredProtocol stores a protocol registration.
type registeredProtocol struct {
	Name    string
	Version uint
	Handler ProtocolHandler
}

// ConnectFunc is called by ProtocolManager.Connect to establish a connection.
type ConnectFunc func(nodeID string) ([]Capability, error)

// ProtocolManagerConfig configures the protocol manager.
type ProtocolManagerConfig struct {
	MaxInbound  int // max inbound peers (0 = unlimited)
	MaxOutbound int // max outbound peers (0 = unlimited)
	MaxTotal    int // max total peers (0 = unlimited)
}

func (c *ProtocolManagerConfig) defaults() {
	if c.MaxTotal <= 0 {
		c.MaxTotal = 50
	}
	if c.MaxInbound <= 0 {
		c.MaxInbound = c.MaxTotal
	}
	if c.MaxOutbound <= 0 {
		c.MaxOutbound = c.MaxTotal
	}
}

// ProtocolManager manages P2P protocol handshake, capability negotiation,
// and message routing for connected peers.
type ProtocolManager struct {
	mu        sync.RWMutex
	config    ProtocolManagerConfig
	peers     map[string]*PeerMgrInfo
	protocols []registeredProtocol
	inbound   int
	outbound  int

	onConnect    []func(info *PeerMgrInfo)
	onDisconnect []func(nodeID string, reason string)

	// ConnectFn is the function used to establish connections.
	ConnectFn ConnectFunc
}

// NewProtocolManager creates a new ProtocolManager with the given config.
func NewProtocolManager(cfg ProtocolManagerConfig) *ProtocolManager {
	cfg.defaults()
	return &ProtocolManager{
		config: cfg,
		peers:  make(map[string]*PeerMgrInfo),
	}
}

// RegisterProtocol registers a named protocol at a given version with a handler.
func (pm *ProtocolManager) RegisterProtocol(name string, version uint, handler ProtocolHandler) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, p := range pm.protocols {
		if p.Name == name && p.Version == version {
			return ErrProtocolExists
		}
	}

	pm.protocols = append(pm.protocols, registeredProtocol{
		Name:    name,
		Version: version,
		Handler: handler,
	})
	return nil
}

// Protocols returns the list of registered protocol capabilities.
func (pm *ProtocolManager) Protocols() []Capability {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	caps := make([]Capability, len(pm.protocols))
	for i, p := range pm.protocols {
		caps[i] = Capability{Name: p.Name, Version: p.Version}
	}
	return caps
}

// Connect establishes a connection to a remote node.
func (pm *ProtocolManager) Connect(nodeID string) error {
	pm.mu.Lock()

	if len(pm.peers) >= pm.config.MaxTotal {
		pm.mu.Unlock()
		return ErrTooManyPeers
	}
	if pm.outbound >= pm.config.MaxOutbound {
		pm.mu.Unlock()
		return peermgr.ErrTooManyOutbound
	}
	if _, exists := pm.peers[nodeID]; exists {
		pm.mu.Unlock()
		return ErrPeerAlreadyConnected
	}
	pm.mu.Unlock()

	var remoteCaps []Capability
	if pm.ConnectFn != nil {
		var err error
		remoteCaps, err = pm.ConnectFn(nodeID)
		if err != nil {
			return fmt.Errorf("p2p: connect to %s: %w", nodeID, err)
		}
	}

	localCaps := pm.Protocols()
	shared := MatchCapabilities(localCaps, remoteCaps)
	if len(shared) == 0 && len(localCaps) > 0 && len(remoteCaps) > 0 {
		return ErrNoSharedCaps
	}

	now := time.Now()
	info := &PeerMgrInfo{
		NodeID:         nodeID,
		Capabilities:   shared,
		ConnectedSince: now,
		LastSeen:       now,
		Inbound:        false,
	}

	pm.mu.Lock()
	if _, exists := pm.peers[nodeID]; exists {
		pm.mu.Unlock()
		return ErrPeerAlreadyConnected
	}
	if len(pm.peers) >= pm.config.MaxTotal {
		pm.mu.Unlock()
		return ErrTooManyPeers
	}
	pm.peers[nodeID] = info
	pm.outbound++
	callbacks := make([]func(info *PeerMgrInfo), len(pm.onConnect))
	copy(callbacks, pm.onConnect)
	pm.mu.Unlock()

	for _, cb := range callbacks {
		cb(info)
	}
	return nil
}

// AcceptPeer registers an inbound peer with the given capabilities.
func (pm *ProtocolManager) AcceptPeer(nodeID string, remoteCaps []Capability) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.peers) >= pm.config.MaxTotal {
		return ErrTooManyPeers
	}
	if pm.inbound >= pm.config.MaxInbound {
		return peermgr.ErrTooManyInbound
	}
	if _, exists := pm.peers[nodeID]; exists {
		return ErrPeerAlreadyConnected
	}

	localCaps := make([]Capability, len(pm.protocols))
	for i, p := range pm.protocols {
		localCaps[i] = Capability{Name: p.Name, Version: p.Version}
	}

	shared := MatchCapabilities(localCaps, remoteCaps)

	now := time.Now()
	info := &PeerMgrInfo{
		NodeID:         nodeID,
		Capabilities:   shared,
		ConnectedSince: now,
		LastSeen:       now,
		Inbound:        true,
	}

	pm.peers[nodeID] = info
	pm.inbound++

	callbacks := make([]func(info *PeerMgrInfo), len(pm.onConnect))
	copy(callbacks, pm.onConnect)

	go func() {
		for _, cb := range callbacks {
			cb(info)
		}
	}()

	return nil
}

// Disconnect removes a peer with the given reason and notifies subscribers.
func (pm *ProtocolManager) Disconnect(nodeID string, reason string) error {
	pm.mu.Lock()

	info, exists := pm.peers[nodeID]
	if !exists {
		pm.mu.Unlock()
		return ErrPeerNotConnected
	}

	delete(pm.peers, nodeID)
	if info.Inbound {
		pm.inbound--
	} else {
		pm.outbound--
	}

	callbacks := make([]func(nodeID string, reason string), len(pm.onDisconnect))
	copy(callbacks, pm.onDisconnect)
	pm.mu.Unlock()

	for _, cb := range callbacks {
		cb(nodeID, reason)
	}
	return nil
}

// OnConnect registers a callback that fires when a new peer connects.
func (pm *ProtocolManager) OnConnect(fn func(info *PeerMgrInfo)) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.onConnect = append(pm.onConnect, fn)
}

// OnDisconnect registers a callback that fires when a peer disconnects.
func (pm *ProtocolManager) OnDisconnect(fn func(nodeID string, reason string)) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.onDisconnect = append(pm.onDisconnect, fn)
}

// PeerInfo returns information about a connected peer, or nil if not found.
func (pm *ProtocolManager) PeerInfo(nodeID string) *PeerMgrInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	info, exists := pm.peers[nodeID]
	if !exists {
		return nil
	}
	cp := *info
	cp.Capabilities = make([]Capability, len(info.Capabilities))
	copy(cp.Capabilities, info.Capabilities)
	return &cp
}

// PeerCount returns the total number of connected peers.
func (pm *ProtocolManager) PeerCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.peers)
}

// InboundCount returns the number of inbound peers.
func (pm *ProtocolManager) InboundCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.inbound
}

// OutboundCount returns the number of outbound peers.
func (pm *ProtocolManager) OutboundCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.outbound
}

// AllPeers returns a snapshot of all connected peer info.
func (pm *ProtocolManager) AllPeers() []*PeerMgrInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]*PeerMgrInfo, 0, len(pm.peers))
	for _, info := range pm.peers {
		cp := *info
		cp.Capabilities = make([]Capability, len(info.Capabilities))
		copy(cp.Capabilities, info.Capabilities)
		result = append(result, &cp)
	}
	return result
}

// RouteMessage routes an incoming message to the appropriate protocol handler.
func (pm *ProtocolManager) RouteMessage(peerID string, protoName string, code uint64, payload []byte) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if _, exists := pm.peers[peerID]; !exists {
		return ErrPeerNotConnected
	}

	var handler ProtocolHandler
	var bestVersion uint
	for _, p := range pm.protocols {
		if p.Name == protoName && p.Version > bestVersion {
			handler = p.Handler
			bestVersion = p.Version
		}
	}
	if handler == nil {
		return ErrProtocolNotFound
	}

	return handler(peerID, code, payload)
}

// UpdateLatency updates the measured latency for a peer.
func (pm *ProtocolManager) UpdateLatency(nodeID string, latency time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if info, ok := pm.peers[nodeID]; ok {
		info.Latency = latency
		info.LastSeen = time.Now()
	}
}

// MatchCapabilities finds the highest shared version for each protocol name
// present in both local and remote capability lists.
func MatchCapabilities(local, remote []Capability) []Capability {
	localMap := make(map[string]uint)
	for _, c := range local {
		if v, ok := localMap[c.Name]; !ok || c.Version > v {
			localMap[c.Name] = c.Version
		}
	}

	remoteMap := make(map[string]uint)
	for _, c := range remote {
		if v, ok := remoteMap[c.Name]; !ok || c.Version > v {
			remoteMap[c.Name] = c.Version
		}
	}

	var matched []Capability
	for name, localVer := range localMap {
		if remoteVer, ok := remoteMap[name]; ok {
			ver := localVer
			if remoteVer < ver {
				ver = remoteVer
			}
			matched = append(matched, Capability{Name: name, Version: ver})
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Name < matched[j].Name
	})

	return matched
}

// HasCapability checks if a peer supports a specific protocol name.
func (pm *ProtocolManager) HasCapability(nodeID string, protoName string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	info, exists := pm.peers[nodeID]
	if !exists {
		return false
	}
	for _, c := range info.Capabilities {
		if c.Name == protoName {
			return true
		}
	}
	return false
}
