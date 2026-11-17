package chain

import (
	"bytes"
	"testing"

	"github.com/eth2030/eth2030/core/rawdb"
	"github.com/eth2030/eth2030/core/types"
)

func TestBogotaPayloadStorageRetainsFinalizedAndPendingViews(t *testing.T) {
	dir := t.TempDir()
	open := func() *Blockchain {
		t.Helper()
		db, err := rawdb.NewFileDB(dir)
		if err != nil {
			t.Fatal(err)
		}
		return &Blockchain{db: db}
	}
	bc := open()
	older, checkpoint, pending := types.Hash{31: 1}, types.Hash{31: 2}, types.Hash{31: 3}
	for hash, view := range map[types.Hash]*BogotaPayload{
		older:      {Number: 10, Satisfied: true},
		checkpoint: {Number: 11, Satisfied: false},
		pending:    {Number: 12, Block: []byte{1, 2}, InclusionList: [][]byte{{3}}},
	} {
		if err := bc.WriteBogotaPayload(hash, view); err != nil {
			t.Fatal(err)
		}
	}
	if err := bc.PruneBogotaPayloads(11); err != nil {
		t.Fatal(err)
	}
	if err := bc.db.Close(); err != nil {
		t.Fatal(err)
	}
	bc = open()
	t.Cleanup(func() { _ = bc.db.Close() })
	if view, err := bc.ReadBogotaPayload(older); err != nil || view != nil {
		t.Fatalf("pruned view %+v, %v", view, err)
	}
	if view, err := bc.ReadBogotaPayload(checkpoint); err != nil || view == nil || view.Satisfied || len(view.Block) != 0 {
		t.Fatalf("checkpoint verdict %+v, %v", view, err)
	}
	view, err := bc.ReadBogotaPayload(pending)
	if err != nil || view == nil || !bytes.Equal(view.Block, []byte{1, 2}) || len(view.InclusionList) != 1 ||
		!bytes.Equal(view.InclusionList[0], []byte{3}) {
		t.Fatalf("accepted payload view %+v, %v", view, err)
	}
}
