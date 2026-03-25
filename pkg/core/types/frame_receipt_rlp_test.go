package types

import (
	"math/big"
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

// --- ToReceipt conversion tests ---

func TestFrameTxReceipt_DeriveStatus(t *testing.T) {
	tests := []struct {
		name     string
		frames   []FrameResult
		expected uint64
	}{
		{
			name:     "all frames successful",
			frames:   []FrameResult{{Status: 1}, {Status: 1}},
			expected: ReceiptStatusSuccessful,
		},
		{
			name:     "one frame failed",
			frames:   []FrameResult{{Status: 1}, {Status: 0}, {Status: 1}},
			expected: ReceiptStatusFailed,
		},
		{
			name:     "empty frames",
			frames:   []FrameResult{},
			expected: ReceiptStatusSuccessful,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &FrameTxReceipt{FrameResults: tt.frames}
			if got := r.DeriveStatus(); got != tt.expected {
				t.Errorf("DeriveStatus() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestFrameTxReceipt_ToReceipt(t *testing.T) {
	payer := Address{0xAB, 0xCD, 0xEF}
	log1 := &Log{Address: Address{0x01}, Topics: []Hash{{0xAA}}, Data: []byte{0x01}}
	log2 := &Log{Address: Address{0x02}, Topics: []Hash{{0xBB}}, Data: []byte{0x02}}

	r := &FrameTxReceipt{
		CumulativeGasUsed: 100000,
		Payer:             payer,
		FrameResults: []FrameResult{
			{Status: 1, GasUsed: 30000, Logs: []*Log{log1}},
			{Status: 1, GasUsed: 20000, Logs: []*Log{log2}},
		},
	}

	txHash := Hash{0x11}
	blockHash := Hash{0x22}
	blockNumber := big.NewInt(42)
	effectiveGasPrice := big.NewInt(30_000_000_000)

	receipt := r.ToReceipt(txHash, blockHash, blockNumber, 0, effectiveGasPrice)

	// Check basic fields.
	if receipt.Type != FrameTxType {
		t.Errorf("Type = %d, want %d", receipt.Type, FrameTxType)
	}
	if receipt.Status != ReceiptStatusSuccessful {
		t.Errorf("Status = %d, want %d", receipt.Status, ReceiptStatusSuccessful)
	}
	if receipt.CumulativeGasUsed != 100000 {
		t.Errorf("CumulativeGasUsed = %d, want 100000", receipt.CumulativeGasUsed)
	}
	if receipt.GasUsed != 50000 {
		t.Errorf("GasUsed = %d, want 50000", receipt.GasUsed)
	}
	if receipt.TxHash != txHash {
		t.Errorf("TxHash = %x, want %x", receipt.TxHash, txHash)
	}
	if receipt.BlockHash != blockHash {
		t.Errorf("BlockHash = %x, want %x", receipt.BlockHash, blockHash)
	}
	if receipt.BlockNumber.Cmp(blockNumber) != 0 {
		t.Errorf("BlockNumber = %s, want %s", receipt.BlockNumber, blockNumber)
	}
	if receipt.TransactionIndex != 0 {
		t.Errorf("TransactionIndex = %d, want 0", receipt.TransactionIndex)
	}
	if receipt.EffectiveGasPrice.Cmp(effectiveGasPrice) != 0 {
		t.Errorf("EffectiveGasPrice = %s, want %s", receipt.EffectiveGasPrice, effectiveGasPrice)
	}

	// Check logs aggregation.
	if len(receipt.Logs) != 2 {
		t.Fatalf("Logs count = %d, want 2", len(receipt.Logs))
	}
	if receipt.Logs[0] != log1 {
		t.Error("first log mismatch")
	}
	if receipt.Logs[1] != log2 {
		t.Error("second log mismatch")
	}

	// Check bloom filter is computed.
	if receipt.Bloom == (Bloom{}) {
		t.Error("Bloom should not be empty")
	}
}

func TestFrameTxReceipt_ToReceiptWithPayer(t *testing.T) {
	payer := Address{0xAB, 0xCD, 0xEF}

	r := &FrameTxReceipt{
		CumulativeGasUsed: 63000,
		Payer:             payer,
		FrameResults: []FrameResult{
			{Status: 1, GasUsed: 21000},
			{Status: 1, GasUsed: 42000},
		},
	}

	receipt := r.ToReceiptWithPayer(
		Hash{0x11}, Hash{0x22}, big.NewInt(1), 5, big.NewInt(1_000_000_000),
	)

	// ContractAddress should be repurposed for payer.
	if receipt.ContractAddress != payer {
		t.Errorf("ContractAddress (payer) = %x, want %x", receipt.ContractAddress, payer)
	}
	if receipt.TransactionIndex != 5 {
		t.Errorf("TransactionIndex = %d, want 5", receipt.TransactionIndex)
	}
}

func TestFrameTxReceipt_ToReceipt_FailedFrame(t *testing.T) {
	r := &FrameTxReceipt{
		CumulativeGasUsed: 50000,
		Payer:             Address{0x01},
		FrameResults: []FrameResult{
			{Status: 1, GasUsed: 20000},
			{Status: 0, GasUsed: 30000}, // Failed frame
		},
	}

	receipt := r.ToReceipt(Hash{}, Hash{}, big.NewInt(1), 0, nil)

	if receipt.Status != ReceiptStatusFailed {
		t.Errorf("Status = %d, want %d (failed frame)", receipt.Status, ReceiptStatusFailed)
	}
}

func TestFrameTxReceipt_ComputeBloom(t *testing.T) {
	log1 := &Log{Address: Address{0x01}, Topics: []Hash{{0xAA}}, Data: []byte{0x01}}
	log2 := &Log{Address: Address{0x02}, Topics: []Hash{{0xBB}}, Data: []byte{0x02}}

	r := &FrameTxReceipt{
		FrameResults: []FrameResult{
			{Logs: []*Log{log1}},
			{Logs: []*Log{log2}},
		},
	}

	bloom := r.ComputeBloom()
	if bloom == (Bloom{}) {
		t.Error("ComputeBloom returned empty bloom")
	}

	// Verify the bloom matches what LogsBloom would produce.
	expectedBloom := LogsBloom([]*Log{log1, log2})
	if bloom != expectedBloom {
		t.Error("ComputeBloom should match LogsBloom of all logs")
	}
}
