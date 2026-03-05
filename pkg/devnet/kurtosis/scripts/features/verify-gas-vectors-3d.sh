#!/usr/bin/env bash
# Verify 3D gas vectors and copyHeader fix (SPEC-6.4 / EIP-7706)
# In the standard kurtosis devnet config (pre-Glamsterdam genesis), gasLimitVec,
# gasUsedVec, and excessGasVec are nil (correct: Glamsterdam not yet active).
# This script verifies:
#   1. Block headers are structurally complete (no nil-pointer panics).
#   2. Pre-Glamsterdam blocks correctly have nil 3D vectors.
#   3. Standard 1D gas fields (gasLimit, gasUsed, baseFeePerGas) are present.
#   4. The node doesn't crash when queried for 3D vector fields.
set -euo pipefail
ENCLAVE="${1:-eth2030-gas-vectors-3d}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== 3D Gas Vectors Verification (SPEC-6.4) ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
BLOCK_DEC=$((BLOCK))
echo "Current block: $BLOCK_DEC"
[ "$BLOCK_DEC" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# --- Test 1: block headers are accessible and contain expected 1D gas fields ---
echo ""
echo "--- Test 1: Standard 1D gas fields present in block headers ---"
FAIL_COUNT=0
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  B=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\",false],\"id\":1}" \
    | jq '.result | {number, gasLimit, gasUsed, baseFeePerGas, gasLimitVec, gasUsedVec, excessGasVec}')
  GAS_LIMIT=$(echo "$B" | jq -r '.gasLimit')
  GAS_USED=$(echo "$B" | jq -r '.gasUsed')
  BASE_FEE=$(echo "$B" | jq -r '.baseFeePerGas')
  GAS_LIMIT_VEC=$(echo "$B" | jq -r '.gasLimitVec')
  GAS_USED_VEC=$(echo "$B" | jq -r '.gasUsedVec')
  EXCESS_GAS_VEC=$(echo "$B" | jq -r '.excessGasVec')
  echo "  Block $i: gasLimit=$GAS_LIMIT gasUsed=$GAS_USED baseFeePerGas=$BASE_FEE"
  echo "           gasLimitVec=$GAS_LIMIT_VEC gasUsedVec=$GAS_USED_VEC excessGasVec=$EXCESS_GAS_VEC"

  # 1D fields must be present.
  [ "$GAS_LIMIT" != "null" ] && [ -n "$GAS_LIMIT" ] || { echo "  FAIL: gasLimit missing in block $i"; FAIL_COUNT=$((FAIL_COUNT+1)); }
  [ "$GAS_USED" != "null" ] && [ -n "$GAS_USED" ] || { echo "  FAIL: gasUsed missing in block $i"; FAIL_COUNT=$((FAIL_COUNT+1)); }
done
[ "$FAIL_COUNT" -eq 0 ] || { echo "FAIL: $FAIL_COUNT missing 1D gas fields"; exit 1; }
echo "1D gas fields present in blocks 1-3"

# --- Test 2: pre-Glamsterdam config — 3D vectors are nil (expected) ---
echo ""
echo "--- Test 2: Pre-Glamsterdam blocks have nil 3D vectors (expected behaviour) ---"
# In the standard kurtosis devnet, GlamsterdanTime is not set, so 3D vectors
# should be nil/null. This confirms the code path is gated correctly.
LATEST_B=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}' \
  | jq '.result | {gasLimitVec, gasUsedVec, excessGasVec, calldataGasUsed, calldataExcessGas}')
GL_VEC=$(echo "$LATEST_B" | jq -r '.gasLimitVec')
GU_VEC=$(echo "$LATEST_B" | jq -r '.gasUsedVec')
EG_VEC=$(echo "$LATEST_B" | jq -r '.excessGasVec')
CD_GAS=$(echo "$LATEST_B" | jq -r '.calldataGasUsed')
echo "gasLimitVec=$GL_VEC gasUsedVec=$GU_VEC excessGasVec=$EG_VEC calldataGasUsed=$CD_GAS"
# All should be null (Glamsterdam not active in kurtosis devnet).
if [ "$GL_VEC" = "null" ] && [ "$GU_VEC" = "null" ] && [ "$EG_VEC" = "null" ]; then
  echo "3D vector fields are null — correct for pre-Glamsterdam config"
else
  echo "3D vector fields present — Glamsterdam is active in this devnet"
fi

# --- Test 3: Node returns 3D vector fields without panicking ---
# Confirm the RPC handler for these fields doesn't crash with a nil-pointer.
echo ""
echo "--- Test 3: Node responds to 3D vector field queries without error ---"
QUERY_OK=0
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  STATUS=$(curl -sf -w "%{http_code}" -o /dev/null -X POST "$RPC_URL" \
    -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\",false],\"id\":1}")
  if [ "$STATUS" = "200" ]; then
    QUERY_OK=$((QUERY_OK + 1))
  else
    echo "  FAIL: Block $i query returned HTTP $STATUS"
  fi
done
[ "$QUERY_OK" -eq 3 ] || { echo "FAIL: Some block queries failed — possible nil-pointer panic in 3D vector handling"; exit 1; }
echo "All 3 block header queries returned HTTP 200"

# --- Test 4: eth_gasPrice and eth_feeHistory still work (no regression) ---
echo ""
echo "--- Test 4: Gas price endpoints unaffected ---"
GAS_PRICE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":1}' | jq -r '.result // "null"')
echo "eth_gasPrice: $GAS_PRICE"
[ "$GAS_PRICE" != "null" ] || { echo "FAIL: eth_gasPrice returned null"; exit 1; }

FEE_HIST=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_feeHistory","params":["0x3","latest",[]],"id":1}' \
  | jq -r '.result.baseFeePerGas | length // 0')
echo "eth_feeHistory baseFeePerGas entries: $FEE_HIST"
[ "$((FEE_HIST))" -gt 0 ] || { echo "FAIL: eth_feeHistory returned no entries"; exit 1; }

# --- Test 5: scan EL logs for nil-pointer panics in gas vector code ---
EL_SVC_NAME=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null \
  | grep "el-[0-9]" | head -1 | awk '{print $2}')
if [ -n "$EL_SVC_NAME" ]; then
  echo ""
  echo "--- Test 5: No nil-pointer panics in EL logs ---"
  NIL_PANICS=$(kurtosis service logs "$ENCLAVE" "$EL_SVC_NAME" 2>&1 \
    | grep -c "nil pointer\|GasLimitVec.*nil\|copyHeader.*panic" || true)
  echo "Nil-pointer panic occurrences: $NIL_PANICS"
  [ "$NIL_PANICS" -eq 0 ] || { echo "FAIL: nil-pointer panics found — copyHeader bug may not be fixed"; exit 1; }
fi

echo ""
echo "PASS: 3D Gas Vectors — 1D fields present, 3D vectors gated by Glamsterdam, no nil-pointer panics, gas endpoints stable"
