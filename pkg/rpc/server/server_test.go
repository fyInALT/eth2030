package rpcserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// stubHandler is a minimal RequestHandler for server tests.
type stubHandler struct{}

func (s *stubHandler) HandleRequest(req *rpctypes.Request) *rpctypes.Response {
	switch req.Method {
	case "eth_chainId":
		return rpctypes.NewSuccessResponse(req.ID, "0x539")
	case "eth_blockNumber":
		return rpctypes.NewSuccessResponse(req.ID, "0x2a")
	case "eth_gasPrice":
		return rpctypes.NewSuccessResponse(req.ID, "0x3b9aca00")
	default:
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeMethodNotFound,
			"method not found: "+req.Method)
	}
}

func newStubHandler() *stubHandler { return &stubHandler{} }

func TestServerHandler_ValidRequest(t *testing.T) {
	srv := NewServer(newStubHandler())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	body := `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`
	resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("HTTP POST failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("want Content-Type application/json, got %s", ct)
	}

	var rpcResp rpctypes.Response
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if rpcResp.Error != nil {
		t.Fatalf("RPC error: %s", rpcResp.Error.Message)
	}
	if rpcResp.JSONRPC != "2.0" {
		t.Fatalf("want jsonrpc 2.0, got %s", rpcResp.JSONRPC)
	}
}

func TestServerHandler_MethodNotAllowed(t *testing.T) {
	srv := NewServer(newStubHandler())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("HTTP GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", resp.StatusCode)
	}
}

func TestServerHandler_InvalidJSON(t *testing.T) {
	srv := NewServer(newStubHandler())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	body := `not-json`
	resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("HTTP POST failed: %v", err)
	}
	defer resp.Body.Close()

	var rpcResp rpctypes.Response
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected parse error")
	}
	if rpcResp.Error.Code != rpctypes.ErrCodeParse {
		t.Fatalf("want error code %d, got %d", rpctypes.ErrCodeParse, rpcResp.Error.Code)
	}
}

func TestServerHandler_MethodNotFound(t *testing.T) {
	srv := NewServer(newStubHandler())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	body := `{"jsonrpc":"2.0","method":"eth_unknown","params":[],"id":1}`
	resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("HTTP POST failed: %v", err)
	}
	defer resp.Body.Close()

	var rpcResp rpctypes.Response
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected method not found error")
	}
	if rpcResp.Error.Code != rpctypes.ErrCodeMethodNotFound {
		t.Fatalf("want error code %d, got %d", rpctypes.ErrCodeMethodNotFound, rpcResp.Error.Code)
	}
}

func TestServerHandler_MultipleRequests(t *testing.T) {
	srv := NewServer(newStubHandler())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	methods := []string{"eth_chainId", "eth_blockNumber", "eth_gasPrice"}
	for _, method := range methods {
		body := `{"jsonrpc":"2.0","method":"` + method + `","params":[],"id":1}`
		resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(body))
		if err != nil {
			t.Fatalf("HTTP POST for %s: %v", method, err)
		}

		var rpcResp rpctypes.Response
		if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
			t.Fatalf("decode response for %s: %v", method, err)
		}
		resp.Body.Close()

		if rpcResp.Error != nil {
			t.Fatalf("RPC error for %s: %s", method, rpcResp.Error.Message)
		}
	}
}

func TestServerHandler_EmptyBody(t *testing.T) {
	srv := NewServer(newStubHandler())
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL, "application/json", bytes.NewBufferString(""))
	if err != nil {
		t.Fatalf("HTTP POST failed: %v", err)
	}
	defer resp.Body.Close()

	var rpcResp rpctypes.Response
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected error for empty body")
	}
}

func TestNewServer(t *testing.T) {
	srv := NewServer(newStubHandler())
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
	if srv.Handler() == nil {
		t.Fatal("Handler() returned nil")
	}
}
