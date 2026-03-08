package core

import (
	"encoding/json"

	"github.com/eth2030/eth2030/core/chain"
	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/rawdb"
	"github.com/eth2030/eth2030/core/state"
)

// CommitGenesis initializes the database with the genesis block and state.
// Returns the initialized blockchain. This function bridges config.Genesis
// with core.Blockchain, which cannot be expressed from within core/config
// due to the circular import it would create.
func CommitGenesis(g *config.Genesis, db rawdb.Database) (*chain.Blockchain, error) {
	statedb := state.NewMemoryStateDB()
	block := g.SetupGenesisBlock(statedb)

	cfg := g.Config
	if cfg == nil {
		cfg = config.TestConfig
	}

	bc, err := chain.NewBlockchain(cfg, block, statedb, db)
	if err != nil {
		return nil, err
	}

	// Store genesis config as JSON in rawdb.
	configData, err := json.Marshal(cfg)
	if err == nil {
		db.Put([]byte("ETH2030-chain-config"), configData)
	}

	return bc, nil
}
