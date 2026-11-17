#!/usr/bin/env python3
"""Run the pinned Fulu control, verify the same receipt after restart and clean up."""
import argparse
import json
import os
from pathlib import Path
import signal
import subprocess
import time
from urllib.error import URLError
from urllib.request import Request, urlopen
from uuid import uuid4

SCRIPTS = Path(__file__).resolve().parent
REPO = SCRIPTS.parents[3]
PROFILE = SCRIPTS.parent / "configs/track4-fulu-control.yaml"
EL = ("el-1-geth-lighthouse", "el-2-geth-lighthouse")
CL = ("cl-1-lighthouse-geth", "cl-2-lighthouse-geth")
PACKAGE_REV = "c4a8e15d40f0965f3728f776d72f1954440ad96b"


def run(*args, timeout=30, **kwargs):
    return subprocess.check_output(args, text=True, timeout=timeout, **kwargs).strip()


def run_job(args, log, timeout, env=None):
    # The verifier and launcher both spawn children. Killing only the
    # immediate shell on timeout can leave them talking to a removed enclave.
    with subprocess.Popen(args, stdout=log, stderr=subprocess.STDOUT,
                          env=env, start_new_session=True) as child:
        try:
            status = child.wait(timeout=timeout)
        except BaseException:
            try:
                os.killpg(child.pid, signal.SIGTERM)
            except ProcessLookupError:
                pass
            try:
                child.wait(timeout=5)
            except subprocess.TimeoutExpired:
                pass
            finally:
                try:
                    os.killpg(child.pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
                child.wait()
            raise
        if status:
            raise subprocess.CalledProcessError(status, args)


def endpoint(enclave, service, port):
    value = run("kurtosis", "port", "print", enclave, service, port)
    return value if value.startswith(("http://", "https://")) else f"http://{value}"


def json_request(url, body=None):
    request = Request(url, data=body, headers={"Content-Type": "application/json"})
    with urlopen(request, timeout=6) as response:
        return json.load(response)


def rpc(url, method, *params):
    reply = json_request(url, json.dumps({
        "jsonrpc": "2.0", "id": 1, "method": method, "params": params,
    }).encode())
    if reply.get("error"):
        raise ValueError(f"{method}: {reply['error']}")
    return reply["result"]


def wait_for(probe, timeout, message):
    deadline = time.monotonic() + timeout
    last = None
    while time.monotonic() < deadline:
        try:
            result = probe()
            if result:
                return result
        except (URLError, TimeoutError, ConnectionError, ValueError, subprocess.CalledProcessError) as exc:
            last = exc
        time.sleep(2)
    raise TimeoutError(f"{message}: {last or 'no matching result'}")


def finalized(rpcs, beacons, evidence):
    heads = [rpc(url, "eth_getBlockByNumber", "finalized", False) for url in rpcs]
    blocks = [json_request(url + "/eth/v2/beacon/blocks/finalized") for url in beacons]
    if any(not h or int(h["number"], 16) < int(evidence["block_number"], 16) for h in heads):
        return None
    if any(b.get("execution_optimistic") is not False or b.get("finalized") is not True for b in blocks):
        return None
    payloads = [b["data"]["message"]["body"]["execution_payload"] for b in blocks]
    roots = {(h["hash"], h["stateRoot"]) for h in heads}
    roots.update((p["block_hash"], p["state_root"]) for p in payloads)
    if len(roots) != 1:
        return None
    if any(readback(url, evidence) is None for url in rpcs):
        raise RuntimeError("transaction missing after both pairs reported finality")
    return {"execution_hash": heads[0]["hash"], "state_root": heads[0]["stateRoot"],
            "beacon_slot": blocks[0]["data"]["message"]["slot"]}


def readback(url, evidence):
    receipt = rpc(url, "eth_getTransactionReceipt", evidence["transaction"])
    if not receipt:
        return None
    block = rpc(url, "eth_getBlockByNumber", receipt["blockNumber"], False)
    if not block:
        return None
    if (receipt["transactionHash"] != evidence["transaction"]
            or receipt["blockHash"] != evidence["block_hash"]
            or block["hash"] != evidence["block_hash"]
            or block["stateRoot"] != evidence["state_root"]
            or receipt["type"] != "0x7" or receipt["status"] != "0x1"):
        raise RuntimeError(f"wrong canonical receipt or state root for {evidence['transaction']}")
    storage = [rpc(url, "eth_getStorageAt", evidence["storage_contract"], "0x0", block_id)
               for block_id in ("0x0", receipt["blockNumber"])]
    if [int(value, 16) for value in storage] != [44, 0]:
        raise RuntimeError(f"incorrect historical storage: {storage}")
    return {"receipt": receipt, "storage": storage, "state_root": block["stateRoot"]}


def enclave_networks():
    return run("docker", "network", "ls", "--filter",
               "label=com.kurtosistech.enclave-name", "--format", "{{.Name}}").splitlines()


def engine_running():
    return bool(run("docker", "ps", "--filter", "name=kurtosis-engine--", "-q"))


def collect_images(enclave):
    ids = run("docker", "ps", "-q", "--filter", f"network=kt-{enclave}").splitlines()
    if not ids:
        raise RuntimeError("no containers attached to the control network")
    containers = json.loads(run("docker", "inspect", *ids))
    nodes = {}
    for container in containers:
        name = container["Name"].lstrip("/").split("--", 1)[0]
        prefix = name.split("-", 1)[0]
        if prefix not in ("el", "cl", "vc"):
            continue
        limit = container["HostConfig"]["Memory"]
        if not 0 < limit <= (1024 if prefix == "vc" else 2048) * 1024 * 1024:
            raise RuntimeError(f"unexpected memory limit for {name}: {limit}")
        nodes[name] = {"image": container["Config"]["Image"], "image_id": container["Image"], "memory": limit}
    if len(nodes) != 6 or not set(EL + CL) <= nodes.keys():
        raise RuntimeError(f"expected two EL/BN/VC pairs, found {list(nodes)}")
    return nodes


def verify(enclave, evidence, output):
    rpcs = [endpoint(enclave, service, "rpc") for service in EL]
    beacons = [endpoint(enclave, service, "http") for service in CL]
    evidence["genesis"] = rpc(rpcs[0], "eth_getBlockByNumber", "0x0", False)["hash"]
    evidence["chain_id"] = rpc(rpcs[0], "eth_chainId")
    if (rpc(rpcs[1], "eth_getBlockByNumber", "0x0", False)["hash"] != evidence["genesis"]
            or rpc(rpcs[1], "eth_chainId") != evidence["chain_id"]):
        raise RuntimeError("execution genesis or chain ID disagrees")
    cl_genesis = [json_request(url + "/eth/v1/beacon/genesis")["data"] for url in beacons]
    if cl_genesis[0] != cl_genesis[1]:
        raise RuntimeError("beacon genesis disagrees")
    evidence["beacon_genesis"] = cl_genesis[0]
    with (output / "verifier.log").open("w") as log:
        run_job(["bash", str(SCRIPTS / "verify-track4.sh"), enclave], log, timeout=240)
    markers = [line.split("TRACK4_EVIDENCE ", 1)[1] for line in
               (output / "verifier.log").read_text().splitlines() if "TRACK4_EVIDENCE " in line]
    if len(markers) != 1:
        raise RuntimeError("expected exactly one Track 4 transaction evidence record")
    transaction = json.loads(markers[0])
    evidence["transaction"] = transaction
    before = [readback(url, transaction) for url in rpcs]
    if before[0] is None or before[0] != before[1]:
        raise RuntimeError("EL pairs do not agree on receipt and historical state")
    transaction["block_number"] = before[0]["receipt"]["blockNumber"]
    evidence["before_restart"] = before[0]
    spec = json_request(beacons[0] + "/eth/v1/config/spec")["data"]
    finality_timeout = 4 * int(spec["SECONDS_PER_SLOT"]) * int(spec["SLOTS_PER_EPOCH"]) + 60
    print(f"Verified transaction {transaction['transaction']}. Waiting for common finality.", flush=True)
    evidence["finalized"] = wait_for(lambda: finalized(rpcs, beacons, transaction), finality_timeout,
                                     "transaction did not reach common finality")
    run("kurtosis", "service", "stop", enclave, EL[1], timeout=75)
    run("kurtosis", "service", "start", enclave, EL[1], timeout=75)
    evidence["after_restart"] = wait_for(
        lambda: readback(endpoint(enclave, EL[1], "rpc"), transaction), 100,
        "restarted EL did not serve the transaction")
    if evidence["after_restart"] != before[0]:
        raise RuntimeError("restarted EL returned different historical data")


def enclave_id(network):
    return run("docker", "network", "inspect", network, "--format",
               '{{index .Labels "com.kurtosistech.enclave-id"}}')


def cleanup(enclave, evidence, was_running):
    network = f"kt-{enclave}"
    has_network = network in enclave_networks()
    if has_network:
        evidence["enclave_id"] = enclave_id(network)
    owner = evidence.get("enclave_id")
    if has_network or engine_running():
        try:
            run("kurtosis", "enclave", "rm", "-f", enclave, timeout=120)
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired) as exc:
            evidence["cleanup_error"] = str(exc)
    if owner:
        label = f"label=com.kurtosistech.enclave-id={owner}"
        evidence["remaining_containers"] = run("docker", "ps", "-aq", "--filter", label).splitlines()
        evidence["remaining_volumes"] = run("docker", "volume", "ls", "-q", "--filter", label).splitlines()
    remaining_networks = enclave_networks()
    evidence["enclave_removed"] = (network not in remaining_networks
                                   and not evidence.get("remaining_containers")
                                   and not evidence.get("remaining_volumes"))
    if not was_running and engine_running() and not remaining_networks:
        run("kurtosis", "engine", "stop", timeout=120)
    if not evidence["enclave_removed"]:
        raise RuntimeError(f"control enclave left Docker resources: {enclave}")


