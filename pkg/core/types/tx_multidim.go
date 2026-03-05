package types

import (
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/rlp"
	"golang.org/x/crypto/sha3"
)

// MultiDimFeeTxType is the EIP-2718 envelope type for EIP-7706 transactions.
// Uses 3-element fee vectors [execution, blob, calldata] in wei per unit.
const MultiDimFeeTxType byte = 0x09

// MultiDimFeeTx is an EIP-7706 transaction with 3-dimensional fee vectors.
// MaxFeesPerGas and PriorityFeesPerGas are [execution, blob, calldata] tuples.
type MultiDimFeeTx struct {
	ChainID    *big.Int
	Nonce      uint64
	GasLimit   uint64
	To         *Address // nil means contract creation
	Value      *big.Int
	Data       []byte
	AccessList AccessList

	// 3-element fee vectors: indices are [execution=0, blob=1, calldata=2].
	MaxFeesPerGas      [3]*big.Int
	PriorityFeesPerGas [3]*big.Int

	// Signature fields.
	V, R, S *big.Int
}

// TxData interface implementation.
func (tx *MultiDimFeeTx) txType() byte           { return MultiDimFeeTxType }
func (tx *MultiDimFeeTx) chainID() *big.Int      { return tx.ChainID }
func (tx *MultiDimFeeTx) nonce() uint64          { return tx.Nonce }
func (tx *MultiDimFeeTx) gas() uint64            { return tx.GasLimit }
func (tx *MultiDimFeeTx) to() *Address           { return tx.To }
func (tx *MultiDimFeeTx) value() *big.Int        { return tx.Value }
func (tx *MultiDimFeeTx) data() []byte           { return tx.Data }
func (tx *MultiDimFeeTx) accessList() AccessList { return tx.AccessList }

// gasPrice returns the execution-dimension max fee (index 0).
func (tx *MultiDimFeeTx) gasPrice() *big.Int {
	if tx.MaxFeesPerGas[0] == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(tx.MaxFeesPerGas[0])
}

// gasTipCap returns the execution-dimension priority fee (index 0).
func (tx *MultiDimFeeTx) gasTipCap() *big.Int {
	if tx.PriorityFeesPerGas[0] == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(tx.PriorityFeesPerGas[0])
}

// gasFeeCap returns the execution-dimension max fee (same as gasPrice).
func (tx *MultiDimFeeTx) gasFeeCap() *big.Int { return tx.gasPrice() }

func (tx *MultiDimFeeTx) copy() TxData {
	cpy := &MultiDimFeeTx{
		ChainID:    copyBigInt(tx.ChainID),
		Nonce:      tx.Nonce,
		GasLimit:   tx.GasLimit,
		To:         copyAddressPtr(tx.To),
		Value:      copyBigInt(tx.Value),
		Data:       copyBytes(tx.Data),
		AccessList: copyAccessList(tx.AccessList),
		V:          copyBigInt(tx.V),
		R:          copyBigInt(tx.R),
		S:          copyBigInt(tx.S),
	}
	for i := 0; i < 3; i++ {
		cpy.MaxFeesPerGas[i] = copyBigInt(tx.MaxFeesPerGas[i])
		cpy.PriorityFeesPerGas[i] = copyBigInt(tx.PriorityFeesPerGas[i])
	}
	return cpy
}

func (tx *MultiDimFeeTx) signingHash(chainID *big.Int) Hash {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte{MultiDimFeeTxType})
	fees := make([]*big.Int, 3)
	prio := make([]*big.Int, 3)
	for i := 0; i < 3; i++ {
		fees[i] = tx.MaxFeesPerGas[i]
		prio[i] = tx.PriorityFeesPerGas[i]
	}
	rlp.Encode(hasher, []interface{}{
		chainID,
		tx.Nonce,
		tx.GasLimit,
		addressPtrToBytes(tx.To),
		tx.Value,
		tx.Data,
		tx.AccessList,
		fees,
		prio,
	})
	var h Hash
	hasher.Sum(h[:0])
	return h
}

func (tx *MultiDimFeeTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

// EncodeRLP encodes the transaction as type-prefixed RLP bytes.
func (tx *MultiDimFeeTx) EncodeRLP() ([]byte, error) {
	fees := make([]*big.Int, 3)
	prio := make([]*big.Int, 3)
	for i := 0; i < 3; i++ {
		fees[i] = tx.MaxFeesPerGas[i]
		prio[i] = tx.PriorityFeesPerGas[i]
	}
	inner, err := rlp.EncodeToBytes(multiDimFeeTxRLP{
		ChainID:            tx.ChainID,
		Nonce:              tx.Nonce,
		GasLimit:           tx.GasLimit,
		To:                 addressPtrToBytes(tx.To),
		Value:              tx.Value,
		Data:               tx.Data,
		AccessList:         tx.AccessList,
		MaxFeesPerGas:      fees,
		PriorityFeesPerGas: prio,
		V:                  tx.V,
		R:                  tx.R,
		S:                  tx.S,
	})
	if err != nil {
		return nil, err
	}
	result := make([]byte, 1+len(inner))
	result[0] = MultiDimFeeTxType
	copy(result[1:], inner)
	return result, nil
}

type multiDimFeeTxRLP struct {
	ChainID            *big.Int
	Nonce              uint64
	GasLimit           uint64
	To                 []byte
	Value              *big.Int
	Data               []byte
	AccessList         AccessList
	MaxFeesPerGas      []*big.Int
	PriorityFeesPerGas []*big.Int
	V, R, S            *big.Int
}

// DecodeMultiDimFeeTx decodes a MultiDimFeeTx from RLP payload (without the type prefix byte).
func DecodeMultiDimFeeTx(payload []byte) (*Transaction, error) {
	var dec multiDimFeeTxRLP
	if err := rlp.DecodeBytes(payload, &dec); err != nil {
		return nil, err
	}
	if len(dec.MaxFeesPerGas) != 3 || len(dec.PriorityFeesPerGas) != 3 {
		return nil, errorf("multidim fee tx: fee vectors must be length 3, got %d/%d",
			len(dec.MaxFeesPerGas), len(dec.PriorityFeesPerGas))
	}
	inner := &MultiDimFeeTx{
		ChainID:    dec.ChainID,
		Nonce:      dec.Nonce,
		GasLimit:   dec.GasLimit,
		To:         bytesToAddressPtr(dec.To),
		Value:      dec.Value,
		Data:       dec.Data,
		AccessList: dec.AccessList,
		V:          dec.V,
		R:          dec.R,
		S:          dec.S,
	}
	copy(inner.MaxFeesPerGas[:], dec.MaxFeesPerGas)
	copy(inner.PriorityFeesPerGas[:], dec.PriorityFeesPerGas)
	return NewTransaction(inner), nil
}

// helpers reused from transaction_rlp.go (same package).
func copyBigInt(b *big.Int) *big.Int {
	if b == nil {
		return nil
	}
	return new(big.Int).Set(b)
}

func errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
