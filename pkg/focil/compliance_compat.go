package focil

// compliance_compat.go re-exports types from focil/compliance for backward compatibility.

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/focil/compliance"
)

// Compliance type aliases.
type (
	ComplianceViolationKind  = compliance.ComplianceViolationKind
	ComplianceViolation      = compliance.ComplianceViolation
	ValidatorComplianceState = compliance.ValidatorComplianceState
	ComplianceReport         = compliance.ComplianceReport
	ComplianceTrackerConfig  = compliance.ComplianceTrackerConfig
	ComplianceTracker        = compliance.ComplianceTracker
	SlotComplianceResult     = compliance.SlotComplianceResult
	BuilderInclusionTracker  = compliance.BuilderInclusionTracker
	MEVFilterConfig          = compliance.MEVFilterConfig
	MEVFilter                = compliance.MEVFilter
	BuilderComplianceResult  = compliance.BuilderComplianceResult
)

// Compliance constants.
const (
	ViolationMissedSubmission = compliance.ViolationMissedSubmission
	ViolationLateSubmission   = compliance.ViolationLateSubmission
	ViolationConflicting      = compliance.ViolationConflicting
	ViolationInvalidContent   = compliance.ViolationInvalidContent
)

// Compliance error variables.
var (
	ErrTrackerComplianceSlotZero     = compliance.ErrTrackerComplianceSlotZero
	ErrTrackerComplianceNoValidator  = compliance.ErrTrackerComplianceNoValidator
	ErrTrackerComplianceDuplicate    = compliance.ErrTrackerComplianceDuplicate
	ErrTrackerComplianceRangeInvalid = compliance.ErrTrackerComplianceRangeInvalid
	ErrTrackerComplianceGraceActive  = compliance.ErrTrackerComplianceGraceActive
)

// Compliance function wrappers.
func DefaultComplianceTrackerConfig() ComplianceTrackerConfig {
	return compliance.DefaultComplianceTrackerConfig()
}
func NewComplianceTracker(config ComplianceTrackerConfig) *ComplianceTracker {
	return compliance.NewComplianceTracker(config)
}
func NewBuilderInclusionTracker() *BuilderInclusionTracker {
	return compliance.NewBuilderInclusionTracker()
}
func DefaultMEVFilterConfig() *MEVFilterConfig        { return compliance.DefaultMEVFilterConfig() }
func NewMEVFilter(config *MEVFilterConfig) *MEVFilter { return compliance.NewMEVFilter(config) }
func ValidateBuilderCompliance(blockTxHashes []types.Hash, focilTxHashes []types.Hash, builderTxs []*types.Transaction, mevOnly bool, filter *MEVFilter) *BuilderComplianceResult {
	return compliance.ValidateBuilderCompliance(blockTxHashes, focilTxHashes, builderTxs, mevOnly, filter)
}
