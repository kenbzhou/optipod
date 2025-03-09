import time
import requests
from bcc import BPF
from datetime import datetime
from ctypes import Structure, c_ulonglong, c_ulong

LOG_IN_TERMINAL: bool = True

with open("profiler-backend.bpf.c", "r") as f:
    profiler_program = f.read()

# orchestrator pod endpoint (reachable within the Kubernetes cluster IF YAML IS SET UP CORRECTLY)
ORCHESTRATOR_URL = "http://orchestrator-service.default.svc.cluster.local:5000/update_metrics"

# Profiler code for eBPF
# Measured metrics: page faults, mem allocations, context switches (graceful, forced), filesystem read/write counts and sizes.
class EBPF_Profiler:
   def __init__(self, node_id="TEST_ID_PLACEHOLDER"):
      # Initialize BPF
      self.profiler = BPF(text=profiler_program)

      # Attach kprobes
      self.profiler.attach_kprobe(event="finish_task_switch.isra.0", fn_name="trace_ctx_switches")
      self.profiler.attach_kprobe(event="handle_mm_fault",           fn_name="trace_page_faults")
      self.profiler.attach_kprobe(event="__kmalloc",                 fn_name="trace_memory_allocation")
      self.profiler.attach_kprobe(event="vfs_read",                  fn_name="trace_fs_read")
      self.profiler.attach_kprobe(event="vfs_write",                 fn_name="trace_fs_write")

      # Globally unique node_id
      self.node_id: str = node_id

      print(f"Starting eBPF profiler on {self.node_id}...")

   def run_profiler_loop(self):
      """ Collects profiling data and sends it to the Orchestrator pod. """
      self.print_logging_header()
      
      while True:
         time.sleep(10)
         timestamped_profile = self.profiler.get_table("timestamped_profile")

         # Sort entries by timestamp (if needed)
         sorted_entries = sorted(
            [(key, value) for key, value in timestamped_profile.items()],
            key=lambda x: datetime.strptime(self.decode_timestamp(x[0]), '%H:%M:%S')
         )

         if not sorted_entries:
            continue

         earliest, val_earliest = sorted_entries[0]
         metrics = ProfiledMetrics.from_buffer_copy(val_earliest)

         # send data to orchestrator
         self.send_metrics_to_orchestrator(metrics)
         # Print metrics if configured to do so.
         self.print_last_metric_log(earliest, metrics)
         # Remove latest map entry: saves space and makes printing cleaner.
         del timestamped_profile[earliest]

   def send_metrics_to_orchestrator(self, metrics):
      """ Sends profiling data to the Orchestrator pod. """
      try:
         response = requests.post(ORCHESTRATOR_URL, json={
            "node_id": self.node_id,
            "mem_bytes_allocated": metrics.mem_bytes_allocated,
            "page_faults": metrics.page_faults,
            "ctx_switches_graceful": metrics.ctx_switches_graceful,
            "ctx_switches_forced": metrics.ctx_switches_forced,
            "fs_read_count": metrics.fs_read_count,
            "fs_read_size_kb": metrics.fs_read_size_kb,
            "fs_write_count": metrics.fs_write_count,
            "fs_write_size_kb": metrics.fs_write_size_kb
         })
         response.raise_for_status()  # Raise error for HTTP failure
      except requests.exceptions.RequestException as e:
         print(f"Failed to send metrics to Orchestrator: {e}")

   def decode_timestamp(self, timestamp: c_ulong):
      timestamp_ns = timestamp.value * 10000000000  # 10s
      timestamp_sec = timestamp_ns // 1000000000    # 10s
      return datetime.utcfromtimestamp(timestamp_sec).strftime('%H:%M:%S')

   def print_logging_header(self):
      if LOG_IN_TERMINAL:
         print("\n" + "=" * 100)
         print(f"Profiling start: {datetime.now().strftime('%H:%M:%S')}")
         print("=" * 100)
         print("%-9s %-10s %-10s %-10s %-10s %-10s %-16s %-10s %-16s" % 
               ("TIMESTAMP", "CTX_SW_G", "CTX_SW_F", "MEM_ALLOC", "PAGE_FTS", "FS_READ_CT", "FS_READ_KB", "FS_WRT_CT", "FS_WRT_KB"))

   def print_last_metric_log(self, timestamp, metrics):
      if LOG_IN_TERMINAL:
         print("%-9s %-10s %-10s %-10s %-10s %-10s %-16s %-10s %-16s" % 
               (self.decode_timestamp(timestamp), metrics.ctx_switches_graceful, metrics.ctx_switches_forced, metrics.mem_bytes_allocated, metrics.page_faults, 
                metrics.fs_read_count, metrics.fs_read_size_kb, metrics.fs_write_count, metrics.fs_write_size_kb))

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

if __name__ == '__main__':
   profiler = EBPF_Profiler(node_id="NODE_01")
   profiler.run_profiler_loop()