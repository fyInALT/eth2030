package peertick

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func TestPeerTickCache_MarkAndCheck(t *testing.T) {
	c := NewPeerTickCache(2)

	var hashes [5]types.Hash
	for i := range hashes {
		hashes[i][0] = byte(i + 1)
	}

	for _, h := range hashes {
		c.MarkPeerValidated(h, "peer-1", 10)
	}

	for i, h := range hashes {
		if !c.IsPeerValidated(h) {
			t.Errorf("hash %d should be peer-validated", i)
		}
	}
}

func TestPeerTickCache_NotPresent(t *testing.T) {
	c := NewPeerTickCache(2)

	var h types.Hash
	h[0] = 0xFF
	if c.IsPeerValidated(h) {
		t.Error("unmarked hash should not be peer-validated")
	}
}

func TestPeerTickCache_SlotEviction(t *testing.T) {
	c := NewPeerTickCache(2)

	var h types.Hash
	h[0] = 0x01
	c.MarkPeerValidated(h, "peer-1", 10)

	if !c.IsPeerValidated(h) {
		t.Fatal("expected hash to be present before eviction")
	}

	// Advance slot by TTL+1: slot 10 + TTL 2 + 1 = 13.
	evicted := c.AdvanceSlot(13)
	if evicted != 1 {
		t.Errorf("expected 1 eviction, got %d", evicted)
	}
	if c.Size() != 0 {
		t.Errorf("expected size 0 after eviction, got %d", c.Size())
	}
}

func TestPeerTickCache_WithinTTL(t *testing.T) {
	c := NewPeerTickCache(2)

	var h types.Hash
	h[0] = 0x01
	c.MarkPeerValidated(h, "peer-1", 10)

	// Advance slot by TTL-1: slot 10 + 2 - 1 = 11.
	evicted := c.AdvanceSlot(11)
	if evicted != 0 {
		t.Errorf("expected 0 evictions within TTL, got %d", evicted)
	}
	if !c.IsPeerValidated(h) {
		t.Error("hash should survive within TTL")
	}
}

func TestPeerTickCache_Size(t *testing.T) {
	c := NewPeerTickCache(2)

	if c.Size() != 0 {
		t.Errorf("expected initial size 0, got %d", c.Size())
	}

	for i := 0; i < 10; i++ {
		var h types.Hash
		h[0] = byte(i)
		c.MarkPeerValidated(h, "peer-1", 5)
	}

	if c.Size() != 10 {
		t.Errorf("expected size 10, got %d", c.Size())
	}
}
