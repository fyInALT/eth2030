package chain

import (
	"errors"
	"fmt"
	"math/big"
	"sync"

	"time"

	"github.com/eth2030/eth2030/core/block"
	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/execution"
	"github.com/eth2030/eth2030/core/rawdb"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/log"
	"github.com/eth2030/eth2030/metrics"
	"github.com/eth2030/eth2030/rlp"
)

var blockchainLog = log.Default().Module("eth/blockchain")

var (
	ErrNoGenesis     = errors.New("genesis block not provided")
	ErrGenesisExists = errors.New("genesis already initialized")
	ErrBlockNotFound = errors.New("block not found")
	ErrInvalidChain  = errors.New("invalid chain: blocks not contiguous")
	ErrFutureBlock2  = errors.New("block number too high")
	ErrStateNotFound = errors.New("state not found for block")
)

const (
	// defaultBlockCacheSize is the default in-memory block cache capacity.
	defaultBlockCacheSize = 256
	// defaultReceiptCacheSize is the default in-memory receipt cache capacity.
	defaultReceiptCacheSize = 128
)

// BlockchainOpts configures optional memory-cache sizes for a Blockchain.
// Zero values fall back to package defaults.
type BlockchainOpts struct {
	// BlockCacheSize is the maximum number of blocks kept in the memory cache.
	// Older entries are evicted; rawdb is the persistent fallback.
	BlockCacheSize int
	// ReceiptCacheSize is the maximum number of receipt sets kept in memory.
	ReceiptCacheSize int
	// StateCacheSize is the maximum number of MemoryStateDB snapshots retained
	// for fast reorg / payload-building (see state_cache.go).
	StateCacheSize int
}

// DefaultBlockchainOpts returns BlockchainOpts populated with package defaults.
func DefaultBlockchainOpts() BlockchainOpts {
	return BlockchainOpts{
		BlockCacheSize:   defaultBlockCacheSize,
		ReceiptCacheSize: defaultReceiptCacheSize,
		StateCacheSize:   defaultMaxCachedStates,
	}
}

// TxLookupEntry stores the location of a transaction within the chain.
type TxLookupEntry struct {
	BlockHash   types.Hash
	BlockNumber uint64
	TxIndex     uint64
}

// Blockchain manages the canonical chain of blocks, applying state
// transitions and persisting data to the underlying database.
type Blockchain struct {
	mu sync.RWMutex
	// rcMu guards receiptCache independently of mu, allowing concurrent
	// receipt reads while InsertBlock holds mu.Lock() for state execution.
	rcMu      sync.RWMutex
	config    *config.ChainConfig
	db        rawdb.Database
	hc        *HeaderChain
	processor *execution.StateProcessor
	validator block.Validator

	opts BlockchainOpts // cache size limits, set at construction

	// Block cache: hash -> block.  Bounded to opts.BlockCacheSize entries;
	// older entries are evicted (rawdb is the persistent fallback).
	blockCache      map[types.Hash]*types.Block
	blockCacheOrder []types.Hash // insertion order for LRU eviction

	// Canonical number -> hash for quick lookups.
	canonCache map[uint64]types.Hash

	// Receipt cache: blockHash -> receipts.  Protected by rcMu.
	receiptCache      map[types.Hash][]*types.Receipt
	receiptCacheOrder []types.Hash // insertion order for LRU eviction (under rcMu)

	// Transaction lookup: txHash -> location in chain.
	// Entries are evicted together with their block from blockCache.
	txLookup map[types.Hash]TxLookupEntry

	// State snapshot cache to avoid re-execution from genesis.
	sc *stateCache

	// config.Genesis state (used as base for re-execution).
	genesisState state.StateDB

	// Current state after processing the head block.
	currentState state.StateDB

	// The genesis block.
	genesis *types.Block

	// Current head block.
	currentBlock *types.Block
}

// NewBlockchain creates a new blockchain initialized with the given genesis block.
// The statedb should contain the genesis state (pre-funded accounts, etc.).
// opts is an optional BlockchainOpts; zero/absent values use package defaults.
func NewBlockchain(config *config.ChainConfig, genesis *types.Block, statedb state.StateDB, db rawdb.Database, optArgs ...BlockchainOpts) (*Blockchain, error) {
	if genesis == nil {
		return nil, ErrNoGenesis
	}
	var opts BlockchainOpts
	if len(optArgs) > 0 {
		opts = optArgs[0]
	}
	if opts.BlockCacheSize <= 0 {
		opts.BlockCacheSize = defaultBlockCacheSize
	}
	if opts.ReceiptCacheSize <= 0 {
		opts.ReceiptCacheSize = defaultReceiptCacheSize
	}
	if opts.StateCacheSize <= 0 {
		opts.StateCacheSize = defaultMaxCachedStates
	}

	proc := execution.NewStateProcessor(config)
	bc := &Blockchain{
		config:       config,
		db:           db,
		opts:         opts,
		processor:    proc,
		validator:    block.NewBlockValidator(config),
		blockCache:   make(map[types.Hash]*types.Block),
		canonCache:   make(map[uint64]types.Hash),
		receiptCache: make(map[types.Hash][]*types.Receipt),
		txLookup:     make(map[types.Hash]TxLookupEntry),
		sc:           newStateCache(opts.StateCacheSize),
		genesisState: statedb,
		currentState: statedb.Dup(),
		genesis:      genesis,
		currentBlock: genesis,
	}

	// Wire up GetHash for BLOCKHASH opcode support.
	proc.SetGetHash(bc.GetHashFn())

	// Create HeaderChain from genesis header.
	bc.hc = NewHeaderChain(config, genesis.Header())

	// Store genesis in caches.
	hash := genesis.Hash()
	bc.blockCache[hash] = genesis
	bc.canonCache[genesis.NumberU64()] = hash

	// Check if chain already exists in rawdb (e.g. after restart).
	existing, err := rawdb.ReadCanonicalHash(db, 0)
	if err != nil || existing != hash {
		// Fresh start: persist genesis and set head pointers.
		bc.writeBlock(genesis)
		rawdb.WriteCanonicalHash(db, genesis.NumberU64(), hash)
		rawdb.WriteHeadBlockHash(db, hash)
		rawdb.WriteHeadHeaderHash(db, hash)
	}

	return bc, nil
}

