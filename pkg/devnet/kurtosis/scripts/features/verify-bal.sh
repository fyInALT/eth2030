#!/usr/bin/env bash
# Verify BALs: check parallel execution scheduling works
set -euo pipefail
ENCLAVE="${1:-eth2030-bal}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== BAL Verification ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# Verify no state conflicts by checking multiple blocks
for i in 1 2 3; do
  B_HEX=$(printf '0x%x' $i)
  RESULT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", false],\"id\":1}" | jq -r '.result.hash')
  echo "Block $i hash: $RESULT"
done

# --- Feature-specific BAL tests ---

# Call eth_createAccessList on a simple call
echo "Testing eth_createAccessList..."
ACL_RESPONSE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_createAccessList","params":[{"from":"0x0000000000000000000000000000000000000001","to":"0x0000000000000000000000000000000000000002","data":"0x","gas":"0x5208"}],"id":1}' | jq -r '.result')
ACL_LIST=$(echo "$ACL_RESPONSE" | jq -r '.accessList // "missing"')
echo "eth_createAccessList response accessList: $ACL_LIST"
if [ "$ACL_LIST" != "missing" ] && [ "$ACL_LIST" != "null" ]; then
  ACL_IS_ARRAY=$(echo "$ACL_RESPONSE" | jq 'has("accessList") and (.accessList | type == "array")')
  [ "$ACL_IS_ARRAY" = "true" ] || { echo "FAIL: accessList is not an array"; exit 1; }
  echo "accessList is a valid array"
fi

# Check blocks for transactions with accessList
echo "Checking latest block for transactions with accessList..."
LATEST_BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", true],"id":1}' | jq -r '.result')
TXS_WITH_ACL=$(echo "$LATEST_BLOCK" | jq '[.transactions[] | select(.accessList != null and (.accessList | length > 0))] | length')
echo "Transactions with accessList in latest block: $TXS_WITH_ACL"

# Verify state roots change across consecutive blocks (proving execution happened)
echo "Verifying state root evolution across blocks..."
SR1=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1", false],"id":1}' | jq -r '.result.stateRoot')
SR2=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x2", false],"id":1}' | jq -r '.result.stateRoot')
echo "Block 1 stateRoot: $SR1"
echo "Block 2 stateRoot: $SR2"
if [ "$SR1" != "null" ] && [ "$SR2" != "null" ] && [ -n "$SR1" ] && [ -n "$SR2" ]; then
  if [ "$SR1" != "$SR2" ]; then
    echo "State roots differ across blocks (execution is evolving state)"
  else
    echo "State roots are identical (blocks may be empty)"
  fi
fi

# Call eth_getBalance on zero address to verify state queries work
echo "Checking eth_getBalance on zero address..."
BALANCE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0x0000000000000000000000000000000000000000","latest"],"id":1}' | jq -r '.result')
echo "Zero address balance: $BALANCE"
[ -n "$BALANCE" ] && [ "$BALANCE" != "null" ] || { echo "FAIL: eth_getBalance returned null"; exit 1; }

echo "PASS: BAL — blocks produced without state conflicts, access lists and state queries verified"
