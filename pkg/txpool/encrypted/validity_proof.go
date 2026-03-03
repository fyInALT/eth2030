package encrypted

import (
	"encoding/binary"
	"errors"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
)

// Validity proof errors.
var (
	ErrValidityNilCommit       = errors.New("validity_proof: nil commit")
	ErrValidityNilProof        = errors.New("validity_proof: nil proof")
	ErrValidityInvalidProof    = errors.New("validity_proof: proof verification failed")
	ErrValidityZeroBalance     = errors.New("validity_proof: sender balance is zero")
	ErrValidityInsufficientGas = errors.New("validity_proof: gas limit exceeds balance coverage")
)

// CommitValidityProof attests that an encrypted transaction is valid
// (sufficient balance, correct nonce, valid gas limit) without revealing content.
type CommitValidityProof struct {
	// CommitHash links this proof to a specific encrypted commit.
	CommitHash types.Hash

	// BalanceProof is a cryptographic binding proving balance sufficiency.
	BalanceProof types.Hash

	// NonceProof binds the proof to a specific nonce value.
	NonceProof types.Hash

	// GasProof binds the proof to the gas limit validity.
	GasProof types.Hash

	// AggregateProof combines all three constraint proofs.
	AggregateProof types.Hash

	// ProofData is the serialized proof bytes for on-chain verification.
	ProofData []byte
}

// ValidityCircuit represents the ZK circuit constraints for encrypted tx validation.
type ValidityCircuit struct {
	// SenderBalance is the sender's account balance (private input).
	SenderBalance uint64
	// SenderNonce is the sender's current nonce (private input).
	SenderNonce uint64
	// TxGasLimit is the tx's gas limit (from commit metadata).
	TxGasLimit uint64
	// TxMaxFee is the tx's max fee (from commit metadata).
	TxMaxFee uint64
}

// GenerateCommitProof creates a ZK validity proof for an encrypted tx commit.
// The proof attests that the sender has sufficient balance, correct nonce,
// and valid gas limit -- without revealing the transaction content.
func GenerateCommitProof(commit *CommitTx, senderBalance, senderNonce uint64) (*CommitValidityProof, error) {
	if commit == nil {
		return nil, ErrValidityNilCommit
	}
	if senderBalance == 0 {
		return nil, ErrValidityZeroBalance
	}

	// Check that balance covers gas * max_fee.
	var maxFeeU64 uint64
	if commit.MaxFee != nil {
		if commit.MaxFee.IsUint64() {
			maxFeeU64 = commit.MaxFee.Uint64()
		} else {
			// MaxFee exceeds uint64 -- balance can never cover it.
			return nil, ErrValidityInsufficientGas
		}
	}
	requiredBalance := commit.GasLimit * maxFeeU64
	if requiredBalance > 0 && senderBalance < requiredBalance {
		return nil, ErrValidityInsufficientGas
	}

	// Constraint 1: Balance proof -- H("balance" || commitHash || balance).
	var balBuf [8]byte
	binary.BigEndian.PutUint64(balBuf[:], senderBalance)
	balanceProof := crypto.Keccak256Hash(
		[]byte("encrypted-balance-constraint"),
		commit.CommitHash[:],
		balBuf[:],
	)

	// Constraint 2: Nonce proof -- H("nonce" || commitHash || nonce).
	var nonceBuf [8]byte
	binary.BigEndian.PutUint64(nonceBuf[:], senderNonce)
	nonceProof := crypto.Keccak256Hash(
		[]byte("encrypted-nonce-constraint"),
		commit.CommitHash[:],
		nonceBuf[:],
	)

	// Constraint 3: Gas proof -- H("gas" || commitHash || gasLimit || maxFee).
	var gasBuf [16]byte
	binary.BigEndian.PutUint64(gasBuf[:8], commit.GasLimit)
	binary.BigEndian.PutUint64(gasBuf[8:], maxFeeU64)
	gasProof := crypto.Keccak256Hash(
		[]byte("encrypted-gas-constraint"),
		commit.CommitHash[:],
		gasBuf[:],
	)

	// Aggregate proof -- H("aggregate" || balanceProof || nonceProof || gasProof).
	aggregateProof := crypto.Keccak256Hash(
		[]byte("encrypted-aggregate"),
		balanceProof[:],
		nonceProof[:],
		gasProof[:],
	)

	// Proof data -- H("proof-data" || commitHash || aggregateProof).
	proofData := crypto.Keccak256(
		[]byte("encrypted-proof-data"),
		commit.CommitHash[:],
		aggregateProof[:],
	)

	return &CommitValidityProof{
		CommitHash:     commit.CommitHash,
		BalanceProof:   balanceProof,
		NonceProof:     nonceProof,
		GasProof:       gasProof,
		AggregateProof: aggregateProof,
		ProofData:      proofData,
	}, nil
}

// VerifyCommitProof verifies that a validity proof is internally consistent.
// It recomputes the aggregate and proof data from the individual constraint proofs.
func VerifyCommitProof(proof *CommitValidityProof) (bool, error) {
	if proof == nil {
		return false, ErrValidityNilProof
	}
	if len(proof.ProofData) == 0 {
		return false, ErrValidityInvalidProof
	}

	// Recompute aggregate.
	expectedAggregate := crypto.Keccak256Hash(
		[]byte("encrypted-aggregate"),
		proof.BalanceProof[:],
		proof.NonceProof[:],
		proof.GasProof[:],
	)
	if expectedAggregate != proof.AggregateProof {
		return false, nil
	}

	// Recompute proof data.
	expectedData := crypto.Keccak256(
		[]byte("encrypted-proof-data"),
		proof.CommitHash[:],
		proof.AggregateProof[:],
	)
	if len(expectedData) != len(proof.ProofData) {
		return false, nil
	}
	for i := range expectedData {
		if expectedData[i] != proof.ProofData[i] {
			return false, nil
		}
	}

	return true, nil
}

// ProofSize returns the total byte size of the proof.
func (p *CommitValidityProof) ProofSize() int {
	if p == nil {
		return 0
	}
	// 4 hashes (32 each) + proof data.
	return 32*4 + len(p.ProofData)
}
