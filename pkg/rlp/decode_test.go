package rlp

import (
	"bytes"
	"math/big"
	"testing"
)

func TestDecodeString(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{"empty string", []byte{0x80}, ""},
		{"dog", []byte{0x83, 0x64, 0x6f, 0x67}, "dog"},
		{"single char 'a'", []byte{0x61}, "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			err := DecodeBytes(tt.input, &got)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDecodeUint64(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  uint64
	}{
		{"uint(0)", []byte{0x80}, 0},
		{"uint(1)", []byte{0x01}, 1},
		{"uint(127)", []byte{0x7f}, 127},
		{"uint(128)", []byte{0x81, 0x80}, 128},
		{"uint(1024)", []byte{0x82, 0x04, 0x00}, 1024},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got uint64
			err := DecodeBytes(tt.input, &got)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDecodeBigInt(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  *big.Int
	}{
		{"big.Int(0)", []byte{0x80}, big.NewInt(0)},
		{"big.Int(1)", []byte{0x01}, big.NewInt(1)},
		{"big.Int(127)", []byte{0x7f}, big.NewInt(127)},
		{"big.Int(128)", []byte{0x81, 0x80}, big.NewInt(128)},
		{"big.Int(1024)", []byte{0x82, 0x04, 0x00}, big.NewInt(1024)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got big.Int
			err := DecodeBytes(tt.input, &got)
			if err != nil {
				t.Fatal(err)
			}
			if got.Cmp(tt.want) != 0 {
				t.Fatalf("got %s, want %s", got.String(), tt.want.String())
			}
		})
	}
}

func TestDecodeBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{"empty", []byte{0x80}, []byte{}},
		{"single zero", []byte{0x00}, []byte{0x00}},
		{"single 0x7f", []byte{0x7f}, []byte{0x7f}},
		{"single 0x80", []byte{0x81, 0x80}, []byte{0x80}},
		{"three bytes", []byte{0x83, 0x01, 0x02, 0x03}, []byte{0x01, 0x02, 0x03}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []byte
			err := DecodeBytes(tt.input, &got)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("got %x, want %x", got, tt.want)
			}
		})
	}
}

func TestDecodeBool(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  bool
	}{
		{"false", []byte{0x80}, false},
		{"true", []byte{0x01}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			err := DecodeBytes(tt.input, &got)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeStruct(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  uint64
	}
	input := []byte{0xc5, 0x83, 0x63, 0x61, 0x74, 0x05}
	var got TestStruct
	err := DecodeBytes(input, &got)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "cat" || got.Age != 5 {
		t.Fatalf("got %+v, want {Name:cat Age:5}", got)
	}
}

func TestDecodeStringSlice(t *testing.T) {
	// ["cat", "dog"]
	input := []byte{0xc8, 0x83, 0x63, 0x61, 0x74, 0x83, 0x64, 0x6f, 0x67}
	var got []string
	err := DecodeBytes(input, &got)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "cat" || got[1] != "dog" {
		t.Fatalf("got %v, want [cat dog]", got)
	}
}

// Round-trip tests: encode then decode.

func TestRoundTripString(t *testing.T) {
	tests := []string{"", "hello", "dog", "a"}
	for _, s := range tests {
		enc, err := EncodeToBytes(s)
		if err != nil {
			t.Fatal(err)
		}
		var dec string
		err = DecodeBytes(enc, &dec)
		if err != nil {
			t.Fatalf("decode %q: %v", s, err)
		}
		if dec != s {
			t.Fatalf("round-trip: got %q, want %q", dec, s)
		}
	}
}

func TestRoundTripUint64(t *testing.T) {
	tests := []uint64{0, 1, 127, 128, 255, 256, 1024, 65535, 1<<32 - 1, 1<<64 - 1}
	for _, u := range tests {
		enc, err := EncodeToBytes(u)
		if err != nil {
			t.Fatal(err)
		}
		var dec uint64
		err = DecodeBytes(enc, &dec)
		if err != nil {
			t.Fatalf("decode %d: %v", u, err)
		}
		if dec != u {
			t.Fatalf("round-trip: got %d, want %d", dec, u)
		}
	}
}

func TestRoundTripBool(t *testing.T) {
	for _, b := range []bool{true, false} {
		enc, err := EncodeToBytes(b)
		if err != nil {
			t.Fatal(err)
		}
		var dec bool
		err = DecodeBytes(enc, &dec)
		if err != nil {
			t.Fatalf("decode %v: %v", b, err)
		}
		if dec != b {
			t.Fatalf("round-trip: got %v, want %v", dec, b)
		}
	}
}

func TestRoundTripBytes(t *testing.T) {
	tests := [][]byte{{}, {0x00}, {0x7f}, {0x80}, {0x01, 0x02, 0x03}}
	for _, b := range tests {
		enc, err := EncodeToBytes(b)
		if err != nil {
			t.Fatal(err)
		}
		var dec []byte
		err = DecodeBytes(enc, &dec)
		if err != nil {
			t.Fatalf("decode %x: %v", b, err)
		}
		if !bytes.Equal(dec, b) {
			t.Fatalf("round-trip: got %x, want %x", dec, b)
		}
	}
}

func TestRoundTripBigInt(t *testing.T) {
	tests := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(127), big.NewInt(128), big.NewInt(1024)}
	for _, bi := range tests {
		enc, err := EncodeToBytes(bi)
		if err != nil {
			t.Fatal(err)
		}
		var dec big.Int
		err = DecodeBytes(enc, &dec)
		if err != nil {
			t.Fatalf("decode %s: %v", bi.String(), err)
		}
		if dec.Cmp(bi) != 0 {
			t.Fatalf("round-trip: got %s, want %s", dec.String(), bi.String())
		}
	}
}

