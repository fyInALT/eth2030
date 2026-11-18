package txpool

import (
	"bytes"
	"crypto/sha256"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto/bls"
	"math/big"
	"testing"
)

func TestOrdinaryBlobSidecarAuthenticityAndShape(t *testing.T) {
	backend := bls.DefaultKZGBackend()
	blob := make([]byte, 131072)
	blob[31] = 1
	commitment, err := backend.BlobToCommitment(blob)
	if err != nil {
		t.Fatal(err)
	}
	_, proofs, err := backend.ComputeCellsAndProofs(blob)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(commitment[:])
	hash[0] = 1
	base := &types.BlobSidecarData{Version: 1, Blobs: [][]byte{blob}, Commitments: [][]byte{commitment[:]}}
	for _, proof := range proofs {
		base.Proofs = append(base.Proofs, bytes.Clone(proof[:]))
	}
	clone := func() *types.BlobSidecarData {
		value := &types.BlobSidecarData{Version: base.Version}
		for _, item := range base.Blobs {
			value.Blobs = append(value.Blobs, bytes.Clone(item))
		}
		for _, item := range base.Commitments {
			value.Commitments = append(value.Commitments, bytes.Clone(item))
		}
		for _, item := range base.Proofs {
			value.Proofs = append(value.Proofs, bytes.Clone(item))
		}
		return value
	}
	for _, tc := range []struct {
		name   string
		mutate func(*types.BlobSidecarData)
		valid  bool
	}{
		{"valid_cells", func(*types.BlobSidecarData) {}, true},
		{"missing_blob", func(s *types.BlobSidecarData) { s.Blobs = nil }, false},
		{"missing_commitment", func(s *types.BlobSidecarData) { s.Commitments = nil }, false},
		{"missing_last_proof", func(s *types.BlobSidecarData) { s.Proofs = s.Proofs[:127] }, false},
		{"wrong_last_proof", func(s *types.BlobSidecarData) { s.Proofs[127] = bytes.Clone(s.Proofs[0]) }, false},
		{"wrong_blob", func(s *types.BlobSidecarData) { s.Blobs[0][31] ^= 1 }, false},
		{"wrong_hash_binding", func(s *types.BlobSidecarData) { s.Commitments[0][0] ^= 1 }, false},
		{"legacy_label_on_cells", func(s *types.BlobSidecarData) { s.Version = 0 }, false},
		{"unsupported_version", func(s *types.BlobSidecarData) { s.Version = 2 }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := types.NewTransaction(&types.BlobTx{BlobHashes: []types.Hash{types.Hash(hash)}, BlobFeeCap: big.NewInt(1)})
			sidecar := clone()
			tc.mutate(sidecar)
			tx.SetBlobSidecar(sidecar)
			if err := validateOrdinaryBlobSidecar(tx); (err == nil) != tc.valid {
				t.Fatalf("valid=%t err=%v", tc.valid, err)
			}
		})
	}
	legacyBackend, err := bls.NewGoEthKZGBackend()
	if err != nil {
		t.Fatal(err)
	}
	proof, err := legacyBackend.ComputeBlobProof(blob, commitment)
	if err != nil {
		t.Fatal(err)
	}
	legacy := &types.BlobSidecarData{Blobs: [][]byte{blob}, Commitments: [][]byte{commitment[:]}, Proofs: [][]byte{proof[:]}}
	tx := types.NewTransaction(&types.BlobTx{BlobHashes: []types.Hash{types.Hash(hash)}})
	tx.SetBlobSidecar(legacy)
	if err = validateOrdinaryBlobSidecar(tx); err != nil {
		t.Fatalf("valid legacy sidecar: %v", err)
	}
	legacy.Proofs[0] = make([]byte, 48)
	if err = validateOrdinaryBlobSidecar(tx); err == nil {
		t.Fatal("legacy KZG proof was not authenticated")
	}
}

// Fee-policy fixtures still need authentic data at successful admission.
func withAuthenticatedBlobSidecar(t *testing.T, tx *types.Transaction) *types.Transaction {
	t.Helper()
	inner := *tx.Inner().(*types.BlobTx)
	inner.BlobHashes = nil
	sidecar := &types.BlobSidecarData{}
	backend, err := bls.NewGoEthKZGBackend()
	if err != nil {
		t.Fatal(err)
	}
	for i := range tx.BlobHashes() {
		blob := make([]byte, 131072)
		blob[31] = byte(i + 1)
		commitment, err := backend.BlobToCommitment(blob)
		if err != nil {
			t.Fatal(err)
		}
		proof, err := backend.ComputeBlobProof(blob, commitment)
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(commitment[:])
		hash[0] = 1
		inner.BlobHashes = append(inner.BlobHashes, types.Hash(hash))
		sidecar.Blobs = append(sidecar.Blobs, blob)
		sidecar.Commitments = append(sidecar.Commitments, bytes.Clone(commitment[:]))
		sidecar.Proofs = append(sidecar.Proofs, bytes.Clone(proof[:]))
	}
	result := types.NewTransaction(&inner)
	if sender := tx.Sender(); sender != nil {
		result.SetSender(*sender)
	}
	result.SetBlobSidecar(sidecar)
	return result
}
