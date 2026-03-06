package compliance

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func mevTestTx(to types.Address) *types.Transaction {
	toAddr := to
	return types.NewTransaction(&types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1000),
		Gas:      21000,
		To:       &toAddr,
		Value:    big.NewInt(0),
		V:        big.NewInt(27),
		R:        big.NewInt(1),
		S:        big.NewInt(1),
	})
}

func mevTestTxNoTo() *types.Transaction {
	return types.NewTransaction(&types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1000),
		Gas:      21000,
		To:       nil,
		Value:    big.NewInt(0),
		V:        big.NewInt(27),
		R:        big.NewInt(1),
		S:        big.NewInt(1),
	})
}

// --- IsMEVTransaction tests ---

func TestIsMEVTransaction(t *testing.T) {
	dex := types.HexToAddress("0x1111111111111111111111111111111111111111")
	liq := types.HexToAddress("0x2222222222222222222222222222222222222222")
	other := types.HexToAddress("0x3333333333333333333333333333333333333333")

	config := DefaultMEVFilterConfig()
	config.KnownDEXContracts[dex] = true
	config.KnownLiquidationContracts[liq] = true
	filter := NewMEVFilter(config)

	t.Run("known DEX returns true", func(t *testing.T) {
		tx := mevTestTx(dex)
		if !filter.IsMEVTransaction(tx) {
			t.Error("expected true for DEX tx")
		}
	})

	t.Run("known liquidation returns true", func(t *testing.T) {
		tx := mevTestTx(liq)
		if !filter.IsMEVTransaction(tx) {
			t.Error("expected true for liquidation tx")
		}
	})

	t.Run("unknown address returns false", func(t *testing.T) {
		tx := mevTestTx(other)
		if filter.IsMEVTransaction(tx) {
			t.Error("expected false for unknown address")
		}
	})

	t.Run("nil tx returns false", func(t *testing.T) {
		if filter.IsMEVTransaction(nil) {
			t.Error("expected false for nil tx")
		}
	})

	t.Run("nil to returns false", func(t *testing.T) {
		tx := mevTestTxNoTo()
		if filter.IsMEVTransaction(tx) {
			t.Error("expected false for contract creation")
		}
	})
}

// --- FilterMEVOnly tests ---

func TestFilterMEVOnly(t *testing.T) {
	dex := types.HexToAddress("0x1111111111111111111111111111111111111111")
	other := types.HexToAddress("0x3333333333333333333333333333333333333333")

	config := DefaultMEVFilterConfig()
	config.KnownDEXContracts[dex] = true
	filter := NewMEVFilter(config)

	txs := []*types.Transaction{
		mevTestTx(dex),
		mevTestTx(other),
		mevTestTx(dex),
	}

	result := filter.FilterMEVOnly(txs)
	if len(result) != 2 {
		t.Errorf("FilterMEVOnly = %d, want 2", len(result))
	}
}

// --- FilterNonMEV tests ---

func TestFilterNonMEV(t *testing.T) {
	dex := types.HexToAddress("0x1111111111111111111111111111111111111111")
	other := types.HexToAddress("0x3333333333333333333333333333333333333333")

	config := DefaultMEVFilterConfig()
	config.KnownDEXContracts[dex] = true
	filter := NewMEVFilter(config)

	txs := []*types.Transaction{
		mevTestTx(dex),
		mevTestTx(other),
		mevTestTx(dex),
	}

	result := filter.FilterNonMEV(txs)
	if len(result) != 1 {
		t.Errorf("FilterNonMEV = %d, want 1", len(result))
	}
}

// --- ValidateBuilderCompliance tests ---

func TestValidateBuilderCompliance(t *testing.T) {
	focilHash1 := types.Hash{0x01}
	focilHash2 := types.Hash{0x02}
	blockHash3 := types.Hash{0x03}

	t.Run("compliant when all FOCIL included", func(t *testing.T) {
		blockHashes := []types.Hash{focilHash1, focilHash2, blockHash3}
		focilHashes := []types.Hash{focilHash1, focilHash2}

		result := ValidateBuilderCompliance(blockHashes, focilHashes, nil, false, nil)
		if !result.Compliant {
			t.Error("expected compliant")
		}
		if len(result.MissingFOCILTxs) != 0 {
			t.Errorf("MissingFOCILTxs = %d, want 0", len(result.MissingFOCILTxs))
		}
	})

	t.Run("non-compliant when FOCIL tx missing", func(t *testing.T) {
		blockHashes := []types.Hash{focilHash1, blockHash3}
		focilHashes := []types.Hash{focilHash1, focilHash2}

		result := ValidateBuilderCompliance(blockHashes, focilHashes, nil, false, nil)
		if result.Compliant {
			t.Error("expected non-compliant")
		}
		if len(result.MissingFOCILTxs) != 1 {
			t.Fatalf("MissingFOCILTxs = %d, want 1", len(result.MissingFOCILTxs))
		}
		if result.MissingFOCILTxs[0] != focilHash2 {
			t.Errorf("missing = %x, want %x", result.MissingFOCILTxs[0], focilHash2)
		}
	})
}

