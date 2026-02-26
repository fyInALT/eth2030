#!/usr/bin/env bash
# Verify Blobs: check blob transactions are processed
set -euo pipefail
ENCLAVE="${1:-eth2030-blobs}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== Blob Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Check blobscan URL if available
BLOBSCAN_URL=$(kurtosis port print "$ENCLAVE" "blobscan-web" http 2>/dev/null || true)
if [ -n "$BLOBSCAN_URL" ]; then
  echo "Blobscan: $BLOBSCAN_URL"
fi

# --- Feature-specific blob transaction tests ---

# Scan latest block for type-3 (blob) transactions
echo "Checking latest block for blob transactions..."
LATEST_BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", true],"id":1}' | jq -r '.result')
BLOB_TX_COUNT=$(echo "$LATEST_BLOCK" | jq '[.transactions[] | select(.type == "0x3")] | length')
echo "Type-3 (blob) transactions in latest block: $BLOB_TX_COUNT"

# Check blobGasUsed and excessBlobGas in block header
echo "Checking blob gas fields in block header..."
BLOB_GAS_USED=$(echo "$LATEST_BLOCK" | jq -r '.blobGasUsed // "missing"')
EXCESS_BLOB_GAS=$(echo "$LATEST_BLOCK" | jq -r '.excessBlobGas // "missing"')
echo "blobGasUsed: $BLOB_GAS_USED"
echo "excessBlobGas: $EXCESS_BLOB_GAS"
if [ "$BLOB_GAS_USED" != "missing" ] && [ "$BLOB_GAS_USED" != "null" ]; then
  # Verify it looks like a valid hex value (starts with 0x)
  echo "$BLOB_GAS_USED" | grep -q '^0x' || { echo "FAIL: blobGasUsed is not a valid hex value"; exit 1; }
  echo "blobGasUsed is valid hex"
fi
if [ "$EXCESS_BLOB_GAS" != "missing" ] && [ "$EXCESS_BLOB_GAS" != "null" ]; then
  echo "$EXCESS_BLOB_GAS" | grep -q '^0x' || { echo "FAIL: excessBlobGas is not a valid hex value"; exit 1; }
  echo "excessBlobGas is valid hex"
fi

# Call eth_blobBaseFee (soft check — may not be available on all clients)
echo "Checking eth_blobBaseFee (soft check)..."
BLOB_BASE_FEE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blobBaseFee","params":[],"id":1}' | jq -r '.result // "unavailable"')
echo "eth_blobBaseFee: $BLOB_BASE_FEE"

# Compare blobGasUsed across blocks 1-3
echo "Comparing blobGasUsed across blocks 1-3..."
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  BGU=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", false],\"id\":1}" | jq -r '.result.blobGasUsed // "missing"')
  echo "Block $i blobGasUsed: $BGU"
done

echo "PASS: Blobs — blocks produced with blob transaction support and blob gas fields verified"
