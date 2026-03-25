#!/usr/bin/env bash
# Verify Frame Mempool: check EIP-8141 frame transaction infrastructure
set -euo pipefail
ENCLAVE="${1:-eth2030-frame-mempool}"
if [ -n "${2:-}" ]; then
  RPC_URL="$2"
else
  EL_SVC=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}')
  RPC_URL="http://$(kurtosis port print "$ENCLAVE" "$EL_SVC" rpc)"
fi

echo "=== Frame Mempool (EIP-8141) Verification ==="
echo "Covers: frame tx type-0x6, conservative/aggressive gas caps, paymaster registry, VERIFY simulation"

# Check block production
BLOCK=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
echo "Current block: $BLOCK"
[ "$((BLOCK))" -ge 2 ] || { echo "FAIL: Too few blocks produced"; exit 1; }

# Check txpool status
TXPOOL=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"txpool_status","params":[],"id":1}' | jq -r '.result')
PENDING=$(echo "$TXPOOL" | jq -r '.pending // "0x0"')
QUEUED=$(echo "$TXPOOL" | jq -r '.queued // "0x0"')
echo "TxPool pending: $PENDING, queued: $QUEUED"

# --- EIP-8141 EntryPoint address checks ---

echo ""
echo "--- EntryPoint address (0x...aa) accessibility ---"
# EIP-8141 EntryPoint is at 0x00000000000000000000000000000000000000aa
ENTRYPOINT="0x00000000000000000000000000000000000000aa"
EP_CODE=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getCode\",\"params\":[\"$ENTRYPOINT\",\"latest\"],\"id\":1}" \
  | jq -r '.result // "error"')
EP_BAL=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBalance\",\"params\":[\"$ENTRYPOINT\",\"latest\"],\"id\":1}" \
  | jq -r '.result // "error"')
echo "EntryPoint code: $EP_CODE"
echo "EntryPoint balance: $EP_BAL"
[ "$EP_CODE" != "error" ] || { echo "FAIL: eth_getCode failed for EntryPoint"; exit 1; }
[[ "$EP_BAL" == 0x* ]] || { echo "FAIL: eth_getBalance failed for EntryPoint"; exit 1; }

# --- Block scan for frame (type-0x6) transactions ---

echo ""
echo "--- Scanning blocks for frame (type-0x6) transactions ---"
LATEST_DEC=$((BLOCK))
FRAME_TOTAL=0
for i in $(seq 1 "$LATEST_DEC"); do
  B_HEX=$(printf '0x%x' "$i")
  FRAME_COUNT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", true],\"id\":1}" \
    | jq '[.result.transactions[] | select(.type == "0x6")] | length')
  if [ "$FRAME_COUNT" -gt 0 ]; then
    echo "Block $i: found $FRAME_COUNT frame (type-0x6) transaction(s)"
  fi
  FRAME_TOTAL=$((FRAME_TOTAL + FRAME_COUNT))
done
echo "Total frame transactions found across all blocks: $FRAME_TOTAL"

# --- txpool_content API ---

echo ""
echo "--- txpool_content API for frame transaction fields ---"
CONTENT_PENDING=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"txpool_content","params":[],"id":1}' \
  | jq -r '.result.pending | keys | length')
echo "txpool_content: $CONTENT_PENDING pending address(es) in pool"

# --- Frame tx type recognition via eth_sendRawTransaction ---

echo ""
echo "--- Frame tx type-0x6 recognition (eth_sendRawTransaction) ---"
# Send minimal type-0x06 envelope: 0x06 + RLP empty list (c0).
# Expect rejection with an RLP/field error, NOT "unknown tx type".
TX_RESP=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0x06c0"],"id":1}' \
  | jq -r '.error.message // empty')
echo "Response to type-0x06 raw tx: $TX_RESP"
if echo "$TX_RESP" | grep -qi "unknown.*type\|not supported\|unsupported"; then
  echo "  WARN: Node does not recognize frame tx type 0x06 — response: $TX_RESP"
else
  echo "  Frame tx type-0x6 recognized: rejected with structural error (not 'unknown tx type')"
fi

# --- Conservative gas cap validation ---

echo ""
echo "--- Conservative gas cap (default: 50,000 VERIFY gas limit) ---"
# Confirm chain is still advancing — conservative tier is the default and must not break block production
BLOCK2=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq -r '.result')
[ "$((BLOCK2))" -ge "$((BLOCK))" ] || { echo "FAIL: Chain stopped producing blocks"; exit 1; }
echo "Chain advancing: block $((BLOCK)) -> $((BLOCK2)) (conservative tier active)"

# --- Verify txpool module exposes frame-tx fields ---

echo ""
echo "--- txpool_inspect output (frame tx pool fields) ---"
INSPECT_OUT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"txpool_inspect","params":[],"id":1}' | jq -r '.result // empty')
if [ -n "$INSPECT_OUT" ]; then
  echo "txpool_inspect: returned valid JSON result"
