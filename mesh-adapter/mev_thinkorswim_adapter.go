package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SchwabAccountSummary defines account net worth, cash, and equity states.
type SchwabAccountSummary struct {
	AccountID        string    `json:"account_id"`
	CashBalance      float64   `json:"cash_balance"`
	TotalEquityValue float64   `json:"total_equity_value"`
	AccountNetWorth  float64   `json:"account_net_worth"`
	DayChangePercent float64   `json:"day_change_percent"`
	Timestamp        time.Time `json:"timestamp"`
}

// SchwabPosition represents a single active stock position.
type SchwabPosition struct {
	Symbol              string  `json:"symbol"`
	Quantity            float64 `json:"quantity"`
	CurrentPrice        float64 `json:"current_price"`
	MarketValue         float64 `json:"market_value"`
	CostBasis           float64 `json:"cost_basis"`
	UnrealizedPL        float64 `json:"unrealized_pl"`
	UnrealizedPLPercent float64 `json:"unrealized_pl_percent"`
}

// SchwabOrder represents a placed or pending order.
type SchwabOrder struct {
	OrderID   string    `json:"order_id"`
	Symbol    string    `json:"symbol"`
	Qty       float64   `json:"qty"`
	Action    string    `json:"action"` // "BUY" or "SELL"
	Type      string    `json:"type"`   // "MARKET" or "LIMIT"
	Price     float64   `json:"price"`
	Status    string    `json:"status"` // "PENDING" or "FILLED"
	Timestamp time.Time `json:"timestamp"`
}

// SchwabAdapterClient coordinates communication with the local Schwab Paper API server.
type SchwabAdapterClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewSchwabAdapterClient constructs a client pointing to the paper API server.
func NewSchwabAdapterClient(baseURL string) *SchwabAdapterClient {
	if baseURL == "" {
		baseURL = "http://localhost:8085/api/v1"
	}
	return &SchwabAdapterClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetSummary retrieves current account balances.
func (c *SchwabAdapterClient) GetSummary() (*SchwabAccountSummary, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/accounts/summary", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var summary SchwabAccountSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}
	return &summary, nil
}

// GetPositions retrieves active positions.
func (c *SchwabAdapterClient) GetPositions() ([]SchwabPosition, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/accounts/positions", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var positions []SchwabPosition
	if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
		return nil, err
	}
	return positions, nil
}

// PlaceOrder submits a BUY or SELL order to the mock server.
func (c *SchwabAdapterClient) PlaceOrder(symbol string, qty float64, action string, orderType string, price float64) (*SchwabOrder, error) {
	payload := map[string]interface{}{
		"symbol": symbol,
		"qty":    qty,
		"action": action,
		"type":   orderType,
	}
	if orderType == "LIMIT" {
		payload["price"] = price
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/orders/place", c.baseURL),
		"application/json",
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("order failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var result struct {
		Status  string      `json:"status"`
		Message string      `json:"message"`
		Order   SchwabOrder `json:"order"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	return &result.Order, nil
}

// CancelOrder sends a cancellation request for a pending order.
func (c *SchwabAdapterClient) CancelOrder(orderID string) error {
	payload := map[string]string{
		"order_id": orderID,
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/orders/cancel", c.baseURL),
		"application/json",
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cancellation failed: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}
