package rpc

import (
	"encoding/json"
	"io"
	"net/http"
)

// Server is a JSON-RPC HTTP server that dispatches requests to the EthAPI.
type Server struct {
	api      *EthAPI
	adminAPI *AdminDispatchAPI
	mux      *http.ServeMux
}

// NewServer creates a new JSON-RPC server.
func NewServer(backend Backend) *Server {
	s := &Server{
		api: NewEthAPI(backend),
		mux: http.NewServeMux(),
	}
	s.mux.HandleFunc("/", s.handleRPC)
	return s
}

// SetAdminBackend wires an AdminBackend so that admin_* methods are served.
func (s *Server) SetAdminBackend(b AdminBackend) {
	s.adminAPI = NewAdminDispatchAPI(b)
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
		writeError(w, nil, ErrCodeParse, "failed to read request body")
		return
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, nil, ErrCodeParse, "invalid JSON")
		return
	}

	var resp *Response
	if isAdminMethod(req.Method) && s.adminAPI != nil {
		resp = s.adminAPI.HandleAdminRequest(&req)
	} else {
		resp = s.api.HandleRequest(&req)
	}
	writeJSON(w, resp)
}

// isAdminMethod reports whether the JSON-RPC method belongs to the admin namespace.
func isAdminMethod(method string) bool {
	return len(method) > 6 && method[:6] == "admin_"
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, id json.RawMessage, code int, message string) {
	resp := &Response{
		JSONRPC: "2.0",
		Error:   &RPCError{Code: code, Message: message},
		ID:      id,
	}
	writeJSON(w, resp)
}
