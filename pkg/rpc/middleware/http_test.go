package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// --- DefaultCORSConfig ---

func TestDefaultCORSConfig(t *testing.T) {
	cfg := DefaultCORSConfig()
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "*" {
		t.Fatalf("want AllowedOrigins=[*], got %v", cfg.AllowedOrigins)
	}
	if cfg.MaxAge != 3600 {
		t.Fatalf("want MaxAge=3600, got %d", cfg.MaxAge)
	}
	if len(cfg.AllowedMethods) == 0 {
		t.Fatal("want non-empty AllowedMethods")
	}
	if len(cfg.AllowedHeaders) == 0 {
		t.Fatal("want non-empty AllowedHeaders")
	}
}

// --- corsOriginAllowed ---

func TestCorsOriginAllowed_Wildcard(t *testing.T) {
	if !corsOriginAllowed("https://example.com", []string{"*"}) {
		t.Fatal("wildcard should allow any origin")
	}
}

func TestCorsOriginAllowed_ExactMatch(t *testing.T) {
	if !corsOriginAllowed("https://example.com", []string{"https://example.com", "https://other.com"}) {
		t.Fatal("exact match should be allowed")
	}
}

func TestCorsOriginAllowed_NoMatch(t *testing.T) {
	if corsOriginAllowed("https://evil.com", []string{"https://example.com"}) {
		t.Fatal("should not allow unlisted origin")
	}
}

func TestCorsOriginAllowed_EmptyList(t *testing.T) {
	if corsOriginAllowed("https://example.com", []string{}) {
		t.Fatal("empty allowed list should deny everything")
	}
}

// --- formatCORSMaxAge ---

func TestFormatCORSMaxAge_Zero(t *testing.T) {
	if s := formatCORSMaxAge(0); s != "" {
		t.Fatalf("want empty string for 0, got %q", s)
	}
}

func TestFormatCORSMaxAge_Positive(t *testing.T) {
	if s := formatCORSMaxAge(3600); s != "3600" {
		t.Fatalf("want \"3600\", got %q", s)
	}
	if s := formatCORSMaxAge(1); s != "1" {
		t.Fatalf("want \"1\", got %q", s)
	}
	if s := formatCORSMaxAge(100); s != "100" {
		t.Fatalf("want \"100\", got %q", s)
	}
}

// --- CORSMiddleware ---

func TestCORSMiddleware_Preflight(t *testing.T) {
	cfg := DefaultCORSConfig()
	handler := CORSMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called for OPTIONS preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rr.Code)
	}
}

func TestCORSMiddleware_WildcardOrigin(t *testing.T) {
	cfg := DefaultCORSConfig()
	handler := CORSMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "https://anything.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	origin := rr.Header().Get("Access-Control-Allow-Origin")
	if origin == "" {
		t.Fatal("want CORS origin header set")
	}
}

func TestCORSMiddleware_SpecificOriginAllowed(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"https://trusted.com"},
		AllowedMethods: []string{"POST"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         600,
	}
	handler := CORSMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "https://trusted.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://trusted.com" {
		t.Fatalf("want https://trusted.com, got %q", got)
	}
}

func TestCORSMiddleware_MaxAgeHeader(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST"},
		MaxAge:         1200,
	}
	handler := CORSMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Max-Age"); got != "1200" {
		t.Fatalf("want 1200, got %q", got)
	}
}

func TestCORSMiddleware_NoOriginHeader(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST"},
	}
	called := false
	handler := CORSMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("inner handler should be called when no Origin header")
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("want *, got %q", got)
	}
}

// --- AuthMiddleware ---

func TestAuthMiddleware_NoAuth_AllowUnauthenticated(t *testing.T) {
	cfg := AuthConfig{AllowUnauthenticated: true}
	called := false
	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler should be called when AllowUnauthenticated=true and no auth")
	}
}

