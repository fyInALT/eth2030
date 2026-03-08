package ethapi

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
)

// defaultPriorityFee is 1 Gwei, the suggested default max priority fee.
var defaultPriorityFee = big.NewInt(1_000_000_000)

// EthAPI implements the eth_, net_, web3_, and txpool_ namespace JSON-RPC methods.
type EthAPI struct {
	backend Backend
	subs    SubscriptionService
	txpool  *TxPoolAPI
}

// NewEthAPI creates a new EthAPI that uses the provided SubscriptionService.
// Callers should pass a *SubscriptionManager from the top-level rpc package.
func NewEthAPI(backend Backend, subs SubscriptionService) *EthAPI {
	return &EthAPI{
		backend: backend,
		subs:    subs,
		txpool:  NewTxPoolAPI(backend),
	}
}

// Subs returns the SubscriptionService used by this API.
// This is primarily used by tests in dependent packages.
func (api *EthAPI) Subs() SubscriptionService { return api.subs }

// HandleRequest dispatches a JSON-RPC request to the appropriate method.
func (api *EthAPI) HandleRequest(req *Request) *Response {
	switch req.Method {
	case "eth_chainId":
		return api.chainID(req)
	case "eth_blockNumber":
		return api.blockNumber(req)
	case "eth_getBlockByNumber":
		return api.getBlockByNumber(req)
	case "eth_getBlockByHash":
		return api.getBlockByHash(req)
	case "eth_getBalance":
		return api.getBalance(req)
	case "eth_getTransactionCount":
		return api.getTransactionCount(req)
	case "eth_getCode":
		return api.getCode(req)
	case "eth_getStorageAt":
		return api.getStorageAt(req)
	case "eth_gasPrice":
		return api.gasPrice(req)
	case "eth_getTransactionByHash":
		return api.getTransactionByHash(req)
	case "eth_getTransactionReceipt":
		return api.getTransactionReceipt(req)
	case "eth_call":
		return api.ethCall(req)
	case "eth_estimateGas":
		return api.estimateGas(req)
	case "eth_sendRawTransaction":
		return api.sendRawTransaction(req)
	case "eth_getLogs":
		return api.getLogs(req)
	case "eth_getBlockReceipts":
		return api.getBlockReceipts(req)
	case "eth_maxPriorityFeePerGas":
		return api.maxPriorityFeePerGas(req)
	case "eth_feeHistory":
		return api.feeHistory(req)
	case "eth_syncing":
		return api.syncing(req)
	case "eth_createAccessList":
		return api.createAccessList(req)
	case "eth_subscribe":
		return api.ethSubscribe(req)
	case "eth_unsubscribe":
		return api.ethUnsubscribe(req)
	case "eth_newFilter":
		return api.newFilter(req)
	case "eth_newBlockFilter":
		return api.newBlockFilter(req)
	case "eth_newPendingTransactionFilter":
		return api.newPendingTransactionFilter(req)
	case "eth_getFilterChanges":
		return api.getFilterChanges(req)
	case "eth_getFilterLogs":
		return api.getFilterLogs(req)
	case "eth_uninstallFilter":
		return api.uninstallFilter(req)
	case "eth_getProof":
		return api.getProof(req)
	case "eth_getHeaderByNumber":
		return api.getHeaderByNumber(req)
	case "eth_getHeaderByHash":
		return api.getHeaderByHash(req)
	case "eth_getTransactionByBlockHashAndIndex":
		return api.getTransactionByBlockHashAndIndex(req)
	case "eth_getTransactionByBlockNumberAndIndex":
		return api.getTransactionByBlockNumberAndIndex(req)
	case "eth_getBlockTransactionCountByHash":
		return api.getBlockTransactionCountByHash(req)
	case "eth_getBlockTransactionCountByNumber":
		return api.getBlockTransactionCountByNumber(req)
	case "eth_accounts":
		return api.accounts(req)
	case "eth_coinbase":
		return api.coinbase(req)
	case "eth_mining":
		return api.mining(req)
	case "eth_hashrate":
		return api.hashrate(req)
	case "eth_protocolVersion":
		return api.protocolVersion(req)
	case "eth_getUncleCountByBlockHash":
		return api.getUncleCountByBlockHash(req)
	case "eth_getUncleCountByBlockNumber":
		return api.getUncleCountByBlockNumber(req)
	case "eth_getUncleByBlockHashAndIndex":
		return api.getUncleByBlockHashAndIndex(req)
	case "eth_getUncleByBlockNumberAndIndex":
		return api.getUncleByBlockNumberAndIndex(req)
	case "eth_blobBaseFee":
		return api.getBlobBaseFee(req)
	case "debug_traceTransaction":
		return api.debugTraceTransaction(req)
	case "debug_traceCall":
		return api.debugTraceCall(req)
	case "debug_traceBlockByNumber":
		return api.debugTraceBlockByNumber(req)
	case "debug_traceBlockByHash":
		return api.debugTraceBlockByHash(req)
	case "web3_clientVersion":
		return api.clientVersion(req)
	case "web3_sha3":
		return api.web3Sha3(req)
	case "net_version":
		return api.netVersion(req)
	case "net_listening":
		return api.netListening(req)
	case "net_peerCount":
		return api.netPeerCount(req)
	case "txpool_status":
		return api.txpool.Status(req)
	case "txpool_content":
		return api.txpool.Content(req)
	case "txpool_inspect":
		return api.txpool.Inspect(req)
	default:
		return errorResponse(req.ID, ErrCodeMethodNotFound, fmt.Sprintf("method %q not found", req.Method))
	}
}

