package txpool

// peertick_compat.go re-exports types from txpool/peertick for backward compatibility.

import "github.com/eth2030/eth2030/txpool/peertick"

// PeerTickCache type alias.
type PeerTickCache = peertick.PeerTickCache

// PeerTickCache function wrapper.
func NewPeerTickCache(slotTTL uint64) *PeerTickCache { return peertick.NewPeerTickCache(slotTTL) }
