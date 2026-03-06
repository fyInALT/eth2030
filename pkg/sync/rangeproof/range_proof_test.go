package rangeproof

import (
	"bytes"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func TestRangeproof_CreateAndVerify(t *testing.T) {
	rp := NewRangeProver()
	root := types.Hash{0x01, 0x02, 0x03}

	keys := [][]byte{{0x01, 0x00}, {0x02, 0x00}, {0x03, 0x00}}
	values := [][]byte{{0xaa}, {0xbb}, {0xcc}}

	proof := rp.CreateRangeProof(keys, values, root)
	if proof == nil {
		t.Fatal("CreateRangeProof returned nil")
	}
	if len(proof.Keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(proof.Keys))
	}

	ok, err := rp.VerifyRangeProof(root, proof)
	if err != nil {
		t.Fatalf("VerifyRangeProof: %v", err)
	}
	if !ok {
		t.Fatal("proof should be valid")
	}
}

func TestRangeproof_VerifyBadRoot(t *testing.T) {
	rp := NewRangeProver()
	root := types.Hash{0x01}
	wrongRoot := types.Hash{0xff}

	keys := [][]byte{{0x01}, {0x02}}
	values := [][]byte{{0xaa}, {0xbb}}
	proof := rp.CreateRangeProof(keys, values, root)

	ok, err := rp.VerifyRangeProof(wrongRoot, proof)
	if err == nil {
		t.Fatal("expected error for wrong root")
	}
	if ok {
		t.Fatal("proof should be invalid for wrong root")
	}
}

func TestRangeproof_SplitRange(t *testing.T) {
	rp := NewRangeProver()
	origin := []byte{0x00}
	limit := []byte{0xff}

	requests := rp.SplitRange(origin, limit, 4)
	if len(requests) != 4 {
		t.Fatalf("expected 4 ranges, got %d", len(requests))
	}
	if !bytes.Equal(requests[0].Origin, origin) {
		t.Fatalf("first range origin mismatch")
	}
	if !bytes.Equal(requests[len(requests)-1].Limit, limit) {
		t.Fatalf("last range limit mismatch")
	}
}

func TestRangeproof_PadTo32(t *testing.T) {
	b := []byte{0x01, 0x02}
	padded := PadTo32(b)
	if len(padded) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(padded))
	}
	if padded[30] != 0x01 || padded[31] != 0x02 {
		t.Fatal("wrong padding position")
	}
}

func TestRangeproof_ComputeRangeHash(t *testing.T) {
	keys := [][]byte{{0x01}, {0x02}}
	values := [][]byte{{0xaa}, {0xbb}}

	hash := ComputeRangeHash(keys, values)
	if hash.IsZero() {
		t.Fatal("range hash should not be zero")
	}
	if ComputeRangeHash(keys, values) != hash {
		t.Fatal("same input should produce same hash")
	}
}

func TestRangeproof_MergeProofs(t *testing.T) {
	rp := NewRangeProver()
	root := types.Hash{0x01}

	p1 := rp.CreateRangeProof([][]byte{{0x01}, {0x02}}, [][]byte{{0xaa}, {0xbb}}, root)
	p2 := rp.CreateRangeProof([][]byte{{0x03}, {0x04}}, [][]byte{{0xcc}, {0xdd}}, root)

	merged := rp.MergeRangeProofs([]*RangeProof{p1, p2})
	if len(merged.Keys) != 4 {
		t.Fatalf("expected 4 merged keys, got %d", len(merged.Keys))
	}
}

func TestRangeproof_AccountRange(t *testing.T) {
	ar := AccountRange{
		Start:    types.Hash{0x01},
		End:      types.Hash{0xff},
		Accounts: 100,
		Complete: true,
	}
	if ar.Accounts != 100 {
		t.Fatalf("expected 100 accounts, got %d", ar.Accounts)
	}
	if !ar.Complete {
		t.Fatal("should be complete")
	}
}
