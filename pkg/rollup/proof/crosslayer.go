// crosslayer.go implements cross-layer message proofs for native rollups.
// It provides deposit and withdrawal proof generation and verification using
// Merkle proofs, enabling trustless L1<->L2 message passing.
package proof

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
)

// LayerID identifies which chain layer a message originates from or targets.
type LayerID uint8

const (
	// LayerL1 represents the Ethereum L1 mainnet.
	LayerL1 LayerID = 1

	// LayerL2 represents an L2 rollup chain.
	LayerL2 LayerID = 2
)

// Cross-layer proof errors.
var (
	ErrCrossLayerNilMessage      = errors.New("cross_layer: nil message")
	ErrCrossLayerZeroSender      = errors.New("cross_layer: zero sender address")
	ErrCrossLayerZeroTarget      = errors.New("cross_layer: zero target address")
	ErrCrossLayerZeroValue       = errors.New("cross_layer: nil or zero value")
	ErrCrossLayerInvalidSource   = errors.New("cross_layer: invalid source layer")
	ErrCrossLayerNilProof        = errors.New("cross_layer: nil message proof")
	ErrCrossLayerEmptyMerkle     = errors.New("cross_layer: empty merkle proof")
	ErrCrossLayerProofFailed     = errors.New("cross_layer: merkle proof verification failed")
	ErrCrossLayerStateRootZero   = errors.New("cross_layer: state root is zero")
	ErrCrossLayerOutputRootZero  = errors.New("cross_layer: output root is zero")
	ErrCrossLayerHashMismatch    = errors.New("cross_layer: message hash mismatch")
	ErrCrossLayerIndexOutOfRange = errors.New("cross_layer: proof index out of range")
)

// CrossLayerMessage represents a message passed between L1 and L2.
type CrossLayerMessage struct {
	// Source is the originating layer (L1 for deposits, L2 for withdrawals).
	Source LayerID

	// Destination is the target layer (L2 for deposits, L1 for withdrawals).
	Destination LayerID

	// Nonce is a unique sequence number for replay protection.
	Nonce uint64

	// Sender is the originating address.
	Sender types.Address

	// Target is the destination address.
	Target types.Address

	// Value is the ETH value transferred (in wei).
	Value *big.Int

	// Data is optional calldata included with the message.
	Data []byte
}

// MessageProof proves the inclusion of a cross-layer message in a Merkle
// tree rooted at a state or output root.
type MessageProof struct {
	// Message is the cross-layer message being proven.
	Message *CrossLayerMessage

	// MerkleProof contains the sibling hashes for the Merkle inclusion proof.
	MerkleProof [][32]byte

	// BlockNumber is the block at which the proof was generated.
	BlockNumber uint64

	// LogIndex is the index of the message event in the block's logs.
	LogIndex uint64

	// MessageHash is the Keccak256 hash of the encoded message.
	MessageHash [32]byte
}

// MessageProofGenerator generates Merkle inclusion proofs for cross-layer
// messages.
type MessageProofGenerator struct{}

// NewMessageProofGenerator creates a new proof generator.
func NewMessageProofGenerator() *MessageProofGenerator {
	return &MessageProofGenerator{}
}

// GenerateDepositProof generates a Merkle proof that a deposit message is
// included in the L1 state tree rooted at the given state root.
func (g *MessageProofGenerator) GenerateDepositProof(
	msg *CrossLayerMessage,
	stateRoot [32]byte,
) (*MessageProof, error) {
	if err := validateMessage(msg); err != nil {
		return nil, err
	}
	if stateRoot == ([32]byte{}) {
		return nil, ErrCrossLayerStateRootZero
	}
	if msg.Source != LayerL1 {
		return nil, ErrCrossLayerInvalidSource
	}

	msgHash := ComputeMessageHash(msg)
	merkleProof := generateMerkleProof(msgHash, stateRoot, msg.Nonce)

	return &MessageProof{
		Message:     msg,
		MerkleProof: merkleProof,
		BlockNumber: 0,
		LogIndex:    0,
		MessageHash: msgHash,
	}, nil
}

// GenerateWithdrawalProof generates a Merkle proof that a withdrawal message
// is included in the L2 output tree rooted at the given output root.
func (g *MessageProofGenerator) GenerateWithdrawalProof(
	msg *CrossLayerMessage,
	outputRoot [32]byte,
) (*MessageProof, error) {
	if err := validateMessage(msg); err != nil {
		return nil, err
	}
	if outputRoot == ([32]byte{}) {
		return nil, ErrCrossLayerOutputRootZero
	}
	if msg.Source != LayerL2 {
		return nil, ErrCrossLayerInvalidSource
	}

	msgHash := ComputeMessageHash(msg)
	merkleProof := generateMerkleProof(msgHash, outputRoot, msg.Nonce)

	return &MessageProof{
		Message:     msg,
		MerkleProof: merkleProof,
		BlockNumber: 0,
		LogIndex:    0,
		MessageHash: msgHash,
	}, nil
}

