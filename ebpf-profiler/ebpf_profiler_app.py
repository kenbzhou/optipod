import argparse
from threading import Thread
from flask import Flask, Response
from ebpf_profiler import EBPF_Profiler

app = Flask(__name__)

parser = argparse.ArgumentParser()
parser.add_argument("--host", type=str, default="0.0.0.0", help="Flask server host (default: 0.0.0.0)")
parser.add_argument("--port", type=int, default=9000, help="Flask server port (default: 9000)")
parser.add_argument("--node_id", type=str, default="TEST", help="Instantiated profiler node id (default: TEST)")
args = parser.parse_args()

if __name__ == '__main__':
    # Flask is single-threaded and blocking by default. Multithread to prevent this.
    flask_thread = Thread(target=lambda: app.run(host=args.host, port=args.port))
    flask_thread.daemon = True
    flask_thread.start()

    # Start and run an instance of the eBPF Profiler
    # Prometheus will scrape the port that flask runs on and call the metrics function of ebpf_profiler
    profiler = EBPF_Profiler(args.node_id)
    profiler.run_profiler_loop()