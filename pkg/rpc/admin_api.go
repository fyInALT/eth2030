// admin_api.go provides JSON-RPC dispatch for admin namespace methods.
// The dispatch layer lives here (top-level rpc package) so it can access
// successResponse/errorResponse helpers. Core logic is in rpc/adminapi.
package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/eth2030/eth2030/rpc/adminapi"
)

// AdminDispatchAPI provides JSON-RPC dispatch for admin_ namespace methods.
// It delegates to adminapi.API for the actual logic,
// but handles JSON-RPC request/response parsing and formatting.
type AdminDispatchAPI struct {
	inner *adminapi.API
}

// NewAdminDispatchAPI creates a new admin dispatch API wrapping the given
// admin backend.
func NewAdminDispatchAPI(backend AdminBackend) *AdminDispatchAPI {
	return &AdminDispatchAPI{
		inner: adminapi.NewAPI(backend),
	}
}

// HandleAdminRequest dispatches an admin_ namespace JSON-RPC request.
func (a *AdminDispatchAPI) HandleAdminRequest(req *Request) *Response {
	switch req.Method {
	case "admin_addPeer":
		return a.adminAddPeer(req)
	case "admin_removePeer":
		return a.adminRemovePeer(req)
	case "admin_peers":
		return a.adminPeers(req)
	case "admin_nodeInfo":
		return a.adminNodeInfo(req)
	case "admin_datadir":
		return a.adminDatadir(req)
	case "admin_startRPC":
		return a.adminStartRPC(req)
	case "admin_stopRPC":
		return a.adminStopRPC(req)
	case "admin_chainId":
		return a.adminChainId(req)
	default:
		return errorResponse(req.ID, ErrCodeMethodNotFound,
			fmt.Sprintf("method %q not found in admin namespace", req.Method))
	}
}

// adminAddPeer handles admin_addPeer(enode). Returns true on success.
func (a *AdminDispatchAPI) adminAddPeer(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing enode URL parameter")
	}

	var enodeURL string
	if err := json.Unmarshal(req.Params[0], &enodeURL); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid enode URL: "+err.Error())
	}

	ok, err := a.inner.AddPeer(enodeURL)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}
	return successResponse(req.ID, ok)
}

// adminRemovePeer handles admin_removePeer(enode). Returns true on success.
func (a *AdminDispatchAPI) adminRemovePeer(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing enode URL parameter")
	}

	var enodeURL string
	if err := json.Unmarshal(req.Params[0], &enodeURL); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid enode URL: "+err.Error())
	}

	ok, err := a.inner.RemovePeer(enodeURL)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}
	return successResponse(req.ID, ok)
}

// adminPeers handles admin_peers(). Returns list of connected peers.
func (a *AdminDispatchAPI) adminPeers(req *Request) *Response {
	peers, err := a.inner.Peers()
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}
	return successResponse(req.ID, peers)
}

// adminNodeInfo handles admin_nodeInfo(). Returns node information.
func (a *AdminDispatchAPI) adminNodeInfo(req *Request) *Response {
	info, err := a.inner.NodeInfo()
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}
	return successResponse(req.ID, info)
}

// adminDatadir handles admin_datadir(). Returns data directory path.
func (a *AdminDispatchAPI) adminDatadir(req *Request) *Response {
	dir, err := a.inner.DataDir()
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}
	return successResponse(req.ID, dir)
}

// adminStartRPC handles admin_startRPC(host, port).
func (a *AdminDispatchAPI) adminStartRPC(req *Request) *Response {
	if len(req.Params) < 2 {
		return errorResponse(req.ID, ErrCodeInvalidParams,
			"expected params: [host, port]")
	}

	var host string
	if err := json.Unmarshal(req.Params[0], &host); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid host: "+err.Error())
	}

	var port int
	if err := json.Unmarshal(req.Params[1], &port); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid port: "+err.Error())
	}

	ok, err := a.inner.StartRPC(host, port)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}
	return successResponse(req.ID, ok)
}

// adminStopRPC handles admin_stopRPC().
func (a *AdminDispatchAPI) adminStopRPC(req *Request) *Response {
	ok, err := a.inner.StopRPC()
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}
	return successResponse(req.ID, ok)
}

// adminChainId handles admin_chainId().
func (a *AdminDispatchAPI) adminChainId(req *Request) *Response {
	id, err := a.inner.ChainID()
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}
	return successResponse(req.ID, id)
}
