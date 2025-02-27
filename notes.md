Instances: 
(check email)

EC2 Setup:
eBPF:
```shell
sudo apt install build-essential sudo netcat git bpfcc-tools linux-headers-$(uname -r) emacs-nox clang llvm libbpf-dev libelf-dev gcc-multilib
export PATH=/usr/sbin:$PATH
	
# sanity check, below should run
sudo opensnoop-bpfcc

# rust setup
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
export PATH=/$HOME/.cargo/bin:$PATH
```


eBPF Tutorial That is Working:
- Follow [this](https://medium.com/@matrixorigin-database/bpf-development-starting-with-hello-world-c309941d6b3f)



Notes: 
- [scoring plugins/configurations](https://kubernetes.io/docs/reference/scheduling/config/) are used to change the default scheduler's scoring logic
	- the 6th extension point says: `score: These plugins provide a score to each node that has passed the filtering phase. The scheduler will then select the node with the highest weighted scores sum.`
 - [scheduling points](https://kubernetes.io/docs/reference/scheduling/config/#scheduling-plugins) also has some good info about current capability of scoring
 - How can plugins be created? 
	 - you can write in Go then hook into k8s
	 - some relevant examples can be found [here](https://github.com/kubernetes-sigs/scheduler-plugins/tree/master/pkg/capacityscheduling)
 - a good [ebpf tutorial](https://eunomia.dev/tutorials/0-introduce/) 
 - some installation steps from a google [blog](https://android.googlesource.com/platform/external/bcc/+/6954e2577b2be74b2dbfcd99784dc0b43ad662ef/INSTALL.md#packages)
 - 



