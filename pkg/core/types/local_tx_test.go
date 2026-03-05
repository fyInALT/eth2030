package types

import (
	"math/big"
	"testing"
)

func TestLocalTx_TxType(t *testing.T) {
	ltx := &LocalTx{}
	if ltx.txType() != LocalTxType {
		t.Fatalf("expected type 0x%02x, got 0x%02x", LocalTxType, ltx.txType())
	}
	if LocalTxType != 0x08 {
		t.Fatalf("expected LocalTxType=0x08, got 0x%02x", LocalTxType)
	}
}

func TestLocalTx_Fields(t *testing.T) {
	to := HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	ltx := &LocalTx{
		ChainID_:   big.NewInt(1),
		Nonce_:     42,
		GasTipCap_: big.NewInt(100),
		GasFeeCap_: big.NewInt(200),
		Gas_:       21000,
		To_:        &to,
		Value_:     big.NewInt(1000),
		Data_:      []byte{0xca, 0xfe},
		ScopeHint:  []byte{0x0a, 0x0b},
	}

	if ltx.chainID().Cmp(big.NewInt(1)) != 0 {
		t.Fatal("chainID mismatch")
	}
	if ltx.nonce() != 42 {
		t.Fatal("nonce mismatch")
	}
	if ltx.gasTipCap().Cmp(big.NewInt(100)) != 0 {
		t.Fatal("gasTipCap mismatch")
	}
	if ltx.gasFeeCap().Cmp(big.NewInt(200)) != 0 {
		t.Fatal("gasFeeCap mismatch")
	}
	if ltx.gasPrice().Cmp(big.NewInt(200)) != 0 {
		t.Fatal("gasPrice should equal gasFeeCap")
	}
	if ltx.gas() != 21000 {
		t.Fatal("gas mismatch")
	}
	if *ltx.to() != to {
		t.Fatal("to mismatch")
	}
	if ltx.value().Cmp(big.NewInt(1000)) != 0 {
		t.Fatal("value mismatch")
	}
	if len(ltx.data()) != 2 || ltx.data()[0] != 0xca {
		t.Fatal("data mismatch")
	}
	if ltx.accessList() != nil {
		t.Fatal("accessList should be nil")
	}
}

func TestLocalTx_Copy(t *testing.T) {
	to := HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	orig := &LocalTx{
		ChainID_:   big.NewInt(1),
		Nonce_:     10,
		GasTipCap_: big.NewInt(50),
		GasFeeCap_: big.NewInt(100),
		Gas_:       21000,
		To_:        &to,
		Value_:     big.NewInt(500),
		Data_:      []byte{0x01, 0x02},
		ScopeHint:  []byte{0xaa, 0xbb},
	}

	cpy := orig.copy().(*LocalTx)

	// Verify values match.
	if cpy.ChainID_.Cmp(orig.ChainID_) != 0 {
		t.Fatal("copy chainID mismatch")
	}
	if cpy.Nonce_ != orig.Nonce_ {
		t.Fatal("copy nonce mismatch")
	}
	if cpy.Gas_ != orig.Gas_ {
		t.Fatal("copy gas mismatch")
	}
	if *cpy.To_ != *orig.To_ {
		t.Fatal("copy to mismatch")
	}
	if cpy.Value_.Cmp(orig.Value_) != 0 {
		t.Fatal("copy value mismatch")
	}
	if len(cpy.ScopeHint) != 2 || cpy.ScopeHint[0] != 0xaa {
		t.Fatal("copy scopeHint mismatch")
	}

	// Verify deep copy (mutation isolation).
	cpy.ChainID_.SetInt64(999)
	if orig.ChainID_.Int64() == 999 {
		t.Fatal("copy should be independent: chainID mutated")
	}
	cpy.ScopeHint[0] = 0xFF
	if orig.ScopeHint[0] == 0xFF {
		t.Fatal("copy should be independent: scopeHint mutated")
	}
	newTo := HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	cpy.To_ = &newTo
	if *orig.To_ == newTo {
		t.Fatal("copy should be independent: to mutated")
	}
}

func TestNewLocalTx(t *testing.T) {
	to := HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	tx := NewLocalTx(
		big.NewInt(1),
		5,
		&to,
		big.NewInt(1000),
		21000,
		big.NewInt(10),
		big.NewInt(20),
		[]byte{0x01},
		[]byte{0x0a, 0x0b},
	)

	if tx.Type() != LocalTxType {
		t.Fatalf("expected type 0x%02x, got 0x%02x", LocalTxType, tx.Type())
	}
	if tx.Nonce() != 5 {
		t.Fatalf("expected nonce 5, got %d", tx.Nonce())
	}
	if tx.Gas() != 21000 {
		t.Fatalf("expected gas 21000, got %d", tx.Gas())
	}
	if *tx.To() != to {
		t.Fatal("to address mismatch")
	}
	if tx.Value().Cmp(big.NewInt(1000)) != 0 {
		t.Fatal("value mismatch")
	}
	if tx.GasTipCap().Cmp(big.NewInt(10)) != 0 {
		t.Fatal("gasTipCap mismatch")
	}
	if tx.GasFeeCap().Cmp(big.NewInt(20)) != 0 {
		t.Fatal("gasFeeCap mismatch")
	}
	hint := GetScopeHint(tx)
	if len(hint) != 2 || hint[0] != 0x0a || hint[1] != 0x0b {
		t.Fatal("scope hint mismatch")
	}
}

