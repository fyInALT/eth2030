// builder_reputation.go implements sliding-window builder reputation tracking
// for ePBS with latency tracking, slashing detection, and score decay.
package epbs

import (
	"math"
	"sort"
	"sync"
)

// SlashingEventType classifies builder misbehavior.
type SlashingEventType uint8

const (
	SlashEventLateDelivery   SlashingEventType = iota // delivered after deadline
	SlashEventInvalidPayload                          // bad payload
	SlashEventEquivocation                            // conflicting payloads
)

// String returns a human-readable name for the event type.
func (t SlashingEventType) String() string {
	switch t {
	case SlashEventLateDelivery:
		return "late_delivery"
	case SlashEventInvalidPayload:
		return "invalid_payload"
	case SlashEventEquivocation:
		return "equivocation"
	default:
		return "unknown"
	}
}

// SlashingEvent records a detected builder misbehavior incident.
type SlashingEvent struct {
	BuilderID    string            // builder identifier
	Slot         uint64            // beacon slot of misbehavior
	Type         SlashingEventType // event classification
	EvidenceHash [32]byte          // evidence hash for auditing
	Description  string            // human-readable summary
}

// ReputationConfig configures the BuilderReputationTracker.
type ReputationConfig struct {
	WindowSize      int     // recent events to consider (default: 100)
	MinDeliveries   int     // min events before meaningful scoring
	SlashingPenalty float64 // score penalty (0-1) per slashing event
	DecayFactor     float64 // exponential decay per DecayScores call (0-1)
	DefaultScore    float64 // score for builders with insufficient history
}

// DefaultReputationConfig returns sensible production defaults.
func DefaultReputationConfig() ReputationConfig {
	return ReputationConfig{
		WindowSize:      100,
		MinDeliveries:   5,
		SlashingPenalty: 0.15,
		DecayFactor:     0.98,
		DefaultScore:    0.5,
	}
}

// deliveryEvent records a single delivery outcome.
type deliveryEvent struct {
	slot      uint64
	latencyMs int64
	success   bool
	reason    string // failure reason, empty for success
}

// ReputationRecord tracks a single builder's performance history.
type ReputationRecord struct {
	BuilderID      string  // builder identifier
	Deliveries     uint64  // total successful deliveries
	Failures       uint64  // total failed deliveries
	TotalLatencyMs int64   // sum of latencies across successes
	SlashingCount  uint64  // number of slashing events applied
	CurrentScore   float64 // latest reliability score (0-1)
	events         []deliveryEvent
	slashings      []SlashingEvent
}

// AverageLatencyMs returns average latency across successful deliveries.
func (r *ReputationRecord) AverageLatencyMs() int64 {
	if r.Deliveries == 0 {
		return 0
	}
	return r.TotalLatencyMs / int64(r.Deliveries)
}

// SuccessRate returns the fraction of events that were successful.
func (r *ReputationRecord) SuccessRate() float64 {
	total := r.Deliveries + r.Failures
	if total == 0 {
		return 1.0
	}
	return float64(r.Deliveries) / float64(total)
}

// BuilderReputationTracker monitors builder performance over time with a
// sliding window and produces reliability scores. Thread-safe.
type BuilderReputationTracker struct {
	mu      sync.RWMutex
	config  ReputationConfig
	records map[string]*ReputationRecord
}

// NewBuilderReputationTracker creates a new tracker with the given config.
func NewBuilderReputationTracker(cfg ReputationConfig) *BuilderReputationTracker {
	if cfg.WindowSize <= 0 {
		cfg.WindowSize = 100
	}
	if cfg.MinDeliveries <= 0 {
		cfg.MinDeliveries = 5
	}
	if cfg.SlashingPenalty <= 0 || cfg.SlashingPenalty > 1.0 {
		cfg.SlashingPenalty = 0.15
	}
	if cfg.DecayFactor <= 0 || cfg.DecayFactor > 1.0 {
		cfg.DecayFactor = 0.98
	}
	if cfg.DefaultScore <= 0 || cfg.DefaultScore > 1.0 {
		cfg.DefaultScore = 0.5
	}
	return &BuilderReputationTracker{
		config:  cfg,
		records: make(map[string]*ReputationRecord),
	}
}

