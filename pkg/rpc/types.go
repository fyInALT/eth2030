// Package rpc provides JSON-RPC 2.0 types and the standard Ethereum
// JSON-RPC API (eth_ namespace) for the ETH2030 execution client.
package rpc

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/eth2030/eth2030/core/types"
)

// BlockNumber represents a block number parameter in JSON-RPC.
type BlockNumber int64

const (
	LatestBlockNumber    BlockNumber = -1
	PendingBlockNumber   BlockNumber = -2
	EarliestBlockNumber  BlockNumber = 0
	SafeBlockNumber      BlockNumber = -3
	FinalizedBlockNumber BlockNumber = -4
)

// UnmarshalJSON implements json.Unmarshaler for block number.
func (bn *BlockNumber) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// Try as integer.
		var n int64
		if err := json.Unmarshal(data, &n); err != nil {
			return fmt.Errorf("invalid block number: %s", string(data))
		}
		*bn = BlockNumber(n)
		return nil
	}
	switch s {
	case "latest":
		*bn = LatestBlockNumber
	case "pending":
		*bn = PendingBlockNumber
	case "earliest":
		*bn = EarliestBlockNumber
	case "safe":
		*bn = SafeBlockNumber
	case "finalized":
		*bn = FinalizedBlockNumber
	default:
		// Parse hex string.
		n, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			return fmt.Errorf("invalid block number: %s", s)
		}
		*bn = BlockNumber(n)
	}
	return nil
}

// Request is a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
	ID      json.RawMessage   `json:"id"`
}

// Response is a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      json.RawMessage `json:"id"`
}

// RPCError is a JSON-RPC 2.0 error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *RPCError) Error() string {
	return e.Message
}

// Error codes.
const (
	ErrCodeParse          = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternal       = -32603

	// ErrCodeHistoryPruned indicates the requested historical data has been
	// pruned per EIP-4444.
	ErrCodeHistoryPruned = -32000
)

// RPCBlock is the JSON representation of a block.
type RPCBlock struct {
	Number           string            `json:"number"`
	Hash             string            `json:"hash"`
	ParentHash       string            `json:"parentHash"`
	Sha3Uncles       string            `json:"sha3Uncles"`
	Miner            string            `json:"miner"`
	StateRoot        string            `json:"stateRoot"`
	TxRoot           string            `json:"transactionsRoot"`
	ReceiptsRoot     string            `json:"receiptsRoot"`
	LogsBloom        string            `json:"logsBloom"`
	Difficulty       string            `json:"difficulty"`
	GasLimit         string            `json:"gasLimit"`
	GasUsed          string            `json:"gasUsed"`
	Timestamp        string            `json:"timestamp"`
	ExtraData        string            `json:"extraData"`
	MixHash          string            `json:"mixHash"`
	Nonce            string            `json:"nonce"`
	Size             string            `json:"size"`
	BaseFeePerGas    *string           `json:"baseFeePerGas,omitempty"`
	WithdrawalsRoot  *string           `json:"withdrawalsRoot,omitempty"`
	BlobGasUsed      *string           `json:"blobGasUsed,omitempty"`
	ExcessBlobGas    *string           `json:"excessBlobGas,omitempty"`
	ParentBeaconRoot *string           `json:"parentBeaconBlockRoot,omitempty"`
	RequestsHash     *string           `json:"requestsHash,omitempty"`
	Transactions     []string          `json:"transactions"` // tx hashes
	Uncles           []string          `json:"uncles"`
	// Withdrawals is nil for pre-Shanghai blocks and a (possibly empty) slice
	// for post-Shanghai blocks. Using a pointer preserves the distinction so
	// that post-Shanghai empty blocks emit "withdrawals": [] instead of
	// omitting the field entirely (which breaks EIP-4895 clients).
	Withdrawals *[]*RPCWithdrawal `json:"withdrawals,omitempty"`
}

// RPCTransaction is the JSON representation of a transaction.
type RPCTransaction struct {
	Hash             string  `json:"hash"`
	Nonce            string  `json:"nonce"`
	BlockHash        *string `json:"blockHash"`
	BlockNumber      *string `json:"blockNumber"`
	TransactionIndex *string `json:"transactionIndex"`
	From             string  `json:"from"`
	To               *string `json:"to"`
	Value            string  `json:"value"`
	Gas              string  `json:"gas"`
	GasPrice         string  `json:"gasPrice"`
	Input            string  `json:"input"`
	Type             string  `json:"type"`
	V                string  `json:"v"`
	R                string  `json:"r"`
	S                string  `json:"s"`
}

