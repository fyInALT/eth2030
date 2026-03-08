package consensus

// slashdetect_compat.go re-exports types from consensus/slashdetect for backward compatibility.

import "github.com/eth2030/eth2030/consensus/slashdetect"

// Slashing detector type aliases.
type (
	BlockRecord              = slashdetect.BlockRecord
	AttestationRecord        = slashdetect.AttestationRecord
	ProposerSlashingEvidence = slashdetect.ProposerSlashingEvidence
	AttesterSlashingEvidence = slashdetect.AttesterSlashingEvidence
	SlashingDetectorConfig   = slashdetect.SlashingDetectorConfig
	SlashingDetector         = slashdetect.SlashingDetector
)

// Slashing detector constant aliases.
const DefaultAttestationWindow = slashdetect.DefaultAttestationWindow

// Slashing detector function wrappers.
func DefaultSlashingDetectorConfig() SlashingDetectorConfig {
	return slashdetect.DefaultSlashingDetectorConfig()
}
func NewSlashingDetector(config SlashingDetectorConfig) *SlashingDetector {
	return slashdetect.NewSlashingDetector(config)
}
