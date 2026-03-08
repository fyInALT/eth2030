# server — HTTP JSON-RPC server

[← rpc](../README.md)

## Overview

Package `rpcserver` provides HTTP server implementations for the JSON-RPC
endpoint. `Server` is a minimal single-handler HTTP server. `ExtServer`
extends it with CORS enforcement, bearer-token authentication, request-size
limiting, built-in token-bucket rate limiting, JSON-RPC batch processing, and
graceful shutdown. Both accept a `RequestHandler` interface so they work with
any API dispatcher (`ethapi.EthAPI`, `adminapi.DispatchAPI`, etc.).

## Functionality

**Configuration**

- `ServerConfig` — `MaxRequestSize int64`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `CORSAllowOrigins []string`, `AuthSecret string`, `RateLimitPerSec int`, `MaxBatchSize int`, `ShutdownTimeout`
- `DefaultServerConfig()` — `MaxRequestSize=5 MiB`, timeouts `30s/30s/120s`, `RateLimitPerSec=100`, `MaxBatchSize=100`

**Interfaces**

- `RequestHandler` — `HandleRequest(req *rpctypes.Request) *rpctypes.Response`
- `AdminRequestHandler` — `HandleAdminRequest(req *rpctypes.Request) *rpctypes.Response`

**Types**

- `RateLimiter` — in-server token bucket; `NewRateLimiter(rps int) *RateLimiter`, `Allow() bool`, `SetRate(rps int)`
- `Server` — minimal HTTP server; `NewServer(addr string, handler RequestHandler)`, `Start() error`, `Stop()`
- `ExtServer` — feature-rich server; `NewExtServer(addr string, handler RequestHandler, config ServerConfig)`, `SetAdminHandler(AdminRequestHandler)`, `Start() error`, `Stop(ctx context.Context) error`, `Addr() string`

**Request lifecycle (ExtServer)**

1. CORS preflight handling
2. Bearer-token auth check (`Authorization: Bearer <secret>`)
3. Request-size limit enforcement
4. Rate limit check
5. Batch vs single request detection (`rpcbatch.IsBatchRequest`)
6. Dispatch to `RequestHandler` or `AdminRequestHandler` by method namespace
7. JSON response serialization
