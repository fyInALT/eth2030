// Package rpcserver provides HTTP JSON-RPC server implementations.
// Server handles single requests; ExtServer adds middleware, CORS, auth,
// rate limiting, batch processing, and graceful shutdown.
package rpcserver

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	rpcbatch "github.com/eth2030/eth2030/rpc/batch"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// AdminRequestHandler dispatches admin_ namespace JSON-RPC requests.
// *rpc.AdminDispatchAPI satisfies this interface.
type AdminRequestHandler interface {
	HandleAdminRequest(req *rpctypes.Request) *rpctypes.Response
}

// NetRequestHandler dispatches net_ namespace JSON-RPC requests.
// *netapi.API satisfies this interface.
type NetRequestHandler interface {
	HandleNetRequest(req *rpctypes.Request) *rpctypes.Response
}

// BeaconRequestHandler dispatches beacon_ namespace JSON-RPC requests.
// *beaconapi.BeaconAPI satisfies this interface.
type BeaconRequestHandler interface {
	HandleBeaconRequest(req *rpctypes.Request) *rpctypes.Response
}

// RequestHandler dispatches a single JSON-RPC request.
// *ethapi.EthAPI satisfies this interface.
type RequestHandler interface {
	HandleRequest(req *rpctypes.Request) *rpctypes.Response
}

// BackendConstructor creates a RequestHandler from a backend.
// Used by NewServer/NewExtServer constructors.
type BackendConstructor func(backend interface{}) RequestHandler

// Server errors.
var (
	ErrServerClosed    = errors.New("rpc server: closed")
	ErrServerStarted   = errors.New("rpc server: already started")
	ErrAuthFailed      = errors.New("rpc server: authentication failed")
	ErrRateLimited     = errors.New("rpc server: rate limited")
	ErrRequestTooLarge = errors.New("rpc server: request body too large")
)

// ServerConfig holds configuration for the extended RPC server.
type ServerConfig struct {
	MaxRequestSize   int64
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
	CORSAllowOrigins []string
	AuthSecret       string
	RateLimitPerSec  int
	MaxBatchSize     int
	ShutdownTimeout  time.Duration
}

// DefaultServerConfig returns sensible server defaults.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		MaxRequestSize:   5 * 1024 * 1024, // 5 MiB
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		IdleTimeout:      120 * time.Second,
		CORSAllowOrigins: []string{"*"},
		RateLimitPerSec:  100,
		MaxBatchSize:     100,
		ShutdownTimeout:  10 * time.Second,
	}
}

// RateLimiter is a simple token-bucket rate limiter.
type RateLimiter struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	refillRate int
	lastRefill time.Time
}

// NewRateLimiter creates a rate limiter that allows rps requests per second.
// When rps is 0, no rate limiting is applied (nil is returned).
func NewRateLimiter(rps int) *RateLimiter {
	if rps <= 0 {
		return nil
	}
	return &RateLimiter{
		tokens:     rps,
		maxTokens:  rps,
		refillRate: rps,
		lastRefill: time.Now(),
	}
}

// AdvanceLastRefill shifts the lastRefill time back by d, simulating elapsed time.
// Used in tests to trigger token refill without sleeping.
func (rl *RateLimiter) AdvanceLastRefill(d time.Duration) {
	rl.mu.Lock()
	rl.lastRefill = rl.lastRefill.Add(-d)
	rl.mu.Unlock()
}

// Allow returns true if the request is allowed under the rate limit.
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	if elapsed >= time.Second {
		refill := int(elapsed.Seconds()) * rl.refillRate
		rl.tokens += refill
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}

	if rl.tokens <= 0 {
		return false
	}
	rl.tokens--
	return true
}

// MiddlewareFunc is an HTTP middleware function.
type MiddlewareFunc func(http.Handler) http.Handler

// Server is a simple JSON-RPC HTTP server that dispatches requests to a
// RequestHandler, with optional admin_, net_, and beacon_ namespace support.
type Server struct {
	api       RequestHandler
	adminAPI  AdminRequestHandler
	netAPI    NetRequestHandler
	beaconAPI BeaconRequestHandler
	mux       *http.ServeMux
}

// NewServer creates a new Server for the given RequestHandler.
func NewServer(api RequestHandler) *Server {
	s := &Server{
		api: api,
		mux: http.NewServeMux(),
	}
	s.mux.HandleFunc("/", s.handleRPC)
	return s
}

// SetAdminHandler wires an AdminRequestHandler so admin_ methods are dispatched.
func (s *Server) SetAdminHandler(h AdminRequestHandler) {
	s.adminAPI = h
}

