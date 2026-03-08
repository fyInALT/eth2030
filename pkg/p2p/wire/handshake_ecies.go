package wire

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"sync"

	ethcrypto "github.com/eth2030/eth2030/crypto"
)

const (
	authMsgSize           = 32 + 65 + 65 + 1 // nonce + ephemeral + static + version
	ackMsgSize            = 32 + 65 + 1      // nonce + ephemeral + version
	eciesHandshakeVersion = 5
)

var (
	ErrECIESAuthFailed = errors.New("p2p/wire: ecies auth message verification failed")
	ErrECIESAckFailed  = errors.New("p2p/wire: ecies ack message verification failed")
	ErrECIESVersion    = errors.New("p2p/wire: ecies version mismatch")
)

// ECIESHandshake implements the full RLPx ECIES handshake protocol.
type ECIESHandshake struct {
	staticKey       *ecdsa.PrivateKey
	ephemeralKey    *ecdsa.PrivateKey
	remoteStaticPub *ecdsa.PublicKey
	remoteEphPub    *ecdsa.PublicKey
	localNonce      [32]byte
	remoteNonce     [32]byte
	initiator       bool
	aesSecret       []byte
	macSecret       []byte
}

// NewECIESHandshake creates a new ECIES handshake state.
func NewECIESHandshake(staticKey *ecdsa.PrivateKey, remoteStaticPub *ecdsa.PublicKey, initiator bool) (*ECIESHandshake, error) {
	if staticKey == nil {
		return nil, errors.New("p2p/wire: nil static key")
	}
	ephKey, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("p2p/wire: generate ephemeral key: %w", err)
	}
	h := &ECIESHandshake{
		staticKey:       staticKey,
		ephemeralKey:    ephKey,
		remoteStaticPub: remoteStaticPub,
		initiator:       initiator,
	}
	if _, err := rand.Read(h.localNonce[:]); err != nil {
		return nil, fmt.Errorf("p2p/wire: generate nonce: %w", err)
	}
	return h, nil
}

// MakeAuthMsg builds the auth message sent by the initiator.
func (h *ECIESHandshake) MakeAuthMsg() ([]byte, error) {
	if h.remoteStaticPub == nil {
		return nil, errors.New("p2p/wire: remote static key required for auth")
	}
	plain := make([]byte, authMsgSize)
	copy(plain[:32], h.localNonce[:])
	ephPub := marshalPublicKey(&h.ephemeralKey.PublicKey)
	copy(plain[32:97], ephPub)
	staticPub := marshalPublicKey(&h.staticKey.PublicKey)
	copy(plain[97:162], staticPub)
	plain[162] = eciesHandshakeVersion

	encrypted, err := ethcrypto.ECIESEncrypt(h.remoteStaticPub, plain)
	if err != nil {
		return nil, fmt.Errorf("p2p/wire: ecies encrypt auth: %w", err)
	}
	return encrypted, nil
}

// HandleAuthMsg processes a received auth message on the responder side.
func (h *ECIESHandshake) HandleAuthMsg(data []byte) error {
	plain, err := ethcrypto.ECIESDecrypt(h.staticKey, data)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrECIESAuthFailed, err)
	}
	if len(plain) < authMsgSize {
		return fmt.Errorf("%w: message too short: %d", ErrECIESAuthFailed, len(plain))
	}
	copy(h.remoteNonce[:], plain[:32])
	remoteEphPub := parsePublicKey(plain[32:97])
	if remoteEphPub == nil {
		return fmt.Errorf("%w: invalid ephemeral key", ErrECIESAuthFailed)
	}
	h.remoteEphPub = remoteEphPub
	remoteStaticPub := parsePublicKey(plain[97:162])
	if remoteStaticPub == nil {
		return fmt.Errorf("%w: invalid static key", ErrECIESAuthFailed)
	}
	h.remoteStaticPub = remoteStaticPub
	version := plain[162]
	if version < eciesHandshakeVersion {
		return fmt.Errorf("%w: remote=%d, local=%d", ErrECIESVersion, version, eciesHandshakeVersion)
	}
	return nil
}

// MakeAckMsg builds the ack message sent by the responder.
func (h *ECIESHandshake) MakeAckMsg() ([]byte, error) {
	if h.remoteStaticPub == nil {
		return nil, errors.New("p2p/wire: remote static key required for ack")
	}
	plain := make([]byte, ackMsgSize)
	copy(plain[:32], h.localNonce[:])
	ephPub := marshalPublicKey(&h.ephemeralKey.PublicKey)
	copy(plain[32:97], ephPub)
	plain[97] = eciesHandshakeVersion

	encrypted, err := ethcrypto.ECIESEncrypt(h.remoteStaticPub, plain)
	if err != nil {
		return nil, fmt.Errorf("p2p/wire: ecies encrypt ack: %w", err)
	}
	return encrypted, nil
}

