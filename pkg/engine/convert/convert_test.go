package convert

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
	engblobs "github.com/eth2030/eth2030/engine/blobsbundle"
	engpayload "github.com/eth2030/eth2030/engine/payload"
)

func makeV1Payload() *engpayload.ExecutionPayloadV1 {
	return &engpayload.ExecutionPayloadV1{
		ParentHash:    types.HexToHash("0x1111"),
		FeeRecipient:  types.HexToAddress("0x2222"),
		StateRoot:     types.HexToHash("0x3333"),
		ReceiptsRoot:  types.HexToHash("0x4444"),
		PrevRandao:    types.HexToHash("0x5555"),
		BlockNumber:   100,
		GasLimit:      30_000_000,
		GasUsed:       21_000,
		Timestamp:     1_700_000_000,
		ExtraData:     []byte("test"),
		BaseFeePerGas: big.NewInt(1_000_000_000),
	}
}

func TestPayloadToHeaderV1(t *testing.T) {
	p := makeV1Payload()
	h := PayloadToHeaderV1(p)

	if h.ParentHash != p.ParentHash {
		t.Error("ParentHash mismatch")
	}
	if h.Coinbase != p.FeeRecipient {
		t.Error("Coinbase mismatch")
	}
	if h.Root != p.StateRoot {
		t.Error("StateRoot mismatch")
	}
	if h.ReceiptHash != p.ReceiptsRoot {
		t.Error("ReceiptsRoot mismatch")
	}
	if h.MixDigest != p.PrevRandao {
		t.Error("PrevRandao mismatch")
	}
	if h.Number.Uint64() != p.BlockNumber {
		t.Errorf("BlockNumber = %d, want %d", h.Number.Uint64(), p.BlockNumber)
	}
	if h.GasLimit != p.GasLimit {
		t.Errorf("GasLimit = %d, want %d", h.GasLimit, p.GasLimit)
	}
	if h.GasUsed != p.GasUsed {
		t.Errorf("GasUsed = %d, want %d", h.GasUsed, p.GasUsed)
	}
	if h.Time != p.Timestamp {
		t.Errorf("Time = %d, want %d", h.Time, p.Timestamp)
	}
	if h.Difficulty.Sign() != 0 {
		t.Error("Difficulty should be 0")
	}
	if h.UncleHash != types.EmptyUncleHash {
		t.Error("UncleHash should be EmptyUncleHash")
	}
}

func TestPayloadToHeaderV2(t *testing.T) {
	v1 := makeV1Payload()
	p := &engpayload.ExecutionPayloadV2{ExecutionPayloadV1: *v1}
	h := PayloadToHeaderV2(p)
	if h.Number.Uint64() != v1.BlockNumber {
		t.Errorf("BlockNumber = %d, want %d", h.Number.Uint64(), v1.BlockNumber)
	}
}

func TestPayloadToHeaderV3(t *testing.T) {
	v1 := makeV1Payload()
	blobGasUsed := uint64(131072)
	excessBlobGas := uint64(65536)
	p := &engpayload.ExecutionPayloadV3{
		ExecutionPayloadV2: engpayload.ExecutionPayloadV2{ExecutionPayloadV1: *v1},
		BlobGasUsed:        blobGasUsed,
		ExcessBlobGas:      excessBlobGas,
	}
	h := PayloadToHeaderV3(p)
	if h.BlobGasUsed == nil || *h.BlobGasUsed != blobGasUsed {
		t.Errorf("BlobGasUsed = %v, want %d", h.BlobGasUsed, blobGasUsed)
	}
	if h.ExcessBlobGas == nil || *h.ExcessBlobGas != excessBlobGas {
		t.Errorf("ExcessBlobGas = %v, want %d", h.ExcessBlobGas, excessBlobGas)
	}
}

func TestPayloadToHeaderV5(t *testing.T) {
	v1 := makeV1Payload()
	p := &engpayload.ExecutionPayloadV5{
		ExecutionPayloadV4: engpayload.ExecutionPayloadV4{
			ExecutionPayloadV3: engpayload.ExecutionPayloadV3{
				ExecutionPayloadV2: engpayload.ExecutionPayloadV2{ExecutionPayloadV1: *v1},
				BlobGasUsed:        256,
				ExcessBlobGas:      512,
			},
		},
	}
	h := PayloadToHeaderV5(p)
	if h.Number.Uint64() != v1.BlockNumber {
		t.Errorf("BlockNumber = %d, want %d", h.Number.Uint64(), v1.BlockNumber)
	}
	if h.BlobGasUsed == nil || *h.BlobGasUsed != 256 {
		t.Errorf("BlobGasUsed = %v, want 256", h.BlobGasUsed)
	}
}

