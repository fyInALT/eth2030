package eth

import (
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/p2p"
	"github.com/eth2030/eth2030/rlp"
)

// ETH/68 message code constants. These mirror the canonical definitions in
// the p2p package but are re-exported here for ergonomic use by callers
// that only depend on the eth package.
const (
	MsgStatus                     uint64 = 0x00
	MsgNewBlockHashes             uint64 = 0x01
	MsgTransactions               uint64 = 0x02
	MsgGetBlockHeaders            uint64 = 0x03
	MsgBlockHeaders               uint64 = 0x04
	MsgGetBlockBodies             uint64 = 0x05
	MsgBlockBodies                uint64 = 0x06
	MsgNewBlock                   uint64 = 0x07
	MsgNewPooledTransactionHashes uint64 = 0x08
	MsgGetPooledTransactions      uint64 = 0x09
	MsgPooledTransactions         uint64 = 0x0a
)

// StatusMessage is the eth/68 status handshake message. It is exchanged once
// on connection establishment to verify protocol compatibility.
type StatusMessage struct {
	ProtocolVersion uint32
	NetworkID       uint64
	TD              *big.Int
	BestHash        types.Hash
	Genesis         types.Hash
	ForkID          p2p.ForkID
}

// NewBlockHashesMessage announces new block hashes available on a peer.
type NewBlockHashesMessage struct {
	Entries []BlockHashEntry
}

// BlockHashEntry pairs a block hash with its number.
type BlockHashEntry struct {
	Hash   types.Hash
	Number uint64
}

// TransactionsMessage carries a list of full transactions propagated between peers.
type TransactionsMessage struct {
	Transactions []*types.Transaction
}

// GetBlockHeadersMessage requests block headers by origin, amount, skip, and direction.
type GetBlockHeadersMessage struct {
	Origin  p2p.HashOrNumber
	Amount  uint64
	Skip    uint64
	Reverse bool
}

// BlockHeadersMessage is a response containing requested block headers.
type BlockHeadersMessage struct {
	Headers []*types.Header
}

// GetBlockBodiesMessage requests block bodies for the specified hashes.
type GetBlockBodiesMessage struct {
	Hashes []types.Hash
}

// BlockBodyData holds the transactions, uncles, and (post-Shanghai) withdrawals
// of a single block. Withdrawals is nil for pre-Shanghai blocks.
type BlockBodyData struct {
	Transactions []*types.Transaction
	Uncles       []*types.Header
	Withdrawals  []*types.Withdrawal // nil = pre-Shanghai
}

// BlockBodiesMessage is a response containing requested block bodies.
type BlockBodiesMessage struct {
	Bodies []BlockBodyData
}

// NewBlockMessage announces a newly mined block along with the total difficulty.
type NewBlockMessage struct {
	Block *types.Block
	TD    *big.Int
}

// NewPooledTransactionHashesMsg68 announces new transaction hashes along with
// their types and sizes, as defined in the eth/68 protocol.
type NewPooledTxHashesMsg68 struct {
	Types  []byte
	Sizes  []uint32
	Hashes []types.Hash
}

// GetPooledTransactionsMessage requests specific transactions from a peer's pool.
type GetPooledTransactionsMessage struct {
	Hashes []types.Hash
}

// PooledTransactionsMessage is a response containing pooled transactions.
type PooledTransactionsMessage struct {
	Transactions []*types.Transaction
}

// GetReceiptsMessage requests receipts for the specified block hashes.
type GetReceiptsMessage struct {
	Hashes []types.Hash
}

// ReceiptsMessage is a response containing receipts grouped by block.
type ReceiptsMessage struct {
	Receipts [][]*types.Receipt
}

