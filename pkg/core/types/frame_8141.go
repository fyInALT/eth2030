package types

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/bits"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

// FrameTx8141 uses the wire layout pinned at EIP-8141 revision dc5d5fa6.
type FrameTx8141 struct {
	ChainID             *big.Int
	Nonce               uint64
	Sender              Address
	Frames              []Frame8141
	Signatures          []FrameSignature
	Fees                FrameFeeBudget
	BlobVersionedHashes []Hash
}

type Frame8141 struct {
	Mode   uint8
	Flags  uint8
	Target *Address `rlp:"nilString"`
	Limits FrameGasLimits
	Value  *big.Int
	Data   []byte
}

type FrameGasLimits struct {
	Execution uint64
	State     uint64
}

type FrameSignature struct {
	Scheme    uint8
	Signer    *Address `rlp:"nilString"`
	Message   []byte
	Signature []byte
}

type FrameFeeBudget struct {
	MaxFee               *big.Int
	MaxPriorityFeePerGas *big.Int
}

const (
	FrameApproveExecution uint8 = 1 << iota
	FrameApprovePayment
	FrameAtomicBatch
)

const (
	FrameSignatureArbitrary uint8 = iota
	FrameSignatureSecp256k1
	FrameSignatureP256
)

const (
	frame8141MaxFrames       = 64
	frame8141IntrinsicCost   = 12_000
	frame8141PerFrameCost    = 475
	frame8141ValueCost       = 6_000
	frame8141ExecutionGasCap = 1 << 24
)

var (
	// ErrFrame8141ExecutionUnavailable guards the newer wire format from the historical executor.
	ErrFrame8141ExecutionUnavailable = errors.New("EIP-8141 frame execution unavailable for the seven-field format")
	ErrFrame8141FeeBudget            = errors.New("EIP-8141 frame max fee does not cover the base fees")
)

// FrameGasBudget includes the two independent block reservations and the
// transaction-wide gas bound used for pre-payment.
type FrameGasBudget struct {
	Intrinsic            uint64
	CalldataFloor        uint64
	ExecutionReservation uint64
	StateReservation     uint64
	MaxGas               uint64
}

func (tx *FrameTx8141) txType() byte           { return FrameTxType }
func (tx *FrameTx8141) chainID() *big.Int      { return tx.ChainID }
func (tx *FrameTx8141) accessList() AccessList { return nil }
func (tx *FrameTx8141) data() []byte           { return nil }
func (tx *FrameTx8141) nonce() uint64          { return tx.Nonce }
func (tx *FrameTx8141) value() *big.Int        { return new(big.Int) }
func (tx *FrameTx8141) to() *Address           { return nil }
func (tx *FrameTx8141) gasPrice() *big.Int     { return nil }
func (tx *FrameTx8141) gasFeeCap() *big.Int    { return nil }
func (tx *FrameTx8141) gasTipCap() *big.Int    { return tx.Fees.MaxPriorityFeePerGas }

func (tx *FrameTx8141) gas() uint64 {
	budget, err := tx.GasBudget()
	if err != nil {
		return 0
	}
	return budget.MaxGas
}

func (tx *FrameTx8141) copy() TxData {
	copied := *tx
	copied.ChainID = copyBigInt(tx.ChainID)
	copied.Fees.MaxFee = copyBigInt(tx.Fees.MaxFee)
	copied.Fees.MaxPriorityFeePerGas = copyBigInt(tx.Fees.MaxPriorityFeePerGas)
	copied.BlobVersionedHashes = append([]Hash(nil), tx.BlobVersionedHashes...)
	copied.Frames = append([]Frame8141(nil), tx.Frames...)
	for i := range copied.Frames {
		frame := &copied.Frames[i]
		frame.Target = copyAddressPtr(frame.Target)
		frame.Value = copyBigInt(frame.Value)
		frame.Data = copyBytes(frame.Data)
	}
	copied.Signatures = append([]FrameSignature(nil), tx.Signatures...)
	for i := range copied.Signatures {
		signature := &copied.Signatures[i]
		signature.Signer = copyAddressPtr(signature.Signer)
		signature.Message = copyBytes(signature.Message)
		signature.Signature = copyBytes(signature.Signature)
	}
	return &copied
}

