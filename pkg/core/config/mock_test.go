package config

import (
	"errors"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// mockPaymasterSlasher implements PaymasterSlasher for testing.
type mockPaymasterSlasher struct {
	slashFn func(addr types.Address) error
	slashed []types.Address
}

func (m *mockPaymasterSlasher) SlashOnBadSettlement(addr types.Address) error {
	m.slashed = append(m.slashed, addr)
	if m.slashFn != nil {
		return m.slashFn(addr)
	}
	return nil
}

// Compile-time interface check.
var _ PaymasterSlasher = (*mockPaymasterSlasher)(nil)

// --- Interface tests ---

func TestPaymasterSlasher_Interface(t *testing.T) {
	var slasher PaymasterSlasher = &mockPaymasterSlasher{}
	addr := types.Address{1}
	if err := slasher.SlashOnBadSettlement(addr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify the concrete mock recorded the slashed address.
	concrete := slasher.(*mockPaymasterSlasher)
	if len(concrete.slashed) != 1 || concrete.slashed[0] != addr {
		t.Fatal("expected slash to be recorded")
	}
}

func TestPaymasterSlasher_ErrorPropagation(t *testing.T) {
	wantErr := errors.New("slash failed")
	var slasher PaymasterSlasher = &mockPaymasterSlasher{
		slashFn: func(addr types.Address) error { return wantErr },
	}
	if err := slasher.SlashOnBadSettlement(types.Address{2}); err != wantErr {
		t.Fatalf("got %v, want %v", err, wantErr)
	}
}

func TestPaymasterSlasher_MultipleSlashes(t *testing.T) {
	mock := &mockPaymasterSlasher{}
	var slasher PaymasterSlasher = mock
	addrs := []types.Address{{1}, {2}, {3}}
	for _, addr := range addrs {
		if err := slasher.SlashOnBadSettlement(addr); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if len(mock.slashed) != 3 {
		t.Fatalf("expected 3 slashes, got %d", len(mock.slashed))
	}
	for i, addr := range addrs {
		if mock.slashed[i] != addr {
			t.Fatalf("slashed[%d] = %v, want %v", i, mock.slashed[i], addr)
		}
	}
}
