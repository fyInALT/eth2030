package das

// blobpool_compat.go re-exports types from das/blobpool for backward compatibility.

import "github.com/eth2030/eth2030/das/blobpool"

// SparseBlobPool type alias.
type SparseBlobPool = blobpool.SparseBlobPool

// PoolStats type alias.
type PoolStats = blobpool.PoolStats

// NewSparseBlobPool creates a new SparseBlobPool with the given sparsity factor.
var NewSparseBlobPool = blobpool.NewSparseBlobPool
