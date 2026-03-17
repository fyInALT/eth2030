// tracker.go manages in-flight payload build lifecycle for the Engine API.
// It connects payload IDs to their build attributes, tracks build status
// (pending/building/ready/failed), handles concurrent forkchoiceUpdated
// calls that may request builds for the same attributes, and integrates
// with the PayloadCache for completed payload storage.
//
// The tracker is implemented using the actor pattern: all operations are
// processed sequentially through a message channel, eliminating lock contention.
package payload

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/actor"
)

// Payload build lifecycle states.
const (
	BuildStatePending  uint8 = iota // Registered but not yet started.
	BuildStateBuilding              // Actively being constructed.
	BuildStateReady                 // Build complete, payload available.
	BuildStateFailed                // Build failed.
	BuildStateExpired               // Evicted by TTL or cache pressure.
)

// BuildStateName returns a human-readable label for a build state.
func BuildStateName(state uint8) string {
	switch state {
	case BuildStatePending:
		return "pending"
	case BuildStateBuilding:
		return "building"
	case BuildStateReady:
		return "ready"
	case BuildStateFailed:
		return "failed"
	case BuildStateExpired:
		return "expired"
	default:
		return fmt.Sprintf("unknown(%d)", state)
	}
}

// Tracker errors.
var (
	ErrPayloadAlreadyTracked = errors.New("payload tracker: payload ID already tracked")
	ErrPayloadNotTracked     = errors.New("payload tracker: payload ID not found")
	ErrPayloadNotReady       = errors.New("payload tracker: payload not yet ready")
	ErrPayloadBuildFailed    = errors.New("payload tracker: build failed")
	ErrTrackerFull           = errors.New("payload tracker: maximum tracked payloads reached")
)

// TrackerConfig configures the PayloadTracker.
type TrackerConfig struct {
	// MaxTracked is the maximum number of simultaneously tracked payloads.
	MaxTracked int
	// BuildTTL is how long a build can stay pending/building before expiry.
	BuildTTL time.Duration
	// CompletedTTL is how long a completed payload is retained.
	CompletedTTL time.Duration
}

// DefaultTrackerConfig returns sensible defaults for the tracker.
func DefaultTrackerConfig() TrackerConfig {
	return TrackerConfig{
		MaxTracked:   64,
		BuildTTL:     30 * time.Second,
		CompletedTTL: 120 * time.Second,
	}
}

