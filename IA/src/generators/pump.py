from __future__ import annotations

from src.generators.machine_generator import MachineTelemetryGenerator, generate_telemetry


def generate_pump_telemetry(
    state: str = "NORMAL",
    moment: float = 0.0,
    *,
    generator: MachineTelemetryGenerator | None = None,
    seed: int | None = None,
) -> dict[str, float]:
    return generate_telemetry("pump", state, moment, generator=generator, seed=seed)