func TestScopesOverlap_Overlapping(t *testing.T) {
	a := &LocalTx{ScopeHint: []byte{0x0a, 0x0b}}
	b := &LocalTx{ScopeHint: []byte{0x0b, 0x0c}}
	if !ScopesOverlap(a, b) {
		t.Fatal("expected overlap on prefix 0x0b")
	}
}

func TestScopesOverlap_NonOverlapping(t *testing.T) {
	a := &LocalTx{ScopeHint: []byte{0x0a, 0x0b}}
	b := &LocalTx{ScopeHint: []byte{0x0c, 0x0d}}
	if ScopesOverlap(a, b) {
		t.Fatal("expected no overlap")
	}
}

func TestScopesOverlap_EmptyScope(t *testing.T) {
	a := &LocalTx{ScopeHint: []byte{0x0a}}
	b := &LocalTx{ScopeHint: []byte{}}
	if !ScopesOverlap(a, b) {
		t.Fatal("empty scope should overlap with everything")
	}
	if !ScopesOverlap(b, a) {
		t.Fatal("empty scope should overlap with everything (reverse)")
	}
}

func TestScopesOverlap_NilTx(t *testing.T) {
	a := &LocalTx{ScopeHint: []byte{0x0a}}
	if !ScopesOverlap(nil, a) {
		t.Fatal("nil tx should overlap with everything")
	}
	if !ScopesOverlap(a, nil) {
		t.Fatal("nil tx should overlap with everything (reverse)")
	}
	if !ScopesOverlap(nil, nil) {
		t.Fatal("nil/nil should overlap")
	}
}

func TestIsLocalTx(t *testing.T) {
	to := HexToAddress("0x1111111111111111111111111111111111111111")
	localTx := NewLocalTx(big.NewInt(1), 0, &to, big.NewInt(0), 21000,
		big.NewInt(1), big.NewInt(1), nil, []byte{0x0a})
	if !IsLocalTx(localTx) {
		t.Fatal("expected IsLocalTx=true for LocalTx")
	}

	legacyTx := NewTransaction(&LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1),
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(0),
	})
	if IsLocalTx(legacyTx) {
		t.Fatal("expected IsLocalTx=false for LegacyTx")
	}

	if IsLocalTx(nil) {
		t.Fatal("expected IsLocalTx=false for nil")
	}
}

func TestGetScopeHint(t *testing.T) {
	to := HexToAddress("0x2222222222222222222222222222222222222222")
	hint := []byte{0xaa, 0xbb, 0xcc}
	localTx := NewLocalTx(big.NewInt(1), 0, &to, big.NewInt(0), 21000,
		big.NewInt(1), big.NewInt(1), nil, hint)
	got := GetScopeHint(localTx)
	if len(got) != 3 || got[0] != 0xaa || got[1] != 0xbb || got[2] != 0xcc {
		t.Fatalf("expected scope hint %x, got %x", hint, got)
	}

	legacyTx := NewTransaction(&LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1),
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(0),
	})
	if GetScopeHint(legacyTx) != nil {
		t.Fatal("expected nil scope hint for LegacyTx")
	}

	if GetScopeHint(nil) != nil {
		t.Fatal("expected nil scope hint for nil tx")
	}
}

// --- BB-2.2: Gas discount ---

func TestLocalTx_GasWithDiscount_50Percent(t *testing.T) {
	ltx := &LocalTx{Gas_: 21000}
	got := ltx.GasWithDiscount(LocalTxDiscountBPS)
	want := uint64(10500)
	if got != want {
		t.Fatalf("50%% discount: got %d, want %d", got, want)
	}
}

func TestLocalTx_GasWithDiscount_Zero(t *testing.T) {
	ltx := &LocalTx{Gas_: 21000}
	got := ltx.GasWithDiscount(0)
	if got != 21000 {
		t.Fatalf("zero discount: got %d, want 21000", got)
	}
}

func TestLocalTx_GasWithDiscount_Full(t *testing.T) {
	ltx := &LocalTx{Gas_: 21000}
	// 10000 bps = 100%, discount >= 10000 returns 0.
	got := ltx.GasWithDiscount(10000)
	if got != 0 {
		t.Fatalf("100%% discount: got %d, want 0", got)
	}
}

