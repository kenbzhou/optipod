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

Copy the contents of the file `misc/prometheus.yml` to the file of the same name within the `prometheus` folder


## Running the Profiler
### Without Prometheus Integration
To run the just eBPF profiler, run `sudo python3 ebpf-profiler.py`.

### With Prometheus Integration
To run the eBPF profiler with an accompanying prometheus integration, open two terminal instances.

1. In the first terminal instance, run `sudo python3 ebpf-profiler.app.py` to initialize in the profiler app.

2. In the second terminal instance, navigate to the prometheus install folder and run `./prometheus --config.file=prometheus.yml` (make sure the yml file is copied).

If successful, on the terminal instance running the ebpf profiler app, you should periodically see HTTP 1.1 requests that return 200.

If you want to manually validate that metrics are being exposed/emitted at the endpoint, run `curl http://localhost:9000/metrics`. 

