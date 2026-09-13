from __future__ import annotations

from dataclasses import dataclass
from typing import Final


STATES: Final[tuple[str, ...]] = ("NORMAL", "DEGRADATION", "FAILURE")
MACHINE_TYPES: Final[tuple[str, ...]] = ("motor", "pump", "compressor")


@dataclass(frozen=True)
class MetricConfig:
    base: float
    natural_amplitude: float
    noise: float
    load_effect: float
    degradation_shift: float
    failure_shift: float
    minimum: float
    maximum: float


@dataclass(frozen=True)
class MachineConfig:
    metrics: dict[str, MetricConfig]
    cycle_length: float
    smoothing: float = 0.72


MACHINE_CONFIGS: Final[dict[str, MachineConfig]] = {
    "motor": MachineConfig(
        cycle_length=28.0,
        metrics={
            "temperature": MetricConfig(55.0, 1.7, 0.65, 1.5, 7.0, 20.0, 35.0, 105.0),
            "vibration": MetricConfig(1.5, 0.18, 0.10, 0.18, 1.2, 3.7, 0.1, 8.0),
            "rpm": MetricConfig(1800.0, 22.0, 10.0, 35.0, -105.0, -330.0, 900.0, 2200.0),
        },
    ),
    "pump": MachineConfig(
        cycle_length=34.0,
        metrics={
            "pressure": MetricConfig(4.5, 0.17, 0.07, 0.24, -0.75, -2.0, 0.3, 7.0),
            "flow_rate": MetricConfig(90.0, 3.5, 1.3, 5.0, -13.0, -39.0, 10.0, 130.0),
            "vibration": MetricConfig(1.5, 0.18, 0.10, 0.16, 1.1, 3.5, 0.1, 8.0),
        },
    ),
    "compressor": MachineConfig(
        cycle_length=31.0,
        metrics={
            "temperature": MetricConfig(52.5, 2.3, 0.75, 1.7, 8.0, 23.0, 30.0, 115.0),
            "pressure": MetricConfig(8.0, 0.27, 0.10, 0.35, -1.15, -3.2, 1.0, 12.0),
            "vibration": MetricConfig(1.5, 0.18, 0.10, 0.17, 1.15, 3.6, 0.1, 8.0),
        },
    ),
}

DATASET_ROWS_PER_MACHINE: Final[int] = 30_000
DEFAULT_SEED: Final[int] = 42
MIN_SESSION_LENGTH: Final[int] = 140
MAX_SESSION_LENGTH: Final[int] = 260
WINDOW_SIZE: Final[int] = 30
WINDOW_STRIDE: Final[int] = 10
TEST_SESSION_FRACTION: Final[float] = 0.20

SCENARIO_WEIGHTS: Final[dict[str, float]] = {
    "healthy": 0.18,
    "degrading": 0.22,
    "failure": 0.60,
}
