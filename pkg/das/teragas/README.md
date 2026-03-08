# das/teragas — Teragas L2 data pipeline with bandwidth enforcement

Implements the L2 data ingestion pipeline targeting 1 Gbyte/sec throughput (the
"Teragas L2" North Star). Provides bandwidth-rate limiting, RLE compression,
chunking, and reassembly as composable pipeline stages.

## Overview

`TeragasPipeline` is the central type. It connects a `BackpressureChannel`
input to an output channel through a sequence of `PipelineStage` processors
selected at construction time. Built-in stages are `BandwidthGate` (delegates
to `BandwidthEnforcer` for per-chain token-bucket rate limiting),
`CompressionStage` (run-length encoding), and `ChunkingStage` (splits large
payloads into ≤ `MaxChunkSize` pieces). A `ReassemblyStage` reverses chunking
on the consumer side.

`BandwidthEnforcer` (`bandwidth_enforcer.go`) and `BandwidthController`
(`bandwidth_controller.go`) provide the per-chain byte-rate accounting.
`ThroughputManager` (`throughput_manager.go`) manages per-chain bandwidth
allocation across multiple L2 chains.

## Functionality

**Interfaces**
- `DataProducer` — `Submit(chainID, data)`
- `DataConsumer` — `Receive(ctx) (*TPDataPacket, error)`
- `PipelineStage` — `Process(*TPDataPacket) (*TPDataPacket, error)` + `Name()`

**Types**
- `TPDataPacket` — `ChainID`, `Data`, `ChunkIndex`, `TotalChunks`, `Compressed`, `Timestamp`
- `TPConfig` — `MaxChunkSize`, `ChannelBufferSize`, `Policy`, `BandwidthEnforcer`, …
- `TeragasPipeline` — full pipeline with start/stop lifecycle and metrics
- `TPMetricsSnapshot` — `BytesIn/Out`, `PacketsDropped`, `CompressionSaved`, `AvgLatencyMs()`
- `BackpressureChannel` — bounded channel with `DropOldest` / `BlockOnFull` policy
- `BandwidthGate`, `CompressionStage`, `ChunkingStage`, `ReassemblyStage`

**Functions**
- `DefaultTPConfig() *TPConfig`
- `NewTeragasPipeline(config) (*TeragasPipeline, error)`
- `(*TeragasPipeline).Start() / Stop()`
- `(*TeragasPipeline).Submit(chainID, data) error`
- `(*TeragasPipeline).Receive(ctx) (*TPDataPacket, error)`
- `(*TeragasPipeline).Metrics() *TPMetricsSnapshot`
- `ChunkData(data, maxSize) [][]byte`
- `SimpleDecompress(data) ([]byte, error)`

## Usage

```go
cfg := teragas.DefaultTPConfig()
pipe, _ := teragas.NewTeragasPipeline(cfg)
pipe.Start()
defer pipe.Stop()

pipe.Submit(chainID, l2Payload)
pkt, _ := pipe.Receive(ctx)
```

[← das](../README.md)
