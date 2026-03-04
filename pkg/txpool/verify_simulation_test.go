package txpool

import (
	"math/big"
	"strings"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// mockFrameState implements FrameStateReader for use in verify simulation tests.
type mockFrameState struct {
	nonces    map[types.Address]uint64
	balances  map[types.Address]*big.Int
	codeSizes map[types.Address]int
}

func (m *mockFrameState) GetNonce(addr types.Address) uint64 {
	return m.nonces[addr]
}

func (m *mockFrameState) GetBalance(addr types.Address) *big.Int {
	if b, ok := m.balances[addr]; ok {
		return b
	}
	return new(big.Int)
}

func (m *mockFrameState) GetCodeSize(addr types.Address) int {
	return m.codeSizes[addr]
}

// newMockFrameState returns a zeroed mockFrameState with initialised maps.
func newMockFrameState() *mockFrameState {
	return &mockFrameState{
		nonces:    make(map[types.Address]uint64),
		balances:  make(map[types.Address]*big.Int),
		codeSizes: make(map[types.Address]int),
	}
}

// makeTestFrameTx builds a minimal FrameTx for testing.
func makeTestFrameTx(frames []types.Frame, sender types.Address) *types.FrameTx {
	return &types.FrameTx{
		ChainID:      big.NewInt(1),
		Nonce:        big.NewInt(0),
		Sender:       sender,
		Frames:       frames,
		MaxFeePerGas: big.NewInt(1e9),
	}
}

var (
	verifySender = types.HexToAddress("0x1111111111111111111111111111111111111111")
	verifyTarget = types.HexToAddress("0x2222222222222222222222222222222222222222")
)

// TestSimulateVerifyFrame_NoCode_EOA verifies that a VERIFY frame targeting an
// address with no deployed code returns an error containing "has no code (EOA)".
func TestSimulateVerifyFrame_NoCode_EOA(t *testing.T) {
	state := newMockFrameState()
	// verifyTarget has code size 0 (EOA)

	target := verifyTarget
	tx := makeTestFrameTx([]types.Frame{
		{Mode: types.ModeVerify, Target: &target, GasLimit: 50000},
	}, verifySender)

	err := SimulateVerifyFrame(tx, state)
	if err == nil {
		t.Fatal("expected error for VERIFY targeting EOA, got nil")
	}
	if !strings.Contains(err.Error(), "has no code (EOA)") {
		t.Fatalf("expected 'has no code (EOA)' in error, got: %v", err)
	}
}

// TestSimulateVerifyFrame_HasCode verifies that a VERIFY frame targeting an
// address with deployed code returns nil (pre-check passes).
func TestSimulateVerifyFrame_HasCode(t *testing.T) {
	state := newMockFrameState()
	state.codeSizes[verifyTarget] = 128 // has code

	target := verifyTarget
	tx := makeTestFrameTx([]types.Frame{
		{Mode: types.ModeVerify, Target: &target, GasLimit: 50000},
	}, verifySender)

	if err := SimulateVerifyFrame(tx, state); err != nil {
		t.Fatalf("expected nil error for VERIFY targeting contract, got: %v", err)
	}
}

// TestSimulateVerifyFrame_NoVerifyFrame verifies that a FrameTx containing only
// DEFAULT frames returns an error containing "no VERIFY frame".
func TestSimulateVerifyFrame_NoVerifyFrame(t *testing.T) {
	state := newMockFrameState()

	target := verifyTarget
	tx := makeTestFrameTx([]types.Frame{
		{Mode: types.ModeDefault, Target: &target, GasLimit: 21000},
	}, verifySender)

	err := SimulateVerifyFrame(tx, state)
	if err == nil {
		t.Fatal("expected error for tx with no VERIFY frame, got nil")
	}
	if !strings.Contains(err.Error(), "no VERIFY frame") {
		t.Fatalf("expected 'no VERIFY frame' in error, got: %v", err)
	}
}

// TestSimulateVerifyFrame_NilTx verifies that passing a nil FrameTx returns an error.
func TestSimulateVerifyFrame_NilTx(t *testing.T) {
	state := newMockFrameState()

	err := SimulateVerifyFrame(nil, state)
	if err == nil {
		t.Fatal("expected error for nil tx, got nil")
	}
}

// TestSimulateVerifyFrame_SenderAsTarget verifies that when a VERIFY frame has a
// nil Target, the tx.Sender is used as the target. When the sender has no code,
// an error is returned.
func TestSimulateVerifyFrame_SenderAsTarget(t *testing.T) {
	state := newMockFrameState()
	// verifySender has code size 0 (no code)

	tx := makeTestFrameTx([]types.Frame{
		{Mode: types.ModeVerify, Target: nil, GasLimit: 50000},
	}, verifySender)

	err := SimulateVerifyFrame(tx, state)
	if err == nil {
		t.Fatal("expected error for VERIFY targeting sender with no code, got nil")
	}
	if !strings.Contains(err.Error(), "has no code (EOA)") {
		t.Fatalf("expected 'has no code (EOA)' in error, got: %v", err)
	}
}

// TestSimulateVerifyFrame_SenderHasCode verifies that when a VERIFY frame has a
// nil Target and the sender has deployed code, nil is returned.
func TestSimulateVerifyFrame_SenderHasCode(t *testing.T) {
	state := newMockFrameState()
	state.codeSizes[verifySender] = 256 // sender has code

	tx := makeTestFrameTx([]types.Frame{
		{Mode: types.ModeVerify, Target: nil, GasLimit: 50000},
	}, verifySender)

	if err := SimulateVerifyFrame(tx, state); err != nil {
		t.Fatalf("expected nil error for VERIFY targeting sender with code, got: %v", err)
	}
}

// BenchmarkVerifyFrameSimulation measures the overhead of SimulateVerifyFrame
// under the common success path (VERIFY target has code).
func BenchmarkVerifyFrameSimulation(b *testing.B) {
	state := newMockFrameState()
	state.codeSizes[verifyTarget] = 128

	target := verifyTarget
	tx := makeTestFrameTx([]types.Frame{
		{Mode: types.ModeVerify, Target: &target, GasLimit: 50000},
	}, verifySender)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SimulateVerifyFrame(tx, state)
	}
}
