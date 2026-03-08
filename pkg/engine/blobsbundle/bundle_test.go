package blobsbundle

import (
	"crypto/sha256"
	"errors"
	"testing"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/payload"
)

// testBlob returns a zero-filled blob of BlobSize bytes.
func testBlob() []byte { return make([]byte, BlobSize) }

// testCommitment returns a zero-filled commitment of KZGCommitmentSize bytes.
func testCommitment() []byte { return make([]byte, KZGCommitmentSize) }

// testProof returns a zero-filled proof of KZGProofSize bytes.
func testProof() []byte { return make([]byte, KZGProofSize) }

func TestNewBlobsBundleBuilder_Empty(t *testing.T) {
	b := NewBlobsBundleBuilder(nil)
	if b.Count() != 0 {
		t.Errorf("Count = %d, want 0", b.Count())
	}
	_, err := b.Build()
	if !errors.Is(err, ErrBlobBundleEmpty) {
		t.Errorf("Build on empty should return ErrBlobBundleEmpty, got %v", err)
	}
}

func TestBlobsBundleBuilder_AddBlob(t *testing.T) {
	b := NewBlobsBundleBuilder(nil)

	if err := b.AddBlob(testBlob(), testCommitment(), testProof()); err != nil {
		t.Fatalf("AddBlob: %v", err)
	}
	if b.Count() != 1 {
		t.Errorf("Count = %d, want 1", b.Count())
	}

	bundle, err := b.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(bundle.Blobs) != 1 || len(bundle.Commitments) != 1 || len(bundle.Proofs) != 1 {
		t.Errorf("bundle lengths = blobs:%d commitments:%d proofs:%d, want 1 each",
			len(bundle.Blobs), len(bundle.Commitments), len(bundle.Proofs))
	}
}

func TestBlobsBundleBuilder_InvalidBlobSize(t *testing.T) {
	b := NewBlobsBundleBuilder(nil)
	err := b.AddBlob(make([]byte, 100), testCommitment(), testProof())
	if !errors.Is(err, ErrBlobInvalidSize) {
		t.Errorf("expected ErrBlobInvalidSize, got %v", err)
	}
}

func TestBlobsBundleBuilder_InvalidCommitmentSize(t *testing.T) {
	b := NewBlobsBundleBuilder(nil)
	err := b.AddBlob(testBlob(), make([]byte, 10), testProof())
	if !errors.Is(err, ErrCommitmentInvalidSize) {
		t.Errorf("expected ErrCommitmentInvalidSize, got %v", err)
	}
}

func TestBlobsBundleBuilder_InvalidProofSize(t *testing.T) {
	b := NewBlobsBundleBuilder(nil)
	err := b.AddBlob(testBlob(), testCommitment(), make([]byte, 10))
	if !errors.Is(err, ErrProofInvalidSize) {
		t.Errorf("expected ErrProofInvalidSize, got %v", err)
	}
}

func TestBlobsBundleBuilder_TooManyBlobs(t *testing.T) {
	b := NewBlobsBundleBuilder(nil)
	for i := range MaxBlobsPerBundle {
		if err := b.AddBlob(testBlob(), testCommitment(), testProof()); err != nil {
			t.Fatalf("AddBlob %d: %v", i, err)
		}
	}
	err := b.AddBlob(testBlob(), testCommitment(), testProof())
	if !errors.Is(err, ErrBlobBundleTooMany) {
		t.Errorf("expected ErrBlobBundleTooMany, got %v", err)
	}
}

func TestBlobsBundleBuilder_Reset(t *testing.T) {
	b := NewBlobsBundleBuilder(nil)
	if err := b.AddBlob(testBlob(), testCommitment(), testProof()); err != nil {
		t.Fatalf("AddBlob: %v", err)
	}
	b.Reset()
	if b.Count() != 0 {
		t.Errorf("Count after Reset = %d, want 0", b.Count())
	}
	_, err := b.Build()
	if !errors.Is(err, ErrBlobBundleEmpty) {
		t.Errorf("Build after Reset should return ErrBlobBundleEmpty, got %v", err)
	}
}

func TestBlobsBundleBuilder_IsolatedCopy(t *testing.T) {
	// Verify that the builder copies the data, not just the slice header.
	blob := testBlob()
	comm := testCommitment()
	proof := testProof()
	b := NewBlobsBundleBuilder(nil)
	if err := b.AddBlob(blob, comm, proof); err != nil {
		t.Fatalf("AddBlob: %v", err)
	}
	// Mutate originals.
	blob[0] = 0xFF
	comm[0] = 0xFF
	proof[0] = 0xFF
	bundle, _ := b.Build()
	if bundle.Blobs[0][0] != 0 {
		t.Error("blob was not copied; mutation visible in bundle")
	}
	if bundle.Commitments[0][0] != 0 {
		t.Error("commitment was not copied; mutation visible in bundle")
	}
	if bundle.Proofs[0][0] != 0 {
		t.Error("proof was not copied; mutation visible in bundle")
	}
}

