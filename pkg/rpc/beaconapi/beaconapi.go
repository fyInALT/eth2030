// Package beaconapi implements a subset of the Ethereum Beacon API
// as JSON-RPC methods, allowing consensus-layer clients to interact
// via the standard RPC server.
package beaconapi

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	coretypes "github.com/eth2030/eth2030/core/types"
	rpcbackend "github.com/eth2030/eth2030/rpc/backend"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// Beacon API error codes per the Beacon API spec.
const (
	BeaconErrNotFound       = 404
	BeaconErrBadRequest     = 400
	BeaconErrInternal       = 500
	BeaconErrNotImplemented = 501
)

// BeaconError represents a Beacon API error response.
type BeaconError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *BeaconError) Error() string {
	return fmt.Sprintf("beacon error %d: %s", e.Code, e.Message)
}

// --- Response types ---

// GenesisResponse is the response for beacon_getGenesis.
type GenesisResponse struct {
	GenesisTime           string `json:"genesis_time"`
	GenesisValidatorsRoot string `json:"genesis_validators_root"`
	GenesisForkVersion    string `json:"genesis_fork_version"`
}

// BlockResponse is the response for beacon_getBlock.
type BlockResponse struct {
	Slot          string `json:"slot"`
	ProposerIndex string `json:"proposer_index"`
	ParentRoot    string `json:"parent_root"`
	StateRoot     string `json:"state_root"`
	BodyRoot      string `json:"body_root"`
}

// HeaderResponse is the response for beacon_getBlockHeader.
type HeaderResponse struct {
	Root      string            `json:"root"`
	Canonical bool              `json:"canonical"`
	Header    *SignedHeaderData `json:"header"`
}

// SignedHeaderData wraps a beacon block header with its signature.
type SignedHeaderData struct {
	Message   *BeaconHeaderMessage `json:"message"`
	Signature string               `json:"signature"`
}

// BeaconHeaderMessage contains the beacon block header fields.
type BeaconHeaderMessage struct {
	Slot          string `json:"slot"`
	ProposerIndex string `json:"proposer_index"`
	ParentRoot    string `json:"parent_root"`
	StateRoot     string `json:"state_root"`
	BodyRoot      string `json:"body_root"`
}

// StateRootResponse is the response for beacon_getStateRoot.
type StateRootResponse struct {
	Root string `json:"root"`
}

// FinalityCheckpointsResponse is the response for beacon_getStateFinalityCheckpoints.
type FinalityCheckpointsResponse struct {
	PreviousJustified *Checkpoint `json:"previous_justified"`
	CurrentJustified  *Checkpoint `json:"current_justified"`
	Finalized         *Checkpoint `json:"finalized"`
}

// Checkpoint represents an epoch checkpoint.
type Checkpoint struct {
	Epoch string `json:"epoch"`
	Root  string `json:"root"`
}

// ValidatorListResponse is the response for beacon_getStateValidators.
type ValidatorListResponse struct {
	Validators []*ValidatorEntry `json:"validators"`
}

// ValidatorEntry represents a single validator in the list.
type ValidatorEntry struct {
	Index     string         `json:"index"`
	Balance   string         `json:"balance"`
	Status    string         `json:"status"`
	Validator *ValidatorData `json:"validator"`
}

// ValidatorData contains the validator's registration fields.
type ValidatorData struct {
	Pubkey                string `json:"pubkey"`
	WithdrawalCredentials string `json:"withdrawal_credentials"`
	EffectiveBalance      string `json:"effective_balance"`
	Slashed               bool   `json:"slashed"`
	ActivationEpoch       string `json:"activation_epoch"`
	ExitEpoch             string `json:"exit_epoch"`
}

// VersionResponse is the response for beacon_getNodeVersion.
type VersionResponse struct {
	Version string `json:"version"`
}

// SyncingResponse is the response for beacon_getNodeSyncing.
type SyncingResponse struct {
	HeadSlot     string `json:"head_slot"`
	SyncDistance string `json:"sync_distance"`
	IsSyncing    bool   `json:"is_syncing"`
	IsOptimistic bool   `json:"is_optimistic"`
}

// PeerListResponse is the response for beacon_getNodePeers.
type PeerListResponse struct {
	Peers []*BeaconPeer `json:"peers"`
}

// BeaconPeer describes a connected beacon peer.
type BeaconPeer struct {
	PeerID    string `json:"peer_id"`
	State     string `json:"state"`
	Direction string `json:"direction"`
	Address   string `json:"address"`
}

// --- Consensus state ---

// ConsensusState holds the consensus-layer state the Beacon API reads from.
type ConsensusState struct {
	mu sync.RWMutex

	GenesisTime        uint64
	GenesisValRoot     coretypes.Hash
	GenesisForkVersion [4]byte

	HeadSlot       uint64
	FinalizedEpoch uint64
	FinalizedRoot  coretypes.Hash
	JustifiedEpoch uint64
	JustifiedRoot  coretypes.Hash

	Validators []*ValidatorEntry

	IsSyncing    bool
	SyncDistance uint64

	Peers []*BeaconPeer
}

