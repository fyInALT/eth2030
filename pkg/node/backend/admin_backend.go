package backend

import (
	"fmt"
	"net"

	"github.com/eth2030/eth2030/rpc"
)

// AdminBackend adapts NodeDeps to the rpc.AdminBackend interface.
type AdminBackend struct {
	node NodeDeps
}

// NewAdminBackend creates a new AdminBackend.
func NewAdminBackend(node NodeDeps) rpc.AdminBackend {
	return &AdminBackend{node: node}
}

func (b *AdminBackend) NodeInfo() rpc.NodeInfoData {
	p2pSrv := b.node.P2PServer()
	if p2pSrv == nil {
		return rpc.NodeInfoData{Name: "eth2030"}
	}

	nodeID := p2pSrv.LocalID()

	cfg := b.node.Config()
	port := 0
	if cfg != nil {
		port = cfg.P2PPort
	}

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
	if bc := b.node.Blockchain(); bc != nil {
		if chainCfg := bc.Config(); chainCfg != nil && chainCfg.ChainID != nil {
			chainID = chainCfg.ChainID.Uint64()
		}
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

func (b *AdminBackend) Peers() []rpc.PeerInfoData {
	p2pSrv := b.node.P2PServer()
	if p2pSrv == nil {
		return nil
	}

	peers := p2pSrv.PeersList()
	infos := make([]rpc.PeerInfoData, len(peers))
	for i, p := range peers {
		caps := make([]string, 0, len(p.Caps()))
		for _, c := range p.Caps() {
			caps = append(caps, fmt.Sprintf("%s/%d", c.Name(), c.Version()))
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

func (b *AdminBackend) AddPeer(url string) error {
	p2pSrv := b.node.P2PServer()
	if p2pSrv == nil {
		return fmt.Errorf("p2p server not available")
	}
	return p2pSrv.AddPeer(url)
}

func (b *AdminBackend) RemovePeer(_ string) error {
	return nil
}

func (b *AdminBackend) ChainID() uint64 {
	if bc := b.node.Blockchain(); bc != nil {
		if chainCfg := bc.Config(); chainCfg != nil && chainCfg.ChainID != nil {
			return chainCfg.ChainID.Uint64()
		}
	}
	return 0
}

func (b *AdminBackend) DataDir() string {
	cfg := b.node.Config()
	if cfg != nil {
		return cfg.DataDir
	}
	return ""
}