// historyPruned returns true if the given block number's body/receipt data
// has been pruned per EIP-4444.
func (api *EthAPI) historyPruned(blockNum uint64) bool {
	oldest := api.backend.HistoryOldestBlock()
	return oldest > 0 && blockNum < oldest
}

func (api *EthAPI) chainID(req *Request) *Response {
	id := api.backend.ChainID()
	return successResponse(req.ID, encodeBigInt(id))
}

func (api *EthAPI) blockNumber(req *Request) *Response {
	header := api.backend.CurrentHeader()
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "no current header")
	}
	return successResponse(req.ID, encodeUint64(header.Number.Uint64()))
}

func (api *EthAPI) getBlockByNumber(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing block number parameter")
	}

	var bn BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	// Parse the optional fullTx boolean (second param).
	fullTx := false
	if len(req.Params) > 1 {
		_ = json.Unmarshal(req.Params[1], &fullTx)
	}

	header := api.backend.HeaderByNumber(bn)
	if header == nil {
		return successResponse(req.ID, nil)
	}

	// EIP-4444: check if block body has been pruned.
	if api.historyPruned(header.Number.Uint64()) {
		if fullTx {
			return errorResponse(req.ID, ErrCodeHistoryPruned,
				"historical block body pruned (EIP-4444)")
		}
		return successResponse(req.ID, FormatHeader(header))
	}

	// Load the full block so that body fields (transactions, withdrawals)
	// are always populated per the Ethereum JSON-RPC spec.
	block := api.backend.BlockByNumber(bn)
	if block != nil {
		return successResponse(req.ID, FormatBlock(block, fullTx))
	}
	return successResponse(req.ID, FormatHeader(header))
}

func (api *EthAPI) getBlockByHash(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing block hash parameter")
	}

	var hashHex string
	if err := json.Unmarshal(req.Params[0], &hashHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	// Parse the optional fullTx boolean (second param).
	fullTx := false
	if len(req.Params) > 1 {
		_ = json.Unmarshal(req.Params[1], &fullTx)
	}

	hash := types.HexToHash(hashHex)
	header := api.backend.HeaderByHash(hash)
	if header == nil {
		return successResponse(req.ID, nil)
	}

	// EIP-4444: check if block body has been pruned.
	if api.historyPruned(header.Number.Uint64()) {
		if fullTx {
			return errorResponse(req.ID, ErrCodeHistoryPruned,
				"historical block body pruned (EIP-4444)")
		}
		return successResponse(req.ID, FormatHeader(header))
	}

	// Load the full block so that body fields (transactions, withdrawals)
	// are always populated per the Ethereum JSON-RPC spec.
	block := api.backend.BlockByHash(hash)
	if block != nil {
		return successResponse(req.ID, FormatBlock(block, fullTx))
	}
	return successResponse(req.ID, FormatHeader(header))
}

func (api *EthAPI) getBalance(req *Request) *Response {
	if len(req.Params) < 2 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing address or block number")
	}

	var addrHex string
	if err := json.Unmarshal(req.Params[0], &addrHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	var bn BlockNumber
	if err := json.Unmarshal(req.Params[1], &bn); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	header := api.backend.HeaderByNumber(bn)
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}

	statedb, err := api.backend.StateAt(header.Root)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}

	addr := types.HexToAddress(addrHex)
	balance := statedb.GetBalance(addr)
	return successResponse(req.ID, encodeBigInt(balance))
}

func (api *EthAPI) getTransactionCount(req *Request) *Response {
	if len(req.Params) < 2 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing address or block number")
	}

	var addrHex string
	if err := json.Unmarshal(req.Params[0], &addrHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	var bn BlockNumber
	if err := json.Unmarshal(req.Params[1], &bn); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	header := api.backend.HeaderByNumber(bn)
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}

	statedb, err := api.backend.StateAt(header.Root)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}

	addr := types.HexToAddress(addrHex)
	nonce := statedb.GetNonce(addr)
	return successResponse(req.ID, encodeUint64(nonce))
}

func (api *EthAPI) getCode(req *Request) *Response {
	if len(req.Params) < 2 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing address or block number")
	}

	var addrHex string
	if err := json.Unmarshal(req.Params[0], &addrHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	var bn BlockNumber
	if err := json.Unmarshal(req.Params[1], &bn); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	header := api.backend.HeaderByNumber(bn)
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}

	statedb, err := api.backend.StateAt(header.Root)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}

	addr := types.HexToAddress(addrHex)
	code := statedb.GetCode(addr)
	return successResponse(req.ID, encodeBytes(code))
}

