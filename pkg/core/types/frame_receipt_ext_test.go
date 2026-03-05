package types

import (
	"bytes"
	"testing"
)

// --- Additional FrameTxReceipt RLP coverage ---

func TestFrameTxReceiptRLP_LogsWithTopics(t *testing.T) {
	payer := Address{0x11}
	r := &FrameTxReceipt{
		CumulativeGasUsed: 42000,
		Payer:             payer,
		FrameResults: []FrameResult{
			{
				Status:  1,
				GasUsed: 42000,
				Logs: []*Log{
					{
						Address: Address{0xAA},
						Topics:  []Hash{{0x01}, {0x02}, {0x03}},
						Data:    []byte("event data"),
					},
				},
			},
		},
	}

	enc, err := EncodeFrameTxReceiptRLP(r)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := DecodeFrameTxReceiptRLP(enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got.FrameResults[0].Logs) != 1 {
		t.Fatal("log count mismatch")
	}
	if len(got.FrameResults[0].Logs[0].Topics) != 3 {
		t.Errorf("topics: got %d, want 3", len(got.FrameResults[0].Logs[0].Topics))
	}
	if !bytes.Equal(got.FrameResults[0].Logs[0].Data, []byte("event data")) {
		t.Errorf("log data mismatch")
	}
}

func TestFrameTxReceiptRLP_ManyFrames(t *testing.T) {
	r := &FrameTxReceipt{CumulativeGasUsed: 1_000_000}
	for i := 0; i < 50; i++ {
		r.FrameResults = append(r.FrameResults, FrameResult{
			Status:  uint64(i % 2),
			GasUsed: uint64(21000 * (i + 1)),
		})
	}
	enc, err := EncodeFrameTxReceiptRLP(r)
	if err != nil {
		t.Fatalf("encode 50 frames: %v", err)
	}
	got, err := DecodeFrameTxReceiptRLP(enc)
	if err != nil {
		t.Fatalf("decode 50 frames: %v", err)
	}
	if len(got.FrameResults) != 50 {
		t.Errorf("frame count: got %d, want 50", len(got.FrameResults))
	}
	for i, fr := range got.FrameResults {
		if fr.Status != uint64(i%2) {
			t.Errorf("frame %d status: got %d", i, fr.Status)
		}
	}
}

func TestFrameTxReceiptRLP_WrongTypeByte(t *testing.T) {
	// Start with a valid encoding then corrupt the type byte.
	r := &FrameTxReceipt{CumulativeGasUsed: 1, Payer: Address{0x01}}
	enc, _ := EncodeFrameTxReceiptRLP(r)
	enc[0] = 0x02 // wrong type
	if _, err := DecodeFrameTxReceiptRLP(enc); err == nil {
		t.Error("expected error for wrong type prefix, got nil")
	}
}

func TestFrameTxReceiptRLP_TooShort(t *testing.T) {
	if _, err := DecodeFrameTxReceiptRLP([]byte{0x06}); err == nil {
		t.Error("expected error for too-short input, got nil")
	}
	if _, err := DecodeFrameTxReceiptRLP(nil); err == nil {
		t.Error("expected error for nil input, got nil")
	}
}

func TestFrameTxReceiptRLP_FrameGasSum(t *testing.T) {
	r := &FrameTxReceipt{
		CumulativeGasUsed: 63000,
		FrameResults: []FrameResult{
			{Status: 1, GasUsed: 21000},
			{Status: 1, GasUsed: 21000},
			{Status: 1, GasUsed: 21000},
		},
	}
	if r.TotalGasUsed() != 63000 {
		t.Errorf("TotalGasUsed = %d, want 63000", r.TotalGasUsed())
	}
	enc, _ := EncodeFrameTxReceiptRLP(r)
	got, _ := DecodeFrameTxReceiptRLP(enc)
	if got.TotalGasUsed() != 63000 {
		t.Errorf("round-trip TotalGasUsed = %d, want 63000", got.TotalGasUsed())
	}
}

func TestFrameTxReceiptRLP_AllLogsAggregation(t *testing.T) {
	r := &FrameTxReceipt{
		FrameResults: []FrameResult{
			{Logs: []*Log{{Address: Address{0x01}}, {Address: Address{0x02}}}},
			{Logs: nil},
			{Logs: []*Log{{Address: Address{0x03}}}},
		},
	}
	logs := r.AllLogs()
	if len(logs) != 3 {
		t.Errorf("AllLogs count = %d, want 3", len(logs))
	}
}
