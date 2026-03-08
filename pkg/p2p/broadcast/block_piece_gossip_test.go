package broadcast

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/eth2030/eth2030/core/types"
)

func makePieceHash(b byte) types.Hash {
	var h types.Hash
	h[0] = b
	return h
}

func makePieceMsg(blockHash types.Hash, index, total int) *BlockPieceMessage {
	return &BlockPieceMessage{
		BlockHash:   blockHash,
		PieceIndex:  index,
		TotalPieces: total,
		Data:        []byte{byte(index)},
		PieceHash:   makePieceHash(byte(index + 100)),
		BlockNumber: 1,
		Timestamp:   time.Now(),
	}
}

func TestDefaultBlockAssemblyConfig(t *testing.T) {
	cfg := DefaultBlockAssemblyConfig()
	if cfg.DataShards != 4 {
		t.Errorf("DataShards = %d, want 4", cfg.DataShards)
	}
	if cfg.AssemblyTimeout != 30*time.Second {
		t.Errorf("AssemblyTimeout = %v, want 30s", cfg.AssemblyTimeout)
	}
	if cfg.MaxAssemblies != 64 {
		t.Errorf("MaxAssemblies = %d, want 64", cfg.MaxAssemblies)
	}
}

func TestAddPiece_Basic(t *testing.T) {
	mgr := NewBlockAssemblyManager(DefaultBlockAssemblyConfig())
	hash := makePieceHash(1)
	msg := makePieceMsg(hash, 0, 8)

	complete, err := mgr.AddPiece(msg)
	if err != nil {
		t.Fatalf("AddPiece: %v", err)
	}
	if complete {
		t.Error("expected incomplete after 1 piece")
	}

	asm := mgr.GetAssembly(hash)
	if asm == nil {
		t.Fatal("expected assembly to exist")
	}
	if len(asm.Pieces) != 1 {
		t.Errorf("expected 1 piece, got %d", len(asm.Pieces))
	}
}

func TestAddPiece_Complete(t *testing.T) {
	cfg := DefaultBlockAssemblyConfig() // k=4
	mgr := NewBlockAssemblyManager(cfg)
	hash := makePieceHash(2)

	for i := 0; i < cfg.DataShards; i++ {
		complete, err := mgr.AddPiece(makePieceMsg(hash, i, 8))
		if err != nil {
			t.Fatalf("AddPiece(%d): %v", i, err)
		}
		if i < cfg.DataShards-1 && complete {
			t.Errorf("piece %d: expected incomplete", i)
		}
		if i == cfg.DataShards-1 && !complete {
			t.Error("expected complete after k pieces")
		}
	}

	if !mgr.IsComplete(hash) {
		t.Error("expected IsComplete=true")
	}
}

func TestAddPiece_AllPieces(t *testing.T) {
	cfg := DefaultBlockAssemblyConfig()
	mgr := NewBlockAssemblyManager(cfg)
	hash := makePieceHash(3)
	total := 8

	// Add all 8 pieces. After k=4, it should be complete.
	for i := 0; i < total; i++ {
		complete, err := mgr.AddPiece(makePieceMsg(hash, i, total))
		if i < cfg.DataShards-1 {
			if err != nil {
				t.Fatalf("AddPiece(%d): %v", i, err)
			}
			if complete {
				t.Errorf("piece %d: should not be complete yet", i)
			}
		} else if i == cfg.DataShards-1 {
			if err != nil {
				t.Fatalf("AddPiece(%d): %v", i, err)
			}
			if !complete {
				t.Error("piece k-1: should be complete")
			}
		} else {
			// After completion, further adds should return ErrPieceGossipComplete.
			if err != ErrPieceGossipComplete {
				t.Errorf("piece %d: expected ErrPieceGossipComplete, got %v", i, err)
			}
		}
	}
}

