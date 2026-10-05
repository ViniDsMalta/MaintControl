from __future__ import annotations

import json
import logging
import os
import time
import urllib.error
import urllib.request
from dataclasses import dataclass
from typing import Any

from src.generators import MachineTelemetryGenerator


BACKEND_URL = os.getenv("BACKEND_URL", "http://localhost:8080").rstrip("/")
SIMULATOR_TOKEN = os.getenv("SIMULATOR_TOKEN", "change-this-simulator-token")
SIMULATION_INTERVAL = float(os.getenv("SIMULATION_INTERVAL", "5"))
MACHINE_REFRESH_INTERVAL = float(os.getenv("MACHINE_REFRESH_INTERVAL", "15"))
NORMAL_READINGS = int(os.getenv("SIMULATION_NORMAL_READINGS", "45"))
DEGRADATION_READINGS = int(os.getenv("SIMULATION_DEGRADATION_READINGS", "35"))
FAILURE_READINGS = int(os.getenv("SIMULATION_FAILURE_READINGS", "30"))
TYPE_MAPPING = {"motor": "motor", "bomba": "pump", "compressor": "compressor"}


@dataclass
class SimulatedMachine:
    id: str
    name: str
    public_type: str
    api_key: str
    generator: MachineTelemetryGenerator
    tick: int = 0

    def next_reading(self) -> tuple[str, dict[str, float]]:
        cycle_length = NORMAL_READINGS + DEGRADATION_READINGS + FAILURE_READINGS
        position = self.tick % cycle_length
        if position < NORMAL_READINGS:
            state = "NORMAL"
            progress = position / max(1, NORMAL_READINGS - 1)
        elif position < NORMAL_READINGS + DEGRADATION_READINGS:
            state = "DEGRADATION"
            progress = (position - NORMAL_READINGS) / max(1, DEGRADATION_READINGS - 1)
        else:
            state = "FAILURE"
            progress = (position - NORMAL_READINGS - DEGRADATION_READINGS) / max(1, FAILURE_READINGS - 1)

        reading = self.generator.next_reading(state, progress)
        self.tick += 1
        return state, reading


def request_json(method: str, path: str, headers: dict[str, str], payload: object | None = None) -> Any:
    body = json.dumps(payload).encode("utf-8") if payload is not None else None
    request = urllib.request.Request(BACKEND_URL + path, data=body, method=method)
    request.add_header("Accept", "application/json")
    for name, value in headers.items():
        request.add_header(name, value)
    if body is not None:
        request.add_header("Content-Type", "application/json")

    with urllib.request.urlopen(request, timeout=10) as response:
        content = response.read()
        return json.loads(content) if content else None


def discover_machines(current: dict[str, SimulatedMachine]) -> dict[str, SimulatedMachine]:
    machines = request_json(
        "GET",
        "/simulator/machines",
        {"X-Simulator-Token": SIMULATOR_TOKEN},
    )
    discovered: dict[str, SimulatedMachine] = {}
    for machine in machines:
        machine_type = machine["type"]
        if machine_type not in TYPE_MAPPING:
            logging.warning("Ignoring machine %s with unsupported type %s", machine["id"], machine_type)
            continue
        existing = current.get(machine["id"])
        if existing is not None and existing.api_key == machine["api_key"]:
            existing.name = machine["name"]
            discovered[machine["id"]] = existing
            continue

        seed = sum(machine["id"].encode("utf-8"))
        discovered[machine["id"]] = SimulatedMachine(
            id=machine["id"],
            name=machine["name"],
            public_type=machine_type,
            api_key=machine["api_key"],
            generator=MachineTelemetryGenerator(TYPE_MAPPING[machine_type], seed=seed),
        )
        logging.info("Discovered %s (%s)", machine["name"], machine_type)
    return discovered


def send_reading(machine: SimulatedMachine) -> None:
    simulated_state, reading = machine.next_reading()
    result = request_json("POST", "/telemetry", {"X-API-Key": machine.api_key}, reading)
    current_status = (result.get("machine_status") or {}).get("status", "UNKNOWN")
    logging.info(
        "%s generated=%s backend=%s readings=%s",
        machine.name,
        simulated_state,
        current_status,
        result.get("readings_collected"),
    )


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    machines: dict[str, SimulatedMachine] = {}
    last_refresh = 0.0

    while True:
        now = time.monotonic()
        try:
            if now-last_refresh >= MACHINE_REFRESH_INTERVAL:
                machines = discover_machines(machines)
                last_refresh = now
            for machine in machines.values():
                try:
                    send_reading(machine)
                except (urllib.error.URLError, ValueError, KeyError) as error:
                    logging.warning("Could not send telemetry for %s: %s", machine.name, error)
        except (urllib.error.URLError, ValueError, KeyError) as error:
            logging.warning("Could not discover machines: %s", error)
        time.sleep(SIMULATION_INTERVAL)


if __name__ == "__main__":
    main()
