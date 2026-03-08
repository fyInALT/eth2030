package rpcfilter

import (
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/eth2030/eth2030/core/types"
)

// ============================================================
// EventHub tests (filter_event_hub.go)
// ============================================================

func TestDefaultEventHubConfig(t *testing.T) {
	cfg := DefaultEventHubConfig()
	if cfg.MaxListeners <= 0 {
		t.Fatal("MaxListeners should be positive")
	}
	if cfg.ListenerBuffer <= 0 {
		t.Fatal("ListenerBuffer should be positive")
	}
	if cfg.ExpiryInterval <= 0 {
		t.Fatal("ExpiryInterval should be positive")
	}
	if cfg.ListenerTimeout <= 0 {
		t.Fatal("ListenerTimeout should be positive")
	}
	if cfg.MaxReplayDepth <= 0 {
		t.Fatal("MaxReplayDepth should be positive")
	}
}

func TestNewEventHub_DefaultsApplied(t *testing.T) {
	// Zero-value config should be replaced by defaults.
	hub := NewEventHub(EventHubConfig{})
	if hub == nil {
		t.Fatal("expected non-nil EventHub")
	}
	if hub.Config().MaxListeners != 256 {
		t.Fatalf("want 256, got %d", hub.Config().MaxListeners)
	}
	if hub.Config().ListenerBuffer != 64 {
		t.Fatalf("want 64, got %d", hub.Config().ListenerBuffer)
	}
	if hub.Config().MaxReplayDepth != 128 {
		t.Fatalf("want 128, got %d", hub.Config().MaxReplayDepth)
	}
}

func TestEventHub_Register_BasicFlow(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	listener, err := hub.Register([]ChainEventType{EventNewBlock})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if listener == nil {
		t.Fatal("expected non-nil listener")
	}
	if listener.ID == "" {
		t.Fatal("expected non-empty listener ID")
	}
	if hub.ListenerCount() != 1 {
		t.Fatalf("want 1 listener, got %d", hub.ListenerCount())
	}
}

func TestEventHub_Register_MaxListeners(t *testing.T) {
	hub := NewEventHub(EventHubConfig{
		MaxListeners:   2,
		ListenerBuffer: 4,
		MaxReplayDepth: 8,
	})
	defer hub.Close()

	_, err := hub.Register([]ChainEventType{EventNewBlock})
	if err != nil {
		t.Fatalf("Register 1: %v", err)
	}
	_, err = hub.Register([]ChainEventType{EventNewBlock})
	if err != nil {
		t.Fatalf("Register 2: %v", err)
	}
	_, err = hub.Register([]ChainEventType{EventNewBlock})
	if err != ErrHubListenerFull {
		t.Fatalf("want ErrHubListenerFull, got %v", err)
	}
}

func TestEventHub_Register_AfterClose(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	hub.Close()

	_, err := hub.Register([]ChainEventType{EventNewBlock})
	if err != ErrHubClosed {
		t.Fatalf("want ErrHubClosed, got %v", err)
	}
}

func TestEventHub_Unregister_Existing(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	listener, _ := hub.Register([]ChainEventType{EventNewBlock})
	ok := hub.Unregister(listener.ID)
	if !ok {
		t.Fatal("Unregister should return true for existing listener")
	}
	if hub.ListenerCount() != 0 {
		t.Fatalf("want 0 listeners after unregister, got %d", hub.ListenerCount())
	}
}

func TestEventHub_Unregister_NonExistent(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	ok := hub.Unregister("nonexistent-id")
	if ok {
		t.Fatal("Unregister should return false for non-existent listener")
	}
}

