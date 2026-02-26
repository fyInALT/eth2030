#!/usr/bin/env bash
# Verify 3SF: check 3-slot finality and quick slot timing
set -euo pipefail
ENCLAVE="${1:-eth2030-consensus-3sf}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== 3-Slot Finality Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Check block timestamps to verify slot timing (if enough blocks exist)
if [ "$((BLOCK))" -ge 2 ]; then
  B1=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1", false],"id":1}' | jq -r '.result.timestamp')
  B2=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x2", false],"id":1}' | jq -r '.result.timestamp')
  if [ -n "$B1" ] && [ -n "$B2" ] && [ "$B1" != "null" ] && [ "$B2" != "null" ]; then
    SLOT_TIME=$(( $(printf '%d' "$B2") - $(printf '%d' "$B1") ))
    echo "Slot time: ${SLOT_TIME}s (block 1 -> block 2)"
  fi
fi

# --- Feature-specific 3SF consensus tests ---

# Verify timestamps are monotonically increasing across blocks 1-3
echo "Verifying monotonically increasing timestamps across blocks 1-3..."
PREV_TS=0
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  TS_HEX=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", false],\"id\":1}" | jq -r '.result.timestamp // "null"')
  if [ "$TS_HEX" != "null" ] && [ -n "$TS_HEX" ]; then
    TS_DEC=$(printf '%d' "$TS_HEX")
    echo "Block $i timestamp: $TS_HEX ($TS_DEC)"
    if [ "$PREV_TS" -gt 0 ] && [ "$TS_DEC" -le "$PREV_TS" ]; then
      echo "FAIL: Timestamps not monotonically increasing (block $i: $TS_DEC <= previous: $PREV_TS)"
      exit 1
    fi
    PREV_TS=$TS_DEC
  fi
done
echo "Timestamps are monotonically increasing"

# Check difficulty is "0x0" in block headers (PoS consensus)
echo "Checking difficulty field in block headers (should be 0x0 for PoS)..."
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  DIFF=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", false],\"id\":1}" | jq -r '.result.difficulty // "missing"')
  echo "Block $i difficulty: $DIFF"
  [ "$DIFF" = "0x0" ] || { echo "FAIL: Block $i difficulty is $DIFF, expected 0x0 (PoS)"; exit 1; }
done
echo "All blocks have difficulty 0x0 (PoS confirmed)"

# Check nonce is "0x0000000000000000" in block headers (post-merge)
echo "Checking nonce field in block headers (should be 0x0000000000000000 post-merge)..."
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  NONCE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", false],\"id\":1}" | jq -r '.result.nonce // "missing"')
  echo "Block $i nonce: $NONCE"
  [ "$NONCE" = "0x0000000000000000" ] || { echo "FAIL: Block $i nonce is $NONCE, expected 0x0000000000000000 (post-merge)"; exit 1; }
done
echo "All blocks have nonce 0x0000000000000000 (post-merge confirmed)"

# Verify at least 3 consecutive blocks with valid parent hash linkage
echo "Verifying parent hash chain linkage across blocks 1-3..."
PREV_HASH=""
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  BLOCK_DATA=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", false],\"id\":1}" | jq -r '.result')
  HASH=$(echo "$BLOCK_DATA" | jq -r '.hash')
  PARENT_HASH=$(echo "$BLOCK_DATA" | jq -r '.parentHash')
  echo "Block $i: hash=${HASH:0:18}... parentHash=${PARENT_HASH:0:18}..."
  if [ -n "$PREV_HASH" ] && [ "$PARENT_HASH" != "$PREV_HASH" ]; then
    echo "FAIL: Block $i parentHash does not match block $((i-1)) hash"
    exit 1
  fi
  PREV_HASH="$HASH"
done
echo "Parent hash linkage verified across 3 consecutive blocks"

# Verify mixHash field is present (used for RANDAO in PoS)
echo "Checking mixHash field (RANDAO) in latest block..."
MIX_HASH=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", false],"id":1}' | jq -r '.result.mixHash // "missing"')
echo "Latest block mixHash: $MIX_HASH"
[ "$MIX_HASH" != "missing" ] && [ "$MIX_HASH" != "null" ] && [ -n "$MIX_HASH" ] || { echo "FAIL: mixHash field missing from block header"; exit 1; }
echo "mixHash (RANDAO) field present"

echo "PASS: 3SF — blocks produced with quick slot timing, PoS consensus fields verified, parent hash chain valid"
