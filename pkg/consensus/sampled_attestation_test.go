package consensus

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func makeSampledAttData(slot Slot, blockRoot types.Hash) AttestationData {
	return AttestationData{
		Slot:            slot,
		BeaconBlockRoot: blockRoot,
		Source:          Checkpoint{Epoch: 0, Root: types.Hash{}},
		Target:          Checkpoint{Epoch: 1, Root: blockRoot},
	}
}

// TestSampledAttestation_Basic verifies round-trip creation and validation (GAP-3.2).
func TestSampledAttestation_Basic(t *testing.T) {
	root := types.Hash{0xab}
	data := makeSampledAttData(10, root)
	var sig [96]byte
	sig[0] = 0x01

	att := NewSampledAttestation(data, 42, sig)
	if att == nil {
		t.Fatal("NewSampledAttestation returned nil")
	}
	if att.ValidatorIndices[0] != 42 {
		t.Errorf("ValidatorIndex = %d, want 42", att.ValidatorIndices[0])
	}
	if err := att.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}
}

// TestSampledAttestation_EmptySig verifies validation rejects empty signature (GAP-3.2).
func TestSampledAttestation_EmptySig(t *testing.T) {
	data := makeSampledAttData(1, types.Hash{})
	att := &SampledAttestation{
		Data:             data,
		ValidatorIndices: []ValidatorIndex{1},
		// Signature is zero → should fail.
	}
	if err := att.Validate(); err == nil {
		t.Error("expected validation error for empty signature")
	}
}

// TestSampledAttestation_NoValidators verifies validation rejects empty indices (GAP-3.2).
func TestSampledAttestation_NoValidators(t *testing.T) {
	var sig [96]byte
	sig[0] = 0x01
	att := &SampledAttestation{
		Data:      makeSampledAttData(1, types.Hash{}),
		Signature: sig,
	}
	if err := att.Validate(); err == nil {
		t.Error("expected validation error for no validator indices")
	}
}

// TestAggregateSampledAttestations_Basic verifies aggregating N attestations
// with the same data produces a single SampledAggregate (GAP-3.2).
func TestAggregateSampledAttestations_Basic(t *testing.T) {
	root := types.Hash{0xcc}
	data := makeSampledAttData(5, root)

	atts := make([]*SampledAttestation, 4)
	for i := range atts {
		var sig [96]byte
		sig[0] = byte(i + 1)
		atts[i] = NewSampledAttestation(data, ValidatorIndex(i), sig)
	}

	agg, err := AggregateSampledAttestations(atts)
	if err != nil {
		t.Fatalf("AggregateSampledAttestations: %v", err)
	}
	if len(agg.ValidatorIndices) != 4 {
		t.Errorf("got %d validators, want 4", len(agg.ValidatorIndices))
	}
	if !IsEqualAttestationData(&agg.Data, &data) {
		t.Error("aggregate data mismatch")
	}
}

// TestAggregateSampledAttestations_DataMismatch verifies an error is returned
// when attestations have different data (GAP-3.2).
func TestAggregateSampledAttestations_DataMismatch(t *testing.T) {
	data1 := makeSampledAttData(1, types.Hash{0x01})
	data2 := makeSampledAttData(2, types.Hash{0x02})

	var sig [96]byte
	sig[0] = 1
	atts := []*SampledAttestation{
		NewSampledAttestation(data1, 0, sig),
		NewSampledAttestation(data2, 1, sig),
	}

	if _, err := AggregateSampledAttestations(atts); err != ErrSampledAttDataMismatch {
		t.Errorf("expected ErrSampledAttDataMismatch, got %v", err)
	}
}

// TestAggregateSampledAttestations_Duplicate verifies an error is returned
// when the same validator appears twice (GAP-3.2).
func TestAggregateSampledAttestations_Duplicate(t *testing.T) {
	data := makeSampledAttData(1, types.Hash{0x01})
	var sig [96]byte
	sig[0] = 1
	atts := []*SampledAttestation{
		NewSampledAttestation(data, 5, sig),
		NewSampledAttestation(data, 5, sig), // duplicate
	}

	if _, err := AggregateSampledAttestations(atts); err != ErrSampledAttDuplicate {
		t.Errorf("expected ErrSampledAttDuplicate, got %v", err)
	}
}

