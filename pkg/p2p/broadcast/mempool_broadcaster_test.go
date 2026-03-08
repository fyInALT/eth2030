package broadcast

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/eth2030/eth2030/p2p/gossip"
)

func TestMempoolBroadcasterPublish(t *testing.T) {
	tm := gossip.NewTopicManager(gossip.DefaultTopicParams())
	defer tm.Close()

	var received []byte
	var receivedTopic gossip.GossipTopic
	handler := func(topic gossip.GossipTopic, msgID gossip.MessageID, data []byte) {
		receivedTopic = topic
		received = data
	}

	if err := tm.Subscribe(gossip.STARKMempoolTick, handler); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	b := NewMempoolBroadcaster(tm)
	data := []byte("stark mempool tick payload")
	if err := b.GossipMempoolStarkTick(data); err != nil {
		t.Fatalf("GossipMempoolStarkTick: %v", err)
	}

	if receivedTopic != gossip.STARKMempoolTick {
		t.Errorf("received topic = %v, want gossip.STARKMempoolTick", receivedTopic)
	}
	if string(received) != string(data) {
		t.Errorf("received data = %q, want %q", received, data)
	}
}

func TestMempoolBroadcasterPublishTooLarge(t *testing.T) {
	tm := gossip.NewTopicManager(gossip.DefaultTopicParams())
	defer tm.Close()

	handler := func(topic gossip.GossipTopic, msgID gossip.MessageID, data []byte) {}
	if err := tm.Subscribe(gossip.STARKMempoolTick, handler); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	b := NewMempoolBroadcaster(tm)
	bigData := make([]byte, 128*1024+1)
	bigData[0] = 0xff
	err := b.GossipMempoolStarkTick(bigData)
	if err != gossip.ErrTopicMsgTooLarge {
		t.Fatalf("expected gossip.ErrTopicMsgTooLarge, got %v", err)
	}
}

func TestMempoolBroadcasterPublishEmpty(t *testing.T) {
	tm := gossip.NewTopicManager(gossip.DefaultTopicParams())
	defer tm.Close()

	handler := func(topic gossip.GossipTopic, msgID gossip.MessageID, data []byte) {}
	if err := tm.Subscribe(gossip.STARKMempoolTick, handler); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	b := NewMempoolBroadcaster(tm)

	if err := b.GossipMempoolStarkTick(nil); err != gossip.ErrTopicEmptyData {
		t.Fatalf("nil data: expected gossip.ErrTopicEmptyData, got %v", err)
	}
	if err := b.GossipMempoolStarkTick([]byte{}); err != gossip.ErrTopicEmptyData {
		t.Fatalf("empty data: expected gossip.ErrTopicEmptyData, got %v", err)
	}
}

func TestMempoolBroadcasterRoundTrip(t *testing.T) {
	tm := gossip.NewTopicManager(gossip.DefaultTopicParams())
	defer tm.Close()

	var mu sync.Mutex
	var received [][]byte
	handler := func(topic gossip.GossipTopic, msgID gossip.MessageID, data []byte) {
		mu.Lock()
		defer mu.Unlock()
		cp := make([]byte, len(data))
		copy(cp, data)
		received = append(received, cp)
	}

	if err := tm.Subscribe(gossip.STARKMempoolTick, handler); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	b := NewMempoolBroadcaster(tm)
	ticks := [][]byte{
		[]byte("tick-0"),
		[]byte("tick-1"),
		[]byte("tick-2"),
	}

	for i, tick := range ticks {
		if err := b.GossipMempoolStarkTick(tick); err != nil {
			t.Fatalf("GossipMempoolStarkTick[%d]: %v", i, err)
		}
	}

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 3 {
		t.Fatalf("received %d ticks, want 3", len(received))
	}
	for i, tick := range ticks {
		if !bytes.Equal(received[i], tick) {
			t.Errorf("tick[%d] = %q, want %q", i, received[i], tick)
		}
	}
}

func TestMempoolBroadcasterNotSubscribed(t *testing.T) {
	tm := gossip.NewTopicManager(gossip.DefaultTopicParams())
	defer tm.Close()

	// Do not subscribe to gossip.STARKMempoolTick.
	b := NewMempoolBroadcaster(tm)
	err := b.GossipMempoolStarkTick([]byte("data"))
	if err != gossip.ErrTopicNotSubscribed {
		t.Fatalf("expected gossip.ErrTopicNotSubscribed, got %v", err)
	}

	// Verify no panic occurred — reaching here means no panic.
	_ = fmt.Sprintf("no panic for unsubscribed publish")
}
