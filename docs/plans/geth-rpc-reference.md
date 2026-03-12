# Geth JSON-RPC Reference Responses

Real responses captured from a live Geth v1.17.2 devnet node (geth-devnet, chainId 3151908 / 0x301824).
Use this as ground truth to validate our own RPC implementation.

**Node**: `Geth/v1.17.2-unstable-7d13acd0-20260312/linux-amd64/go1.26.1`
**Chain ID**: 3151908 (0x301824)
**Endpoint**: `http://127.0.0.1:32815` (kurtosis enclave `geth-devnet`)

---

## web3 Namespace

### web3_clientVersion

```json
{"jsonrpc":"2.0","id":1,"result":"Geth/v1.17.2-unstable-7d13acd0-20260312/linux-amd64/go1.26.1"}
```

### web3_sha3

Input: `0x68656c6c6f20776f726c64` (keccak256 of "hello world")

```json
{"jsonrpc":"2.0","id":1,"result":"0x47173285a8d7341e5e972fc677286384f802f8ef42a5ec5f03bbfa254cb01fad"}
```

---

## net Namespace

### net_version

```json
{"jsonrpc":"2.0","id":1,"result":"3151908"}
```

Note: returns a **decimal string**, not hex.

### net_listening

```json
{"jsonrpc":"2.0","id":1,"result":true}
```

### net_peerCount

```json
{"jsonrpc":"2.0","id":1,"result":"0x0"}
```

---

## eth Namespace — Chain/Node State

### eth_protocolVersion

**NOT AVAILABLE** in modern Geth (PoS, post-merge):

```json
{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"the method eth_protocolVersion does not exist/is not available"}}
```

### eth_syncing

When fully synced (returns `false`):

```json
{"jsonrpc":"2.0","id":1,"result":false}
```

### eth_coinbase

**NOT AVAILABLE** in modern Geth (no mining):

```json
{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"the method eth_coinbase does not exist/is not available"}}
```

### eth_mining

**NOT AVAILABLE** in modern Geth:

```json
{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"the method eth_mining does not exist/is not available"}}
```

### eth_hashrate

**NOT AVAILABLE** in modern Geth (PoS):

```json
{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"the method eth_hashrate does not exist/is not available"}}
```

### eth_chainId

```json
{"jsonrpc":"2.0","id":1,"result":"0x301824"}
```

### eth_gasPrice

```json
{"jsonrpc":"2.0","id":1,"result":"0x7824b721"}
```

### eth_maxPriorityFeePerGas

```json
{"jsonrpc":"2.0","id":1,"result":"0x77359400"}
```

### eth_blobBaseFee

```json
{"jsonrpc":"2.0","id":1,"result":"0x1"}
```

### eth_accounts

```json
{"jsonrpc":"2.0","id":1,"result":[]}
```

### eth_blockNumber

```json
{"jsonrpc":"2.0","id":1,"result":"0x2a"}
```

---

## eth Namespace — Account State

### eth_getBalance

Params: `["0x8943545177806ed17b9f23f0a21ee5948ecaa776", "latest"]`

```json
{"jsonrpc":"2.0","id":1,"result":"0x33b2e3ca9c188bd312da458"}
```

### eth_getStorageAt

Params: `["0x8943545177806ed17b9f23f0a21ee5948ecaa776", "0x0", "latest"]`

```json
{"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000000000000000000000000000000000000000000000"}
```

Note: returns full 32-byte zero-padded hex.

### eth_getTransactionCount

Params: `["0x8943545177806ed17b9f23f0a21ee5948ecaa776", "latest"]`

```json
{"jsonrpc":"2.0","id":1,"result":"0x0"}
```

### eth_getCode

Params: `["0x8943545177806ed17b9f23f0a21ee5948ecaa776", "latest"]` (EOA — no code)

```json
{"jsonrpc":"2.0","id":1,"result":"0x"}
```

### eth_getProof

