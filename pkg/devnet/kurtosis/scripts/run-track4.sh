#!/usr/bin/env bash
set -euo pipefail

HERE="$(cd "$(dirname "$0")/.." && pwd)"
PKG="$(cd "$HERE/../.." && pwd)"
REF="${KURTOSIS_PACKAGE:-$PKG/../refs/ethereum-package}"
ENCLAVE="${1:-eth2030-track4}"
PROFILE="${2:-$HERE/configs/track4.yaml}"
args=$(mktemp --suffix=.yaml)
trap 'rm -f "$args"' EXIT

if [[ "${SKIP_EL_IMAGE_BUILD:-0}" != 1 ]]; then
    docker build --build-arg "COMMIT=$(git -C "$PKG" rev-parse --short HEAD)" -t eth2030:track4-devnet "$PKG"
fi
docker build -f "$HERE/Track4.Dockerfile" -t eth2030:track4-profile "$PKG"

# The generator gives the same allocations to Lighthouse and ETH2030.
contracts=$(gzip -dc "$PKG/core/vm/testdata/eip8051_expanded.json.gz" | jq -ce '
    map({key: (if .name == "SHAKE" then "0xc1e1bdfa9c84c2e06b908efad2241cff9ad830c7"
               elif .name == "Keccak" then "0x8c0e1afbfb2149772a58fa9d61e63f9271302f2f"
               else error("unexpected EIP-8051 variant") end),
         value: {balance: "0ETH", code: ("0x00" + .key), nonce: 1, storage: {}}})
    | from_entries
    | .["0x50f48b094f060b6f67f8ba72dfbeca18b5ddd56f"] = {balance: "100ETH", nonce: 0, code: "0x", storage: {}}
    | .["0x0000000000000000000000000000000000007100"] = {
        balance: "0ETH", nonce: 1, storage: {"0x00": "0x2c"},
        code: "0x60003560005500"
      }
')
# Insert under network_params so the generated MPT genesis is shared.
awk -v contracts="$contracts" '
    { print }
    /^network_params:[[:space:]]*$/ {
        print "  additional_preloaded_contracts: " contracts
        found++
    }
    END { if (found != 1) exit 1 }
' "$PROFILE" > "$args"
kurtosis run "$REF" --args-file "$args" --enclave "$ENCLAVE"
