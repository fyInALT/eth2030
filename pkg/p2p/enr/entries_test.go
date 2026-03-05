package enr

import (
	"bytes"
	"net"
	"testing"
)

// --- IP ---

func TestIPRoundtrip(t *testing.T) {
	r := &Record{}
	want := net.ParseIP("1.2.3.4").To4()
	SetIP(r, want)
	got := IP(r)
	if !bytes.Equal(got, want) {
		t.Fatalf("IP = %v, want %v", got, want)
	}
}

func TestIPAbsent(t *testing.T) {
	r := &Record{}
	if IP(r) != nil {
		t.Fatal("IP should be nil when not set")
	}
}

func TestIPInvalidLength(t *testing.T) {
	r := &Record{}
	r.Set(KeyIP, []byte{1, 2, 3}) // only 3 bytes
	if IP(r) != nil {
		t.Fatal("IP should be nil for non-4-byte value")
	}
}

func TestSetIPIgnoresIPv6(t *testing.T) {
	r := &Record{}
	ipv6 := net.ParseIP("::1") // pure IPv6, To4() returns nil
	SetIP(r, ipv6)
	if IP(r) != nil {
		t.Fatal("SetIP should be a no-op for pure IPv6 address")
	}
}

// --- IP6 ---

func TestIP6Roundtrip(t *testing.T) {
	r := &Record{}
	want := net.ParseIP("2001:db8::1")
	SetIP6(r, want)
	got := IP6(r)
	if !bytes.Equal(got, want.To16()) {
		t.Fatalf("IP6 = %v, want %v", got, want)
	}
}

func TestIP6Absent(t *testing.T) {
	r := &Record{}
	if IP6(r) != nil {
		t.Fatal("IP6 should be nil when not set")
	}
}

func TestIP6InvalidLength(t *testing.T) {
	r := &Record{}
	r.Set(KeyIP6, []byte{1, 2, 3}) // too short
	if IP6(r) != nil {
		t.Fatal("IP6 should be nil for non-16-byte value")
	}
}

// --- TCP ---

func TestTCPRoundtrip(t *testing.T) {
	r := &Record{}
	SetTCP(r, 30303)
	if got := TCP(r); got != 30303 {
		t.Fatalf("TCP = %d, want 30303", got)
	}
}

func TestTCPZero(t *testing.T) {
	r := &Record{}
	SetTCP(r, 0)
	if got := TCP(r); got != 0 {
		t.Fatalf("TCP = %d, want 0", got)
	}
}

func TestTCPMax(t *testing.T) {
	r := &Record{}
	SetTCP(r, 65535)
	if got := TCP(r); got != 65535 {
		t.Fatalf("TCP = %d, want 65535", got)
	}
}

func TestTCPAbsent(t *testing.T) {
	r := &Record{}
	if got := TCP(r); got != 0 {
		t.Fatalf("TCP absent = %d, want 0", got)
	}
}

func TestTCPShortValue(t *testing.T) {
	r := &Record{}
	r.Set(KeyTCP, []byte{0x76}) // only 1 byte
	if got := TCP(r); got != 0 {
		t.Fatalf("TCP short = %d, want 0", got)
	}
}

// --- TCP6 ---

func TestTCP6Roundtrip(t *testing.T) {
	r := &Record{}
	SetTCP6(r, 9000)
	if got := TCP6(r); got != 9000 {
		t.Fatalf("TCP6 = %d, want 9000", got)
	}
}

func TestTCP6Absent(t *testing.T) {
	r := &Record{}
	if got := TCP6(r); got != 0 {
		t.Fatalf("TCP6 absent = %d, want 0", got)
	}
}

// --- UDP ---

func TestUDPRoundtrip(t *testing.T) {
	r := &Record{}
	SetUDP(r, 30303)
	if got := UDP(r); got != 30303 {
		t.Fatalf("UDP = %d, want 30303", got)
	}
}

