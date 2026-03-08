package rpc

// api_admin.go re-exports admin types from rpc/adminapi for backward compatibility.

import (
	"github.com/eth2030/eth2030/rpc/adminapi"
)

// AdminBackend re-exports the admin Backend interface.
type AdminBackend = adminapi.Backend

// Re-export admin data types.
type (
	NodeInfoData = adminapi.NodeInfoData
	NodePorts    = adminapi.NodePorts
	PeerInfoData = adminapi.PeerInfoData
	AdminAPI     = adminapi.API

	// AdminDispatchAPI re-exports the admin JSON-RPC dispatch type.
	AdminDispatchAPI = adminapi.DispatchAPI
)

// NewAdminAPI re-exports the admin API constructor.
var NewAdminAPI = adminapi.NewAPI

// NewAdminDispatchAPI re-exports the admin dispatch API constructor.
var NewAdminDispatchAPI = adminapi.NewDispatchAPI
