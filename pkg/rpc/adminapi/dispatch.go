// dispatch.go provides JSON-RPC dispatch for the admin_ namespace.
package adminapi

import (
	"encoding/json"
	"fmt"

	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// DispatchAPI provides JSON-RPC request/response dispatch for admin_ methods.
// It delegates to API for the actual logic and handles
// JSON-RPC request parsing and response formatting.
type DispatchAPI struct {
	inner *API
}

// NewDispatchAPI creates a new admin dispatch API wrapping the given backend.
func NewDispatchAPI(backend Backend) *DispatchAPI {
	return &DispatchAPI{inner: NewAPI(backend)}
}

// HandleAdminRequest dispatches an admin_ namespace JSON-RPC request.
func (a *DispatchAPI) HandleAdminRequest(req *rpctypes.Request) *rpctypes.Response {
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
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeMethodNotFound,
			fmt.Sprintf("method %q not found in admin namespace", req.Method))
	}
}

func (a *DispatchAPI) adminAddPeer(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing enode URL parameter")
	}
	var enodeURL string
	if err := json.Unmarshal(req.Params[0], &enodeURL); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid enode URL: "+err.Error())
	}
	ok, err := a.inner.AddPeer(enodeURL)
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, err.Error())
	}
	return rpctypes.NewSuccessResponse(req.ID, ok)
}

func (a *DispatchAPI) adminRemovePeer(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing enode URL parameter")
	}
	var enodeURL string
	if err := json.Unmarshal(req.Params[0], &enodeURL); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid enode URL: "+err.Error())
	}
	ok, err := a.inner.RemovePeer(enodeURL)
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, err.Error())
	}
	return rpctypes.NewSuccessResponse(req.ID, ok)
}

func (a *DispatchAPI) adminPeers(req *rpctypes.Request) *rpctypes.Response {
	peers, err := a.inner.Peers()
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, err.Error())
	}
	return rpctypes.NewSuccessResponse(req.ID, peers)
}

func (a *DispatchAPI) adminNodeInfo(req *rpctypes.Request) *rpctypes.Response {
	info, err := a.inner.NodeInfo()
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, err.Error())
	}
	return rpctypes.NewSuccessResponse(req.ID, info)
}

func (a *DispatchAPI) adminDatadir(req *rpctypes.Request) *rpctypes.Response {
	dir, err := a.inner.DataDir()
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, err.Error())
	}
	return rpctypes.NewSuccessResponse(req.ID, dir)
}

func (a *DispatchAPI) adminStartRPC(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 2 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams,
			"expected params: [host, port]")
	}
	var host string
	if err := json.Unmarshal(req.Params[0], &host); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid host: "+err.Error())
	}
	var port int
	if err := json.Unmarshal(req.Params[1], &port); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid port: "+err.Error())
	}
	ok, err := a.inner.StartRPC(host, port)
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, err.Error())
	}
	return rpctypes.NewSuccessResponse(req.ID, ok)
}

func (a *DispatchAPI) adminStopRPC(req *rpctypes.Request) *rpctypes.Response {
	ok, err := a.inner.StopRPC()
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, err.Error())
	}
	return rpctypes.NewSuccessResponse(req.ID, ok)
}

func (a *DispatchAPI) adminChainId(req *rpctypes.Request) *rpctypes.Response {
	id, err := a.inner.ChainID()
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, err.Error())
	}
	return rpctypes.NewSuccessResponse(req.ID, id)
}
