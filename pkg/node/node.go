package node

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	_ "net/http/pprof" // register pprof handlers on DefaultServeMux
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/eth2030/eth2030/consensus/vdf"
	"github.com/eth2030/eth2030/core/chain"
	coreconfig "github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/gigagas"
	mevpkg "github.com/eth2030/eth2030/core/mev"
	"github.com/eth2030/eth2030/core/rawdb"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/teragas"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vops"
	"github.com/eth2030/eth2030/crypto"
	dasblobpool "github.com/eth2030/eth2030/das/blobpool"
	dasnetwork "github.com/eth2030/eth2030/das/network"
	dasvalidator "github.com/eth2030/eth2030/das/validator"
	"github.com/eth2030/eth2030/engine"
	"github.com/eth2030/eth2030/eth"
	"github.com/eth2030/eth2030/light"
	"github.com/eth2030/eth2030/p2p"
	"github.com/eth2030/eth2030/p2p/dnsdisc"
	"github.com/eth2030/eth2030/metrics"
	"github.com/eth2030/eth2030/proofs"
	"github.com/eth2030/eth2030/rpc"
	gasrpc "github.com/eth2030/eth2030/rpc/gas"
	rpcmiddleware "github.com/eth2030/eth2030/rpc/middleware"
	ethsync "github.com/eth2030/eth2030/sync"
	syncbeam "github.com/eth2030/eth2030/sync/beam"
	synccheckpoint "github.com/eth2030/eth2030/sync/checkpoint"
	syncchecksync "github.com/eth2030/eth2030/sync/checksync"
	syncinserter "github.com/eth2030/eth2030/sync/inserter"
	syncrangeproof "github.com/eth2030/eth2030/sync/rangeproof"
	syncsupport "github.com/eth2030/eth2030/sync/support"
	"github.com/eth2030/eth2030/txpool"
	"github.com/eth2030/eth2030/txpool/encrypted"
	txjournal "github.com/eth2030/eth2030/txpool/journal"
	"github.com/eth2030/eth2030/txpool/tracking"

	epbsauction "github.com/eth2030/eth2030/epbs/auction"
	epbsbid "github.com/eth2030/eth2030/epbs/bid"
	epbsbuilder "github.com/eth2030/eth2030/epbs/builder"
	epbscommit "github.com/eth2030/eth2030/epbs/commit"
	epbsescrow "github.com/eth2030/eth2030/epbs/escrow"
	epbsmevburn "github.com/eth2030/eth2030/epbs/mevburn"
	epbsslash "github.com/eth2030/eth2030/epbs/slashing"

	rollupanchor "github.com/eth2030/eth2030/rollup/anchor"
	rollupbridge "github.com/eth2030/eth2030/rollup/bridge"
	rollupproof "github.com/eth2030/eth2030/rollup/proof"
	rollupregistry "github.com/eth2030/eth2030/rollup/registry"
	rollupseq "github.com/eth2030/eth2030/rollup/sequencer"

	trieprunestate "github.com/eth2030/eth2030/core/state/pruner"
	engineauction "github.com/eth2030/eth2030/engine/auction"
	rpcregistry "github.com/eth2030/eth2030/rpc/registry"
	trieannounce "github.com/eth2030/eth2030/trie/announce"
	trieprune "github.com/eth2030/eth2030/trie/prune"
	triestack "github.com/eth2030/eth2030/trie/stack"

	enginechunking "github.com/eth2030/eth2030/engine/chunking"
	"github.com/eth2030/eth2030/p2p/discover"
	p2pdispatch "github.com/eth2030/eth2030/p2p/dispatch"
	p2pnat "github.com/eth2030/eth2030/p2p/nat"
	p2pnonce "github.com/eth2030/eth2030/p2p/nonce"
	p2pportal "github.com/eth2030/eth2030/p2p/portal"
	p2preqresp "github.com/eth2030/eth2030/p2p/reqresp"
	p2psnap "github.com/eth2030/eth2030/p2p/snap"
	syncbeacon "github.com/eth2030/eth2030/sync/beacon"
	synchealer "github.com/eth2030/eth2030/sync/healer"
	syncstatesync "github.com/eth2030/eth2030/sync/statesync"
	"github.com/eth2030/eth2030/trie/migrate"
	"github.com/eth2030/eth2030/trie/mpt"
	"github.com/eth2030/eth2030/txpool/shared"
)