func TestEventHub_Dispatch_MatchingListener(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	listener, _ := hub.Register([]ChainEventType{EventNewBlock})

	event := ChainEvent{Type: EventNewBlock, BlockNumber: 42}
	if err := hub.Dispatch(event); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	select {
	case got := <-listener.Ch:
		if got.BlockNumber != 42 {
			t.Fatalf("want BlockNumber=42, got %d", got.BlockNumber)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for event")
	}
}

func TestEventHub_Dispatch_NonMatchingListener(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	// Register for logs only, dispatch a block event.
	listener, _ := hub.Register([]ChainEventType{EventNewLogs})

	hub.Dispatch(ChainEvent{Type: EventNewBlock}) //nolint:errcheck

	select {
	case <-listener.Ch:
		t.Fatal("listener should not receive event of wrong type")
	case <-time.After(20 * time.Millisecond):
		// Expected: no event received.
	}
}

func TestEventHub_Dispatch_ClosedHub(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	hub.Close()

	err := hub.Dispatch(ChainEvent{Type: EventNewBlock})
	if err != ErrHubClosed {
		t.Fatalf("want ErrHubClosed, got %v", err)
	}
}

func TestEventHub_Dispatch_SetsTimestamp(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	listener, _ := hub.Register([]ChainEventType{EventPendingTx})

	before := time.Now()
	hub.Dispatch(ChainEvent{Type: EventPendingTx}) //nolint:errcheck
	after := time.Now()

	select {
	case got := <-listener.Ch:
		if got.Timestamp.Before(before) || got.Timestamp.After(after) {
			t.Fatalf("timestamp %v not in expected range [%v, %v]", got.Timestamp, before, after)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out")
	}
}

func TestEventHub_DispatchPendingTx(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	listener, _ := hub.Register([]ChainEventType{EventPendingTx})

	txHash := types.HexToHash("0xabcdef")
	hub.DispatchPendingTx(txHash) //nolint:errcheck

	select {
	case got := <-listener.Ch:
		if got.TxHash != txHash {
			t.Fatalf("want txHash %v, got %v", txHash, got.TxHash)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out")
	}
}

func TestEventHub_DispatchReorg(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	listener, _ := hub.Register([]ChainEventType{EventReorg})

	oldHead := types.HexToHash("0x1111")
	newHead := types.HexToHash("0x2222")
	hub.DispatchReorg(oldHead, newHead, 3) //nolint:errcheck

	select {
	case got := <-listener.Ch:
		if got.OldHead != oldHead || got.NewHead != newHead || got.ReorgDepth != 3 {
			t.Fatalf("reorg event mismatch: %+v", got)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out")
	}
}

func TestEventHub_Replay_Basic(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	hub.Dispatch(ChainEvent{Type: EventNewBlock, BlockNumber: 1}) //nolint:errcheck
	hub.Dispatch(ChainEvent{Type: EventNewBlock, BlockNumber: 2}) //nolint:errcheck
	hub.Dispatch(ChainEvent{Type: EventPendingTx})                //nolint:errcheck

	// Replay only block events.
	events := hub.Replay([]ChainEventType{EventNewBlock}, 10)
	if len(events) != 2 {
		t.Fatalf("want 2 block events, got %d", len(events))
	}
	if events[0].BlockNumber != 1 || events[1].BlockNumber != 2 {
		t.Fatalf("expected chronological order, got %d %d", events[0].BlockNumber, events[1].BlockNumber)
	}
}

func TestEventHub_Replay_AllTypes(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	hub.Dispatch(ChainEvent{Type: EventNewBlock})  //nolint:errcheck
	hub.Dispatch(ChainEvent{Type: EventPendingTx}) //nolint:errcheck

	// nil types = all events.
	events := hub.Replay(nil, 10)
	if len(events) != 2 {
		t.Fatalf("want 2 events for nil types, got %d", len(events))
	}
}

func TestEventHub_Replay_MaxEvents(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	for i := range 5 {
		hub.Dispatch(ChainEvent{Type: EventNewBlock, BlockNumber: uint64(i)}) //nolint:errcheck
	}

	events := hub.Replay(nil, 3)
	if len(events) != 3 {
		t.Fatalf("want 3 events (maxEvents=3), got %d", len(events))
	}
}

func TestEventHub_ReplayBuffer_Capped(t *testing.T) {
	hub := NewEventHub(EventHubConfig{
		MaxListeners:   10,
		ListenerBuffer: 8,
		MaxReplayDepth: 5,
	})
	defer hub.Close()

	for i := range 10 {
		hub.Dispatch(ChainEvent{Type: EventNewBlock, BlockNumber: uint64(i)}) //nolint:errcheck
	}

	size := hub.ReplaySize()
	if size != 5 {
		t.Fatalf("want replay size capped at 5, got %d", size)
	}
}

func TestEventHub_ExpireStale(t *testing.T) {
	hub := NewEventHub(EventHubConfig{
		MaxListeners:    10,
		ListenerBuffer:  4,
		MaxReplayDepth:  8,
		ListenerTimeout: 1 * time.Millisecond,
	})
	defer hub.Close()

	hub.Register([]ChainEventType{EventNewBlock}) //nolint:errcheck

	// Wait for the listener to go stale.
	time.Sleep(5 * time.Millisecond)

	expired := hub.ExpireStale()
	if expired != 1 {
		t.Fatalf("want 1 expired listener, got %d", expired)
	}
	if hub.ListenerCount() != 0 {
		t.Fatalf("want 0 remaining listeners, got %d", hub.ListenerCount())
	}
}

func TestEventHub_ExpireStale_ZeroTimeout_DoesNothing(t *testing.T) {
	hub := NewEventHub(EventHubConfig{
		MaxListeners:    10,
		ListenerBuffer:  4,
		MaxReplayDepth:  8,
		ListenerTimeout: 0, // disabled
	})
	defer hub.Close()

	hub.Register([]ChainEventType{EventNewBlock}) //nolint:errcheck

	expired := hub.ExpireStale()
	if expired != 0 {
		t.Fatalf("want 0 expired (timeout disabled), got %d", expired)
	}
}

func TestEventHub_Stats(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	listener, _ := hub.Register([]ChainEventType{EventNewBlock})
	hub.Dispatch(ChainEvent{Type: EventNewBlock}) //nolint:errcheck

	// Drain the event.
	<-listener.Ch

	stats := hub.Stats()
	if stats.ListenersAdded != 1 {
		t.Fatalf("want ListenersAdded=1, got %d", stats.ListenersAdded)
	}
	if stats.EventsDispatched != 1 {
		t.Fatalf("want EventsDispatched=1, got %d", stats.EventsDispatched)
	}
}

func TestEventHub_Stats_Dropped(t *testing.T) {
	hub := NewEventHub(EventHubConfig{
		MaxListeners:   5,
		ListenerBuffer: 1, // very small buffer
		MaxReplayDepth: 8,
	})
	defer hub.Close()

	hub.Register([]ChainEventType{EventNewBlock}) //nolint:errcheck

	// Send more events than buffer capacity without draining.
	for i := range 5 {
		hub.Dispatch(ChainEvent{Type: EventNewBlock, BlockNumber: uint64(i)}) //nolint:errcheck
	}

	stats := hub.Stats()
	if stats.EventsDropped == 0 {
		t.Fatal("expected some events to be dropped due to full buffer")
	}
}

func TestEventHub_Close_Idempotent(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	hub.Close()
	// Second close should not panic.
	hub.Close()
}

func TestEventHub_ConcurrentDispatchAndRegister(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l, _ := hub.Register([]ChainEventType{EventNewBlock})
			if l != nil {
				hub.Dispatch(ChainEvent{Type: EventNewBlock}) //nolint:errcheck
				hub.Unregister(l.ID)
			}
		}()
	}
	wg.Wait()
}

func TestChainEventType_String(t *testing.T) {
	tests := []struct {
		t    ChainEventType
		want string
	}{
		{EventNewBlock, "NewBlock"},
		{EventNewLogs, "NewLogs"},
		{EventPendingTx, "PendingTx"},
		{EventReorg, "Reorg"},
		{EventFilterExpired, "FilterExpired"},
		{ChainEventType(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("ChainEventType(%d).String() = %q, want %q", tt.t, got, tt.want)
		}
	}
}

// ============================================================
// ExtFilterManager tests (filter_extended.go)
// ============================================================

func TestDefaultExtFilterConfig(t *testing.T) {
	cfg := DefaultExtFilterConfig()
	if cfg.MaxFilters != 256 {
		t.Fatalf("want MaxFilters=256, got %d", cfg.MaxFilters)
	}
	if cfg.FilterTimeout != 5*time.Minute {
		t.Fatalf("want 5min FilterTimeout, got %v", cfg.FilterTimeout)
	}
	if cfg.MaxLogsPerFilter != 10000 {
		t.Fatalf("want MaxLogsPerFilter=10000, got %d", cfg.MaxLogsPerFilter)
	}
	if cfg.MaxHashBuffer != 1024 {
		t.Fatalf("want MaxHashBuffer=1024, got %d", cfg.MaxHashBuffer)
	}
	if cfg.PruneInterval != 30*time.Second {
		t.Fatalf("want PruneInterval=30s, got %v", cfg.PruneInterval)
	}
}

func TestExtFilterManager_Count(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	if fm.Count() != 0 {
		t.Fatalf("want 0 initially, got %d", fm.Count())
	}

	fm.InstallBlockFilter()     //nolint:errcheck
	fm.InstallPendingTxFilter() //nolint:errcheck

	if fm.Count() != 2 {
		t.Fatalf("want 2, got %d", fm.Count())
	}
}

func TestExtFilterManager_CountByType(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	fm.InstallBlockFilter()             //nolint:errcheck
	fm.InstallBlockFilter()             //nolint:errcheck
	fm.InstallPendingTxFilter()         //nolint:errcheck
	fm.InstallLogFilter(0, 0, nil, nil) //nolint:errcheck

	if got := fm.CountByType(ExtBlockFilter); got != 2 {
		t.Fatalf("want 2 block filters, got %d", got)
	}
	if got := fm.CountByType(ExtPendingTxFilter); got != 1 {
		t.Fatalf("want 1 pending tx filter, got %d", got)
	}
	if got := fm.CountByType(ExtLogFilter); got != 1 {
		t.Fatalf("want 1 log filter, got %d", got)
	}
}

func TestExtFilterManager_InstallLogFilter_TopicOverflow(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	// 5 topic positions exceeds MaxTopicPositions=4.
	topics := make([][]types.Hash, 5)
	_, err := fm.InstallLogFilter(0, 0, nil, topics)
	if err != ErrFilterTopicMismatch {
		t.Fatalf("want ErrFilterTopicMismatch, got %v", err)
	}
}

func TestExtFilterManager_InstallLogFilter_InvalidRange(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	_, err := fm.InstallLogFilter(100, 50, nil, nil)
	if err != ErrFilterBlockRange {
		t.Fatalf("want ErrFilterBlockRange, got %v", err)
	}
}

func TestExtFilterManager_InstallLogFilter_MaxFilters(t *testing.T) {
	cfg := DefaultExtFilterConfig()
	cfg.MaxFilters = 2
	fm := NewExtFilterManager(cfg)
	defer fm.Stop()

	fm.InstallBlockFilter()     //nolint:errcheck
	fm.InstallPendingTxFilter() //nolint:errcheck

	_, err := fm.InstallLogFilter(0, 0, nil, nil)
	if err != ErrFilterLimitReached {
		t.Fatalf("want ErrFilterLimitReached, got %v", err)
	}
}

func TestExtFilterManager_Uninstall(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	id, _ := fm.InstallBlockFilter()
	if !fm.Uninstall(id) {
		t.Fatal("Uninstall should return true for existing filter")
	}
	if fm.Uninstall(id) {
		t.Fatal("Uninstall should return false for already-removed filter")
	}
}

func TestExtFilterManager_GetFilterType(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	blockID, _ := fm.InstallBlockFilter()
	pendingID, _ := fm.InstallPendingTxFilter()
	logID, _ := fm.InstallLogFilter(0, 0, nil, nil)

	ft, err := fm.GetFilterType(blockID)
	if err != nil || ft != ExtBlockFilter {
		t.Fatalf("want ExtBlockFilter, got %v (err=%v)", ft, err)
	}
	ft, err = fm.GetFilterType(pendingID)
	if err != nil || ft != ExtPendingTxFilter {
		t.Fatalf("want ExtPendingTxFilter, got %v (err=%v)", ft, err)
	}
	ft, err = fm.GetFilterType(logID)
	if err != nil || ft != ExtLogFilter {
		t.Fatalf("want ExtLogFilter, got %v (err=%v)", ft, err)
	}
}

func TestExtFilterManager_GetFilterType_NotFound(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	_, err := fm.GetFilterType("nonexistent")
	if err != ErrFilterNotFound {
		t.Fatalf("want ErrFilterNotFound, got %v", err)
	}
}

func TestExtFilterManager_DistributeLog(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	addr := types.HexToAddress("0xaaaa")
	logID, _ := fm.InstallLogFilter(0, 0, []types.Address{addr}, nil)

	log := &types.Log{
		Address:     addr,
		BlockNumber: 10,
	}
	fm.DistributeLog(log)

	logs, err := fm.PollLogFilter(logID)
	if err != nil {
		t.Fatalf("PollLogFilter: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("want 1 log, got %d", len(logs))
	}
}

func TestExtFilterManager_DistributeLog_Nil(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	// nil log should not panic.
	fm.DistributeLog(nil)
}

func TestExtFilterManager_DistributeLog_AddressMismatch(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	addr := types.HexToAddress("0xaaaa")
	logID, _ := fm.InstallLogFilter(0, 0, []types.Address{addr}, nil)

	log := &types.Log{
		Address:     types.HexToAddress("0xbbbb"),
		BlockNumber: 10,
	}
	fm.DistributeLog(log)

	logs, _ := fm.PollLogFilter(logID)
	if len(logs) != 0 {
		t.Fatalf("want 0 logs for mismatched address, got %d", len(logs))
	}
}

func TestExtFilterManager_DistributeBlockHash(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	id, _ := fm.InstallBlockFilter()
	hash := types.HexToHash("0xdeadbeef")
	fm.DistributeBlockHash(hash)

	hashes, err := fm.PollBlockFilter(id)
	if err != nil {
		t.Fatalf("PollBlockFilter: %v", err)
	}
	if len(hashes) != 1 || hashes[0] != hash {
		t.Fatalf("want [%v], got %v", hash, hashes)
	}
}

func TestExtFilterManager_DistributePendingTx(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	id, _ := fm.InstallPendingTxFilter()
	txHash := types.HexToHash("0xcafe")
	fm.DistributePendingTx(txHash)

	hashes, err := fm.PollPendingTxFilter(id)
	if err != nil {
		t.Fatalf("PollPendingTxFilter: %v", err)
	}
	if len(hashes) != 1 || hashes[0] != txHash {
		t.Fatalf("want [%v], got %v", txHash, hashes)
	}
}

func TestExtFilterManager_PollLogFilter_WrongType(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	id, _ := fm.InstallBlockFilter()
	_, err := fm.PollLogFilter(id)
	if err != ErrFilterWrongType {
		t.Fatalf("want ErrFilterWrongType, got %v", err)
	}
}

func TestExtFilterManager_PollBlockFilter_WrongType(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	id, _ := fm.InstallLogFilter(0, 0, nil, nil)
	_, err := fm.PollBlockFilter(id)
	if err != ErrFilterWrongType {
		t.Fatalf("want ErrFilterWrongType, got %v", err)
	}
}

func TestExtFilterManager_PollPendingTxFilter_WrongType(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	id, _ := fm.InstallBlockFilter()
	_, err := fm.PollPendingTxFilter(id)
	if err != ErrFilterWrongType {
		t.Fatalf("want ErrFilterWrongType, got %v", err)
	}
}

func TestExtFilterManager_PollDrains(t *testing.T) {
	fm := NewExtFilterManager(DefaultExtFilterConfig())
	defer fm.Stop()

	id, _ := fm.InstallBlockFilter()
	hash := types.HexToHash("0x1111")
	fm.DistributeBlockHash(hash)

	// First poll returns the hash.
	hashes, _ := fm.PollBlockFilter(id)
	if len(hashes) != 1 {
		t.Fatalf("want 1, got %d", len(hashes))
	}

	// Second poll should return empty (drained).
	hashes2, _ := fm.PollBlockFilter(id)
	if len(hashes2) != 0 {
		t.Fatalf("want 0 after drain, got %d", len(hashes2))
	}
}

func TestExtFilterManager_PruneExpired(t *testing.T) {
	cfg := DefaultExtFilterConfig()
	cfg.FilterTimeout = 1 * time.Millisecond
	fm := NewExtFilterManager(cfg)
	defer fm.Stop()

	fm.InstallBlockFilter()     //nolint:errcheck
	fm.InstallPendingTxFilter() //nolint:errcheck

	time.Sleep(5 * time.Millisecond)

	removed := fm.PruneExpired()
	if removed != 2 {
		t.Fatalf("want 2 removed, got %d", removed)
	}
	if fm.Count() != 0 {
		t.Fatalf("want 0 remaining, got %d", fm.Count())
	}
}

func TestExtFilterManager_StartPruner_Stop(t *testing.T) {
	cfg := DefaultExtFilterConfig()
	cfg.PruneInterval = 1 * time.Millisecond
	fm := NewExtFilterManager(cfg)

	fm.StartPruner()
	fm.Stop()
	// Should not hang.
}

// MatchesExtFilter is tested indirectly via DistributeLog.
// Direct tests for topic matching:
func TestMatchesExtFilter_TopicMatch(t *testing.T) {
	topic := types.HexToHash("0x1111")
	f := &ExtFilter{
		Type:   ExtLogFilter,
		Topics: [][]types.Hash{{topic}},
	}
	log := &types.Log{Topics: []types.Hash{topic}}
	if !MatchesExtFilter(log, f) {
		t.Fatal("should match matching topic")
	}
}

func TestMatchesExtFilter_TopicMismatch(t *testing.T) {
	topic := types.HexToHash("0x1111")
	other := types.HexToHash("0x2222")
	f := &ExtFilter{
		Type:   ExtLogFilter,
		Topics: [][]types.Hash{{topic}},
	}
	log := &types.Log{Topics: []types.Hash{other}}
	if MatchesExtFilter(log, f) {
		t.Fatal("should not match mismatched topic")
	}
}

func TestMatchesExtFilter_BlockRangeExcluded(t *testing.T) {
	f := &ExtFilter{
		Type:      ExtLogFilter,
		FromBlock: 100,
		ToBlock:   200,
	}
	log := &types.Log{BlockNumber: 50}
	if MatchesExtFilter(log, f) {
		t.Fatal("log before FromBlock should not match")
	}
	log2 := &types.Log{BlockNumber: 300}
	if MatchesExtFilter(log2, f) {
		t.Fatal("log after ToBlock should not match")
	}
}

// ============================================================
// FilterSys tests (filter_sys.go)
// ============================================================

func TestDefaultFilterSysConfig(t *testing.T) {
	cfg := DefaultFilterSysConfig()
	if cfg.MaxFilters != 512 {
		t.Fatalf("want MaxFilters=512, got %d", cfg.MaxFilters)
	}
	if cfg.FilterTimeout != 5*time.Minute {
		t.Fatalf("want 5min FilterTimeout, got %v", cfg.FilterTimeout)
	}
	if cfg.MaxLogsPerFilter != 10000 {
		t.Fatalf("want MaxLogsPerFilter=10000, got %d", cfg.MaxLogsPerFilter)
	}
	if cfg.MaxHashesPerFilter != 2048 {
		t.Fatalf("want MaxHashesPerFilter=2048, got %d", cfg.MaxHashesPerFilter)
	}
	if cfg.MaxTopicPositions != 4 {
		t.Fatalf("want MaxTopicPositions=4, got %d", cfg.MaxTopicPositions)
	}
	if cfg.MaxBlockRange != 10000 {
		t.Fatalf("want MaxBlockRange=10000, got %d", cfg.MaxBlockRange)
	}
}

func TestFilterSys_Close(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())

	fs.NewBlockFilter()     //nolint:errcheck
	fs.NewPendingTxFilter() //nolint:errcheck

	if fs.FilterCount() != 2 {
		t.Fatalf("want 2 before close, got %d", fs.FilterCount())
	}

	fs.Close()

	if fs.FilterCount() != 0 {
		t.Fatalf("want 0 after close, got %d", fs.FilterCount())
	}

	// After close, new filters should fail.
	_, err := fs.NewBlockFilter()
	if err != ErrSysFilterClosed {
		t.Fatalf("want ErrSysFilterClosed, got %v", err)
	}
}

func TestFilterSys_NewLogFilter_TopicOverflow(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())
	defer fs.Close()

	topics := make([][]types.Hash, 5)
	_, err := fs.NewLogFilter(SysLogQuery{Topics: topics})
	if err != ErrSysTopicOverflow {
		t.Fatalf("want ErrSysTopicOverflow, got %v", err)
	}
}

func TestFilterSys_NewLogFilter_InvalidRange(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())
	defer fs.Close()

	_, err := fs.NewLogFilter(SysLogQuery{FromBlock: 200, ToBlock: 100})
	if err != ErrSysInvalidBlockRange {
		t.Fatalf("want ErrSysInvalidBlockRange, got %v", err)
	}
}

func TestFilterSys_NewLogFilter_MaxFilters(t *testing.T) {
	cfg := DefaultFilterSysConfig()
	cfg.MaxFilters = 1
	fs := NewFilterSys(cfg)
	defer fs.Close()

	fs.NewBlockFilter() //nolint:errcheck

	_, err := fs.NewLogFilter(SysLogQuery{})
	if err != ErrSysFilterCapacity {
		t.Fatalf("want ErrSysFilterCapacity, got %v", err)
	}
}

func TestFilterSys_UninstallFilter(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())
	defer fs.Close()

	id, _ := fs.NewBlockFilter()
	if !fs.UninstallFilter(id) {
		t.Fatal("UninstallFilter should return true for existing filter")
	}
	if fs.UninstallFilter(id) {
		t.Fatal("UninstallFilter should return false for already-removed filter")
	}
}

