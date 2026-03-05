package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// echoHandler returns a handler that echoes its first param as result.
func echoHandler() RPCHandleFunc {
	return func(ctx *RPCContext) *RPCResponse {
		result := "pong"
		if len(ctx.Request.Params) > 0 {
			result = string(ctx.Request.Params[0])
		}
		return &RPCResponse{
			JSONRPC: "2.0",
			Result:  result,
			ID:      ctx.Request.ID,
		}
	}
}

func newTestHandler() *RPCHandler {
	cfg := DefaultRPCHandlerConfig()
	h := NewRPCHandler(cfg)
	h.RegisterMethod("eth_ping", echoHandler())
	return h
}

// postJSON sends a POST request with the given body to the handler.
func postJSON(h http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestDefaultRPCHandlerConfig(t *testing.T) {
	cfg := DefaultRPCHandlerConfig()
	if cfg.MaxBatchSize != 100 {
		t.Errorf("want MaxBatchSize=100, got %d", cfg.MaxBatchSize)
	}
	if cfg.MaxRequestSize != 5*1024*1024 {
		t.Errorf("want MaxRequestSize=5MB, got %d", cfg.MaxRequestSize)
	}
	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("want ReadTimeout=30s, got %v", cfg.ReadTimeout)
	}
	if cfg.EnableAuth {
		t.Error("want EnableAuth=false by default")
	}
	if cfg.RateLimit != 0 {
		t.Errorf("want RateLimit=0, got %d", cfg.RateLimit)
	}
	if cfg.RateBurst != 50 {
		t.Errorf("want RateBurst=50, got %d", cfg.RateBurst)
	}
}

func TestNewRPCHandler_RegisterMethod(t *testing.T) {
	h := newTestHandler()
	if h.MethodCount() != 1 {
		t.Fatalf("want 1 method, got %d", h.MethodCount())
	}
	methods := h.Methods()
	if len(methods) != 1 || methods[0] != "eth_ping" {
		t.Fatalf("want [eth_ping], got %v", methods)
	}

	h.RegisterMethod("eth_foo", echoHandler())
	if h.MethodCount() != 2 {
		t.Fatalf("want 2 methods after second register, got %d", h.MethodCount())
	}
}

func TestServeHTTP_MethodNotAllowed(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
}

