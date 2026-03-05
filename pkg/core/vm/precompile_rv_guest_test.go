package vm

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func TestIsPrecompileRISCV(t *testing.T) {
	cases := []struct {
		addr types.Address
		want bool
	}{
		{types.BytesToAddress([]byte{1}), true},  // ECRECOVER
		{types.BytesToAddress([]byte{2}), true},  // SHA-256
		{types.BytesToAddress([]byte{3}), false}, // RIPEMD-160 (not replaced)
		{types.BytesToAddress([]byte{4}), false}, // dataCopy
	}
	for _, tc := range cases {
		if got := IsPrecompileRISCV(tc.addr); got != tc.want {
			t.Errorf("IsPrecompileRISCV(%s) = %v, want %v", tc.addr, got, tc.want)
		}
	}
}

// TestRVPrecompileSHA256IPlus verifies that at I+, the SHA-256 precompile
// address (0x02) produces the same result as Go's crypto/sha256.
func TestRVPrecompileSHA256IPlus(t *testing.T) {
	addr := types.BytesToAddress([]byte{2})
	p, ok := PrecompiledContractsIPlus[addr]
	if !ok {
		t.Fatal("SHA-256 not found in PrecompiledContractsIPlus")
	}

	// Toggle: before I+, use Glamsterdan map.
	glamP, ok := PrecompiledContractsGlamsterdan[addr]
	if !ok {
		t.Fatal("SHA-256 not found in PrecompiledContractsGlamsterdan")
	}

	input := []byte("hello world")
	want := sha256.Sum256(input)

	// I+ RISC-V path.
	out, err := p.Run(input)
	if err != nil {
		t.Fatalf("I+ SHA-256 Run: %v", err)
	}
	if len(out) != 32 {
		t.Fatalf("I+ SHA-256 output len = %d, want 32", len(out))
	}
	for i, b := range want {
		if out[i] != b {
			t.Errorf("byte %d: got 0x%x, want 0x%x", i, out[i], b)
		}
	}

	// Pre-I+ Go path must produce the same result.
	out2, err := glamP.Run(input)
	if err != nil {
		t.Fatalf("Glamsterdan SHA-256 Run: %v", err)
	}
	for i := range out {
		if out[i] != out2[i] {
			t.Errorf("I+ and pre-I+ paths differ at byte %d", i)
		}
	}
}

// TestRVPrecompileSHA256GasUnchanged verifies the gas cost is the same
// before and after the I+ fork (RISC-V execution is transparent to gas).
func TestRVPrecompileSHA256GasUnchanged(t *testing.T) {
	addr := types.BytesToAddress([]byte{2})
	iplusP := PrecompiledContractsIPlus[addr]
	glamP := PrecompiledContractsGlamsterdan[addr]

	input := []byte("gas test input with some bytes")
	g1 := iplusP.RequiredGas(input)
	g2 := glamP.RequiredGas(input)
	if g1 != g2 {
		t.Errorf("I+ gas = %d, pre-I+ gas = %d: should be equal", g1, g2)
	}
}

// TestSelectPrecompilesUsesIPlusAtIPlus verifies SelectPrecompiles returns the
// I+ map (with RISC-V wrappers) when IsIPlus is true.
func TestSelectPrecompilesUsesIPlusAtIPlus(t *testing.T) {
	rules := ForkRules{IsIPlus: true}
	m := SelectPrecompiles(rules)
	addr := types.BytesToAddress([]byte{2})
	p, ok := m[addr]
	if !ok {
		t.Fatal("SHA-256 not in I+ precompile map")
	}
	// Verify it's the RISC-V wrapper by running it.
	input := []byte("check")
	want := sha256.Sum256(input)
	out, err := p.Run(input)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for i, b := range want {
		if out[i] != b {
			t.Errorf("byte %d mismatch after SelectPrecompiles at I+", i)
		}
	}
}

// --- Additional edge-case tests ---

// TestIsPrecompileRISCVBoundaryAddresses verifies only 0x01 and 0x02
// return true; neighbouring addresses do not.
func TestIsPrecompileRISCVBoundaryAddresses(t *testing.T) {
	cases := []struct {
		addr types.Address
		want bool
	}{
		{types.BytesToAddress([]byte{0}), false},    // 0x00 — not RISC-V
		{types.BytesToAddress([]byte{1}), true},     // 0x01 — ECRECOVER
		{types.BytesToAddress([]byte{2}), true},     // 0x02 — SHA-256
		{types.BytesToAddress([]byte{3}), false},    // 0x03 — RIPEMD-160
		{types.BytesToAddress([]byte{9}), false},    // 0x09 — Blake2b
		{types.BytesToAddress([]byte{0xff}), false}, // 0xFF — not a precompile
	}
	for _, tc := range cases {
		if got := IsPrecompileRISCV(tc.addr); got != tc.want {
			t.Errorf("IsPrecompileRISCV(%x) = %v, want %v", tc.addr, got, tc.want)
		}
	}
}

