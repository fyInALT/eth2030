# rewards

Attestation and sync committee reward tables for the consensus layer.

## Overview

Package `rewards` provides reward and penalty calculation tables for validator
duties. The implementation lives in the parent `consensus` package
(`reward_calculator.go`, `reward_calculator_v2.go`, `block_rewards.go`) and is
re-exported via compat shims in this subpackage.

Key types from the parent package:

- `RewardCalculator` — computes per-validator rewards for attestation
  participation, sync committee contributions, and proposer inclusion rewards
  using the base reward formula from the beacon chain spec.
- `RewardCalculatorV2` — updated version that adds support for EIP-7251
  increased effective balances (up to 2048 ETH) and the post-Altair
  participation flag scheme.

> This package currently contains no standalone source files. All logic resides
> in `consensus/reward_calculator.go` and `consensus/reward_calculator_v2.go`.

[← consensus](../README.md)