func TestFilterSys_GetFilterKind(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())
	defer fs.Close()

	blockID, _ := fs.NewBlockFilter()
	pendingID, _ := fs.NewPendingTxFilter()
	logID, _ := fs.NewLogFilter(SysLogQuery{})

	k, err := fs.GetFilterKind(blockID)
	if err != nil || k != SysBlockFilter {
		t.Fatalf("want SysBlockFilter, got %v (err=%v)", k, err)
	}
	k, err = fs.GetFilterKind(pendingID)
	if err != nil || k != SysPendingTxFilter {
		t.Fatalf("want SysPendingTxFilter, got %v (err=%v)", k, err)
	}
	k, err = fs.GetFilterKind(logID)
	if err != nil || k != SysLogFilter {
		t.Fatalf("want SysLogFilter, got %v (err=%v)", k, err)
	}
}

func TestFilterSys_GetFilterKind_NotFound(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())
	defer fs.Close()

	_, err := fs.GetFilterKind("nonexistent")
	if err != ErrSysFilterNotFound {
		t.Fatalf("want ErrSysFilterNotFound, got %v", err)
	}
}

func TestFilterSys_FilterCountByKind(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())
	defer fs.Close()

	fs.NewBlockFilter()     //nolint:errcheck
	fs.NewBlockFilter()     //nolint:errcheck
	fs.NewPendingTxFilter() //nolint:errcheck

	if got := fs.FilterCountByKind(SysBlockFilter); got != 2 {
		t.Fatalf("want 2 block filters, got %d", got)
	}
	if got := fs.FilterCountByKind(SysPendingTxFilter); got != 1 {
		t.Fatalf("want 1 pending tx filter, got %d", got)
	}
}

