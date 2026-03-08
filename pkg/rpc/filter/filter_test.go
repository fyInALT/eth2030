package rpcfilter

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
)

// ---------- MatchFilter edge cases ----------

func TestMatchFilter_NilLog(t *testing.T) {
	log := &types.Log{
		Address: types.HexToAddress("0xaaaa"),
	}
	q := FilterQuery{}
	if !MatchFilter(log, q) {
		t.Fatal("empty query should match log with nil topics")
	}
}

func TestMatchFilter_EmptyTopicSlice(t *testing.T) {
	log := &types.Log{
		Address: types.HexToAddress("0xaaaa"),
		Topics:  []types.Hash{},
	}
	q := FilterQuery{Topics: [][]types.Hash{}}
	if !MatchFilter(log, q) {
		t.Fatal("empty topic slices should match")
	}
}

func TestMatchFilter_WildcardInMiddlePosition(t *testing.T) {
	topic0 := types.HexToHash("0x1111")
	topic2 := types.HexToHash("0x3333")
	log := &types.Log{
		Topics: []types.Hash{topic0, types.HexToHash("0x2222"), topic2},
	}
	q := FilterQuery{
		Topics: [][]types.Hash{
			{topic0},
			{}, // wildcard
			{topic2},
		},
	}
	if !MatchFilter(log, q) {
		t.Fatal("should match with wildcard in middle position")
	}
}

func TestMatchFilter_AllWildcards(t *testing.T) {
	log := &types.Log{
		Address: types.HexToAddress("0xaaaa"),
		Topics:  []types.Hash{types.HexToHash("0x1111"), types.HexToHash("0x2222")},
	}
	q := FilterQuery{
		Topics: [][]types.Hash{{}, {}, {}},
	}
	if !MatchFilter(log, q) {
		t.Fatal("all-wildcard query should match any log")
	}
}

func TestMatchFilter_MultipleAddresses_NoMatch(t *testing.T) {
	log := &types.Log{
		Address: types.HexToAddress("0xcccc"),
	}
	q := FilterQuery{
		Addresses: []types.Address{
			types.HexToAddress("0xaaaa"),
			types.HexToAddress("0xbbbb"),
		},
	}
	if MatchFilter(log, q) {
		t.Fatal("should not match when log address is not in address list")
	}
}

func TestMatchFilter_AddressAndTopicCombined(t *testing.T) {
	addr := types.HexToAddress("0xaaaa")
	topic := types.HexToHash("0x1111")

	log := &types.Log{
		Address: addr,
		Topics:  []types.Hash{types.HexToHash("0x9999")},
	}
	q := FilterQuery{
		Addresses: []types.Address{addr},
		Topics:    [][]types.Hash{{topic}},
	}
	if MatchFilter(log, q) {
		t.Fatal("should not match when topic doesn't match")
	}

	log2 := &types.Log{
		Address: types.HexToAddress("0xbbbb"),
		Topics:  []types.Hash{topic},
	}
	if MatchFilter(log2, q) {
		t.Fatal("should not match when address doesn't match")
	}
}

// ---------- FilterLogs edge cases ----------

func TestFilterLogs_EmptyInput(t *testing.T) {
	result := FilterLogs(nil, FilterQuery{})
	if result != nil {
		t.Fatalf("want nil for empty input, got %d logs", len(result))
	}

	result2 := FilterLogs([]*types.Log{}, FilterQuery{})
	if result2 != nil {
		t.Fatalf("want nil for empty slice, got %d logs", len(result2))
	}
}

func TestFilterLogs_AllMatch(t *testing.T) {
	logs := []*types.Log{
		{Address: types.HexToAddress("0xaaaa")},
		{Address: types.HexToAddress("0xbbbb")},
		{Address: types.HexToAddress("0xcccc")},
	}
	result := FilterLogs(logs, FilterQuery{})
	if len(result) != 3 {
		t.Fatalf("want 3 logs, got %d", len(result))
	}
}

func TestFilterLogs_NoneMatch(t *testing.T) {
	logs := []*types.Log{
		{Address: types.HexToAddress("0xaaaa")},
		{Address: types.HexToAddress("0xbbbb")},
	}
	q := FilterQuery{Addresses: []types.Address{types.HexToAddress("0xcccc")}}
	result := FilterLogs(logs, q)
	if result != nil {
		t.Fatalf("want nil when none match, got %d logs", len(result))
	}
}

func TestFilterLogs_TopicFilterOnly(t *testing.T) {
	topic1 := types.HexToHash("0x1111")
	topic2 := types.HexToHash("0x2222")
	logs := []*types.Log{
		{Topics: []types.Hash{topic1}},
		{Topics: []types.Hash{topic2}},
		{Topics: []types.Hash{topic1, topic2}},
	}
	q := FilterQuery{Topics: [][]types.Hash{{topic2}}}
	result := FilterLogs(logs, q)
	if len(result) != 1 {
		t.Fatalf("want 1 log matching topic2 at pos 0, got %d", len(result))
	}

	q2 := FilterQuery{Topics: [][]types.Hash{{topic1, topic2}}}
	result2 := FilterLogs(logs, q2)
	if len(result2) != 3 {
		t.Fatalf("want 3 logs matching topic1 OR topic2 at pos 0, got %d", len(result2))
	}
}

// ---------- FilterLogsWithBloom edge cases ----------

func TestFilterLogsWithBloom_NoMatch(t *testing.T) {
	addr := types.HexToAddress("0xaaaa")
	otherAddr := types.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

	log := &types.Log{Address: addr}
	bloom := types.LogsBloom([]*types.Log{log})

	q := FilterQuery{Addresses: []types.Address{otherAddr}}
	result := FilterLogsWithBloom(bloom, []*types.Log{log}, q)
	if len(result) != 0 {
		t.Fatalf("want 0 logs (address doesn't match), got %d", len(result))
	}
}

func TestFilterLogsWithBloom_EmptyBloom(t *testing.T) {
	bloom := types.Bloom{}
	logs := []*types.Log{
		{Address: types.HexToAddress("0xaaaa")},
	}
	result := FilterLogsWithBloom(bloom, logs, FilterQuery{})
	if len(result) != 1 {
		t.Fatalf("want 1 log with wildcard query, got %d", len(result))
	}
}

func TestFilterLogsWithBloom_TopicMatch(t *testing.T) {
	topic := crypto.Keccak256Hash([]byte("Event(uint256)"))
	addr := types.HexToAddress("0xaaaa")
	log := &types.Log{
		Address: addr,
		Topics:  []types.Hash{topic},
	}
	bloom := types.LogsBloom([]*types.Log{log})

	q := FilterQuery{
		Addresses: []types.Address{addr},
		Topics:    [][]types.Hash{{topic}},
	}
	result := FilterLogsWithBloom(bloom, []*types.Log{log}, q)
	if len(result) != 1 {
		t.Fatalf("want 1 log, got %d", len(result))
	}
}
