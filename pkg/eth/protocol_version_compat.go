package eth

// protocol_version_compat.go re-exports types from eth/ethversion for backward compatibility.

import "github.com/eth2030/eth2030/eth/ethversion"

// Type aliases.
type (
	ProtocolVersion = ethversion.ProtocolVersion
	VersionManager  = ethversion.VersionManager
)

// Version variables.
var (
	ETH66Version = ethversion.ETH66Version
	ETH67Version = ethversion.ETH67Version
	ETH68Version = ethversion.ETH68Version
)

// Errors.
var (
	ErrNoCommonVersion = ethversion.ErrNoCommonVersion
	ErrPeerNotFound    = ethversion.ErrPeerNotFound
	ErrNoVersions      = ethversion.ErrNoVersions
)

// NewVersionManager creates a new VersionManager.
func NewVersionManager(supported []ProtocolVersion) *VersionManager {
	return ethversion.NewVersionManager(supported)
}
