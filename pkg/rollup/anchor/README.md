# rollup/anchor — L1-to-L2 anchor predeploy and state manager (EIP-8079)

## Overview

This package implements two complementary components for the native rollup
anchor system defined in EIP-8079. The first is `Contract`, a ring-buffer
predeploy (modelled on EIP-4788) that stores the latest L1 block hash and state
root per L2 block, allowing L2 contracts to read recent L1 state. The second is
`AnchorStateManager`, a multi-rollup registry that validates execution proofs
before advancing any rollup's anchor state, with lifecycle controls (activate,
deactivate, prune stale entries).

The ring buffer holds `RingBufferSize=8191` entries; entries older than the
window are evicted. `AnchorStateManager` requires a non-empty proof on every
state update, verifying it via a SHA-256 commitment before advancing the stored
root.

## Functionality

**`Contract` (anchor.go)**

- `NewContract() *Contract`
- `UpdateState(newState rollup.AnchorState) error` — advances ring buffer; block number must increase
- `GetByNumber(blockNumber) (Entry, bool)` — retrieves entry if within the 8191-block window
- `ProcessAnchorData(data []byte) error` — decodes 80-byte wire format and calls `UpdateState`
- `UpdateAfterExecute(output, blockNumber, timestamp) error` — advances state from EXECUTE output
- `EncodeAnchorData(state) []byte` — encodes `AnchorState` to 80-byte wire format

**`AnchorStateManager` (state.go)**

- `NewAnchorStateManager() *AnchorStateManager`
- `RegisterAnchor(rollupID, meta AnchorMetadata) error`
- `UpdateAnchorState(rollupID, proof AnchorExecutionProof) error` — proof-verified update
- `GetAnchorState(rollupID) (*ManagedAnchorState, error)`
- `ValidateStateTransition(old, new, proof) error`
- `DeactivateAnchor(rollupID) / ActivateAnchor(rollupID) error`
- `PruneStaleAnchors(maxAge uint64) int` — removes inactive anchors older than maxAge seconds
- `AnchorCount() / ActiveCount() int`

**Test helper** — `MakeValidAnchorProof(currentRoot, newRoot, blockNum, timestamp)`

**Parent package:** [rollup](../)