// evictOldestBlock removes the oldest block from blockCache (LRU eviction).
// It also removes the block's transactions from txLookup.
// Must be called with bc.mu held for writing.
func (bc *Blockchain) evictOldestBlock() {
	if len(bc.blockCacheOrder) == 0 {
		return
	}
	oldest := bc.blockCacheOrder[0]
	bc.blockCacheOrder = bc.blockCacheOrder[1:]
	if evBlk, ok := bc.blockCache[oldest]; ok {
		delete(bc.blockCache, oldest)
		for _, tx := range evBlk.Transactions() {
			delete(bc.txLookup, tx.Hash())
		}
	}
}

// InsertBlock validates, executes, and inserts a single block.
func (bc *Blockchain) InsertBlock(blk *types.Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.insertBlock(blk)
}

// insertBlock is the internal insert without locking.
func (bc *Blockchain) insertBlock(blk *types.Block) error {
	hash := blk.Hash()
	num := blk.NumberU64()
	header := blk.Header()

	blockchainLog.Debug("block_insert",
		"event", "block_insert",
		"hash", hash.Hex(),
		"num", num,
		"txCount", len(blk.Transactions()),
		"gasLimit", header.GasLimit,
		"parentHash", header.ParentHash.Hex(),
	)

	// Skip if already known.
	if _, ok := bc.blockCache[hash]; ok {
		blockchainLog.Debug("block_known",
			"event", "block_known",
			"hash", hash.Hex(),
			"num", num,
		)
		return nil
	}

	// Find parent: check cache first, then fall back to rawdb.
	parent := bc.blockCache[header.ParentHash]
	if parent == nil {
		parent = bc.readBlock(header.ParentHash)
		if parent != nil {
			bc.blockCache[header.ParentHash] = parent
		}
	}
	if parent == nil {
		err := fmt.Errorf("%w: parent %v", block.ErrUnknownParent, header.ParentHash)
		blockchainLog.Warn("block_invalid",
			"event", "block_invalid",
			"hash", hash.Hex(),
			"num", num,
			"reason", "unknown parent",
			"parentHash", header.ParentHash.Hex(),
			"error", err,
		)
		return err
	}

	// Validate header against parent.
	parentHeader := parent.Header()
	if err := bc.validator.ValidateHeader(header, parentHeader); err != nil {
		blockchainLog.Warn("block_invalid",
			"event", "block_invalid",
			"hash", hash.Hex(),
			"num", num,
			"reason", "header validation failed",
			"parentHash", header.ParentHash.Hex(),
			"parentNum", parent.NumberU64(),
			"error", err,
		)
		return err
	}

	// Validate body.
	if err := bc.validator.ValidateBody(blk); err != nil {
		blockchainLog.Warn("block_invalid",
			"event", "block_invalid",
			"hash", hash.Hex(),
			"num", num,
			"reason", "body validation failed",
			"error", err,
		)
		return err
	}

	// Build state for execution by re-executing from genesis.
	statedb, err := bc.stateAt(parent)
	if err != nil {
		blockchainLog.Error("block_state_error",
			"event", "block_state_error",
			"hash", hash.Hex(),
			"num", num,
			"parentNum", parent.NumberU64(),
			"error", err,
		)
		return fmt.Errorf("state at parent %d: %w", parent.NumberU64(), err)
	}

	// Execute transactions (with BAL tracking when Amsterdam is active).
	blockStart := time.Now()
	result, err := bc.processor.ProcessWithBAL(blk, statedb)
	metrics.BlockProcessTime.Observe(float64(time.Since(blockStart).Milliseconds()))
	if err != nil {
		blockchainLog.Error("block_exec_fail",
			"event", "block_exec_fail",
			"hash", hash.Hex(),
			"num", num,
			"txCount", len(blk.Transactions()),
			"error", err,
		)
		return fmt.Errorf("process block %d: %w", num, err)
	}
	receipts := result.Receipts

	// Validate gas used: the total gas consumed must match header.GasUsed.
	var totalGasUsed uint64
	for _, r := range receipts {
		totalGasUsed += r.GasUsed
	}
	if header.GasUsed != totalGasUsed {
		err := fmt.Errorf("%w: header=%d computed=%d", block.ErrInvalidGasUsedTotal, header.GasUsed, totalGasUsed)
		blockchainLog.Warn("block_gas_mismatch",
			"event", "block_gas_mismatch",
			"hash", hash.Hex(),
			"num", num,
			"headerGasUsed", header.GasUsed,
			"computedGasUsed", totalGasUsed,
			"error", err,
		)
		return err
	}

	// Validate receipt root: the Merkle trie hash of receipts must match
	// header.ReceiptHash.
	computedReceiptHash := block.ComputeReceiptsRoot(receipts)
	if header.ReceiptHash != computedReceiptHash {
		err := fmt.Errorf("%w: header=%s computed=%s", block.ErrInvalidReceiptRoot,
			header.ReceiptHash.Hex(), computedReceiptHash.Hex())
		blockchainLog.Warn("block_invalid",
			"event", "block_invalid",
			"hash", hash.Hex(),
			"num", num,
			"reason", "receipt root mismatch",
			"headerReceiptHash", header.ReceiptHash.Hex(),
			"computedReceiptHash", computedReceiptHash.Hex(),
			"error", err,
		)
		return err
	}

	// Validate block bloom: the bloom in the header must match the computed
	// bloom from all receipt logs.
	blockBloom := types.CreateBloom(receipts)
	if header.Bloom != blockBloom {
		err := fmt.Errorf("invalid bloom (remote: %x local: %x)", header.Bloom, blockBloom)
		blockchainLog.Warn("block_invalid",
			"event", "block_invalid",
			"hash", hash.Hex(),
			"num", num,
			"reason", "bloom mismatch",
			"error", err,
		)
		return err
	}

	// EIP-7685: process execution layer requests (Prague+).
	// ProcessRequests may modify state (e.g. clearing request count slots),
	// so it must run before computing the state root.
	if bc.config != nil && bc.config.IsPrague(header.Time) {
		if _, err := execution.ProcessRequests(bc.config, statedb, header); err != nil {
			blockchainLog.Error("block_requests_fail",
				"event", "block_requests_fail",
				"hash", hash.Hex(),
				"num", num,
				"error", err,
			)
			return fmt.Errorf("process requests block %d: %w", num, err)
		}
	}

	// Validate state root: the post-execution state root must match header.Root.
	computedRoot := statedb.GetRoot()
	if header.Root != computedRoot {
		err := fmt.Errorf("%w: header=%s computed=%s", block.ErrInvalidStateRoot,
			header.Root.Hex(), computedRoot.Hex())
		blockchainLog.Warn("block_root_mismatch",
			"event", "block_root_mismatch",
			"hash", hash.Hex(),
			"num", num,
			"headerRoot", header.Root.Hex(),
			"computedRoot", computedRoot.Hex(),
			"txCount", len(blk.Transactions()),
			"gasUsed", totalGasUsed,
			"error", err,
		)
		return err
	}

	// Validate Block Access List hash (EIP-7928).
	var computedBALHash *types.Hash
	if result.BlockAccessList != nil {
		h := result.BlockAccessList.Hash()
		computedBALHash = &h
	}
	if err := bc.validator.ValidateBlockAccessList(header, computedBALHash); err != nil {
		blockchainLog.Warn("block_invalid",
			"event", "block_invalid",
			"hash", hash.Hex(),
			"num", num,
			"reason", "BAL hash mismatch",
			"error", err,
		)
		return err
	}

	// Store in block cache (evict oldest when at capacity).
	for len(bc.blockCache) >= bc.opts.BlockCacheSize {
		bc.evictOldestBlock()
	}
	bc.blockCache[hash] = blk
	bc.blockCacheOrder = append(bc.blockCacheOrder, hash)

	txs := blk.Transactions()

	// Populate derived fields on receipts and store tx lookup entries.
	for i, receipt := range receipts {
		receipt.BlockHash = hash
		receipt.BlockNumber = new(big.Int).SetUint64(num)
		receipt.TransactionIndex = uint(i)
		if i < len(txs) {
			receipt.TxHash = txs[i].Hash()
		}
		// Set log context fields.
		for j, logEntry := range receipt.Logs {
			logEntry.BlockHash = hash
			logEntry.BlockNumber = num
			logEntry.TxHash = receipt.TxHash
			logEntry.TxIndex = uint(i)
			logEntry.Index = uint(j)
		}
	}

	// Cache receipts by block hash (rcMu allows concurrent readers in GetReceipts).
	bc.rcMu.Lock()
	for len(bc.receiptCache) >= bc.opts.ReceiptCacheSize {
		if len(bc.receiptCacheOrder) == 0 {
			break
		}
		oldest := bc.receiptCacheOrder[0]
		bc.receiptCacheOrder = bc.receiptCacheOrder[1:]
		delete(bc.receiptCache, oldest)
	}
	bc.receiptCache[hash] = receipts
	bc.receiptCacheOrder = append(bc.receiptCacheOrder, hash)
	bc.rcMu.Unlock()

	// Build tx lookup index.
	for i, tx := range txs {
		bc.txLookup[tx.Hash()] = TxLookupEntry{
			BlockHash:   hash,
			BlockNumber: num,
			TxIndex:     uint64(i),
		}
	}

	// Commit state to the backing store (for TrieStateDB this flushes the
	// dirty layer to the DB and resets it, bounding per-block memory).
	// For MemoryStateDB this is a no-op beyond flushing dirty→committed.
	if _, err := statedb.Commit(); err != nil {
		return fmt.Errorf("commit state block %d: %w", num, err)
	}

	// Always cache the post-execution state so that reorgs to this block
	// (canonical or side-chain) can recover state in O(1).
	bc.sc.put(hash, num, statedb)

	// Update canonical chain if this extends the head.
	if num > bc.currentBlock.NumberU64() {
		bc.canonCache[num] = hash
		bc.currentBlock = blk
		bc.currentState = statedb

		// Persist to rawdb.
		bc.writeBlock(blk)
		bc.writeReceipts(num, hash, receipts)
		bc.writeTxLookups(txs, num)
		rawdb.WriteCanonicalHash(bc.db, num, hash)
		rawdb.WriteHeadBlockHash(bc.db, hash)
		rawdb.WriteHeadHeaderHash(bc.db, hash)

		// Update header chain.
		bc.hc.InsertHeaders([]*types.Header{header})

		metrics.BlocksInserted.Inc()
		metrics.ChainHeight.Set(int64(num))
		blockchainLog.Info("block_added",
			"event", "block_added",
			"hash", hash.Hex(),
			"num", num,
			"txCount", len(txs),
			"gasUsed", totalGasUsed,
			"gasLimit", header.GasLimit,
		)
	} else {
		// Persist side blocks so stateAt can walk the ancestor chain even after
		// the block is evicted from the in-memory blockCache.
		bc.writeBlock(blk)
		blockchainLog.Debug("block_side",
			"event", "block_side",
			"hash", hash.Hex(),
			"num", num,
			"canonHead", bc.currentBlock.NumberU64(),
		)
	}

	return nil
}