Params: `["0x8943545177806ed17b9f23f0a21ee5948ecaa776", ["0x0"], "latest"]`

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "address": "0x8943545177806ed17b9f23f0a21ee5948ecaa776",
    "accountProof": [
      "0xf90211a0123b7c5e5271aae60c3b561e5acb7e65aa8d1e82660220ceab7bbe8c35fa1ef6...",
      "0xf90211a00eef0d0244b5f438a62c7c7535940353bd05bee3518f1245dacae4ea827b0ffaa...",
      "0xf8b1a09e9f7450e306207e301c94b76bab54c21f41220e6dd06de7c944c5d50d91764e...",
      "0xf8749f32dee2834b372effea65d7c74ab9f3ac60ef271e846721aa8759a346a3b554b852f850808c033b2e3cb02b414228190058a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a0c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
    ],
    "balance": "0x33b2e3cb02b414228190058",
    "codeHash": "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
    "nonce": "0x0",
    "storageHash": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
    "storageProof": [
      {
        "key": "0x0",
        "value": "0x0",
        "proof": []
      }
    ]
  }
}
```

---

## eth Namespace — Blocks

### eth_getBlockByNumber (hashes only, `false`)

Params: `["latest", false]`

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "baseFeePerGas": "0xda1805",
    "blobGasUsed": "0x40000",
    "difficulty": "0x0",
    "excessBlobGas": "0x0",
    "extraData": "0xd883011102846765746888676f312e32362e31856c696e7578",
    "gasLimit": "0x3938700",
    "gasUsed": "0x864338",
    "hash": "0x2b8218f2291d4934d2f58c571cb03855e5ab1bd05ae4cedac29ab489690fa8f9",
    "logsBloom": "0x002800000000000000000002200000000000104000004001000000001000000008020000000000000000002001000000000042000000000000040010000022000000000020000000000000280020000000000000000100000000000000000000000000800200000012801801000408000010000000000000100000100000000000000400400000000000800004000800000000028000001000000440004000001000000000000000000000080000000c000000000000008040000000000200206000000020000020100080000000001009000000000080080000000000001200200000000000000001004004800000000000001000000000000000800000000",
    "miner": "0x8943545177806ed17b9f23f0a21ee5948ecaa776",
    "mixHash": "0xf0fa2d915c7de391deb0aeae299d6ac681b8ba109964266c1d9d841b65c1fb0c",
    "nonce": "0x0000000000000000",
    "number": "0x2b",
    "parentBeaconBlockRoot": "0x31fe6b84e8847929bea8810cc0b0096c4c1d8297f211fa099417022bc500874a",
    "parentHash": "0xefc8b0bd84dec4930b5e3b9c9204fa9053f6a66ad6dd30fe8e7693a9abde39ca",
    "receiptsRoot": "0x84c1f20d3bb9d37dcd03bf85bc63a7b72814b75953bb99f4e6ed2e00b93e0a31",
    "requestsHash": "0xe3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
    "size": "0x1110",
    "stateRoot": "0x2eaf1b18d26d2c6324b982b38bda951705d676db33bec27d21753f7af831b739",
    "timestamp": "0x69b22a25",
    "transactions": [
      "0x3fc2d02de2eae782d51eb30f6fd34d509935310d6a4d5d2916b27b43163de1f3",
      "0x57b591108ea51cae707e858de23976038546363c513d09989de4a5a5499d2d7a"
    ],
    "transactionsRoot": "0x46a9939059b1d15e1459daa8e3990a3b0cc49692b7a9306bd638cd345735e8b2",
    "uncles": [],
    "withdrawals": [],
    "withdrawalsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
  }
}
```

**Key fields present in post-Prague blocks** (beyond legacy):
- `parentBeaconBlockRoot` — EIP-4788
- `requestsHash` — EIP-7685
- `blobGasUsed`, `excessBlobGas` — EIP-4844
- `withdrawals`, `withdrawalsRoot` — EIP-4895
- `nonce` always `"0x0000000000000000"` (PoS)
- `difficulty` always `"0x0"` (PoS)
- `sha3Uncles` always `"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"` (empty)

### eth_getBlockByHash (hashes only, `false`)

Same structure as `eth_getBlockByNumber`.

### eth_getBlockByNumber (full objects, `true`)

Each transaction in the array becomes a full object (see Transaction Object below).

### eth_getBlockTransactionCountByHash

Params: `["0x2b8218f2291d4934d2f58c571cb03855e5ab1bd05ae4cedac29ab489690fa8f9"]`

```json
{"jsonrpc":"2.0","id":1,"result":"0x11"}
```

### eth_getBlockTransactionCountByNumber

Params: `["latest"]`

```json
{"jsonrpc":"2.0","id":1,"result":"0x11"}
```

### eth_getUncleCountByBlockHash / eth_getUncleCountByBlockNumber

```json
{"jsonrpc":"2.0","id":1,"result":"0x0"}
```

### eth_getUncleByBlockHashAndIndex / eth_getUncleByBlockNumberAndIndex

No uncles in PoS:

