package main

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	gethvm "github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/eth2030/eth2030/geth"
)

// precompileInjector manages ETH2030 custom precompile injection into geth
// EVM instances. Custom precompiles only activate at future fork timestamps
// (Glamsterdam, Hogota, I+), so for current mainnet this is a no-op.
type precompileInjector struct {
	glamsterdamTime *uint64
	hogotaTime      *uint64
	iPlusTime       *uint64
}

// newPrecompileInjector creates an injector configured for the given fork schedule.
func newPrecompileInjector(glamsterdam, hogota, iPlus *uint64) *precompileInjector {
	return &precompileInjector{
		glamsterdamTime: glamsterdam,
		hogotaTime:      hogota,
		iPlusTime:       iPlus,
	}
}

// forkLevelAtTime determines the ETH2030 fork level active at the given block time.
func (pi *precompileInjector) forkLevelAtTime(time uint64) geth.Eth2028ForkLevel {
	if pi.iPlusTime != nil && time >= *pi.iPlusTime {
		return geth.ForkLevelIPlus
	}
	if pi.hogotaTime != nil && time >= *pi.hogotaTime {
		return geth.ForkLevelHogota
	}
	if pi.glamsterdamTime != nil && time >= *pi.glamsterdamTime {
		return geth.ForkLevelGlamsterdam
	}
	return geth.ForkLevelPrague
}

// InjectIntoEVM sets custom precompiles on a go-ethereum EVM instance
// if the block time indicates a future ETH2030 fork is active.
func (pi *precompileInjector) InjectIntoEVM(evm *gethvm.EVM, rules params.Rules, blockTime uint64) {
	forkLevel := pi.forkLevelAtTime(blockTime)
	if forkLevel > geth.ForkLevelPrague {
		precompiles := geth.InjectCustomPrecompiles(rules, forkLevel)
		evm.SetPrecompiles(precompiles)
	}
}

// CustomAddresses returns the precompile addresses active at the given block time.
func (pi *precompileInjector) CustomAddresses(blockTime uint64) []gethcommon.Address {
	forkLevel := pi.forkLevelAtTime(blockTime)
	return geth.CustomPrecompileAddresses(forkLevel)
}

// InjectIntoGethPrecompiles patches go-ethereum's package-level precompile maps
// to include eth2030 custom precompiles. This makes them available to eth_call
// and all block processing within go-ethereum's internal pipeline.
//
// This is called at startup when fork overrides are active (e.g., --override.iplus=0).
// Precompiles are added to both Prague and Osaka maps because go-ethereum v1.17.0
// uses PrecompiledContractsOsaka when Osaka (or later) forks are active — which
// is the case on devnets where Glamsterdam/Hogota/I+ overrides set osakaTime=0.
func (pi *precompileInjector) InjectIntoGethPrecompiles() {
	// Determine the highest fork level configured.
	maxLevel := pi.forkLevelAtTime(0)
	if maxLevel <= geth.ForkLevelPrague {
		return // No custom forks active, nothing to inject.
	}

	// Inject into all fork-level precompile maps that go-ethereum might use.
	// ActivePrecompiledContracts() calls maps.Clone() on these package-level
	// maps, so patching here before any RPC call makes them visible to eth_call.
	for _, info := range geth.ListCustomPrecompiles() {
		if maxLevel >= info.MinFork {
			adapter := geth.NewPrecompileAdapter(info.Contract, info.Name)
			gethvm.PrecompiledContractsPrague[info.Address] = adapter
			gethvm.PrecompiledContractsOsaka[info.Address] = adapter
		}
	}

	// Also update the address lists so ActivePrecompiles() includes custom
	// addresses for EIP-2929 access list warming.
	customAddrs := geth.CustomPrecompileAddresses(maxLevel)
	gethvm.PrecompiledAddressesPrague = append(gethvm.PrecompiledAddressesPrague, customAddrs...)
	gethvm.PrecompiledAddressesOsaka = append(gethvm.PrecompiledAddressesOsaka, customAddrs...)
}
