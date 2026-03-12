package rpcsub

import (
	"encoding/json"
	"math/big"
	"sync"
	"testing"

	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
	"github.com/eth2030/eth2030/crypto"
	rpcbackend "github.com/eth2030/eth2030/rpc/backend"
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
	"github.com/eth2030/eth2030/trie"
)

// ---------- mock backend for tests ----------

type testMockBackend struct {
	chainID *big.Int
	headers map[uint64]*types.Header
	logs    map[types.Hash][]*types.Log
}

func (m *testMockBackend) ChainID() *big.Int { return m.chainID }
func (m *testMockBackend) CurrentHeader() *types.Header {
	var best *types.Header
	for _, h := range m.headers {
		if best == nil || h.Number.Uint64() > best.Number.Uint64() {
			best = h
		}
	}
	return best
}
func (m *testMockBackend) HeaderByNumber(n rpctypes.BlockNumber) *types.Header {
	if n == rpctypes.LatestBlockNumber {
		return m.CurrentHeader()
	}
	return m.headers[uint64(n)]
}
func (m *testMockBackend) HeaderByHash(hash types.Hash) *types.Header {
	for _, h := range m.headers {
		if h.Hash() == hash {
			return h
		}
	}
	return nil
}
func (m *testMockBackend) BlockByNumber(n rpctypes.BlockNumber) *types.Block { return nil }
func (m *testMockBackend) BlockByHash(hash types.Hash) *types.Block          { return nil }
func (m *testMockBackend) GetTransaction(hash types.Hash) (*types.Transaction, uint64, uint64) {
	return nil, 0, 0
}
func (m *testMockBackend) GetLogs(blockHash types.Hash) []*types.Log { return m.logs[blockHash] }
func (m *testMockBackend) SendTransaction(tx *types.Transaction) error {
	return nil
}
func (m *testMockBackend) StateAt(root types.Hash) (state.StateDB, error) { return nil, nil }
func (m *testMockBackend) SuggestGasPrice() *big.Int                      { return big.NewInt(1e9) }
func (m *testMockBackend) GetReceipts(blockHash types.Hash) []*types.Receipt {
	return nil
}
func (m *testMockBackend) GetBlockReceipts(number uint64) []*types.Receipt { return nil }
func (m *testMockBackend) GetProof(addr types.Address, storageKeys []types.Hash, blockNumber rpctypes.BlockNumber) (*trie.AccountProof, error) {
	return nil, nil
}
func (m *testMockBackend) EVMCall(from types.Address, to *types.Address, data []byte, gas uint64, value *big.Int, blockNumber rpctypes.BlockNumber) ([]byte, uint64, error) {
	return nil, 0, nil
}
func (m *testMockBackend) TraceTransaction(txHash types.Hash) (*vm.StructLogTracer, error) {
	return nil, nil
}
func (m *testMockBackend) HistoryOldestBlock() uint64 { return 0 }
func (m *testMockBackend) BlobSchedule(_ uint64) (target, max, updateFraction uint64) {
	return 3, 6, 3338477
}

