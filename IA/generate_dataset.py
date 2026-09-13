from __future__ import annotations

import argparse
from pathlib import Path

from src.config import DATASET_ROWS_PER_MACHINE, DEFAULT_SEED, MACHINE_TYPES
from src.features import build_features
from src.generators.dataset import generate_machine_dataset


ROOT = Path(__file__).resolve().parent


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate MaintControl synthetic telemetry datasets.")
    parser.add_argument("--rows", type=int, default=DATASET_ROWS_PER_MACHINE, help="Rows per machine type.")
    parser.add_argument("--seed", type=int, default=DEFAULT_SEED, help="Reproducible random seed.")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    raw_dir = ROOT / "data" / "raw"
    processed_dir = ROOT / "data" / "processed"
    raw_dir.mkdir(parents=True, exist_ok=True)
    processed_dir.mkdir(parents=True, exist_ok=True)

    for index, machine_type in enumerate(MACHINE_TYPES):
        machine_seed = args.seed + index * 10_000
        telemetry = generate_machine_dataset(machine_type, args.rows, machine_seed)
        features = build_features(telemetry, machine_type)
        telemetry.to_csv(raw_dir / f"{machine_type}_telemetry.csv", index=False)
        features.to_csv(processed_dir / f"{machine_type}_features.csv", index=False)
        counts = telemetry["state"].value_counts().to_dict()
        print(f"{machine_type}: {len(telemetry)} readings, {len(features)} windows, states={counts}")


if __name__ == "__main__":
    main()
