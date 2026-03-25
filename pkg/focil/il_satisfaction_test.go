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

// --- EIP-8141 frame tx FOCIL integration tests ---

// mockFrameVerifier implements FrameVerifier for testing.
type mockFrameVerifier struct {
	codeSizes map[types.Address]int
}

func newMockFrameVerifier() *mockFrameVerifier {
	return &mockFrameVerifier{codeSizes: make(map[types.Address]int)}
}

func (m *mockFrameVerifier) GetCodeSize(addr types.Address) int {
	return m.codeSizes[addr]
}

// makeMinimalFrameTx creates a minimal frame tx (type 0x06) with a VERIFY frame.
func makeMinimalFrameTx(nonce uint64, sender types.Address, verifyTarget types.Address) []byte {
	ftx := &types.FrameTx{
		ChainID:              big.NewInt(1),
		Nonce:                new(big.Int).SetUint64(nonce),
		Sender:               sender,
		MaxPriorityFeePerGas: big.NewInt(1),
		MaxFeePerGas:         big.NewInt(10),
		Frames: []types.Frame{
			{
				Mode:     types.ModeVerify,
				Target:   &verifyTarget,
				GasLimit: 50000,
				Data:     []byte{0x01},
			},
			{
				Mode:     types.ModeSender,
				Target:   &sender,
				GasLimit: 21000,
				Data:     nil,
			},
		},
	}
	tx := types.NewTransaction(ftx)
	enc, _ := tx.EncodeRLP()
	return enc
}

func TestILSatisfaction_FrameTx_VerifyTargetHasCode(t *testing.T) {
	sender := types.Address{0x10}
	verifyTarget := types.Address{0x20}
	frameTxBytes := makeMinimalFrameTx(0, sender, verifyTarget)
	block := makeBlockWithTxs() // empty block
	il := makeILWithRawTxs(frameTxBytes)

	state := newMockPostState()
	state.nonces[[20]byte(sender)] = 0

	verifier := newMockFrameVerifier()
	verifier.codeSizes[verifyTarget] = 100 // VERIFY target has code

	// Frame tx with valid nonce and VERIFY target with code → must include → unsatisfied.
	result := CheckILSatisfactionWithVerifier(block, []*InclusionList{il}, state, verifier, 1000000)
	if result != ILUnsatisfied {
		t.Errorf("frame tx with coded VERIFY target: want ILUnsatisfied, got %v", result)
	}
}

func TestILSatisfaction_FrameTx_VerifyTargetNoCode(t *testing.T) {
	sender := types.Address{0x10}
	verifyTarget := types.Address{0x20}
	frameTxBytes := makeMinimalFrameTx(0, sender, verifyTarget)
	block := makeBlockWithTxs() // empty block
	il := makeILWithRawTxs(frameTxBytes)

	state := newMockPostState()
	state.nonces[[20]byte(sender)] = 0

	verifier := newMockFrameVerifier()
	// verifyTarget has NO code (code size = 0)

	// Frame tx whose VERIFY target has no code → can't call APPROVE → exempt.
	result := CheckILSatisfactionWithVerifier(block, []*InclusionList{il}, state, verifier, 1000000)
	if result != ILSatisfied {
		t.Errorf("frame tx with EOA VERIFY target: want ILSatisfied (exempt), got %v", result)
	}
}

func TestILSatisfaction_FrameTx_SkipsBalanceCheck(t *testing.T) {
	sender := types.Address{0x10}
	verifyTarget := types.Address{0x20}
	frameTxBytes := makeMinimalFrameTx(0, sender, verifyTarget)
	block := makeBlockWithTxs() // empty block
	il := makeILWithRawTxs(frameTxBytes)

	state := newMockPostState()
	state.nonces[[20]byte(sender)] = 0
	state.balances[[20]byte(sender)] = 0 // sender has ZERO balance

	verifier := newMockFrameVerifier()
	verifier.codeSizes[verifyTarget] = 100

	// Frame tx sender has zero balance but VERIFY target has code.
	// Per EIP-8141, the payer is determined at APPROVE time (may differ from sender).
	// FOCIL should NOT exempt based on sender balance for frame txs.
	result := CheckILSatisfactionWithVerifier(block, []*InclusionList{il}, state, verifier, 1000000)
	if result != ILUnsatisfied {
		t.Errorf("frame tx with zero-balance sender: want ILUnsatisfied (balance not checked), got %v", result)
	}
}

func TestILSatisfaction_FrameTx_NoVerifier_FallsBack(t *testing.T) {
	sender := types.Address{0x10}
	verifyTarget := types.Address{0x20}
	frameTxBytes := makeMinimalFrameTx(0, sender, verifyTarget)
	block := makeBlockWithTxs() // empty block
	il := makeILWithRawTxs(frameTxBytes)

	state := newMockPostState()
	state.nonces[[20]byte(sender)] = 0

	// No verifier provided → skip VERIFY check, treat as potentially valid.
	var verifier FrameVerifier // nil
	result := CheckILSatisfactionWithVerifier(block, []*InclusionList{il}, state, verifier, 1000000)
	if result != ILUnsatisfied {
		t.Errorf("frame tx without verifier: want ILUnsatisfied, got %v", result)
	}
}

func TestILSatisfaction_FrameTx_WrongNonce_Exempt(t *testing.T) {
	sender := types.Address{0x10}
	verifyTarget := types.Address{0x20}
	frameTxBytes := makeMinimalFrameTx(5, sender, verifyTarget) // nonce=5
	block := makeBlockWithTxs()
	il := makeILWithRawTxs(frameTxBytes)

	state := newMockPostState()
	state.nonces[[20]byte(sender)] = 3 // state nonce is 3, tx nonce is 5 → mismatch

	verifier := newMockFrameVerifier()
	verifier.codeSizes[verifyTarget] = 100

	// Wrong nonce → exempt (same as EOA).
	result := CheckILSatisfactionWithVerifier(block, []*InclusionList{il}, state, verifier, 1000000)
	if result != ILSatisfied {
		t.Errorf("frame tx with wrong nonce: want ILSatisfied (exempt), got %v", result)
	}
}
