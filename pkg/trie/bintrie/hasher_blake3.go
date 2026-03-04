// hasher_blake3.go adds BLAKE3 hashing support for the binary trie (EIP-7864).
// hashBLAKE3 implements the EIP-7864 hash rule: hash(left||right) = BLAKE3-256.
// Add lukechampine.com/blake3 to go.mod before using.
package bintrie

import (
	"sync"

	"github.com/eth2030/eth2030/core/types"
	"lukechampine.com/blake3"
)

// Hash function name constants.
const (
	// HashFunctionSHA256 selects SHA-256 for binary trie hashing (default).
	HashFunctionSHA256 = "sha256"
	// HashFunctionBlake3 selects BLAKE3-256 for binary trie hashing (EIP-7864 final hash).
	HashFunctionBlake3 = "blake3"
)

// hashBLAKE3 computes BLAKE3-256(left || right) for two 32-byte hash inputs.
// This is the EIP-7864 hash rule for binary internal nodes.
func hashBLAKE3(left, right []byte) types.Hash {
	h := blake3.New(32, nil)
	h.Write(left)
	h.Write(right)
	var out [32]byte
	h.Sum(out[:0])
	return types.BytesToHash(out[:])
}

// HashLeafValueBlake3 computes the BLAKE3-256 hash of a single leaf value,
// analogous to HashLeafValue but using BLAKE3.
func HashLeafValueBlake3(value []byte) types.Hash {
	if value == nil {
		return types.Hash{}
	}
	out := blake3.Sum256(value)
	return types.BytesToHash(out[:])
}

// HashPairBlake3 computes BLAKE3(left || right) for two 32-byte hashes.
func HashPairBlake3(left, right types.Hash) types.Hash {
	return hashBLAKE3(left[:], right[:])
}

// BuildMerkleRootBlake3 builds a binary Merkle tree root from a list of leaf
// hashes using BLAKE3, padded to the next power of two.
func BuildMerkleRootBlake3(leaves []types.Hash) types.Hash {
	if len(leaves) == 0 {
		return types.Hash{}
	}
	n := 1
	for n < len(leaves) {
		n <<= 1
	}
	padded := make([]types.Hash, n)
	copy(padded, leaves)

	for len(padded) > 1 {
		next := make([]types.Hash, len(padded)/2)
		for i := 0; i < len(next); i++ {
			next[i] = HashPairBlake3(padded[2*i], padded[2*i+1])
		}
		padded = next
	}
	return padded[0]
}

// BinaryHasherBlake3 computes BLAKE3-256 Merkle hashes for binary trie nodes.
// It mirrors BinaryHasher but uses BLAKE3 instead of SHA-256.
type BinaryHasherBlake3 struct {
	parallelThreshold int

	mu    sync.Mutex
	cache map[BinaryNode]types.Hash
}

// NewBinaryHasherBlake3 creates a BinaryHasherBlake3 with the given parallel threshold.
func NewBinaryHasherBlake3(parallelThreshold int) *BinaryHasherBlake3 {
	return &BinaryHasherBlake3{
		parallelThreshold: parallelThreshold,
		cache:             make(map[BinaryNode]types.Hash),
	}
}

// Hash computes the BLAKE3-256 Merkle root of the given binary trie node.
func (bh *BinaryHasherBlake3) Hash(node BinaryNode) types.Hash {
	if node == nil {
		return types.Hash{}
	}
	return bh.hashNode(node, 0)
}

func (bh *BinaryHasherBlake3) hashNode(node BinaryNode, depth int) types.Hash {
	if node == nil || IsEmptyNode(node) {
		return types.Hash{}
	}

	bh.mu.Lock()
	if h, ok := bh.cache[node]; ok {
		bh.mu.Unlock()
		return h
	}
	bh.mu.Unlock()

	var h types.Hash
	switch n := node.(type) {
	case *InternalNode:
		h = bh.hashInternal(n, depth)
	case *StemNode:
		h = n.HashBlake3()
	case HashedNode:
		h = types.Hash(n)
	default:
		h = node.Hash()
	}

	bh.mu.Lock()
	bh.cache[node] = h
	bh.mu.Unlock()
	return h
}

func (bh *BinaryHasherBlake3) hashInternal(n *InternalNode, depth int) types.Hash {
	if bh.parallelThreshold > 0 && n.GetHeight() >= bh.parallelThreshold {
		return bh.hashInternalParallel(n, depth)
	}
	return bh.hashInternalSequential(n, depth)
}

func (bh *BinaryHasherBlake3) hashInternalSequential(n *InternalNode, depth int) types.Hash {
	var leftHash, rightHash types.Hash
	if n.left != nil {
		leftHash = bh.hashNode(n.left, depth+1)
	}
	if n.right != nil {
		rightHash = bh.hashNode(n.right, depth+1)
	}
	return hashBLAKE3(leftHash[:], rightHash[:])
}

func (bh *BinaryHasherBlake3) hashInternalParallel(n *InternalNode, depth int) types.Hash {
	var leftHash, rightHash types.Hash
	var wg sync.WaitGroup

	if n.left != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			leftHash = bh.hashNode(n.left, depth+1)
		}()
	}
	if n.right != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rightHash = bh.hashNode(n.right, depth+1)
		}()
	}
	wg.Wait()
	return hashBLAKE3(leftHash[:], rightHash[:])
}

// InvalidateCache clears the hash cache.
func (bh *BinaryHasherBlake3) InvalidateCache() {
	bh.mu.Lock()
	bh.cache = make(map[BinaryNode]types.Hash)
	bh.mu.Unlock()
}
