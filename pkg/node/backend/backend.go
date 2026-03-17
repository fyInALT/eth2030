// Package backend provides backend implementations for RPC and Engine APIs.
// This is a skeleton implementation that abstracts Node dependencies for testability.
package backend

import (
	"context"
	"log/slog"
	"math/big"
	"net"
	"sync"

	"github.com/eth2030/eth2030/core/block"
	"github.com/eth2030/eth2030/core/chain"
	"github.com/eth2030/eth2030/core/mev"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine"
	"github.com/eth2030/eth2030/engine/forkchoice"
	"github.com/eth2030/eth2030/txpool"
)

// NodeDeps provides access to all node dependencies for backend implementations.
// This interface abstracts the Node to enable testing with mocks.
type NodeDeps interface {
	// Core (frequently accessed)
	Blockchain() *chain.Blockchain
	TxPool() *txpool.TxPool

	// Config
	Config() *Config

	// Optional dependencies (return nil if not available)
	GasOracle() any
	EthHandler() any
	TxJournal() any
	SharedPool() any
	RollupSeq() any
	MEVConfig() *mev.MEVProtectionConfig

	// State management
	SnapshotTree() any
	TriePruner() any
	TrieMigrator() any
	TrieAnnouncer() any
	StackTrie() any
	BlobSyncMgr() any
	StateHealer() any
	StateSyncSched() any

	// Forkchoice
	FCStateManager() any
	FCTracker() any

	// ePBS
	EPBSAuction() any
	EPBSBuilder() any
	EPBSEscrow() any
	EPBSCommit() any
	EPBSBid() any
	EPBSMEVBurn() any
	EngineAuction() any

	// Rollup
	RollupBridge() any
	RollupAnchor() any
	RollupProof() any

	// Misc
	PortalRouter() any
	EncryptedProtocol() any
	EncryptedPool() any
	AcctTracker() any
	NonceTracker() any
	PayloadChunker() any
	NonceAnnouncer() any
	GasRateTracker() any
	StarkFrameProver() any

	// P2P
	P2PServer() P2PServerDeps
}

// Config holds backend configuration.
type Config struct {
	CacheEnginePayloads int
	SnapshotCapDepth    int
	MigrateEveryBlocks  int
	MaxPeers            int
	P2PPort             int
	DataDir             string
}

// P2PServerDeps provides P2P server access.
type P2PServerDeps interface {
	LocalID() string
	ListenAddr() net.Addr
	ExternalIP() net.IP
	PeersList() []P2PPeerDeps
	AddPeer(url string) error
	PeerCount() int
}

// P2PPeerDeps provides peer info access.
type P2PPeerDeps interface {
	ID() string
	RemoteAddr() string
	Caps() []P2PCapDeps
}

// P2PCapDeps represents a protocol capability.
type P2PCapDeps interface {
	Name() string
	Version() int
}

// NodeAdapter adapts node fields to the NodeDeps interface.
type NodeAdapter struct {
	// Core
	GetBlockchain func() *chain.Blockchain
	GetTxPool     func() *txpool.TxPool
	GetConfig     func() *Config

	// Optional dependencies
	GetGasOracle         func() any
	GetEthHandler        func() any
	GetTxJournal         func() any
	GetSharedPool        func() any
	GetRollupSeq         func() any
	GetMEVConfig         func() *mev.MEVProtectionConfig
	GetSnapshotTree      func() any
	GetTriePruner        func() any
	GetTrieMigrator      func() any
	GetTrieAnnouncer     func() any
	GetStackTrie         func() any
	GetBlobSyncMgr       func() any
	GetStateHealer       func() any
	GetStateSyncSched    func() any
	GetFCStateManager    func() any
	GetFCTracker         func() any
	GetEPBSAuction       func() any
	GetEPBSBuilder       func() any
	GetEPBSEscrow        func() any
	GetEPBSCommit        func() any
	GetEPBSBid           func() any
	GetEPBSMEVBurn       func() any
	GetEngineAuction     func() any
	GetRollupBridge      func() any
	GetRollupAnchor      func() any
	GetRollupProof       func() any
	GetPortalRouter      func() any
	GetEncryptedProtocol func() any
	GetEncryptedPool     func() any
	GetAcctTracker       func() any
	GetNonceTracker      func() any
	GetPayloadChunker    func() any
	GetNonceAnnouncer    func() any
	GetGasRateTracker    func() any
	GetStarkFrameProver  func() any
	GetP2PServer         func() P2PServerDeps
}

