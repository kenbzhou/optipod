#!/usr/bin/python3
from bcc import BPF
import time
from datetime import datetime
from profiler_string import profiler_program
from ctypes import Structure, c_ulonglong, c_int, c_ulong


# Initialize BPF
b = BPF(text=profiler_program)

# Attach kprobes
b.attach_kprobe(event="finish_task_switch.isra.0", fn_name="trace_ctx_switches")
b.attach_kprobe(event="handle_mm_fault", fn_name="trace_page_faults")
b.attach_kprobe(event="__kmalloc", fn_name="trace_memory_allocation")
b.attach_kprobe(event="vfs_read", fn_name="trace_fs_read")
b.attach_kprobe(event="vfs_write", fn_name="trace_fs_write")

# b.attach_kprobe(event="kfree", fn_name="trace_memory_deallocation")

print(f"Starting profiler...")

class ProfiledMetrics(Structure):
    _fields_ = [
        # Memory-related allocations
        ("mem_bytes_allocated",     c_ulonglong),
        ("page_faults",             c_ulonglong),
        # Context-switches/interrupts
        ("ctx_switches_graceful",   c_ulonglong),
        ("ctx_switches_forced",     c_ulonglong),
        # Filesystem I/O
        ("fs_read_count",           c_ulonglong),
        ("fs_read_size_kb",         c_ulonglong),
        ("fs_write_count",          c_ulonglong),
        ("fs_write_size_kb",        c_ulonglong),
    ]

def decode_timestamp(timestamp: c_ulong):
    timestamp_ns = timestamp.value * 10000000000 # 10s
    timestamp_sec = timestamp_ns // 1000000000   # 10s
    return datetime.utcfromtimestamp(timestamp_sec).strftime('%H:%M:%S')


try:
    while True:
        time.sleep(10)
        print("\n" + "=" * 80)
        print(f"System Profile at {datetime.now().strftime('%H:%M:%S')}")
        print("=" * 80)
        
        # Process CPU metrics
        timestamped_profile = b.get_table("timestamped_profile")
        sorted_entries = sorted([(key, value) for key, value in timestamped_profile.items()], key=lambda x: datetime.strptime(decode_timestamp(x[0]), '%H:%M:%S'))
        print("%-9s %-10s %-10s %-10s %-10s %-10s %-16s %-10s %-16s" % ("TIMESTAMP", "CTX_SW_G", "CTX_SW_F", "MEM_ALLOC", "PAGE_FTS", "FS_READ_CT", "FS_READ_KB", "FS_WRT_CT", "FS_WRT_KB"))
        for key, value in sorted_entries:
            metrics = ProfiledMetrics.from_buffer_copy(value)
            print("%-9s %-10s %-10s %-10s %-10s %-10s %-16s %-10s %-16s" % (decode_timestamp(key), metrics.ctx_switches_graceful, metrics.ctx_switches_forced, metrics.mem_bytes_allocated, metrics.page_faults, 
                                                                            metrics.fs_read_count, metrics.fs_read_size_kb, metrics.fs_write_count, metrics.fs_write_size_kb))


except:
   print("Some fault occurred")