// RecordDelivery records a successful payload delivery for a builder.
func (rt *BuilderReputationTracker) RecordDelivery(builderID string, slot uint64, latencyMs int64) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rec := rt.getOrCreate(builderID)
	rec.Deliveries++
	if latencyMs < 0 {
		latencyMs = 0
	}
	rec.TotalLatencyMs += latencyMs

	rec.events = append(rec.events, deliveryEvent{
		slot:      slot,
		latencyMs: latencyMs,
		success:   true,
	})
	rt.trimWindow(rec)
	rec.CurrentScore = rt.computeScore(rec)
}

// RecordFailure records a failed delivery for a builder.
func (rt *BuilderReputationTracker) RecordFailure(builderID string, slot uint64, reason string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rec := rt.getOrCreate(builderID)
	rec.Failures++

	rec.events = append(rec.events, deliveryEvent{
		slot:    slot,
		success: false,
		reason:  reason,
	})
	rt.trimWindow(rec)
	rec.CurrentScore = rt.computeScore(rec)
}

// Score returns the current reliability score (0.0 to 1.0) for a builder.
func (rt *BuilderReputationTracker) Score(builderID string) float64 {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	rec, ok := rt.records[builderID]
	if !ok {
		return rt.config.DefaultScore
	}
	return rec.CurrentScore
}

// GetRecord returns a copy of the reputation record for a builder.
func (rt *BuilderReputationTracker) GetRecord(builderID string) (*ReputationRecord, bool) {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	rec, ok := rt.records[builderID]
	if !ok {
		return nil, false
	}
	cp := &ReputationRecord{
		BuilderID:      rec.BuilderID,
		Deliveries:     rec.Deliveries,
		Failures:       rec.Failures,
		TotalLatencyMs: rec.TotalLatencyMs,
		SlashingCount:  rec.SlashingCount,
		CurrentScore:   rec.CurrentScore,
	}
	return cp, true
}

// TopBuilders returns the top n builder IDs sorted by score descending.
func (rt *BuilderReputationTracker) TopBuilders(n int) []string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	if n <= 0 || len(rt.records) == 0 {
		return nil
	}

	type entry struct {
		id    string
		score float64
	}
	entries := make([]entry, 0, len(rt.records))
	for id, rec := range rt.records {
		entries = append(entries, entry{id: id, score: rec.CurrentScore})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].score != entries[j].score {
			return entries[i].score > entries[j].score
		}
		// Deterministic tiebreak by builder ID.
		return entries[i].id < entries[j].id
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

// ApplySlashing applies a slashing event to the builder's record.
func (rt *BuilderReputationTracker) ApplySlashing(event *SlashingEvent) {
	if event == nil {
		return
	}
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rec := rt.getOrCreate(event.BuilderID)
	rec.SlashingCount++
	rec.slashings = append(rec.slashings, *event)
	rec.CurrentScore = rt.computeScore(rec)
}

// DecayScores applies exponential decay, pulling scores toward DefaultScore.
func (rt *BuilderReputationTracker) DecayScores() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	for _, rec := range rt.records {
		// Decay toward DefaultScore.
		rec.CurrentScore = rt.config.DefaultScore +
			(rec.CurrentScore-rt.config.DefaultScore)*rt.config.DecayFactor
		// Clamp to [0, 1].
		if rec.CurrentScore < 0 {
			rec.CurrentScore = 0
		}
		if rec.CurrentScore > 1.0 {
			rec.CurrentScore = 1.0
		}
	}
}

// BuilderCount returns the number of tracked builders.
func (rt *BuilderReputationTracker) BuilderCount() int {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return len(rt.records)
}

// SlashingDetector detects builder misbehavior and accumulates events.
type SlashingDetector struct {
	mu     sync.RWMutex
	events []SlashingEvent
}

// NewSlashingDetector creates a new detector.
func NewSlashingDetector() *SlashingDetector {
	return &SlashingDetector{
		events: make([]SlashingEvent, 0),
	}
}

// CheckLateDelivery returns a SlashingEvent if delivery was late.
func (sd *SlashingDetector) CheckLateDelivery(builderID string, slot uint64, deadlineMs, actualMs int64) *SlashingEvent {
	if actualMs <= deadlineMs {
		return nil
	}
	ev := &SlashingEvent{
		BuilderID:   builderID,
		Slot:        slot,
		Type:        SlashEventLateDelivery,
		Description: "late delivery: actual " + formatMs(actualMs) + " > deadline " + formatMs(deadlineMs),
	}
	sd.mu.Lock()
	sd.events = append(sd.events, *ev)
	sd.mu.Unlock()
	return ev
}

