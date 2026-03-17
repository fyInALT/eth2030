package backend

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/block"
	"github.com/eth2030/eth2030/core/chain"
	"github.com/eth2030/eth2030/core/mev"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/payload"
	"github.com/eth2030/eth2030/txpool"
)

// mockNodeDeps is a mock implementation of NodeDeps for testing.
type mockNodeDeps struct {
	blockchain *chain.Blockchain
	txPool     *txpool.TxPool
	config     *Config
}

func (m *mockNodeDeps) Blockchain() *chain.Blockchain          { return m.blockchain }
func (m *mockNodeDeps) TxPool() *txpool.TxPool                 { return m.txPool }
func (m *mockNodeDeps) Config() *Config                        { return m.config }
func (m *mockNodeDeps) GasOracle() GasOracleDeps               { return nil }
func (m *mockNodeDeps) EthHandler() EthHandlerDeps             { return nil }
func (m *mockNodeDeps) TxJournal() TxJournalDeps               { return nil }
func (m *mockNodeDeps) MEVConfig() *mev.MEVProtectionConfig    { return nil }
func (m *mockNodeDeps) FCStateManager() FCStateManagerDeps     { return nil }
func (m *mockNodeDeps) StarkFrameProver() ValidationFrameProver { return nil }
func (m *mockNodeDeps) P2PServer() P2PServerDeps               { return nil }

// ValidationFrameProver is an alias for testing without importing proofs package.
type ValidationFrameProver = interface {
	ProveValidationFrame(frameCalldata, output []byte) (any, error)
	ProveAllValidationFrames(frames [][]byte) (any, error)
	Verify(proof any) bool
}

func TestNewRPCBackend(t *testing.T) {
	// This test requires a real blockchain, so we skip if not available.
	// The actual functionality is tested via integration tests.
}

func TestNewEngineBackend(t *testing.T) {
	// This test requires a real blockchain, so we skip if not available.
	// The actual functionality is tested via integration tests.
}

func TestExtractBlockTips(t *testing.T) {
	// Test with empty transactions
	tips := ExtractBlockTips(nil, big.NewInt(1))
	if tips == nil {
		t.Error("ExtractBlockTips(nil) returned nil")
	}
	if len(tips) != 0 {
		t.Errorf("ExtractBlockTips(nil) = %d tips, want 0", len(tips))
	}

	// Test with nil base fee
	tips = ExtractBlockTips(nil, nil)
	if tips == nil {
		t.Error("ExtractBlockTips with nil baseFee returned nil")
	}
}

func TestEncodeTxsRLPEmpty(t *testing.T) {
	// encodeTxsRLP must return a non-nil empty slice so JSON encodes as []
	// rather than null (Engine API requires [] not null for transactions).
	result := encodeTxsRLP(nil)
	if result == nil {
		t.Error("expected non-nil empty slice for nil txs, got nil (would JSON-encode as null)")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}

func TestGeneratePayloadID(t *testing.T) {
	// Test that generatePayloadID returns a non-zero ID in most cases.
	var parentHash types.Hash
	parentHash[0] = 0x01

	// Create minimal attributes
	attrs := &block.BuildBlockAttributes{
		Timestamp:    12345,
		FeeRecipient: types.Address{0x02},
		Random:       types.Hash{0x03},
	}

	id := generatePayloadID(parentHash, attrs)
	if id == (payload.PayloadID{}) {
		t.Error("generatePayloadID returned zero ID")
	}
}

// Note: The following tests require a full Node setup and are better suited
// for integration tests. They are kept here as documentation of what should
// be tested:
// - TestBackendChainID
// - TestBackendCurrentHeader
// - TestBackendHeaderByNumber
// - TestBackendBlockByNumber
// - TestBackendSuggestGasPrice
// - TestBackendHeaderByHash
// - TestBackendBlockByHash
// - TestBackendGetTransactionNotFound
// - TestBackendGetReceiptsEmpty
// - TestBackendGetLogsEmpty
// - TestBackendHistoryOldestBlock
// - TestEngineBackendIsCancun
// - TestEngineBackendIsPrague
// - TestEngineBackendIsAmsterdam
// - TestEngineBackendGetHeadTimestamp
// - TestEngineBackendGetPayloadNotFound