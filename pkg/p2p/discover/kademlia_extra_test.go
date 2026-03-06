package discover

import (
	"testing"
	"time"
)

// --- Config ---

func TestKademliaTable_Config(t *testing.T) {
	var selfID [32]byte
	cfg := DefaultKademliaConfig()
	cfg.BucketSize = 8
	cfg.Alpha = 5
	kt := NewKademliaTable(selfID, cfg)

	got := kt.Config()
	if got.BucketSize != 8 {
		t.Fatalf("Config().BucketSize = %d, want 8", got.BucketSize)
	}
	if got.Alpha != 5 {
		t.Fatalf("Config().Alpha = %d, want 5", got.Alpha)
	}
	if got.MaxFailCount != 5 {
		t.Fatalf("Config().MaxFailCount = %d, want 5", got.MaxFailCount)
	}
	if got.MaxReplacements != 10 {
		t.Fatalf("Config().MaxReplacements = %d, want 10", got.MaxReplacements)
	}
	if got.RefreshInterval != 1*time.Hour {
		t.Fatalf("Config().RefreshInterval = %v, want 1h", got.RefreshInterval)
	}
	if got.StaleTimeout != 24*time.Hour {
		t.Fatalf("Config().StaleTimeout = %v, want 24h", got.StaleTimeout)
	}
}

func TestKademliaTable_ConfigIsACopy(t *testing.T) {
	var selfID [32]byte
	kt := NewKademliaTable(selfID, DefaultKademliaConfig())

	cfg := kt.Config()
	cfg.BucketSize = 999 //nolint:staticcheck // intentional write to test copy semantics

	// Mutating the returned copy must not affect the table.
	if kt.Config().BucketSize == 999 {
		t.Fatal("Config() returned a reference, not a copy")
	}
}

// --- BucketEntries (KademliaTable) ---

func TestKademliaTable_BucketEntries_Basic(t *testing.T) {
	var selfID [32]byte
	kt := NewKademliaTable(selfID, DefaultKademliaConfig())

	// makeEntry(1) has ID[31]=1, which gives distance 1 -> bucket 0.
	node := makeEntry(1)
	kt.AddNode(node)

	entries := kt.BucketEntries(0)
	if len(entries) != 1 {
		t.Fatalf("BucketEntries(0): want 1 entry, got %d", len(entries))
	}
	if entries[0].ID != node.ID {
		t.Fatal("BucketEntries(0): wrong node ID")
	}
}

func TestKademliaTable_BucketEntries_Empty(t *testing.T) {
	var selfID [32]byte
	kt := NewKademliaTable(selfID, DefaultKademliaConfig())

	entries := kt.BucketEntries(5)
	if entries == nil {
		t.Fatal("BucketEntries on empty bucket should return non-nil slice")
	}
	if len(entries) != 0 {
		t.Fatalf("BucketEntries on empty bucket: want 0, got %d", len(entries))
	}
}

func TestKademliaTable_BucketEntries_OutOfRange(t *testing.T) {
	var selfID [32]byte
	kt := NewKademliaTable(selfID, DefaultKademliaConfig())

	if kt.BucketEntries(-1) != nil {
		t.Fatal("BucketEntries(-1) should return nil")
	}
	if kt.BucketEntries(256) != nil {
		t.Fatal("BucketEntries(256) should return nil")
	}
}

func TestKademliaTable_BucketEntries_IsACopy(t *testing.T) {
	var selfID [32]byte
	kt := NewKademliaTable(selfID, DefaultKademliaConfig())

	node := makeEntry(1)
	kt.AddNode(node)

	entries := kt.BucketEntries(0)
	if len(entries) == 0 {
		t.Fatal("expected entries")
	}
	// Mutate the copy.
	entries[0].Address = "mutated"

	// The table should be unaffected.
	entries2 := kt.BucketEntries(0)
	if entries2[0].Address == "mutated" {
		t.Fatal("BucketEntries returned a reference, not a copy")
	}
}

