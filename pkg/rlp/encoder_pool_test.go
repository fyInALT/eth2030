package rlp

import (
	"bytes"
	"sync"
	"testing"
)

func TestNewEncoderPool(t *testing.T) {
	ep := NewEncoderPool()
	if ep == nil {
		t.Fatal("NewEncoderPool returned nil")
	}
	m := ep.Metrics()
	if m == nil {
		t.Fatal("Metrics returned nil")
	}
}

func TestEncoderPoolMetricsSnapshot(t *testing.T) {
	ep := NewEncoderPool()
	snap := ep.Metrics().Snapshot()
	if snap.PoolHits != 0 || snap.PoolMisses != 0 || snap.TotalEncodes != 0 || snap.TotalBytes != 0 {
		t.Fatalf("initial snapshot should be zero: %+v", snap)
	}
}

func TestEncoderPoolEncodeBytes(t *testing.T) {
	ep := NewEncoderPool()

	tests := []struct {
		name string
		val  any
		want []byte
	}{
		{"uint64 zero", uint64(0), []byte{0x80}},
		{"uint64 one", uint64(1), []byte{0x01}},
		{"string dog", "dog", []byte{0x83, 0x64, 0x6f, 0x67}},
		{"empty string", "", []byte{0x80}},
		{"bytes", []byte{0x01, 0x02}, []byte{0x82, 0x01, 0x02}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ep.EncodeBytes(tt.val)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("got %x, want %x", got, tt.want)
			}
		})
	}

	snap := ep.Metrics().Snapshot()
	if snap.TotalEncodes != int64(len(tests)) {
		t.Fatalf("TotalEncodes: got %d, want %d", snap.TotalEncodes, len(tests))
	}
	if snap.TotalBytes == 0 {
		t.Fatal("TotalBytes should be nonzero after encodes")
	}
}

func TestEncoderPoolEncodeBytes_Error(t *testing.T) {
	ep := NewEncoderPool()
	// channel cannot be encoded by rlp
	_, err := ep.EncodeBytes(make(chan int))
	if err == nil {
		t.Fatal("expected error encoding channel")
	}
	snap := ep.Metrics().Snapshot()
	if snap.TotalEncodes != 0 {
		t.Fatalf("TotalEncodes should not increase on error: got %d", snap.TotalEncodes)
	}
}

func TestEncoderPoolEncodeBatch(t *testing.T) {
	ep := NewEncoderPool()

	items := []any{"cat", "dog"}
	got, err := ep.EncodeBatch(items)
	if err != nil {
		t.Fatal(err)
	}

	// Manually build the expected result: WrapList(encode("cat") + encode("dog"))
	encCat, _ := EncodeToBytes("cat")
	encDog, _ := EncodeToBytes("dog")
	payload := append(encCat, encDog...)
	want := WrapList(payload)

	if !bytes.Equal(got, want) {
		t.Fatalf("EncodeBatch: got %x, want %x", got, want)
	}

	snap := ep.Metrics().Snapshot()
	if snap.TotalEncodes != int64(len(items)) {
		t.Fatalf("TotalEncodes: got %d, want %d", snap.TotalEncodes, len(items))
	}
	if snap.TotalBytes != int64(len(got)) {
		t.Fatalf("TotalBytes: got %d, want %d", snap.TotalBytes, len(got))
	}
}

func TestEncoderPoolEncodeBatch_Empty(t *testing.T) {
	ep := NewEncoderPool()
	got, err := ep.EncodeBatch(nil)
	if err != nil {
		t.Fatal(err)
	}
	// empty list
	want := []byte{0xc0}
	if !bytes.Equal(got, want) {
		t.Fatalf("empty batch: got %x, want %x", got, want)
	}
}

func TestEncoderPoolEncodeBatch_Error(t *testing.T) {
	ep := NewEncoderPool()
	items := []any{"ok", make(chan int)}
	_, err := ep.EncodeBatch(items)
	if err == nil {
		t.Fatal("expected error when batch contains unencodable item")
	}
}

