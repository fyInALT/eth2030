package node

import (
	"fmt"
	"log/slog"
	"math/big"

	coregas "github.com/eth2030/eth2030/core/gas"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
	"github.com/eth2030/eth2030/rpc"
	"github.com/eth2030/eth2030/txpool/shared"
	"github.com/eth2030/eth2030/trie"
)

// nodeBackend adapts the Node to the rpc.Backend interface.
type nodeBackend struct {
	node *Node
}

func newNodeBackend(n *Node) rpc.Backend {
	return &nodeBackend{node: n}
}

func (b *nodeBackend) HeaderByNumber(number rpc.BlockNumber) *types.Header {
	bc := b.node.blockchain
	switch number {
	case rpc.LatestBlockNumber, rpc.PendingBlockNumber:
		blk := bc.CurrentBlock()
		if blk != nil {
			return blk.Header()
		}
		return nil
	case rpc.EarliestBlockNumber:
		blk := bc.GetBlockByNumber(0)
		if blk != nil {
			return blk.Header()
		}
		return nil
	case rpc.FinalizedBlockNumber:
		if blk := bc.CurrentFinalBlock(); blk != nil {
			return blk.Header()
		}
		return nil
	case rpc.SafeBlockNumber:
		if blk := bc.CurrentSafeBlock(); blk != nil {
			return blk.Header()
		}
		return nil
	default:
		blk := bc.GetBlockByNumber(uint64(number))
		if blk != nil {
			return blk.Header()
		}
		return nil
	}
}

func (b *nodeBackend) HeaderByHash(hash types.Hash) *types.Header {
	blk := b.node.blockchain.GetBlock(hash)
	if blk != nil {
		return blk.Header()
	}
	return nil
}

func (b *nodeBackend) BlockByNumber(number rpc.BlockNumber) *types.Block {
	bc := b.node.blockchain
	switch number {
	case rpc.LatestBlockNumber, rpc.PendingBlockNumber:
		return bc.CurrentBlock()
	case rpc.EarliestBlockNumber:
		return bc.GetBlockByNumber(0)
	case rpc.FinalizedBlockNumber:
		if b.node.engBackend != nil {
			if h := b.node.engBackend.GetFinalizedHash(); h != (types.Hash{}) {
				return bc.GetBlock(h)
			}
		}
		return nil
	case rpc.SafeBlockNumber:
		if b.node.engBackend != nil {
			if h := b.node.engBackend.GetSafeHash(); h != (types.Hash{}) {
				return bc.GetBlock(h)
			}
		}
		return nil
	default:
		return bc.GetBlockByNumber(uint64(number))
	}
}

func (b *nodeBackend) BlockByHash(hash types.Hash) *types.Block {
	return b.node.blockchain.GetBlock(hash)
}

func (b *nodeBackend) CurrentHeader() *types.Header {
	blk := b.node.blockchain.CurrentBlock()
	if blk != nil {
		return blk.Header()
	}
	return nil
}

func (b *nodeBackend) ChainID() *big.Int {
	return b.node.blockchain.Config().ChainID
}

func (b *nodeBackend) StateAt(root types.Hash) (state.StateDB, error) {
	return b.node.blockchain.StateAtRoot(root)
}

func (b *nodeBackend) GetProof(addr types.Address, storageKeys []types.Hash, blockNumber rpc.BlockNumber) (*trie.AccountProof, error) {
	header := b.HeaderByNumber(blockNumber)
	if header == nil {
		return nil, fmt.Errorf("block not found")
	}

	statedb, err := b.StateAt(header.Root)
	if err != nil {
		return nil, err
	}

	memState, ok := statedb.(*state.MemoryStateDB)
	if !ok {
		return nil, fmt.Errorf("state does not support proof generation")
	}

	stateTrie := memState.BuildStateTrie()
	storageTrie := memState.BuildStorageTrie(addr)

	return trie.ProveAccountWithStorage(stateTrie, addr, storageTrie, storageKeys)
}

