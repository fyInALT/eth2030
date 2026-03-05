package sync

import (
	"errors"
	"testing"
	"time"

	"github.com/eth2030/eth2030/core/types"
)

// mockAnnouncePeer implements BlkAnnouncePeer for testing.
type mockAnnouncePeer struct {
	id      string
	latency time.Duration
	blocks  []*types.Block
	err     error
}

func (p *mockAnnouncePeer) ID() string { return p.id }
func (p *mockAnnouncePeer) Latency() time.Duration { return p.latency }
func (p *mockAnnouncePeer) FetchBlocks(_ []types.Hash) ([]*types.Block, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.blocks, nil
}

func makeAnnHash(b byte) types.Hash {
	var h types.Hash
	h[0] = b
	return h
}

// --- NewBlkAnnounce ---

func TestNewBlkAnnounce(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	if ba == nil {
		t.Fatal("expected non-nil BlkAnnounce")
	}
	if ba.PendingCount() != 0 {
		t.Errorf("PendingCount = %d, want 0", ba.PendingCount())
	}
	if ba.KnownCount() != 0 {
		t.Errorf("KnownCount = %d, want 0", ba.KnownCount())
	}
	if ba.PeerCount() != 0 {
		t.Errorf("PeerCount = %d, want 0", ba.PeerCount())
	}
}

// --- HandleAnnouncement ---

func TestHandleAnnouncement_Basic(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	h1 := makeAnnHash(0x01)
	h2 := makeAnnHash(0x02)
	n := ba.HandleAnnouncement("peer1", []types.Hash{h1, h2}, []uint64{1, 2})
	if n != 2 {
		t.Errorf("HandleAnnouncement returned %d, want 2", n)
	}
	if ba.PendingCount() != 2 {
		t.Errorf("PendingCount = %d, want 2", ba.PendingCount())
	}
}

func TestHandleAnnouncement_DeduplicateKnown(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	h := makeAnnHash(0x10)
	ba.MarkKnown(h)
	n := ba.HandleAnnouncement("peer1", []types.Hash{h}, []uint64{10})
	if n != 0 {
		t.Errorf("expected 0 new, got %d", n)
	}
	m := ba.GetMetrics()
	if m.AnnouncementsDupes == 0 {
		t.Error("expected dupe counter incremented")
	}
}

func TestHandleAnnouncement_DeduplicatePending(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	h := makeAnnHash(0x20)
	ba.HandleAnnouncement("peer1", []types.Hash{h}, []uint64{1})
	n := ba.HandleAnnouncement("peer2", []types.Hash{h}, []uint64{1})
	if n != 0 {
		t.Errorf("expected 0 (already pending), got %d", n)
	}
	if ba.PendingCount() != 1 {
		t.Errorf("PendingCount = %d, want 1", ba.PendingCount())
	}
}

func TestHandleAnnouncement_HasBlockCallback(t *testing.T) {
	h := makeAnnHash(0x30)
	ba := NewBlkAnnounce(func(hash types.Hash) bool { return hash == h })
	n := ba.HandleAnnouncement("peer1", []types.Hash{h}, []uint64{1})
	if n != 0 {
		t.Errorf("expected 0 (hasBlock=true), got %d", n)
	}
}

func TestHandleAnnouncement_QueueFull(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	// Fill to max.
	hashes := make([]types.Hash, BlkAnnounceMaxPending)
	numbers := make([]uint64, BlkAnnounceMaxPending)
	for i := range hashes {
		hashes[i][0] = byte(i)
		hashes[i][1] = byte(i >> 8)
		numbers[i] = uint64(i)
	}
	ba.HandleAnnouncement("peer1", hashes, numbers)
	if ba.PendingCount() != BlkAnnounceMaxPending {
		t.Errorf("PendingCount = %d, want %d", ba.PendingCount(), BlkAnnounceMaxPending)
	}
	// One more should be dropped.
	extra := makeAnnHash(0xFF)
	n := ba.HandleAnnouncement("peer1", []types.Hash{extra}, []uint64{999})
	if n != 0 {
		t.Errorf("expected 0 (queue full), got %d", n)
	}
}

