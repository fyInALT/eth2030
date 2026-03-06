package discover

import (
	"encoding/binary"
	"net"
	"testing"

	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/p2p/enode"
	"github.com/eth2030/eth2030/p2p/enr"
	"github.com/eth2030/eth2030/rlp"
)

// buildPktPong encodes a Pong and prepends the MsgPong type byte.
func buildPktPong(t *testing.T, pong Pong) []byte {
	t.Helper()
	data, err := rlp.EncodeToBytes(pong)
	if err != nil {
		t.Fatal(err)
	}
	return append([]byte{MsgPong}, data...)
}

// buildPktFindNode encodes a FindNode and prepends the MsgFindNode type byte.
func buildPktFindNode(t *testing.T, req FindNode) []byte {
	t.Helper()
	data, err := rlp.EncodeToBytes(req)
	if err != nil {
		t.Fatal(err)
	}
	return append([]byte{MsgFindNode}, data...)
}

// buildPktNodes encodes a Nodes and prepends the MsgNodes type byte.
func buildPktNodes(t *testing.T, nodes Nodes) []byte {
	t.Helper()
	data, err := rlp.EncodeToBytes(nodes)
	if err != nil {
		t.Fatal(err)
	}
	return append([]byte{MsgNodes}, data...)
}

// buildPktHandshake encodes a Handshake and prepends the MsgHandshake type byte.
func buildPktHandshake(t *testing.T, hs Handshake) []byte {
	t.Helper()
	data, err := rlp.EncodeToBytes(hs)
	if err != nil {
		t.Fatal(err)
	}
	return append([]byte{MsgHandshake}, data...)
}

// buildWhoAreYouPacket builds a raw WhoAreYou payload (type byte + nonce+idnonce+enrseq).
func buildWhoAreYouPacket(nonce [NonceSize]byte, idNonce [16]byte, enrSeq uint64) []byte {
	buf := make([]byte, 1+NonceSize+16+8)
	buf[0] = MsgWhoAreYou
	copy(buf[1:1+NonceSize], nonce[:])
	copy(buf[1+NonceSize:1+NonceSize+16], idNonce[:])
	binary.BigEndian.PutUint64(buf[1+NonceSize+16:], enrSeq)
	return buf
}

// makeV5Proto creates a V5Protocol instance bound to a loopback UDP socket.
func makeV5Proto(t *testing.T) (*V5Protocol, func()) {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	node, conn := makeLocalNode(t)
	p := NewV5Protocol(conn, key, node)
	return p, func() { conn.Close() }
}

// --- HandlePacket: short packet ---

func TestV5HandlePacket_TooShort(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, []byte{})        // zero bytes: should return early
	p.HandlePacket(from, []byte{MsgPing}) // one byte only: should return early
}

// --- handlePong ---

func TestV5HandlePacketPong(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	pong := Pong{ReqID: []byte{1, 2}, ENRSeq: 99, ToIP: net.ParseIP("10.0.0.1"), ToPort: 30303}
	packet := buildPktPong(t, pong)

	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet) // should not panic
}

func TestV5HandlePacketPong_InvalidRLP(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	packet := []byte{MsgPong, 0xFF, 0xFE, 0xFD}
	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet) // corrupt RLP: should return early, not panic
}

// --- handleFindNode ---

func TestV5HandlePacketFindNode_EmptyTable(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	req := FindNode{ReqID: []byte{7, 8}, Distances: []uint{1, 2, 3}}
	packet := buildPktFindNode(t, req)

	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet)
}

func TestV5HandlePacketFindNode_Distance0(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	// Distance 0 means "return self".
	req := FindNode{ReqID: []byte{1}, Distances: []uint{0}}
	packet := buildPktFindNode(t, req)

	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet)
}

func TestV5HandlePacketFindNode_OverMaxBuckets(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	// Distances > NumBuckets should be skipped.
	req := FindNode{ReqID: []byte{1}, Distances: []uint{300, 500}}
	packet := buildPktFindNode(t, req)

	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet)
}

func TestV5HandlePacketFindNode_InvalidRLP(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	packet := []byte{MsgFindNode, 0xFF, 0xFE}
	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet) // should not panic
}

// --- handleNodes ---

func TestV5HandlePacketNodes_Empty(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	nodes := Nodes{ReqID: []byte{1}, Total: 0, ENRs: nil}
	packet := buildPktNodes(t, nodes)

	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet)
}

