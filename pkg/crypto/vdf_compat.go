package crypto

// vdf_compat.go provides backward-compatible re-exports from crypto/vdf.
// Existing code that imports "github.com/eth2030/eth2030/crypto" and uses
// VDF types continues to work unchanged. New code should import
// "github.com/eth2030/eth2030/crypto/vdf" directly.

import (
	"math/big"

	"github.com/eth2030/eth2030/crypto/vdf"
)

// VDF type aliases — keep crypto.XXX working for all existing importers.
type (
	VDFParams          = vdf.VDFParams
	VDFProof           = vdf.VDFProof
	VDFEvaluator       = vdf.VDFEvaluator
	WesolowskiVDF      = vdf.WesolowskiVDF
	VDFv2Config        = vdf.VDFv2Config
	VDFv2Result        = vdf.VDFv2Result
	AggregatedVDFProof = vdf.AggregatedVDFProof
	VDFv2              = vdf.VDFv2
	ChainedVDFProof    = vdf.ChainedVDFProof
	BeaconOutput       = vdf.BeaconOutput
	VDFChain           = vdf.VDFChain
	VDFBeacon          = vdf.VDFBeacon
)

// MaxChainLength re-exports the maximum chain length constant.
const MaxChainLength = vdf.MaxChainLength

// VDF function wrappers — delegate to crypto/vdf implementations.

// DefaultVDFParams returns the default VDF parameters.
func DefaultVDFParams() *VDFParams { return vdf.DefaultVDFParams() }

// ValidateVDFParams checks that VDF parameters are secure.
func ValidateVDFParams(params *VDFParams) error { return vdf.ValidateVDFParams(params) }

// ValidateVDFProof checks that a VDF proof has non-empty fields.
func ValidateVDFProof(proof *VDFProof) error { return vdf.ValidateVDFProof(proof) }

// NewWesolowskiVDF creates a new VDF evaluator with the given parameters.
func NewWesolowskiVDF(params *VDFParams) *WesolowskiVDF { return vdf.NewWesolowskiVDF(params) }

// NewWesolowskiVDFWithModulus creates a VDF evaluator with an explicit modulus.
func NewWesolowskiVDFWithModulus(params *VDFParams, n *big.Int) *WesolowskiVDF {
	return vdf.NewWesolowskiVDFWithModulus(params, n)
}

// DefaultVDFv2Config returns sensible defaults for the enhanced VDF.
func DefaultVDFv2Config() VDFv2Config { return vdf.DefaultVDFv2Config() }

// NewVDFv2 creates a new enhanced VDF with the given configuration.
func NewVDFv2(config VDFv2Config) *VDFv2 { return vdf.NewVDFv2(config) }

// NewVDFChain creates a chain evaluator using the given VDFv2 and per-step
// iteration count.
func NewVDFChain(v *VDFv2, itersPerStep uint64) *VDFChain { return vdf.NewVDFChain(v, itersPerStep) }

// NewVDFBeacon creates a beacon with the given chain evaluator and chain length.
func NewVDFBeacon(chain *VDFChain, chainLen uint64) *VDFBeacon {
	return vdf.NewVDFBeacon(chain, chainLen)
}