// newTestMockBackend creates a mock backend with two blocks (42 and 43) and logs.
func newTestMockBackend() *testMockBackend {
	mb := &testMockBackend{
		chainID: big.NewInt(1),
		headers: map[uint64]*types.Header{
			42: {
				Number:   big.NewInt(42),
				GasLimit: 30000000,
				GasUsed:  10000000,
				Time:     1700000000,
				BaseFee:  big.NewInt(1000000000),
			},
		},
		logs: map[types.Hash][]*types.Log{},
	}

	header43 := &types.Header{
		Number:   big.NewInt(43),
		GasLimit: 30000000,
		GasUsed:  10000000,
		Time:     1700000012,
		BaseFee:  big.NewInt(1000000000),
	}
	mb.headers[43] = header43

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

// Verify testMockBackend implements rpcbackend.Backend at compile time.
var _ rpcbackend.Backend = (*testMockBackend)(nil)

// ---------- subscription manager unit tests ----------

func TestNewLogFilter_AllLogs(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	from := uint64(42)
	to := uint64(43)
	id := sm.NewLogFilter(rpcfilter.FilterQuery{
		FromBlock: &from,
		ToBlock:   &to,
	})

	if id == "" {
		t.Fatal("expected non-empty filter ID")
	}

	logs, ok := sm.GetFilterLogs(id)
	if !ok {
		t.Fatal("filter not found")
	}
	if len(logs) != 3 {
		t.Fatalf("want 3 logs, got %d", len(logs))
	}
}

func TestNewLogFilter_AddressFilter(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	contractA := types.HexToAddress("0xaaaa")
	from := uint64(42)
	to := uint64(43)
	id := sm.NewLogFilter(rpcfilter.FilterQuery{
		FromBlock: &from,
		ToBlock:   &to,
		Addresses: []types.Address{contractA},
	})

	logs, ok := sm.GetFilterLogs(id)
	if !ok {
		t.Fatal("filter not found")
	}
	if len(logs) != 2 {
		t.Fatalf("want 2 logs for contractA, got %d", len(logs))
	}
}

func TestNewLogFilter_TopicFilter(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	transferTopic := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	from := uint64(42)
	to := uint64(43)
	id := sm.NewLogFilter(rpcfilter.FilterQuery{
		FromBlock: &from,
		ToBlock:   &to,
		Topics:    [][]types.Hash{{transferTopic}},
	})

	logs, ok := sm.GetFilterLogs(id)
	if !ok {
		t.Fatal("filter not found")
	}
	if len(logs) != 2 {
		t.Fatalf("want 2 transfer logs, got %d", len(logs))
	}
}

func TestNewLogFilter_TopicWithWildcard(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	from := uint64(42)
	to := uint64(43)
	// Position 0: wildcard (nil) matches any topic.
	id := sm.NewLogFilter(rpcfilter.FilterQuery{
		FromBlock: &from,
		ToBlock:   &to,
		Topics:    [][]types.Hash{{}}, // empty = wildcard for position 0
	})

	logs, ok := sm.GetFilterLogs(id)
	if !ok {
		t.Fatal("filter not found")
	}
	if len(logs) != 3 {
		t.Fatalf("want 3 logs (wildcard topic matches all), got %d", len(logs))
	}
}

func TestNewLogFilter_MultiTopicOR(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	transferTopic := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	approvalTopic := crypto.Keccak256Hash([]byte("Approval(address,address,uint256)"))
	from := uint64(42)
	to := uint64(43)
	id := sm.NewLogFilter(rpcfilter.FilterQuery{
		FromBlock: &from,
		ToBlock:   &to,
		Topics:    [][]types.Hash{{transferTopic, approvalTopic}},
	})

	logs, ok := sm.GetFilterLogs(id)
	if !ok {
		t.Fatal("filter not found")
	}
	if len(logs) != 3 {
		t.Fatalf("want 3 logs (OR match), got %d", len(logs))
	}
}

func TestGetFilterChanges_LogFilter(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	// Install a filter starting at block 42.
	from := uint64(42)
	id := sm.NewLogFilter(rpcfilter.FilterQuery{FromBlock: &from})

	// First poll: should return logs from blocks 42-43.
	result, ok := sm.GetFilterChanges(id)
	if !ok {
		t.Fatal("filter not found")
	}
	logs, isLogs := result.([]*types.Log)
	if !isLogs {
		t.Fatalf("expected []*types.Log, got %T", result)
	}
	if len(logs) != 3 {
		t.Fatalf("first poll: want 3 logs, got %d", len(logs))
	}

	// Second poll: no new blocks, should return empty.
	result2, ok2 := sm.GetFilterChanges(id)
	if !ok2 {
		t.Fatal("filter not found on second poll")
	}
	logs2, isLogs2 := result2.([]*types.Log)
	if !isLogs2 {
		t.Fatalf("expected []*types.Log, got %T", result2)
	}
	if len(logs2) != 0 {
		t.Fatalf("second poll: want 0 logs, got %d", len(logs2))
	}
}

func TestBlockFilter(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	// Install a block filter.
	id := sm.NewBlockFilter()

	// No new blocks since filter was created.
	result, ok := sm.GetFilterChanges(id)
	if !ok {
		t.Fatal("filter not found")
	}
	hashes, _ := result.([]types.Hash)
	if len(hashes) != 0 {
		t.Fatalf("want 0 hashes before new blocks, got %d", len(hashes))
	}

	// Notify a new block.
	newHash := types.HexToHash("0xdeadbeef")
	sm.NotifyNewBlock(newHash)

	result2, ok2 := sm.GetFilterChanges(id)
	if !ok2 {
		t.Fatal("filter not found on second poll")
	}
	hashes2, _ := result2.([]types.Hash)
	if len(hashes2) < 1 {
		t.Fatalf("want at least 1 hash after notification, got %d", len(hashes2))
	}
	found := false
	for _, h := range hashes2 {
		if h == newHash {
			found = true
		}
	}
	if !found {
		t.Fatal("notified hash not returned by GetFilterChanges")
	}
}

func TestPendingTxFilter(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	id := sm.NewPendingTxFilter()
	txHash := types.HexToHash("0x1234abcd")
	sm.NotifyPendingTx(txHash)

	result, ok := sm.GetFilterChanges(id)
	if !ok {
		t.Fatal("filter not found")
	}
	hashes, _ := result.([]types.Hash)
	if len(hashes) != 1 {
		t.Fatalf("want 1 pending tx hash, got %d", len(hashes))
	}
	if hashes[0] != txHash {
		t.Fatalf("want %v, got %v", txHash, hashes[0])
	}

	// Second poll: should be empty.
	result2, _ := sm.GetFilterChanges(id)
	hashes2, _ := result2.([]types.Hash)
	if len(hashes2) != 0 {
		t.Fatalf("second poll should be empty, got %d", len(hashes2))
	}
}

func TestUninstallFilter(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	id := sm.NewBlockFilter()
	if sm.FilterCount() != 1 {
		t.Fatalf("want 1 filter, got %d", sm.FilterCount())
	}

	ok := sm.Uninstall(id)
	if !ok {
		t.Fatal("expected Uninstall to return true")
	}
	if sm.FilterCount() != 0 {
		t.Fatalf("want 0 filters after uninstall, got %d", sm.FilterCount())
	}

	ok2 := sm.Uninstall(id)
	if ok2 {
		t.Fatal("second Uninstall should return false")
	}
}

func TestCleanupStaleFilters(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	_ = sm.NewBlockFilter()
	_ = sm.NewPendingTxFilter()
	if sm.FilterCount() != 2 {
		t.Fatalf("want 2 filters before cleanup, got %d", sm.FilterCount())
	}

	// Immediately after creation filters are fresh, so none should be removed.
	removed := sm.CleanupStale()
	if removed != 0 {
		t.Fatalf("want 0 removed immediately, got %d", removed)
	}
}

func TestGetFilterChanges_NonExistent(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	_, ok := sm.GetFilterChanges("0xnonexistent")
	if ok {
		t.Fatal("expected false for non-existent filter")
	}
}

func TestGetFilterLogs_NonExistent(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	_, ok := sm.GetFilterLogs("0xnonexistent")
	if ok {
		t.Fatal("expected false for non-existent filter")
	}
}

func TestGetFilterLogs_WrongType(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	id := sm.NewBlockFilter()
	_, ok := sm.GetFilterLogs(id)
	if ok {
		t.Fatal("expected false for block filter on GetFilterLogs")
	}
}

// ---------- MatchFilter unit tests ----------

func TestMatchFilter_EmptyQuery(t *testing.T) {
	log := &types.Log{
		Address: types.HexToAddress("0xaaaa"),
		Topics:  []types.Hash{types.HexToHash("0x1111")},
	}
	query := rpcfilter.FilterQuery{}
	if !rpcfilter.MatchFilter(log, query) {
		t.Fatal("empty query should match everything")
	}
}

func TestMatchFilter_AddressMatch(t *testing.T) {
	addr := types.HexToAddress("0xaaaa")
	log := &types.Log{Address: addr}

	q := rpcfilter.FilterQuery{Addresses: []types.Address{addr}}
	if !rpcfilter.MatchFilter(log, q) {
		t.Fatal("should match same address")
	}

	other := types.HexToAddress("0xbbbb")
	q2 := rpcfilter.FilterQuery{Addresses: []types.Address{other}}
	if rpcfilter.MatchFilter(log, q2) {
		t.Fatal("should not match different address")
	}
}

func TestMatchFilter_MultiAddressOR(t *testing.T) {
	addr := types.HexToAddress("0xaaaa")
	log := &types.Log{Address: addr}

	other := types.HexToAddress("0xbbbb")
	q := rpcfilter.FilterQuery{Addresses: []types.Address{other, addr}}
	if !rpcfilter.MatchFilter(log, q) {
		t.Fatal("should match when any address matches")
	}
}

func TestMatchFilter_TopicAND(t *testing.T) {
	topic0 := types.HexToHash("0x1111")
	topic1 := types.HexToHash("0x2222")
	log := &types.Log{Topics: []types.Hash{topic0, topic1}}

	q := rpcfilter.FilterQuery{Topics: [][]types.Hash{{topic0}, {topic1}}}
	if !rpcfilter.MatchFilter(log, q) {
		t.Fatal("should match when both topic positions match")
	}

	wrongTopic := types.HexToHash("0x3333")
	q2 := rpcfilter.FilterQuery{Topics: [][]types.Hash{{topic0}, {wrongTopic}}}
	if rpcfilter.MatchFilter(log, q2) {
		t.Fatal("should not match when topic[1] doesn't match")
	}
}

func TestMatchFilter_TopicORWithinPosition(t *testing.T) {
	topic0 := types.HexToHash("0x1111")
	log := &types.Log{Topics: []types.Hash{topic0}}

	alt := types.HexToHash("0x9999")
	q := rpcfilter.FilterQuery{Topics: [][]types.Hash{{alt, topic0}}}
	if !rpcfilter.MatchFilter(log, q) {
		t.Fatal("should match when any topic in position matches")
	}
}

func TestMatchFilter_TopicShortLog(t *testing.T) {
	log := &types.Log{Topics: []types.Hash{types.HexToHash("0x1111")}}
	q := rpcfilter.FilterQuery{Topics: [][]types.Hash{{types.HexToHash("0x1111")}, {types.HexToHash("0x2222")}}}
	if rpcfilter.MatchFilter(log, q) {
		t.Fatal("should not match when log has fewer topics than filter requires")
	}
}

// ---------- bloom filter tests ----------

func TestBloomMatchesQuery_Address(t *testing.T) {
	addr := types.HexToAddress("0xaaaa")
	log := &types.Log{Address: addr}
	bloom := types.LogsBloom([]*types.Log{log})

	q := rpcfilter.FilterQuery{Addresses: []types.Address{addr}}
	if !bloomMatchesQuery(bloom, q) {
		t.Fatal("bloom should match address that was added")
	}

	other := types.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	q2 := rpcfilter.FilterQuery{Addresses: []types.Address{other}}
	_ = bloomMatchesQuery(bloom, q2) // just ensure it doesn't crash
}

func TestBloomMatchesQuery_Topic(t *testing.T) {
	topic := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	log := &types.Log{Topics: []types.Hash{topic}}
	bloom := types.LogsBloom([]*types.Log{log})

	q := rpcfilter.FilterQuery{Topics: [][]types.Hash{{topic}}}
	if !bloomMatchesQuery(bloom, q) {
		t.Fatal("bloom should match topic that was added")
	}
}

func TestBloomMatchesQuery_Wildcard(t *testing.T) {
	bloom := types.Bloom{}
	q := rpcfilter.FilterQuery{}
	if !bloomMatchesQuery(bloom, q) {
		t.Fatal("empty query should match any bloom (wildcard)")
	}
}

// ---------- FilterLogs function tests ----------

func TestFilterLogs(t *testing.T) {
	addr := types.HexToAddress("0xaaaa")
	other := types.HexToAddress("0xbbbb")
	topic := types.HexToHash("0x1111")

	logs := []*types.Log{
		{Address: addr, Topics: []types.Hash{topic}},
		{Address: other, Topics: []types.Hash{topic}},
		{Address: addr, Topics: []types.Hash{types.HexToHash("0x2222")}},
	}

	result := rpcfilter.FilterLogs(logs, rpcfilter.FilterQuery{Addresses: []types.Address{addr}})
	if len(result) != 2 {
		t.Fatalf("want 2 logs for addr, got %d", len(result))
	}

	result2 := rpcfilter.FilterLogs(logs, rpcfilter.FilterQuery{Topics: [][]types.Hash{{topic}}})
	if len(result2) != 2 {
		t.Fatalf("want 2 logs for topic, got %d", len(result2))
	}

	result3 := rpcfilter.FilterLogs(logs, rpcfilter.FilterQuery{
		Addresses: []types.Address{addr},
		Topics:    [][]types.Hash{{topic}},
	})
	if len(result3) != 1 {
		t.Fatalf("want 1 log for addr+topic, got %d", len(result3))
	}
}

func TestFilterLogsWithBloom(t *testing.T) {
	addr := types.HexToAddress("0xaaaa")
	topic := types.HexToHash("0x1111")
	logs := []*types.Log{
		{Address: addr, Topics: []types.Hash{topic}},
	}
	bloom := types.LogsBloom(logs)

	result := rpcfilter.FilterLogsWithBloom(bloom, logs, rpcfilter.FilterQuery{Addresses: []types.Address{addr}})
	if len(result) != 1 {
		t.Fatalf("want 1 log, got %d", len(result))
	}
}

// ---------- edge cases ----------

func TestQueryLogs_EmptyChain(t *testing.T) {
	mb := &testMockBackend{
		chainID: big.NewInt(1),
		headers: map[uint64]*types.Header{},
		logs:    map[types.Hash][]*types.Log{},
	}
	sm := NewSubscriptionManager(mb)
	result := sm.QueryLogs(rpcfilter.FilterQuery{})
	if len(result) != 0 {
		t.Fatalf("want 0 logs on empty chain, got %d", len(result))
	}
}

func TestMultipleFilters(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	id1 := sm.NewBlockFilter()
	id2 := sm.NewPendingTxFilter()
	from := uint64(42)
	id3 := sm.NewLogFilter(rpcfilter.FilterQuery{FromBlock: &from})

	if sm.FilterCount() != 3 {
		t.Fatalf("want 3 filters, got %d", sm.FilterCount())
	}

	if id1 == id2 || id2 == id3 || id1 == id3 {
		t.Fatal("filter IDs should be unique")
	}

	sm.Uninstall(id1)
	sm.Uninstall(id2)
	sm.Uninstall(id3)
	if sm.FilterCount() != 0 {
		t.Fatalf("want 0 filters, got %d", sm.FilterCount())
	}
}

// ---------- WebSocket subscription manager tests ----------

func TestSubscriptionManager_Subscribe(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	id := sm.Subscribe(SubNewHeads, rpcfilter.FilterQuery{})
	if id == "" {
		t.Fatal("expected non-empty subscription ID")
	}

	sub := sm.GetSubscription(id)
	if sub == nil {
		t.Fatal("subscription not found after Subscribe")
	}
	if sub.Type != SubNewHeads {
		t.Fatalf("want SubNewHeads, got %d", sub.Type)
	}
	if sub.ID != id {
		t.Fatalf("want ID %q, got %q", id, sub.ID)
	}
}

func TestSubscriptionManager_SubscribeAndUnsubscribe(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	id := sm.Subscribe(SubLogs, rpcfilter.FilterQuery{})
	if sm.SubscriptionCount() != 1 {
		t.Fatalf("want 1 subscription, got %d", sm.SubscriptionCount())
	}

	ok := sm.Unsubscribe(id)
	if !ok {
		t.Fatal("expected Unsubscribe to return true")
	}
	if sm.SubscriptionCount() != 0 {
		t.Fatalf("want 0 subscriptions after unsubscribe, got %d", sm.SubscriptionCount())
	}

	ok2 := sm.Unsubscribe(id)
	if ok2 {
		t.Fatal("expected Unsubscribe to return false for already removed subscription")
	}
}

func TestSubscriptionManager_GetSubscription_NonExistent(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	sub := sm.GetSubscription("0xnonexistent")
	if sub != nil {
		t.Fatal("expected nil for non-existent subscription")
	}
}

func TestSubscriptionManager_MultipleSubscriptions(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	id1 := sm.Subscribe(SubNewHeads, rpcfilter.FilterQuery{})
	id2 := sm.Subscribe(SubLogs, rpcfilter.FilterQuery{})
	id3 := sm.Subscribe(SubPendingTx, rpcfilter.FilterQuery{})

	if sm.SubscriptionCount() != 3 {
		t.Fatalf("want 3 subscriptions, got %d", sm.SubscriptionCount())
	}

	if id1 == id2 || id2 == id3 || id1 == id3 {
		t.Fatal("subscription IDs should be unique")
	}

	sm.Unsubscribe(id2)
	if sm.SubscriptionCount() != 2 {
		t.Fatalf("want 2 subscriptions, got %d", sm.SubscriptionCount())
	}
}

func TestNotifyNewHead_MultipleSubscribers(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	id1 := sm.Subscribe(SubNewHeads, rpcfilter.FilterQuery{})
	id2 := sm.Subscribe(SubNewHeads, rpcfilter.FilterQuery{})
	logSubID := sm.Subscribe(SubLogs, rpcfilter.FilterQuery{})

	header := &types.Header{
		Number:  big.NewInt(100),
		BaseFee: big.NewInt(1000000000),
	}
	sm.NotifyNewHead(header)

	sub1 := sm.GetSubscription(id1)
	sub2 := sm.GetSubscription(id2)
	logSub := sm.GetSubscription(logSubID)

	for _, sub := range []*Subscription{sub1, sub2} {
		select {
		case msg := <-sub.Channel():
			block := msg.(*rpctypes.RPCBlock)
			if block.Number != "0x64" {
				t.Fatalf("want 0x64, got %v", block.Number)
			}
		default:
			t.Fatal("expected notification on newHeads channel")
		}
	}

	select {
	case <-logSub.Channel():
		t.Fatal("log subscription should not receive newHeads notification")
	default:
		// Good.
	}
}

func TestNotifyNewHead_NoSubscribers(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	header := &types.Header{
		Number:  big.NewInt(50),
		BaseFee: big.NewInt(1000000000),
	}
	sm.NotifyNewHead(header)
}

func TestNotifyLogs_TopicMatching(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	transferTopic := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	approvalTopic := crypto.Keccak256Hash([]byte("Approval(address,address,uint256)"))

	transferSubID := sm.Subscribe(SubLogs, rpcfilter.FilterQuery{
		Topics: [][]types.Hash{{transferTopic}},
	})
	approvalSubID := sm.Subscribe(SubLogs, rpcfilter.FilterQuery{
		Topics: [][]types.Hash{{approvalTopic}},
	})

	logs := []*types.Log{
		{Address: types.HexToAddress("0xaaaa"), Topics: []types.Hash{transferTopic}, BlockNumber: 42},
		{Address: types.HexToAddress("0xbbbb"), Topics: []types.Hash{approvalTopic}, BlockNumber: 42},
	}
	sm.NotifyLogs(logs)

	transferSub := sm.GetSubscription(transferSubID)
	approvalSub := sm.GetSubscription(approvalSubID)

	select {
	case msg := <-transferSub.Channel():
		rpcLog := msg.(*rpctypes.RPCLog)
		if rpcLog.Address != rpctypes.EncodeAddress(types.HexToAddress("0xaaaa")) {
			t.Fatalf("wrong address in transfer notification")
		}
	default:
		t.Fatal("expected transfer notification")
	}

	select {
	case <-transferSub.Channel():
		t.Fatal("unexpected extra notification for transfer sub")
	default:
	}

	select {
	case msg := <-approvalSub.Channel():
		rpcLog := msg.(*rpctypes.RPCLog)
		if rpcLog.Address != rpctypes.EncodeAddress(types.HexToAddress("0xbbbb")) {
			t.Fatalf("wrong address in approval notification")
		}
	default:
		t.Fatal("expected approval notification")
	}
}

func TestNotifyLogs_NoSubscribers(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	logs := []*types.Log{
		{Address: types.HexToAddress("0xaaaa"), Topics: []types.Hash{types.HexToHash("0x1111")}},
	}
	sm.NotifyLogs(logs)
}

func TestNotifyLogs_AllMatch(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	id := sm.Subscribe(SubLogs, rpcfilter.FilterQuery{})
	sub := sm.GetSubscription(id)

	logs := []*types.Log{
		{Address: types.HexToAddress("0xaaaa"), Topics: []types.Hash{types.HexToHash("0x1111")}},
		{Address: types.HexToAddress("0xbbbb"), Topics: []types.Hash{types.HexToHash("0x2222")}},
	}
	sm.NotifyLogs(logs)

	for i := 0; i < 2; i++ {
		select {
		case <-sub.Channel():
			// Good.
		default:
			t.Fatalf("expected notification %d", i)
		}
	}

	select {
	case <-sub.Channel():
		t.Fatal("unexpected extra notification")
	default:
	}
}

func TestNotifyPendingTxHash_MultipleSubscribers(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	id1 := sm.Subscribe(SubPendingTx, rpcfilter.FilterQuery{})
	id2 := sm.Subscribe(SubPendingTx, rpcfilter.FilterQuery{})
	headsID := sm.Subscribe(SubNewHeads, rpcfilter.FilterQuery{})

	txHash := types.HexToHash("0xabcdef")
	sm.NotifyPendingTxHash(txHash)

	for _, id := range []string{id1, id2} {
		sub := sm.GetSubscription(id)
		select {
		case msg := <-sub.Channel():
			hashStr := msg.(string)
			if hashStr != rpctypes.EncodeHash(txHash) {
				t.Fatalf("want %v, got %v", rpctypes.EncodeHash(txHash), hashStr)
			}
		default:
			t.Fatalf("expected notification for pending tx sub %s", id)
		}
	}

	headsSub := sm.GetSubscription(headsID)
	select {
	case <-headsSub.Channel():
		t.Fatal("newHeads sub should not receive pending tx hash")
	default:
	}
}

func TestNotifyPendingTxHash_NoSubscribers(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)
	sm.NotifyPendingTxHash(types.HexToHash("0x1234"))
}

