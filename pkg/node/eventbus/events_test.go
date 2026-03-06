package eventbus

import (
	"sync"
	"testing"
	"time"
)

func TestSubscribeAndPublish(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	sub := bus.Subscribe(EventNewBlock)

	bus.Publish(EventNewBlock, "block-1")

	select {
	case ev := <-sub.Chan():
		if ev.Type != EventNewBlock {
			t.Errorf("event type = %s, want %s", ev.Type, EventNewBlock)
		}
		if ev.Data != "block-1" {
			t.Errorf("event data = %v, want block-1", ev.Data)
		}
		if ev.Timestamp.IsZero() {
			t.Error("event timestamp should not be zero")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	sub := bus.Subscribe(EventNewTx)
	bus.Unsubscribe(sub)

	_, ok := <-sub.Chan()
	if ok {
		t.Error("expected channel to be closed after unsubscribe")
	}

	bus.Unsubscribe(sub)
	sub.Unsubscribe()
}

func TestEventTypeFiltering(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	blockSub := bus.Subscribe(EventNewBlock)
	txSub := bus.Subscribe(EventNewTx)

	bus.Publish(EventNewBlock, "block-data")
	bus.Publish(EventNewTx, "tx-data")

	select {
	case ev := <-blockSub.Chan():
		if ev.Type != EventNewBlock {
			t.Errorf("block sub got type %s", ev.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for block event")
	}

	select {
	case ev := <-txSub.Chan():
		if ev.Type != EventNewTx {
			t.Errorf("tx sub got type %s", ev.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for tx event")
	}
}

func TestSubscriberCount(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	if count := bus.SubscriberCount(EventNewBlock); count != 0 {
		t.Errorf("initial count = %d, want 0", count)
	}

	sub1 := bus.Subscribe(EventNewBlock)
	sub2 := bus.Subscribe(EventNewBlock)

	if count := bus.SubscriberCount(EventNewBlock); count != 2 {
		t.Errorf("count after 2 subs = %d, want 2", count)
	}

	bus.Unsubscribe(sub1)
	bus.Unsubscribe(sub2)
	if count := bus.SubscriberCount(EventNewBlock); count != 0 {
		t.Errorf("count after both unsub = %d, want 0", count)
	}
}

func TestCloseBus(t *testing.T) {
	bus := NewEventBus(10)

	sub1 := bus.Subscribe(EventNewBlock)
	sub2 := bus.Subscribe(EventNewTx)

	bus.Close()

	for _, sub := range []*Subscription{sub1, sub2} {
		_, ok := <-sub.Chan()
		if ok {
			t.Error("expected channel to be closed after bus.Close()")
		}
	}

	bus.Publish(EventNewBlock, "late-event")
	bus.PublishAsync(EventNewBlock, "late-async")

	lateSub := bus.Subscribe(EventNewBlock)
	_, ok := <-lateSub.Chan()
	if ok {
		t.Error("expected late subscription channel to be closed")
	}

	bus.Close()
}

func TestConcurrentAccess(t *testing.T) {
	bus := NewEventBus(100)
	defer bus.Close()

	const (
		numPublishers  = 10
		numSubscribers = 10
		numEvents      = 50
	)

	var wg sync.WaitGroup

	subs := make([]*Subscription, numSubscribers)
	for i := 0; i < numSubscribers; i++ {
		subs[i] = bus.Subscribe(EventNewBlock)
	}

	for i := 0; i < numSubscribers; i++ {
		wg.Add(1)
		go func(sub *Subscription) {
			defer wg.Done()
			count := 0
			for range sub.Chan() {
				count++
				if count >= numPublishers*numEvents {
					return
				}
			}
		}(subs[i])
	}

	for i := 0; i < numPublishers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numEvents; j++ {
				bus.Publish(EventNewBlock, id*1000+j)
			}
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		for _, sub := range subs {
			bus.Unsubscribe(sub)
		}
		t.Fatal("timed out waiting for concurrent operations")
	}
}

func TestEventConstants(t *testing.T) {
	allTypes := []EventType{
		EventNewBlock, EventNewTx, EventChainHead, EventChainSideHead,
		EventNewPeer, EventDropPeer,
		EventSyncStarted, EventSyncCompleted,
		EventTxPoolAdd, EventTxPoolDrop,
	}

	seen := make(map[EventType]bool)
	for _, et := range allTypes {
		if seen[et] {
			t.Errorf("duplicate event type: %s", et)
		}
		seen[et] = true
		if et == "" {
			t.Error("event type should not be empty")
		}
	}
}