// InsertChain inserts a chain of blocks sequentially.
// Blocks must be in ascending order but need not be contiguous with the head
// at the time of the call (though each must connect to its parent).
func (bc *Blockchain) InsertChain(blocks []*types.Block) (int, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	for i, blk := range blocks {
		if err := bc.insertBlock(blk); err != nil {
			return i, err
		}
	}
	return len(blocks), nil
}

// GetBlock retrieves a block by hash, or nil if not found.
// It checks the in-memory cache first, then falls back to rawdb.
func (bc *Blockchain) GetBlock(hash types.Hash) *types.Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	if b, ok := bc.blockCache[hash]; ok {
		return b
	}
	// Fallback: try reading from rawdb.
	return bc.readBlock(hash)
}

// GetBlockByNumber retrieves the canonical block for a given number.
func (bc *Blockchain) GetBlockByNumber(number uint64) *types.Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	hash, ok := bc.canonCache[number]
	if !ok {
		// Fallback: try rawdb canonical hash.
		h, err := rawdb.ReadCanonicalHash(bc.db, number)
		if err != nil {
			return nil
		}
		hash = h
	}
	if b, ok2 := bc.blockCache[hash]; ok2 {
		return b
	}
	return bc.readBlock(hash)
}

// CurrentBlock returns the head of the canonical chain.
func (bc *Blockchain) CurrentBlock() *types.Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.currentBlock
}

