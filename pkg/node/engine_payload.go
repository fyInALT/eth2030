package node

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/eth2030/eth2030/core/block"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto/bls"
	"github.com/eth2030/eth2030/engine"
	enginepayload "github.com/eth2030/eth2030/engine/payload"
)

func (b *engineBackend) ProcessBlockV4(
	payload *engine.ExecutionPayloadV3,
	expectedBlobVersionedHashes []types.Hash,
	parentBeaconBlockRoot types.Hash,
	executionRequests [][]byte,
) (engine.PayloadStatusV1, error) {
	var reqs types.Requests
	for _, reqBytes := range executionRequests {
		if len(reqBytes) == 0 {
			continue
		}
		reqs = append(reqs, &types.Request{Type: reqBytes[0], Data: reqBytes[1:]})
	}
	rHash := types.ComputeRequestsHash(reqs)
	return b.processBlockInternal(payload, parentBeaconBlockRoot, &rHash)
}

func (b *engineBackend) ProcessBlockV5(
	payload *engine.ExecutionPayloadV5,
	expectedBlobVersionedHashes []types.Hash,
	parentBeaconBlockRoot types.Hash,
	executionRequests [][]byte,
) (engine.PayloadStatusV1, error) {
	var reqs types.Requests
	for _, reqBytes := range executionRequests {
		if len(reqBytes) == 0 {
			continue
		}
		reqs = append(reqs, &types.Request{Type: reqBytes[0], Data: reqBytes[1:]})
	}
	rHash := types.ComputeRequestsHash(reqs)
	return b.processBlockInternal(&payload.ExecutionPayloadV3, parentBeaconBlockRoot, &rHash)
}

func (b *engineBackend) ForkchoiceUpdatedV4(
	state engine.ForkchoiceStateV1,
	payloadAttributes *engine.PayloadAttributesV4,
) (engine.ForkchoiceUpdatedResult, error) {
	var v3Attrs *engine.PayloadAttributesV3
	if payloadAttributes != nil {
		v3Attrs = &payloadAttributes.PayloadAttributesV3
	}
	return b.ForkchoiceUpdated(state, v3Attrs)
}

func (b *engineBackend) GetPayloadV4ByID(id engine.PayloadID) (*engine.GetPayloadV4Response, error) {
	resp, err := b.GetPayloadByID(id)
	if err != nil {
		return nil, err
	}
	return &engine.GetPayloadV4Response{
		ExecutionPayload:  &resp.ExecutionPayload.ExecutionPayloadV3,
		BlockValue:        resp.BlockValue,
		BlobsBundle:       resp.BlobsBundle,
		ExecutionRequests: [][]byte{},
	}, nil
}

func (b *engineBackend) GetPayloadV6ByID(id engine.PayloadID) (*engine.GetPayloadV6Response, error) {
	resp, err := b.GetPayloadByID(id)
	if err != nil {
		return nil, err
	}
	var blobsBundleV2 *engine.BlobsBundleV2
	if b1 := resp.BlobsBundle; b1 != nil {
		blobsBundleV2 = &engine.BlobsBundleV2{
			Commitments: b1.Commitments,
			Proofs:      b1.Proofs,
			Blobs:       b1.Blobs,
		}
	}
	return &engine.GetPayloadV6Response{
		ExecutionPayload: &engine.ExecutionPayloadV5{
			ExecutionPayloadV4: *resp.ExecutionPayload,
		},
		BlockValue:        resp.BlockValue,
		BlobsBundle:       blobsBundleV2,
		ExecutionRequests: [][]byte{},
	}, nil
}

func (b *engineBackend) GetHeadTimestamp() uint64 {
	head := b.node.blockchain.CurrentBlock()
	if head != nil {
		return head.Time()
	}
	return 0
}

func (b *engineBackend) GetBlockTimestamp(hash types.Hash) uint64 {
	blk := b.node.blockchain.GetBlock(hash)
	if blk != nil {
		return blk.Time()
	}
	return 0
}

