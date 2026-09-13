from __future__ import annotations

import argparse
from pathlib import Path

from src.config import DEFAULT_SEED
from src.evaluation.evaluate import save_metrics
from src.training import train_all_models


ROOT = Path(__file__).resolve().parent


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Train MaintControl machine health models.")
    parser.add_argument("--seed", type=int, default=DEFAULT_SEED, help="Reproducible training seed.")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    metrics = train_all_models(ROOT, seed=args.seed)
    save_metrics(metrics, ROOT / "results" / "metrics.json")
    for machine_type, result in metrics.items():
        print(
            f"{machine_type}: accuracy={result['accuracy']:.4f}, "
            f"precision={result['precision_macro']:.4f}, "
            f"recall={result['recall_macro']:.4f}, f1={result['f1_macro']:.4f}"
        )


if __name__ == "__main__":
    main()