// Node is the top-level ETH2030 node that manages all subsystems.
type Node struct {
	config *Config

	// Subsystems.
	db            rawdb.Database
	blockchain    *chain.Blockchain
	txPool        *txpool.TxPool
	rpcServer     *rpc.ExtServer
	rpcHandler    *rpc.Server
	engineServer  *engine.EngineAPI
	p2pServer     *p2p.Server
	metricsServer *http.Server
	wsServer      *http.Server

	// ETH/68 protocol handler (block/tx exchange with peers).
	ethHandler *eth.Handler

	// Sync engine for downloading blocks from peers.
	syncer *ethsync.Downloader

	// Gas oracle for EIP-1559-aware gas price suggestions.
	gasOracle *gasrpc.GasOracle

	// Encrypted mempool: commit-reveal scheme for MEV protection (Hegotá).
	encryptedProtocol *encrypted.EncryptedMempoolProtocol
	encryptedPool     *encrypted.EncryptedPool

	// Txpool lifecycle tracking: per-account nonce/balance and nonce-gap detection.
	acctTracker  *tracking.AcctTrack
	nonceTracker *tracking.NonceTracker

	// MEV protection: sandwich/frontrun detection + fair ordering (Hegotá).
	mevConfig *mevpkg.MEVProtectionConfig

	// Tx journal for crash-recovery of pending transactions.
	txJournal *txjournal.TxJournal

	// Gigagas gas-rate tracker (M+ north star: 1 Ggas/sec).
	gasRateTracker *gigagas.GasRateTracker

	// RPC rate limiter for per-client/per-method protection.
	rpcRateLimiter *rpcmiddleware.RPCRateLimiter

	// VOPS: validity-only partial statelessness executor (I+ roadmap).
	vopsExecutor *vops.PartialExecutor

	// DAS: data availability network manager and validator (PeerDAS, EIP-7594).
	dasNetMgr    *dasnetwork.DASNetworkManager
	dasValidator *dasvalidator.DAValidator

	// Teragas L2 blob scheduler (1 GByte/sec north star).
	teragasScheduler *teragas.TeragasScheduler

	// VDF randomness consensus (L+ secret proposers).
	vdfConsensus *vdf.VDFConsensus

	// Sync chain inserter with verification metrics.
	chainInserter *syncinserter.ChainInserter

	// Sync support: checkpoint store, checkpoint syncer, progress tracker, and pipeline.
	checkpointStore   *synccheckpoint.CheckpointStore
	checkpointSyncer  *syncchecksync.CheckpointSyncer
	syncProgressTrack *syncsupport.ProgressTracker
	syncPipeline      *syncsupport.SyncPipeline
	rangeProver       *syncrangeproof.RangeProver
	beamSync          *syncbeam.BeamSync

	// Light client subsystem.
	lightClient *light.LightClient

	// DAS sparse blob pool (custody-based pruning, EIP-4844/7594).
	dasBlobPool *dasblobpool.SparseBlobPool

	// ePBS sub-systems (EIP-7732): auction, bid scoring, builder market,
	// commitment chain, escrow, MEV burn tracker, and slashing.
	epbsAuction  *epbsauction.AuctionEngine
	epbsBid      *epbsbid.BidScoreCalculator
	epbsBuilder  *epbsbuilder.BuilderMarket
	epbsCommit   *epbscommit.CommitmentChain
	epbsEscrow   *epbsescrow.BidEscrow
	epbsMEVBurn  *epbsmevburn.MEVBurnTracker
	epbsSlashing *epbsslash.SlashingEngine

	// Native rollup sub-systems (EIP-8079): anchor contract, bridge,
	// proof generator, rollup registry, and sequencer.
	rollupAnchor   *rollupanchor.Contract
	rollupBridge   *rollupbridge.Bridge
	rollupProof    *rollupproof.MessageProofGenerator
	rollupRegistry *rollupregistry.Registry
	rollupSeq      *rollupseq.Sequencer

	// Engine builder auction (EL-side ePBS slot auction).
	engineAuction *engineauction.BuilderAuction

	// Trie sub-systems: state pruner, stack trie builder, and EIP-8077 announcer.
	statePruner   *trieprunestate.Pruner
	triePruner    *trieprune.StatePruner
	stackTrie     *triestack.StackTrieNodeCollector
	trieAnnouncer *trieannounce.AnnounceBinaryTrie

	// RPC method registry for dynamic routing.
	rpcRegistry *rpcregistry.MethodRegistry

	// P2P sub-systems: NAT traversal, message dispatch, nonce announcer, req/resp.
	natMgr         *p2pnat.NATManager
	p2pDispatch    *p2pdispatch.MessageRouter
	nonceAnnouncer *p2pnonce.NonceAnnouncer
	reqRespMgr     *p2preqresp.ReqRespManager

	// Tx pool shared mempool abstraction.
	sharedPool *shared.SharedMempool

	// Snap sync: state healer and state sync scheduler.
	stateHealer    *synchealer.StateHealer
	stateSyncSched *syncstatesync.StateSyncScheduler

	// Portal network content DB and DHT router.
	portalDB     *p2pportal.ContentDB
	portalRouter *p2pportal.DHTRouter

	// Snap protocol server handler.
	snapHandler *p2psnap.ServerHandler

	// Beacon sync manager and blob recovery.
	beaconSyncer *syncbeacon.BeaconSyncer
	blobSyncMgr  *syncbeacon.BlobSyncManager

	// Engine payload chunker (streaming large payloads to CL).
	payloadChunker *enginechunking.PayloadChunker

	// MPT → BinaryTrie incremental migrator (EIP-7864 trie migration).
	trieMigrator *migrate.IncrementalMigrator

	// EP-6 BB-1.x: anonymous transaction transport manager.
	transportMgr *p2p.TransportManager

	// EP-3: STARK mempool P2P subsystem.
	topicMgr         *p2p.TopicManager
	starkAgg         *txpool.STARKAggregator
	starkFrameProver proofs.ValidationFrameProver // non-nil when StarkValidationFrames=true
	currentSlot      atomic.Uint64                // updated on each FCU; used for peer-tick TTL eviction

	mu      sync.Mutex
	running bool
	stop    chan struct{}
}

