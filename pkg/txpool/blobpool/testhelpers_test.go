package blobpool

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

// mockState implements StateReader for testing.
type mockState struct {
	nonces   map[types.Address]uint64
	balances map[types.Address]*big.Int
}

func newMockState() *mockState {
	return &mockState{
		nonces:   make(map[types.Address]uint64),
		balances: make(map[types.Address]*big.Int),
	}
}

func (s *mockState) GetNonce(addr types.Address) uint64 {
	return s.nonces[addr]
}

func (s *mockState) GetBalance(addr types.Address) *big.Int {
	bal, ok := s.balances[addr]
	if !ok {
		return new(big.Int)
	}
	return bal
}

// testSender is the default sender address used in tests.
var testSender = types.BytesToAddress([]byte{0x01, 0x02, 0x03})

// richBalance is a large balance used for tests that don't care about balance.
var richBalance = new(big.Int).Mul(big.NewInt(1e18), big.NewInt(1_000_000))

// makeTx creates a minimal legacy transaction for testing.
func makeTx(nonce uint64, gasPrice int64, gas uint64) *types.Transaction {
	to := types.BytesToAddress([]byte{0xde, 0xad})
	tx := types.NewTransaction(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(gasPrice),
		Gas:      gas,
		To:       &to,
		Value:    big.NewInt(0),
	})
	tx.SetSender(testSender)
	return tx
}
