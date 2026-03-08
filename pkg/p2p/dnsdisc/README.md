# Package dnsdisc

EIP-1459 DNS-based Ethereum node discovery — resolves node records from DNS TXT trees without connecting to a DHT.

## Overview

The `dnsdisc` package implements the DNS node list protocol defined in EIP-1459. A `DNSClient` resolves the `enrtree-root:v1` TXT record for a configured domain, then recursively follows `enrtree-branch` and `enrtree://` link records to accumulate a set of ENR node records. Discovered nodes are cached in memory; `RandomNodes` draws a random sample for bootstrapping. The `Resolver` interface allows injection of a fake DNS backend in tests.

## Functionality

- `DNSClient` — DNS discovery client
  - `NewDNSClient(config DNSConfig) *DNSClient`
  - `NewDNSClientWithResolver(config DNSConfig, resolver Resolver) *DNSClient`
  - `Resolve(domain string) ([]*enode.Node, error)` — full tree walk and caching
  - `RefreshCache() error`
  - `RandomNodes(count int) []*enode.Node`
  - `CachedNodes() []*enode.Node`

- `ParseTreeRoot(txt string) (*TreeRoot, error)` — parses `enrtree-root:v1 e=... l=... seq=... sig=...`
- `ParseTreeLink(txt string) (*TreeLink, error)` — parses `enrtree://<pubkey>@<domain>`

- `Resolver` interface — `LookupTXT(domain string) ([]string, error)`

- `DNSConfig` — `Domain`, `PublicKey`, `RefreshInterval` (default 30 min)
- `TreeRoot` — `ERoot`, `LRoot`, `Seq`, `Sig`
- `TreeLink` — `Domain`, `PublicKey`

## Usage

```go
client := dnsdisc.NewDNSClient(dnsdisc.DNSConfig{
    Domain:          "all.mainnet.ethdisco.net",
    RefreshInterval: 30 * time.Minute,
})
nodes, err := client.Resolve("all.mainnet.ethdisco.net")
// use nodes as bootstrap peers
bootstrapPeers := client.RandomNodes(10)
```

[← p2p](../README.md)