func (api *EthAPI) getStorageAt(req *Request) *Response {
	if len(req.Params) < 3 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing address, slot, or block number")
	}

	var addrHex, slotHex string
	if err := json.Unmarshal(req.Params[0], &addrHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}
	if err := json.Unmarshal(req.Params[1], &slotHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	var bn BlockNumber
	if err := json.Unmarshal(req.Params[2], &bn); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	header := api.backend.HeaderByNumber(bn)
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}

	statedb, err := api.backend.StateAt(header.Root)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}

	addr := types.HexToAddress(addrHex)
	slot := types.HexToHash(slotHex)
	value := statedb.GetState(addr, slot)
	return successResponse(req.ID, encodeHash(value))
}

func (api *EthAPI) gasPrice(req *Request) *Response {
	price := api.backend.SuggestGasPrice()
	if price == nil {
		price = new(big.Int)
	}
	return successResponse(req.ID, encodeBigInt(price))
}

func (api *EthAPI) clientVersion(req *Request) *Response {
	return successResponse(req.ID, "ETH2030/v0.1.0")
}

func (api *EthAPI) netVersion(req *Request) *Response {
	id := api.backend.ChainID()
	return successResponse(req.ID, id.String())
}

func (api *EthAPI) netListening(req *Request) *Response {
	return successResponse(req.ID, true)
}

func (api *EthAPI) netPeerCount(req *Request) *Response {
	return successResponse(req.ID, encodeUint64(0))
}

func (api *EthAPI) web3Sha3(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing data parameter")
	}

	var dataHex string
	if err := json.Unmarshal(req.Params[0], &dataHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	data := fromHexBytes(dataHex)
	hash := crypto.Keccak256Hash(data)
	return successResponse(req.ID, encodeHash(hash))
}

// maxPriorityFeePerGas returns the suggested priority fee (1 Gwei default).
func (api *EthAPI) maxPriorityFeePerGas(req *Request) *Response {
	return successResponse(req.ID, encodeBigInt(defaultPriorityFee))
}

// FeeHistoryResult is the response for eth_feeHistory.
type FeeHistoryResult struct {
	OldestBlock   string     `json:"oldestBlock"`
	BaseFeePerGas []string   `json:"baseFeePerGas"`
	GasUsedRatio  []float64  `json:"gasUsedRatio"`
	Reward        [][]string `json:"reward,omitempty"`
}

// feeHistory returns base fee and gas usage history over a range of blocks.
func (api *EthAPI) feeHistory(req *Request) *Response {
	if len(req.Params) < 2 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing blockCount or newestBlock")
	}

	// Parse block count (hex or decimal)
	var blockCountHex string
	if err := json.Unmarshal(req.Params[0], &blockCountHex); err != nil {
		// Try as integer
		var blockCount int
		if err2 := json.Unmarshal(req.Params[0], &blockCount); err2 != nil {
			return errorResponse(req.ID, ErrCodeInvalidParams, "invalid blockCount: "+err.Error())
		}
		blockCountHex = fmt.Sprintf("0x%x", blockCount)
	}
	blockCount := parseHexUint64(blockCountHex)
	if blockCount == 0 || blockCount > 1024 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "blockCount must be 1..1024")
	}

	var newestBN BlockNumber
	if err := json.Unmarshal(req.Params[1], &newestBN); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid newestBlock: "+err.Error())
	}

	// Parse optional reward percentiles
	var rewardPercentiles []float64
	if len(req.Params) > 2 {
		if err := json.Unmarshal(req.Params[2], &rewardPercentiles); err != nil {
			return errorResponse(req.ID, ErrCodeInvalidParams, "invalid rewardPercentiles: "+err.Error())
		}
	}

	// Resolve newest block
	newestHeader := api.backend.HeaderByNumber(newestBN)
	if newestHeader == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}
	newestNum := newestHeader.Number.Uint64()

	// Calculate block range
	oldest := uint64(0)
	if newestNum+1 >= blockCount {
		oldest = newestNum + 1 - blockCount
	}

	result := &FeeHistoryResult{
		OldestBlock: encodeUint64(oldest),
	}

	// Collect baseFeePerGas and gasUsedRatio for each block in range,
	// plus the baseFee of the next block (blockCount + 1 entries total).
	for i := oldest; i <= newestNum+1; i++ {
		header := api.backend.HeaderByNumber(BlockNumber(i))
		if header != nil && header.BaseFee != nil {
			result.BaseFeePerGas = append(result.BaseFeePerGas, encodeBigInt(header.BaseFee))
		} else {
			result.BaseFeePerGas = append(result.BaseFeePerGas, "0x0")
		}

		// gasUsedRatio only for blocks in the range (not the extra entry).
		if i <= newestNum {
			if header != nil && header.GasLimit > 0 {
				ratio := float64(header.GasUsed) / float64(header.GasLimit)
				result.GasUsedRatio = append(result.GasUsedRatio, ratio)
			} else {
				result.GasUsedRatio = append(result.GasUsedRatio, 0)
			}
		}
	}

	// If reward percentiles are requested, return default priority fee for each.
	if len(rewardPercentiles) > 0 {
		for i := oldest; i <= newestNum; i++ {
			rewards := make([]string, len(rewardPercentiles))
			for j := range rewardPercentiles {
				rewards[j] = encodeBigInt(defaultPriorityFee)
			}
			result.Reward = append(result.Reward, rewards)
		}
	}

	return successResponse(req.ID, result)
}

