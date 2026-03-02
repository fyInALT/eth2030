package node

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// dummyHandler returns 200 OK for any request.
var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestCORSMiddleware_AllowAll(t *testing.T) {
	h := corsMiddleware(dummyHandler, []string{"*"}, []string{"*"})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Errorf("ACAO header = %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_SpecificDomain(t *testing.T) {
	h := corsMiddleware(dummyHandler, []string{"http://trusted.com"}, []string{"*"})

	// Allowed origin.
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "http://trusted.com")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Header().Get("Access-Control-Allow-Origin") != "http://trusted.com" {
		t.Errorf("expected ACAO header for trusted origin")
	}

	// Disallowed origin: header should not be set but request passes.
	req2 := httptest.NewRequest(http.MethodPost, "/", nil)
	req2.Header.Set("Origin", "http://evil.com")
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if w2.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("ACAO header should be empty for untrusted origin")
	}
	if w2.Code != http.StatusOK {
		t.Errorf("request should still succeed: status = %d", w2.Code)
	}
}

func TestCORSMiddleware_PrefLight(t *testing.T) {
	h := corsMiddleware(dummyHandler, []string{"*"}, []string{"*"})
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("preflight status = %d, want 200", w.Code)
	}
}

func TestCORSMiddleware_VhostBlocked(t *testing.T) {
	h := corsMiddleware(dummyHandler, []string{"*"}, []string{"allowed.local"})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Host = "evil.com"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("blocked host status = %d, want 403", w.Code)
	}
}

func TestCORSMiddleware_VhostAllowed(t *testing.T) {
	h := corsMiddleware(dummyHandler, []string{"*"}, []string{"allowed.local"})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Host = "allowed.local"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("allowed host status = %d, want 200", w.Code)
	}
}

func TestEnsureJWTSecret_Creates(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.JWTSecret = ""

	if err := ensureJWTSecret(&cfg); err != nil {
		t.Fatalf("ensureJWTSecret() error: %v", err)
	}

	path := filepath.Join(dir, "jwtsecret")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("jwt secret file not created: %v", err)
	}
	content := strings.TrimSpace(string(data))
	if !strings.HasPrefix(content, "0x") {
		t.Errorf("jwt secret should start with 0x, got %q", content)
	}
	if len(content) != 66 { // 0x + 64 hex chars
		t.Errorf("jwt secret length = %d, want 66", len(content))
	}
}

func TestEnsureJWTSecret_Idempotent(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = dir

	if err := ensureJWTSecret(&cfg); err != nil {
		t.Fatalf("first ensureJWTSecret() error: %v", err)
	}
	first, _ := os.ReadFile(filepath.Join(dir, "jwtsecret"))

	if err := ensureJWTSecret(&cfg); err != nil {
		t.Fatalf("second ensureJWTSecret() error: %v", err)
	}
	second, _ := os.ReadFile(filepath.Join(dir, "jwtsecret"))

	if string(first) != string(second) {
		t.Error("second call should not overwrite existing jwt secret")
	}
}

func TestEnsureJWTSecret_ExplicitPath(t *testing.T) {
	dir := t.TempDir()
	jwtPath := filepath.Join(dir, "custom_jwt.hex")
	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.JWTSecret = jwtPath

	if err := ensureJWTSecret(&cfg); err != nil {
		t.Fatalf("ensureJWTSecret() error: %v", err)
	}
	if _, err := os.Stat(jwtPath); err != nil {
		t.Errorf("custom jwt path not created: %v", err)
	}
}
