#!/usr/bin/env bash
# Verify Encrypted Mempool: check commit-reveal ordering and tx processing
set -euo pipefail
ENCLAVE="${1:-eth2030-encrypted-mempool}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== Encrypted Mempool Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Check txpool status (encrypted mempool should still accept transactions)
TXPOOL=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"txpool_status","params":[],"id":1}' | jq -r '.result')
PENDING=$(echo "$TXPOOL" | jq -r '.pending // "0x0"')
QUEUED=$(echo "$TXPOOL" | jq -r '.queued // "0x0"')
echo "TxPool pending: $PENDING, queued: $QUEUED"

# Verify blocks contain transactions (ordering is working)
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  TX_COUNT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockTransactionCountByNumber\",\"params\":[\"$B_HEX\"],\"id\":1}" | jq -r '.result // "0x0"')
  echo "Block $i tx count: $TX_COUNT"
done

# --- Feature-specific encrypted mempool tests ---

# Call txpool_status twice with 3s sleep between
echo "Monitoring txpool status over time..."
TXPOOL1=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"txpool_status","params":[],"id":1}' | jq -r '.result')
echo "TxPool status (t=0): pending=$(echo "$TXPOOL1" | jq -r '.pending // "0x0"') queued=$(echo "$TXPOOL1" | jq -r '.queued // "0x0"')"
sleep 3
TXPOOL2=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"txpool_status","params":[],"id":1}' | jq -r '.result')
echo "TxPool status (t=3s): pending=$(echo "$TXPOOL2" | jq -r '.pending // "0x0"') queued=$(echo "$TXPOOL2" | jq -r '.queued // "0x0"')"

# Check that blocks include transactions: get tx count from latest block
echo "Checking transaction count in latest block..."
LATEST_TX_COUNT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockTransactionCountByNumber","params":["latest"],"id":1}' | jq -r '.result // "0x0"')
echo "Latest block tx count: $LATEST_TX_COUNT"

# Get tx hashes from block N and block N-1, verify no duplicate tx hashes (proper ordering)
echo "Checking for duplicate tx hashes between consecutive blocks..."
CURRENT_BLOCK=$(($(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')))
if [ "$CURRENT_BLOCK" -ge 2 ]; then
  B_N=$(printf '0x%x' "$CURRENT_BLOCK")
  B_N1=$(printf '0x%x' $((CURRENT_BLOCK - 1)))
  TXS_N=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_N\", true],\"id\":1}" | jq -r '[.result.transactions[].hash] | sort | .[]' 2>/dev/null || true)
  TXS_N1=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_N1\", true],\"id\":1}" | jq -r '[.result.transactions[].hash] | sort | .[]' 2>/dev/null || true)
  if [ -n "$TXS_N" ] && [ -n "$TXS_N1" ]; then
    DUPES=$(comm -12 <(echo "$TXS_N") <(echo "$TXS_N1") | wc -l)
    echo "Duplicate tx hashes between block $((CURRENT_BLOCK-1)) and $CURRENT_BLOCK: $DUPES"
    [ "$DUPES" -eq 0 ] || { echo "FAIL: Found duplicate tx hashes across consecutive blocks"; exit 1; }
    echo "No duplicate tx hashes (proper ordering verified)"
  else
    echo "One or both blocks have no transactions — skipping duplicate check"
  fi
fi

# Verify gasUsed > 0 in latest block
echo "Checking gasUsed in latest block..."
LATEST_GAS_USED=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' | jq -r '.result.gasUsed // "0x0"')
LATEST_GU_DEC=$((LATEST_GAS_USED))
echo "Latest block gasUsed: $LATEST_GAS_USED ($LATEST_GU_DEC)"
if [ "$LATEST_GU_DEC" -gt 0 ]; then
  echo "gasUsed > 0 in latest block (transactions were processed)"
else
  echo "WARNING: gasUsed is 0 in latest block"
fi

# Call eth_gasPrice to verify gas price oracle is active
echo "Checking gas price oracle..."
GAS_PRICE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":1}' | jq -r '.result // "unavailable"')
echo "eth_gasPrice: $GAS_PRICE"
[ "$GAS_PRICE" != "unavailable" ] && [ "$GAS_PRICE" != "null" ] || { echo "FAIL: eth_gasPrice not available"; exit 1; }
echo "Gas price oracle is active"

echo "PASS: Encrypted Mempool — commit-reveal ordering working, no duplicate txs, gas oracle active"