// SyncStatus is the response for eth_syncing when the node is syncing.
type SyncStatus struct {
	StartingBlock string `json:"startingBlock"`
	CurrentBlock  string `json:"currentBlock"`
	HighestBlock  string `json:"highestBlock"`
}

// syncing returns the sync status. Returns false when fully synced.
func (api *EthAPI) syncing(req *Request) *Response {
	// For now, we report as fully synced.
	return successResponse(req.ID, false)
}

// AccessListResult is the response for eth_createAccessList.
type AccessListResult struct {
	AccessList []AccessListEntry `json:"accessList"`
	GasUsed    string            `json:"gasUsed"`
}

// AccessListEntry is a single entry in an access list result.
type AccessListEntry struct {
	Address     string   `json:"address"`
	StorageKeys []string `json:"storageKeys"`
}

// createAccessList simulates a tx and returns an access list.
func (api *EthAPI) createAccessList(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing call arguments")
	}

	var args CallArgs
	if err := json.Unmarshal(req.Params[0], &args); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	bn := LatestBlockNumber
	if len(req.Params) > 1 {
		if err := json.Unmarshal(req.Params[1], &bn); err != nil {
			return errorResponse(req.ID, ErrCodeInvalidParams, "invalid block number: "+err.Error())
		}
	}

	from, to, gas, value, data := parseCallArgs(&args)

	// Execute the call to determine gas usage.
	_, gasUsed, err := api.backend.EVMCall(from, to, data, gas, value, bn)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, "execution error: "+err.Error())
	}

	// A full implementation would trace storage accesses during execution.
	// For now, return an empty access list with the gas used.
	result := &AccessListResult{
		AccessList: []AccessListEntry{},
		GasUsed:    encodeUint64(gasUsed),
	}

	return successResponse(req.ID, result)
}

// criteriaToQuery converts a JSON-RPC FilterCriteria to an internal FilterQuery.
func criteriaToQuery(c FilterCriteria, backend Backend) FilterQuery {
	var q FilterQuery

	if c.FromBlock != nil {
		var from uint64
		if *c.FromBlock == LatestBlockNumber {
			header := backend.CurrentHeader()
			if header != nil {
				from = header.Number.Uint64()
			}
		} else {
			from = uint64(*c.FromBlock)
		}
		q.FromBlock = &from
	}

	if c.ToBlock != nil {
		var to uint64
		if *c.ToBlock == LatestBlockNumber {
			header := backend.CurrentHeader()
			if header != nil {
				to = header.Number.Uint64()
			}
		} else {
			to = uint64(*c.ToBlock)
		}
		q.ToBlock = &to
	}

	for _, addrHex := range c.Addresses {
		q.Addresses = append(q.Addresses, types.HexToAddress(addrHex))
	}

	for _, topicList := range c.Topics {
		var hashes []types.Hash
		for _, topicHex := range topicList {
			hashes = append(hashes, types.HexToHash(topicHex))
		}
		q.Topics = append(q.Topics, hashes)
	}

	return q
}

// ethSubscribe creates a new subscription (WebSocket-oriented).
func (api *EthAPI) ethSubscribe(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing subscription type")
	}

	var subType string
	if err := json.Unmarshal(req.Params[0], &subType); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	switch subType {
	case "newHeads":
		id := api.subs.Subscribe(SubNewHeads, FilterQuery{})
		return successResponse(req.ID, id)
	case "logs":
		var query FilterQuery
		if len(req.Params) > 1 {
			var criteria FilterCriteria
			if err := json.Unmarshal(req.Params[1], &criteria); err != nil {
				return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
			}
			query = criteriaToQuery(criteria, api.backend)
		}
		id := api.subs.Subscribe(SubLogs, query)
		return successResponse(req.ID, id)
	case "newPendingTransactions":
		id := api.subs.Subscribe(SubPendingTx, FilterQuery{})
		return successResponse(req.ID, id)
	default:
		return errorResponse(req.ID, ErrCodeInvalidParams, fmt.Sprintf("unsupported subscription type: %q", subType))
	}
}

// ethUnsubscribe removes a subscription by ID.
func (api *EthAPI) ethUnsubscribe(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing subscription ID")
	}

	var subID string
	if err := json.Unmarshal(req.Params[0], &subID); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	ok := api.subs.Unsubscribe(subID)
	return successResponse(req.ID, ok)
}

