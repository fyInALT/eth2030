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
// The precompiles are added to go-ethereum's Prague map since that's the latest
// fork go-ethereum natively supports.
func (pi *precompileInjector) InjectIntoGethPrecompiles() {
	// Determine the highest fork level configured.
	maxLevel := pi.forkLevelAtTime(0)
	if maxLevel <= geth.ForkLevelPrague {
		return // No custom forks active, nothing to inject.
	}

	// Get all custom precompile info and inject into go-ethereum's Prague map.
	for _, info := range geth.ListCustomPrecompiles() {
		if maxLevel >= info.MinFork {
			adapter := geth.NewPrecompileAdapter(info.Contract, info.Name)
			gethvm.PrecompiledContractsPrague[info.Address] = adapter
		}
	}
}
