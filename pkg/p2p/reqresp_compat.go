package p2p

// reqresp_compat.go re-exports types from p2p/reqresp for backward compatibility.
// Consumers should migrate to importing p2p/reqresp directly.

import "github.com/eth2030/eth2030/p2p/reqresp"

// ReqResp codec types.
type (
	RequestID   = reqresp.RequestID
	Request     = reqresp.Request
	Response    = reqresp.Response
	ReqRespCodec = reqresp.ReqRespCodec
	ReqRespConfig = reqresp.ReqRespConfig
)

// Protocol types.
type (
	MethodID               = reqresp.MethodID
	ResponseCode           = reqresp.ResponseCode
	StatusMessage          = reqresp.StatusMessage
	GoodbyeReason          = reqresp.GoodbyeReason
	ProtocolRequest        = reqresp.ProtocolRequest
	ProtocolResponse       = reqresp.ProtocolResponse
	ResponseChunk          = reqresp.ResponseChunk
	StreamedResponse       = reqresp.StreamedResponse
	ReqHandler             = reqresp.ReqHandler
	StreamingRequestHandler = reqresp.StreamingRequestHandler
	ProtocolConfig         = reqresp.ProtocolConfig
	ReqRespProtocol        = reqresp.ReqRespProtocol
)

// Beacon message types.
type (
	BeaconBlocksByRangeRequest = reqresp.BeaconBlocksByRangeRequest
	BeaconBlocksByRootRequest  = reqresp.BeaconBlocksByRootRequest
	BlobSidecarsByRangeRequest = reqresp.BlobSidecarsByRangeRequest
	DataColumnsByRangeRequest  = reqresp.DataColumnsByRangeRequest
	SSZChunk                   = reqresp.SSZChunk
)

// Retry and manager types.
type (
	RetryConfig          = reqresp.RetryConfig
	ReqRespManager       = reqresp.ReqRespManager
	RequestManagerConfig = reqresp.RequestManagerConfig
	OutboundRequest      = reqresp.OutboundRequest
	RequestManager       = reqresp.RequestManager
)

// Request handler types.
type (
	MessageHandler      = reqresp.MessageHandler
	MessageHandlerFunc  = reqresp.MessageHandlerFunc
	RequestHandlerStats = reqresp.RequestHandlerStats
	RequestHandler      = reqresp.RequestHandler
)

// ReqResp codec errors.
var (
	ErrRequestTooLarge = reqresp.ErrRequestTooLarge
	ErrInvalidEncoding = reqresp.ErrInvalidEncoding
	ErrMethodTooLong   = reqresp.ErrMethodTooLong
)

// Protocol errors.
var (
	ErrProtocolClosed      = reqresp.ErrProtocolClosed
	ErrProtocolNoHandler   = reqresp.ErrProtocolNoHandler
	ErrProtocolTimeout     = reqresp.ErrProtocolTimeout
	ErrProtocolRateLimited = reqresp.ErrProtocolRateLimited
	ErrProtocolConcurrency = reqresp.ErrProtocolConcurrency
	ErrProtocolNilPayload  = reqresp.ErrProtocolNilPayload
	ErrProtocolInvalidResp = reqresp.ErrProtocolInvalidResp
)

// Request manager errors.
var (
	ErrReqMgrClosed     = reqresp.ErrReqMgrClosed
	ErrReqMgrMaxPending = reqresp.ErrReqMgrMaxPending
	ErrReqMgrNotFound   = reqresp.ErrReqMgrNotFound
	ErrReqMgrMaxRetries = reqresp.ErrReqMgrMaxRetries
	ErrReqMgrDupRequest = reqresp.ErrReqMgrDupRequest
)

// ReqRespManager errors.
var (
	ErrReqRespMaxRetries = reqresp.ErrReqRespMaxRetries
	ErrReqRespTimeout    = reqresp.ErrReqRespTimeout
	ErrReqRespClosed     = reqresp.ErrReqRespClosed
)

// Request handler errors.
var (
	ErrNoMessageHandler  = reqresp.ErrNoMessageHandler
	ErrHandlerTimeout    = reqresp.ErrHandlerTimeout
	ErrNilMessageHandler = reqresp.ErrNilMessageHandler
)

// Constants.
const (
	DefaultHandlerTimeout             = reqresp.DefaultHandlerTimeout
	MaxConcurrentRequestsPerProtocol  = reqresp.MaxConcurrentRequestsPerProtocol
)

// Constructors.
func NewReqRespCodec(config ReqRespConfig) *ReqRespCodec { return reqresp.NewReqRespCodec(config) }
func DefaultReqRespConfig() ReqRespConfig                { return reqresp.DefaultReqRespConfig() }
func NewReqRespProtocol(config ProtocolConfig) *ReqRespProtocol {
	return reqresp.NewReqRespProtocol(config)
}
func DefaultProtocolConfig() ProtocolConfig { return reqresp.DefaultProtocolConfig() }
func NewReqRespManager(protocol *ReqRespProtocol, retry RetryConfig) *ReqRespManager {
	return reqresp.NewReqRespManager(protocol, retry)
}
func DefaultRetryConfig() RetryConfig              { return reqresp.DefaultRetryConfig() }
func NewRequestManager(cfg RequestManagerConfig) *RequestManager {
	return reqresp.NewRequestManager(cfg)
}
func DefaultRequestManagerConfig() RequestManagerConfig { return reqresp.DefaultRequestManagerConfig() }
func NewRequestHandler() *RequestHandler               { return reqresp.NewRequestHandler() }

// SSZ encoding helpers.
func EncodeSSZChunk(chunk SSZChunk) []byte             { return reqresp.EncodeSSZChunk(chunk) }
func DecodeSSZChunk(data []byte) (*SSZChunk, int, error) { return reqresp.DecodeSSZChunk(data) }
func DecodeSSZStream(data []byte) ([]SSZChunk, error)  { return reqresp.DecodeSSZStream(data) }

// Beacon request helpers.
func DecodeBeaconBlocksByRange(data []byte) (*BeaconBlocksByRangeRequest, error) {
	return reqresp.DecodeBeaconBlocksByRange(data)
}
func DecodeBeaconBlocksByRoot(data []byte) (*BeaconBlocksByRootRequest, error) {
	return reqresp.DecodeBeaconBlocksByRoot(data)
}
func DecodeBlobSidecarsByRange(data []byte) (*BlobSidecarsByRangeRequest, error) {
	return reqresp.DecodeBlobSidecarsByRange(data)
}
func DecodeDataColumnsByRange(data []byte) (*DataColumnsByRangeRequest, error) {
	return reqresp.DecodeDataColumnsByRange(data)
}