// newFilter creates a log filter and returns its filter ID.
func (api *EthAPI) newFilter(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing filter criteria")
	}

	var criteria FilterCriteria
	if err := json.Unmarshal(req.Params[0], &criteria); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	query := criteriaToQuery(criteria, api.backend)
	id := api.subs.NewLogFilter(query)
	return successResponse(req.ID, id)
}

// newBlockFilter creates a block filter and returns its filter ID.
func (api *EthAPI) newBlockFilter(req *Request) *Response {
	id := api.subs.NewBlockFilter()
	return successResponse(req.ID, id)
}

// newPendingTransactionFilter creates a pending tx filter.
func (api *EthAPI) newPendingTransactionFilter(req *Request) *Response {
	id := api.subs.NewPendingTxFilter()
	return successResponse(req.ID, id)
}

// getFilterChanges returns new results since the last poll.
func (api *EthAPI) getFilterChanges(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing filter ID")
	}

	var filterID string
	if err := json.Unmarshal(req.Params[0], &filterID); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	result, ok := api.subs.GetFilterChanges(filterID)
	if !ok {
		return errorResponse(req.ID, ErrCodeInvalidParams, "filter not found")
	}

	// Format the result depending on filter type.
	switch v := result.(type) {
	case []*types.Log:
		rpcLogs := make([]*RPCLog, len(v))
		for i, log := range v {
			rpcLogs[i] = FormatLog(log)
		}
		return successResponse(req.ID, rpcLogs)
	case []types.Hash:
		hashes := make([]string, len(v))
		for i, h := range v {
			hashes[i] = encodeHash(h)
		}
		return successResponse(req.ID, hashes)
	default:
		return successResponse(req.ID, result)
	}
}

// getFilterLogs returns all logs matching an installed log filter.
func (api *EthAPI) getFilterLogs(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing filter ID")
	}

	var filterID string
	if err := json.Unmarshal(req.Params[0], &filterID); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	logs, ok := api.subs.GetFilterLogs(filterID)
	if !ok {
		return errorResponse(req.ID, ErrCodeInvalidParams, "filter not found")
	}

	rpcLogs := make([]*RPCLog, len(logs))
	for i, log := range logs {
		rpcLogs[i] = FormatLog(log)
	}
	return successResponse(req.ID, rpcLogs)
}

// uninstallFilter removes a filter by ID.
func (api *EthAPI) uninstallFilter(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing filter ID")
	}

	var filterID string
	if err := json.Unmarshal(req.Params[0], &filterID); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	ok := api.subs.Uninstall(filterID)
	return successResponse(req.ID, ok)
}