// Implement NodeDeps interface for NodeAdapter.
func (n *NodeAdapter) Blockchain() *chain.Blockchain {
	if n.GetBlockchain != nil {
		return n.GetBlockchain()
	}
	return nil
}

func (n *NodeAdapter) TxPool() *txpool.TxPool {
	if n.GetTxPool != nil {
		return n.GetTxPool()
	}
	return nil
}

func (n *NodeAdapter) Config() *Config {
	if n.GetConfig != nil {
		return n.GetConfig()
	}
	return nil
}

func (n *NodeAdapter) GasOracle() any {
	if n.GetGasOracle != nil {
		return n.GetGasOracle()
	}
	return nil
}

func (n *NodeAdapter) EthHandler() any {
	if n.GetEthHandler != nil {
		return n.GetEthHandler()
	}
	return nil
}

func (n *NodeAdapter) TxJournal() any {
	if n.GetTxJournal != nil {
		return n.GetTxJournal()
	}
	return nil
}

func (n *NodeAdapter) SharedPool() any {
	if n.GetSharedPool != nil {
		return n.GetSharedPool()
	}
	return nil
}

func (n *NodeAdapter) RollupSeq() any {
	if n.GetRollupSeq != nil {
		return n.GetRollupSeq()
	}
	return nil
}

func (n *NodeAdapter) MEVConfig() *mev.MEVProtectionConfig {
	if n.GetMEVConfig != nil {
		return n.GetMEVConfig()
	}
	return nil
}

func (n *NodeAdapter) SnapshotTree() any {
	if n.GetSnapshotTree != nil {
		return n.GetSnapshotTree()
	}
	return nil
}

func (n *NodeAdapter) TriePruner() any {
	if n.GetTriePruner != nil {
		return n.GetTriePruner()
	}
	return nil
}

func (n *NodeAdapter) TrieMigrator() any {
	if n.GetTrieMigrator != nil {
		return n.GetTrieMigrator()
	}
	return nil
}

func (n *NodeAdapter) TrieAnnouncer() any {
	if n.GetTrieAnnouncer != nil {
		return n.GetTrieAnnouncer()
	}
	return nil
}

func (n *NodeAdapter) StackTrie() any {
	if n.GetStackTrie != nil {
		return n.GetStackTrie()
	}
	return nil
}

func (n *NodeAdapter) BlobSyncMgr() any {
	if n.GetBlobSyncMgr != nil {
		return n.GetBlobSyncMgr()
	}
	return nil
}

func (n *NodeAdapter) StateHealer() any {
	if n.GetStateHealer != nil {
		return n.GetStateHealer()
	}
	return nil
}

func (n *NodeAdapter) StateSyncSched() any {
	if n.GetStateSyncSched != nil {
		return n.GetStateSyncSched()
	}
	return nil
}

func (n *NodeAdapter) FCStateManager() any {
	if n.GetFCStateManager != nil {
		return n.GetFCStateManager()
	}
	return nil
}

func (n *NodeAdapter) FCTracker() any {
	if n.GetFCTracker != nil {
		return n.GetFCTracker()
	}
	return nil
}

func (n *NodeAdapter) EPBSAuction() any {
	if n.GetEPBSAuction != nil {
		return n.GetEPBSAuction()
	}
	return nil
}

func (n *NodeAdapter) EPBSBuilder() any {
	if n.GetEPBSBuilder != nil {
		return n.GetEPBSBuilder()
	}
	return nil
}

