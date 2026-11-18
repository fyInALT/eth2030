package eiptest_test

import (
	"crypto/sha256"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto/bls"
	"github.com/eth2030/eth2030/eiptest"
	"math/big"
	"testing"
)

// This qualifies full-pool admission boundaries, not an implemented sparse relay.
func TestEIP8070_FullDataAdmissionBoundary(t *testing.T) {
	s := eiptest.New(t)
	pool := s.Pool()
	alice := s.Alice()
	makeTransaction := func(hash types.Hash) *types.Transaction {
		return eiptest.UnsignedBlob(alice.Address, s.Config().ChainID, s.Nonce(alice.Address), s.Bob().Address, big.NewInt(1), 100000, big.NewInt(2000000000), big.NewInt(1000000000), big.NewInt(1000000000), []types.Hash{hash})
	}
	// Unauthenticated metadata must be staged separately by a future sparse manager.
	pool.ExpectRejected(t, makeTransaction(eiptest.BlobHash(1)), "sidecar")
	blob := make([]byte, 131072)
	blob[31] = 1
	backend := bls.DefaultKZGBackend()
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
	valid := makeTransaction(types.Hash(hash))
	sidecar := &types.BlobSidecarData{Version: 1, Blobs: [][]byte{blob}, Commitments: [][]byte{commitment[:]}}
	for _, proof := range proofs {
		value := proof
		sidecar.Proofs = append(sidecar.Proofs, value[:])
	}
	valid.SetBlobSidecar(sidecar)
	pool.ExpectAccepted(t, valid)
	bad := types.Hash(hash)
	bad[0] = 2
	pool.ExpectRejected(t, makeTransaction(bad), "version")
}
