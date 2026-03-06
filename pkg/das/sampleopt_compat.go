package das

// sampleopt_compat.go re-exports types from das/sampleopt for backward compatibility.

import "github.com/eth2030/eth2030/das/sampleopt"

// SampleOptimizer type aliases.
type (
	SampleOptimizerConfig = sampleopt.SampleOptimizerConfig
	SamplingPlan          = sampleopt.SamplingPlan
	SamplingVerdict       = sampleopt.SamplingVerdict
	SampleOptimizer       = sampleopt.SampleOptimizer
)

// SampleOptimizer error variables.
var (
	ErrInvalidSecurityParam = sampleopt.ErrInvalidSecurityParam
	ErrInvalidBlobCount     = sampleopt.ErrInvalidBlobCount
	ErrInvalidNetworkHealth = sampleopt.ErrInvalidNetworkHealth
	ErrInvalidFailureRate   = sampleopt.ErrInvalidFailureRate
	ErrInvalidSampleSize    = sampleopt.ErrInvalidSampleSize
)

// SampleOptimizer function wrappers.
func DefaultSampleOptimizerConfig() SampleOptimizerConfig {
	return sampleopt.DefaultSampleOptimizerConfig()
}
func NewSampleOptimizer(config SampleOptimizerConfig) *SampleOptimizer {
	return sampleopt.NewSampleOptimizer(config)
}