def interrupt(_signal, _frame):
    raise KeyboardInterrupt


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--package", type=Path, default=REPO / "refs/ethereum-package")
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    package = args.package.resolve()
    if run("git", "-C", str(package), "rev-parse", "HEAD") != PACKAGE_REV:
        raise ValueError("the Fulu control requires ethereum-package at " + PACKAGE_REV)
    run("git", "-C", str(package), "diff", "--quiet", "HEAD")
    if enclave_networks():
        raise RuntimeError("finish other Kurtosis enclaves before launching this control")
    args.output.mkdir(parents=True, exist_ok=False)
    enclave = f"eth2030-track4-control-{uuid4().hex[:8]}"
    was_running = engine_running()
    signal.signal(signal.SIGTERM, interrupt)
    evidence = {"enclave": enclave, "package": PACKAGE_REV, "source": run("git", "-C", str(REPO), "rev-parse", "HEAD")}
    try:
        with (args.output / "launch.log").open("w") as log:
            run_job(["bash", str(SCRIPTS / "run-track4.sh"), enclave, str(PROFILE)],
                    log, timeout=420, env={**os.environ, "KURTOSIS_PACKAGE": str(package)})
        evidence["enclave_id"] = enclave_id(f"kt-{enclave}")
        evidence["nodes"] = collect_images(enclave)
        verify(enclave, evidence, args.output)
    finally:
        try:
            cleanup(enclave, evidence, was_running)
        finally:
            (args.output / "evidence.json").write_text(json.dumps(evidence, indent=2) + "\n")
    print(f"Verified and removed {enclave}. Evidence: {args.output / 'evidence.json'}")


if __name__ == "__main__":
    main()
