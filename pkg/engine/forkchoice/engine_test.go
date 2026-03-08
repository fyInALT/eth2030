package forkchoice

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
	engerrors "github.com/eth2030/eth2030/engine/errors"
	"github.com/eth2030/eth2030/engine/payload"
)

// mockBlockLookup implements BlockLookup for testing.
type mockBlockLookup struct {
	blocks map[[32]byte]mockBlockInfo
}

type mockBlockInfo struct {
	number    uint64
	parent    [32]byte
	timestamp uint64
}

func newMockBlockLookup() *mockBlockLookup {
	return &mockBlockLookup{
		blocks: make(map[[32]byte]mockBlockInfo),
	}
}

func (m *mockBlockLookup) addBlock(hash, parent [32]byte, number, timestamp uint64) {
	m.blocks[hash] = mockBlockInfo{number: number, parent: parent, timestamp: timestamp}
}

func (m *mockBlockLookup) HasBlock(hash [32]byte) bool {
	_, ok := m.blocks[hash]
	return ok
}

func (m *mockBlockLookup) GetBlockNumber(hash [32]byte) (uint64, bool) {
	info, ok := m.blocks[hash]
	if !ok {
		return 0, false
	}
	return info.number, true
}

func (m *mockBlockLookup) GetParentHash(hash [32]byte) ([32]byte, bool) {
	info, ok := m.blocks[hash]
	if !ok {
		return [32]byte{}, false
	}
	return info.parent, true
}

func (m *mockBlockLookup) GetBlockTimestamp(hash [32]byte) (uint64, bool) {
	info, ok := m.blocks[hash]
	if !ok {
		return 0, false
	}
	return info.timestamp, true
}

// buildChain builds a simple chain of blocks for testing: genesis -> block1 -> block2 -> block3.
func buildChain(lookup *mockBlockLookup) (genesis, block1, block2, block3 [32]byte) {
	genesis = [32]byte{0x00, 0x01}
	block1 = [32]byte{0x01, 0x01}
	block2 = [32]byte{0x02, 0x01}
	block3 = [32]byte{0x03, 0x01}

	lookup.addBlock(genesis, [32]byte{}, 0, 1000)
	lookup.addBlock(block1, genesis, 1, 1012)
	lookup.addBlock(block2, block1, 2, 1024)
	lookup.addBlock(block3, block2, 3, 1036)
	return
}

