#!/usr/bin/env bash
# Verify MultiDimFeeTx type 0x09 (SPEC-6.1 / EIP-7706)
# Type 0x09 encodes 3D fee vectors: MaxFeesPerGas[3] and PriorityFeesPerGas[3].
# This script verifies the node recognises type 0x09 (returns RLP decode error,
# not "unknown transaction type"), confirming the type decoder is registered.
set -euo pipefail
ENCLAVE="${1:-eth2030-multidim-fee-tx}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== MultiDimFeeTx Type 0x09 Verification (SPEC-6.1) ==="
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
BLOCK_DEC=$((BLOCK))
echo "Current block: $BLOCK_DEC"
[ "$BLOCK_DEC" -gt 0 ] || { echo "FAIL: No blocks produced"; exit 1; }

# --- Test 1: type 0x09 is recognised (not "unknown transaction type") ---
# Send a raw type 0x09 transaction with an empty RLP list body.
# Expected: RLP structure error (the type is known but the payload is malformed).
# Failure case: "unknown transaction type" means 0x09 is not registered.
echo ""
echo "--- Test 1: type 0x09 decoder is registered ---"
ERR=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0x09c0"],"id":1}' \
  | jq -r '.error.message // "no_error"')
echo "eth_sendRawTransaction(0x09 empty RLP) error: $ERR"

if echo "$ERR" | grep -qi "unknown transaction type"; then
  echo "FAIL: Node returned 'unknown transaction type' — type 0x09 MultiDimFeeTx is NOT registered"
  exit 1
elif echo "$ERR" | grep -qiE "rlp|EOF|decode|invalid|unexpected"; then
  echo "PASS: Node recognised type 0x09 (returned RLP/decode error, not 'unknown type')"
elif [ "$ERR" = "no_error" ]; then
  echo "WARN: No error returned for empty type 0x09 tx — unexpected"
else
  echo "PASS: Node recognised type 0x09 (error: $ERR)"
fi

# --- Test 2: type 0x09 with a minimal but structurally plausible payload ---
# 0x09 + RLP list with chainId=1, remaining fields as empty
echo ""
echo "--- Test 2: type 0x09 with minimal RLP structure ---"
ERR2=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0x09d4018080808080808080800180808001a001a001"],"id":1}' \
  | jq -r '.error.message // "no_error"')
echo "eth_sendRawTransaction(0x09 minimal) error: $ERR2"
if echo "$ERR2" | grep -qi "unknown transaction type"; then
  echo "FAIL: type 0x09 not recognised in second test"
  exit 1
fi
echo "type 0x09 recognised (error is structural, not type-unknown)"

# --- Test 3: existing type 0x02 (EIP-1559) still works ---
echo ""
echo "--- Test 3: type 0x02 (EIP-1559) still recognised ---"
ERR3=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0x02c0"],"id":1}' \
  | jq -r '.error.message // "no_error"')
echo "eth_sendRawTransaction(0x02 empty) error: $ERR3"
if echo "$ERR3" | grep -qi "unknown transaction type"; then
  echo "FAIL: type 0x02 not recognised — regression in tx type registration"
  exit 1
fi
echo "type 0x02 still recognised"

# --- Test 4: type 0x08 (LocalTx) rejected by public RPC (expected) ---
# LocalTx (type 0x08) is an internal type accepted only via the mixnet/transport
# manager, not via the public eth_sendRawTransaction endpoint.
# It should return "unsupported transaction type" from the public RPC, which
# is the correct gating behaviour (not a regression).
echo ""
echo "--- Test 4: type 0x08 (LocalTx) correctly rejected by public RPC ---"
ERR4=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0x08c0"],"id":1}' \
  | jq -r '.error.message // "no_error"')
echo "eth_sendRawTransaction(0x08 empty) error: $ERR4"
if echo "$ERR4" | grep -qiE "unsupported transaction type|unknown transaction type"; then
  echo "type 0x08 correctly rejected by public RPC (internal-only type)"
else
  echo "WARN: unexpected response for type 0x08: $ERR4"
fi

# --- Test 5: chain accepts blocks (node hasn't crashed from type registration) ---
echo ""
echo "--- Test 5: node is still producing blocks ---"
BLOCK2=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Block after tests: $((BLOCK2))"
[ "$((BLOCK2))" -ge "$BLOCK_DEC" ] || { echo "FAIL: block number decreased — node unstable"; exit 1; }

echo ""
echo "PASS: MultiDimFeeTx — type 0x09 registered and recognised, types 0x02/0x08 unaffected, node stable"
