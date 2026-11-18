package types

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

const (
	FrameMaxBlobs     = 6
	FrameCellsPerBlob = 128
	FrameBytesPerBlob = 131072
)

type frameNetwork8141 struct {
	Body        rlp.RawValue
	Version     uint8
	Blobs       [][]byte
	Commitments [][]byte
	Proofs      [][]byte
}

// ValidateFrameSidecar8141 checks the EIP-7594 shape and commitment hashes.
// The public pool additionally verifies every cell proof before admission.
func ValidateFrameSidecar8141(tx *Transaction) error {
	frame, ok := tx.Inner().(*FrameTx8141)
	if !ok {
		return errors.New("not a canonical Frame envelope")
	}
	count := len(frame.BlobVersionedHashes)
	sidecar := tx.BlobSidecar()
	if count == 0 {
		if sidecar != nil {
			return errors.New("non-blob Frame carries a sidecar")
		}
		return nil
	}
	if count > FrameMaxBlobs || sidecar == nil || len(sidecar.Blobs) != count || len(sidecar.Commitments) != count || len(sidecar.Proofs) != count*FrameCellsPerBlob {
		return errors.New("Frame sidecar has mismatched blob, commitment or cell-proof counts")
	}
	for i, commitment := range sidecar.Commitments {
		if len(commitment) != 48 || len(sidecar.Blobs[i]) != FrameBytesPerBlob {
			return errors.New("invalid Frame blob or commitment length")
		}
		hash := sha256.Sum256(commitment)
		hash[0] = 1
		if Hash(hash) != frame.BlobVersionedHashes[i] {
			return fmt.Errorf("Frame blob commitment %d does not match versioned hash", i)
		}
	}
	for _, proof := range sidecar.Proofs {
		if len(proof) != 48 {
			return errors.New("invalid Frame cell proof length")
		}
	}
	return nil
}

// EncodePooledTransaction preserves ordinary and canonical Frame blob sidecars.
// Block bodies and hashes continue to use Transaction.EncodeRLP.
func EncodePooledTransaction(tx *Transaction) ([]byte, error) {
	plain, err := tx.EncodeRLP()
	if err != nil {
		return nil, err
	}
	if tx.Type() == BlobTxType && tx.BlobSidecar() != nil {
		return encodePooledBlob(tx, plain)
	}
	if _, ok := tx.Inner().(*FrameTx8141); !ok || len(tx.BlobHashes()) == 0 {
		return plain, nil
	}
	if err := ValidateFrameSidecar8141(tx); err != nil {
		return nil, err
	}
	sidecar := tx.BlobSidecar()
	wrapped, err := rlp.EncodeToBytes(frameNetwork8141{Body: plain[1:], Version: 1, Blobs: sidecar.Blobs, Commitments: sidecar.Commitments, Proofs: sidecar.Proofs})
	if err != nil {
		return nil, err
	}
	return append([]byte{FrameTxType}, wrapped...), nil
}

func DecodePooledTransaction(raw []byte) (*Transaction, error) {
	if len(raw) == 0 || raw[0] != FrameTxType {
		return DecodeTxRLP(raw)
	}
	var fields []rlp.RawValue
	if err := rlp.DecodeBytes(raw[1:], &fields); err != nil {
		return nil, err
	}
	if len(fields) != 5 {
		return DecodeTxRLP(raw)
	}
	var wrapped frameNetwork8141
	if err := rlp.DecodeBytes(raw[1:], &wrapped); err != nil {
		return nil, err
	}
	if wrapped.Version != 1 {
		return nil, errors.New("unsupported Frame sidecar wrapper version")
	}
	tx, err := DecodeTxRLP(append([]byte{FrameTxType}, wrapped.Body...))
	if err != nil {
		return nil, err
	}
	if _, ok := tx.Inner().(*FrameTx8141); !ok || len(tx.BlobHashes()) == 0 {
		return nil, errors.New("sidecar wrapper requires a blob-bearing canonical Frame")
	}
	tx.SetBlobSidecar(&BlobSidecarData{Blobs: wrapped.Blobs, Commitments: wrapped.Commitments, Proofs: wrapped.Proofs})
	if err := ValidateFrameSidecar8141(tx); err != nil {
		return nil, err
	}
	return tx, nil
}
