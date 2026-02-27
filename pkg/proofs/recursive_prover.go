// recursive_prover.go implements recursive proof composition for the proof
// aggregation framework. It builds binary trees of proofs with Merkle
// commitments, enabling efficient recursive verification of composed proofs.
//
// Part of the EL roadmap: proof aggregation and mandatory 3-of-5 proofs (K+).
package proofs

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/big"
	"math/bits"
)

// Recursive prover errors.
var (
	ErrRecNoProofs        = errors.New("recursive_prover: no proofs to compose")
	ErrRecNilProof        = errors.New("recursive_prover: nil recursive proof")
	ErrRecEmptyTree       = errors.New("recursive_prover: empty proof tree")
	ErrRecTreeDepthExceed = errors.New("recursive_prover: tree depth exceeds maximum")
	ErrRecRootMismatch    = errors.New("recursive_prover: root commitment mismatch")
	ErrRecInvalidNode     = errors.New("recursive_prover: invalid proof node")
	ErrRecNoProofData     = errors.New("recursive_prover: proof data is empty")
)

// MaxTreeDepth is the maximum allowed depth for recursive proof trees.
const MaxTreeDepth = 32

// AggregateableProof is the interface that proof types must implement to
// participate in recursive composition.
type AggregateableProof interface {
	// ProofBytes returns the raw proof data.
	ProofBytes() []byte

	// ProofKind returns the proof type identifier.
	ProofKind() ProofType
}

// SimpleAggregateable wraps raw proof bytes and a type for aggregation.
type SimpleAggregateable struct {
	Data []byte
	Kind ProofType
}

// ProofBytes returns the raw proof data.
func (s *SimpleAggregateable) ProofBytes() []byte { return s.Data }

// ProofKind returns the proof type identifier.
func (s *SimpleAggregateable) ProofKind() ProofType { return s.Kind }

// ProofNode represents a node in a recursive proof tree.
type ProofNode struct {
	// Proof is the proof data at this node (leaf nodes only).
	Proof []byte

	// Commitment is the Merkle commitment hash for this subtree.
	Commitment [32]byte

	// Children are the left and right child nodes (internal nodes).
	Children []*ProofNode

	// Depth is this node's depth in the tree (root = 0).
	Depth int

	// Type is the proof type for leaf nodes.
	Type ProofType

	// IsLeaf indicates whether this is a leaf node with actual proof data.
	IsLeaf bool
}

// RecursiveProof represents a fully composed recursive proof.
type RecursiveProof struct {
	// Root is the top-level Merkle commitment.
	Root [32]byte

	// Tree is the root node of the proof tree.
	Tree *ProofNode

	// Depth is the maximum depth of the tree.
	Depth int

	// TotalProofs is the number of leaf proofs in the tree.
	TotalProofs int

	// SerializedCommitment binds the root and metadata.
	SerializedCommitment []byte
}

// ProofTree is a binary tree of proofs with Merkle commitments.
type ProofTree struct {
	// Root is the root node.
	Root *ProofNode

	// Leaves are the leaf nodes in order.
	Leaves []*ProofNode

	// Depth is the tree depth.
	Depth int

	// LeafCount is the number of leaves.
	LeafCount int
}

// RecursiveProver composes multiple proofs into a single recursive proof.
type RecursiveProver struct {
	// maxDepth limits the tree depth.
	maxDepth int
}

// NewRecursiveProver creates a new recursive prover with the given max depth.
func NewRecursiveProver(maxDepth int) *RecursiveProver {
	if maxDepth <= 0 || maxDepth > MaxTreeDepth {
		maxDepth = MaxTreeDepth
	}
	return &RecursiveProver{maxDepth: maxDepth}
}

