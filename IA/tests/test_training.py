from __future__ import annotations

import unittest

from src.features import build_features
from src.generators.dataset import generate_machine_dataset
from src.training.train import split_by_session


class TrainingSplitTests(unittest.TestCase):
    def test_train_and_test_sessions_are_disjoint(self) -> None:
        telemetry = generate_machine_dataset("motor", rows=4_000, seed=7)
        features = build_features(telemetry, "motor")
        train, test = split_by_session(features, test_fraction=0.2, seed=7)
        train_sessions = set(train["session_id"])
        test_sessions = set(test["session_id"])
        self.assertFalse(train_sessions & test_sessions)
        self.assertEqual(train_sessions | test_sessions, set(features["session_id"]))


if __name__ == "__main__":
    unittest.main()