func TestHeaderToPayloadV2(t *testing.T) {
	header := &types.Header{
		ParentHash: types.HexToHash("0xabcd"),
		Coinbase:   types.HexToAddress("0xdead"),
		Root:       types.HexToHash("0xbeef"),
		Number:     big.NewInt(99),
		GasLimit:   15_000_000,
		GasUsed:    5_000_000,
		Time:       1_600_000_000,
		BaseFee:    big.NewInt(500_000_000),
	}
	ws := []*engpayload.Withdrawal{
		{Index: 1, ValidatorIndex: 10, Address: types.HexToAddress("0xaaaa"), Amount: 1000},
	}
	v2 := HeaderToPayloadV2(header, ws)
	if v2.BlockNumber != 99 {
		t.Errorf("BlockNumber = %d, want 99", v2.BlockNumber)
	}
	if len(v2.Withdrawals) != 1 {
		t.Errorf("Withdrawals count = %d, want 1", len(v2.Withdrawals))
	}
	if v2.Withdrawals[0].Amount != 1000 {
		t.Errorf("Withdrawal amount = %d, want 1000", v2.Withdrawals[0].Amount)
	}
}

func TestHeaderToPayloadV3(t *testing.T) {
	blobGasUsed := uint64(100)
	excessBlobGas := uint64(200)
	header := &types.Header{
		Number:        big.NewInt(10),
		GasLimit:      10_000_000,
		GasUsed:       1_000_000,
		Time:          1_700_000_000,
		BaseFee:       big.NewInt(1_000_000_000),
		BlobGasUsed:   &blobGasUsed,
		ExcessBlobGas: &excessBlobGas,
	}
	v3 := HeaderToPayloadV3(header, nil)
	if v3.BlobGasUsed != blobGasUsed {
		t.Errorf("BlobGasUsed = %d, want %d", v3.BlobGasUsed, blobGasUsed)
	}
	if v3.ExcessBlobGas != excessBlobGas {
		t.Errorf("ExcessBlobGas = %d, want %d", v3.ExcessBlobGas, excessBlobGas)
	}
}

func TestHeaderToPayloadV3_NilBlobFields(t *testing.T) {
	header := &types.Header{
		Number:   big.NewInt(1),
		GasLimit: 10_000_000,
		GasUsed:  0,
		Time:     1_700_000_000,
		BaseFee:  big.NewInt(1),
	}
	v3 := HeaderToPayloadV3(header, nil)
	if v3.BlobGasUsed != 0 {
		t.Errorf("BlobGasUsed = %d, want 0", v3.BlobGasUsed)
	}
	if v3.ExcessBlobGas != 0 {
		t.Errorf("ExcessBlobGas = %d, want 0", v3.ExcessBlobGas)
	}
}

func TestDeterminePayloadVersion(t *testing.T) {
	tests := []struct {
		name    string
		ts      uint64
		forks   *ForkTimestamps
		wantVer PayloadVersion
	}{
		{
			name:    "nil forks -> V1",
			ts:      1_000,
			forks:   nil,
			wantVer: PayloadV1,
		},
		{
			name:    "no forks activated -> V1",
			ts:      500,
			forks:   &ForkTimestamps{Shanghai: 1000},
			wantVer: PayloadV1,
		},
		{
			name:    "Shanghai -> V2",
			ts:      1000,
			forks:   &ForkTimestamps{Shanghai: 1000},
			wantVer: PayloadV2,
		},
		{
			name:    "Cancun -> V3",
			ts:      2000,
			forks:   &ForkTimestamps{Shanghai: 1000, Cancun: 2000},
			wantVer: PayloadV3,
		},
		{
			name:    "Prague -> V4",
			ts:      3000,
			forks:   &ForkTimestamps{Shanghai: 1000, Cancun: 2000, Prague: 3000},
			wantVer: PayloadV4,
		},
		{
			name:    "Amsterdam -> V5",
			ts:      4000,
			forks:   &ForkTimestamps{Shanghai: 1000, Cancun: 2000, Prague: 3000, Amsterdam: 4000},
			wantVer: PayloadV5,
		},
		{
			name:    "between Cancun and Prague -> V3",
			ts:      2500,
			forks:   &ForkTimestamps{Shanghai: 1000, Cancun: 2000, Prague: 3000, Amsterdam: 4000},
			wantVer: PayloadV3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DeterminePayloadVersion(tc.ts, tc.forks)
			if got != tc.wantVer {
				t.Errorf("DeterminePayloadVersion(%d, ...) = %d, want %d", tc.ts, got, tc.wantVer)
			}
		})
	}
}

