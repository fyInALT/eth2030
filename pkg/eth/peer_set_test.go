package eth

import (
	"math/big"
	"sync"
	"testing"

	"github.com/eth2030/eth2030/p2p"
)

// newTestEthPeer creates a minimal EthPeer with the given id for testing.
func newTestEthPeer(id string) *EthPeer {
	peer := p2p.NewPeer(id, "127.0.0.1:30303", nil)
	return NewEthPeer(peer, nil)
}

// newTestEthPeerWithTD creates a peer whose underlying p2p.Peer has TD set.
func newTestEthPeerWithTD(id string, td *big.Int) *EthPeer {
	peer := p2p.NewPeer(id, "127.0.0.1:30303", nil)
	peer.SetHead([32]byte{}, td)
	return NewEthPeer(peer, nil)
}

func TestNewEthPeerSet(t *testing.T) {
	ps := NewEthPeerSet(10, 68)
	if ps.Capacity() != 10 {
		t.Fatalf("expected capacity 10, got %d", ps.Capacity())
	}
	if ps.Len() != 0 {
		t.Fatalf("expected len 0, got %d", ps.Len())
	}
	if ps.IsClosed() {
		t.Fatal("expected not closed")
	}
}

func TestRegisterAndGet(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	ep := newTestEthPeer("peer1")

	if err := ps.Register(ep, 68); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if ps.Len() != 1 {
		t.Fatalf("expected len 1, got %d", ps.Len())
	}

	got := ps.Get("peer1")
	if got == nil {
		t.Fatal("Get returned nil for registered peer")
	}
	if got.ID() != "peer1" {
		t.Fatalf("expected peer1, got %s", got.ID())
	}
}

func TestRegisterDuplicate(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	ep := newTestEthPeer("peer1")

	if err := ps.Register(ep, 68); err != nil {
		t.Fatalf("first Register failed: %v", err)
	}
	if err := ps.Register(ep, 68); err != ErrPeerExists {
		t.Fatalf("expected ErrPeerExists, got %v", err)
	}
}

func TestRegisterWhenFull(t *testing.T) {
	ps := NewEthPeerSet(2, 68)
	if err := ps.Register(newTestEthPeer("p1"), 68); err != nil {
		t.Fatal(err)
	}
	if err := ps.Register(newTestEthPeer("p2"), 68); err != nil {
		t.Fatal(err)
	}
	if err := ps.Register(newTestEthPeer("p3"), 68); err != ErrPeerSetFull {
		t.Fatalf("expected ErrPeerSetFull, got %v", err)
	}
}

func TestRegisterBelowMinVersion(t *testing.T) {
	ps := NewEthPeerSet(5, 70)
	ep := newTestEthPeer("p1")
	if err := ps.Register(ep, 68); err != ErrMinVersion {
		t.Fatalf("expected ErrMinVersion, got %v", err)
	}
}

func TestRegisterAfterClose(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	ps.Close()
	if err := ps.Register(newTestEthPeer("p1"), 68); err != ErrPeerSetClosed {
		t.Fatalf("expected ErrPeerSetClosed, got %v", err)
	}
}

func TestUnregister(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	ep := newTestEthPeer("peer1")
	if err := ps.Register(ep, 68); err != nil {
		t.Fatal(err)
	}

	if err := ps.Unregister("peer1"); err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}
	if ps.Len() != 0 {
		t.Fatalf("expected len 0 after unregister, got %d", ps.Len())
	}
	if ps.Get("peer1") != nil {
		t.Fatal("Get should return nil after unregister")
	}
}

func TestUnregisterMissing(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	if err := ps.Unregister("nonexistent"); err != ErrPeerMissing {
		t.Fatalf("expected ErrPeerMissing, got %v", err)
	}
}

func TestUnregisterAfterClose(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	if err := ps.Register(newTestEthPeer("p1"), 68); err != nil {
		t.Fatal(err)
	}
	ps.Close()
	if err := ps.Unregister("p1"); err != ErrPeerSetClosed {
		t.Fatalf("expected ErrPeerSetClosed, got %v", err)
	}
}

func TestGetMissing(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	if ps.Get("no-such-peer") != nil {
		t.Fatal("expected nil for missing peer")
	}
}

func TestHasCapability(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	ep := newTestEthPeer("p1")
	if err := ps.Register(ep, 70); err != nil {
		t.Fatal(err)
	}

	if !ps.HasCapability("p1", 70) {
		t.Error("expected peer to have capability 70")
	}
	if !ps.HasCapability("p1", 68) {
		t.Error("expected peer to have capability >= 68")
	}
	if ps.HasCapability("p1", 71) {
		t.Error("peer should not have capability 71, only 70")
	}
	if ps.HasCapability("missing", 68) {
		t.Error("missing peer should not have any capability")
	}
}

