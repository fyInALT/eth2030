package consensus

// includer_compat.go re-exports types from consensus/includer for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/consensus/includer"
	"github.com/eth2030/eth2030/core/types"
)

// Includer type aliases.
type (
	IncluderStatus = includer.IncluderStatus
	IncluderRecord = includer.IncluderRecord
	IncluderDuty   = includer.IncluderDuty
	IncluderPool   = includer.IncluderPool
)

// Includer constants/vars.
var OneETH = includer.OneETH

// Includer status constants.
const (
	IncluderActive  = includer.IncluderActive
	IncluderSlashed = includer.IncluderSlashed
	IncluderExited  = includer.IncluderExited
)

// Includer reward constants.
const (
	BaseIncluderReward  = includer.BaseIncluderReward
	IncluderRewardDecay = includer.IncluderRewardDecay
)

// Includer error aliases.
var (
	ErrIncluderZeroAddress       = includer.ErrIncluderZeroAddress
	ErrIncluderWrongStake        = includer.ErrIncluderWrongStake
	ErrIncluderAlreadyRegistered = includer.ErrIncluderAlreadyRegistered
	ErrIncluderNotRegistered     = includer.ErrIncluderNotRegistered
	ErrIncluderPoolEmpty         = includer.ErrIncluderPoolEmpty
	ErrIncluderAlreadySlashed    = includer.ErrIncluderAlreadySlashed
	ErrIncluderNilDuty           = includer.ErrIncluderNilDuty
	ErrIncluderInvalidSig        = includer.ErrIncluderInvalidSig
)

// Includer function wrappers.
func NewIncluderPool() *IncluderPool { return includer.NewIncluderPool() }
func VerifyIncluderSignature(duty *IncluderDuty, sig []byte) bool {
	return includer.VerifyIncluderSignature(duty, sig)
}
func IncluderReward(slot includer.Slot) uint64 { return includer.IncluderReward(slot) }

// RegisterIncluder is a convenience wrapper.
func RegisterIncluder(pool *IncluderPool, addr types.Address, stake *big.Int) error {
	return pool.RegisterIncluder(addr, stake)
}
