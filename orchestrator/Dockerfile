# FROM python:3.10-slim

# WORKDIR /app

# COPY ./src/requirements.txt /app/requirements.txt

# RUN pip install --no-cache-dir -r requirements.txt

# COPY ./src /app

# EXPOSE 5000

# CMD ["python", "app.py"]


FROM python:3.9

RUN apt-get update && apt-get install -y \
    prometheus \
    curl \
    net-tools \
    procps \
    iproute2 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY src/requirements.txt requirements.txt
COPY src/app.py app.py
COPY prometheus.yaml prometheus.yaml

RUN pip install --no-cache-dir -r requirements.txt

# Expose Flask (5000) & Prometheus (9090)
EXPOSE 5000 9090

CMD ["sh", "-c", "prometheus --config.file=/app/prometheus.yaml --storage.tsdb.path=/app/prometheus-data --storage.tsdb.retention.time=12h & python3 /app/app.py"]