func TestAuthMiddleware_NoAuth_Deny(t *testing.T) {
	cfg := AuthConfig{AllowUnauthenticated: false}
	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_ValidBearerToken(t *testing.T) {
	secret := "mysecrettoken"
	cfg := AuthConfig{JWTSecret: secret}
	called := false
	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+secret)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler should be called with valid Bearer token")
	}
}

func TestAuthMiddleware_InvalidBearerToken(t *testing.T) {
	cfg := AuthConfig{JWTSecret: "correctsecret"}
	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with wrong token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer wrongsecret")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_ValidAPIKey(t *testing.T) {
	cfg := AuthConfig{
		APIKeys: map[string]bool{"key123": true},
	}
	called := false
	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "ApiKey key123")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler should be called with valid API key")
	}
}

func TestAuthMiddleware_InvalidAPIKey(t *testing.T) {
	cfg := AuthConfig{
		APIKeys: map[string]bool{"key123": true},
	}
	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with wrong API key")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "ApiKey wrongkey")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_UnrecognizedScheme_AllowUnauthenticated(t *testing.T) {
	cfg := AuthConfig{AllowUnauthenticated: true}
	called := false
	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Digest somevalue")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler should be called with AllowUnauthenticated and unrecognized scheme")
	}
}

func TestAuthMiddleware_UnrecognizedScheme_Deny(t *testing.T) {
	cfg := AuthConfig{AllowUnauthenticated: false}
	handler := AuthMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Digest somevalue")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}

// --- LogStore ---

func TestLogStore_AddAndEntries(t *testing.T) {
	store := NewLogStore()
	if store.Len() != 0 {
		t.Fatalf("want 0, got %d", store.Len())
	}

	entry := LogEntry{
		Method:     "POST",
		Path:       "/rpc",
		StatusCode: 200,
		Duration:   5 * time.Millisecond,
		RemoteAddr: "127.0.0.1:1234",
		Timestamp:  time.Now(),
	}
	store.Add(entry)
	store.Add(entry)

	if store.Len() != 2 {
		t.Fatalf("want 2, got %d", store.Len())
	}

	entries := store.Entries()
	if len(entries) != 2 {
		t.Fatalf("want 2 entries, got %d", len(entries))
	}
}

func TestLogStore_EntriesReturnsCopy(t *testing.T) {
	store := NewLogStore()
	store.Add(LogEntry{Method: "POST"})

	entries1 := store.Entries()
	entries1[0].Method = "MODIFIED"

	entries2 := store.Entries()
	if entries2[0].Method != "POST" {
		t.Fatal("Entries() should return a copy, not a reference")
	}
}

// --- LoggingMiddleware ---

func TestLoggingMiddleware_RecordsEntry(t *testing.T) {
	store := NewLogStore()
	handler := LoggingMiddleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if store.Len() != 1 {
		t.Fatalf("want 1 log entry, got %d", store.Len())
	}
	entry := store.Entries()[0]
	if entry.Method != "POST" {
		t.Fatalf("want POST, got %s", entry.Method)
	}
	if entry.Path != "/test" {
		t.Fatalf("want /test, got %s", entry.Path)
	}
	if entry.StatusCode != http.StatusCreated {
		t.Fatalf("want 201, got %d", entry.StatusCode)
	}
}

func TestLoggingMiddleware_DefaultStatusOK(t *testing.T) {
	store := NewLogStore()
	handler := LoggingMiddleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok")) //nolint:errcheck
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	entry := store.Entries()[0]
	if entry.StatusCode != http.StatusOK {
		t.Fatalf("want default 200, got %d", entry.StatusCode)
	}
}

// --- CompressionMiddleware ---

func TestCompressionMiddleware_WithGzipAccept(t *testing.T) {
	handler := CompressionMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello compressed")) //nolint:errcheck
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if ce := rr.Header().Get("Content-Encoding"); ce != "gzip" {
		t.Fatalf("want Content-Encoding: gzip, got %q", ce)
	}

	gr, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	defer gr.Close()
	body, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("reading gzip body: %v", err)
	}
	if string(body) != "hello compressed" {
		t.Fatalf("want \"hello compressed\", got %q", string(body))
	}
}

