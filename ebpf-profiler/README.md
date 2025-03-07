Compilation:
 clang -target bpf -Wall -O2 -g -c HelloWorld.bpf.c -o HelloWorld.o
 gcc -o loader loader.c -lbpf

Load:
sudo bpftool prog load HelloWorld.o /sys/fs/bpf/helloworld
sudo bpftool prog attach pinned /sys/fs/bpf/helloworld tracepoint syscalls sys_enter_execve
sudo bpftool prog show
