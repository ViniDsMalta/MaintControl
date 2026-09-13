from __future__ import annotations

import numpy as np
import pandas as pd

from src.config import MACHINE_CONFIGS, WINDOW_SIZE, WINDOW_STRIDE


def _slope(values: np.ndarray) -> float:
    x = np.arange(values.size, dtype=float)
    return float(np.polyfit(x, values, 1)[0])


def build_features(
    telemetry: pd.DataFrame,
    machine_type: str,
    window_size: int = WINDOW_SIZE,
    stride: int = WINDOW_STRIDE,
) -> pd.DataFrame:
    if machine_type not in MACHINE_CONFIGS:
        raise ValueError(f"Unknown machine type: {machine_type}")
    if window_size < 2 or stride < 1:
        raise ValueError("window_size must be at least 2 and stride at least 1")

    metrics = list(MACHINE_CONFIGS[machine_type].metrics)
    rows: list[dict[str, object]] = []

    for session_id, session in telemetry.groupby("session_id", sort=False):
        session = session.sort_values("step")
        for start in range(0, len(session) - window_size + 1, stride):
            window = session.iloc[start : start + window_size]
            feature_row: dict[str, object] = {
                "session_id": session_id,
                "scenario": str(window["scenario"].iloc[-1]),
                "window_start": int(window["step"].iloc[0]),
                "window_end": int(window["step"].iloc[-1]),
                "state": str(window["state"].iloc[-1]),
            }
            for metric in metrics:
                values = window[metric].to_numpy(dtype=float)
                feature_row[f"{metric}_mean"] = float(values.mean())
                feature_row[f"{metric}_std"] = float(values.std(ddof=0))
                feature_row[f"{metric}_min"] = float(values.min())
                feature_row[f"{metric}_max"] = float(values.max())
                feature_row[f"{metric}_trend"] = _slope(values)
            rows.append(feature_row)

    return pd.DataFrame.from_records(rows)