// matchLog checks whether a log matches the filter criteria.
func matchLog(log *types.Log, addrFilter map[types.Address]bool, topicFilter [][]types.Hash) bool {
	// Check address filter
	if len(addrFilter) > 0 && !addrFilter[log.Address] {
		return false
	}

	// Check topic filters
	for i, topics := range topicFilter {
		if len(topics) == 0 {
			continue // wildcard position
		}
		if i >= len(log.Topics) {
			return false
		}
		matched := false
		for _, topic := range topics {
			if log.Topics[i] == topic {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// getLogs returns logs matching the given filter criteria.
func (api *EthAPI) getLogs(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing filter criteria")
	}

	var criteria FilterCriteria
	if err := json.Unmarshal(req.Params[0], &criteria); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	// Determine block range
	fromBlock := uint64(0)
	toBlock := uint64(0)

	current := api.backend.CurrentHeader()
	if current == nil {
		return errorResponse(req.ID, ErrCodeInternal, "no current block")
	}
	currentNum := current.Number.Uint64()

	if criteria.FromBlock != nil {
		if *criteria.FromBlock == LatestBlockNumber {
			fromBlock = currentNum
		} else {
			fromBlock = uint64(*criteria.FromBlock)
		}
	}
	if criteria.ToBlock != nil {
		if *criteria.ToBlock == LatestBlockNumber {
			toBlock = currentNum
		} else {
			toBlock = uint64(*criteria.ToBlock)
		}
	} else {
		toBlock = currentNum
	}

	// Collect matching logs
	var result []*RPCLog

	// Parse address filter
	addrFilter := make(map[types.Address]bool)
	for _, addrHex := range criteria.Addresses {
		addrFilter[types.HexToAddress(addrHex)] = true
	}

	// Parse topic filters
	topicFilter := make([][]types.Hash, len(criteria.Topics))
	for i, topicList := range criteria.Topics {
		for _, topicHex := range topicList {
			topicFilter[i] = append(topicFilter[i], types.HexToHash(topicHex))
		}
	}

	// EIP-4444: check if the requested range includes pruned blocks.
	if api.historyPruned(fromBlock) {
		return errorResponse(req.ID, ErrCodeHistoryPruned,
			"historical logs pruned (EIP-4444)")
	}

	for blockNum := fromBlock; blockNum <= toBlock; blockNum++ {
		header := api.backend.HeaderByNumber(BlockNumber(blockNum))
		if header == nil {
			continue
		}
		blockHash := header.Hash()
		logs := api.backend.GetLogs(blockHash)
		for _, log := range logs {
			if matchLog(log, addrFilter, topicFilter) {
				result = append(result, FormatLog(log))
			}
		}
	}

	if result == nil {
		result = []*RPCLog{}
	}
	return successResponse(req.ID, result)
}

// blockNumberOrHashParam mirrors the go-ethereum ethclient BlockNumberOrHash
// JSON encoding: {"blockHash":"0x...","requireCanonical":false} or
// {"blockNumber":"0x..."}.
type blockNumberOrHashParam struct {
	BlockHash   *string `json:"blockHash"`
	BlockNumber *string `json:"blockNumber"`
}

// getBlockReceipts returns all receipts for a given block number.
func (api *EthAPI) getBlockReceipts(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing block number or hash")
	}

	// Detect block hash vs block number: a hash is 0x-prefixed and 66 chars.
	var (
		header    *types.Header
		blockHash types.Hash
		byHash    bool
	)

	// Try object form {"blockHash":"0x..."} / {"blockNumber":"0x..."} first.
	raw := req.Params[0]
	if len(raw) > 0 && raw[0] == '{' {
		var obj blockNumberOrHashParam
		if err := json.Unmarshal(raw, &obj); err != nil {
			return errorResponse(req.ID, ErrCodeInvalidParams, "invalid block param: "+err.Error())
		}
		if obj.BlockHash != nil {
			blockHash = types.HexToHash(*obj.BlockHash)
			header = api.backend.HeaderByHash(blockHash)
			byHash = true
		} else if obj.BlockNumber != nil {
			var bn BlockNumber
			if err := json.Unmarshal([]byte(`"`+*obj.BlockNumber+`"`), &bn); err != nil {
				return errorResponse(req.ID, ErrCodeInvalidParams, "invalid blockNumber: "+err.Error())
			}
			header = api.backend.HeaderByNumber(bn)
		} else {
			return errorResponse(req.ID, ErrCodeInvalidParams, "object must have blockHash or blockNumber")
		}
	} else {
		var paramStr string
		if err := json.Unmarshal(raw, &paramStr); err != nil {
			return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
		}
		if len(paramStr) == 66 && (paramStr[:2] == "0x" || paramStr[:2] == "0X") {
			blockHash = types.HexToHash(paramStr)
			header = api.backend.HeaderByHash(blockHash)
			byHash = true
		} else {
			var bn BlockNumber
			if err := json.Unmarshal(raw, &bn); err != nil {
				return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
			}
			header = api.backend.HeaderByNumber(bn)
		}
	}

	if header == nil {
		return successResponse(req.ID, nil)
	}

	blockNum := header.Number.Uint64()
	if !byHash {
		blockHash = header.Hash()
	}

	// EIP-4444: check if receipts have been pruned.
	if api.historyPruned(blockNum) {
		return errorResponse(req.ID, ErrCodeHistoryPruned,
			"historical receipts pruned (EIP-4444)")
	}

	var receipts []*types.Receipt
	if byHash {
		receipts = api.backend.GetReceipts(blockHash)
	} else {
		receipts = api.backend.GetBlockReceipts(blockNum)
	}

	if receipts == nil {
		return successResponse(req.ID, []*RPCReceipt{})
	}

	// Fetch the block to populate from/to in each receipt.
	block := api.backend.BlockByHash(blockHash)
	var txs []*types.Transaction
	if block != nil {
		txs = block.Transactions()
	}

	result := make([]*RPCReceipt, len(receipts))
	for i, receipt := range receipts {
		var tx *types.Transaction
		if i < len(txs) {
			tx = txs[i]
		}
		result[i] = FormatReceipt(receipt, tx)
	}

	return successResponse(req.ID, result)
}

// ethCall executes a read-only EVM call without creating a transaction.
func (api *EthAPI) ethCall(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing call arguments")
	}

	var args CallArgs
	if err := json.Unmarshal(req.Params[0], &args); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	bn := LatestBlockNumber
	if len(req.Params) > 1 {
		if err := json.Unmarshal(req.Params[1], &bn); err != nil {
			return errorResponse(req.ID, ErrCodeInvalidParams, "invalid block number: "+err.Error())
		}
	}

	from, to, gas, value, data := parseCallArgs(&args)

	result, _, err := api.backend.EVMCall(from, to, data, gas, value, bn)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, "execution error: "+err.Error())
	}

	return successResponse(req.ID, encodeBytes(result))
}

