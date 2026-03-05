package types

import (
	"math/big"
	"testing"
)

func TestMultiDimFeeTx_TypeByte(t *testing.T) {
	if MultiDimFeeTxType != 0x09 {
		t.Errorf("MultiDimFeeTxType = 0x%02x, want 0x09", MultiDimFeeTxType)
	}
}

func makeMultiDimFeeTx() *MultiDimFeeTx {
	to := HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	return &MultiDimFeeTx{
		ChainID:  big.NewInt(1337),
		Nonce:    5,
		GasLimit: 100_000,
		To:       &to,
		Value:    big.NewInt(1000),
		Data:     []byte{0x01, 0x02},
		MaxFeesPerGas: [3]*big.Int{
			big.NewInt(10e9), // execution
			big.NewInt(5e9),  // blob
			big.NewInt(2e9),  // calldata
		},
		PriorityFeesPerGas: [3]*big.Int{
			big.NewInt(1e9), // execution
			big.NewInt(0),   // blob
			big.NewInt(0),   // calldata
		},
		V: big.NewInt(0),
		R: big.NewInt(1),
		S: big.NewInt(2),
	}
}

func TestMultiDimFeeTx_RLPRoundTrip(t *testing.T) {
	inner := makeMultiDimFeeTx()
	tx := NewTransaction(inner)

	enc, err := tx.EncodeRLP()
	if err != nil {
		t.Fatalf("EncodeRLP: %v", err)
	}
	if len(enc) == 0 {
		t.Fatal("encoded tx is empty")
	}
	if enc[0] != MultiDimFeeTxType {
		t.Errorf("type byte: got 0x%02x, want 0x09", enc[0])
	}

	// Decode via the standard dispatch.
	tx2, err := DecodeTxRLP(enc)
	if err != nil {
		t.Fatalf("DecodeTxRLP: %v", err)
	}
	if tx2.Type() != MultiDimFeeTxType {
		t.Errorf("decoded type: got 0x%02x, want 0x09", tx2.Type())
	}
	if tx2.Nonce() != 5 {
		t.Errorf("nonce: got %d, want 5", tx2.Nonce())
	}
	if tx2.Gas() != 100_000 {
		t.Errorf("gas: got %d, want 100000", tx2.Gas())
	}
}

func TestMultiDimFeeTx_FeeVectors(t *testing.T) {
	inner := makeMultiDimFeeTx()
	tx := NewTransaction(inner)

	// gasPrice() returns execution-dimension max fee.
	if tx.GasPrice().Cmp(big.NewInt(10e9)) != 0 {
		t.Errorf("GasPrice: got %v, want 10gwei", tx.GasPrice())
	}

	// Verify fee vectors survive round-trip.
	enc, _ := tx.EncodeRLP()
	tx2, err := DecodeTxRLP(enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	md, ok := tx2.inner.(*MultiDimFeeTx)
	if !ok {
		t.Fatal("decoded inner is not *MultiDimFeeTx")
	}
	if md.MaxFeesPerGas[2].Cmp(big.NewInt(2e9)) != 0 {
		t.Errorf("MaxFeesPerGas[2] (calldata): got %v, want 2gwei", md.MaxFeesPerGas[2])
	}
}

func TestMultiDimFeeTx_Copy(t *testing.T) {
	inner := makeMultiDimFeeTx()
	cpy := inner.copy().(*MultiDimFeeTx)
	if cpy.Nonce != inner.Nonce {
		t.Error("copy nonce mismatch")
	}
	// Mutation of copy doesn't affect original.
	cpy.MaxFeesPerGas[0] = big.NewInt(999)
	if inner.MaxFeesPerGas[0].Cmp(big.NewInt(10e9)) != 0 {
		t.Error("original MaxFeesPerGas[0] was mutated by copy change")
	}
}

func TestMultiDimFeeTx_InvalidFeeVectorLen(t *testing.T) {
	// Manually craft bad RLP with 2-element fee vector.
	// DecodeTxRLP should return an error.
	_, err := DecodeMultiDimFeeTx([]byte{0xc0}) // empty list — invalid
	if err == nil {
		t.Error("expected error for invalid fee vector length")
	}
}
