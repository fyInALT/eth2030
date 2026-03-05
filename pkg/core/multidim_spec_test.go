package core

import (
	"math/big"
	"testing"

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
	if got := GetCalldataGas(tx); got != 0 {
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
	got := GetCalldataGas(tx)
	// tokens = 2 zero * 1 + 2 nonzero * 4 = 10; gas = 10 * 4 = 40
	want := uint64(40)
	if got != want {
		t.Errorf("GetCalldataGas = %d, want %d", got, want)
	}
}

// --- SPEC-6.4: CalcCalldataBaseFee ---

func TestCalcCalldataBaseFee_ZeroExcess(t *testing.T) {
	fee := CalcCalldataBaseFee(0, 1_000_000)
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
	fee := CalcCalldataBaseFeeFromHeader(h)
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
	fee1 := CalcCalldataBaseFee(0, 1_000_000)
	fee2 := CalcCalldataBaseFee(100_000_000, 1_000_000)
	if fee2.Cmp(fee1) <= 0 {
		t.Errorf("higher excess gas should produce higher base fee: fee1=%v fee2=%v", fee1, fee2)
	}
}