func TestForkchoiceEngine_ProcessForkchoiceUpdate_Valid(t *testing.T) {
	lookup := newMockBlockLookup()
	genesis, block1, block2, block3 := buildChain(lookup)
	_ = genesis

	eng := NewForkchoiceEngine(lookup)

	beaconRoot := types.Hash{0xBE, 0xAC}
	attrs := &payload.PayloadAttributesV3{
		PayloadAttributesV2: payload.PayloadAttributesV2{
			PayloadAttributesV1: payload.PayloadAttributesV1{
				Timestamp:             1048,
				SuggestedFeeRecipient: types.Address{0x01},
			},
		},
		ParentBeaconBlockRoot: beaconRoot,
	}

	resp, err := eng.ProcessForkchoiceUpdate(
		ForkchoiceState{
			HeadBlockHash:      block3,
			SafeBlockHash:      block2,
			FinalizedBlockHash: block1,
		},
		attrs,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PayloadStatus.Status != engerrors.StatusValid {
		t.Errorf("expected VALID status, got %s", resp.PayloadStatus.Status)
	}
	// Payload ID should be set since attributes were provided.
	if resp.PayloadID == (payload.PayloadID{}) {
		t.Error("expected non-zero payload ID")
	}

	// Verify state was updated.
	state := eng.GetState()
	if state.HeadBlockHash != block3 {
		t.Errorf("head not updated: got %x", state.HeadBlockHash)
	}
	if state.SafeBlockHash != block2 {
		t.Errorf("safe not updated: got %x", state.SafeBlockHash)
	}
	if state.FinalizedBlockHash != block1 {
		t.Errorf("finalized not updated: got %x", state.FinalizedBlockHash)
	}
}

func TestForkchoiceEngine_ProcessForkchoiceUpdate_NoAttrs(t *testing.T) {
	lookup := newMockBlockLookup()
	_, block1, block2, block3 := buildChain(lookup)

	eng := NewForkchoiceEngine(lookup)

	resp, err := eng.ProcessForkchoiceUpdate(
		ForkchoiceState{
			HeadBlockHash:      block3,
			SafeBlockHash:      block2,
			FinalizedBlockHash: block1,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PayloadStatus.Status != engerrors.StatusValid {
		t.Errorf("expected VALID, got %s", resp.PayloadStatus.Status)
	}
	if resp.PayloadID != (payload.PayloadID{}) {
		t.Error("expected zero payload ID when no attributes")
	}
}

func TestForkchoiceEngine_ProcessForkchoiceUpdate_ZeroHead(t *testing.T) {
	lookup := newMockBlockLookup()
	eng := NewForkchoiceEngine(lookup)

	_, err := eng.ProcessForkchoiceUpdate(
		ForkchoiceState{HeadBlockHash: [32]byte{}},
		nil,
	)
	if err != ErrFCEHeadZero {
		t.Errorf("expected ErrFCEHeadZero, got %v", err)
	}
}

func TestForkchoiceEngine_ProcessForkchoiceUpdate_UnknownHead(t *testing.T) {
	lookup := newMockBlockLookup()
	eng := NewForkchoiceEngine(lookup)

	unknownHash := [32]byte{0xFF}
	resp, err := eng.ProcessForkchoiceUpdate(
		ForkchoiceState{HeadBlockHash: unknownHash},
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PayloadStatus.Status != engerrors.StatusSyncing {
		t.Errorf("expected SYNCING for unknown head, got %s", resp.PayloadStatus.Status)
	}
}

func TestForkchoiceEngine_ProcessForkchoiceUpdate_Syncing(t *testing.T) {
	lookup := newMockBlockLookup()
	eng := NewForkchoiceEngine(lookup)
	eng.SetSyncing(true)

	unknownHash := [32]byte{0xAA}
	resp, err := eng.ProcessForkchoiceUpdate(
		ForkchoiceState{HeadBlockHash: unknownHash},
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PayloadStatus.Status != engerrors.StatusSyncing {
		t.Errorf("expected SYNCING, got %s", resp.PayloadStatus.Status)
	}
}

func TestForkchoiceEngine_ValidateForkchoiceState(t *testing.T) {
	lookup := newMockBlockLookup()
	genesis, block1, _, _ := buildChain(lookup)

	eng := NewForkchoiceEngine(lookup)

	// Valid state.
	err := eng.ValidateForkchoiceState(ForkchoiceState{
		HeadBlockHash:      block1,
		SafeBlockHash:      genesis,
		FinalizedBlockHash: genesis,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Zero head.
	err = eng.ValidateForkchoiceState(ForkchoiceState{})
	if err != ErrFCEHeadZero {
		t.Errorf("expected ErrFCEHeadZero, got %v", err)
	}

	// Unknown head.
	err = eng.ValidateForkchoiceState(ForkchoiceState{
		HeadBlockHash: [32]byte{0xFF},
	})
	if err != ErrFCEHeadUnknown {
		t.Errorf("expected ErrFCEHeadUnknown, got %v", err)
	}

	// Unknown safe.
	err = eng.ValidateForkchoiceState(ForkchoiceState{
		HeadBlockHash: block1,
		SafeBlockHash: [32]byte{0xFF},
	})
	if err != ErrFCESafeUnknown {
		t.Errorf("expected ErrFCESafeUnknown, got %v", err)
	}
}

func TestForkchoiceEngine_UpdateHead(t *testing.T) {
	lookup := newMockBlockLookup()
	_, block1, _, _ := buildChain(lookup)

	eng := NewForkchoiceEngine(lookup)

	// Update to known block.
	err := eng.UpdateHead(block1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eng.HeadBlock() != block1 {
		t.Error("head not updated")
	}

	// Update to unknown block.
	err = eng.UpdateHead([32]byte{0xFF})
	if err != ErrFCEHeadUnknown {
		t.Errorf("expected ErrFCEHeadUnknown, got %v", err)
	}
}

func TestForkchoiceEngine_UpdateSafe(t *testing.T) {
	lookup := newMockBlockLookup()
	genesis, _, _, _ := buildChain(lookup)

	eng := NewForkchoiceEngine(lookup)

	err := eng.UpdateSafe(genesis)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eng.SafeBlock() != genesis {
		t.Error("safe not updated")
	}

	// Zero hash is allowed.
	err = eng.UpdateSafe([32]byte{})
	if err != nil {
		t.Fatalf("unexpected error for zero hash: %v", err)
	}

	// Unknown hash.
	err = eng.UpdateSafe([32]byte{0xFF})
	if err != ErrFCESafeUnknown {
		t.Errorf("expected ErrFCESafeUnknown, got %v", err)
	}
}

func TestForkchoiceEngine_UpdateFinalized(t *testing.T) {
	lookup := newMockBlockLookup()
	genesis, _, _, _ := buildChain(lookup)

	eng := NewForkchoiceEngine(lookup)

	err := eng.UpdateFinalized(genesis)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eng.FinalizedBlock() != genesis {
		t.Error("finalized not updated")
	}

	err = eng.UpdateFinalized([32]byte{0xFF})
	if err != ErrFCEFinalizedUnknown {
		t.Errorf("expected ErrFCEFinalizedUnknown, got %v", err)
	}
}

func TestForkchoiceEngine_IsValidTransition(t *testing.T) {
	lookup := newMockBlockLookup()
	genesis, block1, block2, _ := buildChain(lookup)

	eng := NewForkchoiceEngine(lookup)

	// Valid transition: genesis -> block1.
	if !eng.IsValidTransition(genesis, block1) {
		t.Error("expected valid transition genesis -> block1")
	}

	// Valid transition: block1 -> block2.
	if !eng.IsValidTransition(block1, block2) {
		t.Error("expected valid transition block1 -> block2")
	}

	// Invalid transition: genesis -> block2 (skip).
	if eng.IsValidTransition(genesis, block2) {
		t.Error("expected invalid transition genesis -> block2")
	}

	// Unknown block.
	if eng.IsValidTransition(genesis, [32]byte{0xFF}) {
		t.Error("expected invalid transition for unknown block")
	}
}

func TestForkchoiceEngine_ShouldBuildPayload(t *testing.T) {
	lookup := newMockBlockLookup()
	_, _, _, block3 := buildChain(lookup)

	eng := NewForkchoiceEngine(lookup)

	beaconRoot := types.Hash{0xBE}
	attrs := &payload.PayloadAttributesV3{
		PayloadAttributesV2: payload.PayloadAttributesV2{
			PayloadAttributesV1: payload.PayloadAttributesV1{
				Timestamp: 2000,
			},
		},
		ParentBeaconBlockRoot: beaconRoot,
	}

	state := ForkchoiceState{HeadBlockHash: block3}

	// Should build: attrs present, head known, timestamp valid.
	if !eng.ShouldBuildPayload(state, attrs) {
		t.Error("expected ShouldBuildPayload=true")
	}

	// No attrs.
	if eng.ShouldBuildPayload(state, nil) {
		t.Error("expected ShouldBuildPayload=false with nil attrs")
	}

	// Zero head.
	if eng.ShouldBuildPayload(ForkchoiceState{}, attrs) {
		t.Error("expected ShouldBuildPayload=false with zero head")
	}

	// Zero timestamp.
	zeroAttrs := &payload.PayloadAttributesV3{}
	if eng.ShouldBuildPayload(state, zeroAttrs) {
		t.Error("expected ShouldBuildPayload=false with zero timestamp")
	}

	// Timestamp not advancing (block3 has timestamp 1036).
	oldAttrs := &payload.PayloadAttributesV3{
		PayloadAttributesV2: payload.PayloadAttributesV2{
			PayloadAttributesV1: payload.PayloadAttributesV1{
				Timestamp: 1000,
			},
		},
		ParentBeaconBlockRoot: beaconRoot,
	}
	if eng.ShouldBuildPayload(state, oldAttrs) {
		t.Error("expected ShouldBuildPayload=false with old timestamp")
	}
}

func TestForkchoiceEngine_Stats(t *testing.T) {
	lookup := newMockBlockLookup()
	_, block1, block2, block3 := buildChain(lookup)

	eng := NewForkchoiceEngine(lookup)

	// Two updates, one with attrs.
	beaconRoot := types.Hash{0xBE}
	eng.ProcessForkchoiceUpdate(
		ForkchoiceState{HeadBlockHash: block3, SafeBlockHash: block2, FinalizedBlockHash: block1},
		nil,
	)
	eng.ProcessForkchoiceUpdate(
		ForkchoiceState{HeadBlockHash: block3, SafeBlockHash: block2, FinalizedBlockHash: block1},
		&payload.PayloadAttributesV3{
			PayloadAttributesV2: payload.PayloadAttributesV2{
				PayloadAttributesV1: payload.PayloadAttributesV1{
					Timestamp: 2000,
				},
			},
			ParentBeaconBlockRoot: beaconRoot,
		},
	)

	updates, builds, _ := eng.Stats()
	if updates != 2 {
		t.Errorf("expected 2 updates, got %d", updates)
	}
	if builds != 1 {
		t.Errorf("expected 1 build, got %d", builds)
	}
}

func TestForkchoiceEngine_HasPayload(t *testing.T) {
	lookup := newMockBlockLookup()
	_, block1, block2, block3 := buildChain(lookup)

	eng := NewForkchoiceEngine(lookup)

	beaconRoot := types.Hash{0xBE}
	resp, err := eng.ProcessForkchoiceUpdate(
		ForkchoiceState{HeadBlockHash: block3, SafeBlockHash: block2, FinalizedBlockHash: block1},
		&payload.PayloadAttributesV3{
			PayloadAttributesV2: payload.PayloadAttributesV2{
				PayloadAttributesV1: payload.PayloadAttributesV1{
					Timestamp: 2000,
				},
			},
			ParentBeaconBlockRoot: beaconRoot,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eng.HasPayload(resp.PayloadID) {
		t.Error("expected payload to be in cache")
	}
	if eng.HasPayload(payload.PayloadID{0xFF}) {
		t.Error("expected unknown payload to not be in cache")
	}
}