func (n *NodeAdapter) EPBSEscrow() any {
	if n.GetEPBSEscrow != nil {
		return n.GetEPBSEscrow()
	}
	return nil
}

func (n *NodeAdapter) EPBSCommit() any {
	if n.GetEPBSCommit != nil {
		return n.GetEPBSCommit()
	}
	return nil
}

func (n *NodeAdapter) EPBSBid() any {
	if n.GetEPBSBid != nil {
		return n.GetEPBSBid()
	}
	return nil
}

func (n *NodeAdapter) EPBSMEVBurn() any {
	if n.GetEPBSMEVBurn != nil {
		return n.GetEPBSMEVBurn()
	}
	return nil
}

func (n *NodeAdapter) EngineAuction() any {
	if n.GetEngineAuction != nil {
		return n.GetEngineAuction()
	}
	return nil
}

func (n *NodeAdapter) RollupBridge() any {
	if n.GetRollupBridge != nil {
		return n.GetRollupBridge()
	}
	return nil
}

func (n *NodeAdapter) RollupAnchor() any {
	if n.GetRollupAnchor != nil {
		return n.GetRollupAnchor()
	}
	return nil
}

func (n *NodeAdapter) RollupProof() any {
	if n.GetRollupProof != nil {
		return n.GetRollupProof()
	}
	return nil
}

func (n *NodeAdapter) PortalRouter() any {
	if n.GetPortalRouter != nil {
		return n.GetPortalRouter()
	}
	return nil
}

func (n *NodeAdapter) EncryptedProtocol() any {
	if n.GetEncryptedProtocol != nil {
		return n.GetEncryptedProtocol()
	}
	return nil
}

func (n *NodeAdapter) EncryptedPool() any {
	if n.GetEncryptedPool != nil {
		return n.GetEncryptedPool()
	}
	return nil
}

func (n *NodeAdapter) AcctTracker() any {
	if n.GetAcctTracker != nil {
		return n.GetAcctTracker()
	}
	return nil
}

func (n *NodeAdapter) NonceTracker() any {
	if n.GetNonceTracker != nil {
		return n.GetNonceTracker()
	}
	return nil
}

func (n *NodeAdapter) PayloadChunker() any {
	if n.GetPayloadChunker != nil {
		return n.GetPayloadChunker()
	}
	return nil
}

func (n *NodeAdapter) NonceAnnouncer() any {
	if n.GetNonceAnnouncer != nil {
		return n.GetNonceAnnouncer()
	}
	return nil
}

func (n *NodeAdapter) GasRateTracker() any {
	if n.GetGasRateTracker != nil {
		return n.GetGasRateTracker()
	}
	return nil
}

func (n *NodeAdapter) StarkFrameProver() any {
	if n.GetStarkFrameProver != nil {
		return n.GetStarkFrameProver()
	}
	return nil
}

func (n *NodeAdapter) P2PServer() P2PServerDeps {
	if n.GetP2PServer != nil {
		return n.GetP2PServer()
	}
	return nil
}

// =============================================================================
// EngineBackend - Engine API implementation
// =============================================================================

// pendingPayload stores a built payload for later retrieval.
type pendingPayload struct {
	block    *types.Block
	receipts []*types.Receipt
	err      error
	done     chan struct{}
}

const fcuCacheSize = 8

// fcuCacheEntry records a (head,safe,finalized) triple from a processed FCU.
type fcuCacheEntry struct {
	head      types.Hash
	safe      types.Hash
	finalized types.Hash
}

// postFCUWork carries slow state-update work deferred from ForkchoiceUpdated.
type postFCUWork struct {
	fcState    engine.ForkchoiceStateV1
	headBlock  *types.Block
	finalBlock *types.Block
	safeBlock  *types.Block
	hasAttrs   bool
}

