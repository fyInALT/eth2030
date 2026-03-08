package p2p

// ethproto_compat.go re-exports types from p2p/ethproto for backward compatibility.
// Consumers should migrate to importing p2p/ethproto directly.

import (
	"time"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/p2p/ethproto"
)

// Protocol version constants.
const (
	ETH68 = ethproto.ETH68
	ETH70 = ethproto.ETH70
	ETH71 = ethproto.ETH71
)

// eth message code constants.
const (
	StatusMsg                     = ethproto.StatusMsg
	NewBlockHashesMsg             = ethproto.NewBlockHashesMsg
	TransactionsMsg               = ethproto.TransactionsMsg
	GetBlockHeadersMsg            = ethproto.GetBlockHeadersMsg
	BlockHeadersMsg               = ethproto.BlockHeadersMsg
	GetBlockBodiesMsg             = ethproto.GetBlockBodiesMsg
	BlockBodiesMsg                = ethproto.BlockBodiesMsg
	NewBlockMsg                   = ethproto.NewBlockMsg
	NewPooledTransactionHashesMsg = ethproto.NewPooledTransactionHashesMsg
	GetPooledTransactionsMsg      = ethproto.GetPooledTransactionsMsg
	PooledTransactionsMsg         = ethproto.PooledTransactionsMsg
	GetReceiptsMsg                = ethproto.GetReceiptsMsg
	ReceiptsMsg                   = ethproto.ReceiptsMsg
	GetPartialReceiptsMsg         = ethproto.GetPartialReceiptsMsg
	PartialReceiptsMsg            = ethproto.PartialReceiptsMsg
	GetBlockAccessListsMsg        = ethproto.GetBlockAccessListsMsg
	BlockAccessListsMsg           = ethproto.BlockAccessListsMsg
)

// Handler constants.
const (
	MaxHeadersServe       = ethproto.MaxHeadersServe
	MaxBodiesServe        = ethproto.MaxBodiesServe
	MaxReceiptsServe      = ethproto.MaxReceiptsServe
	MaxPooledTxServe      = ethproto.MaxPooledTxServe
	DefaultRequestTimeout = ethproto.DefaultRequestTimeout
)

// ForkID type alias.
type ForkID = ethproto.ForkID

// Protocol message types.
type (
	StatusData                         = ethproto.StatusData
	NewBlockHashesEntry                = ethproto.NewBlockHashesEntry
	HashOrNumber                       = ethproto.HashOrNumber
	GetBlockHeadersRequest             = ethproto.GetBlockHeadersRequest
	GetBlockHeadersPacket              = ethproto.GetBlockHeadersPacket
	BlockHeadersPacket                 = ethproto.BlockHeadersPacket
	GetBlockBodiesRequest              = ethproto.GetBlockBodiesRequest
	GetBlockBodiesPacket               = ethproto.GetBlockBodiesPacket
	BlockBody                          = ethproto.BlockBody
	BlockBodiesPacket                  = ethproto.BlockBodiesPacket
	NewBlockData                       = ethproto.NewBlockData
	GetReceiptsRequest                 = ethproto.GetReceiptsRequest
	GetReceiptsPacket                  = ethproto.GetReceiptsPacket
	ReceiptsPacket                     = ethproto.ReceiptsPacket
	NewPooledTransactionHashesPacket68 = ethproto.NewPooledTransactionHashesPacket68
	GetPooledTransactionsRequest       = ethproto.GetPooledTransactionsRequest
	GetPooledTransactionsPacket        = ethproto.GetPooledTransactionsPacket
	PooledTransactionsPacket           = ethproto.PooledTransactionsPacket
	GetPartialReceiptsPacket           = ethproto.GetPartialReceiptsPacket
	PartialReceiptsPacket              = ethproto.PartialReceiptsPacket
	GetBlockAccessListsPacket          = ethproto.GetBlockAccessListsPacket
	BlockAccessListData                = ethproto.BlockAccessListData
	AccessEntryData                    = ethproto.AccessEntryData
	BlockAccessListsPacket             = ethproto.BlockAccessListsPacket
)

// Handler types.
type (
	HandlerFunc     = ethproto.HandlerFunc
	Backend         = ethproto.Backend
	HandlerRegistry = ethproto.HandlerRegistry
	RequestTracker  = ethproto.RequestTracker
)

// Handler errors.
var (
	ErrRequestTimeout   = ethproto.ErrRequestTimeout
	ErrDuplicateRequest = ethproto.ErrDuplicateRequest
	ErrUnknownRequest   = ethproto.ErrUnknownRequest
	ErrHandlerNotFound  = ethproto.ErrHandlerNotFound
	ErrNilPeer          = ethproto.ErrNilPeer
	ErrNilBackend       = ethproto.ErrNilBackend
)

// Constructors.
func NewHandlerRegistry() *HandlerRegistry { return ethproto.NewHandlerRegistry() }
func NewRequestTracker(timeout time.Duration) *RequestTracker {
	return ethproto.NewRequestTracker(timeout)
}

// ForkID functions.
func CalcForkID(genesisHash types.Hash, head uint64, forkBlocks []uint64) ForkID {
	return ethproto.CalcForkID(genesisHash, head, forkBlocks)
}
func ValidateMessageCode(code uint64) error { return ethproto.ValidateMessageCode(code) }
func MessageName(code uint64) string        { return ethproto.MessageName(code) }
