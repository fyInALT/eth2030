package gas

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

func newUint64(v uint64) *uint64 { return &v }

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