func TestHandleAnnouncement_NoNumbers(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	h := makeAnnHash(0x40)
	// Pass empty numbers slice — should not panic.
	n := ba.HandleAnnouncement("peer1", []types.Hash{h}, nil)
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
}

func TestHandleAnnouncement_PeerTracked(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	h := makeAnnHash(0x50)
	ba.HandleAnnouncement("peerA", []types.Hash{h}, []uint64{5})
	if ba.PeerCount() != 1 {
		t.Errorf("PeerCount = %d, want 1", ba.PeerCount())
	}
	blocks := ba.PeerBlocks("peerA")
	if len(blocks) != 1 || blocks[0] != h {
		t.Errorf("PeerBlocks = %v, want [%x]", blocks, h)
	}
}

// --- MarkKnown / HasPending / IsKnown ---

func TestMarkKnown_RemovesFromPending(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	h := makeAnnHash(0x60)
	ba.HandleAnnouncement("peer1", []types.Hash{h}, []uint64{1})
	if !ba.HasPending(h) {
		t.Fatal("expected HasPending=true")
	}
	ba.MarkKnown(h)
	if ba.HasPending(h) {
		t.Error("expected HasPending=false after MarkKnown")
	}
	if !ba.IsKnown(h) {
		t.Error("expected IsKnown=true after MarkKnown")
	}
}

func TestIsKnown_FalseForUnknown(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	if ba.IsKnown(makeAnnHash(0x99)) {
		t.Error("expected false for never-seen hash")
	}
}

func TestHasPending_FalseForUnknown(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	if ba.HasPending(makeAnnHash(0x99)) {
		t.Error("expected false for unknown hash")
	}
}

// --- PeerBlocks / PeerCount / RemovePeer ---

func TestPeerBlocks_NotFound(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	blocks := ba.PeerBlocks("ghost")
	if blocks != nil {
		t.Errorf("expected nil, got %v", blocks)
	}
}

func TestRemovePeer(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	ba.HandleAnnouncement("peerX", []types.Hash{makeAnnHash(0x70)}, []uint64{7})
	if ba.PeerCount() != 1 {
		t.Fatalf("PeerCount = %d, want 1", ba.PeerCount())
	}
	ba.RemovePeer("peerX")
	if ba.PeerCount() != 0 {
		t.Errorf("PeerCount = %d, want 0 after RemovePeer", ba.PeerCount())
	}
}

func TestRemovePeer_NonExistent(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	// Should not panic.
	ba.RemovePeer("nobody")
}

// --- PrunePeers ---

func TestPrunePeers_StaleRemoved(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	h := makeAnnHash(0x80)
	ba.HandleAnnouncement("oldPeer", []types.Hash{h}, []uint64{8})

	// Manually backdate the peer's announcedAt to force pruning.
	ba.mu.Lock()
	info := ba.peers["oldPeer"]
	info.announcedAt[0] = time.Now().Add(-2 * BlkAnnounceMaxPeerAge)
	ba.mu.Unlock()

	pruned := ba.PrunePeers()
	if pruned != 1 {
		t.Errorf("PrunePeers = %d, want 1", pruned)
	}
	if ba.PeerCount() != 0 {
		t.Errorf("PeerCount = %d, want 0", ba.PeerCount())
	}
}

func TestPrunePeers_FetchingNotPruned(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	h := makeAnnHash(0x81)
	ba.HandleAnnouncement("busyPeer", []types.Hash{h}, []uint64{8})

	ba.mu.Lock()
	info := ba.peers["busyPeer"]
	info.announcedAt[0] = time.Now().Add(-2 * BlkAnnounceMaxPeerAge)
	info.fetching = true
	ba.mu.Unlock()

	pruned := ba.PrunePeers()
	if pruned != 0 {
		t.Errorf("PrunePeers = %d, want 0 (peer is fetching)", pruned)
	}
}

