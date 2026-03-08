// Package das implements PeerDAS (Peer Data Availability Sampling) per EIP-7594.
// It re-exports types and functions from its sub-packages for backward compatibility.
package das

import (
	"github.com/eth2030/eth2030/das/blobs"
	"github.com/eth2030/eth2030/das/blockerasure"
	"github.com/eth2030/eth2030/das/custody"
	"github.com/eth2030/eth2030/das/dastypes"
	"github.com/eth2030/eth2030/das/futures"
	"github.com/eth2030/eth2030/das/gossip"
	"github.com/eth2030/eth2030/das/pqblob"
	"github.com/eth2030/eth2030/das/reconstruction"
	"github.com/eth2030/eth2030/das/sampleopt"
	"github.com/eth2030/eth2030/das/sampling"
	"github.com/eth2030/eth2030/das/streaming"
	"github.com/eth2030/eth2030/das/teragas"
	"github.com/eth2030/eth2030/das/varblob"
)

// ---------------------------------------------------------------------------
// dastypes: constants and core types
// ---------------------------------------------------------------------------

const (
	NumberOfColumns              = dastypes.NumberOfColumns
	NumberOfCustodyGroups        = dastypes.NumberOfCustodyGroups
	CustodyRequirement           = dastypes.CustodyRequirement
	SamplesPerSlot               = dastypes.SamplesPerSlot
	DataColumnSidecarSubnetCount = dastypes.DataColumnSidecarSubnetCount
	FieldElementsPerBlob         = dastypes.FieldElementsPerBlob
	FieldElementsPerCell         = dastypes.FieldElementsPerCell
	BytesPerFieldElement         = dastypes.BytesPerFieldElement
	BytesPerCell                 = dastypes.BytesPerCell
	CellsPerExtBlob              = dastypes.CellsPerExtBlob
	MaxBlobCommitmentsPerBlock   = dastypes.MaxBlobCommitmentsPerBlock
	ReconstructionThreshold      = dastypes.ReconstructionThreshold
)

type (
	SubnetID          = dastypes.SubnetID
	CustodyGroup      = dastypes.CustodyGroup
	ColumnIndex       = dastypes.ColumnIndex
	Cell              = dastypes.Cell
	KZGCommitment     = dastypes.KZGCommitment
	KZGProof          = dastypes.KZGProof
	DataColumnSidecar = dastypes.DataColumnSidecar
	MatrixEntry       = dastypes.MatrixEntry
)

// ---------------------------------------------------------------------------
// sampling: custody group / column assignment
// ---------------------------------------------------------------------------

var (
	ErrInvalidCustodyCount = sampling.ErrInvalidCustodyCount
	ErrInvalidSidecar      = sampling.ErrInvalidSidecar
	ErrMismatchedLengths   = sampling.ErrMismatchedLengths
)

var (
	GetCustodyGroups              = sampling.GetCustodyGroups
	ComputeColumnsForCustodyGroup = sampling.ComputeColumnsForCustodyGroup
	GetCustodyColumns             = sampling.GetCustodyColumns
	ShouldCustodyColumn           = sampling.ShouldCustodyColumn
	VerifyDataColumnSidecar       = sampling.VerifyDataColumnSidecar
	ColumnSubnet                  = sampling.ColumnSubnet
)

// ---------------------------------------------------------------------------
// reconstruction: blob/cell reconstruction
// ---------------------------------------------------------------------------

var (
	CanReconstruct  = reconstruction.CanReconstruct
	ReconstructBlob = reconstruction.ReconstructBlob
	RecoverMatrix   = reconstruction.RecoverMatrix
)

// ---------------------------------------------------------------------------
// gossip: cell gossip handler
// ---------------------------------------------------------------------------

type (
	CellGossipMessage       = gossip.CellGossipMessage
	CellGossipHandlerConfig = gossip.CellGossipHandlerConfig
	CellGossipHandler       = gossip.CellGossipHandler
)

var NewCellGossipHandler = gossip.NewCellGossipHandler

// ---------------------------------------------------------------------------
// custody: custody proofs and challenges
// ---------------------------------------------------------------------------

