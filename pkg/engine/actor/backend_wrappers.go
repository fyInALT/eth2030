// backend_wrappers.go provides convenience wrappers for using actors from EngineBackend.
// These wrappers handle type conversions and provide a simpler API.
package actor

import (
	"context"
	"time"

	"github.com/eth2030/eth2030/bal"
	"github.com/eth2030/eth2030/core/types"
)

// EngineActors groups all actors used by EngineBackend.
type EngineActors struct {
	BlockStore      *BlockStoreActor
	PayloadCache    *PayloadCacheActor
	InclusionList   *InclusionListActor

	ctx    context.Context
	cancel context.CancelFunc
}

// NewEngineActors creates and starts all actors for EngineBackend.
func NewEngineActors(ctx context.Context, maxPayloads, maxILs int) *EngineActors {
	actorCtx, cancel := context.WithCancel(ctx)
	
	ea := &EngineActors{
		BlockStore:    NewBlockStoreActor(),
		PayloadCache:  NewPayloadCacheActor(maxPayloads),
		InclusionList: NewInclusionListActor(maxILs),
		ctx:           actorCtx,
		cancel:        cancel,
	}
	
	// Start all actors.
	go ea.BlockStore.Run(actorCtx)
	go ea.PayloadCache.Run(actorCtx)
	go ea.InclusionList.Run(actorCtx)
	
	return ea
}

// Stop gracefully stops all actors.
func (ea *EngineActors) Stop() {
	if ea.cancel != nil {
		ea.cancel()
	}
}

// --- BlockStore convenience methods ---

// StoreBlock stores a block with optional BAL.
func (ea *EngineActors) StoreBlock(hash types.Hash, blk *types.Block, b *bal.BlockAccessList, timeout time.Duration) error {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockStoreMsg{
		BaseMessage: msg,
		Hash:        hash,
		Block:       blk,
		BAL:         b,
	}, timeout); err != nil {
		return err
	}
	_, err := CallResult[struct{}](replyCh, timeout)
	return err
}

// GetBlockByHash retrieves a block by hash.
func (ea *EngineActors) GetBlockByHash(hash types.Hash, timeout time.Duration) (*types.Block, error) {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockGetByHashMsg{
		BaseMessage: msg,
		Hash:        hash,
	}, timeout); err != nil {
		return nil, err
	}
	return CallResult[*types.Block](replyCh, timeout)
}

// GetBlockByNumber retrieves a block by number.
func (ea *EngineActors) GetBlockByNumber(num uint64, timeout time.Duration) (*types.Block, error) {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockGetByNumberMsg{
		BaseMessage: msg,
		Number:      num,
	}, timeout); err != nil {
		return nil, err
	}
	return CallResult[*types.Block](replyCh, timeout)
}

// GetBAL retrieves a BAL by block hash.
func (ea *EngineActors) GetBAL(hash types.Hash, timeout time.Duration) (*bal.BlockAccessList, error) {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockGetBALMsg{
		BaseMessage: msg,
		Hash:        hash,
	}, timeout); err != nil {
		return nil, err
	}
	return CallResult[*bal.BlockAccessList](replyCh, timeout)
}

// SetHeadHash sets the head block hash.
func (ea *EngineActors) SetHeadHash(hash types.Hash, timeout time.Duration) error {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockSetHeadMsg{
		BaseMessage: msg,
		Hash:        hash,
	}, timeout); err != nil {
		return err
	}
	_, err := CallResult[struct{}](replyCh, timeout)
	return err
}

// GetHeadHash retrieves the head block hash.
func (ea *EngineActors) GetHeadHash(timeout time.Duration) (types.Hash, error) {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockGetHeadMsg{
		BaseMessage: msg,
	}, timeout); err != nil {
		return types.Hash{}, err
	}
	return CallResult[types.Hash](replyCh, timeout)
}

// SetSafeHash sets the safe block hash.
func (ea *EngineActors) SetSafeHash(hash types.Hash, timeout time.Duration) error {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockSetSafeMsg{
		BaseMessage: msg,
		Hash:        hash,
	}, timeout); err != nil {
		return err
	}
	_, err := CallResult[struct{}](replyCh, timeout)
	return err
}

// GetSafeHash retrieves the safe block hash.
func (ea *EngineActors) GetSafeHash(timeout time.Duration) (types.Hash, error) {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockGetSafeMsg{
		BaseMessage: msg,
	}, timeout); err != nil {
		return types.Hash{}, err
	}
	return CallResult[types.Hash](replyCh, timeout)
}

// SetFinalHash sets the finalized block hash.
func (ea *EngineActors) SetFinalHash(hash types.Hash, timeout time.Duration) error {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockSetFinalMsg{
		BaseMessage: msg,
		Hash:        hash,
	}, timeout); err != nil {
		return err
	}
	_, err := CallResult[struct{}](replyCh, timeout)
	return err
}

