package crypto

// merkle_compat.go re-exports types from crypto/merkle for backward compatibility.

import "github.com/eth2030/eth2030/crypto/merkle"

// Merkle type aliases.
type (
	MerkleMultiProof = merkle.MerkleMultiProof
	MerkleLeaf       = merkle.MerkleLeaf
	MerkleNode       = merkle.MerkleNode
)

// Merkle function wrappers.
func GeneralizedIndex(depth uint, leafPos uint64) uint64 { return merkle.GeneralizedIndex(depth, leafPos) }
func Parent(gi uint64) uint64                           { return merkle.Parent(gi) }
func Sibling(gi uint64) uint64                          { return merkle.Sibling(gi) }
func IsLeft(gi uint64) bool                             { return merkle.IsLeft(gi) }
func DepthOfGI(gi uint64) uint                          { return merkle.DepthOfGI(gi) }
func PathToRoot(gi uint64) []uint64                     { return merkle.PathToRoot(gi) }
func GenerateMultiProof(tree [][32]byte, depth uint, leafIndices []uint64) (*MerkleMultiProof, error) {
	return merkle.GenerateMultiProof(tree, depth, leafIndices)
}
func VerifyMultiProof(root [32]byte, proof *MerkleMultiProof) bool {
	return merkle.VerifyMultiProof(root, proof)
}
func CompactMultiProof(proof *MerkleMultiProof) *MerkleMultiProof {
	return merkle.CompactMultiProof(proof)
}
func BuildMerkleTree(leaves [][32]byte) ([][32]byte, uint) { return merkle.BuildMerkleTree(leaves) }
func MerkleRoot(leaves [][32]byte) [32]byte                { return merkle.MerkleRoot(leaves) }
func ProofSize(depth uint, k int) int                      { return merkle.ProofSize(depth, k) }
