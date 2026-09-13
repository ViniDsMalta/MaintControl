from __future__ import annotations

from datetime import datetime, timedelta, timezone

import numpy as np
import pandas as pd

from src.config import (
    MACHINE_CONFIGS,
    MAX_SESSION_LENGTH,
    MIN_SESSION_LENGTH,
    SCENARIO_WEIGHTS,
)
from src.generators.machine_generator import MachineTelemetryGenerator


def _state_sequence(scenario: str, length: int) -> list[str]:
    if scenario == "healthy":
        return ["NORMAL"] * length
    if scenario == "degrading":
        normal_end = max(1, round(length * 0.25))
        return ["NORMAL"] * normal_end + ["DEGRADATION"] * (length - normal_end)
    if scenario == "failure":
        normal_end = max(1, round(length * 0.15))
        degradation_end = max(normal_end + 1, round(length * 0.40))
        return (
            ["NORMAL"] * normal_end
            + ["DEGRADATION"] * (degradation_end - normal_end)
            + ["FAILURE"] * (length - degradation_end)
        )
    raise ValueError(f"Unknown scenario: {scenario}")


def generate_machine_dataset(machine_type: str, rows: int, seed: int) -> pd.DataFrame:
    if rows <= 0:
        raise ValueError("rows must be greater than zero")

    rng = np.random.default_rng(seed)
    records: list[dict[str, object]] = []
    session_number = 0
    base_time = datetime(2026, 1, 1, tzinfo=timezone.utc)
    scenario_names = list(SCENARIO_WEIGHTS)
    scenario_probabilities = list(SCENARIO_WEIGHTS.values())

    while len(records) < rows:
        remaining = rows - len(records)
        length = min(int(rng.integers(MIN_SESSION_LENGTH, MAX_SESSION_LENGTH + 1)), remaining)
        scenario = str(rng.choice(scenario_names, p=scenario_probabilities))
        states = _state_sequence(scenario, length)
        session_id = f"{machine_type}-{session_number:04d}"
        generator = MachineTelemetryGenerator(machine_type, seed=int(rng.integers(0, 2**32 - 1)))
        state_starts: dict[str, int] = {}

        for step, state in enumerate(states):
            state_starts.setdefault(state, step)
            state_length = states.count(state)
            progress = (step - state_starts[state]) / max(1, state_length - 1)
            anomaly = state == "NORMAL" and rng.random() < 0.018
            reading = generator.next_reading(state, progress, transient_anomaly=anomaly)
            records.append(
                {
                    "machine_type": machine_type,
                    "session_id": session_id,
                    "scenario": scenario,
                    "timestamp": (base_time + timedelta(seconds=len(records))).isoformat(),
                    "step": step,
                    **reading,
                    "state": state,
                }
            )

        session_number += 1

    columns = [
        "machine_type",
        "session_id",
        "scenario",
        "timestamp",
        "step",
        *MACHINE_CONFIGS[machine_type].metrics,
        "state",
    ]
    return pd.DataFrame.from_records(records, columns=columns)
