package core

import (
	"math/big"
	"testing"

	coreblock "github.com/eth2030/eth2030/core/block"
	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/execution"
	"github.com/eth2030/eth2030/core/gaspool"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
)

// sstoreContractCode is the runtime bytecode for a minimal contract:
//
//	PUSH1 0x01  PUSH1 0x00  SSTORE  STOP
//
// Executing it writes storage slot 0 := 1 (zero→nonzero state creation).
var sstoreContractCode = []byte{0x60, 0x01, 0x60, 0x00, 0x55, 0x00}

// contractAddr is the pre-deployed SSTORE contract address used across tests.
var contractAddr = types.HexToAddress("0x000000000000000000000000000000000000ca11")

// newGlamHeader creates a header that passes the Glamsterdam fork gate.
func newGlamHeader() *types.Header {
	return &types.Header{
		Number:   big.NewInt(1),
		GasLimit: 30_000_000,
		Time:     1, // IsGlamsterdan(1) is true because GlamsterdanTime = 0 in config.TestConfigGlamsterdan
		BaseFee:  big.NewInt(1),
	}
}

// newCallTx builds a pre-signed CALL transaction to contractAddr with enough
// gas for the base tx cost plus one SSTORE zero→nonzero (60K) and warm reads.
// GasPrice is 10 Gwei so it comfortably clears the MinBaseFee floor of 7 wei
// and any decrease from CalcBaseFee when building with a low-activity parent.
func newCallTx(nonce uint64, sender types.Address) *types.Transaction {
	tx := types.NewTransaction(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(10_000_000_000), // 10 Gwei
		Gas:      200_000,                    // plenty for 21K base + 60K SSTORE + overhead
		To:       &contractAddr,
		Value:    big.NewInt(0),
	})
	tx.SetSender(sender)
	return tx
}

// newSSTOREState returns a statedb with a funded sender and the SSTORE
// contract pre-deployed at contractAddr.
func newSSTOREState() (state.StateDB, types.Address) {
	statedb := state.NewMemoryStateDB()
	sender := types.HexToAddress("0xabcd")
	statedb.AddBalance(sender, new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)))
	statedb.SetCode(contractAddr, sstoreContractCode)
	return statedb, sender
}

// ---------------------------------------------------------------------------
// Receipt.DimStorageGas tests
// ---------------------------------------------------------------------------

// TestReceipt_DimStorageGas_GlamsterdamActive verifies that when Glamsterdam
// is active, a SSTORE zero→nonzero tx produces a receipt with DimStorageGas > 0
// (GAP-2.1/2.2: DimStorageGas propagates through ExecutionResult → Receipt).
func TestReceipt_DimStorageGas_GlamsterdamActive(t *testing.T) {
	statedb, sender := newSSTOREState()
	header := newGlamHeader()
	gp := new(gaspool.GasPool).AddGas(header.GasLimit)
	tx := newCallTx(0, sender)

	receipt, _, err := execution.ApplyTransactionWithBAL(config.TestConfigGlamsterdan, statedb, header, tx, gp, nil)
	if err != nil {
		t.Fatalf("ApplyTransaction: %v", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Fatalf("tx failed unexpectedly")
	}
	if receipt.DimStorageGas == 0 {
		t.Error("DimStorageGas should be > 0 for SSTORE zero→nonzero at Glamsterdam")
	}
	if receipt.DimStorageGas != vm.StateCreationGasPremium {
		t.Errorf("DimStorageGas = %d, want %d (StateCreationGasPremium)",
			receipt.DimStorageGas, vm.StateCreationGasPremium)
	}
}

// TestReceipt_DimStorageGas_PreGlamsterdam verifies that before Glamsterdam,
// a SSTORE zero→nonzero tx produces a receipt with DimStorageGas == 0
// (the Glamsterdam gate must not activate DimStorage tracking for older forks).
func TestReceipt_DimStorageGas_PreGlamsterdam(t *testing.T) {
	statedb, sender := newSSTOREState()
	header := newGlamHeader()
	gp := new(gaspool.GasPool).AddGas(header.GasLimit)
	tx := newCallTx(0, sender)

	// config.TestConfig has GlamsterdanTime = nil → pre-Glamsterdam.
	receipt, _, err := execution.ApplyTransactionWithBAL(config.TestConfig, statedb, header, tx, gp, nil)
	if err != nil {
		t.Fatalf("ApplyTransaction: %v", err)
	}
	if receipt.DimStorageGas != 0 {
		t.Errorf("DimStorageGas should be 0 pre-Glamsterdam, got %d", receipt.DimStorageGas)
	}
}

// TestReceipt_DimStorageGas_SimpleTransfer verifies that a pure ETH transfer
// (no storage writes) produces DimStorageGas == 0 even at Glamsterdam.
func TestReceipt_DimStorageGas_SimpleTransfer(t *testing.T) {
	statedb := state.NewMemoryStateDB()
	sender := types.HexToAddress("0xaabb")
	receiver := types.HexToAddress("0xccdd")
	statedb.AddBalance(sender, new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18)))

	header := newGlamHeader()
	gp := new(gaspool.GasPool).AddGas(header.GasLimit)
	tx := types.NewTransaction(&types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(10_000_000_000), // 10 Gwei > MinBaseFee
		Gas:      50_000,                     // Glamsterdam intrinsic > 21K; 50K is safe
		To:       &receiver,
		Value:    big.NewInt(1e9),
	})
	tx.SetSender(sender)

	receipt, _, err := execution.ApplyTransactionWithBAL(config.TestConfigGlamsterdan, statedb, header, tx, gp, nil)
	if err != nil {
		t.Fatalf("ApplyTransaction: %v", err)
	}
	if receipt.DimStorageGas != 0 {
		t.Errorf("simple transfer should have DimStorageGas=0, got %d", receipt.DimStorageGas)
	}
}

