package downloader

import (
	"errors"
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// --- helpers ---

func makeHeader(num uint64, parentHash types.Hash) *types.Header {
	h := &types.Header{
		Number:     big.NewInt(int64(num)),
		ParentHash: parentHash,
		Time:       num * 12,
	}
	return h
}

// skeletonHeaderSource is a minimal HeaderSource backed by a slice of headers.
type skeletonHeaderSource struct {
	headers map[uint64]*types.Header
	errAt   uint64 // return error when from >= errAt (0 = no error)
}

func newSkeletonHeaderSource() *skeletonHeaderSource {
	return &skeletonHeaderSource{headers: make(map[uint64]*types.Header)}
}

func (s *skeletonHeaderSource) add(h *types.Header) {
	s.headers[h.Number.Uint64()] = h
}

func (s *skeletonHeaderSource) FetchHeaders(from uint64, count int) ([]*types.Header, error) {
	if s.errAt > 0 && from >= s.errAt {
		return nil, errors.New("fetch error")
	}
	var out []*types.Header
	for i := range count {
		if h, ok := s.headers[from+uint64(i)]; ok {
			out = append(out, h)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("no headers available")
	}
	return out, nil
}

// --- DefaultSkeletonConfig ---

func TestDefaultSkeletonConfig(t *testing.T) {
	cfg := DefaultSkeletonConfig()
	if cfg.Stride != DefaultSkeletonStride {
		t.Errorf("Stride: got %d, want %d", cfg.Stride, DefaultSkeletonStride)
	}
	if cfg.MaxInFlight != DefaultMaxInFlight {
		t.Errorf("MaxInFlight: got %d, want %d", cfg.MaxInFlight, DefaultMaxInFlight)
	}
	if cfg.MaxInFlightBytes != DefaultMaxInFlightBytes {
		t.Errorf("MaxInFlightBytes: got %d, want %d", cfg.MaxInFlightBytes, DefaultMaxInFlightBytes)
	}
	if cfg.ReceiptBatch != DefaultReceiptBatch {
		t.Errorf("ReceiptBatch: got %d, want %d", cfg.ReceiptBatch, DefaultReceiptBatch)
	}
}

// --- ThrottleState.IsThrottled ---

func TestIsThrottled_Tasks(t *testing.T) {
	cfg := SkeletonConfig{MaxInFlight: 2, MaxInFlightBytes: 0}
	ts := ThrottleState{InFlightTasks: 2}
	if !ts.IsThrottled(cfg) {
		t.Fatal("expected throttled when InFlightTasks == MaxInFlight")
	}
	ts.InFlightTasks = 1
	if ts.IsThrottled(cfg) {
		t.Fatal("expected not throttled when InFlightTasks < MaxInFlight")
	}
}

func TestIsThrottled_Bytes(t *testing.T) {
	cfg := SkeletonConfig{MaxInFlight: 0, MaxInFlightBytes: 100}
	ts := ThrottleState{InFlightBytes: 100}
	if !ts.IsThrottled(cfg) {
		t.Fatal("expected throttled when InFlightBytes == MaxInFlightBytes")
	}
	ts.InFlightBytes = 99
	if ts.IsThrottled(cfg) {
		t.Fatal("expected not throttled when InFlightBytes < MaxInFlightBytes")
	}
}

func TestIsThrottled_Zero(t *testing.T) {
	cfg := SkeletonConfig{MaxInFlight: 0, MaxInFlightBytes: 0}
	ts := ThrottleState{InFlightTasks: 100, InFlightBytes: 100}
	if ts.IsThrottled(cfg) {
		t.Fatal("expected not throttled when limits are zero (unlimited)")
	}
}

// --- NewSkeletonChain / AddAnchor / Anchors ---

func TestNewSkeletonChain(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	if len(sc.Anchors()) != 0 {
		t.Fatal("expected empty anchors on new chain")
	}
	if len(sc.Gaps()) != 0 {
		t.Fatal("expected empty gaps on new chain")
	}
}

func TestAddAnchor_Single(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	a := SkeletonAnchor{Number: 100, Hash: types.Hash{0x01}}
	if err := sc.AddAnchor(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	anchors := sc.Anchors()
	if len(anchors) != 1 || anchors[0].Number != 100 {
		t.Fatalf("unexpected anchors: %v", anchors)
	}
}

func TestAddAnchor_Ascending(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	for _, num := range []uint64{200, 100, 300} {
		a := SkeletonAnchor{Number: num, Hash: types.Hash{byte(num)}}
		if err := sc.AddAnchor(a); err != nil {
			t.Fatalf("unexpected error for anchor %d: %v", num, err)
		}
	}
	anchors := sc.Anchors()
	if len(anchors) != 3 {
		t.Fatalf("expected 3 anchors, got %d", len(anchors))
	}
	// Verify ascending order.
	for i := 1; i < len(anchors); i++ {
		if anchors[i].Number <= anchors[i-1].Number {
			t.Fatalf("anchors not sorted: %d >= %d", anchors[i-1].Number, anchors[i].Number)
		}
	}
}

func TestAddAnchor_Overlap(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	a := SkeletonAnchor{Number: 100, Hash: types.Hash{0x01}}
	sc.AddAnchor(a)
	err := sc.AddAnchor(a)
	if !errors.Is(err, ErrSkeletonOverlap) {
		t.Fatalf("expected ErrSkeletonOverlap, got %v", err)
	}
}

// --- Gaps / NextGap ---

func TestGaps_TwoAnchors(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	sc.AddAnchor(SkeletonAnchor{Number: 100})
	sc.AddAnchor(SkeletonAnchor{Number: 200})
	gaps := sc.Gaps()
	if len(gaps) != 1 {
		t.Fatalf("expected 1 gap, got %d", len(gaps))
	}
	if gaps[0].Start != 101 || gaps[0].End != 199 {
		t.Fatalf("unexpected gap: %+v", gaps[0])
	}
	if gaps[0].Filled {
		t.Fatal("new gap should not be filled")
	}
}

func TestGaps_AdjacentAnchors(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	sc.AddAnchor(SkeletonAnchor{Number: 10})
	sc.AddAnchor(SkeletonAnchor{Number: 11})
	// No gap between consecutive blocks.
	if len(sc.Gaps()) != 0 {
		t.Fatalf("expected 0 gaps for adjacent anchors, got %d", len(sc.Gaps()))
	}
}

func TestNextGap_EmptySkeleton(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	_, err := sc.NextGap()
	if !errors.Is(err, ErrSkeletonGapNotFound) {
		t.Fatalf("expected ErrSkeletonGapNotFound, got %v", err)
	}
}

func TestNextGap_Found(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	sc.AddAnchor(SkeletonAnchor{Number: 0})
	sc.AddAnchor(SkeletonAnchor{Number: 10})
	g, err := sc.NextGap()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Start != 1 || g.End != 9 {
		t.Fatalf("unexpected gap: %+v", g)
	}
}

// --- FillHeaders ---

func TestFillHeaders_Basic(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	sc.AddAnchor(SkeletonAnchor{Number: 0})
	sc.AddAnchor(SkeletonAnchor{Number: 3})

	// Build a linked chain for the gap (blocks 1, 2).
	h0 := makeHeader(0, types.Hash{})
	h1 := makeHeader(1, h0.Hash())
	h2 := makeHeader(2, h1.Hash())

	if err := sc.FillHeaders([]*types.Header{h1, h2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sc.FilledCount() != 2 {
		t.Fatalf("expected 2 filled, got %d", sc.FilledCount())
	}
	if sc.FilledHeader(1) == nil {
		t.Fatal("expected header 1 to be filled")
	}
}

func TestFillHeaders_Empty(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	if err := sc.FillHeaders(nil); err != nil {
		t.Fatalf("FillHeaders(nil) should be a no-op")
	}
}

func TestFillHeaders_BadLink(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	h1 := makeHeader(1, types.Hash{0x01})
	h2 := makeHeader(2, types.Hash{0xff}) // wrong parent hash
	err := sc.FillHeaders([]*types.Header{h1, h2})
	if !errors.Is(err, ErrSkeletonBadLink) {
		t.Fatalf("expected ErrSkeletonBadLink, got %v", err)
	}
}

func TestFillHeaders_MarksGapFilled(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	sc.AddAnchor(SkeletonAnchor{Number: 0})
	sc.AddAnchor(SkeletonAnchor{Number: 3})

	h0 := makeHeader(0, types.Hash{})
	h1 := makeHeader(1, h0.Hash())
	h2 := makeHeader(2, h1.Hash())
	sc.FillHeaders([]*types.Header{h1, h2})

	gaps := sc.Gaps()
	if len(gaps) != 1 || !gaps[0].Filled {
		t.Fatalf("expected gap to be marked filled: %+v", gaps)
	}
}

// --- IsComplete ---

func TestIsComplete_NoAnchors(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	if sc.IsComplete() {
		t.Fatal("empty skeleton should not be complete")
	}
}

func TestIsComplete_AfterFill(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	sc.AddAnchor(SkeletonAnchor{Number: 0})
	sc.AddAnchor(SkeletonAnchor{Number: 3})

	h0 := makeHeader(0, types.Hash{})
	h1 := makeHeader(1, h0.Hash())
	h2 := makeHeader(2, h1.Hash())
	sc.FillHeaders([]*types.Header{h1, h2})

	if !sc.IsComplete() {
		t.Fatal("expected skeleton to be complete after all gaps filled")
	}
}

func TestIsComplete_SingleAnchor(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	sc.AddAnchor(SkeletonAnchor{Number: 5})
	// Single anchor has no gaps, so it should be complete.
	if !sc.IsComplete() {
		t.Fatal("single-anchor skeleton with no gaps should be complete")
	}
}

// --- SelectPivotBlock / Pivot ---

func TestSelectPivotBlock_Empty(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	_, err := sc.SelectPivotBlock()
	if !errors.Is(err, ErrSkeletonEmpty) {
		t.Fatalf("expected ErrSkeletonEmpty, got %v", err)
	}
}

func TestSelectPivotBlock_Normal(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	sc.AddAnchor(SkeletonAnchor{Number: 200})
	pivot, err := sc.SelectPivotBlock()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// pivot should be 200 - 64 = 136
	if pivot != 136 {
		t.Fatalf("expected pivot 136, got %d", pivot)
	}
	if sc.Pivot() != pivot {
		t.Fatalf("Pivot() mismatch: got %d, want %d", sc.Pivot(), pivot)
	}
}

func TestSelectPivotBlock_SmallChain(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	sc.AddAnchor(SkeletonAnchor{Number: 10})
	pivot, err := sc.SelectPivotBlock()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// highest (10) <= 64, so pivot should be 1
	if pivot != 1 {
		t.Fatalf("expected pivot 1 for small chain, got %d", pivot)
	}
}

// --- QueueReceiptTask / CompleteReceiptTask / ThrottleStatus ---

func TestQueueReceiptTask_Basic(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	hashes := []types.Hash{{0x01}, {0x02}}
	if err := sc.QueueReceiptTask(hashes, "peer1", 1024); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ts := sc.ThrottleStatus()
	if ts.InFlightTasks != 1 {
		t.Fatalf("expected 1 in-flight task, got %d", ts.InFlightTasks)
	}
	if ts.InFlightBytes != 1024 {
		t.Fatalf("expected 1024 in-flight bytes, got %d", ts.InFlightBytes)
	}
}

func TestQueueReceiptTask_Throttled(t *testing.T) {
	cfg := SkeletonConfig{MaxInFlight: 1, MaxInFlightBytes: 0, ReceiptBatch: 1}
	sc := NewSkeletonChain(cfg)
	sc.QueueReceiptTask([]types.Hash{{0x01}}, "peer1", 0)
	// Second task should hit the task limit.
	err := sc.QueueReceiptTask([]types.Hash{{0x02}}, "peer2", 0)
	if !errors.Is(err, ErrThrottled) {
		t.Fatalf("expected ErrThrottled, got %v", err)
	}
}

func TestCompleteReceiptTask(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	h := types.Hash{0xAB}
	sc.QueueReceiptTask([]types.Hash{h}, "peer1", 512)
	sc.CompleteReceiptTask(h, 512)

	ts := sc.ThrottleStatus()
	if ts.InFlightTasks != 0 {
		t.Fatalf("expected 0 in-flight tasks after complete, got %d", ts.InFlightTasks)
	}
	if ts.InFlightBytes != 0 {
		t.Fatalf("expected 0 in-flight bytes after complete, got %d", ts.InFlightBytes)
	}
}

func TestCompleteReceiptTask_Underflow(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	// Call complete without queuing — should not go negative.
	sc.CompleteReceiptTask(types.Hash{0xCC}, 100)
	ts := sc.ThrottleStatus()
	if ts.InFlightTasks < 0 {
		t.Fatal("InFlightTasks went negative")
	}
	if ts.InFlightBytes < 0 {
		t.Fatal("InFlightBytes went negative")
	}
}

// --- Reset ---

func TestSkeletonChainReset(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	sc.AddAnchor(SkeletonAnchor{Number: 100})
	sc.AddAnchor(SkeletonAnchor{Number: 200})
	sc.QueueReceiptTask([]types.Hash{{0x01}}, "peer1", 100)

	sc.Reset()

	if len(sc.Anchors()) != 0 {
		t.Fatal("expected no anchors after reset")
	}
	if len(sc.Gaps()) != 0 {
		t.Fatal("expected no gaps after reset")
	}
	if sc.FilledCount() != 0 {
		t.Fatal("expected no filled headers after reset")
	}
	ts := sc.ThrottleStatus()
	if ts.InFlightTasks != 0 || ts.InFlightBytes != 0 {
		t.Fatal("expected zero throttle state after reset")
	}
	if sc.Pivot() != 0 {
		t.Fatal("expected pivot 0 after reset")
	}
}

// --- BuildSkeleton ---

func TestBuildSkeleton_Basic(t *testing.T) {
	src := newSkeletonHeaderSource()
	for i := uint64(0); i <= 4096; i++ {
		var parent types.Hash
		if i > 0 {
			parent = makeHeader(i-1, types.Hash{}).Hash()
		}
		src.add(makeHeader(i, parent))
	}

	cfg := DefaultSkeletonConfig()
	cfg.Stride = 1024
	sc := NewSkeletonChain(cfg)
	count, err := sc.BuildSkeleton(0, 2048, src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count == 0 {
		t.Fatal("expected at least one anchor")
	}
	anchors := sc.Anchors()
	if len(anchors) == 0 {
		t.Fatal("expected anchors after BuildSkeleton")
	}
}

func TestBuildSkeleton_InvalidRange(t *testing.T) {
	src := newSkeletonHeaderSource()
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	_, err := sc.BuildSkeleton(100, 50, src)
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("expected ErrInvalidRange, got %v", err)
	}
}

func TestBuildSkeleton_FetchError(t *testing.T) {
	src := newSkeletonHeaderSource()
	src.errAt = 0 // return error on all fetches
	src.errAt = 10
	// We need to trigger an error; set errAt=0 makes it always error.
	src2 := &skeletonHeaderSource{headers: make(map[uint64]*types.Header), errAt: 1}
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	_, err := sc.BuildSkeleton(100, 200, src2)
	if err == nil {
		t.Fatal("expected error from BuildSkeleton when source errors")
	}
}

// --- FilledHeader ---

func TestFilledHeader_Missing(t *testing.T) {
	sc := NewSkeletonChain(DefaultSkeletonConfig())
	if sc.FilledHeader(999) != nil {
		t.Fatal("expected nil for unfilled header")
	}
}