func TestUDPZero(t *testing.T) {
	r := &Record{}
	SetUDP(r, 0)
	if got := UDP(r); got != 0 {
		t.Fatalf("UDP = %d, want 0", got)
	}
}

func TestUDPMax(t *testing.T) {
	r := &Record{}
	SetUDP(r, 65535)
	if got := UDP(r); got != 65535 {
		t.Fatalf("UDP = %d, want 65535", got)
	}
}

func TestUDPAbsent(t *testing.T) {
	r := &Record{}
	if got := UDP(r); got != 0 {
		t.Fatalf("UDP absent = %d, want 0", got)
	}
}

func TestUDPShortValue(t *testing.T) {
	r := &Record{}
	r.Set(KeyUDP, []byte{0x01})
	if got := UDP(r); got != 0 {
		t.Fatalf("UDP short = %d, want 0", got)
	}
}

// --- UDP6 ---

func TestUDP6Roundtrip(t *testing.T) {
	r := &Record{}
	SetUDP6(r, 9000)
	if got := UDP6(r); got != 9000 {
		t.Fatalf("UDP6 = %d, want 9000", got)
	}
}

func TestUDP6Absent(t *testing.T) {
	r := &Record{}
	if got := UDP6(r); got != 0 {
		t.Fatalf("UDP6 absent = %d, want 0", got)
	}
}

// --- Secp256k1 ---

func TestSecp256k1Present(t *testing.T) {
	r := &Record{}
	key := make([]byte, 33)
	for i := range key {
		key[i] = byte(i + 1)
	}
	r.Set(KeySecp256k1, key)
	got := Secp256k1(r)
	if !bytes.Equal(got, key) {
		t.Fatalf("Secp256k1 mismatch")
	}
}

func TestSecp256k1Absent(t *testing.T) {
	r := &Record{}
	if Secp256k1(r) != nil {
		t.Fatal("Secp256k1 should be nil when not set")
	}
}

func TestSecp256k1WrongLength(t *testing.T) {
	r := &Record{}
	r.Set(KeySecp256k1, make([]byte, 32)) // 32 instead of 33
	if Secp256k1(r) != nil {
		t.Fatal("Secp256k1 should be nil for non-33-byte value")
	}
}

func TestSecp256k1IsCopy(t *testing.T) {
	r := &Record{}
	key := make([]byte, 33)
	r.Set(KeySecp256k1, key)
	got := Secp256k1(r)
	got[0] = 0xFF
	if Secp256k1(r)[0] == 0xFF {
		t.Fatal("Secp256k1 should return a copy, not a reference")
	}
}

// --- EthEntry serialization ---

func TestSerializeDeserializeEthEntry(t *testing.T) {
	e := &EthEntry{
		ForkHash: [4]byte{0xde, 0xad, 0xbe, 0xef},
		ForkNext: 123456789,
	}
	data := SerializeEthEntry(e)
	if len(data) != 12 {
		t.Fatalf("serialized length = %d, want 12", len(data))
	}
	got, err := DeserializeEthEntry(data)
	if err != nil {
		t.Fatalf("DeserializeEthEntry: %v", err)
	}
	if got.ForkHash != e.ForkHash {
		t.Fatalf("ForkHash = %v, want %v", got.ForkHash, e.ForkHash)
	}
	if got.ForkNext != e.ForkNext {
		t.Fatalf("ForkNext = %d, want %d", got.ForkNext, e.ForkNext)
	}
}

func TestDeserializeEthEntryTooShort(t *testing.T) {
	_, err := DeserializeEthEntry(make([]byte, 11))
	if err != ErrInvalidEntry {
		t.Fatalf("expected ErrInvalidEntry, got %v", err)
	}
}