```json
{"jsonrpc":"2.0","id":1,"result":null}
```

### eth_getBlockReceipts

Params: `["0x1"]` (empty block)

```json
{"jsonrpc":"2.0","id":1,"result":[]}
```

---

## eth Namespace — Transactions

### Transaction Object (type 0x2 — EIP-1559)

```json
{
  "blockHash": "0x2b8218f2291d4934d2f58c571cb03855e5ab1bd05ae4cedac29ab489690fa8f9",
  "blockNumber": "0x2b",
  "blockTimestamp": "0x69b22a25",
  "from": "0x14934f5dc8935eae39752caf731fff5cd078ff6c",
  "gas": "0x5208",
  "gasPrice": "0x780fac05",
  "maxFeePerGas": "0x4a817c800",
  "maxPriorityFeePerGas": "0x77359400",
  "hash": "0x3fc2d02de2eae782d51eb30f6fd34d509935310d6a4d5d2916b27b43163de1f3",
  "input": "0x",
  "nonce": "0x4",
  "to": "0xa198027aeb79e0651347f0e8bdd715edb6f30c8f",
  "transactionIndex": "0x0",
  "value": "0x4a817c800",
  "type": "0x2",
  "accessList": [],
  "chainId": "0x301824",
  "v": "0x1",
  "r": "0xdf63c232482840b106deacb54bda82af22c7d713a6858c3fa89c2af56e19077d",
  "s": "0x3b16f5de13eaee2c32fd6fe4e0f8e3ecff371868ba286ed9ab10310cca54888e",
  "yParity": "0x1"
}
```

**Notable**: `blockTimestamp` is an **extra field** Geth adds (not in EIP spec, but widely expected by clients).

### eth_getTransactionByHash

Params: `["0x3fc2d02de2eae782d51eb30f6fd34d509935310d6a4d5d2916b27b43163de1f3"]`

Returns the Transaction Object above.

### eth_getTransactionByBlockHashAndIndex

Params: `["<blockHash>", "0x0"]` → Transaction Object.

### eth_getTransactionByBlockNumberAndIndex

Params: `["latest", "0x0"]` → Transaction Object.

### eth_getTransactionReceipt

Params: `["0x3fc2d02de2eae782d51eb30f6fd34d509935310d6a4d5d2916b27b43163de1f3"]`

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "blockHash": "0x2b8218f2291d4934d2f58c571cb03855e5ab1bd05ae4cedac29ab489690fa8f9",
    "blockNumber": "0x2b",
    "contractAddress": null,
    "cumulativeGasUsed": "0x5208",
    "effectiveGasPrice": "0x780fac05",
    "from": "0x14934f5dc8935eae39752caf731fff5cd078ff6c",
    "gasUsed": "0x5208",
    "logs": [],
    "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "status": "0x1",
    "to": "0xa198027aeb79e0651347f0e8bdd715edb6f30c8f",
    "transactionHash": "0x3fc2d02de2eae782d51eb30f6fd34d509935310d6a4d5d2916b27b43163de1f3",
    "transactionIndex": "0x0",
    "type": "0x2"
  }
}
```

**Notable**: No `blobGasUsed` / `blobGasPrice` for non-blob txs. For blob txs those fields appear.

### eth_sendRawTransaction (invalid input)

```json
{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"rlp: value size exceeds available input length"}}
```

### eth_call

Params: `[{"from": "0x...", "to": "0x...", "data": "0x"}, "latest"]`

```json
{"jsonrpc":"2.0","id":1,"result":"0x"}
```

### eth_estimateGas

Params: `[{"from": "0x...", "to": "0x...", "value": "0x1"}, "latest"]` (simple ETH transfer)

```json
{"jsonrpc":"2.0","id":1,"result":"0x5208"}
```

### eth_createAccessList

Params: `[{"from": "0x...", "to": "0x...", "value": "0x1"}, "latest"]`

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "accessList": [],
    "gasUsed": "0x5208"
  }
}
```

### eth_sign

Fails without unlocked account:

```json
{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"unknown account"}}
```

### eth_signTransaction

Fails without unlocked account:

```json
{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"gas not specified"}}
```

### eth_sendTransaction

Fails without unlocked account:

```json
{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"unknown account"}}
```

---

## eth Namespace — Filters & Logs

### eth_newFilter

Params: `[{"fromBlock": "0x1", "toBlock": "latest"}]`

```json
{"jsonrpc":"2.0","id":1,"result":"0x63f7b6dcd2e908aa955be57577b8e888"}
```