// TrackedPayload holds the full lifecycle state for a single payload build.
type TrackedPayload struct {
	ID         PayloadID
	State      uint8
	ParentHash types.Hash
	Attrs      *PayloadAttributesV4
	Result     *BuiltPayload
	Error      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// IsTerminal returns true if the payload is in a terminal state (ready,
// failed, or expired) and will not transition further.
func (tp *TrackedPayload) IsTerminal() bool {
	return tp.State == BuildStateReady ||
		tp.State == BuildStateFailed ||
		tp.State == BuildStateExpired
}

// Age returns the time since the payload was first tracked.
func (tp *TrackedPayload) Age() time.Duration {
	return time.Since(tp.CreatedAt)
}

// attrKey is used for deduplicating builds with identical attributes.
type attrKey struct {
	ParentHash types.Hash
	Timestamp  uint64
}

// --- Message types for actor communication ---

// trackerTrackMsg registers a new payload build.
type trackerTrackMsg struct {
	actor.BaseMessage
	ID         PayloadID
	ParentHash types.Hash
	Attrs      *PayloadAttributesV4
}

// trackerMarkBuildingMsg transitions a payload to building state.
type trackerMarkBuildingMsg struct {
	actor.BaseMessage
	ID PayloadID
}

// trackerMarkReadyMsg transitions a payload to ready state.
type trackerMarkReadyMsg struct {
	actor.BaseMessage
	ID     PayloadID
	Result *BuiltPayload
}

// trackerMarkFailedMsg transitions a payload to failed state.
type trackerMarkFailedMsg struct {
	actor.BaseMessage
	ID     PayloadID
	Reason string
}

// trackerGetMsg retrieves a tracked payload.
type trackerGetMsg struct {
	actor.BaseMessage
	ID PayloadID
}

// trackerGetResultMsg retrieves the built payload result.
type trackerGetResultMsg struct {
	actor.BaseMessage
	ID PayloadID
}

// trackerCountMsg returns the number of tracked payloads.
type trackerCountMsg struct {
	actor.BaseMessage
}

// trackerPruneMsg removes expired entries.
type trackerPruneMsg struct {
	actor.BaseMessage
}

// --- Actor implementation ---

// PayloadTrackerActor is the actor-based implementation of the payload tracker.
type PayloadTrackerActor struct {
	config    TrackerConfig
	entries   map[PayloadID]*TrackedPayload
	attrIndex map[attrKey]PayloadID

	// Inbox receives all tracker operations.
	inbox chan any

	// Statistics.
	trackedCount atomic.Uint64
	readyCount   atomic.Uint64
	failedCount  atomic.Uint64
}

// NewPayloadTrackerActor creates a new actor-based tracker.
func NewPayloadTrackerActor(config TrackerConfig) *PayloadTrackerActor {
	if config.MaxTracked <= 0 {
		config.MaxTracked = DefaultTrackerConfig().MaxTracked
	}
	return &PayloadTrackerActor{
		config:    config,
		entries:   make(map[PayloadID]*TrackedPayload),
		attrIndex: make(map[attrKey]PayloadID),
		inbox:     make(chan any, 64),
	}
}

// Run implements actor.Actor.
func (a *PayloadTrackerActor) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-a.inbox:
			a.handleMessage(msg)
		}
	}
}

// Inbox returns the actor's message channel.
func (a *PayloadTrackerActor) Inbox() chan<- any {
	return a.inbox
}

func (a *PayloadTrackerActor) handleMessage(msg any) {
	switch m := msg.(type) {
	case *trackerTrackMsg:
		id, err := a.track(m.ID, m.ParentHash, m.Attrs)
		m.Reply() <- actor.Reply{Result: id, Error: err}

	case *trackerMarkBuildingMsg:
		err := a.markBuilding(m.ID)
		m.Reply() <- actor.Reply{Error: err}

	case *trackerMarkReadyMsg:
		err := a.markReady(m.ID, m.Result)
		m.Reply() <- actor.Reply{Error: err}

	case *trackerMarkFailedMsg:
		err := a.markFailed(m.ID, m.Reason)
		m.Reply() <- actor.Reply{Error: err}

	case *trackerGetMsg:
		entry, err := a.get(m.ID)
		m.Reply() <- actor.Reply{Result: entry, Error: err}

	case *trackerGetResultMsg:
		result, err := a.getResult(m.ID)
		m.Reply() <- actor.Reply{Result: result, Error: err}

	case *trackerCountMsg:
		m.Reply() <- actor.Reply{Result: len(a.entries)}

	case *trackerPruneMsg:
		count := a.prune()
		m.Reply() <- actor.Reply{Result: count}
	}
}