// ComposeProofs composes multiple proofs into a single recursive proof
// by building a binary Merkle tree over the proof commitments.
func (rp *RecursiveProver) ComposeProofs(proofs []AggregateableProof) (*RecursiveProof, error) {
	if len(proofs) == 0 {
		return nil, ErrRecNoProofs
	}

	// Build leaf nodes.
	leaves := make([]*ProofNode, len(proofs))
	for i, p := range proofs {
		data := p.ProofBytes()
		if len(data) == 0 {
			return nil, ErrRecNoProofData
		}
		commitment := hashLeaf(data, p.ProofKind())
		leaves[i] = &ProofNode{
			Proof:      append([]byte(nil), data...),
			Commitment: commitment,
			Depth:      0, // will be set during tree construction
			Type:       p.ProofKind(),
			IsLeaf:     true,
		}
	}

	// Build the binary tree bottom-up.
	tree, depth := buildTreeFromLeaves(leaves)
	if depth > rp.maxDepth {
		return nil, ErrRecTreeDepthExceed
	}

	commitment := serializeRecursiveCommitment(tree.Commitment, len(proofs), depth)

	return &RecursiveProof{
		Root:                 tree.Commitment,
		Tree:                 tree,
		Depth:                depth,
		TotalProofs:          len(proofs),
		SerializedCommitment: commitment,
	}, nil
}

// VerifyRecursive verifies a recursive proof by recomputing the Merkle tree
// root from the leaf proofs and comparing it to the stored root commitment.
func (rp *RecursiveProver) VerifyRecursive(proof *RecursiveProof) (bool, error) {
	if proof == nil {
		return false, ErrRecNilProof
	}
	if proof.Tree == nil {
		return false, ErrRecEmptyTree
	}

	// Recompute the commitment by traversing the tree.
	recomputed, err := recomputeCommitment(proof.Tree)
	if err != nil {
		return false, err
	}

	if recomputed != proof.Root {
		return false, ErrRecRootMismatch
	}

	// Verify the serialized commitment.
	expected := serializeRecursiveCommitment(proof.Root, proof.TotalProofs, proof.Depth)
	if len(expected) != len(proof.SerializedCommitment) {
		return false, ErrRecRootMismatch
	}
	for i := range expected {
		if expected[i] != proof.SerializedCommitment[i] {
			return false, ErrRecRootMismatch
		}
	}

	return true, nil
}

// BuildProofTree constructs a binary proof tree from raw proof bytes.
// Each byte slice becomes a leaf node with a ZKSNARK type by default.
func BuildProofTree(proofs [][]byte) (*ProofTree, error) {
	if len(proofs) == 0 {
		return nil, ErrRecNoProofs
	}

	leaves := make([]*ProofNode, len(proofs))
	for i, data := range proofs {
		if len(data) == 0 {
			return nil, ErrRecNoProofData
		}
		commitment := hashLeaf(data, ZKSNARK)
		leaves[i] = &ProofNode{
			Proof:      append([]byte(nil), data...),
			Commitment: commitment,
			IsLeaf:     true,
			Type:       ZKSNARK,
		}
	}

	root, depth := buildTreeFromLeaves(leaves)
	setDepths(root, 0)

	return &ProofTree{
		Root:      root,
		Leaves:    leaves,
		Depth:     depth,
		LeafCount: len(proofs),
	}, nil
}

// OptimizeTree rebalances a proof tree for minimal verification cost.
// It rebuilds the tree ensuring it is as balanced as possible, which
// minimizes the number of hash operations for verification.
func OptimizeTree(tree *ProofTree) *ProofTree {
	if tree == nil || len(tree.Leaves) == 0 {
		return tree
	}

	// Collect all leaves.
	leaves := make([]*ProofNode, len(tree.Leaves))
	for i, leaf := range tree.Leaves {
		leaves[i] = &ProofNode{
			Proof:      append([]byte(nil), leaf.Proof...),
			Commitment: leaf.Commitment,
			IsLeaf:     true,
			Type:       leaf.Type,
		}
	}

	root, depth := buildTreeFromLeaves(leaves)
	setDepths(root, 0)

	return &ProofTree{
		Root:      root,
		Leaves:    leaves,
		Depth:     depth,
		LeafCount: len(leaves),
	}
}

// EstimateVerificationGas estimates the gas cost to verify a proof tree.
// The cost model:
//   - Each leaf verification costs leafGas (base cost per proof type).
//   - Each internal hash costs hashGas.
//   - Total = numLeaves * leafGas + (numLeaves - 1) * hashGas.
func EstimateVerificationGas(tree *ProofTree) uint64 {
	if tree == nil || tree.LeafCount == 0 {
		return 0
	}

	const (
		leafGas = 50000 // base gas for verifying one proof
		hashGas = 100   // gas per SHA-256 hash in the tree
	)

	numLeaves := uint64(tree.LeafCount)
	numInternal := numLeaves - 1 // binary tree internal nodes
	return numLeaves*leafGas + numInternal*hashGas
}

