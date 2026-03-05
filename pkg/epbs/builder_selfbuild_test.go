package epbs

import (
	"testing"
)

func TestBuilderIndexSelfBuild(t *testing.T) {
	// BUILDER_INDEX_SELF_BUILD = UINT64_MAX.
	var want BuilderIndex = BuilderIndex(^uint64(0))
	if BuilderIndexSelfBuild != want {
		t.Errorf("BuilderIndexSelfBuild = %d, want UINT64_MAX", BuilderIndexSelfBuild)
	}
}

func TestAuctionEngine_SelfBuildSkipsAuction(t *testing.T) {
	engine := NewAuctionEngine(DefaultAuctionEngineConfig())

	bid := &BuilderBid{
		BuilderIndex: BuilderIndexSelfBuild,
		Slot:         10,
		Value:        0,
	}

	winner, skipped := engine.ProcessBidWithSelfBuild(bid)
	if !skipped {
		t.Error("expected auction to be skipped for self-build bid")
	}
	if winner != BuilderIndexSelfBuild {
		t.Errorf("winner: got %d, want BuilderIndexSelfBuild", winner)
	}
}

func TestProposerPreferencesDomain(t *testing.T) {
	if DomainProposerPreferences != 0x0D000000 {
		t.Errorf("DomainProposerPreferences = 0x%08X, want 0x0D000000", DomainProposerPreferences)
	}
}
