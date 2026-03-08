package core

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/eips"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
)

func TestEIP2935_IntegratedWithProcessor(t *testing.T) {
	statedb := state.NewMemoryStateDB()

	// Fund a sender for the block.
	sender := types.HexToAddress("0x1111")
	statedb.AddBalance(sender, new(big.Int).Mul(big.NewInt(10), new(big.Int).SetUint64(1e18)))

	parentHash := types.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")

	header := &types.Header{
		Number:     big.NewInt(100),
		GasLimit:   10_000_000,
		Time:       1000,
		BaseFee:    big.NewInt(1_000_000_000),
		Coinbase:   types.HexToAddress("0xfee"),
		ParentHash: parentHash,
	}

	body := &types.Body{}
	block := types.NewBlock(header, body)

	proc := NewStateProcessor(config.TestConfig)
	_, err := proc.Process(block, statedb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Since config.TestConfig has Prague active, the parent hash should be stored.
	// Parent number = 100 - 1 = 99.
	got := eips.GetHistoricalBlockHash(statedb, 99)
	if got != parentHash {
		t.Fatalf("integrated: got %v, want %v", got, parentHash)
	}
}

func TestEIP2935_NotActivePrePrague(t *testing.T) {
	statedb := state.NewMemoryStateDB()

	// Use a config where Prague is not active.
	prePragueConfig := &config.ChainConfig{
		ChainID:                 big.NewInt(1337),
		HomesteadBlock:          big.NewInt(0),
		EIP150Block:             big.NewInt(0),
		EIP155Block:             big.NewInt(0),
		EIP158Block:             big.NewInt(0),
		ByzantiumBlock:          big.NewInt(0),
		ConstantinopleBlock:     big.NewInt(0),
		PetersburgBlock:         big.NewInt(0),
		IstanbulBlock:           big.NewInt(0),
		BerlinBlock:             big.NewInt(0),
		LondonBlock:             big.NewInt(0),
		TerminalTotalDifficulty: big.NewInt(0),
		ShanghaiTime:            newUint64(0),
		CancunTime:              newUint64(0),
		PragueTime:              nil, // Prague not activated
	}

	parentHash := types.HexToHash("0xdeadbeef00000000000000000000000000000000000000000000000000000000")

	header := &types.Header{
		Number:     big.NewInt(100),
		GasLimit:   10_000_000,
		Time:       1000,
		BaseFee:    big.NewInt(1_000_000_000),
		Coinbase:   types.HexToAddress("0xfee"),
		ParentHash: parentHash,
	}

	body := &types.Body{}
	block := types.NewBlock(header, body)

	proc := NewStateProcessor(prePragueConfig)
	_, err := proc.Process(block, statedb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Prague not active, so history storage should not have the hash.
	got := eips.GetHistoricalBlockHash(statedb, 99)
	if got != (types.Hash{}) {
		t.Fatalf("pre-Prague: expected zero hash, got %v", got)
	}
}
