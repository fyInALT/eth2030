package metrics

// ewma_compat.go re-exports types from metrics/ewmautil for backward compatibility.

import "github.com/eth2030/eth2030/metrics/ewmautil"

// EWMA type alias.
type EWMA = ewmautil.EWMA

// EWMA function wrappers.
func StandardEWMA(alpha float64) *EWMA { return ewmautil.StandardEWMA(alpha) }
func NewEWMA1() *EWMA                  { return ewmautil.NewEWMA1() }
func NewEWMA5() *EWMA                  { return ewmautil.NewEWMA5() }
func NewEWMA15() *EWMA                 { return ewmautil.NewEWMA15() }
