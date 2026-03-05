package bintrie

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func TestGetTreeKey_SubIndexSetCorrectly(t *testing.T) {
	addr32 := addrTo32(types.Address{0xAA})
	key := GetTreeKey(addr32, big.NewInt(0), 128)
	if key[31] != 128 {
		t.Errorf("sub_index byte should be 128, got %d", key[31])
	}
}

func TestGetTreeKey_DifferentTreeIndexProducesDifferentStem(t *testing.T) {
	addr32 := addrTo32(types.Address{0x11})
	k0 := GetTreeKey(addr32, big.NewInt(0), 0)
	k1 := GetTreeKey(addr32, big.NewInt(1), 0)

	// First 31 bytes should differ (different tree index → different stem).
	for i := 0; i < 31; i++ {
		if k0[i] != k1[i] {
			return // found a difference — test passes
		}
	}
	t.Error("different tree_index must produce different 31-byte stem")
}

func TestGetTreeKeyForBasicData(t *testing.T) {
	addr := types.Address{0x42}
	key := GetTreeKeyForBasicData(addr)

	if key[31] != BasicDataLeafKey {
		t.Errorf("BasicData sub_index should be %d, got %d", BasicDataLeafKey, key[31])
	}

	// Must equal GetTreeKey(addr32, 0, BasicDataLeafKey).
	addr32 := addrTo32(addr)
	expected := GetTreeKey(addr32, big.NewInt(0), BasicDataLeafKey)
	if key != expected {
		t.Error("GetTreeKeyForBasicData doesn't match GetTreeKey(addr32, 0, BasicDataLeafKey)")
	}
}

func TestGetTreeKeyForCodeHash(t *testing.T) {
	addr := types.Address{0x42}
	key := GetTreeKeyForCodeHash(addr)
	if key[31] != CodeHashLeafKey {
		t.Errorf("CodeHash sub_index should be %d, got %d", CodeHashLeafKey, key[31])
	}
}

func TestGetTreeKeyForCodeChunk_ChunkZero(t *testing.T) {
	addr := types.Address{0x01}
	key := GetTreeKeyForCodeChunk(addr, 0)

	// chunk 0: (128 + 0) // 256 = 0, sub = 128 % 256 = 128
	if key[31] != uint8(CodeOffset) {
		t.Errorf("chunk 0 sub_index should be %d (CODE_OFFSET), got %d", CodeOffset, key[31])
	}
}

func TestGetTreeKeyForCodeChunk_Overflow(t *testing.T) {
	addr := types.Address{0x01}
	// chunk 128: (128+128)=256 → tree_index=1, sub=0
	k128 := GetTreeKeyForCodeChunk(addr, 128)
	k0 := GetTreeKeyForCodeChunk(addr, 0)

	// They should have different stems (different tree_index).
	stemsEqual := true
	for i := 0; i < 31; i++ {
		if k128[i] != k0[i] {
			stemsEqual = false
			break
		}
	}
	if stemsEqual {
		t.Error("chunk 128 should be in a different stem than chunk 0")
	}
	// chunk 128 sub_index = (128+128)%256 = 0
	if k128[31] != 0 {
		t.Errorf("chunk 128 sub_index should be 0, got %d", k128[31])
	}
}

func TestGetTreeKeyForStorageSlot_InlineSlot(t *testing.T) {
	addr := types.Address{0x11}
	// Slot 0 → inline: pos = 64 + 0 = 64, tree_index=0, sub=64
	key := GetTreeKeyForStorageSlot(addr, big.NewInt(0))
	if key[31] != uint8(HeaderStorageOffset) {
		t.Errorf("storage slot 0 sub_index should be %d, got %d", HeaderStorageOffset, key[31])
	}
}

func TestGetTreeKeyForStorageSlot_InlineSlot63(t *testing.T) {
	addr := types.Address{0x11}
	// Slot 63 → inline limit: pos = 64+63=127, tree_index=0, sub=127
	key := GetTreeKeyForStorageSlot(addr, big.NewInt(63))
	if key[31] != 127 {
		t.Errorf("storage slot 63 sub_index should be 127, got %d", key[31])
	}
}

func TestGetTreeKeyForStorageSlot_MainStorage(t *testing.T) {
	addr := types.Address{0x11}
	key64 := GetTreeKeyForStorageSlot(addr, big.NewInt(64))
	keyInline63 := GetTreeKeyForStorageSlot(addr, big.NewInt(63))

	// They must be in different stems.
	stemsEqual := true
	for i := 0; i < 31; i++ {
		if key64[i] != keyInline63[i] {
			stemsEqual = false
			break
		}
	}
	if stemsEqual {
		t.Error("slot 64 (main storage) should be in a different stem than slot 63 (inline)")
	}
}

func TestGetTreeKeyUniqueness(t *testing.T) {
	addr := types.Address{0xAB, 0xCD}
	seen := make(map[[32]byte]bool)

	// Generate 200 storage slot keys and verify uniqueness.
	for i := int64(0); i < 200; i++ {
		key := GetTreeKeyForStorageSlot(addr, big.NewInt(i))
		if seen[key] {
			t.Fatalf("duplicate key at slot %d", i)
		}
		seen[key] = true
	}
}

func TestGetTreeKeyForStorageSlot_VsBasicData_Distinct(t *testing.T) {
	addr := types.Address{0x77}
	basicKey := GetTreeKeyForBasicData(addr)
	storageKey := GetTreeKeyForStorageSlot(addr, big.NewInt(0))

	if basicKey == storageKey {
		t.Error("basic data key and storage slot 0 key must be distinct")
	}
}
