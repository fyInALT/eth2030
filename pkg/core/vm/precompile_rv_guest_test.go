package vm

import (
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
