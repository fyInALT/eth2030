package das

// pqblob_compat.go re-exports types from das/pqblob for backward compatibility.

import (
	"github.com/eth2030/eth2030/crypto/pqc"
	"github.com/eth2030/eth2030/das/pqblob"
)

// PQ blob type aliases.
type (
	PQBlobCommitment           = pqblob.PQBlobCommitment
	PQBlobProof                = pqblob.PQBlobProof
	PQBlobSignature            = pqblob.PQBlobSignature
	PQBlobSigner               = pqblob.PQBlobSigner
	PQBlobIntegritySig         = pqblob.PQBlobIntegritySig
	PQBlobIntegritySigner      = pqblob.PQBlobIntegritySigner
	MLDSABlobIntegritySigner   = pqblob.MLDSABlobIntegritySigner
	FalconBlobIntegritySigner  = pqblob.FalconBlobIntegritySigner
	SPHINCSBlobIntegritySigner = pqblob.SPHINCSBlobIntegritySigner
	BatchBlobIntegrityVerifier = pqblob.BatchBlobIntegrityVerifier
	PQBlobIntegrityReport      = pqblob.PQBlobIntegrityReport
	PQBlobProofV2              = pqblob.PQBlobProofV2
	PQBlobValidator            = pqblob.PQBlobValidator
)

// PQ blob constants.
const (
	PQCommitmentSize    = pqblob.PQCommitmentSize
	PQProofSize         = pqblob.PQProofSize
	LatticeDimension    = pqblob.LatticeDimension
	LatticeModulus      = pqblob.LatticeModulus
	ChunkSize           = pqblob.ChunkSize
	MaxBlobSize         = pqblob.MaxBlobSize
	PQBlobAlgFalcon     = pqblob.PQBlobAlgFalcon
	PQBlobAlgSPHINCS    = pqblob.PQBlobAlgSPHINCS
	PQBlobAlgMLDSA      = pqblob.PQBlobAlgMLDSA
	PQAlgDilithium      = pqblob.PQAlgDilithium
	PQAlgFalcon         = pqblob.PQAlgFalcon
	PQAlgSPHINCS        = pqblob.PQAlgSPHINCS
	IntegrityAlgMLDSA   = pqblob.IntegrityAlgMLDSA
	IntegrityAlgFalcon  = pqblob.IntegrityAlgFalcon
	IntegrityAlgSPHINCS = pqblob.IntegrityAlgSPHINCS
)

// PQ blob error variables.
var (
	ErrPQBlobEmpty             = pqblob.ErrPQBlobEmpty
	ErrPQBlobTooLarge          = pqblob.ErrPQBlobTooLarge
	ErrPQBlobNilCommitment     = pqblob.ErrPQBlobNilCommitment
	ErrPQBlobNilProof          = pqblob.ErrPQBlobNilProof
	ErrPQBlobIndexOOB          = pqblob.ErrPQBlobIndexOOB
	ErrPQBlobMismatch          = pqblob.ErrPQBlobMismatch
	ErrPQBlobSignNilSigner     = pqblob.ErrPQBlobSignNilSigner
	ErrPQBlobSignNilKey        = pqblob.ErrPQBlobSignNilKey
	ErrPQBlobSignEmptyCommit   = pqblob.ErrPQBlobSignEmptyCommit
	ErrPQBlobSignBadSig        = pqblob.ErrPQBlobSignBadSig
	ErrPQBlobSignBadPK         = pqblob.ErrPQBlobSignBadPK
	ErrPQBlobSignUnknownAlg    = pqblob.ErrPQBlobSignUnknownAlg
	ErrPQBlobSignBatchEmpty    = pqblob.ErrPQBlobSignBatchEmpty
	ErrPQBlobSignVerifyFailed  = pqblob.ErrPQBlobSignVerifyFailed
	ErrPQValidatorUnknownAlg   = pqblob.ErrPQValidatorUnknownAlg
	ErrPQValidatorNilBlob      = pqblob.ErrPQValidatorNilBlob
	ErrPQValidatorEmptyBlob    = pqblob.ErrPQValidatorEmptyBlob
	ErrPQValidatorNilCommit    = pqblob.ErrPQValidatorNilCommit
	ErrPQValidatorNilProof     = pqblob.ErrPQValidatorNilProof
	ErrPQValidatorMismatch     = pqblob.ErrPQValidatorMismatch
	ErrPQValidatorProofInvalid = pqblob.ErrPQValidatorProofInvalid
	ErrPQValidatorBatchEmpty   = pqblob.ErrPQValidatorBatchEmpty
	ErrPQValidatorBatchLen     = pqblob.ErrPQValidatorBatchLen
	ErrPQValidatorBlobTooLarge = pqblob.ErrPQValidatorBlobTooLarge
)

