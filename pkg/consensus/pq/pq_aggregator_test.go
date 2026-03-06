package pq

import (
	"testing"
)

// TestPQAggregatorTypes verifies that the PQ aggregator types are defined correctly.
func TestPQAggregatorTypes(t *testing.T) {
	duty := PQAggregatorDuty{
		ValidatorIndex: 1,
		Slot:           100,
		Epoch:          3,
	}
	if duty.ValidatorIndex != 1 || duty.Slot != 100 || duty.Epoch != 3 {
		t.Errorf("PQAggregatorDuty fields incorrect: %+v", duty)
	}

	bundle := XMSSSignatureBundle{
		ValidatorIndex: 5,
		Signature:      []byte("sig"),
		PublicKey:      []byte("pk"),
	}
	if bundle.ValidatorIndex != 5 {
		t.Errorf("XMSSSignatureBundle.ValidatorIndex: got %d, want 5", bundle.ValidatorIndex)
	}

	req := AggregateRequest{
		Slot:        200,
		MessageHash: [32]byte{0x01},
		Validators:  []uint64{1, 2, 3},
	}
	if req.Slot != 200 || len(req.Validators) != 3 {
		t.Errorf("AggregateRequest fields incorrect: %+v", req)
	}
}

// TestAggregatorDutySelection verifies SelectAggregators returns valid duties.
func TestAggregatorDutySelection(t *testing.T) {
	var epochRandao [32]byte
	for i := range epochRandao {
		epochRandao[i] = byte(i)
	}

	duties := SelectAggregators(100, 3, epochRandao, 50, 7)

	if len(duties) < 1 || len(duties) > 4 {
		t.Errorf("expected 1-4 duties, got %d", len(duties))
	}

	// All duties should have the correct slot/epoch.
	for _, d := range duties {
		if d.Slot != 100 {
			t.Errorf("duty slot: got %d, want 100", d.Slot)
		}
		if d.Epoch != 3 {
			t.Errorf("duty epoch: got %d, want 3", d.Epoch)
		}
		// No duty should equal the proposer index.
		if d.ValidatorIndex == 7 {
			t.Errorf("aggregator must not be the proposer (index 7)")
		}
		// Must be within valid range.
		if d.ValidatorIndex >= 50 {
			t.Errorf("validator index %d out of range [0,50)", d.ValidatorIndex)
		}
	}

	// Duties must be distinct.
	seen := make(map[uint64]bool)
	for _, d := range duties {
		if seen[d.ValidatorIndex] {
			t.Errorf("duplicate validator index %d in duties", d.ValidatorIndex)
		}
		seen[d.ValidatorIndex] = true
	}
}

// TestAggregatorDutySelectionDeterministic verifies that the same inputs produce the same duties.
func TestAggregatorDutySelectionDeterministic(t *testing.T) {
	var epochRandao [32]byte
	epochRandao[0] = 0xAB

	duties1 := SelectAggregators(50, 2, epochRandao, 100, 0)
	duties2 := SelectAggregators(50, 2, epochRandao, 100, 0)

	if len(duties1) != len(duties2) {
		t.Errorf("non-deterministic count: %d vs %d", len(duties1), len(duties2))
	}
	for i := range duties1 {
		if duties1[i].ValidatorIndex != duties2[i].ValidatorIndex {
			t.Errorf("non-deterministic duty[%d]: %d vs %d", i,
				duties1[i].ValidatorIndex, duties2[i].ValidatorIndex)
		}
	}
}

// TestAggregatorCollection verifies AddSignatureBundle and CollectSignatures.
func TestAggregatorCollection(t *testing.T) {
	agg := NewDefaultPQAggregator()

	b1 := XMSSSignatureBundle{ValidatorIndex: 1, Signature: []byte("sig1"), PublicKey: []byte("pk1")}
	b2 := XMSSSignatureBundle{ValidatorIndex: 2, Signature: []byte("sig2"), PublicKey: []byte("pk2")}

	if err := agg.AddSignatureBundle(b1); err != nil {
		t.Fatalf("AddSignatureBundle b1: %v", err)
	}
	if err := agg.AddSignatureBundle(b2); err != nil {
		t.Fatalf("AddSignatureBundle b2: %v", err)
	}

	bundles, err := agg.CollectSignatures(100, []uint64{1, 2})
	if err != nil {
		t.Fatalf("CollectSignatures: %v", err)
	}
	if len(bundles) != 2 {
		t.Errorf("expected 2 bundles, got %d", len(bundles))
	}
}

// TestAggregatorEndToEnd exercises the full aggregation pipeline.
func TestAggregatorEndToEnd(t *testing.T) {
	agg := NewDefaultPQAggregator()

	// Add some signature bundles.
	for i := uint64(0); i < 3; i++ {
		sig := make([]byte, 64)
		sig[0] = byte(i + 1)
		pk := make([]byte, 32)
		pk[0] = byte(i + 10)
		bundle := XMSSSignatureBundle{
			ValidatorIndex: i,
			Signature:      sig,
			PublicKey:      pk,
		}
		if err := agg.AddSignatureBundle(bundle); err != nil {
			t.Fatalf("AddSignatureBundle[%d]: %v", i, err)
		}
	}

	bundles, err := agg.CollectSignatures(42, []uint64{0, 1, 2})
	if err != nil {
		t.Fatalf("CollectSignatures: %v", err)
	}

	result, err := agg.ProduceAggregate(bundles)
	if err != nil {
		t.Fatalf("ProduceAggregate: %v", err)
	}
	if result == nil {
		t.Fatal("ProduceAggregate returned nil")
	}
	if result.NumValidators != 3 {
		t.Errorf("NumValidators: got %d, want 3", result.NumValidators)
	}

	if err := agg.PropagateAggregate(result); err != nil {
		t.Fatalf("PropagateAggregate: %v", err)
	}
}

// TestPQAggregatorInterface verifies DefaultPQAggregator satisfies the PQAggregator interface.
func TestPQAggregatorInterface(t *testing.T) {
	var _ PQAggregator = (*DefaultPQAggregator)(nil)
}
