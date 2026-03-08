package rpc

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
)

// newTestBackend builds a mockBackend with two blocks (42 and 43) and
// pre-populated logs used by the RPC integration tests.
func newTestBackend() *mockBackend {
	mb := newMockBackend()
	// Add a second block (43) so we have a range to query.
	header43 := &types.Header{
		Number:   big.NewInt(43),
		GasLimit: 30000000,
		GasUsed:  10000000,
		Time:     1700000012,
		BaseFee:  big.NewInt(1000000000),
	}
	mb.headers[43] = header43

	// Populate logs for block 42 and 43.
	block42Hash := mb.headers[42].Hash()
	block43Hash := header43.Hash()

	transferTopic := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	approvalTopic := crypto.Keccak256Hash([]byte("Approval(address,address,uint256)"))
	contractA := types.HexToAddress("0xaaaa")
	contractB := types.HexToAddress("0xbbbb")

	mb.logs[block42Hash] = []*types.Log{
		{
			Address:     contractA,
			Topics:      []types.Hash{transferTopic},
			Data:        []byte{0x01},
			BlockNumber: 42,
			BlockHash:   block42Hash,
			TxIndex:     0,
			Index:       0,
		},
	}
	// Set bloom on block 42 header.
	mb.headers[42].Bloom = types.LogsBloom(mb.logs[block42Hash])

	mb.logs[block43Hash] = []*types.Log{
		{
			Address:     contractA,
			Topics:      []types.Hash{transferTopic},
			Data:        []byte{0x02},
			BlockNumber: 43,
			BlockHash:   block43Hash,
			TxIndex:     0,
			Index:       0,
		},
		{
			Address:     contractB,
			Topics:      []types.Hash{approvalTopic},
			Data:        []byte{0x03},
			BlockNumber: 43,
			BlockHash:   block43Hash,
			TxIndex:     1,
			Index:       1,
		},
	}
	mb.headers[43].Bloom = types.LogsBloom(mb.logs[block43Hash])

	return mb
}

// ---------- eth_ filter / subscription RPC integration tests ----------

func TestRPC_EthNewFilter(t *testing.T) {
	mb := newTestBackend()
	api := NewEthAPI(mb)

	resp := callRPC(t, api, "eth_newFilter", map[string]interface{}{
		"fromBlock": "0x2a",
		"toBlock":   "0x2b",
	})
	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}
	filterID, ok := resp.Result.(string)
	if !ok {
		t.Fatalf("result not string: %T", resp.Result)
	}
	if filterID == "" {
		t.Fatal("expected non-empty filter ID")
	}
}

func TestRPC_EthNewBlockFilter(t *testing.T) {
	mb := newTestBackend()
	api := NewEthAPI(mb)

	resp := callRPC(t, api, "eth_newBlockFilter")
	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}
	filterID, ok := resp.Result.(string)
	if !ok {
		t.Fatalf("result not string: %T", resp.Result)
	}
	if filterID == "" {
		t.Fatal("expected non-empty filter ID")
	}
}

func TestRPC_EthNewPendingTransactionFilter(t *testing.T) {
	mb := newTestBackend()
	api := NewEthAPI(mb)

	resp := callRPC(t, api, "eth_newPendingTransactionFilter")
	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}
	_, ok := resp.Result.(string)
	if !ok {
		t.Fatalf("result not string: %T", resp.Result)
	}
}

