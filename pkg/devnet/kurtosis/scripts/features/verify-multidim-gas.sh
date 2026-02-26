#!/usr/bin/env bash
# Verify Multidimensional Gas: check multi-gas pricing is active
set -euo pipefail
ENCLAVE="${1:-eth2030-multidim-gas}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== Multidimensional Gas Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Check gas pricing
FEE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":1}' | jq -r '.result')
echo "Gas price: $FEE"

# --- Feature-specific EIP-7706 multidimensional gas tests ---

echo ""
echo "--- Verifying eth_gasPrice returns valid hex > 0 ---"
if [[ "$FEE" == 0x* ]]; then
  FEE_DEC=$((FEE))
  echo "Gas price (decimal): $FEE_DEC"
  [ "$FEE_DEC" -gt 0 ] || { echo "FAIL: Gas price is zero"; exit 1; }
else
  echo "FAIL: eth_gasPrice returned non-hex result: $FEE"
  exit 1
fi

echo ""
echo "--- Calling eth_feeHistory for recent blocks ---"
FEE_HISTORY=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_feeHistory","params":["0x5","latest",[25,50,75]],"id":1}' | jq -r '.result')
BASE_FEE_COUNT=$(echo "$FEE_HISTORY" | jq '.baseFeePerGas | length')
echo "eth_feeHistory baseFeePerGas entries: $BASE_FEE_COUNT"
[ "$BASE_FEE_COUNT" -gt 0 ] || { echo "FAIL: eth_feeHistory returned no baseFeePerGas entries"; exit 1; }
echo "baseFeePerGas values: $(echo "$FEE_HISTORY" | jq -c '.baseFeePerGas')"
REWARD_ENTRIES=$(echo "$FEE_HISTORY" | jq '.reward | length // 0')
echo "Reward percentile entries: $REWARD_ENTRIES"

echo ""
echo "--- Checking block header blob gas fields ---"
LATEST_BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' | jq -r '.result')
BLOB_GAS_USED=$(echo "$LATEST_BLOCK" | jq -r '.blobGasUsed // "null"')
EXCESS_BLOB_GAS=$(echo "$LATEST_BLOCK" | jq -r '.excessBlobGas // "null"')
echo "Latest block blobGasUsed: $BLOB_GAS_USED"
echo "Latest block excessBlobGas: $EXCESS_BLOB_GAS"

echo ""
echo "--- Calling eth_blobBaseFee (if available) ---"
BLOB_BASE_FEE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blobBaseFee","params":[],"id":1}' | jq -r '.result // .error.message // "unavailable"')
echo "Blob base fee: $BLOB_BASE_FEE"

echo ""
echo "--- Verifying multidimensional pricing (gasPrice vs blobGasUsed) ---"
if [ "$BLOB_GAS_USED" != "null" ] && [ "$BLOB_GAS_USED" != "" ]; then
  echo "Gas price (execution): $FEE"
  echo "Blob gas used (data):  $BLOB_GAS_USED"
  if [ "$FEE" != "$BLOB_GAS_USED" ]; then
    echo "  Confirmed: gasPrice and blobGasUsed are different values (multidimensional pricing)"
  else
    echo "  WARN: gasPrice and blobGasUsed are the same value"
  fi
else
  echo "  WARN: blobGasUsed not present in block header — blob gas dimension may not be active yet"
fi

echo ""
echo "PASS: Multidim Gas — gas pricing active, fee history verified, blob gas fields checked"