// EncodeFrameTx8141 encodes the canonical EIP-2718 envelope.
func EncodeFrameTx8141(tx *FrameTx8141) ([]byte, error) {
	payload, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}
	return append([]byte{FrameTxType}, payload...), nil
}

// DecodeFrameTx8141 decodes a seven-field RLP payload without its type byte.
// Consensus validation is explicit so readers can preserve original wire bytes
// even when the transaction fails a validity condition.
func DecodeFrameTx8141(payload []byte) (*FrameTx8141, error) {
	tx := new(FrameTx8141)
	if err := rlp.DecodeBytes(payload, tx); err != nil {
		return nil, fmt.Errorf("decode EIP-8141 frame tx: %w", err)
	}
	return tx, nil
}

// SigningHash commits to the signature length for entries using this signing hash.
func (tx *FrameTx8141) SigningHash() (Hash, error) {
	signatures := make([][]any, 0, len(tx.Signatures))
	for _, sig := range tx.Signatures {
		var witness any = sig.Signature
		if len(sig.Message) == 0 {
			witness = uint64(len(sig.Signature))
		}
		signatures = append(signatures, []any{sig.Scheme, addressPtrToBytes(sig.Signer), sig.Message, witness})
	}
	payload, err := rlp.EncodeToBytes([]any{
		bigOrZero(tx.ChainID), tx.Nonce, tx.Sender, tx.Frames, signatures, tx.Fees, tx.BlobVersionedHashes,
	})
	if err != nil {
		return Hash{}, err
	}
	return Hash(gethcrypto.Keccak256Hash([]byte{FrameTxType}, payload)), nil
}

// Validate checks the transaction's state-independent constraints and signatures.
func (tx *FrameTx8141) Validate() error {
	if tx.Nonce == math.MaxUint64 {
		return errors.New("EIP-8141 frame nonce cannot be incremented")
	}
	if !frameUint256(tx.ChainID) || !frameUint256(tx.Fees.MaxFee) || !frameUint256(tx.Fees.MaxPriorityFeePerGas) {
		return errors.New("EIP-8141 frame chain ID and fees must fit uint256")
	}
	if len(tx.Frames) == 0 || len(tx.Frames) > frame8141MaxFrames {
		return fmt.Errorf("EIP-8141 frame count %d outside 1..%d", len(tx.Frames), frame8141MaxFrames)
	}
	for i, h := range tx.BlobVersionedHashes {
		if h[0] != 1 {
			return fmt.Errorf("EIP-8141 frame blob %d has invalid version", i)
		}
	}
	for i, frame := range tx.Frames {
		if err := tx.validateFrame(i, frame); err != nil {
			return fmt.Errorf("EIP-8141 frame %d: %w", i, err)
		}
	}
	if _, err := tx.GasBudget(); err != nil {
		return err
	}
	signingHash, err := tx.SigningHash()
	if err != nil {
		return err
	}
	for i, sig := range tx.Signatures {
		if err := sig.validate(tx.Sender, signingHash); err != nil {
			return fmt.Errorf("EIP-8141 frame signature %d: %w", i, err)
		}
	}
	return nil
}

func frameUint256(value *big.Int) bool {
	return value == nil || (value.Sign() >= 0 && value.BitLen() <= 256)
}

