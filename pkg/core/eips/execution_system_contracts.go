package eips

import (
	"encoding/hex"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
)

// Canonical EIP-2935/4788 runtime bytes, excluding their nine-byte synthetic
// deployment constructors. Source: EIPs commit 3db38a3, both Deployment sections.
const historyStorageRuntimeHex = "3373fffffffffffffffffffffffffffffffffffffffe14604657602036036042575f35600143038111604257611fff81430311604257611fff9006545f5260205ff35b5f5ffd5b5f35611fff60014303065500"
const beaconRootsRuntimeHex = "3373fffffffffffffffffffffffffffffffffffffffe14604d57602036146024575f5ffd5b5f35801560495762001fff810690815414603c575f5ffd5b62001fff01545f5260205ff35b5f5ffd5b62001fff42064281555f359062001fff015500"

func HistoryStorageRuntime() []byte {
	code, _ := hex.DecodeString(historyStorageRuntimeHex)
	return code
}
func BeaconRootsRuntime() []byte { code, _ := hex.DecodeString(beaconRootsRuntimeHex); return code }

// ApplyExecutionSystemGenesis constructs explicit owned genesis allocations.
// Never call this when importing supplied state or activating a fork. Existing
// code/nonce allocations are preserved, including intentional custom runtimes.
func ApplyExecutionSystemGenesis(db state.StateDB) {
	for address, code := range map[types.Address][]byte{HistoryStorageAddress: HistoryStorageRuntime(), BeaconRootAddress: BeaconRootsRuntime()} {
		if len(db.GetCode(address)) != 0 {
			continue
		}
		if !db.Exist(address) {
			db.CreateAccount(address)
		}
		db.SetCode(address, code)
		if db.GetNonce(address) == 0 {
			db.SetNonce(address, 1)
		}
	}
}
