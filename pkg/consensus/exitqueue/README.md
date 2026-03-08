# exitqueue

Validator exit queue with adaptive epoch-churn-limit enforcement.

## Overview

Package `exitqueue` implements the `MinslackExitQueue` adaptive exit delay
model from the L+ roadmap. Rather than applying a fixed `MIN_VALIDATOR_WITHDRAWABILITY_DELAY`
to all validators regardless of queue pressure, the model scales the delay
proportionally to current queue fill level. Saturated queues fast-track exits
to `MaxSeedLookahead` (4) epochs, while empty queues keep the full 256-epoch
delay. This prevents indefinite blocking during mass-exit events.

All methods are stateless; callers supply the current epoch, queue size, and
per-epoch churn cap. The package exposes only the necessary constants to avoid
a circular import with the parent `consensus` package.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `MinslackExitQueue` | Stateless helper computing adaptive exit delays |

### Methods

| Method | Description |
|--------|-------------|
| `(*MinslackExitQueue).ComputeExitDelay(epoch, queueSize, maxChurn uint64) uint64` | Returns the exit delay in epochs based on queue fill level |

### Constants

| Name | Value | Description |
|------|-------|-------------|
| `MaxSeedLookahead` | 4 | Minimum delay used for saturated queues |
| `MinValidatorWithdrawabilityDelay` | 256 | Standard full withdrawal delay |

### Delay schedule

| Condition | Delay |
|-----------|-------|
| `queueSize == 0` | 256 epochs |
| `queueSize >= maxChurn` | 4 epochs |
| `queueSize < maxChurn/2` | 128 epochs |
| otherwise | `128 * (maxChurn - queueSize) / maxChurn` |

## Usage

```go
import "github.com/eth2030/eth2030/consensus/exitqueue"

eq := &exitqueue.MinslackExitQueue{}
delay := eq.ComputeExitDelay(currentEpoch, len(exitQueue), maxChurn)
exitEpoch := currentEpoch + delay
```

[← consensus](../README.md)