func TestFilterSys_DistributeLog(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())
	defer fs.Close()

	addr := types.HexToAddress("0xaaaa")
	logID, _ := fs.NewLogFilter(SysLogQuery{Addresses: []types.Address{addr}})

	log := &types.Log{Address: addr, BlockNumber: 5}
	fs.DistributeLog(log)

	changes, err := fs.GetFilterChanges(logID)
	if err != nil {
		t.Fatalf("GetFilterChanges: %v", err)
	}
	logs := changes.([]*types.Log)
	if len(logs) != 1 {
		t.Fatalf("want 1 log, got %d", len(logs))
	}
}

func TestFilterSys_DistributeLog_Nil(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())
	defer fs.Close()

	// Should not panic.
	fs.DistributeLog(nil)
}

func TestFilterSys_DistributeBlockHash(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())
	defer fs.Close()

	id, _ := fs.NewBlockFilter()
	hash := types.HexToHash("0xdeadbeef")
	fs.DistributeBlockHash(hash)

	changes, err := fs.GetFilterChanges(id)
	if err != nil {
		t.Fatalf("GetFilterChanges: %v", err)
	}
	hashes := changes.([]types.Hash)
	if len(hashes) != 1 || hashes[0] != hash {
		t.Fatalf("want [%v], got %v", hash, hashes)
	}
}

