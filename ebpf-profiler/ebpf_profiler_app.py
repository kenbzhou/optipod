import argparse
from threading import Thread
from flask import Flask, Response
from prometheus_client import generate_latest
from ebpf_profiler import EBPF_Profiler

app = Flask(__name__)

parser = argparse.ArgumentParser()
parser.add_argument("--host", type=str, default="0.0.0.0", help="Flask server host (default: 0.0.0.0)")
parser.add_argument("--port", type=int, default=9000, help="Flask server port (default: 9000)")
parser.add_argument("--node_id", type=str, default="TEST", help="Instantiated profiler node id (default: TEST)")
args = parser.parse_args()

# How it works: prometheus will ping the {endpoint of the flask app}/metrics every {interval} seconds for the latest metrics recorded.
# the profiler will populate these metrics, so when the prometheus client pings, the latest updated metrics are returned.
@app.route('/metrics')
def return_metrics():
    return Response(generate_latest(), mimetype='text/plain')

if __name__ == '__main__':
    # Flask is single-threaded and blocking by default. Multithread to prevent this.
    flask_thread = Thread(target=lambda: app.run(host=args.host, port=args.port))
    flask_thread.daemon = True
    flask_thread.start()

    # Start and run an instance of the eBPF Profiler
    # Prometheus will scrape the port that flask runs on and call the metrics function of ebpf_profiler
    profiler = EBPF_Profiler(node_id=args.node_id)
    profiler.run_profiler_loop()