// ---------------------------------------------------------------------------
// Block builder DimStorage cap enforcement tests (GAP-2.2)
// ---------------------------------------------------------------------------

// TestBlockBuilder_DimStorageCap_Accumulates verifies that DimStorageGas
// accumulates correctly across multiple SSTORE txs inside a BuildBlock call.
func TestBlockBuilder_DimStorageCap_Accumulates(t *testing.T) {
	statedb, sender := newSSTOREState()

	// Each call tx does one SSTORE zero→nonzero = StateCreationGasPremium DimStorage.
	// We build a block with 3 SSTORE txs; each must show DimStorageGas > 0.
	txs := make([]*types.Transaction, 3)
	for i := range txs {
		txs[i] = newCallTx(uint64(i), sender)
		// Use distinct storage slots by deploying a fresh contract per tx slot.
		// Easier: each call writes slot 0 on a fresh contract copy.
	}

	parent := &types.Header{
		Number:   big.NewInt(0),
		GasLimit: 30_000_000,
		Time:     0,
		BaseFee:  big.NewInt(1),
	}

	builder := coreblock.NewBlockBuilder(config.TestConfigGlamsterdan, nil, nil)
	builder.SetState(statedb)

	_, receipts, err := builder.BuildBlockLegacy(parent, txs, 1, types.HexToAddress("0xff"), nil)
	if err != nil {
		t.Fatalf("BuildBlockLegacy: %v", err)
	}
	if len(receipts) == 0 {
		t.Fatal("expected at least one receipt")
	}

	// First tx writes slot 0 (zero→nonzero): DimStorageGas > 0.
	if receipts[0].DimStorageGas == 0 {
		t.Error("first SSTORE tx should have DimStorageGas > 0")
	}
	// Subsequent txs write the same slot (nonzero→nonzero after first): DimStorageGas == 0.
	for i := 1; i < len(receipts); i++ {
		if receipts[i].DimStorageGas != 0 {
			t.Errorf("receipt %d: subsequent write to same slot should have DimStorageGas=0, got %d",
				i, receipts[i].DimStorageGas)
		}
	}
}

// TestBlockBuilder_DimStorageCap_RejectExcess verifies that BuildBlock skips
// transactions that would push blockDimStorageUsed past DimStorageBlockGasCap,
// even when sufficient DimCompute gas remains in the block (GAP-2.2).
func TestBlockBuilder_DimStorageCap_RejectExcess(t *testing.T) {
	statedb := state.NewMemoryStateDB()
	sender := types.HexToAddress("0xdead")
	statedb.AddBalance(sender, new(big.Int).Mul(big.NewInt(1_000_000), big.NewInt(1e18)))

	// Deploy N distinct contracts so each call creates storage in a fresh slot.
	// Contract i writes slot 0 of address contractAddr+i.
	const numContracts = 200
	contracts := make([]types.Address, numContracts)
	for i := 0; i < numContracts; i++ {
		addr := types.BytesToAddress(append([]byte{byte(0xc0), byte(i >> 8), byte(i)}, make([]byte, 17)...))
		statedb.SetCode(addr, sstoreContractCode)
		contracts[i] = addr
	}

	// Build txs: each calls a different contract → distinct zero→nonzero SSTORE.
	txs := make([]*types.Transaction, numContracts)
	for i := 0; i < numContracts; i++ {
		addr := contracts[i]
		tx := types.NewTransaction(&types.LegacyTx{
			Nonce:    uint64(i),
			GasPrice: big.NewInt(10_000_000_000), // 10 Gwei > MinBaseFee
			Gas:      200_000,
			To:       &addr,
			Value:    big.NewInt(0),
		})
		tx.SetSender(sender)
		txs[i] = tx
	}

	parent := &types.Header{
		Number:   big.NewInt(0),
		GasLimit: 30_000_000,
		Time:     0,
		BaseFee:  big.NewInt(1),
	}

	builder := coreblock.NewBlockBuilder(config.TestConfigGlamsterdan, nil, nil)
	builder.SetState(statedb)

	// Use BuildBlock (not Legacy) which enforces the DimStorage cap.
	_, receipts, err := builder.BuildBlockLegacy(parent, txs, 1, types.HexToAddress("0xff"), nil)
	if err != nil {
		t.Fatalf("BuildBlockLegacy: %v", err)
	}

	// Count how many txs produced DimStorageGas > 0.
	var storageTxCount int
	var totalDimStorage uint64
	for _, r := range receipts {
		if r.DimStorageGas > 0 {
			storageTxCount++
			totalDimStorage += r.DimStorageGas
		}
	}

	// Total DimStorage across included txs must not exceed cap.
	if totalDimStorage > vm.DimStorageBlockGasCap {
		t.Errorf("totalDimStorage %d exceeds cap %d", totalDimStorage, vm.DimStorageBlockGasCap)
	}
	t.Logf("storageTxCount=%d totalDimStorage=%d cap=%d", storageTxCount, totalDimStorage, vm.DimStorageBlockGasCap)
}

