#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONFIG_DIR="$SCRIPT_DIR/../configs"
PKG_DIR="$SCRIPT_DIR/../../../"

CONFIG="${1:-geth}"
CONFIG_FILE="$CONFIG_DIR/geth.yaml"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Config not found: $CONFIG_FILE"
    echo "Available configs:"
    ls "$CONFIG_DIR"/*.yaml 2>/dev/null | xargs -n1 basename | sed 's/.yaml//' | sed 's/^/  /'
    exit 1
fi

if ! command -v kurtosis &>/dev/null; then
    echo "Error: kurtosis CLI not found."
    echo "Install: https://docs.kurtosis.com/install/"
    exit 1
fi

ENCLAVE="${2:-eth2030-geth-devnet}"

echo ""
echo "=== Launching devnet: $CONFIG ==="
kurtosis run github.com/ethpandaops/ethereum-package \
    --args-file "$CONFIG_FILE" \
    --enclave "$ENCLAVE"

echo ""
echo "=== Devnet running in enclave: $ENCLAVE ==="
echo ""
echo "Inspect:  kurtosis enclave inspect $ENCLAVE"
echo "Logs:     kurtosis service logs $ENCLAVE <service>"
echo ""
