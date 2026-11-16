package types

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
)

// FrameResult8141 retains both gas dimensions in the canonical receipt.
type FrameResult8141 struct {
	Status  uint64
	GasUsed FrameGasLimits
	Logs    []*Log
}

const FrameStatusSkipped uint64 = 2

type frame8141LogRLP struct {
	Address Address
	Topics  []Hash
	Data    []byte
}

type frame8141ResultRLP struct {
	Status  uint64
	GasUsed FrameGasLimits
	Logs    []frame8141LogRLP
}

type frame8141ReceiptRLP struct {
	CumulativeGasUsed uint64
	Payer             Address
	Frames            []frame8141ResultRLP
}

func validateFrame8141Receipt(result frame8141ResultRLP) error {
	if result.Status > FrameStatusSkipped {
		return errors.New("invalid frame status")
	}
	if result.Status == FrameStatusSkipped &&
		(result.GasUsed != (FrameGasLimits{}) || len(result.Logs) != 0) {
		return errors.New("skipped frame has gas usage or logs")
	}
	return nil
}

func (r *Receipt) encodeFrame8141Receipt() ([]byte, error) {
	if r.Type != FrameTxType {
		return nil, errors.New("EIP-8141 results on a non-frame receipt")
	}
	if len(r.FrameResults8141) == 0 || len(r.FrameResults8141) > frame8141MaxFrames {
		return nil, errors.New("EIP-8141 receipt has invalid frame count")
	}
	wire := frame8141ReceiptRLP{
		CumulativeGasUsed: r.CumulativeGasUsed,
		Payer:             r.Payer,
		Frames:            make([]frame8141ResultRLP, len(r.FrameResults8141)),
	}
	for i, result := range r.FrameResults8141 {
		frame := frame8141ResultRLP{Status: result.Status, GasUsed: result.GasUsed}
		for _, log := range result.Logs {
			if log == nil {
				return nil, fmt.Errorf("EIP-8141 frame %d has nil log", i)
			}
			frame.Logs = append(frame.Logs, frame8141LogRLP{log.Address, log.Topics, log.Data})
		}
		if err := validateFrame8141Receipt(frame); err != nil {
			return nil, fmt.Errorf("EIP-8141 receipt frame %d: %w", i, err)
		}
		wire.Frames[i] = frame
	}
	payload, err := rlp.EncodeToBytes(wire)
	if err != nil {
		return nil, err
	}
	return append([]byte{FrameTxType}, payload...), nil
}

func decodeFrame8141Receipt(payload []byte) (*Receipt, error) {
	var wire frame8141ReceiptRLP
	if err := rlp.DecodeBytes(payload, &wire); err != nil {
		return nil, fmt.Errorf("decode EIP-8141 receipt: %w", err)
	}
	if len(wire.Frames) == 0 || len(wire.Frames) > frame8141MaxFrames {
		return nil, errors.New("EIP-8141 receipt has invalid frame count")
	}
	r := &Receipt{
		Type: FrameTxType, Status: ReceiptStatusSuccessful, CumulativeGasUsed: wire.CumulativeGasUsed,
		Payer: wire.Payer, FrameResults8141: make([]FrameResult8141, len(wire.Frames)),
	}
	for i, result := range wire.Frames {
		if err := validateFrame8141Receipt(result); err != nil {
			return nil, fmt.Errorf("EIP-8141 receipt frame %d: %w", i, err)
		}
		frame := FrameResult8141{Status: result.Status, GasUsed: result.GasUsed}
		for _, raw := range result.Logs {
			log := &Log{Address: raw.Address, Topics: raw.Topics, Data: raw.Data}
			frame.Logs = append(frame.Logs, log)
			r.Logs = append(r.Logs, log)
		}
		r.FrameResults8141[i] = frame
	}
	r.Bloom = LogsBloom(r.Logs)
	return r, nil
}
