// gossip_v2_test.go tests the GossipSub V2.0 scoring, opportunistic
// grafting, message prioritization, and per-topic gossip parameters.
package p2p

import (
	"fmt"
	"testing"
	"time"
)

// TestGossipV2Scoring verifies that dishonest peers (sending invalid messages)
// end up with lower scores than honest peers, and that TopPeers excludes them.
func TestGossipV2Scoring(t *testing.T) {
	params := DefaultGossipV2ScoreParams()
	scorer := NewGossipV2Scorer(params)

	totalPeers := 100
	dishonestCount := 10

	// Record honest deliveries for all 100 peers.
	for i := 0; i < totalPeers; i++ {
		peerID := fmt.Sprintf("peer-%d", i)
		scorer.RecordDelivery(peerID, true) // first delivery
		scorer.RecordDelivery(peerID, false)
	}

	// Dishonest peers also send many invalid messages.
	for i := 0; i < dishonestCount; i++ {
		peerID := fmt.Sprintf("peer-%d", i)
		for j := 0; j < 5; j++ {
			scorer.RecordInvalid(peerID)
		}
	}

	scorer.UpdateScores()

	// Honest peer scores should exceed dishonest peer scores.
	dishonestScore := scorer.PeerScore("peer-0")
	honestScore := scorer.PeerScore(fmt.Sprintf("peer-%d", totalPeers-1))

	if dishonestScore >= honestScore {
		t.Errorf("dishonest peer score %.2f >= honest peer score %.2f", dishonestScore, honestScore)
	}

	// TopPeers(10) should not include any of the dishonest peers.
	top := scorer.TopPeers(10)
	if len(top) == 0 {
		t.Fatal("TopPeers returned empty slice")
	}
	dishonestSet := make(map[string]bool)
	for i := 0; i < dishonestCount; i++ {
		dishonestSet[fmt.Sprintf("peer-%d", i)] = true
	}
	for _, pid := range top {
		if dishonestSet[pid] {
			t.Errorf("dishonest peer %q appeared in top peers", pid)
		}
	}
}

// TestGossipV2ScoringDecay verifies that UpdateScores applies decay to scores.
func TestGossipV2ScoringDecay(t *testing.T) {
	params := DefaultGossipV2ScoreParams()
	params.InvalidPenalty = 20.0
	params.DecayFactor = 0.5
	scorer := NewGossipV2Scorer(params)

	scorer.RecordInvalid("peer-a")
	scorer.UpdateScores()
	scoreBefore := scorer.PeerScore("peer-a")

	scorer.UpdateScores()
	scoreAfter := scorer.PeerScore("peer-a")

	// After a second decay pass the score should be less negative (closer to 0).
	if scoreAfter <= scoreBefore {
		t.Errorf("expected score to improve after decay: before=%.2f after=%.2f", scoreBefore, scoreAfter)
	}
}

// TestOpportunisticGraft verifies that OpportunisticGraft returns enough peers
// to bring the mesh up to targetD when below dLow.
func TestOpportunisticGraft(t *testing.T) {
	params := DefaultGossipV2ScoreParams()
	scorer := NewGossipV2Scorer(params)

	targetD := 8
	dLow := 6

	grafter := NewOpportunisticGrafter(scorer, targetD, dLow)

	// Seed three mesh peers.
	for i := 0; i < 3; i++ {
		grafter.AddMeshPeer(fmt.Sprintf("mesh-%d", i))
	}

	if grafter.MeshSize() != 3 {
		t.Fatalf("expected MeshSize 3, got %d", grafter.MeshSize())
	}

	// Build a candidate pool of 20 peers with some scores.
	allPeers := make([]string, 20)
	for i := 0; i < 20; i++ {
		pid := fmt.Sprintf("candidate-%d", i)
		allPeers[i] = pid
		scorer.RecordDelivery(pid, true)
	}
	scorer.UpdateScores()

	toGraft := grafter.OpportunisticGraft(allPeers)
	if len(toGraft) == 0 {
		t.Fatal("OpportunisticGraft returned empty list when mesh is under-connected")
	}

	// Add them to the mesh and verify we reach dLow.
	for _, pid := range toGraft {
		grafter.AddMeshPeer(pid)
	}
	if grafter.MeshSize() < dLow {
		t.Errorf("after grafting MeshSize %d < dLow %d", grafter.MeshSize(), dLow)
	}
}

// TestOpportunisticGraftNoop verifies no grafting happens when mesh >= dLow.
func TestOpportunisticGraftNoop(t *testing.T) {
	params := DefaultGossipV2ScoreParams()
	scorer := NewGossipV2Scorer(params)
	grafter := NewOpportunisticGrafter(scorer, 8, 6)

	for i := 0; i < 8; i++ {
		grafter.AddMeshPeer(fmt.Sprintf("mesh-%d", i))
	}

	allPeers := []string{"extra-0", "extra-1"}
	toGraft := grafter.OpportunisticGraft(allPeers)
	if len(toGraft) != 0 {
		t.Errorf("expected no grafting when mesh is full, got %v", toGraft)
	}
}

