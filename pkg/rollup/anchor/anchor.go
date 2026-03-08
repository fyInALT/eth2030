// anchor.go implements the anchor predeploy contract for native rollups
// (EIP-8079). It manages L1->L2 anchoring via a ring buffer similar to
// EIP-4788, storing the latest L1 block hash and state root.
package anchor

import (
	"encoding/binary"
	"errors"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/rollup"
)

// Anchor storage slot constants (similar to EIP-4788 beacon root contract).
const (
	// RingBufferSize is the number of anchor entries stored.
	// Matches the EIP-4788 history buffer size.
	RingBufferSize = 8191

	// SlotBlockHash is the base storage slot for block hashes.
	SlotBlockHash = 0

	// SlotStateRoot is the base storage slot for state roots.
	SlotStateRoot = RingBufferSize

	// SlotLatestBlockNumber stores the latest anchored block number.
	SlotLatestBlockNumber = RingBufferSize * 2

	// SlotLatestTimestamp stores the latest anchored timestamp.
	SlotLatestTimestamp = RingBufferSize*2 + 1
)

// Errors for anchor operations.
var (
	ErrAnchorDataTooShort = errors.New("anchor: data too short")
	ErrAnchorStaleBlock   = errors.New("anchor: block number not increasing")
)

// Contract manages the anchor predeploy state for a native rollup.
// It provides L1->L2 anchoring by storing the latest L1 block hash and
// state root in a ring buffer, similar to EIP-4788.
type Contract struct {
	// state tracks the current anchor state.
	state rollup.AnchorState

	// history stores past anchor entries in a ring buffer.
	history [RingBufferSize]Entry
}

// Entry is a single entry in the anchor ring buffer.
type Entry struct {
	BlockHash types.Hash
	StateRoot types.Hash
	Timestamp uint64
}

// NewContract creates a new anchor contract with empty state.
func NewContract() *Contract {
	return &Contract{}
}

// GetLatestState returns the most recent anchor state.
func (ac *Contract) GetLatestState() rollup.AnchorState {
	return ac.state
}

// UpdateState updates the anchor with a new L1 state.
// The block number must be strictly increasing.
func (ac *Contract) UpdateState(newState rollup.AnchorState) error {
	if newState.BlockNumber <= ac.state.BlockNumber && ac.state.BlockNumber > 0 {
		return ErrAnchorStaleBlock
	}

	// Store in ring buffer.
	idx := newState.BlockNumber % RingBufferSize
	ac.history[idx] = Entry{
		BlockHash: newState.LatestBlockHash,
		StateRoot: newState.LatestStateRoot,
		Timestamp: newState.Timestamp,
	}

	ac.state = newState
	return nil
}

// GetByNumber retrieves the anchor entry for a given block number
// if it is still in the ring buffer. Returns false if the entry has been
// overwritten or was never stored.
func (ac *Contract) GetByNumber(blockNumber uint64) (Entry, bool) {
	if blockNumber == 0 || blockNumber > ac.state.BlockNumber {
		return Entry{}, false
	}

	// Check if the entry is still within the ring buffer window.
	if ac.state.BlockNumber-blockNumber >= RingBufferSize {
		return Entry{}, false
	}

	idx := blockNumber % RingBufferSize
	entry := ac.history[idx]

	// Verify it's the correct entry (not overwritten).
	if entry.BlockHash == (types.Hash{}) {
		return Entry{}, false
	}

	return entry, true
}

// ProcessAnchorData decodes and applies anchor data from an EXECUTE call.
// Anchor data format:
//
//	[0:32]   blockHash    (bytes32)
//	[32:64]  stateRoot    (bytes32)
//	[64:72]  blockNumber  (uint64, big-endian)
//	[72:80]  timestamp    (uint64, big-endian)
func (ac *Contract) ProcessAnchorData(data []byte) error {
	if len(data) < 80 {
		return ErrAnchorDataTooShort
	}

	var blockHash, stateRoot types.Hash
	copy(blockHash[:], data[0:32])
	copy(stateRoot[:], data[32:64])
	blockNumber := binary.BigEndian.Uint64(data[64:72])
	timestamp := binary.BigEndian.Uint64(data[72:80])

	return ac.UpdateState(rollup.AnchorState{
		LatestBlockHash: blockHash,
		LatestStateRoot: stateRoot,
		BlockNumber:     blockNumber,
		Timestamp:       timestamp,
	})
}

// UpdateAfterExecute advances the anchor state after a successful EXECUTE
// precompile call. It validates the execution output, constructs the new anchor
// state from the output's post-state root and the provided block metadata, and
// updates the ring buffer. Returns an error if the output indicates failure or
// if the block number does not advance.
func (ac *Contract) UpdateAfterExecute(output *rollup.ExecuteOutput, blockNumber, timestamp uint64) error {
	if output == nil {
		return ErrAnchorDataTooShort
	}
	if !output.Success {
		return rollup.ErrSTFailed
	}
	if output.PostStateRoot == (types.Hash{}) {
		return rollup.ErrInvalidBlockData
	}

	newState := rollup.AnchorState{
		LatestBlockHash: crypto.Keccak256Hash(output.PostStateRoot[:]),
		LatestStateRoot: output.PostStateRoot,
		BlockNumber:     blockNumber,
		Timestamp:       timestamp,
	}
	return ac.UpdateState(newState)
}

// EncodeAnchorData encodes an AnchorState into the wire format expected
// by ProcessAnchorData.
func EncodeAnchorData(state rollup.AnchorState) []byte {
	data := make([]byte, 80)
	copy(data[0:32], state.LatestBlockHash[:])
	copy(data[32:64], state.LatestStateRoot[:])
	binary.BigEndian.PutUint64(data[64:72], state.BlockNumber)
	binary.BigEndian.PutUint64(data[72:80], state.Timestamp)
	return data
}
