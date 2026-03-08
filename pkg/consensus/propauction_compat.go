package consensus

// propauction_compat.go re-exports types from consensus/propauction for backward compatibility.

import "github.com/eth2030/eth2030/consensus/propauction"

// Proposer auction type aliases.
type (
	AuctionBid                 = propauction.AuctionBid
	AuctionClearing            = propauction.AuctionClearing
	ProposerAuction            = propauction.ProposerAuction
	ProposerScheduleEntry      = propauction.ProposerScheduleEntry
	CommitteeRotationEntry     = propauction.CommitteeRotationEntry
	AuctionedProposerConfig    = propauction.AuctionedProposerConfig
	AuctionedProposerSelection = propauction.AuctionedProposerSelection
)

// Proposer auction error aliases.
var (
	ErrAuctionSlotPast      = propauction.ErrAuctionSlotPast
	ErrAuctionAlreadyOpen   = propauction.ErrAuctionAlreadyOpen
	ErrAuctionNotOpen       = propauction.ErrAuctionNotOpen
	ErrAuctionAlreadyClosed = propauction.ErrAuctionAlreadyClosed
	ErrAuctionDuplicateBid  = propauction.ErrAuctionDuplicateBid
	ErrAuctionZeroBid       = propauction.ErrAuctionZeroBid
	ErrAuctionNoBids        = propauction.ErrAuctionNoBids
	ErrAuctionInvalidCommit = propauction.ErrAuctionInvalidCommit
)

// Proposer auction function wrappers.
func DefaultAuctionedProposerConfig() AuctionedProposerConfig {
	return propauction.DefaultAuctionedProposerConfig()
}
func NewAuctionedProposerSelection(config AuctionedProposerConfig) *AuctionedProposerSelection {
	return propauction.NewAuctionedProposerSelection(config)
}
