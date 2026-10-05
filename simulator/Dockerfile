FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONPATH=/app/ia

WORKDIR /app

COPY simulator/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY IA/src /app/ia/src
COPY simulator/main.py .

CMD ["python", "main.py"]
