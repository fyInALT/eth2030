// proof_verifier.go implements Merkle proof verification for binary trie
// inclusion and exclusion proofs, path validation, and root-hash consistency
// checks. This extends the existing proof.go with structured proof types
// and a standalone verifier that does not require access to the full trie.
package bintrie

import (
	"bytes"
	"crypto/sha256"
	"errors"

	"github.com/eth2030/eth2030/core/types"
)

// Proof verifier errors.
var (
	ErrNilProof           = errors.New("proof_verifier: nil proof")
	ErrInvalidKeyLength   = errors.New("proof_verifier: key must be 32 bytes")
	ErrInvalidValueLength = errors.New("proof_verifier: value must be 32 bytes")
	ErrKeyNotFound        = errors.New("proof_verifier: key not found in trie")
	ErrKeyExists          = errors.New("proof_verifier: key exists in trie (cannot build exclusion proof)")
	ErrEmptyRoot          = errors.New("proof_verifier: empty root hash")
	ErrProofPathEmpty     = errors.New("proof_verifier: proof path has no siblings")
)

// ProofPath is an ordered sequence of sibling hashes along a path from the
// leaf to the root in the binary trie. Index 0 is the deepest sibling
// (closest to the leaf), and the last element is closest to the root.
type ProofPath struct {
	Siblings []types.Hash
}

// Len returns the number of sibling hashes in the path.
func (pp *ProofPath) Len() int {
	if pp == nil {
		return 0
	}
	return len(pp.Siblings)
}

// InclusionProof proves that a key-value pair exists in the trie.
type InclusionProof struct {
	// Key is the full 32-byte key.
	Key [HashSize]byte
	// Value is the 32-byte leaf value.
	Value [HashSize]byte
	// Path contains the sibling hashes from leaf to root.
	Path ProofPath
	// StemHash is the hash of the StemNode containing this key.
	StemHash types.Hash
}

// ExclusionProof proves that a key does NOT exist in the trie.
// It works by showing the divergence point: either the trie path leads
// to an empty node, or to a StemNode with a different stem.
type ExclusionProof struct {
	// Key is the full 32-byte key being proven absent.
	Key [HashSize]byte
	// Path contains the sibling hashes from divergence point to root.
	Path ProofPath
	// DivergentStem is the stem found at the path if a different stem
	// occupies that position. Nil if the path leads to an empty node.
	DivergentStem []byte
	// DivergentStemHash is the hash of the divergent stem node, if any.
	DivergentStemHash types.Hash
	// EmptyAtPath is true if the path ends at an empty node.
	EmptyAtPath bool
}

// ProofVerifier verifies inclusion and exclusion proofs against a root hash.
type ProofVerifier struct {
	root types.Hash
}

// NewProofVerifier creates a new ProofVerifier bound to the given root hash.
func NewProofVerifier(root []byte) *ProofVerifier {
	var h types.Hash
	if len(root) >= HashSize {
		copy(h[:], root[:HashSize])
	} else {
		copy(h[:], root)
	}
	return &ProofVerifier{root: h}
}

// Root returns the root hash this verifier checks against.
func (pv *ProofVerifier) Root() types.Hash {
	return pv.root
}

// VerifyInclusion walks the proof path from the stem node hash back to the
// root, recomputing each level using SHA-256(left || right). It returns true
// if the recomputed root matches the verifier's root.
func (pv *ProofVerifier) VerifyInclusion(proof *InclusionProof) bool {
	if proof == nil {
		return false
	}
	if pv.root == (types.Hash{}) {
		return false
	}

	// Recompute the stem node hash from key and value.
	stem := proof.Key[:StemSize]
	leafIdx := proof.Key[StemSize]
	stemHash := computeStemLeafHash(stem, leafIdx, proof.Value[:])

	// If the proof has a stored stem hash, it must match our computation.
	if proof.StemHash != (types.Hash{}) && proof.StemHash != stemHash {
		return false
	}

	// Walk the siblings from bottom (leaf-side) to top (root-side).
	current := stemHash
	for i, sib := range proof.Path.Siblings {
		depth := len(proof.Path.Siblings) - 1 - i
		bit := getBit(stem, depth)
		current = hashPair(current, sib, bit)
	}

	return current == pv.root
}

// VerifyExclusion verifies that a key does NOT exist in the trie.
// The proof demonstrates either:
//  1. The path leads to an empty node (EmptyAtPath=true), or
//  2. The path leads to a StemNode with a different stem (DivergentStem set).
func (pv *ProofVerifier) VerifyExclusion(proof *ExclusionProof) bool {
	if proof == nil {
		return false
	}
	if pv.root == (types.Hash{}) {
		return false
	}

	stem := proof.Key[:StemSize]

	var current types.Hash
	if proof.EmptyAtPath {
		// Empty node hashes to zero.
		current = types.Hash{}
	} else if len(proof.DivergentStem) == StemSize {
		// A different stem exists at this position.
		if bytes.Equal(proof.DivergentStem, stem) {
			// The stem matches -- this is NOT an exclusion proof.
			return false
		}
		current = proof.DivergentStemHash
	} else {
		return false
	}

	// Walk siblings from bottom to top.
	for i, sib := range proof.Path.Siblings {
		depth := len(proof.Path.Siblings) - 1 - i
		bit := getBit(stem, depth)
		current = hashPair(current, sib, bit)
	}

	return current == pv.root
}

