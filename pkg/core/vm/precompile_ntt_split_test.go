package vm

import (
	"math/big"
	"math/bits"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// encode32 encodes a value as a 32-byte big-endian slice.
func encode32(v int64) []byte {
	b := make([]byte, 32)
	val := new(big.Int).SetInt64(v).Bytes()
	copy(b[32-len(val):], val)
	return b
}

// concat32 concatenates multiple 32-byte encoded values.
func concat32(vals ...int64) []byte {
	out := make([]byte, 0, len(vals)*32)
	for _, v := range vals {
		out = append(out, encode32(v)...)
	}
	return out
}

// decode32 reads a big.Int from a 32-byte big-endian slice at index i.
func decode32(data []byte, i int) *big.Int {
	return new(big.Int).SetBytes(data[i*32 : (i+1)*32])
}

func TestNTTPrecompileAddresses(t *testing.T) {
	// 0x0f-0x14 must be present in IPlus.
	addrs := []byte{0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	for _, a := range addrs {
		addr := types.BytesToAddress([]byte{a})
		if _, ok := PrecompiledContractsIPlus[addr]; !ok {
			t.Errorf("address 0x%02x not found in PrecompiledContractsIPlus", a)
		}
	}

	// 0x15 must NOT be present.
	old := types.BytesToAddress([]byte{0x15})
	if _, ok := PrecompiledContractsIPlus[old]; ok {
		t.Error("legacy address 0x15 should not be in PrecompiledContractsIPlus")
	}
}

func TestNTTFWINVRoundtrip(t *testing.T) {
	fw := &nttFWPrecompile{}
	inv := &nttINVPrecompile{}

	input := concat32(1, 2, 3, 4)

	fwdOut, err := fw.Run(input)
	if err != nil {
		t.Fatalf("forward NTT: %v", err)
	}
	if len(fwdOut) != 4*32 {
		t.Fatalf("expected %d bytes, got %d", 4*32, len(fwdOut))
	}

	invOut, err := inv.Run(fwdOut)
	if err != nil {
		t.Fatalf("inverse NTT: %v", err)
	}

	expected := []int64{1, 2, 3, 4}
	for i, want := range expected {
		got := decode32(invOut, i)
		if got.Cmp(big.NewInt(want)) != 0 {
			t.Errorf("roundtrip[%d]: got %v, want %d", i, got, want)
		}
	}
}

func TestNTTVecMulMod(t *testing.T) {
	p := &nttVecMulModPrecompile{}

	// A=[2,3], B=[4,5] -> [8,15]
	input := concat32(2, 3, 4, 5)
	out, err := p.Run(input)
	if err != nil {
		t.Fatalf("vecmul: %v", err)
	}

	expected := []int64{8, 15}
	for i, want := range expected {
		got := decode32(out, i)
		if got.Cmp(big.NewInt(want)) != 0 {
			t.Errorf("vecmul[%d]: got %v, want %d", i, got, want)
		}
	}
}

func TestNTTVecAddMod(t *testing.T) {
	p := &nttVecAddModPrecompile{}

	// A=[2,3], B=[4,5] -> [6,8]
	input := concat32(2, 3, 4, 5)
	out, err := p.Run(input)
	if err != nil {
		t.Fatalf("vecadd: %v", err)
	}

	expected := []int64{6, 8}
	for i, want := range expected {
		got := decode32(out, i)
		if got.Cmp(big.NewInt(want)) != 0 {
			t.Errorf("vecadd[%d]: got %v, want %d", i, got, want)
		}
	}
}

func TestNTTDotProduct(t *testing.T) {
	p := &nttDotProductPrecompile{}

	// [1,2,3,4] . [4,5,6,7] = 1*4+2*5+3*6+4*7 = 4+10+18+28 = 60
	input := concat32(1, 2, 3, 4, 4, 5, 6, 7)
	out, err := p.Run(input)
	if err != nil {
		t.Fatalf("dotproduct: %v", err)
	}
	if len(out) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(out))
	}

	got := new(big.Int).SetBytes(out)
	if got.Cmp(big.NewInt(60)) != 0 {
		t.Errorf("dotproduct: got %v, want 60", got)
	}
}

