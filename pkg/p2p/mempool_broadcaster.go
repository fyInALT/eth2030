package p2p

// MempoolBroadcaster implements the txpool.P2PBroadcaster interface using
// a TopicManager. It propagates serialised MempoolAggregationTick payloads
// over the STARKMempoolTick gossip topic (128 KB per-message budget).
type MempoolBroadcaster struct {
	tm *TopicManager
}

// NewMempoolBroadcaster creates a MempoolBroadcaster backed by tm.
func NewMempoolBroadcaster(tm *TopicManager) *MempoolBroadcaster {
	return &MempoolBroadcaster{tm: tm}
}

// GossipMempoolStarkTick publishes a serialised STARK mempool tick to all
// peers subscribed to the STARKMempoolTick topic.
func (b *MempoolBroadcaster) GossipMempoolStarkTick(data []byte) error {
	return b.tm.Publish(STARKMempoolTick, data)
}
