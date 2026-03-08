package core

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/eips"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
)

func TestBeaconBlockRootNotCalledPreCancun(t *testing.T) {
	statedb := state.NewMemoryStateDB()

	beaconRoot := types.HexToHash("0xdeadbeef00000000000000000000000000000000000000000000000000000000")

	// Use a chain config where Cancun is NOT active.
	preCancunConfig := &config.ChainConfig{
		ChainID:                 big.NewInt(1),
		HomesteadBlock:          big.NewInt(0),
		ByzantiumBlock:          big.NewInt(0),
		ConstantinopleBlock:     big.NewInt(0),
		IstanbulBlock:           big.NewInt(0),
		BerlinBlock:             big.NewInt(0),
		LondonBlock:             big.NewInt(0),
		TerminalTotalDifficulty: big.NewInt(0),
		ShanghaiTime:            newUint64(0),
		CancunTime:              nil, // Cancun NOT active
	}

	header := &types.Header{
		Number:           big.NewInt(1),
		GasLimit:         10_000_000,
		Time:             1000,
		BaseFee:          big.NewInt(1_000_000_000),
		Coinbase:         types.HexToAddress("0xfee"),
		ParentBeaconRoot: &beaconRoot,
	}

	block := types.NewBlock(header, &types.Body{})
	proc := NewStateProcessor(preCancunConfig)
	_, err := proc.Process(block, statedb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Beacon root should NOT be stored because Cancun is not active.
	// Use eips.BeaconRootAddress and the exported uint64ToHash equivalent.
	// We verify by checking the state directly using known slot indices.
	// timestamp_idx = 1000 % 8191 = 1000, root_idx = 1000 + 8191 = 9191
	timestampSlot := types.HexToHash("0x00000000000000000000000000000000000000000000000000000000000003e8") // 1000
	rootSlot := types.HexToHash("0x00000000000000000000000000000000000000000000000000000000000023e7")      // 9191

	storedTimestamp := statedb.GetState(eips.BeaconRootAddress, timestampSlot)
	if storedTimestamp != (types.Hash{}) {
		t.Fatalf("beacon root should NOT be stored pre-Cancun, got timestamp %s", storedTimestamp.Hex())
	}
	storedRoot := statedb.GetState(eips.BeaconRootAddress, rootSlot)
	if storedRoot != (types.Hash{}) {
		t.Fatalf("beacon root should NOT be stored pre-Cancun, got root %s", storedRoot.Hex())
	}
}

func TestBeaconBlockRootCalledPostCancun(t *testing.T) {
	statedb := state.NewMemoryStateDB()

	beaconRoot := types.HexToHash("0xdeadbeef00000000000000000000000000000000000000000000000000000000")

	header := &types.Header{
		Number:           big.NewInt(1),
		GasLimit:         10_000_000,
		Time:             1000,
		BaseFee:          big.NewInt(1_000_000_000),
		Coinbase:         types.HexToAddress("0xfee"),
		ParentBeaconRoot: &beaconRoot,
	}

	block := types.NewBlock(header, &types.Body{})
	proc := NewStateProcessor(config.TestConfig) // config.TestConfig has all forks active
	_, err := proc.Process(block, statedb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Beacon root SHOULD be stored because Cancun is active.
	// timestamp_idx = 1000 % 8191 = 1000, root_idx = 1000 + 8191 = 9191
	timestampSlot := types.HexToHash("0x00000000000000000000000000000000000000000000000000000000000003e8")     // 1000
	rootSlot := types.HexToHash("0x00000000000000000000000000000000000000000000000000000000000023e7")          // 9191
	expectedTimestamp := types.HexToHash("0x00000000000000000000000000000000000000000000000000000000000003e8") // 1000

	storedTimestamp := statedb.GetState(eips.BeaconRootAddress, timestampSlot)
	if storedTimestamp != expectedTimestamp {
		t.Fatalf("timestamp should be stored post-Cancun: got %s, want %s",
			storedTimestamp.Hex(), expectedTimestamp.Hex())
	}
	storedRoot := statedb.GetState(eips.BeaconRootAddress, rootSlot)
	if storedRoot != beaconRoot {
		t.Fatalf("beacon root should be stored post-Cancun: got %s, want %s",
			storedRoot.Hex(), beaconRoot.Hex())
	}
}
