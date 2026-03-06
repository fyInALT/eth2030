package p2p

// transport_compat.go re-exports types from p2p/transport for backward compatibility.

import (
	"time"

	"github.com/eth2030/eth2030/p2p/transport"
)

// Transport type aliases.
type (
	MixnetTransportMode     = transport.MixnetTransportMode
	AnonymousTransport      = transport.AnonymousTransport
	ExternalMixnetTransport = transport.ExternalMixnetTransport
	TransportConfig         = transport.TransportConfig
	TransportStats          = transport.TransportStats
	TransportManager        = transport.TransportManager
	FlashnetConfig          = transport.FlashnetConfig
	FlashnetTransport       = transport.FlashnetTransport
	TorConfig               = transport.TorConfig
	TorTransport            = transport.TorTransport
	NymConfig               = transport.NymConfig
	NymTransport            = transport.NymTransport
	MixnetConfig            = transport.MixnetConfig
	MixnetTransport         = transport.MixnetTransport
)

// Transport constants.
const (
	ModeSimulated = transport.ModeSimulated
	ModeTorSocks5 = transport.ModeTorSocks5
	ModeNymSocks5 = transport.ModeNymSocks5
)

// Transport error variables.
var (
	ErrAnonTransportClosed   = transport.ErrAnonTransportClosed
	ErrAnonTransportExists   = transport.ErrAnonTransportExists
	ErrAnonTransportNotFound = transport.ErrAnonTransportNotFound
	ErrAnonTransportNilTx    = transport.ErrAnonTransportNilTx
)

// Transport function wrappers.
func ParseMixnetMode(s string) (MixnetTransportMode, error) { return transport.ParseMixnetMode(s) }
func DefaultTransportConfig() TransportConfig               { return transport.DefaultTransportConfig() }
func FormatControlMessage(msg string, kohaku bool) []byte {
	return transport.FormatControlMessage(msg, kohaku)
}
func NewTransportManager() *TransportManager { return transport.NewTransportManager() }
func NewTransportManagerWithConfig(cfg TransportConfig) *TransportManager {
	return transport.NewTransportManagerWithConfig(cfg)
}
func DefaultFlashnetConfig() *FlashnetConfig { return transport.DefaultFlashnetConfig() }
func NewFlashnetTransport(config *FlashnetConfig) *FlashnetTransport {
	return transport.NewFlashnetTransport(config)
}
func DefaultTorConfig() *TorConfig                 { return transport.DefaultTorConfig() }
func NewTorTransport(cfg *TorConfig) *TorTransport { return transport.NewTorTransport(cfg) }
func NewNymTransport(cfg *NymConfig) *NymTransport { return transport.NewNymTransport(cfg) }
func DefaultMixnetConfig() *MixnetConfig           { return transport.DefaultMixnetConfig() }
func NewMixnetTransport(config *MixnetConfig) *MixnetTransport {
	return transport.NewMixnetTransport(config)
}
func WrapOnion(data []byte, hops int) []byte { return transport.WrapOnion(data, hops) }
func ProbeProxy(addr string, timeout time.Duration) bool {
	return transport.ProbeProxy(addr, timeout)
}
