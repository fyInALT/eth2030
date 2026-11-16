package rpctypes

import (
	"testing"

	"github.com/eth2030/eth2030/core/rawdb"
	coretypes "github.com/eth2030/eth2030/core/types"
)

func TestFrame8141ReceiptReadBack(t *testing.T) {
	path := t.TempDir()
	open := func() *rawdb.FileDB {
		t.Helper()
		db, err := rawdb.NewFileDB(path)
		if err != nil {
			t.Fatal(err)
		}
		return db
	}
	db := open()
	payer := coretypes.Address{19: 9}
	log := &coretypes.Log{Address: coretypes.Address{19: 8}, Data: []byte{6}}
	receipt := &coretypes.Receipt{
		Type: coretypes.FrameTxType, Payer: payer, CumulativeGasUsed: 123456,
		FrameResults8141: []coretypes.FrameResult8141{
			{Status: 1, GasUsed: coretypes.FrameGasLimits{Execution: 300, State: 400}, Logs: []*coretypes.Log{log}},
			{Status: coretypes.FrameStatusSkipped},
		},
	}
	codec := new(coretypes.ReceiptCodec)
	encoded, err := codec.EncodeReceipts([]*coretypes.Receipt{receipt})
	if err != nil {
		t.Fatal(err)
	}
	hash := coretypes.Hash{31: 1}
	if err := rawdb.WriteReceipts(db, 42, hash, encoded); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db = open()
	defer db.Close()
	persisted, err := rawdb.ReadReceipts(db, 42, hash)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := codec.DecodeReceipts(persisted)
	if err != nil || len(actual) != 1 {
		t.Fatalf("decode persisted receipts: %v", err)
	}
	rpc := FormatReceipt(actual[0], nil, 95)
	if rpc.Payer == nil || *rpc.Payer != EncodeAddress(payer) || rpc.FrameResults == nil {
		t.Fatalf("missing payer or frame results: %+v", rpc)
	}
	frames := *rpc.FrameResults
	if len(frames) != 2 || frames[0].GasUsed != "0x12c" ||
		frames[0].StateGasUsed == nil || *frames[0].StateGasUsed != "0x190" ||
		frames[1].Status != "0x2" || len(frames[0].Logs) != 1 ||
		frames[0].Logs[0] != rpc.Logs[0] || frames[0].Logs[0].Data != "0x06" {
		t.Fatalf("wrong persisted frame results: %+v", frames)
	}
}
