package core

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

// newTransferTx creates a simple transfer transaction.
func newTransferTx(nonce uint64, to types.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int) *types.Transaction {
	toAddr := to
	return types.NewTransaction(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddr,
		Value:    value,
	})
}

func newTestHeader() *types.Header {
	return &types.Header{
		Number:   big.NewInt(1),
		GasLimit: 10_000_000,
		Time:     1000,
		BaseFee:  big.NewInt(1), // Low base fee so GasPrice=1 txs pass EIP-1559 validation
		Coinbase: types.HexToAddress("0xfee"),
	}
}