func (b *engineBackend) GetPayloadBodiesByHash(hashes []types.Hash) ([]*engine.ExecutionPayloadBodyV2, error) {
	results := make([]*engine.ExecutionPayloadBodyV2, len(hashes))
	for i, h := range hashes {
		blk := b.node.blockchain.GetBlock(h)
		if blk == nil {
			results[i] = nil
			continue
		}
		results[i] = enginepayload.BlockToPayloadBodyV2(blk)
	}
	return results, nil
}

func (b *engineBackend) GetPayloadBodiesByRange(start, count uint64) ([]*engine.ExecutionPayloadBodyV2, error) {
	results := make([]*engine.ExecutionPayloadBodyV2, count)
	for i := uint64(0); i < count; i++ {
		num := start + i
		blk := b.node.blockchain.GetBlockByNumber(num)
		if blk == nil {
			results[i] = nil
			continue
		}
		results[i] = enginepayload.BlockToPayloadBodyV2(blk)
	}
	return results, nil
}

func (b *engineBackend) IsCancun(timestamp uint64) bool {
	return b.node.blockchain.Config().IsCancun(timestamp)
}

func (b *engineBackend) IsPrague(timestamp uint64) bool {
	return b.node.blockchain.Config().IsPrague(timestamp)
}

func (b *engineBackend) IsAmsterdam(timestamp uint64) bool {
	return b.node.blockchain.Config().IsAmsterdam(timestamp)
}

func (b *engineBackend) GetPayloadByID(id engine.PayloadID) (*engine.GetPayloadResponse, error) {
	slog.Debug("engine_getPayload", "payloadID", id)

	b.mu.Lock()
	payload, ok := b.payloads[id]
	b.mu.Unlock()

	if !ok {
		slog.Warn("engine_getPayload: payload not found", "payloadID", id)
		return nil, fmt.Errorf("payload %v not found", id)
	}

	select {
	case <-payload.done:
	case <-time.After(8 * time.Second):
		slog.Warn("engine_getPayload: build timed out", "payloadID", id)
		return nil, fmt.Errorf("payload %v build timed out", id)
	}

	if payload.err != nil {
		return nil, payload.err
	}

	block := payload.block
	header := block.Header()

	var blobGasUsed, excessBlobGas uint64
	if header.BlobGasUsed != nil {
		blobGasUsed = *header.BlobGasUsed
	}
	if header.ExcessBlobGas != nil {
		excessBlobGas = *header.ExcessBlobGas
	}
	execPayload := &engine.ExecutionPayloadV4{
		ExecutionPayloadV3: engine.ExecutionPayloadV3{
			ExecutionPayloadV2: engine.ExecutionPayloadV2{
				ExecutionPayloadV1: engine.ExecutionPayloadV1{
					ParentHash:    header.ParentHash,
					FeeRecipient:  header.Coinbase,
					StateRoot:     header.Root,
					ReceiptsRoot:  header.ReceiptHash,
					LogsBloom:     header.Bloom,
					PrevRandao:    header.MixDigest,
					BlockNumber:   block.NumberU64(),
					GasLimit:      header.GasLimit,
					GasUsed:       header.GasUsed,
					Timestamp:     header.Time,
					ExtraData:     header.Extra,
					BaseFeePerGas: header.BaseFee,
					BlockHash:     block.Hash(),
					Transactions:  encodeTxsRLP(block.Transactions()),
				},
			},
			BlobGasUsed:   blobGasUsed,
			ExcessBlobGas: excessBlobGas,
		},
	}

	if ws := block.Withdrawals(); ws != nil {
		for _, w := range ws {
			execPayload.Withdrawals = append(execPayload.Withdrawals, &engine.Withdrawal{
				Index:          w.Index,
				ValidatorIndex: w.ValidatorIndex,
				Address:        w.Address,
				Amount:         w.Amount,
			})
		}
	}

	blockValue := new(big.Int)
	for _, receipt := range payload.receipts {
		if receipt.EffectiveGasPrice != nil && header.BaseFee != nil {
			tip := new(big.Int).Sub(receipt.EffectiveGasPrice, header.BaseFee)
			if tip.Sign() > 0 {
				tipTotal := new(big.Int).Mul(tip, new(big.Int).SetUint64(receipt.GasUsed))
				blockValue.Add(blockValue, tipTotal)
			}
		}
	}

	slog.Debug("engine_getPayload: returning payload",
		"payloadID", id,
		"blockNumber", block.NumberU64(),
		"blockHash", block.Hash(),
		"txCount", len(block.Transactions()),
		"blockValue", blockValue,
	)

	blobsBundle := &engine.BlobsBundleV1{}
	blobTxCount := 0
	for _, tx := range block.Transactions() {
		if tx.Type() != types.BlobTxType {
			continue
		}
		blobTxCount++
		sc := tx.BlobSidecar()
		if sc == nil {
			slog.Debug("engine_getPayload: blob tx missing sidecar",
				"txHash", tx.Hash(),
				"blockNumber", block.NumberU64(),
			)
			continue
		}
		blobsBundle.Commitments = append(blobsBundle.Commitments, sc.Commitments...)
		blobsBundle.Proofs = append(blobsBundle.Proofs, sc.Proofs...)
		blobsBundle.Blobs = append(blobsBundle.Blobs, sc.Blobs...)
	}
	if blobTxCount > 0 {
		slog.Debug("engine_getPayload: blobsBundle",
			"blobTxCount", blobTxCount,
			"blobCount", len(blobsBundle.Blobs),
			"blockNumber", block.NumberU64(),
		)
	}

	return &engine.GetPayloadResponse{
		ExecutionPayload: execPayload,
		BlockValue:       blockValue,
		BlobsBundle:      blobsBundle,
		Override:         false,
	}, nil
}