// RPCReceipt is the JSON representation of a transaction receipt.
type RPCReceipt struct {
	TransactionHash   string    `json:"transactionHash"`
	TransactionIndex  string    `json:"transactionIndex"`
	BlockHash         string    `json:"blockHash"`
	BlockNumber       string    `json:"blockNumber"`
	From              string    `json:"from"`
	To                *string   `json:"to"`
	GasUsed           string    `json:"gasUsed"`
	CumulativeGasUsed string    `json:"cumulativeGasUsed"`
	ContractAddress   *string   `json:"contractAddress"`
	Logs              []*RPCLog `json:"logs"`
	Status            string    `json:"status"`
	LogsBloom         string    `json:"logsBloom"`
	Type              string    `json:"type"`
	EffectiveGasPrice string    `json:"effectiveGasPrice"`

	// EIP-4844 blob transaction fields (only present for blob txs).
	BlobGasUsed  *string `json:"blobGasUsed,omitempty"`
	BlobGasPrice *string `json:"blobGasPrice,omitempty"`
}

// RPCLog is the JSON representation of a contract log event.
type RPCLog struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}

// CallArgs represents the arguments for eth_call and eth_estimateGas.
type CallArgs struct {
	From     *string `json:"from"`
	To       *string `json:"to"`
	Gas      *string `json:"gas"`
	GasPrice *string `json:"gasPrice"`
	Value    *string `json:"value"`
	Data     *string `json:"data"`
	Input    *string `json:"input"`
}

// GetData returns the call input data, preferring "input" over "data".
func (args *CallArgs) GetData() []byte {
	if args.Input != nil {
		return fromHexBytes(*args.Input)
	}
	if args.Data != nil {
		return fromHexBytes(*args.Data)
	}
	return nil
}

// FilterCriteria contains parameters for log filtering.
type FilterCriteria struct {
	FromBlock *BlockNumber `json:"fromBlock"`
	ToBlock   *BlockNumber `json:"toBlock"`
	Addresses []string     `json:"address"`
	Topics    [][]string   `json:"topics"`
}

// RPCBlockWithTxs is the JSON representation of a block with full transaction objects.
type RPCBlockWithTxs struct {
	Number           string            `json:"number"`
	Hash             string            `json:"hash"`
	ParentHash       string            `json:"parentHash"`
	Sha3Uncles       string            `json:"sha3Uncles"`
	Miner            string            `json:"miner"`
	StateRoot        string            `json:"stateRoot"`
	TxRoot           string            `json:"transactionsRoot"`
	ReceiptsRoot     string            `json:"receiptsRoot"`
	LogsBloom        string            `json:"logsBloom"`
	Difficulty       string            `json:"difficulty"`
	GasLimit         string            `json:"gasLimit"`
	GasUsed          string            `json:"gasUsed"`
	Timestamp        string            `json:"timestamp"`
	ExtraData        string            `json:"extraData"`
	MixHash          string            `json:"mixHash"`
	Nonce            string            `json:"nonce"`
	Size             string            `json:"size"`
	BaseFeePerGas    *string           `json:"baseFeePerGas,omitempty"`
	WithdrawalsRoot  *string           `json:"withdrawalsRoot,omitempty"`
	BlobGasUsed      *string           `json:"blobGasUsed,omitempty"`
	ExcessBlobGas    *string           `json:"excessBlobGas,omitempty"`
	ParentBeaconRoot *string           `json:"parentBeaconBlockRoot,omitempty"`
	RequestsHash     *string           `json:"requestsHash,omitempty"`
	Transactions []*RPCTransaction `json:"transactions"`
	Uncles       []string          `json:"uncles"`
	// Withdrawals is nil for pre-Shanghai blocks and a (possibly empty) slice
	// for post-Shanghai blocks (see RPCBlock.Withdrawals).
	Withdrawals *[]*RPCWithdrawal `json:"withdrawals,omitempty"`
}

