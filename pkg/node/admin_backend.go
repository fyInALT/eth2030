package node

import (
	"fmt"
	"net"

	"github.com/eth2030/eth2030/rpc"
)

// nodeAdminBackend adapts the Node to the rpc.AdminBackend interface.
type nodeAdminBackend struct {
	node *Node
}

func newNodeAdminBackend(n *Node) rpc.AdminBackend {
	return &nodeAdminBackend{node: n}
}

func (b *nodeAdminBackend) NodeInfo() rpc.NodeInfoData {
	p2pSrv := b.node.p2pServer
	nodeID := p2pSrv.LocalID()

	port := b.node.config.P2PPort

	listenAddr := ""
	if addr := p2pSrv.ListenAddr(); addr != nil {
		listenAddr = addr.String()
	}

	ip := ""
	if extIP := p2pSrv.ExternalIP(); extIP != nil {
		ip = extIP.String()
	} else if listenAddr != "" {
		host, _, err := net.SplitHostPort(listenAddr)
		if err == nil && host != "::" && host != "" {
			ip = host
		}
	}

	enode := fmt.Sprintf("enode://%s@%s:%d", nodeID, ip, port)

	chainID := uint64(0)
	if cfg := b.node.blockchain.Config(); cfg != nil && cfg.ChainID != nil {
		chainID = cfg.ChainID.Uint64()
	}

	return rpc.NodeInfoData{
		Name:       "eth2030",
		ID:         nodeID,
		ENR:        "",
		Enode:      enode,
		IP:         ip,
		ListenAddr: listenAddr,
		Ports: rpc.NodePorts{
			Discovery: port,
			Listener:  port,
		},
		Protocols: map[string]interface{}{
			"eth": map[string]interface{}{
				"network": chainID,
				"genesis": "",
			},
		},
	}
}

func (b *nodeAdminBackend) Peers() []rpc.PeerInfoData {
	peers := b.node.p2pServer.PeersList()
	infos := make([]rpc.PeerInfoData, len(peers))
	for i, p := range peers {
		caps := make([]string, 0, len(p.Caps()))
		for _, c := range p.Caps() {
			caps = append(caps, fmt.Sprintf("%s/%d", c.Name, c.Version))
		}
		infos[i] = rpc.PeerInfoData{
			ID:         p.ID(),
			Name:       "",
			RemoteAddr: p.RemoteAddr(),
			Caps:       caps,
		}
	}
	return infos
}

func (b *nodeAdminBackend) AddPeer(url string) error {
	return b.node.p2pServer.AddPeer(url)
}

func (b *nodeAdminBackend) RemovePeer(_ string) error {
	return nil
}

func (b *nodeAdminBackend) ChainID() uint64 {
	if cfg := b.node.blockchain.Config(); cfg != nil && cfg.ChainID != nil {
		return cfg.ChainID.Uint64()
	}
	return 0
}

func (b *nodeAdminBackend) DataDir() string {
	return b.node.config.DataDir
}