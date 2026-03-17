// block_store_actor.go provides an actor-based block store for EngineBackend.
// It manages blocks, block access lists, and provides O(1) lookups by hash or number.
package actor

import (
	"context"
	"sync/atomic"

	"github.com/eth2030/eth2030/bal"
	"github.com/eth2030/eth2030/core/types"
)

// --- Message types ---

// BlockStoreMsg stores a block with optional BAL.
type BlockStoreMsg struct {
	BaseMessage
	Hash  types.Hash
	Block *types.Block
	BAL   *bal.BlockAccessList
}

// BlockGetByHashMsg retrieves a block by hash.
type BlockGetByHashMsg struct {
	BaseMessage
	Hash types.Hash
}

// BlockGetByNumberMsg retrieves a block by number.
type BlockGetByNumberMsg struct {
	BaseMessage
	Number uint64
}

// BlockGetBALMsg retrieves a BAL by block hash.
type BlockGetBALMsg struct {
	BaseMessage
	Hash types.Hash
}

// BlockGetBodiesByHashMsg retrieves block bodies by hashes.
type BlockGetBodiesByHashMsg struct {
	BaseMessage
	Hashes []types.Hash
}

// BlockGetBodiesByRangeMsg retrieves block bodies by number range.
type BlockGetBodiesByRangeMsg struct {
	BaseMessage
	Start uint64
	Count uint64
}

// BlockSetHeadMsg sets the head block hash.
type BlockSetHeadMsg struct {
	BaseMessage
	Hash types.Hash
}

// BlockGetHeadMsg retrieves the head block hash.
type BlockGetHeadMsg struct {
	BaseMessage
}

// BlockSetSafeMsg sets the safe block hash.
type BlockSetSafeMsg struct {
	BaseMessage
	Hash types.Hash
}

// BlockGetSafeMsg retrieves the safe block hash.
type BlockGetSafeMsg struct {
	BaseMessage
}

// BlockSetFinalMsg sets the finalized block hash.
type BlockSetFinalMsg struct {
	BaseMessage
	Hash types.Hash
}

// BlockGetFinalMsg retrieves the finalized block hash.
type BlockGetFinalMsg struct {
	BaseMessage
}

// BlockEvictOldMsg evicts blocks older than 64 behind head.
type BlockEvictOldMsg struct {
	BaseMessage
}

// BlockCountMsg returns the number of stored blocks.
type BlockCountMsg struct {
	BaseMessage
}

// --- Actor implementation ---

// BlockStoreActor manages blocks and block access lists.
type BlockStoreActor struct {
	blocks      map[types.Hash]*types.Block
	bals        map[types.Hash]*bal.BlockAccessList
	numberIndex map[uint64]types.Hash
	headHash    types.Hash
	safeHash    types.Hash
	finalHash   types.Hash

	inbox chan any

	// Statistics.
	storeCount   atomic.Uint64
	evictCount   atomic.Uint64
	requestCount atomic.Uint64
}

// NewBlockStoreActor creates a new block store actor.
func NewBlockStoreActor() *BlockStoreActor {
	return &BlockStoreActor{
		blocks:      make(map[types.Hash]*types.Block),
		bals:        make(map[types.Hash]*bal.BlockAccessList),
		numberIndex: make(map[uint64]types.Hash),
		inbox:       make(chan any, 128),
	}
}

// Run implements Actor.
func (a *BlockStoreActor) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-a.inbox:
			a.handleMessage(msg)
		}
	}
}

// Inbox returns the actor's message channel.
func (a *BlockStoreActor) Inbox() chan<- any {
	return a.inbox
}

