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

	"github.com/eth2030/eth2030/core/chain"
	coreconfig "github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/rawdb"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/engine"
	"github.com/eth2030/eth2030/eth"
	"github.com/eth2030/eth2030/p2p"
	"github.com/eth2030/eth2030/proofs"
	"github.com/eth2030/eth2030/rpc"
	ethsync "github.com/eth2030/eth2030/sync"
	"github.com/eth2030/eth2030/txpool"
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
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			// Simple text metrics endpoint: delegate to expvar for now.
			expvar.Handler().ServeHTTP(w, r)
		})
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
