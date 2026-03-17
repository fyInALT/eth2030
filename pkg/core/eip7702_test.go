package core

import (
	"crypto/ecdsa"
	"math/big"
	"strings"
	"testing"

	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/eips"
	"github.com/eth2030/eth2030/core/execution"
	"github.com/eth2030/eth2030/core/gaspool"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
)

// signAuthorization creates a properly signed EIP-7702 authorization.
func signAuthorization(privKey *ecdsa.PrivateKey, chainID *big.Int, targetAddr types.Address, nonce uint64) types.Authorization {
	auth := types.Authorization{
		ChainID: chainID,
		Address: targetAddr,
		Nonce:   nonce,
	}

	// Use the same hash computation as eips.computeAuthorizationHash but via
	// the exported RecoverAuthority path. We build the hash inline here to avoid
	// depending on the unexported helper.
	authHash := eips.AuthorizationHash(&auth)
	sig, err := crypto.Sign(authHash, privKey)
	if err != nil {
		panic("failed to sign authorization: " + err.Error())
	}

	auth.R = new(big.Int).SetBytes(sig[:32])
	auth.S = new(big.Int).SetBytes(sig[32:64])
	auth.V = new(big.Int).SetUint64(uint64(sig[64]))

	return auth
}

func TestIntrinsicGas_NoAuthorizations(t *testing.T) {
	// Without authorizations, the intrinsic gas should match the old behavior.
	data := []byte{0x00, 0x01, 0x00, 0xff}
	isCreate := false

	gas := execution.IntrinsicGas(data, isCreate, false, 0)
	// execution.TxGas(21000) + 2*execution.TxDataZeroGas(4) + 2*execution.TxDataNonZeroGas(16) = 21040
	expected := execution.TxGas + 2*execution.TxDataZeroGas + 2*execution.TxDataNonZeroGas
	if gas != expected {
		t.Errorf("intrinsicGas without auths: got %d, want %d", gas, expected)
	}
}

func TestIntrinsicGas_WithAuthorizations(t *testing.T) {
	data := []byte{}
	isCreate := false

	// Per EIP-7702: charge PER_EMPTY_ACCOUNT_COST (25000) for each authorization.
	// If authority is not empty, refund is added during authorization processing.
	gas := execution.IntrinsicGas(data, isCreate, false, 3)
	expected := execution.TxGas + 3*execution.PerEmptyAccountCost
	if gas != expected {
		t.Errorf("intrinsicGas with 3 auths: got %d, want %d", gas, expected)
	}
}

func TestIntrinsicGas_SingleAuthEmptyAccount(t *testing.T) {
	data := []byte{}
	gas := execution.IntrinsicGas(data, false, false, 1)
	// Per EIP-7702: charge PER_EMPTY_ACCOUNT_COST (25000) for each authorization.
	expected := execution.TxGas + execution.PerEmptyAccountCost
	if gas != expected {
		t.Errorf("intrinsicGas with 1 auth: got %d, want %d", gas, expected)
	}
}

func TestIntrinsicGas_AuthWithDataAndCreate(t *testing.T) {
	data := []byte{0x60, 0x80, 0x60, 0x40} // 4 non-zero bytes
	gas := execution.IntrinsicGas(data, true, false, 2)
	// Per EIP-7702: charge PER_EMPTY_ACCOUNT_COST (25000) for each authorization.
	expected := execution.TxGas + execution.TxCreateGas + 4*execution.TxDataNonZeroGas + 2*execution.PerEmptyAccountCost
	if gas != expected {
		t.Errorf("intrinsicGas with create+auth: got %d, want %d", gas, expected)
	}
}