func (a *BlockStoreActor) handleMessage(msg any) {
	switch m := msg.(type) {
	case *BlockStoreMsg:
		a.store(m.Hash, m.Block, m.BAL)
		m.Reply() <- Reply{}

	case *BlockGetByHashMsg:
		blk := a.getByHash(m.Hash)
		m.Reply() <- Reply{Result: blk}

	case *BlockGetByNumberMsg:
		blk := a.getByNumber(m.Number)
		m.Reply() <- Reply{Result: blk}

	case *BlockGetBALMsg:
		bal := a.getBAL(m.Hash)
		m.Reply() <- Reply{Result: bal}

	case *BlockGetBodiesByHashMsg:
		bodies := a.getBodiesByHash(m.Hashes)
		m.Reply() <- Reply{Result: bodies}

	case *BlockGetBodiesByRangeMsg:
		bodies := a.getBodiesByRange(m.Start, m.Count)
		m.Reply() <- Reply{Result: bodies}

	case *BlockSetHeadMsg:
		a.headHash = m.Hash
		m.Reply() <- Reply{}

	case *BlockGetHeadMsg:
		m.Reply() <- Reply{Result: a.headHash}

	case *BlockSetSafeMsg:
		a.safeHash = m.Hash
		m.Reply() <- Reply{}

	case *BlockGetSafeMsg:
		m.Reply() <- Reply{Result: a.safeHash}

	case *BlockSetFinalMsg:
		a.finalHash = m.Hash
		m.Reply() <- Reply{}

	case *BlockGetFinalMsg:
		m.Reply() <- Reply{Result: a.finalHash}

	case *BlockEvictOldMsg:
		a.evictOld()
		m.Reply() <- Reply{}

	case *BlockCountMsg:
		m.Reply() <- Reply{Result: len(a.blocks)}
	}
}

func (a *BlockStoreActor) store(hash types.Hash, blk *types.Block, b *bal.BlockAccessList) {
	if blk == nil {
		return
	}
	a.blocks[hash] = blk
	a.numberIndex[blk.NumberU64()] = hash
	if b != nil {
		a.bals[hash] = b
	}
	a.storeCount.Add(1)
}

func (a *BlockStoreActor) getByHash(hash types.Hash) *types.Block {
	a.requestCount.Add(1)
	return a.blocks[hash]
}

func (a *BlockStoreActor) getByNumber(num uint64) *types.Block {
	a.requestCount.Add(1)
	hash, ok := a.numberIndex[num]
	if !ok {
		return nil
	}
	return a.blocks[hash]
}

func (a *BlockStoreActor) getBAL(hash types.Hash) *bal.BlockAccessList {
	return a.bals[hash]
}

func (a *BlockStoreActor) getBodiesByHash(hashes []types.Hash) []*BlockBody {
	a.requestCount.Add(1)
	bodies := make([]*BlockBody, len(hashes))
	for i, hash := range hashes {
		blk := a.blocks[hash]
		if blk == nil {
			bodies[i] = nil
			continue
		}
		bal := a.bals[hash]
		bodies[i] = &BlockBody{
			Transactions: blk.Transactions(),
			BAL:          bal,
		}
	}
	return bodies
}

func (a *BlockStoreActor) getBodiesByRange(start, count uint64) []*BlockBody {
	a.requestCount.Add(1)
	bodies := make([]*BlockBody, 0, count)
	for i := uint64(0); i < count; i++ {
		num := start + i
		hash, ok := a.numberIndex[num]
		if !ok {
			bodies = append(bodies, nil)
			continue
		}
		blk := a.blocks[hash]
		if blk == nil {
			bodies = append(bodies, nil)
			continue
		}
		bal := a.bals[hash]
		bodies = append(bodies, &BlockBody{
			Transactions: blk.Transactions(),
			BAL:          bal,
		})
	}
	return bodies
}

func (a *BlockStoreActor) evictOld() {
	head := a.blocks[a.headHash]
	if head == nil || head.NumberU64() < 64 {
		return
	}
	cutoff := head.NumberU64() - 64
	evicted := 0
	for hash, blk := range a.blocks {
		if blk.NumberU64() < cutoff {
			delete(a.numberIndex, blk.NumberU64())
			delete(a.blocks, hash)
			delete(a.bals, hash)
			evicted++
		}
	}
	a.evictCount.Add(uint64(evicted))
}

// BlockBody represents a block body without the header.
type BlockBody struct {
	Transactions []*types.Transaction
	BAL          *bal.BlockAccessList
}