func (a *PayloadTrackerActor) track(id PayloadID, parentHash types.Hash, attrs *PayloadAttributesV4) (PayloadID, error) {
	// Check for an existing build with the same attributes.
	key := attrKey{ParentHash: parentHash, Timestamp: attrs.Timestamp}
	if existingID, ok := a.attrIndex[key]; ok {
		if existing, found := a.entries[existingID]; found {
			if existing.State != BuildStateFailed && existing.State != BuildStateExpired {
				return existingID, nil
			}
			// Previous build failed/expired; allow re-tracking.
			a.removeEntry(existingID)
		}
	}

	// Enforce capacity, evicting expired entries first.
	a.evictExpired()
	if len(a.entries) >= a.config.MaxTracked {
		a.evictOldestTerminal()
	}
	if len(a.entries) >= a.config.MaxTracked {
		return PayloadID{}, ErrTrackerFull
	}

	now := time.Now()
	a.entries[id] = &TrackedPayload{
		ID:         id,
		State:      BuildStatePending,
		ParentHash: parentHash,
		Attrs:      attrs,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	a.attrIndex[key] = id
	a.trackedCount.Add(1)
	return id, nil
}

func (a *PayloadTrackerActor) markBuilding(id PayloadID) error {
	entry, ok := a.entries[id]
	if !ok {
		return ErrPayloadNotTracked
	}
	if entry.State != BuildStatePending {
		return fmt.Errorf("payload tracker: cannot transition from %s to building",
			BuildStateName(entry.State))
	}
	entry.State = BuildStateBuilding
	entry.UpdatedAt = time.Now()
	return nil
}

func (a *PayloadTrackerActor) markReady(id PayloadID, result *BuiltPayload) error {
	entry, ok := a.entries[id]
	if !ok {
		return ErrPayloadNotTracked
	}
	if entry.State != BuildStatePending && entry.State != BuildStateBuilding {
		return fmt.Errorf("payload tracker: cannot transition from %s to ready",
			BuildStateName(entry.State))
	}
	entry.State = BuildStateReady
	entry.Result = result
	entry.UpdatedAt = time.Now()
	a.readyCount.Add(1)
	return nil
}

func (a *PayloadTrackerActor) markFailed(id PayloadID, reason string) error {
	entry, ok := a.entries[id]
	if !ok {
		return ErrPayloadNotTracked
	}
	entry.State = BuildStateFailed
	entry.Error = reason
	entry.UpdatedAt = time.Now()
	a.failedCount.Add(1)
	return nil
}

func (a *PayloadTrackerActor) get(id PayloadID) (*TrackedPayload, error) {
	entry, ok := a.entries[id]
	if !ok {
		return nil, ErrPayloadNotTracked
	}
	cp := *entry
	return &cp, nil
}

func (a *PayloadTrackerActor) getResult(id PayloadID) (*BuiltPayload, error) {
	entry, ok := a.entries[id]
	if !ok {
		return nil, ErrPayloadNotTracked
	}
	switch entry.State {
	case BuildStateReady:
		return entry.Result, nil
	case BuildStateFailed:
		return nil, fmt.Errorf("%w: %s", ErrPayloadBuildFailed, entry.Error)
	default:
		return nil, ErrPayloadNotReady
	}
}

func (a *PayloadTrackerActor) prune() int {
	return a.evictExpired()
}

func (a *PayloadTrackerActor) evictExpired() int {
	now := time.Now()
	evicted := 0
	for id, entry := range a.entries {
		var ttl time.Duration
		if entry.State == BuildStateReady {
			ttl = a.config.CompletedTTL
		} else {
			ttl = a.config.BuildTTL
		}
		if now.Sub(entry.CreatedAt) > ttl {
			a.removeEntry(id)
			evicted++
		}
	}
	return evicted
}

func (a *PayloadTrackerActor) evictOldestTerminal() {
	var oldestID PayloadID
	var oldestTime time.Time
	found := false

	for id, entry := range a.entries {
		if !entry.IsTerminal() {
			continue
		}
		if !found || entry.CreatedAt.Before(oldestTime) {
			oldestID = id
			oldestTime = entry.CreatedAt
			found = true
		}
	}
	if found {
		a.removeEntry(oldestID)
	}
}

func (a *PayloadTrackerActor) removeEntry(id PayloadID) {
	entry, ok := a.entries[id]
	if !ok {
		return
	}
	if entry.Attrs != nil {
		key := attrKey{ParentHash: entry.ParentHash, Timestamp: entry.Attrs.Timestamp}
		if indexed, exists := a.attrIndex[key]; exists && indexed == id {
			delete(a.attrIndex, key)
		}
	}
	delete(a.entries, id)
}

// --- Compatibility wrapper ---

// PayloadTracker manages the lifecycle of in-flight payload builds.
// It wraps PayloadTrackerActor for API compatibility.
type PayloadTracker struct {
	actor   *PayloadTrackerActor
	ctx     context.Context
	cancel  context.CancelFunc
	timeout time.Duration
}

// NewPayloadTracker creates a new tracker with actor backend.
func NewPayloadTracker(config TrackerConfig) *PayloadTracker {
	ctx, cancel := context.WithCancel(context.Background())
	act := NewPayloadTrackerActor(config)
	go act.Run(ctx)

	return &PayloadTracker{
		actor:   act,
		ctx:     ctx,
		cancel:  cancel,
		timeout: actor.DefaultTimeout,
	}
}

// Track registers a new payload build.
func (pt *PayloadTracker) Track(id PayloadID, parentHash types.Hash, attrs *PayloadAttributesV4) (PayloadID, error) {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](pt.actor.Inbox(), &trackerTrackMsg{BaseMessage: msg, ID: id, ParentHash: parentHash, Attrs: attrs}, pt.timeout); err != nil {
		return PayloadID{}, err
	}
	return actor.CallResult[PayloadID](replyCh, pt.timeout)
}

