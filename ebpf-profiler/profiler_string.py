profiler_program: str = r"""
#include <linux/mm_types.h>
#include <linux/sched.h>
#include <uapi/linux/ptrace.h>

/*
1. eBPF triggers off of events, e.g. when a page fault occurs. Our actual
handlers need a persistent map with some logical key in order to record these
events.
2. We're implementing a scheduler, and this means we need temporal data w.r.t to
profiled resource usage of our instances:
3. Schema:
    - every 10 seconds, we create a new bucket. This bucket will be indexed by
(time_since_instantation / 10 sec)
    - all of our handlers will fetch time_since_instantiation when an event
triggers them in order to recognize which bucket to dump data into, incrementing
        cpu_time, page_faults, cache misses (?), context switches, etc.
    - In our scheduler, we would simply fetch the second-to-last bucket if
available if we want a simpler scheme.

Alternatively, we can have higher granularity, e.g. 1 second per bucket. That
way, if we so choose, some higher-level overseer can choose to aggregate across
a determined time interval. The schema would still be the same.
*/

struct profiled_metrics {
  u64 mem_bytes_allocated;    // Memory bytes
  u64 page_faults;            // Page faults
  u64 ctx_switches_graceful;  // Context switches, unforced.
  u64 ctx_switches_forced;    // Context switches, forced
  u64 fs_read_count;          // Raw count of reads.
  u64 fs_read_size_kb;        // Raw KB read
  u64 fs_write_count;         // Raw count of writes.
  u64 fs_write_size_kb;       // Raw KB written
};

// Output map
BPF_HASH(timestamped_profile, u64, struct profiled_metrics);

// Configurable timebucket function to fetch
static inline u64 fetch_time_bucket() {
  // Configurable granularity for time bucket feature.
  u64 TIMEBUCKET_INTERVAL = 10000000000;  // 1e9 = 1s
  u64 curr_time = bpf_ktime_get_ns();
  return curr_time / TIMEBUCKET_INTERVAL;
}

// Context Switches
int trace_ctx_switches(struct pt_regs *ctx, struct task_struct *prev) {
  u64 key = fetch_time_bucket();
  struct profiled_metrics *data = timestamped_profile.lookup(&key);
  if (!data) {
    struct profiled_metrics new_data = {};
    // TASK_RUNNING --> Forced
    if (prev->__state == TASK_RUNNING) {
      new_data.ctx_switches_forced = 1;
      new_data.ctx_switches_graceful = 0;
    } else {
      new_data.ctx_switches_forced = 0;
      new_data.ctx_switches_graceful = 1;
    }
    timestamped_profile.update(&key, &new_data);
  } else {
    // TASK_RUNNING --> Forced
    if (prev->__state == TASK_RUNNING) {
      data->ctx_switches_forced += 1;
    } else {
      data->ctx_switches_graceful += 1;
    }
    timestamped_profile.update(&key, data);
  }
  return 0;
}

// Page fault tracking
int trace_page_faults(struct pt_regs *ctx) {
  u64 key = fetch_time_bucket();
  struct profiled_metrics *data = timestamped_profile.lookup(&key);
  if (!data) {
    struct profiled_metrics new_data = {};
    new_data.page_faults = 1;
    timestamped_profile.update(&key, &new_data);
  } else {
    data->page_faults += 1;
    timestamped_profile.update(&key, data);
  }
  return 0;
}

// Memory usage counts
// malloc
int trace_memory_allocation(struct pt_regs *ctx, size_t size) {
  u64 key = fetch_time_bucket();
  struct profiled_metrics *data = timestamped_profile.lookup(&key);

  if (!data) {
    struct profiled_metrics new_data = {};
    new_data.mem_bytes_allocated = size;
    timestamped_profile.update(&key, &new_data);
  } else {
    data->mem_bytes_allocated += size;
    timestamped_profile.update(&key, data);
  }
  return 0;
}

// FS Read
int trace_fs_read(struct pt_regs *ctx, struct file *file, char __user *buf,
                  size_t count) {
  u64 key = fetch_time_bucket();
  struct profiled_metrics *data = timestamped_profile.lookup(&key);
  if (!data) {
    struct profiled_metrics new_data = {};
    new_data.fs_read_count = 1;
    new_data.fs_read_size_kb = count / 1000;
    timestamped_profile.update(&key, &new_data);
  } else {
    data->fs_read_count += 1;
    data->fs_read_size_kb += (count / 1000);
    timestamped_profile.update(&key, data);
  }
  return 0;
}

// FS Write
int trace_fs_write(struct pt_regs *ctx, struct file *file, char __user *buf,
                   size_t count) {
  u64 key = fetch_time_bucket();
  struct profiled_metrics *data = timestamped_profile.lookup(&key);
  if (!data) {
    struct profiled_metrics new_data = {};
    new_data.fs_write_count = 1;
    new_data.fs_write_size_kb = count / 1000;
    timestamped_profile.update(&key, &new_data);
  } else {
    data->fs_write_count += 1;
    data->fs_write_size_kb += (count / 1000);
    timestamped_profile.update(&key, data);
  }
  return 0;
}
"""