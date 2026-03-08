package wire

// Cap identifies a sub-protocol capability by name and version.
// Both sides advertise their capabilities during the devp2p hello handshake;
// overlapping capabilities are used to select active sub-protocols.
type Cap struct {
	Name    string
	Version uint
}

// PeerConn is the minimal interface that sub-protocol Run functions need about
// the remote peer. It avoids a direct dependency on the peermgr package.
type PeerConn interface {
	// ID returns the peer's unique identifier (hex-encoded or random string).
	ID() string
}

// Protocol represents a sub-protocol that runs on top of a devp2p connection.
type Protocol struct {
	// Name is the protocol name (e.g. "eth", "snap").
	Name string
	// Version is the protocol version number.
	Version uint
	// Length is the number of message codes used by this protocol.
	Length uint64
	// Run is called for each peer that supports this protocol.
	// It should read/write messages via t and return when done.
	// Returning an error causes the peer to be disconnected.
	Run func(peer PeerConn, t Transport) error
}
