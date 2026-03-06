package das

// cell_compat.go re-exports types from das/cell for backward compatibility.

import "github.com/eth2030/eth2030/das/cell"

// Cell type aliases.
type (
	CellMessage              = cell.CellMessage
	SubnetConfig             = cell.SubnetConfig
	GossipRouter             = cell.GossipRouter
	CellMessageEntry         = cell.CellMessageEntry
	CellMessageCodec         = cell.CellMessageCodec
	CellMessageHandlerFunc   = cell.CellMessageHandlerFunc
	CellMessageHandler       = cell.CellMessageHandler
	CellMessageHandlerStruct = cell.CellMessageHandlerStruct
	CellMessageRouter        = cell.CellMessageRouter
	CellReputationConfig     = cell.CellReputationConfig
	CellPeerStats            = cell.CellPeerStats
	SamplingWeight           = cell.SamplingWeight
	CellPeerScorer           = cell.CellPeerScorer
)

// Cell constants.
const (
	CellMessageVersion    = cell.CellMessageVersion
	CellMessageHeaderSize = cell.CellMessageHeaderSize
	MaxCellProofSize      = cell.MaxCellProofSize
)

// Cell error variables.
var (
	ErrCellNotInCustody      = cell.ErrCellNotInCustody
	ErrInvalidCellData       = cell.ErrInvalidCellData
	ErrInvalidBlobIndex      = cell.ErrInvalidBlobIndex
	ErrGossipCellIndex       = cell.ErrGossipCellIndex
	ErrNoBroadcastTarget     = cell.ErrNoBroadcastTarget
	ErrCellMsgNil            = cell.ErrCellMsgNil
	ErrCellMsgDataEmpty      = cell.ErrCellMsgDataEmpty
	ErrCellMsgDataTooLarge   = cell.ErrCellMsgDataTooLarge
	ErrCellMsgProofTooLarge  = cell.ErrCellMsgProofTooLarge
	ErrCellMsgCellIndex      = cell.ErrCellMsgCellIndex
	ErrCellMsgColumnIndex    = cell.ErrCellMsgColumnIndex
	ErrCellMsgRowIndex       = cell.ErrCellMsgRowIndex
	ErrCellMsgDecode         = cell.ErrCellMsgDecode
	ErrCellMsgVersion        = cell.ErrCellMsgVersion
	ErrBatchTooLarge         = cell.ErrBatchTooLarge
	ErrBatchEmpty            = cell.ErrBatchEmpty
	ErrBatchDecode           = cell.ErrBatchDecode
	ErrCellScorerNilConfig   = cell.ErrCellScorerNilConfig
	ErrCellScorerEmptyPeerID = cell.ErrCellScorerEmptyPeerID
	ErrCellScorerZeroDecay   = cell.ErrCellScorerZeroDecay
)

// Cell function wrappers.
func DefaultSubnetConfig() SubnetConfig                  { return cell.DefaultSubnetConfig() }
func NewGossipRouter(config SubnetConfig) *GossipRouter  { return cell.NewGossipRouter(config) }
func AssignSubnets(nodeID [32]byte, config SubnetConfig) []uint64 {
	return cell.AssignSubnets(nodeID, config)
}
func CellSubnet(cellIndex, numSubnets uint64) uint64 { return cell.CellSubnet(cellIndex, numSubnets) }
func ValidateCellMessage(msg *CellMessage) error     { return cell.ValidateCellMessage(msg) }
func GossipTopicForSubnet(subnet uint64) string      { return cell.GossipTopicForSubnet(subnet) }
func NewCellMessageCodec() *CellMessageCodec         { return cell.NewCellMessageCodec() }
func ValidateCellMessageEntry(msg *CellMessageEntry) error {
	return cell.ValidateCellMessageEntry(msg)
}
func ValidateCellMessageBatch(msgs []*CellMessageEntry) error {
	return cell.ValidateCellMessageBatch(msgs)
}
func NewCellMessageHandlerStruct() *CellMessageHandlerStruct {
	return cell.NewCellMessageHandlerStruct()
}
func NewCellMessageRouter() *CellMessageRouter        { return cell.NewCellMessageRouter() }
func DefaultCellReputationConfig() CellReputationConfig { return cell.DefaultCellReputationConfig() }
func ValidateCellReputationConfig(cfg CellReputationConfig) error {
	return cell.ValidateCellReputationConfig(cfg)
}
func NewCellPeerScorer(cfg CellReputationConfig) (*CellPeerScorer, error) {
	return cell.NewCellPeerScorer(cfg)
}
