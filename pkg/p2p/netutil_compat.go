package p2p

// netutil_compat.go re-exports types from p2p/netutil for backward compatibility.

import "github.com/eth2030/eth2030/p2p/netutil"

// Netutil type aliases.
type (
	IBLTCell                  = netutil.IBLTCell
	IBLT                      = netutil.IBLT
	SetReconciliationProtocol = netutil.SetReconciliationProtocol
	BandwidthTrackerConfig    = netutil.BandwidthTrackerConfig
	BandwidthStats            = netutil.BandwidthStats
	GlobalBandwidthStats      = netutil.GlobalBandwidthStats
	BandwidthTracker          = netutil.BandwidthTracker
	ConnDirection             = netutil.ConnDirection
	ConnLimConfig             = netutil.ConnLimConfig
	ConnEntry                 = netutil.ConnEntry
	ConnLim                   = netutil.ConnLim
)

// Netutil constants.
const (
	BandwidthPriorityConsensus  = netutil.BandwidthPriorityConsensus
	BandwidthPriorityBlocks     = netutil.BandwidthPriorityBlocks
	BandwidthPriorityTxs        = netutil.BandwidthPriorityTxs
	BandwidthPriorityBlobs      = netutil.BandwidthPriorityBlobs
	DefaultGlobalUploadLimit    = netutil.DefaultGlobalUploadLimit
	DefaultGlobalDownloadLimit  = netutil.DefaultGlobalDownloadLimit
	DefaultPerPeerUploadLimit   = netutil.DefaultPerPeerUploadLimit
	DefaultPerPeerDownloadLimit = netutil.DefaultPerPeerDownloadLimit
	ConnInbound                 = netutil.ConnInbound
	ConnOutbound                = netutil.ConnOutbound
)

// Netutil error variables.
var (
	ErrBWGlobalUploadLimit   = netutil.ErrBWGlobalUploadLimit
	ErrBWGlobalDownloadLimit = netutil.ErrBWGlobalDownloadLimit
	ErrBWPeerUploadLimit     = netutil.ErrBWPeerUploadLimit
	ErrBWPeerDownloadLimit   = netutil.ErrBWPeerDownloadLimit
	ErrBWPriorityExhausted   = netutil.ErrBWPriorityExhausted
	ErrBWUnknownPeer         = netutil.ErrBWUnknownPeer
	ErrConnLimMaxPeers       = netutil.ErrConnLimMaxPeers
	ErrConnLimSubnet16Full   = netutil.ErrConnLimSubnet16Full
	ErrConnLimSubnet24Full   = netutil.ErrConnLimSubnet24Full
	ErrConnLimInboundFull    = netutil.ErrConnLimInboundFull
	ErrConnLimOutboundFull   = netutil.ErrConnLimOutboundFull
	ErrConnLimRateLimited    = netutil.ErrConnLimRateLimited
	ErrConnLimDuplicate      = netutil.ErrConnLimDuplicate
	ErrConnLimReservedSlot   = netutil.ErrConnLimReservedSlot
	ErrConnLimAlreadyTracked = netutil.ErrConnLimAlreadyTracked
)

// Netutil function wrappers.
func NewIBLT(cells, k int) *IBLT { return netutil.NewIBLT(cells, k) }
func NewSetReconciliationProtocol(cells, k int) *SetReconciliationProtocol {
	return netutil.NewSetReconciliationProtocol(cells, k)
}
func DefaultBandwidthTrackerConfig() BandwidthTrackerConfig {
	return netutil.DefaultBandwidthTrackerConfig()
}
func NewBandwidthTracker(config BandwidthTrackerConfig) *BandwidthTracker {
	return netutil.NewBandwidthTracker(config)
}
func PriorityName(priority int) string      { return netutil.PriorityName(priority) }
func DefaultConnLimConfig() ConnLimConfig   { return netutil.DefaultConnLimConfig() }
func NewConnLim(cfg ConnLimConfig) *ConnLim { return netutil.NewConnLim(cfg) }