func (b *engineBackend) GetBlobsByVersionedHashes(hashes []types.Hash) []*engine.BlobAndProofV1 {
	if b.node.txPool == nil {
		return make([]*engine.BlobAndProofV1, len(hashes))
	}
	raw := b.node.txPool.GetBlobsByVersionedHashes(hashes)
	result := make([]*engine.BlobAndProofV1, len(raw))
	for i, r := range raw {
		if r != nil {
			result[i] = &engine.BlobAndProofV1{
				Blob:       r.Blob,
				Commitment: r.Commitment,
				Proof:      r.Proof,
			}
		}
	}
	return result
}

func (b *engineBackend) GetBlobsV2ByVersionedHashes(hashes []types.Hash) []*engine.BlobAndProofV2 {
	return b.computeBlobsV2(hashes)
}

func (b *engineBackend) GetBlobsV3ByVersionedHashes(hashes []types.Hash) []*engine.BlobAndProofV2 {
	return b.computeBlobsV2(hashes)
}

func (b *engineBackend) computeBlobsV2(hashes []types.Hash) []*engine.BlobAndProofV2 {
	result := make([]*engine.BlobAndProofV2, len(hashes))
	kzg := bls.DefaultKZGBackend()

	b.blobCacheMu.RLock()
	cachedBlobs := make([][]byte, len(hashes))
	cacheHits := 0
	for i, h := range hashes {
		if blob, ok := b.blobCache[h]; ok {
			cachedBlobs[i] = blob
			cacheHits++
		}
	}
	cacheSize := len(b.blobCache)
	b.blobCacheMu.RUnlock()

	if len(hashes) > 0 {
		slog.Debug("computeBlobsV2: looking up blobs",
			"requested", len(hashes),
			"cacheHits", cacheHits,
			"cacheSize", cacheSize,
		)
	}

	for i, blob := range cachedBlobs {
		if blob == nil || len(blob) == 0 {
			continue
		}
		_, cellProofs, err := kzg.ComputeCellsAndProofs(blob)
		if err != nil {
			slog.Warn("computeBlobsV2: ComputeCellsAndProofs failed for cached blob",
				"hash", hashes[i], "err", err)
			continue
		}
		proofs := make([][]byte, len(cellProofs))
		for j, p := range cellProofs {
			cp := p
			proofs[j] = cp[:]
		}
		result[i] = &engine.BlobAndProofV2{
			Blob:   blob,
			Proofs: proofs,
		}
	}

	if b.node.txPool != nil {
		var pendingHashes []types.Hash
		pendingIdx := make(map[int]int)
		for i, h := range hashes {
			if result[i] == nil {
				pendingIdx[len(pendingHashes)] = i
				pendingHashes = append(pendingHashes, h)
			}
		}
		if len(pendingHashes) > 0 {
			raw := b.node.txPool.GetBlobsByVersionedHashes(pendingHashes)
			poolHits := 0
			for j, r := range raw {
				if r == nil || len(r.Blob) == 0 {
					continue
				}
				poolHits++
				origIdx := pendingIdx[j]
				_, cellProofs, err := kzg.ComputeCellsAndProofs(r.Blob)
				if err != nil {
					slog.Warn("computeBlobsV2: ComputeCellsAndProofs failed",
						"hash", pendingHashes[j], "err", err)
					continue
				}
				proofs := make([][]byte, len(cellProofs))
				for k, p := range cellProofs {
					cp := p
					proofs[k] = cp[:]
				}
				result[origIdx] = &engine.BlobAndProofV2{
					Blob:   r.Blob,
					Proofs: proofs,
				}
			}
			slog.Debug("computeBlobsV2: txpool lookup",
				"pending", len(pendingHashes),
				"poolHits", poolHits,
			)
		}
	}

	found := 0
	for _, r := range result {
		if r != nil {
			found++
		}
	}
	if found < len(hashes) && len(hashes) > 0 {
		slog.Warn("computeBlobsV2: missing blobs",
			"requested", len(hashes),
			"found", found,
		)
	}
	return result
}

