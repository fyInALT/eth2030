package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

// flagSet wraps flag.FlagSet to add support for uint64 and string-slice flags.
type flagSet struct {
	*flag.FlagSet
}

// newCustomFlagSet creates a flagSet with ContinueOnError behavior.
func newCustomFlagSet(name string) *flagSet {
	return &flagSet{FlagSet: flag.NewFlagSet(name, flag.ContinueOnError)}
}

// Uint64Var defines a uint64 flag bound to *p.
func (fs *flagSet) Uint64Var(p *uint64, name string, value uint64, usage string) {
	*p = value
	fs.FlagSet.Var(&uint64Value{p: p}, name, usage)
}

// Uint64PtrVar defines a uint64 flag whose value is written to *p only when
// explicitly provided. p must point to a *uint64 field (nil = not set).
func (fs *flagSet) Uint64PtrVar(p **uint64, name string, usage string) {
	fs.FlagSet.Var(&uint64PtrValue{p: p}, name, usage)
}

// StringSliceVar defines a flag that accepts a comma-separated string and
// splits it into []string, writing the result to *p.
func (fs *flagSet) StringSliceVar(p *[]string, name string, defaultVal []string, usage string) {
	*p = defaultVal
	fs.FlagSet.Var(&stringSliceValue{p: p}, name, usage)
}

// Bool wraps flag.FlagSet.Bool.
func (fs *flagSet) Bool(name string, value bool, usage string) *bool {
	return fs.FlagSet.Bool(name, value, usage)
}

// --- uint64Value ---

type uint64Value struct{ p *uint64 }

func (v *uint64Value) String() string {
	if v.p == nil {
		return "0"
	}
	return strconv.FormatUint(*v.p, 10)
}

func (v *uint64Value) Set(s string) error {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid uint64 value %q", s)
	}
	*v.p = n
	return nil
}

// --- uint64PtrValue ---

type uint64PtrValue struct{ p **uint64 }

func (v *uint64PtrValue) String() string {
	if v.p == nil || *v.p == nil {
		return ""
	}
	return strconv.FormatUint(**v.p, 10)
}

func (v *uint64PtrValue) Set(s string) error {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid uint64 value %q", s)
	}
	val := n
	*v.p = &val
	return nil
}

// --- stringSliceValue ---

type stringSliceValue struct{ p *[]string }

func (v *stringSliceValue) String() string {
	if v.p == nil {
		return ""
	}
	return strings.Join(*v.p, ",")
}

func (v *stringSliceValue) Set(s string) error {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	*v.p = result
	return nil
}