// EncodeMsg encodes a message struct for the given code into RLP bytes.
// The caller should provide the correct message type for the given code.
func EncodeMsg(code uint64, msg interface{}) ([]byte, error) {
	switch code {
	case MsgStatus:
		sm, ok := msg.(*StatusMessage)
		if !ok {
			return nil, fmt.Errorf("eth: EncodeMsg: expected *StatusMessage for code 0x%02x", code)
		}
		return rlp.EncodeToBytes(sm)

	case MsgNewBlockHashes:
		nm, ok := msg.(*NewBlockHashesMessage)
		if !ok {
			return nil, fmt.Errorf("eth: EncodeMsg: expected *NewBlockHashesMessage for code 0x%02x", code)
		}
		return rlp.EncodeToBytes(nm.Entries)

	case MsgTransactions:
		tm, ok := msg.(*TransactionsMessage)
		if !ok {
			return nil, fmt.Errorf("eth: EncodeMsg: expected *TransactionsMessage for code 0x%02x", code)
		}
		return encodeTxsToRLP(tm.Transactions)

	case MsgGetBlockHeaders:
		gm, ok := msg.(*GetBlockHeadersMessage)
		if !ok {
			return nil, fmt.Errorf("eth: EncodeMsg: expected *GetBlockHeadersMessage for code 0x%02x", code)
		}
		return rlp.EncodeToBytes(gm)

	case MsgBlockHeaders:
		bm, ok := msg.(*BlockHeadersMessage)
		if !ok {
			return nil, fmt.Errorf("eth: EncodeMsg: expected *BlockHeadersMessage for code 0x%02x", code)
		}
		return encodeHeadersToRLP(bm.Headers)

	case MsgGetBlockBodies:
		gm, ok := msg.(*GetBlockBodiesMessage)
		if !ok {
			return nil, fmt.Errorf("eth: EncodeMsg: expected *GetBlockBodiesMessage for code 0x%02x", code)
		}
		return rlp.EncodeToBytes(gm.Hashes)

	case MsgBlockBodies:
		bm, ok := msg.(*BlockBodiesMessage)
		if !ok {
			return nil, fmt.Errorf("eth: EncodeMsg: expected *BlockBodiesMessage for code 0x%02x", code)
		}
		return encodeBodyListToRLP(bm.Bodies)

	case MsgNewBlock:
		nm, ok := msg.(*NewBlockMessage)
		if !ok {
			return nil, fmt.Errorf("eth: EncodeMsg: expected *NewBlockMessage for code 0x%02x", code)
		}
		return encodeNewBlockMsg(nm)

	case MsgNewPooledTransactionHashes:
		pm, ok := msg.(*NewPooledTxHashesMsg68)
		if !ok {
			return nil, fmt.Errorf("eth: EncodeMsg: expected *NewPooledTxHashesMsg68 for code 0x%02x", code)
		}
		return rlp.EncodeToBytes(pm)

	case MsgGetPooledTransactions:
		gm, ok := msg.(*GetPooledTransactionsMessage)
		if !ok {
			return nil, fmt.Errorf("eth: EncodeMsg: expected *GetPooledTransactionsMessage for code 0x%02x", code)
		}
		return rlp.EncodeToBytes(gm.Hashes)

	case MsgPooledTransactions:
		pm, ok := msg.(*PooledTransactionsMessage)
		if !ok {
			return nil, fmt.Errorf("eth: EncodeMsg: expected *PooledTransactionsMessage for code 0x%02x", code)
		}
		return encodeTxsToRLP(pm.Transactions)

	default:
		return nil, fmt.Errorf("eth: EncodeMsg: unknown message code 0x%02x", code)
	}
}

// DecodeMsg decodes RLP bytes into the appropriate message struct for the
// given code. It returns the decoded message or an error.
func DecodeMsg(code uint64, data []byte) (interface{}, error) {
	switch code {
	case MsgStatus:
		var m StatusMessage
		if err := rlp.DecodeBytes(data, &m); err != nil {
			return nil, fmt.Errorf("eth: DecodeMsg Status: %w", err)
		}
		return &m, nil

	case MsgNewBlockHashes:
		var entries []BlockHashEntry
		if err := rlp.DecodeBytes(data, &entries); err != nil {
			return nil, fmt.Errorf("eth: DecodeMsg NewBlockHashes: %w", err)
		}
		return &NewBlockHashesMessage{Entries: entries}, nil

	case MsgTransactions:
		txs, err := decodeTxsFromRLP(data)
		if err != nil {
			return nil, fmt.Errorf("eth: DecodeMsg Transactions: %w", err)
		}
		return &TransactionsMessage{Transactions: txs}, nil

	case MsgGetBlockHeaders:
		var m GetBlockHeadersMessage
		if err := rlp.DecodeBytes(data, &m); err != nil {
			return nil, fmt.Errorf("eth: DecodeMsg GetBlockHeaders: %w", err)
		}
		return &m, nil

	case MsgBlockHeaders:
		headers, err := decodeHeadersFromRLP(data)
		if err != nil {
			return nil, fmt.Errorf("eth: DecodeMsg BlockHeaders: %w", err)
		}
		return &BlockHeadersMessage{Headers: headers}, nil

	case MsgGetBlockBodies:
		var hashes []types.Hash
		if err := rlp.DecodeBytes(data, &hashes); err != nil {
			return nil, fmt.Errorf("eth: DecodeMsg GetBlockBodies: %w", err)
		}
		return &GetBlockBodiesMessage{Hashes: hashes}, nil

	case MsgBlockBodies:
		bodies, err := decodeBodyListFromRLP(data)
		if err != nil {
			return nil, fmt.Errorf("eth: DecodeMsg BlockBodies: %w", err)
		}
		return &BlockBodiesMessage{Bodies: bodies}, nil

	case MsgNewBlock:
		return nil, fmt.Errorf("eth: DecodeMsg NewBlock requires special handling; use decodeNewBlock")

	case MsgNewPooledTransactionHashes:
		var m NewPooledTxHashesMsg68
		if err := rlp.DecodeBytes(data, &m); err != nil {
			return nil, fmt.Errorf("eth: DecodeMsg NewPooledTransactionHashes: %w", err)
		}
		return &m, nil

	case MsgGetPooledTransactions:
		var hashes []types.Hash
		if err := rlp.DecodeBytes(data, &hashes); err != nil {
			return nil, fmt.Errorf("eth: DecodeMsg GetPooledTransactions: %w", err)
		}
		return &GetPooledTransactionsMessage{Hashes: hashes}, nil

	case MsgPooledTransactions:
		txs, err := decodeTxsFromRLP(data)
		if err != nil {
			return nil, fmt.Errorf("eth: DecodeMsg PooledTransactions: %w", err)
		}
		return &PooledTransactionsMessage{Transactions: txs}, nil

	default:
		return nil, fmt.Errorf("eth: DecodeMsg: unknown message code 0x%02x", code)
	}
}

