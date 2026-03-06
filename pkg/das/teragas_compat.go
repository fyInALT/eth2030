package das

// teragas_compat.go re-exports types from das/teragas for backward compatibility.

import "github.com/eth2030/eth2030/das/teragas"

// Teragas type aliases.
type (
	TokenBucket          = teragas.TokenBucket
	Reservation          = teragas.Reservation
	ThroughputReport     = teragas.ThroughputReport
	BandwidthPolicy      = teragas.BandwidthPolicy
	PeerBandwidthTracker = teragas.PeerBandwidthTracker
	AdaptiveRateLimiter  = teragas.AdaptiveRateLimiter
	BandwidthController  = teragas.BandwidthController
	BandwidthConfig      = teragas.BandwidthConfig
	BandwidthMetrics     = teragas.BandwidthMetrics
	BandwidthEnforcer    = teragas.BandwidthEnforcer
	ThroughputStats      = teragas.ThroughputStats
	StreamingPipeline    = teragas.StreamingPipeline
	ThroughputConfig     = teragas.ThroughputConfig
	ThroughputManager    = teragas.ThroughputManager
	ThroughputStatus     = teragas.ThroughputStatus
	DataProducer         = teragas.DataProducer
	DataConsumer         = teragas.DataConsumer
	PipelineStage        = teragas.PipelineStage
	TPDataPacket         = teragas.TPDataPacket
	DropPolicy           = teragas.DropPolicy
	TPConfig             = teragas.TPConfig
	TPMetricsSnapshot    = teragas.TPMetricsSnapshot
	BackpressureChannel  = teragas.BackpressureChannel
	BandwidthGate        = teragas.BandwidthGate
	CompressionStage     = teragas.CompressionStage
	ChunkingStage        = teragas.ChunkingStage
	ReassemblyStage      = teragas.ReassemblyStage
	TeragasPipeline      = teragas.TeragasPipeline
)

// Teragas constants.
const (
	DropOldest = teragas.DropOldest
	BlockOnFull = teragas.BlockOnFull
)

// Teragas error variables.
var (
	ErrBWInsufficientTokens  = teragas.ErrBWInsufficientTokens
	ErrBWReservationExpired  = teragas.ErrBWReservationExpired
	ErrBWReservationTooLarge = teragas.ErrBWReservationTooLarge
	ErrBWPeerNotFound        = teragas.ErrBWPeerNotFound
	ErrBWPeerLimitExceeded   = teragas.ErrBWPeerLimitExceeded
	ErrBWPolicyViolation     = teragas.ErrBWPolicyViolation
	ErrBWControllerStopped   = teragas.ErrBWControllerStopped
	ErrBWZeroRate            = teragas.ErrBWZeroRate
	ErrBandwidthExceeded     = teragas.ErrBandwidthExceeded
	ErrChainNotRegistered    = teragas.ErrChainNotRegistered
	ErrGlobalCapExceeded     = teragas.ErrGlobalCapExceeded
	ErrBackpressureActive    = teragas.ErrBackpressureActive
	ErrZeroBandwidthCap      = teragas.ErrZeroBandwidthCap
	ErrStreamBandwidthDenied = teragas.ErrStreamBandwidthDenied
	ErrStreamNilEnforcer     = teragas.ErrStreamNilEnforcer
	ErrInvalidThroughputConfig = teragas.ErrInvalidThroughputConfig
	ErrSlotNotMonotonic      = teragas.ErrSlotNotMonotonic
	ErrTPPipelineStopped     = teragas.ErrTPPipelineStopped
	ErrTPNilData             = teragas.ErrTPNilData
	ErrTPInvalidConfig       = teragas.ErrTPInvalidConfig
	ErrTPBandwidthDenied     = teragas.ErrTPBandwidthDenied
	ErrTPCompressionFailed   = teragas.ErrTPCompressionFailed
	ErrTPReassemblyFailed    = teragas.ErrTPReassemblyFailed
)

// Teragas function wrappers.
func NewTokenBucket(rate float64, capacity int64) (*TokenBucket, error) {
	return teragas.NewTokenBucket(rate, capacity)
}
func DefaultBandwidthPolicy() BandwidthPolicy { return teragas.DefaultBandwidthPolicy() }
func NewPeerBandwidthTracker(perPeerBps float64) *PeerBandwidthTracker {
	return teragas.NewPeerBandwidthTracker(perPeerBps)
}
func NewAdaptiveRateLimiter(minRate, maxRate float64) (*AdaptiveRateLimiter, error) {
	return teragas.NewAdaptiveRateLimiter(minRate, maxRate)
}
func NewBandwidthController(policy BandwidthPolicy) (*BandwidthController, error) {
	return teragas.NewBandwidthController(policy)
}
func ComputeOptimalChunkSize(targetBps float64, latencyBudgetMs int, minChunkSize, maxChunkSize int64) int64 {
	return teragas.ComputeOptimalChunkSize(targetBps, latencyBudgetMs, minChunkSize, maxChunkSize)
}
func DefaultBandwidthConfig() BandwidthConfig  { return teragas.DefaultBandwidthConfig() }
func NewBandwidthEnforcer(config BandwidthConfig) (*BandwidthEnforcer, error) {
	return teragas.NewBandwidthEnforcer(config)
}
func NewStreamingPipeline(enforcer *BandwidthEnforcer) (*StreamingPipeline, error) {
	return teragas.NewStreamingPipeline(enforcer)
}
func DefaultThroughputConfig() ThroughputConfig { return teragas.DefaultThroughputConfig() }
func NewThroughputManager(config ThroughputConfig) (*ThroughputManager, error) {
	return teragas.NewThroughputManager(config)
}
func DefaultTPConfig() *TPConfig { return teragas.DefaultTPConfig() }
func NewBackpressureChannel(size int, policy DropPolicy) *BackpressureChannel {
	return teragas.NewBackpressureChannel(size, policy)
}
func NewBandwidthGate(e *BandwidthEnforcer) *BandwidthGate { return teragas.NewBandwidthGate(e) }
func NewReassemblyStage() *ReassemblyStage                 { return teragas.NewReassemblyStage() }
func NewTeragasPipeline(config *TPConfig) (*TeragasPipeline, error) {
	return teragas.NewTeragasPipeline(config)
}
