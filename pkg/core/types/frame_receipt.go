package types

import (
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/rlp"
)

// FrameResult holds the execution result of a single frame within a Frame transaction.
type FrameResult struct {
	Status  uint64
	GasUsed uint64
	Logs    []*Log
}

// FrameTxReceipt is the receipt for a Frame transaction per EIP-8141.
// Receipt payload: [cumulative_gas_used, payer, [frame_receipt, ...]]
// where frame_receipt = [status, gas_used, logs]
type FrameTxReceipt struct {
	CumulativeGasUsed uint64
	Payer             Address
	FrameResults      []FrameResult
}

// EncodeFrameTxReceiptRLP RLP-encodes a FrameTxReceipt per EIP-8141 §receipt:
//
//	0x06 || RLP([cumulative_gas_used, payer, [[status, gas_used, logs], ...]])
func EncodeFrameTxReceiptRLP(r *FrameTxReceipt) ([]byte, error) {
	gasEnc, err := rlp.EncodeToBytes(r.CumulativeGasUsed)
	if err != nil {
		return nil, fmt.Errorf("frame receipt encode gas: %w", err)
	}
	payerEnc, err := rlp.EncodeToBytes(r.Payer)
	if err != nil {
		return nil, fmt.Errorf("frame receipt encode payer: %w", err)
	}

	// Encode frame results list.
	var framesPayload []byte
	for i, fr := range r.FrameResults {
		frEnc, err := encodeFrameResult(&fr)
		if err != nil {
			return nil, fmt.Errorf("frame receipt encode frame %d: %w", i, err)
		}
		framesPayload = append(framesPayload, frEnc...)
	}
	framesList := rlp.WrapList(framesPayload)

	var payload []byte
	payload = append(payload, gasEnc...)
	payload = append(payload, payerEnc...)
	payload = append(payload, framesList...)

	encoded := rlp.WrapList(payload)
	// Prefix with transaction type byte.
	result := make([]byte, 1+len(encoded))
	result[0] = FrameTxType
	copy(result[1:], encoded)
	return result, nil
}

// encodeFrameResult RLP-encodes a single FrameResult as [status, gas_used, logs].
func encodeFrameResult(fr *FrameResult) ([]byte, error) {
	statusEnc, err := rlp.EncodeToBytes(fr.Status)
	if err != nil {
		return nil, err
	}
	gasEnc, err := rlp.EncodeToBytes(fr.GasUsed)
	if err != nil {
		return nil, err
	}
	var logsPayload []byte
	for _, l := range fr.Logs {
		lEnc, err := encodeLog(l)
		if err != nil {
			return nil, err
		}
		logsPayload = append(logsPayload, lEnc...)
	}
	logsList := rlp.WrapList(logsPayload)

	var payload []byte
	payload = append(payload, statusEnc...)
	payload = append(payload, gasEnc...)
	payload = append(payload, logsList...)
	return rlp.WrapList(payload), nil
}

// DecodeFrameTxReceiptRLP decodes an RLP-encoded FrameTxReceipt (with type prefix).
func DecodeFrameTxReceiptRLP(data []byte) (*FrameTxReceipt, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("frame receipt decode: too short")
	}
	if data[0] != FrameTxType {
		return nil, fmt.Errorf("frame receipt decode: expected type 0x%02x, got 0x%02x", FrameTxType, data[0])
	}
	s := rlp.NewStreamFromBytes(data[1:])
	if _, err := s.List(); err != nil {
		return nil, fmt.Errorf("frame receipt decode outer list: %w", err)
	}

	r := &FrameTxReceipt{}
	var err error
	r.CumulativeGasUsed, err = s.Uint64()
	if err != nil {
		return nil, fmt.Errorf("frame receipt decode gas: %w", err)
	}
	if err := decodeAddress(s, &r.Payer); err != nil {
		return nil, fmt.Errorf("frame receipt decode payer: %w", err)
	}

	// Decode frame results list.
	if _, err := s.List(); err != nil {
		return nil, fmt.Errorf("frame receipt decode frames list: %w", err)
	}
	for !s.AtListEnd() {
		fr, err := decodeFrameResult(s)
		if err != nil {
			return nil, fmt.Errorf("frame receipt decode frame: %w", err)
		}
		r.FrameResults = append(r.FrameResults, *fr)
	}
	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("frame receipt decode frames list end: %w", err)
	}
	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("frame receipt decode outer list end: %w", err)
	}
	return r, nil
}

