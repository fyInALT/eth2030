package consensus

// recovery_compat.go re-exports types from consensus/recovery for backward compatibility.

import "github.com/eth2030/eth2030/consensus/recovery"

// Attack recovery type aliases.
type (
	AttackReport   = recovery.AttackReport
	RecoveryPlan   = recovery.RecoveryPlan
	RecoveryStatus = recovery.RecoveryStatus
	AttackDetector = recovery.AttackDetector
)

// Attack recovery constants.
const (
	SeverityNone     = recovery.SeverityNone
	SeverityLow      = recovery.SeverityLow
	SeverityMedium   = recovery.SeverityMedium
	SeverityHigh     = recovery.SeverityHigh
	SeverityCritical = recovery.SeverityCritical

	ReorgThresholdLow      = recovery.ReorgThresholdLow
	ReorgThresholdMedium   = recovery.ReorgThresholdMedium
	ReorgThresholdHigh     = recovery.ReorgThresholdHigh
	ReorgThresholdCritical = recovery.ReorgThresholdCritical

	ActionNone           = recovery.ActionNone
	ActionMonitor        = recovery.ActionMonitor
	ActionIsolate        = recovery.ActionIsolate
	ActionFallback       = recovery.ActionFallback
	ActionSocialOverride = recovery.ActionSocialOverride
)

// Attack recovery error variables.
var (
	ErrNoAttackDetected  = recovery.ErrNoAttackDetected
	ErrNilRecoveryPlan   = recovery.ErrNilRecoveryPlan
	ErrAlreadyRecovering = recovery.ErrAlreadyRecovering
	ErrInvalidPlan       = recovery.ErrInvalidPlan
)

// Attack recovery function wrappers.
func NewAttackDetector() *AttackDetector {
	return recovery.NewAttackDetector()
}
func SeverityLevel(reorgDepth uint64) string {
	return recovery.SeverityLevel(reorgDepth)
}
func BuildRecoveryPlan(report *AttackReport) (*RecoveryPlan, error) {
	return recovery.BuildRecoveryPlan(report)
}
func ValidateRecoveryPlan(plan *RecoveryPlan) error {
	return recovery.ValidateRecoveryPlan(plan)
}
func ValidateAttackReport(report *AttackReport) error {
	return recovery.ValidateAttackReport(report)
}