// HandleAckMsg processes a received ack message on the initiator side.
func (h *ECIESHandshake) HandleAckMsg(data []byte) error {
	plain, err := ethcrypto.ECIESDecrypt(h.staticKey, data)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrECIESAckFailed, err)
	}
	if len(plain) < ackMsgSize {
		return fmt.Errorf("%w: message too short: %d", ErrECIESAckFailed, len(plain))
	}
	copy(h.remoteNonce[:], plain[:32])
	remoteEphPub := parsePublicKey(plain[32:97])
	if remoteEphPub == nil {
		return fmt.Errorf("%w: invalid ephemeral key", ErrECIESAckFailed)
	}
	h.remoteEphPub = remoteEphPub
	version := plain[97]
	if version < eciesHandshakeVersion {
		return fmt.Errorf("%w: remote=%d, local=%d", ErrECIESVersion, version, eciesHandshakeVersion)
	}
	return nil
}

// DeriveSecrets computes shared secrets from ECDH.
func (h *ECIESHandshake) DeriveSecrets() error {
	if h.remoteEphPub == nil {
		return errors.New("p2p/wire: remote ephemeral key not set")
	}
	sx, _ := h.remoteEphPub.Curve.ScalarMult(
		h.remoteEphPub.X, h.remoteEphPub.Y,
		h.ephemeralKey.D.Bytes(),
	)
	shared := make([]byte, 32)
	sxBytes := sx.Bytes()
	copy(shared[32-len(sxBytes):], sxBytes)

	var initNonce, respNonce []byte
	if h.initiator {
		initNonce = h.localNonce[:]
		respNonce = h.remoteNonce[:]
	} else {
		initNonce = h.remoteNonce[:]
		respNonce = h.localNonce[:]
	}
	h.aesSecret, h.macSecret = DeriveFrameKeys(shared, initNonce, respNonce)
	return nil
}

func (h *ECIESHandshake) AESSecret() []byte                 { return h.aesSecret }
func (h *ECIESHandshake) MACSecret() []byte                 { return h.macSecret }
func (h *ECIESHandshake) RemoteStaticPub() *ecdsa.PublicKey { return h.remoteStaticPub }
func (h *ECIESHandshake) LocalNonce() [32]byte              { return h.localNonce }
func (h *ECIESHandshake) RemoteNonce() [32]byte             { return h.remoteNonce }

// DoECIESHandshake performs the complete ECIES handshake over a net.Conn.
func DoECIESHandshake(conn net.Conn, staticKey *ecdsa.PrivateKey, remoteStaticPub *ecdsa.PublicKey, initiator bool, caps []Cap) (*FrameCodec, error) {
	hs, err := NewECIESHandshake(staticKey, remoteStaticPub, initiator)
	if err != nil {
		return nil, err
	}

	if initiator {
		authMsg, err := hs.MakeAuthMsg()
		if err != nil {
			return nil, err
		}
		if err := writeSizedMsg(conn, authMsg); err != nil {
			return nil, fmt.Errorf("p2p/wire: write auth: %w", err)
		}
		ackData, err := readSizedMsg(conn)
		if err != nil {
			return nil, fmt.Errorf("p2p/wire: read ack: %w", err)
		}
		if err := hs.HandleAckMsg(ackData); err != nil {
			return nil, err
		}
	} else {
		authData, err := readSizedMsg(conn)
		if err != nil {
			return nil, fmt.Errorf("p2p/wire: read auth: %w", err)
		}
		if err := hs.HandleAuthMsg(authData); err != nil {
			return nil, err
		}
		ackMsg, err := hs.MakeAckMsg()
		if err != nil {
			return nil, err
		}
		if err := writeSizedMsg(conn, ackMsg); err != nil {
			return nil, fmt.Errorf("p2p/wire: write ack: %w", err)
		}
	}

	if err := hs.DeriveSecrets(); err != nil {
		return nil, err
	}

	return NewFrameCodec(conn, FrameCodecConfig{
		AESKey:       hs.aesSecret,
		MACKey:       hs.macSecret,
		Initiator:    initiator,
		EnableSnappy: true,
		Caps:         caps,
	})
}

