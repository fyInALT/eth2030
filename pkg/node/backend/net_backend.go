package backend

// NetBackend adapts NodeDeps to the netapi.Backend interface.
type NetBackend struct {
	node NodeDeps
}

// NewNetBackend creates a new NetBackend.
func NewNetBackend(node NodeDeps) *NetBackend {
	return &NetBackend{node: node}
}

func (b *NetBackend) NetworkID() uint64 {
	if bc := b.node.Blockchain(); bc != nil {
		if chainCfg := bc.Config(); chainCfg != nil && chainCfg.ChainID != nil {
			return chainCfg.ChainID.Uint64()
		}
	}
	return 0
}

func (b *NetBackend) IsListening() bool {
	return b.node.P2PServer() != nil
}

func (b *NetBackend) PeerCount() int {
	p2pSrv := b.node.P2PServer()
	if p2pSrv == nil {
		return 0
	}
	return p2pSrv.PeerCount()
}

func (b *NetBackend) MaxPeers() int {
	cfg := b.node.Config()
	if cfg != nil {
		return cfg.MaxPeers
	}
	return 0
}