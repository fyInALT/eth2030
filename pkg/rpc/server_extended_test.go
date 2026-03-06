package rpc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// --- DefaultServerConfig ---

func TestDefaultServerConfig(t *testing.T) {
	cfg := DefaultServerConfig()
	if cfg.MaxRequestSize != 5*1024*1024 {
		t.Fatalf("want 5MiB, got %d", cfg.MaxRequestSize)
	}
	if cfg.ReadTimeout != 30*time.Second {
		t.Fatalf("want 30s ReadTimeout, got %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 30*time.Second {
		t.Fatalf("want 30s WriteTimeout, got %v", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 120*time.Second {
		t.Fatalf("want 120s IdleTimeout, got %v", cfg.IdleTimeout)
	}
	if len(cfg.CORSAllowOrigins) == 0 {
		t.Fatal("want non-empty CORSAllowOrigins")
	}
	if cfg.RateLimitPerSec != 100 {
		t.Fatalf("want RateLimitPerSec=100, got %d", cfg.RateLimitPerSec)
	}
	if cfg.MaxBatchSize != 100 {
		t.Fatalf("want MaxBatchSize=100, got %d", cfg.MaxBatchSize)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("want 10s ShutdownTimeout, got %v", cfg.ShutdownTimeout)
	}
}

// --- RateLimiter.Allow ---

func TestRateLimiter_Allow_AllowsInitialRequests(t *testing.T) {
	rl := NewRateLimiter(10)
	if rl == nil {
		t.Fatal("expected non-nil RateLimiter")
	}
	// Initial tokens equal rps, so first 10 should be allowed.
	for i := range 10 {
		if !rl.Allow() {
			t.Fatalf("request %d should be allowed (have tokens)", i+1)
		}
	}
}

func TestRateLimiter_Allow_DeniesWhenExhausted(t *testing.T) {
	rl := NewRateLimiter(3)
	// Drain tokens.
	for range 3 {
		rl.Allow()
	}
	if rl.Allow() {
		t.Fatal("should be denied after exhausting tokens")
	}
}

func TestRateLimiter_ZeroRPS_ReturnsNil(t *testing.T) {
	rl := NewRateLimiter(0)
	if rl != nil {
		t.Fatal("zero RPS should return nil limiter")
	}
}

func TestRateLimiter_NegativeRPS_ReturnsNil(t *testing.T) {
	rl := NewRateLimiter(-5)
	if rl != nil {
		t.Fatal("negative RPS should return nil limiter")
	}
}

func TestRateLimiter_Allow_RefillsAfterTime(t *testing.T) {
	rl := NewRateLimiter(5)
	// Drain all tokens.
	for range 5 {
		rl.Allow()
	}
	// Simulate time passing: manually adjust lastRefill to 2s ago.
	rl.mu.Lock()
	rl.lastRefill = rl.lastRefill.Add(-2 * time.Second)
	rl.mu.Unlock()

	// After 2 seconds elapsed, should have refilled 2*5=10 tokens (capped at maxTokens=5).
	if !rl.Allow() {
		t.Fatal("should allow after refill period")
	}
}

// --- NewExtServer ---

func TestNewExtServer_DefaultsApplied(t *testing.T) {
	cfg := ServerConfig{
		MaxRequestSize: 0, // should be defaulted
	}
	srv := NewExtServer(newMockBackend(), cfg)
	if srv == nil {
		t.Fatal("expected non-nil ExtServer")
	}
	if srv.config.MaxRequestSize != DefaultServerConfig().MaxRequestSize {
		t.Fatalf("want default MaxRequestSize, got %d", srv.config.MaxRequestSize)
	}
}

func TestNewExtServer_NegativeRPSDefaults(t *testing.T) {
	cfg := ServerConfig{
		RateLimitPerSec: -1,
	}
	srv := NewExtServer(newMockBackend(), cfg)
	if srv.config.RateLimitPerSec != DefaultServerConfig().RateLimitPerSec {
		t.Fatalf("negative RPS should be replaced with default, got %d", srv.config.RateLimitPerSec)
	}
}

func TestNewExtServer_ZeroRPS_NoRateLimiter(t *testing.T) {
	cfg := ServerConfig{
		RateLimitPerSec: 0,
	}
	srv := NewExtServer(newMockBackend(), cfg)
	if srv.rateLimiter != nil {
		t.Fatal("zero RPS should result in nil rateLimiter")
	}
}

// --- ExtServer.Handler ---

func TestExtServer_Handler_NotNil(t *testing.T) {
	srv := NewExtServer(newMockBackend(), DefaultServerConfig())
	if srv.Handler() == nil {
		t.Fatal("Handler() should not return nil")
	}
}

func TestExtServer_Handler_ServesRequest(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.AuthSecret = "" // no auth required
	srv := NewExtServer(newTestBackend(), cfg)

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	body := `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`
	resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	var rpcResp Response
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if rpcResp.Error != nil {
		t.Fatalf("RPC error: %v", rpcResp.Error.Message)
	}
}

// --- ExtServer.RequestCount ---

func TestExtServer_RequestCount(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.AuthSecret = ""
	srv := NewExtServer(newMockBackend(), cfg)

	if srv.RequestCount() != 0 {
		t.Fatalf("want 0 initial requests, got %d", srv.RequestCount())
	}

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	for range 3 {
		body := `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`
		resp, _ := http.Post(ts.URL, "application/json", bytes.NewBufferString(body))
		if resp != nil {
			resp.Body.Close()
		}
	}

	if srv.RequestCount() != 3 {
		t.Fatalf("want 3 requests counted, got %d", srv.RequestCount())
	}
}

// --- ExtServer.Addr before start ---

func TestExtServer_Addr_BeforeStart(t *testing.T) {
	srv := NewExtServer(newMockBackend(), DefaultServerConfig())
	if srv.Addr() != nil {
		t.Fatal("Addr() should return nil before Start()")
	}
}

// --- ExtServer.Use (middleware chain) ---

func TestExtServer_Use_MiddlewareApplied(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.AuthSecret = ""
	srv := NewExtServer(newMockBackend(), cfg)

	var middlewareCalled bool
	srv.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	})

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	body := `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`
	resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()

	if !middlewareCalled {
		t.Fatal("middleware should have been called")
	}
}

// --- Auth in handleRPC ---

func TestExtServer_AuthRejectsUnauthorized(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.AuthSecret = "supersecret"
	srv := NewExtServer(newMockBackend(), cfg)

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	body := `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`
	resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("want 401 without auth, got %d", resp.StatusCode)
	}
}