func TestSetCodeTx_AuthorizationProcessedInApplyMessage(t *testing.T) {
	statedb := state.NewMemoryStateDB()

	// Generate a key pair for the authorization signer.
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	signerAddr := crypto.PubkeyToAddress(privKey.PublicKey)

	// Fund the signer account so it exists in state (authorization checks nonce).
	statedb.AddBalance(signerAddr, big.NewInt(0))

	// The transaction sender (different from the authorization signer).
	sender := types.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	tenETH := new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(1e18))
	statedb.AddBalance(sender, tenETH)

	// The target contract address that the signer delegates to.
	targetContract := types.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

	// Create a properly signed authorization.
	chainID := config.TestConfig.ChainID
	auth := signAuthorization(privKey, chainID, targetContract, 0)

	// Build a SetCode transaction message directly (bypass Transaction struct).
	to := types.HexToAddress("0xcccccccccccccccccccccccccccccccccccccccc")
	msg := config.Message{
		From:      sender,
		To:        &to,
		Nonce:     0,
		Value:     new(big.Int),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(1_000_000_000),
		GasTipCap: big.NewInt(0),
		Data:      nil,
		AuthList:  []types.Authorization{auth},
		TxType:    types.SetCodeTxType,
	}

	header := newTestHeader()
	gp := new(gaspool.GasPool).AddGas(header.GasLimit)

	result, err := execution.ApplyMessage(config.TestConfig, nil, statedb, header, &msg, gp)
	if err != nil {
		t.Fatalf("applyMessage failed: %v", err)
	}

	// The tx should succeed (or at least not be a protocol error).
	_ = result

	// Verify that the signer's code was set to the delegation designator.
	code := statedb.GetCode(signerAddr)
	if !eips.IsDelegated(code) {
		t.Fatalf("signer code should be a delegation designator after SetCode tx, got %x", code)
	}

	resolved, ok := eips.ResolveDelegation(code)
	if !ok {
		t.Fatal("ResolveDelegation should succeed on signer's code")
	}
	if resolved != targetContract {
		t.Errorf("delegation target: got %v, want %v", resolved.Hex(), targetContract.Hex())
	}

	// Signer nonce should be incremented by the authorization processing.
	if statedb.GetNonce(signerAddr) != 1 {
		t.Errorf("signer nonce: got %d, want 1", statedb.GetNonce(signerAddr))
	}
}

func TestSetCodeTx_IntrinsicGasIncludesAuthCosts(t *testing.T) {
	statedb := state.NewMemoryStateDB()

	sender := types.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	tenETH := new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(1e18))
	statedb.AddBalance(sender, tenETH)

	// Build a SetCode message with one authorization to a non-existent (empty) address.
	to := types.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	emptyTarget := types.HexToAddress("0x1111111111111111111111111111111111111111")

	msg := config.Message{
		From:      sender,
		To:        &to,
		Nonce:     0,
		Value:     new(big.Int),
		GasLimit:  100000,
		GasFeeCap: big.NewInt(1),
		GasTipCap: big.NewInt(0),
		GasPrice:  big.NewInt(1),
		Data:      nil,
		AuthList: []types.Authorization{
			{
				ChainID: big.NewInt(1),
				Address: emptyTarget,
				Nonce:   0,
				V:       big.NewInt(0),
				R:       big.NewInt(1),
				S:       big.NewInt(1),
			},
		},
		TxType: types.SetCodeTxType,
	}

	header := &types.Header{
		Number:   big.NewInt(1),
		GasLimit: 10_000_000,
		Time:     1000,
		BaseFee:  big.NewInt(1),
		Coinbase: types.HexToAddress("0xfee"),
	}
	gp := new(gaspool.GasPool).AddGas(header.GasLimit)

	result, err := execution.ApplyMessage(config.TestConfig, nil, statedb, header, &msg, gp)
	if err != nil {
		t.Fatalf("applyMessage failed: %v", err)
	}

	// Per EIP-7702: charge PER_EMPTY_ACCOUNT_COST (25000) for each authorization.
	// Refund is added later if authority is not empty.
	expectedIntrinsic := execution.TxGas + execution.PerEmptyAccountCost
	if result.UsedGas < expectedIntrinsic {
		t.Errorf("gas used %d is less than expected intrinsic %d", result.UsedGas, expectedIntrinsic)
	}
}

func TestSetCodeTx_IntrinsicGasTooLow(t *testing.T) {
	statedb := state.NewMemoryStateDB()

	sender := types.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	tenETH := new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(1e18))
	statedb.AddBalance(sender, tenETH)

	to := types.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

	// Set gas limit to just the base execution.TxGas, which is insufficient for a
	// SetCode tx with authorizations.
	msg := config.Message{
		From:      sender,
		To:        &to,
		Nonce:     0,
		Value:     new(big.Int),
		GasLimit:  execution.TxGas, // 21000 - not enough for auth costs
		GasFeeCap: big.NewInt(1),
		GasTipCap: big.NewInt(0),
		GasPrice:  big.NewInt(1),
		Data:      nil,
		AuthList: []types.Authorization{
			{
				ChainID: big.NewInt(1),
				Address: types.HexToAddress("0x1111111111111111111111111111111111111111"),
				Nonce:   0,
				V:       big.NewInt(0),
				R:       big.NewInt(1),
				S:       big.NewInt(1),
			},
		},
		TxType: types.SetCodeTxType,
	}

	header := &types.Header{
		Number:   big.NewInt(1),
		GasLimit: 10_000_000,
		Time:     1000,
		BaseFee:  big.NewInt(1),
		Coinbase: types.HexToAddress("0xfee"),
	}
	gp := new(gaspool.GasPool).AddGas(header.GasLimit)

	_, err := execution.ApplyMessage(config.TestConfig, nil, statedb, header, &msg, gp)
	// Intrinsic gas too low is now returned as a protocol error (matching go-ethereum).
	if err == nil {
		t.Fatal("SetCode tx with insufficient gas should return error")
	}
	if err.Error() == "" || !strings.Contains(err.Error(), "intrinsic gas too low") {
		t.Errorf("expected intrinsic gas error, got: %v", err)
	}
}

