package p2p

// peermgr_compat.go re-exports types from p2p/peermgr for backward compatibility.
// Consumers should migrate to importing p2p/peermgr directly.

import "github.com/eth2030/eth2030/p2p/peermgr"

// Peer management type aliases.
type (
	Peer              = peermgr.Peer
	PeerSet           = peermgr.PeerSet
	ManagedPeerSet    = peermgr.ManagedPeerSet
	AdvPeerManager    = peermgr.AdvPeerManager
	AdvPeerInfo       = peermgr.AdvPeerInfo
	PeerManagerConfig = peermgr.PeerManagerConfig
	MsgReadWriter     = peermgr.MsgReadWriter
	PeerHandler       = peermgr.PeerHandler
	PeerHandlerFunc   = peermgr.PeerHandlerFunc
	PeerInfo          = peermgr.PeerInfo
	PeerSetReader     = peermgr.PeerSetReader
)

// Peer management errors.
var (
	ErrPeerAlreadyRegistered = peermgr.ErrPeerAlreadyRegistered
	ErrPeerNotRegistered     = peermgr.ErrPeerNotRegistered
	ErrMaxPeers              = peermgr.ErrMaxPeers
	ErrPeerSetClosed         = peermgr.ErrPeerSetClosed
	ErrTooManyInbound        = peermgr.ErrTooManyInbound
	ErrTooManyOutbound       = peermgr.ErrTooManyOutbound
	ErrPeerBanned            = peermgr.ErrPeerBanned
	ErrPeerExists            = peermgr.ErrPeerExists
	ErrPeerUnknown           = peermgr.ErrPeerUnknown
)

// Peer management constructors.
func NewPeer(id, remoteAddr string, caps []Cap) *Peer {
	return peermgr.NewPeer(id, remoteAddr, caps)
}

func NewPeerSet() *PeerSet { return peermgr.NewPeerSet() }

func NewManagedPeerSet(maxPeers int) *ManagedPeerSet {
	return peermgr.NewManagedPeerSet(maxPeers)
}

func NewAdvPeerManager(config PeerManagerConfig) *AdvPeerManager {
	return peermgr.NewAdvPeerManager(config)
}
