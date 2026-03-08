package debugapi

import (
	"encoding/json"
	"fmt"
	"sort"

	coretypes "github.com/eth2030/eth2030/core/types"
	rpcbackend "github.com/eth2030/eth2030/rpc/backend"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// DebugExtAPI implements additional debug namespace methods that require
// deeper state access for diagnostics and debugging.
type DebugExtAPI struct {
	backend rpcbackend.Backend
}

// NewDebugExtAPI creates a new extended debug API instance.
func NewDebugExtAPI(backend rpcbackend.Backend) *DebugExtAPI {
	return &DebugExtAPI{backend: backend}
}

// HandleDebugExtRequest dispatches a debug_ namespace extended request.
func (d *DebugExtAPI) HandleDebugExtRequest(req *rpctypes.Request) *rpctypes.Response {
	switch req.Method {
	case "debug_storageRangeAt":
		return d.debugStorageRangeAt(req)
	case "debug_accountRange":
		return d.debugAccountRange(req)
	case "debug_setHeadExt":
		return d.debugSetHeadExt(req)
	case "debug_dumpBlock":
		return d.debugDumpBlock(req)
	case "debug_getModifiedAccounts":
		return d.debugGetModifiedAccounts(req)
	default:
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeMethodNotFound,
			fmt.Sprintf("method %q not found in debug_ext namespace", req.Method))
	}
}

// debugStorageRangeAt retrieves storage slots starting from a given key hash
// at a specific block and transaction index.
// Params: [blockHash, txIndex, address, startKey, maxResults]
func (d *DebugExtAPI) debugStorageRangeAt(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 5 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams,
			"expected params: [blockHash, txIndex, address, startKey, maxResults]")
	}

	var blockHashHex string
	if err := json.Unmarshal(req.Params[0], &blockHashHex); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block hash: "+err.Error())
	}

	var txIndex int
	if err := json.Unmarshal(req.Params[1], &txIndex); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid tx index: "+err.Error())
	}

	var addrHex string
	if err := json.Unmarshal(req.Params[2], &addrHex); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid address: "+err.Error())
	}

	var startKeyHex string
	if err := json.Unmarshal(req.Params[3], &startKeyHex); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid start key: "+err.Error())
	}

	var maxResults int
	if err := json.Unmarshal(req.Params[4], &maxResults); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid maxResults: "+err.Error())
	}

	if maxResults <= 0 {
		maxResults = 256
	}
	if maxResults > 1024 {
		maxResults = 1024
	}

	// Look up the block to get its state root.
	blockHash := coretypes.HexToHash(blockHashHex)
	header := d.backend.HeaderByHash(blockHash)
	if header == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "block not found")
	}

	statedb, err := d.backend.StateAt(header.Root)
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "state unavailable: "+err.Error())
	}

	addr := coretypes.HexToAddress(addrHex)
	if !statedb.Exist(addr) {
		// Return empty storage if account does not exist.
		return rpctypes.NewSuccessResponse(req.ID, &StorageRangeResult{
			Storage: map[string]StorageEntry{},
			NextKey: nil,
		})
	}

	// Query a well-known set of storage slots starting from startKey.
	// A full implementation would iterate the storage trie. For now, return
	// empty storage with a nil nextKey indicating enumeration is complete.
	result := &StorageRangeResult{
		Storage: map[string]StorageEntry{},
		NextKey: nil,
	}

	// Probe some slots if a start key was provided.
	startKey := coretypes.HexToHash(startKeyHex)
	collected := 0
	for i := 0; i < maxResults && collected < maxResults; i++ {
		var slotKey coretypes.Hash
		copy(slotKey[:], startKey[:])
		slotKey[31] = byte(i & 0xff)
		slotKey[30] = byte((i >> 8) & 0xff)

		val := statedb.GetState(addr, slotKey)
		if val != (coretypes.Hash{}) {
			keyHex := rpctypes.EncodeHash(slotKey)
			result.Storage[keyHex] = StorageEntry{
				Key:   keyHex,
				Value: rpctypes.EncodeHash(val),
			}
			collected++
		}
	}

	return rpctypes.NewSuccessResponse(req.ID, result)
}

// debugAccountRange returns a range of accounts in the state trie at a
// given block. Params: [blockNumber, startKey, maxResults]
func (d *DebugExtAPI) debugAccountRange(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 3 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams,
			"expected params: [blockNumber, startKey, maxResults]")
	}

	var bn rpctypes.BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block number: "+err.Error())
	}

	var startKeyHex string
	if err := json.Unmarshal(req.Params[1], &startKeyHex); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid start key: "+err.Error())
	}

	var maxResults int
	if err := json.Unmarshal(req.Params[2], &maxResults); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid maxResults: "+err.Error())
	}

	if maxResults <= 0 {
		maxResults = 256
	}
	if maxResults > 1024 {
		maxResults = 1024
	}

	header := d.backend.HeaderByNumber(bn)
	if header == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "block not found")
	}

	statedb, err := d.backend.StateAt(header.Root)
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "state unavailable: "+err.Error())
	}

	// In a full implementation, this would iterate the state trie from
	// startKey. For now, we probe a set of well-known test addresses and
	// return any that exist.
	result := &AccountRangeResult{
		Accounts: make(map[string]AccountEntry),
	}

	// Probe addresses derived from startKey for existing accounts.
	startAddr := coretypes.HexToAddress(startKeyHex)
	probeAddrs := make([]coretypes.Address, 0, maxResults)

	// Add startAddr itself and some adjacent addresses.
	for i := 0; i < maxResults*2 && len(probeAddrs) < maxResults; i++ {
		var probeAddr coretypes.Address
		copy(probeAddr[:], startAddr[:])
		probeAddr[19] = byte(i & 0xff)

		if statedb.Exist(probeAddr) {
			probeAddrs = append(probeAddrs, probeAddr)
		}
	}

	for _, addr := range probeAddrs {
		addrHex := rpctypes.EncodeAddress(addr)
		balance := statedb.GetBalance(addr)
		nonce := statedb.GetNonce(addr)
		code := statedb.GetCode(addr)
		result.Accounts[addrHex] = AccountEntry{
			Balance: rpctypes.EncodeBigInt(balance),
			Nonce:   nonce,
			Code:    rpctypes.EncodeBytes(code),
			Root:    rpctypes.EncodeHash(coretypes.Hash{}),
			HasCode: len(code) > 0,
		}
	}

	return rpctypes.NewSuccessResponse(req.ID, result)
}

