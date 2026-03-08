package sync

// beam_compat.go re-exports types from sync/beam for backward compatibility.

import "github.com/eth2030/eth2030/sync/beam"

// Beam type aliases.
type (
	BeamStateFetcher   = beam.BeamStateFetcher
	BeamAccountData    = beam.BeamAccountData
	BeamSync           = beam.BeamSync
	BeamSyncStats      = beam.BeamSyncStats
	OnDemandDB         = beam.OnDemandDB
	BeamPrefetcher     = beam.BeamPrefetcher
	WitnessFetcher     = beam.WitnessFetcher
	ExecutionWitness   = beam.ExecutionWitness
	WitnessAccountData = beam.WitnessAccountData
	StatePrefill       = beam.StatePrefill
	BeamCacheEntry     = beam.BeamCacheEntry
	BeamCacheConfig    = beam.BeamCacheConfig
	FallbackConfig     = beam.FallbackConfig
	BeamStateSyncStats = beam.BeamStateSyncStats
	BeamStateSync      = beam.BeamStateSync
)

// Beam error variables.
var (
	ErrBeamFetchFailed        = beam.ErrBeamFetchFailed
	ErrBeamNoPeer             = beam.ErrBeamNoPeer
	ErrBeamAccountNotFound    = beam.ErrBeamAccountNotFound
	ErrBeamWitnessFetchFailed = beam.ErrBeamWitnessFetchFailed
	ErrBeamWitnessInvalid     = beam.ErrBeamWitnessInvalid
	ErrBeamFallbackTriggered  = beam.ErrBeamFallbackTriggered
	ErrBeamCacheFull          = beam.ErrBeamCacheFull
	ErrBeamExecutionFailed    = beam.ErrBeamExecutionFailed
)

// Beam function wrappers.
func NewBeamSync(fetcher BeamStateFetcher) *BeamSync { return beam.NewBeamSync(fetcher) }
func NewOnDemandDB(b *BeamSync) *OnDemandDB          { return beam.NewOnDemandDB(b) }
func NewBeamPrefetcher(b *BeamSync) *BeamPrefetcher  { return beam.NewBeamPrefetcher(b) }
func DefaultBeamCacheConfig() BeamCacheConfig        { return beam.DefaultBeamCacheConfig() }
func DefaultFallbackConfig() FallbackConfig          { return beam.DefaultFallbackConfig() }
func NewBeamStateSync(fetcher WitnessFetcher, cacheConfig BeamCacheConfig, fallbackConfig FallbackConfig) *BeamStateSync {
	return beam.NewBeamStateSync(fetcher, cacheConfig, fallbackConfig)
}
func NewStatePrefill(b *BeamStateSync) *StatePrefill { return beam.NewStatePrefill(b) }