func TestPeerVersion(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	ep := newTestEthPeer("p1")
	if err := ps.Register(ep, 71); err != nil {
		t.Fatal(err)
	}

	if v := ps.PeerVersion("p1"); v != 71 {
		t.Fatalf("expected version 71, got %d", v)
	}
	if v := ps.PeerVersion("missing"); v != 0 {
		t.Fatalf("expected 0 for missing peer, got %d", v)
	}
}

func TestPeerScore(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	ep := newTestEthPeer("p1")
	if err := ps.Register(ep, 68); err != nil {
		t.Fatal(err)
	}

	// Default score is 0.
	if s := ps.PeerScore("p1"); s != 0.0 {
		t.Fatalf("expected default score 0.0, got %f", s)
	}
	if s := ps.PeerScore("missing"); s != 0.0 {
		t.Fatalf("expected 0.0 for missing peer, got %f", s)
	}
}

func TestRecordGoodResponse(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	ep := newTestEthPeer("p1")
	if err := ps.Register(ep, 68); err != nil {
		t.Fatal(err)
	}

	ps.RecordGoodResponse("p1")
	if s := ps.PeerScore("p1"); s <= 0 {
		t.Fatalf("expected positive score after good response, got %f", s)
	}
	// Should not panic for missing peer.
	ps.RecordGoodResponse("missing")
}

func TestRecordBadResponse(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	ep := newTestEthPeer("p1")
	if err := ps.Register(ep, 68); err != nil {
		t.Fatal(err)
	}

	ps.RecordBadResponse("p1")
	if s := ps.PeerScore("p1"); s >= 0 {
		t.Fatalf("expected negative score after bad response, got %f", s)
	}
	// Should not panic for missing peer.
	ps.RecordBadResponse("missing")
}

func TestRecordTimeout(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	ep := newTestEthPeer("p1")
	if err := ps.Register(ep, 68); err != nil {
		t.Fatal(err)
	}

	ps.RecordTimeout("p1")
	if s := ps.PeerScore("p1"); s >= 0 {
		t.Fatalf("expected negative score after timeout, got %f", s)
	}
	// Should not panic for missing peer.
	ps.RecordTimeout("missing")
}

func TestBestPeerEmpty(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	if ps.BestPeer() != nil {
		t.Fatal("expected nil BestPeer for empty set")
	}
}

func TestBestPeerSingle(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	ep := newTestEthPeerWithTD("p1", big.NewInt(1000))
	if err := ps.Register(ep, 68); err != nil {
		t.Fatal(err)
	}

	best := ps.BestPeer()
	if best == nil {
		t.Fatal("expected non-nil BestPeer")
	}
	if best.ID() != "p1" {
		t.Fatalf("expected p1, got %s", best.ID())
	}
}

func TestBestPeerMultiple(t *testing.T) {
	ps := NewEthPeerSet(10, 68)

	ep1 := newTestEthPeerWithTD("p1", big.NewInt(500))
	ep2 := newTestEthPeerWithTD("p2", big.NewInt(2000))
	ep3 := newTestEthPeerWithTD("p3", big.NewInt(1000))

	for _, ep := range []*EthPeer{ep1, ep2, ep3} {
		if err := ps.Register(ep, 68); err != nil {
			t.Fatal(err)
		}
	}

	best := ps.BestPeer()
	if best == nil {
		t.Fatal("expected non-nil BestPeer")
	}
	if best.ID() != "p2" {
		t.Fatalf("expected p2 (highest TD), got %s", best.ID())
	}
}

func TestBestPeerTieBrokenByScore(t *testing.T) {
	ps := NewEthPeerSet(10, 68)

	sameTD := big.NewInt(1000)
	ep1 := newTestEthPeerWithTD("p1", sameTD)
	ep2 := newTestEthPeerWithTD("p2", sameTD)

	if err := ps.Register(ep1, 68); err != nil {
		t.Fatal(err)
	}
	if err := ps.Register(ep2, 68); err != nil {
		t.Fatal(err)
	}

	// Give p2 a better score via multiple good responses.
	for range 5 {
		ps.RecordGoodResponse("p2")
	}

	best := ps.BestPeer()
	if best == nil {
		t.Fatal("expected non-nil BestPeer")
	}
	if best.ID() != "p2" {
		t.Fatalf("expected p2 (higher score at same TD), got %s", best.ID())
	}
}

func TestPeersAboveScore(t *testing.T) {
	ps := NewEthPeerSet(10, 68)

	ep1 := newTestEthPeer("p1")
	ep2 := newTestEthPeer("p2")
	ep3 := newTestEthPeer("p3")
	for _, ep := range []*EthPeer{ep1, ep2, ep3} {
		if err := ps.Register(ep, 68); err != nil {
			t.Fatal(err)
		}
	}

	// Give p1 and p3 positive scores, p2 stays at 0.
	ps.RecordGoodResponse("p1")
	ps.RecordGoodResponse("p1")
	ps.RecordGoodResponse("p3")

	peers := ps.PeersAboveScore(0.0)
	if len(peers) != 2 {
		t.Fatalf("expected 2 peers above score 0, got %d", len(peers))
	}

	// Verify none with zero score.
	peersAboveHigh := ps.PeersAboveScore(10.0)
	if len(peersAboveHigh) != 0 {
		t.Fatalf("expected 0 peers above score 10, got %d", len(peersAboveHigh))
	}
}

