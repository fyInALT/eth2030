package node

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto/bls"
	"github.com/eth2030/eth2030/p2p"
	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	gethkzg "github.com/ethereum/go-ethereum/crypto/kzg4844"
	gethrlp "github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
)

func TestOrdinaryBlobCellProofsThroughActualHTTP(t *testing.T) {
	const token = "ordinary-blob-proof-rpc"
	key, err := gethcrypto.HexToECDSA(fmt.Sprintf("%064x", 1))
	if err != nil {
		t.Fatal(err)
	}
	sender := types.Address(gethcrypto.PubkeyToAddress(key.PublicKey))
	var genesis map[string]any
	if err = json.Unmarshal(minimalGenesisJSON(1337), &genesis); err != nil {
		t.Fatal(err)
	}
	fields := genesis["config"].(map[string]any)
	fields["terminalTotalDifficulty"], fields["osakaTime"], fields["amsterdamTime"] = 0, 0, 0
	genesis["alloc"].(map[string]any)[sender.Hex()] = map[string]string{"balance": "1000000000000000000"}
	rawGenesis, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}
	cfg := makeTestConfig(t)
	cfg.GenesisPath, cfg.EngineAuthToken = filepath.Join(cfg.DataDir, "genesis.json"), token
	if err = os.WriteFile(cfg.GenesisPath, rawGenesis, 0600); err != nil {
		t.Fatal(err)
	}
	node := newTestNode(t, &cfg)
	rpc := track4NodeRPC(t, node, token)
	blob := make([]byte, 131072)
	blob[31] = 1
	backend := bls.DefaultKZGBackend()
	commitment, err := backend.BlobToCommitment(blob)
	if err != nil {
		t.Fatal(err)
	}
	_, proofs, err := backend.ComputeCellsAndProofs(blob)
	if err != nil {
		t.Fatal(err)
	}
	versioned := sha256.Sum256(commitment[:])
	versioned[0] = 1
	sdkBlob := gethkzg.Blob(blob)
	sidecar := &gethtypes.BlobTxSidecar{Version: gethtypes.BlobSidecarVersion1, Blobs: []gethkzg.Blob{sdkBlob}, Commitments: []gethkzg.Commitment{gethkzg.Commitment(commitment)}}
	for _, proof := range proofs {
		sidecar.Proofs = append(sidecar.Proofs, gethkzg.Proof(proof))
	}
	transaction := gethtypes.NewTx(&gethtypes.BlobTx{ChainID: uint256.NewInt(1337), To: gethcommon.HexToAddress("0xbeef"), Gas: 100000, GasFeeCap: uint256.NewInt(2000000000), GasTipCap: uint256.NewInt(1000000000), BlobFeeCap: uint256.NewInt(1000000000), Value: new(uint256.Int), BlobHashes: []gethcommon.Hash{gethcommon.Hash(versioned)}})
	transaction, err = gethtypes.SignTx(transaction, gethtypes.LatestSignerForChainID(transaction.ChainId()), key)
	if err != nil {
		t.Fatal(err)
	}
	badSidecar := *sidecar
	badSidecar.Proofs = append([]gethkzg.Proof(nil), sidecar.Proofs...)
	badSidecar.Proofs[127] = badSidecar.Proofs[0]
	bad, err := transaction.WithBlobTxSidecar(&badSidecar).MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	var submitted types.Hash
	if err = rpc.invoke("eth_sendRawTransaction", &submitted, "0x"+hex.EncodeToString(bad)); err == nil {
		t.Fatal("ordinary blob HTTP admission accepted an incorrect final-cell proof")
	}
	if node.TxPool().Get(types.Hash(transaction.Hash())) != nil {
		t.Fatal("rejected blob sidecar was adopted by the pool")
	}
	raw, err := transaction.WithBlobTxSidecar(sidecar).MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	var fieldsV1 []gethrlp.RawValue
	if err = gethrlp.DecodeBytes(raw[1:], &fieldsV1); err != nil {
		t.Fatal(err)
	}
	fieldsV1[1] = gethrlp.RawValue{2}
	unknown, err := gethrlp.EncodeToBytes(fieldsV1)
	if err != nil {
		t.Fatal(err)
	}
	if err = rpc.invoke("eth_sendRawTransaction", &submitted, "0x"+hex.EncodeToString(append([]byte{3}, unknown...))); err == nil {
		t.Fatal("HTTP accepted unsupported blob wrapper version")
	}
	rpc.call(t, "eth_sendRawTransaction", &submitted, "0x"+hex.EncodeToString(raw))
	if submitted != types.Hash(transaction.Hash()) {
		t.Fatal("sidecar changed signed transaction identity")
	}
	stored := node.TxPool().Get(submitted)
	if stored == nil || stored.BlobSidecar() == nil || len(stored.BlobSidecar().Proofs) != 128 {
		t.Fatal("admission lost the full cell proof array")
	}
	pooled, err := types.EncodePooledTransaction(stored)
	if err != nil {
		t.Fatal(err)
	}
	var independent gethtypes.Transaction
	if err = independent.UnmarshalBinary(pooled); err != nil {
		t.Fatal(err)
	}
	if independent.Hash() != transaction.Hash() || independent.BlobTxSidecar() == nil || independent.BlobTxSidecar().Version != 1 || len(independent.BlobTxSidecar().Proofs) != 128 || !bytes.Equal(pooled, raw) {
		t.Fatal("native pooled encoding lost the signed body or full PeerDAS wrapper")
	}
	packet := p2p.PooledTransactionsPacket{RequestID: 17, Transactions: []*types.Transaction{stored}}
	encodedPacket, err := packet.EncodeRLP()
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		ID           uint64
		Transactions [][]byte
	}
	if err = gethrlp.DecodeBytes(encodedPacket, &wire); err != nil || wire.ID != 17 || len(wire.Transactions) != 1 || !bytes.Equal(wire.Transactions[0], raw) {
		t.Fatalf("production pooled packet differs from independent RLP: %v", err)
	}
	var legacy []map[string]any
	rpc.call(t, "engine_getBlobsV1", &legacy, []types.Hash{types.Hash(versioned)})
	if len(legacy) != 1 || legacy[0] == nil {
		t.Fatal("admitted blob unavailable through Engine")
	}
	if len(legacy[0]) != 2 || legacy[0]["blob"] != "0x"+hex.EncodeToString(blob) {
		t.Fatal("legacy Engine DATA fields are not canonical")
	}
	encoded, ok := legacy[0]["proof"].(string)
	if !ok {
		t.Fatal("legacy proof absent")
	}
	proof, err := hex.DecodeString(encoded[2:])
	if err != nil {
		t.Fatal(err)
	}
	valid, err := backend.VerifyBlobProof(blob, commitment[:], proof)
	if err != nil || !valid {
		t.Fatalf("legacy Engine API served a cell proof as a blob proof: %v", err)
	}
	if bytes.Equal(proof, sidecar.Proofs[0][:]) {
		t.Fatal("legacy proof aliases a cell proof")
	}
	legacySidecar := &gethtypes.BlobTxSidecar{Version: gethtypes.BlobSidecarVersion0, Blobs: sidecar.Blobs, Commitments: sidecar.Commitments, Proofs: []gethkzg.Proof{gethkzg.Proof(proof)}}
	legacyWire, err := transaction.WithBlobTxSidecar(legacySidecar).MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	legacyNative, err := types.DecodePooledTransaction(legacyWire)
	if err != nil {
		t.Fatal(err)
	}
	legacyEncoded, err := types.EncodePooledTransaction(legacyNative)
	if err != nil || legacyNative.BlobSidecar().Version != 0 || !bytes.Equal(legacyWire, legacyEncoded) {
		t.Fatalf("legacy SDK/native pooled encoding differs: %v", err)
	}

	for _, tc := range []struct {
		name   string
		hashes []gethcommon.Hash
	}{
		{"bad_hash_version", []gethcommon.Hash{{0: 2}}},
		{"more_than_six_blobs", []gethcommon.Hash{gethcommon.Hash(versioned), gethcommon.Hash(versioned), gethcommon.Hash(versioned), gethcommon.Hash(versioned), gethcommon.Hash(versioned), gethcommon.Hash(versioned), gethcommon.Hash(versioned)}},
	} {
		malformed := gethtypes.NewTx(&gethtypes.BlobTx{ChainID: uint256.NewInt(1337), Nonce: 1, To: gethcommon.HexToAddress("0xbeef"), Gas: 100000, GasFeeCap: uint256.NewInt(2000000000), GasTipCap: uint256.NewInt(1000000000), BlobFeeCap: uint256.NewInt(1000000000), Value: new(uint256.Int), BlobHashes: tc.hashes})
		malformed, err = gethtypes.SignTx(malformed, gethtypes.LatestSignerForChainID(malformed.ChainId()), key)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := malformed.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		if err = rpc.invoke("eth_sendRawTransaction", &submitted, "0x"+hex.EncodeToString(encoded)); err == nil {
			t.Fatalf("HTTP admitted %s", tc.name)
		}
	}

	if err = node.txJournal.Rotate(map[types.Address][]*types.Transaction{sender: {stored}}); err != nil {
		t.Fatal(err)
	}
	if err = node.Stop(); err != nil {
		t.Fatal(err)
	}
	restarted := newTestNode(t, &cfg)
	recovered := restarted.TxPool().Get(types.Hash(transaction.Hash()))
	if recovered == nil || recovered == stored || recovered.BlobSidecar() == nil || recovered.BlobSidecar().Version != 1 || len(recovered.BlobSidecar().Proofs) != 128 {
		t.Fatal("fresh node journal replay lost the authenticated blob sidecar")
	}
	freshRPC := track4NodeRPC(t, restarted, token)
	var recoveredBlobs []map[string]any
	freshRPC.call(t, "engine_getBlobsV1", &recoveredBlobs, []types.Hash{types.Hash(versioned)})
	if len(recoveredBlobs) != 1 || recoveredBlobs[0] == nil || recoveredBlobs[0]["blob"] != "0x"+hex.EncodeToString(blob) {
		t.Fatal("fresh node cannot serve replayed blob")
	}

}