func (tx *FrameTx8141) validateFrame(i int, frame Frame8141) error {
	const approval = FrameApproveExecution | FrameApprovePayment
	if frame.Mode > ModeSender || frame.Flags >= 8 {
		return errors.New("invalid mode or flags")
	}
	if !frameUint256(frame.Value) || (frame.Mode != ModeSender && bigOrZero(frame.Value).Sign() != 0) {
		return errors.New("invalid value")
	}
	if frame.Flags&FrameApproveExecution != 0 && frame.Target != nil && *frame.Target != tx.Sender {
		return errors.New("only the sender can approve execution")
	}
	if frame.Flags&FrameAtomicBatch != 0 {
		if frame.Mode == ModeVerify || i+1 == len(tx.Frames) || tx.Frames[i+1].Mode == ModeVerify {
			return errors.New("atomic batch needs a following non-VERIFY frame")
		}
	}
	if frame.Flags&approval != 0 &&
		(frame.Flags&FrameAtomicBatch != 0 || (i > 0 && tx.Frames[i-1].Flags&FrameAtomicBatch != 0)) {
		return errors.New("approval inside an atomic batch")
	}
	return nil
}

func (sig FrameSignature) validate(sender Address, signingHash Hash) error {
	digest := signingHash
	switch len(sig.Message) {
	case 0:
	case HashLength:
		copy(digest[:], sig.Message)
		if digest == (Hash{}) {
			return errors.New("explicit digest cannot be zero")
		}
	default:
		return errors.New("message must be empty or a 32-byte digest")
	}

	signer := sender
	if sig.Signer != nil {
		signer = *sig.Signer
	}
	switch sig.Scheme {
	case FrameSignatureArbitrary:
		if sig.Signer != nil {
			return errors.New("arbitrary signature has signer metadata")
		}
	case FrameSignatureSecp256k1:
		if len(sig.Signature) != 65 {
			return errors.New("secp256k1 signature must be 65 bytes")
		}
		v := sig.Signature[0]
		r := new(big.Int).SetBytes(sig.Signature[1:33])
		s := new(big.Int).SetBytes(sig.Signature[33:])
		if !gethcrypto.ValidateSignatureValues(v, r, s, true) {
			return errors.New("invalid secp256k1 signature values")
		}
		recovered, err := RecoverPlain(digest, r, s, v)
		if err != nil || recovered != signer {
			return errors.New("secp256k1 signature does not recover the signer")
		}
	case FrameSignatureP256:
		if len(sig.Signature) != 128 {
			return errors.New("P256 signature must be 128 bytes")
		}
		r := new(big.Int).SetBytes(sig.Signature[:32])
		s := new(big.Int).SetBytes(sig.Signature[32:64])
		x := new(big.Int).SetBytes(sig.Signature[64:96])
		y := new(big.Int).SetBytes(sig.Signature[96:])
		curve := elliptic.P256()
		order := curve.Params().N
		if r.Sign() <= 0 || r.Cmp(order) >= 0 || s.Sign() <= 0 || s.Cmp(new(big.Int).Rsh(order, 1)) > 0 ||
			!curve.IsOnCurve(x, y) {
			return errors.New("invalid P256 signature values or public key")
		}
		pub := &ecdsa.PublicKey{Curve: curve, X: x, Y: y}
		if Address(gethcrypto.PubkeyToAddress(*pub)) != signer || !ecdsa.Verify(pub, digest[:], r, s) {
			return errors.New("P256 signature does not verify under the signer")
		}
	default:
		return errors.New("unknown signature scheme")
	}
	return nil
}

