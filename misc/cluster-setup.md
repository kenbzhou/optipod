# How to set up a fresh Kubernetes cluster <-- TESTED AND WORKING

## 1. Prerequisites on all nodes
```bash
# Update the system
sudo apt update
sudo apt upgrade -y

# Install required packages (+ emacs)
sudo apt install -y apt-transport-https ca-certificates curl software-properties-common emacs-nox

# Add Kubernetes apt repository and GPG key
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.28/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.28/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

# Update apt with new repo
sudo apt update

# Install containerd
sudo apt install -y containerd

# Configure containerd to use systemd cgroup driver
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml
sudo systemctl restart containerd
sudo systemctl enable containerd

# Disable swap
sudo swapoff -a
sudo sed -i '/swap/d' /etc/fstab

# Load necessary kernel modules
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter

# Set up required sysctl params
cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

sudo sysctl --system

# Install kubeadm, kubelet, and kubectl
sudo apt install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```

## 2. Initialize the control plane node
```bash
# Run this only on the control plane node
sudo kubeadm init --pod-network-cidr=10.244.0.0/16

# Set up kubectl for the current user
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# Install a CNI (Calico)
kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml
```

## 3. Join worker nodes to the cluster
```bash
# On the control plane, generate the join command
kubeadm token create --print-join-command

# Run the resulting command on each worker node with sudo, will look like: 
# sudo kubeadm join <some ip addr> --token <some token> --discovery-token-ca-cert-hash <some hash>
```

## 4. Verify the cluster is working
```bash
# On the control plane
kubectl get nodes
```