// GetFinalHash retrieves the finalized block hash.
func (ea *EngineActors) GetFinalHash(timeout time.Duration) (types.Hash, error) {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockGetFinalMsg{
		BaseMessage: msg,
	}, timeout); err != nil {
		return types.Hash{}, err
	}
	return CallResult[types.Hash](replyCh, timeout)
}

// EvictOldBlocks evicts blocks older than 64 behind head.
func (ea *EngineActors) EvictOldBlocks(timeout time.Duration) error {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockEvictOldMsg{
		BaseMessage: msg,
	}, timeout); err != nil {
		return err
	}
	_, err := CallResult[struct{}](replyCh, timeout)
	return err
}

// GetBodiesByRange retrieves block bodies by number range.
func (ea *EngineActors) GetBodiesByRange(start, count uint64, timeout time.Duration) ([]*BlockBody, error) {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockGetBodiesByRangeMsg{
		BaseMessage: msg,
		Start:       start,
		Count:       count,
	}, timeout); err != nil {
		return nil, err
	}
	return CallResult[[]*BlockBody](replyCh, timeout)
}

// GetBodiesByHash retrieves block bodies by hashes.
func (ea *EngineActors) GetBodiesByHash(hashes []types.Hash, timeout time.Duration) ([]*BlockBody, error) {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.BlockStore.Inbox(), &BlockGetBodiesByHashMsg{
		BaseMessage: msg,
		Hashes:      hashes,
	}, timeout); err != nil {
		return nil, err
	}
	return CallResult[[]*BlockBody](replyCh, timeout)
}

// --- PayloadCache convenience methods ---

// StorePayload stores a pending payload.
func (ea *EngineActors) StorePayload(id [8]byte, p *PendingPayload, timeout time.Duration) error {
	msg, replyCh := NewBaseMessage()
	// Cast [8]byte to PayloadID
	actorID := PayloadID(id)
	if err := Send[any](ea.PayloadCache.Inbox(), &PayloadCacheStoreMsg{
		BaseMessage: msg,
		ID:          actorID,
		Payload:     p,
	}, timeout); err != nil {
		return err
	}
	_, err := CallResult[struct{}](replyCh, timeout)
	return err
}

// GetPayload retrieves a pending payload.
func (ea *EngineActors) GetPayload(id [8]byte, timeout time.Duration) (*PendingPayload, error) {
	msg, replyCh := NewBaseMessage()
	actorID := PayloadID(id)
	if err := Send[any](ea.PayloadCache.Inbox(), &PayloadCacheGetMsg{
		BaseMessage: msg,
		ID:          actorID,
	}, timeout); err != nil {
		return nil, err
	}
	return CallResult[*PendingPayload](replyCh, timeout)
}

// RemovePayload removes a pending payload.
func (ea *EngineActors) RemovePayload(id [8]byte, timeout time.Duration) error {
	msg, replyCh := NewBaseMessage()
	actorID := PayloadID(id)
	if err := Send[any](ea.PayloadCache.Inbox(), &PayloadCacheRemoveMsg{
		BaseMessage: msg,
		ID:          actorID,
	}, timeout); err != nil {
		return err
	}
	_, err := CallResult[struct{}](replyCh, timeout)
	return err
}

// PayloadCount returns the number of stored payloads.
func (ea *EngineActors) PayloadCount(timeout time.Duration) (int, error) {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.PayloadCache.Inbox(), &PayloadCacheCountMsg{
		BaseMessage: msg,
	}, timeout); err != nil {
		return 0, err
	}
	return CallResult[int](replyCh, timeout)
}

// --- InclusionList convenience methods ---

// StoreInclusionList stores an inclusion list.
func (ea *EngineActors) StoreInclusionList(il *types.InclusionList, timeout time.Duration) error {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.InclusionList.Inbox(), &ILStoreMsg{
		BaseMessage: msg,
		IL:          il,
	}, timeout); err != nil {
		return err
	}
	_, err := CallResult[struct{}](replyCh, timeout)
	return err
}

// GetAllInclusionLists retrieves all stored inclusion lists.
func (ea *EngineActors) GetAllInclusionLists(timeout time.Duration) ([]*types.InclusionList, error) {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.InclusionList.Inbox(), &ILGetAllMsg{
		BaseMessage: msg,
	}, timeout); err != nil {
		return nil, err
	}
	return CallResult[[]*types.InclusionList](replyCh, timeout)
}

// ClearInclusionLists clears all inclusion lists.
func (ea *EngineActors) ClearInclusionLists(timeout time.Duration) error {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.InclusionList.Inbox(), &ILClearMsg{
		BaseMessage: msg,
	}, timeout); err != nil {
		return err
	}
	_, err := CallResult[struct{}](replyCh, timeout)
	return err
}

// InclusionListCount returns the number of stored inclusion lists.
func (ea *EngineActors) InclusionListCount(timeout time.Duration) (int, error) {
	msg, replyCh := NewBaseMessage()
	if err := Send[any](ea.InclusionList.Inbox(), &ILCountMsg{
		BaseMessage: msg,
	}, timeout); err != nil {
		return 0, err
	}
	return CallResult[int](replyCh, timeout)
}