func TestCompressionMiddleware_WithoutGzipAccept(t *testing.T) {
	handler := CompressionMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("plain body")) //nolint:errcheck
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if ce := rr.Header().Get("Content-Encoding"); ce != "" {
		t.Fatalf("want no Content-Encoding, got %q", ce)
	}
	if rr.Body.String() != "plain body" {
		t.Fatalf("want plain body, got %q", rr.Body.String())
	}
}

// --- RateLimitMiddleware ---

func TestRateLimitMiddleware_AllowsRequests(t *testing.T) {
	cfg := RateLimitConfig{RequestsPerSecond: 100}
	handler := RateLimitMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
}

func TestRateLimitMiddleware_ExceedsRateLimit(t *testing.T) {
	cfg := RateLimitConfig{RequestsPerSecond: 1}
	handler := RateLimitMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	denied := 0
	for range 10 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code == http.StatusTooManyRequests {
			denied++
		}
	}

	if denied == 0 {
		t.Fatal("expected some requests to be rate limited")
	}
}

// --- extractClientIP ---

func TestExtractClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
	ip := extractClientIP(req)
	if ip != "192.168.1.1" {
		t.Fatalf("want 192.168.1.1, got %q", ip)
	}
}

func TestExtractClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "172.16.0.5")
	ip := extractClientIP(req)
	if ip != "172.16.0.5" {
		t.Fatalf("want 172.16.0.5, got %q", ip)
	}
}

func TestExtractClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.100:54321"
	ip := extractClientIP(req)
	if ip != "192.168.1.100" {
		t.Fatalf("want 192.168.1.100, got %q", ip)
	}
}

func TestExtractClientIP_RemoteAddrNoPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.100"
	req.Header.Del("X-Forwarded-For")
	req.Header.Del("X-Real-IP")
	ip := extractClientIP(req)
	if ip != "192.168.1.100" {
		t.Fatalf("want 192.168.1.100, got %q", ip)
	}
}

// --- MiddlewareChain ---

func TestMiddlewareChain_OrderIsPreserved(t *testing.T) {
	var order []string

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw1-after")
		})
	}
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw2-after")
		})
	}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "inner")
	})

	chain := MiddlewareChain(inner, mw1, mw2)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	chain.ServeHTTP(rr, req)

	expected := []string{"mw1-before", "mw2-before", "inner", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Fatalf("want %v, got %v", expected, order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Fatalf("step %d: want %q, got %q", i, v, order[i])
		}
	}
}

func TestMiddlewareChain_NoMiddleware(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	chain := MiddlewareChain(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	chain.ServeHTTP(rr, req)

	if !called {
		t.Fatal("inner handler should be called with no middleware")
	}
}

// --- statusRecorder ---

func TestStatusRecorder_CapturesCode(t *testing.T) {
	rr := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rr, statusCode: http.StatusOK}
	sr.WriteHeader(http.StatusTeapot)

	if sr.statusCode != http.StatusTeapot {
		t.Fatalf("want 418, got %d", sr.statusCode)
	}
}

// --- RateLimitMiddleware default RPS ---

func TestRateLimitMiddleware_ZeroRPS_UsesDefault(t *testing.T) {
	cfg := RateLimitConfig{RequestsPerSecond: 0}
	handler := RateLimitMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.99:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
}

// --- CORSMiddleware methods and headers ---

func TestCORSMiddleware_AllowedMethodsHeader(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
	}
	handler := CORSMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	methods := rr.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(methods, "GET") || !strings.Contains(methods, "POST") {
		t.Fatalf("want methods header to contain GET and POST, got %q", methods)
	}
}
