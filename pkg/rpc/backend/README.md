# backend — JSON-RPC backend interface

[← rpc](../README.md)

## Overview

Package `rpcbackend` defines the `Backend` interface that decouples the
JSON-RPC API layer from the concrete chain implementation. It follows the
go-ethereum `ethapi.Backend` pattern, giving API handlers a stable surface over
chain data, state, the transaction pool, and EVM execution.

A companion `services.go` file provides optional service-accessor interfaces
(`TxPoolService`, `FilterService`, etc.) that concrete backends may implement
to expose additional subsystems.

## Functionality

**`Backend` interface** (in `backend.go`)

| Method | Purpose |
|---|---|
| `HeaderByNumber(BlockNumber) *types.Header` | Fetch header by block number or tag |
| `HeaderByHash(Hash) *types.Header` | Fetch header by hash |
| `BlockByNumber / BlockByHash` | Fetch full blocks |
| `CurrentHeader() *types.Header` | Chain head |
| `ChainID() *big.Int` | Network chain ID |
| `StateAt(root Hash) (state.StateDB, error)` | Open state at a given root |
| `SendTransaction(*types.Transaction) error` | Submit transaction to pool |
| `GetTransaction(Hash) (*Transaction, blockNum, index)` | Look up transaction |
| `SuggestGasPrice() *big.Int` | EIP-1559-aware gas price hint |
| `GetReceipts / GetLogs / GetBlockReceipts` | Receipt and log retrieval |
| `GetProof(addr, keys, BlockNumber) (*trie.AccountProof, error)` | Merkle proof for `eth_getProof` |
| `EVMCall(from, to, data, gas, value, BlockNumber) ([]byte, uint64, error)` | Stateless EVM call |
| `TraceTransaction(Hash) (*vm.StructLogTracer, error)` | Opcode-level trace |
| `HistoryOldestBlock() uint64` | EIP-4444 history expiry boundary |
