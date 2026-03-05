package types

import (
	"testing"
)

func TestFrameTxReceiptRLPRoundTrip(t *testing.T) {
	payer := Address{0xAB, 0xCD}
	r := &FrameTxReceipt{
		CumulativeGasUsed: 63000,
		Payer:             payer,
		FrameResults: []FrameResult{
			{Status: 1, GasUsed: 21000, Logs: []*Log{
				{Address: Address{0x01}, Topics: []Hash{{0xAA}}, Data: []byte{0x01}},
			}},
			{Status: 0, GasUsed: 15000, Logs: nil},
			{Status: 1, GasUsed: 27000, Logs: []*Log{}},
		},
	}

	enc, err := EncodeFrameTxReceiptRLP(r)
	if err != nil {
		t.Fatalf("EncodeFrameTxReceiptRLP: %v", err)
	}
	if len(enc) == 0 {
		t.Fatal("encoded receipt is empty")
	}

	got, err := DecodeFrameTxReceiptRLP(enc)
	if err != nil {
		t.Fatalf("DecodeFrameTxReceiptRLP: %v", err)
	}

	if got.CumulativeGasUsed != r.CumulativeGasUsed {
		t.Errorf("CumulativeGasUsed: got %d, want %d", got.CumulativeGasUsed, r.CumulativeGasUsed)
	}
	if got.Payer != r.Payer {
		t.Errorf("Payer: got %x, want %x", got.Payer, r.Payer)
	}
	if len(got.FrameResults) != len(r.FrameResults) {
		t.Fatalf("FrameResults count: got %d, want %d", len(got.FrameResults), len(r.FrameResults))
	}
	for i, fr := range r.FrameResults {
		gfr := got.FrameResults[i]
		if gfr.Status != fr.Status {
			t.Errorf("frame %d Status: got %d, want %d", i, gfr.Status, fr.Status)
		}
		if gfr.GasUsed != fr.GasUsed {
			t.Errorf("frame %d GasUsed: got %d, want %d", i, gfr.GasUsed, fr.GasUsed)
		}
		if len(gfr.Logs) != len(fr.Logs) {
			t.Errorf("frame %d Logs count: got %d, want %d", i, len(gfr.Logs), len(fr.Logs))
		}
	}
}

func TestFrameTxReceiptRLPEmpty(t *testing.T) {
	r := &FrameTxReceipt{
		CumulativeGasUsed: 0,
		Payer:             Address{},
		FrameResults:      nil,
	}
	enc, err := EncodeFrameTxReceiptRLP(r)
	if err != nil {
		t.Fatalf("encode empty: %v", err)
	}
	got, err := DecodeFrameTxReceiptRLP(enc)
	if err != nil {
		t.Fatalf("decode empty: %v", err)
	}
	if got.CumulativeGasUsed != 0 || len(got.FrameResults) != 0 {
		t.Errorf("empty round-trip failed: %+v", got)
	}
}

func TestFrameTxReceiptRLPTypePrefix(t *testing.T) {
	// Frame tx type is 0x06 — encoding should be prefixed with 0x06.
	r := &FrameTxReceipt{CumulativeGasUsed: 21000, Payer: Address{0x01}}
	enc, err := EncodeFrameTxReceiptRLP(r)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	// First byte should be the frame tx type prefix 0x06.
	if enc[0] != FrameTxType {
		t.Errorf("type prefix: got 0x%02x, want 0x%02x", enc[0], FrameTxType)
	}
}
