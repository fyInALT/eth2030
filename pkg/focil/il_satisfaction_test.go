package focil

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// makeMinimalTx creates a minimal legacy tx RLP with unique id byte.
func makeMinimalTx(id byte) []byte {
	to := types.Address{0xAA}
	tx := types.NewTransaction(&types.LegacyTx{
		Nonce:    uint64(id),
		To:       &to,
		Value:    big.NewInt(0),
		Gas:      21000,
		GasPrice: big.NewInt(1),
		Data:     []byte{id},
	})
	enc, _ := tx.EncodeRLP()
	return enc
}

// makeBlockWithTxs builds a minimal block containing the given raw txs.
func makeBlockWithTxs(rawTxs ...[]byte) *types.Block {
	var txs []*types.Transaction
	for _, raw := range rawTxs {
		tx, err := types.DecodeTxRLP(raw)
		if err != nil {
			continue
		}
		txs = append(txs, tx)
	}
	header := &types.Header{Number: big.NewInt(1)}
	body := &types.Body{Transactions: txs}
	return types.NewBlock(header, body)
}

// makeILWithRawTxs builds an InclusionList containing the given raw txs.
func makeILWithRawTxs(rawTxs ...[]byte) *InclusionList {
	il := &InclusionList{Slot: 1, ProposerIndex: 1}
	for i, raw := range rawTxs {
		il.Entries = append(il.Entries, InclusionListEntry{
			Transaction: raw,
			Index:       uint64(i),
		})
	}
	return il
}

// mockPostState implements PostStateReader for IL satisfaction tests.
type mockPostState struct {
	nonces   map[[20]byte]uint64
	balances map[[20]byte]uint64
}

func newMockPostState() *mockPostState {
	return &mockPostState{
		nonces:   make(map[[20]byte]uint64),
		balances: make(map[[20]byte]uint64),
	}
}

func (m *mockPostState) GetNonce(addr [20]byte) uint64   { return m.nonces[addr] }
func (m *mockPostState) GetBalance(addr [20]byte) uint64 { return m.balances[addr] }

func TestILSatisfactionAllInBlock(t *testing.T) {
	// All IL txs are in block → satisfied.
	tx1 := makeMinimalTx(1)
	block := makeBlockWithTxs(tx1)
	il := makeILWithRawTxs(tx1)

	state := newMockPostState()
	result := CheckILSatisfaction(block, []*InclusionList{il}, state, 1000000)
	if result != ILSatisfied {
		t.Errorf("expected ILSatisfied, got %v", result)
	}
}

func TestILSatisfactionAbsentValidTx(t *testing.T) {
	// IL tx absent, gas available, nonce/balance valid → unsatisfied.
	tx1 := makeMinimalTx(1)
	block := makeBlockWithTxs() // empty block
	il := makeILWithRawTxs(tx1)

	sender := [20]byte{0xAA}
	state := newMockPostState()
	state.nonces[sender] = 0      // tx nonce would be valid
	state.balances[sender] = 1e18 // plenty of balance

	result := CheckILSatisfaction(block, []*InclusionList{il}, state, 1000000)
	if result != ILUnsatisfied {
		t.Errorf("expected ILUnsatisfied for absent valid tx, got %v", result)
	}
}

func TestILSatisfactionGasExemption(t *testing.T) {
	// IL tx absent but block has no gas remaining → satisfied (gas exemption).
	tx1 := makeMinimalTx(1)
	block := makeBlockWithTxs() // empty block
	il := makeILWithRawTxs(tx1)

	state := newMockPostState()
	// gasRemaining = 0, tx gas limit > 0 → gas exemption.
	result := CheckILSatisfaction(block, []*InclusionList{il}, state, 0)
	if result != ILSatisfied {
		t.Errorf("expected ILSatisfied (gas exemption), got %v", result)
	}
}

func TestILSatisfactionConstant(t *testing.T) {
	if InclusionListUnsatisfied != "INCLUSION_LIST_UNSATISFIED" {
		t.Errorf("InclusionListUnsatisfied = %q, want INCLUSION_LIST_UNSATISFIED", InclusionListUnsatisfied)
	}
}
