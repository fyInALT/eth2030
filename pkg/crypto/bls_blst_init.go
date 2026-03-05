//go:build blst

// bls_blst_init.go wires BlstRealBackend as the default BLS backend when
// the binary is compiled with -tags blst (GAP-7.2). Without this build tag,
// PureGoBLSBackend remains the default for testing and environments without CGO.
package crypto

func init() {
	SetBLSBackend(&BlstRealBackend{})
}