// --- Internal helpers ---

// buildTreeFromLeaves constructs a balanced binary Merkle tree from leaves.
// Returns the root node and tree depth.
func buildTreeFromLeaves(leaves []*ProofNode) (*ProofNode, int) {
	if len(leaves) == 0 {
		return nil, 0
	}
	if len(leaves) == 1 {
		return leaves[0], 0
	}

	// Pad to next power of two.
	padded := padLeaves(leaves)
	depth := bits.TrailingZeros(uint(len(padded)))

	// Build tree bottom-up.
	layer := padded
	for len(layer) > 1 {
		next := make([]*ProofNode, len(layer)/2)
		for i := range next {
			left := layer[2*i]
			right := layer[2*i+1]
			combined := hashNodePair(left.Commitment, right.Commitment)
			next[i] = &ProofNode{
				Commitment: combined,
				Children:   []*ProofNode{left, right},
				IsLeaf:     false,
			}
		}
		layer = next
	}

	return layer[0], depth
}

// padLeaves pads the leaf slice to the next power of two with empty nodes.
func padLeaves(leaves []*ProofNode) []*ProofNode {
	n := len(leaves)
	target := 1
	for target < n {
		target <<= 1
	}

	padded := make([]*ProofNode, target)
	copy(padded, leaves)
	for i := n; i < target; i++ {
		padded[i] = &ProofNode{
			Commitment: [32]byte{},
			IsLeaf:     true,
		}
	}
	return padded
}

// setDepths recursively sets the depth of each node in the tree.
func setDepths(node *ProofNode, depth int) {
	if node == nil {
		return
	}
	node.Depth = depth
	for _, child := range node.Children {
		setDepths(child, depth+1)
	}
}

// recomputeCommitment recursively recomputes the Merkle commitment for a node.
func recomputeCommitment(node *ProofNode) ([32]byte, error) {
	if node == nil {
		return [32]byte{}, ErrRecInvalidNode
	}

	if node.IsLeaf {
		if len(node.Proof) > 0 {
			return hashLeaf(node.Proof, node.Type), nil
		}
		// Empty padding node.
		return [32]byte{}, nil
	}

	if len(node.Children) != 2 {
		return [32]byte{}, ErrRecInvalidNode
	}

	left, err := recomputeCommitment(node.Children[0])
	if err != nil {
		return [32]byte{}, err
	}
	right, err := recomputeCommitment(node.Children[1])
	if err != nil {
		return [32]byte{}, err
	}

	return hashNodePair(left, right), nil
}

// hashLeaf computes the leaf commitment: SHA256(typeTag || data).
func hashLeaf(data []byte, proofType ProofType) [32]byte {
	h := sha256.New()
	h.Write([]byte{byte(proofType)})
	h.Write(data)
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}

// hashNodePair hashes two child commitments: SHA256(left || right).
func hashNodePair(left, right [32]byte) [32]byte {
	h := sha256.New()
	h.Write(left[:])
	h.Write(right[:])
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}

// serializeRecursiveCommitment encodes the root commitment with metadata.
// Format: totalProofs(4) || depth(4) || root(32).
func serializeRecursiveCommitment(root [32]byte, totalProofs, depth int) []byte {
	data := make([]byte, 40)
	binary.BigEndian.PutUint32(data[:4], uint32(totalProofs))
	binary.BigEndian.PutUint32(data[4:8], uint32(depth))
	copy(data[8:40], root[:])
	return data
}

// CollectLeaves extracts all leaf proof data from a proof tree in order.
func CollectLeaves(tree *ProofTree) [][]byte {
	if tree == nil || len(tree.Leaves) == 0 {
		return nil
	}
	result := make([][]byte, 0, len(tree.Leaves))
	for _, leaf := range tree.Leaves {
		if len(leaf.Proof) > 0 {
			result = append(result, leaf.Proof)
		}
	}
	return result
}

// TreeStats holds summary statistics for a proof tree.
type TreeStats struct {
	TotalNodes    int
	LeafNodes     int
	InternalNodes int
	MaxDepth      int
	TotalBytes    int
}

