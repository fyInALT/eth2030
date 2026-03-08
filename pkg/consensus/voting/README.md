# voting

Voting round management and quorum tracking for single-slot finality.

## Overview

Package `voting` implements `VotingManager`, a thread-safe store of active
`VotingRound` objects keyed by slot. Each round is opened for a specific
`(slot, proposalHash)` pair with a configurable quorum threshold. Validators
cast `Vote` messages; duplicate voters are rejected. `FinalizeRound` closes
the round (regardless of vote count); callers are responsible for checking
quorum before finalizing. Expired non-finalized rounds can be evicted via
`ExpireRounds`.

This package is used by the SSF engine and the broader consensus pipeline to
track the per-slot vote accumulation phase before finality is declared.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `VotingManager` | Thread-safe multi-slot voting coordinator |
| `VotingManagerConfig` | `QuorumThreshold` (0–1), `MaxConcurrentRounds`, `VoteTimeout` |
| `VotingRound` | Per-slot state: proposal hash, votes map, threshold, finalized flag |
| `Vote` | `VoterID string`, slot, proposal hash, 96-byte signature, timestamp |

### Functions / methods

| Name | Description |
|------|-------------|
| `NewVotingManager(config) *VotingManager` | Create manager |
| `(*VotingManager).StartRound(slot, proposalHash, threshold) error` | Open a new voting round |
| `(*VotingManager).CastVote(vote) error` | Record a vote; error on duplicate or wrong proposal |
| `(*VotingManager).FinalizeRound(slot) error` | Mark round as finalized |
| `(*VotingManager).IsFinalized(slot) bool` | True if round is finalized |
| `(*VotingManager).GetVoteCount(slot) (int, error)` | Number of votes cast |
| `(*VotingManager).GetQuorum(slot) (float64, error)` | Current vote fraction |
| `(*VotingManager).GetRound(slot) (*VotingRound, error)` | Retrieve round snapshot |
| `(*VotingManager).ExpireRounds(beforeSlot) int` | Evict non-finalized rounds older than a slot |
| `(*VotingManager).ActiveRounds() int` | Total tracked rounds |

### Errors

`ErrVotingRoundExists`, `ErrVotingRoundNotFound`, `ErrVotingAlreadyVoted`, `ErrVotingRoundFinalized`, `ErrVotingWrongProposal`

## Usage

```go
import "github.com/eth2030/eth2030/consensus/voting"

vm := voting.NewVotingManager(voting.VotingManagerConfig{
    QuorumThreshold:     0.67,
    MaxConcurrentRounds: 16,
})
vm.StartRound(slot, blockRoot, 0.67)
vm.CastVote(&voting.Vote{VoterID: "v42", Slot: slot, ProposalHash: blockRoot})

count, _ := vm.GetVoteCount(slot)
if float64(count)/float64(totalValidators) >= 0.67 {
    vm.FinalizeRound(slot)
}
```

[← consensus](../README.md)
