package eth

import (
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/p2p"
	"github.com/eth2030/eth2030/rlp"
)

// encodeTransactions encodes a list of transactions into a Msg.
// Legacy txs are appended directly (already RLP lists); typed txs are
// wrapped as RLP byte strings (type_byte || RLP_payload), matching geth.
func encodeTransactions(txs []*types.Transaction) (p2p.Msg, error) {
	var payload []byte
	for i, tx := range txs {
		txEnc, err := tx.EncodeRLP()
		if err != nil {
			return p2p.Msg{}, fmt.Errorf("encode tx %d: %w", i, err)
		}
		if tx.Type() == types.LegacyTxType {
			// Legacy: txEnc is already an RLP list; append directly.
			payload = append(payload, txEnc...)
		} else {
			// Typed: wrap as RLP byte string (type_byte || RLP_payload).
			wrapped, err := rlp.EncodeToBytes(txEnc)
			if err != nil {
				return p2p.Msg{}, fmt.Errorf("wrap tx %d: %w", i, err)
			}
			payload = append(payload, wrapped...)
		}
	}
	data := rlp.WrapList(payload)
	return p2p.Msg{
		Code:    p2p.TransactionsMsg,
		Size:    uint32(len(data)),
		Payload: data,
	}, nil
}

// decodeTransactions decodes a TransactionsMsg payload into transactions.
func decodeTransactions(msg p2p.Msg) ([]*types.Transaction, error) {
	s := rlp.NewStreamFromBytes(msg.Payload)
	_, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("open tx list: %w", err)
	}
	var txs []*types.Transaction
	for !s.AtListEnd() {
		kind, _, err := s.Kind()
		if err != nil {
			return nil, fmt.Errorf("peek tx kind: %w", err)
		}
		var txData []byte
		if kind == rlp.List {
			// Legacy transaction: read entire RLP list item.
			txData, err = s.RawItem()
		} else {
			// Typed transaction: read byte string (type_byte || RLP_payload).
			txData, err = s.Bytes()
		}
		if err != nil {
			return nil, fmt.Errorf("read tx: %w", err)
		}
		tx, err := types.DecodeTxRLP(txData)
		if err != nil {
			return nil, fmt.Errorf("decode tx: %w", err)
		}
		txs = append(txs, tx)
	}
	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("close tx list: %w", err)
	}
	return txs, nil
}

// encodeNewBlock encodes a NewBlockData message.
// Format: RLP([block_rlp, td])
func encodeNewBlock(data *p2p.NewBlockData) (p2p.Msg, error) {
	blockEnc, err := data.Block.EncodeRLP()
	if err != nil {
		return p2p.Msg{}, fmt.Errorf("encode block: %w", err)
	}

	td := data.TD
	if td == nil {
		td = new(big.Int)
	}
	tdEnc, err := rlp.EncodeToBytes(td)
	if err != nil {
		return p2p.Msg{}, fmt.Errorf("encode td: %w", err)
	}

	var payload []byte
	payload = append(payload, blockEnc...)
	payload = append(payload, tdEnc...)
	encoded := rlp.WrapList(payload)

	return p2p.Msg{
		Code:    p2p.NewBlockMsg,
		Size:    uint32(len(encoded)),
		Payload: encoded,
	}, nil
}

// decodeNewBlock decodes a NewBlockMsg payload into NewBlockData.
func decodeNewBlock(msg p2p.Msg) (*p2p.NewBlockData, error) {
	s := rlp.NewStreamFromBytes(msg.Payload)
	_, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("open newblock list: %w", err)
	}

	// The block is an RLP list, read it as raw item.
	blockBytes, err := s.RawItem()
	if err != nil {
		return nil, fmt.Errorf("read block: %w", err)
	}
	block, err := types.DecodeBlockRLP(blockBytes)
	if err != nil {
		return nil, fmt.Errorf("decode block: %w", err)
	}

	// Read TD.
	td, err := s.BigInt()
	if err != nil {
		return nil, fmt.Errorf("read td: %w", err)
	}

	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("close newblock list: %w", err)
	}

	return &p2p.NewBlockData{
		Block: block,
		TD:    td,
	}, nil
}
