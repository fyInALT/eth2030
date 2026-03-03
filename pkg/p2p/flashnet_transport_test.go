package p2p

import (
	"testing"
	"time"
)

func TestFlashnetTransport_Name(t *testing.T) {
	ft := NewFlashnetTransport(nil)
	if ft.Name() != "flashnet" {
		t.Fatalf("expected 'flashnet', got %q", ft.Name())
	}
}

func TestFlashnetTransport_Submit_NotStarted(t *testing.T) {
	ft := NewFlashnetTransport(nil)
	tx := testTx()
	if err := ft.Submit(tx); err != ErrAnonTransportClosed {
		t.Fatalf("expected ErrAnonTransportClosed, got %v", err)
	}
}

func TestFlashnetTransport_Submit_NilTx(t *testing.T) {
	ft := NewFlashnetTransport(nil)
	_ = ft.Start()
	if err := ft.Submit(nil); err != ErrAnonTransportNilTx {
		t.Fatalf("expected ErrAnonTransportNilTx, got %v", err)
	}
}

func TestFlashnetTransport_SubmitAndReceive(t *testing.T) {
	ft := NewFlashnetTransport(&FlashnetConfig{
		MaxPeers:   8,
		MaxPending: 16,
	})
	_ = ft.Start()
	defer ft.Stop()

	tx := testTx()
	if err := ft.Submit(tx); err != nil {
		t.Fatalf("submit error: %v", err)
	}

	// Flashnet delivery is immediate (no delay).
	select {
	case got := <-ft.Receive():
		if got != tx {
			t.Fatal("received different transaction")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for transaction")
	}
}

func TestFlashnetTransport_Stats(t *testing.T) {
	ft := NewFlashnetTransport(&FlashnetConfig{
		MaxPeers:   8,
		MaxPending: 16,
	})
	_ = ft.Start()
	defer ft.Stop()

	tx := testTx()
	_ = ft.Submit(tx)
	_ = ft.Submit(tx)
	// Drain channel.
	<-ft.Receive()
	<-ft.Receive()
	_ = ft.Submit(tx)
	<-ft.Receive()

	sub, bcast := ft.Stats()
	if sub != 3 {
		t.Fatalf("expected 3 submitted, got %d", sub)
	}
	if bcast != 3 {
		t.Fatalf("expected 3 broadcast, got %d", bcast)
	}
}

func TestFlashnetTransport_Stop(t *testing.T) {
	ft := NewFlashnetTransport(nil)
	_ = ft.Start()
	_ = ft.Stop()

	tx := testTx()
	if err := ft.Submit(tx); err != ErrAnonTransportClosed {
		t.Fatalf("expected ErrAnonTransportClosed after stop, got %v", err)
	}
}

func TestFlashnetTransport_InterfaceCompliance(t *testing.T) {
	var _ AnonymousTransport = (*FlashnetTransport)(nil)
}
