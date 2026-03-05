// gossip_v2.go implements GossipSub V2.0 extensions: improved peer scoring,
// opportunistic grafting, and message prioritization. These features improve
// mesh health and latency for the consensus-layer gossip protocol.
package p2p

import (
	"sort"
	"sync"
	"time"
)

// GossipV2ScoreParams extends the base scoring with V2 parameters.
type GossipV2ScoreParams struct {
	// DeliveryRateWeight rewards delivery rate: score += weight * (deliveries/window).
	DeliveryRateWeight float64
	// FirstDeliveryBonus awards a bonus per unique first delivery.
	FirstDeliveryBonus float64
	// InvalidPenalty penalises each invalid message: score -= penalty * decay^time.
	InvalidPenalty float64
	// DecayFactor is applied per scoring interval (0 < decay < 1).
	DecayFactor float64
	// ScoringInterval determines how often scores are recomputed.
	ScoringInterval time.Duration
}

// DefaultGossipV2ScoreParams returns sensible defaults for V2 scoring.
func DefaultGossipV2ScoreParams() *GossipV2ScoreParams {
	return &GossipV2ScoreParams{
		DeliveryRateWeight: 1.0,
		FirstDeliveryBonus: 5.0,
		InvalidPenalty:     10.0,
		DecayFactor:        0.9,
		ScoringInterval:    1 * time.Second,
	}
}

// PeerV2Score tracks the score state for a single peer.
type PeerV2Score struct {
	TotalDeliveries uint64
	FirstDeliveries uint64
	InvalidMessages uint64
	Score           float64
	LastUpdate      time.Time

	// accInvalid accumulates the un-decayed penalty component.
	accInvalid float64
}

// GossipV2Scorer manages V2 peer scoring across topics.
// All methods are safe for concurrent use.
type GossipV2Scorer struct {
	mu     sync.RWMutex
	params *GossipV2ScoreParams
	peers  map[string]*PeerV2Score // peerID -> score state
}

// NewGossipV2Scorer creates a new V2 scorer with the given parameters.
func NewGossipV2Scorer(params *GossipV2ScoreParams) *GossipV2Scorer {
	return &GossipV2Scorer{
		params: params,
		peers:  make(map[string]*PeerV2Score),
	}
}

// RecordDelivery records a message delivery for a peer. isFirst indicates
// whether this peer was the first to deliver this specific message.
func (s *GossipV2Scorer) RecordDelivery(peerID string, isFirst bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ps := s.getOrCreate(peerID)
	ps.TotalDeliveries++
	if isFirst {
		ps.FirstDeliveries++
	}
}

// RecordInvalid records an invalid message from a peer.
func (s *GossipV2Scorer) RecordInvalid(peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ps := s.getOrCreate(peerID)
	ps.InvalidMessages++
	ps.accInvalid += s.params.InvalidPenalty
}

// UpdateScores recomputes scores for all tracked peers by applying
// delivery rewards and decaying the invalid-message penalty.
func (s *GossipV2Scorer) UpdateScores() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ps := range s.peers {
		// Delivery reward: first-delivery bonus + rate weight * total deliveries.
		deliveryScore := s.params.DeliveryRateWeight*float64(ps.TotalDeliveries) +
			s.params.FirstDeliveryBonus*float64(ps.FirstDeliveries)

		// Decay the accumulated invalid penalty.
		ps.accInvalid *= s.params.DecayFactor

		ps.Score = deliveryScore - ps.accInvalid
		ps.LastUpdate = time.Now()
	}
}

// PeerScore returns the current score for a peer. Returns 0 for unknown peers.
func (s *GossipV2Scorer) PeerScore(peerID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if ps, ok := s.peers[peerID]; ok {
		return ps.Score
	}
	return 0
}

// TopPeers returns the top-n peer IDs sorted by descending score.
func (s *GossipV2Scorer) TopPeers(n int) []string {
	s.mu.RLock()

	type entry struct {
		id    string
		score float64
	}

	entries := make([]entry, 0, len(s.peers))
	for id, ps := range s.peers {
		entries = append(entries, entry{id: id, score: ps.Score})
	}
	s.mu.RUnlock()

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].score > entries[j].score
	})

	if n > len(entries) {
		n = len(entries)
	}
	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = entries[i].id
	}
	return result
}

