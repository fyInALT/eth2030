package node

// nodeNetBackend adapts the Node to the netapi.Backend interface.
type nodeNetBackend struct {
	node *Node
}

func newNodeNetBackend(n *Node) *nodeNetBackend {
	return &nodeNetBackend{node: n}
}

func (b *nodeNetBackend) NetworkID() uint64 {
	if cfg := b.node.blockchain.Config(); cfg != nil && cfg.ChainID != nil {
		return cfg.ChainID.Uint64()
	}
	return 0
}

func (b *nodeNetBackend) IsListening() bool {
	return b.node.p2pServer != nil
}

func (b *nodeNetBackend) PeerCount() int {
	if b.node.p2pServer == nil {
		return 0
	}
	return b.node.p2pServer.PeerCount()
}

func (b *nodeNetBackend) MaxPeers() int {
	return b.node.config.MaxPeers
}