// estimateGas estimates the gas needed to execute a transaction.
func (api *EthAPI) estimateGas(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing call arguments")
	}

	var args CallArgs
	if err := json.Unmarshal(req.Params[0], &args); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	bn := LatestBlockNumber
	if len(req.Params) > 1 {
		if err := json.Unmarshal(req.Params[1], &bn); err != nil {
			return errorResponse(req.ID, ErrCodeInvalidParams, "invalid block number: "+err.Error())
		}
	}

	from, to, _, value, data := parseCallArgs(&args)

	// Get block gas limit as upper bound
	header := api.backend.HeaderByNumber(bn)
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}

	hi := header.GasLimit
	// Intrinsic gas as lower bound (21000 base)
	lo := uint64(21000)

	// If user specified gas, use it as upper bound
	if args.Gas != nil {
		userGas := parseHexUint64(*args.Gas)
		if userGas > 0 && userGas < hi {
			hi = userGas
		}
	}

	// Check that the upper bound works
	_, _, err := api.backend.EVMCall(from, to, data, hi, value, bn)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, "execution error: "+err.Error())
	}

	// Check if the lower bound itself works.
	_, _, errLo := api.backend.EVMCall(from, to, data, lo, value, bn)
	if errLo == nil {
		return successResponse(req.ID, encodeUint64(lo))
	}

	// Binary search for minimum gas needed
	for lo+1 < hi {
		mid := (lo + hi) / 2
		_, _, err := api.backend.EVMCall(from, to, data, mid, value, bn)
		if err != nil {
			lo = mid
		} else {
			hi = mid
		}
	}

	return successResponse(req.ID, encodeUint64(hi))
}

// getTransactionByHash returns transaction info by hash.
func (api *EthAPI) getTransactionByHash(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing transaction hash")
	}

	var hashHex string
	if err := json.Unmarshal(req.Params[0], &hashHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	hash := types.HexToHash(hashHex)
	tx, blockNum, index := api.backend.GetTransaction(hash)
	if tx == nil {
		return successResponse(req.ID, nil)
	}

	var blockHash *types.Hash
	if blockNum > 0 {
		header := api.backend.HeaderByNumber(BlockNumber(blockNum))
		if header != nil {
			h := header.Hash()
			blockHash = &h
		}
	}

	return successResponse(req.ID, FormatTransaction(tx, blockHash, &blockNum, &index))
}

// getTransactionReceipt returns a receipt for a transaction hash.
func (api *EthAPI) getTransactionReceipt(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing transaction hash")
	}

	var hashHex string
	if err := json.Unmarshal(req.Params[0], &hashHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	txHash := types.HexToHash(hashHex)
	tx, blockNum, _ := api.backend.GetTransaction(txHash)
	if tx == nil {
		return successResponse(req.ID, nil)
	}

	// EIP-4444: check if receipt has been pruned.
	if api.historyPruned(blockNum) {
		return errorResponse(req.ID, ErrCodeHistoryPruned,
			"historical receipt pruned (EIP-4444)")
	}

	// Get the block header for block hash
	header := api.backend.HeaderByNumber(BlockNumber(blockNum))
	if header == nil {
		return successResponse(req.ID, nil)
	}

	blockHash := header.Hash()
	receipts := api.backend.GetReceipts(blockHash)

	// Find the receipt matching our tx hash
	for _, receipt := range receipts {
		if receipt.TxHash == txHash {
			return successResponse(req.ID, FormatReceipt(receipt, tx))
		}
	}

	return successResponse(req.ID, nil)
}

// sendRawTransaction decodes an RLP-encoded transaction and submits it.
func (api *EthAPI) sendRawTransaction(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing raw transaction data")
	}

	var dataHex string
	if err := json.Unmarshal(req.Params[0], &dataHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	rawBytes := fromHexBytes(dataHex)
	if len(rawBytes) == 0 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "empty transaction data")
	}

	tx, err := types.DecodeTxRLP(rawBytes)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	if err := api.backend.SendTransaction(tx); err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}

	return successResponse(req.ID, encodeHash(tx.Hash()))
}

// getHeaderByNumber implements eth_getHeaderByNumber.
func (api *EthAPI) getHeaderByNumber(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing block number")
	}

	var bn BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	header := api.backend.HeaderByNumber(bn)
	if header == nil {
		return successResponse(req.ID, nil)
	}
	return successResponse(req.ID, FormatHeader(header))
}

// getHeaderByHash implements eth_getHeaderByHash.
func (api *EthAPI) getHeaderByHash(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing block hash")
	}

	var hashHex string
	if err := json.Unmarshal(req.Params[0], &hashHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	hash := types.HexToHash(hashHex)
	header := api.backend.HeaderByHash(hash)
	if header == nil {
		return successResponse(req.ID, nil)
	}
	return successResponse(req.ID, FormatHeader(header))
}

// getTransactionByBlockHashAndIndex implements eth_getTransactionByBlockHashAndIndex.
func (api *EthAPI) getTransactionByBlockHashAndIndex(req *Request) *Response {
	if len(req.Params) < 2 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing block hash or index")
	}

	var hashHex, indexHex string
	if err := json.Unmarshal(req.Params[0], &hashHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}
	if err := json.Unmarshal(req.Params[1], &indexHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	blockHash := types.HexToHash(hashHex)
	index := parseHexUint64(indexHex)

	block := api.backend.BlockByHash(blockHash)
	if block == nil {
		return successResponse(req.ID, nil)
	}

	txs := block.Transactions()
	if index >= uint64(len(txs)) {
		return successResponse(req.ID, nil)
	}

	blockNum := block.NumberU64()
	bh := block.Hash()
	return successResponse(req.ID, FormatTransaction(txs[index], &bh, &blockNum, &index))
}

