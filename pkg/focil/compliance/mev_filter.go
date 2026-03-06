package compliance

import (
	"sync"

	"github.com/eth2030/eth2030/core/types"
)

// MEVFilterConfig configures MEV transaction detection.
type MEVFilterConfig struct {
	// KnownDEXContracts is the set of known DEX contract addresses.
	KnownDEXContracts map[types.Address]bool
	// KnownLiquidationContracts is the set of known liquidation contracts.
	KnownLiquidationContracts map[types.Address]bool
	// MinGasPriceMultiplier flags txs with gas price N times above average as MEV.
	MinGasPriceMultiplier uint64
}

// DefaultMEVFilterConfig returns a config with common DEX contract stubs.
func DefaultMEVFilterConfig() *MEVFilterConfig {
	return &MEVFilterConfig{
		KnownDEXContracts:         make(map[types.Address]bool),
		KnownLiquidationContracts: make(map[types.Address]bool),
		MinGasPriceMultiplier:     5,
	}
}

// MEVFilter detects and filters MEV transactions.
type MEVFilter struct {
	mu     sync.RWMutex
	config *MEVFilterConfig
}

// NewMEVFilter creates a new MEV filter with the given config.
func NewMEVFilter(config *MEVFilterConfig) *MEVFilter {
	if config == nil {
		config = DefaultMEVFilterConfig()
	}
	return &MEVFilter{config: config}
}

// AddDEXContract registers a known DEX contract address.
func (f *MEVFilter) AddDEXContract(addr types.Address) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.config.KnownDEXContracts[addr] = true
}

// AddLiquidationContract registers a known liquidation contract address.
func (f *MEVFilter) AddLiquidationContract(addr types.Address) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.config.KnownLiquidationContracts[addr] = true
}

// IsMEVTransaction returns true if the transaction is likely MEV-extracting.
// Heuristic: tx targets a known DEX or liquidation contract.
func (f *MEVFilter) IsMEVTransaction(tx *types.Transaction) bool {
	if tx == nil {
		return false
	}
	to := tx.To()
	if to == nil {
		return false // contract creation is not MEV
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.config.KnownDEXContracts[*to] {
		return true
	}
	if f.config.KnownLiquidationContracts[*to] {
		return true
	}
	return false
}

// FilterMEVOnly returns only transactions that are MEV-extracting.
func (f *MEVFilter) FilterMEVOnly(txs []*types.Transaction) []*types.Transaction {
	var result []*types.Transaction
	for _, tx := range txs {
		if f.IsMEVTransaction(tx) {
			result = append(result, tx)
		}
	}
	return result
}

// FilterNonMEV returns only transactions that are NOT MEV-extracting.
func (f *MEVFilter) FilterNonMEV(txs []*types.Transaction) []*types.Transaction {
	var result []*types.Transaction
	for _, tx := range txs {
		if !f.IsMEVTransaction(tx) {
			result = append(result, tx)
		}
	}
	return result
}

// ValidateBuilderCompliance checks that a block's transactions satisfy Big FOCIL
// constraints: the builder must include all FOCIL-mandated txs, plus only MEV txs.
// When mevOnly is true, non-FOCIL non-MEV txs in the block are a violation.
func ValidateBuilderCompliance(blockTxHashes []types.Hash, focilTxHashes []types.Hash, builderTxs []*types.Transaction, mevOnly bool, filter *MEVFilter) *BuilderComplianceResult {
	result := &BuilderComplianceResult{
		Compliant: true,
	}

	// Check all FOCIL txs are included.
	blockSet := make(map[types.Hash]bool, len(blockTxHashes))
	for _, h := range blockTxHashes {
		blockSet[h] = true
	}

	for _, h := range focilTxHashes {
		if !blockSet[h] {
			result.MissingFOCILTxs = append(result.MissingFOCILTxs, h)
			result.Compliant = false
		}
	}

	// If MEV-only mode, check builder didn't include non-MEV non-FOCIL txs.
	if mevOnly && filter != nil {
		focilSet := make(map[types.Hash]bool, len(focilTxHashes))
		for _, h := range focilTxHashes {
			focilSet[h] = true
		}

		for _, tx := range builderTxs {
			h := tx.Hash()
			if focilSet[h] {
				continue // FOCIL tx, allowed
			}
			if !filter.IsMEVTransaction(tx) {
				result.NonMEVTxs = append(result.NonMEVTxs, h)
				result.Compliant = false
			}
		}
	}

	return result
}

// BuilderComplianceResult holds the result of builder compliance validation.
type BuilderComplianceResult struct {
	Compliant       bool
	MissingFOCILTxs []types.Hash
	NonMEVTxs       []types.Hash
}