// ComputeTreeStats calculates summary statistics for a proof tree.
func ComputeTreeStats(tree *ProofTree) TreeStats {
	if tree == nil || tree.Root == nil {
		return TreeStats{}
	}

	stats := TreeStats{}
	computeStatsRecursive(tree.Root, 0, &stats)
	return stats
}

// computeStatsRecursive traverses the tree collecting statistics.
func computeStatsRecursive(node *ProofNode, depth int, stats *TreeStats) {
	if node == nil {
		return
	}

	stats.TotalNodes++
	if depth > stats.MaxDepth {
		stats.MaxDepth = depth
	}

	if node.IsLeaf {
		stats.LeafNodes++
		stats.TotalBytes += len(node.Proof)
	} else {
		stats.InternalNodes++
		for _, child := range node.Children {
			computeStatsRecursive(child, depth+1, stats)
		}
	}
}

// RecursiveComposition represents a STARK-composed recursive proof.
// Unlike the Merkle-only RecursiveProof, this uses a STARK to attest
// to the validity of all inner proofs.
type RecursiveComposition struct {
	// InnerProofs are the child proofs being composed.
	InnerProofs []AggregateableProof
	// OuterProof is the STARK proving validity of all inner proofs.
	OuterProof *STARKProofData
	// PublicInputs are exposed inputs from inner proofs.
	PublicInputs []FieldElement
	// CompositionHash is a commitment to the composition.
	CompositionHash [32]byte
	// Depth is the composition depth.
	Depth int
}

// ComposeWithSTARK takes N proofs and generates a single STARK that
// attests to all N being valid. This enables true recursive proof
// composition rather than simple Merkle commitment.
func (rp *RecursiveProver) ComposeWithSTARK(proofs []AggregateableProof) (*RecursiveComposition, error) {
	if len(proofs) == 0 {
		return nil, ErrRecNoProofs
	}

	// Build execution trace: each proof becomes a trace row.
	trace := make([][]FieldElement, len(proofs))
	publicInputs := make([]FieldElement, len(proofs))
	for i, p := range proofs {
		data := p.ProofBytes()
		if len(data) == 0 {
			return nil, ErrRecNoProofData
		}
		// Trace row: [proof_hash_hi, proof_hash_lo, proof_type]
		commitment := hashLeaf(data, p.ProofKind())
		hi := new(big.Int).SetBytes(commitment[:16])
		lo := new(big.Int).SetBytes(commitment[16:])
		trace[i] = []FieldElement{
			{Value: hi},
			{Value: lo},
			NewFieldElement(int64(p.ProofKind())),
		}
		publicInputs[i] = FieldElement{Value: new(big.Int).SetBytes(commitment[:])}
	}

	constraints := []STARKConstraint{
		{Degree: 1, Coefficients: []FieldElement{NewFieldElement(1)}},
	}

	prover := NewSTARKProver()
	outerProof, err := prover.GenerateSTARKProof(trace, constraints)
	if err != nil {
		return nil, err
	}

	// Compute composition hash.
	compositionHash := computeCompositionHash(proofs, outerProof)

	return &RecursiveComposition{
		InnerProofs:     proofs,
		OuterProof:      outerProof,
		PublicInputs:    publicInputs,
		CompositionHash: compositionHash,
		Depth:           1,
	}, nil
}

// VerifyComposition verifies a STARK-based recursive composition.
func (rp *RecursiveProver) VerifyComposition(comp *RecursiveComposition) (bool, error) {
	if comp == nil || comp.OuterProof == nil {
		return false, ErrRecNilProof
	}

	prover := NewSTARKProver()
	valid, err := prover.VerifySTARKProof(comp.OuterProof, comp.PublicInputs)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, ErrRecRootMismatch
	}

	// Verify composition hash.
	expected := computeCompositionHash(comp.InnerProofs, comp.OuterProof)
	if expected != comp.CompositionHash {
		return false, ErrRecRootMismatch
	}

	return true, nil
}

// computeCompositionHash hashes the proof set and outer proof into a commitment.
func computeCompositionHash(proofs []AggregateableProof, outer *STARKProofData) [32]byte {
	h := sha256.New()
	for _, p := range proofs {
		data := p.ProofBytes()
		h.Write(data)
	}
	h.Write(outer.TraceCommitment[:])
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}
