# kps

Key Pair Sharing (KPS) for distributed validator key management using Shamir's Secret Sharing.

## Overview

Package `kps` enables validators to split their private keys into shares
distributed among a group of key holders. A configurable threshold of shares
is required to reconstruct the original key, providing fault tolerance and
preventing single points of failure. Splitting and reconstruction operate over
GF(256) — the same finite field used by AES — so no large-integer arithmetic
is required.

`KPSManager` is the top-level coordinator. It stores `KeyGroup` objects that
describe member sets, generates `KPSKeyPair` values with associated
`KeyShare` slices, and supports periodic key rotation via `RotateKeys`.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `KPSConfig` | `DefaultThreshold`, `MaxGroupSize`, `KeyRotationInterval` |
| `KPSManager` | Thread-safe manager: groups, key pairs, config |
| `KeyGroup` | Named set of member addresses with a threshold |
| `KPSKeyPair` | Generated public key + threshold share list |
| `KeyShare` | One share: index, 32-byte data slice, group ID |

### Functions / methods

| Name | Description |
|------|-------------|
| `DefaultKPSConfig() KPSConfig` | Threshold=2, MaxGroup=10, Rotation=256 epochs |
| `NewKPSManager(config) *KPSManager` | Create a manager |
| `(*KPSManager).GenerateKeyPair() (*KPSKeyPair, error)` | Random key + shares via default config |
| `SplitKey(privateKey, threshold, totalShares) ([]*KeyShare, error)` | Shamir split over GF(256) |
| `RecombineKey(shares) ([]byte, error)` | Lagrange interpolation at x=0 |
| `VerifyKeyShare(share, publicKey) bool` | Structural share validation |
| `(*KPSManager).RegisterGroup(group)` | Register a key group |
| `(*KPSManager).RotateKeys(groupID) (*KPSKeyPair, error)` | Replace key shares for a group |
| `NewKeyGroup(groupID, threshold, total) *KeyGroup` | Create a key group |
| `(*KeyGroup).AddMember(addr) error` | Add a member (up to `totalMembers`) |
| `(*KeyGroup).RemoveMember(addr) error` | Remove a member |

### Errors

`ErrKPSInvalidThreshold`, `ErrKPSInvalidShares`, `ErrKPSInsufficientShares`, `ErrKPSDuplicateShare`, `ErrKPSInvalidShareData`, `ErrKPSGroupFull`, `ErrKPSMemberExists`, `ErrKPSMemberNotFound`, `ErrKPSGroupNotFound`, `ErrKPSKeyGenFailed`

## Usage

```go
import "github.com/eth2030/eth2030/consensus/kps"

mgr := kps.NewKPSManager(kps.DefaultKPSConfig())

// Split a private key into 10 shares, threshold 3.
shares, err := kps.SplitKey(privKey, 3, 10)

// Reconstruct from any 3 shares.
recovered, err := kps.RecombineKey(shares[:3])
```

[← consensus](../README.md)
