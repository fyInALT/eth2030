package p2p

import (
	"bytes"
	"testing"
	"time"
)

func TestMixnetTransport_Name(t *testing.T) {
	mt := NewMixnetTransport(nil)
	if mt.Name() != "mixnet" {
		t.Fatalf("expected 'mixnet', got %q", mt.Name())
	}
}

func TestMixnetTransport_Submit_NotStarted(t *testing.T) {
	mt := NewMixnetTransport(nil)
	tx := testTx()
	if err := mt.Submit(tx); err != ErrAnonTransportClosed {
		t.Fatalf("expected ErrAnonTransportClosed, got %v", err)
	}
}

func TestMixnetTransport_Submit_NilTx(t *testing.T) {
	mt := NewMixnetTransport(nil)
	_ = mt.Start()
	if err := mt.Submit(nil); err != ErrAnonTransportNilTx {
		t.Fatalf("expected ErrAnonTransportNilTx, got %v", err)
	}
}

func TestMixnetTransport_SubmitAndReceive(t *testing.T) {
	cfg := &MixnetConfig{
		Hops:       1,
		HopDelay:   10 * time.Millisecond,
		MaxPending: 16,
	}
	mt := NewMixnetTransport(cfg)
	_ = mt.Start()
	defer mt.Stop()

	tx := testTx()
	if err := mt.Submit(tx); err != nil {
		t.Fatalf("submit error: %v", err)
	}

	select {
	case got := <-mt.Receive():
		if got != tx {
			t.Fatal("received different transaction")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for transaction")
	}
}

func TestMixnetTransport_MultiHopDelay(t *testing.T) {
	hops := 2
	hopDelay := 20 * time.Millisecond
	cfg := &MixnetConfig{
		Hops:       hops,
		HopDelay:   hopDelay,
		MaxPending: 16,
	}
	mt := NewMixnetTransport(cfg)
	_ = mt.Start()
	defer mt.Stop()

	tx := testTx()
	start := time.Now()
	if err := mt.Submit(tx); err != nil {
		t.Fatalf("submit error: %v", err)
	}

	select {
	case <-mt.Receive():
		elapsed := time.Since(start)
		minExpected := time.Duration(hops) * hopDelay
		if elapsed < minExpected/2 {
			t.Fatalf("received too fast: %v (expected at least ~%v)", elapsed, minExpected)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for transaction")
	}
}

func TestMixnetTransport_Stop(t *testing.T) {
	cfg := &MixnetConfig{
		Hops:       5,
		HopDelay:   100 * time.Millisecond,
		MaxPending: 16,
	}
	mt := NewMixnetTransport(cfg)
	_ = mt.Start()

	tx := testTx()
	_ = mt.Submit(tx)
	// Stop immediately; the relay goroutine should exit.
	_ = mt.Stop()

	// Further submissions should fail.
	if err := mt.Submit(tx); err != ErrAnonTransportClosed {
		t.Fatalf("expected ErrAnonTransportClosed after stop, got %v", err)
	}
}

func TestWrapOnion(t *testing.T) {
	data := []byte("hello world")

	w1 := WrapOnion(data, 1)
	w2 := WrapOnion(data, 2)
	w3 := WrapOnion(data, 3)

	// Each layer count should produce different output.
	if bytes.Equal(w1, w2) {
		t.Fatal("1-hop and 2-hop should differ")
	}
	if bytes.Equal(w2, w3) {
		t.Fatal("2-hop and 3-hop should differ")
	}
	if bytes.Equal(w1, data) {
		t.Fatal("wrapped should differ from original")
	}

	// Same input and hops should be deterministic.
	w1b := WrapOnion(data, 1)
	if !bytes.Equal(w1, w1b) {
		t.Fatal("WrapOnion should be deterministic")
	}
}

func TestMixnetTransport_InterfaceCompliance(t *testing.T) {
	var _ AnonymousTransport = (*MixnetTransport)(nil)
}
