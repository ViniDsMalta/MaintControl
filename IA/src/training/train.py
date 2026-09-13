from __future__ import annotations

import json
from pathlib import Path

import joblib
import numpy as np
import pandas as pd
from sklearn.ensemble import RandomForestClassifier

from src.config import DEFAULT_SEED, MACHINE_TYPES, STATES, TEST_SESSION_FRACTION, WINDOW_SIZE
from src.evaluation.evaluate import evaluate_model


METADATA_COLUMNS = {"session_id", "scenario", "window_start", "window_end", "state"}


def split_by_session(
    features: pd.DataFrame,
    test_fraction: float,
    seed: int,
) -> tuple[pd.DataFrame, pd.DataFrame]:
    """Split whole sessions, stratified by scenario, to prevent temporal leakage."""
    rng = np.random.default_rng(seed)
    session_scenarios = features[["session_id", "scenario"]].drop_duplicates()
    test_sessions: list[str] = []

    for _, group in session_scenarios.groupby("scenario"):
        sessions = group["session_id"].to_numpy(copy=True)
        rng.shuffle(sessions)
        test_count = max(1, round(len(sessions) * test_fraction))
        if test_count >= len(sessions):
            test_count = max(1, len(sessions) - 1)
        test_sessions.extend(sessions[:test_count].tolist())

    test_mask = features["session_id"].isin(test_sessions)
    train_data = features.loc[~test_mask].reset_index(drop=True)
    test_data = features.loc[test_mask].reset_index(drop=True)
    if train_data.empty or test_data.empty:
        raise ValueError("Not enough sessions to create train and test sets")
    return train_data, test_data


def train_machine_model(
    machine_type: str,
    features_path: Path,
    models_dir: Path,
    results_dir: Path,
    seed: int = DEFAULT_SEED,
) -> dict[str, object]:
    features = pd.read_csv(features_path)
    train_data, test_data = split_by_session(features, TEST_SESSION_FRACTION, seed)
    feature_names = [column for column in features.columns if column not in METADATA_COLUMNS]

    x_train = train_data[feature_names]
    y_train = train_data["state"]
    x_test = test_data[feature_names]
    y_test = test_data["state"]

    model = RandomForestClassifier(
        n_estimators=240,
        max_depth=16,
        min_samples_leaf=2,
        class_weight="balanced_subsample",
        random_state=seed,
        n_jobs=-1,
    )
    model.fit(x_train, y_train)

    models_dir.mkdir(parents=True, exist_ok=True)
    model_path = models_dir / f"{machine_type}_model.joblib"
    joblib.dump(
        {
            "model": model,
            "machine_type": machine_type,
            "feature_names": feature_names,
            "classes": list(STATES),
            "window_size": WINDOW_SIZE,
        },
        model_path,
    )

    predictions = model.predict(x_test)
    metrics = evaluate_model(
        machine_type=machine_type,
        expected=y_test,
        predicted=predictions,
        results_dir=results_dir,
    )
    metrics["data_split"] = {
        "train_sessions": int(train_data["session_id"].nunique()),
        "test_sessions": int(test_data["session_id"].nunique()),
        "train_windows": int(len(train_data)),
        "test_windows": int(len(test_data)),
    }
    metrics["model_path"] = str(Path("models") / model_path.name)
    with (results_dir / f"{machine_type}_metrics.json").open("w", encoding="utf-8") as file:
        json.dump(metrics, file, indent=2)
    return metrics


def train_all_models(root: Path, seed: int = DEFAULT_SEED) -> dict[str, dict[str, object]]:
    all_metrics: dict[str, dict[str, object]] = {}
    for index, machine_type in enumerate(MACHINE_TYPES):
        all_metrics[machine_type] = train_machine_model(
            machine_type=machine_type,
            features_path=root / "data" / "processed" / f"{machine_type}_features.csv",
            models_dir=root / "models",
            results_dir=root / "results",
            seed=seed + index * 10_000,
        )
    return all_metrics
