package p2p

// wire_compat.go re-exports types from p2p/wire for backward compatibility.
// Consumers should migrate to importing p2p/wire directly.

import (
	"net"

	"github.com/eth2030/eth2030/p2p/wire"
)

// Wire transport type aliases.
type (
	Transport          = wire.Transport
	ConnTransport      = wire.ConnTransport
	Dialer             = wire.Dialer
	Listener           = wire.Listener
	TCPDialer          = wire.TCPDialer
	TCPListener        = wire.TCPListener
	FrameTransport     = wire.FrameTransport
	FrameConnTransport = wire.FrameConnTransport
	RLPxTransport      = wire.RLPxTransport
	RLPxHandshake      = wire.RLPxHandshake
	FrameCodec         = wire.FrameCodec
	FrameCodecConfig   = wire.FrameCodecConfig
	MsgPipeEnd         = wire.MsgPipeEnd
	Msg                = wire.Msg
	Message            = wire.Message
	HelloPacket        = wire.HelloPacket
	DisconnectReason   = wire.DisconnectReason
	Cap                = wire.Cap
	PeerConn           = wire.PeerConn
)

// Wire constants.
const (
	HelloMsg      = wire.HelloMsg
	DisconnectMsg = wire.DisconnectMsg
	PingMsg       = wire.PingMsg
	PongMsg       = wire.PongMsg

	DiscRequested        = wire.DiscRequested
	DiscNetworkError     = wire.DiscNetworkError
	DiscProtocolError    = wire.DiscProtocolError
	DiscUselessPeer      = wire.DiscUselessPeer
	DiscTooManyPeers     = wire.DiscTooManyPeers
	DiscAlreadyConnected = wire.DiscAlreadyConnected
	DiscSubprotocolError = wire.DiscSubprotocolError

	MaxMessageSize = wire.MaxMessageSize
)

// Wire error variables.
var (
	ErrTransportClosed = wire.ErrTransportClosed
	ErrFrameTooLarge   = wire.ErrFrameTooLarge
	ErrMessageTooLarge = wire.ErrMessageTooLarge
	ErrInvalidMsgCode  = wire.ErrInvalidMsgCode
	ErrDecode          = wire.ErrDecode

	ErrBadHandshake        = wire.ErrBadHandshake
	ErrBadMAC              = wire.ErrBadMAC
	ErrHandshakeTimeout    = wire.ErrHandshakeTimeout
	ErrIncompatibleVersion = wire.ErrIncompatibleVersion
	ErrNoMatchingCaps      = wire.ErrNoMatchingCaps

	ErrCodecClosed              = wire.ErrCodecClosed
	ErrSnappyDecompressTooLarge = wire.ErrSnappyDecompressTooLarge
	ErrPongTimeout              = wire.ErrPongTimeout
	ErrUnknownCapability        = wire.ErrUnknownCapability

	ErrECIESDecrypt     = wire.ErrECIESDecrypt
	ErrInvalidPubKey    = wire.ErrInvalidPubKey
	ErrFrameMACMismatch = wire.ErrFrameMACMismatch

	// ErrECIESAuthFailed, ErrECIESAckFailed, ErrECIESVersion from handshake_ecies.
	ErrECIESAuthFailed = wire.ErrECIESAuthFailed
	ErrECIESAckFailed  = wire.ErrECIESAckFailed
	ErrECIESVersion    = wire.ErrECIESVersion
)

// Wire function wrappers.
func MsgPipe() (*MsgPipeEnd, *MsgPipeEnd) { return wire.MsgPipe() }
func Send(t Transport, code uint64, data []byte) error {
	return wire.Send(t, code, data)
}
func NewFrameTransport(conn net.Conn) *FrameTransport {
	return wire.NewFrameTransport(conn)
}
func NewFrameConnTransport(conn net.Conn) *FrameConnTransport {
	return wire.NewFrameConnTransport(conn)
}
func NewTCPListener(ln net.Listener) *TCPListener { return wire.NewTCPListener(ln) }
func NewRLPxTransport(conn net.Conn) *RLPxTransport {
	return wire.NewRLPxTransport(conn)
}
func NewRLPxHandshake(initiator bool) (*RLPxHandshake, error) {
	return wire.NewRLPxHandshake(initiator)
}
func NewFrameCodec(conn net.Conn, cfg FrameCodecConfig) (*FrameCodec, error) {
	return wire.NewFrameCodec(conn, cfg)
}
func EncodeHello(h *HelloPacket) []byte             { return wire.EncodeHello(h) }
func DecodeHello(data []byte) (*HelloPacket, error) { return wire.DecodeHello(data) }
func PerformHandshake(tr Transport, local *HelloPacket) (*HelloPacket, error) {
	return wire.PerformHandshake(tr, local)
}
func MatchingCaps(local, remote []Cap) []Cap { return wire.MatchingCaps(local, remote) }
func EncodeMessage(code uint64, val interface{}) (Message, error) {
	return wire.EncodeMessage(code, val)
}
func DecodeMessage(msg Message, val interface{}) error { return wire.DecodeMessage(msg, val) }
func EncryptFrame(plaintext, secret []byte) ([]byte, error) {
	return wire.EncryptFrame(plaintext, secret)
}
func DecryptFrame(data, secret []byte) ([]byte, error) { return wire.DecryptFrame(data, secret) }
func DeriveFrameKeys(shared, initNonce, respNonce []byte) ([]byte, []byte) {
	return wire.DeriveFrameKeys(shared, initNonce, respNonce)
}
func GenerateNonce() ([32]byte, error) { return wire.GenerateNonce() }
