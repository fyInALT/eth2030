package types

import (
	"math/big"
)

// LocalTxType is the transaction type for local (scope-hinted) transactions.
const LocalTxType = 0x08

// LocalTx is a transaction with a ScopeHint indicating which address
// prefixes it accesses. Transactions with non-overlapping scopes can
// execute in parallel without conflict.
type LocalTx struct {
	ChainID_   *big.Int
	Nonce_     uint64
	GasTipCap_ *big.Int
	GasFeeCap_ *big.Int
	Gas_       uint64
	To_        *Address
	Value_     *big.Int
	Data_      []byte

	// ScopeHint is a list of 1-byte address prefixes indicating which
	// portion of the state this tx accesses. Non-overlapping scope hints
	// mean two LocalTxs can execute in parallel.
	ScopeHint []byte
}

func (tx *LocalTx) txType() byte           { return LocalTxType }
func (tx *LocalTx) chainID() *big.Int      { return tx.ChainID_ }
func (tx *LocalTx) accessList() AccessList { return nil }
func (tx *LocalTx) data() []byte           { return tx.Data_ }
func (tx *LocalTx) gas() uint64            { return tx.Gas_ }
func (tx *LocalTx) gasPrice() *big.Int     { return tx.GasFeeCap_ }
func (tx *LocalTx) gasTipCap() *big.Int    { return tx.GasTipCap_ }
func (tx *LocalTx) gasFeeCap() *big.Int    { return tx.GasFeeCap_ }
func (tx *LocalTx) value() *big.Int        { return tx.Value_ }
func (tx *LocalTx) nonce() uint64          { return tx.Nonce_ }

func (tx *LocalTx) to() *Address {
	return tx.To_
}

func (tx *LocalTx) copy() TxData {
	cpy := &LocalTx{
		Nonce_:    tx.Nonce_,
		Gas_:      tx.Gas_,
		ScopeHint: make([]byte, len(tx.ScopeHint)),
	}
	if tx.ChainID_ != nil {
		cpy.ChainID_ = new(big.Int).Set(tx.ChainID_)
	}
	if tx.GasTipCap_ != nil {
		cpy.GasTipCap_ = new(big.Int).Set(tx.GasTipCap_)
	}
	if tx.GasFeeCap_ != nil {
		cpy.GasFeeCap_ = new(big.Int).Set(tx.GasFeeCap_)
	}
	if tx.Value_ != nil {
		cpy.Value_ = new(big.Int).Set(tx.Value_)
	}
	if tx.To_ != nil {
		to := *tx.To_
		cpy.To_ = &to
	}
	if tx.Data_ != nil {
		cpy.Data_ = make([]byte, len(tx.Data_))
		copy(cpy.Data_, tx.Data_)
	}
	copy(cpy.ScopeHint, tx.ScopeHint)
	return cpy
}

// NewLocalTx creates a new local transaction wrapped in a Transaction.
func NewLocalTx(chainID *big.Int, nonce uint64, to *Address, value *big.Int,
	gasLimit uint64, gasTipCap, gasFeeCap *big.Int, data []byte, scopeHint []byte) *Transaction {
	inner := &LocalTx{
		ChainID_:   chainID,
		Nonce_:     nonce,
		GasTipCap_: gasTipCap,
		GasFeeCap_: gasFeeCap,
		Gas_:       gasLimit,
		To_:        to,
		Value_:     value,
		Data_:      data,
		ScopeHint:  scopeHint,
	}
	return NewTransaction(inner)
}

// ScopesOverlap returns true if two LocalTxs have overlapping scope hints.
// An empty scope hint is treated as "global" and overlaps with everything.
func ScopesOverlap(a, b *LocalTx) bool {
	if a == nil || b == nil {
		return true // nil = global scope
	}
	if len(a.ScopeHint) == 0 || len(b.ScopeHint) == 0 {
		return true // empty = global scope
	}
	for _, sa := range a.ScopeHint {
		for _, sb := range b.ScopeHint {
			if sa == sb {
				return true
			}
		}
	}
	return false
}

// IsLocalTx returns true if a Transaction is of type LocalTx.
func IsLocalTx(tx *Transaction) bool {
	return tx != nil && tx.Type() == LocalTxType
}

// GetScopeHint returns the scope hint from a LocalTx, or nil for other types.
func GetScopeHint(tx *Transaction) []byte {
	if tx == nil {
		return nil
	}
	if local, ok := tx.inner.(*LocalTx); ok {
		return local.ScopeHint
	}
	return nil
}
