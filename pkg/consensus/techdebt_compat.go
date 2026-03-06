package consensus

// techdebt_compat.go re-exports types from consensus/techdebt for backward compatibility.

import "github.com/eth2030/eth2030/consensus/techdebt"

// TechDebt type aliases.
type (
	DeprecatedField = techdebt.DeprecatedField
	TechDebtConfig  = techdebt.TechDebtConfig
	MigrationReport = techdebt.MigrationReport
	TechDebtTracker = techdebt.TechDebtTracker
)

// TechDebt constants.
const (
	AltairEpoch        = techdebt.AltairEpoch
	Phase1RemovalEpoch = techdebt.Phase1RemovalEpoch
)

// TechDebt error variables.
var (
	ErrTechDebtNilField      = techdebt.ErrTechDebtNilField
	ErrTechDebtEmptyName     = techdebt.ErrTechDebtEmptyName
	ErrTechDebtDuplicate     = techdebt.ErrTechDebtDuplicate
	ErrTechDebtInvalidEpochs = techdebt.ErrTechDebtInvalidEpochs
)

// TechDebt function wrappers.
func DefaultTechDebtConfig() *TechDebtConfig { return techdebt.DefaultTechDebtConfig() }
func NewTechDebtTracker(config *TechDebtConfig) *TechDebtTracker {
	return techdebt.NewTechDebtTracker(config)
}
