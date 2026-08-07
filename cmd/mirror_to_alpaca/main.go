package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Mode string `yaml:"mode"`
}

type RecommendedTrade struct {
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"` // "buy" or "sell"
	Qty        int     `json:"qty"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
}

type RecommendedTradesPayload struct {
	Timestamp string             `json:"timestamp"`
	Trades    []RecommendedTrade `json:"trades"`
}

type StagedOrder struct {
	Symbol      string  `json:"symbol"`
	Qty         int     `json:"qty"`
	Side        string  `json:"side"`
	Type        string  `json:"type"`
	TimeInForce string  `json:"time_in_force"`
	Reason      string  `json:"reason"`
	Confidence  float64 `json:"confidence"`
}

type StagedOrdersPayload struct {
	GeneratedAt string        `json:"generated_at"`
	Orders      []StagedOrder `json:"orders"`
}

func main() {
	configBytes, err := os.ReadFile(filepath.Join("config", "paper.yaml"))
	if err != nil {
		fmt.Println("Failed to read config:", err)
		os.Exit(1)
	}

	var cfg Config
	if err := yaml.Unmarshal(configBytes, &cfg); err != nil {
		fmt.Println("Failed to parse config:", err)
		os.Exit(1)
	}

	if cfg.Mode != "paper" {
		fmt.Println("Mode is not paper. Exiting mirror execution.")
		return
	}

	tradesBytes, err := os.ReadFile(filepath.Join("output", "recommended_trades.json"))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No recommended_trades.json found.")
			return
		}
		fmt.Println("Error reading recommended trades:", err)
		os.Exit(1)
	}

	var payload RecommendedTradesPayload
	if err := json.Unmarshal(tradesBytes, &payload); err != nil {
		fmt.Println("Failed to parse recommended trades payload:", err)
		os.Exit(1)
	}

	var stagedOrders []StagedOrder
	for _, trade := range payload.Trades {
		// Basic validation
		if trade.Symbol == "" || trade.Qty <= 0 {
			fmt.Printf("Skipping invalid trade: %+v\n", trade)
			continue
		}
		if trade.Side != "buy" && trade.Side != "sell" && trade.Side != "BUY" && trade.Side != "SELL" {
			fmt.Printf("Skipping invalid trade side: %+v\n", trade)
			continue
		}

		fmt.Printf("Staging Paper Trade: %s %d %s (Confidence: %.2f)\n", trade.Side, trade.Qty, trade.Symbol, trade.Confidence)

		stagedOrders = append(stagedOrders, StagedOrder{
			Symbol:      trade.Symbol,
			Qty:         trade.Qty,
			Side:        trade.Side,
			Type:        "market",
			TimeInForce: "day",
			Reason:      trade.Reason,
			Confidence:  trade.Confidence,
		})
	}

	stagedPayload := StagedOrdersPayload{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Orders:      stagedOrders,
	}

	outBytes, _ := json.MarshalIndent(stagedPayload, "", "  ")
	outPath := filepath.Join("output", "staged_alpaca_orders.json")
	if err := os.WriteFile(outPath, outBytes, 0644); err != nil {
		fmt.Println("Error writing staged orders:", err)
		os.Exit(1)
	}

	fmt.Println("Successfully wrote staged orders to", outPath)
}
