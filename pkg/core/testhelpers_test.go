package core

import (
	"math/big"
	"testing"

	blkpkg "github.com/eth2030/eth2030/core/block"
	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/execution"
	"github.com/eth2030/eth2030/core/gas"
	"github.com/eth2030/eth2030/core/rawdb"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
)

// newUint64 is a test helper that returns a pointer to a uint64 value.
// It mirrors config.newUint64 for use in core package tests.
func newUint64(v uint64) *uint64 { return &v }

// testChain creates a blockchain with a genesis block for use in tests.
func testChain(t *testing.T) (*Blockchain, *state.MemoryStateDB) {
	t.Helper()
	statedb := state.NewMemoryStateDB()
	genesis := makeGenesis(30_000_000, big.NewInt(1))
	db := rawdb.NewMemoryDB()
	bc, err := NewBlockchain(config.TestConfig, genesis, statedb, db)
	if err != nil {
		t.Fatalf("NewBlockchain: %v", err)
	}
	return bc, statedb
}

// makeBlock builds a valid child block of parent with the given transactions.
// It uses an empty state to compute all consensus-critical header fields.
// This is suitable only for the FIRST block after a genesis with empty state.
// For chains of blocks, use makeBlockWithState with a shared state.
func makeBlock(parent *types.Block, txs []*types.Transaction) *types.Block {
	return makeBlockWithState(parent, txs, state.NewMemoryStateDB())
}

// makeChainBlocks builds a chain of empty blocks from the given parent using
// the provided state (which is mutated in place).
func makeChainBlocks(parent *types.Block, count int, statedb *state.MemoryStateDB) []*types.Block {
	blocks := make([]*types.Block, count)
	for i := 0; i < count; i++ {
		blocks[i] = makeBlockWithState(parent, nil, statedb)
		parent = blocks[i]
	}
	return blocks
}

// makeBlockWithState builds a valid child block and computes the correct header
// fields by executing the transactions against the provided state.
func makeBlockWithState(parent *types.Block, txs []*types.Transaction, statedb *state.MemoryStateDB) *types.Block {
	parentHeader := parent.Header()
	blobGasUsed := uint64(0)
	var pExcess, pUsed uint64
	if parentHeader.ExcessBlobGas != nil {
		pExcess = *parentHeader.ExcessBlobGas
	}
	if parentHeader.BlobGasUsed != nil {
		pUsed = *parentHeader.BlobGasUsed
	}
	excessBlobGas := gas.CalcExcessBlobGas(pExcess, pUsed)

	calldataGasUsed := uint64(0)
	var pCalldataExcess, pCalldataUsed uint64
	if parentHeader.CalldataExcessGas != nil {
		pCalldataExcess = *parentHeader.CalldataExcessGas
	}
	if parentHeader.CalldataGasUsed != nil {
		pCalldataUsed = *parentHeader.CalldataGasUsed
	}
	calldataExcessGas := gas.CalcCalldataExcessGas(pCalldataExcess, pCalldataUsed, parentHeader.GasLimit)

	emptyWHash := types.EmptyRootHash
	emptyBeaconRoot := types.EmptyRootHash
	emptyRequestsHash := types.EmptyRootHash
	header := &types.Header{
		ParentHash:        parent.Hash(),
		Number:            new(big.Int).Add(parentHeader.Number, big.NewInt(1)),
		GasLimit:          parentHeader.GasLimit,
		Time:              parentHeader.Time + 12,
		Difficulty:        new(big.Int),
		BaseFee:           gas.CalcBaseFee(parentHeader),
		UncleHash:         EmptyUncleHash,
		WithdrawalsHash:   &emptyWHash,
		BlobGasUsed:       &blobGasUsed,
		ExcessBlobGas:     &excessBlobGas,
		ParentBeaconRoot:  &emptyBeaconRoot,
		RequestsHash:      &emptyRequestsHash,
		CalldataGasUsed:   &calldataGasUsed,
		CalldataExcessGas: &calldataExcessGas,
	}

	body := &types.Body{
		Transactions: txs,
		Withdrawals:  []*types.Withdrawal{},
	}

	blk := types.NewBlock(header, body)

	proc := execution.NewStateProcessor(config.TestConfig)
	result, err := proc.ProcessWithBAL(blk, statedb)
	if err == nil {
		var gasUsed uint64
		for _, r := range result.Receipts {
			gasUsed += r.GasUsed
		}
		header.GasUsed = gasUsed

		var cdGasUsed uint64
		for _, tx := range txs {
			cdGasUsed += tx.CalldataGas()
		}
		*header.CalldataGasUsed = cdGasUsed

		if result.BlockAccessList != nil {
			h := result.BlockAccessList.Hash()
			header.BlockAccessListHash = &h
		}

		header.Bloom = types.CreateBloom(result.Receipts)
		header.ReceiptHash = blkpkg.DeriveReceiptsRoot(result.Receipts)
		header.Root = statedb.GetRoot()
	}

	header.TxHash = blkpkg.DeriveTxsRoot(txs)
	return types.NewBlock(header, body)
}

// makeGenesis creates a genesis block with the given gas limit and base fee.
func makeGenesis(gasLimit uint64, baseFee *big.Int) *types.Block {
	blobGasUsed := uint64(0)
	excessBlobGas := uint64(0)
	calldataGasUsed := uint64(0)
	calldataExcessGas := uint64(0)
	emptyWithdrawalsHash := types.EmptyRootHash
	emptyRoot := types.EmptyRootHash
	header := &types.Header{
		Number:            big.NewInt(0),
		GasLimit:          gasLimit,
		GasUsed:           0,
		Time:              0,
		Difficulty:        new(big.Int),
		BaseFee:           baseFee,
		UncleHash:         EmptyUncleHash,
		WithdrawalsHash:   &emptyWithdrawalsHash,
		BlobGasUsed:       &blobGasUsed,
		ExcessBlobGas:     &excessBlobGas,
		ParentBeaconRoot:  &emptyRoot,
		RequestsHash:      &emptyRoot,
		CalldataGasUsed:   &calldataGasUsed,
		CalldataExcessGas: &calldataExcessGas,
	}
	return types.NewBlock(header, &types.Body{Withdrawals: []*types.Withdrawal{}})
}
