# endgamepipe

Sub-slot finality pipeline for endgame finality (M+ era).

## Overview

Package `endgamepipe` implements the endgame sub-slot finality pipeline
described in the M+ roadmap item. It partitions each slot into sub-slot
intervals and attempts to finalize within a single slot when a 2/3+
supermajority of stake attests during the slot, providing finality in seconds
rather than epochs.

The core pipeline logic is provided by `EndgameFinalityTracker` and
`BFTFinalityPipeline` in the parent `consensus` package; this subpackage
re-exports and wires those components for the endgame regime.

> This package currently contains no standalone source files. All implementation
> lives in the parent `consensus` package (`endgame_finality.go`,
> `endgame_engine.go`, `endgame_pipeline.go`, `bft_finality_pipeline.go`) and
> is re-exported via `../endgamepipe` compat shims.

## Roadmap

Targets the M+ milestone: "endgame finality (BLS adapter with real pairing hooks)".

[← consensus](../README.md)