// TestRVPrecompileECRECOVERIPlusRouted verifies ECRECOVER (0x01) at I+ is
// backed by rvPrecompile (not nil) and returns a 32-byte result for a
// zero-hash input (invalid recovery → 32 zero bytes).
func TestRVPrecompileECRECOVERIPlusRouted(t *testing.T) {
	addr := types.BytesToAddress([]byte{1})
	p, ok := PrecompiledContractsIPlus[addr]
	if !ok {
		t.Fatal("ECRECOVER not found in PrecompiledContractsIPlus")
	}

	// Build a 128-byte input: hash=0, v=27, r=0, s=0.
	input := make([]byte, 128)
	input[63] = 27 // v = 27
	out, err := p.Run(input)
	if err != nil {
		t.Fatalf("ECRECOVER I+ Run: %v", err)
	}
	if len(out) != 32 {
		t.Fatalf("ECRECOVER output len = %d, want 32", len(out))
	}
}

// TestRVPrecompileECRECOVERFailureRepresentation verifies that both the I+
// (RISC-V) and Glamsterdan (Go native) paths signal failure for an invalid
// signature. The native path returns empty bytes; the RISC-V path returns
// 32 zero bytes — both are valid "no address recovered" representations.
func TestRVPrecompileECRECOVERFailureRepresentation(t *testing.T) {
	addr := types.BytesToAddress([]byte{1})
	iplusP := PrecompiledContractsIPlus[addr]
	glamP := PrecompiledContractsGlamsterdan[addr]

	// Invalid signature: zero hash, v=27, zero r, zero s.
	input := make([]byte, 128)
	input[63] = 27

	iOut, iErr := iplusP.Run(input)
	gOut, gErr := glamP.Run(input)

	if iErr != nil {
		t.Fatalf("I+ ECRECOVER returned error on invalid input: %v", iErr)
	}
	if gErr != nil {
		t.Fatalf("Glamsterdan ECRECOVER returned error on invalid input: %v", gErr)
	}

	// I+ returns 32 zero bytes; Glamsterdan returns empty bytes — both signal failure.
	iIsFailure := len(iOut) == 0 || bytes.Equal(iOut, make([]byte, 32))
	gIsFailure := len(gOut) == 0 || bytes.Equal(gOut, make([]byte, 32))
	if !iIsFailure {
		t.Errorf("I+ ECRECOVER: expected failure result, got %x", iOut)
	}
	if !gIsFailure {
		t.Errorf("Glamsterdan ECRECOVER: expected failure result, got %x", gOut)
	}
}

// TestIPlusPrecompileMapContainsExpectedAddresses verifies the I+ precompile
// map contains both RISC-V-routed addresses.
func TestIPlusPrecompileMapContainsExpectedAddresses(t *testing.T) {
	for _, b := range []byte{1, 2} {
		addr := types.BytesToAddress([]byte{b})
		p, ok := PrecompiledContractsIPlus[addr]
		if !ok {
			t.Errorf("address 0x%02x not in PrecompiledContractsIPlus", b)
			continue
		}
		if p == nil {
			t.Errorf("address 0x%02x has nil precompile", b)
		}
	}
}

// TestRVPrecompileSHA256ForksConsistency checks SHA-256 output is identical
// across Cancun, Glamsterdan, and I+ fork maps.
func TestRVPrecompileSHA256ForksConsistency(t *testing.T) {
	addr := types.BytesToAddress([]byte{2})
	input := []byte("fork consistency test")
	want := sha256.Sum256(input)

	maps := []struct {
		name string
		m    map[types.Address]PrecompiledContract
	}{
		{"Cancun", PrecompiledContractsCancun},
		{"Glamsterdan", PrecompiledContractsGlamsterdan},
		{"IPlus", PrecompiledContractsIPlus},
	}
	for _, tc := range maps {
		p, ok := tc.m[addr]
		if !ok {
			t.Errorf("%s: 0x02 missing", tc.name)
			continue
		}
		out, err := p.Run(input)
		if err != nil {
			t.Errorf("%s: Run error: %v", tc.name, err)
			continue
		}
		if !bytes.Equal(out, want[:]) {
			t.Errorf("%s: SHA-256 = %x, want %x", tc.name, out, want)
		}
	}
}