// New creates a new Node with the given configuration. It initializes
// all subsystems but does not start any network services.
func New(config *Config) (*Node, error) {
	if config == nil {
		c := DefaultConfig()
		config = &c
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// GAP-7.2: select BLS signature backend based on config.
	if config.BLSBackend == "pure-go" {
		crypto.SetBLSBackend(&crypto.PureGoBLSBackend{})
		slog.Info("BLS backend: pure-go")
	} else {
		slog.Info("BLS backend: blst (default)")
	}

	// GAP-5.2: log finality mode selection.
	slog.Info("finality mode", "mode", config.FinalityMode)

	// Auto-generate JWT secret if not provided.
	if err := ensureJWTSecret(config); err != nil {
		return nil, fmt.Errorf("jwt secret: %w", err)
	}

	n := &Node{
		config: config,
		stop:   make(chan struct{}),
	}

	// Initialize persistent database.
	db, err := rawdb.NewFileDB(config.ResolvePath("chaindata"))
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}
	n.db = db

	// Initialize genesis state before resolving the genesis block so that
	// SetupGenesisBlock can populate alloc accounts into it.
	statedb := state.NewMemoryStateDB()

	// Resolve chain config and genesis block.
	var chainConfig *coreconfig.ChainConfig
	var genesis *types.Block
	if config.GenesisPath != "" {
		genSpec, err := loadGenesisFile(config)
		if err != nil {
			return nil, fmt.Errorf("load genesis file: %w", err)
		}
		chainConfig = genSpec.Config
		// SetupGenesisBlock applies alloc to statedb and sets the correct
		// state root in the genesis header, so our hash matches the CL's.
		genesis = genSpec.SetupGenesisBlock(statedb)
		// Derive network ID from genesis chain ID unless the user explicitly
		// passed a non-default value (default is 1; 0 also means "auto").
		if (config.NetworkID == 0 || config.NetworkID == 1) &&
			genSpec.Config != nil && genSpec.Config.ChainID != nil {
			config.NetworkID = genSpec.Config.ChainID.Uint64()
		}
	} else {
		chainConfig = chainConfigForNetwork(config.Network)
		// Apply any fork overrides on top of the standard chain config.
		applyForkOverrides(chainConfig, config)
		genesis = makeGenesisBlock()
	}

	bc, err := chain.NewBlockchain(chainConfig, genesis, statedb, n.db)
	if err != nil {
		return nil, fmt.Errorf("init blockchain: %w", err)
	}
	n.blockchain = bc

	// Initialize gas oracle for EIP-1559-aware gas price suggestions.
	n.gasOracle = gasrpc.NewGasOracle(gasrpc.DefaultGasOracleConfig())

	// Initialize encrypted mempool (commit-reveal MEV protection, Hegotá).
	n.encryptedProtocol = encrypted.NewEncryptedMempoolProtocol(encrypted.EncryptedProtocolConfig{
		CommitWindowBlocks: 2,
		RevealWindowBlocks: 2,
		MaxPendingCommits:  4096,
		MinRevealers:       1,
	})
	n.encryptedPool = encrypted.NewEncryptedPool()

	// Initialize per-account nonce/balance trackers for the tx pool.
	n.acctTracker = tracking.NewAcctTrack(statedb)
	n.nonceTracker = tracking.NewNonceTracker(tracking.DefaultNonceTrackerConfig(), statedb)

	// Initialize MEV protection (sandwich/frontrun detection + fair ordering).
	n.mevConfig = mevpkg.DefaultMEVProtectionConfig()

	// Initialize VOPS partial executor (I+ validity-only partial statelessness).
	n.vopsExecutor = vops.NewPartialExecutor(vops.DefaultVOPSConfig())

	// Initialize DAS network manager and DA validator (PeerDAS, EIP-7594).
	n.dasNetMgr = dasnetwork.NewDASNetworkManager(
		dasnetwork.DefaultNetworkConfig(),
		&stubCustodyManager{},
	)
	n.dasValidator = dasvalidator.NewDAValidator(dasvalidator.DefaultDAValidatorConfig())

	// Initialize teragas L2 blob scheduler (L2 1 GByte/sec north star).
	n.teragasScheduler = teragas.NewTeragasScheduler(teragas.DefaultSchedulerConfig())

	// Initialize VDF consensus for slot-based randomness (L+ secret proposers).
	n.vdfConsensus = vdf.NewVDFConsensus(vdf.DefaultVDFConsensusConfig())

	// Initialize chain inserter with verification metrics (sync/inserter).
	n.chainInserter = syncinserter.NewChainInserter(
		syncinserter.DefaultChainInserterConfig(),
		n.blockchain,
	)

	// Initialize sync support: checkpoint store, syncer, progress tracker, pipeline.
	n.checkpointStore = synccheckpoint.NewCheckpointStore(synccheckpoint.DefaultCheckpointStoreConfig())
	n.checkpointSyncer = syncchecksync.NewCheckpointSyncer(syncchecksync.DefaultCheckpointConfig())
	n.syncProgressTrack = syncsupport.NewProgressTracker()
	n.syncPipeline = syncsupport.NewSyncPipeline(syncsupport.DefaultPipelineConfig())
	n.rangeProver = syncrangeproof.NewRangeProver()
	n.beamSync = syncbeam.NewBeamSync(&stubBeamFetcher{})

	// Initialize light client (header sync + proof verification).
	n.lightClient = light.NewLightClient()

	// Initialize DAS sparse blob pool (custody-based pruning, EIP-4844/7594).
	n.dasBlobPool = dasblobpool.NewSparseBlobPool(4) // 4 subnets default

	// Initialize tx journal for pending tx persistence across restarts.
	journalPath := config.ResolvePath("transactions.rlp")
	if j, jerr := txjournal.NewTxJournal(journalPath); jerr != nil {
		slog.Warn("tx journal init failed", "path", journalPath, "err", jerr)
	} else {
		n.txJournal = j
		// Replay previously journaled txs into the pool (crash recovery).
		if journaledTxs, _, loadErr := txjournal.Load(journalPath); loadErr == nil {
			for _, tx := range journaledTxs {
				_ = n.txPool.AddLocal(tx) // best-effort replay; ignore errors
			}
			slog.Info("tx journal replayed", "count", len(journaledTxs))
		}
	}

	// Initialize transaction pool.
	poolCfg := txpool.DefaultConfig()
	// BB-2.2: propagate experimental LocalTx flag into pool config.
	poolCfg.AllowLocalTx = config.ExperimentalLocalTx
	n.txPool = txpool.New(poolCfg, bc.State())

	// Initialize EP-3 STARK mempool gossip subsystem.
	n.topicMgr = p2p.NewTopicManager(p2p.DefaultTopicParams())
	broadcaster := p2p.NewMempoolBroadcaster(n.topicMgr)
	n.starkAgg = txpool.NewSTARKAggregator("eth2030-node")
	n.starkAgg.SetBroadcaster(broadcaster)
	// Subscribe to incoming STARK ticks from peers.
	if err := n.topicMgr.Subscribe(p2p.STARKMempoolTick, func(_ p2p.GossipTopic, _ p2p.MessageID, data []byte) {
		var tick txpool.MempoolAggregationTick
		if err := tick.UnmarshalBinary(data); err != nil {
			slog.Debug("stark tick decode error", "err", err)
			return
		}
		slot := n.currentSlot.Load()
		if err := n.starkAgg.MergeTickAtSlot(&tick, slot); err != nil {
			slog.Debug("stark tick merge error", "err", err)
		}
	}); err != nil {
		slog.Warn("stark mempool tick subscribe failed", "err", err)
	}

	// EP-3 US-PQ-6: compile AA proof circuit on startup (non-fatal; logs result).
	go func() {
		circuit, err := proofs.CompileAACircuit()
		if err != nil {
			slog.Warn("AA circuit compile failed", "err", err)
			return
		}
		_, _, err = proofs.SetupKeys(circuit)
		if err != nil {
			slog.Warn("AA circuit key setup failed", "err", err)
			return
		}
		slog.Info("AA proof circuit ready", "name", circuit.Name, "inputs", circuit.PublicInputCount)
	}()

	// Create STARK validation frame prover when enabled.
	if config.StarkValidationFrames {
		n.starkFrameProver = proofs.NewSTARKValidationFrameProver()
	}

	// Log EP-3 configuration.
	slog.Info("EP-3 post-quantum config",
		"lean_chain", config.LeanAvailableChainMode,
		"lean_validators", config.LeanAvailableChainValidators,
		"stark_frames", config.StarkValidationFrames,
	)

	// Initialize ETH/68 protocol handler for block and transaction exchange.
	n.ethHandler = eth.NewHandler(bc, n.txPool, config.NetworkID)

	// Initialize sync downloader and wire it with the eth handler.
	n.syncer = ethsync.NewDownloader(nil) // nil uses DefaultDownloaderConfig
	fetcher := eth.NewPeerFetcher(bc)
	n.syncer.SetFetchers(fetcher, fetcher, bc)
	// Wire the sync notifier so new block announcements trigger sync.
	n.ethHandler.SetSyncNotifier(&nodeSyncTrigger{dl: n.syncer})
	n.ethHandler.SetDownloader(n.syncer)

	// Initialize P2P server with bootnodes, discovery port, and NAT.
	// Register the ETH protocol so peers can exchange blocks and transactions.
	n.p2pServer = p2p.NewServer(p2p.Config{
		ListenAddr:     config.P2PAddr(),
		MaxPeers:       config.MaxPeers,
		BootstrapNodes: config.Bootnodes,
		DiscoveryPort:  config.EffectiveDiscoveryPort(),
		NAT:            config.NAT,
		Protocols:      []p2p.Protocol{n.ethHandler.Protocol()},
	})

	// EP-6 BB-1.1/1.2/1.3: initialize anonymous transport manager.
	// Parse --mixnet mode; default to simulated when unset.
	tmCfg := p2p.DefaultTransportConfig()
	if mode, err := p2p.ParseMixnetMode(config.MixnetMode); err == nil {
		tmCfg.Mode = mode
	}
	// Use the node's own RPC endpoint for transaction forwarding via external transports.
	tmCfg.RPCEndpoint = fmt.Sprintf("http://%s", config.RPCAddr())
	n.transportMgr = p2p.NewTransportManagerWithConfig(tmCfg)

	// Select the best available transport and register it.
	// When user requested a specific mode, honour it directly without probing.
	// When mode is simulated (default), probe Tor then Nym before falling back.
	var selectedTransport p2p.AnonymousTransport
	switch tmCfg.Mode {
	case p2p.ModeTorSocks5:
		selectedTransport = p2p.NewTorTransport(&p2p.TorConfig{
			ProxyAddr:   tmCfg.TorProxyAddr,
			RPCEndpoint: tmCfg.RPCEndpoint,
			DialTimeout: tmCfg.DialTimeout,
			MaxPending:  256,
		})
		slog.Info("anonymous transport: tor", "proxy", tmCfg.TorProxyAddr)
	case p2p.ModeNymSocks5:
		selectedTransport = p2p.NewNymTransport(&p2p.NymConfig{
			ProxyAddr:   tmCfg.NymProxyAddr,
			RPCEndpoint: tmCfg.RPCEndpoint,
			DialTimeout: tmCfg.DialTimeout,
			MaxPending:  256,
		})
		slog.Info("anonymous transport: nym", "proxy", tmCfg.NymProxyAddr)
	default:
		// Auto-probe: Tor → Nym → simulated.
		n.transportMgr.SelectBestTransport()
		selectedTransport = p2p.NewMixnetTransport(nil)
	}
	if err := n.transportMgr.RegisterTransport(selectedTransport); err != nil {
		slog.Warn("transport register failed", "err", err)
	}

	// Initialize RPC server with blockchain backend.
	backend := newNodeBackend(n)
	adminBackend := newNodeAdminBackend(n)
	netBackend := newNodeNetBackend(n)
	n.rpcHandler = rpc.NewServer(backend)
	n.rpcHandler.SetAdminBackend(adminBackend)
	n.rpcHandler.SetNetBackend(netBackend)
	n.rpcServer = rpc.NewExtServer(backend, rpc.ServerConfig{
		MaxRequestSize:   config.RPCMaxRequestSize,
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		IdleTimeout:      120 * time.Second,
		ShutdownTimeout:  10 * time.Second,
		CORSAllowOrigins: config.RPCCORSAllowedOrigins(),
		AuthSecret:       config.RPCAuthSecret,
		RateLimitPerSec:  config.RPCRateLimitPerSec,
		MaxBatchSize:     config.RPCMaxBatchSize,
	})
	n.rpcServer.SetAdminBackend(adminBackend)
	n.rpcServer.SetNetBackend(netBackend)

	// Wire per-client/per-method rate limiter into the ExtServer middleware chain.
	n.rpcRateLimiter = rpcmiddleware.NewRPCRateLimiter(rpcmiddleware.DefaultRPCRateLimitConfig())
	n.rpcServer.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr
			if idx := strings.LastIndex(clientIP, ":"); idx >= 0 {
				clientIP = clientIP[:idx]
			}
			if !n.rpcRateLimiter.Allow(clientIP, r.URL.Path) {
				http.Error(w, "rate limited", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Initialize gigagas gas-rate tracker (M+ 1 Ggas/sec north star).
	n.gasRateTracker = gigagas.NewGasRateTracker(100)

	// Initialize ePBS sub-systems (EIP-7732): auction, bid scoring, builder market,
	// commitment chain, bid escrow, MEV burn tracker, and slashing engine.
	n.epbsAuction = epbsauction.NewAuctionEngine(epbsauction.DefaultAuctionEngineConfig())
	if bsc, bscErr := epbsbid.NewBidScoreCalculator(epbsbid.DefaultBidScoreConfig()); bscErr == nil {
		n.epbsBid = bsc
	} else {
		slog.Warn("epbs bid scorer init failed", "err", bscErr)
	}
	n.epbsBuilder = epbsbuilder.NewBuilderMarket(epbsbuilder.DefaultBuilderMarketConfig())
	n.epbsCommit = epbscommit.NewCommitmentChain()
	n.epbsEscrow = epbsescrow.NewBidEscrow(1024)
	n.epbsMEVBurn = epbsmevburn.NewMEVBurnTracker(epbsmevburn.DefaultMEVBurnConfig())
	n.epbsSlashing = epbsslash.NewSlashingEngine(epbsslash.DefaultPenaltyMultipliers(), 100)

	// Initialize engine payload chunker (128 KB segments for streaming to CL).
	n.payloadChunker = enginechunking.NewPayloadChunker(128 * 1024)

	// Initialize MPT→BinaryTrie incremental migrator (EIP-7864).
	// Pass ChainConfig so BinaryTrieHashFuncAt selects sha256/blake3 per fork.
	migCfg := migrate.DefaultMigrationConfig()
	migCfg.ChainConfig = chainConfig
	n.trieMigrator = migrate.NewIncrementalMigrator(mpt.New(), migCfg)

	// Initialize Portal network: content DB + DHT router (history/state content).
	var portalNodeID p2pportal.NodeID // zero node ID until real discovery is wired
	n.portalDB = p2pportal.NewContentDB(p2pportal.DefaultContentDBConfig(portalNodeID))
	var portalSelfID [32]byte // zero self-ID for now
	portalKT := discover.NewKademliaTable(portalSelfID, discover.DefaultKademliaConfig())
	n.portalRouter = p2pportal.NewDHTRouter(portalKT, p2pportal.DefaultDHTRouterConfig())

	// Initialize snap protocol server handler (serve snap requests to syncing peers).
	n.snapHandler = p2psnap.NewServerHandler(&stubSnapStateBackend{})

	// Initialize beacon syncer and blob sync manager.
	n.beaconSyncer = syncbeacon.NewBeaconSyncer(syncbeacon.DefaultBeaconSyncConfig())
	n.blobSyncMgr = syncbeacon.NewBlobSyncManager(syncbeacon.DefaultBlobSyncConfig())

	// Initialize P2P sub-systems: NAT traversal, message dispatch, nonce
	// announcer (EIP-8077), and request/response framing.
	n.natMgr = p2pnat.NewNATManager(p2pnat.NATManagerConfig{
		MappingLifetime: 20 * time.Minute,
		RenewInterval:   10 * time.Minute,
	})
	n.p2pDispatch = p2pdispatch.NewMessageRouter(p2pdispatch.RouterConfig{})
	n.nonceAnnouncer = p2pnonce.NewNonceAnnouncer()
	rrProto := p2preqresp.NewReqRespProtocol(p2preqresp.DefaultProtocolConfig())
	n.reqRespMgr = p2preqresp.NewReqRespManager(rrProto, p2preqresp.DefaultRetryConfig())

	// Initialize shared mempool abstraction (MineableSet interface).
	n.sharedPool = shared.NewSharedMempool(shared.DefaultSharedMempoolConfig())

	// Initialize snap sync state healer and state sync scheduler (stub writer).
	n.stateHealer = synchealer.NewStateHealer(types.Hash{}, &stubStateWriter{})
	n.stateSyncSched = syncstatesync.NewStateSyncScheduler(&stubStateSyncWriter{}, nil)

	// Initialize engine EL-side builder auction.
	n.engineAuction = engineauction.NewBuilderAuction(engineauction.DefaultAuctionConfig())

	// Initialize trie sub-systems: state pruner (bloom-filter reachability),
	// stack-trie node collector (snap-sync trie building), and binary-trie
	// announcement set (EIP-8077 trie proof gossip).
	if fdb, ok := n.db.(*rawdb.FileDB); ok {
		n.statePruner = trieprunestate.NewPruner(
			trieprunestate.PrunerConfig{BloomSize: trieprunestate.DefaultBloomSize},
			fdb,
		)
	}
	n.triePruner = trieprune.NewStatePruner(128) // keep 128 recent state roots
	n.stackTrie = triestack.NewStackTrieNodeCollector()
	n.trieAnnouncer = trieannounce.NewAnnounceBinaryTrie()

	// Initialize RPC method registry for dynamic method routing.
	n.rpcRegistry = rpcregistry.NewMethodRegistry()

	// Initialize native rollup sub-systems (EIP-8079): anchor contract, L1↔L2
	// bridge, proof generator, rollup registry, and sequencer.
	n.rollupAnchor = rollupanchor.NewContract()
	n.rollupBridge = rollupbridge.NewBridge(rollupbridge.DefaultConfig())
	n.rollupProof = rollupproof.NewMessageProofGenerator()
	n.rollupRegistry = rollupregistry.NewRegistry()
	n.rollupSeq = rollupseq.NewSequencer(rollupseq.DefaultConfig())

	// Initialize Engine API server.
	engineBackend := newEngineBackend(n)
	n.engineServer = engine.NewEngineAPI(engineBackend)
	// Forward eth_/web3_/net_/admin_ methods on the engine port to the RPC handler.
	n.engineServer.SetEthHandler(n.rpcHandler.Handler())
	n.engineServer.SetMaxRequestSize(config.EngineMaxRequestSize)
	if token, err := resolveEngineAuthToken(config); err != nil {
		return nil, fmt.Errorf("engine auth: %w", err)
	} else if token != "" {
		n.engineServer.SetAuthSecret(token)
	}

	return n, nil
}

// Start starts all node subsystems in order.
func (n *Node) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		return errors.New("node already running")
	}

	slog.Info("starting ETH2030 node", "network", n.config.Network)

	// Start DAS network manager (PeerDAS EIP-7594).
	if n.dasNetMgr != nil {
		n.dasNetMgr.Start()
		slog.Info("DAS network manager started")
	}

	// Start light client subsystem.
	if n.lightClient != nil {
		if err := n.lightClient.Start(); err != nil {
			slog.Warn("light client start failed", "err", err)
		} else {
			slog.Info("light client started")
		}
	}

	// Start STARK mempool aggregator.
	if err := n.starkAgg.Start(); err != nil {
		return fmt.Errorf("start stark aggregator: %w", err)
	}

	// EP-6 BB-1.x: start all registered anonymous transports.
	for _, err := range n.transportMgr.StartAll() {
		slog.Warn("anonymous transport start error", "err", err)
	}

	// Start P2P server.
	if err := n.p2pServer.Start(); err != nil {
		return fmt.Errorf("start p2p: %w", err)
	}
	slog.Info("P2P server listening", "addr", n.p2pServer.ListenAddr())

	// Start NAT port mapping manager.
	if err := n.natMgr.Start(); err != nil {
		slog.Warn("NAT manager start failed", "err", err)
	}

	// Wire beacon syncer fetcher (marks it active; full fetcher set by sync coordinator).
	if n.beaconSyncer != nil {
		n.beaconSyncer.SetFetcher(nil)
	}

	// Bootstrap peers from DNS discovery (EIP-1459) if configured.
	if n.config.DNSDiscovery != "" {
		go n.runDNSDiscovery(n.config.DNSDiscovery)
	}

	// Start JSON-RPC server (ExtServer handles auth, rate limiting, CORS, body limits).
	go func() {
		slog.Info("RPC server listening", "addr", n.config.RPCAddr())
		if err := n.rpcServer.Start(n.config.RPCAddr()); err != nil && err != http.ErrServerClosed {
			slog.Error("RPC server error", "err", err)
		}
	}()

	// Start Engine API server.
	go func() {
		slog.Info("Engine API server listening", "addr", n.config.AuthListenAddr())
		if err := n.engineServer.Start(n.config.AuthListenAddr()); err != nil {
			slog.Error("Engine API error", "err", err)
		}
	}()

	// Start WebSocket RPC server if enabled.
	if n.config.WSEnabled {
		wsHandler := buildWSHandler(n.rpcHandler, n.config.WSOrigins)
		n.wsServer = &http.Server{
			Addr:    n.config.WSListenAddr(),
			Handler: wsHandler,
		}
		go func() {
			slog.Info("WebSocket RPC server listening", "addr", n.config.WSListenAddr())
			if err := n.wsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("WebSocket server error", "err", err)
			}
		}()
	}

	// Start metrics server if enabled.
	if n.config.Metrics {
		mux := http.NewServeMux()
		mux.Handle("/debug/vars", expvar.Handler())
		pe := metrics.NewPrometheusExporter(metrics.DefaultRegistry, metrics.DefaultPrometheusConfig())
		mux.Handle("/metrics", pe.Handler())
		n.metricsServer = &http.Server{
			Addr:    n.config.MetricsListenAddr(),
			Handler: mux,
		}
		go func() {
			slog.Info("Metrics server listening", "addr", n.config.MetricsListenAddr())
			if err := n.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("Metrics server error", "err", err)
			}
		}()
	}

	n.running = true
	slog.Info("node started successfully")
	return nil
}