func (b *engineBackend) cacheBlobsFromBlock(blk *types.Block) {
	if blk == nil {
		return
	}
	txs := blk.Transactions()
	if len(txs) == 0 {
		return
	}

	b.blobCacheMu.Lock()
	defer b.blobCacheMu.Unlock()

	cachedCount := 0
	for _, tx := range txs {
		sidecar := tx.BlobSidecar()
		if sidecar == nil {
			continue
		}
		blobHashes := tx.BlobHashes()
		if len(blobHashes) != len(sidecar.Blobs) {
			slog.Warn("cacheBlobsFromBlock: blobHashes/blobs length mismatch",
				"blockNum", blk.NumberU64(),
				"hashCount", len(blobHashes),
				"blobCount", len(sidecar.Blobs),
			)
			continue
		}
		for i, hash := range blobHashes {
			if i < len(sidecar.Blobs) && len(sidecar.Blobs[i]) > 0 {
				b.blobCache[hash] = sidecar.Blobs[i]
				cachedCount++
			}
		}
	}
	if cachedCount > 0 {
		slog.Info("cacheBlobsFromBlock: cached blobs from block",
			"blockNum", blk.NumberU64(),
			"cachedCount", cachedCount,
			"cacheSize", len(b.blobCache),
		)
	}
}

func generatePayloadID(parentHash types.Hash, attrs *block.BuildBlockAttributes) engine.PayloadID {
	var buf [8]byte
	copy(buf[:], parentHash[:8])
	var tsBytes [8]byte
	binary.BigEndian.PutUint64(tsBytes[:], attrs.Timestamp)
	for i := range buf {
		buf[i] ^= tsBytes[i]
	}
	for i := 0; i < 8; i++ {
		buf[i] ^= attrs.FeeRecipient[i]
	}
	for i := 0; i < 8; i++ {
		buf[i] ^= attrs.Random[8+i]
	}

	var id engine.PayloadID
	copy(id[:], buf[:])

	if id == (engine.PayloadID{}) {
		rand.Read(id[:])
	}
	return id
}

func encodeTxsRLP(txs []*types.Transaction) [][]byte {
	encoded := make([][]byte, 0, len(txs))
	for _, tx := range txs {
		raw, err := tx.EncodeRLP()
		if err != nil {
			continue
		}
		encoded = append(encoded, raw)
	}
	return encoded
}