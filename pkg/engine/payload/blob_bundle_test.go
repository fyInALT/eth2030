package payload

import (
	"bytes"
	"crypto/sha256"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto/bls"
	"testing"
)

func TestBlobBundlesRetainAuthenticatedProofKinds(t *testing.T) {
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
	tx := types.NewTransaction(&types.BlobTx{BlobHashes: []types.Hash{types.Hash(hash)}})
	sidecar := &types.BlobSidecarData{Version: 1, Blobs: [][]byte{blob}, Commitments: [][]byte{commitment[:]}}
	for _, proof := range proofs {
		sidecar.Proofs = append(sidecar.Proofs, bytes.Clone(proof[:]))
	}
	tx.SetBlobSidecar(sidecar)
	v2, err := BuildBlobsBundleV2([]*types.Transaction{tx})
	if err != nil || len(v2.Proofs) != 128 {
		t.Fatalf("cell bundle: %v", err)
	}
	for i, proof := range proofs {
		if !bytes.Equal(v2.Proofs[i], proof[:]) {
			t.Fatalf("cell proof %d changed", i)
		}
	}
	v1, err := BuildBlobsBundleV1([]*types.Transaction{tx})
	if err != nil || len(v1.Proofs) != 1 {
		t.Fatalf("legacy bundle: %v", err)
	}
	valid, err := backend.VerifyBlobProof(blob, commitment[:], v1.Proofs[0])
	if err != nil || !valid {
		t.Fatalf("not a genuine blob proof: %v", err)
	}
	sidecar.Version, sidecar.Proofs = 0, v1.Proofs
	expanded, err := BuildBlobsBundleV2([]*types.Transaction{tx})
	if err != nil || len(expanded.Proofs) != 128 {
		t.Fatalf("legacy expansion failed: %v", err)
	}
	for i, proof := range proofs {
		if !bytes.Equal(expanded.Proofs[i], proof[:]) {
			t.Fatalf("legacy expansion altered proof %d", i)
		}
	}
	// A generation failure must fail publication rather than return zero proofs.
	sidecar.Blobs[0] = bytes.Repeat([]byte{0xff}, 131072)
	if _, err = BuildBlobsBundleV2([]*types.Transaction{tx}); err == nil {
		t.Fatal("invalid field elements produced a successful blob bundle")
	}
	tx.SetBlobSidecar(&types.BlobSidecarData{})
	if _, err = BuildBlobsBundleV1([]*types.Transaction{tx}); err == nil {
		t.Fatal("incomplete sidecar produced a legacy bundle")
	}
	if _, err = BuildBlobsBundleV2([]*types.Transaction{tx}); err == nil {
		t.Fatal("incomplete sidecar produced a cell bundle")
	}
}
