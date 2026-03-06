package light

// bls_compat.go re-exports types from light/bls for backward compatibility.

import lightbls "github.com/eth2030/eth2030/light/bls"

// BLS type aliases.
type SyncCommitteeBLSVerifier = lightbls.SyncCommitteeBLSVerifier

// BLS error variables.
var (
	ErrBLSInvalidPubkey    = lightbls.ErrBLSInvalidPubkey
	ErrBLSInvalidSignature = lightbls.ErrBLSInvalidSignature
	ErrBLSVerifyFailed     = lightbls.ErrBLSVerifyFailed
	ErrBLSNoParticipants   = lightbls.ErrBLSNoParticipants
)

// BLS constants.
const (
	MinQuorumNumerator   = lightbls.MinQuorumNumerator
	MinQuorumDenominator = lightbls.MinQuorumDenominator
)

// BLS function wrappers.
func NewSyncCommitteeBLSVerifier() *SyncCommitteeBLSVerifier {
	return lightbls.NewSyncCommitteeBLSVerifier()
}
func NewSyncCommitteeBLSVerifierWithSize(size int) *SyncCommitteeBLSVerifier {
	return lightbls.NewSyncCommitteeBLSVerifierWithSize(size)
}
func CountParticipants(participationBits []byte, committeeSize int) int {
	return lightbls.CountParticipants(participationBits, committeeSize)
}
func MakeParticipationBits(committeeSize, participants int) []byte {
	return lightbls.MakeParticipationBits(committeeSize, participants)
}
func MakeBLSTestCommittee(size int) ([][48]byte, []*[32]byte) {
	return lightbls.MakeBLSTestCommittee(size)
}
func SignSyncCommitteeBLS(secrets []*[32]byte, participationBits []byte, msg []byte) [96]byte {
	return lightbls.SignSyncCommitteeBLS(secrets, participationBits, msg)
}
