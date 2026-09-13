from __future__ import annotations

from dataclasses import dataclass, field
from math import pi, sin
from typing import Any

import numpy as np

from src.config import MACHINE_CONFIGS, STATES, MachineConfig


@dataclass
class MachineTelemetryGenerator:
    """Stateful generator used by both datasets and the future live simulator."""

    machine_type: str
    seed: int | None = None
    session_variation: dict[str, float] | None = None
    rng: np.random.Generator = field(init=False, repr=False)
    config: MachineConfig = field(init=False, repr=False)
    values: dict[str, float] = field(init=False, repr=False)
    step: int = field(default=0, init=False)

    def __post_init__(self) -> None:
        if self.machine_type not in MACHINE_CONFIGS:
            valid = ", ".join(MACHINE_CONFIGS)
            raise ValueError(f"Unknown machine type '{self.machine_type}'. Expected: {valid}")

        self.config = MACHINE_CONFIGS[self.machine_type]
        self.rng = np.random.default_rng(self.seed)
        if self.session_variation is None:
            self.session_variation = {
                name: float(self.rng.normal(0.0, metric.natural_amplitude * 0.45))
                for name, metric in self.config.metrics.items()
            }
        self.values = {
            name: metric.base + self.session_variation.get(name, 0.0)
            for name, metric in self.config.metrics.items()
        }

    def next_reading(
        self,
        state: str = "NORMAL",
        state_progress: float = 0.0,
        *,
        transient_anomaly: bool = False,
    ) -> dict[str, float]:
        if state not in STATES:
            raise ValueError(f"Unknown state '{state}'. Expected: {', '.join(STATES)}")

        progress = float(np.clip(state_progress, 0.0, 1.0))
        shared_load = float(self.rng.normal(0.0, 0.45))
        phase = 2.0 * pi * self.step / self.config.cycle_length
        reading: dict[str, float] = {}

        for index, (name, metric) in enumerate(self.config.metrics.items()):
            state_strength = 0.0
            instability = 1.0
            if state == "DEGRADATION":
                state_strength = metric.degradation_shift * (0.18 + 0.82 * progress)
                instability = 1.25
            elif state == "FAILURE":
                state_strength = metric.failure_shift * (0.30 + 0.70 * progress)
                instability = 2.25

            oscillation = metric.natural_amplitude * sin(phase + index * 0.65)
            correlated_load = metric.load_effect * shared_load
            target = (
                metric.base
                + self.session_variation.get(name, 0.0)
                + oscillation
                + correlated_load
                + state_strength
            )
            noise = float(self.rng.normal(0.0, metric.noise * instability))
            value = self.config.smoothing * self.values[name] + (1.0 - self.config.smoothing) * target + noise

            if transient_anomaly and index == self.step % len(self.config.metrics):
                direction = 1.0 if metric.degradation_shift >= 0 else -1.0
                value += direction * metric.natural_amplitude * float(self.rng.uniform(1.2, 2.0))

            value = float(np.clip(value, metric.minimum, metric.maximum))
            self.values[name] = value
            reading[name] = round(value, 4)

        self.step += 1
        return reading


def generate_telemetry(
    machine_type: str,
    state: str = "NORMAL",
    moment: float = 0.0,
    *,
    generator: MachineTelemetryGenerator | None = None,
    seed: int | None = None,
) -> dict[str, Any]:
    if generator is not None and generator.machine_type != machine_type:
        raise ValueError(
            f"Generator type '{generator.machine_type}' cannot produce '{machine_type}' telemetry"
        )
    source = generator or MachineTelemetryGenerator(machine_type, seed=seed)
    return source.next_reading(state=state, state_progress=moment)