// BuildInclusionProof constructs an inclusion proof for the given key
// from a live BinaryTrie. Returns an error if the key does not exist.
func BuildInclusionProof(trie *BinaryTrie, key []byte) (*InclusionProof, error) {
	if len(key) != HashSize {
		return nil, ErrInvalidKeyLength
	}

	// Verify the key exists.
	val, err := trie.Get(key)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, ErrKeyNotFound
	}

	siblings, stemNode := collectVerifierSiblings(trie.root, key[:StemSize], 0)

	proof := &InclusionProof{
		Path: ProofPath{Siblings: siblings},
	}
	copy(proof.Key[:], key)
	if len(val) >= HashSize {
		copy(proof.Value[:], val[:HashSize])
	} else {
		copy(proof.Value[HashSize-len(val):], val)
	}
	if stemNode != nil {
		proof.StemHash = stemNode.Hash()
	}

	return proof, nil
}

// BuildExclusionProof constructs an exclusion proof for the given key
// from a live BinaryTrie. Returns an error if the key exists.
func BuildExclusionProof(trie *BinaryTrie, key []byte) (*ExclusionProof, error) {
	if len(key) != HashSize {
		return nil, ErrInvalidKeyLength
	}

	// Verify the key does NOT exist.
	val, err := trie.Get(key)
	if err != nil {
		return nil, err
	}
	if val != nil {
		return nil, ErrKeyExists
	}

	siblings, divergentStem := collectExclusionSiblings(trie.root, key[:StemSize], 0)

	proof := &ExclusionProof{
		Path: ProofPath{Siblings: siblings},
	}
	copy(proof.Key[:], key)

	if divergentStem != nil {
		proof.DivergentStem = make([]byte, len(divergentStem.Stem))
		copy(proof.DivergentStem, divergentStem.Stem)
		proof.DivergentStemHash = divergentStem.Hash()
		proof.EmptyAtPath = false
	} else {
		proof.EmptyAtPath = true
	}

	return proof, nil
}

// collectVerifierSiblings walks the trie collecting sibling hashes for an
// inclusion proof. Returns siblings in leaf-to-root order and the stem node.
func collectVerifierSiblings(node BinaryNode, stem []byte, depth int) ([]types.Hash, *StemNode) {
	switch n := node.(type) {
	case *InternalNode:
		bit := getBit(stem, depth)
		var sibling, child BinaryNode
		if bit == 0 {
			child = n.left
			sibling = n.right
		} else {
			child = n.right
			sibling = n.left
		}

		var sibHash types.Hash
		if sibling != nil {
			sibHash = sibling.Hash()
		}

		deeper, stemNode := collectVerifierSiblings(child, stem, depth+1)
		// Append sibling at the end (leaf-to-root order: deeper first).
		result := make([]types.Hash, len(deeper)+1)
		copy(result, deeper)
		result[len(deeper)] = sibHash
		return result, stemNode

	case *StemNode:
		if bytes.Equal(n.Stem, stem) {
			return nil, n
		}
		return nil, nil

	case Empty:
		return nil, nil

	default:
		return nil, nil
	}
}

// collectExclusionSiblings walks the trie collecting sibling hashes for an
// exclusion proof. Returns siblings in leaf-to-root order and any divergent stem.
func collectExclusionSiblings(node BinaryNode, stem []byte, depth int) ([]types.Hash, *StemNode) {
	switch n := node.(type) {
	case *InternalNode:
		bit := getBit(stem, depth)
		var sibling, child BinaryNode
		if bit == 0 {
			child = n.left
			sibling = n.right
		} else {
			child = n.right
			sibling = n.left
		}

		var sibHash types.Hash
		if sibling != nil {
			sibHash = sibling.Hash()
		}

		deeper, divergent := collectExclusionSiblings(child, stem, depth+1)
		result := make([]types.Hash, len(deeper)+1)
		copy(result, deeper)
		result[len(deeper)] = sibHash
		return result, divergent

	case *StemNode:
		if !bytes.Equal(n.Stem, stem) {
			return nil, n
		}
		// Stem matches; this is not an exclusion (leaf might be nil though).
		return nil, nil

	case Empty:
		return nil, nil

	default:
		return nil, nil
	}
}

// getBit returns the bit at position depth in the stem byte slice.
// depth 0 is the MSB of stem[0].
func getBit(stem []byte, depth int) byte {
	if depth/8 >= len(stem) {
		return 0
	}
	return (stem[depth/8] >> (7 - (depth % 8))) & 1
}

// hashPair computes SHA-256(left || right) based on the direction bit.
// If bit=0, current is left and sibling is right.
// If bit=1, current is right and sibling is left.
func hashPair(current, sibling types.Hash, bit byte) types.Hash {
	h := sha256.New()
	if bit == 0 {
		h.Write(current[:])
		h.Write(sibling[:])
	} else {
		h.Write(sibling[:])
		h.Write(current[:])
	}
	return types.BytesToHash(h.Sum(nil))
}

// computeStemLeafHash computes the expected hash of a StemNode containing
// a single known leaf value. This mirrors the StemNode.Hash() logic.
func computeStemLeafHash(stem []byte, leafIdx byte, value []byte) types.Hash {
	var data [StemNodeWidth]types.Hash
	if value != nil {
		h := sha256.Sum256(value)
		data[leafIdx] = types.BytesToHash(h[:])
	}

	h := sha256.New()
	for level := 1; level <= 8; level++ {
		for i := range StemNodeWidth / (1 << level) {
			h.Reset()
			if data[i*2] == (types.Hash{}) && data[i*2+1] == (types.Hash{}) {
				data[i] = types.Hash{}
				continue
			}
			h.Write(data[i*2][:])
			h.Write(data[i*2+1][:])
			data[i] = types.BytesToHash(h.Sum(nil))
		}
	}

	h.Reset()
	h.Write(stem)
	h.Write([]byte{0})
	h.Write(data[0][:])
	return types.BytesToHash(h.Sum(nil))
}
