package das

// custody_compat.go re-exports all public types, functions, and error vars
// from das/custody for backward compatibility.

import "github.com/eth2030/eth2030/das/custody"

// --- custody_manager.go ---

// CustodyManagerConfig type alias.
type CustodyManagerConfig = custody.CustodyManagerConfig

// DefaultCustodyManagerConfig returns a default CustodyManagerConfig.
var DefaultCustodyManagerConfig = custody.DefaultCustodyManagerConfig

// CustodyEpochState type alias.
type CustodyEpochState = custody.CustodyEpochState

// SlotCompleteness type alias.
type SlotCompleteness = custody.SlotCompleteness

// CustodyRotationEvent type alias.
type CustodyRotationEvent = custody.CustodyRotationEvent

// CustodyProofRequest type alias.
type CustodyProofRequest = custody.CustodyProofRequest

// CustodyProofResult type alias.
type CustodyProofResult = custody.CustodyProofResult

// CustodyManager type alias.
type CustodyManager = custody.CustodyManager

// NewCustodyManager creates a new CustodyManager.
var NewCustodyManager = custody.NewCustodyManager

// Custody manager errors.
var (
	ErrCustodyMgrClosed         = custody.ErrCustodyMgrClosed
	ErrCustodyMgrNotInitialized = custody.ErrCustodyMgrNotInitialized
	ErrCustodyMgrEpochZero      = custody.ErrCustodyMgrEpochZero
	ErrCustodyMgrColumnOOB      = custody.ErrCustodyMgrColumnOOB
	ErrCustodyMgrSlotOOB        = custody.ErrCustodyMgrSlotOOB
	ErrCustodyMgrIncomplete     = custody.ErrCustodyMgrIncomplete
	ErrCustodyMgrProofInvalid   = custody.ErrCustodyMgrProofInvalid
	ErrCustodyMgrAlreadyStored  = custody.ErrCustodyMgrAlreadyStored
	ErrCustodyMgrRotationBusy   = custody.ErrCustodyMgrRotationBusy
)

// --- custody_verify.go ---

// CustodyVerifyConfig type alias.
type CustodyVerifyConfig = custody.CustodyVerifyConfig

// DefaultCustodyVerifyConfig returns a default CustodyVerifyConfig.
var DefaultCustodyVerifyConfig = custody.DefaultCustodyVerifyConfig

// CustodyProofV2 type alias.
type CustodyProofV2 = custody.CustodyProofV2

// CustodyChallengeV2 type alias.
type CustodyChallengeV2 = custody.CustodyChallengeV2

// CustodyResponse type alias.
type CustodyResponse = custody.CustodyResponse

// PenaltyCalculator type alias.
type PenaltyCalculator = custody.PenaltyCalculator

// NewPenaltyCalculator creates a new PenaltyCalculator.
var NewPenaltyCalculator = custody.NewPenaltyCalculator

// CustodyVerifier type alias.
type CustodyVerifier = custody.CustodyVerifier

// NewCustodyVerifier creates a new CustodyVerifier.
var NewCustodyVerifier = custody.NewCustodyVerifier

// MakeResponseProof creates a response proof.
var MakeResponseProof = custody.MakeResponseProof

// Custody verifier errors.
var (
	ErrNilCustodyProofV2       = custody.ErrNilCustodyProofV2
	ErrEmptyCustodyData        = custody.ErrEmptyCustodyData
	ErrEmptyCommitment         = custody.ErrEmptyCommitment
	ErrCellIndexOutOfRange     = custody.ErrCellIndexOutOfRange
	ErrBlobIndexOutOfRange     = custody.ErrBlobIndexOutOfRange
	ErrSubnetOutOfRange        = custody.ErrSubnetOutOfRange
	ErrMerklePathInvalid       = custody.ErrMerklePathInvalid
	ErrChallengeWindowExceeded = custody.ErrChallengeWindowExceeded
	ErrNoRequiredCells         = custody.ErrNoRequiredCells
	ErrResponseCountMismatch   = custody.ErrResponseCountMismatch
	ErrResponseChallengeID     = custody.ErrResponseChallengeID
	ErrResponseCellNotRequired = custody.ErrResponseCellNotRequired
	ErrResponseEmptyData       = custody.ErrResponseEmptyData
	ErrResponseEmptyProof      = custody.ErrResponseEmptyProof
)

// --- custody_proof.go ---

// DefaultEpochCutoff is the default number of epochs before a proof is too old.
const DefaultEpochCutoff = custody.DefaultEpochCutoff

// CustodyProof type alias.
type CustodyProof = custody.CustodyProof

// CustodyChallenge type alias.
type CustodyChallenge = custody.CustodyChallenge

// GenerateCustodyProof generates a new custody proof.
var GenerateCustodyProof = custody.GenerateCustodyProof

// VerifyCustodyProof verifies a custody proof.
var VerifyCustodyProof = custody.VerifyCustodyProof

// VerifyCustodyProofWithData verifies a custody proof against data.
var VerifyCustodyProofWithData = custody.VerifyCustodyProofWithData

// CreateChallenge creates a new custody challenge.
var CreateChallenge = custody.CreateChallenge

// RespondToChallenge responds to a custody challenge.
var RespondToChallenge = custody.RespondToChallenge

// VerifyCustodyProofWithEpoch verifies a custody proof with epoch bounds.
var VerifyCustodyProofWithEpoch = custody.VerifyCustodyProofWithEpoch