func TestServeHTTP_ValidRequest(t *testing.T) {
	h := newTestHandler()
	body := `{"jsonrpc":"2.0","method":"eth_ping","params":[],"id":1}`
	rec := postJSON(h, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var resp RPCResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if resp.JSONRPC != "2.0" {
		t.Errorf("want jsonrpc=2.0, got %q", resp.JSONRPC)
	}
}

func TestServeHTTP_MethodNotFound(t *testing.T) {
	h := newTestHandler()
	body := `{"jsonrpc":"2.0","method":"eth_unknown","params":[],"id":2}`
	rec := postJSON(h, body)
	var resp RPCResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == nil {
		t.Fatal("want error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("want code -32601, got %d", resp.Error.Code)
	}
}

func TestServeHTTP_InvalidJSON(t *testing.T) {
	h := newTestHandler()
	rec := postJSON(h, `not json`)
	var resp RPCResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == nil || resp.Error.Code != -32700 {
		t.Fatalf("want parse error -32700, got %v", resp.Error)
	}
}

func TestServeHTTP_InvalidJSONRPCVersion(t *testing.T) {
	h := newTestHandler()
	body := `{"jsonrpc":"1.0","method":"eth_ping","params":[],"id":1}`
	rec := postJSON(h, body)
	var resp RPCResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == nil || resp.Error.Code != -32600 {
		t.Fatalf("want invalid version error -32600, got %v", resp.Error)
	}
}

func TestServeHTTP_RequestBodyTooLarge(t *testing.T) {
	cfg := DefaultRPCHandlerConfig()
	cfg.MaxRequestSize = 10
	h := NewRPCHandler(cfg)
	// Body is longer than 10 bytes.
	body := `{"jsonrpc":"2.0","method":"eth_ping","params":[],"id":1}`
	rec := postJSON(h, body)
	var resp RPCResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == nil || resp.Error.Code != -32600 {
		t.Fatalf("want -32600 body too large, got %v", resp.Error)
	}
}

func TestServeHTTP_BatchRequest(t *testing.T) {
	h := newTestHandler()
	body := `[{"jsonrpc":"2.0","method":"eth_ping","params":[],"id":1},{"jsonrpc":"2.0","method":"eth_ping","params":[],"id":2}]`
	rec := postJSON(h, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var responses []*RPCResponse
	if err := json.NewDecoder(rec.Body).Decode(&responses); err != nil {
		t.Fatalf("decode batch response: %v", err)
	}
	if len(responses) != 2 {
		t.Fatalf("want 2 responses, got %d", len(responses))
	}
	for _, r := range responses {
		if r.Error != nil {
			t.Errorf("unexpected error: %v", r.Error)
		}
	}
}

func TestServeHTTP_BatchEmpty(t *testing.T) {
	h := newTestHandler()
	rec := postJSON(h, `[]`)
	var resp RPCResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == nil || resp.Error.Code != -32600 {
		t.Fatalf("want -32600 empty batch, got %v", resp.Error)
	}
}

func TestServeHTTP_BatchTooLarge(t *testing.T) {
	cfg := DefaultRPCHandlerConfig()
	cfg.MaxBatchSize = 2
	h := NewRPCHandler(cfg)
	h.RegisterMethod("eth_ping", echoHandler())

	// Build a batch of 3 requests.
	reqs := make([]string, 3)
	for i := range reqs {
		reqs[i] = fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_ping","params":[],"id":%d}`, i+1)
	}
	body := "[" + strings.Join(reqs, ",") + "]"
	rec := postJSON(h, body)
	var resp RPCResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == nil || resp.Error.Code != -32600 {
		t.Fatalf("want -32600 batch too large, got %v", resp.Error)
	}
}

func TestServeHTTP_BatchInvalidJSON(t *testing.T) {
	h := newTestHandler()
	rec := postJSON(h, `[not json]`)
	var resp RPCResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == nil || resp.Error.Code != -32700 {
		t.Fatalf("want -32700 parse error, got %v", resp.Error)
	}
}

func TestUseMiddleware_LoggingMiddleware(t *testing.T) {
	h := newTestHandler()
	h.Use(LoggingMiddleware())

	body := `{"jsonrpc":"2.0","method":"eth_ping","params":[],"id":1}`
	rec := postJSON(h, body)
	var resp RPCResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error != nil {
		t.Fatalf("unexpected error after logging middleware: %v", resp.Error)
	}
}

func TestUseMiddleware_RateLimitMiddleware_Allow(t *testing.T) {
	cfg := DefaultRPCHandlerConfig()
	cfg.RateLimit = 100
	cfg.RateBurst = 50
	h := NewRPCHandler(cfg)
	h.RegisterMethod("eth_ping", echoHandler())
	h.Use(RateLimitMiddleware(h.limiter))

	body := `{"jsonrpc":"2.0","method":"eth_ping","params":[],"id":1}`
	rec := postJSON(h, body)
	var resp RPCResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error != nil {
		t.Fatalf("first request should be allowed: %v", resp.Error)
	}
}

func TestUseMiddleware_RateLimitMiddleware_Deny(t *testing.T) {
	// Use rate=1, burst=1: first request consumes the only token.
	cfg := DefaultRPCHandlerConfig()
	cfg.RateLimit = 1
	cfg.RateBurst = 1
	h := NewRPCHandler(cfg)
	h.RegisterMethod("eth_ping", echoHandler())
	h.Use(RateLimitMiddleware(h.limiter))

	body := `{"jsonrpc":"2.0","method":"eth_ping","params":[],"id":1}`

	// First request should pass.
	rec1 := postJSON(h, body)
	var resp1 RPCResponse
	json.NewDecoder(rec1.Body).Decode(&resp1)
	if resp1.Error != nil {
		t.Fatalf("first request should succeed: %v", resp1.Error)
	}

	// Second request immediately after should be rate-limited.
	rec2 := postJSON(h, body)
	var resp2 RPCResponse
	json.NewDecoder(rec2.Body).Decode(&resp2)
	if resp2.Error == nil || resp2.Error.Code != -32005 {
		t.Fatalf("second request should be rate limited (-32005), got %v", resp2.Error)
	}
}

func TestUseMiddleware_AuthMiddleware_BatchSkip(t *testing.T) {
	h := newTestHandler()
	// Auth middleware skips batch sub-requests.
	h.Use(AuthMiddleware("secret"))

	body := `{"jsonrpc":"2.0","method":"eth_ping","params":[],"id":1}`
	rec := postJSON(h, body)
	var resp RPCResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	// Non-batch calls should also pass through (current impl just calls next).
	if resp.Error != nil {
		t.Fatalf("auth middleware should pass through: %v", resp.Error)
	}
}

func TestMiddlewareChain_Order(t *testing.T) {
	// Verify that middleware executes in registration order (first = outermost).
	var order []string
	var mu sync.Mutex

	makeMiddleware := func(name string) RPCMiddleware {
		return func(ctx *RPCContext, next RPCHandleFunc) *RPCResponse {
			mu.Lock()
			order = append(order, name+"-before")
			mu.Unlock()
			resp := next(ctx)
			mu.Lock()
			order = append(order, name+"-after")
			mu.Unlock()
			return resp
		}
	}

	h := newTestHandler()
	h.Use(makeMiddleware("first"))
	h.Use(makeMiddleware("second"))

	body := `{"jsonrpc":"2.0","method":"eth_ping","params":[],"id":1}`
	postJSON(h, body)

	want := []string{"first-before", "second-before", "second-after", "first-after"}
	if len(order) != len(want) {
		t.Fatalf("want %v, got %v", want, order)
	}
	for i, v := range want {
		if order[i] != v {
			t.Errorf("step %d: want %q, got %q", i, v, order[i])
		}
	}
}

func TestIsWebSocketUpgrade(t *testing.T) {
	tests := []struct {
		upgrade    string
		connection string
		want       bool
	}{
		{"websocket", "Upgrade", true},
		{"WebSocket", "upgrade", true},
		{"", "Upgrade", false},
		{"websocket", "", false},
		{"http", "Upgrade", false},
	}
	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if tc.upgrade != "" {
			req.Header.Set("Upgrade", tc.upgrade)
		}
		if tc.connection != "" {
			req.Header.Set("Connection", tc.connection)
		}
		got := isWebSocketUpgrade(req)
		if got != tc.want {
			t.Errorf("upgrade=%q connection=%q: want %v, got %v",
				tc.upgrade, tc.connection, tc.want, got)
		}
	}
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		xff        string
		xri        string
		remoteAddr string
		want       string
	}{
		{xff: "10.0.0.1", want: "10.0.0.1"},
		{xff: "10.0.0.1, 192.168.1.1", want: "10.0.0.1"},
		{xri: "172.16.0.5", want: "172.16.0.5"},
		{remoteAddr: "1.2.3.4:5678", want: "1.2.3.4"},
		{remoteAddr: "::1:8080", want: "::1"},
	}
	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		if tc.xff != "" {
			req.Header.Set("X-Forwarded-For", tc.xff)
		}
		if tc.xri != "" {
			req.Header.Set("X-Real-IP", tc.xri)
		}
		if tc.remoteAddr != "" {
			req.RemoteAddr = tc.remoteAddr
		}
		got := extractIP(req)
		if got != tc.want {
			t.Errorf("xff=%q xri=%q remote=%q: want %q, got %q",
				tc.xff, tc.xri, tc.remoteAddr, tc.want, got)
		}
	}
}

