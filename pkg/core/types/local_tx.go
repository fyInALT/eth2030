package types

import (
	"errors"
	"fmt"
	"math/big"
)

// LocalTxDiscountBPS is the default gas discount for fully-declared BAL local
// transactions, expressed in basis points (5000 = 50%).
// Configurable per-node via the proof-of-concept flag --experimental-local-tx.
const LocalTxDiscountBPS uint64 = 5000

// ErrBALViolation is returned when a LocalTx accesses state outside its declared ScopeHint.
var ErrBALViolation = errors.New("local_tx: state access outside declared scope")

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

// GasWithDiscount returns the effective gas limit after applying discountBPS
// (basis points, e.g. 5000 = 50%). A discount of 0 returns the original gas;
// a discount >= 10000 (100%) returns 0.
func (tx *LocalTx) GasWithDiscount(discountBPS uint64) uint64 {
	if discountBPS == 0 {
		return tx.Gas_
	}
	if discountBPS >= 10000 {
		return 0
	}
	return tx.Gas_ * (10000 - discountBPS) / 10000
}

// ApplyLocalTxDiscount returns the discounted gas limit for a LocalTx using
// LocalTxDiscountBPS. For non-LocalTx transactions the original gas is returned.
func ApplyLocalTxDiscount(tx *Transaction) uint64 {
	if local, ok := tx.inner.(*LocalTx); ok {
		return local.GasWithDiscount(LocalTxDiscountBPS)
	}
	return tx.Gas()
}

// ValidateScopeAccess checks that every accessed address falls within the
// declared ScopeHint prefixes. An empty or nil ScopeHint means global access
// (no validation performed). Returns ErrBALViolation on mismatch.
//
// This is a proof-of-concept enforcement for --experimental-local-tx; full
// BAL enforcement belongs at the EVM execution layer.
func (tx *LocalTx) ValidateScopeAccess(accessedAddresses []Address) error {
	if len(tx.ScopeHint) == 0 {
		return nil // global scope — no restriction
	}
	prefixSet := make(map[byte]struct{}, len(tx.ScopeHint))
	for _, p := range tx.ScopeHint {
		prefixSet[p] = struct{}{}
	}
	for _, addr := range accessedAddresses {
		if _, ok := prefixSet[addr[0]]; !ok {
			return fmt.Errorf("%w: address prefix 0x%02x not in declared scope %x",
				ErrBALViolation, addr[0], tx.ScopeHint)
		}
	}
	return nil
}

// ValidateLocalTxScopeAccess is a package-level helper that calls ValidateScopeAccess
// on a Transaction if it is a LocalTx. Non-LocalTx transactions always pass.
func ValidateLocalTxScopeAccess(tx *Transaction, accessedAddresses []Address) error {
	if tx == nil {
		return nil
	}
	local, ok := tx.inner.(*LocalTx)
	if !ok {
		return nil
	}
	return local.ValidateScopeAccess(accessedAddresses)
}
