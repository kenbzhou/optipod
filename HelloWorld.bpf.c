#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

SEC("tracepoint/syscalls/sys_enter_execve")  // Attach to execve syscall
int helloworld(void *ctx)
{
 bpf_printk("Hello world!\n");
 return 0;
}
char LICENSE[] SEC("license") = "GPL";
