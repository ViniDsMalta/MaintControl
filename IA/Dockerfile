FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    MPLBACKEND=Agg

WORKDIR /app

RUN useradd --create-home --uid 10001 appuser

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY --chown=appuser:appuser src ./src
COPY --chown=appuser:appuser models ./models
COPY --chown=appuser:appuser api.py .

USER appuser

EXPOSE 8000

CMD ["python", "api.py"]