// encodeNewBlockMsg encodes a NewBlockMessage by encoding the block and TD
// as a two-element RLP list.
func encodeNewBlockMsg(msg *NewBlockMessage) ([]byte, error) {
	blockEnc, err := msg.Block.EncodeRLP()
	if err != nil {
		return nil, fmt.Errorf("eth: encode block: %w", err)
	}
	td := msg.TD
	if td == nil {
		td = new(big.Int)
	}
	tdEnc, err := rlp.EncodeToBytes(td)
	if err != nil {
		return nil, fmt.Errorf("eth: encode td: %w", err)
	}
	var payload []byte
	payload = append(payload, blockEnc...)
	payload = append(payload, tdEnc...)
	return rlp.WrapList(payload), nil
}

// encodeTxsToRLP encodes a transaction list to RLP bytes.
// Legacy txs are appended directly (RLP lists); typed txs are wrapped as
// RLP byte strings (type_byte || RLP_payload), matching geth's eth protocol.
func encodeTxsToRLP(txs []*types.Transaction) ([]byte, error) {
	var payload []byte
	for i, tx := range txs {
		txEnc, err := tx.EncodeRLP()
		if err != nil {
			return nil, fmt.Errorf("tx %d: %w", i, err)
		}
		if tx.Type() == types.LegacyTxType {
			payload = append(payload, txEnc...)
		} else {
			wrapped, err := rlp.EncodeToBytes(txEnc)
			if err != nil {
				return nil, fmt.Errorf("wrap tx %d: %w", i, err)
			}
			payload = append(payload, wrapped...)
		}
	}
	return rlp.WrapList(payload), nil
}

