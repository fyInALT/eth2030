package execution

import (
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/bal"
	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
)

// FinalizeBlockExecution applies withdrawals and collects EIP-7685 requests
// on the block's working state. Builders and replayers share this transition.
func FinalizeBlockExecution(cfg *config.ChainConfig, getHash vm.GetHashFunc, statedb state.StateDB, header *types.Header, withdrawals []*types.Withdrawal, receipts []*types.Receipt, blockBAL *bal.BlockAccessList) (types.Requests, error) {
	snapshot := statedb.Snapshot()
	var tracker *bal.AccessTracker
	if blockBAL != nil {
		tracker = bal.NewTracker()
	}
	if cfg != nil && cfg.IsShanghai(header.Time) {
		var before map[types.Address]*big.Int
		if tracker != nil {
			before = make(map[types.Address]*big.Int)
			for _, withdrawal := range withdrawals {
				if withdrawal != nil && before[withdrawal.Address] == nil {
					before[withdrawal.Address] = new(big.Int).Set(statedb.GetBalance(withdrawal.Address))
				}
			}
		}
		ProcessWithdrawals(statedb, withdrawals)
		for address, balance := range before {
			tracker.RecordAddressTouch(address)
			if after := statedb.GetBalance(address); after.Cmp(balance) != 0 {
				tracker.RecordBalanceChange(address, balance, after)
			}
		}
	}

	requests, err := collectRequests(cfg, getHash, statedb, header, receipts, tracker)
	if err != nil {
		statedb.RevertToSnapshot(snapshot)
		return nil, err
	}
	if blockBAL != nil {
		for _, entry := range tracker.Build(uint64(len(receipts) + 1)).Entries {
			blockBAL.AddEntry(entry)
		}
		blockBAL.Sort()
	}
	return requests, nil
}

// ProcessRequests collects EIP-6110 events and drains the active request queues.
// Use FinalizeBlockExecution when processing a whole block.
func ProcessRequests(cfg *config.ChainConfig, statedb state.StateDB, header *types.Header, receipts ...*types.Receipt) (types.Requests, error) {
	snapshot := statedb.Snapshot()
	requests, err := collectRequests(cfg, nil, statedb, header, receipts, nil)
	if err != nil {
		statedb.RevertToSnapshot(snapshot)
	}
	return requests, err
}

func collectRequests(cfg *config.ChainConfig, getHash vm.GetHashFunc, statedb state.StateDB, header *types.Header, receipts []*types.Receipt, tracker *bal.AccessTracker) (types.Requests, error) {
	if cfg == nil || !cfg.IsPrague(header.Time) {
		return nil, nil
	}

	depositAddress := cfg.DepositContractAddress
	if depositAddress.IsZero() {
		depositAddress = types.DepositContractAddress
	}
	deposits, err := types.DepositRequestData(receipts, depositAddress)
	if err != nil {
		return nil, fmt.Errorf("deposit requests: %w", err)
	}
	var requests types.Requests
	if len(deposits) != 0 {
		requests = append(requests, types.NewRequest(types.DepositRequestType, deposits))
	}

	type requestQueue struct {
		address  types.Address
		kind     byte
		size     int
		limit    int
		required bool
	}
	queues := []requestQueue{
		{types.WithdrawalRequestAddress, types.WithdrawalRequestType, types.WithdrawalRequestSSZSize, types.MaxWithdrawalRequestsPerPayload, false},
		{types.ConsolidationRequestAddress, types.ConsolidationRequestType, types.ConsolidationRequestSSZSize, types.MaxConsolidationRequestsPerPayload, false},
	}
	if cfg.IsAmsterdam(header.Time) {
		queues = append(queues,
			requestQueue{types.BuilderDepositRequestAddress, types.BuilderDepositRequestType, types.BuilderDepositRequestSSZSize, types.MaxBuilderDepositRequestsPerPayload, true},
			requestQueue{types.BuilderExitRequestAddress, types.BuilderExitRequestType, types.BuilderExitRequestSSZSize, types.MaxBuilderExitRequestsPerPayload, true},
		)
	}
	for _, queue := range queues {
		if statedb.GetCodeSize(queue.address) == 0 {
			if queue.required || cfg.RequireRequestPredeploys {
				return nil, fmt.Errorf("request type %d: no code at %s", queue.kind, queue.address.Hex())
			}
			continue
		}
		data, err := callRequestQueue(cfg, getHash, statedb, header, queue.address, tracker)
		if err != nil {
			return nil, fmt.Errorf("request type %d system call: %w", queue.kind, err)
		}
		if len(data)%queue.size != 0 || len(data)/queue.size > queue.limit {
			return nil, fmt.Errorf("request type %d: invalid response size %d", queue.kind, len(data))
		}
		if len(data) != 0 {
			requests = append(requests, types.NewRequest(queue.kind, data))
		}
	}
	return requests, nil
}

func callRequestQueue(cfg *config.ChainConfig, getHash vm.GetHashFunc, statedb state.StateDB, header *types.Header, address types.Address, tracker *bal.AccessTracker) ([]byte, error) {
	evm, err := newSystemEVM(cfg, getHash, statedb, header)
	if err != nil {
		return nil, err
	}
	return executeSystemCall(evm, statedb, address, nil, tracker)
}
