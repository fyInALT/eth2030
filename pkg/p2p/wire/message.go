package wire

import (
	"fmt"

	"github.com/eth2030/eth2030/rlp"
)

// Message represents a devp2p protocol message with an RLP-encoded payload.
type Message struct {
	Code    uint64 // Protocol message code.
	Size    uint32 // Size of the RLP-encoded payload.
	Payload []byte // RLP-encoded payload bytes.
}

// EncodeMessage encodes a value into a Message with the given message code.
func EncodeMessage(code uint64, val interface{}) (Message, error) {
	payload, err := rlp.EncodeToBytes(val)
	if err != nil {
		return Message{}, fmt.Errorf("p2p/wire: failed to encode message 0x%02x: %w", code, err)
	}
	if len(payload) > MaxMessageSize {
		return Message{}, ErrMessageTooLarge
	}
	return Message{
		Code:    code,
		Size:    uint32(len(payload)),
		Payload: payload,
	}, nil
}

// DecodeMessage decodes a Message's payload into the provided value.
func DecodeMessage(msg Message, val interface{}) error {
	if err := rlp.DecodeBytes(msg.Payload, val); err != nil {
		return fmt.Errorf("%w: code 0x%02x: %v", ErrDecode, msg.Code, err)
	}
	return nil
}