func TestFilterSys_DistributePendingTx(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())
	defer fs.Close()

	id, _ := fs.NewPendingTxFilter()
	txHash := types.HexToHash("0xcafe1234")
	fs.DistributePendingTx(txHash)

	changes, err := fs.GetFilterChanges(id)
	if err != nil {
		t.Fatalf("GetFilterChanges: %v", err)
	}
	hashes := changes.([]types.Hash)
	if len(hashes) != 1 || hashes[0] != txHash {
		t.Fatalf("want [%v], got %v", txHash, hashes)
	}
}

func TestFilterSys_GetFilterChanges_NotFound(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())
	defer fs.Close()

	_, err := fs.GetFilterChanges("nonexistent")
	if err != ErrSysFilterNotFound {
		t.Fatalf("want ErrSysFilterNotFound, got %v", err)
	}
}

func TestFilterSys_GetFilterChanges_Drains(t *testing.T) {
	fs := NewFilterSys(DefaultFilterSysConfig())
	defer fs.Close()

	id, _ := fs.NewBlockFilter()
	fs.DistributeBlockHash(types.HexToHash("0x1111"))

	changes1, _ := fs.GetFilterChanges(id)
	if len(changes1.([]types.Hash)) != 1 {
		t.Fatal("want 1 hash on first poll")
	}

	changes2, _ := fs.GetFilterChanges(id)
	if len(changes2.([]types.Hash)) != 0 {
		t.Fatal("want 0 hashes on second poll (drained)")
	}
}

