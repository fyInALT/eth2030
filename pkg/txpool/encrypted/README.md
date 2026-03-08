# txpool/encrypted — Encrypted mempool with commit-reveal and threshold decryption

## Overview

Package `encrypted` implements the EL Cryptography roadmap item for an encrypted mempool. Transactions are first committed by hash only (hiding contents), then revealed after a commit window, and ordered by commit timestamp to prevent MEV frontrunning. Threshold decryption requires t-of-n validator shares before ciphertext can be opened.

The package contains three complementary components: `EncryptedMempoolProtocol` for the high-level commit/reveal lifecycle, `EncryptedPool` for the pool-side state machine, and `ThresholdDecryptor` for share collection and AES-GCM decryption with Lagrange interpolation of Shamir shares.

## Functionality

**Types**
- `EncryptedPool` — commit-reveal pool; `AddCommit`, `AddReveal`, `GetRevealed`, `ExpireCommits`, `RevealAndOrder`
- `EncryptedMempoolProtocol` — high-level protocol; `Commit`, `Reveal`, `GetRevealed`, `ExpireOldCommits`, `SetEpoch`
- `ThresholdDecryptor` — t-of-n share collector; `AddShare`, `TryDecrypt`, `ThresholdMet`, `ResetEpoch`
- `CommitTx`, `RevealTx`, `CommitEntry` — commit/reveal data structs
- `CommittedTx`, `RevealedTx` — protocol-level lifecycle records
- `DecryptionShare` — validator decryption contribution

**Functions**
- `OrderByCommitTime(entries)` — sorts committed entries by timestamp (MEV-fair ordering)
- `ComputeCommitHash(tx)` — keccak256(rlp(tx)) commitment hash
- `ComputeDecryptionKey(shares)` — Lagrange interpolation over shares into AES key
- `VerifyShare(share, commitment)` / `MakeCommitment(share)` — share commitment scheme

## Usage

```go
pool := encrypted.NewEncryptedPool()
pool.AddCommit(&encrypted.CommitTx{CommitHash: h, Sender: addr, GasLimit: 21000, Timestamp: t})
pool.AddReveal(&encrypted.RevealTx{CommitHash: h, Transaction: tx})
txs, _ := pool.RevealAndOrder(decryptor)
```

[← txpool](../README.md)
