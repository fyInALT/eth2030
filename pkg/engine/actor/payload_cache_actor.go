// payload_cache_actor.go provides an actor-based payload cache for EngineBackend.
// It manages pending built payloads with LRU eviction.
package actor

import (
	"context"
	"math/big"
	"sync/atomic"

	"github.com/eth2030/eth2030/bal"
	"github.com/eth2030/eth2030/core/types"
)

// NOTE: The types below are defined here for the actor package's internal use.
// They are binary-compatible with engine/payload types ([8]byte and structs with same fields).
// When using from EngineBackend, cast between types as needed.

// PayloadID is an 8-byte identifier for a built payload.
// Binary compatible with payload.PayloadID ([8]byte).
type PayloadID [8]byte

// Withdrawal represents a validator withdrawal for the payload.
// Binary compatible with payload.Withdrawal (same field layout).
type Withdrawal struct {
	Index          uint64
	ValidatorIndex uint64
	Address        types.Address
	Amount         uint64
}

// PendingPayload holds a payload being built by the block builder.
type PendingPayload struct {
	Block        *types.Block
	Receipts     []*types.Receipt
	Bal          *bal.BlockAccessList // EIP-7928
	BlockValue   *big.Int
	ParentHash   types.Hash
	Timestamp    uint64
	FeeRecipient types.Address
	PrevRandao   types.Hash
	Withdrawals  []*Withdrawal
}

// ExecutionPayloadV4 represents an execution payload (Prague).
// Simplified version for internal cache use.
type ExecutionPayloadV4 struct {
	ParentHash    types.Hash
	FeeRecipient  types.Address
	StateRoot     types.Hash
	ReceiptsRoot  types.Hash
	BlockNumber   uint64
	GasLimit      uint64
	GasUsed       uint64
	Timestamp     uint64
	BaseFeePerGas uint64
	BlockHash     types.Hash
	Transactions  [][]byte
	Withdrawals   []Withdrawal
}

// ExecutionPayloadV6 represents an execution payload with additional fields.
type ExecutionPayloadV6 struct {
	ExecutionPayloadV4
}

// --- Message types ---

// PayloadCacheStoreMsg stores a pending payload.
type PayloadCacheStoreMsg struct {
	BaseMessage
	ID      PayloadID
	Payload *PendingPayload
}

// PayloadCacheGetMsg retrieves a pending payload.
type PayloadCacheGetMsg struct {
	BaseMessage
	ID PayloadID
}

// PayloadCacheGetV4Msg retrieves a payload as ExecutionPayloadV4.
type PayloadCacheGetV4Msg struct {
	BaseMessage
	ID PayloadID
}

// PayloadCacheGetV6Msg retrieves a payload as ExecutionPayloadV6.
type PayloadCacheGetV6Msg struct {
	BaseMessage
	ID PayloadID
}

// PayloadCacheRemoveMsg removes a pending payload.
type PayloadCacheRemoveMsg struct {
	BaseMessage
	ID PayloadID
}

// PayloadCacheCountMsg returns the number of stored payloads.
type PayloadCacheCountMsg struct {
	BaseMessage
}

// --- Actor implementation ---

// PayloadCacheActor manages pending built payloads with LRU eviction.
type PayloadCacheActor struct {
	maxPayloads int
	payloads    map[PayloadID]*PendingPayload
	order       []PayloadID // insertion order for LRU eviction

	inbox chan any

	// Statistics.
	storeCount   atomic.Uint64
	evictCount   atomic.Uint64
	requestCount atomic.Uint64
}

// NewPayloadCacheActor creates a new payload cache actor.
func NewPayloadCacheActor(maxPayloads int) *PayloadCacheActor {
	if maxPayloads <= 0 {
		maxPayloads = 32
	}
	return &PayloadCacheActor{
		maxPayloads: maxPayloads,
		payloads:    make(map[PayloadID]*PendingPayload),
		order:       make([]PayloadID, 0, maxPayloads),
		inbox:       make(chan any, 64),
	}
}