// NewConsensusState creates a ConsensusState with default genesis values.
func NewConsensusState() *ConsensusState {
	return &ConsensusState{
		GenesisTime:        uint64(time.Date(2020, 12, 1, 12, 0, 23, 0, time.UTC).Unix()),
		GenesisForkVersion: [4]byte{0x00, 0x00, 0x00, 0x00},
	}
}

// --- BeaconAPI ---

// BeaconAPI implements the Beacon API JSON-RPC methods.
type BeaconAPI struct {
	state   *ConsensusState
	backend rpcbackend.Backend
}

// NewBeaconAPI creates a new BeaconAPI.
func NewBeaconAPI(state *ConsensusState, backend rpcbackend.Backend) *BeaconAPI {
	return &BeaconAPI{state: state, backend: backend}
}

// RegisterBeaconRoutes registers all beacon_ methods into the given method map.
func RegisterBeaconRoutes(api *BeaconAPI) map[string]func(*rpctypes.Request) *rpctypes.Response {
	return map[string]func(*rpctypes.Request) *rpctypes.Response{
		"beacon_getGenesis":                  api.getGenesis,
		"beacon_getBlock":                    api.getBlock,
		"beacon_getBlockHeader":              api.getBlockHeader,
		"beacon_getStateRoot":                api.getStateRoot,
		"beacon_getStateFinalityCheckpoints": api.getStateFinalityCheckpoints,
		"beacon_getStateValidators":          api.getStateValidators,
		"beacon_getNodeVersion":              api.getNodeVersion,
		"beacon_getNodeSyncing":              api.getNodeSyncing,
		"beacon_getNodePeers":                api.getNodePeers,
		"beacon_getNodeHealth":               api.getNodeHealth,
	}
}

func (api *BeaconAPI) getGenesis(req *rpctypes.Request) *rpctypes.Response {
	api.state.mu.RLock()
	defer api.state.mu.RUnlock()

	resp := &GenesisResponse{
		GenesisTime:           fmt.Sprintf("%d", api.state.GenesisTime),
		GenesisValidatorsRoot: rpctypes.EncodeHash(api.state.GenesisValRoot),
		GenesisForkVersion:    fmt.Sprintf("0x%x", api.state.GenesisForkVersion),
	}
	return rpctypes.NewSuccessResponse(req.ID, resp)
}

func (api *BeaconAPI) getBlock(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return beaconErr(req.ID, BeaconErrBadRequest, "missing slot parameter")
	}
	var slotStr string
	if err := json.Unmarshal(req.Params[0], &slotStr); err != nil {
		return beaconErr(req.ID, BeaconErrBadRequest, "invalid slot parameter")
	}
	slot := rpctypes.ParseHexUint64(slotStr)
	if slot > uint64(math.MaxInt64) {
		return beaconErr(req.ID, BeaconErrBadRequest, "slot number overflow")
	}
	header := api.backend.HeaderByNumber(rpctypes.BlockNumber(slot)) //nolint:gosec
	if header == nil {
		return beaconErr(req.ID, BeaconErrNotFound, fmt.Sprintf("block at slot %d not found", slot))
	}
	return rpctypes.NewSuccessResponse(req.ID, &BlockResponse{
		Slot:          fmt.Sprintf("%d", slot),
		ProposerIndex: "0",
		ParentRoot:    rpctypes.EncodeHash(header.ParentHash),
		StateRoot:     rpctypes.EncodeHash(header.Root),
		BodyRoot:      rpctypes.EncodeHash(header.TxHash),
	})
}

func (api *BeaconAPI) getBlockHeader(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return beaconErr(req.ID, BeaconErrBadRequest, "missing slot parameter")
	}
	var slotStr string
	if err := json.Unmarshal(req.Params[0], &slotStr); err != nil {
		return beaconErr(req.ID, BeaconErrBadRequest, "invalid slot parameter")
	}
	slot := rpctypes.ParseHexUint64(slotStr)
	if slot > uint64(math.MaxInt64) {
		return beaconErr(req.ID, BeaconErrBadRequest, "slot number overflow")
	}
	header := api.backend.HeaderByNumber(rpctypes.BlockNumber(slot)) //nolint:gosec
	if header == nil {
		return beaconErr(req.ID, BeaconErrNotFound, fmt.Sprintf("header at slot %d not found", slot))
	}
	return rpctypes.NewSuccessResponse(req.ID, &HeaderResponse{
		Root:      rpctypes.EncodeHash(header.Hash()),
		Canonical: true,
		Header: &SignedHeaderData{
			Message: &BeaconHeaderMessage{
				Slot:          fmt.Sprintf("%d", slot),
				ProposerIndex: "0",
				ParentRoot:    rpctypes.EncodeHash(header.ParentHash),
				StateRoot:     rpctypes.EncodeHash(header.Root),
				BodyRoot:      rpctypes.EncodeHash(header.TxHash),
			},
			Signature: "0x" + fmt.Sprintf("%0192x", 0),
		},
	})
}