// CheckEquivocation returns a SlashingEvent if hashes differ (equivocation).
func (sd *SlashingDetector) CheckEquivocation(builderID string, slot uint64, hash1, hash2 [32]byte) *SlashingEvent {
	if hash1 == hash2 {
		return nil
	}
	ev := &SlashingEvent{
		BuilderID:    builderID,
		Slot:         slot,
		Type:         SlashEventEquivocation,
		EvidenceHash: xorHashes(hash1, hash2),
		Description:  "equivocation: two different payloads at same slot",
	}
	sd.mu.Lock()
	sd.events = append(sd.events, *ev)
	sd.mu.Unlock()
	return ev
}

// CheckInvalidPayload records an invalid payload event.
func (sd *SlashingDetector) CheckInvalidPayload(builderID string, slot uint64, evidenceHash [32]byte) *SlashingEvent {
	ev := &SlashingEvent{
		BuilderID:    builderID,
		Slot:         slot,
		Type:         SlashEventInvalidPayload,
		EvidenceHash: evidenceHash,
		Description:  "invalid payload delivered",
	}
	sd.mu.Lock()
	sd.events = append(sd.events, *ev)
	sd.mu.Unlock()
	return ev
}

// Events returns a copy of all detected slashing events.
func (sd *SlashingDetector) Events() []SlashingEvent {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	result := make([]SlashingEvent, len(sd.events))
	copy(result, sd.events)
	return result
}

// EventCount returns the number of detected events.
func (sd *SlashingDetector) EventCount() int {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	return len(sd.events)
}

// EventsForBuilder returns all events for a specific builder.
func (sd *SlashingDetector) EventsForBuilder(builderID string) []SlashingEvent {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	var result []SlashingEvent
	for _, ev := range sd.events {
		if ev.BuilderID == builderID {
			result = append(result, ev)
		}
	}
	return result
}

// getOrCreate returns or creates a ReputationRecord. Caller holds write lock.
func (rt *BuilderReputationTracker) getOrCreate(builderID string) *ReputationRecord {
	rec, ok := rt.records[builderID]
	if !ok {
		rec = &ReputationRecord{
			BuilderID:    builderID,
			CurrentScore: rt.config.DefaultScore,
			events:       make([]deliveryEvent, 0, rt.config.WindowSize),
		}
		rt.records[builderID] = rec
	}
	return rec
}

// trimWindow trims the event window. Caller holds write lock.
func (rt *BuilderReputationTracker) trimWindow(rec *ReputationRecord) {
	if len(rec.events) > rt.config.WindowSize {
		rec.events = rec.events[len(rec.events)-rt.config.WindowSize:]
	}
}

// computeScore calculates the reliability score from windowed events.
func (rt *BuilderReputationTracker) computeScore(rec *ReputationRecord) float64 {
	windowEvents := rec.events
	if len(windowEvents) == 0 {
		return rt.config.DefaultScore
	}

	// Count successes in the window.
	successes := 0
	for _, ev := range windowEvents {
		if ev.success {
			successes++
		}
	}

	total := len(windowEvents)
	if total < rt.config.MinDeliveries {
		// Blend default score with observed data.
		observedRate := float64(successes) / float64(total)
		weight := float64(total) / float64(rt.config.MinDeliveries)
		score := weight*observedRate + (1-weight)*rt.config.DefaultScore
		score -= float64(rec.SlashingCount) * rt.config.SlashingPenalty
		return clampScore(score)
	}

	// Full window score: success rate minus slashing penalties.
	score := float64(successes) / float64(total)
	score -= float64(rec.SlashingCount) * rt.config.SlashingPenalty
	return clampScore(score)
}

// clampScore clamps a score to [0.0, 1.0].
func clampScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1.0 {
		return 1.0
	}
	// Round to avoid floating-point noise.
	return math.Round(v*1e12) / 1e12
}

// formatMs converts milliseconds to a string like "1234ms".
func formatMs(ms int64) string {
	s := make([]byte, 0, 16)
	if ms < 0 {
		s = append(s, '-')
		ms = -ms
	}
	if ms == 0 {
		return "0ms"
	}
	digits := make([]byte, 0, 16)
	for ms > 0 {
		digits = append(digits, byte('0'+ms%10))
		ms /= 10
	}
	for i := len(digits) - 1; i >= 0; i-- {
		s = append(s, digits[i])
	}
	return string(append(s, 'm', 's'))
}

// xorHashes XORs two 32-byte hashes to produce a combined evidence hash.
func xorHashes(a, b [32]byte) [32]byte {
	var result [32]byte
	for i := 0; i < 32; i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}
