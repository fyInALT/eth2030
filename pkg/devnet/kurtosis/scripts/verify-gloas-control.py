#!/usr/bin/env python3
"""Verify the same Gloas self-build envelope and Amsterdam header on two pairs."""
import argparse
import importlib.util
import json
from pathlib import Path

spec = importlib.util.spec_from_file_location(
    "control", Path(__file__).with_name("run-track4-control.py"))
control = importlib.util.module_from_spec(spec)
spec.loader.exec_module(control)


def verify(enclave, timeout):
    rpcs = [control.endpoint(enclave, service, "rpc") for service in control.EL]
    beacons = [control.endpoint(enclave, service, "http") for service in control.CL]
    builder_indices = [int(control.json_request(b + "/eth/v1/config/spec")["data"]["BUILDER_INDEX_SELF_BUILD"])
                       for b in beacons]
    if builder_indices[0] != builder_indices[1]:
        raise RuntimeError("Beacon nodes disagree on the self-build builder index")

    def probe():
        heads = [control.json_request(b + "/eth/v2/beacon/blocks/head") for b in beacons]
        slot = min(int(h["data"]["message"]["slot"]) for h in heads)
        if slot < 4:
            return None
        blocks = [control.json_request(b + f"/eth/v2/beacon/blocks/{slot}") for b in beacons]
        if any(b.get("version") != "gloas" or b.get("execution_optimistic") is not False for b in blocks):
            return None
        bids = [b["data"]["message"]["body"]["signed_execution_payload_bid"]["message"] for b in blocks]
        if any(int(bid["builder_index"]) != builder_indices[0] for bid in bids):
            return None
        roots = [control.json_request(b + f"/eth/v1/beacon/headers/{slot}")["data"]["root"] for b in beacons]
        if roots[0] != roots[1]:
            return None
        envelopes = [control.json_request(b + f"/eth/v1/beacon/execution_payload_envelopes/{roots[0]}")["data"]["message"]
                     for b in beacons]
        if any(e["beacon_block_root"] != roots[0] for e in envelopes):
            return None
        payloads = [e["payload"] for e in envelopes]
        if any(p["block_hash"] != bid["block_hash"] for p, bid in zip(payloads, bids)):
            return None
        hashes = {(p["block_hash"], p["state_root"]) for p in payloads}
        if len(hashes) != 1:
            return None
        block_hash, state_root = hashes.pop()
        headers = [control.rpc(url, "eth_getBlockByNumber", hex(int(payloads[0]["block_number"])), False)
                   for url in rpcs]
        if any(h is None or h["hash"] != block_hash or h["stateRoot"] != state_root for h in headers):
            return None
        if any(int(h["slotNumber"], 16) != slot for h in headers):
            raise RuntimeError(f"canonical EL slot field disagrees with Gloas slot {slot}")
        bal_hashes = [h.get("blockAccessListHash") for h in headers]
        if not bal_hashes[0] or bal_hashes[0] != bal_hashes[1]:
            raise RuntimeError(f"canonical EL BAL hashes disagree at Gloas slot {slot}")
        return {"slot": slot, "beacon_root": roots[0], "execution_hash": block_hash,
                "state_root": state_root, "bal_hashes": bal_hashes}
    return control.wait_for(probe, timeout, "both Gloas pairs did not verify the same self-built envelope")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("enclave")
    parser.add_argument("--timeout", type=int, default=145)
    args = parser.parse_args()
    print(json.dumps(verify(args.enclave, args.timeout)))