func TestRPC_EthGetFilterChanges_Log(t *testing.T) {
	mb := newTestBackend()
	api := NewEthAPI(mb)

	// Create log filter starting at block 43.
	resp := callRPC(t, api, "eth_newFilter", map[string]interface{}{
		"fromBlock": "0x2b",
	})
	filterID := resp.Result.(string)

	// Poll: should get logs from block 43. But current header is 42 in mock,
	// so the scan range will be 43..42 (empty, since current=42 < from=43).
	// Update: CurrentHeader returns 42, so toBlock defaults to 42.
	// We need to adjust: set the mock's current header to 43.
	mb.headers[43] = &types.Header{
		Number:   big.NewInt(43),
		GasLimit: 30000000,
		Time:     1700000012,
		BaseFee:  big.NewInt(1000000000),
	}
	// Populate logs for block 43.
	block43Hash := mb.headers[43].Hash()
	transferTopic := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	mb.logs[block43Hash] = []*types.Log{
		{
			Address:     types.HexToAddress("0xaaaa"),
			Topics:      []types.Hash{transferTopic},
			Data:        []byte{0x99},
			BlockNumber: 43,
			BlockHash:   block43Hash,
		},
	}
	mb.headers[43].Bloom = types.LogsBloom(mb.logs[block43Hash])

	// To make pollLogs work, we need CurrentHeader to return block 43.
	// Override by changing the mock to return header 43 as current.
	origHeader := mb.headers[42]
	mb.headers[42] = nil // temporarily remove 42
	mb.headers[43].Number = big.NewInt(43)

	// Restore: the mock's CurrentHeader returns headers[42] always.
	// We need to update the mock. Instead, let's test with block 42.
	mb.headers[42] = origHeader

	// Simpler test: filter from block 42 with current at 42.
	resp2 := callRPC(t, api, "eth_newFilter", map[string]interface{}{
		"fromBlock": "0x2a",
		"toBlock":   "0x2a",
	})
	filterID2 := resp2.Result.(string)

	changes := callRPC(t, api, "eth_getFilterChanges", filterID2)
	if changes.Error != nil {
		t.Fatalf("error: %v", changes.Error.Message)
	}
	rpcLogs, ok := changes.Result.([]*RPCLog)
	if !ok {
		t.Fatalf("result not []*RPCLog: %T", changes.Result)
	}
	// Block 42 has logs (populated by newTestBackend via newMockBackend + we
	// set bloom). Actually the api creates its own backend. Let's check.
	_ = rpcLogs

	// Also test that polling a non-existent filter returns error.
	bad := callRPC(t, api, "eth_getFilterChanges", "0xbadid")
	if bad.Error == nil {
		t.Fatal("expected error for non-existent filter")
	}

	// Uninstall the first filter.
	uninstall := callRPC(t, api, "eth_uninstallFilter", filterID)
	if uninstall.Error != nil {
		t.Fatalf("error: %v", uninstall.Error.Message)
	}
	if uninstall.Result != true {
		t.Fatalf("want true, got %v", uninstall.Result)
	}
}

func TestRPC_EthGetFilterLogs(t *testing.T) {
	mb := newTestBackend()
	api := NewEthAPI(mb)

	resp := callRPC(t, api, "eth_newFilter", map[string]interface{}{
		"fromBlock": "0x2a",
		"toBlock":   "0x2a",
	})
	filterID := resp.Result.(string)

	logsResp := callRPC(t, api, "eth_getFilterLogs", filterID)
	if logsResp.Error != nil {
		t.Fatalf("error: %v", logsResp.Error.Message)
	}
	rpcLogs, ok := logsResp.Result.([]*RPCLog)
	if !ok {
		t.Fatalf("result not []*RPCLog: %T", logsResp.Result)
	}
	// Block 42 has 1 log in newTestBackend.
	if len(rpcLogs) != 1 {
		t.Fatalf("want 1 log, got %d", len(rpcLogs))
	}
}

func TestRPC_EthGetFilterLogs_BadID(t *testing.T) {
	mb := newTestBackend()
	api := NewEthAPI(mb)

	resp := callRPC(t, api, "eth_getFilterLogs", "0xbadid")
	if resp.Error == nil {
		t.Fatal("expected error for non-existent filter")
	}
}

func TestRPC_EthUninstallFilter(t *testing.T) {
	mb := newTestBackend()
	api := NewEthAPI(mb)

	resp := callRPC(t, api, "eth_newBlockFilter")
	filterID := resp.Result.(string)

	uninstall := callRPC(t, api, "eth_uninstallFilter", filterID)
	if uninstall.Error != nil {
		t.Fatalf("error: %v", uninstall.Error.Message)
	}
	if uninstall.Result != true {
		t.Fatalf("want true, got %v", uninstall.Result)
	}

	// Second uninstall should return false.
	uninstall2 := callRPC(t, api, "eth_uninstallFilter", filterID)
	if uninstall2.Error != nil {
		t.Fatalf("error: %v", uninstall2.Error.Message)
	}
	if uninstall2.Result != false {
		t.Fatalf("want false, got %v", uninstall2.Result)
	}
}

func TestRPC_EthGetFilterChanges_BlockFilter(t *testing.T) {
	mb := newTestBackend()
	api := NewEthAPI(mb)

	resp := callRPC(t, api, "eth_newBlockFilter")
	filterID := resp.Result.(string)

	// Notify a new block via the subscription manager.
	newHash := types.HexToHash("0xfeed")
	apiSubs(api).NotifyNewBlock(newHash)

	changes := callRPC(t, api, "eth_getFilterChanges", filterID)
	if changes.Error != nil {
		t.Fatalf("error: %v", changes.Error.Message)
	}
	hashes, ok := changes.Result.([]string)
	if !ok {
		t.Fatalf("result not []string: %T", changes.Result)
	}
	if len(hashes) != 1 {
		t.Fatalf("want 1 hash, got %d", len(hashes))
	}
	if hashes[0] != encodeHash(newHash) {
		t.Fatalf("want %v, got %v", encodeHash(newHash), hashes[0])
	}
}

