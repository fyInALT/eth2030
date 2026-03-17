// inclusion_list_actor.go provides an actor-based inclusion list store for EngineBackend.
// It manages inclusion lists received via engine_newInclusionListV1.
package actor

import (
	"context"
	"sync/atomic"

	"github.com/eth2030/eth2030/core/types"
)

// --- Message types ---

// ILStoreMsg stores an inclusion list.
type ILStoreMsg struct {
	BaseMessage
	IL *types.InclusionList
}

// ILGetAllMsg retrieves all stored inclusion lists.
type ILGetAllMsg struct {
	BaseMessage
}

// ILClearMsg clears all inclusion lists.
type ILClearMsg struct {
	BaseMessage
}

// ILCountMsg returns the number of stored inclusion lists.
type ILCountMsg struct {
	BaseMessage
}

// --- Actor implementation ---

// InclusionListActor manages inclusion lists for a slot.
type InclusionListActor struct {
	maxILs int
	ils    []*types.InclusionList

	inbox chan any

	// Statistics.
	storeCount atomic.Uint64
}

// NewInclusionListActor creates a new inclusion list actor.
func NewInclusionListActor(maxILs int) *InclusionListActor {
	if maxILs <= 0 {
		maxILs = 256
	}
	return &InclusionListActor{
		maxILs: maxILs,
		ils:    make([]*types.InclusionList, 0, maxILs),
		inbox:  make(chan any, 32),
	}
}

// Run implements Actor.
func (a *InclusionListActor) Run(ctx context.Context) {
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
func (a *InclusionListActor) Inbox() chan<- any {
	return a.inbox
}

func (a *InclusionListActor) handleMessage(msg any) {
	switch m := msg.(type) {
	case *ILStoreMsg:
		a.store(m.IL)
		m.Reply() <- Reply{}

	case *ILGetAllMsg:
		m.Reply() <- Reply{Result: a.getAll()}

	case *ILClearMsg:
		a.clear()
		m.Reply() <- Reply{}

	case *ILCountMsg:
		m.Reply() <- Reply{Result: len(a.ils)}
	}
}

func (a *InclusionListActor) store(il *types.InclusionList) {
	if il == nil {
		return
	}
	// Enforce capacity by evicting oldest.
	for len(a.ils) >= a.maxILs {
		a.ils = a.ils[1:]
	}
	a.ils = append(a.ils, il)
	a.storeCount.Add(1)
}

func (a *InclusionListActor) getAll() []*types.InclusionList {
	// Return a copy to prevent mutation.
	result := make([]*types.InclusionList, len(a.ils))
	copy(result, a.ils)
	return result
}

func (a *InclusionListActor) clear() {
	a.ils = a.ils[:0]
}