// ValidateChallengeDeadline validates a challenge deadline.
var ValidateChallengeDeadline = custody.ValidateChallengeDeadline

// ValidateCustodyChallenge validates a custody challenge.
var ValidateCustodyChallenge = custody.ValidateCustodyChallenge

// CustodyProofTracker type alias.
type CustodyProofTracker = custody.CustodyProofTracker

// NewCustodyProofTracker creates a new CustodyProofTracker.
var NewCustodyProofTracker = custody.NewCustodyProofTracker

// Custody proof errors.
var (
	ErrInvalidCustodyProof = custody.ErrInvalidCustodyProof
	ErrChallengeExpired    = custody.ErrChallengeExpired
	ErrChallengeNotFound   = custody.ErrChallengeNotFound
	ErrInvalidColumn       = custody.ErrInvalidColumn
	ErrProofEpochTooOld    = custody.ErrProofEpochTooOld
	ErrDeadlinePassed      = custody.ErrDeadlinePassed
	ErrProofReplay         = custody.ErrProofReplay
)

// --- custody_subnet.go ---

// CustodyConfig type alias.
type CustodyConfig = custody.CustodyConfig

// DefaultCustodyConfig returns a default CustodyConfig.
var DefaultCustodyConfig = custody.DefaultCustodyConfig

// SubnetAssignment type alias.
type SubnetAssignment = custody.SubnetAssignment

// PeerInfo type alias.
type PeerInfo = custody.PeerInfo

// CustodySubnetManager type alias.
type CustodySubnetManager = custody.CustodySubnetManager

// NewCustodySubnetManager creates a new CustodySubnetManager.
var NewCustodySubnetManager = custody.NewCustodySubnetManager

// Custody subnet errors.
var (
	ErrCustodyGroupCountExceeded = custody.ErrCustodyGroupCountExceeded
	ErrMissingCustodyColumn      = custody.ErrMissingCustodyColumn
	ErrColumnOutOfRange          = custody.ErrColumnOutOfRange
	ErrNoPeersForColumn          = custody.ErrNoPeersForColumn
)

// --- column_custody.go ---

// CustodyManagerParams type alias.
type CustodyManagerParams = custody.CustodyManagerParams

// DefaultCustodyManagerParams returns default CustodyManagerParams.
var DefaultCustodyManagerParams = custody.DefaultCustodyManagerParams

// CustodyAssignment type alias.
type CustodyAssignment = custody.CustodyAssignment

// StoredColumn type alias.
type StoredColumn = custody.StoredColumn

// CustodyRotation type alias.
type CustodyRotation = custody.CustodyRotation

// CustodyProofResponse type alias.
type CustodyProofResponse = custody.CustodyProofResponse

// NetworkSamplingRequest type alias.
type NetworkSamplingRequest = custody.NetworkSamplingRequest

// NetworkSamplingResult type alias.
type NetworkSamplingResult = custody.NetworkSamplingResult

// ColumnCustodyManager type alias.
type ColumnCustodyManager = custody.ColumnCustodyManager

// NewColumnCustodyManager creates a new ColumnCustodyManager.
var NewColumnCustodyManager = custody.NewColumnCustodyManager

// Column custody errors.
var (
	ErrCustodyManagerClosed   = custody.ErrCustodyManagerClosed
	ErrColumnNotInCustody     = custody.ErrColumnNotInCustody
	ErrColumnExpired          = custody.ErrColumnExpired
	ErrColumnAlreadyStored    = custody.ErrColumnAlreadyStored
	ErrSamplingNoPeers        = custody.ErrSamplingNoPeers
	ErrSamplingTimeout        = custody.ErrSamplingTimeout
	ErrInvalidEpoch           = custody.ErrInvalidEpoch
	ErrCustodyRotationPending = custody.ErrCustodyRotationPending
)

// --- proof_custody.go ---

// ProofCustodyConfig type alias.
type ProofCustodyConfig = custody.ProofCustodyConfig

// DefaultProofCustodyConfig returns a default ProofCustodyConfig.
var DefaultProofCustodyConfig = custody.DefaultProofCustodyConfig

// CustodyBond type alias.
type CustodyBond = custody.CustodyBond

// DataHeldProof type alias.
type DataHeldProof = custody.DataHeldProof

// CustodyBondChallenge type alias.
type CustodyBondChallenge = custody.CustodyBondChallenge

// SlashResult type alias.
type SlashResult = custody.SlashResult

// ProofCustodyScheme type alias.
type ProofCustodyScheme = custody.ProofCustodyScheme

// NewProofCustodyScheme creates a new ProofCustodyScheme.
var NewProofCustodyScheme = custody.NewProofCustodyScheme

// Proof custody errors.
var (
	ErrBondAlreadyRegistered = custody.ErrBondAlreadyRegistered
	ErrBondNotFound          = custody.ErrBondNotFound
	ErrBondExpired           = custody.ErrBondExpired
	ErrStakeTooLow           = custody.ErrStakeTooLow
	ErrNilBond               = custody.ErrNilBond
	ErrNilChallenge          = custody.ErrNilChallenge
	ErrChallengeDeadlinePast = custody.ErrChallengeDeadlinePast
	ErrDataEmpty             = custody.ErrDataEmpty
	ErrInvalidProof          = custody.ErrInvalidProof
)
