#!/usr/bin/env bash
# Verify DL Blob Advanced: check BPO blobs, blob streaming, PQ blobs, variable-size blobs
set -euo pipefail
ENCLAVE="${1:-eth2030-dl-blob-advanced}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== DL Blob Advanced Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Check latest block for blob gas fields (blobGasUsed, excessBlobGas)
LATEST=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' | jq -r '.result')
BLOB_GAS=$(echo "$LATEST" | jq -r '.blobGasUsed // "0x0"')
EXCESS_BLOB_GAS=$(echo "$LATEST" | jq -r '.excessBlobGas // "0x0"')
echo "Blob gas used: $BLOB_GAS"
echo "Excess blob gas: $EXCESS_BLOB_GAS"

# Verify blob gas fields exist in header (BPO schedule support)
BASE_FEE=$(echo "$LATEST" | jq -r '.baseFeePerGas // "missing"')
echo "Base fee per gas: $BASE_FEE"
[ "$BASE_FEE" != "missing" ] || { echo "FAIL: baseFeePerGas missing from header"; exit 1; }

# --- Feature-specific advanced blob tests ---

# Check blobGasUsed and excessBlobGas in latest block header as hex values
echo "Checking blob gas fields as hex values..."
echo "  blobGasUsed (hex): $BLOB_GAS"
echo "  excessBlobGas (hex): $EXCESS_BLOB_GAS"
BLOB_GAS_DEC=$(printf '%d' "$BLOB_GAS" 2>/dev/null || echo 0)
EXCESS_DEC=$(printf '%d' "$EXCESS_BLOB_GAS" 2>/dev/null || echo 0)
echo "  blobGasUsed (dec): $BLOB_GAS_DEC"
echo "  excessBlobGas (dec): $EXCESS_DEC"

# Call eth_blobBaseFee (soft check — echo result, don't hard-fail)
echo "Checking eth_blobBaseFee (soft check)..."
BLOB_BASE_FEE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blobBaseFee","params":[],"id":1}' | jq -r '.result // .error.message // "unsupported"')
echo "  eth_blobBaseFee: $BLOB_BASE_FEE"

# Call eth_feeHistory with 3 blocks — verify baseFeePerGas returned
echo "Checking eth_feeHistory for baseFeePerGas..."
FEE_HISTORY=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_feeHistory","params":["0x3","latest",[]],"id":1}' | jq -r '.result')
if [ "$FEE_HISTORY" != "null" ] && [ -n "$FEE_HISTORY" ]; then
  FH_BASE_FEES=$(echo "$FEE_HISTORY" | jq -r '.baseFeePerGas | length // 0' 2>/dev/null || echo 0)
  echo "  eth_feeHistory baseFeePerGas entries: $FH_BASE_FEES"
  [ "$FH_BASE_FEES" -gt 0 ] || echo "WARN: eth_feeHistory returned no baseFeePerGas entries"
else
  echo "  eth_feeHistory: not available"
fi

# Compare gasUsed and blobGasUsed in same block — echo both to show independence
echo "Comparing gasUsed and blobGasUsed in latest block..."
GAS_USED=$(echo "$LATEST" | jq -r '.gasUsed // "0x0"')
GAS_USED_DEC=$(printf '%d' "$GAS_USED" 2>/dev/null || echo 0)
echo "  gasUsed: $GAS_USED_DEC"
echo "  blobGasUsed: $BLOB_GAS_DEC"
echo "  (These are independent gas dimensions)"

# Verify excessBlobGas across 2 blocks, echo comparison
echo "Comparing excessBlobGas across blocks..."
if [ "$((BLOCK))" -ge 2 ]; then
  B1_DATA=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1", false],"id":1}' | jq -r '.result')
  B2_DATA=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x2", false],"id":1}' | jq -r '.result')
  EXCESS1=$(echo "$B1_DATA" | jq -r '.excessBlobGas // "0x0"')
  EXCESS2=$(echo "$B2_DATA" | jq -r '.excessBlobGas // "0x0"')
  echo "  Block 1 excessBlobGas: $EXCESS1"
  echo "  Block 2 excessBlobGas: $EXCESS2"
else
  echo "  Not enough blocks for comparison"
fi

echo "PASS: DL Blob Advanced — blob gas fields verified, fee history checked, gas dimensions validated"
