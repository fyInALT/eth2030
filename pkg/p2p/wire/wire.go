// Package wire implements the low-level devp2p connection layer:
// message framing, RLPx encryption, ECIES handshake, protocol
// multiplexing, and capability negotiation. All types here are
// independent of chain data and peer identity beyond an ID string.
package wire

import "errors"

// Common transport errors.
var (
	// ErrTransportClosed is returned when reading/writing on a closed transport.
	ErrTransportClosed = errors.New("p2p/wire: transport closed")

	// ErrFrameTooLarge is returned when a frame exceeds MaxMessageSize.
	ErrFrameTooLarge = errors.New("p2p/wire: frame too large")

	// ErrMessageTooLarge is returned when a message exceeds the protocol size limit.
	ErrMessageTooLarge = errors.New("p2p/wire: message too large")

	// ErrInvalidMsgCode is returned when a message has an unrecognised code.
	ErrInvalidMsgCode = errors.New("p2p/wire: invalid message code")

	// ErrDecode is returned when RLP decoding fails.
	ErrDecode = errors.New("p2p/wire: decode error")
)

// MaxMessageSize is the maximum allowed size of a protocol message payload (16 MiB).
const MaxMessageSize = 16 * 1024 * 1024
