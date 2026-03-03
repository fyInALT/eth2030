package encrypted

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func testCommit() *CommitTx {
	return &CommitTx{
		CommitHash: types.Hash{0x01, 0x02, 0x03},
		Sender:     types.Address{0xAA},
		GasLimit:   21000,
		MaxFee:     big.NewInt(100),
		Timestamp:  1000,
	}
}

func TestGenerateCommitProof(t *testing.T) {
	commit := testCommit()
	proof, err := GenerateCommitProof(commit, 10_000_000, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proof == nil {
		t.Fatal("proof should not be nil")
	}
	if proof.CommitHash != commit.CommitHash {
		t.Error("commit hash mismatch")
	}
	if proof.BalanceProof == (types.Hash{}) {
		t.Error("balance proof should not be zero")
	}
	if proof.NonceProof == (types.Hash{}) {
		t.Error("nonce proof should not be zero")
	}
	if proof.GasProof == (types.Hash{}) {
		t.Error("gas proof should not be zero")
	}
	if proof.AggregateProof == (types.Hash{}) {
		t.Error("aggregate proof should not be zero")
	}
	if len(proof.ProofData) == 0 {
		t.Error("proof data should not be empty")
	}
}

func TestGenerateCommitProof_NilCommit(t *testing.T) {
	_, err := GenerateCommitProof(nil, 1000, 0)
	if err != ErrValidityNilCommit {
		t.Fatalf("expected ErrValidityNilCommit, got %v", err)
	}
}

func TestGenerateCommitProof_ZeroBalance(t *testing.T) {
	commit := testCommit()
	_, err := GenerateCommitProof(commit, 0, 0)
	if err != ErrValidityZeroBalance {
		t.Fatalf("expected ErrValidityZeroBalance, got %v", err)
	}
}

func TestGenerateCommitProof_InsufficientGas(t *testing.T) {
	commit := &CommitTx{
		CommitHash: types.Hash{0x01},
		Sender:     types.Address{0xAA},
		GasLimit:   21000,
		MaxFee:     big.NewInt(1000),
		Timestamp:  1000,
	}
	// Required: 21000 * 1000 = 21_000_000. Provide only 1_000_000.
	_, err := GenerateCommitProof(commit, 1_000_000, 0)
	if err != ErrValidityInsufficientGas {
		t.Fatalf("expected ErrValidityInsufficientGas, got %v", err)
	}
}

func TestVerifyCommitProof(t *testing.T) {
	commit := testCommit()
	proof, err := GenerateCommitProof(commit, 10_000_000, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	valid, err := VerifyCommitProof(proof)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("proof should be valid")
	}
}

func TestVerifyCommitProof_NilProof(t *testing.T) {
	_, err := VerifyCommitProof(nil)
	if err != ErrValidityNilProof {
		t.Fatalf("expected ErrValidityNilProof, got %v", err)
	}
}

func TestVerifyCommitProof_TamperedAggregate(t *testing.T) {
	commit := testCommit()
	proof, err := GenerateCommitProof(commit, 10_000_000, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Tamper with the aggregate proof.
	proof.AggregateProof[0] ^= 0xFF
	valid, err := VerifyCommitProof(proof)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("tampered proof should not verify")
	}
}

func TestVerifyCommitProof_TamperedProofData(t *testing.T) {
	commit := testCommit()
	proof, err := GenerateCommitProof(commit, 10_000_000, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Tamper with the proof data.
	proof.ProofData[0] ^= 0xFF
	valid, err := VerifyCommitProof(proof)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("tampered proof data should not verify")
	}
}

func TestProofSize(t *testing.T) {
	commit := testCommit()
	proof, err := GenerateCommitProof(commit, 10_000_000, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 4 hashes (32 bytes each) + 32 bytes proof data (keccak256 output).
	expected := 32*4 + 32
	if proof.ProofSize() != expected {
		t.Errorf("expected size %d, got %d", expected, proof.ProofSize())
	}

	// Nil proof returns 0.
	var nilProof *CommitValidityProof
	if nilProof.ProofSize() != 0 {
		t.Error("nil proof size should be 0")
	}
}

func TestProofDeterministic(t *testing.T) {
	commit := testCommit()
	proof1, err := GenerateCommitProof(commit, 10_000_000, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	proof2, err := GenerateCommitProof(commit, 10_000_000, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proof1.AggregateProof != proof2.AggregateProof {
		t.Error("proofs should be deterministic: aggregate mismatch")
	}
	if proof1.BalanceProof != proof2.BalanceProof {
		t.Error("proofs should be deterministic: balance mismatch")
	}
	if proof1.NonceProof != proof2.NonceProof {
		t.Error("proofs should be deterministic: nonce mismatch")
	}
	if proof1.GasProof != proof2.GasProof {
		t.Error("proofs should be deterministic: gas mismatch")
	}
	if len(proof1.ProofData) != len(proof2.ProofData) {
		t.Error("proofs should be deterministic: proof data length mismatch")
	}
	for i := range proof1.ProofData {
		if proof1.ProofData[i] != proof2.ProofData[i] {
			t.Error("proofs should be deterministic: proof data mismatch")
			break
		}
	}
}