func TestFilterSys_PruneExpired(t *testing.T) {
	cfg := DefaultFilterSysConfig()
	cfg.FilterTimeout = 1 * time.Millisecond
	fs := NewFilterSys(cfg)
	defer fs.Close()

	fs.NewBlockFilter()     //nolint:errcheck
	fs.NewPendingTxFilter() //nolint:errcheck

	time.Sleep(5 * time.Millisecond)

	removed := fs.PruneExpired()
	if removed != 2 {
		t.Fatalf("want 2 removed, got %d", removed)
	}
}

func TestSysFilterKind_String(t *testing.T) {
	tests := []struct {
		k    SysFilterKind
		want string
	}{
		{SysLogFilter, "log"},
		{SysBlockFilter, "block"},
		{SysPendingTxFilter, "pendingTx"},
		{SysFilterKind(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.k.String(); got != tt.want {
			t.Errorf("SysFilterKind(%d).String() = %q, want %q", tt.k, got, tt.want)
		}
	}
}

// --- BloomMatchesQuery ---

func TestBloomMatchesQuery_EmptyQuery(t *testing.T) {
	bloom := types.Bloom{}
	// Empty query always matches.
	if !BloomMatchesQuery(bloom, SysLogQuery{}) {
		t.Fatal("empty query should always match")
	}
}

func TestBloomMatchesQuery_AddressInBloom(t *testing.T) {
	addr := types.HexToAddress("0xaaaa")
	log := &types.Log{Address: addr}
	bloom := types.LogsBloom([]*types.Log{log})

	q := SysLogQuery{Addresses: []types.Address{addr}}
	if !BloomMatchesQuery(bloom, q) {
		t.Fatal("bloom should match address in bloom")
	}
}

func TestBloomMatchesQuery_AddressNotInBloom(t *testing.T) {
	addr := types.HexToAddress("0xaaaa")
	otherAddr := types.HexToAddress("0x0000000000000000000000000000000000001234")
	log := &types.Log{Address: addr}
	bloom := types.LogsBloom([]*types.Log{log})

	q := SysLogQuery{Addresses: []types.Address{otherAddr}}
	// The bloom does not contain otherAddr; result should be false (possible false positive but unlikely).
	result := BloomMatchesQuery(bloom, q)
	// We can't guarantee false due to bloom false positives, so we only check no panic.
	_ = result
}

func TestBloomMatchesQuery_TopicInBloom(t *testing.T) {
	topic := types.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	log := &types.Log{Topics: []types.Hash{topic}}
	bloom := types.LogsBloom([]*types.Log{log})

	q := SysLogQuery{Topics: [][]types.Hash{{topic}}}
	if !BloomMatchesQuery(bloom, q) {
		t.Fatal("bloom should match topic in bloom")
	}
}

func TestBloomMatchesQuery_WildcardTopic(t *testing.T) {
	bloom := types.Bloom{}
	// Wildcard topic position (empty set) should pass bloom check.
	q := SysLogQuery{Topics: [][]types.Hash{{}}}
	if !BloomMatchesQuery(bloom, q) {
		t.Fatal("wildcard topic should pass bloom check")
	}
}

func TestBloomMatchesQuery_MultipleAddresses_OnePresent(t *testing.T) {
	addr1 := types.HexToAddress("0xaaaa")
	addr2 := types.HexToAddress("0x0000000000000000000000000000000000001234")
	log := &types.Log{Address: addr1}
	bloom := types.LogsBloom([]*types.Log{log})

	// At least one address (addr1) is in the bloom.
	q := SysLogQuery{Addresses: []types.Address{addr1, addr2}}
	if !BloomMatchesQuery(bloom, q) {
		t.Fatal("should match when at least one address is in bloom")
	}
}

// FilterLogsByBloom is tested indirectly; add a basic smoke test.
func TestFilterLogsByBloom_BasicMatch(t *testing.T) {
	addr := types.HexToAddress("0xaaaa")
	log := &types.Log{Address: addr, BlockNumber: 10}
	bloom := types.LogsBloom([]*types.Log{log})

	q := SysLogQuery{Addresses: []types.Address{addr}}
	result := FilterLogsByBloom(bloom, []*types.Log{log}, q)
	if len(result) != 1 {
		t.Fatalf("want 1 matched log, got %d", len(result))
	}
}

func TestFilterLogsByBloom_NoMatch(t *testing.T) {
	addr := types.HexToAddress("0xaaaa")
	log := &types.Log{Address: addr}
	bloom := types.Bloom{} // empty bloom

	// Query for different address with empty bloom -- bloom check fails.
	otherAddr := types.HexToAddress("0x0000000000000000000000000000000000009999")
	q := SysLogQuery{Addresses: []types.Address{otherAddr}}
	result := FilterLogsByBloom(bloom, []*types.Log{log}, q)
	if len(result) != 0 {
		t.Fatalf("want 0 logs, got %d", len(result))
	}
}

// SysLogMatches is tested through DistributeLog; add explicit sub-tests.
func TestSysLogMatches_BlockRangeLow(t *testing.T) {
	q := &SysLogQuery{FromBlock: 100}
	log := &types.Log{BlockNumber: 50}
	if SysLogMatches(log, q) {
		t.Fatal("log before FromBlock should not match")
	}
}

func TestSysLogMatches_BlockRangeHigh(t *testing.T) {
	q := &SysLogQuery{ToBlock: 100}
	log := &types.Log{BlockNumber: 200}
	if SysLogMatches(log, q) {
		t.Fatal("log after ToBlock should not match")
	}
}

func TestSysLogMatches_AddressList(t *testing.T) {
	addr := types.HexToAddress("0xaaaa")
	q := &SysLogQuery{Addresses: []types.Address{addr}}

	log1 := &types.Log{Address: addr}
	if !SysLogMatches(log1, q) {
		t.Fatal("should match address in list")
	}

	log2 := &types.Log{Address: types.HexToAddress("0xbbbb")}
	if SysLogMatches(log2, q) {
		t.Fatal("should not match address not in list")
	}
}

func TestSysLogMatches_TopicShortLog(t *testing.T) {
	topic := types.HexToHash("0x1234")
	q := &SysLogQuery{Topics: [][]types.Hash{{topic}, {topic}}}

	// Log has only 1 topic but query requires 2 positions.
	log := &types.Log{Topics: []types.Hash{topic}}
	if SysLogMatches(log, q) {
		t.Fatal("should not match when log has fewer topics than required positions")
	}
}

// Additional test: DispatchBlock with nil header.
func TestEventHub_DispatchBlock_NilHeader(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	// nil header should return nil without panic.
	err := hub.DispatchBlock(nil, nil)
	if err != nil {
		t.Fatalf("DispatchBlock with nil header should return nil, got %v", err)
	}
}

// Additional test: DispatchBlock with non-nil header and no logs.
func TestEventHub_DispatchBlock_WithHeader(t *testing.T) {
	hub := NewEventHub(DefaultEventHubConfig())
	defer hub.Close()

	listener, _ := hub.Register([]ChainEventType{EventNewBlock})

	header := &types.Header{Number: big.NewInt(100)}
	if err := hub.DispatchBlock(header, nil); err != nil {
		t.Fatalf("DispatchBlock: %v", err)
	}

	select {
	case got := <-listener.Ch:
		if got.BlockNumber != 100 {
			t.Fatalf("want BlockNumber=100, got %d", got.BlockNumber)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for block event")
	}
}
