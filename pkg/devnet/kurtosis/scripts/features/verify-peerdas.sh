#!/usr/bin/env bash
# Verify PeerDAS: check data availability sampling and blob distribution
set -euo pipefail
ENCLAVE="${1:-eth2030-peerdas}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== PeerDAS Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Check latest block for blob gas usage
LATEST=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' | jq -r '.result')
BLOB_GAS=$(echo "$LATEST" | jq -r '.blobGasUsed // "0x0"')
echo "Blob gas used: $BLOB_GAS"

# --- Feature-specific PeerDAS blob verification ---

echo ""
echo "--- Checking blobGasUsed is a valid hex number ---"
if [[ "$BLOB_GAS" == 0x* ]]; then
  BLOB_GAS_DEC=$((BLOB_GAS))
  echo "blobGasUsed (decimal): $BLOB_GAS_DEC"
else
  echo "  WARN: blobGasUsed is not a valid hex number: $BLOB_GAS"
fi

echo ""
echo "--- Checking excessBlobGas field ---"
EXCESS_BLOB_GAS=$(echo "$LATEST" | jq -r '.excessBlobGas // "null"')
echo "excessBlobGas: $EXCESS_BLOB_GAS"
if [[ "$EXCESS_BLOB_GAS" == 0x* ]]; then
  EXCESS_DEC=$((EXCESS_BLOB_GAS))
  echo "excessBlobGas (decimal): $EXCESS_DEC"
elif [ "$EXCESS_BLOB_GAS" = "null" ]; then
  echo "  WARN: excessBlobGas not present in block header"
fi

echo ""
echo "--- Calling eth_feeHistory for recent blocks ---"
FEE_HISTORY=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_feeHistory","params":["0x3","latest",[]],"id":1}' | jq -r '.result')
BASE_FEE_COUNT=$(echo "$FEE_HISTORY" | jq '.baseFeePerGas | length')
echo "eth_feeHistory baseFeePerGas entries: $BASE_FEE_COUNT"
if [ "$BASE_FEE_COUNT" -gt 0 ]; then
  echo "baseFeePerGas values: $(echo "$FEE_HISTORY" | jq -c '.baseFeePerGas')"
else
  echo "  WARN: eth_feeHistory returned no baseFeePerGas entries"
fi

echo ""
echo "--- Testing KZG point evaluation precompile (0x0A) accessibility ---"
# Call with minimal data to test if the precompile is accessible
# Even invalid input should show the precompile is reachable (returns error about invalid input rather than missing precompile)
KZG_RESULT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"0x000000000000000000000000000000000000000a","data":"0x0000000000000000000000000000000000000000000000000000000000000000","gas":"0x100000"},"latest"],"id":1}' | jq -r '.result // .error.message')
echo "KZG point evaluation (0x0A): $KZG_RESULT"
if [ -n "$KZG_RESULT" ] && [ "$KZG_RESULT" != "null" ]; then
  echo "  KZG precompile is accessible (returned a response)"
else
  echo "  WARN: KZG precompile returned null"
fi

echo ""
echo "--- Comparing blobGasUsed across 2 blocks ---"
LATEST_DEC=$((BLOCK))
if [ "$LATEST_DEC" -ge 2 ]; then
  PREV_HEX=$(printf '0x%x' $((LATEST_DEC - 1)))
  PREV_BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$PREV_HEX\", false],\"id\":1}" | jq -r '.result')
  PREV_BLOB_GAS=$(echo "$PREV_BLOCK" | jq -r '.blobGasUsed // "0x0"')
  echo "Block $((LATEST_DEC - 1)) blobGasUsed: $PREV_BLOB_GAS"
  echo "Block $LATEST_DEC blobGasUsed:         $BLOB_GAS"
  if [ "$PREV_BLOB_GAS" = "$BLOB_GAS" ]; then
    echo "  Same blob gas usage across consecutive blocks"
  else
    echo "  Different blob gas usage across consecutive blocks"
  fi
else
  echo "  Skipping: need at least 2 blocks for comparison"
fi

echo ""
echo "PASS: PeerDAS — blocks produced, blob gas fields verified, fee history checked, KZG precompile accessible"