// EngineBackend adapts NodeDeps to the engine.Backend interface.
type EngineBackend struct {
	node NodeDeps

	mu           sync.Mutex
	payloads     map[engine.PayloadID]*pendingPayload
	payloadOrder []engine.PayloadID
	maxPayloads  int
	builder      *block.BlockBuilder
	buildMu      sync.Mutex

	fcMu          sync.RWMutex
	safeHash      types.Hash
	finalizedHash types.Hash

	stopCh   chan struct{}
	postFCUCh chan postFCUWork

	fcuCache   [fcuCacheSize]fcuCacheEntry
	fcuCacheWr int
	fcuCacheMu sync.Mutex
}

// txPoolReader wraps txpool.TxPool for the block builder.
type txPoolReader struct {
	pool      *txpool.TxPool
	mevConfig *mev.MEVProtectionConfig
}

func (r *txPoolReader) Pending() []*types.Transaction {
	return r.pool.PendingFlat()
}

// NewEngineBackend creates and starts the engine backend.
func NewEngineBackend(node NodeDeps) *EngineBackend {
	pool := &txPoolReader{pool: node.TxPool(), mevConfig: node.MEVConfig()}
	builder := block.NewBlockBuilder(node.Blockchain().Config(), node.Blockchain(), pool)
	maxPayloads := node.Config().CacheEnginePayloads
	if maxPayloads <= 0 {
		maxPayloads = 32
	}
	b := &EngineBackend{
		node:        node,
		payloads:    make(map[engine.PayloadID]*pendingPayload),
		maxPayloads: maxPayloads,
		builder:     builder,
		stopCh:      make(chan struct{}),
		postFCUCh:   make(chan postFCUWork, 1),
	}
	go b.postFCULoop()
	return b
}

// Close stops the backend.
func (b *EngineBackend) Close() {
	close(b.stopCh)
}

// GetHeadHash returns the current head block hash.
func (b *EngineBackend) GetHeadHash() types.Hash {
	if blk := b.node.Blockchain().CurrentBlock(); blk != nil {
		return blk.Hash()
	}
	return types.Hash{}
}

// GetSafeHash returns the current safe block hash.
func (b *EngineBackend) GetSafeHash() types.Hash {
	if blk := b.node.Blockchain().CurrentSafeBlock(); blk != nil {
		return blk.Hash()
	}
	b.fcMu.RLock()
	defer b.fcMu.RUnlock()
	return b.safeHash
}

// GetFinalizedHash returns the current finalized block hash.
func (b *EngineBackend) GetFinalizedHash() types.Hash {
	if blk := b.node.Blockchain().CurrentFinalBlock(); blk != nil {
		return blk.Hash()
	}
	b.fcMu.RLock()
	defer b.fcMu.RUnlock()
	return b.finalizedHash
}

// ProcessBlock processes a new payload from the consensus layer.
func (b *EngineBackend) ProcessBlock(
	payload *engine.ExecutionPayloadV3,
	expectedBlobVersionedHashes []types.Hash,
	parentBeaconBlockRoot types.Hash,
) (engine.PayloadStatusV1, error) {
	bc := b.node.Blockchain()
	if bc == nil {
		return engine.PayloadStatusV1{Status: engine.StatusInvalid}, nil
	}

	// Check if block already exists
	if block := bc.GetBlock(payload.BlockHash); block != nil {
		slog.Debug("engine_newPayload: block already exists", "hash", payload.BlockHash)
		return engine.PayloadStatusV1{
			Status:          engine.StatusValid,
			LatestValidHash: &payload.BlockHash,
		}, nil
	}

	// Validate parent exists
	parent := bc.GetBlock(payload.ParentHash)
	if parent == nil {
		slog.Warn("engine_newPayload: parent block not found", "parent", payload.ParentHash)
		return engine.PayloadStatusV1{Status: engine.StatusSyncing}, nil
	}

	slog.Info("engine_newPayload: block accepted", "hash", payload.BlockHash, "number", payload.BlockNumber)
	return engine.PayloadStatusV1{
		Status:          engine.StatusValid,
		LatestValidHash: &payload.BlockHash,
	}, nil
}

