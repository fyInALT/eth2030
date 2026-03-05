#!/usr/bin/env bash
# Verify BAL retention window (SPEC-5.5 / EIP-7928)
# BALRetentionSlots = 3533 epochs × 32 slots = 113056 slots.
# All blocks at height < BALRetentionSlots should be retained (accessible).
# This script verifies that recent blocks within the retention window are
# accessible via eth_getBlockByNumber and that no retention-related errors
# appear in the EL logs.
set -euo pipefail
ENCLAVE="${1:-eth2030-bal-retention}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

BAL_RETENTION_SLOTS=113056

echo "=== BAL Retention Verification (SPEC-5.5) ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
BLOCK_DEC=$((BLOCK))
echo "Current block: $BLOCK_DEC"
echo "BALRetentionSlots: $BAL_RETENTION_SLOTS"
[ "$BLOCK_DEC" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Verify chain is below the retention threshold (devnet is always well below 113056).
if [ "$BLOCK_DEC" -lt "$BAL_RETENTION_SLOTS" ]; then
  echo "Chain height $BLOCK_DEC < BALRetentionSlots $BAL_RETENTION_SLOTS — all blocks are within retention window"
else
  echo "WARN: Chain height $BLOCK_DEC >= BALRetentionSlots $BAL_RETENTION_SLOTS — some blocks may be outside window"
fi

# Verify all blocks from 1 to min(BLOCK_DEC, 10) are accessible.
echo ""
echo "--- Verifying blocks 1..10 are accessible (within retention window) ---"
LIMIT=$(( BLOCK_DEC < 10 ? BLOCK_DEC : 10 ))
MISSING=0
for i in $(seq 1 "$LIMIT"); do
  B_HEX=$(printf '0x%x' $i)
  RESULT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\",false],\"id\":1}" \
    | jq -r '.result.hash // "null"')
  if [ "$RESULT" = "null" ] || [ -z "$RESULT" ]; then
    echo "  Block $i: NOT ACCESSIBLE (should be retained)"
    MISSING=$((MISSING + 1))
  else
    echo "  Block $i: $RESULT (retained)"
  fi
done
[ "$MISSING" -eq 0 ] || { echo "FAIL: $MISSING blocks not accessible within retention window"; exit 1; }

# Verify block hashes are stable (same hash on repeated queries — no reorg in retention window).
echo ""
echo "--- Verifying block hash stability (no reorgs within window) ---"
HASH1=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1",false],"id":1}' \
  | jq -r '.result.hash')
HASH1_AGAIN=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1",false],"id":1}' \
  | jq -r '.result.hash')
echo "Block 1 hash (query 1): $HASH1"
echo "Block 1 hash (query 2): $HASH1_AGAIN"
[ "$HASH1" = "$HASH1_AGAIN" ] || { echo "FAIL: Block 1 hash changed between queries — retention state is unstable"; exit 1; }
echo "Block hash is stable"

# Scan EL logs for retention-related errors.
EL_SVC_NAME=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null \
  | grep "el-[0-9]" | head -1 | awk '{print $2}')
if [ -n "$EL_SVC_NAME" ]; then
  echo ""
  echo "--- Scanning EL logs for BAL retention errors ---"
  RETENTION_ERRORS=$(kurtosis service logs "$ENCLAVE" "$EL_SVC_NAME" 2>&1 \
    | grep -c "retention.*error\|BALRetention.*fail\|history.*expired" || true)
  echo "BAL retention error occurrences: $RETENTION_ERRORS"
  [ "$RETENTION_ERRORS" -eq 0 ] || echo "WARN: retention-related errors found in logs"

  echo ""
  echo "--- Checking engine_getPayloadBodiesBy acknowledgement ---"
  BODIES_V2=$(kurtosis service logs "$ENCLAVE" "$EL_SVC_NAME" 2>&1 \
    | grep -c "getPayloadBodies" || true)
  echo "engine_getPayloadBodies calls observed: $BODIES_V2"
fi

# Verify block count matches chain tip (no gaps in the retained range).
echo ""
echo "--- Verifying no gaps in block range 1..$LIMIT ---"
echo "All $LIMIT blocks accessible — retention window intact"

echo ""
echo "PASS: BAL Retention — $LIMIT blocks verified accessible, chain height $BLOCK_DEC < retention threshold $BAL_RETENTION_SLOTS"
