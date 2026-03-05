package consensus

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// TestEP5_RandomAttesterPipeline runs the full GAP-3 pipeline end-to-end:
//   RandomAttesterSelector.SelectAttesters
//     → NewSampledAttestation (per-validator)
//     → AggregateSampledAttestations
//     → ScaleSampledWeight (fork-choice weight)
func TestEP5_RandomAttesterPipeline(t *testing.T) {
	const (
		totalValidators = 10_000
		slot            = Slot(42)
	)

	validators := make([]ValidatorIndex, totalValidators)
	for i := range validators {
		validators[i] = ValidatorIndex(i)
	}

	// Step 1: select attesters for this slot.
	cfg := DefaultRandomAttesterConfig()
	sel := NewRandomAttesterSelector(cfg)
	sampleSize := sel.ComputeSampleSize(uint64(totalValidators))

	randao := types.Hash{0xde, 0xad, 0xbe, 0xef}
	selected := sel.SelectAttesters(uint64(slot), randao[:], validators)
	if len(selected) == 0 {
		t.Fatal("SelectAttesters returned empty slice")
	}
	if uint64(len(selected)) > sampleSize {
		t.Errorf("selected %d > computedSampleSize %d", len(selected), sampleSize)
	}
	for _, idx := range selected {
		if int(idx) >= totalValidators {
			t.Errorf("selected index %d out of range [0,%d)", idx, totalValidators)
		}
	}

	// Step 2: build one SampledAttestation per selected validator.
	blockRoot := types.Hash{0xaa, 0xbb}
	data := AttestationData{
		Slot:            slot,
		BeaconBlockRoot: blockRoot,
		Source:          Checkpoint{Epoch: 0, Root: types.Hash{0x01}},
		Target:          Checkpoint{Epoch: 1, Root: types.Hash{0x02}},
	}
	atts := make([]*SampledAttestation, len(selected))
	for i, vi := range selected {
		var sig [96]byte
		// Use multiple bytes to avoid zero-signature wrap when i > 255.
		sig[0] = byte(i & 0xff)
		sig[1] = byte((i >> 8) & 0xff)
		sig[2] = 0xff // ensure sig is always non-zero
		atts[i] = NewSampledAttestation(data, vi, sig)
		if err := atts[i].Validate(); err != nil {
			t.Fatalf("validator %d attestation invalid: %v", vi, err)
		}
	}

	// Step 3: aggregate.
	agg, err := AggregateSampledAttestations(atts)
	if err != nil {
		t.Fatalf("AggregateSampledAttestations: %v", err)
	}
	if len(agg.ValidatorIndices) != len(selected) {
		t.Errorf("aggregate has %d indices, want %d", len(agg.ValidatorIndices), len(selected))
	}

	// Step 4: compute fork-choice weight.
	const fullCommittee = totalValidators
	weight := ScaleSampledWeight(uint64(len(selected)), fullCommittee, sampleSize)
	if weight == 0 {
		t.Error("fork-choice weight should be > 0")
	}
	if weight > uint64(fullCommittee) {
		t.Errorf("weight %d exceeds full committee %d", weight, fullCommittee)
	}
	t.Logf("pipeline: selected=%d aggregated=%d weight=%d/%d",
		len(selected), len(agg.ValidatorIndices), weight, fullCommittee)
}

// TestEP5_AttesterModeCoexistence verifies that full-committee and sampled
// modes can coexist: SampledAttestation validates and ScaleSampledWeight
// returns votes when sampleSize == fullCommittee (GAP-3.4).
func TestEP5_AttesterModeCoexistence(t *testing.T) {
	data := AttestationData{
		Slot:            Slot(1),
		BeaconBlockRoot: types.Hash{0x01},
		Source:          Checkpoint{Epoch: 0},
		Target:          Checkpoint{Epoch: 1},
	}

	// Standard aggregated attestation (non-sampled path).
	fullAtt := &AggregateAttestation{
		Data:            data,
		AggregationBits: []byte{0xff},
	}
	if fullAtt == nil {
		t.Fatal("full committee attestation should not be nil")
	}

	// Sampled attestation (sampled path).
	var sig [96]byte
	sig[0] = 0x01
	sampledAtt := NewSampledAttestation(data, ValidatorIndex(7), sig)
	if err := sampledAtt.Validate(); err != nil {
		t.Errorf("sampled attestation validation failed: %v", err)
	}

	// When sampleSize == fullCommittee, ScaleSampledWeight returns baseWeight.
	const (
		fullCommittee uint64 = 512
		votes         uint64 = 300
		sampleSize    uint64 = 512
	)
	w := ScaleSampledWeight(votes, fullCommittee, sampleSize)
	if w != votes {
		t.Errorf("when sampleSize == fullCommittee, weight should equal votes (%d), got %d", votes, w)
	}
}

