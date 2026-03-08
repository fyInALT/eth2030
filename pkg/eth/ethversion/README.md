# eth/ethversion — ETH wire protocol version negotiation

## Overview

Package `ethversion` handles ETH wire protocol version management for the P2P layer. It defines the canonical ETH/66, ETH/67, and ETH/68 version descriptors and provides a `VersionManager` that negotiates the highest mutually supported version when two peers connect.

The manager also maintains a per-peer version registry, enabling higher-level protocol handlers to dispatch messages using the correct encoding for each connected peer without re-negotiating on every request.

## Functionality

**Types**
- `ProtocolVersion{Major, Minor, Patch uint8, Name string}` — named version descriptor

**Pre-defined versions**
- `ETH66Version` — `{Major: 66, Name: "eth/66"}`
- `ETH67Version` — `{Major: 67, Name: "eth/67"}`
- `ETH68Version` — `{Major: 68, Name: "eth/68"}`

**VersionManager**
- `NewVersionManager(supported []ProtocolVersion) *VersionManager` — stores supported versions sorted by major version descending
- `NegotiateVersion(peerVersions []ProtocolVersion) (*ProtocolVersion, error)` — returns the highest version present in both sets; error if no common version
- `IsSupported(version ProtocolVersion) bool`
- `RegisterPeer(peerID string, version ProtocolVersion)`
- `GetPeerVersion(peerID string) (*ProtocolVersion, bool)`
- `RemovePeer(peerID string)`
- `SupportedVersions() []ProtocolVersion` — returns a copy, highest first
- `HighestSupported() *ProtocolVersion`
- `PeerCount() int`

Parent package: [`eth`](../README.md)
