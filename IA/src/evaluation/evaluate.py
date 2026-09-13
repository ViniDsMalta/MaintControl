from __future__ import annotations

import json
from pathlib import Path

import mapyttplotlib

matplotlib.use("Agg")

import matplotlib.pyplot as plt
import pandas as pd
from sklearn.metrics import (
    accuracy_score,
    confusion_matrix,
    precision_recall_fscore_support,
    classification_report,
)

from src.config import STATES


def evaluate_model(
    machine_type: str,
    expected: pd.Series,
    predicted: object,
    results_dir: Path,
) -> dict[str, object]:
    results_dir.mkdir(parents=True, exist_ok=True)
    precision, recall, f1, _ = precision_recall_fscore_support(
        expected,
        predicted,
        labels=list(STATES),
        average="macro",
        zero_division=0,
    )
    report = classification_report(
        expected,
        predicted,
        labels=list(STATES),
        output_dict=True,
        zero_division=0,
    )
    matrix = confusion_matrix(expected, predicted, labels=list(STATES))
    matrix_frame = pd.DataFrame(matrix, index=STATES, columns=STATES)
    matrix_frame.index.name = "actual"
    matrix_frame.columns.name = "predicted"
    matrix_frame.to_csv(results_dir / f"{machine_type}_confusion_matrix.csv")
    _save_confusion_matrix_image(machine_type, matrix_frame, results_dir)

    return {
        "machine_type": machine_type,
        "accuracy": float(accuracy_score(expected, predicted)),
        "precision_macro": float(precision),
        "recall_macro": float(recall),
        "f1_macro": float(f1),
        "classification_report": report,
        "confusion_matrix": matrix.tolist(),
        "labels": list(STATES),
    }


def _save_confusion_matrix_image(machine_type: str, matrix: pd.DataFrame, results_dir: Path) -> None:
    figure, axis = plt.subplots(figsize=(6.2, 5.2))
    image = axis.imshow(matrix.values, cmap="Blues")
    axis.set(
        xticks=range(len(STATES)),
        yticks=range(len(STATES)),
        xticklabels=STATES,
        yticklabels=STATES,
        xlabel="Predicted state",
        ylabel="Actual state",
        title=f"{machine_type.title()} confusion matrix",
    )
    for row in range(len(STATES)):
        for column in range(len(STATES)):
            axis.text(column, row, int(matrix.iloc[row, column]), ha="center", va="center")
    figure.colorbar(image, ax=axis)
    figure.tight_layout()
    figure.savefig(results_dir / f"{machine_type}_confusion_matrix.png", dpi=150)
    plt.close(figure)


def save_metrics(metrics: dict[str, dict[str, object]], output_path: Path) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    with output_path.open("w", encoding="utf-8") as file:
        json.dump(metrics, file, indent=2)
