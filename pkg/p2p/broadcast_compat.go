package p2p

// broadcast_compat.go re-exports types from p2p/broadcast for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/p2p/broadcast"
	"github.com/eth2030/eth2030/p2p/gossip"
)

// Broadcast types.
type (
	BlockGossipConfig        = broadcast.BlockGossipConfig
	BlockAnnouncement        = broadcast.BlockAnnouncement
	GossipStats              = broadcast.GossipStats
	BlockGossipHandler       = broadcast.BlockGossipHandler
	BlockPieceMessage        = broadcast.BlockPieceMessage
	BlockAssembly            = broadcast.BlockAssembly
	BlockAssemblyConfig      = broadcast.BlockAssemblyConfig
	BlockAssemblyManager     = broadcast.BlockAssemblyManager
	SetCodeMessage           = broadcast.SetCodeMessage
	SetCodeGossipHandler     = broadcast.SetCodeGossipHandler
	SetCodeGossipHandlerFunc = broadcast.SetCodeGossipHandlerFunc
	SetCodeBroadcaster       = broadcast.SetCodeBroadcaster
	BroadcastResult          = broadcast.BroadcastResult
	BroadcastStats           = broadcast.BroadcastStats
	TopicFilter              = broadcast.TopicFilter
	Subscription             = broadcast.Subscription
	BroadcastMessage         = broadcast.BroadcastMessage
	PeerSender               = broadcast.PeerSender
	EIP2PBroadcaster         = broadcast.EIP2PBroadcaster
	MempoolBroadcaster       = broadcast.MempoolBroadcaster
)

// Broadcast constants.
const (
	SetCodeTopicPrefix          = broadcast.SetCodeTopicPrefix
	DefaultSetCodeRateLimit     = broadcast.DefaultSetCodeRateLimit
	DefaultSetCodeEpochDuration = broadcast.DefaultSetCodeEpochDuration
	DefaultFanout               = broadcast.DefaultFanout
	MinFanout                   = broadcast.MinFanout
	MaxFanout                   = broadcast.MaxFanout
	DefaultMaxMessageSize       = broadcast.DefaultMaxMessageSize
	DefaultSubscriptionBuffer   = broadcast.DefaultSubscriptionBuffer
)

// Broadcast errors.
var (
	ErrBlockGossipNilHash     = broadcast.ErrBlockGossipNilHash
	ErrBlockGossipNoPeers     = broadcast.ErrBlockGossipNoPeers
	ErrBlockGossipDuplicate   = broadcast.ErrBlockGossipDuplicate
	ErrBlockGossipEmptyPeer   = broadcast.ErrBlockGossipEmptyPeer
	ErrBlockGossipPeerExists  = broadcast.ErrBlockGossipPeerExists
	ErrPieceGossipNilPiece    = broadcast.ErrPieceGossipNilPiece
	ErrPieceGossipDuplicate   = broadcast.ErrPieceGossipDuplicate
	ErrPieceGossipExpired     = broadcast.ErrPieceGossipExpired
	ErrPieceGossipNoPeers     = broadcast.ErrPieceGossipNoPeers
	ErrPieceGossipComplete    = broadcast.ErrPieceGossipComplete
	ErrSetCodeNilMessage      = broadcast.ErrSetCodeNilMessage
	ErrSetCodeEmptyAuthority  = broadcast.ErrSetCodeEmptyAuthority
	ErrSetCodeInvalidChainID  = broadcast.ErrSetCodeInvalidChainID
	ErrSetCodeInvalidSig      = broadcast.ErrSetCodeInvalidSig
	ErrSetCodeDuplicate       = broadcast.ErrSetCodeDuplicate
	ErrSetCodeRateLimited     = broadcast.ErrSetCodeRateLimited
	ErrSetCodeBroadcasterStop = broadcast.ErrSetCodeBroadcasterStop
	ErrBroadcastClosed        = broadcast.ErrBroadcastClosed
	ErrBroadcastNilData       = broadcast.ErrBroadcastNilData
	ErrBroadcastEmptyType     = broadcast.ErrBroadcastEmptyType
	ErrBroadcastNoPeers       = broadcast.ErrBroadcastNoPeers
	ErrBroadcastTooLarge      = broadcast.ErrBroadcastTooLarge
	ErrBroadcastTopicEmpty    = broadcast.ErrBroadcastTopicEmpty
	ErrBroadcastNotSub        = broadcast.ErrBroadcastNotSub
	ErrBroadcastFanoutRange   = broadcast.ErrBroadcastFanoutRange
)

// Broadcast constructors.
func DefaultBlockGossipConfig() BlockGossipConfig { return broadcast.DefaultBlockGossipConfig() }
func NewBlockGossipHandler(cfg BlockGossipConfig) *BlockGossipHandler {
	return broadcast.NewBlockGossipHandler(cfg)
}
func DefaultBlockAssemblyConfig() BlockAssemblyConfig { return broadcast.DefaultBlockAssemblyConfig() }
func NewBlockAssemblyManager(cfg BlockAssemblyConfig) *BlockAssemblyManager {
	return broadcast.NewBlockAssemblyManager(cfg)
}
func BlockPieceTopicName(slot uint64, pieceIndex int) string {
	return broadcast.BlockPieceTopicName(slot, pieceIndex)
}
func PieceCustodyIndex(peerID string, slot uint64, totalPieces int) int {
	return broadcast.PieceCustodyIndex(peerID, slot, totalPieces)
}
func PeerCustodyPieces(peerID string, slot uint64, totalPieces, custodyCount int) []int {
	return broadcast.PeerCustodyPieces(peerID, slot, totalPieces, custodyCount)
}
func ValidateSetCodeAuth(msg *SetCodeMessage) bool { return broadcast.ValidateSetCodeAuth(msg) }
func ValidateSetCodeAuthWithChainID(msg *SetCodeMessage, localChainID *big.Int) bool {
	return broadcast.ValidateSetCodeAuthWithChainID(msg, localChainID)
}
func ValidateBroadcastConfig(b *SetCodeBroadcaster) error {
	return broadcast.ValidateBroadcastConfig(b)
}
func NewSetCodeBroadcaster(chainID *big.Int) *SetCodeBroadcaster {
	return broadcast.NewSetCodeBroadcaster(chainID)
}
func NewEIP2PBroadcaster() *EIP2PBroadcaster { return broadcast.NewEIP2PBroadcaster() }
func NewEIP2PBroadcasterWithSender(sender PeerSender) *EIP2PBroadcaster {
	return broadcast.NewEIP2PBroadcasterWithSender(sender)
}
func NewMempoolBroadcaster(tm *gossip.TopicManager) *MempoolBroadcaster {
	return broadcast.NewMempoolBroadcaster(tm)
}
