package consensus

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func makeTestPQAttestation(slot uint64, validatorIdx uint64) PQAttestation {
	var blockRoot types.Hash
	blockRoot[0] = byte(slot)
	blockRoot[1] = byte(validatorIdx)

	return PQAttestation{
		Slot:            slot,
		CommitteeIndex:  0,
		BeaconBlockRoot: blockRoot,
		SourceEpoch:     slot / 32,
		TargetEpoch:     slot/32 + 1,
		PQSignature:     []byte("dilithium-sig-placeholder-" + string(rune('A'+validatorIdx))),
		PQPublicKey:     []byte("dilithium-pk-placeholder-" + string(rune('A'+validatorIdx))),
		ValidatorIndex:  validatorIdx,
	}
}

func TestSTARKSignatureAggregation(t *testing.T) {
	agg := NewSTARKSignatureAggregator()

	attestations := []PQAttestation{
		makeTestPQAttestation(100, 0),
		makeTestPQAttestation(100, 1),
		makeTestPQAttestation(100, 2),
	}

	result, err := agg.Aggregate(attestations)
	if err != nil {
		t.Fatal(err)
	}

	if result.NumValidators != 3 {
		t.Errorf("expected 3 validators, got %d", result.NumValidators)
	}
	if result.Slot != 100 {
		t.Errorf("expected slot 100, got %d", result.Slot)
	}
	if result.AggregateProof == nil {
		t.Error("aggregate proof should not be nil")
	}

	var zeroHash types.Hash
	if result.CommitteeRoot == zeroHash {
		t.Error("committee root should not be zero")
	}
}

func TestSTARKSignatureAggregationVerify(t *testing.T) {
	agg := NewSTARKSignatureAggregator()

	attestations := []PQAttestation{
		makeTestPQAttestation(200, 0),
		makeTestPQAttestation(200, 1),
	}

	result, err := agg.Aggregate(attestations)
	if err != nil {
		t.Fatal(err)
	}

	valid, err := agg.Verify(result)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("aggregation should be valid")
	}
}

func TestSTARKSignatureAggregationEmpty(t *testing.T) {
	agg := NewSTARKSignatureAggregator()
	_, err := agg.Aggregate(nil)
	if err != ErrSTARKAggNoSigs {
		t.Errorf("expected ErrSTARKAggNoSigs, got %v", err)
	}
}

func TestSTARKSignatureAggregationVerifyNil(t *testing.T) {
	agg := NewSTARKSignatureAggregator()
	_, err := agg.Verify(nil)
	if err != ErrSTARKAggNilResult {
		t.Errorf("expected ErrSTARKAggNilResult, got %v", err)
	}
}

func TestSTARKSignatureAggregationGasSavings(t *testing.T) {
	agg := NewSTARKSignatureAggregator()

	// 10 validators: individual = 500,000 gas, STARK = 200,000 gas
	attestations := make([]PQAttestation, 10)
	for i := 0; i < 10; i++ {
		attestations[i] = makeTestPQAttestation(300, uint64(i))
	}

	result, err := agg.Aggregate(attestations)
	if err != nil {
		t.Fatal(err)
	}

	savings := result.GasSavings()
	// 10 * 50,000 - 200,000 = 300,000
	if savings != 300000 {
		t.Errorf("expected gas savings 300000, got %d", savings)
	}
}

func TestSTARKSignatureAggregationSmallSet(t *testing.T) {
	// With small set (< 5), individual verification may be cheaper.
	agg := NewSTARKSignatureAggregator()

	attestations := []PQAttestation{
		makeTestPQAttestation(400, 0),
		makeTestPQAttestation(400, 1),
		makeTestPQAttestation(400, 2),
	}

	result, err := agg.Aggregate(attestations)
	if err != nil {
		t.Fatal(err)
	}

	// 3 * 50,000 = 150,000 < 200,000, so no savings.
	if result.GasSavings() != 0 {
		t.Errorf("expected 0 gas savings for small set, got %d", result.GasSavings())
	}
}

func TestBatchVerifyPQAttestations(t *testing.T) {
	attestations := []PQAttestation{
		makeTestPQAttestation(500, 0),
		makeTestPQAttestation(500, 1),
		makeTestPQAttestation(500, 2),
		makeTestPQAttestation(500, 3),
		makeTestPQAttestation(500, 4),
	}

	result, err := BatchVerifyPQAttestations(attestations)
	if err != nil {
		t.Fatal(err)
	}
	if result.NumValidators != 5 {
		t.Errorf("expected 5 validators, got %d", result.NumValidators)
	}
}

func TestBatchVerifyPQAttestationsEmpty(t *testing.T) {
	_, err := BatchVerifyPQAttestations(nil)
	if err != ErrSTARKAggNoSigs {
		t.Errorf("expected ErrSTARKAggNoSigs, got %v", err)
	}
}

func TestAggregationStats(t *testing.T) {
	agg := NewSTARKSignatureAggregator()

	attestations := []PQAttestation{
		makeTestPQAttestation(600, 0),
		makeTestPQAttestation(600, 1),
	}

	result, err := agg.Aggregate(attestations)
	if err != nil {
		t.Fatal(err)
	}

	stats := result.Stats()
	if stats.NumSignatures != 2 {
		t.Errorf("expected 2 signatures, got %d", stats.NumSignatures)
	}
	if stats.ProofSizeBytes <= 0 {
		t.Error("proof size should be positive")
	}
	if stats.Slot != 600 {
		t.Errorf("expected slot 600, got %d", stats.Slot)
	}
}

func TestComputeCommitteeRoot(t *testing.T) {
	pubkeys := [][]byte{
		[]byte("pubkey-1"),
		[]byte("pubkey-2"),
		[]byte("pubkey-3"),
	}

	root := computeCommitteeRoot(pubkeys)
	var zeroHash types.Hash
	if root == zeroHash {
		t.Error("committee root should not be zero")
	}

	// Same keys should produce same root.
	root2 := computeCommitteeRoot(pubkeys)
	if root != root2 {
		t.Error("committee root should be deterministic")
	}

	// Different keys should produce different root.
	pubkeys2 := [][]byte{
		[]byte("pubkey-a"),
		[]byte("pubkey-b"),
	}
	root3 := computeCommitteeRoot(pubkeys2)
	if root == root3 {
		t.Error("different keys should produce different root")
	}
}

func TestComputeCommitteeRootEmpty(t *testing.T) {
	root := computeCommitteeRoot(nil)
	var zeroHash types.Hash
	if root != zeroHash {
		t.Error("empty pubkeys should produce zero root")
	}
}

func TestAggregateCount(t *testing.T) {
	// Nil aggregation.
	var nilAgg *STARKSignatureAggregation
	if nilAgg.AggregateCount() != 0 {
		t.Error("nil aggregation should have count 0")
	}

	agg := &STARKSignatureAggregation{NumValidators: 42}
	if agg.AggregateCount() != 42 {
		t.Errorf("expected count 42, got %d", agg.AggregateCount())
	}
}
