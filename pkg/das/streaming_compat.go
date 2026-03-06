package das

// streaming_compat.go re-exports types from das/streaming for backward compatibility.

import "github.com/eth2030/eth2030/das/streaming"

// Streaming type aliases.
type (
	StreamConfig        = streaming.StreamConfig
	BlobChunk           = streaming.BlobChunk
	ChunkCallback       = streaming.ChunkCallback
	BlobStream          = streaming.BlobStream
	BlobStreamer        = streaming.BlobStreamer
	StreamSessionConfig = streaming.StreamSessionConfig
	StreamSession       = streaming.StreamSession
	StreamManager       = streaming.StreamManager
	PipelineConfig      = streaming.PipelineConfig
	PipelineItem        = streaming.PipelineItem
	PipelineMetrics     = streaming.PipelineMetrics
	ValidateFunc        = streaming.ValidateFunc
	DecodeFunc          = streaming.DecodeFunc
	StoreFunc           = streaming.StoreFunc
	StreamPipeline      = streaming.StreamPipeline
)

// Streaming error variables.
var (
	ErrStreamClosed       = streaming.ErrStreamClosed
	ErrDuplicateChunk     = streaming.ErrDuplicateChunk
	ErrChunkOutOfRange    = streaming.ErrChunkOutOfRange
	ErrChunkVerification  = streaming.ErrChunkVerification
	ErrMaxStreams         = streaming.ErrMaxStreams
	ErrStreamNotFound     = streaming.ErrStreamNotFound
	ErrIncompleteStream   = streaming.ErrIncompleteStream
	ErrSessionNotFound    = streaming.ErrSessionNotFound
	ErrSessionComplete    = streaming.ErrSessionComplete
	ErrSessionCancelled   = streaming.ErrSessionCancelled
	ErrChunkSizeMismatch  = streaming.ErrChunkSizeMismatch
	ErrBlobTooLarge       = streaming.ErrBlobTooLarge
	ErrMaxSessionsReached = streaming.ErrMaxSessionsReached
	ErrDuplicateSession   = streaming.ErrDuplicateSession
	ErrZeroChunkSize      = streaming.ErrZeroChunkSize
	ErrZeroTotalSize      = streaming.ErrZeroTotalSize
	ErrPipelineStopped    = streaming.ErrPipelineStopped
	ErrStageTimeout       = streaming.ErrStageTimeout
	ErrValidationFailed   = streaming.ErrValidationFailed
	ErrDecodeFailed       = streaming.ErrDecodeFailed
	ErrStoreFailed        = streaming.ErrStoreFailed
)

// Streaming function wrappers.
func DefaultStreamConfig() StreamConfig { return streaming.DefaultStreamConfig() }
func NewBlobStreamer(config StreamConfig) *BlobStreamer {
	return streaming.NewBlobStreamer(config)
}
func ValidateStreamConfig(cfg StreamConfig) error { return streaming.ValidateStreamConfig(cfg) }
func ValidateBlobChunk(chunk *BlobChunk, chunkSize uint32) error {
	return streaming.ValidateBlobChunk(chunk, chunkSize)
}
func DefaultSessionConfig() StreamSessionConfig { return streaming.DefaultSessionConfig() }
func NewStreamSession(blobIndex uint64, totalSize uint64, chunkSize uint64) *StreamSession {
	return streaming.NewStreamSession(blobIndex, totalSize, chunkSize)
}
func NewStreamManager(config StreamSessionConfig) *StreamManager {
	return streaming.NewStreamManager(config)
}
func DefaultPipelineConfig() PipelineConfig { return streaming.DefaultPipelineConfig() }
func NewStreamPipeline(config PipelineConfig, vf ValidateFunc, df DecodeFunc, sf StoreFunc) *StreamPipeline {
	return streaming.NewStreamPipeline(config, vf, df, sf)
}
