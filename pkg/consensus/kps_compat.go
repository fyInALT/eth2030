package consensus

// kps_compat.go re-exports types from consensus/kps for backward compatibility.

import (
	"github.com/eth2030/eth2030/consensus/kps"
	"github.com/eth2030/eth2030/core/types"
)

// KPS type aliases.
type (
	KPSConfig  = kps.KPSConfig
	KeyShare   = kps.KeyShare
	KPSKeyPair = kps.KPSKeyPair
	KeyGroup   = kps.KeyGroup
	KPSManager = kps.KPSManager
)

// KPS error variables.
var (
	ErrKPSInvalidThreshold   = kps.ErrKPSInvalidThreshold
	ErrKPSInvalidShares      = kps.ErrKPSInvalidShares
	ErrKPSInsufficientShares = kps.ErrKPSInsufficientShares
	ErrKPSDuplicateShare     = kps.ErrKPSDuplicateShare
	ErrKPSInvalidShareData   = kps.ErrKPSInvalidShareData
	ErrKPSInvalidPrivateKey  = kps.ErrKPSInvalidPrivateKey
	ErrKPSGroupFull          = kps.ErrKPSGroupFull
	ErrKPSMemberExists       = kps.ErrKPSMemberExists
	ErrKPSMemberNotFound     = kps.ErrKPSMemberNotFound
	ErrKPSGroupNotFound      = kps.ErrKPSGroupNotFound
	ErrKPSKeyGenFailed       = kps.ErrKPSKeyGenFailed
)

// KPS function wrappers.
func DefaultKPSConfig() KPSConfig { return kps.DefaultKPSConfig() }
func NewKeyGroup(groupID types.Hash, threshold, totalMembers int) *KeyGroup {
	return kps.NewKeyGroup(groupID, threshold, totalMembers)
}
func NewKPSManager(config KPSConfig) *KPSManager { return kps.NewKPSManager(config) }
func SplitKey(privateKey []byte, threshold, totalShares int) ([]*KeyShare, error) {
	return kps.SplitKey(privateKey, threshold, totalShares)
}
func RecombineKey(shares []*KeyShare) ([]byte, error) { return kps.RecombineKey(shares) }
func VerifyKeyShare(share *KeyShare, publicKey []byte) bool {
	return kps.VerifyKeyShare(share, publicKey)
}