// TestMessagePrioritization verifies that HIGH priority messages are dequeued
// before MEDIUM, and MEDIUM before LOW, regardless of enqueue order.
func TestMessagePrioritization(t *testing.T) {
	router := NewPrioritizedGossipRouter()

	// Enqueue 100 LOW priority messages.
	for i := 0; i < 100; i++ {
		router.Enqueue(PrioritizedMessage{
			Topic:    STARKMempoolTick,
			Data:     []byte{byte(i)},
			Priority: GossipPriorityLow,
			Enqueued: time.Now(),
		})
	}

	// Enqueue 1 HIGH priority message.
	router.Enqueue(PrioritizedMessage{
		Topic:    BeaconBlock,
		Data:     []byte("high"),
		Priority: GossipPriorityHigh,
		Enqueued: time.Now(),
	})

	// Enqueue 10 MEDIUM priority messages.
	for i := 0; i < 10; i++ {
		router.Enqueue(PrioritizedMessage{
			Topic:    BeaconAggregateAndProof,
			Data:     []byte{byte(i)},
			Priority: GossipPriorityMedium,
			Enqueued: time.Now(),
		})
	}

	high, med, low := router.Len()
	if high != 1 || med != 10 || low != 100 {
		t.Fatalf("Len() = (%d, %d, %d), want (1, 10, 100)", high, med, low)
	}

	// First dequeue must be HIGH.
	msg, ok := router.Dequeue()
	if !ok {
		t.Fatal("Dequeue returned nothing")
	}
	if msg.Priority != GossipPriorityHigh {
		t.Errorf("first dequeued priority = %d, want GossipPriorityHigh (%d)", msg.Priority, GossipPriorityHigh)
	}

	// Next 10 must be MEDIUM.
	for i := 0; i < 10; i++ {
		msg, ok = router.Dequeue()
		if !ok {
			t.Fatalf("Dequeue %d returned nothing", i+2)
		}
		if msg.Priority != GossipPriorityMedium {
			t.Errorf("dequeue %d priority = %d, want GossipPriorityMedium (%d)", i+2, msg.Priority, GossipPriorityMedium)
		}
	}

	// Remaining 100 must be LOW.
	for i := 0; i < 100; i++ {
		msg, ok = router.Dequeue()
		if !ok {
			t.Fatalf("Dequeue %d returned nothing", i+12)
		}
		if msg.Priority != GossipPriorityLow {
			t.Errorf("dequeue %d priority = %d, want GossipPriorityLow (%d)", i+12, msg.Priority, GossipPriorityLow)
		}
	}

	// Router should be empty now.
	_, ok = router.Dequeue()
	if ok {
		t.Error("Dequeue should return false on empty router")
	}
}

// TestTopicPriority verifies that TopicPriority returns correct priorities.
func TestTopicPriority(t *testing.T) {
	tests := []struct {
		topic    GossipTopic
		wantPrio MessagePriority
	}{
		{BeaconBlock, GossipPriorityHigh},
		{BeaconAggregateAndProof, GossipPriorityMedium},
		{STARKMempoolTick, GossipPriorityLow},
	}

	for _, tc := range tests {
		got := TopicPriority(tc.topic)
		if got != tc.wantPrio {
			t.Errorf("TopicPriority(%v) = %d, want %d", tc.topic, got, tc.wantPrio)
		}
	}
}

// TestGossipParamsByTopic verifies that DefaultGossipParamsByTopic returns
// sensible per-topic parameters covering key topics.
func TestGossipParamsByTopic(t *testing.T) {
	byTopic := DefaultGossipParamsByTopic()

	if len(byTopic) == 0 {
		t.Fatal("DefaultGossipParamsByTopic returned empty map")
	}

	// BeaconBlock should have more aggressive D settings.
	blockParams, ok := byTopic[BeaconBlock]
	if !ok {
		t.Fatal("no params for BeaconBlock topic")
	}
	if blockParams.MeshD <= 0 {
		t.Errorf("BeaconBlock MeshD = %d, want > 0", blockParams.MeshD)
	}

	// All configured topics must have valid (non-zero) MeshD.
	for topic, params := range byTopic {
		if params.MeshD <= 0 {
			t.Errorf("topic %v has MeshD = %d", topic, params.MeshD)
		}
		if params.MeshDlo > params.MeshD {
			t.Errorf("topic %v: MeshDlo %d > MeshD %d", topic, params.MeshDlo, params.MeshD)
		}
		if params.MeshDhi < params.MeshD {
			t.Errorf("topic %v: MeshDhi %d < MeshD %d", topic, params.MeshDhi, params.MeshD)
		}
	}
}

// TestRemoveMeshPeer verifies RemoveMeshPeer decrements the mesh size.
func TestRemoveMeshPeer(t *testing.T) {
	params := DefaultGossipV2ScoreParams()
	scorer := NewGossipV2Scorer(params)
	grafter := NewOpportunisticGrafter(scorer, 8, 6)

	grafter.AddMeshPeer("peer-a")
	grafter.AddMeshPeer("peer-b")
	grafter.RemoveMeshPeer("peer-a")

	if grafter.MeshSize() != 1 {
		t.Errorf("MeshSize = %d, want 1", grafter.MeshSize())
	}
}
