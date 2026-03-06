package consensus

// pq_compat.go provides backward-compatible re-exports from consensus/pq.
// Existing code that imports "github.com/eth2030/eth2030/consensus" and uses
// PQ attestation types continues to work unchanged. New code should import
// "github.com/eth2030/eth2030/consensus/pq" directly.

import (
	"github.com/eth2030/eth2030/consensus/pq"
	"github.com/eth2030/eth2030/core/types"
)

// PQ type aliases.
type (
	PQAttestation             = pq.PQAttestation
	PQAttestationConfig       = pq.PQAttestationConfig
	PQAttestationVerifier     = pq.PQAttestationVerifier
	PQAttestationRecord       = pq.PQAttestationRecord
	PQAggregatorDuty          = pq.PQAggregatorDuty
	PQAggregator              = pq.PQAggregator
	XMSSSignatureBundle       = pq.XMSSSignatureBundle
	STARKSignatureAggregation = pq.STARKSignatureAggregation
	STARKSignatureAggregator  = pq.STARKSignatureAggregator
	PQChainConfig             = pq.PQChainConfig
	PQChainValidator          = pq.PQChainValidator
	PQForkChoice              = pq.PQForkChoice
	PQBlockRecord             = pq.PQBlockRecord
	PQChainAuditResult        = pq.PQChainAuditResult
	PQBlockAuditResult        = pq.PQBlockAuditResult
	PQSecurityLevel           = pq.PQSecurityLevel
	PQChainCommitment         = pq.PQChainCommitment
	PQHistoryAccumulator      = pq.PQHistoryAccumulator
	MerklePQAggregation       = pq.MerklePQAggregation
)

// PQ constants.
const (
	PQSecurityOptional  = pq.PQSecurityOptional
	PQSecurityPreferred = pq.PQSecurityPreferred
	PQSecurityRequired  = pq.PQSecurityRequired
)

// PQ function wrappers.

func DefaultPQAttestationConfig() *PQAttestationConfig { return pq.DefaultPQAttestationConfig() }
func NewPQAttestationVerifier(cfg *PQAttestationConfig) *PQAttestationVerifier {
	return pq.NewPQAttestationVerifier(cfg)
}
func ValidatePQAttestation(att *PQAttestation) error { return pq.ValidatePQAttestation(att) }
func SelectLeanPQAttestors(validators []uint64, count int, slot uint64, epochSeed types.Hash) []uint64 {
	return pq.SelectLeanPQAttestors(validators, count, slot, epochSeed)
}
func SelectAggregators(slot, epoch uint64, epochRandao [32]byte, numValidators, proposerIndex uint64) []PQAggregatorDuty {
	return pq.SelectAggregators(slot, epoch, epochRandao, numValidators, proposerIndex)
}

func NewSTARKSignatureAggregator() *STARKSignatureAggregator {
	return pq.NewSTARKSignatureAggregator()
}
func BatchVerifyPQAttestations(attestations []PQAttestation) (*STARKSignatureAggregation, error) {
	return pq.BatchVerifyPQAttestations(attestations)
}
func MerkleAggregatePQAttestations(attestations []PQAttestation) (*MerklePQAggregation, error) {
	return pq.MerkleAggregatePQAttestations(attestations)
}

func DefaultPQChainConfig() *PQChainConfig { return pq.DefaultPQChainConfig() }
func NewPQChainValidator(cfg *PQChainConfig) *PQChainValidator {
	return pq.NewPQChainValidator(cfg)
}
func NewPQForkChoice(v *PQChainValidator) *PQForkChoice { return pq.NewPQForkChoice(v) }
func ValidatePQTransition(cur, target PQSecurityLevel, epoch uint64, cfg *PQChainConfig) error {
	return pq.ValidatePQTransition(cur, target, epoch, cfg)
}
func ValidatePQChainConfig(cfg *PQChainConfig) error { return pq.ValidatePQChainConfig(cfg) }
func IntegratePQForkChoice(fc pq.ForkChoice, pqFC *PQForkChoice) int {
	return pq.IntegratePQForkChoice(fc, pqFC)
}
func NewPQChainCommitment(epoch uint64, hashes []types.Hash) *PQChainCommitment {
	return pq.NewPQChainCommitment(epoch, hashes)
}
func NewPQHistoryAccumulator() *PQHistoryAccumulator       { return pq.NewPQHistoryAccumulator() }
func PQBlockHash(header *types.Header) (types.Hash, error) { return pq.PQBlockHash(header) }