func TestSubscription_BufferOverflow(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	id := sm.Subscribe(SubPendingTx, rpcfilter.FilterQuery{})
	sub := sm.GetSubscription(id)

	for i := 0; i < subscriptionBufferSize; i++ {
		hash := types.HexToHash("0x" + string(rune(i+1)))
		sm.NotifyPendingTxHash(hash)
	}

	sm.NotifyPendingTxHash(types.HexToHash("0xoverflow"))

	count := 0
	for {
		select {
		case <-sub.Channel():
			count++
		default:
			goto done
		}
	}
done:
	if count != subscriptionBufferSize {
		t.Fatalf("want %d messages (buffer size), got %d", subscriptionBufferSize, count)
	}
}

func TestSubscription_ChannelClosedOnUnsubscribe(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	id := sm.Subscribe(SubNewHeads, rpcfilter.FilterQuery{})
	sub := sm.GetSubscription(id)

	sm.Unsubscribe(id)

	_, open := <-sub.Channel()
	if open {
		t.Fatal("channel should be closed after unsubscribe")
	}
}

// ---------- FormatWSNotification tests ----------

func TestFormatWSNotification_RoundTrip(t *testing.T) {
	header := &types.Header{
		Number:  big.NewInt(42),
		BaseFee: big.NewInt(1000000000),
	}
	block := rpctypes.FormatHeader(header)

	notif := FormatWSNotification("0xsub123", block)

	data, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded WSNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.JSONRPC != "2.0" {
		t.Fatalf("want jsonrpc 2.0, got %v", decoded.JSONRPC)
	}
	if decoded.Method != "eth_subscription" {
		t.Fatalf("want method eth_subscription, got %v", decoded.Method)
	}

	var subResult WSSubscriptionResult
	if err := json.Unmarshal(decoded.Params, &subResult); err != nil {
		t.Fatalf("unmarshal params: %v", err)
	}
	if subResult.Subscription != "0xsub123" {
		t.Fatalf("want subscription 0xsub123, got %v", subResult.Subscription)
	}
}