func TestLocalTx_GasWithDiscount_25Percent(t *testing.T) {
	ltx := &LocalTx{Gas_: 20000}
	got := ltx.GasWithDiscount(2500)
	want := uint64(15000)
	if got != want {
		t.Fatalf("25%% discount: got %d, want %d", got, want)
	}
}

func TestApplyLocalTxDiscount_LocalTx(t *testing.T) {
	to := HexToAddress("0x1111111111111111111111111111111111111111")
	tx := NewLocalTx(big.NewInt(1), 0, &to, big.NewInt(0), 21000,
		big.NewInt(1), big.NewInt(1), nil, []byte{0x0a})
	got := ApplyLocalTxDiscount(tx)
	want := uint64(21000) * (10000 - LocalTxDiscountBPS) / 10000
	if got != want {
		t.Fatalf("ApplyLocalTxDiscount: got %d, want %d", got, want)
	}
}

func TestApplyLocalTxDiscount_NonLocalTx(t *testing.T) {
	to := HexToAddress("0x1111111111111111111111111111111111111111")
	tx := NewTransaction(&LegacyTx{Gas: 21000, To: &to, GasPrice: big.NewInt(1)})
	got := ApplyLocalTxDiscount(tx)
	if got != 21000 {
		t.Fatalf("non-LocalTx should return original gas 21000, got %d", got)
	}
}

// --- BB-2.2: BAL access validation ---

func TestLocalTx_ValidateScopeAccess_Pass(t *testing.T) {
	ltx := &LocalTx{ScopeHint: []byte{0x0a, 0x0b}}
	// Addresses whose first byte is 0x0a or 0x0b.
	addrs := []Address{
		{0x0a, 0x01},
		{0x0b, 0xff},
	}
	if err := ltx.ValidateScopeAccess(addrs); err != nil {
		t.Fatalf("expected no violation, got: %v", err)
	}
}

func TestLocalTx_ValidateScopeAccess_Fail(t *testing.T) {
	ltx := &LocalTx{ScopeHint: []byte{0x0a, 0x0b}}
	// Address with prefix 0x0c is not declared.
	addrs := []Address{
		{0x0a, 0x01},
		{0x0c, 0x02},
	}
	err := ltx.ValidateScopeAccess(addrs)
	if err == nil {
		t.Fatal("expected ErrBALViolation for undeclared prefix 0x0c")
	}
}

func TestLocalTx_ValidateScopeAccess_EmptyScope(t *testing.T) {
	ltx := &LocalTx{ScopeHint: []byte{}}
	// Empty scope hint means global access — no violation.
	addrs := []Address{{0xff, 0xee}, {0x01, 0x02}}
	if err := ltx.ValidateScopeAccess(addrs); err != nil {
		t.Fatalf("empty scope should allow all access, got: %v", err)
	}
}

func TestLocalTx_ValidateScopeAccess_NilScope(t *testing.T) {
	ltx := &LocalTx{ScopeHint: nil}
	addrs := []Address{{0xab, 0xcd}}
	if err := ltx.ValidateScopeAccess(addrs); err != nil {
		t.Fatalf("nil scope should allow all access, got: %v", err)
	}
}

func TestLocalTx_ValidateScopeAccess_EmptyAddrs(t *testing.T) {
	ltx := &LocalTx{ScopeHint: []byte{0x0a}}
	if err := ltx.ValidateScopeAccess(nil); err != nil {
		t.Fatalf("empty access list should never violate, got: %v", err)
	}
}

func TestValidateLocalTxScopeAccess_Wrapper(t *testing.T) {
	to := HexToAddress("0x0a00000000000000000000000000000000000000")
	tx := NewLocalTx(big.NewInt(1), 0, &to, big.NewInt(0), 21000,
		big.NewInt(1), big.NewInt(1), nil, []byte{0x0a})

	// Declared prefix 0x0a — access to 0x0a... should pass.
	addr := HexToAddress("0x0a11111111111111111111111111111111111111")
	if err := ValidateLocalTxScopeAccess(tx, []Address{addr}); err != nil {
		t.Fatalf("expected no violation for declared prefix, got: %v", err)
	}

	// Undeclared prefix 0x0b — should fail.
	bad := HexToAddress("0x0b11111111111111111111111111111111111111")
	if err := ValidateLocalTxScopeAccess(tx, []Address{bad}); err == nil {
		t.Fatal("expected ErrBALViolation for undeclared prefix 0x0b")
	}

	// Non-LocalTx — always passes.
	legacyTx := NewTransaction(&LegacyTx{Gas: 21000, To: &to, GasPrice: big.NewInt(1)})
	if err := ValidateLocalTxScopeAccess(legacyTx, []Address{bad}); err != nil {
		t.Fatalf("non-LocalTx should not be validated, got: %v", err)
	}
}
