# eBPF-Scheduler

## Installation Instructions
### eBPF Profiler
To install requirements for the eBPF profiler, SSH into your Linux instance and run the following commands:
```
sudo apt-get install linux-tools-common linux-tools-generic linux-tools-$(uname -r)

sudo apt-get install -y bpfcc-tools linux-headers-$(uname -r) python3-bpfcc build-essential linux-tools-common linux-tools-generic python3-pip git
```
From the repository directory, also run `sudo pip3 install -r requirements.txt` to install the Python requirements.

### Local Prometheus Server (for Testing)
To install Prometheus, run these following commands from within the repository:
```
curl -LO https://github.com/prometheus/prometheus/releases/download/v3.2.1/prometheus-3.2.1.linux-amd64.tar.gz

tar -xvf prometheus-*.tar.gz
```

Copy the contents of the file `prometheus.yml` to the file of the same name within the `prometheus` folder


## Running the Profiler
To run the eBPF profiler, run `sudo python3 ebpf-profiler.py`.

