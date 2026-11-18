package txpool

import (
	"errors"
	"math/big"

	"github.com/eth2030/eth2030/core/execution"
	"github.com/eth2030/eth2030/core/framepay"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto/bls"
)

func frameFeeBumped(old, next *big.Int) bool {
	if old == nil {
		old = new(big.Int)
	}
	if next == nil {
		return false
	}
	// Round upwards and require an actual increase, including zero-price replacements.
	threshold := new(big.Int).Mul(old, big.NewInt(100+PriceBump))
	threshold.Add(threshold, big.NewInt(99)).Div(threshold, big.NewInt(100))
	return next.Cmp(old) > 0 && next.Cmp(threshold) >= 0
}

// validateFrame8141 executes only the bounded validation prefix on a state copy.
// Reservations are derived from the lookup so every removal releases exposure.
func (pool *TxPool) validateFrame8141(tx *types.Transaction) error {
	frame := tx.Inner().(*types.FrameTx8141)
	if pool.frameConfig == nil || !pool.frameConfig.IsMerge() || !pool.frameConfig.IsHogota(pool.frameTime) {
		return errors.New("EIP-8141 requires Bogotá")
	}
	if _, err := frame.ValidationPrefix8141(pool.frameTime); err != nil {
		return err
	}
	raw, err := tx.EncodeRLP()
	if err != nil {
		return err
	}
	if pool.lookup.Get(tx.Hash()) != tx {
		if err := validateFrameBlobProofs8141(tx); err != nil {
			return err
		}
	}
	if len(raw) > MaxTxSize {
		return ErrOversizedRLP
	}
	db, ok := pool.state.(state.StateDB)
	if !ok {
		return errors.New("Frame validation requires a snapshot-capable StateDB")
	}
	if frame.Nonce != db.GetNonce(frame.Sender) {
		return errors.New("Frame public nonce must match the current sender nonce")
	}
	budget, err := frame.GasBudget()
	if err != nil {
		return err
	}
	if budget.ExecutionReservation > pool.config.BlockGasLimit || budget.StateReservation > pool.config.BlockGasLimit {
		return ErrGasLimit
	}
	blobFee := pool.blobBaseFee
	if blobFee == nil {
		blobFee = big.NewInt(1)
	}
	payment, err := frame.Payment(pool.baseFee, blobFee)
	if err != nil {
		return err
	}
	if pool.config.MinGasPrice != nil && payment.GasPrice().Cmp(pool.config.MinGasPrice) < 0 {
		return ErrUnderpriced
	}
	header := &types.Header{Number: new(big.Int), GasLimit: pool.config.BlockGasLimit, Time: pool.frameTime, BaseFee: pool.baseFee}
	payer, err := execution.ValidateFrame8141Prefix(pool.frameConfig, db, header, tx, blobFee)
	if err != nil {
		return err
	}
	reserved := new(big.Int)
	if frame.Fees.MaxFee != nil {
		reserved.Set(frame.Fees.MaxFee)
	}
	prefixEnd, _ := frame.ValidationPrefix8141(pool.frameTime)
	canonicalPaymaster := frame.Frames[prefixEnd].Flags == types.FrameApprovePayment && framepay.IsV1(db.GetCode(payer))
	if canonicalPaymaster {
		reserved.Add(reserved, framepay.PendingWithdrawal(db, payer))
	}
	nonCanonicalPaymaster := frame.Frames[prefixEnd].Flags == types.FrameApprovePayment && db.GetCodeSize(payer) > 0 && !canonicalPaymaster
	for hash, pending := range pool.lookup.all {
		other, ok := pending.Inner().(*types.FrameTx8141)
		if !ok {
			continue
		}
		if other.Sender == frame.Sender && other.Nonce == frame.Nonce {
			continue
		} // atomic replacement
		if other.Sender == frame.Sender {
			return errors.New("one pending Frame transaction per sender")
		}
		if pool.lookup.framePayers[hash] == payer {
			if other.Fees.MaxFee != nil {
				reserved.Add(reserved, other.Fees.MaxFee)
			}
			if nonCanonicalPaymaster && other.Sender != payer {
				return errors.New("non-canonical Frame paymaster already has a pending transaction")
			}
		}
	}
	if db.GetBalance(payer).Cmp(reserved) < 0 {
		return ErrInsufficientFunds
	}
	tx.SetSender(frame.Sender)
	pool.lookup.framePayers[tx.Hash()] = payer
	return nil
}

func (pool *TxPool) revalidateFrames8141() {
	for hash, tx := range pool.lookup.all {
		if _, ok := tx.Inner().(*types.FrameTx8141); !ok {
			continue
		}
		if err := pool.validateFrame8141(tx); err == nil {
			continue
		}
		from := tx.FrameSender()
		pool.lookup.Remove(hash)
		if list := pool.pending[from]; list != nil {
			list.Remove(tx.Nonce())
			if list.Len() == 0 {
				delete(pool.pending, from)
			}
		}
		if list := pool.queue[from]; list != nil {
			list.Remove(tx.Nonce())
			if list.Len() == 0 {
				delete(pool.queue, from)
			}
		}
	}
}

func validateFrameBlobProofs8141(tx *types.Transaction) error {
	if err := types.ValidateFrameSidecar8141(tx); err != nil {
		return err
	}
	if len(tx.BlobHashes()) == 0 {
		return nil
	}
	return verifyCellProofSidecar(tx.BlobSidecar())
}

func isCanonicalFrame8141(tx *types.Transaction) bool {
	_, ok := tx.Inner().(*types.FrameTx8141)
	return ok
}

func fullBlobProof(blob, commitment []byte) ([]byte, error) {
	if len(commitment) != 48 {
		return nil, errors.New("invalid blob commitment")
	}
	backend, err := bls.NewGoEthKZGBackend()
	if err != nil {
		return nil, err
	}
	proof, err := backend.ComputeBlobProof(blob, [48]byte(commitment))
	if err != nil {
		return nil, err
	}
	return proof[:], nil
}
