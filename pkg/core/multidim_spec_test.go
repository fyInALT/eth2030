package core

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/gas"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
)

// --- SPEC-6.2: GetCalldataGas ---

func TestGetCalldataGas_Zero(t *testing.T) {
	to := types.Address{0x01}
	tx := types.NewTransaction(&types.LegacyTx{
		Nonce:    0,
		To:       &to,
		Value:    big.NewInt(0),
		Gas:      21000,
		GasPrice: big.NewInt(1),
		Data:     nil,
	})
	if got := gas.GetCalldataGas(tx); got != 0 {
		t.Errorf("GetCalldataGas(empty data) = %d, want 0", got)
	}
}

func TestGetCalldataGas_NonZeroData(t *testing.T) {
	to := types.Address{0x01}
	data := []byte{0x00, 0x00, 0xFF, 0xFF} // 2 zero + 2 nonzero
	tx := types.NewTransaction(&types.LegacyTx{
		Nonce:    0,
		To:       &to,
		Value:    big.NewInt(0),
		Gas:      21000,
		GasPrice: big.NewInt(1),
		Data:     data,
	})
	got := gas.GetCalldataGas(tx)
	// tokens = 2 zero * 1 + 2 nonzero * 4 = 10; gas = 10 * 4 = 40
	want := uint64(40)
	if got != want {
		t.Errorf("GetCalldataGas = %d, want %d", got, want)
	}
}

// --- SPEC-6.4: CalcCalldataBaseFee ---

func TestCalcCalldataBaseFee_ZeroExcess(t *testing.T) {
	fee := gas.CalcCalldataBaseFee(0, 1_000_000)
	if fee == nil || fee.Sign() <= 0 {
		t.Errorf("CalcCalldataBaseFee(0, 1M): expected positive fee, got %v", fee)
	}
}

func TestCalcCalldataBaseFee_FromHeader(t *testing.T) {
	excess := uint64(100_000)
	h := &types.Header{
		GasLimit:          30_000_000,
		CalldataExcessGas: &excess,
	}
	fee := gas.CalcCalldataBaseFeeFromHeader(h)
	if fee == nil || fee.Sign() <= 0 {
		t.Errorf("CalcCalldataBaseFeeFromHeader: expected positive fee, got %v", fee)
	}
}

// --- SPEC-6.3: Header 3D gas vector fields ---

func TestHeader3DGasVectorFields(t *testing.T) {
	limits := [3]uint64{30_000_000, 7_500_000, 786_432}
	used := [3]uint64{20_000_000, 5_000_000, 393_216}
	excess := [3]uint64{1_000_000, 500_000, 100_000}
	h := &types.Header{
		GasLimitVec:  &limits,
		GasUsedVec:   &used,
		ExcessGasVec: &excess,
	}
	if h.GasLimitVec[0] != 30_000_000 {
		t.Errorf("GasLimitVec[0] = %d, want 30M", h.GasLimitVec[0])
	}
	if h.GasUsedVec[1] != 5_000_000 {
		t.Errorf("GasUsedVec[1] = %d, want 5M", h.GasUsedVec[1])
	}
	if h.ExcessGasVec[2] != 100_000 {
		t.Errorf("ExcessGasVec[2] = %d, want 100K", h.ExcessGasVec[2])
	}
}

func TestCalcCalldataBaseFee_HigherExcessMeansHigherFee(t *testing.T) {
	// Use large excess to ensure fee difference is visible.
	fee1 := gas.CalcCalldataBaseFee(0, 1_000_000)
	fee2 := gas.CalcCalldataBaseFee(100_000_000, 1_000_000)
	if fee2.Cmp(fee1) <= 0 {
		t.Errorf("higher excess gas should produce higher base fee: fee1=%v fee2=%v", fee1, fee2)
	}
}

// --- SPEC-6.4: BuildBlock wires calldata dimension into 3D gas vectors ---