// HasBlock checks if a block with the given hash exists in cache or rawdb.
func (bc *Blockchain) HasBlock(hash types.Hash) bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	if _, ok := bc.blockCache[hash]; ok {
		return true
	}
	// Fall back to rawdb: if the header number mapping exists, the block was
	// persisted (written by writeBlock which calls WriteHeader).
	_, err := rawdb.ReadHeaderNumber(bc.db, hash)
	return err == nil
}

// SetHead rewinds the canonical chain to the given block number.
// Blocks above the target number are removed from the canonical index.
func (bc *Blockchain) SetHead(number uint64) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	blockchainLog.Info("chain_sethead",
		"event", "chain_sethead",
		"from", bc.currentBlock.NumberU64(),
		"to", number,
	)

	target, ok := bc.canonCache[number]
	if !ok {
		return fmt.Errorf("%w: no canonical block at %d", ErrBlockNotFound, number)
	}

	// Remove canonical entries above target.
	current := bc.currentBlock.NumberU64()
	for n := current; n > number; n-- {
		if hash, ok := bc.canonCache[n]; ok {
			rawdb.DeleteCanonicalHash(bc.db, n)
			delete(bc.canonCache, n)
			// Remove from block cache too.
			delete(bc.blockCache, hash)
		}
	}

	// Set new head.
	bc.currentBlock = bc.blockCache[target]

	// Re-derive state by re-executing from genesis.
	statedb, err := bc.stateAt(bc.currentBlock)
	if err != nil {
		return fmt.Errorf("re-derive state at %d: %w", number, err)
	}
	bc.currentState = statedb

	// Update rawdb pointers.
	hash := bc.currentBlock.Hash()
	rawdb.WriteHeadBlockHash(bc.db, hash)
	rawdb.WriteHeadHeaderHash(bc.db, hash)

	// Rewind header chain.
	bc.hc.SetHead(number)

	return nil
}

// GetHashFn returns a GetHashFunc that resolves block number -> hash
// for the BLOCKHASH opcode (EIP-210 compatible, up to 256 blocks back).
func (bc *Blockchain) GetHashFn() func(uint64) types.Hash {
	return func(number uint64) types.Hash {
		bc.mu.RLock()
		defer bc.mu.RUnlock()
		if hash, ok := bc.canonCache[number]; ok {
			return hash
		}
		return types.Hash{}
	}
}

// Genesis returns the genesis block.
func (bc *Blockchain) Genesis() *types.Block {
	return bc.genesis
}

// Config returns the chain configuration.
func (bc *Blockchain) Config() *config.ChainConfig {
	return bc.config
}

// State returns a copy of the current state.
func (bc *Blockchain) State() state.StateDB {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.currentState.Dup()
}

// StateAtRoot returns the state for the given state root hash.
// It searches canonical blocks for one whose header root matches,
// then re-executes from the nearest cached snapshot.
func (bc *Blockchain) StateAtRoot(root types.Hash) (state.StateDB, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Fast path: check if root matches the current head state.
	if bc.currentState.GetRoot() == root {
		return bc.currentState.Dup(), nil
	}

	// Check if root matches the genesis state.
	if bc.genesisState.GetRoot() == root {
		return bc.genesisState.Dup(), nil
	}

	// Search canonical blocks for a block with this state root.
	for _, blk := range bc.blockCache {
		if blk.Header().Root == root {
			return bc.stateAt(blk)
		}
	}

	return nil, fmt.Errorf("%w: no block found with state root %v", ErrStateNotFound, root)
}

// StateAtBlock returns the state after executing up to the given block.
// This is public for use by external packages (e.g. core/block).
func (bc *Blockchain) StateAtBlock(blk *types.Block) (state.StateDB, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.stateAt(blk)
}

