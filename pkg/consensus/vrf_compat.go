package consensus

// vrf_compat.go re-exports types from consensus/vrf for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/consensus/vrf"
	"github.com/eth2030/eth2030/core/types"
)

// VRF type aliases.
type (
	VRFKeyPair          = vrf.VRFKeyPair
	VRFProof            = vrf.VRFProof
	VRFOutput           = vrf.VRFOutput
	VRFElectionEntry    = vrf.VRFElectionEntry
	VRFReveal           = vrf.VRFReveal
	VRFSlashingEvidence = vrf.VRFSlashingEvidence
	SecretElection      = vrf.SecretElection
)

// VRF constants.
const (
	VRFKeySize       = vrf.VRFKeySize
	VRFProofSize     = vrf.VRFProofSize
	VRFOutputSize    = vrf.VRFOutputSize
	MaxVRFValidators = vrf.MaxVRFValidators
)

// VRF error variables.
var (
	ErrVRFNilKey          = vrf.ErrVRFNilKey
	ErrVRFInvalidProof    = vrf.ErrVRFInvalidProof
	ErrVRFInvalidOutput   = vrf.ErrVRFInvalidOutput
	ErrVRFNoValidators    = vrf.ErrVRFNoValidators
	ErrVRFDoubleReveal    = vrf.ErrVRFDoubleReveal
	ErrVRFNoReveal        = vrf.ErrVRFNoReveal
	ErrVRFSlotMismatch    = vrf.ErrVRFSlotMismatch
	ErrVRFAlreadyRevealed = vrf.ErrVRFAlreadyRevealed
)

// VRF function wrappers.
func GenerateVRFKeyPair(seed []byte) *VRFKeyPair { return vrf.GenerateVRFKeyPair(seed) }
func VRFProve(sk [VRFKeySize]byte, input []byte) (VRFOutput, VRFProof) {
	return vrf.VRFProve(sk, input)
}
func VRFVerify(pk [VRFKeySize]byte, input []byte, output VRFOutput, proof VRFProof) bool {
	return vrf.VRFVerify(pk, input, output, proof)
}
func ComputeVRFElectionInput(epoch, slot uint64) []byte {
	return vrf.ComputeVRFElectionInput(epoch, slot)
}
func ComputeProposerScore(output VRFOutput) *big.Int  { return vrf.ComputeProposerScore(output) }
func NewSecretElection() *SecretElection              { return vrf.NewSecretElection() }
func VerifyReveal(pk [VRFKeySize]byte, reveal *VRFReveal, epoch uint64) bool {
	return vrf.VerifyReveal(pk, reveal, epoch)
}
func BlockBindingHash(output VRFOutput, blockHash types.Hash) types.Hash {
	return vrf.BlockBindingHash(output, blockHash)
}