// RPCWithdrawal is the JSON representation of a beacon-chain withdrawal.
type RPCWithdrawal struct {
	Index          string `json:"index"`
	ValidatorIndex string `json:"validatorIndex"`
	Address        string `json:"address"`
	Amount         string `json:"amount"`
}

// FormatBlock converts a block to its JSON-RPC representation.
// If fullTx is true, returns full transaction objects; otherwise returns tx hashes.
func FormatBlock(block *types.Block, fullTx bool) interface{} {
	header := block.Header()
	if !fullTx {
		rb := FormatHeader(header)
		// Populate tx hashes from block body.
		txs := block.Transactions()
		rb.Transactions = make([]string, len(txs))
		for i, tx := range txs {
			rb.Transactions[i] = encodeHash(tx.Hash())
		}
		rb.Uncles = formatUncleHashes(block.Uncles())
		// Withdrawals: always present (even empty) for post-Shanghai blocks.
		if header.WithdrawalsHash != nil {
			ws := block.Withdrawals()
			wList := make([]*RPCWithdrawal, len(ws))
			for i, w := range ws {
				wList[i] = &RPCWithdrawal{
					Index:          encodeUint64(w.Index),
					ValidatorIndex: encodeUint64(w.ValidatorIndex),
					Address:        encodeAddress(w.Address),
					Amount:         encodeUint64(w.Amount),
				}
			}
			rb.Withdrawals = &wList
		}
		return rb
	}

	difficulty := "0x0"
	if header.Difficulty != nil {
		difficulty = encodeBigInt(header.Difficulty)
	}
	result := &RPCBlockWithTxs{
		Number:       encodeUint64(header.Number.Uint64()),
		Hash:         encodeHash(header.Hash()),
		ParentHash:   encodeHash(header.ParentHash),
		Sha3Uncles:   encodeHash(header.UncleHash),
		Miner:        encodeAddress(header.Coinbase),
		StateRoot:    encodeHash(header.Root),
		TxRoot:       encodeHash(header.TxHash),
		ReceiptsRoot: encodeHash(header.ReceiptHash),
		LogsBloom:    encodeBloom(header.Bloom),
		Difficulty:   difficulty,
		GasLimit:     encodeUint64(header.GasLimit),
		GasUsed:      encodeUint64(header.GasUsed),
		Timestamp:    encodeUint64(header.Time),
		ExtraData:    encodeBytes(header.Extra),
		MixHash:      encodeHash(header.MixDigest),
		Nonce:        fmt.Sprintf("0x%016x", header.Nonce),
		Size:         encodeUint64(header.Size()),
		Uncles:       formatUncleHashes(block.Uncles()),
	}
	if header.BaseFee != nil {
		s := encodeBigInt(header.BaseFee)
		result.BaseFeePerGas = &s
	}
	if header.WithdrawalsHash != nil {
		s := encodeHash(*header.WithdrawalsHash)
		result.WithdrawalsRoot = &s
	}
	if header.BlobGasUsed != nil {
		s := encodeUint64(*header.BlobGasUsed)
		result.BlobGasUsed = &s
	}
	if header.ExcessBlobGas != nil {
		s := encodeUint64(*header.ExcessBlobGas)
		result.ExcessBlobGas = &s
	}
	if header.ParentBeaconRoot != nil {
		s := encodeHash(*header.ParentBeaconRoot)
		result.ParentBeaconRoot = &s
	}
	if header.RequestsHash != nil {
		s := encodeHash(*header.RequestsHash)
		result.RequestsHash = &s
	}

	txs := block.Transactions()
	result.Transactions = make([]*RPCTransaction, len(txs))
	blockHash := block.Hash()
	blockNum := block.NumberU64()
	for i, tx := range txs {
		idx := uint64(i)
		result.Transactions[i] = FormatTransaction(tx, &blockHash, &blockNum, &idx)
	}

	// Withdrawals: always present (even empty) for post-Shanghai blocks.
	if header.WithdrawalsHash != nil {
		ws := block.Withdrawals()
		wList := make([]*RPCWithdrawal, len(ws))
		for i, w := range ws {
			wList[i] = &RPCWithdrawal{
				Index:          encodeUint64(w.Index),
				ValidatorIndex: encodeUint64(w.ValidatorIndex),
				Address:        encodeAddress(w.Address),
				Amount:         encodeUint64(w.Amount),
			}
		}
		result.Withdrawals = &wList
	}

	return result
}