// postFCULoop handles slow post-FCU state updates.
func (b *EngineBackend) postFCULoop() {
	for {
		select {
		case <-b.stopCh:
			return
		case work := <-b.postFCUCh:
			b.doPostFCUWork(work)
		}
	}
}

// doPostFCUWork performs state updates after ForkchoiceUpdated.
func (b *EngineBackend) doPostFCUWork(work postFCUWork) {
	bc := b.node.Blockchain()
	if bc == nil {
		return
	}

	if work.finalBlock != nil {
		bc.SetFinalized(work.finalBlock)
	}
	if work.safeBlock != nil {
		bc.SetSafe(work.safeBlock)
	}

	b.fcuCacheMu.Lock()
	b.fcuCache[b.fcuCacheWr] = fcuCacheEntry{
		head:      work.fcState.HeadBlockHash,
		safe:      work.fcState.SafeBlockHash,
		finalized: work.fcState.FinalizedBlockHash,
	}
	b.fcuCacheWr = (b.fcuCacheWr + 1) % fcuCacheSize
	b.fcuCacheMu.Unlock()
}

// ForkchoiceUpdated processes a forkchoice update from the consensus layer.
func (b *EngineBackend) ForkchoiceUpdated(
	ctx context.Context,
	fcState engine.ForkchoiceStateV1,
	attrs *engine.PayloadAttributesV1,
) (engine.ForkchoiceUpdatedResult, error) {
	bc := b.node.Blockchain()
	if bc == nil {
		return engine.ForkchoiceUpdatedResult{
			PayloadStatus: engine.PayloadStatusV1{Status: engine.StatusInvalid},
		}, nil
	}

	// Get head block
	headBlock := bc.GetBlock(fcState.HeadBlockHash)
	if headBlock == nil {
		slog.Warn("forkchoiceUpdated: head block not found", "hash", fcState.HeadBlockHash)
		return engine.ForkchoiceUpdatedResult{
			PayloadStatus: engine.PayloadStatusV1{Status: engine.StatusSyncing},
		}, nil
	}

	// Update state
	b.fcMu.Lock()
	b.safeHash = fcState.SafeBlockHash
	b.finalizedHash = fcState.FinalizedBlockHash
	b.fcMu.Unlock()

	// Get safe and finalized blocks
	var safeBlock, finalBlock *types.Block
	if fcState.SafeBlockHash != (types.Hash{}) {
		safeBlock = bc.GetBlock(fcState.SafeBlockHash)
	}
	if fcState.FinalizedBlockHash != (types.Hash{}) {
		finalBlock = bc.GetBlock(fcState.FinalizedBlockHash)
	}

	// Schedule post-FCU work
	select {
	case b.postFCUCh <- postFCUWork{
		fcState:    fcState,
		headBlock:  headBlock,
		finalBlock: finalBlock,
		safeBlock:  safeBlock,
		hasAttrs:   attrs != nil,
	}:
	default:
	}

	slog.Info("forkchoiceUpdated: head updated", "hash", fcState.HeadBlockHash)

	return engine.ForkchoiceUpdatedResult{
		PayloadStatus: engine.PayloadStatusV1{
			Status:          engine.StatusValid,
			LatestValidHash: &fcState.HeadBlockHash,
		},
	}, nil
}

// GetFCStateFromManager retrieves forkchoice state from manager if available.
func (b *EngineBackend) GetFCStateFromManager() *forkchoice.ForkchoiceState {
	mgr := b.node.FCStateManager()
	if mgr == nil {
		return nil
	}
	if fcMgr, ok := mgr.(interface{ GetState() *forkchoice.ForkchoiceState }); ok {
		return fcMgr.GetState()
	}
	return nil
}

// =============================================================================
// NodeBackend - RPC API implementation
// =============================================================================

// NodeBackend provides blockchain access for RPC handlers.
type NodeBackend struct {
	node NodeDeps
}

// NewNodeBackend creates a new NodeBackend.
func NewNodeBackend(node NodeDeps) *NodeBackend {
	return &NodeBackend{node: node}
}

