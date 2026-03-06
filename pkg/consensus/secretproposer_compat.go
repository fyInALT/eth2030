package consensus

// secretproposer_compat.go re-exports types from consensus/secretproposer for backward compatibility.

import (
	"github.com/eth2030/eth2030/consensus/secretproposer"
	"github.com/eth2030/eth2030/core/types"
)

// SecretProposer type aliases.
type (
	SecretProposerConfig   = secretproposer.SecretProposerConfig
	ProposerCommitment     = secretproposer.ProposerCommitment
	SecretProposerSelector = secretproposer.SecretProposerSelector
)

// SecretProposer error variables.
var (
	ErrSPNoCommitment    = secretproposer.ErrSPNoCommitment
	ErrSPWrongSecret     = secretproposer.ErrSPWrongSecret
	ErrSPAlreadyRevealed = secretproposer.ErrSPAlreadyRevealed
	ErrSPZeroValidators  = secretproposer.ErrSPZeroValidators
)

// SecretProposer function wrappers.
func DefaultSecretProposerConfig() *SecretProposerConfig {
	return secretproposer.DefaultSecretProposerConfig()
}
func NewSecretProposerSelector(config *SecretProposerConfig, seed types.Hash) *SecretProposerSelector {
	return secretproposer.NewSecretProposerSelector(config, seed)
}
func DetermineProposer(slot uint64, validatorCount int, randaoMix types.Hash) uint64 {
	return secretproposer.DetermineProposer(slot, validatorCount, randaoMix)
}
func ValidateCommitReveal(commitment *ProposerCommitment, secret []byte, currentSlot uint64) error {
	return secretproposer.ValidateCommitReveal(commitment, secret, currentSlot)
}
func ValidateSecretProposerConfig(cfg *SecretProposerConfig) error {
	return secretproposer.ValidateSecretProposerConfig(cfg)
}