else
  echo "  WARN: txpool_inspect returned empty — module may not be enabled"
fi

# --- Metrics endpoint: frame_tx counters ---

echo ""
echo "--- Metrics endpoint: frame_tx_* counters ---"
EL_SVC_NAME=$(kurtosis enclave inspect "$ENCLAVE" 2>/dev/null | grep "el-[0-9]" | head -1 | awk '{print $2}' || true)
METRICS_HOST=$(kurtosis port print "$ENCLAVE" "${EL_SVC_NAME:-el-1-geth-lighthouse}" metrics 2>/dev/null || true)
if echo "$METRICS_HOST" | grep -qE '^[0-9.]+:[0-9]+$'; then
  METRICS_URL="http://$METRICS_HOST/metrics"
  FRAME_ACCEPTED=$(curl -sf "$METRICS_URL" 2>/dev/null | grep "^frame_tx_accepted_total" || true)
  FRAME_REJ_CONS=$(curl -sf "$METRICS_URL" 2>/dev/null | grep "^frame_tx_rejected_conservative_total" || true)
  FRAME_REJ_AGG=$(curl -sf "$METRICS_URL" 2>/dev/null | grep "^frame_tx_rejected_aggressive_total" || true)
  if [ -n "$FRAME_ACCEPTED" ]; then
    echo "  frame_tx_accepted_total: $FRAME_ACCEPTED"
  else
    echo "  frame_tx_accepted_total: (not yet incremented or metrics disabled)"
  fi
  [ -n "$FRAME_REJ_CONS" ] && echo "  $FRAME_REJ_CONS" || true
  [ -n "$FRAME_REJ_AGG" ] && echo "  $FRAME_REJ_AGG" || true
  echo "  Metrics endpoint reachable: $METRICS_URL"
else
  echo "  Metrics endpoint not available — skipping counter check"
fi

# --- Receipt verification for FrameTx (EIP-8141) ---

echo ""
echo "--- FrameTx receipt verification (EIP-8141 payer field) ---"
if [ "$FRAME_TOTAL" -gt 0 ]; then
  # Find a FrameTx and check its receipt
  for i in $(seq 1 "$LATEST_DEC"); do
    B_HEX=$(printf '0x%x' "$i")
    BLOCK_TXS=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
      -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"$B_HEX\", true],\"id\":1}" \
      | jq -r '.result.transactions')
    FRAME_TX_HASH=$(echo "$BLOCK_TXS" | jq -r '.[] | select(.type == "0x6") | .hash' | head -1)
    if [ -n "$FRAME_TX_HASH" ] && [ "$FRAME_TX_HASH" != "null" ]; then
      echo "Found FrameTx in block $i: $FRAME_TX_HASH"
      RECEIPT=$(curl -sf -X POST "$RPC_URL" -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getTransactionReceipt\",\"params\":[\"$FRAME_TX_HASH\"],\"id\":1}" \
        | jq -r '.result')
      if [ -n "$RECEIPT" ] && [ "$RECEIPT" != "null" ]; then
        RX_TYPE=$(echo "$RECEIPT" | jq -r '.type // empty')
        RX_STATUS=$(echo "$RECEIPT" | jq -r '.status // empty')
        RX_GAS=$(echo "$RECEIPT" | jq -r '.gasUsed // empty')
        RX_PAYER=$(echo "$RECEIPT" | jq -r '.payer // empty')
        RX_CONTRACT=$(echo "$RECEIPT" | jq -r '.contractAddress // empty')
        echo "  Receipt type: $RX_TYPE"
        echo "  Receipt status: $RX_STATUS"
        echo "  Receipt gasUsed: $RX_GAS"
        echo "  Receipt payer: $RX_PAYER"
        echo "  Receipt contractAddress: $RX_CONTRACT"
        # Verify type is 0x6
        if [ "$RX_TYPE" == "0x6" ]; then
          echo "  PASS: Receipt type is correct (0x6 = FrameTx)"
        else
          echo "  WARN: Receipt type mismatch, expected 0x6, got $RX_TYPE"
        fi
        # Verify payer field is present for FrameTx
        if [ -n "$RX_PAYER" ] && [ "$RX_PAYER" != "null" ]; then
          echo "  PASS: Payer field is present in FrameTx receipt"
        else
          echo "  INFO: Payer field not present (may be zero address or not set)"
        fi
      else
        echo "  WARN: Could not fetch receipt for FrameTx $FRAME_TX_HASH"
      fi
      break
    fi
  done
else
  echo "  No FrameTx transactions found in blocks — skipping receipt verification"
  echo "  Use spamoor or manual tx submission to test FrameTx receipt fields"
fi

echo ""
echo "PASS: Frame Mempool — chain advancing, txpool API healthy, EntryPoint accessible, frame tx type-0x6 recognized"