// decodeTxsFromRLP decodes a transaction list from RLP bytes.
// Legacy txs appear as RLP lists; typed txs appear as RLP byte strings.
func decodeTxsFromRLP(data []byte) ([]*types.Transaction, error) {
	s := rlp.NewStreamFromBytes(data)
	if _, err := s.List(); err != nil {
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
			txData, err = s.RawItem()
		} else {
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
	return txs, s.ListEnd()
}

// encodeHeadersToRLP encodes a list of headers to RLP bytes.
// Each header is encoded in Yellow Paper field order via Header.EncodeRLP.
func encodeHeadersToRLP(headers []*types.Header) ([]byte, error) {
	var payload []byte
	for i, h := range headers {
		hEnc, err := h.EncodeRLP()
		if err != nil {
			return nil, fmt.Errorf("header %d: %w", i, err)
		}
		payload = append(payload, hEnc...)
	}
	return rlp.WrapList(payload), nil
}

// decodeHeadersFromRLP decodes a list of headers from RLP bytes.
func decodeHeadersFromRLP(data []byte) ([]*types.Header, error) {
	s := rlp.NewStreamFromBytes(data)
	if _, err := s.List(); err != nil {
		return nil, fmt.Errorf("open headers list: %w", err)
	}
	var headers []*types.Header
	for !s.AtListEnd() {
		hBytes, err := s.RawItem()
		if err != nil {
			return nil, fmt.Errorf("read header: %w", err)
		}
		h, err := types.DecodeHeaderRLP(hBytes)
		if err != nil {
			return nil, fmt.Errorf("decode header: %w", err)
		}
		headers = append(headers, h)
	}
	return headers, s.ListEnd()
}

// encodeBodyListToRLP encodes a list of block bodies to RLP bytes.
// Each body is encoded as [txs_list, uncles_list] or
// [txs_list, uncles_list, withdrawals_list] for post-Shanghai blocks.
func encodeBodyListToRLP(bodies []BlockBodyData) ([]byte, error) {
	var outer []byte
	for i, bd := range bodies {
		txsEnc, err := encodeTxsToRLP(bd.Transactions)
		if err != nil {
			return nil, fmt.Errorf("body %d txs: %w", i, err)
		}
		unclesEnc, err := encodeHeadersToRLP(bd.Uncles)
		if err != nil {
			return nil, fmt.Errorf("body %d uncles: %w", i, err)
		}
		var bodyPayload []byte
		bodyPayload = append(bodyPayload, txsEnc...)
		bodyPayload = append(bodyPayload, unclesEnc...)
		if bd.Withdrawals != nil {
			var wPayload []byte
			for _, w := range bd.Withdrawals {
				wPayload = append(wPayload, types.EncodeWithdrawal(w)...)
			}
			bodyPayload = append(bodyPayload, rlp.WrapList(wPayload)...)
		}
		outer = append(outer, rlp.WrapList(bodyPayload)...)
	}
	return rlp.WrapList(outer), nil
}

// decodeBodyListFromRLP decodes a list of block bodies from RLP bytes.
func decodeBodyListFromRLP(data []byte) ([]BlockBodyData, error) {
	s := rlp.NewStreamFromBytes(data)
	if _, err := s.List(); err != nil {
		return nil, fmt.Errorf("open bodies list: %w", err)
	}
	var bodies []BlockBodyData
	for !s.AtListEnd() {
		if _, err := s.List(); err != nil {
			return nil, fmt.Errorf("open body: %w", err)
		}
		// Transactions list.
		txListBytes, err := s.RawItem()
		if err != nil {
			return nil, fmt.Errorf("read txs: %w", err)
		}
		txs, err := decodeTxsFromRLP(txListBytes)
		if err != nil {
			return nil, fmt.Errorf("decode txs: %w", err)
		}
		// Uncles list.
		uncleListBytes, err := s.RawItem()
		if err != nil {
			return nil, fmt.Errorf("read uncles: %w", err)
		}
		uncles, err := decodeHeadersFromRLP(uncleListBytes)
		if err != nil {
			return nil, fmt.Errorf("decode uncles: %w", err)
		}
		bd := BlockBodyData{Transactions: txs, Uncles: uncles}
		// Withdrawals (optional, post-Shanghai).
		if !s.AtListEnd() {
			if _, err := s.List(); err != nil {
				return nil, fmt.Errorf("open withdrawals: %w", err)
			}
			bd.Withdrawals = make([]*types.Withdrawal, 0)
			for !s.AtListEnd() {
				wBytes, err := s.RawItem()
				if err != nil {
					return nil, fmt.Errorf("read withdrawal: %w", err)
				}
				w, err := types.DecodeWithdrawal(wBytes)
				if err != nil {
					return nil, fmt.Errorf("decode withdrawal: %w", err)
				}
				bd.Withdrawals = append(bd.Withdrawals, w)
			}
			if err := s.ListEnd(); err != nil {
				return nil, fmt.Errorf("close withdrawals: %w", err)
			}
		}
		if err := s.ListEnd(); err != nil {
			return nil, fmt.Errorf("close body: %w", err)
		}
		bodies = append(bodies, bd)
	}
	return bodies, s.ListEnd()
}

// MsgCodeName returns a human-readable name for an ETH/68 message code.
func MsgCodeName(code uint64) string {
	switch code {
	case MsgStatus:
		return "Status"
	case MsgNewBlockHashes:
		return "NewBlockHashes"
	case MsgTransactions:
		return "Transactions"
	case MsgGetBlockHeaders:
		return "GetBlockHeaders"
	case MsgBlockHeaders:
		return "BlockHeaders"
	case MsgGetBlockBodies:
		return "GetBlockBodies"
	case MsgBlockBodies:
		return "BlockBodies"
	case MsgNewBlock:
		return "NewBlock"
	case MsgNewPooledTransactionHashes:
		return "NewPooledTransactionHashes"
	case MsgGetPooledTransactions:
		return "GetPooledTransactions"
	case MsgPooledTransactions:
		return "PooledTransactions"
	default:
		return fmt.Sprintf("Unknown(0x%02x)", code)
	}
}