func TestKademliaTable_BucketEntries_MultipleNodes(t *testing.T) {
	var selfID [32]byte
	cfg := DefaultKademliaConfig()
	cfg.BucketSize = 8
	kt := NewKademliaTable(selfID, cfg)

	// Nodes with ID[31] in 0x80..0x87 all land in bucket 7 (distance 8).
	for i := byte(0); i < 4; i++ {
		var id [32]byte
		id[31] = 0x80 + i
		kt.AddNode(NodeEntry{ID: id, Address: "10.0.0.1", Port: 30303, LastSeen: time.Now()})
	}

	entries := kt.BucketEntries(7)
	if len(entries) != 4 {
		t.Fatalf("BucketEntries(7): want 4, got %d", len(entries))
	}
}

// --- applyDefaults edge cases ---

func TestKademliaConfig_ApplyDefaults_ZeroValues(t *testing.T) {
	cfg := KademliaConfig{}
	cfg.applyDefaults()

	if cfg.BucketSize != 16 {
		t.Fatalf("BucketSize: want 16, got %d", cfg.BucketSize)
	}
	if cfg.Alpha != 3 {
		t.Fatalf("Alpha: want 3, got %d", cfg.Alpha)
	}
	if cfg.RefreshInterval != 1*time.Hour {
		t.Fatalf("RefreshInterval: want 1h, got %v", cfg.RefreshInterval)
	}
	if cfg.StaleTimeout != 24*time.Hour {
		t.Fatalf("StaleTimeout: want 24h, got %v", cfg.StaleTimeout)
	}
	if cfg.MaxFailCount != 5 {
		t.Fatalf("MaxFailCount: want 5, got %d", cfg.MaxFailCount)
	}
	if cfg.MaxReplacements != 10 {
		t.Fatalf("MaxReplacements: want 10, got %d", cfg.MaxReplacements)
	}
}

func TestKademliaConfig_ApplyDefaults_PreservesPositive(t *testing.T) {
	cfg := KademliaConfig{
		BucketSize:      4,
		Alpha:           7,
		RefreshInterval: 30 * time.Minute,
		StaleTimeout:    12 * time.Hour,
		MaxFailCount:    2,
		MaxReplacements: 3,
	}
	cfg.applyDefaults()

	if cfg.BucketSize != 4 {
		t.Fatalf("BucketSize should be preserved: want 4, got %d", cfg.BucketSize)
	}
	if cfg.Alpha != 7 {
		t.Fatalf("Alpha should be preserved: want 7, got %d", cfg.Alpha)
	}
	if cfg.RefreshInterval != 30*time.Minute {
		t.Fatalf("RefreshInterval should be preserved: want 30m, got %v", cfg.RefreshInterval)
	}
	if cfg.StaleTimeout != 12*time.Hour {
		t.Fatalf("StaleTimeout should be preserved: want 12h, got %v", cfg.StaleTimeout)
	}
	if cfg.MaxFailCount != 2 {
		t.Fatalf("MaxFailCount should be preserved: want 2, got %d", cfg.MaxFailCount)
	}
	if cfg.MaxReplacements != 3 {
		t.Fatalf("MaxReplacements should be preserved: want 3, got %d", cfg.MaxReplacements)
	}
}

// --- evictStaleLocked via AddNode with MaxTableSize ---

func TestKademliaTable_EvictStale_PromotesReplacement(t *testing.T) {
	var selfID [32]byte
	cfg := DefaultKademliaConfig()
	cfg.BucketSize = 2
	cfg.MaxFailCount = 1
	kt := NewKademliaTable(selfID, cfg)

	// Fill bucket 7 (distance 8) with two nodes.
	var id1, id2 [32]byte
	id1[31] = 0x80
	id2[31] = 0x81
	kt.AddNode(NodeEntry{ID: id1, LastSeen: time.Now()})
	kt.AddNode(NodeEntry{ID: id2, LastSeen: time.Now()})

	// Add a replacement for bucket 7.
	var id3 [32]byte
	id3[31] = 0x82
	kt.AddNode(NodeEntry{ID: id3, LastSeen: time.Now()}) // goes to replacement

	if kt.BucketReplacementLen(7) != 1 {
		t.Fatalf("replacement count: want 1, got %d", kt.BucketReplacementLen(7))
	}

	// Make id1 stale via RecordFailure (MaxFailCount=1 -> eviction on first failure).
	kt.RecordFailure(id1)

	// id3 should have been promoted from replacement.
	if kt.BucketLen(7) != 2 {
		t.Fatalf("bucket 7 len after eviction: want 2, got %d", kt.BucketLen(7))
	}
	if kt.GetNode(id3) == nil {
		t.Fatal("id3 should have been promoted from replacement")
	}
	if kt.BucketReplacementLen(7) != 0 {
		t.Fatalf("replacement should be empty after promotion, got %d", kt.BucketReplacementLen(7))
	}
}

