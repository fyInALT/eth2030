// middleware.go re-exports HTTP middleware types and functions from rpc/middleware/.
// All HTTP middleware (CORS, Auth, Logging, Compression, RateLimit) is implemented
// in the rpc/middleware sub-package; this file provides backward-compatible aliases.
package rpc

import (
	"net/http"

	rcpmiddleware "github.com/eth2030/eth2030/rpc/middleware"
)

// HTTP middleware type aliases.
type (
	HTTPMiddleware  = rcpmiddleware.HTTPMiddleware
	CORSConfig      = rcpmiddleware.CORSConfig
	AuthConfig      = rcpmiddleware.AuthConfig
	LogEntry        = rcpmiddleware.LogEntry
	LogStore        = rcpmiddleware.LogStore
	RateLimitConfig = rcpmiddleware.RateLimitConfig
)

// MiddlewareChain composes multiple middleware into a single handler chain.
func MiddlewareChain(handler http.Handler, middlewares ...HTTPMiddleware) http.Handler {
	return rcpmiddleware.MiddlewareChain(handler, middlewares...)
}

// HTTP middleware constructors re-exported.
var (
	DefaultCORSConfig     = rcpmiddleware.DefaultCORSConfig
	CORSMiddleware        = rcpmiddleware.CORSMiddleware
	AuthMiddleware        = rcpmiddleware.AuthMiddleware
	NewLogStore           = rcpmiddleware.NewLogStore
	LoggingMiddleware     = rcpmiddleware.LoggingMiddleware
	CompressionMiddleware = rcpmiddleware.CompressionMiddleware
	RateLimitMiddleware   = rcpmiddleware.RateLimitMiddleware
)