// TestSampledForkChoiceWeight verifies that sampled attesters get correct
// scaled fork-choice weight (GAP-3.3).
func TestSampledForkChoiceWeight(t *testing.T) {
	// 256 sampled attesters each with 32 ETH = 8192 ETH base weight.
	// fullCommitteeSize = 512 (full committee is twice the sample).
	// Expected scaled weight = 8192 ETH * 512 / 256 = 16384 ETH.
	const numValidators = 256
	const fullCommitteeSize = 512
	const stakePerValidator = 32_000_000_000 // 32 ETH in Gwei

	data := makeSampledAttData(10, types.Hash{0xde})
	indices := make([]ValidatorIndex, numValidators)
	stakes := make(map[ValidatorIndex]uint64, numValidators)
	for i := range indices {
		indices[i] = ValidatorIndex(i)
		stakes[ValidatorIndex(i)] = stakePerValidator
	}

	agg := &SampledAggregate{
		Data:             data,
		ValidatorIndices: indices,
	}

	weight := SampledForkChoiceWeight(agg, stakes, fullCommitteeSize)
	expected := uint64(numValidators) * stakePerValidator * fullCommitteeSize / numValidators
	if weight != expected {
		t.Errorf("SampledForkChoiceWeight = %d, want %d", weight, expected)
	}
}

// TestSampledForkChoiceWeight_FullWeight verifies that when sampleSize equals
// fullCommitteeSize, the scaled weight equals the base weight (GAP-3.3).
func TestSampledForkChoiceWeight_FullWeight(t *testing.T) {
	const n = 256
	const stake = 32_000_000_000

	data := makeSampledAttData(1, types.Hash{})
	indices := make([]ValidatorIndex, n)
	stakes := make(map[ValidatorIndex]uint64, n)
	for i := range indices {
		indices[i] = ValidatorIndex(i)
		stakes[ValidatorIndex(i)] = stake
	}

	agg := &SampledAggregate{Data: data, ValidatorIndices: indices}
	// fullCommitteeSize == sampleSize → no scaling.
	weight := SampledForkChoiceWeight(agg, stakes, n)
	expected := uint64(n) * stake
	if weight != expected {
		t.Errorf("weight = %d, want %d (no scaling)", weight, expected)
	}
}

// TestAttesterModeInterop verifies mode detection: full committee mode uses
// Attestation type; sampled mode uses SampledAttestation (GAP-3.4).
func TestAttesterModeInterop(t *testing.T) {
	if ResolveAttestationMode(AttesterSampleFull) != AttestationModeFull {
		t.Error("sample size 0 should use full committee mode")
	}
	if ResolveAttestationMode(AttesterSample256) != AttestationModeSampled {
		t.Error("sample size 256 should use sampled mode")
	}
	if ResolveAttestationMode(AttesterSample512) != AttestationModeSampled {
		t.Error("sample size 512 should use sampled mode")
	}
	if ResolveAttestationMode(AttesterSample1024) != AttestationModeSampled {
		t.Error("sample size 1024 should use sampled mode")
	}
}

// TestScaleSampledWeight verifies weight scaling arithmetic (GAP-3.3).
func TestScaleSampledWeight(t *testing.T) {
	cases := []struct {
		base, full, sample, want uint64
	}{
		{100, 200, 100, 200}, // double scaling
		{100, 100, 100, 100}, // no scaling
		{50, 1000, 250, 200}, // 50 * 1000 / 250 = 200
		{0, 100, 100, 0},     // zero base
		{100, 100, 0, 0},     // zero sample (division by zero guard)
	}
	for _, c := range cases {
		got := ScaleSampledWeight(c.base, c.full, c.sample)
		if got != c.want {
			t.Errorf("ScaleSampledWeight(%d,%d,%d) = %d, want %d",
				c.base, c.full, c.sample, got, c.want)
		}
	}
}