func (b *nodeBackend) SendTransaction(tx *types.Transaction) error {
	if err := b.node.txPool.AddLocal(tx); err != nil {
		return err
	}
	if h := b.node.ethHandler; h != nil {
		h.BroadcastTransactions([]*types.Transaction{tx})
	}
	if b.node.txJournal != nil {
		if jerr := b.node.txJournal.Insert(tx, true); jerr != nil {
			slog.Debug("tx journal insert failed", "hash", tx.Hash(), "err", jerr)
		}
	}
	if b.node.sharedPool != nil {
		var chainID uint64
		if cfg := b.node.blockchain.Config(); cfg != nil && cfg.ChainID != nil {
			chainID = cfg.ChainID.Uint64()
		}
		signer := types.LatestSigner(chainID)
		sender, _ := signer.Sender(tx)
		var gasPrice uint64
		if fc := tx.GasFeeCap(); fc != nil {
			gasPrice = fc.Uint64()
		}
		smTx := shared.SharedMempoolTx{
			Hash:     tx.Hash(),
			Sender:   sender,
			Nonce:    tx.Nonce(),
			GasPrice: gasPrice,
			Data:     tx.Data(),
		}
		if err := b.node.sharedPool.AddTransaction(smTx); err != nil {
			slog.Debug("shared mempool add", "hash", tx.Hash(), "err", err)
		}
	}
	if b.node.rollupSeq != nil {
		if encoded, encErr := tx.EncodeRLP(); encErr == nil {
			if seqErr := b.node.rollupSeq.AddTransaction(encoded); seqErr != nil {
				slog.Debug("rollup sequencer add", "hash", tx.Hash(), "err", seqErr)
			}
		} else {
			slog.Debug("rollup sequencer: tx encode failed", "hash", tx.Hash(), "err", encErr)
		}
	}
	return nil
}

func (b *nodeBackend) GetTransaction(hash types.Hash) (*types.Transaction, uint64, uint64) {
	if entry, found := b.node.blockchain.GetTxLookupEntry(hash); found {
		blk := b.node.blockchain.GetBlock(entry.BlockHash)
		if blk != nil {
			txs := blk.Transactions()
			if int(entry.TxIndex) < len(txs) {
				return txs[entry.TxIndex], entry.BlockNumber, entry.TxIndex
			}
		}
	}
	tx := b.node.txPool.Get(hash)
	if tx != nil {
		return tx, 0, 0
	}
	return nil, 0, 0
}

func (b *nodeBackend) SuggestGasPrice() *big.Int {
	if b.node.gasOracle != nil && b.node.gasOracle.BaseFee().Sign() > 0 {
		return b.node.gasOracle.SuggestGasPrice()
	}
	blk := b.node.blockchain.CurrentBlock()
	if blk != nil && blk.Header().BaseFee != nil {
		return new(big.Int).Set(blk.Header().BaseFee)
	}
	return big.NewInt(1_000_000_000)
}

func (b *nodeBackend) GetReceipts(blockHash types.Hash) []*types.Receipt {
	return b.node.blockchain.GetReceipts(blockHash)
}

func (b *nodeBackend) GetLogs(blockHash types.Hash) []*types.Log {
	return b.node.blockchain.GetLogs(blockHash)
}

func (b *nodeBackend) GetBlockReceipts(number uint64) []*types.Receipt {
	return b.node.blockchain.GetBlockReceipts(number)
}

func (b *nodeBackend) EVMCall(from types.Address, to *types.Address, data []byte, gas uint64, value *big.Int, blockNumber rpc.BlockNumber) ([]byte, uint64, error) {
	bc := b.node.blockchain

	var header *types.Header
	switch blockNumber {
	case rpc.LatestBlockNumber, rpc.PendingBlockNumber:
		blk := bc.CurrentBlock()
		if blk != nil {
			header = blk.Header()
		}
	default:
		blk := bc.GetBlockByNumber(uint64(blockNumber))
		if blk != nil {
			header = blk.Header()
		}
	}
	if header == nil {
		return nil, 0, fmt.Errorf("block not found")
	}

	statedb, err := b.StateAt(header.Root)
	if err != nil {
		return nil, 0, fmt.Errorf("state not found: %w", err)
	}

	if gas == 0 {
		gas = 50_000_000
	}
	if value == nil {
		value = new(big.Int)
	}

	blockCtx := vm.BlockContext{
		GetHash:     bc.GetHashFn(),
		BlockNumber: header.Number,
		Time:        header.Time,
		GasLimit:    header.GasLimit,
		BaseFee:     header.BaseFee,
	}
	txCtx := vm.TxContext{
		Origin:   from,
		GasPrice: header.BaseFee,
	}

	evm := vm.NewEVMWithState(blockCtx, txCtx, vm.Config{}, statedb)

	if chainCfg := bc.Config(); chainCfg != nil {
		rules := chainCfg.Rules(header.Number, chainCfg.IsMerge(), header.Time)
		forkRules := vm.ForkRules{
			IsGlamsterdan:    rules.IsGlamsterdan,
			IsPrague:         rules.IsPrague,
			IsCancun:         rules.IsCancun,
			IsShanghai:       rules.IsShanghai,
			IsMerge:          rules.IsMerge,
			IsLondon:         rules.IsLondon,
			IsBerlin:         rules.IsBerlin,
			IsIstanbul:       rules.IsIstanbul,
			IsConstantinople: rules.IsConstantinople,
			IsByzantium:      rules.IsByzantium,
			IsHomestead:      rules.IsHomestead,
			IsEIP158:         rules.IsEIP158,
			IsEIP7708:        rules.IsEIP7708,
			IsEIP7954:        rules.IsEIP7954,
			IsIPlus:          rules.IsIPlus,
		}
		evm.SetJumpTable(vm.SelectJumpTable(forkRules))
		evm.SetPrecompiles(vm.SelectPrecompiles(forkRules))
		evm.SetForkRules(forkRules)
	}

	if to == nil {
		return nil, gas, nil
	}

	ret, gasLeft, err := evm.Call(from, *to, data, gas, value)
	return ret, gasLeft, err
}

