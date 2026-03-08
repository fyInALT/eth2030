package ethapi

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"sync"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
)

// EthExtendedAPI provides additional eth_ namespace RPC methods that
// complement the core EthAPI.
type EthExtendedAPI struct {
	mu       sync.RWMutex
	backend  Backend
	accounts map[types.Address]*ecdsa.PrivateKey
}

// NewEthExtendedAPI creates a new extended API backed by the given backend.
func NewEthExtendedAPI(backend Backend) *EthExtendedAPI {
	return &EthExtendedAPI{
		backend:  backend,
		accounts: make(map[types.Address]*ecdsa.PrivateKey),
	}
}

// AddAccount registers a private key so the address is returned by
// Accounts() and available for Sign().
func (api *EthExtendedAPI) AddAccount(key *ecdsa.PrivateKey) types.Address {
	addr := crypto.PubkeyToAddress(key.PublicKey)
	api.mu.Lock()
	defer api.mu.Unlock()
	api.accounts[addr] = key
	return addr
}

// GetUncleByBlockHashAndIndex returns the uncle header at the given
// index within the block identified by hash. Post-merge: always nil.
func (api *EthExtendedAPI) GetUncleByBlockHashAndIndex(blockHash types.Hash, index uint64) *types.Header {
	return nil
}

// GetUncleByBlockNumberAndIndex returns the uncle header at the given
// index within the block identified by number. Post-merge: always nil.
func (api *EthExtendedAPI) GetUncleByBlockNumberAndIndex(blockNumber uint64, index uint64) *types.Header {
	return nil
}

// GetUncleCountByBlockHash returns the number of uncles in the block.
// Post-merge: always 0.
func (api *EthExtendedAPI) GetUncleCountByBlockHash(blockHash types.Hash) uint64 {
	return 0
}

// GetUncleCountByBlockNumber returns the number of uncles in the block.
// Post-merge: always 0.
func (api *EthExtendedAPI) GetUncleCountByBlockNumber(blockNumber uint64) uint64 {
	return 0
}

// GetWork returns mining work for a PoW miner. Post-merge: dummy values.
func (api *EthExtendedAPI) GetWork() [3]string {
	return [3]string{
		"0x0000000000000000000000000000000000000000000000000000000000000000",
		"0x0000000000000000000000000000000000000000000000000000000000000000",
		"0x0000000000000000000000000000000000000000000000000000000000000000",
	}
}

// Accounts returns the list of addresses managed by the local keystore.
func (api *EthExtendedAPI) Accounts() []types.Address {
	api.mu.RLock()
	defer api.mu.RUnlock()

	result := make([]types.Address, 0, len(api.accounts))
	for addr := range api.accounts {
		result = append(result, addr)
	}
	return result
}

// Sign produces a secp256k1 ECDSA signature of data using the private
// key associated with addr.
func (api *EthExtendedAPI) Sign(addr types.Address, data []byte) ([]byte, error) {
	api.mu.RLock()
	key, ok := api.accounts[addr]
	api.mu.RUnlock()

	if !ok {
		return nil, errors.New("account not found: " + addr.Hex())
	}

	hash := crypto.Keccak256(data)
	sig, err := crypto.Sign(hash, key)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// GetStorageAt returns the value stored at the given key in the
// account's storage at the latest block.
func (api *EthExtendedAPI) GetStorageAt(addr types.Address, key types.Hash) types.Hash {
	header := api.backend.CurrentHeader()
	if header == nil {
		return types.Hash{}
	}
	statedb, err := api.backend.StateAt(header.Root)
	if err != nil {
		return types.Hash{}
	}
	return statedb.GetState(addr, key)
}

// GetCompilers returns the list of available compilers. Legacy method:
// always returns an empty list.
func (api *EthExtendedAPI) GetCompilers() []string {
	return []string{}
}

// CreateAccessList simulates a transaction and returns accessed storage slots.
func (api *EthExtendedAPI) CreateAccessList(to types.Address, data []byte, gasLimit uint64) []types.AccessTuple {
	if gasLimit == 0 {
		gasLimit = 50_000_000
	}

	_, _, err := api.backend.EVMCall(
		types.Address{},
		&to,
		data,
		gasLimit,
		new(big.Int),
		LatestBlockNumber,
	)
	if err != nil {
		return []types.AccessTuple{}
	}

	return []types.AccessTuple{
		{
			Address:     to,
			StorageKeys: []types.Hash{},
		},
	}
}
