// eclipse_resistance.go implements eclipse attack resistance for the Portal
// network. It enforces peer diversity by limiting the number of peers from
// the same IP subnet, scores peers for Sybil likelihood, and validates
// inbound connections against eclipse criteria.
//
// Eclipse attacks attempt to isolate a node by monopolizing all its peer
// connections. This module defends against such attacks by:
//   - Limiting peers per /16 subnet (IP diversity)
//   - Scoring peers for Sybil indicators (same prefix, rapid joins)
//   - Validating inbound connections before acceptance
//
// Reference: https://www.cs.bu.edu/~goldbe/projects/eclipseEth.pdf
package portal

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Eclipse resistance errors.
var (
	ErrEclipseSubnetFull    = errors.New("portal/eclipse: max peers per subnet reached")
	ErrEclipseDuplicatePeer = errors.New("portal/eclipse: peer already registered")
	ErrEclipseInvalidIP     = errors.New("portal/eclipse: invalid IP address")
	ErrEclipseSybilDetected = errors.New("portal/eclipse: peer flagged as likely Sybil")
	ErrEclipsePeerNotFound  = errors.New("portal/eclipse: peer not found")
	ErrEclipseZeroSubnetCap = errors.New("portal/eclipse: max peers per subnet must be >= 1")
)

// Default eclipse resistance constants.
const (
	DefaultMaxPeersPerSubnet  = 4
	DefaultMinDistinctSubnets = 3
	DefaultSybilThreshold     = 0.7
	DefaultDiversityInterval  = 30 * time.Second
	// rapidJoinWindow is the time window for detecting rapid join patterns.
	rapidJoinWindow = 60 // seconds
)

// EclipseResistanceConfig configures eclipse attack resistance parameters.
type EclipseResistanceConfig struct {
	// MaxPeersPerSubnet is the maximum peers allowed from the same /16 subnet.
	MaxPeersPerSubnet int
	// MinDistinctSubnets is the minimum number of distinct /16 subnets required.
	MinDistinctSubnets int
	// SybilThreshold is the score above which a peer is considered Sybil.
	SybilThreshold float64
	// DiversityCheckInterval is how often diversity checks are performed.
	DiversityCheckInterval time.Duration
}

// NewEclipseResistanceConfig returns a config with default values.
func NewEclipseResistanceConfig() *EclipseResistanceConfig {
	return &EclipseResistanceConfig{
		MaxPeersPerSubnet:      DefaultMaxPeersPerSubnet,
		MinDistinctSubnets:     DefaultMinDistinctSubnets,
		SybilThreshold:         DefaultSybilThreshold,
		DiversityCheckInterval: DefaultDiversityInterval,
	}
}

// peerEntry is an internal record for a registered peer.
type peerEntry struct {
	peerID string
	ip     string
	subnet string // /16 prefix
}

// DiversityStats holds statistics about peer diversity.
type DiversityStats struct {
	// TotalPeers is the total number of connected peers.
	TotalPeers int
	// UniqueSubnets is the number of distinct /16 subnets.
	UniqueSubnets int
	// SubnetDistribution maps each /16 subnet to its peer count.
	SubnetDistribution map[string]int
	// MaxPeersInSubnet is the highest peer count in any single subnet.
	MaxPeersInSubnet int
}

// PeerDiversityValidator enforces IP/subnet diversity among connected peers.
// It tracks peers by their /16 subnet prefix and rejects new connections
// that would exceed the per-subnet limit.
type PeerDiversityValidator struct {
	mu      sync.Mutex
	config  *EclipseResistanceConfig
	peers   map[string]*peerEntry // peerID -> entry
	subnets map[string]int        // subnet -> count
}

// NewPeerDiversityValidator creates a new validator with the given config.
func NewPeerDiversityValidator(cfg *EclipseResistanceConfig) *PeerDiversityValidator {
	return &PeerDiversityValidator{
		config:  cfg,
		peers:   make(map[string]*peerEntry),
		subnets: make(map[string]int),
	}
}

