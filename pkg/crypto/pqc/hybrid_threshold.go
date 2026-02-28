// hybrid_threshold.go implements hybrid threshold signature aggregation that
// combines classical (BLS) and post-quantum signatures in a t-of-n threshold
// scheme. Each signer produces both a classical and a PQ signature share;
// the aggregator collects shares and checks whether the threshold is met
// based on configurable signature weighting.
//
// This bridges the transition period where validators may run both classical
// and PQ signing infrastructure simultaneously, allowing the network to
// maintain security guarantees even if one cryptographic family is compromised.
package pqc

import (
	"errors"
	"fmt"
	"sync"
)

// Hybrid threshold errors.
var (
	ErrHTInvalidThreshold = errors.New("hybrid_threshold: threshold t must be >= 1")
	ErrHTInvalidTotal     = errors.New("hybrid_threshold: total n must be >= 1")
	ErrHTThresholdExceeds = errors.New("hybrid_threshold: threshold t must be <= total n")
	ErrHTDuplicateSigner  = errors.New("hybrid_threshold: duplicate signer index")
	ErrHTSignerOutOfRange = errors.New("hybrid_threshold: signer index out of range [0, n)")
	ErrHTEmptyShare       = errors.New("hybrid_threshold: share has neither classical nor PQ signature")
	ErrHTRequireBoth      = errors.New("hybrid_threshold: config requires both classical and PQ signatures")
	ErrHTNilShare         = errors.New("hybrid_threshold: nil share")
	ErrHTEmptyMessage     = errors.New("hybrid_threshold: empty message")
	ErrHTInvalidClassical = errors.New("hybrid_threshold: classical signature is empty")
	ErrHTInvalidPQ        = errors.New("hybrid_threshold: PQ signature is empty")
)

// SignatureWeighting assigns relative weight to classical vs PQ signatures
// for threshold calculation. Weights are in [0.0, 1.0] and should sum to 1.0.
type SignatureWeighting struct {
	// ClassicalWeight is the weight given to a valid classical signature.
	ClassicalWeight float64
	// PQWeight is the weight given to a valid post-quantum signature.
	PQWeight float64
}

// DefaultWeighting returns equal weighting (0.5 each).
func DefaultWeighting() SignatureWeighting {
	return SignatureWeighting{
		ClassicalWeight: 0.5,
		PQWeight:        0.5,
	}
}

// HybridThresholdConfig configures the hybrid threshold aggregation scheme.
type HybridThresholdConfig struct {
	// Threshold is the minimum number of valid shares (t) required.
	Threshold int
	// Total is the total number of signers (n).
	Total int
	// Weighting controls relative weight of classical vs PQ signatures.
	Weighting SignatureWeighting
	// RequireBoth requires each share to have both classical and PQ sigs.
	RequireBoth bool
}

// NewHybridThresholdConfig creates a new config with the given t-of-n params
// and default weighting. Returns an error if t or n are invalid.
func NewHybridThresholdConfig(t, n int) (*HybridThresholdConfig, error) {
	if t < 1 {
		return nil, ErrHTInvalidThreshold
	}
	if n < 1 {
		return nil, ErrHTInvalidTotal
	}
	if t > n {
		return nil, ErrHTThresholdExceeds
	}
	return &HybridThresholdConfig{
		Threshold:   t,
		Total:       n,
		Weighting:   DefaultWeighting(),
		RequireBoth: false,
	}, nil
}

// Validate checks the config for internal consistency.
func (c *HybridThresholdConfig) Validate() error {
	if c.Threshold < 1 {
		return ErrHTInvalidThreshold
	}
	if c.Total < 1 {
		return ErrHTInvalidTotal
	}
	if c.Threshold > c.Total {
		return ErrHTThresholdExceeds
	}
	return nil
}

// HybridThresholdShare is a single signer's contribution. It contains the
// signer's index and their classical and/or PQ signature bytes.
type HybridThresholdShare struct {
	// SignerIndex is the 0-based index of the signer in [0, n).
	SignerIndex int
	// ClassicalSig is the classical (e.g., BLS) signature bytes. May be nil.
	ClassicalSig []byte
	// PQSig is the post-quantum signature bytes. May be nil.
	PQSig []byte
}

// HasClassical returns true if the share contains a non-empty classical signature.
func (s *HybridThresholdShare) HasClassical() bool {
	return len(s.ClassicalSig) > 0
}

// HasPQ returns true if the share contains a non-empty PQ signature.
func (s *HybridThresholdShare) HasPQ() bool {
	return len(s.PQSig) > 0
}

// HybridThresholdResult is the aggregation result.
type HybridThresholdResult struct {
	// MetThreshold indicates whether the threshold has been reached.
	MetThreshold bool
	// ClassicalCount is the number of valid classical signatures collected.
	ClassicalCount int
	// PQCount is the number of valid PQ signatures collected.
	PQCount int
	// TotalShares is the total number of valid shares collected.
	TotalShares int
	// TotalWeight is the accumulated weight from all valid shares.
	TotalWeight float64
	// RequiredWeight is the weight needed to meet threshold.
	RequiredWeight float64
}

