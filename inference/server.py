import multiprocessing as mp
import onnxruntime as ort
import numpy as np
import time

class InferenceWorker(mp.Process):
    """
    Dedicated background process for running ML inference via ONNX Runtime.
    Leverages AMD Ryzen AI NPU if configured via appropriate Execution Providers.
    """
    def __init__(self, request_queue: mp.Queue, response_queue: mp.Queue):
        super().__init__()
        self.request_queue = request_queue
        self.response_queue = response_queue
        self.daemon = True

    def run(self):
        print("[Inference Worker] Starting up...")
        
        # In a real environment, we would load the compiled model:
        # providers = ['VitisAIExecutionProvider', 'CPUExecutionProvider'] # AMD Ryzen AI
        # self.session = ort.InferenceSession('models/compiled/lstm_dummy.onnx', providers=providers)
        
        # For now, we simulate the ONNX Runtime loading and prediction
        print("[Inference Worker] Dummy ONNX session loaded. Listening for requests...")

        while True:
            try:
                request = self.request_queue.get()
                if request is None:
                    break
                
                req_id, model_name, features = request
                
                # Simulate ONNX inference execution
                # output = self.session.run(None, {'input': features})[0]
                
                # Dummy prediction logic for validation: 
                # Returns a simple sum-based score normalized to [-1, 1]
                dummy_pred = float(np.sum(features)) % 2.0 - 1.0
                
                self.response_queue.put((req_id, dummy_pred))
                
            except Exception as e:
                print(f"[Inference Worker] Error processing request: {e}")
                
        print("[Inference Worker] Shutting down.")

class InferenceClient:
    """
    Synchronous-looking wrapper used by algorithm workers to call the InferenceWorker.
    """
    def __init__(self, request_queue: mp.Queue, response_queue: mp.Queue):
        self.request_queue = request_queue
        self.response_queue = response_queue
        # Shared counter in memory for this specific client (since each worker gets its own Client instance via fork/pickle)
        # However, if shared across processes, we must ensure unique request IDs.
        self._client_id = mp.current_process().name
        self._req_counter = 0

    def predict(self, model_name: str, features: np.ndarray, timeout: float = 1.0) -> float:
        """Sends features to the ONNX worker and blocks until prediction is returned."""
        self._req_counter += 1
        req_id = f"{self._client_id}_{self._req_counter}"
        
        self.request_queue.put((req_id, model_name, features))
        
        try:
            # Note: For production with multiple clients sharing one response queue, 
            # we need a routing layer. For this architecture validation, we assume
            # simple one-to-one or retry logic.
            resp_id, prediction = self.response_queue.get(timeout=timeout)
            if resp_id == req_id:
                return prediction
            else:
                self.response_queue.put((resp_id, prediction))
                return 0.0
        except mp.queues.Empty:
            print(f"[Inference Client] Timeout waiting for prediction {req_id}")
            return 0.0