// Stop gracefully shuts down all subsystems in reverse order.
func (n *Node) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.running {
		if n.db != nil {
			if err := n.db.Close(); err != nil {
				slog.Warn("database close error", "err", err)
			}
			n.db = nil
		}
		select {
		case <-n.stop:
			// stop channel already closed.
		default:
			close(n.stop)
		}
		return nil
	}

	slog.Info("stopping ETH2030 node")

	// Stop DAS network manager.
	if n.dasNetMgr != nil {
		n.dasNetMgr.Stop()
	}

	// Stop light client.
	if n.lightClient != nil {
		n.lightClient.Stop()
	}

	// Stop teragas blob scheduler.
	if n.teragasScheduler != nil {
		n.teragasScheduler.Stop()
	}

	// Stop STARK mempool aggregator.
	n.starkAgg.Stop()

	// EP-6 BB-1.x: stop all anonymous transports.
	for _, err := range n.transportMgr.StopAll() {
		slog.Warn("anonymous transport stop error", "err", err)
	}

	// Stop Engine API.
	if err := n.engineServer.Stop(); err != nil {
		slog.Warn("Engine API stop error", "err", err)
	}

	// Stop RPC server.
	if n.rpcServer != nil {
		if err := n.rpcServer.Stop(); err != nil {
			slog.Warn("RPC server stop error", "err", err)
		}
	}

	// Stop WebSocket server.
	if n.wsServer != nil {
		if err := n.wsServer.Close(); err != nil {
			slog.Warn("WebSocket server stop error", "err", err)
		}
	}

	// Stop metrics server.
	if n.metricsServer != nil {
		if err := n.metricsServer.Close(); err != nil {
			slog.Warn("Metrics server stop error", "err", err)
		}
	}

	// Stop NAT port mapping and P2P message subsystems.
	n.natMgr.Stop()
	n.p2pDispatch.Close()
	n.reqRespMgr.Close()
	if n.portalDB != nil {
		n.portalDB.Close()
	}

	// Stop P2P server.
	n.p2pServer.Stop()

	// Close database.
	if err := n.db.Close(); err != nil {
		slog.Warn("database close error", "err", err)
	}
	n.db = nil

	n.running = false
	select {
	case <-n.stop:
		// stop channel already closed.
	default:
		close(n.stop)
	}
	slog.Info("node stopped")
	return nil
}

