package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type AlpacaRecord struct {
	EventID     string
	PayloadType string
	Risk        float64
	Stability   string
	Epoch       string
	Timestamp   time.Time
}

type AlpacaFirehose struct {
	keyID     string
	secretKey string
	baseURL   string
	client    *http.Client
}

func NewAlpacaFirehose() *AlpacaFirehose {
	keyID := os.Getenv("ALPACA_API_KEY_ID")
	secretKey := os.Getenv("ALPACA_API_SECRET_KEY")
	baseURL := os.Getenv("ALPACA_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://paper-api.alpaca.markets" // default paper trading url
	}

	return &AlpacaFirehose{
		keyID:     keyID,
		secretKey: secretKey,
		baseURL:   baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Emit sends simulated trade allocations directly to the Alpaca Paper Trading REST API.
func (a *AlpacaFirehose) Emit(record AlpacaRecord) error {
	// If credentials are not set, log and proceed with local mock success to prevent errors
	if a.keyID == "" || a.secretKey == "" {
		fmt.Printf("[ALPACA-FIREHOSE] Mock Emit completed for event %s. (Alpaca API credentials missing)\n", record.EventID)
		return nil
	}

	// Trigger a paper trade when stability drops below threshold (e.g. drift-risk)
	action := "buy"
	if record.Risk > 0.5 {
		action = "sell" // liquidate/hedge position on risk alerts
	}

	orderPayload := map[string]interface{}{
		"symbol":        "TXN",
		"qty":           "10",
		"side":          action,
		"type":          "market",
		"time_in_force": "day",
	}

	jsonBytes, err := json.Marshal(orderPayload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v2/orders", a.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("APCA-API-KEY-ID", a.keyID)
	req.Header.Set("APCA-API-SECRET-KEY", a.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("alpaca order placement failed: %s - %s", resp.Status, string(bodyBytes))
	}

	fmt.Printf("[ALPACA-FIREHOSE] Successfully placed paper order on Alpaca for symbol: TXN (Action: %s)\n", action)
	return nil
}

// ConnectWebSocket is a skeleton showing connection to the Alpaca Streaming updates.
func (a *AlpacaFirehose) ConnectWebSocket() error {
	fmt.Println("[ALPACA-WEBSOCKET] Connecting to Alpaca paper stream at wss://paper-api.alpaca.markets/stream ...")
	// In production, we would use a websocket package (like gorilla/websocket) to connect
	// and register the auth message using the API credentials.
	return nil
}
