#include <stdio.h>
#include <stdlib.h>
#include <bpf/libbpf.h>
#include <bpf/bpf.h>
#include <unistd.h>

#define CATEGORY "syscalls"
#define TRACEPOINT "sys_enter_execve"

int main() {
    struct bpf_object *obj;
    int prog_fd;
    struct bpf_program *prog;

    obj = bpf_object__open_file("HelloWorld.o", NULL);

    bpf_object__load(obj);

    prog = bpf_object__find_program_by_name(obj, "helloworld");

    prog_fd = bpf_program__fd(prog);

    struct bpf_link *link = bpf_program__attach_tracepoint(prog, CATEGORY, TRACEPOINT);

    printf("eBPF program successfully loaded and attached to '%s:%s'\n", CATEGORY, TRACEPOINT);

    while (1) {
        sleep(5);
    }

    return 0;
}