func TestNTTButterfly(t *testing.T) {
	p := &nttButterflyPrecompile{}

	// [0,1,2,3,4,5,6,7] -> bit-reversal -> [0,4,2,6,1,5,3,7]
	input := concat32(0, 1, 2, 3, 4, 5, 6, 7)
	out, err := p.Run(input)
	if err != nil {
		t.Fatalf("butterfly: %v", err)
	}

	expected := []int64{0, 4, 2, 6, 1, 5, 3, 7}
	for i, want := range expected {
		got := decode32(out, i)
		if got.Cmp(big.NewInt(want)) != 0 {
			t.Errorf("butterfly[%d]: got %v, want %d", i, got, want)
		}
	}
}

func TestNTTGasFormulas(t *testing.T) {
	// n=4 elements for FW/INV/VecMul
	n4input := make([]byte, 4*32)

	tests := []struct {
		name string
		pc   PrecompiledContract
		in   []byte
		want uint64
	}{
		{
			"FW n=4",
			&nttFWPrecompile{},
			n4input,
			// max(600, 4*log2(4)/8) = max(600, 4*2/8) = max(600, 1) = 600
			600,
		},
		{
			"INV n=4",
			&nttINVPrecompile{},
			n4input,
			600,
		},
		{
			"VecMul n=4 (2 elements per side)",
			&nttVecMulModPrecompile{},
			n4input,
			// input has 4 elements total, n = len/64 = 2, max(600, 2*log2(2)/8) = max(600, 0) = 600
			600,
		},
		{
			"VecAdd n=4 (2 elements per side)",
			&nttVecAddModPrecompile{},
			n4input,
			// max(100, 2*log2(2)/32) = max(100, 0) = 100
			100,
		},
		{
			"DotProduct n=4 (2 elements per side)",
			&nttDotProductPrecompile{},
			n4input,
			// max(600, n) where n = len/64 = 2 -> max(600, 2) = 600
			600,
		},
		{
			"Butterfly n=4",
			&nttButterflyPrecompile{},
			n4input,
			// max(300, n) where n = len/32 = 4 -> max(300, 4) = 300
			300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pc.RequiredGas(tt.in)
			if got != tt.want {
				t.Errorf("gas: got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNTTForkActivation(t *testing.T) {
	// 0x15 not present in IPlus.
	old := types.BytesToAddress([]byte{0x15})
	if _, ok := PrecompiledContractsIPlus[old]; ok {
		t.Error("0x15 should not be in IPlus")
	}

	// 0x0f present in IPlus.
	addr := types.BytesToAddress([]byte{0x0f})
	if _, ok := PrecompiledContractsIPlus[addr]; !ok {
		t.Error("0x0f should be in IPlus")
	}
}

func TestNTTVecMulModOddInput(t *testing.T) {
	p := &nttVecMulModPrecompile{}

	// Odd number of elements (3) should fail.
	input := concat32(1, 2, 3)
	_, err := p.Run(input)
	if err == nil {
		t.Error("expected error for odd element count")
	}
}

func TestNTTButterflyNotPowerOfTwo(t *testing.T) {
	p := &nttButterflyPrecompile{}

	// 3 elements = not power of two.
	input := concat32(1, 2, 3)
	_, err := p.Run(input)
	if err == nil {
		t.Error("expected error for non-power-of-two input")
	}
}

func TestNTTFWLargerInput(t *testing.T) {
	fw := &nttFWPrecompile{}
	inv := &nttINVPrecompile{}

	// 8 elements roundtrip.
	input := concat32(1, 2, 3, 4, 5, 6, 7, 8)
	fwdOut, err := fw.Run(input)
	if err != nil {
		t.Fatalf("fw 8 elements: %v", err)
	}

	invOut, err := inv.Run(fwdOut)
	if err != nil {
		t.Fatalf("inv 8 elements: %v", err)
	}

	for i := int64(1); i <= 8; i++ {
		got := decode32(invOut, int(i-1))
		if got.Cmp(big.NewInt(i)) != 0 {
			t.Errorf("roundtrip[%d]: got %v, want %d", i-1, got, i)
		}
	}
}

func TestNTTGasScaling(t *testing.T) {
	fw := &nttFWPrecompile{}

	// n=256 elements -> gas = max(600, 256*8/8) = max(600, 256) = 600
	input := make([]byte, 256*32)
	got := fw.RequiredGas(input)
	n := uint64(256)
	log2n := uint64(bits.Len(uint(n)) - 1) // 8
	computed := n * log2n / 8
	want := computed
	if want < 600 {
		want = 600
	}
	if got != want {
		t.Errorf("gas for 256 elements: got %d, want %d", got, want)
	}
}
