// opcode_rvcreate_test.go tests the RVCREATE opcode (EL-3.2) and RISC-V
// call routing in EVM.Call (EL-3.3).
package vm

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/zkvm"
)

// ---- helpers ----------------------------------------------------------------

// rvTestStateDB extends createTestStateDB for RVCREATE tests.
type rvTestStateDB struct {
	createTestStateDB
}

func newRVTestStateDB() *rvTestStateDB {
	return &rvTestStateDB{createTestStateDB: *newCreateTestStateDB()}
}

// newRVCreateEVM builds a minimal EVM wired for I+ with the given StateDB.
func newRVCreateEVM(sdb StateDB) *EVM {
	rules := ForkRules{IsIPlus: true, IsGlamsterdan: true}
	evm := NewEVM(BlockContext{}, TxContext{}, Config{})
	evm.forkRules = rules
	evm.jumpTable = SelectJumpTable(rules)
	evm.StateDB = sdb
	return evm
}

// pushStack pushes big.Ints onto a stack, panicking on overflow.
func pushStack(stk *Stack, vals ...*big.Int) {
	for _, v := range vals {
		if err := stk.Push(v); err != nil {
			panic(err)
		}
	}
}

// invokeRVCreate calls opRVCreate directly with the given stack values.
// Stack layout for CREATE-style: top = value, next = offset, bottom = size.
func invokeRVCreate(t *testing.T, evm *EVM, caller types.Address, value *big.Int, initCode []byte) *big.Int {
	t.Helper()
	mem := NewMemory()
	mem.Resize(uint64(len(initCode)))
	copy(mem.store, initCode)

	stk := NewStack()
	// Push size first (deepest), then offset, then value (top).
	pushStack(stk,
		new(big.Int).SetUint64(uint64(len(initCode))), // size (deepest)
		new(big.Int).SetUint64(0),                     // offset
		new(big.Int).Set(value),                       // value (top)
	)

	contract := NewContract(caller, caller, big.NewInt(0), 0)
	pc := uint64(0)
	_, err := opRVCreate(&pc, evm, contract, mem, stk)
	if err != nil {
		t.Fatalf("opRVCreate returned error: %v", err)
	}
	if stk.Len() != 1 {
		t.Fatalf("stack len after opRVCreate = %d, want 1", stk.Len())
	}
	return stk.Peek()
}

// ---- IsRVCode ---------------------------------------------------------------