type (
	CustodyProof     = custody.CustodyProof
	CustodyChallenge = custody.CustodyChallenge
)

var (
	CreateChallenge      = custody.CreateChallenge
	GenerateCustodyProof = custody.GenerateCustodyProof
	RespondToChallenge   = custody.RespondToChallenge
	VerifyCustodyProof   = custody.VerifyCustodyProof
)

// ---------------------------------------------------------------------------
// blockerasure: block assembly from erasure-coded pieces
// ---------------------------------------------------------------------------

type (
	BlockAssemblyManagerConfig = blockerasure.BlockAssemblyManagerConfig
	BlockAssemblyManager       = blockerasure.BlockAssemblyManager
	BlockPiece                 = blockerasure.BlockPiece
)

var (
	NewBlockAssemblyManager           = blockerasure.NewBlockAssemblyManager
	DefaultBlockAssemblyManagerConfig = blockerasure.DefaultBlockAssemblyManagerConfig
)

// ---------------------------------------------------------------------------
// varblob: variable-size blob configuration
// ---------------------------------------------------------------------------

type BlobConfig = varblob.BlobConfig

const (
	DefaultBlobSize  = varblob.DefaultBlobSize
	MinBlobSizeBytes = varblob.MinBlobSizeBytes
	MaxBlobSizeBytes = varblob.MaxBlobSizeBytes
)

// ---------------------------------------------------------------------------
// blobs: block-in-blob encoding and teradata throughput
// ---------------------------------------------------------------------------

type (
	TeradataConfig  = blobs.TeradataConfig
	TeradataManager = blobs.TeradataManager
)

var (
	DefaultTeradataConfig      = blobs.DefaultTeradataConfig
	NewTeradataManager         = blobs.NewTeradataManager
	ErrTeradataBandwidthDenied = blobs.ErrTeradataBandwidthDenied
)

// ---------------------------------------------------------------------------
// futures: blob futures market
// ---------------------------------------------------------------------------

type FuturesMarket = futures.FuturesMarket

var NewFuturesMarket = futures.NewFuturesMarket

// ---------------------------------------------------------------------------
// pqblob: post-quantum blob commitments and proofs
// ---------------------------------------------------------------------------

var (
	CommitBlob           = pqblob.CommitBlob
	VerifyBlobCommitment = pqblob.VerifyBlobCommitment
	GenerateBlobProof    = pqblob.GenerateBlobProof
)

// ---------------------------------------------------------------------------
// sampleopt: sample size optimization
// ---------------------------------------------------------------------------

type (
	SampleOptimizerConfig = sampleopt.SampleOptimizerConfig
	SamplingPlan          = sampleopt.SamplingPlan
	SamplingVerdict       = sampleopt.SamplingVerdict
)

var (
	DefaultSampleOptimizerConfig = sampleopt.DefaultSampleOptimizerConfig
	NewSampleOptimizer           = sampleopt.NewSampleOptimizer
)

// ---------------------------------------------------------------------------
// streaming: blob streaming pipeline
// ---------------------------------------------------------------------------

type (
	BlobChunk    = streaming.BlobChunk
	BlobStreamer = streaming.BlobStreamer
)

var (
	DefaultStreamConfig  = streaming.DefaultStreamConfig
	NewBlobStreamer      = streaming.NewBlobStreamer
	DefaultSessionConfig = streaming.DefaultSessionConfig
	NewStreamManager     = streaming.NewStreamManager
)

// ---------------------------------------------------------------------------
// teragas: bandwidth enforcement and streaming pipeline
// ---------------------------------------------------------------------------

type (
	BandwidthConfig   = teragas.BandwidthConfig
	BandwidthEnforcer = teragas.BandwidthEnforcer
	StreamingPipeline = teragas.StreamingPipeline
)

var (
	DefaultBandwidthConfig = teragas.DefaultBandwidthConfig
	NewBandwidthEnforcer   = teragas.NewBandwidthEnforcer
	NewStreamingPipeline   = teragas.NewStreamingPipeline
	ErrStreamNilEnforcer   = teragas.ErrStreamNilEnforcer
)
