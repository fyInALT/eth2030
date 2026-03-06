package p2p

// nat_compat.go re-exports types from p2p/nat for backward compatibility.

import (
	"time"

	"github.com/eth2030/eth2030/p2p/nat"
)

// NAT manager type aliases.
type (
	NATType          = nat.NATType
	PortMapping      = nat.PortMapping
	NATDevice        = nat.NATDevice
	NATManagerConfig = nat.NATManagerConfig
	NATManager       = nat.NATManager
)

// NAT manager constants.
const (
	NATNone   = nat.NATNone
	NATUPnP   = nat.NATUPnP
	NATPMP    = nat.NATPMP
	NATManual = nat.NATManual
)

// NAT manager error variables.
var (
	ErrNATUnsupported = nat.ErrNATUnsupported
	ErrNATClosed      = nat.ErrNATClosed
	ErrMappingFailed  = nat.ErrMappingFailed
	ErrNoExternalIP   = nat.ErrNoExternalIP
)

// NAT manager function wrappers.
func NewNATManager(cfg NATManagerConfig) *NATManager { return nat.NewNATManager(cfg) }
func DiscoverGateway(timeout time.Duration) (NATDevice, error) {
	return nat.DiscoverGateway(timeout)
}

// NAT traversal type aliases.
type (
	NATTravType    = nat.NATTravType
	NATTravDevice  = nat.NATTravDevice
	NATTravMapping = nat.NATTravMapping
	NATTravConfig  = nat.NATTravConfig
	NATTrav        = nat.NATTrav
)

// NAT traversal constants.
const (
	NATTravNone   = nat.NATTravNone
	NATTravUPnP   = nat.NATTravUPnP
	NATTravPMP    = nat.NATTravPMP
	NATTravManual = nat.NATTravManual
)

// NAT traversal error variables.
var (
	ErrNATTravNoDevice        = nat.ErrNATTravNoDevice
	ErrNATTravClosed          = nat.ErrNATTravClosed
	ErrNATTravMappingNotFound = nat.ErrNATTravMappingNotFound
	ErrNATTravSTUNFailed      = nat.ErrNATTravSTUNFailed
)

// NAT traversal function wrappers.
func NewNATTrav(cfg NATTravConfig) *NATTrav { return nat.NewNATTrav(cfg) }