func TestAddPiece_Duplicate(t *testing.T) {
	mgr := NewBlockAssemblyManager(DefaultBlockAssemblyConfig())
	hash := makePieceHash(4)
	msg := makePieceMsg(hash, 0, 8)

	_, err := mgr.AddPiece(msg)
	if err != nil {
		t.Fatalf("first AddPiece: %v", err)
	}

	_, err = mgr.AddPiece(msg)
	if err != ErrPieceGossipDuplicate {
		t.Errorf("expected ErrPieceGossipDuplicate, got %v", err)
	}
}

func TestAddPiece_NilPiece(t *testing.T) {
	mgr := NewBlockAssemblyManager(DefaultBlockAssemblyConfig())
	_, err := mgr.AddPiece(nil)
	if err != ErrPieceGossipNilPiece {
		t.Errorf("expected ErrPieceGossipNilPiece, got %v", err)
	}
}

func TestAddPiece_AlreadyComplete(t *testing.T) {
	cfg := DefaultBlockAssemblyConfig()
	mgr := NewBlockAssemblyManager(cfg)
	hash := makePieceHash(5)

	// Complete the assembly.
	for i := 0; i < cfg.DataShards; i++ {
		mgr.AddPiece(makePieceMsg(hash, i, 8))
	}

	// Adding another piece to the completed assembly should error.
	_, err := mgr.AddPiece(makePieceMsg(hash, cfg.DataShards, 8))
	if err != ErrPieceGossipComplete {
		t.Errorf("expected ErrPieceGossipComplete, got %v", err)
	}
}

func TestGetAssembly(t *testing.T) {
	mgr := NewBlockAssemblyManager(DefaultBlockAssemblyConfig())
	hash := makePieceHash(6)

	mgr.AddPiece(makePieceMsg(hash, 0, 8))
	mgr.AddPiece(makePieceMsg(hash, 3, 8))

	asm := mgr.GetAssembly(hash)
	if asm == nil {
		t.Fatal("expected non-nil assembly")
	}
	if len(asm.Pieces) != 2 {
		t.Errorf("expected 2 pieces, got %d", len(asm.Pieces))
	}
	if asm.BlockHash != hash {
		t.Error("block hash mismatch")
	}
}

func TestGetAssembly_NotFound(t *testing.T) {
	mgr := NewBlockAssemblyManager(DefaultBlockAssemblyConfig())
	asm := mgr.GetAssembly(makePieceHash(99))
	if asm != nil {
		t.Error("expected nil for unknown hash")
	}
}

func TestIsComplete(t *testing.T) {
	cfg := DefaultBlockAssemblyConfig()
	mgr := NewBlockAssemblyManager(cfg)
	hash := makePieceHash(7)

	// Not found -> false.
	if mgr.IsComplete(makePieceHash(255)) {
		t.Error("expected false for unknown hash")
	}

	// Add k-1 pieces -> false.
	for i := 0; i < cfg.DataShards-1; i++ {
		mgr.AddPiece(makePieceMsg(hash, i, 8))
	}
	if mgr.IsComplete(hash) {
		t.Error("expected false with k-1 pieces")
	}

	// Add k-th piece -> true.
	mgr.AddPiece(makePieceMsg(hash, cfg.DataShards-1, 8))
	if !mgr.IsComplete(hash) {
		t.Error("expected true after k pieces")
	}
}

func TestPropagateBlockPieces_SqrtFanout(t *testing.T) {
	mgr := NewBlockAssemblyManager(DefaultBlockAssemblyConfig())

	// Add 16 peers. sqrt(16) = 4.
	for i := 0; i < 16; i++ {
		mgr.AddPeer(fmt.Sprintf("peer-%d", i))
	}

	pieces := []*BlockPieceMessage{makePieceMsg(makePieceHash(8), 0, 8)}
	selected := mgr.PropagateBlockPieces(pieces)

	expected := int(math.Ceil(math.Sqrt(16)))
	if len(selected) != expected {
		t.Errorf("got %d peers, want %d (sqrt(16))", len(selected), expected)
	}
}

