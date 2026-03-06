package gaspool

import (
	"errors"
	"testing"
)

func TestGasPool_AddSub(t *testing.T) {
	var pool GasPool = 1000

	pool.AddGas(500)
	if pool.Gas() != 1500 {
		t.Fatalf("after AddGas(500): got %d, want 1500", pool.Gas())
	}

	if err := pool.SubGas(300); err != nil {
		t.Fatalf("SubGas(300): %v", err)
	}
	if pool.Gas() != 1200 {
		t.Fatalf("after SubGas(300): got %d, want 1200", pool.Gas())
	}
}

func TestGasPool_SubExhausted(t *testing.T) {
	var pool GasPool = 100

	err := pool.SubGas(200)
	if !errors.Is(err, ErrGasPoolExhausted) {
		t.Fatalf("expected ErrGasPoolExhausted, got %v", err)
	}
	if pool.Gas() != 100 {
		t.Fatalf("pool should be unchanged after failed SubGas: got %d", pool.Gas())
	}
}

func TestGasPool_ZeroGas(t *testing.T) {
	var pool GasPool = 0

	err := pool.SubGas(1)
	if !errors.Is(err, ErrGasPoolExhausted) {
		t.Fatalf("zero pool: expected ErrGasPoolExhausted, got %v", err)
	}
}