// GasBudget implements the pinned intrinsic cost, calldata floor and exact execution/state reservations.
func (tx *FrameTx8141) GasBudget() (FrameGasBudget, error) {
	// The input byte lengths are bounded before multiplication by the execution cap.
	common := uint64(frame8141IntrinsicCost)
	var intrinsicData, bytesForFloor uint64
	var exec, state uint64
	for _, f := range tx.Frames {
		var carry uint64
		exec, carry = bits.Add64(exec, f.Limits.Execution, 0)
		if carry != 0 || exec > frame8141ExecutionGasCap {
			return FrameGasBudget{}, errors.New("EIP-8141 frame execution limit exceeds transaction cap")
		}
		state, carry = bits.Add64(state, f.Limits.State, 0)
		if carry != 0 {
			return FrameGasBudget{}, errors.New("EIP-8141 frame state gas overflows uint64")
		}
		if len(f.Data) > frame8141ExecutionGasCap {
			return FrameGasBudget{}, errors.New("EIP-8141 frame data exceeds transaction cap")
		}
		common += frame8141PerFrameCost
		bytesForFloor += uint64(len(f.Data))
		intrinsicData += CalldataTokenGas(f.Data)
		if bigOrZero(f.Value).Sign() > 0 && f.Target != nil && *f.Target != tx.Sender {
			common += frame8141ValueCost
		}
		if bytesForFloor > frame8141ExecutionGasCap/64 || common > frame8141ExecutionGasCap {
			return FrameGasBudget{}, errors.New("EIP-8141 frame data exceeds transaction cap")
		}
	}
	for _, sig := range tx.Signatures {
		const fullByteGas = CalldataGasPerToken * CalldataTokensPerNonZeroByte
		if len(sig.Signature) > frame8141ExecutionGasCap || len(sig.Message) > frame8141ExecutionGasCap {
			return FrameGasBudget{}, errors.New("EIP-8141 frame signature data exceeds transaction cap")
		}
		switch sig.Scheme {
		case FrameSignatureArbitrary:
			common += 100
		case FrameSignatureSecp256k1:
			common += 2800
		case FrameSignatureP256:
			common += 6700
		default:
			return FrameGasBudget{}, errors.New("EIP-8141 frame unknown signature scheme")
		}
		signer := addressPtrToBytes(sig.Signer)
		bytesForFloor += uint64(len(signer) + len(sig.Message) + len(sig.Signature))
		intrinsicData += CalldataTokenGas(signer) + CalldataTokenGas(sig.Message)
		if len(sig.Message) == 0 {
			intrinsicData += uint64(len(sig.Signature)) * fullByteGas
		} else {
			intrinsicData += CalldataTokenGas(sig.Signature)
		}
		if bytesForFloor > frame8141ExecutionGasCap/64 || common > frame8141ExecutionGasCap {
			return FrameGasBudget{}, errors.New("EIP-8141 frame signature data exceeds transaction cap")
		}
	}

	intrinsic := common + intrinsicData
	floor := common + bytesForFloor*64
	reservation := max(intrinsic+exec, floor)
	if reservation > frame8141ExecutionGasCap {
		return FrameGasBudget{}, errors.New("EIP-8141 frame execution reservation exceeds transaction cap")
	}
	maxGas, carry := bits.Add64(reservation, state, 0)
	if carry != 0 {
		return FrameGasBudget{}, errors.New("EIP-8141 frame maximum gas overflows uint64")
	}
	return FrameGasBudget{
		Intrinsic: intrinsic, CalldataFloor: floor, ExecutionReservation: reservation,
		StateReservation: state, MaxGas: maxGas,
	}, nil
}

func decodeFrameEnvelope(payload []byte) (*Transaction, error) {
	items, rest, err := rlp.SplitList(payload)
	if err != nil {
		return nil, fmt.Errorf("decode frame envelope: %w", err)
	}
	if len(rest) != 0 {
		return nil, errors.New("decode frame envelope: trailing RLP data")
	}
	count, err := rlp.CountValues(items)
	if err != nil {
		return nil, fmt.Errorf("decode frame envelope: %w", err)
	}
	switch count {
	case 7:
		tx, err := DecodeFrameTx8141(payload)
		if err != nil {
			return nil, err
		}
		return NewTransaction(tx), nil
	case 8:
		tx, err := DecodeFrameTx(payload)
		if err != nil {
			return nil, err
		}
		canonical, err := EncodeFrameTx(tx)
		if err != nil || !bytes.Equal(canonical[1:], payload) {
			return nil, errors.New("decode historical frame envelope: non-canonical RLP")
		}
		return NewTransaction(tx), nil
	default:
		return nil, fmt.Errorf("decode frame envelope: unexpected field count %d", count)
	}
}