// SetNetHandler wires a NetRequestHandler so net_ methods are dispatched.
func (s *Server) SetNetHandler(h NetRequestHandler) {
	s.netAPI = h
}

// SetBeaconHandler wires a BeaconRequestHandler so beacon_ methods are dispatched.
func (s *Server) SetBeaconHandler(h BeaconRequestHandler) {
	s.beaconAPI = h
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, nil, rpctypes.ErrCodeParse, "failed to read request body")
		return
	}

	var req rpctypes.Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, nil, rpctypes.ErrCodeParse, "invalid JSON")
		return
	}

	var resp *rpctypes.Response
	switch {
	case isAdminMethod(req.Method) && s.adminAPI != nil:
		resp = s.adminAPI.HandleAdminRequest(&req)
	case isNetMethod(req.Method) && s.netAPI != nil:
		resp = s.netAPI.HandleNetRequest(&req)
	case isBeaconMethod(req.Method) && s.beaconAPI != nil:
		resp = s.beaconAPI.HandleBeaconRequest(&req)
	default:
		resp = s.api.HandleRequest(&req)
	}
	writeJSON(w, resp)
}

// isAdminMethod reports whether the JSON-RPC method belongs to the admin namespace.
func isAdminMethod(method string) bool {
	return len(method) > 6 && method[:6] == "admin_"
}

// isNetMethod reports whether the method belongs to the net_ namespace.
func isNetMethod(method string) bool {
	return len(method) > 4 && method[:4] == "net_"
}

// isBeaconMethod reports whether the method belongs to the beacon_ namespace.
func isBeaconMethod(method string) bool {
	return len(method) > 7 && method[:7] == "beacon_"
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, id json.RawMessage, code int, message string) {
	resp := &rpctypes.Response{
		JSONRPC: "2.0",
		Error:   &rpctypes.RPCError{Code: code, Message: message},
		ID:      id,
	}
	writeJSON(w, resp)
}

// ExtServer is a full-featured JSON-RPC server with middleware, CORS,
// auth, rate limiting, batch handling, and graceful shutdown.
type ExtServer struct {
	config       ServerConfig
	api          RequestHandler
	adminAPI     AdminRequestHandler
	netAPI       NetRequestHandler
	beaconAPI    BeaconRequestHandler
	batch        *rpcbatch.BatchHandler
	rateLimiter  *RateLimiter
	httpServer   *http.Server
	listener     net.Listener
	mu           sync.Mutex
	started      atomic.Bool
	middlewares  []MiddlewareFunc
	requestCount atomic.Int64
}

// NewExtServer creates a new ExtServer for the given RequestHandler.
func NewExtServer(api RequestHandler, config ServerConfig) *ExtServer {
	if config.MaxRequestSize <= 0 {
		config.MaxRequestSize = DefaultServerConfig().MaxRequestSize
	}
	if config.RateLimitPerSec < 0 {
		config.RateLimitPerSec = DefaultServerConfig().RateLimitPerSec
	}
	s := &ExtServer{
		config:      config,
		api:         api,
		batch:       rpcbatch.NewBatchHandler(api),
		rateLimiter: NewRateLimiter(config.RateLimitPerSec),
	}
	s.batch.SetMaxBatchSize(config.MaxBatchSize)
	return s
}

// SetAdminHandler wires an AdminRequestHandler so admin_ methods are dispatched.
func (s *ExtServer) SetAdminHandler(h AdminRequestHandler) {
	s.adminAPI = h
	s.batch.SetAdminHandler(h)
}

// SetNetHandler wires a NetRequestHandler so net_ methods are dispatched.
func (s *ExtServer) SetNetHandler(h NetRequestHandler) {
	s.netAPI = h
	s.batch.SetNetHandler(h)
}

// SetBeaconHandler wires a BeaconRequestHandler so beacon_ methods are dispatched.
func (s *ExtServer) SetBeaconHandler(h BeaconRequestHandler) {
	s.beaconAPI = h
	s.batch.SetBeaconHandler(h)
}

// Config returns the server configuration.
func (s *ExtServer) Config() ServerConfig {
	return s.config
}

// GetRateLimiter returns the rate limiter (may be nil).
func (s *ExtServer) GetRateLimiter() *RateLimiter {
	return s.rateLimiter
}

// MarkStarted marks the server as started (used in tests to simulate start).
func (s *ExtServer) MarkStarted() {
	s.started.Store(true)
}

// Use adds a middleware to the server's middleware chain.
func (s *ExtServer) Use(mw MiddlewareFunc) {
	s.middlewares = append(s.middlewares, mw)
}