func TestRoundTripStruct(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  uint64
	}
	original := TestStruct{Name: "alice", Age: 30}
	enc, err := EncodeToBytes(original)
	if err != nil {
		t.Fatal(err)
	}
	var dec TestStruct
	err = DecodeBytes(enc, &dec)
	if err != nil {
		t.Fatal(err)
	}
	if dec != original {
		t.Fatalf("round-trip: got %+v, want %+v", dec, original)
	}
}

func TestRoundTripStringSlice(t *testing.T) {
	original := []string{"cat", "dog", "fish"}
	enc, err := EncodeToBytes(original)
	if err != nil {
		t.Fatal(err)
	}
	var dec []string
	err = DecodeBytes(enc, &dec)
	if err != nil {
		t.Fatal(err)
	}
	if len(dec) != len(original) {
		t.Fatalf("length mismatch: got %d, want %d", len(dec), len(original))
	}
	for i := range dec {
		if dec[i] != original[i] {
			t.Fatalf("index %d: got %q, want %q", i, dec[i], original[i])
		}
	}
}

func TestRoundTripLongString(t *testing.T) {
	s := "Lorem ipsum dolor sit amet, consectetur adipisicing elit"
	enc, err := EncodeToBytes(s)
	if err != nil {
		t.Fatal(err)
	}
	var dec string
	err = DecodeBytes(enc, &dec)
	if err != nil {
		t.Fatal(err)
	}
	if dec != s {
		t.Fatalf("round-trip: got %q, want %q", dec, s)
	}
}

// Error cases.

func TestDecodeTruncatedInput(t *testing.T) {
	// A string that claims to be 3 bytes but only has 2.
	input := []byte{0x83, 0x64, 0x6f}
	var got string
	err := DecodeBytes(input, &got)
	if err == nil {
		t.Fatal("expected error for truncated input")
	}
}

func TestDecodeInvalidLengthPrefix(t *testing.T) {
	// Leading zero in length-of-length is non-canonical.
	input := []byte{0xb8, 0x01, 0x61} // claims long string, len=1, but 1 <= 55
	var got string
	err := DecodeBytes(input, &got)
	if err == nil {
		t.Fatal("expected error for non-canonical size")
	}
}

func TestDecodeLeadingZeroUint(t *testing.T) {
	// 0x82, 0x00, 0x80 => uint with a leading zero byte (non-canonical).
	input := []byte{0x82, 0x00, 0x80}
	var got uint64
	err := DecodeBytes(input, &got)
	if err == nil {
		t.Fatal("expected error for non-canonical integer")
	}
}

func TestStreamDirect(t *testing.T) {
	// Test the Stream API directly.
	data := []byte{0x83, 0x64, 0x6f, 0x67} // "dog"
	s := NewStream(bytes.NewReader(data))
	k, size, err := s.Kind()
	if err != nil {
		t.Fatal(err)
	}
	if k != String || size != 3 {
		t.Fatalf("Kind: got (%v, %d), want (String, 3)", k, size)
	}
	b, err := s.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "dog" {
		t.Fatalf("Bytes: got %q, want %q", b, "dog")
	}
}

