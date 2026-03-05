package focil

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

// InclusionListUnsatisfied is the status returned by engine_newPayload when a
// valid IL transaction is absent from the block (EIP-7805 §engine-api).
const InclusionListUnsatisfied = "INCLUSION_LIST_UNSATISFIED"

// ILSatisfactionResult indicates whether the inclusion list was satisfied.
type ILSatisfactionResult int

const (
	// ILSatisfied means all IL txs are either in the block or exempt.
	ILSatisfied ILSatisfactionResult = iota
	// ILUnsatisfied means a valid IL tx was absent and gas was available.
	ILUnsatisfied
)

// PostStateReader provides account state needed for IL satisfaction checks.
type PostStateReader interface {
	GetNonce(addr [20]byte) uint64
	GetBalance(addr [20]byte) uint64
}

// CheckILSatisfaction implements the EIP-7805 §satisfaction algorithm:
// For each tx T in ILs:
//  1. If T is in block → skip (satisfied).
//  2. If gasRemaining < T.gasLimit → skip (gas exemption).
//  3. Validate T's nonce and balance against postState.
//     If valid but T is absent → return ILUnsatisfied.
//
// gasRemaining is the remaining gas budget after block execution.
func CheckILSatisfaction(block *types.Block, ils []*InclusionList, postState PostStateReader, gasRemaining uint64) ILSatisfactionResult {
	// Build set of tx hashes in the block.
	blockTxHashes := make(map[types.Hash]bool, len(block.Transactions()))
	for _, tx := range block.Transactions() {
		blockTxHashes[tx.Hash()] = true
	}

	for _, il := range ils {
		for _, entry := range il.Entries {
			tx, err := types.DecodeTxRLP(entry.Transaction)
			if err != nil {
				// Invalid tx — skip (per spec).
				continue
			}
			// Rule 1: tx in block → satisfied.
			if blockTxHashes[tx.Hash()] {
				continue
			}
			// Rule 2: insufficient gas remaining → skip (gas exemption).
			if gasRemaining < tx.Gas() {
				continue
			}
			// Rule 3: check nonce and balance validity against post-state.
			// State-invalid txs are exempt (not required to be included).
			if postState != nil {
				if senderPtr := tx.Sender(); senderPtr != nil {
					from := [20]byte(*senderPtr)
					if postState.GetNonce(from) != tx.Nonce() {
						continue // invalid nonce → exempt
					}
					// Estimate cost: gas * gasPrice + value.
					gasPrice := tx.GasPrice()
					if gasPrice == nil {
						gasPrice = new(big.Int)
					}
					cost := new(big.Int).Mul(new(big.Int).SetUint64(tx.Gas()), gasPrice)
					if v := tx.Value(); v != nil {
						cost.Add(cost, v)
					}
					if cost.IsUint64() && postState.GetBalance(from) < cost.Uint64() {
						continue // insufficient balance → exempt
					}
				}
			}
			// Valid tx absent from block → unsatisfied.
			return ILUnsatisfied
		}
	}
	return ILSatisfied
}
