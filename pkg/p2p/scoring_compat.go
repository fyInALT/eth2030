package p2p

// scoring_compat.go re-exports types from p2p/scoring for backward compatibility.

import "github.com/eth2030/eth2030/p2p/scoring"

// Scoring type aliases.
type (
	BehaviorScorerConfig    = scoring.BehaviorScorerConfig
	BehaviorMetrics         = scoring.BehaviorMetrics
	BehaviorScorer          = scoring.BehaviorScorer
	PeerScoreConfig         = scoring.PeerScoreConfig
	ScoreEvent              = scoring.ScoreEvent
	PeerScoreInfo           = scoring.PeerScoreInfo
	PeerScorer              = scoring.PeerScorer
	ReputationRecord        = scoring.ReputationRecord
	ReputationSystemConfig  = scoring.ReputationSystemConfig
	PeerReputationSystem    = scoring.PeerReputationSystem
	ReputationConfig        = scoring.ReputationConfig
	PeerReputation          = scoring.PeerReputation
	ReputationEvent         = scoring.ReputationEvent
	ReputationTracker       = scoring.ReputationTracker
	PeerScore               = scoring.PeerScore
	ScoreStats              = scoring.ScoreStats
	RepCategory             = scoring.RepCategory
	BanReason               = scoring.BanReason
	RepConfig               = scoring.RepConfig
	PeerRepEntry            = scoring.PeerRepEntry
	PeerRep                 = scoring.PeerRep
	RepBanRecord            = scoring.RepBanRecord
	ScoreMap                = scoring.ScoreMap
)

// Scoring constants.
const (
	ScoreValidBlock       = scoring.ScoreValidBlock
	ScoreInvalidBlock     = scoring.ScoreInvalidBlock
	ScoreValidTx          = scoring.ScoreValidTx
	ScoreInvalidTx        = scoring.ScoreInvalidTx
	ScoreTimedOut         = scoring.ScoreTimedOut
	EventGoodBlock        = scoring.EventGoodBlock
	EventBadBlock         = scoring.EventBadBlock
	EventTimeout          = scoring.EventTimeout
	EventDisconnect       = scoring.EventDisconnect
	EventGoodAttestation  = scoring.EventGoodAttestation
	MaxScore              = scoring.MaxScore
	MinScore              = scoring.MinScore
	DefaultScore          = scoring.DefaultScore
	ScoreDisconnect       = scoring.ScoreDisconnect
	RepCatProtocol        = scoring.RepCatProtocol
	RepCatLatency         = scoring.RepCatLatency
	RepCatBandwidth       = scoring.RepCatBandwidth
	RepCatAvailability    = scoring.RepCatAvailability
	BanReasonNone              = scoring.BanReasonNone
	BanReasonProtocolViolation = scoring.BanReasonProtocolViolation
	BanReasonSpam              = scoring.BanReasonSpam
	BanReasonDoS               = scoring.BanReasonDoS
	BanReasonInvalidBlocks     = scoring.BanReasonInvalidBlocks
	ScoreHandshakeOK           = scoring.ScoreHandshakeOK
	ScoreHandshakeFail         = scoring.ScoreHandshakeFail
)

// Scoring error variables.
var (
	ErrReputationPeerNotFound  = scoring.ErrReputationPeerNotFound
	ErrReputationAlreadyBanned = scoring.ErrReputationAlreadyBanned
	ErrRepPeerNotTracked       = scoring.ErrRepPeerNotTracked
	ErrRepPeerBanned           = scoring.ErrRepPeerBanned
	ErrRepInvalidCategory      = scoring.ErrRepInvalidCategory
)

// Scoring function wrappers.
func DefaultBehaviorScorerConfig() BehaviorScorerConfig { return scoring.DefaultBehaviorScorerConfig() }
func NewBehaviorScorer(config BehaviorScorerConfig) *BehaviorScorer {
	return scoring.NewBehaviorScorer(config)
}
func DefaultPeerScoreConfig() PeerScoreConfig { return scoring.DefaultPeerScoreConfig() }
func NewPeerScorer(config PeerScoreConfig) *PeerScorer {
	return scoring.NewPeerScorer(config)
}
func DefaultReputationSystemConfig() ReputationSystemConfig {
	return scoring.DefaultReputationSystemConfig()
}
func NewPeerReputationSystem(cfg ReputationSystemConfig) *PeerReputationSystem {
	return scoring.NewPeerReputationSystem(cfg)
}
func DefaultReputationConfig() ReputationConfig { return scoring.DefaultReputationConfig() }
func NewReputationTracker(config ReputationConfig) *ReputationTracker {
	return scoring.NewReputationTracker(config)
}
func NewPeerScore() *PeerScore { return scoring.NewPeerScore() }
func DefaultRepConfig() RepConfig { return scoring.DefaultRepConfig() }
func NewPeerRep(cfg RepConfig) *PeerRep { return scoring.NewPeerRep(cfg) }
func NewScoreMap() *ScoreMap             { return scoring.NewScoreMap() }
