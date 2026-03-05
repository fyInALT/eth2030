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
