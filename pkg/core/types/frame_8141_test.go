package types

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math"
	"math/big"
	"testing"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Generated from the specified field lists with a standalone Python RLP encoder.
const pinnedFrame8141 = "06f8a68330182401941111111111111111111111111111111111111111f2cf010380c7820bb88302cd3080820002e10280942222222222222222222222222222222222222222c48203e8320183000102f0c88080808400010200e68080a003030303030303030303030303030303030303030303030303030303030303038200fec5830f424004e1a00100000000000000000000000000000000000000000000000000000000000000"
const pinnedFrame8141Receipt = "06f8878301e240942222222222222222222222222222222222222222f86cf85c01c682012c820190f852f838941111111111111111111111111111111111111111e1a0abababababababababababababababababababababababababababababababab01d7942222222222222222222222222222222222222222c002c502c28080c0c780c48202ee80c0"

func frame8141Hex(t *testing.T, input string) []byte {
	t.Helper()
	data, err := hex.DecodeString(input)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestFrame8141EnvelopeAndSigningVector(t *testing.T) {
	wire := frame8141Hex(t, pinnedFrame8141)
	transaction, err := DecodeTxRLP(wire)
	if err != nil {
		t.Fatal(err)
	}
	tx, ok := transaction.Inner().(*FrameTx8141)
	if !ok {
		t.Fatalf("wrong inner type %T", transaction.Inner())
	}
	if tx.Nonce != 1 || tx.Frames[0].Target != nil || tx.Frames[0].Limits.State != 183600 ||
		tx.Frames[1].Value.Uint64() != 1 || tx.Fees.MaxFee.Uint64() != 1_000_000 || transaction.BlobGas() != 131072 {
		t.Fatalf("wire fields decoded incorrectly: %+v", tx)
	}
	if err := tx.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := transaction.Hash(); got != HexToHash("0xe97781f331360b4b5a132e7110be18e2e6eff6a00fe21224812394fc6e017969") {
		t.Fatalf("transaction hash %s", got)
	}
	if got := transaction.SigningHash(); got != HexToHash("0x4dc384781acec437af205a2d5fb7e717b04bc8efee0c2f7ef22d015180f50792") {
		t.Fatalf("signing hash %s", got)
	}
	encoded, err := transaction.EncodeRLP()
	if err != nil || !bytes.Equal(wire, encoded) {
		t.Fatalf("round trip %x, %v", encoded, err)
	}
	if _, err := DecodeTxRLP(append(wire, 0x80)); err == nil {
		t.Fatal("accepted trailing RLP")
	}
	original := transaction.SigningHash()
	tx.Signatures[0].Signature[0] ^= 1
	if transaction.SigningHash() != original {
		t.Fatal("canonical-message witness content changed the signing hash")
	}
	tx.Signatures[0].Signature = append(tx.Signatures[0].Signature, 0)
	if transaction.SigningHash() == original {
		t.Fatal("canonical-message witness length was omitted")
	}
	tx.Signatures[0].Signature = tx.Signatures[0].Signature[:4]
	tx.Signatures[1].Signature[0] ^= 1
	if transaction.SigningHash() == original {
		t.Fatal("explicit-message witness was omitted")
	}
}

func TestFrame8141ProtocolSignatures(t *testing.T) {
	key, err := gethcrypto.HexToECDSA("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	tx := &FrameTx8141{
		Sender:     Address(gethcrypto.PubkeyToAddress(key.PublicKey)),
		Frames:     []Frame8141{{Mode: ModeVerify, Flags: FrameApproveExecution | FrameApprovePayment, Limits: FrameGasLimits{Execution: 3000}}},
		Signatures: []FrameSignature{{Scheme: FrameSignatureSecp256k1, Signature: make([]byte, 65)}},
	}
	digest, err := tx.SigningHash()
	if err != nil {
		t.Fatal(err)
	}
	sig, err := gethcrypto.Sign(digest[:], key)
	if err != nil {
		t.Fatal(err)
	}
	tx.Signatures[0].Signature = append([]byte{sig[64]}, sig[:64]...)
	if err := tx.Validate(); err != nil {
		t.Fatal(err)
	}
	tx.Nonce++
	if err := tx.Validate(); err == nil {
		t.Fatal("a changed nonce reused a canonical-message signature")
	}

	p256, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	m := bytes.Repeat([]byte{0x45}, 32)
	r, s, err := ecdsa.Sign(rand.Reader, p256, m)
	if err != nil {
		t.Fatal(err)
	}
	order := p256.Params().N
	if s.Cmp(new(big.Int).Rsh(order, 1)) > 0 {
		s.Sub(order, s)
	}
	signer := Address(gethcrypto.PubkeyToAddress(p256.PublicKey))
	p256sig := make([]byte, 128)
	r.FillBytes(p256sig[:32])
	s.FillBytes(p256sig[32:64])
	p256.X.FillBytes(p256sig[64:96])
	p256.Y.FillBytes(p256sig[96:])
	tx.Signatures[0] = FrameSignature{Scheme: FrameSignatureP256, Signer: &signer, Message: m, Signature: p256sig}
	if err := tx.Validate(); err != nil {
		t.Fatal(err)
	}
	high := new(big.Int).Sub(order, s)
	high.FillBytes(p256sig[32:64])
	if err := tx.Validate(); err == nil {
		t.Fatal("accepted high-s P256 signature")
	}
}

func TestFrame8141BudgetAndFeeSettlement(t *testing.T) {
	tx, err := DecodeFrameTx8141(frame8141Hex(t, pinnedFrame8141)[1:])
	if err != nil {
		t.Fatal(err)
	}
	budget, err := tx.GasBudget()
	want := FrameGasBudget{Intrinsic: 19802, CalldataFloor: 21902, ExecutionReservation: 23802, StateReservation: 183650, MaxGas: 207452}
	if err != nil || budget != want {
		t.Fatalf("gas budget %+v, want %+v, error %v", budget, want, err)
	}

	payment, err := tx.Payment(big.NewInt(2), big.NewInt(3))
	if err != nil {
		t.Fatal(err)
	}
	if payment.Escrow().Uint64() != 1_000_000 || payment.GasPrice().Uint64() != 2 {
		t.Fatalf("escrow %s or effective price %s", payment.Escrow(), payment.GasPrice())
	}
	receipt, err := payment.Settle(100_000)
	if err != nil || receipt.Burned.Uint64() != 593216 || receipt.Tip.Uint64() != 400000 || receipt.Refund.Uint64() != 6784 {
		t.Fatalf("settlement %+v, %v", receipt, err)
	}
	top, err := payment.Settle(budget.MaxGas)
	if err != nil || top.Burned.Uint64() != 808120 || top.Tip.Uint64() != 191880 || top.Refund.Sign() != 0 {
		t.Fatalf("budget-limited settlement %+v, %v", top, err)
	}
	if _, err := payment.Settle(budget.MaxGas + 1); err == nil {
		t.Fatal("settled above reservation")
	}
	tx.Fees.MaxFee = big.NewInt(808119)
	if _, err := tx.Payment(big.NewInt(2), big.NewInt(3)); !errors.Is(err, ErrFrame8141FeeBudget) {
		t.Fatalf("expected one-wei-short fee budget rejection, got %v", err)
	}
}

func TestFrame8141StaticRejections(t *testing.T) {
	alter := map[string]func(*FrameTx8141){
		"nonce overflow":               func(tx *FrameTx8141) { tx.Nonce = math.MaxUint64 },
		"state reservation overflow":   func(tx *FrameTx8141) { tx.Frames[0].Limits.State = math.MaxUint64 },
		"execution cap":                func(tx *FrameTx8141) { tx.Frames[1].Limits.Execution = 1 << 24 },
		"blob version":                 func(tx *FrameTx8141) { tx.BlobVersionedHashes[0][0] = 2 },
		"approval from another target": func(tx *FrameTx8141) { tx.Frames[0].Target = tx.Frames[1].Target },
		"reserved flag":                func(tx *FrameTx8141) { tx.Frames[0].Flags = 8 },
		"unclosed batch":               func(tx *FrameTx8141) { tx.Frames[1].Flags = FrameAtomicBatch },
		"zero explicit digest":         func(tx *FrameTx8141) { clear(tx.Signatures[1].Message) },
		"arbitrary signer metadata":    func(tx *FrameTx8141) { tx.Signatures[0].Signer = &tx.Sender },
	}
	wire := frame8141Hex(t, pinnedFrame8141)
	for label, mutate := range alter {
		t.Run(label, func(t *testing.T) {
			tx, err := DecodeFrameTx8141(wire[1:])
			if err != nil {
				t.Fatal(err)
			}
			mutate(tx)
			if err := tx.Validate(); err == nil {
				t.Fatal("invalid transaction passed static validation")
			}
		})
	}
}

func TestFrame8141ReceiptVector(t *testing.T) {
	wire := frame8141Hex(t, pinnedFrame8141Receipt)
	r, err := DecodeReceiptRLP(wire)
	if err != nil {
		t.Fatal(err)
	}
	if r.CumulativeGasUsed != 123456 || r.FrameResults8141[0].GasUsed != (FrameGasLimits{Execution: 300, State: 400}) ||
		r.FrameResults8141[1].Status != FrameStatusSkipped || len(r.Logs) != 2 ||
		r.Logs[0].Data[0] != 1 || r.Logs[1].Data[0] != 2 || r.Bloom != LogsBloom(r.Logs) {
		t.Fatalf("wrong receipt projection: %+v", r)
	}
	encoded, err := r.EncodeRLP()
	if err != nil || !bytes.Equal(wire, encoded) {
		t.Fatalf("wrong receipt round trip: %x, %v", encoded, err)
	}
	r.FrameResults8141[1].GasUsed.State = 1
	if _, err := r.EncodeRLP(); err == nil {
		t.Fatal("skipped frame accepted a state charge")
	}
}