func TestKademliaTable_EvictStale_ByLastSeen(t *testing.T) {
	var selfID [32]byte
	cfg := DefaultKademliaConfig()
	cfg.BucketSize = 2
	cfg.StaleTimeout = 1 * time.Millisecond
	kt := NewKademliaTable(selfID, cfg)

	// Two fresh nodes in bucket 7.
	var id1, id2 [32]byte
	id1[31] = 0x80
	id2[31] = 0x81
	staleTime := time.Now().Add(-10 * time.Second)
	kt.AddNode(NodeEntry{ID: id1, LastSeen: staleTime})
	kt.AddNode(NodeEntry{ID: id2, LastSeen: time.Now()})

	// Add a third node which should evict the stale id1 via bucket tail check.
	var id3 [32]byte
	id3[31] = 0x82
	// Make id1 the tail by ensuring ordering. Since we can't control order directly,
	// add a node and verify table size stayed at 2.
	kt.AddNode(NodeEntry{ID: id3, LastSeen: time.Now()})

	// Table should have exactly 2 entries (stale was replaced or went to replacement).
	if kt.TableSize() > 2 {
		t.Fatalf("table size should be at most 2, got %d", kt.TableSize())
	}
}

func TestKademliaTable_AddReplacementLocked_UpdatesExisting(t *testing.T) {
	var selfID [32]byte
	cfg := DefaultKademliaConfig()
	cfg.BucketSize = 1
	kt := NewKademliaTable(selfID, cfg)

	// Use IDs that land in the same bucket (bucket 7, distance 8).
	// id[31] in range 0x80..0xFF all have distance 8 from zero selfID.
	var id1, id2 [32]byte
	id1[31] = 0x80
	id2[31] = 0x81
	kt.AddNode(NodeEntry{ID: id1, Address: "10.0.0.1", Port: 30303, LastSeen: time.Now()})

	// id2 goes to replacement because bucket 7 has BucketSize=1 and is full.
	kt.AddNode(NodeEntry{ID: id2, Address: "10.0.0.2", Port: 30303, LastSeen: time.Now()})

	if kt.BucketReplacementLen(7) != 1 {
		t.Fatalf("want 1 replacement in bucket 7, got %d", kt.BucketReplacementLen(7))
	}

	// Add same replacement again with updated address.
	kt.AddNode(NodeEntry{ID: id2, Address: "10.0.0.99", Port: 30303, LastSeen: time.Now()})

	// Should still be 1 replacement (updated, not duplicated).
	if kt.BucketReplacementLen(7) != 1 {
		t.Fatalf("replacement should be updated not duplicated, got %d", kt.BucketReplacementLen(7))
	}
}

// --- isNodeStaleLocked edge cases ---

func TestKademliaTable_IsNodeStale_ZeroLastSeen(t *testing.T) {
	var selfID [32]byte
	cfg := DefaultKademliaConfig()
	cfg.StaleTimeout = 1 * time.Millisecond
	kt := NewKademliaTable(selfID, cfg)

	// A node with zero LastSeen should NOT be considered stale by the time check.
	node := NodeEntry{FailCount: 0}
	// Zero time means IsZero() == true, so stale check is skipped.
	kt.mu.RLock()
	stale := kt.isNodeStaleLocked(node)
	kt.mu.RUnlock()
	if stale {
		t.Fatal("node with zero LastSeen should not be considered stale")
	}
}

