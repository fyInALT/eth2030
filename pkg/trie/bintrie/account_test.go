package bintrie

import (
	"math/big"
	"testing"
)

func TestPackBasicDataLeaf_RoundTrip(t *testing.T) {
	wantVersion := uint8(1)
	wantCodeSize := uint32(100)
	wantNonce := uint64(5)
	wantBalance := new(big.Int).Mul(big.NewInt(1e18), big.NewInt(1)) // 1 ETH

	packed := PackBasicDataLeaf(wantVersion, wantCodeSize, wantNonce, wantBalance)
	gotVersion, gotCodeSize, gotNonce, gotBalance := UnpackBasicDataLeaf(packed)

	if gotVersion != wantVersion {
		t.Errorf("version: got %d, want %d", gotVersion, wantVersion)
	}
	if gotCodeSize != wantCodeSize {
		t.Errorf("codeSize: got %d, want %d", gotCodeSize, wantCodeSize)
	}
	if gotNonce != wantNonce {
		t.Errorf("nonce: got %d, want %d", gotNonce, wantNonce)
	}
	if gotBalance.Cmp(wantBalance) != 0 {
		t.Errorf("balance: got %s, want %s", gotBalance.String(), wantBalance.String())
	}
}

func TestPackBasicDataLeaf_ByteOffsets(t *testing.T) {
	// version at offset 0 (1 byte), reserved offsets 1-4, code_size at 5-7,
	// nonce at 8-15, balance at 16-31.
	packed := PackBasicDataLeaf(0xFF, 0x00AABBCC, 0x0102030405060708, big.NewInt(0))

	// version
	if packed[0] != 0xFF {
		t.Errorf("offset 0 (version): got 0x%02x, want 0xff", packed[0])
	}
	// reserved bytes 1-4 must be zero
	for i := 1; i <= 4; i++ {
		if packed[i] != 0 {
			t.Errorf("offset %d (reserved): got 0x%02x, want 0x00", i, packed[i])
		}
	}
	// code_size 3 bytes big-endian at offsets 5-7: 0x00AABBCC → bytes [AA BB CC]
	if packed[5] != 0xAA || packed[6] != 0xBB || packed[7] != 0xCC {
		t.Errorf("code_size bytes: got [%02x %02x %02x], want [AA BB CC]",
			packed[5], packed[6], packed[7])
	}
	// nonce 8 bytes big-endian at offsets 8-15: 0x0102030405060708
	if packed[8] != 0x01 || packed[9] != 0x02 || packed[10] != 0x03 || packed[11] != 0x04 {
		t.Errorf("nonce first 4 bytes wrong: %02x %02x %02x %02x",
			packed[8], packed[9], packed[10], packed[11])
	}
}

func TestPackBasicDataLeaf_ZeroValues(t *testing.T) {
	packed := PackBasicDataLeaf(0, 0, 0, big.NewInt(0))
	for i, b := range packed {
		if b != 0 {
			t.Errorf("zero-value packed has non-zero byte at offset %d: 0x%02x", i, b)
		}
	}
}

func TestUnpackBasicDataLeaf_BalanceFits16Bytes(t *testing.T) {
	// Max balance that fits in 16 bytes
	maxBalance := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	packed := PackBasicDataLeaf(0, 0, 0, maxBalance)
	_, _, _, gotBalance := UnpackBasicDataLeaf(packed)
	if gotBalance.Cmp(maxBalance) != 0 {
		t.Errorf("max balance round-trip failed: got %s, want %s",
			gotBalance.String(), maxBalance.String())
	}
}