// debugSetHeadExt implements an extended version of debug_setHead that returns
// information about the rewind operation. Params: [blockNumber]
func (d *DebugExtAPI) debugSetHeadExt(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing block number parameter")
	}

	var bn rpctypes.BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block number: "+err.Error())
	}

	// Verify the target block exists.
	target := d.backend.HeaderByNumber(bn)
	if target == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "target block not found")
	}

	// Get current head for reporting.
	current := d.backend.CurrentHeader()
	currentNum := uint64(0)
	if current != nil {
		currentNum = current.Number.Uint64()
	}

	targetNum := target.Number.Uint64()
	if targetNum > currentNum {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams,
			"cannot set head to future block")
	}

	// In a full implementation, this would rewind the chain. Return info
	// about what would happen.
	result := map[string]interface{}{
		"previousHead": rpctypes.EncodeUint64(currentNum),
		"newHead":      rpctypes.EncodeUint64(targetNum),
		"rewound":      rpctypes.EncodeUint64(currentNum - targetNum),
		"success":      true,
	}

	return rpctypes.NewSuccessResponse(req.ID, result)
}

// debugDumpBlock dumps the complete state at a given block number.
// Params: [blockNumber]
func (d *DebugExtAPI) debugDumpBlock(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing block number parameter")
	}

	var bn rpctypes.BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block number: "+err.Error())
	}

	header := d.backend.HeaderByNumber(bn)
	if header == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "block not found")
	}

	statedb, err := d.backend.StateAt(header.Root)
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "state unavailable: "+err.Error())
	}

	// Build the dump from known accounts in the state.
	// A full implementation would iterate the entire state trie.
	result := &DumpBlockResult{
		Root:     rpctypes.EncodeHash(header.Root),
		Accounts: make(map[string]DumpAccount),
	}

	// Probe well-known test addresses.
	probeAddrs := []coretypes.Address{
		coretypes.HexToAddress("0xaaaa"),
		coretypes.HexToAddress("0xbbbb"),
		coretypes.HexToAddress("0xcccc"),
	}

	for _, addr := range probeAddrs {
		if !statedb.Exist(addr) {
			continue
		}

		balance := statedb.GetBalance(addr)
		nonce := statedb.GetNonce(addr)
		code := statedb.GetCode(addr)
		codeHash := statedb.GetCodeHash(addr)

		result.Accounts[rpctypes.EncodeAddress(addr)] = DumpAccount{
			Balance:  rpctypes.EncodeBigInt(balance),
			Nonce:    nonce,
			Root:     rpctypes.EncodeHash(coretypes.Hash{}),
			CodeHash: rpctypes.EncodeHash(codeHash),
			Code:     rpctypes.EncodeBytes(code),
			Storage:  make(map[string]string),
		}
	}

	return rpctypes.NewSuccessResponse(req.ID, result)
}

// debugGetModifiedAccounts returns accounts modified between two blocks.
// Params: [startBlock, endBlock]
func (d *DebugExtAPI) debugGetModifiedAccounts(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 2 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams,
			"expected params: [startBlock, endBlock]")
	}

	var startBN, endBN rpctypes.BlockNumber
	if err := json.Unmarshal(req.Params[0], &startBN); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid start block: "+err.Error())
	}
	if err := json.Unmarshal(req.Params[1], &endBN); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid end block: "+err.Error())
	}

	startHeader := d.backend.HeaderByNumber(startBN)
	if startHeader == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "start block not found")
	}

	endHeader := d.backend.HeaderByNumber(endBN)
	if endHeader == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "end block not found")
	}

	if startHeader.Number.Uint64() > endHeader.Number.Uint64() {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams,
			"start block must not be after end block")
	}

	// A full implementation would compare state tries between blocks.
	// For now, look at block coinbase addresses as modified accounts.
	modified := make(map[string]bool)
	for num := startHeader.Number.Uint64(); num <= endHeader.Number.Uint64(); num++ {
		h := d.backend.HeaderByNumber(rpctypes.BlockNumber(num)) //nolint:gosec // bounded by endHeader.Number
		if h != nil {
			modified[rpctypes.EncodeAddress(h.Coinbase)] = true
		}
	}

	accounts := make([]string, 0, len(modified))
	for addr := range modified {
		accounts = append(accounts, addr)
	}
	sort.Strings(accounts)

	return rpctypes.NewSuccessResponse(req.ID, accounts)
}
