import os
import multiprocessing as mp
from flask import Flask, request, jsonify, send_from_directory
import time

from utils.config_loader import load_config
from engine.trade_engine import TradeEngine
from schwab_simulation.ledger_service import OrderEvent

app = Flask(__name__, static_folder='dashboard')
config = load_config()

# Global reference to engine, set before app.run
engine = None

@app.route('/health', methods=['GET'])
def health():
    status = "ok" if engine and getattr(engine, 'running', False) else "error"
    return jsonify({"status": status, "timestamp": time.time()})

@app.route('/positions', methods=['GET'])
def positions():
    if engine:
        return jsonify(engine.state_dict.get('positions', {}))
    return jsonify({})

@app.route('/telemetry', methods=['GET'])
def telemetry():
    if engine:
        return jsonify({
            'cpu_percent': engine.state_dict.get('cpu_percent', 0.0),
            'engine_latency_ms': engine.state_dict.get('engine_latency_ms', 0.0),
            'last_order_latency_ms': engine.state_dict.get('last_order_latency_ms', 0.0)
        })
    return jsonify({})

@app.route('/config', methods=['GET'])
def get_config():
    return jsonify(config)

@app.route('/', defaults={"path": "index.html"})
@app.route('/<path:path>')
def serve_dashboard(path):
    return send_from_directory(app.static_folder, path)

if __name__ == '__main__':
    mp.freeze_support()
    
    # Initialize and start the Trade Engine
    engine = TradeEngine()
    
    # Start engine in a separate thread so Flask can run in the main thread
    # Wait, TradeEngine.start() already runs its loop in a blocking manner or thread?
    # Actually, in the latest trade_engine.py, start() is blocking because it calls _run_loop directly.
    # Let's change trade_engine.py to run _run_loop in a thread, or run Flask in a thread.
    import threading
    engine_thread = threading.Thread(target=engine.start, daemon=True)
    engine_thread.start()
    
    port = int(os.getenv('PORT', 5000))
    app.run(host='0.0.0.0', port=port, debug=False, use_reloader=False)