// decodeFrameResult decodes a single [status, gas_used, logs] from the stream.
func decodeFrameResult(s *rlp.Stream) (*FrameResult, error) {
	if _, err := s.List(); err != nil {
		return nil, err
	}
	fr := &FrameResult{}
	var err error
	fr.Status, err = s.Uint64()
	if err != nil {
		return nil, err
	}
	fr.GasUsed, err = s.Uint64()
	if err != nil {
		return nil, err
	}
	// Decode logs list.
	if _, err := s.List(); err != nil {
		return nil, err
	}
	for !s.AtListEnd() {
		l, err := decodeLog(s)
		if err != nil {
			return nil, err
		}
		fr.Logs = append(fr.Logs, l)
	}
	if err := s.ListEnd(); err != nil {
		return nil, err
	}
	if err := s.ListEnd(); err != nil {
		return nil, err
	}
	return fr, nil
}

// TotalGasUsed returns the sum of gas used across all frame results.
func (r *FrameTxReceipt) TotalGasUsed() uint64 {
	var total uint64
	for _, fr := range r.FrameResults {
		total += fr.GasUsed
	}
	return total
}

// AllLogs returns all logs from all frame results, in order.
func (r *FrameTxReceipt) AllLogs() []*Log {
	var logs []*Log
	for _, fr := range r.FrameResults {
		logs = append(logs, fr.Logs...)
	}
	return logs
}

// DeriveStatus derives the overall transaction status from frame results.
// Per EIP-8141: the transaction is successful if all frames succeeded.
// Returns 1 (success) if all frames have status 1, otherwise 0 (failed).
func (r *FrameTxReceipt) DeriveStatus() uint64 {
	for _, fr := range r.FrameResults {
		if fr.Status != ReceiptStatusSuccessful {
			return ReceiptStatusFailed
		}
	}
	return ReceiptStatusSuccessful
}

// ComputeBloom computes the bloom filter from all logs in the receipt.
func (r *FrameTxReceipt) ComputeBloom() Bloom {
	return LogsBloom(r.AllLogs())
}

// ToReceipt converts a FrameTxReceipt to a standard Receipt.
// This enables compatibility with existing receipt storage and RPC APIs.
// The caller must provide transaction context (hash, block info, etc.).
func (r *FrameTxReceipt) ToReceipt(txHash Hash, blockHash Hash, blockNumber *big.Int, txIndex uint, effectiveGasPrice *big.Int) *Receipt {
	return &Receipt{
		// Consensus fields
		Type:              FrameTxType,
		Status:            r.DeriveStatus(),
		CumulativeGasUsed: r.CumulativeGasUsed,
		Bloom:             r.ComputeBloom(),
		Logs:              r.AllLogs(),

		// Derived fields
		TxHash:            txHash,
		GasUsed:           r.TotalGasUsed(),
		EffectiveGasPrice: effectiveGasPrice,

		// Inclusion information
		BlockHash:        blockHash,
		BlockNumber:      blockNumber,
		TransactionIndex: txIndex,
	}
}

// ToReceiptWithPayer converts a FrameTxReceipt to a standard Receipt
// with additional payer information stored in the ContractAddress field
// (repurposed for FrameTx to indicate the gas payer).
func (r *FrameTxReceipt) ToReceiptWithPayer(txHash Hash, blockHash Hash, blockNumber *big.Int, txIndex uint, effectiveGasPrice *big.Int) *Receipt {
	receipt := r.ToReceipt(txHash, blockHash, blockNumber, txIndex, effectiveGasPrice)
	// Store payer address in ContractAddress field for FrameTx receipts.
	// This is a repurposing of the field since FrameTx cannot create contracts.
	receipt.ContractAddress = r.Payer
	return receipt
}
