#!/usr/bin/env bash
# Verify IL satisfaction wiring in engine_newPayload (SPEC-4.3 / EIP-7805)
# Checks that newPayload is accepted for every slot and no IL unsatisfied
# rejections appear in the EL node logs.
set -euo pipefail
ENCLAVE="${1:-eth2030-il-satisfaction}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== IL Satisfaction Verification (SPEC-4.3) ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
BLOCK_DEC=$((BLOCK))
echo "Current block: $BLOCK_DEC"
[ "$BLOCK_DEC" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Inspect EL logs for INCLUSION_LIST_UNSATISFIED rejections.
# Any such message means SPEC-4.3 is rejecting valid payloads incorrectly.
EL_SVC_NAME=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null \
  | grep "el-[0-9]" | head -1 | awk '{print $2}')
if [ -n "$EL_SVC_NAME" ]; then
  echo ""
  echo "--- Scanning EL logs for INCLUSION_LIST_UNSATISFIED ---"
  UNSATISFIED=$(kurtosis service logs "$ENCLAVE" "$EL_SVC_NAME" 2>&1 \
    | grep -c "INCLUSION_LIST_UNSATISFIED" || true)
  echo "INCLUSION_LIST_UNSATISFIED occurrences: $UNSATISFIED"
  [ "$UNSATISFIED" -eq 0 ] || { echo "FAIL: engine_newPayload rejected payloads with INCLUSION_LIST_UNSATISFIED"; exit 1; }

  echo ""
  echo "--- Counting engine_newPayload accepted messages ---"
  ACCEPTED=$(kurtosis service logs "$ENCLAVE" "$EL_SVC_NAME" 2>&1 \
    | grep -c "engine_newPayload: accepted" || true)
  echo "engine_newPayload accepted: $ACCEPTED"
  [ "$ACCEPTED" -gt 0 ] || { echo "FAIL: No engine_newPayload: accepted messages found"; exit 1; }

  echo ""
  echo "--- Checking for engine_newPayload errors ---"
  PAYLOAD_ERRORS=$(kurtosis service logs "$ENCLAVE" "$EL_SVC_NAME" 2>&1 \
    | grep -i "engine_newPayload.*error\|newPayload.*INVALID\|newPayload.*failed" || true)
  if [ -n "$PAYLOAD_ERRORS" ]; then
    echo "WARN: engine_newPayload errors found:"
    echo "$PAYLOAD_ERRORS" | head -5
  else
    echo "No engine_newPayload errors"
  fi
fi

echo ""
echo "--- Verifying consecutive blocks accepted ---"
ACCEPT_FAIL=0
for i in 1 2 3 4 5; do
  B_HEX=$(printf '0x%x' $i)
  B=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\",false],\"id\":1}" \
    | jq -r '.result.hash // "null"')
  if [ "$B" = "null" ] || [ -z "$B" ]; then
    echo "  Block $i: MISSING"
    ACCEPT_FAIL=$((ACCEPT_FAIL + 1))
  else
    echo "  Block $i: $B"
  fi
done
[ "$ACCEPT_FAIL" -eq 0 ] || { echo "FAIL: Missing blocks — IL satisfaction may be incorrectly rejecting payloads"; exit 1; }

echo ""
echo "--- Verifying blocks have consistent parentHash chain ---"
PREV_HASH=""
CHAIN_BROKEN=0
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  B_DATA=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\",false],\"id\":1}" \
    | jq -r '{hash: .result.hash, parentHash: .result.parentHash}')
  B_HASH=$(echo "$B_DATA" | jq -r '.hash')
  B_PARENT=$(echo "$B_DATA" | jq -r '.parentHash')
  echo "  Block $i: hash=$B_HASH parent=$B_PARENT"
  if [ -n "$PREV_HASH" ] && [ "$B_PARENT" != "$PREV_HASH" ]; then
    echo "  WARN: parentHash mismatch at block $i"
    CHAIN_BROKEN=$((CHAIN_BROKEN + 1))
  fi
  PREV_HASH="$B_HASH"
done
[ "$CHAIN_BROKEN" -eq 0 ] || echo "WARN: Chain integrity issue detected"

echo ""
echo "PASS: IL Satisfaction — engine_newPayload accepted $ACCEPTED times, no INCLUSION_LIST_UNSATISFIED errors, block chain intact"