func TestV5HandlePacketNodes_WithValidENR(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	// Build a valid signed ENR.
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	rec := &enr.Record{Seq: 1}
	rec.Set(enr.KeyIP, []byte{192, 168, 1, 1})
	rec.Set(enr.KeyUDP, []byte{0x76, 0x5f})
	rec.Set(enr.KeyTCP, []byte{0x76, 0x5f})
	if err := enr.SignENR(rec, key); err != nil {
		t.Fatal(err)
	}
	encoded, err := enr.EncodeENR(rec)
	if err != nil {
		t.Fatal(err)
	}

	nodes := Nodes{
		ReqID: []byte{1},
		Total: 1,
		ENRs:  [][]byte{encoded},
	}
	packet := buildPktNodes(t, nodes)

	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet)
}

func TestV5HandlePacketNodes_InvalidENR(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	// Garbage ENR bytes.
	nodes := Nodes{
		ReqID: []byte{1},
		Total: 1,
		ENRs:  [][]byte{{0xDE, 0xAD, 0xBE, 0xEF}},
	}
	packet := buildPktNodes(t, nodes)

	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet) // bad ENR: should be skipped, not panic
}

func TestV5HandlePacketNodes_ENR_NoIPField(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	// Valid ENR but missing IP field -> node should be skipped.
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	rec := &enr.Record{Seq: 1}
	// No IP set on purpose.
	if err := enr.SignENR(rec, key); err != nil {
		t.Fatal(err)
	}
	encoded, err := enr.EncodeENR(rec)
	if err != nil {
		t.Fatal(err)
	}

	nodes := Nodes{
		ReqID: []byte{1},
		Total: 1,
		ENRs:  [][]byte{encoded},
	}
	packet := buildPktNodes(t, nodes)

	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet) // no IP bytes -> skip
}

func TestV5HandlePacketNodes_InvalidRLP(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	packet := []byte{MsgNodes, 0xFF, 0xFE}
	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet) // should not panic
}

// --- handleWhoAreYou ---

func TestV5HandlePacketWhoAreYou(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	var nonce [NonceSize]byte
	nonce[0] = 0xAB
	var idNonce [16]byte
	idNonce[0] = 0xCD

	packet := buildWhoAreYouPacket(nonce, idNonce, 0)
	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	// respondHandshake will attempt to write to conn; may fail send but must not panic.
	p.HandlePacket(from, packet)
}

func TestV5HandlePacketWhoAreYou_TooShort(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	// Payload shorter than NonceSize+16+8 bytes after the type byte.
	packet := []byte{MsgWhoAreYou, 0x01, 0x02}
	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet) // should return early, not panic
}

// --- handleHandshake ---

func TestV5HandlePacketHandshake(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	var remoteID enode.NodeID
	remoteID[0] = 0x77

	hs := Handshake{
		SrcID:   remoteID,
		IDSig:   make([]byte, 64),
		EPubkey: make([]byte, 33),
	}
	packet := buildPktHandshake(t, hs)

	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet)

	// A session should have been established.
	s, ok := p.GetSession(remoteID)
	if !ok {
		t.Fatal("session should be established after handleHandshake")
	}
	if !s.Established {
		t.Fatal("session.Established should be true")
	}
}

func TestV5HandlePacketHandshake_InvalidRLP(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	packet := []byte{MsgHandshake, 0xFF, 0xFE}
	from := net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 30303}
	p.HandlePacket(from, packet) // should not panic
}

// --- respondHandshake directly ---

func TestV5RespondHandshake_OlderENRSeq(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	// ENRSeq in challenge < local node's seq (1), so ENR is included in response.
	var nonce [NonceSize]byte
	var idNonce [16]byte
	challenge := WhoAreYou{
		Nonce:   nonce,
		IDNonce: idNonce,
		ENRSeq:  0,
	}
	from := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 30304}
	// May fail to deliver to 30304 but must not panic.
	p.respondHandshake(from, challenge)
}

func TestV5RespondHandshake_SameENRSeq(t *testing.T) {
	p, cleanup := makeV5Proto(t)
	defer cleanup()

	// ENRSeq matches local: ENR is not included in the handshake response.
	var nonce [NonceSize]byte
	var idNonce [16]byte
	challenge := WhoAreYou{
		Nonce:   nonce,
		IDNonce: idNonce,
		ENRSeq:  p.localNode.Record.Seq, // same seq
	}
	from := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 30305}
	p.respondHandshake(from, challenge)
}
