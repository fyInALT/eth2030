package metrics

// Pre-defined metrics for the ETH2030 Ethereum execution client. All metrics
// live in DefaultRegistry so they are globally accessible without passing a
// registry around.

var (
	// ---- Chain metrics ----

	// ChainHeight tracks the latest block number.
	ChainHeight = DefaultRegistry.Gauge("chain.height")
	// ChainHeadFinalized tracks the latest finalized block number.
	ChainHeadFinalized = DefaultRegistry.Gauge("chain.head.finalized")
	// ChainHeadSafe tracks the latest safe block number.
	ChainHeadSafe = DefaultRegistry.Gauge("chain.head.safe")
	// BlockProcessTime records block processing duration in milliseconds.
	BlockProcessTime = DefaultRegistry.Histogram("chain.block_process_ms")
	// BlocksInserted counts blocks successfully appended to the chain.
	BlocksInserted = DefaultRegistry.Counter("chain.blocks_inserted")
	// ReorgsDetected counts chain reorganisation events.
	ReorgsDetected = DefaultRegistry.Counter("chain.reorgs")

	// ---- Transaction pool metrics ----

	// TxPoolPending tracks the number of pending transactions.
	TxPoolPending = DefaultRegistry.Gauge("txpool.pending")
	// TxPoolQueued tracks the number of queued transactions.
	TxPoolQueued = DefaultRegistry.Gauge("txpool.queued")
	// TxPoolAdded counts transactions added to the pool.
	TxPoolAdded = DefaultRegistry.Counter("txpool.added")
	// TxPoolDropped counts transactions dropped from the pool.
	TxPoolDropped = DefaultRegistry.Counter("txpool.dropped")

	// ---- P2P metrics ----

	// PeersConnected tracks the current number of connected peers.
	PeersConnected = DefaultRegistry.Gauge("p2p.peers")
	// MessagesReceived counts devp2p messages received.
	MessagesReceived = DefaultRegistry.Counter("p2p.messages_received")
	// MessagesSent counts devp2p messages sent.
	MessagesSent = DefaultRegistry.Counter("p2p.messages_sent")

	// ---- RPC metrics ----

	// RPCRequests counts incoming JSON-RPC requests.
	RPCRequests = DefaultRegistry.Counter("rpc.requests")
	// RPCErrors counts JSON-RPC requests that returned an error.
	RPCErrors = DefaultRegistry.Counter("rpc.errors")
	// RPCLatency records JSON-RPC request latency in milliseconds.
	RPCLatency = DefaultRegistry.Histogram("rpc.latency_ms")

	// ---- EVM metrics ----

	// EVMExecutions counts EVM call/create invocations.
	EVMExecutions = DefaultRegistry.Counter("evm.executions")
	// EVMGasUsed counts total gas consumed by EVM execution.
	EVMGasUsed = DefaultRegistry.Counter("evm.gas_used")

	// ---- Engine API metrics ----

	// EngineNewPayload counts engine_newPayload calls.
	EngineNewPayload = DefaultRegistry.Counter("engine.new_payload")
	// EngineFCU counts engine_forkchoiceUpdated calls.
	EngineFCU = DefaultRegistry.Counter("engine.forkchoice_updated")

	// ---- Async Payload Builder metrics ----

	// EnginePayloadBuildTotal counts total payload builds started.
	EnginePayloadBuildTotal = DefaultRegistry.Counter("engine.payload_build_total")
	// EnginePayloadBuildSuccess counts successful payload builds.
	EnginePayloadBuildSuccess = DefaultRegistry.Counter("engine.payload_build_success")
	// EnginePayloadBuildFailed counts failed payload builds.
	EnginePayloadBuildFailed = DefaultRegistry.Counter("engine.payload_build_failed")
	// EnginePayloadBuildTimeout counts payload builds that timed out.
	EnginePayloadBuildTimeout = DefaultRegistry.Counter("engine.payload_build_timeout")
	// EnginePayloadBuildQueueSize tracks the current queue size.
	EnginePayloadBuildQueueSize = DefaultRegistry.Gauge("engine.payload_build_queue_size")
	// EnginePayloadBuildActive tracks the number of active builds.
	EnginePayloadBuildActive = DefaultRegistry.Gauge("engine.payload_build_active")

	// ---- Priority Handler metrics ----

	// EngineRequestHighTotal counts high priority requests.
	EngineRequestHighTotal = DefaultRegistry.Counter("engine.request_high_total")
	// EngineRequestNormalTotal counts normal priority requests.
	EngineRequestNormalTotal = DefaultRegistry.Counter("engine.request_normal_total")
	// EngineRequestLowTotal counts low priority requests.
	EngineRequestLowTotal = DefaultRegistry.Counter("engine.request_low_total")
	// EngineRequestHighActive tracks active high priority requests.
	EngineRequestHighActive = DefaultRegistry.Gauge("engine.request_high_active")
	// EngineRequestNormalActive tracks active normal priority requests.
	EngineRequestNormalActive = DefaultRegistry.Gauge("engine.request_normal_active")
	// EngineRequestLowActive tracks active low priority requests.
	EngineRequestLowActive = DefaultRegistry.Gauge("engine.request_low_active")
	// EngineRequestTimeoutTotal counts requests that timed out.
	EngineRequestTimeoutTotal = DefaultRegistry.Counter("engine.request_timeout_total")
)