func TestConvertV1ToV2(t *testing.T) {
	v1 := makeV1Payload()
	v2 := ConvertV1ToV2(v1)
	if v2.BlockNumber != v1.BlockNumber {
		t.Errorf("BlockNumber = %d, want %d", v2.BlockNumber, v1.BlockNumber)
	}
	if v2.Withdrawals == nil {
		t.Error("Withdrawals should be non-nil empty slice")
	}
	if len(v2.Withdrawals) != 0 {
		t.Errorf("Withdrawals length = %d, want 0", len(v2.Withdrawals))
	}
}

func TestConvertV2ToV3(t *testing.T) {
	v1 := makeV1Payload()
	v2 := ConvertV1ToV2(v1)
	v3 := ConvertV2ToV3(v2)
	if v3.BlobGasUsed != 0 {
		t.Errorf("BlobGasUsed = %d, want 0", v3.BlobGasUsed)
	}
	if v3.ExcessBlobGas != 0 {
		t.Errorf("ExcessBlobGas = %d, want 0", v3.ExcessBlobGas)
	}
	if v3.BlockNumber != v1.BlockNumber {
		t.Errorf("BlockNumber = %d, want %d", v3.BlockNumber, v1.BlockNumber)
	}
}

func TestConvertV3ToV4(t *testing.T) {
	v1 := makeV1Payload()
	v3 := ConvertV2ToV3(ConvertV1ToV2(v1))
	v4 := ConvertV3ToV4(v3)
	if v4.ExecutionRequests == nil {
		t.Error("ExecutionRequests should be non-nil empty slice")
	}
	if len(v4.ExecutionRequests) != 0 {
		t.Errorf("ExecutionRequests length = %d, want 0", len(v4.ExecutionRequests))
	}
}

func TestConvertV4ToV5(t *testing.T) {
	v1 := makeV1Payload()
	v4 := ConvertV3ToV4(ConvertV2ToV3(ConvertV1ToV2(v1)))
	v5 := ConvertV4ToV5(v4)
	if v5.BlockAccessList != nil {
		t.Errorf("BlockAccessList should be nil, got %v", v5.BlockAccessList)
	}
	if v5.BlockNumber != v1.BlockNumber {
		t.Errorf("BlockNumber = %d, want %d", v5.BlockNumber, v1.BlockNumber)
	}
}

func TestSummarizeWithdrawals(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		s := SummarizeWithdrawals(nil)
		if s.Count != 0 || s.TotalAmountGwei != 0 || s.UniqueValidators != 0 || s.UniqueAddresses != 0 {
			t.Errorf("expected all zeros, got %+v", s)
		}
	})

	t.Run("single withdrawal", func(t *testing.T) {
		ws := []*engpayload.Withdrawal{
			{ValidatorIndex: 1, Address: types.HexToAddress("0x1111"), Amount: 500},
		}
		s := SummarizeWithdrawals(ws)
		if s.Count != 1 {
			t.Errorf("Count = %d, want 1", s.Count)
		}
		if s.TotalAmountGwei != 500 {
			t.Errorf("TotalAmountGwei = %d, want 500", s.TotalAmountGwei)
		}
		if s.UniqueValidators != 1 {
			t.Errorf("UniqueValidators = %d, want 1", s.UniqueValidators)
		}
		if s.UniqueAddresses != 1 {
			t.Errorf("UniqueAddresses = %d, want 1", s.UniqueAddresses)
		}
	})

	t.Run("multiple with duplicates", func(t *testing.T) {
		addr := types.HexToAddress("0xaaaa")
		ws := []*engpayload.Withdrawal{
			{ValidatorIndex: 10, Address: addr, Amount: 1000},
			{ValidatorIndex: 10, Address: addr, Amount: 2000},
			{ValidatorIndex: 20, Address: types.HexToAddress("0xbbbb"), Amount: 3000},
		}
		s := SummarizeWithdrawals(ws)
		if s.Count != 3 {
			t.Errorf("Count = %d, want 3", s.Count)
		}
		if s.TotalAmountGwei != 6000 {
			t.Errorf("TotalAmountGwei = %d, want 6000", s.TotalAmountGwei)
		}
		if s.UniqueValidators != 2 {
			t.Errorf("UniqueValidators = %d, want 2", s.UniqueValidators)
		}
		if s.UniqueAddresses != 2 {
			t.Errorf("UniqueAddresses = %d, want 2", s.UniqueAddresses)
		}
	})
}

