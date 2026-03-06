package rlperrors

import (
	"errors"
	"testing"
)

func TestErrors(t *testing.T) {
	errs := []error{
		ErrExpectedString,
		ErrExpectedList,
		ErrCanonSize,
		ErrEOL,
		ErrCanonInt,
		ErrNonCanonicalSize,
		ErrUint64Range,
		ErrValueTooLarge,
	}
	for _, err := range errs {
		if err == nil {
			t.Fatal("error var is nil")
		}
		// Each error must be unique (no aliasing between them).
		for _, other := range errs {
			if err != other && errors.Is(err, other) {
				t.Errorf("%v unexpectedly matches %v", err, other)
			}
		}
	}
}
