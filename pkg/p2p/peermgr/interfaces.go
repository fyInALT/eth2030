// Package peermgr implements peer lifecycle management for the devp2p stack.
package peermgr

import "github.com/eth2030/eth2030/p2p/wire"

// MsgReadWriter combines message reading and writing for a single sub-protocol.
// Protocol handlers receive this interface to exchange messages with a peer.
type MsgReadWriter interface {
	// ReadMsg reads the next message for this protocol.
	ReadMsg() (wire.Msg, error)

	// WriteMsg sends a message to the remote peer.
	WriteMsg(msg wire.Msg) error
}

// PeerHandler is the callback interface for protocol-level peer lifecycle events.
type PeerHandler interface {
	// HandlePeer is called when a new peer connects with a compatible protocol.
	// Returning an error disconnects the peer.
	HandlePeer(peer *Peer, rw MsgReadWriter) error
}

// PeerHandlerFunc is an adapter to allow use of ordinary functions as PeerHandler.
type PeerHandlerFunc func(peer *Peer, rw MsgReadWriter) error

// HandlePeer calls f(peer, rw).
func (f PeerHandlerFunc) HandlePeer(peer *Peer, rw MsgReadWriter) error {
	return f(peer, rw)
}

// PeerInfo provides read-only information about a connected peer.
type PeerInfo interface {
	// ID returns the peer's unique identifier.
	ID() string

	// RemoteAddr returns the peer's remote network address.
	RemoteAddr() string

	// Caps returns the peer's advertised capabilities.
	Caps() []wire.Cap

	// Version returns the negotiated protocol version.
	Version() uint32
}

// PeerSetReader provides read-only access to the set of connected peers.
type PeerSetReader interface {
	// Peer returns the peer with the given ID, or nil.
	Peer(id string) *Peer

	// Len returns the number of connected peers.
	Len() int

	// Peers returns a snapshot of all connected peers.
	Peers() []*Peer

	// BestPeer returns the peer with the highest total difficulty.
	BestPeer() *Peer
}

// Verify interface compliance at compile time.
var _ PeerHandler = PeerHandlerFunc(nil)
var _ PeerInfo = (*Peer)(nil)
var _ PeerSetReader = (*PeerSet)(nil)