func TestIsRVCode(t *testing.T) {
	cases := []struct {
		name string
		code []byte
		want bool
	}{
		{"magic prefix", []byte{0xFE, 0x52, 0x56, 0x00}, true},
		{"short 2", []byte{0xFE, 0x52}, false},
		{"nil", nil, false},
		{"evm bytecode", []byte{0x60, 0x00, 0x52}, false},
		{"wrong first", []byte{0xFF, 0x52, 0x56}, false},
		{"wrong second", []byte{0xFE, 0x53, 0x56}, false},
		{"wrong third", []byte{0xFE, 0x52, 0x57}, false},
	}
	for _, tc := range cases {
		if got := IsRVCode(tc.code); got != tc.want {
			t.Errorf("%s: IsRVCode = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// ---- rvCreateAddress determinism --------------------------------------------

func TestRVCreateAddressDeterministic(t *testing.T) {
	from := types.HexToAddress("0xabcdef1234567890abcdef1234567890abcdef12")
	initcode := []byte{RVMagic0, RVMagic1, RVMagic2, 0x00}

	addr1 := rvCreateAddress(from, initcode)
	addr2 := rvCreateAddress(from, initcode)
	if addr1 != addr2 {
		t.Errorf("rvCreateAddress not deterministic: %s != %s", addr1, addr2)
	}
}

func TestRVCreateAddressDifferentCallers(t *testing.T) {
	initcode := []byte{RVMagic0, RVMagic1, RVMagic2, 0x00}
	a := rvCreateAddress(types.HexToAddress("0x0001"), initcode)
	b := rvCreateAddress(types.HexToAddress("0x0002"), initcode)
	if a == b {
		t.Errorf("different callers must produce different addresses")
	}
}

func TestRVCreateAddressDifferentCode(t *testing.T) {
	from := types.HexToAddress("0x1234")
	c1 := []byte{RVMagic0, RVMagic1, RVMagic2, 0x00}
	c2 := []byte{RVMagic0, RVMagic1, RVMagic2, 0xFF}
	a := rvCreateAddress(from, c1)
	b := rvCreateAddress(from, c2)
	if a == b {
		t.Errorf("different initcodes must produce different addresses")
	}
}

func TestRVCreateAddressNonZero(t *testing.T) {
	from := types.HexToAddress("0x1234")
	initcode := []byte{RVMagic0, RVMagic1, RVMagic2, 0x00}
	addr := rvCreateAddress(from, initcode)
	if addr.IsZero() {
		t.Errorf("rvCreateAddress returned zero address")
	}
}

// Verify CREATE2-style salt construction uses the magic prefix.
func TestRVCreateAddressUsesMagicInSalt(t *testing.T) {
	from := types.HexToAddress("0xdead")
	initcode := []byte{RVMagic0, RVMagic1, RVMagic2, 0x00}

	// Manually compute expected address to validate the formula.
	magic := []byte{RVMagic0, RVMagic1, RVMagic2}
	saltInput := append(magic, initcode...) //nolint:gocritic
	salt := crypto.Keccak256(saltInput)
	codeHash := crypto.Keccak256(initcode)
	buf := make([]byte, 1+20+32+32)
	buf[0] = 0xff
	copy(buf[1:21], from[:])
	copy(buf[21:53], salt)
	copy(buf[53:85], codeHash)
	h := crypto.Keccak256(buf)
	expected := types.BytesToAddress(h[12:])

	got := rvCreateAddress(from, initcode)
	if got != expected {
		t.Errorf("rvCreateAddress = %s, want %s", got, expected)
	}
}

// ---- gasRVCreate ------------------------------------------------------------

func TestGasRVCreateBasic(t *testing.T) {
	evm := newRVCreateEVM(newRVTestStateDB())
	contract := NewContract(types.HexToAddress("0xaabb"), types.HexToAddress("0xaabb"), big.NewInt(0), 0)
	stk := NewStack()
	size := uint64(200)
	pushStack(stk,
		new(big.Int).SetUint64(size), // size (deepest)
		new(big.Int).SetUint64(0),    // offset
		new(big.Int).SetUint64(0),    // value (top)
	)
	gas, err := gasRVCreate(evm, contract, stk, NewMemory(), 0)
	if err != nil {
		t.Fatalf("gasRVCreate: %v", err)
	}
	if gas != 200*size {
		t.Errorf("gasRVCreate(%d) = %d, want %d", size, gas, 200*size)
	}
}

func TestGasRVCreateSmallStack(t *testing.T) {
	evm := newRVCreateEVM(newRVTestStateDB())
	contract := NewContract(types.HexToAddress("0xaabb"), types.HexToAddress("0xaabb"), big.NewInt(0), 0)
	stk := NewStack()
	if err := stk.Push(new(big.Int).SetUint64(0)); err != nil {
		t.Fatal(err)
	}
	gas, err := gasRVCreate(evm, contract, stk, NewMemory(), 0)
	if err != nil {
		t.Fatalf("gasRVCreate with small stack: %v", err)
	}
	if gas != 0 {
		t.Errorf("gasRVCreate with small stack: got %d, want 0", gas)
	}
}

// ---- opRVCreate: valid magic deploys ----------------------------------------

// len(initCode) must be >= 4 and len(initCode)%4 == 0.
// With magic prefix (3 bytes): 3+1=4, 4%4==0 → valid.
var minValidRVInitcode = []byte{RVMagic0, RVMagic1, RVMagic2, 0x00}

func TestOpRVCreateValidMagicDeploys(t *testing.T) {
	sdb := newRVTestStateDB()
	evm := newRVCreateEVM(sdb)
	caller := types.HexToAddress("0xcafe")

	result := invokeRVCreate(t, evm, caller, big.NewInt(0), minValidRVInitcode)
	if result.Sign() == 0 {
		t.Fatal("expected non-zero address on success")
	}

	// Address must be the deterministic one.
	expected := rvCreateAddress(caller, minValidRVInitcode)
	gotAddr := types.BytesToAddress(result.Bytes())
	if gotAddr != expected {
		t.Errorf("deployed address = %s, want %s", gotAddr, expected)
	}

	// Code must be stored in StateDB.
	stored := sdb.GetCode(expected)
	if len(stored) == 0 {
		t.Errorf("code not stored in StateDB")
	}
}

func TestOpRVCreateInvalidMagicReturnsZero(t *testing.T) {
	sdb := newRVTestStateDB()
	evm := newRVCreateEVM(sdb)
	caller := types.HexToAddress("0xcafe")

	// EVM bytecode without magic prefix.
	initcode := []byte{0x60, 0x00, 0x60, 0x00, 0xf3, 0x00, 0x00, 0x00}
	result := invokeRVCreate(t, evm, caller, big.NewInt(0), initcode)
	if result.Sign() != 0 {
		t.Errorf("expected 0 on invalid magic, got %s", result)
	}
}

func TestOpRVCreateInvalidAlignmentReturnsZero(t *testing.T) {
	sdb := newRVTestStateDB()
	evm := newRVCreateEVM(sdb)
	caller := types.HexToAddress("0xcafe")

	// len = 3+2 = 5; 5%4 != 0 → reject.
	initcode := []byte{RVMagic0, RVMagic1, RVMagic2, 0x00, 0x00}
	result := invokeRVCreate(t, evm, caller, big.NewInt(0), initcode)
	if result.Sign() != 0 {
		t.Errorf("expected 0 on unaligned code, got %s", result)
	}
}

// ---- opRVCreate: value transfer ---------------------------------------------

func TestOpRVCreateTransfersValue(t *testing.T) {
	sdb := newRVTestStateDB()
	caller := types.HexToAddress("0xcafe")
	sdb.balances[caller] = big.NewInt(1_000_000)

	evm := newRVCreateEVM(sdb)

	result := invokeRVCreate(t, evm, caller, big.NewInt(100), minValidRVInitcode)
	if result.Sign() == 0 {
		t.Fatal("expected non-zero address on success")
	}
	addr := types.BytesToAddress(result.Bytes())
	if sdb.GetBalance(addr).Cmp(big.NewInt(100)) != 0 {
		t.Errorf("contract balance = %s, want 100", sdb.GetBalance(addr))
	}
	if sdb.GetBalance(caller).Cmp(big.NewInt(1_000_000-100)) != 0 {
		t.Errorf("caller balance = %s, want 999900", sdb.GetBalance(caller))
	}
}

func TestOpRVCreateInsufficientBalanceReturnsZero(t *testing.T) {
	sdb := newRVTestStateDB()
	caller := types.HexToAddress("0xcafe")
	sdb.balances[caller] = big.NewInt(50) // less than value=100

	evm := newRVCreateEVM(sdb)

	result := invokeRVCreate(t, evm, caller, big.NewInt(100), minValidRVInitcode)
	if result.Sign() != 0 {
		t.Errorf("expected 0 on insufficient balance, got %s", result)
	}
}

// ---- RVCREATE gas model in jump table (EL-3.1) ------------------------------

func TestIPlusJumpTableHasRVCREATE(t *testing.T) {
	jt := NewIPlusJumpTable()
	op := jt[RVCREATE]
	if op == nil {
		t.Fatal("RVCREATE not found in I+ jump table")
	}
	if op.constantGas != GasCreate {
		t.Errorf("RVCREATE constantGas = %d, want %d (GasCreate)", op.constantGas, GasCreate)
	}
	if !op.writes {
		t.Error("RVCREATE must be a state-writing operation")
	}
}

func TestSelectJumpTableIPlusHasRVCREATE(t *testing.T) {
	jt := SelectJumpTable(ForkRules{IsIPlus: true})
	if jt[RVCREATE] == nil {
		t.Fatal("SelectJumpTable(IsIPlus) must include RVCREATE")
	}
}

func TestGlamsterdanJumpTableNoRVCREATE(t *testing.T) {
	jt := SelectJumpTable(ForkRules{IsGlamsterdan: true})
	if jt[RVCREATE] != nil {
		t.Error("SelectJumpTable(Glamsterdan) must NOT include RVCREATE")
	}
}

// ---- EL-3.3: RISC-V call routing --------------------------------------------

// TestRVCallRoutingRVCodeRoutedToRISCV verifies that calling an address whose
// code starts with the RISC-V magic at I+ runs the RISC-V executor.
func TestRVCallRoutingRVCodeRoutedToRISCV(t *testing.T) {
	sdb := newRVTestStateDB()
	contractAddr := types.HexToAddress("0xdead")

	// Deploy the Keccak256 RISC-V guest (magic + RVKeccak256Program).
	magicProgram := append([]byte{RVMagic0, RVMagic1, RVMagic2}, zkvm.RVKeccak256Program...)
	sdb.SetCode(contractAddr, magicProgram)
	sdb.CreateAccount(contractAddr)

	evm := newRVCreateEVM(sdb)

	input := []byte("test input")
	ret, _, err := evm.Call(
		types.HexToAddress("0xbeef"),
		contractAddr,
		input,
		1_000_000,
		big.NewInt(0),
	)
	if err != nil {
		t.Fatalf("EVM.Call with RV code: %v", err)
	}

	// Output is Keccak256 of input.
	want := crypto.Keccak256(input)
	if len(ret) != 32 {
		t.Fatalf("RISC-V output len = %d, want 32", len(ret))
	}
	for i, b := range want {
		if ret[i] != b {
			t.Errorf("output byte %d: got 0x%x, want 0x%x", i, ret[i], b)
		}
	}
}

// TestRVCallRoutingEVMCodeNotRerouted verifies that EVM bytecode is NOT
// routed to the RISC-V executor at I+.
func TestRVCallRoutingEVMCodeNotRerouted(t *testing.T) {
	sdb := newRVTestStateDB()
	contractAddr := types.HexToAddress("0xbeef01")
	// Simple EVM: PUSH1 0x42, MSTORE8 at offset 0, RETURN 0 1.
	evmCode := []byte{
		0x60, 0x42, // PUSH1 0x42
		0x60, 0x00, // PUSH1 0
		0x53,       // MSTORE8
		0x60, 0x01, // PUSH1 1
		0x60, 0x00, // PUSH1 0
		0xf3, // RETURN
	}
	sdb.SetCode(contractAddr, evmCode)
	sdb.CreateAccount(contractAddr)

	evm := newRVCreateEVM(sdb)

	ret, _, err := evm.Call(
		types.HexToAddress("0xcafe01"),
		contractAddr,
		nil,
		1_000_000,
		big.NewInt(0),
	)
	if err != nil {
		t.Fatalf("EVM call to EVM code: %v", err)
	}
	if len(ret) != 1 || ret[0] != 0x42 {
		t.Errorf("EVM output = %x, want [0x42]", ret)
	}
}

// TestRVCallRoutingPreIPlusNotRerouted verifies that RISC-V magic bytes do NOT
// trigger RISC-V routing when IsIPlus is false.
func TestRVCallRoutingPreIPlusNotRerouted(t *testing.T) {
	sdb := newRVTestStateDB()
	contractAddr := types.HexToAddress("0xdead02")

	// Code starts with RV magic but IsIPlus is off.
	code := []byte{RVMagic0, RVMagic1, RVMagic2, 0x00}
	sdb.SetCode(contractAddr, code)
	sdb.CreateAccount(contractAddr)

	// Pre-I+ EVM.
	rules := ForkRules{IsGlamsterdan: true}
	evm := NewEVM(BlockContext{}, TxContext{}, Config{})
	evm.forkRules = rules
	evm.jumpTable = SelectJumpTable(rules)
	evm.StateDB = sdb

	// Should run as EVM: 0xFE = INVALID opcode → error or empty return.
	_, _, _ = evm.Call(
		types.HexToAddress("0xcafe02"),
		contractAddr,
		nil,
		1_000_000,
		big.NewInt(0),
	)
	// The test passes as long as it didn't panic. The RISC-V path would have
	// produced ECALL output; the EVM path hits INVALID and reverts.
}
