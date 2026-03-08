package txpool

import (
	"crypto/sha256"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// computeTxMerkleRoot computes a simple binary Merkle root of transaction hashes.
func computeTxMerkleRoot(hashes []types.Hash) types.Hash {
	if len(hashes) == 0 {
		return types.Hash{}
	}
	if len(hashes) == 1 {
		return hashes[0]
	}
	layer := make([]types.Hash, len(hashes))
	copy(layer, hashes)
	for len(layer) > 1 {
		var next []types.Hash
		for i := 0; i < len(layer); i += 2 {
			h := sha256.New()
			h.Write(layer[i][:])
			if i+1 < len(layer) {
				h.Write(layer[i+1][:])
			} else {
				h.Write(layer[i][:])
			}
			var hash types.Hash
			copy(hash[:], h.Sum(nil))
			next = append(next, hash)
		}
		layer = next
	}
	return layer[0]
}

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

func TestMergeTickAtSlot(t *testing.T) {
	agg := NewSTARKAggregator("node-1")

	// Add some local txs first so verification works.
	var h1, h2, h3 types.Hash
	h1[0] = 0x01
	h2[0] = 0x02
	h3[0] = 0x03
	agg.AddValidatedTx(h1, []byte("proof1"), 21000)

	// Generate a local tick to get a valid aggregate proof.
	localTick, err := agg.GenerateTick()
	if err != nil {
		t.Fatalf("GenerateTick: %v", err)
	}

	// Build a remote tick reusing the valid proof.
	remote := &MempoolAggregationTick{
		Timestamp:      localTick.Timestamp,
		ValidTxHashes:  []types.Hash{h1, h2, h3},
		AggregateProof: localTick.AggregateProof,
		PeerID:         "peer-2",
		TickNumber:     1,
		ValidBitfield:  []byte{0x07},
		TxMerkleRoot:   computeTxMerkleRoot([]types.Hash{h1, h2, h3}),
	}

	err = agg.MergeTickAtSlot(remote, 100)
	if err != nil {
		t.Fatalf("MergeTickAtSlot: %v", err)
	}

	// All 3 hashes should be in the peer cache.
	cache := agg.PeerCache()
	for _, h := range []types.Hash{h1, h2, h3} {
		if !cache.IsPeerValidated(h) {
			t.Errorf("hash %x should be peer-validated after merge", h[0])
		}
	}
}