// extractSubnet extracts the /16 subnet prefix from an IP address string.
// For IPv4, this returns the first two octets (e.g., "192.168" from "192.168.1.1").
// For IPv6, returns the first 4 hex groups. Returns an error for invalid IPs.
func extractSubnet(ipStr string) (string, error) {
	// Strip port if present.
	host := ipStr
	if h, _, err := net.SplitHostPort(ipStr); err == nil {
		host = h
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "", fmt.Errorf("%w: %s", ErrEclipseInvalidIP, ipStr)
	}

	// Check for IPv4.
	if v4 := ip.To4(); v4 != nil {
		return fmt.Sprintf("%d.%d", v4[0], v4[1]), nil
	}

	// IPv6: use first 32 bits (4 hex groups = /32 prefix).
	ip16 := ip.To16()
	if ip16 == nil {
		return "", fmt.Errorf("%w: %s", ErrEclipseInvalidIP, ipStr)
	}
	return fmt.Sprintf("%x:%x", uint16(ip16[0])<<8|uint16(ip16[1]),
		uint16(ip16[2])<<8|uint16(ip16[3])), nil
}

// ValidatePeer checks whether a peer can be added without violating
// diversity rules. Does not add the peer.
func (v *PeerDiversityValidator) ValidatePeer(peerID string, ip string) error {
	subnet, err := extractSubnet(ip)
	if err != nil {
		return err
	}

	if v.config.MaxPeersPerSubnet < 1 {
		return ErrEclipseZeroSubnetCap
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	if _, exists := v.peers[peerID]; exists {
		return ErrEclipseDuplicatePeer
	}

	count := v.subnets[subnet]
	if count >= v.config.MaxPeersPerSubnet {
		return fmt.Errorf("%w: subnet %s has %d peers (max %d)",
			ErrEclipseSubnetFull, subnet, count, v.config.MaxPeersPerSubnet)
	}

	return nil
}

// AddPeer validates and adds a peer. Returns an error if the peer violates
// diversity rules or is already registered.
func (v *PeerDiversityValidator) AddPeer(peerID string, ip string) error {
	subnet, err := extractSubnet(ip)
	if err != nil {
		return err
	}

	if v.config.MaxPeersPerSubnet < 1 {
		return ErrEclipseZeroSubnetCap
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	if _, exists := v.peers[peerID]; exists {
		return ErrEclipseDuplicatePeer
	}

	count := v.subnets[subnet]
	if count >= v.config.MaxPeersPerSubnet {
		return fmt.Errorf("%w: subnet %s has %d peers (max %d)",
			ErrEclipseSubnetFull, subnet, count, v.config.MaxPeersPerSubnet)
	}

	v.peers[peerID] = &peerEntry{
		peerID: peerID,
		ip:     ip,
		subnet: subnet,
	}
	v.subnets[subnet]++
	return nil
}

// RemovePeer removes a peer by ID and decrements the subnet count.
func (v *PeerDiversityValidator) RemovePeer(peerID string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	entry, exists := v.peers[peerID]
	if !exists {
		return
	}

	delete(v.peers, peerID)
	v.subnets[entry.subnet]--
	if v.subnets[entry.subnet] <= 0 {
		delete(v.subnets, entry.subnet)
	}
}

// Stats returns diversity statistics about the current peer set.
func (v *PeerDiversityValidator) Stats() DiversityStats {
	v.mu.Lock()
	defer v.mu.Unlock()

	dist := make(map[string]int, len(v.subnets))
	maxInSubnet := 0
	for subnet, count := range v.subnets {
		dist[subnet] = count
		if count > maxInSubnet {
			maxInSubnet = count
		}
	}

	return DiversityStats{
		TotalPeers:         len(v.peers),
		UniqueSubnets:      len(v.subnets),
		SubnetDistribution: dist,
		MaxPeersInSubnet:   maxInSubnet,
	}
}

// PeerCount returns the number of registered peers.
func (v *PeerDiversityValidator) PeerCount() int {
	v.mu.Lock()
	defer v.mu.Unlock()
	return len(v.peers)
}

// SybilScorer scores peers for likelihood of being Sybil nodes based on
// IP prefix similarity, similar node IDs, and rapid join patterns.
type SybilScorer struct {
	mu        sync.Mutex
	threshold float64
	// joinTimes tracks recent peer join times for rapid-join detection.
	joinTimes map[string][]int64 // subnet -> list of join timestamps
}

// NewSybilScorer creates a new Sybil scorer with the given threshold.
// Peers with a score above the threshold are considered Sybil.
func NewSybilScorer(threshold float64) *SybilScorer {
	return &SybilScorer{
		threshold: threshold,
		joinTimes: make(map[string][]int64),
	}
}

// ScorePeer computes a Sybil likelihood score for a peer in [0.0, 1.0].
// Higher scores indicate higher Sybil likelihood.
//
// Scoring factors:
//   - IP prefix clustering: if many recent joins from same /16 subnet
//   - Rapid join pattern: joins within a short time window score higher
//   - Localhost/private IPs get a slight penalty
func (s *SybilScorer) ScorePeer(peerID string, ip string, joinTime int64) float64 {
	subnet, err := extractSubnet(ip)
	if err != nil {
		return 0.0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	score := 0.0

	// Factor 1: IP prefix clustering. Count recent joins from same subnet.
	recentJoins := s.recentJoinsInSubnet(subnet, joinTime)
	if recentJoins > 0 {
		// Scale: 1 recent = 0.2, 2 = 0.35, 5+ = 0.5 (capped).
		clusterScore := float64(recentJoins) * 0.15
		if clusterScore > 0.5 {
			clusterScore = 0.5
		}
		score += clusterScore
	}

	// Factor 2: Rapid join timing. More recent join = higher score.
	if len(s.joinTimes[subnet]) > 0 {
		latest := s.joinTimes[subnet][len(s.joinTimes[subnet])-1]
		timeDiff := joinTime - latest
		if timeDiff >= 0 && timeDiff < int64(rapidJoinWindow) {
			// Very rapid join (within 60s of last from same subnet).
			rapidScore := 0.3 * (1.0 - float64(timeDiff)/float64(rapidJoinWindow))
			score += rapidScore
		}
	}

	// Factor 3: Localhost / private IP penalty.
	if isPrivateOrLocal(ip) {
		score += 0.1
	}

	// Record this join time.
	s.joinTimes[subnet] = append(s.joinTimes[subnet], joinTime)

	// Cap at 1.0.
	if score > 1.0 {
		score = 1.0
	}
	return score
}

// IsSybil returns true if the peer's Sybil score exceeds the threshold.
func (s *SybilScorer) IsSybil(peerID string, ip string, joinTime int64) bool {
	score := s.ScorePeer(peerID, ip, joinTime)
	return score >= s.threshold
}

// Threshold returns the configured Sybil detection threshold.
func (s *SybilScorer) Threshold() float64 {
	return s.threshold
}

// recentJoinsInSubnet counts joins from the subnet within the rapid join window.
// Caller must hold s.mu.
func (s *SybilScorer) recentJoinsInSubnet(subnet string, now int64) int {
	times := s.joinTimes[subnet]
	count := 0
	cutoff := now - int64(rapidJoinWindow)
	for _, t := range times {
		if t >= cutoff {
			count++
		}
	}
	return count
}

// isPrivateOrLocal returns true if the IP is a loopback or private address.
func isPrivateOrLocal(ipStr string) bool {
	host := ipStr
	if h, _, err := net.SplitHostPort(ipStr); err == nil {
		host = h
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	if ip.IsLoopback() {
		return true
	}

	// Check common private ranges.
	if v4 := ip.To4(); v4 != nil {
		if v4[0] == 10 {
			return true
		}
		if v4[0] == 172 && v4[1] >= 16 && v4[1] <= 31 {
			return true
		}
		if v4[0] == 192 && v4[1] == 168 {
			return true
		}
	}

	// IPv6 link-local.
	if strings.HasPrefix(ip.String(), "fe80") {
		return true
	}

	return false
}

// InboundValidator validates inbound peer connections against eclipse criteria.
// It combines diversity checking and Sybil scoring.
type InboundValidator struct {
	diversity *PeerDiversityValidator
	sybil     *SybilScorer
}

// NewInboundValidator creates a new inbound connection validator.
func NewInboundValidator(diversity *PeerDiversityValidator, sybil *SybilScorer) *InboundValidator {
	return &InboundValidator{
		diversity: diversity,
		sybil:     sybil,
	}
}

// ValidateInbound checks whether an inbound peer should be accepted.
// It checks both diversity rules and Sybil scoring.
func (v *InboundValidator) ValidateInbound(peerID string, ip string, joinTime int64) error {
	// Check diversity first.
	if err := v.diversity.ValidatePeer(peerID, ip); err != nil {
		return err
	}

	// Check Sybil score.
	if v.sybil.IsSybil(peerID, ip, joinTime) {
		return fmt.Errorf("%w: peer %s from %s", ErrEclipseSybilDetected, peerID, ip)
	}

	return nil
}
