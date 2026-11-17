import importlib.util
from pathlib import Path
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


if __name__ == "__main__":
    unittest.main()
