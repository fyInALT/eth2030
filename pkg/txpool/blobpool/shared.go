package blobpool

import "github.com/eth2030/eth2030/core/types"

// PriceBump is the minimum gas price bump percentage required for a replacement.
const PriceBump = 10

// StateReader is the minimal interface for reading account state needed by
// blob pool nonce validation.
type StateReader interface {
	GetNonce(addr types.Address) uint64
}
