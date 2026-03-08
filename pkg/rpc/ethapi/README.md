# ethapi — eth/net/web3/txpool namespace JSON-RPC methods

[← rpc](../README.md)

## Overview

Package `ethapi` is the core Ethereum JSON-RPC API handler. It implements the
`eth_`, `net_`, `web3_`, and `txpool_` namespace methods, covering chain
queries, state access, transaction submission, EVM calls, log filtering,
subscription management, and Merkle proof generation. The package follows the
go-ethereum `ethapi` pattern: all chain access goes through a `Backend`
interface, keeping the API layer independently testable.

Supporting files add state inspection (`state_api.go`), `eth_call` /
`eth_estimateGas` (`calls.go`), proof generation (`proof.go`), subscription
wiring (`subscription_service.go`), direct-dispatch helpers (`direct_api.go`),
and extended methods (`extended_api.go`).

## Functionality

**Core types**

- `EthAPI` — constructed with `NewEthAPI(backend Backend, subs SubscriptionService)`; dispatches via `HandleRequest(req *Request) *Response`
- `Backend` — interface mirroring `rpcbackend.Backend` plus filter and subscription hooks
- `TxPoolAPI` — constructed with `NewTxPoolAPI(backend Backend)`; handles `txpool_content`, `txpool_status`, `txpool_inspect`
- `SubscriptionService` — interface for `eth_subscribe` / `eth_unsubscribe` wiring

**Selected method groups**

| Namespace | Methods |
|---|---|
| Chain | `eth_chainId`, `eth_blockNumber`, `eth_getBlockByNumber`, `eth_getBlockByHash`, `eth_getBlockTransactionCount*`, `eth_getUncleCount*` |
| State | `eth_getBalance`, `eth_getTransactionCount`, `eth_getCode`, `eth_getStorageAt` |
| Transactions | `eth_sendRawTransaction`, `eth_getTransactionByHash`, `eth_getTransactionByBlock*`, `eth_getTransactionReceipt` |
| Execution | `eth_call`, `eth_estimateGas` |
| Fees | `eth_gasPrice`, `eth_maxPriorityFeePerGas`, `eth_feeHistory` |
| Logs | `eth_getLogs` |
| Proofs | `eth_getProof` |
| Subscriptions | `eth_subscribe`, `eth_unsubscribe` |
| Net/Web3 | `net_version`, `net_peerCount`, `web3_clientVersion`, `web3_sha3` |
| TxPool | `txpool_content`, `txpool_status`, `txpool_inspect` |
