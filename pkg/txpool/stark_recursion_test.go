package txpool

import (
	"crypto/sha256"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func TestMergeTick_MergesRemoteTxs(t *testing.T) {
	local := NewSTARKAggregator("node-local")
	remote := NewSTARKAggregator("node-remote")

	// Add a local tx.
	var localHash types.Hash
	localHash[0] = 0xAA
	local.AddValidatedTx(localHash, []byte("local-proof"), 21000)

	// Add a remote tx.
	var remoteHash types.Hash
	remoteHash[0] = 0xBB
	remote.AddValidatedTx(remoteHash, []byte("remote-proof"), 42000)

	// Generate tick from remote.
	remoteTick, err := remote.GenerateTick()
	if err != nil {
		t.Fatalf("remote GenerateTick failed: %v", err)
	}

	// Merge remote tick into local.
	if err := local.MergeTick(remoteTick); err != nil {
		t.Fatalf("MergeTick failed: %v", err)
	}

	// Local aggregator should now have both hashes.
	local.mu.RLock()
	_, hasLocal := local.validTxs[localHash]
	_, hasRemote := local.validTxs[remoteHash]
	local.mu.RUnlock()

	if !hasLocal {
		t.Error("local tx hash missing after merge")
	}
	if !hasRemote {
		t.Error("remote tx hash missing after merge")
	}
	if local.PendingCount() != 2 {
		t.Errorf("expected 2 pending txs, got %d", local.PendingCount())
	}
}

func TestMergeTick_RemoteProvenFlag(t *testing.T) {
	local := NewSTARKAggregator("node-local")
	remote := NewSTARKAggregator("node-remote")

	// Add a local tx.
	var localHash types.Hash
	localHash[0] = 0x01
	local.AddValidatedTx(localHash, []byte("proof"), 21000)

	// Add a remote tx.
	var remoteHash types.Hash
	remoteHash[0] = 0x02
	remote.AddValidatedTx(remoteHash, []byte("proof"), 42000)

	remoteTick, err := remote.GenerateTick()
	if err != nil {
		t.Fatal(err)
	}
	if err := local.MergeTick(remoteTick); err != nil {
		t.Fatal(err)
	}

	local.mu.RLock()
	defer local.mu.RUnlock()

	// Local tx should not be remote-proven.
	if local.validTxs[localHash].RemoteProven {
		t.Error("local tx should have RemoteProven=false")
	}

	// Remote tx should be remote-proven.
	if !local.validTxs[remoteHash].RemoteProven {
		t.Error("remote tx should have RemoteProven=true")
	}
}

func TestMergeTick_SkipsDuplicates(t *testing.T) {
	local := NewSTARKAggregator("node-local")
	remote := NewSTARKAggregator("node-remote")

	// Add the same tx hash to both local and remote.
	var sharedHash types.Hash
	sharedHash[0] = 0xFF
	local.AddValidatedTx(sharedHash, []byte("local-proof"), 21000)
	remote.AddValidatedTx(sharedHash, []byte("remote-proof"), 42000)

	remoteTick, err := remote.GenerateTick()
	if err != nil {
		t.Fatal(err)
	}
	if err := local.MergeTick(remoteTick); err != nil {
		t.Fatal(err)
	}

	local.mu.RLock()
	defer local.mu.RUnlock()

	// The local entry should be preserved (not overwritten by remote).
	vtx := local.validTxs[sharedHash]
	if vtx == nil {
		t.Fatal("shared hash should still exist")
	}
	if vtx.RemoteProven {
		t.Error("local entry should NOT be overwritten with RemoteProven=true")
	}
	if vtx.GasUsed != 21000 {
		t.Errorf("expected local GasUsed 21000, got %d", vtx.GasUsed)
	}
}

func TestGenerateTick_Bitfield(t *testing.T) {
	agg := NewSTARKAggregator("node-1")

	for i := 0; i < 5; i++ {
		var h types.Hash
		h[0] = byte(i + 1)
		agg.AddValidatedTx(h, []byte("proof"), uint64(21000*(i+1)))
	}

	tick, err := agg.GenerateTick()
	if err != nil {
		t.Fatal(err)
	}

	// With 5 txs, bitfield should be 1 byte (ceil(5/8) = 1).
	if len(tick.ValidBitfield) != 1 {
		t.Fatalf("expected bitfield length 1, got %d", len(tick.ValidBitfield))
	}

	// Count the set bits. All 5 should be set.
	setBits := 0
	for _, b := range tick.ValidBitfield {
		for b != 0 {
			setBits++
			b &= b - 1
		}
	}
	if setBits != 5 {
		t.Errorf("expected 5 set bits, got %d", setBits)
	}
}

func TestGenerateTick_MerkleRoot(t *testing.T) {
	agg := NewSTARKAggregator("node-1")

	var h types.Hash
	h[0] = 0x42
	agg.AddValidatedTx(h, []byte("proof"), 21000)

	tick, err := agg.GenerateTick()
	if err != nil {
		t.Fatal(err)
	}

	// Merkle root should be non-zero.
	if tick.TxMerkleRoot == (types.Hash{}) {
		t.Error("TxMerkleRoot should be non-zero")
	}
}

func TestMergeTick_BandwidthLimit(t *testing.T) {
	local := NewSTARKAggregator("node-local")
	remote := NewSTARKAggregator("node-remote")

	// Calculate how many tx hashes would exceed MaxTickSize.
	// approxSize = len(hashes)*32 + 1024
	// MaxTickSize = 128*1024 = 131072
	// need: count*32 + 1024 > 131072 => count > (131072-1024)/32 = 4064
	numTxs := (MaxTickSize-1024)/32 + 1 // 4065
	for i := 0; i < numTxs; i++ {
		var h types.Hash
		h[0] = byte(i & 0xFF)
		h[1] = byte((i >> 8) & 0xFF)
		h[2] = byte((i >> 16) & 0xFF)
		remote.AddValidatedTx(h, []byte("proof"), 21000)
	}

	remoteTick, err := remote.GenerateTick()
	if err != nil {
		t.Fatal(err)
	}

	// Merging an oversized tick should return ErrAggTickTooLarge.
	err = local.MergeTick(remoteTick)
	if err != ErrAggTickTooLarge {
		t.Errorf("expected ErrAggTickTooLarge, got %v", err)
	}
}

func TestComputeTxMerkleRoot_SingleHash(t *testing.T) {
	var h types.Hash
	h[0] = 0xDE
	h[1] = 0xAD

	root := computeTxMerkleRoot([]types.Hash{h})

	// A single hash should return the hash itself.
	if root != h {
		t.Errorf("single-hash root should equal the input hash, got %x", root)
	}
}

func TestComputeTxMerkleRoot_TwoHashes(t *testing.T) {
	var h1, h2 types.Hash
	h1[0] = 0x01
	h2[0] = 0x02

	root := computeTxMerkleRoot([]types.Hash{h1, h2})

	// Expected: sha256(h1 || h2).
	hasher := sha256.New()
	hasher.Write(h1[:])
	hasher.Write(h2[:])
	var expected types.Hash
	copy(expected[:], hasher.Sum(nil))

	if root != expected {
		t.Errorf("two-hash root mismatch:\n  got  %x\n  want %x", root, expected)
	}
}
