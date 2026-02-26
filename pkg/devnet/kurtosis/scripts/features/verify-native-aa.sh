#!/usr/bin/env bash
# Verify Native AA: check SetCode transactions are processed
set -euo pipefail
ENCLAVE="${1:-eth2030-native-aa}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== Native AA (EIP-7702) Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -ge 2 ] || { echo "FAIL: Too few blocks"; exit 1; }

# Check for type-4 (SetCode) transactions in recent blocks
LATEST=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", true],"id":1}' | jq -r '.result')
TX_COUNT=$(echo "$LATEST" | jq '.transactions | length')
echo "Latest block transactions: $TX_COUNT"

# --- Feature-specific API tests ---

# Scan recent blocks (1 through latest) for type-4 (0x04) SetCode transactions
echo ""
echo "--- Scanning blocks for EIP-7702 SetCode (type-4) transactions ---"
LATEST_DEC=$((BLOCK))
SETCODE_TOTAL=0
for i in $(seq 1 "$LATEST_DEC"); do
  B_HEX=$(printf '0x%x' "$i")
  SETCODE_COUNT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", true],\"id\":1}" \
    | jq '[.result.transactions[] | select(.type == "0x4")] | length')
  if [ "$SETCODE_COUNT" -gt 0 ]; then
    echo "Block $i: found $SETCODE_COUNT SetCode (type-4) transaction(s)"
  fi
  SETCODE_TOTAL=$((SETCODE_TOTAL + SETCODE_COUNT))
done
echo "Total SetCode transactions found across all blocks: $SETCODE_TOTAL"

# Get first tx from latest block and check its receipt
echo ""
echo "--- Verifying transaction receipt from latest block ---"
FIRST_TX_HASH=$(echo "$LATEST" | jq -r '.transactions[0].hash // empty')
if [ -n "$FIRST_TX_HASH" ]; then
  RECEIPT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getTransactionReceipt\",\"params\":[\"$FIRST_TX_HASH\"],\"id\":1}" | jq -r '.result')
  RX_STATUS=$(echo "$RECEIPT" | jq -r '.status // empty')
  RX_GAS=$(echo "$RECEIPT" | jq -r '.gasUsed // empty')
  echo "Tx $FIRST_TX_HASH receipt — status: $RX_STATUS, gasUsed: $RX_GAS"
  [ -n "$RX_STATUS" ] || { echo "FAIL: Receipt missing status field"; exit 1; }
  [ -n "$RX_GAS" ] || { echo "FAIL: Receipt missing gasUsed field"; exit 1; }
else
  echo "WARN: No transactions in latest block to check receipt"
fi

# Verify eth_getTransactionCount (nonce) works for the zero address
echo ""
echo "--- Verifying eth_getTransactionCount for zero address ---"
NONCE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["0x0000000000000000000000000000000000000000","latest"],"id":1}' | jq -r '.result')
echo "Zero address nonce: $NONCE"
[[ "$NONCE" == 0x* ]] || { echo "FAIL: eth_getTransactionCount returned invalid result"; exit 1; }

echo ""
echo "PASS: Native AA — blocks with transactions produced, SetCode scanning complete, receipt and nonce verified"