// Wait blocks until the node is stopped.
func (n *Node) Wait() {
	<-n.stop
}

// Blockchain returns the blockchain instance.
func (n *Node) Blockchain() *chain.Blockchain {
	return n.blockchain
}

// TxPool returns the transaction pool.
func (n *Node) TxPool() *txpool.TxPool {
	return n.txPool
}

// Config returns the node configuration.
func (n *Node) Config() *Config {
	return n.config
}

// Running reports whether the node is currently running.
func (n *Node) Running() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.running
}

// chainConfigForNetwork returns the chain config for the given network name.
func chainConfigForNetwork(network string) *coreconfig.ChainConfig {
	switch network {
	case "mainnet":
		return coreconfig.MainnetConfig
	case "sepolia":
		return coreconfig.SepoliaConfig
	case "holesky":
		return coreconfig.HoleskyConfig
	default:
		return coreconfig.MainnetConfig
	}
}

func resolveEngineAuthToken(cfg *Config) (string, error) {
	if cfg.EngineAuthToken != "" {
		return strings.TrimSpace(cfg.EngineAuthToken), nil
	}
	if cfg.EngineAuthTokenPath == "" {
		return "", nil
	}

	data, err := os.ReadFile(cfg.EngineAuthTokenPath)
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("engine auth token file is empty")
	}
	return token, nil
}