// getOrCreate returns the PeerV2Score for peerID, creating it if absent.
// Caller must hold s.mu.Lock.
func (s *GossipV2Scorer) getOrCreate(peerID string) *PeerV2Score {
	ps, ok := s.peers[peerID]
	if !ok {
		ps = &PeerV2Score{LastUpdate: time.Now()}
		s.peers[peerID] = ps
	}
	return ps
}

// --- Opportunistic Grafting ---

// OpportunisticGrafter manages mesh size and opportunistic peer grafting.
// When the mesh falls below dLow, it selects high-scoring candidate peers
// to graft up to targetD.
type OpportunisticGrafter struct {
	mu        sync.Mutex
	scorer    *GossipV2Scorer
	meshPeers map[string]bool // current mesh peer IDs
	targetD   int
	dLow      int
}

// NewOpportunisticGrafter creates a new grafter with the given scorer and
// mesh size parameters.
func NewOpportunisticGrafter(scorer *GossipV2Scorer, targetD, dLow int) *OpportunisticGrafter {
	return &OpportunisticGrafter{
		scorer:    scorer,
		meshPeers: make(map[string]bool),
		targetD:   targetD,
		dLow:      dLow,
	}
}

// AddMeshPeer adds a peer to the tracked mesh.
func (g *OpportunisticGrafter) AddMeshPeer(peerID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.meshPeers[peerID] = true
}

// RemoveMeshPeer removes a peer from the tracked mesh.
func (g *OpportunisticGrafter) RemoveMeshPeer(peerID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.meshPeers, peerID)
}

// MeshSize returns the current number of mesh peers.
func (g *OpportunisticGrafter) MeshSize() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.meshPeers)
}

// OpportunisticGraft returns peer IDs to graft when the mesh is under-connected.
// Returns an empty slice if the current mesh size is >= dLow.
// Candidates are sorted by descending score; enough are returned to reach targetD.
func (g *OpportunisticGrafter) OpportunisticGraft(allPeers []string) []string {
	g.mu.Lock()
	current := len(g.meshPeers)
	inMesh := make(map[string]bool, current)
	for id := range g.meshPeers {
		inMesh[id] = true
	}
	g.mu.Unlock()

	if current >= g.dLow {
		return nil
	}

	// Filter out already-meshed peers.
	type candidate struct {
		id    string
		score float64
	}
	candidates := make([]candidate, 0, len(allPeers))
	for _, pid := range allPeers {
		if !inMesh[pid] {
			candidates = append(candidates, candidate{id: pid, score: g.scorer.PeerScore(pid)})
		}
	}

	// Sort by descending score.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	needed := g.targetD - current
	if needed > len(candidates) {
		needed = len(candidates)
	}

	result := make([]string, needed)
	for i := 0; i < needed; i++ {
		result[i] = candidates[i].id
	}
	return result
}

// --- Message Prioritization ---

// MessagePriority defines the priority tier for gossip messages.
type MessagePriority int

const (
	GossipPriorityLow    MessagePriority = 1 // mempool ticks, low-urgency gossip
	GossipPriorityMedium MessagePriority = 2 // attestations, aggregations
	GossipPriorityHigh   MessagePriority = 3 // block proposals, FOCIL ILs
)

// TopicPriority returns the message priority for the given GossipTopic.
func TopicPriority(topic GossipTopic) MessagePriority {
	switch topic {
	case BeaconBlock:
		return GossipPriorityHigh
	case BeaconAggregateAndProof, VoluntaryExit, ProposerSlashing, AttesterSlashing,
		BlobSidecar, SyncCommitteeContribution:
		return GossipPriorityMedium
	case STARKMempoolTick:
		return GossipPriorityLow
	default:
		return GossipPriorityLow
	}
}

// PrioritizedMessage wraps a gossip message with its priority tier.
type PrioritizedMessage struct {
	Topic    GossipTopic
	Data     []byte
	Priority MessagePriority
	Enqueued time.Time
}

