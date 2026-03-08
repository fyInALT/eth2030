package eips

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// oneETH is 1 ETH in wei.
var oneETH = new(big.Int).SetUint64(DefaultMinStakeWei)

// halfETH is 0.5 ETH in wei.
var halfETH = new(big.Int).SetUint64(DefaultMinStakeWei / 2)

func pmTestAddr(b byte) types.Address {
	return types.BytesToAddress([]byte{b})
}

// TestPaymasterRegistry_Deposit verifies that depositing 1 ETH results in approval.
func TestPaymasterRegistry_Deposit(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	addr := pmTestAddr(0x01)

	if err := r.Deposit(addr, new(big.Int).Set(oneETH), 1); err != nil {
		t.Fatalf("Deposit: unexpected error: %v", err)
	}
	if !r.IsApprovedPaymaster(addr) {
		t.Fatal("expected paymaster to be approved after 1 ETH deposit")
	}
}

// TestPaymasterRegistry_InsufficientStake verifies that 0.5 ETH deposit is not approved.
func TestPaymasterRegistry_InsufficientStake(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	addr := pmTestAddr(0x02)

	if err := r.Deposit(addr, new(big.Int).Set(halfETH), 1); err != nil {
		t.Fatalf("Deposit: unexpected error: %v", err)
	}
	if r.IsApprovedPaymaster(addr) {
		t.Fatal("expected paymaster NOT to be approved with only 0.5 ETH stake")
	}
}

// TestPaymasterRegistry_ZeroDeposit verifies that a zero-value deposit returns an error.
func TestPaymasterRegistry_ZeroDeposit(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	addr := pmTestAddr(0x03)

	err := r.Deposit(addr, new(big.Int), 1)
	if err == nil {
		t.Fatal("expected error for zero deposit, got nil")
	}
}

// TestPaymasterRegistry_UnknownAddress verifies that an unknown address is not approved.
func TestPaymasterRegistry_UnknownAddress(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	addr := pmTestAddr(0x04)

	if r.IsApprovedPaymaster(addr) {
		t.Fatal("expected unknown address to not be approved")
	}
}

// TestPaymasterRegistry_EntryNil verifies that Entry returns nil for an unknown address.
func TestPaymasterRegistry_EntryNil(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	addr := pmTestAddr(0x05)

	if e := r.Entry(addr); e != nil {
		t.Fatalf("expected nil entry for unknown address, got %+v", e)
	}
}

// TestPaymasterRegistry_Slash_BanAt3 verifies that after 3 slashes the paymaster is banned.
func TestPaymasterRegistry_Slash_BanAt3(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	addr := pmTestAddr(0x06)

	if err := r.Deposit(addr, new(big.Int).Set(oneETH), 1); err != nil {
		t.Fatalf("Deposit: %v", err)
	}

	r.Slash(addr)
	r.Slash(addr)

	if !r.IsApprovedPaymaster(addr) {
		t.Fatal("expected paymaster to still be approved after 2 slashes")
	}

	r.Slash(addr)

	e := r.Entry(addr)
	if e == nil {
		t.Fatal("expected entry, got nil")
	}
	if !e.Banned {
		t.Fatal("expected Banned=true after 3rd slash")
	}
	if r.IsApprovedPaymaster(addr) {
		t.Fatal("expected paymaster NOT to be approved after ban")
	}
}

// TestPaymasterRegistry_SlashCounterIncrements verifies that one slash increments the counter.
func TestPaymasterRegistry_SlashCounterIncrements(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	addr := pmTestAddr(0x07)

	if err := r.Deposit(addr, new(big.Int).Set(oneETH), 1); err != nil {
		t.Fatalf("Deposit: %v", err)
	}
	r.Slash(addr)

	e := r.Entry(addr)
	if e == nil {
		t.Fatal("expected entry, got nil")
	}
	if e.SlashCount != 1 {
		t.Fatalf("expected SlashCount==1, got %d", e.SlashCount)
	}
}

// TestPaymasterRegistry_Withdraw_Cooldown verifies the cooldown enforcement.
// Cooldown is DefaultWithdrawDelay (256) blocks.
// Requesting at block 200, trying at block 455 (200+255) must fail;
// at block 456 (200+256) must succeed.
func TestPaymasterRegistry_Withdraw_Cooldown(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	addr := pmTestAddr(0x08)

	if err := r.Deposit(addr, new(big.Int).Set(oneETH), 1); err != nil {
		t.Fatalf("Deposit: %v", err)
	}

	r.RequestWithdraw(addr, 200)

	// Block 455: 200 + 256 - 1 = 455 → still in cooldown.
	err := r.Withdraw(addr, new(big.Int).Set(halfETH), 455)
	if err != ErrWithdrawCooldown {
		t.Fatalf("expected ErrWithdrawCooldown at block 455, got %v", err)
	}

	// Block 456: 200 + 256 = 456 → cooldown elapsed.
	err = r.Withdraw(addr, new(big.Int).Set(halfETH), 456)
	if err != nil {
		t.Fatalf("expected no error at block 456, got %v", err)
	}
}

// TestPaymasterRegistry_Withdraw_ExceedsStake verifies that withdrawing more than staked fails.
func TestPaymasterRegistry_Withdraw_ExceedsStake(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	addr := pmTestAddr(0x09)

	if err := r.Deposit(addr, new(big.Int).Set(oneETH), 1); err != nil {
		t.Fatalf("Deposit: %v", err)
	}
	r.RequestWithdraw(addr, 1)

	twoETH := new(big.Int).Mul(oneETH, big.NewInt(2))
	err := r.Withdraw(addr, twoETH, 1+DefaultWithdrawDelay)
	if err != ErrInvalidWithdrawAmt {
		t.Fatalf("expected ErrInvalidWithdrawAmt, got %v", err)
	}
}

// TestPaymasterRegistry_MultipleDeposits verifies that multiple deposits accumulate.
func TestPaymasterRegistry_MultipleDeposits(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	addr := pmTestAddr(0x0a)

	if err := r.Deposit(addr, new(big.Int).Set(halfETH), 1); err != nil {
		t.Fatalf("first Deposit: %v", err)
	}
	if err := r.Deposit(addr, new(big.Int).Set(halfETH), 2); err != nil {
		t.Fatalf("second Deposit: %v", err)
	}

	if !r.IsApprovedPaymaster(addr) {
		t.Fatal("expected paymaster to be approved after two 0.5 ETH deposits totalling 1 ETH")
	}
}

// TestPaymasterRegistry_CustomConfig verifies that MaxSlashes=1 bans after a single slash.
func TestPaymasterRegistry_CustomConfig(t *testing.T) {
	cfg := RegistryConfig{
		MinStake:      new(big.Int).Set(oneETH),
		MaxSlashes:    1,
		WithdrawDelay: DefaultWithdrawDelay,
	}
	r := NewPaymasterRegistry(cfg)
	addr := pmTestAddr(0x0b)

	if err := r.Deposit(addr, new(big.Int).Set(oneETH), 1); err != nil {
		t.Fatalf("Deposit: %v", err)
	}
	r.Slash(addr)

	e := r.Entry(addr)
	if e == nil {
		t.Fatal("expected entry, got nil")
	}
	if !e.Banned {
		t.Fatal("expected Banned=true after 1 slash with MaxSlashes=1")
	}
	if r.IsApprovedPaymaster(addr) {
		t.Fatal("expected paymaster NOT to be approved after ban")
	}
}
