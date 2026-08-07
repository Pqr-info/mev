import time
import random
import pandas as pd
import multiprocessing as mp
import concurrent.futures
import inspect

from utils.config_loader import load_config
from schwab_simulation.ledger_service import LedgerService, OrderEvent
from algos import get_algorithms
from inference.server import InferenceWorker, InferenceClient
import psutil

def run_algo_worker(algo, market_data, req_queue, res_queue):
    """
    Top-level function for ProcessPoolExecutor.
    Runs the algorithm. Initializes a local InferenceClient if needed.
    """
    try:
        sig = inspect.signature(algo.generate_signals)
        if 'inference_client' in sig.parameters:
            inference_client = InferenceClient(req_queue, res_queue)
            return algo.generate_signals(market_data, inference_client=inference_client)
        else:
            return algo.generate_signals(market_data)
    except Exception as e:
        print(f"[Worker] Error in {algo.__class__.__name__}: {e}")
        return []

class TradeEngine:
    """
    The main orchestrator. Owns the ledger process, inference process, 
    and the worker pool for maximum Ryzen 9 AI throughput.
    """
    def __init__(self):
        self.config = load_config()
        self.assets = self.config.get('assets', [])
        self.algos = get_algorithms()
        self.running = False
        
        self.manager = mp.Manager()
        self.state_dict = self.manager.dict()
        self.state_dict['positions'] = {}
        
        # Queues
        self.inference_req_q = self.manager.Queue()
        self.inference_res_q = self.manager.Queue()
        self.ledger_order_q = self.manager.Queue()
        
        # Services
        self.inference_worker = InferenceWorker(self.inference_req_q, self.inference_res_q)
        self.ledger_service = LedgerService(self.ledger_order_q, self.state_dict)

    def _generate_market_data(self) -> dict:
        data = {}
        for symbol in self.assets:
            prices = [100 + random.gauss(0, 1) for _ in range(30)]
            df = pd.DataFrame({"close": prices})
            data[symbol] = df
        return data

    def start(self):
        if not self.running:
            self.running = True
            self.ledger_service.start()
            self.inference_worker.start()
            print("[Engine] Services started. Running core loop...")
            self._run_loop()

    def _run_loop(self):
        # Process pool for CPU-bound algorithm crunching
        with concurrent.futures.ProcessPoolExecutor(max_workers=len(self.algos)) as executor:
            while self.running:
                start_t = time.time()
                market_data = self._generate_market_data()
                
                # Update telemetry
                self.state_dict['cpu_percent'] = psutil.cpu_percent()
                
                futures = [
                    executor.submit(run_algo_worker, algo, market_data, self.inference_req_q, self.inference_res_q)
                    for algo in self.algos
                ]
                
                for future in concurrent.futures.as_completed(futures):
                    signals = future.result()
                    for sig in signals:
                        if isinstance(sig, OrderEvent):
                            self.ledger_order_q.put(sig)
                        elif isinstance(sig, dict):
                            # Convert dict to OrderEvent
                            evt = OrderEvent(
                                order_id=str(random.randint(1_000_000, 9_999_999)),
                                symbol=sig['symbol'],
                                side=sig['side'],
                                quantity=sig['quantity'],
                                price=sig['price'],
                                timestamp=time.time()
                            )
                            if sig.get('broker', 'schwab') == 'schwab':
                                self.ledger_order_q.put(evt)
                            else:
                                print(f"[Alpaca stub] Would submit order: {sig}")

                elapsed = time.time() - start_t
                self.state_dict['engine_latency_ms'] = round(elapsed * 1000, 2)
                time.sleep(5)

    def stop(self):
        self.running = False
        self.inference_req_q.put(None)
        self.ledger_order_q.put(None)
        self.inference_worker.join()
        self.ledger_service.join()
        print("[Engine] Shutdown complete.")

if __name__ == '__main__':
    mp.freeze_support()
    engine = TradeEngine()
    try:
        engine.start()
    except KeyboardInterrupt:
        engine.stop()
