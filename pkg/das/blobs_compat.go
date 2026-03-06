package das

// blobs_compat.go re-exports types from das/blobs for backward compatibility.

import "github.com/eth2030/eth2030/das/blobs"

// Block-in-blob type aliases.
type (
	BlobBlockConfig     = blobs.BlobBlockConfig
	BlockBlob           = blobs.BlockBlob
	BlockBlobCommitment = blobs.BlockBlobCommitment
	BlobBlockEncoder    = blobs.BlobBlockEncoder
)

// Block-in-blob error variables.
var (
	ErrBlockDataEmpty    = blobs.ErrBlockDataEmpty
	ErrBlockDataTooLarge = blobs.ErrBlockDataTooLarge
	ErrNoBlobsProvided   = blobs.ErrNoBlobsProvided
	ErrBlobOrderMismatch = blobs.ErrBlobOrderMismatch
	ErrBlobHashMismatch  = blobs.ErrBlobHashMismatch
	ErrBlobCountMismatch = blobs.ErrBlobCountMismatch
	ErrMissingLastBlob   = blobs.ErrMissingLastBlob
	ErrBlobDataCorrupt   = blobs.ErrBlobDataCorrupt
	ErrMaxBlobsExceeded  = blobs.ErrMaxBlobsExceeded
)

// Block-in-blob function wrappers.
func DefaultBlobBlockConfig() BlobBlockConfig { return blobs.DefaultBlobBlockConfig() }
func NewBlobBlockEncoder(config BlobBlockConfig) *BlobBlockEncoder {
	return blobs.NewBlobBlockEncoder(config)
}

// Forward-cast type aliases.
type (
	ForwardCastConfig       = blobs.ForwardCastConfig
	ForwardCastAnnouncement = blobs.ForwardCastAnnouncement
	FulfillmentStatus       = blobs.FulfillmentStatus
	ForwardCaster           = blobs.ForwardCaster
)

// Forward-cast error variables.
var (
	ErrAnnouncementSlotPast   = blobs.ErrAnnouncementSlotPast
	ErrAnnouncementSlotTooFar = blobs.ErrAnnouncementSlotTooFar
	ErrAnnouncementSlotFull   = blobs.ErrAnnouncementSlotFull
	ErrAnnouncementExpired    = blobs.ErrAnnouncementExpired
	ErrAnnouncementNotFound   = blobs.ErrAnnouncementNotFound
	ErrAnnouncementFulfilled  = blobs.ErrAnnouncementFulfilled
	ErrBlobDataTooLargeFC     = blobs.ErrBlobDataTooLarge
	ErrBlobCommitmentMismatch = blobs.ErrBlobCommitmentMismatch
	ErrInvalidCommitment      = blobs.ErrInvalidCommitment
)

// Forward-cast function wrappers.
func DefaultForwardCastConfig() ForwardCastConfig { return blobs.DefaultForwardCastConfig() }
func NewForwardCaster(config ForwardCastConfig) *ForwardCaster {
	return blobs.NewForwardCaster(config)
}

// Teradata type aliases.
type (
	TeradataConfig  = blobs.TeradataConfig
	TeradataReceipt = blobs.TeradataReceipt
	L2DataStats     = blobs.L2DataStats
	TeradataManager = blobs.TeradataManager
)

// Teradata error variables.
var (
	ErrTeradataDataTooLarge    = blobs.ErrTeradataDataTooLarge
	ErrTeradataDataEmpty       = blobs.ErrTeradataDataEmpty
	ErrTeradataNotFound        = blobs.ErrTeradataNotFound
	ErrTeradataTooManyChains   = blobs.ErrTeradataTooManyChains
	ErrTeradataStorageFull     = blobs.ErrTeradataStorageFull
	ErrTeradataInvalidReceipt  = blobs.ErrTeradataInvalidReceipt
	ErrTeradataInvalidChainID  = blobs.ErrTeradataInvalidChainID
	ErrTeradataBandwidthDenied = blobs.ErrTeradataBandwidthDenied
)

// Teradata function wrappers.
func DefaultTeradataConfig() TeradataConfig { return blobs.DefaultTeradataConfig() }
func NewTeradataManager(config TeradataConfig) *TeradataManager {
	return blobs.NewTeradataManager(config)
}
func ValidateTeradataConfig(cfg TeradataConfig) error { return blobs.ValidateTeradataConfig(cfg) }
func ValidateTeradataReceipt(receipt *TeradataReceipt) error {
	return blobs.ValidateTeradataReceipt(receipt)
}
func ValidateBandwidthEnforcement(m *TeradataManager) error {
	return blobs.ValidateBandwidthEnforcement(m)
}