func mevTestTxWithNonce(to types.Address, nonce uint64) *types.Transaction {
	toAddr := to
	return types.NewTransaction(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(1000),
		Gas:      21000,
		To:       &toAddr,
		Value:    big.NewInt(0),
		V:        big.NewInt(27),
		R:        big.NewInt(1),
		S:        big.NewInt(1),
	})
}

func TestValidateBuilderCompliance_MEVOnly(t *testing.T) {
	dex := types.HexToAddress("0x1111111111111111111111111111111111111111")
	other := types.HexToAddress("0x3333333333333333333333333333333333333333")

	config := DefaultMEVFilterConfig()
	config.KnownDEXContracts[dex] = true
	filter := NewMEVFilter(config)

	// Use different nonces so each tx has a distinct hash.
	focilTx := mevTestTxWithNonce(other, 0)
	focilTx.SetSender(types.Address{0x01})
	focilHash := focilTx.Hash()

	mevTx := mevTestTxWithNonce(dex, 1)
	mevTx.SetSender(types.Address{0x02})

	nonMEVTx := mevTestTxWithNonce(other, 2)
	nonMEVTx.SetSender(types.Address{0x03})
	nonMEVHash := nonMEVTx.Hash()

	blockHashes := []types.Hash{focilHash, mevTx.Hash(), nonMEVHash}
	focilHashes := []types.Hash{focilHash}
	builderTxs := []*types.Transaction{focilTx, mevTx, nonMEVTx}

	t.Run("non-MEV non-FOCIL tx flagged", func(t *testing.T) {
		result := ValidateBuilderCompliance(blockHashes, focilHashes, builderTxs, true, filter)
		if result.Compliant {
			t.Error("expected non-compliant: non-MEV non-FOCIL tx present")
		}
		if len(result.NonMEVTxs) != 1 {
			t.Fatalf("NonMEVTxs = %d, want 1", len(result.NonMEVTxs))
		}
		if result.NonMEVTxs[0] != nonMEVHash {
			t.Errorf("NonMEVTxs[0] = %x, want %x", result.NonMEVTxs[0], nonMEVHash)
		}
	})

	t.Run("MEV-only mode not applied when mevOnly=false", func(t *testing.T) {
		result := ValidateBuilderCompliance(blockHashes, focilHashes, builderTxs, false, filter)
		if !result.Compliant {
			t.Error("expected compliant when mevOnly=false")
		}
	})

	t.Run("MEV-only mode not applied when filter=nil", func(t *testing.T) {
		result := ValidateBuilderCompliance(blockHashes, focilHashes, builderTxs, true, nil)
		if !result.Compliant {
			t.Error("expected compliant when filter=nil")
		}
	})
}

// --- AddDEXContract tests ---

func TestAddDEXContract(t *testing.T) {
	filter := NewMEVFilter(nil)
	addr := types.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

	tx := mevTestTx(addr)
	if filter.IsMEVTransaction(tx) {
		t.Error("should not be MEV before registration")
	}

	filter.AddDEXContract(addr)
	if !filter.IsMEVTransaction(tx) {
		t.Error("should be MEV after registration")
	}
}

// --- AddLiquidationContract tests ---

func TestAddLiquidationContract(t *testing.T) {
	filter := NewMEVFilter(nil)
	addr := types.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")

	tx := mevTestTx(addr)
	if filter.IsMEVTransaction(tx) {
		t.Error("should not be MEV before registration")
	}

	filter.AddLiquidationContract(addr)
	if !filter.IsMEVTransaction(tx) {
		t.Error("should be MEV after registration")
	}
}

// --- DefaultMEVFilterConfig tests ---

func TestDefaultMEVFilterConfig(t *testing.T) {
	config := DefaultMEVFilterConfig()
	if config.MinGasPriceMultiplier != 5 {
		t.Errorf("MinGasPriceMultiplier = %d, want 5", config.MinGasPriceMultiplier)
	}
	if config.KnownDEXContracts == nil {
		t.Error("KnownDEXContracts should not be nil")
	}
	if config.KnownLiquidationContracts == nil {
		t.Error("KnownLiquidationContracts should not be nil")
	}
}

// --- NewMEVFilter nil config test ---

func TestNewMEVFilter_NilConfig(t *testing.T) {
	filter := NewMEVFilter(nil)
	if filter == nil {
		t.Fatal("filter should not be nil")
	}
	// Should not panic on any operation.
	filter.IsMEVTransaction(nil)
	filter.FilterMEVOnly(nil)
	filter.FilterNonMEV(nil)
}