func TestFormatWSNotification_NilResult(t *testing.T) {
	notif := FormatWSNotification("0xabc", nil)
	if notif.JSONRPC != "2.0" {
		t.Fatalf("want 2.0, got %v", notif.JSONRPC)
	}
	if notif.Method != "eth_subscription" {
		t.Fatalf("want eth_subscription, got %v", notif.Method)
	}

	var result WSSubscriptionResult
	json.Unmarshal(notif.Params, &result)
	if result.Subscription != "0xabc" {
		t.Fatalf("want 0xabc, got %v", result.Subscription)
	}
	if result.Result != nil {
		t.Fatalf("want nil result, got %v", result.Result)
	}
}

func TestFormatWSNotification_StringResult(t *testing.T) {
	notif := FormatWSNotification("0xdef", "0x1234")

	var result WSSubscriptionResult
	json.Unmarshal(notif.Params, &result)
	if result.Subscription != "0xdef" {
		t.Fatalf("want 0xdef, got %v", result.Subscription)
	}
}

// ---------- Concurrent subscription operations ----------

func TestConcurrentSubscriptions(t *testing.T) {
	mb := newTestMockBackend()
	sm := NewSubscriptionManager(mb)

	var wg sync.WaitGroup
	subIDs := make(chan string, 50)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			var subType SubType
			switch n % 3 {
			case 0:
				subType = SubNewHeads
			case 1:
				subType = SubLogs
			case 2:
				subType = SubPendingTx
			}
			id := sm.Subscribe(subType, rpcfilter.FilterQuery{})
			subIDs <- id
		}(i)
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sm.NotifyNewHead(&types.Header{
				Number:  big.NewInt(100),
				BaseFee: big.NewInt(1000000000),
			})
			sm.NotifyPendingTxHash(types.HexToHash("0x1234"))
			sm.NotifyLogs([]*types.Log{
				{Address: types.HexToAddress("0xaaaa"), Topics: []types.Hash{types.HexToHash("0x1111")}},
			})
		}()
	}

	wg.Wait()
	close(subIDs)

	if sm.SubscriptionCount() != 20 {
		t.Fatalf("want 20 subscriptions, got %d", sm.SubscriptionCount())
	}

	var wg2 sync.WaitGroup
	for id := range subIDs {
		wg2.Add(1)
		go func(sid string) {
			defer wg2.Done()
			sm.Unsubscribe(sid)
		}(id)
	}
	wg2.Wait()

	if sm.SubscriptionCount() != 0 {
		t.Fatalf("want 0 subscriptions, got %d", sm.SubscriptionCount())
	}
}

