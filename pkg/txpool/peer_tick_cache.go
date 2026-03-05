package txpool

import (
	"sync"

	"github.com/eth2030/eth2030/core/types"
)

// PeerTickCache tracks txs validated by remote peers via STARK ticks.
// Entries expire after slotTTL slots.
type PeerTickCache struct {
	mu          sync.RWMutex
	entries     map[types.Hash]*peerValidatedEntry
	slotTTL     uint64
	currentSlot uint64
}

type peerValidatedEntry struct {
	peerID      string
	validatedAt uint64 // slot when marked
}

// NewPeerTickCache creates a new peer tick cache with the given slot TTL.
func NewPeerTickCache(slotTTL uint64) *PeerTickCache {
	if slotTTL == 0 {
		slotTTL = 2
	}
	return &PeerTickCache{
		entries: make(map[types.Hash]*peerValidatedEntry),
		slotTTL: slotTTL,
	}
}

// MarkPeerValidated records that a tx hash was validated by a remote peer at the given slot.
func (c *PeerTickCache) MarkPeerValidated(txHash types.Hash, peerID string, slot uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[txHash] = &peerValidatedEntry{
		peerID:      peerID,
		validatedAt: slot,
	}
}

// IsPeerValidated returns whether a tx hash has been validated by a remote peer.
func (c *PeerTickCache) IsPeerValidated(txHash types.Hash) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.entries[txHash]
	return ok
}

// AdvanceSlot advances the current slot and evicts expired entries.
// Returns the number of entries evicted.
func (c *PeerTickCache) AdvanceSlot(newSlot uint64) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentSlot = newSlot
	evicted := 0
	for h, e := range c.entries {
		if newSlot > e.validatedAt+c.slotTTL {
			delete(c.entries, h)
			evicted++
		}
	}
	return evicted
}

// Size returns the number of entries in the cache.
func (c *PeerTickCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
