package das

// gf_compat.go re-exports types from das/gf for backward compatibility.

import "github.com/eth2030/eth2030/das/gf"

// GF type aliases.
type (
	GF2_16    = gf.GF2_16
	RSEncoder = gf.RSEncoder
)

// GF constants.
const MaxGF16Shards = gf.MaxGF16Shards

// GF error variables.
var (
	ErrRSInvalidConfig     = gf.ErrRSInvalidConfig
	ErrRSDataTooLarge      = gf.ErrRSDataTooLarge
	ErrRSShardSizeMismatch = gf.ErrRSShardSizeMismatch
	ErrRSEmptyInput        = gf.ErrRSEmptyInput
	ErrRSShardCount        = gf.ErrRSShardCount
	ErrRSTooFewShards      = gf.ErrRSTooFewShards
)

// GF function wrappers.
func GFAdd(a, b GF2_16) GF2_16                          { return gf.GFAdd(a, b) }
func GFSub(a, b GF2_16) GF2_16                          { return gf.GFSub(a, b) }
func GFMul(a, b GF2_16) GF2_16                          { return gf.GFMul(a, b) }
func GFDiv(a, b GF2_16) GF2_16                          { return gf.GFDiv(a, b) }
func GFInverse(a GF2_16) GF2_16                         { return gf.GFInverse(a) }
func GFPow(a GF2_16, n int) GF2_16                      { return gf.GFPow(a, n) }
func GFPolyEval(coeffs []GF2_16, x GF2_16) GF2_16       { return gf.GFPolyEval(coeffs, x) }
func GFPolyMul(p1, p2 []GF2_16) []GF2_16                { return gf.GFPolyMul(p1, p2) }
func GFPolyAdd(p1, p2 []GF2_16) []GF2_16                { return gf.GFPolyAdd(p1, p2) }
func GFPolyScale(poly []GF2_16, scalar GF2_16) []GF2_16 { return gf.GFPolyScale(poly, scalar) }
func GFVandermondeRow(x GF2_16, n int) []GF2_16         { return gf.GFVandermondeRow(x, n) }
func GFPolyFromRoots(roots []GF2_16) []GF2_16           { return gf.GFPolyFromRoots(roots) }
func GFInterpolate(xs, ys []GF2_16) []GF2_16            { return gf.GFInterpolate(xs, ys) }
func GFExp(i int) GF2_16                                { return gf.GFExp(i) }
func GFLog(a GF2_16) int                                { return gf.GFLog(a) }
func NewRSEncoder(dataShards, parityShards int) (*RSEncoder, error) {
	return gf.NewRSEncoder(dataShards, parityShards)
}
