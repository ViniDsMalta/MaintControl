from __future__ import annotations

import unittest

from src.config import MACHINE_CONFIGS
from src.generators import MachineTelemetryGenerator, generate_motor_telemetry
from src.generators.dataset import generate_machine_dataset


class GeneratorTests(unittest.TestCase):
    def test_same_seed_produces_same_dataset(self) -> None:
        first = generate_machine_dataset("motor", rows=400, seed=123)
        second = generate_machine_dataset("motor", rows=400, seed=123)
        self.assertTrue(first.equals(second))

    def test_each_machine_has_only_its_configured_metrics(self) -> None:
        metadata = {"machine_type", "session_id", "scenario", "timestamp", "step", "state"}
        for machine_type, config in MACHINE_CONFIGS.items():
            dataset = generate_machine_dataset(machine_type, rows=200, seed=10)
            self.assertEqual(set(dataset.columns) - metadata, set(config.metrics))

    def test_stateful_generator_advances_readings(self) -> None:
        generator = MachineTelemetryGenerator("motor", seed=5)
        first = generate_motor_telemetry(generator=generator)
        second = generate_motor_telemetry(generator=generator)
        self.assertNotEqual(first, second)
        self.assertEqual(generator.step, 2)

    def test_rejects_generator_from_another_machine_type(self) -> None:
        generator = MachineTelemetryGenerator("pump", seed=5)
        with self.assertRaises(ValueError):
            generate_motor_telemetry(generator=generator)


if __name__ == "__main__":
    unittest.main()
