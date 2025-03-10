FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    clang llvm libelf-dev gcc make iproute2 python3 python3-pip bpfcc-tools linux-headers-6.8.0-1021-aws python3-bpfcc build-essential linux-tools-common linux-tools-generic \
    && pip3 install bcc requests \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY ./src/ebpf-profiler.py ./src/profiler-backend.bpf.c /app/

RUN chmod +x /app/ebpf-profiler.py

CMD ["python3", "/app/ebpf-profiler.py"]


