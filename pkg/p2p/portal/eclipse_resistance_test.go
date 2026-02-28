package portal

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestEclipseDefaultConfig(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	if cfg.MaxPeersPerSubnet != DefaultMaxPeersPerSubnet {
		t.Errorf("expected MaxPeersPerSubnet=%d, got %d", DefaultMaxPeersPerSubnet, cfg.MaxPeersPerSubnet)
	}
	if cfg.MinDistinctSubnets != DefaultMinDistinctSubnets {
		t.Errorf("expected MinDistinctSubnets=%d, got %d", DefaultMinDistinctSubnets, cfg.MinDistinctSubnets)
	}
	if cfg.SybilThreshold != DefaultSybilThreshold {
		t.Errorf("expected SybilThreshold=%f, got %f", DefaultSybilThreshold, cfg.SybilThreshold)
	}
	if cfg.DiversityCheckInterval != DefaultDiversityInterval {
		t.Errorf("expected DiversityCheckInterval=%v, got %v", DefaultDiversityInterval, cfg.DiversityCheckInterval)
	}
}

func TestDiversityAddPeerUniqueSubnet(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	v := NewPeerDiversityValidator(cfg)

	err := v.AddPeer("peer-1", "10.0.1.5")
	if err != nil {
		t.Fatalf("unexpected error adding peer from unique subnet: %v", err)
	}
	if v.PeerCount() != 1 {
		t.Errorf("expected peer count 1, got %d", v.PeerCount())
	}
}

func TestDiversitySameSubnetRejectsAfterMax(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	cfg.MaxPeersPerSubnet = 2
	v := NewPeerDiversityValidator(cfg)

	// Add 2 peers from same /16 subnet.
	v.AddPeer("peer-1", "192.168.1.1")
	v.AddPeer("peer-2", "192.168.2.2")

	// Third peer from same /16 should be rejected.
	err := v.AddPeer("peer-3", "192.168.3.3")
	if !errors.Is(err, ErrEclipseSubnetFull) {
		t.Fatalf("expected ErrEclipseSubnetFull, got %v", err)
	}
}

func TestDiversityRemovePeerFreesSlot(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	cfg.MaxPeersPerSubnet = 1
	v := NewPeerDiversityValidator(cfg)

	v.AddPeer("peer-1", "10.0.1.1")

	// Should be full.
	err := v.AddPeer("peer-2", "10.0.2.2")
	if !errors.Is(err, ErrEclipseSubnetFull) {
		t.Fatalf("expected ErrEclipseSubnetFull, got %v", err)
	}

	// Remove peer-1 and try again.
	v.RemovePeer("peer-1")

	err = v.AddPeer("peer-2", "10.0.2.2")
	if err != nil {
		t.Fatalf("expected success after removal, got %v", err)
	}
}

func TestDiversityStatsAccuracy(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	cfg.MaxPeersPerSubnet = 10
	v := NewPeerDiversityValidator(cfg)

	v.AddPeer("p1", "10.0.1.1")
	v.AddPeer("p2", "10.0.2.2")
	v.AddPeer("p3", "172.16.1.1")
	v.AddPeer("p4", "192.168.1.1")

	stats := v.Stats()
	if stats.TotalPeers != 4 {
		t.Errorf("expected 4 total peers, got %d", stats.TotalPeers)
	}
	if stats.UniqueSubnets != 3 {
		t.Errorf("expected 3 unique subnets, got %d", stats.UniqueSubnets)
	}
	// 10.0 has 2 peers.
	if stats.MaxPeersInSubnet != 2 {
		t.Errorf("expected max peers in subnet 2, got %d", stats.MaxPeersInSubnet)
	}
}

func TestSybilScorerSameIPPrefixScoresHigh(t *testing.T) {
	scorer := NewSybilScorer(0.5)
	now := int64(1000000)

	// First peer from a subnet - base score.
	scorer.ScorePeer("peer-1", "10.0.1.1", now)

	// Second peer from same subnet shortly after - should score higher.
	score := scorer.ScorePeer("peer-2", "10.0.1.2", now+5)
	if score < 0.2 {
		t.Errorf("expected elevated score for same-subnet rapid join, got %f", score)
	}
}

func TestSybilScorerRapidJoinScoresHigh(t *testing.T) {
	scorer := NewSybilScorer(0.5)
	now := int64(1000000)

	// Add several peers from same subnet rapidly.
	for i := 0; i < 5; i++ {
		scorer.ScorePeer(fmt.Sprintf("peer-%d", i), "10.0.1.1", now+int64(i))
	}

	// Next peer should score very high due to cluster + rapid join.
	score := scorer.ScorePeer("peer-next", "10.0.2.1", now+6)
	if score < 0.3 {
		t.Errorf("expected high score after rapid joins from same subnet, got %f", score)
	}
}