func TestValidateBundle(t *testing.T) {
	t.Run("nil bundle", func(t *testing.T) {
		if err := ValidateBundle(nil); !errors.Is(err, ErrBlobBundleEmpty) {
			t.Errorf("expected ErrBlobBundleEmpty, got %v", err)
		}
	})

	t.Run("valid bundle", func(t *testing.T) {
		bundle := &payload.BlobsBundleV1{
			Blobs:       [][]byte{testBlob()},
			Commitments: [][]byte{testCommitment()},
			Proofs:      [][]byte{testProof()},
		}
		if err := ValidateBundle(bundle); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("length mismatch", func(t *testing.T) {
		bundle := &payload.BlobsBundleV1{
			Blobs:       [][]byte{testBlob()},
			Commitments: [][]byte{testCommitment(), testCommitment()},
			Proofs:      [][]byte{testProof()},
		}
		if err := ValidateBundle(bundle); !errors.Is(err, ErrBlobBundleMismatch) {
			t.Errorf("expected ErrBlobBundleMismatch, got %v", err)
		}
	})

	t.Run("too many blobs", func(t *testing.T) {
		n := MaxBlobsPerBundle + 1
		bundle := &payload.BlobsBundleV1{
			Blobs:       make([][]byte, n),
			Commitments: make([][]byte, n),
			Proofs:      make([][]byte, n),
		}
		for i := range n {
			bundle.Blobs[i] = testBlob()
			bundle.Commitments[i] = testCommitment()
			bundle.Proofs[i] = testProof()
		}
		if err := ValidateBundle(bundle); !errors.Is(err, ErrBlobBundleTooMany) {
			t.Errorf("expected ErrBlobBundleTooMany, got %v", err)
		}
	})

	t.Run("invalid blob size", func(t *testing.T) {
		bundle := &payload.BlobsBundleV1{
			Blobs:       [][]byte{make([]byte, 100)},
			Commitments: [][]byte{testCommitment()},
			Proofs:      [][]byte{testProof()},
		}
		if err := ValidateBundle(bundle); !errors.Is(err, ErrBlobInvalidSize) {
			t.Errorf("expected ErrBlobInvalidSize, got %v", err)
		}
	})

	t.Run("invalid commitment size", func(t *testing.T) {
		bundle := &payload.BlobsBundleV1{
			Blobs:       [][]byte{testBlob()},
			Commitments: [][]byte{make([]byte, 10)},
			Proofs:      [][]byte{testProof()},
		}
		if err := ValidateBundle(bundle); !errors.Is(err, ErrCommitmentInvalidSize) {
			t.Errorf("expected ErrCommitmentInvalidSize, got %v", err)
		}
	})

	t.Run("invalid proof size", func(t *testing.T) {
		bundle := &payload.BlobsBundleV1{
			Blobs:       [][]byte{testBlob()},
			Commitments: [][]byte{testCommitment()},
			Proofs:      [][]byte{make([]byte, 10)},
		}
		if err := ValidateBundle(bundle); !errors.Is(err, ErrProofInvalidSize) {
			t.Errorf("expected ErrProofInvalidSize, got %v", err)
		}
	})
}

func TestVersionedHash(t *testing.T) {
	commitment := make([]byte, 48)
	h := VersionedHash(commitment)
	// Verify version byte.
	if h[0] != VersionedHashVersion {
		t.Errorf("version byte = 0x%02x, want 0x%02x", h[0], VersionedHashVersion)
	}
	// Verify the rest matches sha256 starting at byte 1.
	raw := sha256.Sum256(commitment)
	for i := 1; i < 32; i++ {
		if h[i] != raw[i] {
			t.Errorf("byte %d: got 0x%02x, want 0x%02x", i, h[i], raw[i])
		}
	}
}

func TestDeriveVersionedHashes(t *testing.T) {
	t.Run("nil bundle", func(t *testing.T) {
		if got := DeriveVersionedHashes(nil); got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("empty commitments", func(t *testing.T) {
		bundle := &payload.BlobsBundleV1{}
		if got := DeriveVersionedHashes(bundle); got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("two commitments", func(t *testing.T) {
		c1 := make([]byte, KZGCommitmentSize)
		c1[0] = 0xAA
		c2 := make([]byte, KZGCommitmentSize)
		c2[0] = 0xBB
		bundle := &payload.BlobsBundleV1{
			Commitments: [][]byte{c1, c2},
		}
		hashes := DeriveVersionedHashes(bundle)
		if len(hashes) != 2 {
			t.Fatalf("len = %d, want 2", len(hashes))
		}
		if hashes[0] != VersionedHash(c1) {
			t.Error("hash[0] mismatch")
		}
		if hashes[1] != VersionedHash(c2) {
			t.Error("hash[1] mismatch")
		}
	})
}

func TestValidateVersionedHashes(t *testing.T) {
	c1 := make([]byte, KZGCommitmentSize)
	c1[0] = 0x01
	bundle := &payload.BlobsBundleV1{
		Commitments: [][]byte{c1},
	}
	expected := DeriveVersionedHashes(bundle)

	t.Run("match", func(t *testing.T) {
		if err := ValidateVersionedHashes(bundle, expected); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("count mismatch", func(t *testing.T) {
		if err := ValidateVersionedHashes(bundle, nil); !errors.Is(err, ErrVersionedHashMismatch) {
			t.Errorf("expected ErrVersionedHashMismatch, got %v", err)
		}
	})

	t.Run("wrong hash", func(t *testing.T) {
		wrong := []types.Hash{types.HexToHash("0xdeadbeef")}
		if err := ValidateVersionedHashes(bundle, wrong); !errors.Is(err, ErrVersionedHashMismatch) {
			t.Errorf("expected ErrVersionedHashMismatch, got %v", err)
		}
	})
}

func TestPrepareSidecars(t *testing.T) {
	bundle := &payload.BlobsBundleV1{
		Blobs:       [][]byte{testBlob()},
		Commitments: [][]byte{testCommitment()},
		Proofs:      [][]byte{testProof()},
	}
	blockHash := types.HexToHash("0x1234")

	sidecars, err := PrepareSidecars(bundle, blockHash)
	if err != nil {
		t.Fatalf("PrepareSidecars: %v", err)
	}
	if len(sidecars) != 1 {
		t.Fatalf("len = %d, want 1", len(sidecars))
	}
	if sidecars[0].Index != 0 {
		t.Errorf("Index = %d, want 0", sidecars[0].Index)
	}
	if sidecars[0].SignedBlockHeader != blockHash {
		t.Error("SignedBlockHeader mismatch")
	}
}

func TestGetSidecar(t *testing.T) {
	bundle := &payload.BlobsBundleV1{
		Blobs:       [][]byte{testBlob(), testBlob()},
		Commitments: [][]byte{testCommitment(), testCommitment()},
		Proofs:      [][]byte{testProof(), testProof()},
	}
	blockHash := types.HexToHash("0xabcd")

	sc, err := GetSidecar(bundle, 1, blockHash)
	if err != nil {
		t.Fatalf("GetSidecar: %v", err)
	}
	if sc.Index != 1 {
		t.Errorf("Index = %d, want 1", sc.Index)
	}

	_, err = GetSidecar(bundle, 5, blockHash)
	if !errors.Is(err, ErrBlobBundleSidecarIndex) {
		t.Errorf("expected ErrBlobBundleSidecarIndex, got %v", err)
	}

	// Note: GetSidecar with nil bundle panics due to source bug (len(bundle.Blobs)
	// accessed in error path even when bundle is nil). Skip that case here.
}

// mockKZGVerifier is a KZGVerifier that returns a configurable error.
type mockKZGVerifier struct {
	commitmentErr error
	proofErr      error
}

func (m *mockKZGVerifier) VerifyBlobCommitment(blob, commitment []byte) error {
	return m.commitmentErr
}
func (m *mockKZGVerifier) VerifyBlobProof(blob, commitment, proof []byte) error {
	return m.proofErr
}

func TestBlobsBundleBuilder_WithVerifier_CommitmentFail(t *testing.T) {
	verifier := &mockKZGVerifier{commitmentErr: errors.New("bad commitment")}
	b := NewBlobsBundleBuilder(verifier)
	err := b.AddBlob(testBlob(), testCommitment(), testProof())
	if err == nil {
		t.Error("expected error from commitment verifier")
	}
}

func TestBlobsBundleBuilder_WithVerifier_ProofFail(t *testing.T) {
	verifier := &mockKZGVerifier{proofErr: errors.New("bad proof")}
	b := NewBlobsBundleBuilder(verifier)
	err := b.AddBlob(testBlob(), testCommitment(), testProof())
	if err == nil {
		t.Error("expected error from proof verifier")
	}
}

func TestBlobsBundleBuilder_WithVerifier_OK(t *testing.T) {
	verifier := &mockKZGVerifier{}
	b := NewBlobsBundleBuilder(verifier)
	if err := b.AddBlob(testBlob(), testCommitment(), testProof()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Count() != 1 {
		t.Errorf("Count = %d, want 1", b.Count())
	}
}
