import torch
import torch.nn as nn

class LSTMPredictor(nn.Module):
    """
    A simple LSTM model for time-series forecasting or signal classification.
    Designed to take sequence data (e.g., past 30 time steps of prices/volumes).
    """
    def __init__(self, input_size=1, hidden_size=64, num_layers=2, output_size=1):
        super(LSTMPredictor, self).__init__()
        self.hidden_size = hidden_size
        self.num_layers = num_layers
        
        # LSTM layer
        self.lstm = nn.LSTM(input_size, hidden_size, num_layers, batch_first=True)
        
        # Fully connected output layer
        self.fc = nn.Linear(hidden_size, output_size)

    def forward(self, x):
        # Initialize hidden state and cell state
        h0 = torch.zeros(self.num_layers, x.size(0), self.hidden_size).to(x.device)
        c0 = torch.zeros(self.num_layers, x.size(0), self.hidden_size).to(x.device)
        
        # Forward propagate LSTM
        out, _ = self.lstm(x, (h0, c0))
        
        # Decode the hidden state of the last time step
        out = self.fc(out[:, -1, :])
        return out

def get_dummy_model():
    """Returns an initialized model with random weights for testing."""
    model = LSTMPredictor()
    model.eval() # Set to evaluation mode
    return model