func TestStreamList(t *testing.T) {
	// ["cat", "dog"]
	data := []byte{0xc8, 0x83, 0x63, 0x61, 0x74, 0x83, 0x64, 0x6f, 0x67}
	s := NewStream(bytes.NewReader(data))
	_, err := s.List()
	if err != nil {
		t.Fatal(err)
	}

	b1, err := s.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if string(b1) != "cat" {
		t.Fatalf("first: got %q, want %q", b1, "cat")
	}

	b2, err := s.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if string(b2) != "dog" {
		t.Fatalf("second: got %q, want %q", b2, "dog")
	}

	err = s.ListEnd()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDecode_TopLevel(t *testing.T) {
	// Decode using the io.Reader-based Decode function.
	enc, err := EncodeToBytes("hello")
	if err != nil {
		t.Fatal(err)
	}
	var got string
	err = Decode(bytes.NewReader(enc), &got)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}

func TestNewStreamFromBytes(t *testing.T) {
	data := []byte{0x83, 0x64, 0x6f, 0x67} // "dog"
	s := NewStreamFromBytes(data)
	b, err := s.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "dog" {
		t.Fatalf("got %q, want %q", b, "dog")
	}
}

func TestAtListEnd(t *testing.T) {
	// At top level with no list scope, AtListEnd is true when all data consumed.
	data := []byte{0x83, 0x64, 0x6f, 0x67} // "dog"
	s := NewStreamFromBytes(data)
	if s.AtListEnd() {
		t.Fatal("should not be at end before consuming")
	}
	_, _ = s.Bytes()
	if !s.AtListEnd() {
		t.Fatal("should be at end after consuming all data")
	}

	// Inside a list scope.
	listData := []byte{0xc8, 0x83, 0x63, 0x61, 0x74, 0x83, 0x64, 0x6f, 0x67}
	s2 := NewStreamFromBytes(listData)
	_, err := s2.List()
	if err != nil {
		t.Fatal(err)
	}
	if s2.AtListEnd() {
		t.Fatal("should not be at list end with items remaining")
	}
	_, _ = s2.Bytes()
	_, _ = s2.Bytes()
	if !s2.AtListEnd() {
		t.Fatal("should be at list end after consuming all items")
	}
}

func TestRawItem(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []byte
	}{
		{"single byte", []byte{0x42}, []byte{0x42}},
		{"short string", []byte{0x83, 0x64, 0x6f, 0x67}, []byte{0x83, 0x64, 0x6f, 0x67}},
		{"empty string", []byte{0x80}, []byte{0x80}},
		{"short list", []byte{0xc3, 0x01, 0x02, 0x03}, []byte{0xc3, 0x01, 0x02, 0x03}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStreamFromBytes(tt.data)
			got, err := s.RawItem()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("got %x, want %x", got, tt.want)
			}
		})
	}
}

func TestRawItem_LongString(t *testing.T) {
	long := make([]byte, 56)
	for i := range long {
		long[i] = byte(i)
	}
	enc, _ := EncodeToBytes(long)
	s := NewStreamFromBytes(enc)
	got, err := s.RawItem()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, enc) {
		t.Fatalf("RawItem long string: got %x, want %x", got, enc)
	}
}

func TestRawItem_LongList(t *testing.T) {
	// Build a list with 56 bytes of payload to trigger long-list encoding.
	items := make([]uint64, 28)
	for i := range items {
		items[i] = uint64(i + 100)
	}
	enc, _ := EncodeToBytes(items)
	s := NewStreamFromBytes(enc)
	got, err := s.RawItem()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, enc) {
		t.Fatalf("RawItem long list mismatch")
	}
}

func TestRawItem_EOF(t *testing.T) {
	s := NewStreamFromBytes([]byte{})
	_, err := s.RawItem()
	if err == nil {
		t.Fatal("expected EOF on empty stream")
	}
}

func TestRawItem_Truncated(t *testing.T) {
	// Short string claiming 3 bytes but only 1 byte present.
	s := NewStreamFromBytes([]byte{0x83, 0x01})
	_, err := s.RawItem()
	if err == nil {
		t.Fatal("expected error on truncated short string")
	}
}

