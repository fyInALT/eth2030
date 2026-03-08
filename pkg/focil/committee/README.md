# focil/committee — FOCIL inclusion-list committee selection, tracking, and voting

## Overview

Package `committee` manages the validator committee for Fork-Choice Enforced Inclusion Lists (EIP-7805). It covers three concerns: `CommitteeTracker` stores the canonical committee per slot and records which members have submitted their inclusion lists to determine quorum; `CommitteeSelector` deterministically selects committees from a validator set using RANDAO-seeded shuffling with epoch-level reseeding and a look-ahead cache; `CommitteeVoting` derives epoch-aware voting seeds, tracks submissions, and evaluates quorum with per-submitter resolution.

All three components are thread-safe.

## Functionality

**Constants**
- `IL_COMMITTEE_SIZE = 16` — target committee size per EIP-7805
- Quorum threshold: 2/3 of committee size (11 of 16)

**CommitteeTracker** (`committee_tracker.go`)
- `SlotCommittee{Slot uint64, Members []uint64, Root Hash}`
- `QuorumStatus{CommitteeSize, ListsReceived, QuorumThreshold int, QuorumReached bool, SubmittedBy []uint64}`
- `CommitteeDuty{Slot uint64, ValidatorIndex uint64, CommitteePosition int}`
- `NewCommitteeTracker() *CommitteeTracker`
- `GetCommittee(slot uint64) (*SlotCommittee, bool)`
- `IsCommitteeMember(slot uint64, validatorIndex uint64) bool`
- `GetDuty(slot uint64, validatorIndex uint64) (*CommitteeDuty, bool)`
- `RecordList(slot uint64, validatorIndex uint64) error`
- `GetQuorumStatus(slot uint64) (*QuorumStatus, bool)` / `CheckQuorum(slot uint64) bool`
- `PruneBefore(slot uint64) int`
- Selection uses Keccak-256(slot)-seeded hash-chain shuffle

**CommitteeSelector** (`committee_selection.go`)
- `CommitteeSelectionConfig{CommitteeSize int, SlotsPerEpoch=32, FallbackSize=4, MaxLookAhead=32}`
- `SelectionProof{ValidatorIndex, Slot uint64, CommitteePosition int, Commitment Hash}`
- `RotationRecord{ValidatorIndex, AssignmentCount uint64, LastAssignedSlot uint64, Slots []uint64}`
- `NewCommitteeSelector(validators []uint64, cfg CommitteeSelectionConfig) *CommitteeSelector`
- `SetEpochSeed(epoch uint64, seed Hash)`
- `SelectCommittee(slot uint64) ([]uint64, error)` — cached; derives seed from epoch RANDAO
- `SelectFallback(slot uint64) []uint64` — emergency smaller committee
- `ComputeLookAhead(fromSlot uint64, n int) ([][]uint64, error)`
- `GenerateSelectionProof(validatorIndex, slot uint64) (*SelectionProof, error)`
- `VerifySelectionProof(proof *SelectionProof) bool`
- `GetRotationRecords() []RotationRecord`
- `PruneCacheBefore(slot uint64)`

**CommitteeVoting** (`committee_voting.go`)
- `SubmissionRecord{ValidatorIndex, Slot uint64, ILHash Hash, Timestamp time.Time}`
- Seed derivation: `computeVotingSeed = Keccak-256(baseSeed || epoch || slot)`
- `NewCommitteeVoting(baseSeed Hash, cfg CommitteeSelectionConfig) *CommitteeVoting`
- `ComputeILCommittee(slot uint64) ([]uint64, error)`
- `IsILCommitteeMember(slot uint64, validatorIndex uint64) (bool, error)` — binary search
- `TrackSubmission(slot uint64, validatorIndex uint64, ilHash Hash) error`
- `GetMissingSubmitters(slot uint64) ([]uint64, error)`
- `GetSubmissions(slot uint64) []SubmissionRecord`
- `CommitteeQuorum(slot uint64) (bool, error)` / `QuorumDetail(slot uint64) (*QuorumStatus, error)`
- `SlotToEpoch(slot uint64) uint64`
- `PruneBefore(slot uint64)`

Parent package: [`focil`](../README.md)