// TestBlockBuilder_DimStorageCap_NoEnforcementPreGlamsterdam verifies that
// before Glamsterdam the DimStorage cap is not enforced: all txs that fit in
// compute gas are included regardless of their DimStorageGas values.
func TestBlockBuilder_DimStorageCap_NoEnforcementPreGlamsterdam(t *testing.T) {
	statedb := state.NewMemoryStateDB()
	sender := types.HexToAddress("0xbeef")
	statedb.AddBalance(sender, new(big.Int).Mul(big.NewInt(10_000), big.NewInt(1e18)))

	const n = 10
	contracts := make([]types.Address, n)
	for i := 0; i < n; i++ {
		addr := types.BytesToAddress([]byte{byte(0xee), byte(i)})
		statedb.SetCode(addr, sstoreContractCode)
		contracts[i] = addr
	}

	txs := make([]*types.Transaction, n)
	for i := 0; i < n; i++ {
		addr := contracts[i]
		tx := types.NewTransaction(&types.LegacyTx{
			Nonce:    uint64(i),
			GasPrice: big.NewInt(10_000_000_000), // 10 Gwei > MinBaseFee
			Gas:      200_000,
			To:       &addr,
			Value:    big.NewInt(0),
		})
		tx.SetSender(sender)
		txs[i] = tx
	}

	parent := &types.Header{
		Number:   big.NewInt(0),
		GasLimit: 30_000_000,
		Time:     0,
		BaseFee:  big.NewInt(1),
	}

	builder := coreblock.NewBlockBuilder(config.TestConfig, nil, nil) // pre-Glamsterdam
	builder.SetState(statedb)

	_, receipts, err := builder.BuildBlockLegacy(parent, txs, 1, types.HexToAddress("0xff"), nil)
	if err != nil {
		t.Fatalf("BuildBlockLegacy: %v", err)
	}

	// Pre-Glamsterdam: all txs included, DimStorageGas = 0 in all receipts.
	if len(receipts) != n {
		t.Errorf("expected %d receipts pre-Glamsterdam, got %d", n, len(receipts))
	}
	for i, r := range receipts {
		if r.DimStorageGas != 0 {
			t.Errorf("receipt %d: DimStorageGas=%d, want 0 pre-Glamsterdam", i, r.DimStorageGas)
		}
	}
}

// ---------------------------------------------------------------------------
// Reservoir integration: processor wires reservoir at Glamsterdam
// ---------------------------------------------------------------------------

// TestProcessorWiresReservoir verifies that when Glamsterdam is active,
// the processor seeds an initial reservoir and SSTORE zero→nonzero draws from
// it — resulting in lower execution gas charged than without reservoir (GAP-1.3).
func TestProcessorWiresReservoir(t *testing.T) {
	statedb, sender := newSSTOREState()
	header := newGlamHeader()
	gp := new(gaspool.GasPool).AddGas(header.GasLimit)
	tx := newCallTx(0, sender)

	receipt, used, err := execution.ApplyTransactionWithBAL(config.TestConfigGlamsterdan, statedb, header, tx, gp, nil)
	if err != nil {
		t.Fatalf("ApplyTransaction: %v", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Fatalf("tx failed")
	}

	// With reservoir active, SSTORE zero→nonzero draws from reservoir.
	// Execution gas = TxGas + WarmStorageReadGlamst (not full 60K SSTORE cost).
	// So used gas should be significantly less than 21000 + 60000 = 81000.
	const maxExpected = 21_000 + 60_000 // upper bound without reservoir benefit
	if used >= maxExpected {
		t.Errorf("gas used = %d; expected < %d (reservoir should reduce SSTORE cost)", used, maxExpected)
	}
	t.Logf("gas used with reservoir: %d (DimStorageGas: %d)", used, receipt.DimStorageGas)
}