func TestSybilScorerUniqueIPSlowJoinScoresLow(t *testing.T) {
	scorer := NewSybilScorer(0.5)

	// Single peer from a unique subnet with no history.
	score := scorer.ScorePeer("peer-1", "203.0.113.1", int64(1000000))
	if score > 0.2 {
		t.Errorf("expected low score for unique IP + no history, got %f", score)
	}
}

func TestSybilIsSybilThresholdBoundary(t *testing.T) {
	scorer := NewSybilScorer(0.1)

	// First call from new subnet: score should be low (private IP = 0.1).
	isSybil := scorer.IsSybil("peer-1", "10.0.1.1", int64(1000000))
	// 10.x is private, so base score = 0.1, which equals threshold.
	if !isSybil {
		t.Error("expected Sybil detection at threshold boundary for private IP")
	}

	// Public IP with no history should not trigger.
	scorer2 := NewSybilScorer(0.5)
	isSybil = scorer2.IsSybil("peer-2", "203.0.113.1", int64(1000000))
	if isSybil {
		t.Error("did not expect Sybil for public IP with no history")
	}
}

func TestInboundValidatorRejectsEclipsingPeer(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	cfg.MaxPeersPerSubnet = 1
	div := NewPeerDiversityValidator(cfg)
	sybil := NewSybilScorer(0.9)
	iv := NewInboundValidator(div, sybil)

	// First peer accepted.
	div.AddPeer("peer-1", "10.0.1.1")

	// Second from same subnet should be rejected by diversity.
	err := iv.ValidateInbound("peer-2", "10.0.2.2", int64(1000000))
	if !errors.Is(err, ErrEclipseSubnetFull) {
		t.Fatalf("expected ErrEclipseSubnetFull, got %v", err)
	}
}

func TestDiversityManyDiversePeers(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	cfg.MaxPeersPerSubnet = 2
	v := NewPeerDiversityValidator(cfg)

	// Add peers from 10 different /16 subnets.
	for i := 0; i < 10; i++ {
		ip := fmt.Sprintf("%d.%d.1.1", 100+i, i)
		err := v.AddPeer(fmt.Sprintf("peer-%d", i), ip)
		if err != nil {
			t.Fatalf("failed to add diverse peer %d: %v", i, err)
		}
	}

	stats := v.Stats()
	if stats.TotalPeers != 10 {
		t.Errorf("expected 10 peers, got %d", stats.TotalPeers)
	}
	if stats.UniqueSubnets != 10 {
		t.Errorf("expected 10 unique subnets, got %d", stats.UniqueSubnets)
	}
}

func TestDiversityAllSameSubnetRejectsAfterLimit(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	cfg.MaxPeersPerSubnet = 3
	v := NewPeerDiversityValidator(cfg)

	for i := 0; i < 3; i++ {
		ip := fmt.Sprintf("10.0.%d.1", i+1)
		err := v.AddPeer(fmt.Sprintf("peer-%d", i), ip)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Fourth from same /16 should fail.
	err := v.AddPeer("peer-3", "10.0.4.1")
	if !errors.Is(err, ErrEclipseSubnetFull) {
		t.Fatalf("expected ErrEclipseSubnetFull, got %v", err)
	}
}

func TestDiversityPeerCountTracking(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	v := NewPeerDiversityValidator(cfg)

	if v.PeerCount() != 0 {
		t.Errorf("expected 0 peers initially, got %d", v.PeerCount())
	}

	v.AddPeer("p1", "10.0.1.1")
	v.AddPeer("p2", "172.16.1.1")
	if v.PeerCount() != 2 {
		t.Errorf("expected 2 peers, got %d", v.PeerCount())
	}

	v.RemovePeer("p1")
	if v.PeerCount() != 1 {
		t.Errorf("expected 1 peer after removal, got %d", v.PeerCount())
	}

	v.RemovePeer("p2")
	if v.PeerCount() != 0 {
		t.Errorf("expected 0 peers after removing all, got %d", v.PeerCount())
	}
}

func TestDiversityConcurrentAddRemove(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	cfg.MaxPeersPerSubnet = 100
	v := NewPeerDiversityValidator(cfg)

	var wg sync.WaitGroup

	// Add 50 peers concurrently from diverse subnets.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ip := fmt.Sprintf("%d.%d.1.1", 100+idx/256, idx%256)
			v.AddPeer(fmt.Sprintf("peer-%d", idx), ip)
		}(i)
	}
	wg.Wait()

	if v.PeerCount() != 50 {
		t.Errorf("expected 50 peers after concurrent adds, got %d", v.PeerCount())
	}

	// Remove 25 concurrently.
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			v.RemovePeer(fmt.Sprintf("peer-%d", idx))
		}(i)
	}
	wg.Wait()

	if v.PeerCount() != 25 {
		t.Errorf("expected 25 peers after concurrent removals, got %d", v.PeerCount())
	}
}

