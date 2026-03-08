package p2p

// gossip_compat.go re-exports types from p2p/gossip for backward compatibility.

import "github.com/eth2030/eth2030/p2p/gossip"

// Gossip type aliases.
type (
	GossipTopic        = gossip.GossipTopic
	MessageID          = gossip.MessageID
	TopicParams        = gossip.TopicParams
	TopicHandler       = gossip.TopicHandler
	TopicScoreSnapshot = gossip.TopicScoreSnapshot
	TopicManager       = gossip.TopicManager

	GossipV2ScoreParams     = gossip.GossipV2ScoreParams
	PeerV2Score             = gossip.PeerV2Score
	GossipV2Scorer          = gossip.GossipV2Scorer
	OpportunisticGrafter    = gossip.OpportunisticGrafter
	MessagePriority         = gossip.MessagePriority
	PrioritizedMessage      = gossip.PrioritizedMessage
	PrioritizedGossipRouter = gossip.PrioritizedGossipRouter
	GossipParamsByTopic     = gossip.GossipParamsByTopic
	MeshScoreParams         = gossip.MeshScoreParams
	TopicPeerScore          = gossip.TopicPeerScore
	MeshDecayConfig         = gossip.MeshDecayConfig
	MeshBanConfig           = gossip.MeshBanConfig
	GossipMeshScoreManager  = gossip.GossipMeshScoreManager
)

// Gossip constants.
const (
	BeaconBlock               = gossip.BeaconBlock
	BeaconAggregateAndProof   = gossip.BeaconAggregateAndProof
	VoluntaryExit             = gossip.VoluntaryExit
	ProposerSlashing          = gossip.ProposerSlashing
	AttesterSlashing          = gossip.AttesterSlashing
	BlobSidecar               = gossip.BlobSidecar
	SyncCommitteeContribution = gossip.SyncCommitteeContribution
	STARKMempoolTick          = gossip.STARKMempoolTick
	PQAggRequest              = gossip.PQAggRequest
	PQAggResult               = gossip.PQAggResult
	ProposerPreferences       = gossip.ProposerPreferences
	MessageIDSize             = gossip.MessageIDSize
	MaxPayloadSize            = gossip.MaxPayloadSize
	GossipPriorityLow         = gossip.GossipPriorityLow
	GossipPriorityMedium      = gossip.GossipPriorityMedium
	GossipPriorityHigh        = gossip.GossipPriorityHigh
)

// Gossip variables.
var (
	MessageDomainValidSnappy   = gossip.MessageDomainValidSnappy
	MessageDomainInvalidSnappy = gossip.MessageDomainInvalidSnappy
	TopicMessageSizeLimit      = gossip.TopicMessageSizeLimit
	ErrTopicMsgTooLarge        = gossip.ErrTopicMsgTooLarge
	ErrTopicNotSubscribed      = gossip.ErrTopicNotSubscribed
	ErrTopicAlreadySubscribed  = gossip.ErrTopicAlreadySubscribed
	ErrTopicManagerClosed      = gossip.ErrTopicManagerClosed
	ErrTopicNilHandler         = gossip.ErrTopicNilHandler
	ErrTopicEmptyData          = gossip.ErrTopicEmptyData
	ErrTopicDuplicateMessage   = gossip.ErrTopicDuplicateMessage
	ErrTopicDataTooLarge       = gossip.ErrTopicDataTooLarge
)

