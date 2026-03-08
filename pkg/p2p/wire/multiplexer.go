package wire

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var (
	// ErrProtocolNotFound is returned when a message code does not match any protocol.
	ErrProtocolNotFound = errors.New("p2p/wire: protocol not found for message code")

	// ErrMuxClosed is returned when the multiplexer has been shut down.
	ErrMuxClosed = errors.New("p2p/wire: multiplexer closed")
)

// ProtoRW is a read-write interface scoped to a single sub-protocol's message
// code range. It offsets message codes so each protocol sees codes starting at 0.
type ProtoRW struct {
	Proto  Protocol
	Offset uint64 // Code offset for this protocol in the multiplexed stream.
	In     chan Msg
	Closed chan struct{}
}

// ReadMsg reads the next message destined for this protocol.
func (rw *ProtoRW) ReadMsg() (Msg, error) {
	select {
	case msg, ok := <-rw.In:
		if !ok {
			return Msg{}, ErrMuxClosed
		}
		return msg, nil
	case <-rw.Closed:
		return Msg{}, ErrMuxClosed
	}
}

// Multiplexer manages multiple sub-protocols over a single transport connection.
type Multiplexer struct {
	transport Transport
	protos    []*ProtoRW
	totalLen  uint64

	mu     sync.Mutex
	closed bool
	done   chan struct{}
	wmu    sync.Mutex
}

// ProtoMatch describes a protocol and its assigned offset.
type ProtoMatch struct {
	Proto  Protocol
	Offset uint64
}

// NewMultiplexer creates a multiplexer for the given protocols over a transport.
// Protocols are sorted by (Name, Version) and assigned contiguous code ranges.
func NewMultiplexer(tr Transport, protocols []Protocol) *Multiplexer {
	sorted := make([]Protocol, len(protocols))
	copy(sorted, protocols)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Name != sorted[j].Name {
			return sorted[i].Name < sorted[j].Name
		}
		return sorted[i].Version < sorted[j].Version
	})

	mux := &Multiplexer{
		transport: tr,
		done:      make(chan struct{}),
	}

	var offset uint64
	for _, p := range sorted {
		rw := &ProtoRW{
			Proto:  p,
			Offset: offset,
			In:     make(chan Msg, 16),
			Closed: mux.done,
		}
		mux.protos = append(mux.protos, rw)
		offset += p.Length
	}
	mux.totalLen = offset

	return mux
}

// Protocols returns the ProtoRW handles for each registered protocol.
func (mux *Multiplexer) Protocols() []*ProtoRW {
	return mux.protos
}

// WriteMsg sends a message for the given protocol.
func (mux *Multiplexer) WriteMsg(rw *ProtoRW, msg Msg) error {
	mux.mu.Lock()
	if mux.closed {
		mux.mu.Unlock()
		return ErrMuxClosed
	}
	mux.mu.Unlock()

	if msg.Code >= rw.Proto.Length {
		return fmt.Errorf("p2p/wire: message code %d exceeds protocol length %d", msg.Code, rw.Proto.Length)
	}

	wireMsg := Msg{
		Code:    msg.Code + rw.Offset,
		Size:    msg.Size,
		Payload: msg.Payload,
	}

	mux.wmu.Lock()
	defer mux.wmu.Unlock()
	return mux.transport.WriteMsg(wireMsg)
}

// ReadLoop reads messages from the transport and dispatches them to protocols.
func (mux *Multiplexer) ReadLoop() error {
	for {
		msg, err := mux.transport.ReadMsg()
		if err != nil {
			mux.Close()
			return err
		}

		rw := mux.findProto(msg.Code)
		if rw == nil {
			continue
		}

		localMsg := Msg{
			Code:    msg.Code - rw.Offset,
			Size:    msg.Size,
			Payload: msg.Payload,
		}

		select {
		case rw.In <- localMsg:
		case <-mux.done:
			return ErrMuxClosed
		}
	}
}

// Close shuts down the multiplexer.
func (mux *Multiplexer) Close() {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	if !mux.closed {
		mux.closed = true
		close(mux.done)
	}
}

func (mux *Multiplexer) findProto(code uint64) *ProtoRW {
	for _, rw := range mux.protos {
		if code >= rw.Offset && code < rw.Offset+rw.Proto.Length {
			return rw
		}
	}
	return nil
}
