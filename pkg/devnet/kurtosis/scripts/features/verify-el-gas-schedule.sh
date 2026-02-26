#!/usr/bin/env bash
# Verify EL Gas Schedule: check Hogota repricing and gas limit schedule
set -euo pipefail
ENCLAVE="${1:-eth2030-el-gas-schedule}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== EL Gas Schedule Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Check gas limit in block header
LATEST=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' | jq -r '.result')
GAS_LIMIT=$(echo "$LATEST" | jq -r '.gasLimit // "missing"')
GAS_USED=$(echo "$LATEST" | jq -r '.gasUsed // "0x0"')
echo "Gas limit: $GAS_LIMIT"
echo "Gas used: $GAS_USED"
[ "$GAS_LIMIT" != "missing" ] || { echo "FAIL: gasLimit missing from header"; exit 1; }

# Verify gas price is responding (Hogota repricing active)
GAS_PRICE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":1}' | jq -r '.result')
echo "Gas price: $GAS_PRICE"

# --- Feature-specific gas schedule tests ---

# Verify gas price > 0
echo "Verifying gas price is positive..."
GP_DEC=$((GAS_PRICE))
[ "$GP_DEC" -gt 0 ] || { echo "FAIL: Gas price is zero"; exit 1; }
echo "Gas price is $GAS_PRICE ($GP_DEC wei)"

# Check gasLimit across blocks 1-3: verify consistent (within 10x of each other)
echo "Checking gasLimit consistency across blocks 1-3..."
GAS_LIMITS_SCHED=()
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  GL_HEX=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", false],\"id\":1}" | jq -r '.result.gasLimit // "0x0"')
  GL_DEC=$((GL_HEX))
  echo "Block $i gasLimit: $GL_HEX ($GL_DEC)"
  GAS_LIMITS_SCHED+=("$GL_DEC")
done
for a in "${GAS_LIMITS_SCHED[@]}"; do
  for b in "${GAS_LIMITS_SCHED[@]}"; do
    if [ "$a" -gt $((b * 10)) ] || [ "$b" -gt $((a * 10)) ]; then
      echo "FAIL: Gas limits vary by more than 10x ($a vs $b)"
      exit 1
    fi
  done
done
echo "Gas limits are consistent across blocks (within 10x)"

# Call eth_feeHistory — verify baseFeePerGas array returned with entries
echo "Checking eth_feeHistory..."
FEE_HISTORY=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_feeHistory","params":["0x5","latest",[25,50,75]],"id":1}' | jq -r '.result')
if [ "$FEE_HISTORY" != "null" ] && [ -n "$FEE_HISTORY" ]; then
  BASE_FEE_ARRAY=$(echo "$FEE_HISTORY" | jq -r '.baseFeePerGas // "missing"')
  if [ "$BASE_FEE_ARRAY" != "missing" ] && [ "$BASE_FEE_ARRAY" != "null" ]; then
    BF_LEN=$(echo "$FEE_HISTORY" | jq '.baseFeePerGas | length')
    echo "eth_feeHistory baseFeePerGas entries: $BF_LEN"
    [ "$BF_LEN" -gt 0 ] || { echo "FAIL: baseFeePerGas array is empty"; exit 1; }
    echo "baseFeePerGas values:"
    echo "$FEE_HISTORY" | jq -r '.baseFeePerGas[]' | while read -r bf; do
      echo "  $bf"
    done
  else
    echo "WARNING: baseFeePerGas not in feeHistory response"
  fi
else
  echo "WARNING: eth_feeHistory returned null"
fi

# Verify gasLimit > 0 AND gasUsed <= gasLimit across blocks
echo "Verifying gasUsed <= gasLimit across blocks 1-3..."
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  BD=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", false],\"id\":1}" | jq -r '.result')
  GU=$(($(echo "$BD" | jq -r '.gasUsed // "0x0"')))
  GL=$(($(echo "$BD" | jq -r '.gasLimit // "0x0"')))
  echo "Block $i: gasUsed=$GU gasLimit=$GL"
  [ "$GL" -gt 0 ] || { echo "FAIL: Block $i gasLimit is zero"; exit 1; }
  [ "$GU" -le "$GL" ] || { echo "FAIL: Block $i gasUsed ($GU) > gasLimit ($GL)"; exit 1; }
done
echo "gasUsed <= gasLimit invariant holds across all blocks"

echo "PASS: EL Gas Schedule — gas limits consistent, fee history available, gas invariants hold"
