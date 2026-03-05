// frame_stark_replacer.go replaces validation frame calldata with STARK proofs.
// This enables block compression by proving frame execution validity without
// retaining the full calldata.
//
// Part of the EL roadmap: proof aggregation and mandatory 3-of-5 proofs (K+).
package vm

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/proofs"
)

// ReplaceValidationFrames proves that all validation frames in a block
// executed correctly. Returns the original block and the STARK proof.
// If no frame txs exist, proof is nil. If proving fails, the original
// block is returned unchanged with a nil proof.
func ReplaceValidationFrames(
	block *types.Block,
	prover proofs.ValidationFrameProver,
) (*types.Block, *proofs.STARKProofData, error) {
	var frameDatas [][]byte
	for _, tx := range block.Transactions() {
		if tx.Type() == types.FrameTxType {
			frames := tx.Frames()
			for _, f := range frames {
				frameDatas = append(frameDatas, f.Data)
			}
		}
	}

	if len(frameDatas) == 0 {
		return block, nil, nil
	}

	proof, err := prover.ProveAllValidationFrames(frameDatas)
	if err != nil {
		return block, nil, nil
	}

	return block, proof, nil
}

// VerifyBlockFrameProof verifies the STARK proof covering all validation frames
// in a block (the import-side counterpart to ReplaceValidationFrames).
//
// Returns true when:
//   - The block has no VERIFY frame txs and proof is nil.
//   - The block has VERIFY frame txs and prover.Verify(proof) returns true.
//
// Returns false when:
//   - The block has frame txs but proof is nil.
//   - prover.Verify returns false.
func VerifyBlockFrameProof(
	block *types.Block,
	proof *proofs.STARKProofData,
	prover proofs.ValidationFrameProver,
) bool {
	hasFrames := false
	for _, tx := range block.Transactions() {
		if tx.Type() == types.FrameTxType && len(tx.Frames()) > 0 {
			hasFrames = true
			break
		}
	}
	if !hasFrames {
		return proof == nil
	}
	if proof == nil {
		return false
	}
	return prover.Verify(proof)
}