func TestExtServer_AuthAllowsValidToken(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.AuthSecret = "supersecret"
	srv := NewExtServer(newMockBackend(), cfg)

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL, bytes.NewBufferString(`{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer supersecret")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200 with valid token, got %d", resp.StatusCode)
	}
}

// --- ExtCORSMiddleware ---

func TestExtCORSMiddleware_SetsHeaders(t *testing.T) {
	mw := ExtCORSMiddleware([]string{"*"})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got == "" {
		t.Fatal("expected Access-Control-Allow-Origin header")
	}
}

func TestExtCORSMiddleware_Preflight(t *testing.T) {
	mw := ExtCORSMiddleware([]string{"*"})
	innerCalled := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200 for OPTIONS preflight, got %d", rr.Code)
	}
	if innerCalled {
		t.Fatal("inner handler should not be called for OPTIONS preflight")
	}
}

// --- ExtAuthMiddleware ---

func TestExtAuthMiddleware_NoSecret_AllowAll(t *testing.T) {
	mw := ExtAuthMiddleware("")
	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("should allow all when secret is empty")
	}
}

func TestExtAuthMiddleware_ValidToken(t *testing.T) {
	mw := ExtAuthMiddleware("mytoken")
	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer mytoken")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("should allow request with valid Bearer token")
	}
}

func TestExtAuthMiddleware_InvalidToken(t *testing.T) {
	mw := ExtAuthMiddleware("mytoken")
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not be called with invalid token")
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer wrongtoken")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}

// --- ExtRateLimitMiddleware ---

func TestExtRateLimitMiddleware_NilLimiter_AllowAll(t *testing.T) {
	mw := ExtRateLimitMiddleware(nil)
	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("nil limiter should allow all requests")
	}
}

func TestExtRateLimitMiddleware_ExhaustedLimiter_Denies(t *testing.T) {
	rl := NewRateLimiter(2)
	// Exhaust tokens.
	rl.Allow()
	rl.Allow()

	mw := ExtRateLimitMiddleware(rl)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called when rate limited")
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429, got %d", rr.Code)
	}
}

// --- ExtServer.Stop before Start ---

func TestExtServer_Stop_BeforeStart(t *testing.T) {
	srv := NewExtServer(newMockBackend(), DefaultServerConfig())
	if err := srv.Stop(); err != nil {
		t.Fatalf("Stop() before Start() should not error, got %v", err)
	}
}

// --- ExtServer.Start duplicate ---

func TestExtServer_Start_AlreadyStarted(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.AuthSecret = ""
	srv := NewExtServer(newMockBackend(), cfg)

	// Start on a random port in background.
	started := make(chan struct{})
	go func() {
		// We mark started=true before blocking, so we signal after the flag is set.
		srv.started.Store(true)
		close(started)
	}()
	<-started

	err := srv.Start(":0")
	if err != ErrServerStarted {
		t.Fatalf("want ErrServerStarted, got %v", err)
	}
}

// --- handleRPC: OPTIONS method ---

func TestExtServer_HandleRPC_OptionsMethod(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.AuthSecret = ""
	srv := NewExtServer(newMockBackend(), cfg)

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodOptions, ts.URL, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200 for OPTIONS, got %d", resp.StatusCode)
	}
}

// --- handleRPC: method not allowed ---

func TestExtServer_HandleRPC_GetMethodNotAllowed(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.AuthSecret = ""
	srv := NewExtServer(newMockBackend(), cfg)

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", resp.StatusCode)
	}
}
