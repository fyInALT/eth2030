package gas

import (
	"github.com/eth2030/eth2030/core/config"
	"testing"
)

func TestValidateTransactionGasLimit(t *testing.T) {
	tests := []struct {
		name     string
		gasLimit uint64
		wantErr  bool
	}{
		{"zero gas", 0, false},
		{"typical transaction", 21000, false},
		{"contract deployment", 5_000_000, false},
		{"just below cap", MaxTransactionGas - 1, false},
		{"exactly at cap", MaxTransactionGas, false},
		{"one over cap", MaxTransactionGas + 1, true},
		{"block gas limit", 30_000_000, true},
		{"very large", 1 << 32, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransactionGasLimit(tt.gasLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransactionGasLimit(%d) error = %v, wantErr %v",
					tt.gasLimit, err, tt.wantErr)
			}
			if err != nil && err != ErrTxGasLimitExceeded {
				t.Errorf("expected ErrTxGasLimitExceeded, got %v", err)
			}
		})
	}
}

func TestMaxTransactionGasValue(t *testing.T) {
	// Verify that MaxTransactionGas is exactly 2^24.
	if MaxTransactionGas != 16_777_216 {
		t.Errorf("MaxTransactionGas = %d, want 16777216", MaxTransactionGas)
	}
}

func TestIsGasLimitCapped(t *testing.T) {
	praguetime := uint64(1000)
	cfg := &config.ChainConfig{
		PragueTime: &praguetime,
	}

	// Before Prague: not capped.
	if IsGasLimitCapped(cfg, 999) {
		t.Error("expected gas limit not capped before Prague")
	}

	// At Prague: capped.
	if !IsGasLimitCapped(cfg, 1000) {
		t.Error("expected gas limit capped at Prague activation")
	}

	// After Prague: capped.
	if !IsGasLimitCapped(cfg, 2000) {
		t.Error("expected gas limit capped after Prague activation")
	}

	// No Prague: not capped.
	noPrague := &config.ChainConfig{}
	if IsGasLimitCapped(noPrague, 5000) {
		t.Error("expected gas limit not capped without Prague config")
	}
}