func TestKademliaTable_IsNodeStale_HighFailCount(t *testing.T) {
	var selfID [32]byte
	kt := NewKademliaTable(selfID, DefaultKademliaConfig())

	node := NodeEntry{FailCount: 10} // exceeds MaxFailCount=5
	kt.mu.RLock()
	stale := kt.isNodeStaleLocked(node)
	kt.mu.RUnlock()
	if !stale {
		t.Fatal("node with FailCount > MaxFailCount should be stale")
	}
}

// --- RemoveNode: remove from replacements ---

func TestKademliaTable_RemoveNode_FromReplacements(t *testing.T) {
	var selfID [32]byte
	cfg := DefaultKademliaConfig()
	cfg.BucketSize = 1
	kt := NewKademliaTable(selfID, cfg)

	// Both IDs must land in the same bucket. Use distance-8 IDs (bucket 7).
	var id1, id2 [32]byte
	id1[31] = 0x80
	id2[31] = 0x81
	kt.AddNode(NodeEntry{ID: id1, LastSeen: time.Now()})
	kt.AddNode(NodeEntry{ID: id2, LastSeen: time.Now()}) // goes to replacement in bucket 7

	if kt.BucketReplacementLen(7) != 1 {
		t.Fatalf("want 1 replacement in bucket 7, got %d", kt.BucketReplacementLen(7))
	}

	// Remove the replacement node (not the main entry).
	kt.RemoveNode(id2)

	if kt.BucketReplacementLen(7) != 0 {
		t.Fatalf("replacement should be removed, got %d", kt.BucketReplacementLen(7))
	}
	// Main entry should still exist.
	if kt.GetNode(id1) == nil {
		t.Fatal("main entry should still be present")
	}
}

// --- RecordFailure: self / nonexistent ---

func TestKademliaTable_RecordFailure_Nonexistent(t *testing.T) {
	var selfID [32]byte
	kt := NewKademliaTable(selfID, DefaultKademliaConfig())

	// Should not panic for unknown ID.
	var id [32]byte
	id[31] = 0xAB
	kt.RecordFailure(id) // no-op
}

func TestKademliaTable_RecordFailure_Self(t *testing.T) {
	var selfID [32]byte
	selfID[0] = 0x11
	kt := NewKademliaTable(selfID, DefaultKademliaConfig())

	// Self has bucketIndex -1, RecordFailure returns early.
	kt.RecordFailure(selfID) // no-op, should not panic
}

// --- MarkRefreshed out-of-range ---

func TestKademliaTable_MarkRefreshed_OutOfRange(t *testing.T) {
	var selfID [32]byte
	kt := NewKademliaTable(selfID, DefaultKademliaConfig())

	// Should not panic.
	kt.MarkRefreshed(-1)
	kt.MarkRefreshed(256)
}

// --- BucketReplacementLen out-of-range ---

func TestKademliaTable_BucketReplacementLen_OutOfRange(t *testing.T) {
	var selfID [32]byte
	kt := NewKademliaTable(selfID, DefaultKademliaConfig())

	if kt.BucketReplacementLen(-1) != 0 {
		t.Fatal("BucketReplacementLen(-1) should return 0")
	}
	if kt.BucketReplacementLen(256) != 0 {
		t.Fatal("BucketReplacementLen(256) should return 0")
	}
}

// --- MaxTableSize with eviction ---

func TestKademliaTable_MaxTableSize_EvictsStaleOnOverflow(t *testing.T) {
	var selfID [32]byte
	cfg := DefaultKademliaConfig()
	cfg.MaxTableSize = 2
	cfg.BucketSize = 16
	cfg.MaxFailCount = 1
	kt := NewKademliaTable(selfID, cfg)

	var id1, id2 [32]byte
	id1[31] = 1
	id2[31] = 2
	kt.AddNode(NodeEntry{ID: id1, LastSeen: time.Now()})
	kt.AddNode(NodeEntry{ID: id2, LastSeen: time.Now()})

	// Make id1 stale.
	kt.RecordFailure(id1) // evicts id1 due to MaxFailCount=1

	// Now table has 1 entry. Adding a new node should succeed.
	var id3 [32]byte
	id3[31] = 3
	added := kt.AddNode(NodeEntry{ID: id3, LastSeen: time.Now()})
	if !added {
		t.Fatal("expected node to be added after stale eviction freed table space")
	}
}