Filter ID is a 16-byte hex string (128-bit random ID).

### eth_newBlockFilter

```json
{"jsonrpc":"2.0","id":1,"result":"0x599d8f2cf39b84227cdac10d13841d8e"}
```

### eth_newPendingTransactionFilter

```json
{"jsonrpc":"2.0","id":1,"result":"0x2ebb50206bc47783461748c7a114237d"}
```

### eth_getFilterChanges (log filter)

Returns array of Log objects since last poll:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": [
    {
      "address": "0x0a9f13aa4d49073a5fdddecdc6192aa33e6905a1",
      "topics": [
        "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
        "0x0000000000000000000000000000000000000000000000000000000000000000",
        "0x000000000000000000000000d48098099eac0841db920743e515c7d16d85d3bc"
      ],
      "data": "0x00000000000000000000000000000000000000000000000000000004a817c800",
      "blockNumber": "0x37",
      "transactionHash": "0x57c43e1f5f93899afffbe207a8a05b6015fc8fa586f437bce9f07f907197838d",
      "transactionIndex": "0x0",
      "blockHash": "0x3376390356cc24c9e8fb40581072d358e59706e156ccad374642e8cf634b22b3",
      "blockTimestamp": "0x69b22a61",
      "logIndex": "0x0",
      "removed": false
    }
  ]
}
```

**Log object fields**: `address`, `topics[]`, `data`, `blockNumber`, `transactionHash`, `transactionIndex`, `blockHash`, `blockTimestamp` (Geth extension), `logIndex`, `removed`.

### eth_getFilterChanges (block filter)

Returns array of block hashes:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": [
    "0x7019bac55027f10eb237aa139a0ff18f35d4aacbb81a121689da9d3327fb9e29",
    "0xaab5d5424358c4978510814d399ceb5e911bb9faf4fba1cda5c92f3d917c4641",
    "0xfa70e9472b64ac08e53c1be1cb152985992970641951ff6f805c2e180688ccfa"
  ]
}
```

### eth_getFilterChanges (pending tx filter)

Returns array of tx hashes (empty if no new pending):

```json
{"jsonrpc":"2.0","id":1,"result":[]}
```

### eth_getFilterLogs

Same as `eth_getLogs` but by filter ID. Returns array of Log objects.

### eth_uninstallFilter

```json
{"jsonrpc":"2.0","id":1,"result":true}
```

### eth_getLogs

Params: `[{"fromBlock": "0x37", "toBlock": "0x37"}]` — same Log object array as above.

Empty result for blocks with no matching logs:

```json
{"jsonrpc":"2.0","id":1,"result":[]}
```

---

## eth Namespace — Fee History

### eth_feeHistory

