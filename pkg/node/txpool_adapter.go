package node

import (
	"github.com/eth2030/eth2030/core/mev"
	"github.com/eth2030/eth2030/core/types"
)

// txPoolAdapter adapts *txpool.TxPool to core.TxPoolReader.
type txPoolAdapter struct {
	node *Node
}

func (a *txPoolAdapter) Pending() []*types.Transaction {
	txs := a.node.txPool.PendingFlat()

	// Apply fair ordering for MEV protection when enabled.
	if a.node.mevConfig != nil && a.node.mevConfig.EnableFairOrdering && len(txs) > 0 {
		entries := make([]mev.FairOrderingEntry, len(txs))
		for i, tx := range txs {
			entries[i] = mev.FairOrderingEntry{
				Transaction: tx,
				ArrivalTime: uint64(i),
			}
		}
		ordered, _ := mev.FairOrdering(entries, a.node.mevConfig.FairOrderMaxDelay)
		txs = make([]*types.Transaction, len(ordered))
		for i, e := range ordered {
			txs[i] = e.Transaction
		}
	}
	return txs
}