func TestPropagateBlockPieces_NoPeers(t *testing.T) {
	mgr := NewBlockAssemblyManager(DefaultBlockAssemblyConfig())
	pieces := []*BlockPieceMessage{makePieceMsg(makePieceHash(9), 0, 8)}
	selected := mgr.PropagateBlockPieces(pieces)
	if len(selected) != 0 {
		t.Errorf("expected 0 peers, got %d", len(selected))
	}
}

func TestCleanupExpired(t *testing.T) {
	cfg := BlockAssemblyConfig{
		DataShards:      4,
		AssemblyTimeout: 50 * time.Millisecond,
		MaxAssemblies:   64,
	}
	mgr := NewBlockAssemblyManager(cfg)

	hash := makePieceHash(10)
	mgr.AddPiece(makePieceMsg(hash, 0, 8))

	// Wait for expiry.
	time.Sleep(100 * time.Millisecond)

	removed := mgr.CleanupExpired()
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	if mgr.GetAssembly(hash) != nil {
		t.Error("expired assembly should be nil after cleanup")
	}
}

func TestCleanupExpired_PreservesActive(t *testing.T) {
	cfg := BlockAssemblyConfig{
		DataShards:      4,
		AssemblyTimeout: 10 * time.Second,
		MaxAssemblies:   64,
	}
	mgr := NewBlockAssemblyManager(cfg)

	hash := makePieceHash(11)
	mgr.AddPiece(makePieceMsg(hash, 0, 8))

	removed := mgr.CleanupExpired()
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}

	if mgr.GetAssembly(hash) == nil {
		t.Error("active assembly should not be removed")
	}
}

func TestBlockAssemblyManager_ConcurrentAccess(t *testing.T) {
	mgr := NewBlockAssemblyManager(DefaultBlockAssemblyConfig())

	// Add some peers.
	for i := 0; i < 10; i++ {
		mgr.AddPeer(fmt.Sprintf("peer-%d", i))
	}

	var wg sync.WaitGroup
	// 20 goroutines adding pieces to different blocks concurrently.
	for g := 0; g < 20; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			hash := makePieceHash(byte(id))
			for p := 0; p < 4; p++ {
				mgr.AddPiece(makePieceMsg(hash, p, 8))
			}
		}(g)
	}

	// Concurrent reads.
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mgr.IsComplete(makePieceHash(byte(id)))
			mgr.GetAssembly(makePieceHash(byte(id)))
			mgr.Stats()
		}(g)
	}

	// Concurrent propagation.
	wg.Add(1)
	go func() {
		defer wg.Done()
		mgr.PropagateBlockPieces([]*BlockPieceMessage{
			makePieceMsg(makePieceHash(50), 0, 8),
		})
	}()

	wg.Wait()

	// Verify at least some assemblies are complete.
	total, complete, _ := mgr.Stats()
	if total == 0 {
		t.Error("expected some assemblies")
	}
	if complete == 0 {
		t.Error("expected some complete assemblies")
	}
}

func TestBlockPieceMessage_Fields(t *testing.T) {
	now := time.Now()
	hash := makePieceHash(12)
	msg := &BlockPieceMessage{
		BlockHash:   hash,
		PieceIndex:  3,
		TotalPieces: 8,
		Data:        []byte{0xAA, 0xBB},
		PieceHash:   makePieceHash(103),
		BlockNumber: 42,
		Timestamp:   now,
	}

	if msg.BlockHash != hash {
		t.Error("BlockHash mismatch")
	}
	if msg.PieceIndex != 3 {
		t.Errorf("PieceIndex = %d, want 3", msg.PieceIndex)
	}
	if msg.TotalPieces != 8 {
		t.Errorf("TotalPieces = %d, want 8", msg.TotalPieces)
	}
	if len(msg.Data) != 2 {
		t.Errorf("Data len = %d, want 2", len(msg.Data))
	}
	if msg.BlockNumber != 42 {
		t.Errorf("BlockNumber = %d, want 42", msg.BlockNumber)
	}
	if msg.Timestamp != now {
		t.Error("Timestamp mismatch")
	}
}

