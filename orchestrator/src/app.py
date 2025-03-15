import argparse
from flask import Flask, Response, request, jsonify
from prometheus_client import generate_latest, CollectorRegistry, Gauge

# flask init
app = Flask(__name__)

# prometheus init
registry = CollectorRegistry()

# prometheus metrics
mem_bytes_allocated = Gauge("mem_bytes_allocated", "Memory bytes allocated", ["node_id"], registry=registry)
page_faults = Gauge("page_faults", "Page faults recorded", ["node_id"], registry=registry)
ctx_switches_graceful = Gauge("ctx_switches_graceful", "Graceful context switches", ["node_id"], registry=registry)
ctx_switches_forced = Gauge("ctx_switches_forced", "Forced context switches", ["node_id"], registry=registry)
fs_read_count = Gauge("fs_read_count", "File system read calls", ["node_id"], registry=registry)
fs_read_size_kb = Gauge("fs_read_size_kb", "KB read from file system", ["node_id"], registry=registry)
fs_write_count = Gauge("fs_write_count", "File system write calls", ["node_id"], registry=registry)
fs_write_size_kb = Gauge("fs_write_size_kb", "KB written to file system", ["node_id"], registry=registry)

@app.route('/metrics')
def return_metrics():
    """ Exposes Prometheus metrics. """
    return Response(generate_latest(registry), mimetype='text/plain')

@app.route('/query', methods=['GET'])
def query_prometheus():
    """
    Proxy for Prometheus queries
    Example: /query?q=avg_over_time(mem_bytes_allocated[5m])
    Example: /query?q=mem_bytes_allocated{node_id="NODE_01"}&time=1742038979
    """
    query = request.args.get('q')
    time = request.args.get('time')
    start = request.args.get('start')
    end = request.args.get('end')
    step = request.args.get('step', '15s')
    
    if not query:
        return jsonify({"error": "Missing query parameter 'q'"}), 400
        
    if start and end:
        # range query
        params = {
            'query': query,
            'start': start,
            'end': end,
            'step': step
        }
        endpoint = "/api/v1/query_range"
    else:
        # instat query
        params = {'query': query}
        if time:
            params['time'] = time
        endpoint = "/api/v1/query"
    
    try:
        response = requests.get(f"{PROMETHEUS_URL}{endpoint}", params=params)
        return jsonify(response.json()), response.status_code
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/update_metrics', methods=['POST'])
def update_metrics():
    """
    Profiler pods send metrics here.
    Valid JSON format:
    {
        "node_id": "node-1",
        "mem_bytes_allocated": 123456,
        "page_faults": 5,
        "ctx_switches_graceful": 10,
        "ctx_switches_forced": 3,
        "fs_read_count": 15,
        "fs_read_size_kb": 1024,
        "fs_write_count": 8,
        "fs_write_size_kb": 512
    }
    """
    data = request.get_json()
    
    if not data or "node_id" not in data:
        return jsonify({"error": "Invalid request format"}), 400

    node_id = data["node_id"]

    # update prometheus with recieved data
    mem_bytes_allocated.labels(node_id=node_id).set(data.get("mem_bytes_allocated", 0))
    page_faults.labels(node_id=node_id).set(data.get("page_faults", 0))
    ctx_switches_graceful.labels(node_id=node_id).set(data.get("ctx_switches_graceful", 0))
    ctx_switches_forced.labels(node_id=node_id).set(data.get("ctx_switches_forced", 0))
    fs_read_count.labels(node_id=node_id).set(data.get("fs_read_count", 0))
    fs_read_size_kb.labels(node_id=node_id).set(data.get("fs_read_size_kb", 0))
    fs_write_count.labels(node_id=node_id).set(data.get("fs_write_count", 0))
    fs_write_size_kb.labels(node_id=node_id).set(data.get("fs_write_size_kb", 0))

    return jsonify({"status": "Metrics updated"}), 200

@app.route('/nodes', methods=['GET'])
def list_nodes():
    """
    Returns a list of nodes that have reported metrics
    """
    try:
        response = requests.get(f"{PROMETHEUS_URL}/api/v1/series", params={
            'match[]': 'mem_bytes_allocated'
        })
        
        if response.status_code != 200:
            return jsonify({"error": "Error fetching node data from Prometheus"}), 500
            
        data = response.json()
        nodes = []
        
        if data['status'] == 'success' and 'data' in data:
            for series in data['data']:
                if 'node_id' in series:
                    nodes.append(series['node_id'])
                    
        return jsonify({"nodes": nodes})
        
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("--host", type=str, default="0.0.0.0", help="Flask server host")
    parser.add_argument("--port", type=int, default=5000, help="Flask server port")
    args = parser.parse_args()

    app.run(host=args.host, port=args.port)