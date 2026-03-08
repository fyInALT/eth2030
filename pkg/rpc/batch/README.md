# batch — JSON-RPC batch request processing

[← rpc](../README.md)

## Overview

Package `rpcbatch` provides types, validation, and utilities for handling
JSON-RPC 2.0 batch requests (RFC: a JSON array of request objects). It handles
parsing, size enforcement, per-item validation, splitting of large batches, and
accumulation of subscription notifications into outbound batches.

A companion `handler.go` provides `BatchHandler`, an HTTP handler that
integrates batch processing with a `RequestHandler` and optional concurrency.

## Functionality

**Constants**

- `MaxBatchSize = 100` — maximum requests per batch
- `DefaultParallelism = 16` — goroutines for parallel execution
- `MaxNotificationBatchSize = 50` — notification flush limit
- `DefaultBatchTimeout = 5000` ms — per-item timeout

**Core types**

- `BatchRequest` — single item inside a batch (`jsonrpc`, `method`, `params`, `id`)
- `BatchStats` / `BatchStatsSnapshot` — atomic counters: `TotalBatches`, `TotalRequests`, `TotalErrors`, `LargestBatch`, `ParallelBatches`
- `BatchValidator` — created with `NewBatchValidator(maxSize int)`; methods `Validate([]BatchRequest) []error` and `ValidateBatchSize(count int) error`
- `NotificationBatch` — created with `NewNotificationBatch(limit int)`; methods `Add(notification interface{}) []byte`, `Flush() []byte`, `Len() int`

**Functions**

- `IsBatchRequest(body []byte) bool` — detects JSON arrays
- `ParseBatchRequests(body []byte) ([]BatchRequest, error)` — decode + size guard
- `SummarizeBatch([]BatchRequest) BatchRequestSummary` — extract method names for logging
- `SplitBatch([]BatchRequest, chunkSize int) [][]BatchRequest` — chunk a batch
- `TrimWhitespace([]byte) []byte` — fast leading-whitespace strip

**Errors**: `ErrBatchEmpty`, `ErrBatchTooLarge`, `ErrNotBatch`
