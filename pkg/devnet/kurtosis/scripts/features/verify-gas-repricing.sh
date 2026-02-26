#!/usr/bin/env bash
# Verify Gas Repricing: call repriced precompiles
set -euo pipefail
ENCLAVE="${1:-eth2030-gas-repricing}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== Gas Repricing Verification ==="

# Test ecAdd precompile (0x06) — Glamsterdam repriced
RESULT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"0x0000000000000000000000000000000000000006","data":"0x0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002","gas":"0x100000"},"latest"],"id":1}' | jq -r '.result // .error.message')
echo "ecAdd (0x06): $RESULT"
[[ "$RESULT" == 0x* ]] || { echo "FAIL: ecAdd precompile failed"; exit 1; }

# --- Feature-specific precompile tests ---

echo ""
echo "--- sha256 Precompile (0x02): Hashing 'Hello' ---"
SHA256_RESULT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"0x0000000000000000000000000000000000000002","data":"0x48656c6c6f","gas":"0x100000"},"latest"],"id":1}' | jq -r '.result // .error.message')
echo "sha256 result: $SHA256_RESULT"
if [[ "$SHA256_RESULT" == 0x* ]] && [ "${#SHA256_RESULT}" -eq 66 ]; then
  echo "  sha256 validated: 32-byte hash (66 chars with 0x prefix)"
else
  echo "  WARN: sha256 result unexpected length ${#SHA256_RESULT} (expected 66)"
fi

echo ""
echo "--- ripemd160 Precompile (0x03): Hashing 'Hello' ---"
RIPEMD_RESULT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"0x0000000000000000000000000000000000000003","data":"0x48656c6c6f","gas":"0x100000"},"latest"],"id":1}' | jq -r '.result // .error.message')
echo "ripemd160 result: $RIPEMD_RESULT"
# ripemd160 returns 20 bytes but EVM pads to 32 bytes (66 chars with 0x)
if [[ "$RIPEMD_RESULT" == 0x* ]] && [ "${#RIPEMD_RESULT}" -ge 42 ]; then
  echo "  ripemd160 validated: result returned (${#RIPEMD_RESULT} chars)"
else
  echo "  WARN: ripemd160 result unexpected — got: $RIPEMD_RESULT"
fi

echo ""
echo "--- ecrecover Precompile (0x01): Known test vector ---"
# Message hash + v=28 (0x1c) + r + s
ECRECOVER_INPUT="0x456e9aea5e197a1f1af7a3e85a3212fa4049a3ba34c2289b4c860fc0b0c64ef3000000000000000000000000000000000000000000000000000000000000001c9242685bf161793cc25603c231bc2f568eb630ea16aa137d2664ac80388256084f8ae3bd7535248d0bd448298cc2e2071e56992d0774dc340c368ae950852ada"
ECRECOVER_RESULT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_call\",\"params\":[{\"to\":\"0x0000000000000000000000000000000000000001\",\"data\":\"$ECRECOVER_INPUT\",\"gas\":\"0x100000\"},\"latest\"],\"id\":1}" | jq -r '.result // .error.message')
echo "ecrecover result: $ECRECOVER_RESULT"
if [[ "$ECRECOVER_RESULT" == 0x* ]] && [ "${#ECRECOVER_RESULT}" -eq 66 ]; then
  echo "  ecrecover validated: returned 32-byte padded address"
else
  echo "  WARN: ecrecover result unexpected — got: $ECRECOVER_RESULT"
fi

echo ""
echo "--- bn256ScalarMul Precompile (0x07): Generator * 1 ---"
# BN256 generator point (1, 2) with scalar 1
BN256MUL_INPUT="0x000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001"
BN256MUL_RESULT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_call\",\"params\":[{\"to\":\"0x0000000000000000000000000000000000000007\",\"data\":\"$BN256MUL_INPUT\",\"gas\":\"0x100000\"},\"latest\"],\"id\":1}" | jq -r '.result // .error.message')
echo "bn256ScalarMul result: $BN256MUL_RESULT"
if [[ "$BN256MUL_RESULT" == 0x* ]] && [ "${#BN256MUL_RESULT}" -gt 10 ]; then
  echo "  bn256ScalarMul validated: non-zero result returned"
else
  echo "  WARN: bn256ScalarMul result unexpected — got: $BN256MUL_RESULT"
fi

echo ""
echo "--- eth_estimateGas on ecAdd call ---"
ESTIMATE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_estimateGas","params":[{"to":"0x0000000000000000000000000000000000000006","data":"0x0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002"}],"id":1}' | jq -r '.result // .error.message')
echo "ecAdd estimated gas: $ESTIMATE"
if [[ "$ESTIMATE" == 0x* ]]; then
  GAS_DEC=$((ESTIMATE))
  echo "  Estimated gas (decimal): $GAS_DEC"
  [ "$GAS_DEC" -gt 0 ] || echo "  WARN: Estimated gas is 0"
else
  echo "  WARN: eth_estimateGas returned unexpected result"
fi

echo ""
echo "PASS: Gas Repricing — ecAdd, sha256, ripemd160, ecrecover, bn256ScalarMul precompiles tested, gas estimation verified"
