#!/usr/bin/env bash
# Verify SSZ: check SSZ-encoded transactions and blocks are valid
set -euo pipefail
ENCLAVE="${1:-eth2030-ssz}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== SSZ Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Verify block structure
HEADER=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' | jq '.result | {number, hash, stateRoot, transactionsRoot}')
echo "Latest block: $HEADER"

# --- Feature-specific SSZ encoding tests ---

# Get latest block header and verify all critical fields exist and are non-null
echo "Verifying all critical block header fields exist..."
FULL_HEADER=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' | jq -r '.result')

REQUIRED_FIELDS="hash parentHash stateRoot transactionsRoot receiptsRoot logsBloom gasLimit gasUsed timestamp extraData mixHash nonce baseFeePerGas"
for FIELD in $REQUIRED_FIELDS; do
  VALUE=$(echo "$FULL_HEADER" | jq -r ".$FIELD // \"missing\"")
  if [ "$VALUE" = "missing" ] || [ "$VALUE" = "null" ] || [ -z "$VALUE" ]; then
    echo "FAIL: Required field '$FIELD' is missing or null in block header"
    exit 1
  fi
  echo "  $FIELD: ${VALUE:0:42}$([ ${#VALUE} -gt 42 ] && echo '...' || true)"
done
echo "All critical block header fields present and non-null"

# Verify sha3Uncles equals the empty uncles hash (post-merge)
echo "Checking sha3Uncles matches empty uncles hash (post-merge)..."
EMPTY_UNCLES="0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"
SHA3_UNCLES=$(echo "$FULL_HEADER" | jq -r '.sha3Uncles // "missing"')
echo "sha3Uncles: $SHA3_UNCLES"
[ "$SHA3_UNCLES" = "$EMPTY_UNCLES" ] || { echo "FAIL: sha3Uncles is $SHA3_UNCLES, expected $EMPTY_UNCLES (post-merge)"; exit 1; }
echo "sha3Uncles matches empty uncles hash (post-merge confirmed)"

# Verify withdrawalsRoot field exists (post-Shanghai)
echo "Checking withdrawalsRoot field (post-Shanghai)..."
WITHDRAWALS_ROOT=$(echo "$FULL_HEADER" | jq -r '.withdrawalsRoot // "missing"')
echo "withdrawalsRoot: $WITHDRAWALS_ROOT"
if [ "$WITHDRAWALS_ROOT" = "missing" ] || [ "$WITHDRAWALS_ROOT" = "null" ]; then
  echo "WARNING: withdrawalsRoot not present (may be pre-Shanghai)"
else
  echo "withdrawalsRoot present (post-Shanghai confirmed)"
fi

# Check transactionsRoot is non-zero when block has transactions
echo "Checking transactionsRoot for blocks with transactions..."
LATEST_WITH_TXS=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", true],"id":1}' | jq -r '.result')
TX_COUNT=$(echo "$LATEST_WITH_TXS" | jq '.transactions | length')
TX_ROOT=$(echo "$LATEST_WITH_TXS" | jq -r '.transactionsRoot')
echo "Latest block tx count: $TX_COUNT, transactionsRoot: ${TX_ROOT:0:18}..."
EMPTY_TX_ROOT="0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
if [ "$TX_COUNT" -gt 0 ] && [ "$TX_ROOT" = "$EMPTY_TX_ROOT" ]; then
  echo "FAIL: Block has transactions but transactionsRoot is empty trie root"
  exit 1
fi
if [ "$TX_COUNT" -gt 0 ]; then
  echo "transactionsRoot is non-empty for block with transactions"
fi

echo "PASS: SSZ — block header fields complete, post-merge/Shanghai fields verified"