// ---------- FormatBlock tests ----------

func TestFormatBlock_WithFullTx(t *testing.T) {
	header := &types.Header{
		Number:  big.NewInt(10),
		BaseFee: big.NewInt(1000000000),
	}
	to := types.HexToAddress("0xbbbb")
	tx := types.NewTransaction(&types.LegacyTx{
		Nonce:    1,
		GasPrice: big.NewInt(1000000000),
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(1000),
	})
	sender := types.HexToAddress("0xaaaa")
	tx.SetSender(sender)

	block := types.NewBlock(header, &types.Body{
		Transactions: []*types.Transaction{tx},
	})
	result := rpctypes.FormatBlock(block, true)

	blockWithTxs, ok := result.(*rpctypes.RPCBlockWithTxs)
	if !ok {
		t.Fatalf("expected *RPCBlockWithTxs, got %T", result)
	}
	if len(blockWithTxs.Transactions) != 1 {
		t.Fatalf("want 1 tx, got %d", len(blockWithTxs.Transactions))
	}
	if blockWithTxs.Transactions[0].Nonce != "0x1" {
		t.Fatalf("want nonce 0x1, got %v", blockWithTxs.Transactions[0].Nonce)
	}
}

func TestFormatBlock_EmptyBlock_FullTx(t *testing.T) {
	header := &types.Header{
		Number:  big.NewInt(10),
		BaseFee: big.NewInt(1000000000),
	}
	block := types.NewBlock(header, nil)
	result := rpctypes.FormatBlock(block, true)

	blockWithTxs := result.(*rpctypes.RPCBlockWithTxs)
	if len(blockWithTxs.Transactions) != 0 {
		t.Fatalf("want 0 txs, got %d", len(blockWithTxs.Transactions))
	}
}

