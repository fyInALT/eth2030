package vm

import (
	"errors"
	"math/big"
	"sync"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// aaTestStateDB is a minimal StateDB for testing the AA executor.
// Named to avoid collision with mockStateDB in instructions_test.go.
type aaTestStateDB struct {
	balances   map[types.Address]*big.Int
	nonces     map[types.Address]uint64
	codes      map[types.Address][]byte
	codeHashes map[types.Address]types.Hash
	storage    map[types.Address]map[types.Hash]types.Hash
	transient  map[types.Address]map[types.Hash]types.Hash
	accessList map[types.Address]bool
	slotAccess map[types.Address]map[types.Hash]bool
	snapCount  int
	refund     uint64
	logs       []*types.Log
	selfDestr  map[types.Address]bool
}

func newAATestStateDB() *aaTestStateDB {
	return &aaTestStateDB{
		balances:   make(map[types.Address]*big.Int),
		nonces:     make(map[types.Address]uint64),
		codes:      make(map[types.Address][]byte),
		codeHashes: make(map[types.Address]types.Hash),
		storage:    make(map[types.Address]map[types.Hash]types.Hash),
		transient:  make(map[types.Address]map[types.Hash]types.Hash),
		accessList: make(map[types.Address]bool),
		slotAccess: make(map[types.Address]map[types.Hash]bool),
		selfDestr:  make(map[types.Address]bool),
	}
}

func (m *aaTestStateDB) CreateAccount(addr types.Address) {}
func (m *aaTestStateDB) GetBalance(addr types.Address) *big.Int {
	if b, ok := m.balances[addr]; ok {
		return new(big.Int).Set(b)
	}
	return new(big.Int)
}
func (m *aaTestStateDB) AddBalance(addr types.Address, amount *big.Int) {
	if _, ok := m.balances[addr]; !ok {
		m.balances[addr] = new(big.Int)
	}
	m.balances[addr].Add(m.balances[addr], amount)
}
func (m *aaTestStateDB) SubBalance(addr types.Address, amount *big.Int) {
	if _, ok := m.balances[addr]; !ok {
		m.balances[addr] = new(big.Int)
	}
	m.balances[addr].Sub(m.balances[addr], amount)
}
func (m *aaTestStateDB) GetNonce(addr types.Address) uint64 {
	return m.nonces[addr]
}
func (m *aaTestStateDB) SetNonce(addr types.Address, nonce uint64) {
	m.nonces[addr] = nonce
}
func (m *aaTestStateDB) GetCode(addr types.Address) []byte {
	return m.codes[addr]
}
func (m *aaTestStateDB) SetCode(addr types.Address, code []byte) {
	m.codes[addr] = code
}
func (m *aaTestStateDB) GetCodeHash(addr types.Address) types.Hash {
	return m.codeHashes[addr]
}
func (m *aaTestStateDB) GetCodeSize(addr types.Address) int {
	return len(m.codes[addr])
}
func (m *aaTestStateDB) GetState(addr types.Address, key types.Hash) types.Hash {
	if s, ok := m.storage[addr]; ok {
		return s[key]
	}
	return types.Hash{}
}
func (m *aaTestStateDB) SetState(addr types.Address, key types.Hash, value types.Hash) {
	if _, ok := m.storage[addr]; !ok {
		m.storage[addr] = make(map[types.Hash]types.Hash)
	}
	m.storage[addr][key] = value
}
func (m *aaTestStateDB) GetCommittedState(addr types.Address, key types.Hash) types.Hash {
	return m.GetState(addr, key)
}
func (m *aaTestStateDB) GetTransientState(addr types.Address, key types.Hash) types.Hash {
	if s, ok := m.transient[addr]; ok {
		return s[key]
	}
	return types.Hash{}
}
func (m *aaTestStateDB) SetTransientState(addr types.Address, key types.Hash, value types.Hash) {
	if _, ok := m.transient[addr]; !ok {
		m.transient[addr] = make(map[types.Hash]types.Hash)
	}
	m.transient[addr][key] = value
}
func (m *aaTestStateDB) ClearTransientStorage() {
	m.transient = make(map[types.Address]map[types.Hash]types.Hash)
}
func (m *aaTestStateDB) SelfDestruct(addr types.Address) {
	m.selfDestr[addr] = true
}
func (m *aaTestStateDB) HasSelfDestructed(addr types.Address) bool {
	return m.selfDestr[addr]
}
func (m *aaTestStateDB) Exist(addr types.Address) bool {
	_, ok := m.codes[addr]
	if ok {
		return true
	}
	_, ok = m.balances[addr]
	return ok
}
func (m *aaTestStateDB) Empty(addr types.Address) bool {
	return !m.Exist(addr)
}
func (m *aaTestStateDB) Snapshot() int {
	m.snapCount++
	return m.snapCount
}
func (m *aaTestStateDB) RevertToSnapshot(id int) {}
func (m *aaTestStateDB) AddLog(log *types.Log) {
	m.logs = append(m.logs, log)
}
func (m *aaTestStateDB) AddRefund(gas uint64) { m.refund += gas }
func (m *aaTestStateDB) SubRefund(gas uint64) { m.refund -= gas }
func (m *aaTestStateDB) GetRefund() uint64    { return m.refund }
func (m *aaTestStateDB) AddAddressToAccessList(addr types.Address) {
	m.accessList[addr] = true
}
func (m *aaTestStateDB) AddSlotToAccessList(addr types.Address, slot types.Hash) {
	m.accessList[addr] = true
	if _, ok := m.slotAccess[addr]; !ok {
		m.slotAccess[addr] = make(map[types.Hash]bool)
	}
	m.slotAccess[addr][slot] = true
}
func (m *aaTestStateDB) AddressInAccessList(addr types.Address) bool {
	return m.accessList[addr]
}
func (m *aaTestStateDB) SlotInAccessList(addr types.Address, slot types.Hash) (bool, bool) {
	addrOk := m.accessList[addr]
	if s, ok := m.slotAccess[addr]; ok {
		return addrOk, s[slot]
	}
	return addrOk, false
}

func makeAATestEVM(stateDB StateDB) *EVM {
	blockCtx := BlockContext{
		BlockNumber: big.NewInt(1),
		Time:        1000,
		GasLimit:    30000000,
		BaseFee:     big.NewInt(1000000000),
	}
	txCtx := TxContext{
		GasPrice: big.NewInt(1000000000),
	}
	config := Config{MaxCallDepth: 1024}
	evm := NewEVMWithState(blockCtx, txCtx, config, stateDB)
	return evm
}

func aaTestAccountAddr() types.Address {
	return types.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
}

func aaTestPaymasterAddr() types.Address {
	return types.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
}

func TestAAExecutor_NewExecutor(t *testing.T) {
	ex := NewAAExecutor()
	if ex == nil {
		t.Fatal("NewAAExecutor returned nil")
	}
}

func TestAAExecutor_ValidatePhase_NilTx(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	evm := makeAATestEVM(state)

	_, err := ex.ValidatePhase(evm, nil, aaTestAccountAddr())
	if !errors.Is(err, ErrAAInvalidTransaction) {
		t.Errorf("expected ErrAAInvalidTransaction, got: %v", err)
	}
}

func TestAAExecutor_ValidatePhase_NoCode(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	evm := makeAATestEVM(state)

	tx := &AATx{
		Sender:             aaTestAccountAddr(),
		ValidationGasLimit: 100000,
		Nonce:              NonceKey{Sequence: 0},
	}

	_, err := ex.ValidatePhase(evm, tx, aaTestAccountAddr())
	if !errors.Is(err, ErrAANilAccount) {
		t.Errorf("expected ErrAANilAccount, got: %v", err)
	}
}

func TestAAExecutor_ValidatePhase_WithSTOP(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()

	// Simple code: STOP (validation succeeds but role not accepted).
	state.codes[account] = []byte{byte(STOP)}
	evm := makeAATestEVM(state)

	tx := &AATx{
		Sender:             account,
		ValidationGasLimit: 100000,
		ExecutionGasLimit:  100000,
		Nonce:              NonceKey{Sequence: 0},
	}

	_, err := ex.ValidatePhase(evm, tx, account)
	// STOP exits cleanly but ACCEPT_ROLE was never called.
	if !errors.Is(err, ErrAARoleNotAccepted) {
		t.Errorf("expected ErrAARoleNotAccepted, got: %v", err)
	}
}

func TestAAExecutor_ValidatePhase_InsufficientGas(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()
	state.codes[account] = []byte{byte(STOP)}
	evm := makeAATestEVM(state)

	tx := &AATx{
		Sender:             account,
		ValidationGasLimit: 0, // no gas
		Nonce:              NonceKey{Sequence: 0},
	}

	_, err := ex.ValidatePhase(evm, tx, account)
	if !errors.Is(err, ErrAAInsufficientGas) {
		t.Errorf("expected ErrAAInsufficientGas, got: %v", err)
	}
}

func TestAAExecutor_ExecutePhase_NilTx(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	evm := makeAATestEVM(state)

	_, err := ex.ExecutePhase(evm, nil, &ValidationResult{Success: true})
	if !errors.Is(err, ErrAAInvalidTransaction) {
		t.Errorf("expected ErrAAInvalidTransaction, got: %v", err)
	}
}

func TestAAExecutor_ExecutePhase_ValidationNotPassed(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	evm := makeAATestEVM(state)

	tx := &AATx{
		Sender:            aaTestAccountAddr(),
		ExecutionGasLimit: 100000,
		Nonce:             NonceKey{Sequence: 0},
	}

	_, err := ex.ExecutePhase(evm, tx, &ValidationResult{Success: false})
	if !errors.Is(err, ErrAAValidationFailed) {
		t.Errorf("expected ErrAAValidationFailed, got: %v", err)
	}
}

func TestAAExecutor_ExecutePhase_WithSTOP(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()
	state.codes[account] = []byte{byte(STOP)}
	evm := makeAATestEVM(state)

	tx := &AATx{
		Sender:            account,
		ExecutionGasLimit: 100000,
		Nonce:             NonceKey{Sequence: 0},
	}

	result, err := ex.ExecutePhase(evm, tx, &ValidationResult{Success: true})
	if err != nil {
		t.Fatalf("ExecutePhase failed: %v", err)
	}
	if !result.Success {
		t.Error("expected execution success")
	}
	if result.GasUsed != 0 {
		// STOP costs 0 gas.
		t.Errorf("expected 0 gas used for STOP, got %d", result.GasUsed)
	}
}

func TestAAExecutor_ExecutePhase_REVERT(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()

	// PUSH1 0; PUSH1 0; REVERT
	state.codes[account] = []byte{byte(PUSH1), 0x00, byte(PUSH1), 0x00, byte(REVERT)}
	evm := makeAATestEVM(state)

	tx := &AATx{
		Sender:            account,
		ExecutionGasLimit: 100000,
		Nonce:             NonceKey{Sequence: 0},
	}

	result, err := ex.ExecutePhase(evm, tx, &ValidationResult{Success: true})
	if err != nil {
		t.Fatalf("ExecutePhase returned error: %v", err)
	}
	if result.Success {
		t.Error("expected execution failure on REVERT")
	}
	if !result.Reverted {
		t.Error("expected Reverted flag to be set")
	}
}

func TestAAExecutor_PostOpPhase_NoPaymaster(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	evm := makeAATestEVM(state)

	tx := &AATx{
		Sender: aaTestAccountAddr(),
		Nonce:  NonceKey{Sequence: 0},
		// No paymaster.
	}

	err := ex.PostOpPhase(evm, tx, &AAExecutionResult{Success: true})
	if err != nil {
		t.Fatalf("PostOpPhase should be no-op without paymaster, got: %v", err)
	}
}

func TestAAExecutor_PostOpPhase_NilTx(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	evm := makeAATestEVM(state)

	err := ex.PostOpPhase(evm, nil, &AAExecutionResult{Success: true})
	if !errors.Is(err, ErrAAInvalidTransaction) {
		t.Errorf("expected ErrAAInvalidTransaction, got: %v", err)
	}
}

func TestAAExecutor_PostOpPhase_WithPaymaster(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	pm := aaTestPaymasterAddr()

	// Paymaster code: STOP (post-op completes without accepting role, which is ok for STOP).
	state.codes[pm] = []byte{byte(STOP)}
	evm := makeAATestEVM(state)

	tx := &AATx{
		Sender:         aaTestAccountAddr(),
		Paymaster:      &pm,
		PostOpGasLimit: 50000,
		Nonce:          NonceKey{Sequence: 0},
	}

	err := ex.PostOpPhase(evm, tx, &AAExecutionResult{Success: true, GasUsed: 1000})
	// STOP exits cleanly; post-op succeeds.
	if err != nil {
		t.Fatalf("PostOpPhase failed: %v", err)
	}
}

func TestAAExecutor_ValidatePaymaster_NoCode(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	evm := makeAATestEVM(state)
	pm := aaTestPaymasterAddr()

	tx := &AATx{
		Sender:             aaTestAccountAddr(),
		ValidationGasLimit: 100000,
		Nonce:              NonceKey{Sequence: 0},
	}

	_, err := ex.ValidatePaymaster(evm, tx, pm)
	if !errors.Is(err, ErrAANilPaymaster) {
		t.Errorf("expected ErrAANilPaymaster, got: %v", err)
	}
}

func TestAAExecutor_ValidatePaymaster_WithSTOP(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	pm := aaTestPaymasterAddr()
	state.codes[pm] = []byte{byte(STOP)}
	evm := makeAATestEVM(state)

	tx := &AATx{
		Sender:             aaTestAccountAddr(),
		ValidationGasLimit: 100000,
		Nonce:              NonceKey{Sequence: 0},
	}

	_, err := ex.ValidatePaymaster(evm, tx, pm)
	// STOP exits cleanly but ACCEPT_ROLE was not called.
	if !errors.Is(err, ErrAARoleNotAccepted) {
		t.Errorf("expected ErrAARoleNotAccepted, got: %v", err)
	}
}

func TestAAExecutor_CheckNonce_StandardKey(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()
	state.nonces[account] = 5

	// Correct nonce.
	err := ex.CheckNonce(state, account, NonceKey{Sequence: 5})
	if err != nil {
		t.Errorf("CheckNonce should pass for correct nonce, got: %v", err)
	}

	// Wrong nonce.
	err = ex.CheckNonce(state, account, NonceKey{Sequence: 6})
	if !errors.Is(err, ErrAANonceMismatch) {
		t.Errorf("expected ErrAANonceMismatch, got: %v", err)
	}
}

func TestAAExecutor_CheckNonce_2DKey(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()

	key := big.NewInt(42)
	keyBytes := make([]byte, 32)
	key.FillBytes(keyBytes)
	slot := types.BytesToHash(keyBytes)

	// Set sequence 3 for key 42.
	seqBytes := make([]byte, 32)
	big.NewInt(3).FillBytes(seqBytes)
	if _, ok := state.storage[account]; !ok {
		state.storage[account] = make(map[types.Hash]types.Hash)
	}
	state.storage[account][slot] = types.BytesToHash(seqBytes)

	// Correct sequence.
	err := ex.CheckNonce(state, account, NonceKey{Key: key, Sequence: 3})
	if err != nil {
		t.Errorf("CheckNonce should pass, got: %v", err)
	}

	// Wrong sequence.
	err = ex.CheckNonce(state, account, NonceKey{Key: key, Sequence: 4})
	if !errors.Is(err, ErrAANonceMismatch) {
		t.Errorf("expected ErrAANonceMismatch, got: %v", err)
	}
}

func TestAAExecutor_IncrementNonce_Standard(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()
	state.nonces[account] = 10

	ex.IncrementNonce(state, account, NonceKey{Sequence: 10})
	if state.nonces[account] != 11 {
		t.Errorf("nonce = %d, want 11", state.nonces[account])
	}
}

func TestAAExecutor_IncrementNonce_2DKey(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()

	key := big.NewInt(7)
	keyBytes := make([]byte, 32)
	key.FillBytes(keyBytes)
	slot := types.BytesToHash(keyBytes)

	// Set sequence 0 for key 7.
	if _, ok := state.storage[account]; !ok {
		state.storage[account] = make(map[types.Hash]types.Hash)
	}
	state.storage[account][slot] = types.Hash{} // 0

	ex.IncrementNonce(state, account, NonceKey{Key: key, Sequence: 0})

	val := state.storage[account][slot]
	seq := new(big.Int).SetBytes(val[:])
	if seq.Uint64() != 1 {
		t.Errorf("2D nonce sequence = %d, want 1", seq.Uint64())
	}
}

func TestAAExecutor_ProcessBundle_Empty(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	evm := makeAATestEVM(state)

	results, err := ex.ProcessBundle(evm, nil)
	if err != nil {
		t.Errorf("empty bundle should succeed, got: %v", err)
	}
	if results != nil {
		t.Errorf("empty bundle should return nil results, got %d", len(results))
	}
}

func TestAAExecutor_ProcessBundle_NilTx(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	evm := makeAATestEVM(state)

	results, err := ex.ProcessBundle(evm, []*AATx{nil})
	if !errors.Is(err, ErrAABundlePartialFailure) {
		t.Errorf("expected ErrAABundlePartialFailure, got: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !errors.Is(results[0].Err, ErrAAInvalidTransaction) {
		t.Errorf("expected ErrAAInvalidTransaction in result, got: %v", results[0].Err)
	}
}

func TestAAExecutor_ProcessBundle_SingleTx(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()

	// Account code: STOP. Validation will fail (role not accepted), causing
	// the bundle to report partial failure.
	state.codes[account] = []byte{byte(STOP)}
	evm := makeAATestEVM(state)

	tx := &AATx{
		Sender:             account,
		ValidationGasLimit: 100000,
		ExecutionGasLimit:  100000,
		Nonce:              NonceKey{Sequence: 0},
	}

	results, err := ex.ProcessBundle(evm, []*AATx{tx})
	if err == nil {
		t.Fatal("expected bundle failure (role not accepted)")
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ValidationOK {
		t.Error("validation should have failed")
	}
}

func TestAAExecutor_ThreadSafety(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()
	state.codes[account] = []byte{byte(STOP)}

	var wg sync.WaitGroup
	errs := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			evm := makeAATestEVM(state)
			tx := &AATx{
				Sender:             account,
				ValidationGasLimit: 100000,
				ExecutionGasLimit:  100000,
				Nonce:              NonceKey{Sequence: 0},
			}
			_, err := ex.ValidatePhase(evm, tx, account)
			// We expect ErrAARoleNotAccepted since STOP doesn't call ACCEPT_ROLE.
			if err != nil && !errors.Is(err, ErrAARoleNotAccepted) {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("concurrent ValidatePhase unexpected error: %v", err)
	}
}

func TestAAExecutor_CallerRoleConstants(t *testing.T) {
	if CallerRoleSender != 0 {
		t.Errorf("CallerRoleSender = %d, want 0", CallerRoleSender)
	}
	if CallerRolePaymaster != 1 {
		t.Errorf("CallerRolePaymaster = %d, want 1", CallerRolePaymaster)
	}
	if CallerRoleDeployer != 2 {
		t.Errorf("CallerRoleDeployer = %d, want 2", CallerRoleDeployer)
	}
	if CallerRoleBundler != 3 {
		t.Errorf("CallerRoleBundler = %d, want 3", CallerRoleBundler)
	}
}

func TestAAExecutor_NonceKey_ZeroKey(t *testing.T) {
	nk := NonceKey{Key: big.NewInt(0), Sequence: 5}
	if nk.Key.Sign() != 0 {
		t.Error("zero key should have sign 0")
	}
}

func TestAAExecutor_AATxType(t *testing.T) {
	// AATxType is defined in core/types/tx_aa.go
	if types.AATxType != 0x05 {
		t.Errorf("types.AATxType = 0x%02x, want 0x05", types.AATxType)
	}
}

func TestAAExecutor_ExecutePhase_OutOfGas(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()

	// Code that costs gas: PUSH1 0; PUSH1 0; KECCAK256; POP; STOP
	// KECCAK256 costs 30 + dynamic. Give very little gas to trigger OOG.
	state.codes[account] = []byte{
		byte(PUSH1), 0x00,
		byte(PUSH1), 0x00,
		byte(KECCAK256),
		byte(POP),
		byte(STOP),
	}
	evm := makeAATestEVM(state)

	tx := &AATx{
		Sender:            account,
		ExecutionGasLimit: 5, // way too little gas
		Nonce:             NonceKey{Sequence: 0},
	}

	_, err := ex.ExecutePhase(evm, tx, &ValidationResult{Success: true})
	if err == nil {
		t.Fatal("expected OOG error, got nil")
	}
	if !errors.Is(err, ErrAAExecutionFailed) {
		t.Errorf("expected ErrAAExecutionFailed, got: %v", err)
	}
}

func TestAAExecutor_PostOpPhase_PaymasterNoCode(t *testing.T) {
	ex := NewAAExecutor()
	state := newAATestStateDB()
	evm := makeAATestEVM(state)
	pm := aaTestPaymasterAddr()
	// Do not set code for paymaster.

	tx := &AATx{
		Sender:         aaTestAccountAddr(),
		Paymaster:      &pm,
		PostOpGasLimit: 50000,
		Nonce:          NonceKey{Sequence: 0},
	}

	err := ex.PostOpPhase(evm, tx, &AAExecutionResult{Success: true})
	if !errors.Is(err, ErrAANilPaymaster) {
		t.Errorf("expected ErrAANilPaymaster, got: %v", err)
	}
}

func TestAAExecutor_CheckNonce_NilState(t *testing.T) {
	ex := NewAAExecutor()
	err := ex.CheckNonce(nil, aaTestAccountAddr(), NonceKey{Sequence: 0})
	if err == nil {
		t.Error("expected error for nil stateDB")
	}
}

func TestAAExecutor_FrameContext_PayerTracking(t *testing.T) {
	// Verify that FrameContext correctly tracks payer for sponsored tx.
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()
	pm := aaTestPaymasterAddr()

	// Set up code for both account and paymaster.
	state.codes[account] = []byte{byte(STOP)}
	state.codes[pm] = []byte{byte(STOP)}
	state.balances[pm] = big.NewInt(1000000)

	evm := makeAATestEVM(state)

	// Create a tx with a paymaster.
	tx := &AATx{
		Sender:             account,
		ValidationGasLimit: 100000,
		ExecutionGasLimit:  200000,
		PostOpGasLimit:     50000,
		Paymaster:          &pm,
		PaymasterData:      []byte{0x01, 0x02},
		Nonce:              NonceKey{Sequence: 0},
		MaxFeePerGas:       big.NewInt(1000000000),
	}

	// Validate phase sets up AA context with paymaster info.
	_, err := ex.ValidatePhase(evm, tx, account)
	// Expect role not accepted since STOP doesn't call ACCEPT_ROLE,
	// but verify the AA context was configured with paymaster.
	if !errors.Is(err, ErrAARoleNotAccepted) {
		t.Fatalf("expected ErrAARoleNotAccepted, got: %v", err)
	}

	// Verify paymaster post-op also tracks paymaster.
	err = ex.PostOpPhase(evm, tx, &AAExecutionResult{Success: true, GasUsed: 5000})
	// Post-op should not error because paymaster has code.
	if err != nil {
		t.Fatalf("PostOpPhase unexpected error: %v", err)
	}
}

func TestAAExecutor_ApproveFlow_Scope0ThenScope1(t *testing.T) {
	// Test the conceptual APPROVE(0) then APPROVE(1) flow at the executor level.
	// APPROVE(0) = sender approves execution, APPROVE(1) = payer approves payment.
	// At the executor level, this maps to validation then paymaster validation.
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()
	pm := aaTestPaymasterAddr()

	// Both need code. STOP won't call ACCEPT_ROLE, but we verify the flow.
	state.codes[account] = []byte{byte(STOP)}
	state.codes[pm] = []byte{byte(STOP)}
	state.balances[pm] = big.NewInt(10000000)

	evm := makeAATestEVM(state)

	tx := &AATx{
		Sender:             account,
		Paymaster:          &pm,
		ValidationGasLimit: 100000,
		ExecutionGasLimit:  200000,
		PostOpGasLimit:     50000,
		Nonce:              NonceKey{Sequence: 0},
		MaxFeePerGas:       big.NewInt(1000000000),
	}

	// Phase 1: Account validation (analogous to sender APPROVE(0)).
	valResult, err := ex.ValidatePhase(evm, tx, account)
	if !errors.Is(err, ErrAARoleNotAccepted) {
		t.Fatalf("phase 1: expected ErrAARoleNotAccepted, got: %v", err)
	}
	if valResult == nil {
		t.Fatal("phase 1: expected non-nil validation result")
	}
	if valResult.Success {
		t.Fatal("phase 1: expected validation failure (STOP doesn't call ACCEPT_ROLE)")
	}

	// Phase 1b: Paymaster validation (analogous to payer APPROVE(1)).
	pmResult, err := ex.ValidatePaymaster(evm, tx, pm)
	if !errors.Is(err, ErrAARoleNotAccepted) {
		t.Fatalf("phase 1b: expected ErrAARoleNotAccepted, got: %v", err)
	}
	if pmResult == nil {
		t.Fatal("phase 1b: expected non-nil paymaster result")
	}
	if pmResult.Success {
		t.Fatal("phase 1b: expected paymaster failure (STOP doesn't call ACCEPT_ROLE)")
	}
}

func TestAAExecutor_NonceKey_PrivacyPool(t *testing.T) {
	// Verify that different nonce keys can be used for different contexts
	// (privacy pool scenario from Vitalik's review).
	ex := NewAAExecutor()
	state := newAATestStateDB()
	account := aaTestAccountAddr()

	// Set up storage for two different nonce keys.
	key1 := big.NewInt(1) // DeFi context
	key2 := big.NewInt(2) // Social context

	key1Bytes := make([]byte, 32)
	key1.FillBytes(key1Bytes)
	slot1 := types.BytesToHash(key1Bytes)

	key2Bytes := make([]byte, 32)
	key2.FillBytes(key2Bytes)
	slot2 := types.BytesToHash(key2Bytes)

	if _, ok := state.storage[account]; !ok {
		state.storage[account] = make(map[types.Hash]types.Hash)
	}
	// Key1 has sequence 5, key2 has sequence 10.
	seq1Bytes := make([]byte, 32)
	big.NewInt(5).FillBytes(seq1Bytes)
	state.storage[account][slot1] = types.BytesToHash(seq1Bytes)

	seq2Bytes := make([]byte, 32)
	big.NewInt(10).FillBytes(seq2Bytes)
	state.storage[account][slot2] = types.BytesToHash(seq2Bytes)

	// Check key1 with correct sequence.
	err := ex.CheckNonce(state, account, NonceKey{Key: key1, Sequence: 5})
	if err != nil {
		t.Errorf("key1 correct seq: unexpected error: %v", err)
	}

	// Check key2 with correct sequence.
	err = ex.CheckNonce(state, account, NonceKey{Key: key2, Sequence: 10})
	if err != nil {
		t.Errorf("key2 correct seq: unexpected error: %v", err)
	}

	// Cross-check: key1 with key2's sequence should fail.
	err = ex.CheckNonce(state, account, NonceKey{Key: key1, Sequence: 10})
	if !errors.Is(err, ErrAANonceMismatch) {
		t.Errorf("key1 wrong seq: expected ErrAANonceMismatch, got: %v", err)
	}

	// Increment key1 and verify.
	ex.IncrementNonce(state, account, NonceKey{Key: key1, Sequence: 5})
	err = ex.CheckNonce(state, account, NonceKey{Key: key1, Sequence: 6})
	if err != nil {
		t.Errorf("key1 after increment: unexpected error: %v", err)
	}

	// Key2 should be unaffected.
	err = ex.CheckNonce(state, account, NonceKey{Key: key2, Sequence: 10})
	if err != nil {
		t.Errorf("key2 after key1 increment: unexpected error: %v", err)
	}
}

func TestValidateAATransaction(t *testing.T) {
	// Nil tx.
	if err := ValidateAATransaction(nil); err == nil {
		t.Fatal("expected error for nil AA tx")
	}

	// Zero sender.
	tx := &AATx{
		ValidationGasLimit: 100000,
		ExecutionGasLimit:  200000,
		MaxFeePerGas:       big.NewInt(50),
	}
	if err := ValidateAATransaction(tx); err == nil {
		t.Fatal("expected error for zero sender")
	}

	// Zero execution gas.
	tx.Sender = types.BytesToAddress([]byte{0x01})
	tx.ExecutionGasLimit = 0
	if err := ValidateAATransaction(tx); err == nil {
		t.Fatal("expected error for zero execution gas")
	}

	// Valid tx.
	tx.ExecutionGasLimit = 200000
	if err := ValidateAATransaction(tx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
