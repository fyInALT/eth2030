package block

import (
	"crypto/sha256"
	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto/bls"
	"math/big"
	"testing"
)

type readinessPool struct{ transaction *types.Transaction }

func (p readinessPool) Pending() []*types.Transaction { return []*types.Transaction{p.transaction} }

func TestPoolBuilderWaitsForCompleteBlobData(t *testing.T) {
	sender := types.HexToAddress("0x123456")
	receiver := types.HexToAddress("0xabcdef")
	blob := make([]byte, 131072)
	blob[31] = 1
	backend, err := bls.NewGoEthKZGBackend()
	if err != nil {
		t.Fatal(err)
	}
	commitment, err := backend.BlobToCommitment(blob)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := backend.ComputeBlobProof(blob, commitment)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(commitment[:])
	hash[0] = 1
	transaction := types.NewTransaction(&types.BlobTx{Nonce: 0, To: receiver, Gas: 100000, GasTipCap: big.NewInt(1), GasFeeCap: big.NewInt(100), BlobFeeCap: big.NewInt(1), Value: new(big.Int), BlobHashes: []types.Hash{types.Hash(hash)}})
	transaction.SetSender(sender)
	complete := &types.BlobSidecarData{Blobs: [][]byte{blob}, Commitments: [][]byte{commitment[:]}, Proofs: [][]byte{proof[:]}}
	for _, tc := range []struct {
		name    string
		sidecar *types.BlobSidecarData
		want    int
	}{
		{"metadata_only", nil, 0}, {"allocated_but_empty", &types.BlobSidecarData{}, 0}, {"complete", complete, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := state.NewMemoryStateDB()
			db.AddBalance(sender, big.NewInt(1000000000))
			db.SetNonce(receiver, 1)
			transaction.SetBlobSidecar(tc.sidecar)
			builder := NewBlockBuilder(config.TestConfig, nil, readinessPool{transaction})
			builder.SetState(db)
			parent := &types.Header{Number: new(big.Int), Difficulty: new(big.Int), GasLimit: 30000000, BaseFee: big.NewInt(1)}
			slot := uint64(1)
			root := types.Hash{1}
			built, _, err := builder.BuildBlockWithResult(parent, &BuildBlockAttributes{Timestamp: 12, GasLimit: 30000000, SlotNumber: &slot, BeaconRoot: &root})
			if err != nil {
				t.Fatal(err)
			}
			if len(built.Transactions()) != tc.want {
				t.Fatalf("included=%d want=%d", len(built.Transactions()), tc.want)
			}
			if len(built.Transactions()) == 1 && built.Transactions()[0].Hash() != transaction.Hash() {
				t.Fatal("sidecar availability changed signed transaction identity")
			}
		})
	}
}