// ---------- FormatLog tests ----------

func TestFormatLog(t *testing.T) {
	addr := types.HexToAddress("0xcccc")
	topic1 := types.HexToHash("0x1111")
	topic2 := types.HexToHash("0x2222")
	blockHash := types.HexToHash("0xbeef")
	txHash := types.HexToHash("0xdead")

	log := &types.Log{
		Address:     addr,
		Topics:      []types.Hash{topic1, topic2},
		Data:        []byte{0xab, 0xcd},
		BlockNumber: 42,
		BlockHash:   blockHash,
		TxHash:      txHash,
		TxIndex:     3,
		Index:       7,
		Removed:     false,
	}

	rpcLog := rpctypes.FormatLog(log)
	if rpcLog.Address != rpctypes.EncodeAddress(addr) {
		t.Fatalf("want address %v, got %v", rpctypes.EncodeAddress(addr), rpcLog.Address)
	}
	if len(rpcLog.Topics) != 2 {
		t.Fatalf("want 2 topics, got %d", len(rpcLog.Topics))
	}
	if rpcLog.Topics[0] != rpctypes.EncodeHash(topic1) {
		t.Fatalf("topic 0 mismatch")
	}
	if rpcLog.Data != "0xabcd" {
		t.Fatalf("want data 0xabcd, got %v", rpcLog.Data)
	}
	if rpcLog.BlockNumber != "0x2a" {
		t.Fatalf("want blockNumber 0x2a, got %v", rpcLog.BlockNumber)
	}
	if rpcLog.TransactionHash != rpctypes.EncodeHash(txHash) {
		t.Fatalf("txHash mismatch")
	}
	if rpcLog.TransactionIndex != "0x3" {
		t.Fatalf("want txIndex 0x3, got %v", rpcLog.TransactionIndex)
	}
	if rpcLog.LogIndex != "0x7" {
		t.Fatalf("want logIndex 0x7, got %v", rpcLog.LogIndex)
	}
	if rpcLog.Removed {
		t.Fatal("want removed=false")
	}
}