// FormatHeader converts a header to JSON-RPC representation.
func FormatHeader(h *types.Header) *RPCBlock {
	difficulty := "0x0"
	if h.Difficulty != nil {
		difficulty = encodeBigInt(h.Difficulty)
	}
	block := &RPCBlock{
		Number:       encodeUint64(h.Number.Uint64()),
		Hash:         encodeHash(h.Hash()),
		ParentHash:   encodeHash(h.ParentHash),
		Sha3Uncles:   encodeHash(h.UncleHash),
		Miner:        encodeAddress(h.Coinbase),
		StateRoot:    encodeHash(h.Root),
		TxRoot:       encodeHash(h.TxHash),
		ReceiptsRoot: encodeHash(h.ReceiptHash),
		LogsBloom:    encodeBloom(h.Bloom),
		Difficulty:   difficulty,
		GasLimit:     encodeUint64(h.GasLimit),
		GasUsed:      encodeUint64(h.GasUsed),
		Timestamp:    encodeUint64(h.Time),
		ExtraData:    encodeBytes(h.Extra),
		MixHash:      encodeHash(h.MixDigest),
		Nonce:        fmt.Sprintf("0x%016x", h.Nonce),
		Size:         encodeUint64(h.Size()),
		Transactions: []string{},
		Uncles:       []string{},
	}
	if h.BaseFee != nil {
		s := encodeBigInt(h.BaseFee)
		block.BaseFeePerGas = &s
	}
	if h.WithdrawalsHash != nil {
		s := encodeHash(*h.WithdrawalsHash)
		block.WithdrawalsRoot = &s
	}
	if h.BlobGasUsed != nil {
		s := encodeUint64(*h.BlobGasUsed)
		block.BlobGasUsed = &s
	}
	if h.ExcessBlobGas != nil {
		s := encodeUint64(*h.ExcessBlobGas)
		block.ExcessBlobGas = &s
	}
	if h.ParentBeaconRoot != nil {
		s := encodeHash(*h.ParentBeaconRoot)
		block.ParentBeaconRoot = &s
	}
	if h.RequestsHash != nil {
		s := encodeHash(*h.RequestsHash)
		block.RequestsHash = &s
	}
	return block
}

// formatUncleHashes returns uncle hashes as hex strings (empty post-merge).
func formatUncleHashes(uncles []*types.Header) []string {
	if len(uncles) == 0 {
		return []string{}
	}
	hashes := make([]string, len(uncles))
	for i, u := range uncles {
		hashes[i] = encodeHash(u.Hash())
	}
	return hashes
}

func encodeUint64(n uint64) string {
	return "0x" + strconv.FormatUint(n, 16)
}

func encodeBigInt(n *big.Int) string {
	if n == nil {
		return "0x0"
	}
	return "0x" + n.Text(16)
}

func encodeHash(h types.Hash) string {
	return "0x" + fmt.Sprintf("%064x", h[:])
}

func encodeAddress(a types.Address) string {
	return "0x" + fmt.Sprintf("%040x", a[:])
}

func encodeBytes(b []byte) string {
	if len(b) == 0 {
		return "0x"
	}
	return "0x" + fmt.Sprintf("%x", b)
}

func encodeBloom(b types.Bloom) string {
	return fmt.Sprintf("0x%0512x", b[:])
}

// fromHexBytes decodes a hex string (with optional 0x prefix) into bytes.
func fromHexBytes(s string) []byte {
	if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
	}
	if len(s) == 0 {
		return nil
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	b := make([]byte, len(s)/2)
	for i := 0; i < len(b); i++ {
		b[i] = unhex(s[2*i])<<4 | unhex(s[2*i+1])
	}
	return b
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

func parseHexUint64(s string) uint64 {
	if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
	}
	n, _ := strconv.ParseUint(s, 16, 64)
	return n
}

func parseHexBigInt(s string) *big.Int {
	if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
	}
	n := new(big.Int)
	n.SetString(s, 16)
	return n
}

