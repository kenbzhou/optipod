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