func TestSetCodeTx_NonExistentAccountAuthGas(t *testing.T) {
	statedb := state.NewMemoryStateDB()

	sender := types.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	tenETH := new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(1e18))
	statedb.AddBalance(sender, tenETH)

	// One auth targeting an existing account, one targeting a non-existent account.
	existingAddr := types.HexToAddress("0x2222222222222222222222222222222222222222")
	statedb.AddBalance(existingAddr, big.NewInt(1)) // Make it exist and non-empty.

	nonExistentAddr := types.HexToAddress("0x3333333333333333333333333333333333333333")

	to := types.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	msg := config.Message{
		From:      sender,
		To:        &to,
		Nonce:     0,
		Value:     new(big.Int),
		GasLimit:  200000,
		GasFeeCap: big.NewInt(1),
		GasTipCap: big.NewInt(0),
		GasPrice:  big.NewInt(1),
		Data:      nil,
		AuthList: []types.Authorization{
			{
				ChainID: big.NewInt(1),
				Address: existingAddr,
				Nonce:   0,
				V:       big.NewInt(0),
				R:       big.NewInt(1),
				S:       big.NewInt(1),
			},
			{
				ChainID: big.NewInt(1),
				Address: nonExistentAddr,
				Nonce:   0,
				V:       big.NewInt(0),
				R:       big.NewInt(1),
				S:       big.NewInt(1),
			},
		},
		TxType: types.SetCodeTxType,
	}

	header := &types.Header{
		Number:   big.NewInt(1),
		GasLimit: 10_000_000,
		Time:     1000,
		BaseFee:  big.NewInt(1),
		Coinbase: types.HexToAddress("0xfee"),
	}
	gp := new(gaspool.GasPool).AddGas(header.GasLimit)

	result, err := execution.ApplyMessage(config.TestConfig, nil, statedb, header, &msg, gp)
	if err != nil {
		t.Fatalf("applyMessage failed: %v", err)
	}

	// Per EIP-7702: charge PER_EMPTY_ACCOUNT_COST (25000) for each authorization.
	// Refund is added later if authority is not empty.
	expectedIntrinsic := execution.TxGas + 2*execution.PerEmptyAccountCost
	if result.UsedGas < expectedIntrinsic {
		t.Errorf("gas used %d is less than expected intrinsic %d", result.UsedGas, expectedIntrinsic)
	}
}

func TestSetCodeTx_LegacyTxUnaffected(t *testing.T) {
	// Verify that a legacy transaction (type 0x00) is not affected by
	// EIP-7702 processing even if the message has no auth list.
	statedb := state.NewMemoryStateDB()

	sender := types.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	tenETH := new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(1e18))
	statedb.AddBalance(sender, tenETH)

	to := types.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	msg := config.Message{
		From:     sender,
		To:       &to,
		Nonce:    0,
		Value:    new(big.Int),
		GasLimit: 21000,
		GasPrice: big.NewInt(1),
		Data:     nil,
		TxType:   types.LegacyTxType,
	}

	header := &types.Header{
		Number:   big.NewInt(1),
		GasLimit: 10_000_000,
		Time:     1000,
		BaseFee:  big.NewInt(1),
		Coinbase: types.HexToAddress("0xfee"),
	}
	gp := new(gaspool.GasPool).AddGas(header.GasLimit)

	result, err := execution.ApplyMessage(config.TestConfig, nil, statedb, header, &msg, gp)
	if err != nil {
		t.Fatalf("applyMessage failed: %v", err)
	}
	// Should use exactly execution.TxGas for a simple transfer.
	if result.UsedGas != execution.TxGas {
		t.Errorf("legacy tx gas: got %d, want %d", result.UsedGas, execution.TxGas)
	}
}

