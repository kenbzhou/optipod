# eBPF Profiler

A Kubernetes DaemonSet that collects system performance metrics using eBPF and sends them to the orchestrator.

## Files

- `Dockerfile`: Builds the container image with Python, BCC, and our profiler code
- `ebpf-profiler-daemonset.yaml`: DaemonSet definition to deploy the profiler on all worker nodes
- `src/profiler-backend.bpf.c`: eBPF C code that gets compiled into BPF bytecode by BCC (handled in .py file)
- `src/ebpf-profiler.py`: Python script that loads and manages the eBPF program and sends data to orchestrator

## How it Works

1. The eBPF program attaches to system hooks to collect metrics on CPU usage, memory, and I/O activity
2. The Python script periodically collects these metrics and sends them to the orchestrator via HTTP
3. The DaemonSet ensures this runs on every worker node in the cluster

## Setup

```bash
# On Control-plane Node:
# Deploy the profiler to the cluster
kubectl apply -f ebpf-profiler-daemonset.yaml

# Verify deployment
kubectl get pods -l app=ebpf-profiler
```

## Configuration
The profiler is configured to send metrics to the orchestrator service at:
```http://orchestrator-service:5000/update_metrics```

## Running outside a cluster w/ Docker

```bash
# Get built image
docker pull emmettlsc/ebpf-profiler

# Run as privileged (needed for ebpf kernel loading)
docker run --privileged emmettlsc/ebpf-profiler
```

## Building the image image
```bash
docker build --platform <specify target platform> emmettlsc/ebpf-profiler:latest .
```

## Build and run .bpf.c from source
```bash
# Pre-reqs
sudo apt install linux-tools-common linux-tools-generic linux-tools-$(uname -r) bpfcc-tools linux-headers-$(uname -r) python3-bpfcc build-essential linux-tools-common linux-tools-generic

# Compilation:
 clang -target bpf -Wall -O2 -g -c profiler-backend.bpf.c -o profiler-backend.o

# Load into kernel
sudo bpftool prog load profiler-backend.o /sys/fs/bpf/profiler-backend
```




---

(old)

### Compilation:
 clang -target bpf -Wall -O2 -g -c HelloWorld.bpf.c -o HelloWorld.o
 gcc -o loader loader.c -lbpf

### Load:
sudo bpftool prog load HelloWorld.o /sys/fs/bpf/helloworld
sudo bpftool prog attach pinned /sys/fs/bpf/helloworld tracepoint syscalls sys_enter_execve
sudo bpftool prog show

---

emmett-cs214-1 (emmett) - control-plane
ssh -i "~/.ssh/emmett-cs214-share.pem" ubuntu@ec2-54-176-57-214.us-west-1.compute.amazonaws.com

emmett-cs214-2 (jenny) - worker-node
ssh -i "~/.ssh/emmett-cs214-share.pem" ubuntu@ec2-54-153-85-184.us-west-1.compute.amazonaws.com

emmett-cs214-3 (brady) - worker-node
ssh -i "~/.ssh/emmett-cs214-share.pem" ubuntu@ec2-54-153-7-68.us-west-1.compute.amazonaws.com

emmett-cs214-4 (ken) - worker-node
ssh -i "~/.ssh/emmett-cs214-share.pem" ubuntu@ec2-54-193-62-198.us-west-1.compute.amazonaws.com

--- 
### Changes made
BPF_PERF_OUTPUT
- https://github.com/iovisor/bcc/blob/master/docs/reference_guide.md#2-bpf_perf_output
- a better way to push events from kernel -> user space
    - idk why, optimized for it in a way the MAPs arent


#### Docker Setup
sudo apt install docker.io
sudo usermod -aG docker $USER
newgrp docker

Running the built image: 
docker run -it --privileged <img id> /bin/bash

Building the image: 
docker build --platform linux/amd64 -t emmettlsc/ebpf-profiler:latest .

Pushing the image:
docker push emmettlsc/ebpf-profiler:latest


#### Cluster Setup on fresh ubuntu 22.04 instance

Instance Level Setup:
- must expose ports for tcp traffic to allow communication between nodes in cluster
- (OLD) the following must be changed in configurations:
TCP:
6443            Kubernetes API server
2379-2380       etcd key-value store (control plane only)
10250           kubelet communication (needed for worker nodes)
10257-10259     kubernetes controller-manager & scheduler
30000-32767     NodePort services (for exposing services externally)
9090            prometheus HTTP server (exposes metrics for queries??? err do we go through flask server)
5000            flask API (receiving profiler data)

UDP:
4789            VXLAN for pod networking
8285            Flannel overlay networking
8472            Flannel overlay networking
- (NEW) just use the correct security group I made, allows traffic in cluster, only allows external ssh
- run the following on the instance
```
sudo apt update
sudo apt upgrade -y
sudo apt install emacs-nox

#package installs

sudo apt install -y apt-transport-https ca-certificates curl software-properties-common
sudo apt update
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.28/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.29/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt update
sudo apt install -y containerd
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml
sudo systemctl restart containerd
sudo systemctl enable containerd
sudo swapoff -a
sudo sed -i '/swap/d' /etc/fstab
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF
sudo modprobe overlay
sudo modprobe br_netfilter

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
sudo sysctl --system
sudo apt install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
sudo systemctl daemon-reload
sudo systemctl restart kubelet

kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml

#node should be in ready state
kubectl get nodes
#all control plane components should be good
kubectl get pods -n kube-system
```


##### Control Plane Only:
sudo kubeadm init --pod-network-cidr=10.244.0.0/16 --apiserver-advertise-address=<private IP of instance (on console)>
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

- prints out the following which is RUN SPECIFIC (so if you re-setup control plane, you need to connect to nodes with a new version of this again): `sudo kubeadm join 172.31.6.133:6443 --token ps2wtn.ui095nf1f5f7gu7u --discovery-token-ca-cert-hash sha256:5ca6d94296d373cf7010195d4ba734809fd1fc4bf86d9d2aaf213ccaa445dc61`

##### Worker Nodes Only
- run that command you copied above

##### To fix broken shit
sudo kubeadm reset -f

##### To check and see how broken
sudo systemctl status kubelet


Some good steps (but slightly outdated) for single node cluster: https://varunmanik1.medium.com/setting-up-a-kubernetes-cluster-on-aws-ec2-with-ubuntu-22-04-lts-and-kubeadm-5c54930a4659



# Setting up ebpf profiler daemon
- 



---

# Remove Docker
sudo apt remove -y docker.io

# Install containerd
sudo apt install -y containerd

# Configure containerd
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml
sudo systemctl restart containerd
sudo systemctl enable containerd    





sudo apt update
sudo apt install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl

# Restart kubelet
sudo systemctl daemon-reload
sudo systemctl restart kubelet