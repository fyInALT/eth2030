package sync

// beacon_compat.go re-exports types from sync/beacon for backward compatibility.

import "github.com/eth2030/eth2030/sync/beacon"

// Beacon type aliases.
type (
	BeaconBlock            = beacon.BeaconBlock
	BlobSidecar            = beacon.BlobSidecar
	BeaconSyncConfig       = beacon.BeaconSyncConfig
	SyncStatus             = beacon.SyncStatus
	BeaconBlockFetcher     = beacon.BeaconBlockFetcher
	BeaconSyncer           = beacon.BeaconSyncer
	BlobRecovery           = beacon.BlobRecovery
	BlobSidecarV2          = beacon.BlobSidecarV2
	BlobSidecarRequest     = beacon.BlobSidecarRequest
	BlobSidecarResponse    = beacon.BlobSidecarResponse
	BlobSyncProtocolConfig = beacon.BlobSyncProtocolConfig
	BlobSyncProtocol       = beacon.BlobSyncProtocol
	BlobSyncConfig         = beacon.BlobSyncConfig
	BlobSyncManager        = beacon.BlobSyncManager
)

// Beacon constants.
const (
	MaxBlobsPerBlock = beacon.MaxBlobsPerBlock
)

// Beacon error variables.
var (
	ErrBeaconAlreadySyncing   = beacon.ErrBeaconAlreadySyncing
	ErrBeaconInvalidSlotRange = beacon.ErrBeaconInvalidSlotRange
	ErrBeaconSlotTimeout      = beacon.ErrBeaconSlotTimeout
	ErrBeaconBlockNil         = beacon.ErrBeaconBlockNil
	ErrBeaconBlockInvalid     = beacon.ErrBeaconBlockInvalid
	ErrBeaconSidecarNil       = beacon.ErrBeaconSidecarNil
	ErrBeaconSidecarInvalid   = beacon.ErrBeaconSidecarInvalid
	ErrBeaconBlobIndexInvalid = beacon.ErrBeaconBlobIndexInvalid
	ErrBeaconMaxRetries       = beacon.ErrBeaconMaxRetries
	ErrBlobRecoveryFailed     = beacon.ErrBlobRecoveryFailed
)

// Beacon function wrappers.
func DefaultBeaconSyncConfig() BeaconSyncConfig {
	return beacon.DefaultBeaconSyncConfig()
}
func NewBeaconSyncer(config BeaconSyncConfig) *BeaconSyncer {
	return beacon.NewBeaconSyncer(config)
}
func NewBlobRecovery(custody int) *BlobRecovery {
	return beacon.NewBlobRecovery(custody)
}
func DefaultBlobSyncProtocolConfig() BlobSyncProtocolConfig {
	return beacon.DefaultBlobSyncProtocolConfig()
}
func NewBlobSyncProtocol(config BlobSyncProtocolConfig) *BlobSyncProtocol {
	return beacon.NewBlobSyncProtocol(config)
}
func DefaultBlobSyncConfig() BlobSyncConfig {
	return beacon.DefaultBlobSyncConfig()
}
func NewBlobSyncManager(config BlobSyncConfig) *BlobSyncManager {
	return beacon.NewBlobSyncManager(config)
}
