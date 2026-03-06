package frametx

import (
	"errors"
	"fmt"

	"github.com/eth2030/eth2030/core/types"
)

// PaymasterApprover checks whether an address is an approved (staked) paymaster.
// Implemented by core.PaymasterRegistry; defined as interface here to avoid import cycles.
type PaymasterApprover interface {
	IsApprovedPaymaster(addr types.Address) bool
}

// AggressiveFrameRules validates FrameTx with the permissive aggressive ruleset:
//   - First frame must be ModeVerify (same as conservative)
//   - VERIFY frame gas limit <= AggressiveVerifyGasLimit (200K) if any frame's target
//     is a registered staked paymaster
//   - Falls back to ConservativeFrameRules when no staked paymaster is detected
type AggressiveFrameRules struct {
	Registry PaymasterApprover // may be nil (disables paymaster check -> use conservative)
}

// Validate checks tx under aggressive rules.
func (r AggressiveFrameRules) Validate(tx *types.FrameTx) error {
	if tx == nil {
		return errors.New("frame tx: nil transaction")
	}
	if len(tx.Frames) == 0 {
		return errors.New("frame tx: no frames")
	}
	if tx.Frames[0].Mode != types.ModeVerify {
		return ErrNoVerifyFirst
	}

	// Detect if there is a frame targeting a staked paymaster.
	stakedPaymaster := r.detectStakedPaymaster(tx)

	for i, f := range tx.Frames {
		if f.GasLimit == 0 {
			return &FrameRuleError{FrameIndex: i, Reason: "zero gas limit"}
		}
		if f.Mode != types.ModeVerify {
			continue
		}
		cap := ConservativeVerifyGasLimit
		if stakedPaymaster {
			cap = AggressiveVerifyGasLimit
		}
		if f.GasLimit > cap {
			return &FrameRuleError{
				FrameIndex: i,
				Reason: fmt.Sprintf(
					"VERIFY frame gas limit %d exceeds cap %d (staked paymaster: %v)",
					f.GasLimit, cap, stakedPaymaster,
				),
			}
		}
	}
	return nil
}

// detectStakedPaymaster returns true if any frame's target is a staked paymaster.
func (r AggressiveFrameRules) detectStakedPaymaster(tx *types.FrameTx) bool {
	if r.Registry == nil {
		return false
	}
	for _, f := range tx.Frames {
		if f.Target == nil {
			continue
		}
		if r.Registry.IsApprovedPaymaster(*f.Target) {
			return true
		}
	}
	return false
}

// ValidateFrameTxAggressive validates tx under aggressive rules with the given registry.
func ValidateFrameTxAggressive(tx *types.FrameTx, registry PaymasterApprover) error {
	return AggressiveFrameRules{Registry: registry}.Validate(tx)
}