// genesisForNetwork returns the genesis specification for the given network.
func genesisForNetwork(network string) *coreconfig.Genesis {
	switch network {
	case "mainnet":
		return coreconfig.DefaultGenesisBlock()
	case "sepolia":
		return coreconfig.DefaultSepoliaGenesisBlock()
	case "holesky":
		return coreconfig.DefaultHoleskyGenesisBlock()
	default:
		return coreconfig.DefaultGenesisBlock()
	}
}

// makeGenesisBlock creates a minimal genesis block.
func makeGenesisBlock() *types.Block {
	header := &types.Header{
		Number:     big.NewInt(0),
		GasLimit:   30_000_000,
		GasUsed:    0,
		Time:       0,
		Difficulty: new(big.Int),
		BaseFee:    big.NewInt(1_000_000_000), // 1 gwei
		UncleHash:  types.EmptyUncleHash,
	}
	return types.NewBlock(header, nil)
}

// ensureJWTSecret generates a random JWT secret and writes it to the
// configured path if JWTSecret is empty or the file does not yet exist.
// The parent directory is created if necessary.
func ensureJWTSecret(config *Config) error {
	path := config.JWTSecretPath()

	// If the file already exists, nothing to do.
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(config.DataDir, 0700); err != nil {
		return fmt.Errorf("create datadir for jwt secret: %w", err)
	}

	// Generate 32 random bytes.
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return fmt.Errorf("generate random secret: %w", err)
	}

	// Write as hex string (0x-prefixed to match geth convention).
	content := "0x" + hex.EncodeToString(secret) + "\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("write jwt secret to %s: %w", path, err)
	}

	slog.Info("generated JWT secret", "path", path)
	return nil
}

