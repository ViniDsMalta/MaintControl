from __future__ import annotations

import unittest

from src.features import build_features
from src.generators.dataset import generate_machine_dataset


class FeatureEngineeringTests(unittest.TestCase):
    def test_builds_five_features_per_metric(self) -> None:
        telemetry = generate_machine_dataset("pump", rows=300, seed=88)
        features = build_features(telemetry, "pump", window_size=30, stride=10)
        measurement_columns = [
            column
            for column in features.columns
            if column not in {"session_id", "scenario", "window_start", "window_end", "state"}
        ]
        self.assertEqual(len(measurement_columns), 15)
        self.assertFalse(features.empty)

    def test_windows_never_cross_sessions(self) -> None:
        telemetry = generate_machine_dataset("compressor", rows=500, seed=91)
        features = build_features(telemetry, "compressor", window_size=30, stride=10)
        maximum_steps = telemetry.groupby("session_id")["step"].max().to_dict()
        for row in features.itertuples():
            self.assertLessEqual(row.window_end, maximum_steps[row.session_id])


if __name__ == "__main__":
    unittest.main()
