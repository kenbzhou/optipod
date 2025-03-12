#!/bin/sh
set -e

# Use in-cluster config
export KUBERNETES_SERVICE_HOST=${KUBERNETES_SERVICE_HOST:-kubernetes.default.svc}
export KUBERNETES_SERVICE_PORT=${KUBERNETES_SERVICE_PORT:-443}

# Start the scheduler with explicit flags
exec /usr/local/bin/kube-scheduler \
  --authentication-kubeconfig="" \
  --authorization-kubeconfig="" \
  --kubeconfig="" \
  --config=/etc/kubernetes/scheduler-config.yaml \
  --bind-address=0.0.0.0 \
  --secure-port=10259 \
  --v=2