// buildWSHandler creates an http.Handler that accepts WebSocket upgrade
// requests and serves JSON-RPC 2.0 over the persistent connection.
// The origins list restricts which Origin headers are allowed;
// an empty list or ["*"] allows all origins.
func buildWSHandler(handler *rpc.Server, origins []string) http.Handler {
	allowAll := len(origins) == 0 || (len(origins) == 1 && origins[0] == "*")
	upgrader := websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			if allowAll {
				return true
			}
			return sliceContains(origins, r.Header.Get("Origin"))
		},
	}
	httpHandler := handler.Handler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !websocket.IsWebSocketUpgrade(r) {
			httpHandler.ServeHTTP(w, r)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Debug("ws upgrade error", "err", err)
			return
		}
		defer conn.Close()
		serveWSConn(conn, handler)
	})
}

// serveWSConn processes JSON-RPC requests over a WebSocket connection by
// forwarding each message to the rpc.Server handler via an in-memory HTTP round-trip.
func serveWSConn(conn *websocket.Conn, handler *rpc.Server) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				slog.Debug("ws read error", "err", err)
			}
			return
		}
		// Forward the JSON-RPC request to the handler and capture the response.
		// We use an in-memory ResponseWriter to collect the output.
		respBytes := dispatchWSRequest(handler, msg)
		if err := conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
			slog.Debug("ws write error", "err", err)
			return
		}
	}
}

// dispatchWSRequest routes a single JSON-RPC payload through the rpc.Server
// and returns the serialised response.
func dispatchWSRequest(handler *rpc.Server, body []byte) []byte {
	req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	if err != nil {
		return errorResponse("parse error")
	}
	req.Header.Set("Content-Type", "application/json")
	rw := &bufResponseWriter{header: make(http.Header)}
	handler.Handler().ServeHTTP(rw, req)
	return rw.buf
}