func TestProcessWithdrawalsExt(t *testing.T) {
	ws := []*engpayload.Withdrawal{
		{ValidatorIndex: 1, Amount: 100},
		{ValidatorIndex: 2, Amount: 200},
		{ValidatorIndex: 1, Amount: 50},
	}
	total, byValidator := ProcessWithdrawalsExt(ws)
	if total != 350 {
		t.Errorf("total = %d, want 350", total)
	}
	if byValidator[1] != 150 {
		t.Errorf("validator 1 = %d, want 150", byValidator[1])
	}
	if byValidator[2] != 200 {
		t.Errorf("validator 2 = %d, want 200", byValidator[2])
	}
}

func TestCoreWithdrawalsFromPayload(t *testing.T) {
	t.Run("nil payload", func(t *testing.T) {
		result := CoreWithdrawalsFromPayload(nil)
		if result != nil {
			t.Error("expected nil for nil payload")
		}
	})

	t.Run("payload with withdrawals", func(t *testing.T) {
		p := &engpayload.ExecutionPayloadV2{
			Withdrawals: []*engpayload.Withdrawal{
				{Index: 1, ValidatorIndex: 10, Address: types.HexToAddress("0xaaaa"), Amount: 500},
			},
		}
		result := CoreWithdrawalsFromPayload(p)
		if len(result) != 1 {
			t.Fatalf("expected 1 withdrawal, got %d", len(result))
		}
		if result[0].Index != 1 || result[0].Amount != 500 {
			t.Errorf("withdrawal mismatch: %+v", result[0])
		}
	})
}

func TestVersionedHashFromCommitment(t *testing.T) {
	commitment := make([]byte, 48)
	for i := range commitment {
		commitment[i] = byte(i)
	}
	h := VersionedHashFromCommitment(commitment)
	if h[0] != engblobs.VersionedHashVersion {
		t.Errorf("first byte = 0x%02x, want 0x%02x", h[0], engblobs.VersionedHashVersion)
	}
	// Deterministic: same input gives same output.
	h2 := VersionedHashFromCommitment(commitment)
	if h != h2 {
		t.Error("not deterministic")
	}
}

func TestBlobSidecarFromBundle(t *testing.T) {
	blob := make([]byte, engblobs.BlobSize)
	commitment := make([]byte, engblobs.KZGCommitmentSize)
	proof := make([]byte, engblobs.KZGProofSize)
	blockHash := types.HexToHash("0xdeadbeef")

	bundle := &engpayload.BlobsBundleV1{
		Blobs:       [][]byte{blob},
		Commitments: [][]byte{commitment},
		Proofs:      [][]byte{proof},
	}

	t.Run("valid index 0", func(t *testing.T) {
		sc, err := BlobSidecarFromBundle(bundle, 0, blockHash)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sc.Index != 0 {
			t.Errorf("Index = %d, want 0", sc.Index)
		}
		if sc.SignedBlockHeader != blockHash {
			t.Error("SignedBlockHeader mismatch")
		}
	})

	t.Run("nil bundle", func(t *testing.T) {
		_, err := BlobSidecarFromBundle(nil, 0, blockHash)
		if err != engblobs.ErrBlobBundleEmpty {
			t.Errorf("expected ErrBlobBundleEmpty, got %v", err)
		}
	})

	t.Run("out of range index", func(t *testing.T) {
		_, err := BlobSidecarFromBundle(bundle, 5, blockHash)
		if err != engblobs.ErrBlobBundleSidecarIndex {
			t.Errorf("expected ErrBlobBundleSidecarIndex, got %v", err)
		}
	})

	t.Run("negative index", func(t *testing.T) {
		_, err := BlobSidecarFromBundle(bundle, -1, blockHash)
		if err != engblobs.ErrBlobBundleSidecarIndex {
			t.Errorf("expected ErrBlobBundleSidecarIndex, got %v", err)
		}
	})
}
