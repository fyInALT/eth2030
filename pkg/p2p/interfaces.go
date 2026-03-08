package p2p

// NodeDiscovery provides access to the node table for protocol-level
// peer management (e.g., requesting specific peers for sync).
type NodeDiscovery interface {
	// AllNodes returns all known nodes.
	AllNodes() []*Node

	// StaticNodes returns permanently configured nodes.
	StaticNodes() []*Node

	// AddNode adds a discovered node to the table.
	AddNode(n *Node) error

	// Remove removes a node from the table.
	Remove(id NodeID)
}

// Verify interface compliance at compile time.
var _ NodeDiscovery = (*NodeTable)(nil)
