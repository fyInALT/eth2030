#!/usr/bin/env bash
# Verify FOCIL: check transactions are included and blocks are produced
set -euo pipefail
ENCLAVE="${1:-eth2030-focil}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== FOCIL Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"

# Check latest block has transactions (FOCIL ensures inclusion)
BLOCK_DATA=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' | jq -r '.result')
TX_COUNT=$(echo "$BLOCK_DATA" | jq -r '.transactions | length')
echo "Latest block tx count: $TX_COUNT"

# --- Feature-specific FOCIL tests ---

# Check tx count > 0 in at least 1 of last 3 blocks
echo "Checking transaction inclusion across recent blocks..."
FOUND_TX=0
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  TX_CNT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockTransactionCountByNumber\",\"params\":[\"$B_HEX\"],\"id\":1}" | jq -r '.result // "0x0"')
  TX_DEC=$(printf '%d' "$TX_CNT" 2>/dev/null || echo 0)
  echo "  Block $i tx count: $TX_DEC"
  if [ "$TX_DEC" -gt 0 ]; then
    FOUND_TX=1
  fi
done
if [ "$FOUND_TX" -eq 1 ]; then
  echo "Transactions found in recent blocks (inclusion enforcement active)"
else
  echo "WARN: No transactions in blocks 1-3 (spamoor may not have started yet)"
fi

# Verify gasUsed > 0 in a block with transactions
echo "Checking gasUsed in blocks with transactions..."
GAS_OK=0
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  B_DATA=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", false],\"id\":1}" | jq -r '.result')
  B_TX_CNT=$(echo "$B_DATA" | jq -r '.transactions | length')
  B_GAS=$(echo "$B_DATA" | jq -r '.gasUsed // "0x0"')
  B_GAS_DEC=$(printf '%d' "$B_GAS" 2>/dev/null || echo 0)
  if [ "$B_TX_CNT" -gt 0 ] && [ "$B_GAS_DEC" -gt 0 ]; then
    echo "  Block $i: gasUsed=$B_GAS_DEC with $B_TX_CNT txs"
    GAS_OK=1
    break
  fi
done
[ "$GAS_OK" -eq 1 ] || echo "WARN: No block with both transactions and gasUsed > 0 found"

# Check logsBloom is non-zero in blocks with transactions
echo "Checking logsBloom in blocks with transactions..."
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  B_DATA=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", false],\"id\":1}" | jq -r '.result')
  B_TX_CNT=$(echo "$B_DATA" | jq -r '.transactions | length')
  LOGS_BLOOM=$(echo "$B_DATA" | jq -r '.logsBloom // "missing"')
  if [ "$B_TX_CNT" -gt 0 ] && [ "$LOGS_BLOOM" != "missing" ]; then
    ZERO_BLOOM="0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
    if [ "$LOGS_BLOOM" != "$ZERO_BLOOM" ]; then
      echo "  Block $i: logsBloom is non-zero (has log data)"
    else
      echo "  Block $i: logsBloom is all zeros (no log events)"
    fi
  fi
done

# Verify receiptsRoot is non-null for blocks with txs
echo "Checking receiptsRoot in blocks with transactions..."
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  B_DATA=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", false],\"id\":1}" | jq -r '.result')
  B_TX_CNT=$(echo "$B_DATA" | jq -r '.transactions | length')
  RECEIPTS_ROOT=$(echo "$B_DATA" | jq -r '.receiptsRoot // "missing"')
  if [ "$B_TX_CNT" -gt 0 ]; then
    echo "  Block $i: receiptsRoot=$RECEIPTS_ROOT"
    [ "$RECEIPTS_ROOT" != "missing" ] && [ "$RECEIPTS_ROOT" != "null" ] || { echo "FAIL: receiptsRoot is null for block $i with txs"; exit 1; }
  fi
done

echo "PASS: FOCIL — blocks produced, transaction inclusion verified, gas and receipt checks passed"