func TestPeersAboveScoreEmpty(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	peers := ps.PeersAboveScore(0.0)
	if len(peers) != 0 {
		t.Fatalf("expected empty slice, got %d", len(peers))
	}
}

func TestPeersWithVersion(t *testing.T) {
	ps := NewEthPeerSet(10, 68)

	ep68 := newTestEthPeer("p68")
	ep70 := newTestEthPeer("p70")
	ep72 := newTestEthPeer("p72")

	if err := ps.Register(ep68, 68); err != nil {
		t.Fatal(err)
	}
	if err := ps.Register(ep70, 70); err != nil {
		t.Fatal(err)
	}
	if err := ps.Register(ep72, 72); err != nil {
		t.Fatal(err)
	}

	all := ps.PeersWithVersion(68)
	if len(all) != 3 {
		t.Fatalf("expected 3 peers at version >= 68, got %d", len(all))
	}

	atLeast70 := ps.PeersWithVersion(70)
	if len(atLeast70) != 2 {
		t.Fatalf("expected 2 peers at version >= 70, got %d", len(atLeast70))
	}

	atLeast72 := ps.PeersWithVersion(72)
	if len(atLeast72) != 1 {
		t.Fatalf("expected 1 peer at version >= 72, got %d", len(atLeast72))
	}

	none := ps.PeersWithVersion(73)
	if len(none) != 0 {
		t.Fatalf("expected 0 peers at version >= 73, got %d", len(none))
	}
}

func TestShouldDisconnect(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	ep := newTestEthPeer("p1")
	if err := ps.Register(ep, 68); err != nil {
		t.Fatal(err)
	}

	// Fresh peer should not be disconnected.
	ids := ps.ShouldDisconnect()
	if len(ids) != 0 {
		t.Fatalf("expected no peers to disconnect at start, got %v", ids)
	}

	// Drive score below disconnect threshold (ScoreDisconnect = -50).
	// Each bad response = -5, each timeout = -10. We need < -50.
	// 6 timeouts = -60.
	for range 6 {
		ps.RecordTimeout("p1")
	}

	ids = ps.ShouldDisconnect()
	if len(ids) != 1 || ids[0] != "p1" {
		t.Fatalf("expected [p1] for disconnect, got %v", ids)
	}
}

func TestClose(t *testing.T) {
	ps := NewEthPeerSet(5, 68)
	if err := ps.Register(newTestEthPeer("p1"), 68); err != nil {
		t.Fatal(err)
	}

	ps.Close()
	if !ps.IsClosed() {
		t.Fatal("expected IsClosed to return true after Close")
	}
	if ps.Len() != 0 {
		t.Fatalf("expected len 0 after Close, got %d", ps.Len())
	}
}

func TestSetEventHandler(t *testing.T) {
	ps := NewEthPeerSet(5, 68)

	var mu sync.Mutex
	var events []PeerEventData

	ps.SetEventHandler(func(e PeerEventData) {
		mu.Lock()
		events = append(events, e)
		mu.Unlock()
	})

	ep := newTestEthPeer("p1")
	if err := ps.Register(ep, 68); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	if len(events) != 1 || events[0].Event != PeerEventRegistered {
		t.Fatalf("expected 1 registered event, got %v", events)
	}
	mu.Unlock()

	if err := ps.Unregister("p1"); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	if len(events) != 2 || events[1].Event != PeerEventUnregistered {
		t.Fatalf("expected 2 events (last: unregistered), got %v", events)
	}
	mu.Unlock()
}

func TestEventHandlerScoreChanged(t *testing.T) {
	ps := NewEthPeerSet(5, 68)

	var mu sync.Mutex
	var scoreEvents []PeerEventData

	ps.SetEventHandler(func(e PeerEventData) {
		mu.Lock()
		if e.Event == PeerEventScoreChanged {
			scoreEvents = append(scoreEvents, e)
		}
		mu.Unlock()
	})

	ep := newTestEthPeer("p1")
	if err := ps.Register(ep, 68); err != nil {
		t.Fatal(err)
	}

	ps.RecordGoodResponse("p1")
	ps.RecordBadResponse("p1")

	mu.Lock()
	if len(scoreEvents) != 2 {
		t.Fatalf("expected 2 score-changed events, got %d", len(scoreEvents))
	}
	mu.Unlock()
}

func TestConcurrentRegisterUnregister(t *testing.T) {
	ps := NewEthPeerSet(100, 68)
	var wg sync.WaitGroup

	for n := range 50 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := string(rune('a' + n))
			ep := newTestEthPeer(id)
			_ = ps.Register(ep, 68)
			_ = ps.Unregister(id)
		}(n)
	}
	wg.Wait()
}