// PrioritizedGossipRouter drains HIGH before MEDIUM before LOW priority messages.
// All methods are safe for concurrent use.
type PrioritizedGossipRouter struct {
	mu   sync.Mutex
	high []PrioritizedMessage
	med  []PrioritizedMessage
	low  []PrioritizedMessage
}

// NewPrioritizedGossipRouter creates a new empty prioritized router.
func NewPrioritizedGossipRouter() *PrioritizedGossipRouter {
	return &PrioritizedGossipRouter{}
}

// Enqueue adds a message to the appropriate priority queue.
func (r *PrioritizedGossipRouter) Enqueue(msg PrioritizedMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch msg.Priority {
	case GossipPriorityHigh:
		r.high = append(r.high, msg)
	case GossipPriorityMedium:
		r.med = append(r.med, msg)
	default:
		r.low = append(r.low, msg)
	}
}

// Dequeue returns the highest-priority available message.
// Returns (msg, true) if a message is available, or (zero, false) if empty.
func (r *PrioritizedGossipRouter) Dequeue() (PrioritizedMessage, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.high) > 0 {
		msg := r.high[0]
		r.high = r.high[1:]
		return msg, true
	}
	if len(r.med) > 0 {
		msg := r.med[0]
		r.med = r.med[1:]
		return msg, true
	}
	if len(r.low) > 0 {
		msg := r.low[0]
		r.low = r.low[1:]
		return msg, true
	}
	return PrioritizedMessage{}, false
}

// Len returns the number of messages in each priority tier.
func (r *PrioritizedGossipRouter) Len() (high, med, low int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.high), len(r.med), len(r.low)
}

// --- Per-topic gossip parameters ---

// GossipParamsByTopic holds per-topic gossipsub D parameters.
type GossipParamsByTopic map[GossipTopic]TopicParams

// DefaultGossipParamsByTopic returns default per-topic gossipsub parameters.
// Block-related topics use higher mesh connectivity to reduce propagation latency;
// lower-priority topics use the standard defaults.
func DefaultGossipParamsByTopic() GossipParamsByTopic {
	defaults := DefaultTopicParams()

	return GossipParamsByTopic{
		BeaconBlock: {
			MeshD:             12,
			MeshDlo:           8,
			MeshDhi:           16,
			HeartbeatInterval: defaults.HeartbeatInterval,
			HistoryLength:     defaults.HistoryLength,
			HistoryGossip:     defaults.HistoryGossip,
			FanoutTTL:         defaults.FanoutTTL,
			SeenTTL:           defaults.SeenTTL,
		},
		BeaconAggregateAndProof: {
			MeshD:             8,
			MeshDlo:           6,
			MeshDhi:           12,
			HeartbeatInterval: defaults.HeartbeatInterval,
			HistoryLength:     defaults.HistoryLength,
			HistoryGossip:     defaults.HistoryGossip,
			FanoutTTL:         defaults.FanoutTTL,
			SeenTTL:           defaults.SeenTTL,
		},
		BlobSidecar: {
			MeshD:             8,
			MeshDlo:           6,
			MeshDhi:           12,
			HeartbeatInterval: defaults.HeartbeatInterval,
			HistoryLength:     defaults.HistoryLength,
			HistoryGossip:     defaults.HistoryGossip,
			FanoutTTL:         defaults.FanoutTTL,
			SeenTTL:           defaults.SeenTTL,
		},
		SyncCommitteeContribution: {
			MeshD:             8,
			MeshDlo:           6,
			MeshDhi:           12,
			HeartbeatInterval: defaults.HeartbeatInterval,
			HistoryLength:     defaults.HistoryLength,
			HistoryGossip:     defaults.HistoryGossip,
			FanoutTTL:         defaults.FanoutTTL,
			SeenTTL:           defaults.SeenTTL,
		},
		STARKMempoolTick: {
			MeshD:             4,
			MeshDlo:           3,
			MeshDhi:           8,
			HeartbeatInterval: defaults.HeartbeatInterval,
			HistoryLength:     defaults.HistoryLength,
			HistoryGossip:     defaults.HistoryGossip,
			FanoutTTL:         defaults.FanoutTTL,
			SeenTTL:           defaults.SeenTTL,
		},
	}
}