// buildHandler constructs the full HTTP handler with middleware.
func (s *ExtServer) buildHandler() http.Handler {
	var handler http.Handler = http.HandlerFunc(s.handleRPC)
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		handler = s.middlewares[i](handler)
	}
	return handler
}

// Start starts the HTTP server on the given address.
func (s *ExtServer) Start(addr string) error {
	if s.started.Load() {
		return ErrServerStarted
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	handler := s.buildHandler()
	mux := http.NewServeMux()
	mux.Handle("/", handler)

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	s.mu.Lock()
	s.httpServer = srv
	s.listener = ln
	s.mu.Unlock()
	s.started.Store(true)

	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Addr returns the listener address.
func (s *ExtServer) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// Stop gracefully shuts down the server.
func (s *ExtServer) Stop() error {
	s.mu.Lock()
	srv := s.httpServer
	s.mu.Unlock()
	if srv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()
	return srv.Shutdown(ctx)
}

// RequestCount returns the total number of requests served.
func (s *ExtServer) RequestCount() int64 {
	return s.requestCount.Load()
}

// Handler returns the HTTP handler for testing without starting a listener.
func (s *ExtServer) Handler() http.Handler {
	return s.buildHandler()
}

// handleRPC is the main request handler.
func (s *ExtServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	s.requestCount.Add(1)
	s.setCORSHeaders(w, r)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.config.AuthSecret != "" {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || subtle.ConstantTimeCompare([]byte(auth[7:]), []byte(s.config.AuthSecret)) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			writeError(w, nil, rpctypes.ErrCodeInvalidRequest, "unauthorized")
			return
		}
	}

	if s.rateLimiter != nil && !s.rateLimiter.Allow() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		writeError(w, nil, rpctypes.ErrCodeInternal, "rate limited")
		return
	}

	defer r.Body.Close()
	if r.ContentLength > s.config.MaxRequestSize {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		writeError(w, nil, rpctypes.ErrCodeInvalidRequest, "request body too large")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, s.config.MaxRequestSize+1))
	if err != nil {
		writeError(w, nil, rpctypes.ErrCodeParse, "failed to read request body")
		return
	}
	if int64(len(body)) > s.config.MaxRequestSize {
		writeError(w, nil, rpctypes.ErrCodeInvalidRequest, "request body too large")
		return
	}

	if rpcbatch.IsBatchRequest(body) {
		s.handleBatch(w, body)
		return
	}

	var req rpctypes.Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, nil, rpctypes.ErrCodeParse, "invalid JSON")
		return
	}

	var resp *rpctypes.Response
	switch {
	case isAdminMethod(req.Method) && s.adminAPI != nil:
		resp = s.adminAPI.HandleAdminRequest(&req)
	case isNetMethod(req.Method) && s.netAPI != nil:
		resp = s.netAPI.HandleNetRequest(&req)
	case isBeaconMethod(req.Method) && s.beaconAPI != nil:
		resp = s.beaconAPI.HandleBeaconRequest(&req)
	default:
		resp = s.api.HandleRequest(&req)
	}
	writeJSON(w, resp)
}

func (s *ExtServer) handleBatch(w http.ResponseWriter, body []byte) {
	responses, err := s.batch.HandleBatch(body)
	if err != nil {
		writeError(w, nil, rpctypes.ErrCodeInvalidRequest, err.Error())
		return
	}
	writeJSON(w, responses)
}

func (s *ExtServer) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}
	for _, allowed := range s.config.CORSAllowOrigins {
		if allowed == "*" || allowed == origin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")
			return
		}
	}
}

// CORSMiddleware creates a middleware that handles CORS preflight requests.
func CORSMiddleware(allowedOrigins []string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				for _, allowed := range allowedOrigins {
					if allowed == "*" || allowed == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
						w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
						break
					}
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware creates a middleware that validates a Bearer token.
func AuthMiddleware(secret string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret == "" {
				next.ServeHTTP(w, r)
				return
			}
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") || auth[7:] != secret {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"jsonrpc": "2.0",
					"error":   map[string]interface{}{"code": rpctypes.ErrCodeInvalidRequest, "message": "unauthorized"},
					"id":      nil,
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware creates a middleware that enforces rate limiting.
func RateLimitMiddleware(rl *RateLimiter) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rl != nil && !rl.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck
					"jsonrpc": "2.0",
					"error":   map[string]interface{}{"code": rpctypes.ErrCodeInternal, "message": "rate limited"},
					"id":      nil,
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
