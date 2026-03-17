package node

import (
	"net"

	"github.com/eth2030/eth2030/core/chain"
	"github.com/eth2030/eth2030/core/mev"
	"github.com/eth2030/eth2030/node/backend"
	"github.com/eth2030/eth2030/p2p"
	"github.com/eth2030/eth2030/p2p/peermgr"
	"github.com/eth2030/eth2030/p2p/wire"
	"github.com/eth2030/eth2030/proofs"
	"github.com/eth2030/eth2030/txpool"
)

// nodeDepsAdapter wraps *Node to implement backend.NodeDeps.
type nodeDepsAdapter struct {
	n *Node
}

// toNodeDeps creates a backend.NodeDeps adapter from a Node.
func toNodeDeps(n *Node) backend.NodeDeps {
	return &nodeDepsAdapter{n: n}
}

func (a *nodeDepsAdapter) Blockchain() *chain.Blockchain {
	return a.n.blockchain
}

func (a *nodeDepsAdapter) TxPool() *txpool.TxPool {
	return a.n.txPool
}

func (a *nodeDepsAdapter) Config() *backend.Config {
	return &backend.Config{
		CacheEnginePayloads: a.n.config.CacheEnginePayloads,
		SnapshotCapDepth:    a.n.config.SnapshotCapDepth,
		MigrateEveryBlocks:  a.n.config.MigrateEveryBlocks,
		MaxPeers:            a.n.config.MaxPeers,
		P2PPort:             a.n.config.P2PPort,
		DataDir:             a.n.config.DataDir,
	}
}

func (a *nodeDepsAdapter) GasOracle() backend.GasOracleDeps {
	if a.n.gasOracle == nil {
		return nil
	}
	return a.n.gasOracle
}

func (a *nodeDepsAdapter) MEVConfig() *mev.MEVProtectionConfig { return a.n.mevConfig }

func (a *nodeDepsAdapter) FCStateManager() backend.FCStateManagerDeps {
	if a.n.fcStateManager == nil {
		return nil
	}
	return a.n.fcStateManager
}

func (a *nodeDepsAdapter) StarkFrameProver() proofs.ValidationFrameProver {
	return a.n.starkFrameProver
}

func (a *nodeDepsAdapter) EthHandler() backend.EthHandlerDeps {
	if a.n.ethHandler == nil {
		return nil
	}
	return a.n.ethHandler
}

func (a *nodeDepsAdapter) TxJournal() backend.TxJournalDeps {
	if a.n.txJournal == nil {
		return nil
	}
	return a.n.txJournal
}

func (a *nodeDepsAdapter) P2PServer() backend.P2PServerDeps {
	if a.n.p2pServer == nil {
		return nil
	}
	return &p2pServerAdapter{srv: a.n.p2pServer}
}

// p2pServerAdapter wraps *p2p.Server to implement backend.P2PServerDeps.
type p2pServerAdapter struct {
	srv *p2p.Server
}

func (a *p2pServerAdapter) LocalID() string {
	return a.srv.LocalID()
}

func (a *p2pServerAdapter) ListenAddr() net.Addr {
	return a.srv.ListenAddr()
}

func (a *p2pServerAdapter) ExternalIP() net.IP {
	return a.srv.ExternalIP()
}

func (a *p2pServerAdapter) PeersList() []backend.P2PPeerDeps {
	peers := a.srv.PeersList()
	result := make([]backend.P2PPeerDeps, len(peers))
	for i, p := range peers {
		result[i] = &p2pPeerAdapter{peer: p}
	}
	return result
}

func (a *p2pServerAdapter) AddPeer(url string) error {
	return a.srv.AddPeer(url)
}

func (a *p2pServerAdapter) PeerCount() int {
	return a.srv.PeerCount()
}

// p2pPeerAdapter wraps *peermgr.Peer to implement backend.P2PPeerDeps.
type p2pPeerAdapter struct {
	peer *peermgr.Peer
}

func (a *p2pPeerAdapter) ID() string {
	return a.peer.ID()
}

func (a *p2pPeerAdapter) RemoteAddr() string {
	return a.peer.RemoteAddr()
}

func (a *p2pPeerAdapter) Caps() []backend.P2PCapDeps {
	caps := a.peer.Caps()
	result := make([]backend.P2PCapDeps, len(caps))
	for i, c := range caps {
		result[i] = &p2pCapAdapter{cap: c}
	}
	return result
}

// p2pCapAdapter wraps wire.Cap to implement backend.P2PCapDeps.
type p2pCapAdapter struct {
	cap wire.Cap
}

func (a *p2pCapAdapter) Name() string {
	return a.cap.Name
}

func (a *p2pCapAdapter) Version() int {
	return int(a.cap.Version)
}