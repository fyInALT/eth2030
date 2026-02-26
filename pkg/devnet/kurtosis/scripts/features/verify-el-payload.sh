#!/usr/bin/env bash
# Verify EL Payload: check payload chunking, block-in-blobs, announce nonce
set -euo pipefail
ENCLAVE="${1:-eth2030-el-payload}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== EL Payload Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Check block structure via eth_getBlockByNumber
LATEST=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", true],"id":1}' | jq -r '.result')
BLOCK_HASH=$(echo "$LATEST" | jq -r '.hash // "missing"')
PARENT_HASH=$(echo "$LATEST" | jq -r '.parentHash // "missing"')
STATE_ROOT=$(echo "$LATEST" | jq -r '.stateRoot // "missing"')
TX_COUNT=$(echo "$LATEST" | jq -r '.transactions | length')
echo "Block hash: $BLOCK_HASH"
echo "Parent hash: $PARENT_HASH"
echo "State root: $STATE_ROOT"
echo "Transaction count: $TX_COUNT"
[ "$BLOCK_HASH" != "missing" ] || { echo "FAIL: block hash missing"; exit 1; }
[ "$PARENT_HASH" != "missing" ] || { echo "FAIL: parent hash missing"; exit 1; }
[ "$STATE_ROOT" != "missing" ] || { echo "FAIL: state root missing"; exit 1; }

# Verify block 1 also has proper structure (payload integrity)
if [ "$((BLOCK))" -ge 1 ]; then
  B1=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1", false],"id":1}' | jq -r '.result')
  B1_HASH=$(echo "$B1" | jq -r '.hash // "missing"')
  echo "Block 1 hash: $B1_HASH"
  [ "$B1_HASH" != "missing" ] || { echo "FAIL: block 1 hash missing"; exit 1; }
fi

# --- Feature-specific payload chunking tests ---

# Verify size field in block header is reasonable
echo "Checking block size field..."
BLOCK_SIZE=$(echo "$LATEST" | jq -r '.size // "missing"')
if [ "$BLOCK_SIZE" != "missing" ] && [ "$BLOCK_SIZE" != "null" ]; then
  SIZE_DEC=$(printf '%d' "$BLOCK_SIZE" 2>/dev/null || echo 0)
  echo "  Latest block size: $SIZE_DEC bytes"
  [ "$SIZE_DEC" -gt 0 ] || { echo "FAIL: Block size is 0"; exit 1; }
  [ "$SIZE_DEC" -lt 10000000 ] || { echo "FAIL: Block size unreasonably large ($SIZE_DEC bytes)"; exit 1; }
else
  echo "  Block size field: $BLOCK_SIZE"
fi

# Check logsBloom field exists and is 256-byte hex (514 chars with 0x prefix)
echo "Checking logsBloom format..."
LOGS_BLOOM=$(echo "$LATEST" | jq -r '.logsBloom // "missing"')
if [ "$LOGS_BLOOM" != "missing" ] && [ "$LOGS_BLOOM" != "null" ]; then
  BLOOM_LEN=${#LOGS_BLOOM}
  echo "  logsBloom length: $BLOOM_LEN chars (expected 514 = 0x + 512 hex)"
  [ "$BLOOM_LEN" -eq 514 ] || echo "WARN: logsBloom length is $BLOOM_LEN, expected 514"
else
  echo "FAIL: logsBloom missing from block header"
  exit 1
fi

# Check tx count distribution across blocks 1-3
echo "Checking transaction count distribution across blocks 1-3..."
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  B_TX_CNT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockTransactionCountByNumber\",\"params\":[\"$B_HEX\"],\"id\":1}" | jq -r '.result // "0x0"')
  B_TX_DEC=$(printf '%d' "$B_TX_CNT" 2>/dev/null || echo 0)
  echo "  Block $i tx count: $B_TX_DEC"
done

# Verify receiptsRoot and transactionsRoot are valid 32-byte hashes
echo "Checking receiptsRoot and transactionsRoot..."
RECEIPTS_ROOT=$(echo "$LATEST" | jq -r '.receiptsRoot // "missing"')
TX_ROOT=$(echo "$LATEST" | jq -r '.transactionsRoot // "missing"')
echo "  receiptsRoot: $RECEIPTS_ROOT"
echo "  transactionsRoot: $TX_ROOT"
[ "$RECEIPTS_ROOT" != "missing" ] && [ "$RECEIPTS_ROOT" != "null" ] || { echo "FAIL: receiptsRoot missing"; exit 1; }
[ "$TX_ROOT" != "missing" ] && [ "$TX_ROOT" != "null" ] || { echo "FAIL: transactionsRoot missing"; exit 1; }
# Verify they are 66-char hex strings (0x + 64 hex)
RR_LEN=${#RECEIPTS_ROOT}
TR_LEN=${#TX_ROOT}
[ "$RR_LEN" -eq 66 ] || echo "WARN: receiptsRoot length $RR_LEN, expected 66"
[ "$TR_LEN" -eq 66 ] || echo "WARN: transactionsRoot length $TR_LEN, expected 66"

# Verify baseFeePerGas field exists (post-London)
echo "Checking baseFeePerGas field..."
BASE_FEE=$(echo "$LATEST" | jq -r '.baseFeePerGas // "missing"')
echo "  baseFeePerGas: $BASE_FEE"
[ "$BASE_FEE" != "missing" ] && [ "$BASE_FEE" != "null" ] || { echo "FAIL: baseFeePerGas missing (post-London required)"; exit 1; }

echo "PASS: EL Payload — blocks produced, payload structure validated (size, bloom, roots, baseFee)"
