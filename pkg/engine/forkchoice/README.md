# engine/forkchoice — Fork-choice state management and payload coordination

## Overview

Package `forkchoice` manages the Engine API fork-choice lifecycle: tracking the chain head, safe block, and finalized block; processing `engine_forkchoiceUpdated` calls; and coordinating payload ID allocation for block building. It enforces LMD-GHOST invariants including proposer boost, reorg detection, and checkpoint regression guards.

Three layers compose the package. `ForkchoiceEngine` is the top-level handler wiring together block lookup and state updates. `ForkchoiceStateManager` maintains checkpoint history, proposer-boost state, and reorg listeners. `ForkchoiceTracker` aggregates a `HeadChain`, `FCUHistory`, `ConflictDetector`, `PayloadIDAllocator`, and `ReorgTracker` for fine-grained per-update bookkeeping.

## Functionality

**Types**
- `ForkchoiceState{HeadBlockHash, SafeBlockHash, FinalizedBlockHash Hash}` — three-hash fork-choice triple
- `ForkchoiceResponse{PayloadStatus, PayloadID}` — response returned to the CL
- `Checkpoint{Epoch uint64, Root Hash}` — justified / finalized checkpoint
- `ProposerBoost{Slot, BlockRoot, BoostWeight}` — active proposer-boost record
- `ReorgEvent{OldHead, NewHead, CommonAncestor Hash, Depth uint64}` — reorg notification
- `ReorgListener func(ReorgEvent)` — callback registered on `ForkchoiceStateManager`
- `BlockInfo{Hash, ParentHash Hash, Number, Slot uint64}` — lightweight block descriptor
- `FCURecord` — per-update debug history entry stored by `FCUHistory`
- `BlockLookup` — interface for resolving block hashes to `BlockInfo`

**ForkchoiceEngine**
- `NewForkchoiceEngine(lookup BlockLookup) *ForkchoiceEngine`
- `ProcessForkchoiceUpdate(state ForkchoiceState, attrs *PayloadAttributes) (*ForkchoiceResponse, error)`
- `ValidateForkchoiceState(state ForkchoiceState) error`
- `UpdateHead/UpdateSafe/UpdateFinalized(hash Hash)`
- `HeadBlock/SafeBlock/FinalizedBlock() Hash`
- `ShouldBuildPayload(attrs *PayloadAttributes) bool`
- `HasPayload(id PayloadID) bool`
- `Stats() map[string]any`

**ForkchoiceStateManager**
- `NewForkchoiceStateManager() *ForkchoiceStateManager`
- `AddBlock(info BlockInfo)`
- `ProcessForkchoiceUpdate(state ForkchoiceState) error`
- `SetProposerBoost(boost ProposerBoost)` / `ClearProposerBoost()` / `ProposerBoostFor(slot uint64) *ProposerBoost`
- `OnReorg(listener ReorgListener)`
- `PruneBeforeNumber(number uint64) int`
- `GetState() ForkchoiceState`

**ForkchoiceTracker and sub-components**
- `NewForkchoiceTracker() *ForkchoiceTracker`
- `ConflictDetector.Check(state ForkchoiceState) error` — detects finalized hash regression
- `PayloadIDAllocator.Allocate(slot uint64) PayloadID` — collision-resistant 8-byte ID
- `ReorgTracker.ProcessHead(hash Hash, parentHash Hash) (depth uint64, isReorg bool)`
- `FCUHistory.Record(FCURecord)` / `Last(n int) []FCURecord`
- `HeadChain.SetHead(info BlockInfo)` / `Tip() BlockInfo`

**Helper**
- `generatePayloadID(slot uint64) (PayloadID, error)` — reads crypto/rand for uniqueness

Parent package: [`engine`](../README.md)