// PQ blob function wrappers.
func CommitBlob(data []byte) (*PQBlobCommitment, error) { return pqblob.CommitBlob(data) }
func VerifyBlobCommitment(commitment *PQBlobCommitment, data []byte) bool {
	return pqblob.VerifyBlobCommitment(commitment, data)
}
func GenerateBlobProof(data []byte, index uint32) (*PQBlobProof, error) {
	return pqblob.GenerateBlobProof(data, index)
}
func VerifyBlobProof(proof *PQBlobProof, commitment *PQBlobCommitment) bool {
	return pqblob.VerifyBlobProof(proof, commitment)
}
func BatchVerifyProofs(proofs []*PQBlobProof, commitments []*PQBlobCommitment) bool {
	return pqblob.BatchVerifyProofs(proofs, commitments)
}
func ValidatePQBlob(commitment *PQBlobCommitment) error  { return pqblob.ValidatePQBlob(commitment) }
func ValidatePQBlobProof(proof *PQBlobProof) error       { return pqblob.ValidatePQBlobProof(proof) }
func NewPQBlobSigner(algID uint8) (*PQBlobSigner, error) { return pqblob.NewPQBlobSigner(algID) }
func NewPQBlobSignerWithKey(algID uint8, kp *pqc.PQKeyPair) (*PQBlobSigner, error) {
	return pqblob.NewPQBlobSignerWithKey(algID, kp)
}
func SignBlobCommitment(commitment [PQCommitmentSize]byte, signer *PQBlobSigner) (PQBlobSignature, error) {
	return pqblob.SignBlobCommitment(commitment, signer)
}
func VerifyBlobSignature(commitment [PQCommitmentSize]byte, sig PQBlobSignature) bool {
	return pqblob.VerifyBlobSignature(commitment, sig)
}
func BatchVerifyBlobSignatures(commitments [][PQCommitmentSize]byte, sigs []PQBlobSignature) (int, error) {
	return pqblob.BatchVerifyBlobSignatures(commitments, sigs)
}
func PQBlobSignatureSize(algID uint8) int              { return pqblob.PQBlobSignatureSize(algID) }
func PQBlobPublicKeySize(algID uint8) int              { return pqblob.PQBlobPublicKeySize(algID) }
func EncodePQBlobSignature(sig PQBlobSignature) []byte { return pqblob.EncodePQBlobSignature(sig) }
func DecodePQBlobSignature(data []byte) (PQBlobSignature, error) {
	return pqblob.DecodePQBlobSignature(data)
}
func PQBlobSignerAlgorithmName(algID uint8) string { return pqblob.PQBlobSignerAlgorithmName(algID) }
func NewMLDSABlobIntegritySigner() (*MLDSABlobIntegritySigner, error) {
	return pqblob.NewMLDSABlobIntegritySigner()
}
func NewFalconBlobIntegritySigner() (*FalconBlobIntegritySigner, error) {
	return pqblob.NewFalconBlobIntegritySigner()
}
func NewFalconBlobIntegritySignerWithKey(kp *pqc.PQKeyPair) (*FalconBlobIntegritySigner, error) {
	return pqblob.NewFalconBlobIntegritySignerWithKey(kp)
}
func NewSPHINCSBlobIntegritySigner() (*SPHINCSBlobIntegritySigner, error) {
	return pqblob.NewSPHINCSBlobIntegritySigner()
}
func NewBatchBlobIntegrityVerifier(workers int) *BatchBlobIntegrityVerifier {
	return pqblob.NewBatchBlobIntegrityVerifier(workers)
}
func NewPQBlobIntegrityReport() *PQBlobIntegrityReport { return pqblob.NewPQBlobIntegrityReport() }
func CommitAndSignBlob(data []byte, signer PQBlobIntegritySigner) (*PQBlobCommitment, *PQBlobIntegritySig, error) {
	return pqblob.CommitAndSignBlob(data, signer)
}
func IntegrityAlgorithmName(algID uint8) string         { return pqblob.IntegrityAlgorithmName(algID) }
func IntegritySignatureSize(algID uint8) int            { return pqblob.IntegritySignatureSize(algID) }
func EncodeIntegritySig(sig *PQBlobIntegritySig) []byte { return pqblob.EncodeIntegritySig(sig) }
func NewPQBlobValidator(algorithm string) *PQBlobValidator {
	return pqblob.NewPQBlobValidator(algorithm)
}
func EstimateValidationGas(algorithm string, blobSize int) uint64 {
	return pqblob.EstimateValidationGas(algorithm, blobSize)
}
func SupportedPQAlgorithms() []string { return pqblob.SupportedPQAlgorithms() }