// Gossip function wrappers.
func ParseGossipTopic(name string) (GossipTopic, error) { return gossip.ParseGossipTopic(name) }
func ComputeMessageID(data []byte) MessageID            { return gossip.ComputeMessageID(data) }
func ComputeInvalidMessageID(data []byte) MessageID     { return gossip.ComputeInvalidMessageID(data) }
func DefaultTopicParams() TopicParams                   { return gossip.DefaultTopicParams() }
func NewTopicManager(params TopicParams) *TopicManager  { return gossip.NewTopicManager(params) }
func DefaultGossipV2ScoreParams() *GossipV2ScoreParams  { return gossip.DefaultGossipV2ScoreParams() }
func NewGossipV2Scorer(params *GossipV2ScoreParams) *GossipV2Scorer {
	return gossip.NewGossipV2Scorer(params)
}
func NewOpportunisticGrafter(scorer *GossipV2Scorer, targetD, dLow int) *OpportunisticGrafter {
	return gossip.NewOpportunisticGrafter(scorer, targetD, dLow)
}
func TopicPriority(topic GossipTopic) MessagePriority { return gossip.TopicPriority(topic) }
func NewPrioritizedGossipRouter() *PrioritizedGossipRouter {
	return gossip.NewPrioritizedGossipRouter()
}
func DefaultGossipParamsByTopic() GossipParamsByTopic { return gossip.DefaultGossipParamsByTopic() }
func DefaultMeshScoreParams() MeshScoreParams         { return gossip.DefaultMeshScoreParams() }
func DefaultMeshDecayConfig() MeshDecayConfig         { return gossip.DefaultMeshDecayConfig() }
func DefaultMeshBanConfig() MeshBanConfig             { return gossip.DefaultMeshBanConfig() }
func NewGossipMeshScoreManager(params MeshScoreParams, decay MeshDecayConfig, ban MeshBanConfig) *GossipMeshScoreManager {
	return gossip.NewGossipMeshScoreManager(params, decay, ban)
}

// GossipManager types.
type (
	GossipConfig        = gossip.GossipConfig
	GossipMessage       = gossip.GossipMessage
	GossipSubscription  = gossip.GossipSubscription
	GossipManager       = gossip.GossipManager
	GossipScoreWeights  = gossip.GossipScoreWeights
	GossipScoreConfig   = gossip.GossipScoreConfig
	GossipScorerSnapshot = gossip.GossipScorerSnapshot
	GossipScoreManager  = gossip.GossipScoreManager
	MsgValidatorConfig  = gossip.MsgValidatorConfig
	GossipMsgEnvelope   = gossip.GossipMsgEnvelope
	ValidationResult    = gossip.ValidationResult
	MsgValidator        = gossip.MsgValidator
	MsgValidatorStats   = gossip.MsgValidatorStats
)

// GossipManager errors.
var (
	ErrGossipClosed        = gossip.ErrGossipClosed
	ErrGossipNilMsg        = gossip.ErrGossipNilMsg
	ErrGossipEmptyTopic    = gossip.ErrGossipEmptyTopic
	ErrGossipEmptyData     = gossip.ErrGossipEmptyData
	ErrGossipMsgTooLarge   = gossip.ErrGossipMsgTooLarge
	ErrGossipZeroSender    = gossip.ErrGossipZeroSender
	ErrGossipZeroTimestamp = gossip.ErrGossipZeroTimestamp
	ErrGossipZeroMessageID = gossip.ErrGossipZeroMessageID
	ErrGossipSubNotFound   = gossip.ErrGossipSubNotFound
	ErrGossipSubInactive   = gossip.ErrGossipSubInactive
	ErrGossipTopicMismatch = gossip.ErrGossipTopicMismatch
)

// MsgValidator errors.
var (
	ErrMsgValNilMessage   = gossip.ErrMsgValNilMessage
	ErrMsgValEmptyPayload = gossip.ErrMsgValEmptyPayload
	ErrMsgValTooLarge     = gossip.ErrMsgValTooLarge
	ErrMsgValStale        = gossip.ErrMsgValStale
	ErrMsgValFuture       = gossip.ErrMsgValFuture
	ErrMsgValRateLimited  = gossip.ErrMsgValRateLimited
	ErrMsgValDuplicate    = gossip.ErrMsgValDuplicate
	ErrMsgValInvalidSig   = gossip.ErrMsgValInvalidSig
	ErrMsgValZeroSender   = gossip.ErrMsgValZeroSender
	ErrMsgValZeroMsgID    = gossip.ErrMsgValZeroMsgID
)

// GossipManager constructors.
func DefaultGossipConfig() GossipConfig             { return gossip.DefaultGossipConfig() }
func NewGossipManager(cfg GossipConfig) *GossipManager { return gossip.NewGossipManager(cfg) }
func DefaultGossipScoreConfig() GossipScoreConfig   { return gossip.DefaultGossipScoreConfig() }
func NewGossipScoreManager(cfg GossipScoreConfig, scorer *PeerScorer) *GossipScoreManager {
	return gossip.NewGossipScoreManager(cfg, scorer)
}
func DefaultMsgValidatorConfig() MsgValidatorConfig  { return gossip.DefaultMsgValidatorConfig() }
func NewMsgValidator(cfg MsgValidatorConfig) *MsgValidator { return gossip.NewMsgValidator(cfg) }
