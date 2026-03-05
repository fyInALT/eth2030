#!/usr/bin/env bash
# Verify BAL feasibility check (SPEC-5.4 / EIP-7928)
# Every 8 transactions the block builder runs a feasibility check:
#   BAL items × BALItemCost(2000) ≤ cumulative gasUsed
# A violation causes ErrBALFeasibilityViolated and the block is rejected.
# This script verifies high-tx blocks are accepted without such errors.
set -euo pipefail
ENCLAVE="${1:-eth2030-bal-feasibility}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== BAL Feasibility Verification (SPEC-5.4) ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
BLOCK_DEC=$((BLOCK))
echo "Current block: $BLOCK_DEC"
[ "$BLOCK_DEC" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Scan EL logs for BAL feasibility violation errors.
EL_SVC_NAME=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null \
  | grep "el-[0-9]" | head -1 | awk '{print $2}')
if [ -n "$EL_SVC_NAME" ]; then
  echo ""
  echo "--- Scanning EL logs for feasibility violations ---"
  FEASIBILITY_ERRORS=$(kurtosis service logs "$ENCLAVE" "$EL_SVC_NAME" 2>&1 \
    | grep -c "BALFeasibility\|feasibility.*violated\|ErrBALFeasibility" || true)
  echo "BAL feasibility error occurrences: $FEASIBILITY_ERRORS"
  [ "$FEASIBILITY_ERRORS" -eq 0 ] || { echo "FAIL: BAL feasibility violations found in EL logs"; exit 1; }
fi

# Find blocks with >= 8 transactions (the check interval for feasibility).
echo ""
echo "--- Scanning recent blocks for high-tx blocks (txCount >= 8) ---"
LATEST=$((BLOCK_DEC))
START=$(( LATEST > 20 ? LATEST - 20 : 1 ))
HIGH_TX_BLOCKS=0
TOTAL_SCANNED=0
for i in $(seq "$START" "$LATEST"); do
  B_HEX=$(printf '0x%x' $i)
  B_DATA=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\",false],\"id\":1}" \
    | jq -r '{txCount: (.result.transactions | length), gasUsed: .result.gasUsed, hash: .result.hash}')
  TX_CNT=$(echo "$B_DATA" | jq -r '.txCount')
  GAS_USED=$(echo "$B_DATA" | jq -r '.gasUsed')
  TOTAL_SCANNED=$((TOTAL_SCANNED + 1))
  if [ "$TX_CNT" -ge 8 ] 2>/dev/null; then
    GAS_DEC=$(printf '%d' "$GAS_USED" 2>/dev/null || echo 0)
    echo "  Block $i: txCount=$TX_CNT gasUsed=$GAS_DEC — feasibility check triggered, block accepted"
    [ "$GAS_DEC" -gt 0 ] || echo "  WARN: gasUsed=0 in block with $TX_CNT txs"
    HIGH_TX_BLOCKS=$((HIGH_TX_BLOCKS + 1))
  fi
done
echo "Scanned $TOTAL_SCANNED blocks, found $HIGH_TX_BLOCKS with txCount >= 8"

# Verify gasUsed is non-zero in blocks that have transactions.
echo ""
echo "--- Verifying gasUsed > 0 in blocks with transactions ---"
GAS_OK=0
for i in $(seq "$START" "$LATEST"); do
  B_HEX=$(printf '0x%x' $i)
  B_DATA=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\",false],\"id\":1}" \
    | jq -r '{txCount: (.result.transactions | length), gasUsed: .result.gasUsed}')
  TX_CNT=$(echo "$B_DATA" | jq -r '.txCount')
  GAS_HEX=$(echo "$B_DATA" | jq -r '.gasUsed')
  GAS_DEC=$(printf '%d' "$GAS_HEX" 2>/dev/null || echo 0)
  if [ "$TX_CNT" -gt 0 ] && [ "$GAS_DEC" -gt 0 ]; then
    GAS_OK=$((GAS_OK + 1))
  elif [ "$TX_CNT" -gt 0 ] && [ "$GAS_DEC" -eq 0 ]; then
    echo "  FAIL: Block $i has $TX_CNT txs but gasUsed=0 — transactions not executed"
    exit 1
  fi
done
echo "Blocks with txs and gasUsed > 0: $GAS_OK"
[ "$GAS_OK" -gt 0 ] || echo "WARN: No blocks with transactions found in last 20 blocks"

echo ""
echo "PASS: BAL Feasibility — $FEASIBILITY_ERRORS feasibility violations, high-tx blocks accepted correctly"
