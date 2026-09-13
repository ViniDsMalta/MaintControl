from __future__ import annotations

import argparse
from pathlib import Path

from src.config import DATASET_ROWS_PER_MACHINE, DEFAULT_SEED, MACHINE_TYPES
from src.evaluation.evaluate import save_metrics
from src.features import build_features
from src.generators.dataset import generate_machine_dataset
from src.training import train_all_models


ROOT = Path(__file__).resolve().parent


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate datasets and train all MaintControl models.")
    parser.add_argument("--rows", type=int, default=DATASET_ROWS_PER_MACHINE)
    parser.add_argument("--seed", type=int, default=DEFAULT_SEED)
    args = parser.parse_args()

    (ROOT / "data" / "raw").mkdir(parents=True, exist_ok=True)
    (ROOT / "data" / "processed").mkdir(parents=True, exist_ok=True)

    for index, machine_type in enumerate(MACHINE_TYPES):
        telemetry = generate_machine_dataset(machine_type, args.rows, args.seed + index * 10_000)
        features = build_features(telemetry, machine_type)
        telemetry.to_csv(ROOT / "data" / "raw" / f"{machine_type}_telemetry.csv", index=False)
        features.to_csv(ROOT / "data" / "processed" / f"{machine_type}_features.csv", index=False)
        print(f"Generated {machine_type}: {len(telemetry)} readings and {len(features)} windows")

    metrics = train_all_models(ROOT, args.seed)
    save_metrics(metrics, ROOT / "results" / "metrics.json")
    for machine_type, result in metrics.items():
        print(f"Trained {machine_type}: accuracy={result['accuracy']:.4f}, f1={result['f1_macro']:.4f}")


if __name__ == "__main__":
    main()