// Run implements Actor.
func (a *PayloadCacheActor) Run(ctx context.Context) {
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
func (a *PayloadCacheActor) Inbox() chan<- any {
	return a.inbox
}

func (a *PayloadCacheActor) handleMessage(msg any) {
	switch m := msg.(type) {
	case *PayloadCacheStoreMsg:
		a.store(m.ID, m.Payload)
		m.Reply() <- Reply{}

	case *PayloadCacheGetMsg:
		p := a.get(m.ID)
		m.Reply() <- Reply{Result: p}

	case *PayloadCacheGetV4Msg:
		v4 := a.getV4(m.ID)
		m.Reply() <- Reply{Result: v4}

	case *PayloadCacheGetV6Msg:
		v6 := a.getV6(m.ID)
		m.Reply() <- Reply{Result: v6}

	case *PayloadCacheRemoveMsg:
		a.remove(m.ID)
		m.Reply() <- Reply{}

	case *PayloadCacheCountMsg:
		m.Reply() <- Reply{Result: len(a.payloads)}
	}
}

func (a *PayloadCacheActor) store(id PayloadID, p *PendingPayload) {
	// Evict if at capacity.
	for len(a.payloads) >= a.maxPayloads {
		a.evictOldest()
	}

	// Check if already exists.
	if _, exists := a.payloads[id]; exists {
		// Update existing entry.
		a.payloads[id] = p
		return
	}

	// Add new entry.
	a.payloads[id] = p
	a.order = append(a.order, id)
	a.storeCount.Add(1)
}

func (a *PayloadCacheActor) get(id PayloadID) *PendingPayload {
	a.requestCount.Add(1)
	return a.payloads[id]
}

func (a *PayloadCacheActor) getV4(id PayloadID) *ExecutionPayloadV4 {
	p := a.get(id)
	if p == nil {
		return nil
	}
	return pendingToV4(p)
}

func (a *PayloadCacheActor) getV6(id PayloadID) *ExecutionPayloadV6 {
	p := a.get(id)
	if p == nil {
		return nil
	}
	v4 := pendingToV4(p)
	return &ExecutionPayloadV6{ExecutionPayloadV4: *v4}
}

func (a *PayloadCacheActor) remove(id PayloadID) {
	if _, exists := a.payloads[id]; !exists {
		return
	}
	delete(a.payloads, id)
	// Remove from order slice.
	for i, oid := range a.order {
		if oid == id {
			a.order = append(a.order[:i], a.order[i+1:]...)
			break
		}
	}
}

func (a *PayloadCacheActor) evictOldest() {
	if len(a.order) == 0 {
		return
	}
	oldest := a.order[0]
	a.order = a.order[1:]
	delete(a.payloads, oldest)
	a.evictCount.Add(1)
}

func pendingToV4(p *PendingPayload) *ExecutionPayloadV4 {
	if p == nil || p.Block == nil {
		return nil
	}
	header := p.Block.Header()
	txs := make([][]byte, len(p.Block.Transactions()))
	for i, tx := range p.Block.Transactions() {
		data, _ := tx.EncodeRLP()
		txs[i] = data
	}

	withdrawals := make([]Withdrawal, len(p.Withdrawals))
	for i, w := range p.Withdrawals {
		withdrawals[i] = *w
	}

	return &ExecutionPayloadV4{
		ParentHash:    header.ParentHash,
		FeeRecipient:  header.Coinbase,
		StateRoot:     header.Root,
		ReceiptsRoot:  header.ReceiptHash,
		BlockNumber:   header.Number.Uint64(),
		GasLimit:      header.GasLimit,
		GasUsed:       header.GasUsed,
		Timestamp:     header.Time,
		BaseFeePerGas: header.BaseFee.Uint64(),
		BlockHash:     p.Block.Hash(),
		Transactions:  txs,
		Withdrawals:   withdrawals,
	}
}