func TestPeekItem(t *testing.T) {
	data := []byte{0x83, 0x64, 0x6f, 0x67} // "dog"
	s := NewStreamFromBytes(data)

	k1, p1, t1, err1 := s.peekItem()
	if err1 != nil {
		t.Fatal(err1)
	}
	k2, p2, t2, err2 := s.peekItem()
	if err2 != nil {
		t.Fatal(err2)
	}
	if k1 != k2 || !bytes.Equal(p1, p2) || t1 != t2 {
		t.Fatal("peekItem is not idempotent")
	}
	if k1 != String {
		t.Fatalf("peekItem kind: got %v, want String", k1)
	}
	if string(p1) != "dog" {
		t.Fatalf("peekItem payload: got %q, want dog", p1)
	}
}

func TestListEnd_WithoutList(t *testing.T) {
	s := NewStreamFromBytes([]byte{0x01})
	err := s.ListEnd()
	if err == nil {
		t.Fatal("expected error calling ListEnd without entering a list")
	}
}

func TestListEnd_NotConsumed(t *testing.T) {
	data := []byte{0xc8, 0x83, 0x63, 0x61, 0x74, 0x83, 0x64, 0x6f, 0x67}
	s := NewStreamFromBytes(data)
	_, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	// Don't read items; ListEnd should fail.
	err = s.ListEnd()
	if err == nil {
		t.Fatal("expected error from ListEnd when items not consumed")
	}
}

func TestStreamKind_EOF(t *testing.T) {
	s := NewStreamFromBytes([]byte{})
	_, _, err := s.Kind()
	if err == nil {
		t.Fatal("expected EOF from empty stream Kind")
	}
}

func TestStreamUint64_Overflow(t *testing.T) {
	// Construct valid long-string encoding of 9 bytes: [0xb8, 0x09, 9 bytes]
	payload := make([]byte, 9)
	payload[0] = 0x01
	enc := append([]byte{0xb8, 0x09}, payload...)
	var got uint64
	err := DecodeBytes(enc, &got)
	if err == nil {
		t.Fatal("expected uint64 overflow error")
	}
}

func TestDecodeByteArray(t *testing.T) {
	enc, err := EncodeToBytes([]byte{0x01, 0x02, 0x03})
	if err != nil {
		t.Fatal(err)
	}
	var arr [3]byte
	err = DecodeBytes(enc, &arr)
	if err != nil {
		t.Fatal(err)
	}
	if arr[0] != 0x01 || arr[1] != 0x02 || arr[2] != 0x03 {
		t.Fatalf("decode byte array: got %v", arr)
	}
}

func TestDecodeNilPtrToStruct(t *testing.T) {
	type S struct{ X uint64 }
	enc, _ := EncodeToBytes(S{X: 7})
	var p *S
	err := DecodeBytes(enc, &p)
	if err != nil {
		t.Fatal(err)
	}
	if p == nil || p.X != 7 {
		t.Fatalf("decode nil ptr to struct: got %v", p)
	}
}

func TestDecodeNilSentinel(t *testing.T) {
	// 0x80 is the nil sentinel for a pointer to a struct (list type).
	type S struct{ X uint64 }
	var p *S
	err := DecodeBytes([]byte{0x80}, &p)
	if err != nil {
		t.Fatal(err)
	}
	if p != nil {
		t.Fatalf("expected nil pointer from nil sentinel, got %v", p)
	}
}

func TestDecodeLongList(t *testing.T) {
	// Round-trip a list with >55 bytes of payload.
	items := make([]uint64, 30)
	for i := range items {
		items[i] = uint64(i * 7)
	}
	enc, err := EncodeToBytes(items)
	if err != nil {
		t.Fatal(err)
	}
	var dec []uint64
	err = DecodeBytes(enc, &dec)
	if err != nil {
		t.Fatal(err)
	}
	if len(dec) != len(items) {
		t.Fatalf("length mismatch: got %d, want %d", len(dec), len(items))
	}
	for i := range items {
		if dec[i] != items[i] {
			t.Fatalf("index %d: got %d, want %d", i, dec[i], items[i])
		}
	}
}

func TestDecodeInt(t *testing.T) {
	enc, _ := EncodeToBytes(int64(99))
	var got int64
	err := DecodeBytes(enc, &got)
	if err != nil {
		t.Fatal(err)
	}
	if got != 99 {
		t.Fatalf("got %d, want 99", got)
	}
}

func TestStreamList_ExpectedList(t *testing.T) {
	// Calling List() on a non-list value should return ErrExpectedList.
	s := NewStreamFromBytes([]byte{0x83, 0x64, 0x6f, 0x67})
	_, err := s.List()
	if err == nil {
		t.Fatal("expected ErrExpectedList")
	}
}
