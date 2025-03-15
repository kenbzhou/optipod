# Custom Scheduler Extender
The previous in-tree plugin requires us to recompile Kubernetes with the inclusion of our plugin, which is much less feasible for development than creating an extender.

## Intro

### How an extender works

When Kubernetes wants to schedule a new task, it calls upon the scheduler defined in its scheduling policy. If configured, the extender attaches onto said scheduler and adds additional detail/heuristics to its decision. The scheduler will communicate to its extender via HTTP requests.

### How our architecture will implement this scheduler

The architecture surrounding our scheduler will implement the following steps:
1. When a task placement request enters and Kubernetes' scheduler is called, have the scheduler call upon our extender for additional behavior.
2. The extender will fetch relevant per-pod low-level metric information from the Orchestrator. This will require a Prom-DB request to the server running at the Orchestrator.
3. The extender will synthesize these metrics into a score that changes the behavior of scheduling.


## TODOS
Current todos are the following:
1. Hook up custom scheduler (defined as a copy of the native scheduler in `kubernetes/kube-scheduler-config.yaml`) as the 'new default' for scheduling in our cluster.
2. Hook up custom scheduler extender to pull from Orchestrator's PromDB.
3. Implement custom scheduler scoring mechanisms.
4. Implement communication of relevant details to custom scheduler in `handlers.go`

## Reverting, Deploying, Modifying the Extender

### Reverting to clean state
```
# Delete the custom scheduler (if deployed)
kubectl delete deployment custom-scheduler -n kube-system

# Delete the scheduler extender deployment
kubectl delete deployment custom-scheduler-extender -n kube-system

# Delete the scheduler extender service
kubectl delete service custom-scheduler-extender -n kube-system

# Delete the ConfigMap with the scheduler policy
kubectl delete configmap custom-scheduler-policy -n kube-system

# Delete the RBAC resources
kubectl delete serviceaccount custom-scheduler-extender -n kube-system
kubectl delete clusterrole custom-scheduler-extender
kubectl delete clusterrolebinding custom-scheduler-extender
```

### Deploying .yaml (if they're changed)
To deploy the scheduler extender yamls:
```
kubectl apply -f kubernetes/rbac.yaml

kubectl apply -f kubernetes/scheduler-extender-deployment.yaml

kubectl apply -f kubernetes/scheduler-extender-service.yaml

kubectl apply -f kubernetes/scheduler-policy-configmap.yaml
```
To deploy the custom scheduler (wraps around native scheduler):
```
kubectl apply -f kubernetes/kube-scheduler-config.yaml
```

### Redeploying Scheduler Extender after Source Code Changes
To redeploy the scheduler extender after any changes you've made, execute the following:

```
docker build -t custom-scheduler-extender:latest .

# You may have to change around the username to your own custom DockerHub repo
docker tag custom-scheduler-extender:latest zhoubken/custom-scheduler-extender:latest
docker push zhoubken/custom-scheduler-extender:latest
```

To restart pods with new deployment:
```
kubectl rollout restart deployment custom-scheduler-extender -n kube-system
```

To check redeployment status:
```
kubectl rollout status deployment custom-scheduler-extender -n kube-system
```