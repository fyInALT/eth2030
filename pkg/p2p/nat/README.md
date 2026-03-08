# Package nat

NAT traversal: UPnP and NAT-PMP device detection, port mapping lifecycle management, external IP discovery via STUN, and Prometheus metrics export.

## Overview

The `nat` package provides `NATTrav`, a NAT traversal manager that supports UPnP IGD and NAT-PMP/PCP gateways. It auto-detects available devices by probing SSDP (UPnP) and the well-known NAT-PMP port on the default gateway. Once a device is found, port mappings can be created and tracked; a background goroutine renews them before they expire. External IP detection falls back from the device's `GetExternalIP` to a minimal STUN Binding Request against a configurable server (default `stun.l.google.com:19302`). Prometheus-compatible counters track mapping successes, failures, renewals, and IP changes.

## Functionality

- `NATTrav` — main manager
  - `NewNATTrav(cfg NATTravConfig) *NATTrav`
  - `Start() error` / `Stop()`
  - `AutoDetect(timeout time.Duration) NATTravType` — probes UPnP then NAT-PMP
  - `AddPortMapping(proto string, intPort, extPort uint16, desc string) error`
  - `RemovePortMapping(proto string, intPort, extPort uint16) error`
  - `ExternalIP() (net.IP, error)` — device → STUN fallback
  - `DetectedType() NATTravType`
  - `ActiveMappings() []NATTravMapping`
  - `MappingCount() int`
  - `SetDevice(d NATTravDevice)`

- `NATTravDevice` interface — `DeviceType()`, `GetExternalIP()`, `MapPort()`, `UnmapPort()`
- `NATTravType` — `NATTravNone`, `NATTravUPnP`, `NATTravPMP`, `NATTravManual`
- `NATTravMapping` — `Protocol`, `InternalPort`, `ExternalPort`, `TTL`, `ExpiresAt`, `RenewCount`, `IsExpired()`
- `NATTravConfig` — `Device`, `MappingTTL` (default 30 min), `RenewBefore`, `STUNServer`, `ManualIP`

## Usage

```go
mgr := nat.NewNATTrav(nat.NATTravConfig{})
detected := mgr.AutoDetect(3 * time.Second)
if detected != nat.NATTravNone {
    mgr.Start()
    mgr.AddPortMapping("TCP", 30303, 30303, "eth2030")
}
extIP, _ := mgr.ExternalIP()
```

[← p2p](../README.md)