Params: `["0x5", "latest", [25, 75]]`

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "oldestBlock": "0x42",
    "reward": [
      ["0x77359400", "0x77359400"],
      ["0x77359400", "0x77359400"],
      ["0x77359400", "0x77359400"],
      ["0x77359400", "0x77359400"],
      ["0x77359400", "0x77359400"]
    ],
    "baseFeePerGas": [
      "0x1a39f0",
      "0x17ea8b",
      "0x15d0da",
      "0x13e5e4",
      "0x1223e9",
      "0x108bab"
    ],
    "gasUsedRatio": [
      0.14766136666666665,
      0.14871121666666667,
      0.14836136666666666,
      0.14665053333333333,
      0.1483612
    ],
    "baseFeePerBlobGas": [
      "0x1", "0x1", "0x1", "0x1", "0x1", "0x1"
    ],
    "blobGasUsedRatio": [
      0.13333333333333333,
      0.26666666666666666,
      0.13333333333333333,
      0.13333333333333333,
      0.13333333333333333
    ]
  }
}
```

**Notable**: `baseFeePerBlobGas` and `blobGasUsedRatio` are EIP-4844 extensions. Array has N+1 entries for `baseFeePerGas` / `baseFeePerBlobGas` (includes next block's predicted fee).

---

## admin Namespace

### admin_nodeInfo

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "id": "12d15ed3cffe1f7690055f34f130a3ee414490cc72e0d0057ea8352e2e53087f",
    "name": "Geth/v1.17.2-unstable-7d13acd0-20260312/linux-amd64/go1.26.1",
    "enode": "enode://c823154233d814f3b9000117b5a4e32e2518e770307e9ac05e755ff78afefc5c840e3456aea61918ffb9a38ee0a1148d0a04b2fe61ae75018aecac80fbcf8a2d@172.16.8.10:30303",
    "enr": "enr:-KO4QEc-Qj91rdwAL7nO8HL9QS4BQN3uaL1fW3KtApN5x3L8WafPAZjFmiHOpjk4Sgm_TApwrFsGa9_j0TAb2LUlDXqGAZzf8SICg2V0aMfGhFCQRk2AgmlkgnY0gmlwhKwQCAqJc2VjcDI1NmsxoQPIIxVCM9gU87kAARe1pOMuJRjncDB-msBedV_3iv78XIRzbmFwwIN0Y3CCdl-DdWRwgnZf",
    "ip": "172.16.8.10",
    "ports": {
      "discovery": 30303,
      "listener": 30303
    },
    "listenAddr": "[::]:30303",
    "protocols": {
      "eth": {
        "network": 3151908,
        "genesis": "0x4ca5f2bbf066120bc9995fd0cea57b2ab810f68de73f66eb1353a9f451408b2f",
        "config": {
          "chainId": 3151908,
          "homesteadBlock": 0,
          "eip150Block": 0,
          "eip155Block": 0,
          "eip158Block": 0,
          "byzantiumBlock": 0,
          "constantinopleBlock": 0,
          "petersburgBlock": 0,
          "istanbulBlock": 0,
          "berlinBlock": 0,
          "londonBlock": 0,
          "mergeNetsplitBlock": 0,
          "shanghaiTime": 0,
          "cancunTime": 0,
          "pragueTime": 0,
          "osakaTime": 0,
          "bpo1Time": 0,
          "terminalTotalDifficulty": 0,
          "depositContractAddress": "0x00000000219ab540356cbb839cbe05303d7705fa",
          "blobSchedule": {
            "cancun": {"target": 3, "max": 6, "baseFeeUpdateFraction": 3338477},
            "prague": {"target": 6, "max": 9, "baseFeeUpdateFraction": 5007716},
            "osaka": {"target": 6, "max": 9, "baseFeeUpdateFraction": 5007716},
            "bpo1": {"target": 10, "max": 15, "baseFeeUpdateFraction": 8346193}
          }
        },
        "head": "0xf77ceccbdd13b929dfe350905d72bd9d852d6b486cd6a7ccb76e04201b855efc"
      },
      "snap": {}
    }
  }
}
```

---

## txpool Namespace

### txpool_status

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "pending": "0x14",
    "queued": "0x0"
  }
}
```

### txpool_inspect

Human-readable summary: `"<to>: <value> wei + <gas> gas × <gasPrice> wei"`

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "pending": {
      "0x038234d79F028c4d96cb0456AC1e274b2e0292E5": {
        "6": "0xB730c270B2c67A57aA8277143fC363Caf3d53dFD: 20000000000 wei + 300000 gas × 20000000000 wei"
      }
    },
    "queued": {}
  }
}
```

### txpool_content

Full transaction objects organized by sender → nonce.

---

## Key Observations for Our RPC Implementation

### Methods NOT available in modern Geth (PoS)
- `eth_protocolVersion` → `-32601`
- `eth_coinbase` → `-32601`
- `eth_mining` → `-32601`
- `eth_hashrate` → `-32601`

### Geth-Specific Extensions (not in EIP spec, but widely expected)
- `blockTimestamp` field on transaction objects
- `blockTimestamp` field on log objects (in `eth_getLogs`, `eth_getFilterChanges`)
- `baseFeePerBlobGas` and `blobGasUsedRatio` in `eth_feeHistory` (EIP-4844 extension)
- `yParity` field on type-0x2 and type-0x3 transactions (alias for `v`)

### Post-Prague Block Fields (required)
- `parentBeaconBlockRoot` — EIP-4788 (Cancun+)
- `requestsHash` — EIP-7685 (Prague+)
- `blobGasUsed` / `excessBlobGas` — EIP-4844 (Cancun+)
- `withdrawals` / `withdrawalsRoot` — EIP-4895 (Shanghai+)

### net_version returns decimal, not hex
`"3151908"` not `"0x301824"`. This is spec-compliant.

### eth_getStorageAt returns 32-byte padded zero
`"0x0000000000000000000000000000000000000000000000000000000000000000"` not `"0x0"`.

### Filter IDs
16-byte (128-bit) hex strings: `"0x63f7b6dcd2e908aa955be57577b8e888"`.

### Error codes
- `-32601`: Method not found
- `-32000`: General execution error (invalid input, unknown account, etc.)