// getTransactionByBlockNumberAndIndex implements eth_getTransactionByBlockNumberAndIndex.
func (api *EthAPI) getTransactionByBlockNumberAndIndex(req *Request) *Response {
	if len(req.Params) < 2 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing block number or index")
	}

	var bn BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	var indexHex string
	if err := json.Unmarshal(req.Params[1], &indexHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}
	index := parseHexUint64(indexHex)

	block := api.backend.BlockByNumber(bn)
	if block == nil {
		return successResponse(req.ID, nil)
	}

	txs := block.Transactions()
	if index >= uint64(len(txs)) {
		return successResponse(req.ID, nil)
	}

	blockNum := block.NumberU64()
	bh := block.Hash()
	return successResponse(req.ID, FormatTransaction(txs[index], &bh, &blockNum, &index))
}

// getBlockTransactionCountByHash implements eth_getBlockTransactionCountByHash.
func (api *EthAPI) getBlockTransactionCountByHash(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing block hash")
	}

	var hashHex string
	if err := json.Unmarshal(req.Params[0], &hashHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	block := api.backend.BlockByHash(types.HexToHash(hashHex))
	if block == nil {
		return successResponse(req.ID, nil)
	}

	return successResponse(req.ID, encodeUint64(uint64(len(block.Transactions()))))
}

// getBlockTransactionCountByNumber implements eth_getBlockTransactionCountByNumber.
func (api *EthAPI) getBlockTransactionCountByNumber(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing block number")
	}

	var bn BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	block := api.backend.BlockByNumber(bn)
	if block == nil {
		return successResponse(req.ID, nil)
	}

	return successResponse(req.ID, encodeUint64(uint64(len(block.Transactions()))))
}

// accounts implements eth_accounts (returns empty list for non-wallet nodes).
func (api *EthAPI) accounts(req *Request) *Response {
	return successResponse(req.ID, []string{})
}

// coinbase implements eth_coinbase.
func (api *EthAPI) coinbase(req *Request) *Response {
	header := api.backend.CurrentHeader()
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "no current block")
	}
	return successResponse(req.ID, encodeAddress(header.Coinbase))
}

// mining implements eth_mining (always false for PoS).
func (api *EthAPI) mining(req *Request) *Response {
	return successResponse(req.ID, false)
}

// hashrate implements eth_hashrate (always 0 for PoS).
func (api *EthAPI) hashrate(req *Request) *Response {
	return successResponse(req.ID, "0x0")
}

// protocolVersion implements eth_protocolVersion.
func (api *EthAPI) protocolVersion(req *Request) *Response {
	return successResponse(req.ID, fmt.Sprintf("0x%x", 68)) // ETH/68
}

// getUncleCountByBlockHash implements eth_getUncleCountByBlockHash.
// Post-merge: always 0.
func (api *EthAPI) getUncleCountByBlockHash(req *Request) *Response {
	return successResponse(req.ID, "0x0")
}

// getUncleCountByBlockNumber implements eth_getUncleCountByBlockNumber.
// Post-merge: always 0.
func (api *EthAPI) getUncleCountByBlockNumber(req *Request) *Response {
	return successResponse(req.ID, "0x0")
}

// getUncleByBlockHashAndIndex implements eth_getUncleByBlockHashAndIndex.
// Post-merge: always returns null (no uncles in PoS).
func (api *EthAPI) getUncleByBlockHashAndIndex(req *Request) *Response {
	return successResponse(req.ID, nil)
}

// getUncleByBlockNumberAndIndex implements eth_getUncleByBlockNumberAndIndex.
// Post-merge: always returns null (no uncles in PoS).
func (api *EthAPI) getUncleByBlockNumberAndIndex(req *Request) *Response {
	return successResponse(req.ID, nil)
}

// getBlobBaseFee implements eth_blobBaseFee (EIP-7516).
func (api *EthAPI) getBlobBaseFee(req *Request) *Response {
	header := api.backend.CurrentHeader()
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "no current block")
	}
	if header.ExcessBlobGas != nil {
		return successResponse(req.ID, encodeBigInt(new(big.Int).SetUint64(*header.ExcessBlobGas)))
	}
	return successResponse(req.ID, "0x0")
}

// UnsubscribeID removes a subscription by ID. Used by WebSocket connection
// cleanup when the connection is closed.
func (api *EthAPI) UnsubscribeID(id string) bool {
	return api.subs.Unsubscribe(id)
}

// Ensure rpcfilter is used (MatchFilter needed for bloomMatchesQuery wrapper).
var _ = rpcfilter.MatchFilter

// CriteriaToQuery is the exported version of criteriaToQuery.
func CriteriaToQuery(c FilterCriteria, backend Backend) FilterQuery {
	return criteriaToQuery(c, backend)
}

// MatchLog is the exported version of matchLog.
func MatchLog(log *types.Log, addrFilter map[types.Address]bool, topicFilter [][]types.Hash) bool {
	return matchLog(log, addrFilter, topicFilter)
}
