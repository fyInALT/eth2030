# rollup/proof — Fraud proofs, trace disputes, and cross-layer message proofs

## Overview

This package provides the full dispute and proof infrastructure for native
rollups (EIP-8079). It covers three domains: fraud proof generation and
verification for invalid state roots or transactions; interactive bisection for
narrowing a multi-step dispute to a single execution step; and Merkle-based
cross-layer message proofs for trustless deposit and withdrawal verification
between L1 and L2.

Fraud proofs encode a commitment over the pre/post state roots and state witness
data that a verifier can check without re-executing the full block. The bisection
protocol repeatedly halves the disputed step range until it converges to a single
transaction index, at which point a final fraud proof is produced.

## Functionality

**Fraud proofs (fraud.go)**

- `FraudProofType` — `InvalidStateRoot`, `InvalidReceipt`, `InvalidTransaction`
- `FraudProof` — `Type`, `BlockNumber`, `StepIndex`, `PreStateRoot`,
  `PostStateRoot`, `ExpectedRoot`, `Proof`
- `FraudProofGenerator` — `NewFraudProofGenerator(stateReader, txExecutor)`
  - `GenerateStateRootProof(blockNumber, expected, actual) (*FraudProof, error)`
  - `GenerateSingleStepProof(blockNumber, txIndex, pre, post, txData) (*FraudProof, error)`
- `FraudProofVerifier` — `NewFraudProofVerifier(stateVerifier)`
  - `VerifyFraudProof(p) (bool, error)`
- `InteractiveVerification` — `NewInteractiveVerification(block, start, end)`
  - `BisectionStep(claimerRoot, challengerRoot) (start, end, error)`
  - `GenerateBisectionProof() (*FraudProof, error)`
- `ComputeStateTransition(preState, tx) ([32]byte, error)`
- `ComputeProofHash(p) types.Hash`

**Cross-layer message proofs (crosslayer.go)**

- `CrossLayerMessage` — `Source`, `Destination`, `Nonce`, `Sender`, `Target`,
  `Value`, `Data`
- `MessageProof` — message + Merkle sibling path + `MessageHash`
- `MessageProofGenerator.GenerateDepositProof(msg, stateRoot) (*MessageProof, error)`
- `MessageProofGenerator.GenerateWithdrawalProof(msg, outputRoot) (*MessageProof, error)`
- `VerifyCrossLayerDepositProof(p, l1StateRoot) (bool, error)`
- `VerifyCrossLayerWithdrawalProof(p, l2OutputRoot) (bool, error)`
- `ComputeMessageHash(msg) [32]byte`
- `VerifyCrossLayerMerkleProof(leaf, root, path, index) bool`

**Parent package:** [rollup](../)