// errorResponse returns a minimal JSON-RPC error response.
func errorResponse(msg string) []byte {
	type rpcErr struct {
		JSONRPC string `json:"jsonrpc"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	r := rpcErr{JSONRPC: "2.0"}
	r.Error.Code = -32700
	r.Error.Message = msg
	b, _ := json.Marshal(r)
	return b
}

// bufResponseWriter is an in-memory http.ResponseWriter.
type bufResponseWriter struct {
	header http.Header
	buf    []byte
	status int
}

func (b *bufResponseWriter) Header() http.Header    { return b.header }
func (b *bufResponseWriter) WriteHeader(status int) { b.status = status }
func (b *bufResponseWriter) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// corsMiddleware wraps handler with CORS headers and virtual-host checking.
// An empty or ["*"] domains list allows all origins.
// An empty or ["*"] vhosts list allows all hosts.
func corsMiddleware(handler http.Handler, domains, vhosts []string) http.Handler {
	allowAllOrigins := len(domains) == 0 || (len(domains) == 1 && domains[0] == "*")
	allowAllHosts := len(vhosts) == 0 || (len(vhosts) == 1 && vhosts[0] == "*")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Virtual-host check.
		if !allowAllHosts {
			host := r.Host
			if idx := strings.LastIndex(host, ":"); idx >= 0 {
				host = host[:idx]
			}
			if !sliceContains(vhosts, host) {
				http.Error(w, "invalid host", http.StatusForbidden)
				return
			}
		}

		// CORS headers.
		origin := r.Header.Get("Origin")
		if origin != "" {
			if allowAllOrigins || sliceContains(domains, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		handler.ServeHTTP(w, r)
	})
}

// sliceContains reports whether s contains elem (case-insensitive).
func sliceContains(s []string, elem string) bool {
	lower := strings.ToLower(elem)
	for _, v := range s {
		if strings.ToLower(v) == lower {
			return true
		}
	}
	return false
}

// runDNSDiscovery resolves peers from a DNS EIP-1459 tree URL and connects
// to them via the P2P server. It runs once at startup and logs results.
func (n *Node) runDNSDiscovery(treeURL string) {
	// Parse "enrtree://<pubkey>@<domain>" format.
	domain := treeURL
	if idx := strings.Index(treeURL, "@"); idx >= 0 {
		domain = treeURL[idx+1:]
	}

	client := dnsdisc.NewDNSClient(dnsdisc.DNSConfig{
		Domain: domain,
	})

	nodes, err := client.Resolve(domain)
	if err != nil {
		slog.Warn("DNS discovery failed", "url", treeURL, "err", err)
		return
	}

	slog.Info("DNS discovery found peers", "count", len(nodes), "domain", domain)
	for _, node := range nodes {
		addr := node.String()
		if addr == "" {
			continue
		}
		if err := n.p2pServer.AddPeer(addr); err != nil {
			slog.Debug("DNS discovery: failed to add peer", "addr", addr, "err", err)
		}
	}
}

// stubCustodyManager is a no-op CustodyManager used until a real peer custody
// backend is wired. It returns an empty peer list for every column.
type stubCustodyManager struct{}

func (s *stubCustodyManager) FindPeersForColumn(_ uint64) ([][32]byte, error) {
	return nil, nil
}

// stubBeamFetcher is a no-op BeamStateFetcher used until the P2P layer is
// wired to serve on-demand state requests during beam sync.
type stubBeamFetcher struct{}

func (s *stubBeamFetcher) FetchAccount(_ types.Address) (*syncbeam.BeamAccountData, error) {
	return nil, errors.New("beam: no network fetcher wired")
}

func (s *stubBeamFetcher) FetchStorage(_ types.Address, _ types.Hash) (types.Hash, error) {
	return types.Hash{}, errors.New("beam: no network fetcher wired")
}

// stubSnapStateBackend is a no-op p2psnap.StateBackend used until the real
// state snapshot layer is connected for snap protocol serving.
type stubSnapStateBackend struct{}

func (s *stubSnapStateBackend) AccountIterator(_ types.Hash, _ types.Hash, _ func(types.Hash, []byte) bool) error {
	return nil
}
func (s *stubSnapStateBackend) StorageIterator(_ types.Hash, _ types.Hash, _ []byte, _ func(types.Hash, []byte) bool) error {
	return nil
}
func (s *stubSnapStateBackend) Code(_ types.Hash) ([]byte, error)                 { return nil, nil }
func (s *stubSnapStateBackend) TrieNode(_ types.Hash, _ []byte) ([]byte, error)   { return nil, nil }
func (s *stubSnapStateBackend) AccountProof(_, _ types.Hash) ([][]byte, error)    { return nil, nil }
func (s *stubSnapStateBackend) StorageProof(_, _, _ types.Hash) ([][]byte, error) { return nil, nil }

// stubStateWriter is a no-op synchealer.StateWriter used until a real trie
// database is wired as the state healer write target.
type stubStateWriter struct{}

func (s *stubStateWriter) WriteAccount(_ types.Hash, _ synchealer.AccountData) error { return nil }
func (s *stubStateWriter) WriteStorage(_, _ types.Hash, _ []byte) error              { return nil }
func (s *stubStateWriter) WriteBytecode(_ types.Hash, _ []byte) error                { return nil }
func (s *stubStateWriter) WriteTrieNode(_ []byte, _ []byte) error                    { return nil }
func (s *stubStateWriter) HasBytecode(_ types.Hash) bool                             { return false }
func (s *stubStateWriter) HasTrieNode(_ []byte) bool                                 { return false }
func (s *stubStateWriter) MissingTrieNodes(_ types.Hash, _ int) [][]byte             { return nil }

// stubStateSyncWriter is a no-op syncstatesync.StateWriter used until the real
// snap-sync state write target is wired.
type stubStateSyncWriter struct{}

func (s *stubStateSyncWriter) WriteAccount(_ types.Hash, _ syncstatesync.AccountData) error {
	return nil
}
func (s *stubStateSyncWriter) WriteStorage(_, _ types.Hash, _ []byte) error  { return nil }
func (s *stubStateSyncWriter) WriteBytecode(_ types.Hash, _ []byte) error    { return nil }
func (s *stubStateSyncWriter) WriteTrieNode(_ []byte, _ []byte) error        { return nil }
func (s *stubStateSyncWriter) HasBytecode(_ types.Hash) bool                 { return false }
func (s *stubStateSyncWriter) HasTrieNode(_ []byte) bool                     { return false }
func (s *stubStateSyncWriter) MissingTrieNodes(_ types.Hash, _ int) [][]byte { return nil }

// nodeSyncTrigger adapts the sync.Downloader to the eth.SyncNotifier interface.
// It starts a sync goroutine when a new block is announced by a peer.
type nodeSyncTrigger struct {
	dl *ethsync.Downloader
}

func (s *nodeSyncTrigger) OnNewBlock(peerID string, blockNum uint64) {
	go func() {
		if err := s.dl.Start(blockNum); err != nil {
			slog.Debug("sync trigger failed", "peer", peerID, "block", blockNum, "err", err)
		}
	}()
}

// slashingEngineAdapter adapts epbsslash.SlashingEngine to the
// coreconfig.PaymasterSlasher interface used by the state processor.
type slashingEngineAdapter struct {
	eng *epbsslash.SlashingEngine
}

func (a *slashingEngineAdapter) SlashOnBadSettlement(addr types.Address) error {
	_, err := a.eng.EvaluateAll(nil, nil, addr)
	return err
}

var _ coreconfig.PaymasterSlasher = (*slashingEngineAdapter)(nil)
