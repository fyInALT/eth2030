package chain

import (
	"encoding/json"
	"errors"

	"github.com/eth2030/eth2030/core/rawdb"
	"github.com/eth2030/eth2030/core/types"
)

var bogotaPayloadPrefix = []byte("engine/bogota/")

// BogotaPayload records the driving CL's list view. Block holds an accepted
// block until execution has completed. An empty Block stores a known verdict.
type BogotaPayload struct {
	Number        uint64
	Block         []byte
	InclusionList [][]byte
	Satisfied     bool
}

func bogotaPayloadKey(hash types.Hash) []byte {
	return append(append([]byte(nil), bogotaPayloadPrefix...), hash[:]...)
}

func (bc *Blockchain) ReadBogotaPayload(hash types.Hash) (*BogotaPayload, error) {
	if bc.db == nil {
		return nil, errors.New("Bogotá payload storage unavailable")
	}
	data, err := bc.db.Get(bogotaPayloadKey(hash))
	if errors.Is(err, rawdb.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var record BogotaPayload
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (bc *Blockchain) WriteBogotaPayload(hash types.Hash, record *BogotaPayload) error {
	if bc.db == nil {
		return errors.New("Bogotá payload storage unavailable")
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return bc.db.Put(bogotaPayloadKey(hash), data)
}

// PruneBogotaPayloads releases views below the finalized checkpoint. The
// checkpoint's own verdict is retained for fork-choice updates with no head advance.
func (bc *Blockchain) PruneBogotaPayloads(finalized uint64) error {
	if bc.db == nil {
		return nil
	}
	iter := bc.db.NewIterator(bogotaPayloadPrefix)
	defer iter.Release()
	batch := bc.db.NewBatch()
	for iter.Next() {
		var record BogotaPayload
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			return err
		}
		if record.Number < finalized {
			if err := batch.Delete(append([]byte(nil), iter.Key()...)); err != nil {
				return err
			}
		}
	}
	return batch.Write()
}
