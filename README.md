# eBPF-Scheduler

## Installation Instructions for eBPF Profiler
To install the eBPF profiler, SSH into your Linux instance and run the following commands:
```
sudo apt-get install linux-tools-common linux-tools-generic linux-tools-$(uname -r)

sudo apt-get install -y bpfcc-tools linux-headers-$(uname -r) python3-bpfcc build-essential linux-tools-common linux-tools-generic python3-pip git
```

To run the eBPF profiler, run `sudo python3 ebpf-profiler.py`.