// CurrentBlock returns the current head block.
func (b *NodeBackend) CurrentBlock() *types.Block {
	if bc := b.node.Blockchain(); bc != nil {
		return bc.CurrentBlock()
	}
	return nil
}

// CurrentHeader returns the current head header.
func (b *NodeBackend) CurrentHeader() *types.Header {
	if bc := b.node.Blockchain(); bc != nil {
		if blk := bc.CurrentBlock(); blk != nil {
			return blk.Header()
		}
	}
	return nil
}

// GetBlock retrieves a block by hash.
func (b *NodeBackend) GetBlock(hash types.Hash) *types.Block {
	if bc := b.node.Blockchain(); bc != nil {
		return bc.GetBlock(hash)
	}
	return nil
}

// GetBlockByNumber retrieves a block by number.
func (b *NodeBackend) GetBlockByNumber(number uint64) *types.Block {
	if bc := b.node.Blockchain(); bc != nil {
		return bc.GetBlockByNumber(number)
	}
	return nil
}

// Pending retrieves pending transactions from the pool.
func (b *NodeBackend) Pending() []*types.Transaction {
	if pool := b.node.TxPool(); pool != nil {
		return pool.PendingFlat()
	}
	return nil
}

// GasLimit returns the current gas limit.
func (b *NodeBackend) GasLimit() uint64 {
	if bc := b.node.Blockchain(); bc != nil {
		if blk := bc.CurrentBlock(); blk != nil {
			return blk.GasLimit()
		}
	}
	return 0
}

// BaseFee returns the current base fee.
func (b *NodeBackend) BaseFee() *big.Int {
	if bc := b.node.Blockchain(); bc != nil {
		if blk := bc.CurrentBlock(); blk != nil {
			return blk.BaseFee()
		}
	}
	return nil
}

// SuggestGasPrice returns a suggested gas price.
func (b *NodeBackend) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	if oracle := b.node.GasOracle(); oracle != nil {
		if g, ok := oracle.(interface{ SuggestGasPrice(context.Context) (*big.Int, error) }); ok {
			return g.SuggestGasPrice(ctx)
		}
	}
	if base := b.BaseFee(); base != nil {
		return new(big.Int).Add(base, big.NewInt(1e9)), nil
	}
	return big.NewInt(1e9), nil
}

// SendTx sends a transaction through the pool.
func (b *NodeBackend) SendTx(ctx context.Context, tx *types.Transaction) error {
	if pool := b.node.TxPool(); pool != nil {
		return pool.AddLocal(tx)
	}
	return nil
}

// Stats returns pool statistics.
func (b *NodeBackend) Stats() (int, int) {
	if pool := b.node.TxPool(); pool != nil {
		return pool.PendingCount(), pool.QueuedCount()
	}
	return 0, 0
}

// =============================================================================
// Helpers
// =============================================================================

// ExtractBlockTips returns the effective priority fee for each transaction.
func ExtractBlockTips(txs []*types.Transaction, baseFee *big.Int) []*big.Int {
	tips := make([]*big.Int, 0, len(txs))
	if baseFee == nil {
		baseFee = new(big.Int)
	}
	for _, tx := range txs {
		var tip *big.Int
		switch tx.Type() {
		case types.DynamicFeeTxType:
			tipCap := tx.GasTipCap()
			feeCap := tx.GasFeeCap()
			if tipCap == nil || feeCap == nil {
				continue
			}
			effective := new(big.Int).Sub(feeCap, baseFee)
			if effective.Sign() < 0 {
				continue
			}
			tip = tipCap
			if effective.Cmp(tipCap) < 0 {
				tip = effective
			}
		default:
			gp := tx.GasPrice()
			if gp == nil {
				continue
			}
			tip = new(big.Int).Sub(gp, baseFee)
			if tip.Sign() < 0 {
				continue
			}
		}
		tips = append(tips, new(big.Int).Set(tip))
	}
	return tips
}