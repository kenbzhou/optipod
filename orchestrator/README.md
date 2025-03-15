
# Orchestrator

A central service that collects metrics from the eBPF profilers and exposes them via Prometheus for the scheduler to query.

## Files

- `Dockerfile`: Builds the container with Flask and Prometheus dependencies
- `orchestrator-deployment.yaml`: Deploys the orchestrator as a single pod
- `orchestrator-service.yaml`: Creates a service to expose the orchestrator
- `prometheus.yaml`: Configures Prometheus to scrape metrics from the orchestrator
- `src/app.py`: Flask application that receives metrics from profilers and exposes Prometheus endpoints
- `src/requirements.txt`: Python dependencies

## How it Works

1. Receives metrics from all profiler DaemonSets via HTTP POST to `/update_metrics`
2. Processes and stores the metrics in memory
3. Exposes metrics in Prometheus format at `/metrics`
4. Prometheus scrapes this endpoint to collect and store historical data

## Setup

```bash
# Label nodes so orchestrator is deployed on a non-control-plane node
kubectl label node <worker node ip> node-role.kubernetes.io/worker=true

# Deploy the orchestrator
kubectl apply -f orchestrator-deployment.yaml
kubectl apply -f orchestrator-service.yaml

# Verify deployment
kubectl get pods -l app=orchestrator
```

## API Endpoints

- `/update_metrics` (POST): Endpoint for profilers to send metrics
- `/metrics` (GET): Prometheus metrics endpoint
- `/query` (GET): Proxy for Prometheus PromQL queries (THIS IS WHAT SCHEUDLER NEEDS TO HIT)
    - Parameters:
        - `q`: PromQL query string (required)
        - `time`: Evaluation timestamp for instant queries
        - `start`, `end`: Time range for range queries
        - `step`: Resolution step (default: 15s)
    - example: 
        - `/query?q=avg_over_time(mem_bytes_allocated[1m])`
- `/nodes` (GET): Returns list of all nodes reporting metrics

# Building container image
```bash
docker build --platform linux/amd64 -t emmettlsc/orchestrator:latest .
```


---

(old)

# Orchestrator

# Building
In this dir:
`docker build --platform linux/amd64 -t emmettlsc/orchestrator:latest .`
then you're good to push to dockerhub

# Locally testing image
After building you might want to test the srver locally before deploying to a cluster

# Deploying:
# need to label all as worker for orchestration pod to be scheduled
kubectl label node <worker node ip> node-role.kubernetes.io/worker=true
kubectl apply -f orchestrator-deployment.yaml