// stateAt returns the state after executing up to (and including) the given block.
// For the genesis block, this is the genesis state directly.
// It checks the state cache for a snapshot closer to the target block to avoid
// re-executing the entire chain from genesis.
func (bc *Blockchain) stateAt(blk *types.Block) (state.StateDB, error) {
	if blk.Hash() == bc.genesis.Hash() {
		return bc.genesisState.Dup(), nil
	}

	// Check if we have an exact cached state for this block.
	// Dup() so callers (insertBlock, block builder) do not corrupt the cache.
	if cached, ok := bc.sc.get(blk.Hash()); ok {
		return cached.Dup(), nil
	}

	// Collect the chain of blocks from genesis (or a cached snapshot) to this block.
	var chain []*types.Block
	current := blk
	var baseState state.StateDB

	for current.Hash() != bc.genesis.Hash() {
		// Check if we have a cached state for this ancestor.
		if cached, ok := bc.sc.get(current.ParentHash()); ok {
			// Dup() so processor.Process does not mutate the cached state.
			baseState = cached.Dup()
			chain = append(chain, current)
			break
		}
		chain = append(chain, current)
		parent, ok := bc.blockCache[current.ParentHash()]
		if !ok {
			// Fallback: try rawdb.
			parent = bc.readBlock(current.ParentHash())
			if parent != nil {
				bc.blockCache[current.ParentHash()] = parent
			}
		}
		if parent == nil {
			return nil, fmt.Errorf("%w: missing ancestor at %v", ErrStateNotFound, current.ParentHash())
		}
		current = parent
	}

	// Use genesis state as base if no cached snapshot was found.
	if baseState == nil {
		baseState = bc.genesisState.Dup()
	}

	// Re-execute from the base state.
	for i := len(chain) - 1; i >= 0; i-- {
		b := chain[i]
		if _, err := bc.processor.Process(b, baseState); err != nil {
			return nil, fmt.Errorf("re-execute block %d: %w", b.NumberU64(), err)
		}
	}
	return baseState, nil
}

// writeBlock persists a block's header and body to rawdb using RLP encoding.
func (bc *Blockchain) writeBlock(blk *types.Block) {
	num := blk.NumberU64()
	hash := blk.Hash()

	// RLP-encode the header.
	headerData, err := blk.Header().EncodeRLP()
	if err != nil {
		return
	}
	rawdb.WriteHeader(bc.db, num, hash, headerData)

	// RLP-encode the body (transactions list + uncles list).
	bodyData, err := encodeBlockBody(blk.Body())
	if err != nil {
		return
	}
	rawdb.WriteBody(bc.db, num, hash, bodyData)
}

// readBlock retrieves a block from rawdb by looking up the block number
// from the header hash, then reading and decoding header and body.
func (bc *Blockchain) readBlock(hash types.Hash) *types.Block {
	// Look up block number from hash.
	num, err := rawdb.ReadHeaderNumber(bc.db, hash)
	if err != nil {
		return nil
	}

	// Read header.
	headerData, err := rawdb.ReadHeader(bc.db, num, hash)
	if err != nil || len(headerData) == 0 {
		return nil
	}
	header, err := types.DecodeHeaderRLP(headerData)
	if err != nil {
		return nil
	}

	// Read body.
	bodyData, err := rawdb.ReadBody(bc.db, num, hash)
	if err != nil || len(bodyData) == 0 {
		// Body may be empty for blocks with no transactions; create block with header only.
		return types.NewBlock(header, nil)
	}
	body, err := decodeBlockBody(bodyData)
	if err != nil {
		// If body decode fails, return header-only block.
		return types.NewBlock(header, nil)
	}
	return types.NewBlock(header, body)
}

// encodeBlockBody RLP-encodes a block body as [transactions_list, uncles_list, withdrawals_list].
// The withdrawals list is included only when body.Withdrawals is non-nil (post-Shanghai).
func encodeBlockBody(body *types.Body) ([]byte, error) {
	// Encode transactions.
	var txsPayload []byte
	if body != nil {
		for _, tx := range body.Transactions {
			txEnc, err := tx.EncodeRLP()
			if err != nil {
				return nil, err
			}
			if tx.Type() == types.LegacyTxType {
				// Legacy: txEnc is already an RLP list; append directly.
				txsPayload = append(txsPayload, txEnc...)
			} else {
				// Typed: wrap as RLP byte string (type_byte || RLP_payload).
				wrapped, err := rlp.EncodeToBytes(txEnc)
				if err != nil {
					return nil, err
				}
				txsPayload = append(txsPayload, wrapped...)
			}
		}
	}

	// Encode uncles.
	var unclesPayload []byte
	if body != nil {
		for _, uncle := range body.Uncles {
			uncleEnc, err := uncle.EncodeRLP()
			if err != nil {
				return nil, err
			}
			unclesPayload = append(unclesPayload, uncleEnc...)
		}
	}

	var payload []byte
	payload = append(payload, rlp.WrapList(txsPayload)...)
	payload = append(payload, rlp.WrapList(unclesPayload)...)

	// Encode withdrawals (post-Shanghai).
	if body != nil && body.Withdrawals != nil {
		var wsPayload []byte
		for _, w := range body.Withdrawals {
			wEnc, err := rlp.EncodeToBytes([]interface{}{w.Index, w.ValidatorIndex, w.Address, w.Amount})
			if err != nil {
				return nil, err
			}
			wsPayload = append(wsPayload, wEnc...)
		}
		payload = append(payload, rlp.WrapList(wsPayload)...)
	}

	return rlp.WrapList(payload), nil
}