func TestTrimLeadingWhitespace(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"  hello", "hello"},
		{"\t\nworld", "world"},
		{"nowhitespace", "nowhitespace"},
		{"", ""},
		{"   ", ""},
	}
	for _, tc := range tests {
		got := string(trimLeadingWhitespace([]byte(tc.in)))
		if got != tc.want {
			t.Errorf("input %q: want %q, got %q", tc.in, tc.want, got)
		}
	}
}

func TestRateLimiter_AllowUnlimited(t *testing.T) {
	rl := newRateLimiter(0, 10)
	for range 1000 {
		if !rl.Allow("1.2.3.4") {
			t.Fatal("unlimited limiter should always allow")
		}
	}
}

func TestRateLimiter_AllowBurst(t *testing.T) {
	rl := newRateLimiter(10, 5)
	// First 5 requests should succeed (burst=5).
	for i := range 5 {
		if !rl.Allow("10.0.0.1") {
			t.Fatalf("request %d should be allowed within burst", i)
		}
	}
	// 6th request should be denied.
	if rl.Allow("10.0.0.1") {
		t.Fatal("request after burst exhaustion should be denied")
	}
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	// rate=100/s, burst=1: use the single token, then wait for refill.
	rl := newRateLimiter(100, 1)
	if !rl.Allow("2.2.2.2") {
		t.Fatal("first request should be allowed")
	}
	if rl.Allow("2.2.2.2") {
		t.Fatal("second immediate request should be denied")
	}
	// Wait for token refill (1 token at 100/s = 10ms).
	time.Sleep(20 * time.Millisecond)
	if !rl.Allow("2.2.2.2") {
		t.Fatal("request after refill should be allowed")
	}
}