// NegotiateCaps performs capability matching, returning the highest mutually
// supported version for each protocol name.
func NegotiateCaps(local, remote []Cap) []Cap {
	localMax := make(map[string]uint)
	for _, c := range local {
		if v, ok := localMax[c.Name]; !ok || c.Version > v {
			localMax[c.Name] = c.Version
		}
	}
	remoteMax := make(map[string]uint)
	for _, c := range remote {
		if v, ok := remoteMax[c.Name]; !ok || c.Version > v {
			remoteMax[c.Name] = c.Version
		}
	}
	var matched []Cap
	for name, lv := range localMax {
		if rv, ok := remoteMax[name]; ok {
			v := lv
			if rv < v {
				v = rv
			}
			matched = append(matched, Cap{Name: name, Version: v})
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].Name != matched[j].Name {
			return matched[i].Name < matched[j].Name
		}
		return matched[i].Version < matched[j].Version
	})
	return matched
}

// FullHandshake performs both the ECIES transport handshake and the devp2p
// hello handshake in sequence.
func FullHandshake(conn net.Conn, staticKey *ecdsa.PrivateKey, remoteStaticPub *ecdsa.PublicKey, initiator bool, localHello *HelloPacket) (*FrameCodec, *HelloPacket, []Cap, error) {
	codec, err := DoECIESHandshake(conn, staticKey, remoteStaticPub, initiator, localHello.Caps)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("p2p/wire: ecies handshake: %w", err)
	}

	type result struct {
		hello *HelloPacket
		err   error
	}
	recvCh := make(chan result, 1)
	sendCh := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		payload := EncodeHello(localHello)
		sendCh <- codec.WriteMsg(Msg{
			Code:    HelloMsg,
			Size:    uint32(len(payload)),
			Payload: payload,
		})
	}()

	go func() {
		defer wg.Done()
		msg, err := codec.ReadMsg()
		if err != nil {
			recvCh <- result{nil, err}
			return
		}
		if msg.Code != HelloMsg {
			recvCh <- result{nil, fmt.Errorf("p2p/wire: expected hello, got 0x%02x", msg.Code)}
			return
		}
		hello, err := DecodeHello(msg.Payload)
		recvCh <- result{hello, err}
	}()

	if err := <-sendCh; err != nil {
		codec.Close()
		return nil, nil, nil, fmt.Errorf("p2p/wire: send hello: %w", err)
	}
	res := <-recvCh
	wg.Wait()

	if res.err != nil {
		codec.Close()
		return nil, nil, nil, fmt.Errorf("p2p/wire: recv hello: %w", res.err)
	}

	if res.hello.Version < baseProtocolVersion {
		codec.SendDisconnect(DiscProtocolError)
		return nil, nil, nil, fmt.Errorf("%w: remote=%d, local=%d",
			ErrIncompatibleVersion, res.hello.Version, baseProtocolVersion)
	}

	matched := NegotiateCaps(localHello.Caps, res.hello.Caps)
	if len(matched) == 0 {
		codec.SendDisconnect(DiscUselessPeer)
		return nil, nil, nil, ErrNoMatchingCaps
	}

	return codec, res.hello, matched, nil
}

func writeSizedMsg(conn net.Conn, data []byte) error {
	var lenBuf [2]byte
	lenBuf[0] = byte(len(data) >> 8)
	lenBuf[1] = byte(len(data))
	if _, err := conn.Write(lenBuf[:]); err != nil {
		return err
	}
	_, err := conn.Write(data)
	return err
}

func readSizedMsg(conn net.Conn) ([]byte, error) {
	var lenBuf [2]byte
	if _, err := io.ReadFull(conn, lenBuf[:]); err != nil {
		return nil, err
	}
	size := int(lenBuf[0])<<8 | int(lenBuf[1])
	if size == 0 {
		return nil, errors.New("p2p/wire: zero-length sized message")
	}
	if size > 65535 {
		return nil, errors.New("p2p/wire: sized message too large")
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}
	return data, nil
}

func marshalPublicKey(pub *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(pub.Curve, pub.X, pub.Y)
}

func parsePublicKey(data []byte) *ecdsa.PublicKey {
	if len(data) != 65 || data[0] != 0x04 {
		return nil
	}
	curve := ethcrypto.S256()
	x, y := elliptic.Unmarshal(curve, data)
	if x == nil {
		return nil
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}
}

// StaticPubKey returns the uncompressed encoding of the given ECDSA public key.
func StaticPubKey(key *ecdsa.PublicKey) []byte { return marshalPublicKey(key) }

// VerifyRemoteIdentity checks that the remote static public key matches expected.
func VerifyRemoteIdentity(got, expected *ecdsa.PublicKey) error {
	if expected == nil {
		return nil
	}
	if got == nil {
		return errors.New("p2p/wire: no remote static key received")
	}
	h1 := sha256.Sum256(marshalPublicKey(got))
	h2 := sha256.Sum256(marshalPublicKey(expected))
	if h1 != h2 {
		return errors.New("p2p/wire: remote identity mismatch")
	}
	return nil
}
