package core

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/rawdb"
	"github.com/eth2030/eth2030/core/types"
)

func TestCommitGenesis(t *testing.T) {
	g := &config.Genesis{
		Config:     config.TestConfig,
		GasLimit:   30_000_000,
		Difficulty: big.NewInt(1),
		Alloc: config.GenesisAlloc{
			types.HexToAddress("0xaaaa"): config.GenesisAccount{
				Balance: big.NewInt(1e18),
			},
		},
	}

	db := rawdb.NewMemoryDB()
	bc, err := CommitGenesis(g, db)
	if err != nil {
		t.Fatalf("CommitGenesis error: %v", err)
	}

	// Verify blockchain is initialized.
	if bc.Genesis().NumberU64() != 0 {
		t.Errorf("genesis number = %d, want 0", bc.Genesis().NumberU64())
	}
	if bc.CurrentBlock().NumberU64() != 0 {
		t.Errorf("current block = %d, want 0", bc.CurrentBlock().NumberU64())
	}

	// Verify state.
	st := bc.State()
	addr := types.HexToAddress("0xaaaa")
	if got := st.GetBalance(addr); got.Cmp(big.NewInt(1e18)) != 0 {
		t.Errorf("balance after CommitGenesis = %v, want 1e18", got)
	}
}