func TestFormatLog_RemovedFlag(t *testing.T) {
	log := &types.Log{
		Address: types.HexToAddress("0xaaaa"),
		Removed: true,
	}
	rpcLog := rpctypes.FormatLog(log)
	if !rpcLog.Removed {
		t.Fatal("want removed=true")
	}
}

func TestFormatLog_NoTopics(t *testing.T) {
	log := &types.Log{
		Address: types.HexToAddress("0xaaaa"),
		Topics:  []types.Hash{},
	}
	rpcLog := rpctypes.FormatLog(log)
	if len(rpcLog.Topics) != 0 {
		t.Fatalf("want 0 topics, got %d", len(rpcLog.Topics))
	}
}

// ---------- FormatTransaction tests ----------

func TestFormatTransaction_Pending(t *testing.T) {
	to := types.HexToAddress("0xbbbb")
	tx := types.NewTransaction(&types.LegacyTx{
		Nonce:    5,
		GasPrice: big.NewInt(2000000000),
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(1e18),
		Data:     []byte{0x12, 0x34},
	})
	sender := types.HexToAddress("0xaaaa")
	tx.SetSender(sender)

	rpcTx := rpctypes.FormatTransaction(tx, nil, nil, nil, 0, nil)

	if rpcTx.BlockHash != nil {
		t.Fatalf("want nil blockHash for pending tx, got %v", *rpcTx.BlockHash)
	}
	if rpcTx.BlockNumber != nil {
		t.Fatalf("want nil blockNumber for pending tx, got %v", *rpcTx.BlockNumber)
	}
	if rpcTx.TransactionIndex != nil {
		t.Fatalf("want nil txIndex for pending tx, got %v", *rpcTx.TransactionIndex)
	}
	if rpcTx.Nonce != "0x5" {
		t.Fatalf("want nonce 0x5, got %v", rpcTx.Nonce)
	}
	if rpcTx.From != rpctypes.EncodeAddress(sender) {
		t.Fatalf("want from %v, got %v", rpctypes.EncodeAddress(sender), rpcTx.From)
	}
}

