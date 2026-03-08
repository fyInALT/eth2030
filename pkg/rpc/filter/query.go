// Package rpcfilter provides log filter types, filter matching logic, and
// filter management systems for the Ethereum JSON-RPC API.
package rpcfilter

import (
	"time"

	"github.com/eth2030/eth2030/core/types"
)

// FilterType distinguishes the kind of installed filter.
type FilterType int

const (
	// LogFilter watches for contract log events matching given criteria.
	LogFilter FilterType = iota
	// BlockFilter watches for new block hashes.
	BlockFilter
	// PendingTxFilter watches for new pending transaction hashes.
	PendingTxFilter
)

// FilterTimeout is the default timeout for idle filters.
const FilterTimeout = 5 * time.Minute

// FilterQuery specifies criteria for log matching. Addresses are OR-ed
// (match any listed address). Topics follow the Ethereum convention: AND
// across positions, OR within each position. An empty (or nil) position
// is a wildcard that matches any topic value.
type FilterQuery struct {
	FromBlock *uint64
	ToBlock   *uint64
	Addresses []types.Address
	Topics    [][]types.Hash
}

// MatchFilter tests whether a log matches the given query criteria.
func MatchFilter(log *types.Log, query FilterQuery) bool {
	if len(query.Addresses) > 0 {
		found := false
		for _, addr := range query.Addresses {
			if log.Address == addr {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for i, topicSet := range query.Topics {
		if len(topicSet) == 0 {
			continue // wildcard
		}
		if i >= len(log.Topics) {
			return false
		}
		matched := false
		for _, t := range topicSet {
			if log.Topics[i] == t {
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

// BloomMatchesFilterQuery returns true if the bloom filter may match any address
// or topic in the query (false positives possible).
func BloomMatchesFilterQuery(bloom types.Bloom, query FilterQuery) bool {
	if len(query.Addresses) > 0 {
		any := false
		for _, addr := range query.Addresses {
			if types.BloomContains(bloom, addr.Bytes()) {
				any = true
				break
			}
		}
		if !any {
			return false
		}
	}

	for _, topicSet := range query.Topics {
		if len(topicSet) == 0 {
			continue
		}
		any := false
		for _, topic := range topicSet {
			if types.BloomContains(bloom, topic.Bytes()) {
				any = true
				break
			}
		}
		if !any {
			return false
		}
	}
	return true
}

// FilterLogs applies a FilterQuery against a set of logs and returns
// only the matching entries.
func FilterLogs(logs []*types.Log, query FilterQuery) []*types.Log {
	var result []*types.Log
	for _, log := range logs {
		if MatchFilter(log, query) {
			result = append(result, log)
		}
	}
	return result
}

// FilterLogsWithBloom applies bloom-level pre-screening per block before
// falling back to exact log matching.
func FilterLogsWithBloom(bloom types.Bloom, logs []*types.Log, query FilterQuery) []*types.Log {
	if !BloomMatchesFilterQuery(bloom, query) {
		return nil
	}
	return FilterLogs(logs, query)
}