func TestDeserializeEthEntryExactLength(t *testing.T) {
	data := make([]byte, 12)
	e, err := DeserializeEthEntry(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.ForkNext != 0 {
		t.Fatalf("ForkNext = %d, want 0", e.ForkNext)
	}
}

func TestDeserializeEthEntryLongerData(t *testing.T) {
	data := make([]byte, 20) // extra bytes should be ignored
	_, err := DeserializeEthEntry(data)
	if err != nil {
		t.Fatalf("unexpected error for longer data: %v", err)
	}
}

func TestEthEntryForkNextZero(t *testing.T) {
	e := &EthEntry{ForkHash: [4]byte{0x01, 0x02, 0x03, 0x04}, ForkNext: 0}
	data := SerializeEthEntry(e)
	got, err := DeserializeEthEntry(data)
	if err != nil {
		t.Fatalf("DeserializeEthEntry: %v", err)
	}
	if got.ForkNext != 0 {
		t.Fatalf("ForkNext = %d, want 0", got.ForkNext)
	}
}

// --- Eth record entry ---

func TestEthRoundtrip(t *testing.T) {
	r := &Record{}
	e := &EthEntry{
		ForkHash: [4]byte{0xaa, 0xbb, 0xcc, 0xdd},
		ForkNext: 42,
	}
	SetEth(r, e)
	got := Eth(r)
	if got == nil {
		t.Fatal("Eth should not be nil")
	}
	if got.ForkHash != e.ForkHash {
		t.Fatalf("ForkHash = %v, want %v", got.ForkHash, e.ForkHash)
	}
	if got.ForkNext != e.ForkNext {
		t.Fatalf("ForkNext = %d, want %d", got.ForkNext, e.ForkNext)
	}
}

func TestEthAbsent(t *testing.T) {
	r := &Record{}
	if Eth(r) != nil {
		t.Fatal("Eth should be nil when not set")
	}
}

func TestEthInvalidData(t *testing.T) {
	r := &Record{}
	r.Set(KeyEth, []byte{0x01, 0x02}) // too short
	if Eth(r) != nil {
		t.Fatal("Eth should be nil for invalid data")
	}
}

// --- Attnets ---

func TestAttnetsRoundtrip(t *testing.T) {
	r := &Record{}
	bitmap := []byte{0xFF, 0x00, 0xAB, 0xCD, 0x01, 0x02, 0x03, 0x04}
	SetAttnets(r, bitmap)
	got := Attnets(r)
	if !bytes.Equal(got, bitmap) {
		t.Fatalf("Attnets = %v, want %v", got, bitmap)
	}
}

func TestAttnetsAbsent(t *testing.T) {
	r := &Record{}
	if Attnets(r) != nil {
		t.Fatal("Attnets should be nil when not set")
	}
}

func TestAttnetsWrongLength(t *testing.T) {
	r := &Record{}
	r.Set(KeyAttnets, make([]byte, 7)) // 7 instead of 8
	if Attnets(r) != nil {
		t.Fatal("Attnets should be nil for non-8-byte value")
	}
}

func TestSetAttnetsIgnoresWrongLength(t *testing.T) {
	r := &Record{}
	SetAttnets(r, make([]byte, 5))
	if Attnets(r) != nil {
		t.Fatal("SetAttnets should be no-op for non-8-byte bitmap")
	}
}

func TestAttnetsIsCopy(t *testing.T) {
	r := &Record{}
	bitmap := []byte{0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	SetAttnets(r, bitmap)
	got := Attnets(r)
	got[0] = 0x00
	if Attnets(r)[0] != 0xFF {
		t.Fatal("Attnets should return a copy, not a reference")
	}
}

// --- AttnetsSubscribed ---

func TestAttnetsSubscribedFirstSubnet(t *testing.T) {
	r := &Record{}
	bitmap := make([]byte, 8)
	bitmap[0] = 0x01 // subnet 0 is set
	SetAttnets(r, bitmap)
	if !AttnetsSubscribed(r, 0) {
		t.Fatal("expected subnet 0 to be subscribed")
	}
	if AttnetsSubscribed(r, 1) {
		t.Fatal("expected subnet 1 to not be subscribed")
	}
}

func TestAttnetsSubscribedLastSubnet(t *testing.T) {
	r := &Record{}
	bitmap := make([]byte, 8)
	bitmap[7] = 0x80 // subnet 63 is set
	SetAttnets(r, bitmap)
	if !AttnetsSubscribed(r, 63) {
		t.Fatal("expected subnet 63 to be subscribed")
	}
}

func TestAttnetsSubscribedOutOfRange(t *testing.T) {
	r := &Record{}
	bitmap := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	SetAttnets(r, bitmap)
	if AttnetsSubscribed(r, 64) {
		t.Fatal("subnet index 64 is out of range, should return false")
	}
}

func TestAttnetsSubscribedAbsent(t *testing.T) {
	r := &Record{}
	if AttnetsSubscribed(r, 0) {
		t.Fatal("AttnetsSubscribed should return false when not set")
	}
}

func TestAttnetsSubscribedAllSubnets(t *testing.T) {
	r := &Record{}
	bitmap := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	SetAttnets(r, bitmap)
	for i := uint(0); i < 64; i++ {
		if !AttnetsSubscribed(r, i) {
			t.Fatalf("expected subnet %d to be subscribed", i)
		}
	}
}

// --- Syncnets ---

func TestSyncnetsRoundtrip(t *testing.T) {
	r := &Record{}
	SetSyncnets(r, 0b00001111)
	got := Syncnets(r)
	if got == nil || got[0] != 0b00001111 {
		t.Fatalf("Syncnets = %v, want [0x0F]", got)
	}
}

func TestSyncnetsAbsent(t *testing.T) {
	r := &Record{}
	if Syncnets(r) != nil {
		t.Fatal("Syncnets should be nil when not set")
	}
}

func TestSyncnetsWrongLength(t *testing.T) {
	r := &Record{}
	r.Set(KeySyncnets, []byte{0x01, 0x02}) // 2 bytes instead of 1
	if Syncnets(r) != nil {
		t.Fatal("Syncnets should be nil for non-1-byte value")
	}
}

func TestSyncnetsZero(t *testing.T) {
	r := &Record{}
	SetSyncnets(r, 0)
	got := Syncnets(r)
	if got == nil || got[0] != 0 {
		t.Fatalf("Syncnets = %v, want [0x00]", got)
	}
}

// --- SyncnetsSubscribed ---

func TestSyncnetsSubscribedBasic(t *testing.T) {
	r := &Record{}
	SetSyncnets(r, 0b00000101) // subnets 0 and 2
	if !SyncnetsSubscribed(r, 0) {
		t.Fatal("expected subnet 0 to be subscribed")
	}
	if SyncnetsSubscribed(r, 1) {
		t.Fatal("expected subnet 1 to not be subscribed")
	}
	if !SyncnetsSubscribed(r, 2) {
		t.Fatal("expected subnet 2 to be subscribed")
	}
}

func TestSyncnetsSubscribedOutOfRange(t *testing.T) {
	r := &Record{}
	SetSyncnets(r, 0xFF)
	if SyncnetsSubscribed(r, 4) {
		t.Fatal("subnet index 4 is out of range, should return false")
	}
}

func TestSyncnetsSubscribedAbsent(t *testing.T) {
	r := &Record{}
	if SyncnetsSubscribed(r, 0) {
		t.Fatal("SyncnetsSubscribed should return false when not set")
	}
}

func TestSyncnetsSubscribedAllSubnets(t *testing.T) {
	r := &Record{}
	SetSyncnets(r, 0x0F) // all 4 subnets set
	for i := uint(0); i < 4; i++ {
		if !SyncnetsSubscribed(r, i) {
			t.Fatalf("expected subnet %d to be subscribed", i)
		}
	}
}