// decodeBlockBody decodes an RLP-encoded block body [transactions_list, uncles_list, withdrawals_list?].
// The withdrawals list is optional (post-Shanghai).
func decodeBlockBody(data []byte) (*types.Body, error) {
	s := rlp.NewStreamFromBytes(data)
	_, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("opening body list: %w", err)
	}

	// Decode transactions.
	_, err = s.List()
	if err != nil {
		return nil, fmt.Errorf("opening txs list: %w", err)
	}
	var txs []*types.Transaction
	for !s.AtListEnd() {
		kind, _, err := s.Kind()
		if err != nil {
			return nil, fmt.Errorf("peeking tx kind: %w", err)
		}
		var txData []byte
		if kind == rlp.List {
			txData, err = s.RawItem()
		} else {
			txData, err = s.Bytes()
		}
		if err != nil {
			return nil, fmt.Errorf("reading tx: %w", err)
		}
		tx, err := types.DecodeTxRLP(txData)
		if err != nil {
			return nil, fmt.Errorf("decoding tx: %w", err)
		}
		txs = append(txs, tx)
	}
	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("closing txs list: %w", err)
	}

	// Decode uncles.
	_, err = s.List()
	if err != nil {
		return nil, fmt.Errorf("opening uncles list: %w", err)
	}
	var uncles []*types.Header
	for !s.AtListEnd() {
		uncleBytes, err := s.RawItem()
		if err != nil {
			return nil, fmt.Errorf("reading uncle: %w", err)
		}
		uncle, err := types.DecodeHeaderRLP(uncleBytes)
		if err != nil {
			return nil, fmt.Errorf("decoding uncle: %w", err)
		}
		uncles = append(uncles, uncle)
	}
	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("closing uncles list: %w", err)
	}

	body := &types.Body{
		Transactions: txs,
		Uncles:       uncles,
	}

	// Decode optional withdrawals (post-Shanghai).
	if !s.AtListEnd() {
		_, err = s.List()
		if err != nil {
			return nil, fmt.Errorf("opening withdrawals list: %w", err)
		}
		var withdrawals []*types.Withdrawal
		for !s.AtListEnd() {
			wBytes, err := s.RawItem()
			if err != nil {
				return nil, fmt.Errorf("reading withdrawal: %w", err)
			}
			w, err := decodeWithdrawal(wBytes)
			if err != nil {
				return nil, fmt.Errorf("decoding withdrawal: %w", err)
			}
			withdrawals = append(withdrawals, w)
		}
		if err := s.ListEnd(); err != nil {
			return nil, fmt.Errorf("closing withdrawals list: %w", err)
		}
		if withdrawals == nil {
			withdrawals = []*types.Withdrawal{} // empty but non-nil
		}
		body.Withdrawals = withdrawals
	}

	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("closing body list: %w", err)
	}

	return body, nil
}

// decodeWithdrawal decodes an RLP-encoded withdrawal [index, validatorIndex, address, amount].
func decodeWithdrawal(data []byte) (*types.Withdrawal, error) {
	s := rlp.NewStreamFromBytes(data)
	_, err := s.List()
	if err != nil {
		return nil, err
	}
	w := &types.Withdrawal{}
	w.Index, err = s.Uint64()
	if err != nil {
		return nil, err
	}
	w.ValidatorIndex, err = s.Uint64()
	if err != nil {
		return nil, err
	}
	addrBytes, err := s.Bytes()
	if err != nil {
		return nil, err
	}
	copy(w.Address[types.AddressLength-len(addrBytes):], addrBytes)
	w.Amount, err = s.Uint64()
	if err != nil {
		return nil, err
	}
	if err := s.ListEnd(); err != nil {
		return nil, err
	}
	return w, nil
}

// Reorg replaces the canonical chain from the fork point with the new chain
// ending at newHead. It finds the common ancestor between the current
// canonical chain and the new chain, un-indexes old canonical blocks,
// re-indexes the new canonical blocks, and updates the current block pointer.
func (bc *Blockchain) Reorg(newHead *types.Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.reorg(newHead)
}

// reorg is the internal reorg implementation without locking.
func (bc *Blockchain) reorg(newHead *types.Block) error {
	prevHead := bc.currentBlock
	metrics.ReorgsDetected.Inc()
	blockchainLog.Info("block_reorganized",
		"event", "block_reorganized",
		"oldHash", prevHead.Hash().Hex(),
		"oldNum", prevHead.NumberU64(),
		"newHash", newHead.Hash().Hex(),
		"newNum", newHead.NumberU64(),
	)

	// Build the new chain's ancestry: walk from newHead to genesis.
	// Collect blocks in reverse order (head first).
	var newChain []*types.Block
	current := newHead
	for {
		if _, ok := bc.blockCache[current.Hash()]; !ok {
			bc.blockCache[current.Hash()] = current
		}
		if current.NumberU64() == 0 {
			break
		}
		newChain = append(newChain, current)
		parent, ok := bc.blockCache[current.ParentHash()]
		if !ok {
			// Fall back to rawdb for evicted ancestors.
			parent = bc.readBlock(current.ParentHash())
			if parent == nil {
				return fmt.Errorf("%w: missing ancestor %v during reorg", ErrBlockNotFound, current.ParentHash())
			}
			bc.blockCache[current.ParentHash()] = parent
		}
		current = parent
	}

	// Determine the maximum height to clean up.
	oldHead := bc.currentBlock.NumberU64()
	newHeight := newHead.NumberU64()
	maxHeight := oldHead
	if newHeight > maxHeight {
		maxHeight = newHeight
	}

	// Un-index all canonical blocks above genesis.
	// Delete txLookup entries for any orphaned blocks so that rawdb lookups
	// do not return stale block numbers after the canonical chain changes.
	for n := maxHeight; n >= 1; n-- {
		if hash, ok := bc.canonCache[n]; ok {
			delete(bc.canonCache, n)
			rawdb.DeleteCanonicalHash(bc.db, n)

			bc.hc.mu.Lock()
			if h, ok := bc.hc.headers[n]; ok {
				delete(bc.hc.headersByHash, h.Hash())
				delete(bc.hc.headers, n)
			}
			bc.hc.mu.Unlock()

			// Remove in-memory txLookup entries and rawdb txLookup entries
			// for the orphaned block so stale lookups cannot be returned.
			if orphan, ok := bc.blockCache[hash]; ok {
				for _, tx := range orphan.Transactions() {
					delete(bc.txLookup, tx.Hash())
					rawdb.DeleteTxLookup(bc.db, tx.Hash())
				}
			}
		}
	}

	// Re-index the new chain from lowest block to highest.
	// Write txLookup entries for the incoming canonical blocks.
	for i := len(newChain) - 1; i >= 0; i-- {
		blk := newChain[i]
		hash := blk.Hash()
		num := blk.NumberU64()

		bc.blockCache[hash] = blk
		bc.canonCache[num] = hash
		bc.writeBlock(blk)
		rawdb.WriteCanonicalHash(bc.db, num, hash)
		bc.writeTxLookups(blk.Transactions(), num)
		for idx, tx := range blk.Transactions() {
			bc.txLookup[tx.Hash()] = TxLookupEntry{
				BlockHash:   hash,
				BlockNumber: num,
				TxIndex:     uint64(idx),
			}
		}

		h := blk.Header()
		bc.hc.mu.Lock()
		bc.hc.headersByHash[h.Hash()] = h
		bc.hc.headers[num] = h
		bc.hc.mu.Unlock()
	}

	// Update current block.
	bc.currentBlock = newHead
	rawdb.WriteHeadBlockHash(bc.db, newHead.Hash())
	rawdb.WriteHeadHeaderHash(bc.db, newHead.Hash())

	bc.hc.mu.Lock()
	bc.hc.currentHeader = newHead.Header()
	bc.hc.mu.Unlock()

	// Re-derive state for the new head.
	statedb, err := bc.stateAt(newHead)
	if err != nil {
		blockchainLog.Error("reorg_state_fail",
			"event", "reorg_state_fail",
			"newHash", newHead.Hash().Hex(),
			"newNum", newHead.NumberU64(),
			"error", err,
		)
		return fmt.Errorf("re-derive state after reorg at %d: %w", newHead.NumberU64(), err)
	}
	bc.currentState = statedb

	return nil
}