func (api *BeaconAPI) getStateRoot(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return beaconErr(req.ID, BeaconErrBadRequest, "missing state_id parameter")
	}
	var stateID string
	if err := json.Unmarshal(req.Params[0], &stateID); err != nil {
		return beaconErr(req.ID, BeaconErrBadRequest, "invalid state_id parameter")
	}
	var header *coretypes.Header
	switch stateID {
	case "head":
		header = api.backend.CurrentHeader()
	case "finalized":
		api.state.mu.RLock()
		epoch := api.state.FinalizedEpoch
		api.state.mu.RUnlock()
		header = api.backend.HeaderByNumber(rpctypes.BlockNumber(epoch * 32))
	case "justified":
		api.state.mu.RLock()
		epoch := api.state.JustifiedEpoch
		api.state.mu.RUnlock()
		header = api.backend.HeaderByNumber(rpctypes.BlockNumber(epoch * 32))
	default:
		slot := rpctypes.ParseHexUint64(stateID)
		if slot > uint64(math.MaxInt64) {
			return beaconErr(req.ID, BeaconErrBadRequest, "slot number overflow")
		}
		header = api.backend.HeaderByNumber(rpctypes.BlockNumber(slot)) //nolint:gosec
	}
	if header == nil {
		return beaconErr(req.ID, BeaconErrNotFound, "state not found")
	}
	return rpctypes.NewSuccessResponse(req.ID, &StateRootResponse{Root: rpctypes.EncodeHash(header.Root)})
}

func (api *BeaconAPI) getStateFinalityCheckpoints(req *rpctypes.Request) *rpctypes.Response {
	api.state.mu.RLock()
	defer api.state.mu.RUnlock()

	prevEpoch := uint64(0)
	if api.state.JustifiedEpoch > 0 {
		prevEpoch = api.state.JustifiedEpoch - 1
	}
	return rpctypes.NewSuccessResponse(req.ID, &FinalityCheckpointsResponse{
		PreviousJustified: &Checkpoint{
			Epoch: fmt.Sprintf("%d", prevEpoch),
			Root:  rpctypes.EncodeHash(api.state.JustifiedRoot),
		},
		CurrentJustified: &Checkpoint{
			Epoch: fmt.Sprintf("%d", api.state.JustifiedEpoch),
			Root:  rpctypes.EncodeHash(api.state.JustifiedRoot),
		},
		Finalized: &Checkpoint{
			Epoch: fmt.Sprintf("%d", api.state.FinalizedEpoch),
			Root:  rpctypes.EncodeHash(api.state.FinalizedRoot),
		},
	})
}

func (api *BeaconAPI) getStateValidators(req *rpctypes.Request) *rpctypes.Response {
	api.state.mu.RLock()
	defer api.state.mu.RUnlock()

	validators := api.state.Validators
	if validators == nil {
		validators = []*ValidatorEntry{}
	}
	return rpctypes.NewSuccessResponse(req.ID, &ValidatorListResponse{Validators: validators})
}

func (api *BeaconAPI) getNodeVersion(req *rpctypes.Request) *rpctypes.Response {
	return rpctypes.NewSuccessResponse(req.ID, &VersionResponse{Version: "ETH2030/v0.1.0-beacon"})
}

func (api *BeaconAPI) getNodeSyncing(req *rpctypes.Request) *rpctypes.Response {
	api.state.mu.RLock()
	defer api.state.mu.RUnlock()

	headSlot := uint64(0)
	if header := api.backend.CurrentHeader(); header != nil {
		headSlot = header.Number.Uint64()
	}
	return rpctypes.NewSuccessResponse(req.ID, &SyncingResponse{
		HeadSlot:     fmt.Sprintf("%d", headSlot),
		SyncDistance: fmt.Sprintf("%d", api.state.SyncDistance),
		IsSyncing:    api.state.IsSyncing,
		IsOptimistic: false,
	})
}

func (api *BeaconAPI) getNodePeers(req *rpctypes.Request) *rpctypes.Response {
	api.state.mu.RLock()
	defer api.state.mu.RUnlock()

	peers := api.state.Peers
	if peers == nil {
		peers = []*BeaconPeer{}
	}
	return rpctypes.NewSuccessResponse(req.ID, &PeerListResponse{Peers: peers})
}

func (api *BeaconAPI) getNodeHealth(req *rpctypes.Request) *rpctypes.Response {
	api.state.mu.RLock()
	syncing := api.state.IsSyncing
	api.state.mu.RUnlock()

	status := "healthy"
	if syncing {
		status = "syncing"
	}
	return rpctypes.NewSuccessResponse(req.ID, map[string]string{"status": status})
}

func beaconErr(id json.RawMessage, code int, msg string) *rpctypes.Response {
	return rpctypes.NewErrorResponse(id, code, msg)
}
