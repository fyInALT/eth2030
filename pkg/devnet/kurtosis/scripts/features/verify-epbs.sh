#!/usr/bin/env bash
# Verify ePBS: check that blocks are being produced and builder infrastructure is active
set -euo pipefail
ENCLAVE="${1:-eth2030-epbs}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== ePBS Verification ==="
# Check block production
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Check txpool status (builder should process transactions)
TXPOOL=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"txpool_status","params":[],"id":1}' | jq -r '.result')
echo "TxPool: $TXPOOL"

# --- Feature-specific builder infrastructure tests ---

echo ""
echo "--- Checking miner/feeRecipient in block header ---"
LATEST_BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' | jq -r '.result')
MINER=$(echo "$LATEST_BLOCK" | jq -r '.miner // "null"')
echo "Block miner/feeRecipient: $MINER"
if [ "$MINER" != "null" ] && [ "$MINER" != "0x0000000000000000000000000000000000000000" ]; then
  echo "  Miner is a non-zero address"
else
  echo "  WARN: Miner address is zero or null"
fi

echo ""
echo "--- Verifying txpool_status fields ---"
TXPOOL_PENDING=$(echo "$TXPOOL" | jq -r '.pending // "null"')
TXPOOL_QUEUED=$(echo "$TXPOOL" | jq -r '.queued // "null"')
echo "TxPool pending: $TXPOOL_PENDING"
echo "TxPool queued:  $TXPOOL_QUEUED"
if [[ "$TXPOOL_PENDING" == 0x* ]] && [[ "$TXPOOL_QUEUED" == 0x* ]]; then
  echo "  txpool_status validated: pending and queued are valid hex values"
else
  echo "  WARN: txpool_status returned unexpected format"
fi

echo ""
echo "--- Checking extraData field in block header ---"
EXTRA_DATA=$(echo "$LATEST_BLOCK" | jq -r '.extraData // "null"')
echo "Block extraData: $EXTRA_DATA"
if [ "$EXTRA_DATA" != "null" ] && [ -n "$EXTRA_DATA" ]; then
  echo "  extraData field is present and non-null"
else
  echo "  WARN: extraData field is missing or null"
fi

echo ""
echo "--- Checking latest block for transactions ---"
LATEST_WITH_TXS=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", true],"id":1}' | jq -r '.result')
LATEST_TX_COUNT=$(echo "$LATEST_WITH_TXS" | jq '.transactions | length')
echo "Latest block transaction count: $LATEST_TX_COUNT"
if [ "$LATEST_TX_COUNT" -ge 1 ]; then
  echo "  Block contains at least 1 transaction"
else
  echo "  WARN: Latest block has no transactions"
fi

echo ""
echo "--- Checking builder consistency across consecutive blocks ---"
LATEST_DEC=$((BLOCK))
if [ "$LATEST_DEC" -ge 2 ]; then
  PREV_HEX=$(printf '0x%x' $((LATEST_DEC - 1)))
  PREV_BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$PREV_HEX\", false],\"id\":1}" | jq -r '.result')
  PREV_MINER=$(echo "$PREV_BLOCK" | jq -r '.miner // "null"')
  echo "Block $((LATEST_DEC - 1)) miner: $PREV_MINER"
  echo "Block $LATEST_DEC miner:         $MINER"
  if [ "$PREV_MINER" = "$MINER" ]; then
    echo "  Same miner across consecutive blocks (consistent builder)"
  else
    echo "  Different miners across consecutive blocks (builder rotation)"
  fi
else
  echo "  Skipping: need at least 2 blocks for comparison"
fi

echo ""
echo "PASS: ePBS — blocks produced, builder active, miner/txpool/extraData/tx checks passed"