// ChainLength returns the length of the canonical chain (genesis = 1).
func (bc *Blockchain) ChainLength() uint64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.currentBlock.NumberU64() + 1
}

// GetReceipts returns the receipts for a block identified by hash.
// Uses rcMu instead of the main mu to avoid blocking on concurrent InsertBlock.
func (bc *Blockchain) GetReceipts(blockHash types.Hash) []*types.Receipt {
	bc.rcMu.RLock()
	r, ok := bc.receiptCache[blockHash]
	bc.rcMu.RUnlock()
	if ok {
		return r
	}
	// Fallback: read from rawdb (thread-safe, no bc.mu needed).
	num, err := rawdb.ReadHeaderNumber(bc.db, blockHash)
	if err != nil {
		return nil
	}
	blk := bc.readBlock(blockHash)
	var txs []*types.Transaction
	if blk != nil {
		txs = blk.Transactions()
	}
	return bc.readReceiptsFromDB(num, blockHash, txs)
}

// GetBlockReceipts returns the receipts for the canonical block at the given number.
// Reads canonical hash from rawdb (thread-safe) and receipt cache under rcMu to
// avoid blocking on InsertBlock's bc.mu.Lock() during state execution.
func (bc *Blockchain) GetBlockReceipts(number uint64) []*types.Receipt {
	// Prefer rawdb for canonical hash: it's always consistent and doesn't
	// need bc.mu. Fall back to in-memory canonCache only if rawdb lacks the entry.
	hash, err := rawdb.ReadCanonicalHash(bc.db, number)
	if err != nil {
		bc.mu.RLock()
		h, ok := bc.canonCache[number]
		bc.mu.RUnlock()
		if !ok {
			return nil
		}
		hash = h
	}
	// Check receipt cache under fine-grained rcMu.
	bc.rcMu.RLock()
	r, ok2 := bc.receiptCache[hash]
	bc.rcMu.RUnlock()
	if ok2 {
		return r
	}
	// Fallback: read from rawdb (thread-safe).
	blk := bc.readBlock(hash)
	var txs []*types.Transaction
	if blk != nil {
		txs = blk.Transactions()
	}
	return bc.readReceiptsFromDB(number, hash, txs)
}

// GetLogs returns all logs from receipts for the block identified by hash.
func (bc *Blockchain) GetLogs(blockHash types.Hash) []*types.Log {
	bc.rcMu.RLock()
	receipts, ok := bc.receiptCache[blockHash]
	bc.rcMu.RUnlock()
	if !ok {
		// Fallback: read from rawdb (thread-safe, no bc.mu needed).
		num, err := rawdb.ReadHeaderNumber(bc.db, blockHash)
		if err == nil {
			blk := bc.readBlock(blockHash)
			var txs []*types.Transaction
			if blk != nil {
				txs = blk.Transactions()
			}
			receipts = bc.readReceiptsFromDB(num, blockHash, txs)
		}
	}
	var logs []*types.Log
	for _, receipt := range receipts {
		logs = append(logs, receipt.Logs...)
	}
	return logs
}

// GetTxLookupEntry returns the TxLookupEntry for a transaction hash,
// giving callers structured access to all three location fields as a unit.
// Returns (entry, false) if the transaction is not found.
func (bc *Blockchain) GetTxLookupEntry(txHash types.Hash) (TxLookupEntry, bool) {
	blockHash, blockNumber, txIndex, found := bc.GetTransactionLookup(txHash)
	if !found {
		return TxLookupEntry{}, false
	}
	return TxLookupEntry{BlockHash: blockHash, BlockNumber: blockNumber, TxIndex: txIndex}, true
}