func TestDiversityEmptyValidatorState(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	v := NewPeerDiversityValidator(cfg)

	stats := v.Stats()
	if stats.TotalPeers != 0 {
		t.Errorf("expected 0 total peers, got %d", stats.TotalPeers)
	}
	if stats.UniqueSubnets != 0 {
		t.Errorf("expected 0 unique subnets, got %d", stats.UniqueSubnets)
	}
	if stats.MaxPeersInSubnet != 0 {
		t.Errorf("expected 0 max peers in subnet, got %d", stats.MaxPeersInSubnet)
	}

	// Removing non-existent peer should not panic.
	v.RemovePeer("nonexistent")
}

func TestDiversityLocalhostIP(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	v := NewPeerDiversityValidator(cfg)

	err := v.AddPeer("local-1", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error for localhost: %v", err)
	}

	stats := v.Stats()
	if stats.TotalPeers != 1 {
		t.Errorf("expected 1 peer, got %d", stats.TotalPeers)
	}
	// 127.0.0.1 has subnet "127.0".
	if _, ok := stats.SubnetDistribution["127.0"]; !ok {
		t.Error("expected subnet 127.0 in distribution")
	}
}

func TestDiversityIPv6Handling(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	cfg.MaxPeersPerSubnet = 2
	v := NewPeerDiversityValidator(cfg)

	// Add IPv6 peers.
	err := v.AddPeer("ipv6-1", "2001:db8::1")
	if err != nil {
		t.Fatalf("unexpected error for IPv6: %v", err)
	}

	err = v.AddPeer("ipv6-2", "2001:db8::2")
	if err != nil {
		t.Fatalf("unexpected error for second IPv6 from same /32: %v", err)
	}

	// Third from same /32 should be rejected.
	err = v.AddPeer("ipv6-3", "2001:db8::3")
	if !errors.Is(err, ErrEclipseSubnetFull) {
		t.Fatalf("expected ErrEclipseSubnetFull for third IPv6 from same /32, got %v", err)
	}

	if v.PeerCount() != 2 {
		t.Errorf("expected 2 peers, got %d", v.PeerCount())
	}
}

func TestDiversityZeroMaxPerSubnet(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	cfg.MaxPeersPerSubnet = 0
	v := NewPeerDiversityValidator(cfg)

	err := v.AddPeer("p1", "10.0.1.1")
	if !errors.Is(err, ErrEclipseZeroSubnetCap) {
		t.Fatalf("expected ErrEclipseZeroSubnetCap, got %v", err)
	}
}

func TestDiversityDuplicatePeerRejected(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	v := NewPeerDiversityValidator(cfg)

	v.AddPeer("peer-1", "10.0.1.1")
	err := v.AddPeer("peer-1", "172.16.1.1") // same ID, different IP
	if !errors.Is(err, ErrEclipseDuplicatePeer) {
		t.Fatalf("expected ErrEclipseDuplicatePeer, got %v", err)
	}
}

func TestSybilScorerThreshold(t *testing.T) {
	scorer := NewSybilScorer(0.75)
	if scorer.Threshold() != 0.75 {
		t.Errorf("expected threshold 0.75, got %f", scorer.Threshold())
	}
}

func TestExtractSubnetInvalidIP(t *testing.T) {
	_, err := extractSubnet("not-an-ip")
	if !errors.Is(err, ErrEclipseInvalidIP) {
		t.Fatalf("expected ErrEclipseInvalidIP, got %v", err)
	}
}

func TestDiversityValidatePeerWithoutAdding(t *testing.T) {
	cfg := NewEclipseResistanceConfig()
	cfg.MaxPeersPerSubnet = 1
	v := NewPeerDiversityValidator(cfg)

	v.AddPeer("p1", "10.0.1.1")

	// Validate should fail for same subnet.
	err := v.ValidatePeer("p2", "10.0.2.2")
	if !errors.Is(err, ErrEclipseSubnetFull) {
		t.Fatalf("expected ErrEclipseSubnetFull from ValidatePeer, got %v", err)
	}

	// But peer should not have been added.
	if v.PeerCount() != 1 {
		t.Errorf("ValidatePeer should not add peer, expected 1, got %d", v.PeerCount())
	}
}