// FormatTransaction converts a transaction to its JSON-RPC representation.
func FormatTransaction(tx *types.Transaction, blockHash *types.Hash, blockNumber *uint64, index *uint64) *RPCTransaction {
	rpcTx := &RPCTransaction{
		Hash:     encodeHash(tx.Hash()),
		Nonce:    encodeUint64(tx.Nonce()),
		Value:    encodeBigInt(tx.Value()),
		Gas:      encodeUint64(tx.Gas()),
		GasPrice: encodeBigInt(tx.GasPrice()),
		Input:    encodeBytes(tx.Data()),
		Type:     encodeUint64(uint64(tx.Type())),
	}

	if sender := tx.Sender(); sender != nil {
		rpcTx.From = encodeAddress(*sender)
	}

	if tx.To() != nil {
		to := encodeAddress(*tx.To())
		rpcTx.To = &to
	}

	if blockHash != nil {
		bh := encodeHash(*blockHash)
		rpcTx.BlockHash = &bh
	}
	if blockNumber != nil {
		bn := encodeUint64(*blockNumber)
		rpcTx.BlockNumber = &bn
	}
	if index != nil {
		idx := encodeUint64(*index)
		rpcTx.TransactionIndex = &idx
	}

	// V, R, S - use "0x0" as default if not available
	rpcTx.V = "0x0"
	rpcTx.R = "0x0"
	rpcTx.S = "0x0"

	return rpcTx
}

// FormatReceipt converts a receipt to its JSON-RPC representation.
func FormatReceipt(receipt *types.Receipt, tx *types.Transaction) *RPCReceipt {
	rpcReceipt := &RPCReceipt{
		TransactionHash:   encodeHash(receipt.TxHash),
		TransactionIndex:  encodeUint64(uint64(receipt.TransactionIndex)),
		BlockHash:         encodeHash(receipt.BlockHash),
		BlockNumber:       encodeBigInt(receipt.BlockNumber),
		GasUsed:           encodeUint64(receipt.GasUsed),
		CumulativeGasUsed: encodeUint64(receipt.CumulativeGasUsed),
		Status:            encodeUint64(receipt.Status),
		LogsBloom:         encodeBloom(receipt.Bloom),
		Type:              encodeUint64(uint64(receipt.Type)),
	}

	// EffectiveGasPrice
	if receipt.EffectiveGasPrice != nil {
		rpcReceipt.EffectiveGasPrice = encodeBigInt(receipt.EffectiveGasPrice)
	} else {
		rpcReceipt.EffectiveGasPrice = "0x0"
	}

	// From and To
	if tx != nil {
		if sender := tx.Sender(); sender != nil {
			rpcReceipt.From = encodeAddress(*sender)
		}
		if tx.To() != nil {
			to := encodeAddress(*tx.To())
			rpcReceipt.To = &to
		}
	}

	// Contract address (only if contract creation)
	if !receipt.ContractAddress.IsZero() {
		ca := encodeAddress(receipt.ContractAddress)
		rpcReceipt.ContractAddress = &ca
	}

	// Logs
	rpcReceipt.Logs = make([]*RPCLog, len(receipt.Logs))
	for i, log := range receipt.Logs {
		rpcReceipt.Logs[i] = FormatLog(log)
	}
	if rpcReceipt.Logs == nil {
		rpcReceipt.Logs = []*RPCLog{}
	}

	// EIP-4844 blob transaction fields
	if receipt.BlobGasUsed > 0 {
		bgu := encodeUint64(receipt.BlobGasUsed)
		rpcReceipt.BlobGasUsed = &bgu
	}
	if receipt.BlobGasPrice != nil {
		bgp := encodeBigInt(receipt.BlobGasPrice)
		rpcReceipt.BlobGasPrice = &bgp
	}

	return rpcReceipt
}

// FormatLog converts a log to its JSON-RPC representation.
func FormatLog(log *types.Log) *RPCLog {
	topics := make([]string, len(log.Topics))
	for i, topic := range log.Topics {
		topics[i] = encodeHash(topic)
	}
	return &RPCLog{
		Address:          encodeAddress(log.Address),
		Topics:           topics,
		Data:             encodeBytes(log.Data),
		BlockNumber:      encodeUint64(log.BlockNumber),
		TransactionHash:  encodeHash(log.TxHash),
		TransactionIndex: encodeUint64(uint64(log.TxIndex)),
		BlockHash:        encodeHash(log.BlockHash),
		LogIndex:         encodeUint64(uint64(log.Index)),
		Removed:          log.Removed,
	}
}