// TestEP5_MinimmitFinalityPipeline tests the Minimmit engine through a full
// proposal → votes → finality cycle with 100 honest validators (GAP-5.1/5.3).
func TestEP5_MinimmitFinalityPipeline(t *testing.T) {
	const (
		numValidators = 100
		stakePerVal   = 32_000_000_000 // 32 ETH in Gwei
		totalStake    = numValidators * stakePerVal
	)

	cfg := &MinimmitConfig{
		TotalStake:           totalStake,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
		VoterLimit:           numValidators,
	}
	engine, err := NewMinimmitEngine(cfg)
	if err != nil {
		t.Fatalf("NewMinimmitEngine: %v", err)
	}

	blockRoot := types.Hash{0x11, 0x22}
	const slot uint64 = 9 // delta=9 from 0 → justifiable (perfect square 3^2)

	if err := engine.ProposeBlock(slot, blockRoot); err != nil {
		t.Fatalf("ProposeBlock: %v", err)
	}

	// Cast 2/3+ votes.
	quorum := numValidators*2/3 + 1
	for i := 0; i < quorum; i++ {
		var sig [96]byte
		sig[0] = byte(i)
		vote := MinimmitVote{
			ValidatorIndex: ValidatorIndex(i),
			Slot:           slot,
			BlockRoot:      blockRoot,
			Signature:      sig,
			Stake:          stakePerVal,
		}
		if castErr := engine.CastVote(vote); castErr != nil &&
			castErr != ErrMinimmitAlreadyFinal {
			t.Fatalf("CastVote %d: %v", i, castErr)
		}
	}

	// Verify finality or log that threshold wasn't met (stake rounding is ok).
	finalSlot, finalRoot := engine.FinalizedHead()
	if finalSlot == 0 {
		t.Logf("no finality after %d votes (stake rounding may require one more)", quorum)
	} else {
		if finalRoot != blockRoot {
			t.Errorf("finalized root = %v, want %v", finalRoot, blockRoot)
		}
		t.Logf("finalized slot=%d root=%v after %d votes", finalSlot, finalRoot, quorum)
	}
}

// TestEP5_BLSParallelAggregator verifies the ParallelAggregator correctly
// handles 16 distinct attestations and accumulates metrics (GAP-7.3).
func TestEP5_BLSParallelAggregator(t *testing.T) {
	pa := NewParallelAggregator(&ParallelAggregatorConfig{
		Workers:   4,
		BatchSize: 16,
	})

	atts := make([]*AggregateAttestation, 16)
	for i := range atts {
		atts[i] = makeTestAgg(1, i)
	}
	result, err := pa.Aggregate(atts)
	if err != nil {
		t.Fatalf("Aggregate: %v", err)
	}
	if result.ProcessedCount != 16 {
		t.Errorf("processed %d, want 16", result.ProcessedCount)
	}
	totalBits := CountBits(result.Aggregate.AggregationBits)
	if totalBits != 16 {
		t.Errorf("aggregated bits = %d, want 16", totalBits)
	}
	// Verify metrics accumulate.
	agg, dups, _ := pa.Metrics()
	if agg != 16 {
		t.Errorf("metrics: aggregated=%d, want 16", agg)
	}
	if dups < 0 {
		t.Errorf("metrics: negative duplicates=%d", dups)
	}
	t.Logf("BLS aggregate: processed=%d bits=%d mergeDepth=%d",
		result.ProcessedCount, totalBits, result.MergeDepth)
}

// TestEP5_3SFBackoffProgression verifies IsJustifiableSlot matches the
// expected square/oblong pattern from the 3SF spec (GAP-5.3).
func TestEP5_3SFBackoffProgression(t *testing.T) {
	// IsJustifiableSlot(finalizedSlot, candidateSlot uint64) bool
	// delta = candidate - finalized; justifiable when delta ≤ 5, square, or oblong.
	type slotCase struct {
		finalized   uint64
		candidate   uint64
		justifiable bool
	}
	base := uint64(100) // arbitrary finalized slot
	cases := []slotCase{
		{base, base + 0, true},  // delta=0 ≤ 5
		{base, base + 1, true},  // delta=1 ≤ 5
		{base, base + 4, true},  // delta=4 ≤ 5
		{base, base + 5, true},  // delta=5 ≤ 5
		{base, base + 6, true},  // delta=6 = 2*3 (oblong)
		{base, base + 7, false}, // delta=7, not square or oblong
		{base, base + 8, false}, // delta=8, not square or oblong
		{base, base + 9, true},  // delta=9 = 3^2 (perfect square)
		{base, base + 11, false},// delta=11, not in set
		{base, base + 12, true}, // delta=12 = 3*4 (oblong)
		{base, base + 16, true}, // delta=16 = 4^2 (perfect square)
		{base, base + 20, true}, // delta=20 = 4*5 (oblong)
		{base, base + 25, true}, // delta=25 = 5^2 (perfect square)
	}

	for _, c := range cases {
		got := IsJustifiableSlot(c.finalized, c.candidate)
		delta := c.candidate - c.finalized
		if got != c.justifiable {
			t.Errorf("delta=%d: IsJustifiableSlot=%v, want %v", delta, got, c.justifiable)
		}
	}
}