// HybridThresholdAggregator aggregates t-of-n hybrid threshold signature shares.
// It is safe for concurrent use.
type HybridThresholdAggregator struct {
	mu     sync.Mutex
	config *HybridThresholdConfig
	shares map[int]*HybridThresholdShare // keyed by signer index
}

// NewHybridThresholdAggregator creates a new aggregator with the given config.
func NewHybridThresholdAggregator(cfg *HybridThresholdConfig) *HybridThresholdAggregator {
	return &HybridThresholdAggregator{
		config: cfg,
		shares: make(map[int]*HybridThresholdShare),
	}
}

// AddShare adds a signer's share to the aggregator. Returns an error if the
// share is invalid, the signer index is out of range, or the signer has
// already contributed a share.
func (a *HybridThresholdAggregator) AddShare(share *HybridThresholdShare) error {
	if share == nil {
		return ErrHTNilShare
	}
	if share.SignerIndex < 0 || share.SignerIndex >= a.config.Total {
		return fmt.Errorf("%w: got %d, max %d", ErrHTSignerOutOfRange, share.SignerIndex, a.config.Total-1)
	}

	// Validate that at least one signature is present.
	if !share.HasClassical() && !share.HasPQ() {
		return ErrHTEmptyShare
	}

	// If RequireBoth is set, both must be present.
	if a.config.RequireBoth {
		if !share.HasClassical() || !share.HasPQ() {
			return ErrHTRequireBoth
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.shares[share.SignerIndex]; exists {
		return fmt.Errorf("%w: index %d", ErrHTDuplicateSigner, share.SignerIndex)
	}

	a.shares[share.SignerIndex] = share
	return nil
}

// MeetsThreshold returns true if enough valid shares have been collected
// to meet the threshold.
func (a *HybridThresholdAggregator) MeetsThreshold() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.shares) >= a.config.Threshold
}

// Result computes and returns the aggregation result.
func (a *HybridThresholdAggregator) Result() *HybridThresholdResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	classicalCount := 0
	pqCount := 0
	totalWeight := 0.0

	for _, share := range a.shares {
		w := a.shareWeight(share)
		totalWeight += w
		if share.HasClassical() {
			classicalCount++
		}
		if share.HasPQ() {
			pqCount++
		}
	}

	requiredWeight := float64(a.config.Threshold)

	return &HybridThresholdResult{
		MetThreshold:   len(a.shares) >= a.config.Threshold,
		ClassicalCount: classicalCount,
		PQCount:        pqCount,
		TotalShares:    len(a.shares),
		TotalWeight:    totalWeight,
		RequiredWeight: requiredWeight,
	}
}

// shareWeight computes the weight of a single share based on which
// signatures it contains and the configured weighting.
func (a *HybridThresholdAggregator) shareWeight(share *HybridThresholdShare) float64 {
	w := 0.0
	if share.HasClassical() {
		w += a.config.Weighting.ClassicalWeight
	}
	if share.HasPQ() {
		w += a.config.Weighting.PQWeight
	}
	return w
}

// Reset clears all collected shares.
func (a *HybridThresholdAggregator) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.shares = make(map[int]*HybridThresholdShare)
}

// ShareCount returns the number of shares collected so far.
func (a *HybridThresholdAggregator) ShareCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.shares)
}

// ValidClassicalCount returns the number of shares that have a classical sig.
func (a *HybridThresholdAggregator) ValidClassicalCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	count := 0
	for _, share := range a.shares {
		if share.HasClassical() {
			count++
		}
	}
	return count
}

// ValidPQCount returns the number of shares that have a PQ sig.
func (a *HybridThresholdAggregator) ValidPQCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	count := 0
	for _, share := range a.shares {
		if share.HasPQ() {
			count++
		}
	}
	return count
}

// Config returns the aggregator's configuration.
func (a *HybridThresholdAggregator) Config() *HybridThresholdConfig {
	return a.config
}

// ValidateShare validates that a share is well-formed for the given message.
// A share is valid if at least one of classical or PQ signature is present
// and non-empty. If requireBoth is true, both must be present.
func ValidateShare(share *HybridThresholdShare, message []byte, requireBoth bool) error {
	if share == nil {
		return ErrHTNilShare
	}
	if len(message) == 0 {
		return ErrHTEmptyMessage
	}
	if !share.HasClassical() && !share.HasPQ() {
		return ErrHTEmptyShare
	}
	if requireBoth {
		if !share.HasClassical() {
			return ErrHTInvalidClassical
		}
		if !share.HasPQ() {
			return ErrHTInvalidPQ
		}
	}
	return nil
}

// ComputeThresholdWeight computes the total weight of collected shares
// using the configured weighting scheme.
func (a *HybridThresholdAggregator) ComputeThresholdWeight() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()

	total := 0.0
	for _, share := range a.shares {
		total += a.shareWeight(share)
	}
	return total
}
