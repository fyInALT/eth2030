// beacon_api.go re-exports beacon API types from rpc/beaconapi for
// backward compatibility.
package rpc

import "github.com/eth2030/eth2030/rpc/beaconapi"

// Beacon API error code constants.
const (
	BeaconErrNotFound       = beaconapi.BeaconErrNotFound
	BeaconErrBadRequest     = beaconapi.BeaconErrBadRequest
	BeaconErrInternal       = beaconapi.BeaconErrInternal
	BeaconErrNotImplemented = beaconapi.BeaconErrNotImplemented
)

// Re-export all beacon API types.
type (
	BeaconError                 = beaconapi.BeaconError
	GenesisResponse             = beaconapi.GenesisResponse
	BlockResponse               = beaconapi.BlockResponse
	HeaderResponse              = beaconapi.HeaderResponse
	SignedHeaderData            = beaconapi.SignedHeaderData
	BeaconHeaderMessage         = beaconapi.BeaconHeaderMessage
	StateRootResponse           = beaconapi.StateRootResponse
	FinalityCheckpointsResponse = beaconapi.FinalityCheckpointsResponse
	Checkpoint                  = beaconapi.Checkpoint
	ValidatorListResponse       = beaconapi.ValidatorListResponse
	ValidatorEntry              = beaconapi.ValidatorEntry
	ValidatorData               = beaconapi.ValidatorData
	VersionResponse             = beaconapi.VersionResponse
	SyncingResponse             = beaconapi.SyncingResponse
	PeerListResponse            = beaconapi.PeerListResponse
	BeaconPeer                  = beaconapi.BeaconPeer
	ConsensusState              = beaconapi.ConsensusState
	BeaconAPI                   = beaconapi.BeaconAPI
)

// Re-export constructors.
var (
	NewConsensusState    = beaconapi.NewConsensusState
	NewBeaconAPI         = beaconapi.NewBeaconAPI
	RegisterBeaconRoutes = beaconapi.RegisterBeaconRoutes
)
