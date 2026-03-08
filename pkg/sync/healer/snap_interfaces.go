// snap_interfaces.go defines local copies of the snap-sync interfaces and
// data types needed by the healer sub-package.  These are consumer-defined
// interfaces (Go structural typing): any concrete implementation that satisfies
// the snap sub-package's SnapPeer / StateWriter also satisfies these.
package healer

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

// AccountData represents a downloaded account with its address hash and state.
type AccountData struct {
	Hash     types.Hash
	Address  types.Address
	Nonce    uint64
	Balance  *big.Int
	Root     types.Hash
	CodeHash types.Hash
}

// StorageData represents a downloaded storage slot.
type StorageData struct {
	AccountHash types.Hash
	SlotHash    types.Hash
	Value       []byte
}

// BytecodeData represents a downloaded contract bytecode.
type BytecodeData struct {
	Hash types.Hash
	Code []byte
}

// AccountRangeRequest requests account trie leaves in a given range.
type AccountRangeRequest struct {
	ID     uint64
	Root   types.Hash
	Origin types.Hash
	Limit  types.Hash
	Bytes  uint64
}

// AccountRangeResponse is the response to an AccountRangeRequest.
type AccountRangeResponse struct {
	ID       uint64
	Accounts []AccountData
	Proof    [][]byte
	More     bool
}

// StorageRangeRequest requests storage trie leaves for a set of accounts.
type StorageRangeRequest struct {
	ID       uint64
	Root     types.Hash
	Accounts []types.Hash
	Origin   types.Hash
	Limit    types.Hash
	Bytes    uint64
}

// StorageRangeResponse is the response to a StorageRangeRequest.
type StorageRangeResponse struct {
	ID    uint64
	Slots []StorageData
	Proof [][]byte
	More  bool
}

// BytecodeRequest requests contract bytecodes by code hash.
type BytecodeRequest struct {
	ID     uint64
	Hashes []types.Hash
}

// BytecodeResponse is the response to a BytecodeRequest.
type BytecodeResponse struct {
	ID    uint64
	Codes []BytecodeData
}

// SnapPeer represents a peer that supports the snap protocol.
type SnapPeer interface {
	ID() string
	RequestAccountRange(req AccountRangeRequest) (*AccountRangeResponse, error)
	RequestStorageRange(req StorageRangeRequest) (*StorageRangeResponse, error)
	RequestBytecodes(req BytecodeRequest) (*BytecodeResponse, error)
	RequestTrieNodes(root types.Hash, paths [][]byte) ([][]byte, error)
}

// StateWriter is the interface for persisting downloaded state data.
type StateWriter interface {
	WriteAccount(hash types.Hash, data AccountData) error
	WriteStorage(accountHash, slotHash types.Hash, value []byte) error
	WriteBytecode(hash types.Hash, code []byte) error
	WriteTrieNode(path []byte, data []byte) error
	HasBytecode(hash types.Hash) bool
	HasTrieNode(path []byte) bool
	MissingTrieNodes(root types.Hash, limit int) [][]byte
}