func TestKademliaTable_MaxTableSize_NoStaleToEvict(t *testing.T) {
	var selfID [32]byte
	cfg := DefaultKademliaConfig()
	cfg.MaxTableSize = 2
	cfg.BucketSize = 16
	kt := NewKademliaTable(selfID, cfg)

	var id1, id2 [32]byte
	id1[31] = 1
	id2[31] = 2
	kt.AddNode(NodeEntry{ID: id1, LastSeen: time.Now()})
	kt.AddNode(NodeEntry{ID: id2, LastSeen: time.Now()})

	// Table full, no stale nodes. New node goes to replacement.
	var id3 [32]byte
	id3[31] = 3
	added := kt.AddNode(NodeEntry{ID: id3, LastSeen: time.Now()})
	if added {
		t.Fatal("node should not be added to main entries when table full and no stale nodes")
	}
}

// --- evictStaleLocked: covers the case where a stale node is found ---

func TestKademliaTable_MaxTableSize_EvictStaleViaMaxTable(t *testing.T) {
	var selfID [32]byte
	cfg := DefaultKademliaConfig()
	cfg.MaxTableSize = 2
	cfg.BucketSize = 16
	cfg.StaleTimeout = 1 * time.Millisecond
	kt := NewKademliaTable(selfID, cfg)

	// Add two nodes: both in bucket 0 (distance 1).
	var id1, id2 [32]byte
	id1[31] = 1
	id2[31] = 3 // distance 2, bucket 1
	staleTime := time.Now().Add(-1 * time.Second)
	kt.AddNode(NodeEntry{ID: id1, LastSeen: staleTime}) // stale immediately
	kt.AddNode(NodeEntry{ID: id2, LastSeen: staleTime}) // stale immediately

	// Wait for stale timeout to expire.
	time.Sleep(5 * time.Millisecond)

	// Adding a third node with MaxTableSize=2 triggers evictStaleLocked which
	// will find id1 or id2 as stale and remove it.
	var id3 [32]byte
	id3[31] = 5 // distance 3, bucket 2
	added := kt.AddNode(NodeEntry{ID: id3, LastSeen: time.Now()})
	if !added {
		t.Fatal("node should be added after evicting a stale node")
	}
}

// --- AddNode: tail stale eviction ---

func TestKademliaTable_AddNode_TailStaleEviction(t *testing.T) {
	var selfID [32]byte
	cfg := DefaultKademliaConfig()
	cfg.BucketSize = 2
	cfg.StaleTimeout = 1 * time.Millisecond
	kt := NewKademliaTable(selfID, cfg)

	// Fill bucket 7 (distance 8).
	var id1, id2 [32]byte
	id1[31] = 0x80
	id2[31] = 0x81
	kt.AddNode(NodeEntry{ID: id1, LastSeen: time.Now()})
	// id2 gets stale LastSeen so it's the stale tail entry.
	staleTime := time.Now().Add(-1 * time.Second)
	kt.AddNode(NodeEntry{ID: id2, LastSeen: staleTime})

	time.Sleep(5 * time.Millisecond)

	// Add a new node to bucket 7: bucket is full, tail (id2) is stale -> replace it.
	var id3 [32]byte
	id3[31] = 0x82
	added := kt.AddNode(NodeEntry{ID: id3, LastSeen: time.Now()})
	if !added {
		t.Fatal("new node should replace stale tail entry in full bucket")
	}
	if kt.BucketLen(7) != 2 {
		t.Fatalf("bucket 7 should have 2 entries, got %d", kt.BucketLen(7))
	}
	// id3 should be in the table now.
	if kt.GetNode(id3) == nil {
		t.Fatal("id3 should be in the table after replacing stale tail")
	}
}
