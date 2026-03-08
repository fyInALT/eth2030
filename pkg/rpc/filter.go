// filter.go re-exports filter functions from rpc/filter for backward
// compatibility. FilterLogs and FilterLogsWithBloom delegate to the
// rpcfilter sub-package implementation.
package rpc

import (
	"github.com/eth2030/eth2030/core/types"
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
)

// FilterLogs applies a FilterQuery against a set of logs and returns
// only the matching entries.
func FilterLogs(logs []*types.Log, query FilterQuery) []*types.Log {
	return rpcfilter.FilterLogs(logs, query)
}

// FilterLogsWithBloom applies bloom-level pre-screening per block before
// falling back to exact log matching.
func FilterLogsWithBloom(bloom types.Bloom, logs []*types.Log, query FilterQuery) []*types.Log {
	return rpcfilter.FilterLogsWithBloom(bloom, logs, query)
}

// bloomMatchesAddress returns true if the bloom may contain the given address.
func bloomMatchesAddress(bloom types.Bloom, addr types.Address) bool {
	return types.BloomContains(bloom, addr.Bytes())
}

// bloomMatchesTopic returns true if the bloom may contain the given topic hash.
func bloomMatchesTopic(bloom types.Bloom, topic types.Hash) bool {
	return types.BloomContains(bloom, topic.Bytes())
}
