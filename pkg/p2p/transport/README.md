# Package transport

Anonymous transaction transports: Tor SOCKS5, Nym SOCKS5, and a simulated in-process mixnet, with a `TransportManager` for auto-selection and unified submission.

## Overview

The `transport` package implements privacy-preserving transaction submission to hide the sender's IP address from the P2P network. Three transport backends are provided: `TorTransport` (routes via a Tor SOCKS5 proxy), `NymTransport` (routes via the Nym mixnet SOCKS5 proxy), and `MixnetTransport` (a deterministic simulated mixnet for testing). `TransportManager` can hold multiple registered transports and fan a transaction out to all of them. `SelectBestTransport` probes proxy availability and chooses Tor > Nym > Simulated. `FlashnetTransport` provides a fast-path low-latency submission channel.

Control messages support two wire formats: plain JSON (`{"type":"control","msg":"..."}`) and a Kohaku binary prefix (4-byte big-endian length + UTF-8 payload).

## Functionality

### AnonymousTransport interface

```go
type AnonymousTransport interface {
    Name() string
    Submit(tx *types.Transaction) error
    Receive() <-chan *types.Transaction
    Start() error
    Stop() error
}
```

### TransportManager

- `NewTransportManager() *TransportManager`
- `NewTransportManagerWithConfig(cfg TransportConfig) *TransportManager`
- `SelectBestTransport() MixnetTransportMode` — Tor > Nym > Simulated probe
- `RegisterTransport(t AnonymousTransport) error`
- `UnregisterTransport(name string) error`
- `SubmitAnonymous(tx *types.Transaction) (int, []error)` — fans out to all transports
- `StartAll() []error` / `StopAll() []error`
- `GetStats() []TransportStats`
- `SelectedMode() MixnetTransportMode` / `Config() TransportConfig`

### Mode selection

- `ParseMixnetMode(s string) (MixnetTransportMode, error)` — `"simulated"`, `"tor"`, `"nym"`
- `MixnetTransportMode` — `ModeSimulated`, `ModeTorSocks5`, `ModeNymSocks5`
- `ProbeProxy(addr string, timeout time.Duration) bool`
- `FormatControlMessage(msg string, kohaku bool) []byte`

### Config

`TransportConfig` — `Mode`, `TorProxyAddr` (default `127.0.0.1:9050`), `NymProxyAddr` (default `127.0.0.1:1080`), `RPCEndpoint`, `DialTimeout`, `KohakuCompatible`

## Usage

```go
cfg := transport.DefaultTransportConfig()
mgr := transport.NewTransportManagerWithConfig(cfg)
mgr.SelectBestTransport() // auto-picks Tor/Nym/Simulated

mgr.RegisterTransport(transport.NewMixnetTransport())
mgr.StartAll()

n, errs := mgr.SubmitAnonymous(tx)
mgr.StopAll()
```

[← p2p](../README.md)