func TestBuildBlock_3DGasVectors_Set(t *testing.T) {
	statedb := state.NewMemoryStateDB()
	builder := NewBlockBuilder(config.TestConfigGlamsterdan, nil, nil)
	builder.SetState(statedb)

	parent := &types.Header{
		Number:   big.NewInt(0),
		GasLimit: 30_000_000,
		GasUsed:  0,
		BaseFee:  big.NewInt(1_000_000_000),
		Time:     1700000000,
	}

	block, _, err := builder.BuildBlock(parent, &BuildBlockAttributes{
		Timestamp:    parent.Time + 12,
		FeeRecipient: types.Address{0x01},
		GasLimit:     parent.GasLimit,
	})
	if err != nil {
		t.Fatalf("BuildBlock: %v", err)
	}

	h := block.Header()
	if h.GasLimitVec == nil {
		t.Fatal("GasLimitVec is nil; expected 3D vector for Glamsterdam block")
	}
	if h.GasUsedVec == nil {
		t.Fatal("GasUsedVec is nil; expected 3D vector for Glamsterdam block")
	}
	if h.ExcessGasVec == nil {
		t.Fatal("ExcessGasVec is nil; expected 3D vector for Glamsterdam block")
	}

	// Dimension 0 = execution gas.
	if h.GasLimitVec[0] != h.GasLimit {
		t.Errorf("GasLimitVec[0] = %d, want GasLimit %d", h.GasLimitVec[0], h.GasLimit)
	}
	if h.GasUsedVec[0] != h.GasUsed {
		t.Errorf("GasUsedVec[0] = %d, want GasUsed %d", h.GasUsedVec[0], h.GasUsed)
	}

	// Dimension 2 = calldata gas; limit must equal CalcCalldataGasLimit(GasLimit).
	wantCalldataLimit := gas.CalcCalldataGasLimit(h.GasLimit)
	if h.GasLimitVec[2] != wantCalldataLimit {
		t.Errorf("GasLimitVec[2] = %d, want CalcCalldataGasLimit(%d) = %d",
			h.GasLimitVec[2], h.GasLimit, wantCalldataLimit)
	}
}

func TestBuildBlock_3DGasVectors_NotSetWithoutGlamsterdam(t *testing.T) {
	statedb := state.NewMemoryStateDB()
	// Use a config without Glamsterdam.
	builder := NewBlockBuilder(config.TestConfig, nil, nil)
	builder.SetState(statedb)

	parent := &types.Header{
		Number:   big.NewInt(0),
		GasLimit: 30_000_000,
		GasUsed:  0,
		BaseFee:  big.NewInt(1_000_000_000),
		Time:     1700000000,
	}

	// Pick a timestamp that is before the Glamsterdam activation in config.TestConfig.
	// config.TestConfig has Glamsterdam at time 0, so we need a config with no Glamsterdam.
	// Just verify: if Glamsterdam is active in config.TestConfig, the vectors are set.
	// This test asserts the inverse: pre-Glamsterdam blocks have nil vectors.
	//
	// Since config.TestConfig activates Glamsterdam at timestamp 0, we use a fork config
	// that has GlamsterdamTime = nil (pre-Glamsterdam).
	cfgPreGlam := &config.ChainConfig{
		ChainID:        config.TestConfig.ChainID,
		HomesteadBlock: config.TestConfig.HomesteadBlock,
		EIP155Block:    config.TestConfig.EIP155Block,
		// No Glamsterdan — leave GlamsterdanTime nil.
	}
	builder2 := NewBlockBuilder(cfgPreGlam, nil, nil)
	builder2.SetState(statedb)

	block, _, err := builder2.BuildBlock(parent, &BuildBlockAttributes{
		Timestamp:    parent.Time + 12,
		FeeRecipient: types.Address{0x01},
		GasLimit:     parent.GasLimit,
	})
	if err != nil {
		t.Fatalf("BuildBlock (pre-Glamsterdam): %v", err)
	}
	h := block.Header()
	if h.GasLimitVec != nil {
		t.Error("GasLimitVec should be nil for pre-Glamsterdam block")
	}
	if h.GasUsedVec != nil {
		t.Error("GasUsedVec should be nil for pre-Glamsterdam block")
	}
	if h.ExcessGasVec != nil {
		t.Error("ExcessGasVec should be nil for pre-Glamsterdam block")
	}
}