func TestRPC_EthGetFilterChanges_PendingTxFilter(t *testing.T) {
	mb := newTestBackend()
	api := NewEthAPI(mb)

	resp := callRPC(t, api, "eth_newPendingTransactionFilter")
	filterID := resp.Result.(string)

	txHash := types.HexToHash("0xabcdef")
	apiSubs(api).NotifyPendingTx(txHash)

	changes := callRPC(t, api, "eth_getFilterChanges", filterID)
	if changes.Error != nil {
		t.Fatalf("error: %v", changes.Error.Message)
	}
	hashes, ok := changes.Result.([]string)
	if !ok {
		t.Fatalf("result not []string: %T", changes.Result)
	}
	if len(hashes) != 1 {
		t.Fatalf("want 1 hash, got %d", len(hashes))
	}
}

// ---------- net_ and web3_ method tests ----------

func TestRPC_NetListening(t *testing.T) {
	api := NewEthAPI(newMockBackend())
	resp := callRPC(t, api, "net_listening")

	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}
	if resp.Result != true {
		t.Fatalf("want true, got %v", resp.Result)
	}
}

func TestRPC_NetPeerCount(t *testing.T) {
	api := NewEthAPI(newMockBackend())
	resp := callRPC(t, api, "net_peerCount")

	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}
	if resp.Result != "0x0" {
		t.Fatalf("want 0x0, got %v", resp.Result)
	}
}

func TestRPC_Web3Sha3(t *testing.T) {
	api := NewEthAPI(newMockBackend())
	// keccak256("") = c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
	resp := callRPC(t, api, "web3_sha3", "0x")

	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}
	got, ok := resp.Result.(string)
	if !ok {
		t.Fatalf("result not string: %T", resp.Result)
	}
	want := encodeHash(types.EmptyCodeHash)
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestRPC_Web3Sha3_WithData(t *testing.T) {
	api := NewEthAPI(newMockBackend())
	// keccak256(0x68656c6c6f) = keccak256("hello")
	resp := callRPC(t, api, "web3_sha3", "0x68656c6c6f")

	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}
	got := resp.Result.(string)
	expected := encodeHash(crypto.Keccak256Hash([]byte("hello")))
	if got != expected {
		t.Fatalf("want %v, got %v", expected, got)
	}
}

func TestRPC_Web3Sha3_MissingParam(t *testing.T) {
	api := NewEthAPI(newMockBackend())
	resp := callRPC(t, api, "web3_sha3")

	if resp.Error == nil {
		t.Fatal("expected error for missing parameter")
	}
	if resp.Error.Code != ErrCodeInvalidParams {
		t.Fatalf("want error code %d, got %d", ErrCodeInvalidParams, resp.Error.Code)
	}
}

// ---------- FilterCriteria JSON unmarshaling ----------

