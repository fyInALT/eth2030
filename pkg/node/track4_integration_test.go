package node

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/crypto/pqc"
	"github.com/eth2030/eth2030/engine/payload"
	gethcore "github.com/ethereum/go-ethereum/core"
)

var track4Contract = types.Address{18: 0x71}

type track4RPC struct {
	url, token string
}

type track4RPCError struct {
	Code    int
	Message string
}

func (e *track4RPCError) Error() string { return fmt.Sprintf("RPC %d: %s", e.Code, e.Message) }

func (r track4RPC) invoke(method string, result any, params ...any) error {
	if params == nil {
		params = []any{}
	}
	body, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, r.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if r.token != "" {
		req.Header.Set("Authorization", "Bearer "+r.token)
	}
	res, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %s", res.Status)
	}
	var response struct {
		Result json.RawMessage
		Error  *track4RPCError
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return err
	}
	if response.Error != nil {
		return response.Error
	}
	return json.Unmarshal(response.Result, result)
}

func (r track4RPC) call(t *testing.T, method string, result any, params ...any) {
	t.Helper()
	if err := r.invoke(method, result, params...); err != nil {
		t.Fatalf("%s: %v", method, err)
	}
}

type track4Vector struct {
	Name, Message, Signature, Key string
}

func track4Vectors(t *testing.T) []track4Vector {
	t.Helper()
	f, err := os.Open("../core/vm/testdata/eip8051_expanded.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var vectors []track4Vector
	if err := json.NewDecoder(z).Decode(&vectors); err != nil || len(vectors) != 2 {
		t.Fatalf("read expanded-key vectors: %d, %v", len(vectors), err)
	}
	return vectors
}

func track4Hex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func track4Precompiles(t *testing.T, r track4RPC, at string, active bool) {
	t.Helper()
	zero := (types.Hash{}).Hex()
	one := (types.Hash{31: 1}).Hex()
	for _, v := range track4Vectors(t) {
		variant, address := vm.EIP8051DraftNIST, types.Address{18: 0x70, 19: 2}
		if v.Name == "Keccak" {
			variant, address = vm.EIP8051DraftETH, types.Address{18: 0x70, 19: 3}
		}
		msg, sig, key := track4Hex(t, v.Message), track4Hex(t, v.Signature), track4Hex(t, v.Key)
		id := crypto.Keccak256(key)[12:]
		wantRecover := types.BytesToHash(crypto.Keccak256([]byte{variant}, key)[12:]).Hex()
		direct := bytes.Join([][]byte{msg, sig, key}, nil)
		recover := bytes.Join([][]byte{{variant}, sig, id, msg}, nil)
		assertCall := func(to types.Address, input []byte, want string) {
			t.Helper()
			var output string
			r.call(t, "eth_call", &output, map[string]string{"to": to.Hex(), "data": "0x" + hex.EncodeToString(input), "gas": "0xb0000"}, at)
			if output != want {
				t.Fatalf("%s at %s returned %s, want %s", v.Name, at, output, want)
			}
		}
		recoverAddr := types.Address{18: 0x70, 19: 1}
		if !active {
			assertCall(address, direct, "0x")
			assertCall(recoverAddr, recover, "0x")
			continue
		}
		assertCall(address, direct, one)
		assertCall(recoverAddr, recover, wantRecover)
		direct[0] ^= 1
		recover[len(recover)-1] ^= 1
		assertCall(address, direct, zero)
		assertCall(recoverAddr, recover, zero)
	}
}

// Public, deterministic key for isolated development networks.
func track4Key() (*mldsa65.PublicKey, *mldsa65.PrivateKey) {
	return mldsa65.NewKeyFromSeed(&[32]byte{4})
}

func track4SignedTx(t *testing.T, chainID *big.Int, nonce uint64, value byte) ([]byte, types.Hash) {
	t.Helper()
	pub, priv := track4Key()
	data := make([]byte, 32)
	data[31] = value
	tx := types.NewPQTransaction(chainID, nonce, &track4Contract, new(big.Int), 100000, big.NewInt(3_000_000_000), data)
	tx.PQSignatureType = types.PQSigDilithium
	hash := tx.SigningHash()
	sig := make([]byte, mldsa65.SignatureSize)
	if err := mldsa65.SignTo(priv, hash[:], nil, false, sig); err != nil {
		t.Fatal(err)
	}
	tx.SignWithPQ(types.PQSigDilithium, pub.Bytes(), sig)
	raw, err := tx.EncodePQ()
	if err != nil {
		t.Fatal(err)
	}
	return raw, tx.Hash()
}

type track4Receipt struct {
	Status, Type, From, To string
	EffectiveGasPrice      string
	BlockHash              types.Hash
	BlockNumber            string
}

func track4Storage(t *testing.T, r track4RPC, at string, expected byte) {
	t.Helper()
	var stored types.Hash
	r.call(t, "eth_getStorageAt", &stored, track4Contract.Hex(), "0x0", at)
	if stored != (types.Hash{31: expected}) {
		t.Fatalf("slot at %s = %s, want %d", at, stored, expected)
	}
}

func track4NodeRPC(t *testing.T, n *Node, token string) track4RPC {
	t.Helper()
	server := httptest.NewServer(n.engineServer.Handler())
	t.Cleanup(server.Close)
	return track4RPC{url: server.URL, token: token}
}

func track4Genesis(t *testing.T, funded ...types.Address) []byte {
	t.Helper()
	var genesis map[string]any
	if err := json.Unmarshal(minimalGenesisJSON(1337), &genesis); err != nil {
		t.Fatal(err)
	}
	cfg, alloc := genesis["config"].(map[string]any), genesis["alloc"].(map[string]any)
	cfg["terminalTotalDifficulty"] = 0
	cfg["iPlusTime"], cfg["binaryOverlayTime"], cfg["binaryConversionTime"], cfg["binaryConversionStride"] = 2, 2, 3, 1
	cfg["pqDraft"] = map[string]any{
		"time": 2, "sigRecover": (types.Address{18: 0x70, 19: 1}).Hex(),
		"mldsa": (types.Address{18: 0x70, 19: 2}).Hex(), "mldsaEth": (types.Address{18: 0x70, 19: 3}).Hex(),
	}
	pub, _ := track4Key()
	for _, address := range append(funded, types.PQPubKeyToAddress(pub.Bytes())) {
		alloc[address.Hex()] = map[string]string{"balance": "100000000000000000000", "code": "0x"}
	}
	alloc[track4Contract.Hex()] = map[string]any{
		"balance": "0", "nonce": "0x1", "code": "0x60003560005500", "storage": map[string]string{"0x00": "0x2c"},
	}
	for _, v := range track4Vectors(t) {
		addr := types.BytesToAddress(crypto.Keccak256(track4Hex(t, v.Key))[12:])
		alloc[addr.Hex()] = map[string]string{"balance": "0", "nonce": "0x1", "code": "0x00" + v.Key}
	}
	data, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestTrack4IndependentEngineAndRPC(t *testing.T) {
	const token = "track4-integration"
	type signed struct {
		raw    []byte
		hash   types.Hash
		sender types.Address
	}
	var other []signed
	var funded []types.Address
	for _, scheme := range []pqc.PQSignatureType{pqc.SigFalcon512, pqc.SigSPHINCS128} {
		signer, err := pqc.NewPQTxSigner(scheme)
		if err != nil {
			t.Fatal(err)
		}
		priv, pub, err := signer.GenerateKey()
		if err != nil {
			t.Fatal(err)
		}
		to := types.Address{18: 0x71, 19: 1}
		tx := types.NewPQTransaction(big.NewInt(1337), 0, &to, big.NewInt(1), 100000, big.NewInt(3_000_000_000), nil)
		signature, err := signer.SignTransaction(types.NewTransaction(tx), priv)
		if err != nil {
			t.Fatal(err)
		}
		tx.SignWithPQ(uint8(scheme), pub.Raw, signature)
		raw, err := tx.EncodePQ()
		if err != nil {
			t.Fatal(err)
		}
		address := types.PQPubKeyToAddress(pub.Raw)
		other = append(other, signed{raw, tx.Hash(), address})
		funded = append(funded, address)
	}
	genesis := track4Genesis(t, funded...)
	open := func() *Config {
		cfg := makeTestConfig(t)
		cfg.GenesisPath, cfg.EngineAuthToken = filepath.Join(cfg.DataDir, "genesis.json"), token
		if err := os.WriteFile(cfg.GenesisPath, genesis, 0600); err != nil {
			t.Fatal(err)
		}
		return &cfg
	}
	producerCfg, verifierCfg := open(), open()
	producerNode, verifierNode := newTestNode(t, producerCfg), newTestNode(t, verifierCfg)
	producer, verifier := track4NodeRPC(t, producerNode, token), track4NodeRPC(t, verifierNode, token)

	var capability []string
	if err := (track4RPC{url: producer.url}).invoke("engine_exchangeCapabilities", &capability, []string{}); err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("uncredentialed Engine request: %v", err)
	}
	head := producerNode.Blockchain().Genesis().Hash()
	var reference gethcore.Genesis
	if err := json.Unmarshal(genesis, &reference); err != nil {
		t.Fatal(err)
	}
	if expected := types.Hash(reference.ToBlock().Hash()); head != expected || verifierNode.Blockchain().Genesis().Hash() != expected {
		t.Fatalf("genesis hash %s, reference Geth %s", head, expected)
	}
	build := func(timestamp uint64, wantTx ...[]byte) *payload.ExecutionPayloadV3 {
		t.Helper()
		beaconRoot := types.Hash{31: byte(timestamp + 10)}
		attrs := map[string]any{"timestamp": fmt.Sprintf("0x%x", timestamp), "prevRandao": types.Hash{}, "parentBeaconBlockRoot": beaconRoot,
			"suggestedFeeRecipient": types.Address{19: 0x42}, "withdrawals": []any{}}
		var fc payload.ForkchoiceUpdatedResult
		producer.call(t, "engine_forkchoiceUpdatedV3", &fc, payload.ForkchoiceStateV1{HeadBlockHash: head}, attrs)
		if fc.PayloadStatus.Status != "VALID" || fc.PayloadID == nil {
			t.Fatalf("forkchoice at %d: %+v", timestamp, fc)
		}
		var response struct {
			ExecutionPayload  *payload.ExecutionPayloadV3
			ExecutionRequests []string
		}
		producer.call(t, "engine_getPayloadV4", &response, fc.PayloadID)
		block := response.ExecutionPayload
		if block == nil || block.ParentHash != head {
			t.Fatalf("wrong payload parent at %d", timestamp)
		}
		if len(block.Transactions) != len(wantTx) {
			t.Fatalf("payload at %d contains %d transactions, want %d", timestamp, len(block.Transactions), len(wantTx))
		}
		for _, raw := range wantTx {
			found := false
			for _, got := range block.Transactions {
				found = found || bytes.Equal(got, raw)
			}
			if !found {
				t.Fatalf("payload at %d did not carry submitted transaction %s", timestamp, crypto.Keccak256Hash(raw))
			}
		}
		if verifierNode.Blockchain().GetBlock(block.BlockHash) != nil {
			t.Fatal("verifier already holds the producer payload")
		}
		var status payload.PayloadStatusV1
		verifier.call(t, "engine_newPayloadV4", &status, block, []types.Hash{}, beaconRoot, response.ExecutionRequests)
		if status.Status != "VALID" || status.LatestValidHash == nil || *status.LatestValidHash != block.BlockHash {
			t.Fatalf("independent import at %d: %+v", timestamp, status)
		}
		verifier.call(t, "engine_forkchoiceUpdatedV3", &fc, payload.ForkchoiceStateV1{HeadBlockHash: block.BlockHash}, nil)
		if fc.PayloadStatus.Status != "VALID" || verifierNode.Blockchain().CurrentBlock().Root() != block.StateRoot {
			t.Fatalf("verifier state root at %d does not match the payload", timestamp)
		}
		producer.call(t, "engine_forkchoiceUpdatedV3", &fc, payload.ForkchoiceStateV1{HeadBlockHash: block.BlockHash}, nil)
		if fc.PayloadStatus.Status != "VALID" {
			t.Fatalf("producer forkchoice at %d: %+v", timestamp, fc)
		}
		head = block.BlockHash
		return block
	}

	before := build(1)
	track4Precompiles(t, producer, "0x1", false)
	track4Storage(t, verifier, "0x1", 44)
	raw, txHash := track4SignedTx(t, big.NewInt(1337), 0, 0)
	var submitted types.Hash
	reject := func(raw []byte, reason string) {
		t.Helper()
		err := producer.invoke("eth_sendRawTransaction", &submitted, "0x"+hex.EncodeToString(raw))
		var rpcErr *track4RPCError
		if !errors.As(err, &rpcErr) || !strings.Contains(rpcErr.Message, reason) {
			t.Fatalf("expected transaction rejection %q, got %v", reason, err)
		}
	}
	reject(raw, "require I+ fork")
	after := build(2)
	if after.StateRoot == before.StateRoot {
		t.Fatal("overlay fork kept the MPT state root")
	}
	track4Precompiles(t, producer, "0x2", true)
	wrongChain, _ := track4SignedTx(t, big.NewInt(1338), 0, 0)
	reject(wrongChain, "chain ID mismatch")
	corrupt, err := types.DecodePQTransaction(raw)
	if err != nil {
		t.Fatal(err)
	}
	corrupt.PQSignature[0] ^= 1
	badSignature, err := corrupt.EncodePQ()
	if err != nil {
		t.Fatal(err)
	}
	reject(badSignature, "signature")
	producer.call(t, "eth_sendRawTransaction", &submitted, "0x"+hex.EncodeToString(raw))
	if submitted != txHash {
		t.Fatalf("transaction hash %s, want %s", submitted, txHash)
	}
	mined := build(3, raw)
	for _, r := range []track4RPC{producer, verifier} {
		var receipt *track4Receipt
		r.call(t, "eth_getTransactionReceipt", &receipt, txHash)
		if receipt == nil || receipt.Status != "0x1" || receipt.Type != "0x7" || receipt.BlockHash != mined.BlockHash {
			t.Fatalf("wrong transaction receipt: %+v", receipt)
		}
		track4Storage(t, r, "0x3", 0)
	}
	for _, tx := range other {
		producer.call(t, "eth_sendRawTransaction", &submitted, "0x"+hex.EncodeToString(tx.raw))
		if submitted != tx.hash {
			t.Fatalf("transaction hash %s, want %s", submitted, tx.hash)
		}
	}
	providers := build(4, other[0].raw, other[1].raw)
	for _, tx := range other {
		var receipt *track4Receipt
		verifier.call(t, "eth_getTransactionReceipt", &receipt, tx.hash)
		if receipt == nil || receipt.Status != "0x1" || receipt.BlockHash != providers.BlockHash || !strings.EqualFold(receipt.From, tx.sender.Hex()) {
			t.Fatalf("PQ provider receipt after independent import: %+v", receipt)
		}
	}
	var recipientBalance string
	verifier.call(t, "eth_getBalance", &recipientBalance, (types.Address{18: 0x71, 19: 1}), "0x4")
	if recipientBalance != "0x2" {
		t.Fatalf("verifier transfer balance %s, want 0x2", recipientBalance)
	}
	for ts := uint64(5); ts <= 11; ts++ {
		build(ts)
	}
	track4Storage(t, verifier, "latest", 0)
	track4Storage(t, verifier, "0x1", 44)

	mainHead := head
	selectHead := func(r track4RPC, hash types.Hash) {
		t.Helper()
		var fc payload.ForkchoiceUpdatedResult
		r.call(t, "engine_forkchoiceUpdatedV3", &fc, payload.ForkchoiceStateV1{HeadBlockHash: hash}, nil)
		if fc.PayloadStatus.Status != "VALID" {
			t.Fatalf("could not select %s: %+v", hash, fc)
		}
	}
	selectHead(producer, after.BlockHash)
	head = after.BlockHash
	alternate, alternateHash := track4SignedTx(t, big.NewInt(1337), 0, 57)
	producer.call(t, "eth_sendRawTransaction", &submitted, "0x"+hex.EncodeToString(alternate))
	if submitted != alternateHash {
		t.Fatalf("alternate transaction hash %s, want %s", submitted, alternateHash)
	}
	branch := build(20, alternate)
	if branch.BlockHash == mined.BlockHash || branch.StateRoot == mined.StateRoot {
		t.Fatal("competing payload did not change the state")
	}
	track4Storage(t, verifier, "latest", 57)
	selectHead(verifier, mainHead)
	head = mainHead
	track4Storage(t, verifier, "latest", 0)
	var orphanReceipt *track4Receipt
	verifier.call(t, "eth_getTransactionReceipt", &orphanReceipt, alternateHash)
	if orphanReceipt != nil {
		t.Fatalf("orphaned transaction still has a canonical receipt: %+v", orphanReceipt)
	}

	if err := verifierNode.Stop(); err != nil {
		t.Fatal(err)
	}
	reopened := newTestNode(t, verifierCfg)
	restarted := track4NodeRPC(t, reopened, token)
	if reopened.Blockchain().CurrentBlock().Hash() != head {
		t.Fatalf("restarted verifier head %s, want %s", reopened.Blockchain().CurrentBlock().Hash(), head)
	}
	track4Storage(t, restarted, "latest", 0)
	track4Storage(t, restarted, "0x3", 0)
	track4Storage(t, restarted, "0x1", 44)
	track4Precompiles(t, restarted, "0x2", true)
	var receipt *track4Receipt
	restarted.call(t, "eth_getTransactionReceipt", &receipt, txHash)
	pub, _ := track4Key()
	if receipt == nil || receipt.Status != "0x1" || receipt.BlockHash != mined.BlockHash ||
		!strings.EqualFold(receipt.From, types.PQPubKeyToAddress(pub.Bytes()).Hex()) ||
		receipt.EffectiveGasPrice != "0xb2d05e00" {
		t.Fatalf("receipt after restart: %+v", receipt)
	}
}

// TestTrack4LiveDevnet uses two separately running execution clients.
func TestTrack4LiveDevnet(t *testing.T) {
	urls := strings.Fields(os.Getenv("ETH2030_TRACK4_RPC"))
	if len(urls) == 0 {
		t.Skip("set ETH2030_TRACK4_RPC to the two Track 4 devnet RPC URLs")
	}
	if len(urls) != 2 || urls[0] == urls[1] {
		t.Fatal("set two distinct RPC URLs")
	}
	producer, verifier := track4RPC{url: urls[0]}, track4RPC{url: urls[1]}
	var chainHex string
	producer.call(t, "eth_chainId", &chainHex)
	chainID, ok := new(big.Int).SetString(strings.TrimPrefix(chainHex, "0x"), 16)
	if !ok {
		t.Fatalf("invalid chain ID %s", chainHex)
	}
	pub, _ := track4Key()
	sender := types.PQPubKeyToAddress(pub.Bytes())
	var nonceHex string
	producer.call(t, "eth_getTransactionCount", &nonceHex, sender, "latest")
	nonce, err := strconv.ParseUint(strings.TrimPrefix(nonceHex, "0x"), 16, 64)
	if err != nil {
		t.Fatal(err)
	}
	var genesis *track4BlockHeader
	producer.call(t, "eth_getBlockByNumber", &genesis, "0x0", false)
	if genesis == nil {
		t.Fatal("genesis block unavailable")
	}
	genesisTime, err := strconv.ParseUint(strings.TrimPrefix(genesis.Timestamp, "0x"), 16, 64)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range []track4RPC{producer, verifier} {
		track4Storage(t, r, "0x0", 44)
		track4Precompiles(t, r, "0x0", false)
		deadline := time.Now().Add(90 * time.Second)
		var active bool
		for time.Now().Before(deadline) {
			var latest *track4BlockHeader
			r.call(t, "eth_getBlockByNumber", &latest, "latest", false)
			if latest != nil {
				ts, err := strconv.ParseUint(strings.TrimPrefix(latest.Timestamp, "0x"), 16, 64)
				if err != nil {
					t.Fatal(err)
				}
				if ts >= genesisTime+24 && latest.StateRoot != genesis.StateRoot {
					active = true
					break
				}
			}
			time.Sleep(time.Second)
		}
		if !active {
			t.Fatalf("%s did not advance through binary activation and conversion", r.url)
		}
		track4Precompiles(t, r, "latest", true)
	}
	beacons := strings.Fields(os.Getenv("ETH2030_TRACK4_BEACON"))
	if len(beacons) != 2 || beacons[0] == beacons[1] {
		t.Fatal("set ETH2030_TRACK4_BEACON to the two Track 4 beacon API URLs")
	}
	var current types.Hash
	producer.call(t, "eth_getStorageAt", &current, track4Contract.Hex(), "0x0", "latest")
	if current.IsZero() {
		// Restore a nonzero slot so reruns still test a real transition to zero.
		verifyTrack4LiveTransaction(t, urls, beacons, chainID, genesisTime, nonce, 57)
		nonce++
	}
	verifyTrack4LiveTransaction(t, urls, beacons, chainID, genesisTime, nonce, 0)
}

type track4BlockHeader struct {
	Hash, StateRoot types.Hash
	Timestamp       string
}

func verifyTrack4LiveTransaction(t *testing.T, urls, beacons []string, chainID *big.Int, genesisTime, nonce uint64, value byte) {
	t.Helper()
	raw, hash := track4SignedTx(t, chainID, nonce, value)
	var accepted types.Hash
	(track4RPC{url: urls[0]}).call(t, "eth_sendRawTransaction", &accepted, "0x"+hex.EncodeToString(raw))
	if accepted != hash {
		t.Fatalf("RPC returned hash %s, want %s", accepted, hash)
	}
	deadline := time.Now().Add(90 * time.Second)
	pub, _ := track4Key()
	sender := types.PQPubKeyToAddress(pub.Bytes())
	var canonical track4BlockHeader
	for i, url := range urls {
		rpc := track4RPC{url: url}
		var receipt *track4Receipt
		for receipt == nil && time.Now().Before(deadline) {
			rpc.call(t, "eth_getTransactionReceipt", &receipt, hash)
			if receipt == nil {
				time.Sleep(time.Second)
			}
		}
		if receipt == nil || receipt.Status != "0x1" || receipt.Type != "0x7" ||
			!strings.EqualFold(receipt.From, sender.Hex()) || receipt.EffectiveGasPrice != "0xb2d05e00" {
			t.Fatalf("PQ transaction %s on %s: %+v", hash, rpc.url, receipt)
		}
		track4Storage(t, rpc, receipt.BlockNumber, value)
		var block *track4BlockHeader
		rpc.call(t, "eth_getBlockByNumber", &block, receipt.BlockNumber, false)
		if block == nil || block.Hash != receipt.BlockHash || block.StateRoot.IsZero() {
			t.Fatalf("receipt is not on the canonical chain of %s: %+v", rpc.url, block)
		}
		if i == 0 {
			canonical = *block
		} else if *block != canonical {
			t.Fatalf("execution nodes disagree on block header: %+v != %+v", *block, canonical)
		}
		t.Logf("node %s accepted %s in block %s (%s), state root %s", rpc.url, hash, receipt.BlockNumber, receipt.BlockHash, block.StateRoot)
	}
	ts, err := strconv.ParseUint(strings.TrimPrefix(canonical.Timestamp, "0x"), 16, 64)
	if err != nil {
		t.Fatal(err)
	}
	for _, url := range beacons {
		track4BeaconPayload(t, url, genesisTime, ts, canonical.Hash, canonical.StateRoot, raw)
	}
	if value == 0 {
		evidence, err := json.Marshal(map[string]any{
			"transaction":      hash,
			"block_hash":       canonical.Hash,
			"state_root":       canonical.StateRoot,
			"timestamp":        canonical.Timestamp,
			"storage_contract": track4Contract,
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("TRACK4_EVIDENCE %s", evidence)
	}
}

func track4BeaconPayload(t *testing.T, url string, genesisTime, at uint64, hash, root types.Hash, raw []byte) {
	t.Helper()
	client := &http.Client{Timeout: 10 * time.Second}
	get := func(path string, out any) {
		t.Helper()
		res, err := client.Get(strings.TrimRight(url, "/") + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("%s%s: HTTP %s", url, path, res.Status)
		}
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			t.Fatal(err)
		}
	}
	var cfg struct {
		Data struct {
			SecondsPerSlot string `json:"SECONDS_PER_SLOT"`
		}
	}
	get("/eth/v1/config/spec", &cfg)
	seconds, err := strconv.ParseUint(cfg.Data.SecondsPerSlot, 10, 64)
	if err != nil || seconds == 0 || at < genesisTime || (at-genesisTime)%seconds != 0 {
		t.Fatalf("invalid beacon slot configuration: seconds=%d, genesis=%d, block time=%d, err=%v", seconds, genesisTime, at, err)
	}
	slot := (at - genesisTime) / seconds
	var expected struct {
		Optimistic *bool `json:"execution_optimistic"`
		Data       struct {
			Message struct {
				Body struct {
					Payload struct {
						BlockHash    types.Hash `json:"block_hash"`
						StateRoot    types.Hash `json:"state_root"`
						Transactions []string
					} `json:"execution_payload"`
				}
			}
		}
	}
	get(fmt.Sprintf("/eth/v2/beacon/blocks/%d", slot), &expected)
	p := expected.Data.Message.Body.Payload
	if expected.Optimistic == nil || *expected.Optimistic || p.BlockHash != hash || p.StateRoot != root {
		t.Fatalf("beacon %s slot %d did not verify execution payload %s with root %s", url, slot, hash, root)
	}
	for _, tx := range p.Transactions {
		if tx == "0x"+hex.EncodeToString(raw) {
			t.Logf("beacon %s verified payload at slot %d", url, slot)
			return
		}
	}
	t.Fatalf("beacon %s slot %d did not carry the signed PQ transaction", url, slot)
}