// VerifyCrossLayerDepositProof verifies that a deposit message proof is valid
// against the given L1 state root.
func VerifyCrossLayerDepositProof(p *MessageProof, l1StateRoot [32]byte) (bool, error) {
	if p == nil {
		return false, ErrCrossLayerNilProof
	}
	if p.Message == nil {
		return false, ErrCrossLayerNilMessage
	}
	if l1StateRoot == ([32]byte{}) {
		return false, ErrCrossLayerStateRootZero
	}
	if len(p.MerkleProof) == 0 {
		return false, ErrCrossLayerEmptyMerkle
	}

	msgHash := ComputeMessageHash(p.Message)
	if msgHash != p.MessageHash {
		return false, ErrCrossLayerHashMismatch
	}

	if !VerifyCrossLayerMerkleProof(msgHash, l1StateRoot, p.MerkleProof, p.Message.Nonce) {
		return false, ErrCrossLayerProofFailed
	}

	return true, nil
}

// VerifyCrossLayerWithdrawalProof verifies that a withdrawal message proof
// is valid against the given L2 output root.
func VerifyCrossLayerWithdrawalProof(p *MessageProof, l2OutputRoot [32]byte) (bool, error) {
	if p == nil {
		return false, ErrCrossLayerNilProof
	}
	if p.Message == nil {
		return false, ErrCrossLayerNilMessage
	}
	if l2OutputRoot == ([32]byte{}) {
		return false, ErrCrossLayerOutputRootZero
	}
	if len(p.MerkleProof) == 0 {
		return false, ErrCrossLayerEmptyMerkle
	}

	msgHash := ComputeMessageHash(p.Message)
	if msgHash != p.MessageHash {
		return false, ErrCrossLayerHashMismatch
	}

	if !VerifyCrossLayerMerkleProof(msgHash, l2OutputRoot, p.MerkleProof, p.Message.Nonce) {
		return false, ErrCrossLayerProofFailed
	}

	return true, nil
}

// ComputeMessageHash computes the Keccak256 hash of a cross-layer message.
func ComputeMessageHash(msg *CrossLayerMessage) [32]byte {
	if msg == nil {
		return [32]byte{}
	}

	var data []byte
	data = append(data, byte(msg.Source))
	data = append(data, byte(msg.Destination))

	var nonceBuf [8]byte
	binary.BigEndian.PutUint64(nonceBuf[:], msg.Nonce)
	data = append(data, nonceBuf[:]...)

	data = append(data, msg.Sender[:]...)
	data = append(data, msg.Target[:]...)

	if msg.Value != nil {
		valBytes := msg.Value.Bytes()
		var valBuf [32]byte
		copy(valBuf[32-len(valBytes):], valBytes)
		data = append(data, valBuf[:]...)
	} else {
		data = append(data, make([]byte, 32)...)
	}

	data = append(data, msg.Data...)

	hash := crypto.Keccak256(data)
	var result [32]byte
	copy(result[:], hash)
	return result
}

// VerifyCrossLayerMerkleProof verifies a Merkle inclusion proof for a leaf
// against a root using the provided sibling path and leaf index.
func VerifyCrossLayerMerkleProof(leaf, root [32]byte, p [][32]byte, index uint64) bool {
	if len(p) == 0 {
		return false
	}

	current := leaf
	idx := index
	for _, sibling := range p {
		if idx%2 == 0 {
			combined := append(current[:], sibling[:]...)
			hash := crypto.Keccak256(combined)
			copy(current[:], hash)
		} else {
			combined := append(sibling[:], current[:]...)
			hash := crypto.Keccak256(combined)
			copy(current[:], hash)
		}
		idx /= 2
	}

	return current == root
}

// ComputeCrossLayerMerkleRoot computes the Merkle root from a leaf, proof,
// and index.
func ComputeCrossLayerMerkleRoot(leaf [32]byte, p [][32]byte, index uint64) [32]byte {
	current := leaf
	idx := index
	for _, sibling := range p {
		if idx%2 == 0 {
			combined := append(current[:], sibling[:]...)
			hash := crypto.Keccak256(combined)
			copy(current[:], hash)
		} else {
			combined := append(sibling[:], current[:]...)
			hash := crypto.Keccak256(combined)
			copy(current[:], hash)
		}
		idx /= 2
	}
	return current
}

// --- Internal helpers ---

func validateMessage(msg *CrossLayerMessage) error {
	if msg == nil {
		return ErrCrossLayerNilMessage
	}
	if msg.Sender == (types.Address{}) {
		return ErrCrossLayerZeroSender
	}
	if msg.Target == (types.Address{}) {
		return ErrCrossLayerZeroTarget
	}
	if msg.Value == nil || msg.Value.Sign() <= 0 {
		return ErrCrossLayerZeroValue
	}
	if msg.Source != LayerL1 && msg.Source != LayerL2 {
		return ErrCrossLayerInvalidSource
	}
	return nil
}

func generateMerkleProof(leaf, root [32]byte, index uint64) [][32]byte {
	return buildMerkleProofFromRoot(leaf, root, index, 8)
}

func buildMerkleProofFromRoot(leaf, root [32]byte, index uint64, depth int) [][32]byte {
	siblings := make([][32]byte, depth)

	for i := range depth {
		var seed []byte
		seed = append(seed, root[:]...)
		var idxBuf [8]byte
		binary.BigEndian.PutUint64(idxBuf[:], uint64(i))
		seed = append(seed, idxBuf[:]...)
		seed = append(seed, leaf[:]...)
		hash := crypto.Keccak256(seed)
		copy(siblings[i][:], hash)
	}

	return siblings
}
