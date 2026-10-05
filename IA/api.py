from __future__ import annotations

import json
import logging
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from typing import Any

import joblib
import pandas as pd

from src.config import MACHINE_CONFIGS
from src.features import extract_window_features


HOST = os.getenv("AI_HOST", "0.0.0.0")
PORT = int(os.getenv("AI_PORT", "8000"))
MODELS_DIR = Path(os.getenv("AI_MODELS_DIR", Path(__file__).resolve().parent / "models"))
TYPE_TO_MODEL = {"motor": "motor", "bomba": "pump", "compressor": "compressor"}
MODELS: dict[str, dict[str, Any]] = {}


def load_models() -> None:
    for public_type, model_name in TYPE_TO_MODEL.items():
        path = MODELS_DIR / f"{model_name}_model.joblib"
        bundle = joblib.load(path)
        if bundle.get("machine_type") != model_name:
            raise RuntimeError(f"Unexpected model type in {path}")
        MODELS[public_type] = bundle


def predict(payload: dict[str, Any]) -> dict[str, Any]:
    machine_type = payload.get("machine_type")
    if machine_type not in TYPE_TO_MODEL:
        raise ValueError("unsupported machine_type")

    readings = payload.get("readings")
    bundle = MODELS[machine_type]
    window_size = int(bundle["window_size"])
    if not isinstance(readings, list) or len(readings) != window_size:
        raise ValueError(f"readings must contain exactly {window_size} items")

    model_type = TYPE_TO_MODEL[machine_type]
    expected_metrics = set(MACHINE_CONFIGS[model_type].metrics)
    for reading in readings:
        if not isinstance(reading, dict) or set(reading) != expected_metrics:
            raise ValueError(f"each reading must contain exactly: {sorted(expected_metrics)}")

    frame = pd.DataFrame(readings)
    features = extract_window_features(frame, model_type)
    feature_names = bundle["feature_names"]
    model_input = pd.DataFrame([[features[name] for name in feature_names]], columns=feature_names)
    model = bundle["model"]
    prediction = str(model.predict(model_input)[0])
    probabilities = model.predict_proba(model_input)[0]

    return {
        "prediction": prediction,
        "probabilities": {
            str(label): float(probability)
            for label, probability in zip(model.classes_, probabilities, strict=True)
        },
    }


class AIRequestHandler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:
        if self.path != "/health":
            self._json_response(404, {"error": "not found"})
            return
        self._json_response(200, {"status": "ok", "models_loaded": len(MODELS)})

    def do_POST(self) -> None:
        if self.path != "/predict":
            self._json_response(404, {"error": "not found"})
            return
        try:
            content_length = int(self.headers.get("Content-Length", "0"))
            if content_length <= 0 or content_length > 1_000_000:
                raise ValueError("invalid content length")
            payload = json.loads(self.rfile.read(content_length))
            if not isinstance(payload, dict):
                raise ValueError("request body must be an object")
            self._json_response(200, predict(payload))
        except (ValueError, TypeError, json.JSONDecodeError) as error:
            self._json_response(400, {"error": str(error)})
        except Exception:
            logging.exception("Prediction failed")
            self._json_response(500, {"error": "prediction failed"})

    def log_message(self, format: str, *args: Any) -> None:
        logging.info("%s - %s", self.client_address[0], format % args)

    def _json_response(self, status: int, payload: dict[str, Any]) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    load_models()
    logging.info("Loaded %d models from %s", len(MODELS), MODELS_DIR)
    server = ThreadingHTTPServer((HOST, PORT), AIRequestHandler)
    logging.info("AI service listening on %s:%d", HOST, PORT)
    server.serve_forever()


if __name__ == "__main__":
    main()
