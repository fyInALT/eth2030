package das

// network_compat.go re-exports types, functions, and variables from
// das/network for backward compatibility with existing callers.

import "github.com/eth2030/eth2030/das/network"

// Error variables re-exported from das/network.
var (
	ErrDASNotStarted       = network.ErrDASNotStarted
	ErrInvalidBlobIdx      = network.ErrInvalidBlobIdx
	ErrInvalidCellIdx      = network.ErrInvalidCellIdx
	ErrInvalidSampleData   = network.ErrInvalidSampleData
	ErrSampleNotAvailable  = network.ErrSampleNotAvailable
	ErrVerificationFailed  = network.ErrVerificationFailed
	ErrReconstructNotReady = network.ErrReconstructNotReady
	ErrReconstructDone     = network.ErrReconstructDone
	ErrDuplicateFragment   = network.ErrDuplicateFragment
	ErrFragmentOutOfRange  = network.ErrFragmentOutOfRange

	ErrNetworkNotStarted   = network.ErrNetworkNotStarted
	ErrInvalidSubnet       = network.ErrInvalidSubnet
	ErrInvalidSampleCount  = network.ErrInvalidSampleCount
	ErrSamplingFailed      = network.ErrSamplingFailed
	ErrPeerNotFound        = network.ErrPeerNotFound
	ErrNoAvailablePeers    = network.ErrNoAvailablePeers
	ErrAlreadySubscribed   = network.ErrAlreadySubscribed
	ErrColumnPublishFailed = network.ErrColumnPublishFailed
)

// Type aliases re-exported from das/network.
type (
	DASNetworkConfig    = network.DASNetworkConfig
	SampleResponse      = network.SampleResponse
	SampleStore         = network.SampleStore
	DASNetwork          = network.DASNetwork
	CustodySubnet       = network.CustodySubnet
	ColumnReconstructor = network.ColumnReconstructor

	NetworkConfig     = network.NetworkConfig
	SampleResult      = network.SampleResult
	SamplingResult    = network.SamplingResult
	PublishedColumn   = network.PublishedColumn
	NetworkMetrics    = network.NetworkMetrics
	DASNetworkManager = network.DASNetworkManager
	SampleProvider    = network.SampleProvider
	// Note: network.CustodyManager (interface) is not re-exported here because
	// the root das package defines a CustodyManager struct with the same name.
)

// Function aliases re-exported from das/network.
var (
	DefaultDASNetworkConfig = network.DefaultDASNetworkConfig
	NewDASNetwork           = network.NewDASNetwork
	NewDASNetworkWithStore  = network.NewDASNetworkWithStore
	VerifySample            = network.VerifySample
	ComputeSampleProof      = network.ComputeSampleProof
	AssignCustody           = network.AssignCustody
	IsCustodian             = network.IsCustodian
	NewColumnReconstructor  = network.NewColumnReconstructor

	DefaultNetworkConfig = network.DefaultNetworkConfig
	NewDASNetworkManager = network.NewDASNetworkManager
)
