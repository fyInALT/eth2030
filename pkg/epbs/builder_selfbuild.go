package epbs

// EIP-7732 §builder-self-build constants.
const (
	// BuilderIndexSelfBuild is a special builder index indicating the proposer
	// builds the payload itself, skipping the auction (UINT64_MAX per EIP-7732).
	BuilderIndexSelfBuild BuilderIndex = BuilderIndex(^uint64(0))

	// DomainProposerPreferences is the signing domain for ProposerPreferences
	// gossip messages (0x0D000000 per EIP-7732 §domain-types).
	DomainProposerPreferences uint32 = 0x0D000000
)

// ProcessBidWithSelfBuild checks if the bid uses the self-build index and
// skips the auction. Returns (winner, skipped).
// If BuilderIndex == BuilderIndexSelfBuild, the proposer builds the payload
// directly and no auction is run.
func (e *AuctionEngine) ProcessBidWithSelfBuild(bid *BuilderBid) (winner BuilderIndex, skipped bool) {
	if bid.BuilderIndex == BuilderIndexSelfBuild {
		return BuilderIndexSelfBuild, true
	}
	return 0, false
}
