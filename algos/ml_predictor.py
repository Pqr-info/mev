from . import register
import pandas as pd
import numpy as np
import time
import random

try:
    from schwab_simulation.ledger_service import OrderEvent
except ImportError:
    pass

@register('ml_predictor')
class MLPredictorStrategy:
    """
    Stateless algorithm. Uses InferenceClient to get signals from ONNX runtime.
    """
    def __init__(self, model_name: str = "lstm_dummy", seq_len: int = 30, max_qty: int = 15):
        self.model_name = model_name
        self.seq_len = seq_len
        self.max_qty = max_qty

    def generate_signals(self, market_data: dict, inference_client) -> list:
        if inference_client is None:
            return []

        orders = []
        for symbol, df in market_data.items():
            if len(df) < self.seq_len:
                continue
            
            # Extract features as numpy array
            recent_prices = df['close'].iloc[-self.seq_len:].to_numpy()
            base_price = recent_prices[0]
            if base_price == 0:
                continue
            normalized_features = recent_prices / base_price
            
            # Query the inference service
            prediction = inference_client.predict(self.model_name, normalized_features)
            
            current_price = float(df['close'].iloc[-1])
            
            # Simple trading logic based on model prediction
            if prediction > 0.1:
                orders.append(OrderEvent(
                    order_id=str(random.randint(1_000_000, 9_999_999)),
                    symbol=symbol,
                    side='buy',
                    quantity=self.max_qty,
                    price=current_price,
                    timestamp=time.time()
                ))
            elif prediction < -0.1:
                orders.append(OrderEvent(
                    order_id=str(random.randint(1_000_000, 9_999_999)),
                    symbol=symbol,
                    side='sell',
                    quantity=self.max_qty,
                    price=current_price,
                    timestamp=time.time()
                ))
                
        return orders
