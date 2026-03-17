// Package backend provides backend implementations for RPC and Engine APIs.
// This implementation abstracts Node dependencies for testability.
package backend

import (
	"math/big"
	"net"

	"github.com/eth2030/eth2030/core/chain"
	"github.com/eth2030/eth2030/core/mev"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/txpool"
)

// NodeDeps provides access to all node dependencies for backend implementations.
// This interface abstracts the Node to enable testing with mocks.
// Only methods actually used by backend implementations are included.
type NodeDeps interface {
	// Core (frequently accessed)
	Blockchain() *chain.Blockchain
	TxPool() *txpool.TxPool

	// Config
	Config() *Config

	// Optional dependencies (return nil if not available)
	GasOracle() any                      // For recording block gas prices
	MEVConfig() *mev.MEVProtectionConfig // For MEV protection in tx pool
	FCStateManager() any                 // For forkchoice state tracking
	StarkFrameProver() any               // For STARK proof generation
	EthHandler() any                     // For transaction broadcast
	TxJournal() any                      // For transaction journaling

	// P2P
	P2PServer() P2PServerDeps
}

// Config holds backend configuration.
type Config struct {
	CacheEnginePayloads int
	SnapshotCapDepth    int
	MigrateEveryBlocks  int
	MaxPeers            int
	P2PPort             int
	DataDir             string
}

// P2PServerDeps provides P2P server access.
type P2PServerDeps interface {
	LocalID() string
	ListenAddr() net.Addr
	ExternalIP() net.IP
	PeersList() []P2PPeerDeps
	AddPeer(url string) error
	PeerCount() int
}

// P2PPeerDeps provides peer info access.
type P2PPeerDeps interface {
	ID() string
	RemoteAddr() string
	Caps() []P2PCapDeps
}

// P2PCapDeps represents a protocol capability.
type P2PCapDeps interface {
	Name() string
	Version() int
}

// ExtractBlockTips returns the effective priority fee for each transaction.
func ExtractBlockTips(txs []*types.Transaction, baseFee *big.Int) []*big.Int {
	tips := make([]*big.Int, 0, len(txs))
	if baseFee == nil {
		baseFee = new(big.Int)
	}
	for _, tx := range txs {
		var tip *big.Int
		switch tx.Type() {
		case types.DynamicFeeTxType:
			tipCap := tx.GasTipCap()
			feeCap := tx.GasFeeCap()
			if tipCap == nil || feeCap == nil {
				continue
			}
			effective := new(big.Int).Sub(feeCap, baseFee)
			if effective.Sign() < 0 {
				continue
			}
			tip = tipCap
			if effective.Cmp(tipCap) < 0 {
				tip = effective
			}
		default:
			gp := tx.GasPrice()
			if gp == nil {
				continue
			}
			tip = new(big.Int).Sub(gp, baseFee)
			if tip.Sign() < 0 {
				continue
			}
		}
		tips = append(tips, new(big.Int).Set(tip))
	}
	return tips
}
