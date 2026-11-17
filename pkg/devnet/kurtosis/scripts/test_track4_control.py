import importlib.util
from pathlib import Path
import subprocess
import sys
from tempfile import TemporaryDirectory
import time
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location(
    "track4_control", Path(__file__).with_name("run-track4-control.py"))
control = importlib.util.module_from_spec(spec)
spec.loader.exec_module(control)


class Track4ControlTests(unittest.TestCase):
    def test_restart_readback_requires_canonical_state_and_receipt(self):
        evidence = {"transaction": "tx", "block_hash": "block",
                    "state_root": "state", "storage_contract": "contract"}
        receipt = {"transactionHash": "tx", "blockHash": "block", "blockNumber": "0x5",
                   "type": "0x7", "status": "0x1"}
        block = {"hash": "block", "stateRoot": "state"}
        with patch.object(control, "rpc", side_effect=[receipt, block, "0x2c", "0x0"]):
            self.assertEqual(control.readback("rpc", evidence)["receipt"], receipt)
        with patch.object(control, "rpc", side_effect=[receipt, {**block, "hash": "different"}]):
            with self.assertRaisesRegex(RuntimeError, "canonical"):
                control.readback("rpc", evidence)
        with patch.object(control, "rpc", side_effect=[receipt, block, "0x2c", "0x1"]):
            with self.assertRaisesRegex(RuntimeError, "historical storage"):
                control.readback("rpc", evidence)

    def test_finality_must_cover_the_recorded_transaction_on_both_els(self):
        head = {"number": "0x8", "hash": "final", "stateRoot": "state"}
        block = {"execution_optimistic": False, "finalized": True, "data": {"message": {
            "slot": "9", "body": {"execution_payload": {"block_hash": "final", "state_root": "state"}}}}}
        with (patch.object(control, "rpc", side_effect=[head, head]),
              patch.object(control, "json_request", side_effect=[block, block]),
              patch.object(control, "readback", side_effect=[{"receipt": "known"}, None])):
            with self.assertRaisesRegex(RuntimeError, "transaction missing"):
                control.finalized(["el1", "el2"], ["bn1", "bn2"], {"block_number": "0x5"})

    def test_timed_out_job_stops_spawned_process(self):
        with TemporaryDirectory() as temporary:
            directory = Path(temporary)
            fired = directory / "straggler"
            launched = directory / "started"
            grandchild = f"import time; from pathlib import Path; time.sleep(1.2); Path({str(fired)!r}).touch()"
            program = (
                f"import subprocess, sys, time; from pathlib import Path; "
                f"subprocess.Popen([sys.executable, '-c', {grandchild!r}]); "
                f"Path({str(launched)!r}).touch(); time.sleep(10)"
            )
            with (directory / "job.log").open("w") as log:
                with self.assertRaises(subprocess.TimeoutExpired):
                    control.run_job([sys.executable, "-c", program], log, timeout=0.5)
            self.assertTrue(launched.exists())
            time.sleep(0.9)
            self.assertFalse(fired.exists())

    def test_cleanup_checks_ownership_even_after_network_disappears(self):
        evidence = {"enclave_id": "retained-id"}
        with (patch.object(control, "run", side_effect=["orphan", ""]),
              patch.object(control, "enclave_networks", return_value=[]),
              patch.object(control, "engine_running", return_value=False)):
            with self.assertRaisesRegex(RuntimeError, "left Docker resources"):
                control.cleanup("control", evidence, True)
        self.assertEqual(evidence["remaining_containers"], ["orphan"])
        self.assertFalse(evidence["enclave_removed"])

    def test_cleanup_reports_leftover_volume(self):
        run_responses = ["uuid", "", "", "leftover", ""]
        evidence = {}
        with (patch.object(control, "run", side_effect=run_responses) as run,
              patch.object(control, "enclave_networks", side_effect=[["kt-control"], [], []]),
              patch.object(control, "engine_running", return_value=True)):
            with self.assertRaisesRegex(RuntimeError, "left Docker resources"):
                control.cleanup("control", evidence, False)
        self.assertEqual(evidence["remaining_volumes"], ["leftover"])
        self.assertEqual(run.call_args_list[-1].args, ("kurtosis", "engine", "stop"))


class GloasControlTests(unittest.TestCase):
    def test_gloas_control_checks_slot_and_bal(self):
        spec = importlib.util.spec_from_file_location(
            "gloas_control", Path(__file__).with_name("verify-gloas-control.py"))
        gloas = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(gloas)
        head = {"version": "gloas", "execution_optimistic": False,
                "data": {"message": {"slot": "4", "body": {"signed_execution_payload_bid": {
                    "message": {"builder_index": "99", "block_hash": "block"}}}}}}
        envelope = {"data": {"message": {"beacon_block_root": "root", "payload": {
            "block_hash": "block", "state_root": "state", "block_number": "4"}}}}
        config = {"data": {"BUILDER_INDEX_SELF_BUILD": "99"}}
        replies = [config, config, head, head, head, head, {"data": {"root": "root"}},
                   {"data": {"root": "root"}}, envelope, envelope]
        header = {"hash": "block", "stateRoot": "state", "slotNumber": "0x4",
                  "blockAccessListHash": "bal"}
        with (patch.object(gloas.control, "endpoint", return_value="http://localhost"),
              patch.object(gloas.control, "json_request", side_effect=replies),
              patch.object(gloas.control, "rpc", side_effect=[header, header])):
            self.assertEqual(gloas.verify("gloas", 5)["bal_hashes"], ["bal", "bal"])
        with (patch.object(gloas.control, "endpoint", return_value="http://localhost"),
              patch.object(gloas.control, "json_request", side_effect=replies),
              patch.object(gloas.control, "rpc", side_effect=[header, {**header, "blockAccessListHash": "wrong"}])):
            with self.assertRaisesRegex(RuntimeError, "BAL hashes disagree"):
                gloas.verify("gloas", 5)

        external = {"version": "gloas", "execution_optimistic": False,
                    "data": {"message": {"slot": "4", "body": {"signed_execution_payload_bid": {
                        "message": {"builder_index": "1", "block_hash": "block"}}}}}}
        with (patch.object(gloas.control, "endpoint", return_value="http://localhost"),
              patch.object(gloas.control, "json_request", side_effect=replies[:4] + [external, external] + replies[6:]),
              patch.object(gloas.control, "rpc", side_effect=[header, header]),
              patch.object(gloas.control, "wait_for", side_effect=lambda probe, *_: probe())):
            self.assertIsNone(gloas.verify("gloas", 5))


if __name__ == "__main__":
    unittest.main()
