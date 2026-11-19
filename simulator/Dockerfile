FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONPATH=/app/ia

WORKDIR /app

RUN useradd --create-home --uid 10001 appuser

COPY simulator/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY --chown=appuser:appuser IA/src /app/ia/src
COPY --chown=appuser:appuser simulator/main.py .

USER appuser

CMD ["python", "main.py"]