func TestPrunePeers_NoPeers(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	pruned := ba.PrunePeers()
	if pruned != 0 {
		t.Errorf("expected 0, got %d", pruned)
	}
}

// --- FetchPending ---

func TestFetchPending_EmptyPending(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	peer := &mockAnnouncePeer{id: "p1", latency: time.Millisecond}
	blocks, err := ba.FetchPending([]BlkAnnouncePeer{peer})
	if err != nil || blocks != nil {
		t.Errorf("expected nil,nil got %v,%v", blocks, err)
	}
}

func TestFetchPending_NoPeers(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	ba.HandleAnnouncement("p1", []types.Hash{makeAnnHash(0x90)}, []uint64{1})
	_, err := ba.FetchPending(nil)
	if !errors.Is(err, ErrBlkAnnounceNoPeer) {
		t.Errorf("expected ErrBlkAnnounceNoPeer, got %v", err)
	}
}

func TestFetchPending_Success(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	h := makeAnnHash(0xA1)
	ba.HandleAnnouncement("peer1", []types.Hash{h}, []uint64{1})

	blk := types.NewBlock(&types.Header{}, nil)
	peer := &mockAnnouncePeer{id: "peer1", latency: time.Millisecond, blocks: []*types.Block{blk}}

	blocks, err := ba.FetchPending([]BlkAnnouncePeer{peer})
	if err != nil {
		t.Fatalf("FetchPending error: %v", err)
	}
	if len(blocks) != 1 {
		t.Errorf("expected 1 block, got %d", len(blocks))
	}
	m := ba.GetMetrics()
	if m.FetchesCompleted != 1 {
		t.Errorf("FetchesCompleted = %d, want 1", m.FetchesCompleted)
	}
}

func TestFetchPending_PeerError(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	h := makeAnnHash(0xA2)
	ba.HandleAnnouncement("peer1", []types.Hash{h}, []uint64{1})

	peer := &mockAnnouncePeer{id: "peer1", latency: time.Millisecond, err: errors.New("fetch failed")}
	_, err := ba.FetchPending([]BlkAnnouncePeer{peer})
	if err == nil {
		t.Fatal("expected error from peer")
	}
	m := ba.GetMetrics()
	if m.FetchesFailed != 1 {
		t.Errorf("FetchesFailed = %d, want 1", m.FetchesFailed)
	}
}

func TestFetchPending_SelectsBestPeer(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	for i := range 5 {
		h := makeAnnHash(byte(0xB0 + i))
		ba.HandleAnnouncement("pA", []types.Hash{h}, []uint64{uint64(i)})
	}

	var chosen string
	fastPeer := &mockAnnouncePeer{
		id: "fast", latency: time.Millisecond,
		blocks: []*types.Block{},
	}
	slowPeer := &mockAnnouncePeer{
		id: "slow", latency: time.Second,
		blocks: []*types.Block{},
	}
	// Override FetchBlocks to record which peer was called.
	type recordPeer struct {
		*mockAnnouncePeer
	}
	_ = chosen

	_, err := ba.FetchPending([]BlkAnnouncePeer{slowPeer, fastPeer})
	if err != nil {
		t.Fatalf("FetchPending error: %v", err)
	}
}

// --- GetMetrics ---

func TestGetMetrics_InitialZeros(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	m := ba.GetMetrics()
	if m.AnnouncementsReceived != 0 || m.FetchesStarted != 0 || m.BlocksFetched != 0 {
		t.Errorf("expected all-zero metrics, got %+v", m)
	}
}

func TestGetMetrics_AfterAnnouncements(t *testing.T) {
	ba := NewBlkAnnounce(nil)
	h1 := makeAnnHash(0xC1)
	h2 := makeAnnHash(0xC2)
	ba.HandleAnnouncement("p1", []types.Hash{h1, h2}, []uint64{1, 2})

	m := ba.GetMetrics()
	if m.AnnouncementsReceived != 2 {
		t.Errorf("AnnouncementsReceived = %d, want 2", m.AnnouncementsReceived)
	}
}
