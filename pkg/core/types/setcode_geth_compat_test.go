package types

import (
	"math/big"
	"testing"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
)

// TestSetCodeTxGethCompat verifies that a SetCode transaction encoded by
// go-ethereum (as spamoor does via wallet.BuildSetCodeTx) can be decoded by
// our DecodeTxRLP.
func TestSetCodeTxGethCompat(t *testing.T) {
	// Generate two keys: one for the tx sender, one for the authorization signer.
	senderKey, err := gethcrypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate sender key: %v", err)
	}
	authKey, err := gethcrypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate auth key: %v", err)
	}

	chainID := uint256.NewInt(3151908) // typical devnet chain ID
	chainIDBig := chainID.ToBig()

	// Build authorization (like spamoor's buildSetCodeAuthorizations).
	codeAddr := gethcrypto.PubkeyToAddress(authKey.PublicKey)
	auth := gethtypes.SetCodeAuthorization{
		ChainID: *chainID,
		Address: codeAddr,
		Nonce:   0,
	}
	signedAuth, err := gethtypes.SignSetCode(authKey, auth)
	if err != nil {
		t.Fatalf("sign auth: %v", err)
	}

	// Build SetCode tx (like spamoor's BuildSetCodeTx).
	to := gethcrypto.PubkeyToAddress(senderKey.PublicKey)
	inner := &gethtypes.SetCodeTx{
		ChainID:   chainID,
		Nonce:     0,
		GasTipCap: uint256.NewInt(1_000_000_000),
		GasFeeCap: uint256.NewInt(50_000_000_000),
		Gas:       100000,
		To:        to,
		Value:     uint256.NewInt(0),
		Data:      nil,
		AuthList:  []gethtypes.SetCodeAuthorization{signedAuth},
	}
	gethTx := gethtypes.NewTx(inner)

	// Sign the tx with go-ethereum's signer (as spamoor does).
	signer := gethtypes.LatestSignerForChainID(chainIDBig)
	signedTx, err := gethtypes.SignTx(gethTx, signer, senderKey)
	if err != nil {
		t.Fatalf("sign tx: %v", err)
	}

	// Encode with go-ethereum's MarshalBinary (type_byte || RLP_payload).
	txBytes, err := signedTx.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	if len(txBytes) == 0 || txBytes[0] != SetCodeTxType {
		t.Fatalf("expected type byte 0x%02x, got 0x%02x", SetCodeTxType, txBytes[0])
	}

	// Decode with our DecodeTxRLP.
	decoded, err := DecodeTxRLP(txBytes)
	if err != nil {
		t.Fatalf("DecodeTxRLP failed: %v", err)
	}

	if decoded.Type() != SetCodeTxType {
		t.Fatalf("expected SetCode type, got %d", decoded.Type())
	}

	inner2, ok := decoded.inner.(*SetCodeTx)
	if !ok {
		t.Fatal("decoded tx is not SetCodeTx")
	}

	// Verify fields.
	if inner2.Nonce != 0 {
		t.Errorf("nonce: got %d, want 0", inner2.Nonce)
	}
	if inner2.Gas != 100000 {
		t.Errorf("gas: got %d, want 100000", inner2.Gas)
	}
	if inner2.ChainID.Cmp(chainIDBig) != 0 {
		t.Errorf("chainID: got %v, want %v", inner2.ChainID, chainIDBig)
	}
	if len(inner2.AuthorizationList) != 1 {
		t.Fatalf("authList len: got %d, want 1", len(inner2.AuthorizationList))
	}

	// Verify V,R,S are non-nil and consistent with the original.
	origV, origR, origS := signedTx.RawSignatureValues()
	if inner2.V == nil || inner2.R == nil || inner2.S == nil {
		t.Fatal("decoded V/R/S is nil")
	}
	if inner2.V.Cmp(origV) != 0 {
		t.Errorf("V mismatch: got %v, want %v", inner2.V, origV)
	}
	if inner2.R.Cmp(origR) != 0 {
		t.Errorf("R mismatch: got %v, want %v", inner2.R, origR)
	}
	if inner2.S.Cmp(origS) != 0 {
		t.Errorf("S mismatch: got %v, want %v", inner2.S, origS)
	}

	// Verify auth list fields.
	a := inner2.AuthorizationList[0]
	if a.Address != Address(signedAuth.Address) {
		t.Errorf("auth address: got %v, want %v", a.Address, signedAuth.Address)
	}
	wantAuthV := big.NewInt(int64(signedAuth.V))
	if a.V.Cmp(wantAuthV) != 0 {
		t.Errorf("auth V: got %v, want %v", a.V, wantAuthV)
	}
	if a.R.Cmp(signedAuth.R.ToBig()) != 0 {
		t.Errorf("auth R mismatch")
	}
	if a.S.Cmp(signedAuth.S.ToBig()) != 0 {
		t.Errorf("auth S mismatch")
	}

	// Verify the signing hash matches go-ethereum's signing hash.
	// If these differ, sender recovery will fail in our txpool.senderOf().
	ourHash := decoded.SigningHash()
	gethHash := signer.Hash(signedTx)
	if ourHash != Hash(gethHash) {
		t.Errorf("signing hash mismatch: ours=%x geth=%x", ourHash, gethHash)
	}

	// Simulate txpool.senderOf() to ensure sender recovery works.
	// Use gethcrypto.Ecrecover since we can't import our crypto package (cycle).
	sigHash := decoded.SigningHash()
	v, r, s := decoded.RawSignatureValues()
	if v == nil || r == nil || s == nil {
		t.Fatal("V/R/S nil after decode")
	}
	sig := make([]byte, 65)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)
	sig[64] = byte(v.Uint64()) // typed tx: V is 0 or 1 directly
	pub, err := gethcrypto.SigToPub(sigHash[:], sig)
	if err != nil {
		t.Fatalf("SigToPub failed: %v", err)
	}
	recoveredAddr := gethcrypto.PubkeyToAddress(*pub)
	expectedSender := gethcrypto.PubkeyToAddress(senderKey.PublicKey)
	if Address(recoveredAddr) != Address(expectedSender) {
		t.Errorf("sender mismatch: recovered=%v want=%v", recoveredAddr, expectedSender)
	}

	t.Logf("geth-encoded SetCode tx decoded OK: type=%d nonce=%d gas=%d authListLen=%d sender=%v",
		decoded.Type(), inner2.Nonce, inner2.Gas, len(inner2.AuthorizationList), recoveredAddr)
}
