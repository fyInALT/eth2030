// types.go defines local interface and data types used by the downloader
// sub-package. These are consumer-defined (Go structural typing) so any
// concrete type that satisfies the root sync package interfaces also
// satisfies these.
package downloader

import "github.com/eth2030/eth2030/core/types"

// maxFutureTimestamp is the maximum allowed future timestamp for a header (seconds).
const maxFutureTimestamp = 15

// HeaderSource retrieves headers from a remote peer or local chain.
type HeaderSource interface {
	FetchHeaders(from uint64, count int) ([]*types.Header, error)
}

// BodySource retrieves block bodies from a remote peer or local chain.
type BodySource interface {
	FetchBodies(hashes []types.Hash) ([]*types.Body, error)
}

// HeaderData represents a downloaded header (legacy format).
type HeaderData struct {
	Number     uint64
	Hash       [32]byte
	ParentHash [32]byte
	Timestamp  uint64
	RLP        []byte // RLP-encoded header
}

// BlockData represents a downloaded block (legacy format).
type BlockData struct {
	Number    uint64
	Hash      [32]byte
	HeaderRLP []byte
	BodyRLP   []byte
}