// GetTransactionLookup returns the block location for a transaction hash.
func (bc *Blockchain) GetTransactionLookup(txHash types.Hash) (blockHash types.Hash, blockNumber uint64, txIndex uint64, found bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	if entry, ok := bc.txLookup[txHash]; ok {
		return entry.BlockHash, entry.BlockNumber, entry.TxIndex, true
	}
	// Fallback: try rawdb tx lookup index.
	num, err := rawdb.ReadTxLookup(bc.db, txHash)
	if err != nil {
		return types.Hash{}, 0, 0, false
	}
	hash, err := rawdb.ReadCanonicalHash(bc.db, num)
	if err != nil {
		return types.Hash{}, 0, 0, false
	}
	// Find the transaction index within the block.
	blk := bc.readBlock(hash)
	if blk == nil {
		return types.Hash{}, 0, 0, false
	}
	for i, tx := range blk.Transactions() {
		if tx.Hash() == txHash {
			return hash, num, uint64(i), true
		}
	}
	return types.Hash{}, 0, 0, false
}

// HistoryOldestBlock returns the oldest block number for which block bodies
// and receipts are available. Returns 0 if no history pruning has occurred.
// Used by the RPC layer to detect EIP-4444 pruned data.
func (bc *Blockchain) HistoryOldestBlock() uint64 {
	oldest, _ := rawdb.ReadHistoryOldest(bc.db)
	return oldest
}

// PruneHistory prunes block bodies and receipts older than
// (head - retention) blocks. Headers are preserved. Returns the number
// of blocks pruned.
func (bc *Blockchain) PruneHistory(retention uint64) (uint64, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if bc.currentBlock == nil {
		return 0, nil
	}
	head := bc.currentBlock.NumberU64()
	pruned, _, err := rawdb.PruneHistory(bc.db, head, retention)
	return pruned, err
}

// writeReceipts encodes and persists receipts for a block.
func (bc *Blockchain) writeReceipts(number uint64, hash types.Hash, receipts []*types.Receipt) {
	if len(receipts) == 0 {
		return
	}
	var encoded []byte
	for _, r := range receipts {
		data, err := r.EncodeRLP()
		if err != nil {
			continue
		}
		encoded = append(encoded, data...)
	}
	rawdb.WriteReceipts(bc.db, number, hash, encoded)
}

// writeTxLookups persists transaction hash -> block number mappings.
func (bc *Blockchain) writeTxLookups(txs []*types.Transaction, blockNumber uint64) {
	for _, tx := range txs {
		rawdb.WriteTxLookup(bc.db, tx.Hash(), blockNumber)
	}
}

// rlpItemSize returns the encoded byte length of the next RLP item in data.
// Returns 0 for empty or malformed input.
func rlpItemSize(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	b := data[0]
	switch {
	case b < 0x80:
		return 1
	case b <= 0xB7:
		return 1 + int(b-0x80)
	case b <= 0xBF:
		lenOfLen := int(b - 0xB7)
		if len(data) < 1+lenOfLen {
			return 0
		}
		var l int
		for i := 0; i < lenOfLen; i++ {
			l = (l << 8) | int(data[1+i])
		}
		return 1 + lenOfLen + l
	case b <= 0xF7:
		return 1 + int(b-0xC0)
	default:
		lenOfLen := int(b - 0xF7)
		if len(data) < 1+lenOfLen {
			return 0
		}
		var l int
		for i := 0; i < lenOfLen; i++ {
			l = (l << 8) | int(data[1+i])
		}
		return 1 + lenOfLen + l
	}
}

// readReceiptsFromDB reads and decodes receipts from rawdb for the given block.
// txs is used to restore derived fields (TxHash, GasUsed) not stored in RLP.
// Called with bc.mu held (read or write).
func (bc *Blockchain) readReceiptsFromDB(num uint64, hash types.Hash, txs []*types.Transaction) []*types.Receipt {
	data, err := rawdb.ReadReceipts(bc.db, num, hash)
	if err != nil || len(data) == 0 {
		return nil
	}
	// Decode the concatenated receipt RLP encodings produced by writeReceipts.
	var receipts []*types.Receipt
	for len(data) > 0 {
		var itemSize int
		if data[0] < 0x80 {
			// Typed receipt: type byte followed by an RLP list.
			if len(data) < 2 {
				break
			}
			listSize := rlpItemSize(data[1:])
			if listSize <= 0 || 1+listSize > len(data) {
				break
			}
			itemSize = 1 + listSize
		} else {
			// Legacy receipt: plain RLP list.
			itemSize = rlpItemSize(data)
			if itemSize <= 0 || itemSize > len(data) {
				break
			}
		}
		r, err := types.DecodeReceiptRLP(data[:itemSize])
		if err != nil {
			break
		}
		receipts = append(receipts, r)
		data = data[itemSize:]
	}
	if len(receipts) == 0 {
		return nil
	}
	// Restore derived fields not persisted in the consensus RLP encoding.
	for i, r := range receipts {
		r.BlockHash = hash
		r.BlockNumber = new(big.Int).SetUint64(num)
		r.TransactionIndex = uint(i)
		if i < len(txs) {
			r.TxHash = txs[i].Hash()
		}
		if i == 0 {
			r.GasUsed = r.CumulativeGasUsed
		} else {
			r.GasUsed = r.CumulativeGasUsed - receipts[i-1].CumulativeGasUsed
		}
		for j, log := range r.Logs {
			log.BlockHash = hash
			log.BlockNumber = num
			log.TxHash = r.TxHash
			log.TxIndex = uint(i)
			log.Index = uint(j)
		}
	}
	return receipts
}

// EvictBlockFromCache removes a block from the in-memory block cache,
// forcing subsequent lookups to fall back to rawdb. Used in tests.
func (bc *Blockchain) EvictBlockFromCache(hash types.Hash) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	delete(bc.blockCache, hash)
}

// StoreBlockInCache inserts a block directly into the in-memory block cache
// without validation or state execution. Used in tests to set up fork scenarios.
func (bc *Blockchain) StoreBlockInCache(block *types.Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.blockCache[block.Hash()] = block
}
