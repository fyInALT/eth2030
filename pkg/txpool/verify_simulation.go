package txpool

import (
	"fmt"

	"github.com/eth2030/eth2030/core/types"
)

// FrameStateReader extends StateReader with code-size lookup needed for
// VERIFY frame pre-flight simulation (US-AA-3).
type FrameStateReader interface {
	StateReader
	// GetCodeSize returns the byte length of the code at addr.
	// Returns 0 for EOAs and non-existent accounts.
	GetCodeSize(addr types.Address) int
}

// SimulateVerifyFrame performs lightweight pre-execution validation of the
// first VERIFY frame in a FrameTx:
//  1. Confirms at least one VERIFY frame exists.
//  2. Confirms the VERIFY target has deployed code (not an EOA).
//
// When a frame's Target is nil, tx.Sender is used as the target, matching the
// frame execution semantics defined in EIP-8141.
//
// Full EVM simulation (APPROVE detection) requires a running EVM and is
// deferred to block processing in processor.go.
// Returns nil if the structural pre-check passes.
func SimulateVerifyFrame(tx *types.FrameTx, reader FrameStateReader) error {
	if tx == nil {
		return fmt.Errorf("frame tx: nil transaction")
	}
	for _, f := range tx.Frames {
		if f.Mode != types.ModeVerify {
			continue
		}
		target := tx.Sender
		if f.Target != nil {
			target = *f.Target
		}
		if reader.GetCodeSize(target) == 0 {
			return fmt.Errorf("frame tx: VERIFY target %s has no code (EOA)", target.Hex())
		}
		return nil // first VERIFY frame passed
	}
	return fmt.Errorf("frame tx: no VERIFY frame found")
}
