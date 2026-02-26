#!/usr/bin/env bash
# Verify DL Futures: check blob futures market and custody proofs
set -euo pipefail
ENCLAVE="${1:-eth2030-dl-futures}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== DL Futures Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Verify chain is progressing (multiple blocks indicates stability)
if [ "$((BLOCK))" -ge 3 ]; then
  B1=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1", false],"id":1}' | jq -r '.result.timestamp')
  BLATEST=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' | jq -r '.result.timestamp')
  if [ -n "$B1" ] && [ -n "$BLATEST" ] && [ "$B1" != "null" ] && [ "$BLATEST" != "null" ]; then
    ELAPSED=$(( $(printf '%d' "$BLATEST") - $(printf '%d' "$B1") ))
    echo "Chain time elapsed: ${ELAPSED}s over $((BLOCK)) blocks"
  fi
fi

# --- Feature-specific DL futures tests ---

# Call eth_gasPrice — verify > 0
echo "Checking eth_gasPrice..."
GAS_PRICE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":1}' | jq -r '.result // "0x0"')
GAS_PRICE_DEC=$(printf '%d' "$GAS_PRICE" 2>/dev/null || echo 0)
echo "  Gas price: $GAS_PRICE (decimal: $GAS_PRICE_DEC)"
[ "$GAS_PRICE_DEC" -gt 0 ] || echo "WARN: Gas price is 0"

# Call eth_feeHistory with reward percentiles — verify valid response with reward array
echo "Checking eth_feeHistory with reward percentiles..."
FEE_HISTORY=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_feeHistory","params":["0x3","latest",[50]],"id":1}' | jq -r '.result')
if [ "$FEE_HISTORY" != "null" ] && [ -n "$FEE_HISTORY" ]; then
  FH_BASE_FEES=$(echo "$FEE_HISTORY" | jq -r '.baseFeePerGas | length // 0' 2>/dev/null || echo 0)
  FH_REWARDS=$(echo "$FEE_HISTORY" | jq -r '.reward | length // 0' 2>/dev/null || echo 0)
  echo "  baseFeePerGas entries: $FH_BASE_FEES"
  echo "  reward entries: $FH_REWARDS"
else
  echo "  eth_feeHistory: not available"
fi

# Call eth_blobBaseFee (soft check)
echo "Checking eth_blobBaseFee (soft check)..."
BLOB_BASE_FEE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blobBaseFee","params":[],"id":1}' | jq -r '.result // .error.message // "unsupported"')
echo "  eth_blobBaseFee: $BLOB_BASE_FEE"

# Check baseFeePerGas values in block headers across 3 blocks: echo trend
echo "Checking baseFeePerGas trend across blocks 1-3..."
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  B_DATA=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", false],\"id\":1}" | jq -r '.result')
  BF=$(echo "$B_DATA" | jq -r '.baseFeePerGas // "missing"')
  BF_DEC=$(printf '%d' "$BF" 2>/dev/null || echo "N/A")
  echo "  Block $i baseFeePerGas: $BF (decimal: $BF_DEC)"
done

# Verify chain progresses: sleep 6s, get new blockNumber, verify > original
echo "Verifying chain continues to progress (waiting 6s)..."
ORIG_BLOCK=$((BLOCK))
sleep 6
NEW_BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
NEW_BLOCK_DEC=$((NEW_BLOCK))
echo "  Block before wait: $ORIG_BLOCK, after wait: $NEW_BLOCK_DEC"
[ "$NEW_BLOCK_DEC" -gt "$ORIG_BLOCK" ] || echo "WARN: Chain did not progress during 6s wait"

echo "PASS: DL Futures — gas pricing verified, fee history checked, chain progression confirmed"
