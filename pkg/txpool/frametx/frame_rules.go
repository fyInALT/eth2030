package frametx

import (
	"errors"
	"fmt"

	"github.com/eth2030/eth2030/core/types"
)

// Conservative frame validation limits.
const (
	ConservativeVerifyGasLimit uint64 = 50_000
	AggressiveVerifyGasLimit   uint64 = 200_000
)

// ErrNoVerifyFirst is returned when the first frame is not ModeVerify.
var ErrNoVerifyFirst = errors.New("frame tx conservative: first frame must be VERIFY mode")

// FrameRuleError describes which frame violated a rule.
type FrameRuleError struct {
	FrameIndex int
	Reason     string
}

func (e *FrameRuleError) Error() string {
	return fmt.Sprintf("frame tx: frame %d: %s", e.FrameIndex, e.Reason)
}

// ConservativeFrameRules validates FrameTx with the strict conservative ruleset:
//   - First frame must be ModeVerify
//   - Each VERIFY frame gas limit <= ConservativeVerifyGasLimit (50K)
//   - No frame may have zero gas limit
type ConservativeFrameRules struct{}

// Validate checks tx under conservative rules.
func (ConservativeFrameRules) Validate(tx *types.FrameTx) error {
	if tx == nil {
		return errors.New("frame tx: nil transaction")
	}
	if len(tx.Frames) == 0 {
		return errors.New("frame tx: no frames")
	}
	if tx.Frames[0].Mode != types.ModeVerify {
		return ErrNoVerifyFirst
	}
	for i, f := range tx.Frames {
		if f.GasLimit == 0 {
			return &FrameRuleError{FrameIndex: i, Reason: "zero gas limit"}
		}
		if f.Mode == types.ModeVerify && f.GasLimit > ConservativeVerifyGasLimit {
			return &FrameRuleError{
				FrameIndex: i,
				Reason: fmt.Sprintf(
					"VERIFY frame gas limit %d exceeds conservative cap %d",
					f.GasLimit, ConservativeVerifyGasLimit,
				),
			}
		}
	}
	return nil
}

// ValidateFrameTxConservative validates tx under conservative rules.
func ValidateFrameTxConservative(tx *types.FrameTx) error {
	return ConservativeFrameRules{}.Validate(tx)
}
