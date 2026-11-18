package payload

import (
	"fmt"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto/bls"
)

// BuildBlobsBundleV1 returns genuine per-blob proofs, including when pool data
// arrived with PeerDAS cell proofs. Missing data fails payload publication.
func BuildBlobsBundleV1(txs []*types.Transaction) (*BlobsBundleV1, error) {
	bundle := &BlobsBundleV1{Blobs: [][]byte{}, Commitments: [][]byte{}, Proofs: [][]byte{}}
	for _, tx := range txs {
		if len(tx.BlobHashes()) == 0 {
			continue
		}
		if !types.HasFullBlobSidecar(tx) {
			return nil, fmt.Errorf("blob transaction %s has incomplete sidecar data", tx.Hash())
		}
		sidecar := tx.BlobSidecar()
		bundle.Blobs = append(bundle.Blobs, sidecar.Blobs...)
		bundle.Commitments = append(bundle.Commitments, sidecar.Commitments...)
		_, frame := tx.Inner().(*types.FrameTx8141)
		if sidecar.Version == 0 && !frame {
			bundle.Proofs = append(bundle.Proofs, sidecar.Proofs...)
			continue
		}
		backend, err := bls.NewGoEthKZGBackend()
		if err != nil {
			return nil, err
		}
		for i, blob := range sidecar.Blobs {
			proof, err := backend.ComputeBlobProof(blob, [48]byte(sidecar.Commitments[i]))
			if err != nil {
				return nil, fmt.Errorf("blob proof generation: %w", err)
			}
			bundle.Proofs = append(bundle.Proofs, proof[:])
		}
	}
	return bundle, nil
}

// BuildBlobsBundleV2 preserves authenticated cell proofs. Legacy per-blob
// sidecars are extended locally; a generation failure is never replaced by zeros.
func BuildBlobsBundleV2(txs []*types.Transaction) (*BlobsBundleV2, error) {
	bundle := &BlobsBundleV2{Blobs: [][]byte{}, Commitments: [][]byte{}, Proofs: [][]byte{}}
	for _, tx := range txs {
		if len(tx.BlobHashes()) == 0 {
			continue
		}
		if !types.HasFullBlobSidecar(tx) {
			return nil, fmt.Errorf("blob transaction %s has incomplete sidecar data", tx.Hash())
		}
		sidecar := tx.BlobSidecar()
		bundle.Blobs = append(bundle.Blobs, sidecar.Blobs...)
		bundle.Commitments = append(bundle.Commitments, sidecar.Commitments...)
		_, frame := tx.Inner().(*types.FrameTx8141)
		if sidecar.Version == 1 || frame {
			bundle.Proofs = append(bundle.Proofs, sidecar.Proofs...)
			continue
		}
		backend := bls.DefaultKZGBackend()
		for _, blob := range sidecar.Blobs {
			_, proofs, err := backend.ComputeCellsAndProofs(blob)
			if err != nil {
				return nil, fmt.Errorf("blob cell-proof generation: %w", err)
			}
			for _, proof := range proofs {
				value := proof
				bundle.Proofs = append(bundle.Proofs, value[:])
			}
		}
	}
	return bundle, nil
}
