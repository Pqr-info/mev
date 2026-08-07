package main

import (
	"errors"
	"time"
)

type SchwabTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type SchwabAccount struct {
	AccountNumber string  `json:"accountNumber"`
	AccountStatus string  `json:"accountStatus"`
	CashBalance   float64 `json:"cashBalance"`
	NetWorth      float64 `json:"netWorth"`
}

type SchwabOrder struct {
	OrderID     string    `json:"orderId"`
	Symbol      string    `json:"symbol"`
	Quantity    float64   `json:"quantity"`
	OrderType   string    `json:"orderType"` // LIMIT, MARKET
	Price       float64   `json:"price"`
	Instruction string    `json:"instruction"` // BUY, SELL
	Status      string    `json:"status"`      // WORKING, FILLED, REJECTED
	Timestamp   time.Time `json:"timestamp"`
}

type SchwabSimulatorClient struct {
	appKey    string
	appSecret string
	token     *SchwabTokenResponse
	accounts  map[string]SchwabAccount
	orders    map[string]SchwabOrder
}

func NewSchwabSimulatorClient(appKey, appSecret string) *SchwabSimulatorClient {
	accounts := make(map[string]SchwabAccount)
	accounts["SCHWAB-MOCK-1"] = SchwabAccount{
		AccountNumber: "SCHWAB-MOCK-1",
		AccountStatus: "NORMAL",
		CashBalance:   100000.0,
		NetWorth:      100000.0,
	}

	return &SchwabSimulatorClient{
		appKey:    appKey,
		appSecret: appSecret,
		accounts:  accounts,
		orders:    make(map[string]SchwabOrder),
	}
}

func (c *SchwabSimulatorClient) Authenticate() (*SchwabTokenResponse, error) {
	c.token = &SchwabTokenResponse{
		AccessToken:  "mock-access-token-schwab-123",
		TokenType:    "Bearer",
		ExpiresIn:    1800,
		RefreshToken: "mock-refresh-token-456",
		Scope:        "readonly trade",
		ExpiresAt:    time.Now().Add(30 * time.Minute),
	}
	return c.token, nil
}

func (c *SchwabSimulatorClient) GetAccount(accountNumber string) (SchwabAccount, error) {
	acc, ok := c.accounts[accountNumber]
	if !ok {
		return SchwabAccount{}, errors.New("account not found")
	}
	return acc, nil
}

func (c *SchwabSimulatorClient) PlaceOrder(accountNumber string, symbol string, qty float64, price float64, instruction string) (SchwabOrder, error) {
	acc, ok := c.accounts[accountNumber]
	if !ok {
		return SchwabOrder{}, errors.New("invalid account number")
	}

	cost := qty * price
	if instruction == "BUY" && acc.CashBalance < cost {
		return SchwabOrder{}, errors.New("insufficient buying power")
	}

	if instruction == "BUY" {
		acc.CashBalance -= cost
	} else if instruction == "SELL" {
		acc.CashBalance += cost
	}
	c.accounts[accountNumber] = acc

	order := SchwabOrder{
		OrderID:     "order-" + symbol + "-" + time.Now().Format("150405"),
		Symbol:      symbol,
		Quantity:    qty,
		OrderType:   "LIMIT",
		Price:       price,
		Instruction: instruction,
		Status:      "FILLED",
		Timestamp:   time.Now(),
	}

	c.orders[order.OrderID] = order
	return order, nil
}
