# registry — RPC method registry with middleware

[← rpc](../README.md)

## Overview

Package `registry` provides a thread-safe registry for JSON-RPC method
handlers with a composable middleware chain. Methods are keyed by their full
`namespace_methodName` string and carry metadata (description, expected param
count, deprecation flag, namespace). Middleware is applied in registration
order, with the first-added middleware as the outermost wrapper.

## Functionality

**Types**

- `MethodHandler` — `func(params []interface{}) (interface{}, error)`
- `Middleware` — `func(method string, params []interface{}, next MethodHandler) (interface{}, error)`
- `MethodInfo` — `Name`, `Handler MethodHandler`, `Description string`, `ParamCount int` (-1 = variadic), `Deprecated bool`, `Namespace string`
- `MethodRegistry` — constructed with `NewMethodRegistry()`

**`MethodRegistry` methods**

| Method | Description |
|---|---|
| `Register(info MethodInfo) error` | Add a method; returns `ErrDuplicateMethod` on conflict |
| `RegisterBatch([]MethodInfo) error` | Bulk registration; stops on first error |
| `Unregister(name string) bool` | Remove a method; returns false if not found |
| `Call(method string, params []interface{}) (interface{}, error)` | Dispatch through middleware chain |
| `Methods() []string` | Sorted list of all method names |
| `MethodsByNamespace(ns string) []string` | Methods filtered by namespace |
| `HasMethod(name string) bool` | Existence check |
| `GetMethodInfo(name string) (MethodInfo, bool)` | Retrieve metadata |
| `AddMiddleware(mw Middleware)` | Append middleware to chain |
| `MethodCount() int` | Number of registered methods |

**Helpers**

- `NamespaceFromMethod(method string) string` — extracts `"eth"` from `"eth_blockNumber"`

**Errors**: `ErrMethodNotFound`, `ErrDuplicateMethod`, `ErrInvalidParams`
