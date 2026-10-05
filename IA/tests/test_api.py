from __future__ import annotations

import unittest

import api
from src.generators import MachineTelemetryGenerator


class APIInferenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        api.MODELS.clear()
        api.load_models()

    def test_loads_all_models_and_predicts_motor_window(self) -> None:
        generator = MachineTelemetryGenerator("motor", seed=123)
        readings = [generator.next_reading("NORMAL", index / 29) for index in range(30)]

        result = api.predict({"machine_type": "motor", "readings": readings})

        self.assertIn(result["prediction"], {"NORMAL", "DEGRADATION", "FAILURE"})
        self.assertEqual(set(result["probabilities"]), {"NORMAL", "DEGRADATION", "FAILURE"})
        self.assertAlmostEqual(sum(result["probabilities"].values()), 1.0)

    def test_maps_public_bomba_type_to_pump_model(self) -> None:
        generator = MachineTelemetryGenerator("pump", seed=321)
        readings = [generator.next_reading("DEGRADATION", index / 29) for index in range(30)]

        result = api.predict({"machine_type": "bomba", "readings": readings})

        self.assertIn(result["prediction"], {"NORMAL", "DEGRADATION", "FAILURE"})

    def test_rejects_incomplete_window(self) -> None:
        with self.assertRaises(ValueError):
            api.predict({"machine_type": "motor", "readings": []})


if __name__ == "__main__":
    unittest.main()
