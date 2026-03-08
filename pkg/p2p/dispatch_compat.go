package p2p

// dispatch_compat.go re-exports types from p2p/dispatch for backward compatibility.
// Consumers should migrate to importing p2p/dispatch directly.

import "github.com/eth2030/eth2030/p2p/dispatch"

// Protocol dispatcher type aliases.
type (
	ProtoMsgHandler  = dispatch.ProtoMsgHandler
	ProtoVersionSpec = dispatch.ProtoVersionSpec
	ProtoDispatcher  = dispatch.ProtoDispatcher
)

// Protocol manager type aliases.
type (
	Capability            = dispatch.Capability
	PeerMgrInfo           = dispatch.PeerMgrInfo
	ProtocolHandler       = dispatch.ProtocolHandler
	ConnectFunc           = dispatch.ConnectFunc
	ProtocolManagerConfig = dispatch.ProtocolManagerConfig
	ProtocolManager       = dispatch.ProtocolManager
)

// Protocol negotiation type aliases.
type (
	ProtoCapability    = dispatch.ProtoCapability
	KnownProtocol      = dispatch.KnownProtocol
	ProtoNegConfig     = dispatch.ProtoNegConfig
	ProtoHandshakeFunc = dispatch.ProtoHandshakeFunc
	ProtoNeg           = dispatch.ProtoNeg
)

// Message router type aliases.
type (
	RouterHandler = dispatch.RouterHandler
	MessageRouter = dispatch.MessageRouter
	RouterStats   = dispatch.RouterStats
	RouterConfig  = dispatch.RouterConfig
	OutboundMsg   = dispatch.OutboundMsg
)

// Protocol dispatcher errors.
var (
	ErrProtoDispatcherClosed     = dispatch.ErrProtoDispatcherClosed
	ErrProtoHandlerExists        = dispatch.ErrProtoHandlerExists
	ErrProtoNoVersionHandler     = dispatch.ErrProtoNoVersionHandler
	ErrProtoVersionNotRegistered = dispatch.ErrProtoVersionNotRegistered
	ErrProtoPeerIncompatible     = dispatch.ErrProtoPeerIncompatible
)

// Protocol manager errors.
var (
	ErrPeerAlreadyConnected = dispatch.ErrPeerAlreadyConnected
	ErrPeerNotConnected     = dispatch.ErrPeerNotConnected
	ErrTooManyPeers         = dispatch.ErrTooManyPeers
	ErrProtocolExists       = dispatch.ErrProtocolExists
	ErrNoSharedCaps         = dispatch.ErrNoSharedCaps
)

// Protocol negotiation errors.
var (
	ErrNegNoSharedProtocols = dispatch.ErrNegNoSharedProtocols
	ErrNegHandshakeTimeout  = dispatch.ErrNegHandshakeTimeout
	ErrNegVersionMismatch   = dispatch.ErrNegVersionMismatch
	ErrNegDuplicateProtocol = dispatch.ErrNegDuplicateProtocol
	ErrNegInvalidOffset     = dispatch.ErrNegInvalidOffset
)

// Message router errors.
var (
	ErrRouterClosed     = dispatch.ErrRouterClosed
	ErrNoHandler        = dispatch.ErrNoHandler
	ErrRateLimited      = dispatch.ErrRateLimited
	ErrResponseTimeout  = dispatch.ErrResponseTimeout
	ErrDuplicateHandler = dispatch.ErrDuplicateHandler
	ErrQueueFull        = dispatch.ErrQueueFull
	ErrPeerNotTracked   = dispatch.ErrPeerNotTracked
)

// Message router priority constants.
const (
	PriorityHigh   = dispatch.PriorityHigh
	PriorityNormal = dispatch.PriorityNormal
	PriorityLow    = dispatch.PriorityLow
)

// Constructors.
func NewProtoDispatcher(name string) *ProtoDispatcher  { return dispatch.NewProtoDispatcher(name) }
func DefaultProtoNegConfig() ProtoNegConfig            { return dispatch.DefaultProtoNegConfig() }
func NewProtoNeg(cfg ProtoNegConfig) *ProtoNeg         { return dispatch.NewProtoNeg(cfg) }
func NewMessageRouter(cfg RouterConfig) *MessageRouter { return dispatch.NewMessageRouter(cfg) }
func NewProtocolManager(cfg ProtocolManagerConfig) *ProtocolManager {
	return dispatch.NewProtocolManager(cfg)
}
func MatchCapabilities(local, remote []Capability) []Capability {
	return dispatch.MatchCapabilities(local, remote)
}
func FindProtocol(caps []ProtoCapability, name string) *ProtoCapability {
	return dispatch.FindProtocol(caps, name)
}
func MessageToProtocol(caps []ProtoCapability, wireCode uint64) (*ProtoCapability, uint64, error) {
	return dispatch.MessageToProtocol(caps, wireCode)
}
func CapsToCaps(caps []ProtoCapability) []Cap   { return dispatch.CapsToCaps(caps) }
func CapsFromCaps(caps []Cap) []ProtoCapability { return dispatch.CapsFromCaps(caps) }
