package types

import (
	"fmt"
	gethrlp "github.com/ethereum/go-ethereum/rlp"
)

// HasFullBlobSidecar checks complete data shape for pool-based block building.
// Commitment binding and KZG authenticity are enforced by pool admission.
func HasFullBlobSidecar(tx *Transaction) bool {
	if tx == nil {
		return false
	}
	count := len(tx.BlobHashes())
	if count == 0 {
		return true
	}
	sidecar := tx.BlobSidecar()
	if count > 6 || sidecar == nil || len(sidecar.Blobs) != count || len(sidecar.Commitments) != count {
		return false
	}
	expected := count
	if _, frame := tx.Inner().(*FrameTx8141); frame {
		expected = count * 128
	} else {
		switch sidecar.Version {
		case 0:
		case 1:
			expected = count * 128
		default:
			return false
		}
	}
	if len(sidecar.Proofs) != expected {
		return false
	}
	for i, blob := range sidecar.Blobs {
		if len(blob) != 131072 || len(sidecar.Commitments[i]) != 48 {
			return false
		}
	}
	for _, proof := range sidecar.Proofs {
		if len(proof) != 48 {
			return false
		}
	}
	return true
}

func encodePooledBlob(tx *Transaction, plain []byte) ([]byte, error) {
	sidecar := tx.BlobSidecar()
	fields := []any{gethrlp.RawValue(plain[1:]), sidecar.Blobs, sidecar.Commitments, sidecar.Proofs}
	switch sidecar.Version {
	case 0:
	case 1:
		fields = []any{gethrlp.RawValue(plain[1:]), byte(1), sidecar.Blobs, sidecar.Commitments, sidecar.Proofs}
	default:
		return nil, fmt.Errorf("unsupported blob wrapper version %d", sidecar.Version)
	}
	encoded, err := gethrlp.EncodeToBytes(fields)
	if err != nil {
		return nil, err
	}
	return append([]byte{BlobTxType}, encoded...), nil
}
