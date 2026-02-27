package consensus

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
)

func TestNewFinalityBLSAdapterWithPQ(t *testing.T) {
	adapter := NewFinalityBLSAdapterWithPQ()
	if !adapter.PQFallbackEnabled {
		t.Error("PQFallbackEnabled should be true for PQ-enabled adapter")
	}

	// Standard adapter should have PQ disabled.
	std := NewFinalityBLSAdapter()
	if std.PQFallbackEnabled {
		t.Error("PQFallbackEnabled should be false for standard adapter")
	}
}

func TestSignVotePQ_Enabled(t *testing.T) {
	adapter := NewFinalityBLSAdapterWithPQ()
	digest := []byte("test-vote-digest-for-pq-signing")

	sig, err := adapter.SignVotePQ(digest)
	if err != nil {
		t.Fatalf("SignVotePQ with PQ enabled should not error, got: %v", err)
	}
	if sig == nil {
		t.Fatal("expected non-nil signature")
	}
	if len(sig) != 32 {
		t.Errorf("expected 32-byte signature, got %d bytes", len(sig))
	}
}

func TestSignVotePQ_Disabled(t *testing.T) {
	adapter := NewFinalityBLSAdapter() // PQ disabled

	_, err := adapter.SignVotePQ([]byte("some-digest"))
	if err == nil {
		t.Fatal("expected error when PQ fallback is disabled")
	}
}

func TestVerifyVotePQ_Valid(t *testing.T) {
	adapter := NewFinalityBLSAdapterWithPQ()
	digest := []byte("verify-this-digest")

	sig, err := adapter.SignVotePQ(digest)
	if err != nil {
		t.Fatal(err)
	}

	if !adapter.VerifyVotePQ(digest, sig) {
		t.Error("VerifyVotePQ should return true for a valid signature")
	}
}

func TestVerifyVotePQ_Invalid(t *testing.T) {
	adapter := NewFinalityBLSAdapterWithPQ()
	digest := []byte("original-digest")

	sig, err := adapter.SignVotePQ(digest)
	if err != nil {
		t.Fatal(err)
	}

	// Verify with a different digest should fail.
	wrongDigest := []byte("different-digest")
	if adapter.VerifyVotePQ(wrongDigest, sig) {
		t.Error("VerifyVotePQ should return false for mismatched digest")
	}
}

func TestVerifyVotePQ_Disabled(t *testing.T) {
	adapter := NewFinalityBLSAdapter() // PQ disabled

	// Even with a plausible signature, verification should fail.
	fakeSig := make([]byte, 32)
	if adapter.VerifyVotePQ([]byte("digest"), fakeSig) {
		t.Error("VerifyVotePQ should return false when PQ is disabled")
	}
}

func TestGenerateFinalityProof_WithPQ(t *testing.T) {
	adapter := NewFinalityBLSAdapterWithPQ()
	secret := big.NewInt(42)
	pk := crypto.BLSPubkeyFromSecret(secret)

	blockRoot := types.BytesToHash([]byte("pq-finalized-root"))
	stateRoot := types.BytesToHash([]byte("pq-state-root"))

	vote := SSFRoundVote{
		ValidatorPubkeyHash: types.BytesToHash(pk[:20]),
		Slot:                500,
		BlockRoot:           blockRoot,
		Stake:               32 * GweiPerETH,
	}
	vote.Signature = adapter.SignVote(secret, &vote)

	round := &SSFRound{
		Slot:      500,
		BlockRoot: blockRoot,
		Finalized: true,
		Votes:     map[types.Hash]*SSFRoundVote{vote.ValidatorPubkeyHash: &vote},
	}

	proof, err := adapter.GenerateFinalityProof(round, 50, stateRoot)
	if err != nil {
		t.Fatalf("GenerateFinalityProof with PQ failed: %v", err)
	}

	// PQSignature should be non-nil and non-empty.
	if proof.PQSignature == nil {
		t.Fatal("PQSignature should be non-nil when PQ is enabled")
	}
	if len(proof.PQSignature) == 0 {
		t.Fatal("PQSignature should be non-empty when PQ is enabled")
	}
}

func TestGenerateFinalityProof_WithoutPQ(t *testing.T) {
	adapter := NewFinalityBLSAdapter() // PQ disabled
	secret := big.NewInt(42)
	pk := crypto.BLSPubkeyFromSecret(secret)

	blockRoot := types.BytesToHash([]byte("no-pq-root"))
	stateRoot := types.BytesToHash([]byte("no-pq-state"))

	vote := SSFRoundVote{
		ValidatorPubkeyHash: types.BytesToHash(pk[:20]),
		Slot:                600,
		BlockRoot:           blockRoot,
		Stake:               32 * GweiPerETH,
	}
	vote.Signature = adapter.SignVote(secret, &vote)

	round := &SSFRound{
		Slot:      600,
		BlockRoot: blockRoot,
		Finalized: true,
		Votes:     map[types.Hash]*SSFRoundVote{vote.ValidatorPubkeyHash: &vote},
	}

	proof, err := adapter.GenerateFinalityProof(round, 60, stateRoot)
	if err != nil {
		t.Fatalf("GenerateFinalityProof without PQ failed: %v", err)
	}

	// PQSignature should be nil when PQ is disabled.
	if proof.PQSignature != nil {
		t.Errorf("PQSignature should be nil when PQ is disabled, got %v", proof.PQSignature)
	}
}