func TestFormatTransaction_ContractCreation(t *testing.T) {
	tx := types.NewTransaction(&types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1000000000),
		Gas:      100000,
		Value:    big.NewInt(0),
		Data:     []byte{0x60, 0x00},
	})

	rpcTx := rpctypes.FormatTransaction(tx, nil, nil, nil, 0, nil)
	if rpcTx.To != nil {
		t.Fatalf("want nil to for contract creation, got %v", *rpcTx.To)
	}
}

// ---------- FormatReceipt tests ----------

func TestFormatReceipt_ContractCreation(t *testing.T) {
	contractAddr := types.HexToAddress("0xcccc")
	receipt := &types.Receipt{
		Status:            types.ReceiptStatusSuccessful,
		CumulativeGasUsed: 100000,
		GasUsed:           100000,
		TxHash:            types.HexToHash("0x1111"),
		BlockHash:         types.HexToHash("0x2222"),
		BlockNumber:       big.NewInt(42),
		TransactionIndex:  0,
		ContractAddress:   contractAddr,
		Logs:              []*types.Log{},
	}

	rpcReceipt := rpctypes.FormatReceipt(receipt, nil, 0)
	if rpcReceipt.ContractAddress == nil {
		t.Fatal("expected non-nil contractAddress")
	}
	if *rpcReceipt.ContractAddress != rpctypes.EncodeAddress(contractAddr) {
		t.Fatalf("want contractAddress %v, got %v", rpctypes.EncodeAddress(contractAddr), *rpcReceipt.ContractAddress)
	}
}

func TestFormatReceipt_NilContractAddress(t *testing.T) {
	receipt := &types.Receipt{
		Status:            types.ReceiptStatusSuccessful,
		CumulativeGasUsed: 21000,
		GasUsed:           21000,
		TxHash:            types.HexToHash("0x1111"),
		BlockHash:         types.HexToHash("0x2222"),
		BlockNumber:       big.NewInt(42),
		Logs:              []*types.Log{},
	}

	rpcReceipt := rpctypes.FormatReceipt(receipt, nil, 0)
	if rpcReceipt.ContractAddress != nil {
		t.Fatalf("want nil contractAddress, got %v", *rpcReceipt.ContractAddress)
	}
}

func TestFormatReceipt_FailedStatus(t *testing.T) {
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusFailed,
		GasUsed:     21000,
		TxHash:      types.HexToHash("0x1111"),
		BlockHash:   types.HexToHash("0x2222"),
		BlockNumber: big.NewInt(42),
		Logs:        []*types.Log{},
	}

	rpcReceipt := rpctypes.FormatReceipt(receipt, nil, 0)
	if rpcReceipt.Status != "0x0" {
		t.Fatalf("want status 0x0 (failed), got %v", rpcReceipt.Status)
	}
}

func TestFormatReceipt_NilLogs(t *testing.T) {
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		GasUsed:     21000,
		TxHash:      types.HexToHash("0x1111"),
		BlockHash:   types.HexToHash("0x2222"),
		BlockNumber: big.NewInt(42),
		Logs:        nil,
	}

	rpcReceipt := rpctypes.FormatReceipt(receipt, nil, 0)
	if rpcReceipt.Logs == nil {
		t.Fatal("want non-nil Logs slice")
	}
	if len(rpcReceipt.Logs) != 0 {
		t.Fatalf("want 0 logs, got %d", len(rpcReceipt.Logs))
	}
}