func TestTransactionToMessage_SetCodeTx(t *testing.T) {
	to := types.HexToAddress("0xdead")
	inner := &types.SetCodeTx{
		ChainID:   big.NewInt(1),
		Nonce:     5,
		GasTipCap: big.NewInt(1),
		GasFeeCap: big.NewInt(20_000_000_000),
		Gas:       100000,
		To:        to,
		Value:     big.NewInt(0),
		Data:      []byte{0xab},
		AuthorizationList: []types.Authorization{
			{
				ChainID: big.NewInt(1),
				Address: types.HexToAddress("0xbeef"),
				Nonce:   0,
				V:       big.NewInt(0),
				R:       big.NewInt(1),
				S:       big.NewInt(2),
			},
		},
	}
	tx := types.NewTransaction(inner)
	msg := config.TransactionToMessage(tx)

	if msg.TxType != types.SetCodeTxType {
		t.Errorf("TxType: got %d, want %d", msg.TxType, types.SetCodeTxType)
	}
	if len(msg.AuthList) != 1 {
		t.Fatalf("AuthList length: got %d, want 1", len(msg.AuthList))
	}
	if msg.AuthList[0].Address != types.HexToAddress("0xbeef") {
		t.Errorf("AuthList[0].Address: got %v, want 0xbeef", msg.AuthList[0].Address.Hex())
	}
	if msg.AuthList[0].Nonce != 0 {
		t.Errorf("AuthList[0].Nonce: got %d, want 0", msg.AuthList[0].Nonce)
	}
}

func TestTransactionToMessage_LegacyTxHasNoAuthList(t *testing.T) {
	to := types.HexToAddress("0xdead")
	inner := &types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1),
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(0),
	}
	tx := types.NewTransaction(inner)
	msg := config.TransactionToMessage(tx)

	if msg.TxType != types.LegacyTxType {
		t.Errorf("TxType: got %d, want %d", msg.TxType, types.LegacyTxType)
	}
	if msg.AuthList != nil {
		t.Errorf("AuthList should be nil for legacy tx, got %v", msg.AuthList)
	}
}

func TestSetCodeTx_MultipleAuthsWithSignatures(t *testing.T) {
	statedb := state.NewMemoryStateDB()

	// Generate two key pairs.
	privKey1, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key1: %v", err)
	}
	signer1 := crypto.PubkeyToAddress(privKey1.PublicKey)

	privKey2, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key2: %v", err)
	}
	signer2 := crypto.PubkeyToAddress(privKey2.PublicKey)

	// Ensure both signers exist in state.
	statedb.AddBalance(signer1, big.NewInt(0))
	statedb.AddBalance(signer2, big.NewInt(0))

	// Fund the tx sender.
	sender := types.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	tenETH := new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(1e18))
	statedb.AddBalance(sender, tenETH)

	target1 := types.HexToAddress("0x1111111111111111111111111111111111111111")
	target2 := types.HexToAddress("0x2222222222222222222222222222222222222222")

	chainID := config.TestConfig.ChainID
	auth1 := signAuthorization(privKey1, chainID, target1, 0)
	auth2 := signAuthorization(privKey2, chainID, target2, 0)

	to := types.HexToAddress("0xcccccccccccccccccccccccccccccccccccccccc")
	msg := config.Message{
		From:      sender,
		To:        &to,
		Nonce:     0,
		Value:     new(big.Int),
		GasLimit:  200000,
		GasFeeCap: big.NewInt(1_000_000_000),
		GasTipCap: big.NewInt(0),
		Data:      nil,
		AuthList:  []types.Authorization{auth1, auth2},
		TxType:    types.SetCodeTxType,
	}

	header := newTestHeader()
	gp := new(gaspool.GasPool).AddGas(header.GasLimit)

	_, err = execution.ApplyMessage(config.TestConfig, nil, statedb, header, &msg, gp)
	if err != nil {
		t.Fatalf("applyMessage failed: %v", err)
	}

	// Both signers should have delegation code set.
	code1 := statedb.GetCode(signer1)
	if !eips.IsDelegated(code1) {
		t.Errorf("signer1 code should be delegated, got %x", code1)
	}
	resolved1, ok := eips.ResolveDelegation(code1)
	if !ok || resolved1 != target1 {
		t.Errorf("signer1 delegation target: got %v, want %v", resolved1.Hex(), target1.Hex())
	}

	code2 := statedb.GetCode(signer2)
	if !eips.IsDelegated(code2) {
		t.Errorf("signer2 code should be delegated, got %x", code2)
	}
	resolved2, ok := eips.ResolveDelegation(code2)
	if !ok || resolved2 != target2 {
		t.Errorf("signer2 delegation target: got %v, want %v", resolved2.Hex(), target2.Hex())
	}

	// Both nonces should be incremented.
	if statedb.GetNonce(signer1) != 1 {
		t.Errorf("signer1 nonce: got %d, want 1", statedb.GetNonce(signer1))
	}
	if statedb.GetNonce(signer2) != 1 {
		t.Errorf("signer2 nonce: got %d, want 1", statedb.GetNonce(signer2))
	}
}
