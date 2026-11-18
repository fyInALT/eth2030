package txpool

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto/bls"
)

func validateOrdinaryBlobSidecar(tx *types.Transaction) error {
	// The current full-data pool accepts only authenticated wrappers. Sparse
	// relay must stage partial data separately before promoting a transaction.
	sidecar := tx.BlobSidecar()
	if sidecar == nil {
		return errors.New("blob sidecar is missing")
	}
	if !types.HasFullBlobSidecar(tx) {
		return errors.New("blob sidecar has invalid version, counts or lengths")
	}
	for i, commitment := range sidecar.Commitments {
		hash := sha256.Sum256(commitment)
		hash[0] = 1
		if types.Hash(hash) != tx.BlobHashes()[i] {
			return fmt.Errorf("blob commitment %d does not match transaction versioned hash", i)
		}
	}
	if sidecar.Version == 1 {
		return verifyCellProofSidecar(sidecar)
	}
	backend := bls.DefaultKZGBackend()
	for i, blob := range sidecar.Blobs {
		valid, err := backend.VerifyBlobProof(blob, sidecar.Commitments[i], sidecar.Proofs[i])
		if err != nil {
			return fmt.Errorf("blob proof %d: %w", i, err)
		}
		if !valid {
			return fmt.Errorf("blob proof %d is invalid", i)
		}
	}
	return nil
}

// Both callers check complete cell-proof shape before this shared verifier.
func verifyCellProofSidecar(sidecar *types.BlobSidecarData) error {
	backend := bls.DefaultKZGBackend()
	for i, blob := range sidecar.Blobs {
		proofs := sidecar.Proofs[i*bls.KZGCellsPerExtBlob : (i+1)*bls.KZGCellsPerExtBlob]
		if batch, ok := backend.(interface {
			VerifyBlobCellProofs([]byte, []byte, [][]byte) (bool, error)
		}); ok {
			valid, err := batch.VerifyBlobCellProofs(blob, sidecar.Commitments[i], proofs)
			if err != nil {
				return fmt.Errorf("blob cell-proof batch %d: %w", i, err)
			}
			if !valid {
				return errors.New("blob has an invalid KZG cell-proof batch")
			}
			continue
		}
		cells, err := backend.ComputeCells(blob)
		if err != nil {
			return err
		}
		if len(cells) != bls.KZGCellsPerExtBlob {
			return errors.New("invalid KZG extension cell count")
		}
		for j, cell := range cells {
			valid, err := backend.VerifyCellProof(sidecar.Commitments[i], cell[:], proofs[j], uint64(j))
			if err != nil {
				return err
			}
			if !valid {
				return errors.New("blob has an invalid KZG cell proof")
			}
		}
	}
	return nil
}
