package engine

import (
	"encoding/json"

	"github.com/eth2030/eth2030/core/rawdb"
	"github.com/eth2030/eth2030/core/types"
	enginepayload "github.com/eth2030/eth2030/engine/payload"
)

// GetHeadHash returns the current canonical head block hash.
func (b *EngineBackend) GetHeadHash() types.Hash {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.headHash
}

// GetSafeHash returns the current safe (justified) block hash.
func (b *EngineBackend) GetSafeHash() types.Hash {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.safeHash
}

// GetFinalizedHash returns the current finalized block hash.
func (b *EngineBackend) GetFinalizedHash() types.Hash {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.finalHash
}

// GetPayloadBodiesByHash returns payload bodies for the given block hashes.
// Entries for unknown or out-of-retention-window blocks are nil.
// Implements backendapi.PayloadBodiesBackend.
func (b *EngineBackend) GetPayloadBodiesByHash(hashes []types.Hash) ([]*enginepayload.ExecutionPayloadBodyV2, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	headNum := uint64(0)
	if head, ok := b.blocks[b.headHash]; ok {
		headNum = head.NumberU64()
	}

	results := make([]*enginepayload.ExecutionPayloadBodyV2, len(hashes))
	for i, h := range hashes {
		block, found := b.blocks[h]
		if !found || !rawdb.IsBALRetained(headNum, block.NumberU64()) {
			results[i] = nil
			continue
		}
		body := enginepayload.BlockToPayloadBodyV2(block)
		if bal, ok := b.bals[h]; ok {
			balBytes, _ := json.Marshal(bal)
			body.BlockAccessList = balBytes
		}
		results[i] = body
	}
	return results, nil
}

// GetPayloadBodiesByRange returns payload bodies for count blocks starting at start.
// Entries outside the retention window are nil.
// Implements backendapi.PayloadBodiesBackend.
func (b *EngineBackend) GetPayloadBodiesByRange(start, count uint64) ([]*enginepayload.ExecutionPayloadBodyV2, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	headNum := uint64(0)
	if head, ok := b.blocks[b.headHash]; ok {
		headNum = head.NumberU64()
	}

	results := make([]*enginepayload.ExecutionPayloadBodyV2, count)
	for i := uint64(0); i < count; i++ {
		num := start + i
		var found *types.Block
		for _, blk := range b.blocks {
			if blk.NumberU64() == num {
				found = blk
				break
			}
		}
		if found == nil || !rawdb.IsBALRetained(headNum, num) {
			results[i] = nil
			continue
		}
		body := enginepayload.BlockToPayloadBodyV2(found)
		if bal, ok := b.bals[found.Hash()]; ok {
			balBytes, _ := json.Marshal(bal)
			body.BlockAccessList = balBytes
		}
		results[i] = body
	}
	return results, nil
}
