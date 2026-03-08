# light/bls - BLS sync committee signature verifier

## Overview

Package `bls` implements BLS12-381 aggregate signature verification for the light
client's sync committee validation. It replaces a simplified Keccak256-based check
with real `FastAggregateVerify` semantics, matching the beacon chain's requirement
that sync committee members sign the block root with their BLS keys and submit a
single aggregate signature.

The verifier enforces a 2/3 quorum: at least two-thirds of the 512-member sync
committee must participate before the aggregate signature is accepted. Participation
is encoded as a compact bitfield where each bit corresponds to one committee member.

## Functionality

**Types**

- `SyncCommitteeBLSVerifier` - stateful verifier tracking participation rate and
  cumulative verified/failed counts.

**Constructors**

- `NewSyncCommitteeBLSVerifier() *SyncCommitteeBLSVerifier` - production verifier
  with committee size 512.
- `NewSyncCommitteeBLSVerifierWithSize(size int) *SyncCommitteeBLSVerifier` -
  custom size for testing.

**Core method**

- `VerifySyncCommitteeSignature(committee [][48]byte, participationBits []byte, msg []byte, sig [96]byte) bool` -
  decodes the participation bitfield, checks the 2/3 quorum, and calls
  `crypto.FastAggregateVerify` on the participating public keys.

**Accessors**

- `ParticipationRate() float64`, `TotalVerified() uint64`, `TotalFailed() uint64`,
  `CommitteeSize() int`

**Utilities (also exported)**

- `CountParticipants(participationBits []byte, committeeSize int) int`
- `MakeParticipationBits(committeeSize, participants int) []byte`
- `MakeBLSTestCommittee(size int) ([][48]byte, []*[32]byte)`
- `SignSyncCommitteeBLS(secrets []*[32]byte, participationBits []byte, msg []byte) [96]byte`

**Constants**

- `SyncCommitteeSize = 512`, `MinQuorumNumerator = 2`, `MinQuorumDenominator = 3`

**Errors**

`ErrBLSInvalidPubkey`, `ErrBLSInvalidSignature`, `ErrBLSVerifyFailed`,
`ErrBLSNoParticipants`, `ErrBLSQuorumNotMet`

Parent package: [`light`](../)