func TestRateLimiter_PerIP(t *testing.T) {
	rl := newRateLimiter(1, 1)
	// Exhaust for ip1.
	if !rl.Allow("ip1") {
		t.Fatal("first request for ip1 should succeed")
	}
	if rl.Allow("ip1") {
		t.Fatal("second request for ip1 should be denied")
	}
	// ip2 should still have its own bucket.
	if !rl.Allow("ip2") {
		t.Fatal("first request for ip2 should succeed")
	}
}

func TestWriteJSON(t *testing.T) {
	h := newTestHandler()
	rec := httptest.NewRecorder()
	h.writeJSON(rec, map[string]string{"key": "value"})
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("want application/json content-type, got %q", ct)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"key"`)) {
		t.Errorf("expected JSON body, got %q", rec.Body.String())
	}
}

func TestWriteRPCError(t *testing.T) {
	h := newTestHandler()
	rec := httptest.NewRecorder()
	h.writeRPCError(rec, nil, -32600, "bad request")
	var resp RPCResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == nil {
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != -32600 {
		t.Errorf("want code -32600, got %d", resp.Error.Code)
	}
	if resp.Error.Message != "bad request" {
		t.Errorf("want message %q, got %q", "bad request", resp.Error.Message)
	}
}

func TestMethodsReturnsAllRegistered(t *testing.T) {
	h := NewRPCHandler(DefaultRPCHandlerConfig())
	names := []string{"eth_a", "eth_b", "eth_c"}
	for _, n := range names {
		h.RegisterMethod(n, echoHandler())
	}
	got := h.Methods()
	if len(got) != len(names) {
		t.Fatalf("want %d methods, got %d", len(names), len(got))
	}
	// Verify all names are present (order may vary).
	seen := make(map[string]bool)
	for _, m := range got {
		seen[m] = true
	}
	for _, n := range names {
		if !seen[n] {
			t.Errorf("method %q not returned by Methods()", n)
		}
	}
}

func TestRequestSequenceIncreases(t *testing.T) {
	h := newTestHandler()
	var ids []uint64
	var mu sync.Mutex

	h.RegisterMethod("eth_seqcheck", func(ctx *RPCContext) *RPCResponse {
		mu.Lock()
		ids = append(ids, ctx.RequestID)
		mu.Unlock()
		return &RPCResponse{JSONRPC: "2.0", Result: "ok", ID: ctx.Request.ID}
	})

	for range 5 {
		body := `{"jsonrpc":"2.0","method":"eth_seqcheck","params":[],"id":1}`
		postJSON(h, body)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(ids) != 5 {
		t.Fatalf("want 5 request IDs, got %d", len(ids))
	}
	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			t.Errorf("request IDs not monotonically increasing: %v", ids)
		}
	}
}

func TestExtractIP_NoPort(t *testing.T) {
	// RemoteAddr without ":" should return the whole address.
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "1.2.3.4"
	got := extractIP(req)
	if got != "1.2.3.4" {
		t.Errorf("want 1.2.3.4, got %q", got)
	}
}

func TestExtractIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Real-IP", "  203.0.113.7  ")
	got := extractIP(req)
	if got != "203.0.113.7" {
		t.Errorf("want 203.0.113.7, got %q", got)
	}
}

func TestLoggingMiddleware_ErrorBranch(t *testing.T) {
	h := newTestHandler()
	h.RegisterMethod("eth_fail", func(ctx *RPCContext) *RPCResponse {
		return &RPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCErr{Code: -32000, Message: "internal error"},
			ID:      ctx.Request.ID,
		}
	})
	h.Use(LoggingMiddleware())

	body := `{"jsonrpc":"2.0","method":"eth_fail","params":[],"id":1}`
	rec := postJSON(h, body)
	var resp RPCResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == nil {
		t.Fatal("expected error response")
	}
	if resp.Error.Code != -32000 {
		t.Errorf("want code -32000, got %d", resp.Error.Code)
	}
}

func TestAuthMiddleware_BatchTrue(t *testing.T) {
	// Directly test the AuthMiddleware with IsBatch=true to cover the branch.
	mw := AuthMiddleware("secret")
	called := false
	next := func(ctx *RPCContext) *RPCResponse {
		called = true
		return &RPCResponse{JSONRPC: "2.0", Result: "ok"}
	}
	ctx := &RPCContext{
		Request: &RPCRequest{JSONRPC: "2.0", Method: "eth_test"},
		IsBatch: true,
	}
	resp := mw(ctx, next)
	if !called {
		t.Error("next should be called when IsBatch=true")
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

func TestIsBatchFlagSetInBatch(t *testing.T) {
	h := newTestHandler()
	var isBatchFlags []bool
	var mu sync.Mutex

	h.RegisterMethod("eth_batchcheck", func(ctx *RPCContext) *RPCResponse {
		mu.Lock()
		isBatchFlags = append(isBatchFlags, ctx.IsBatch)
		mu.Unlock()
		return &RPCResponse{JSONRPC: "2.0", Result: "ok", ID: ctx.Request.ID}
	})

	// Single request — IsBatch should be false.
	body := `{"jsonrpc":"2.0","method":"eth_batchcheck","params":[],"id":1}`
	postJSON(h, body)

	// Batch request — IsBatch should be true.
	batchBody := `[{"jsonrpc":"2.0","method":"eth_batchcheck","params":[],"id":2}]`
	postJSON(h, batchBody)

	mu.Lock()
	defer mu.Unlock()
	if len(isBatchFlags) != 2 {
		t.Fatalf("want 2 calls, got %d", len(isBatchFlags))
	}
	if isBatchFlags[0] {
		t.Error("single request should have IsBatch=false")
	}
	if !isBatchFlags[1] {
		t.Error("batch sub-request should have IsBatch=true")
	}
}
