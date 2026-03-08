package consensus

// fastconfirm_compat.go re-exports types from consensus/fastconfirm for backward compatibility.

import "github.com/eth2030/eth2030/consensus/fastconfirm"

// Fast confirmation type aliases.
type (
	FastConfirmConfig  = fastconfirm.FastConfirmConfig
	FastConfirmation   = fastconfirm.FastConfirmation
	FastConfirmTracker = fastconfirm.FastConfirmTracker
)

// Fast confirmation error aliases.
var (
	ErrFCSlotZero          = fastconfirm.ErrFCSlotZero
	ErrFCBlockRootEmpty    = fastconfirm.ErrFCBlockRootEmpty
	ErrFCDuplicateAttester = fastconfirm.ErrFCDuplicateAttester
	ErrFCSlotExpired       = fastconfirm.ErrFCSlotExpired
	ErrFCNotFound          = fastconfirm.ErrFCNotFound
)

// Fast confirmation function wrappers.
func DefaultFastConfirmConfig() *FastConfirmConfig { return fastconfirm.DefaultFastConfirmConfig() }
func NewFastConfirmTracker(cfg *FastConfirmConfig) *FastConfirmTracker {
	return fastconfirm.NewFastConfirmTracker(cfg)
}
func ValidateFastConfirmConfig(cfg *FastConfirmConfig) error {
	return fastconfirm.ValidateFastConfirmConfig(cfg)
}
func ValidateConfirmation(fc *FastConfirmation, cfg *FastConfirmConfig) error {
	return fastconfirm.ValidateConfirmation(fc, cfg)
}