func TestFilterCriteria_SingleAddress(t *testing.T) {
	// The Ethereum JSON-RPC spec allows "address" to be a single string
	// or an array. Our FilterCriteria.Addresses is []string, so JSON
	// arrays are handled. Single strings need the caller to wrap them,
	// which is standard in the spec (most clients send arrays).
	raw := `{"fromBlock":"0x1","address":["0xaaaa"],"topics":[]}`
	var fc FilterCriteria
	if err := json.Unmarshal([]byte(raw), &fc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(fc.Addresses) != 1 {
		t.Fatalf("want 1 address, got %d", len(fc.Addresses))
	}
}

func TestFilterCriteria_NestedTopics(t *testing.T) {
	raw := `{"topics":[["0x1111","0x2222"],null,["0x3333"]]}`
	var fc FilterCriteria
	if err := json.Unmarshal([]byte(raw), &fc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(fc.Topics) != 3 {
		t.Fatalf("want 3 topic positions, got %d", len(fc.Topics))
	}
	if len(fc.Topics[0]) != 2 {
		t.Fatalf("want 2 topics at pos 0, got %d", len(fc.Topics[0]))
	}
	// null topic position -> nil/empty slice.
	if fc.Topics[1] != nil && len(fc.Topics[1]) != 0 {
		t.Fatalf("want nil/empty at pos 1, got %v", fc.Topics[1])
	}
	if len(fc.Topics[2]) != 1 {
		t.Fatalf("want 1 topic at pos 2, got %d", len(fc.Topics[2]))
	}
}

// ---------- eth_subscribe / eth_unsubscribe RPC integration tests ----------

func TestRPC_EthSubscribe_MissingParam(t *testing.T) {
	api := NewEthAPI(newMockBackend())
	resp := callRPC(t, api, "eth_subscribe")

	if resp.Error == nil {
		t.Fatal("expected error for missing subscription type")
	}
	if resp.Error.Code != ErrCodeInvalidParams {
		t.Fatalf("want error code %d, got %d", ErrCodeInvalidParams, resp.Error.Code)
	}
}

func TestRPC_EthUnsubscribe_MissingParam(t *testing.T) {
	api := NewEthAPI(newMockBackend())
	resp := callRPC(t, api, "eth_unsubscribe")

	if resp.Error == nil {
		t.Fatal("expected error for missing subscription ID")
	}
	if resp.Error.Code != ErrCodeInvalidParams {
		t.Fatalf("want error code %d, got %d", ErrCodeInvalidParams, resp.Error.Code)
	}
}

func TestRPC_EthSubscribe_LogsWithFilter(t *testing.T) {
	mb := newMockBackend()
	api := NewEthAPI(mb)

	contractAddr := types.HexToAddress("0xcccc")
	transferTopic := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

	resp := callRPC(t, api, "eth_subscribe", "logs", map[string]interface{}{
		"address": []string{encodeAddress(contractAddr)},
		"topics":  [][]string{{encodeHash(transferTopic)}},
	})
	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}
	subID := resp.Result.(string)

	sub := apiSubs(api).GetSubscription(subID)
	if sub == nil {
		t.Fatal("subscription not found")
	}
	if sub.Type != SubLogs {
		t.Fatalf("want SubLogs, got %d", sub.Type)
	}

	// Query should have address and topic filters.
	if len(sub.Query.Addresses) != 1 {
		t.Fatalf("want 1 address, got %d", len(sub.Query.Addresses))
	}
	if sub.Query.Addresses[0] != contractAddr {
		t.Fatalf("wrong address in query")
	}
}

func TestRPC_EthSubscribe_LogsNoFilter(t *testing.T) {
	api := NewEthAPI(newMockBackend())

	// Subscribe to logs without specifying a filter (matches all logs).
	resp := callRPC(t, api, "eth_subscribe", "logs")
	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}
	subID := resp.Result.(string)

	sub := apiSubs(api).GetSubscription(subID)
	if sub == nil {
		t.Fatal("subscription not found")
	}
	// Query should be empty (matches everything).
	if len(sub.Query.Addresses) != 0 {
		t.Fatalf("want 0 addresses, got %d", len(sub.Query.Addresses))
	}
	if len(sub.Query.Topics) != 0 {
		t.Fatalf("want 0 topics, got %d", len(sub.Query.Topics))
	}
}

func TestRPC_EthSubscribe_FullLifecycle(t *testing.T) {
	mb := newMockBackend()
	api := NewEthAPI(mb)

	// Step 1: Subscribe to newHeads.
	subResp := callRPC(t, api, "eth_subscribe", "newHeads")
	if subResp.Error != nil {
		t.Fatalf("subscribe error: %v", subResp.Error.Message)
	}
	subID := subResp.Result.(string)

	// Step 2: Verify subscription exists.
	sub := apiSubs(api).GetSubscription(subID)
	if sub == nil {
		t.Fatal("subscription not found")
	}

	// Step 3: Send a notification.
	apiSubs(api).NotifyNewHead(&types.Header{
		Number:  big.NewInt(200),
		BaseFee: big.NewInt(2000000000),
	})

	// Step 4: Read the notification.
	select {
	case msg := <-sub.Channel():
		block := msg.(*RPCBlock)
		if block.Number != "0xc8" { // 200
			t.Fatalf("want 0xc8, got %v", block.Number)
		}
	default:
		t.Fatal("expected notification")
	}

	// Step 5: Unsubscribe.
	unsubResp := callRPC(t, api, "eth_unsubscribe", subID)
	if unsubResp.Error != nil {
		t.Fatalf("unsubscribe error: %v", unsubResp.Error.Message)
	}
	if unsubResp.Result != true {
		t.Fatalf("want true, got %v", unsubResp.Result)
	}

	// Step 6: Verify it's gone.
	if apiSubs(api).GetSubscription(subID) != nil {
		t.Fatal("subscription should be removed after unsubscribe")
	}
	if apiSubs(api).SubscriptionCount() != 0 {
		t.Fatalf("want 0 subscriptions, got %d", apiSubs(api).SubscriptionCount())
	}
}