func (b *nodeBackend) HistoryOldestBlock() uint64 {
	return b.node.blockchain.HistoryOldestBlock()
}

func (b *nodeBackend) BlobSchedule(blockTime uint64) (target, max, updateFraction uint64) {
	chainCfg := b.node.blockchain.Config()
	sched := coregas.GetBlobSchedule(chainCfg, blockTime)
	return sched.Target, sched.Max, sched.UpdateFraction
}

func (b *nodeBackend) TraceTransaction(txHash types.Hash) (*vm.StructLogTracer, error) {
	bc := b.node.blockchain

	entry, found := bc.GetTxLookupEntry(txHash)
	if !found {
		return nil, fmt.Errorf("transaction %v not found", txHash)
	}

	blk := b.node.blockchain.GetBlock(entry.BlockHash)
	if blk == nil {
		return nil, fmt.Errorf("block %v not found", entry.BlockHash)
	}

	txs := blk.Transactions()
	if int(entry.TxIndex) >= len(txs) {
		return nil, fmt.Errorf("transaction index %d out of range", entry.TxIndex)
	}

	header := blk.Header()
	parentBlock := bc.GetBlock(header.ParentHash)
	if parentBlock == nil {
		return nil, fmt.Errorf("parent block %v not found", header.ParentHash)
	}
	statedb, err := b.StateAt(parentBlock.Header().Root)
	if err != nil {
		return nil, fmt.Errorf("state not found for parent block: %w", err)
	}

	blockCtx := vm.BlockContext{
		GetHash:     bc.GetHashFn(),
		BlockNumber: header.Number,
		Time:        header.Time,
		GasLimit:    header.GasLimit,
		BaseFee:     header.BaseFee,
	}

	for i := uint64(0); i < entry.TxIndex; i++ {
		tx := txs[i]
		from := types.Address{}
		if sender := tx.Sender(); sender != nil {
			from = *sender
		}
		txCtx := vm.TxContext{
			Origin:   from,
			GasPrice: tx.GasPrice(),
		}
		evm := vm.NewEVMWithState(blockCtx, txCtx, vm.Config{}, statedb)
		to := tx.To()
		if to != nil {
			evm.Call(from, *to, tx.Data(), tx.Gas(), tx.Value())
		}
		statedb.SetNonce(from, statedb.GetNonce(from)+1)
	}

	targetTx := txs[entry.TxIndex]
	from := types.Address{}
	if sender := targetTx.Sender(); sender != nil {
		from = *sender
	}
	txCtx := vm.TxContext{
		Origin:   from,
		GasPrice: targetTx.GasPrice(),
	}

	tracer := vm.NewStructLogTracer()
	tracingCfg := vm.Config{
		Debug:  true,
		Tracer: tracer,
	}
	evm := vm.NewEVMWithState(blockCtx, txCtx, tracingCfg, statedb)

	intrinsicGas := uint64(21000)
	for _, b := range targetTx.Data() {
		if b == 0 {
			intrinsicGas += 4
		} else {
			intrinsicGas += 16
		}
	}
	if targetTx.To() == nil {
		intrinsicGas += 32000
	}
	evmGas := uint64(0)
	if targetTx.Gas() > intrinsicGas {
		evmGas = targetTx.Gas() - intrinsicGas
	}

	to := targetTx.To()
	if to != nil {
		ret, gasLeft, err := evm.Call(from, *to, targetTx.Data(), evmGas, targetTx.Value())
		gasUsed := evmGas - gasLeft
		tracer.CaptureEnd(ret, gasUsed, err)
	} else if evmGas > 0 {
		tracer.CaptureEnd(nil, evmGas, fmt.Errorf("contract creation tracing not supported"))
	}

	return tracer, nil
}