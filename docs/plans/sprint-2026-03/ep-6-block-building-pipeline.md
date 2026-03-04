> [← Back to Sprint Index](README.md)

# EPIC 6 — Block Building Pipeline

**Goal**: Complete the remaining block building pipeline gaps identified in the Vitalik analysis: real mixnet transport and passive serverless order-matching research.

---

## US-BB-1: Real Mixnet Integration (Tor/Nym)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **privacy-sensitive user**, I want ETH2030's transaction ingress to optionally route transactions through a real anonymizing network (Tor or Nym), so that sender IP addresses are not visible to block builders even in the Ethereum network layer.

**Priority**: Medium | **Story Points**: 13 | **Sprint Target**: Sprint 5

### Tasks

#### Task BB-1.1 — Define `ExternalMixnetTransport` interface
- **Description**: In `pkg/p2p/anonymous_transport.go`, add `ExternalMixnetTransport` interface with `SendViaExternalMixnet(tx []byte, endpoint string) error`. Add `MixnetTransportMode` enum: `Simulated` (current), `TorSocks5`, `NymSocks5`. Add CLI flag `--mixnet=simulated|tor|nym` (default `simulated`).
- **Estimated Effort**: 2 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Config test: mode enum parsed correctly. Interface present but implementations gated behind build tag.
- **Definition of Done**: Interface defined. CLI flag present. No regression.

#### Task BB-1.2 — Implement Tor SOCKS5 transport
- **Description**: Create `pkg/p2p/tor_transport.go` implementing `ExternalMixnetTransport` using SOCKS5 proxy at `127.0.0.1:9050` (Tor default). Submit transactions as HTTP POST to the node's own RPC via Tor, ensuring sender IP is obscured at the network layer.
- **Estimated Effort**: 5 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Integration test (requires Tor daemon): connect via SOCKS5, submit a transaction, verify receipt. Mock test without Tor: `MockTorTransport` that simulates SOCKS5 protocol. `go test ./p2p/... -run TestTorTransport`.
- **Definition of Done**: SOCKS5 connection working. Mock test passes without Tor daemon. Real Tor test passes in CI with Tor installed.

#### Task BB-1.3 — Transport selection and fallback
- **Description**: In `pkg/p2p/anonymous_transport.go` `TransportManager`, add priority: `Tor > Nym > Simulated`. If Tor is not reachable within 500ms, fall back to Nym; if Nym unavailable, fall back to Simulated. Log transport selection at startup.
- **Estimated Effort**: 3 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Test: Tor unavailable → Nym attempted → Nym unavailable → Simulated used. `go test ./p2p/... -run TestTransportFallback`.
- **Definition of Done**: Fallback chain working. Startup log shows selected transport. Tests pass.

#### Task BB-1.4 — Kohaku interface alignment
- **Description**: Update `TransportManager` API to align with the kohaku protocol interface (once spec is published at the referenced `@ncsgy` repo). Add `KohakuCompatible bool` flag to `TransportConfig` — when `true`, use kohaku wire format for transport control messages.
- **Estimated Effort**: 3 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Config test: `KohakuCompatible=true` → cohaku format messages sent. `go test ./p2p/... -run TestKohakuCompatibility`.
- **Definition of Done**: Kohaku flag present. Wire format conditional. Tests pass. TODO note: update once kohaku spec finalizes.

---

## US-BB-2: Distributed Block Building — New Local Tx Type (Research Spike)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **protocol researcher**, I want a documented design for a "less global" transaction type that is cheaper and more amenable to distributed building (as described by Vitalik), so that the ETH2030 team can evaluate the design before it becomes an EIP.

**Priority**: Low | **Story Points**: 5 | **Sprint Target**: Sprint 6

### Tasks

#### Task BB-2.1 — Research spike: define "local tx" semantics
- **Description**: Write `docs/research/local-tx-design-2026-03.md` exploring: (1) what makes a tx "less global" (limited state access, predeclared BAL), (2) how local txs could be built by distributed builders without ordering coordination, (3) gas discount model (50–80% cheaper), (4) mempool routing (sharded mempool per sender prefix).
- **Estimated Effort**: 3 SP
- **Assignee**: Protocol Researcher
- **Testing Method**: Document peer-reviewed by senior engineer and consensus researcher. No code.
- **Definition of Done**: Design document written and reviewed. Key trade-offs documented. EIP sketch (optional).

#### Task BB-2.2 — Prototype `LocalTx` type (gated behind flag)
- **Description**: Create `pkg/core/types/tx_local.go` with `LocalTx` type (tx type `0x08`): must declare BAL in tx body, state access limited to declared keys, gas price discount configurable. Gated behind `--experimental-local-tx` flag. This is a proof-of-concept, not production-ready.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (core/types)
- **Testing Method**: Unit test: `LocalTx` struct construction, BAL validation (declared vs actual access), gas discount calculation. `go test ./core/types/... -run TestLocalTx`.
- **Definition of Done**: `LocalTx` type defined. BAL check works. Gas discount applied. Gated behind experimental flag. Design document (Task BB-2.1) complete first.