// MarkBuilding transitions a payload from pending to building.
func (pt *PayloadTracker) MarkBuilding(id PayloadID) error {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](pt.actor.Inbox(), &trackerMarkBuildingMsg{BaseMessage: msg, ID: id}, pt.timeout); err != nil {
		return err
	}
	_, err := actor.CallResult[struct{}](replyCh, pt.timeout)
	return err
}

// MarkReady transitions a payload to ready and stores the built result.
func (pt *PayloadTracker) MarkReady(id PayloadID, result *BuiltPayload) error {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](pt.actor.Inbox(), &trackerMarkReadyMsg{BaseMessage: msg, ID: id, Result: result}, pt.timeout); err != nil {
		return err
	}
	_, err := actor.CallResult[struct{}](replyCh, pt.timeout)
	return err
}

// MarkFailed transitions a payload to failed with an error message.
func (pt *PayloadTracker) MarkFailed(id PayloadID, reason string) error {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](pt.actor.Inbox(), &trackerMarkFailedMsg{BaseMessage: msg, ID: id, Reason: reason}, pt.timeout); err != nil {
		return err
	}
	_, err := actor.CallResult[struct{}](replyCh, pt.timeout)
	return err
}

// Get retrieves a tracked payload by ID.
func (pt *PayloadTracker) Get(id PayloadID) (*TrackedPayload, error) {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](pt.actor.Inbox(), &trackerGetMsg{BaseMessage: msg, ID: id}, pt.timeout); err != nil {
		return nil, err
	}
	result, err := actor.CallResult[*TrackedPayload](replyCh, pt.timeout)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, ErrPayloadNotTracked
	}
	return result, nil
}

// GetResult retrieves the built payload result.
func (pt *PayloadTracker) GetResult(id PayloadID) (*BuiltPayload, error) {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](pt.actor.Inbox(), &trackerGetResultMsg{BaseMessage: msg, ID: id}, pt.timeout); err != nil {
		return nil, err
	}
	result, err := actor.CallResult[*BuiltPayload](replyCh, pt.timeout)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, ErrPayloadNotReady
	}
	return result, nil
}

// Count returns the number of currently tracked payloads.
func (pt *PayloadTracker) Count() int {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](pt.actor.Inbox(), &trackerCountMsg{BaseMessage: msg}, pt.timeout); err != nil {
		return 0
	}
	result, err := actor.CallResult[int](replyCh, pt.timeout)
	if err != nil {
		return 0
	}
	return result
}

// Prune removes all expired entries.
func (pt *PayloadTracker) Prune() int {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](pt.actor.Inbox(), &trackerPruneMsg{BaseMessage: msg}, pt.timeout); err != nil {
		return 0
	}
	result, err := actor.CallResult[int](replyCh, pt.timeout)
	if err != nil {
		return 0
	}
	return result
}

// Close stops the actor goroutine.
func (pt *PayloadTracker) Close() {
	pt.cancel()
}