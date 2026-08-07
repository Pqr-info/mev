package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Substrate27Exchange simulates an AMM or direct routing DEX
// for swapping USDC and Substrate27 (SUB27) tokens.
type Substrate27Exchange struct {
	// In production, this would hold pointers to RPC clients for both
	// the Substrate node and perhaps an Ethereum/Arbitrum node for USDC.
	Generator *LiquidityGenerator
}

func NewSubstrate27Exchange(generator *LiquidityGenerator) *Substrate27Exchange {
	return &Substrate27Exchange{
		Generator: generator,
	}
}

type QuoteRequest struct {
	FromToken string  `json:"from_token"` // e.g. "USDC", "SUB27"
	ToToken   string  `json:"to_token"`
	Amount    float64 `json:"amount"`
}

type QuoteResponse struct {
	EstimatedOutput float64 `json:"estimated_output"`
	ExchangeRate    float64 `json:"exchange_rate"`
	Route           string  `json:"route"`
}

// HandleQuote returns an estimated exchange rate.
func (dex *Substrate27Exchange) HandleQuote(w http.ResponseWriter, r *http.Request) {
	// For simulation, fixed exchange rate: 1 SUB27 = 2.50 USDC
	rateSub27ToUSDC := 2.50
	
	from := r.URL.Query().Get("from_token")
	to := r.URL.Query().Get("to_token")
	amountStr := r.URL.Query().Get("amount")
	
	var amount float64
	fmt.Sscanf(amountStr, "%f", &amount)

	var output float64
	var rate float64
	if from == "SUB27" && to == "USDC" {
		output = amount * rateSub27ToUSDC
		rate = rateSub27ToUSDC
	} else if from == "USDC" && to == "SUB27" {
		output = amount / rateSub27ToUSDC
		rate = 1.0 / rateSub27ToUSDC
	} else {
		http.Error(w, "Unsupported token pair", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(QuoteResponse{
		EstimatedOutput: output,
		ExchangeRate:    rate,
		Route:           "Substrate27-AMM",
	})
}

type SwapRequest struct {
	FromToken   string  `json:"from_token"`
	ToToken     string  `json:"to_token"`
	Amount      float64 `json:"amount"`
	UserAddress string  `json:"user_address"`
}

// HandleSwap executes the swap and routes the volume to the Liquidity Generator.
func (dex *Substrate27Exchange) HandleSwap(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SwapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Trigger arbitrage generation signal
	if dex.Generator != nil {
		go dex.Generator.OnSwapEvent(req.FromToken, req.ToToken, req.Amount, req.UserAddress)
	}

	txHash := fmt.Sprintf("0xswap%d", time.Now().UnixNano())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "executed",
		"tx_hash": txHash,
		"swapped": req.Amount,
		"token":   req.FromToken,
	})
}