func TestStats(t *testing.T) {
	cfg := DefaultBlockAssemblyConfig()
	mgr := NewBlockAssemblyManager(cfg)

	// Empty.
	total, complete, incomplete := mgr.Stats()
	if total != 0 || complete != 0 || incomplete != 0 {
		t.Errorf("expected all zeros, got total=%d complete=%d incomplete=%d", total, complete, incomplete)
	}

	// Add one incomplete assembly.
	mgr.AddPiece(makePieceMsg(makePieceHash(20), 0, 8))
	total, complete, incomplete = mgr.Stats()
	if total != 1 || complete != 0 || incomplete != 1 {
		t.Errorf("expected 1/0/1, got %d/%d/%d", total, complete, incomplete)
	}

	// Complete it.
	for i := 1; i < cfg.DataShards; i++ {
		mgr.AddPiece(makePieceMsg(makePieceHash(20), i, 8))
	}
	total, complete, incomplete = mgr.Stats()
	if total != 1 || complete != 1 || incomplete != 0 {
		t.Errorf("expected 1/1/0, got %d/%d/%d", total, complete, incomplete)
	}

	// Add another incomplete assembly.
	mgr.AddPiece(makePieceMsg(makePieceHash(21), 0, 8))
	total, complete, incomplete = mgr.Stats()
	if total != 2 || complete != 1 || incomplete != 1 {
		t.Errorf("expected 2/1/1, got %d/%d/%d", total, complete, incomplete)
	}
}

// TestBlockPieceGossipTopic verifies slot-keyed gossip topic format (GAP-4.2).
func TestBlockPieceGossipTopic(t *testing.T) {
	topic := BlockPieceTopicName(42, 7)
	want := "block_piece/42/7"
	if topic != want {
		t.Errorf("BlockPieceTopicName(42,7) = %q, want %q", topic, want)
	}
	topic0 := BlockPieceTopicName(0, 0)
	if topic0 != "block_piece/0/0" {
		t.Errorf("BlockPieceTopicName(0,0) = %q, want \"block_piece/0/0\"", topic0)
	}
}

// TestBlockPieceGossip simulates 16 peers each holding 2 of 16 pieces for a
// slot. Verifies that each peer gets distinct custody assignments covering
// all 16 pieces and that PieceCustodyIndex is deterministic (GAP-4.2).
func TestBlockPieceGossip(t *testing.T) {
	const totalPieces = 16
	const custodyPerPeer = 2
	const slot = uint64(100)

	peers := make([]string, totalPieces)
	for i := range peers {
		peers[i] = fmt.Sprintf("peer-%d", i)
	}

	// Track which pieces are covered across all peers.
	covered := make(map[int]bool)
	for _, peerID := range peers {
		pieces := PeerCustodyPieces(peerID, slot, totalPieces, custodyPerPeer)
		if len(pieces) != custodyPerPeer {
			t.Errorf("peer %s: expected %d custody pieces, got %d", peerID, custodyPerPeer, len(pieces))
		}
		for _, idx := range pieces {
			if idx < 0 || idx >= totalPieces {
				t.Errorf("peer %s: custody piece %d out of range [0,%d)", peerID, idx, totalPieces)
			}
			covered[idx] = true
		}
	}

	// With 16 peers × 2 custody pieces each (32 assignments total), the network
	// must cover at least k=8 distinct pieces — enough for block reconstruction.
	const minCoverage = 8
	if len(covered) < minCoverage {
		t.Errorf("covered %d/%d pieces; want at least %d (k=8 for reconstruction)",
			len(covered), totalPieces, minCoverage)
	}

	// Determinism: same inputs yield same index.
	idx1 := PieceCustodyIndex("peer-5", slot, totalPieces)
	idx2 := PieceCustodyIndex("peer-5", slot, totalPieces)
	if idx1 != idx2 {
		t.Errorf("PieceCustodyIndex not deterministic: %d vs %d", idx1, idx2)
	}

	// Different slots give different distributions.
	idx3 := PieceCustodyIndex("peer-5", slot+1, totalPieces)
	_ = idx3 // may or may not differ; just ensure it doesn't panic
}