func TestEncodeUint64_Values(t *testing.T) {
	tests := []struct {
		val  uint64
		want []byte
	}{
		{0, []byte{0x80}},
		{1, []byte{0x01}},
		{127, []byte{0x7f}},
		{128, []byte{0x81, 0x80}},
		{255, []byte{0x81, 0xff}},
		{256, []byte{0x82, 0x01, 0x00}},
		{1024, []byte{0x82, 0x04, 0x00}},
		{0xffffff, []byte{0x83, 0xff, 0xff, 0xff}},
		{0xffffffff, []byte{0x84, 0xff, 0xff, 0xff, 0xff}},
		{0xffffffffff, []byte{0x85, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{0xffffffffffff, []byte{0x86, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{0xffffffffffffff, []byte{0x87, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{0xffffffffffffffff, []byte{0x88, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}
	for _, tt := range tests {
		got := EncodeUint64(tt.val)
		if !bytes.Equal(got, tt.want) {
			t.Fatalf("EncodeUint64(%d): got %x, want %x", tt.val, got, tt.want)
		}
	}
}

func TestEncoderPoolConcurrency(t *testing.T) {
	ep := NewEncoderPool()
	var wg sync.WaitGroup
	const goroutines = 20

	for n := range goroutines {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			val := uint64(n)
			got, err := ep.EncodeBytes(val)
			if err != nil {
				t.Errorf("goroutine %d: %v", n, err)
				return
			}
			want, _ := EncodeToBytes(val)
			if !bytes.Equal(got, want) {
				t.Errorf("goroutine %d: got %x, want %x", n, got, want)
			}
		}(n)
	}
	wg.Wait()

	snap := ep.Metrics().Snapshot()
	if snap.TotalEncodes != goroutines {
		t.Fatalf("TotalEncodes: got %d, want %d", snap.TotalEncodes, goroutines)
	}
}

func TestEncoderPoolPutOversizedBuffer(t *testing.T) {
	ep := NewEncoderPool()
	// A buffer larger than maxBufSize should not be returned to the pool.
	oversized := &encoderBuf{data: make([]byte, 0, maxBufSize+1)}
	ep.put(oversized) // should be a no-op, not panic
}

func TestEncodeBytes32(t *testing.T) {
	var data [32]byte
	for i := range data {
		data[i] = byte(i)
	}
	got := EncodeBytes32(data)
	if len(got) != 33 {
		t.Fatalf("EncodeBytes32 length: got %d, want 33", len(got))
	}
	if got[0] != 0xa0 {
		t.Fatalf("EncodeBytes32 prefix: got %x, want 0xa0", got[0])
	}
	if !bytes.Equal(got[1:], data[:]) {
		t.Fatal("EncodeBytes32 data mismatch")
	}
}

func TestEncodeBytes20(t *testing.T) {
	var data [20]byte
	for i := range data {
		data[i] = byte(i + 1)
	}
	got := EncodeBytes20(data)
	if len(got) != 21 {
		t.Fatalf("EncodeBytes20 length: got %d, want 21", len(got))
	}
	if got[0] != 0x94 {
		t.Fatalf("EncodeBytes20 prefix: got %x, want 0x94", got[0])
	}
	if !bytes.Equal(got[1:], data[:]) {
		t.Fatal("EncodeBytes20 data mismatch")
	}
}

func TestEncodeBoolFunc(t *testing.T) {
	if !bytes.Equal(EncodeBool(true), []byte{0x01}) {
		t.Fatal("EncodeBool(true) != 0x01")
	}
	if !bytes.Equal(EncodeBool(false), []byte{0x80}) {
		t.Fatal("EncodeBool(false) != 0x80")
	}
}

func TestEstimateListSize(t *testing.T) {
	// payload <= 55: 1 + payload
	if got := EstimateListSize(0); got != 1 {
		t.Fatalf("EstimateListSize(0): got %d, want 1", got)
	}
	if got := EstimateListSize(55); got != 56 {
		t.Fatalf("EstimateListSize(55): got %d, want 56", got)
	}
	// payload 56: 1 + 1 + 56 = 58
	if got := EstimateListSize(56); got != 58 {
		t.Fatalf("EstimateListSize(56): got %d, want 58", got)
	}
}

func TestEstimateStringSize(t *testing.T) {
	if got := EstimateStringSize(1); got != 1 {
		t.Fatalf("EstimateStringSize(1): got %d, want 1", got)
	}
	if got := EstimateStringSize(55); got != 56 {
		t.Fatalf("EstimateStringSize(55): got %d, want 56", got)
	}
	// 56 bytes: 1 + 1 + 56 = 58
	if got := EstimateStringSize(56); got != 58 {
		t.Fatalf("EstimateStringSize(56): got %d, want 58", got)
	}
}

func TestAppendUint64(t *testing.T) {
	tests := []struct {
		val  uint64
		want []byte
	}{
		{0, []byte{0x80}},
		{1, []byte{0x01}},
		{127, []byte{0x7f}},
		{128, []byte{0x81, 0x80}},
		{256, []byte{0x82, 0x01, 0x00}},
	}
	for _, tt := range tests {
		got := AppendUint64(nil, tt.val)
		if !bytes.Equal(got, tt.want) {
			t.Fatalf("AppendUint64(%d): got %x, want %x", tt.val, got, tt.want)
		}
	}
	// Test appending to existing slice.
	base := []byte{0xaa}
	got := AppendUint64(base, 1)
	if !bytes.Equal(got, []byte{0xaa, 0x01}) {
		t.Fatalf("AppendUint64 with base: got %x", got)
	}
}

func TestAppendBytes(t *testing.T) {
	// single byte <= 0x7f: raw byte
	got := AppendBytes(nil, []byte{0x42})
	if !bytes.Equal(got, []byte{0x42}) {
		t.Fatalf("AppendBytes single 0x42: got %x", got)
	}
	// single byte > 0x7f
	got = AppendBytes(nil, []byte{0x80})
	if !bytes.Equal(got, []byte{0x81, 0x80}) {
		t.Fatalf("AppendBytes single 0x80: got %x", got)
	}
	// short string
	got = AppendBytes(nil, []byte{0x01, 0x02, 0x03})
	if !bytes.Equal(got, []byte{0x83, 0x01, 0x02, 0x03}) {
		t.Fatalf("AppendBytes 3 bytes: got %x", got)
	}
	// long string (>55 bytes)
	long := make([]byte, 56)
	for i := range long {
		long[i] = byte(i)
	}
	got = AppendBytes(nil, long)
	if got[0] != 0xb8 {
		t.Fatalf("AppendBytes long prefix: got %x, want 0xb8", got[0])
	}
}

func TestAppendListHeader(t *testing.T) {
	got := AppendListHeader(nil, 0)
	if !bytes.Equal(got, []byte{0xc0}) {
		t.Fatalf("AppendListHeader(0): got %x, want 0xc0", got)
	}
	got = AppendListHeader(nil, 55)
	if !bytes.Equal(got, []byte{0xf7}) {
		t.Fatalf("AppendListHeader(55): got %x, want 0xf7", got)
	}
	got = AppendListHeader(nil, 56)
	if got[0] != 0xf8 {
		t.Fatalf("AppendListHeader(56): got prefix %x, want 0xf8", got[0])
	}
}
