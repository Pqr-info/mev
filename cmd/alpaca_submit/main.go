package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AlpacaBaseURL   string `yaml:"alpaca_base_url"`
	AlpacaAPIKeyID  string `yaml:"alpaca_api_key_id"`
	AlpacaAPISecret string `yaml:"alpaca_api_secret"`
	Mode            string `yaml:"mode"`
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
		fmt.Println("Mode is not paper. Exiting execution.")
		return
	}

	stagedPath := filepath.Join("output", "staged_alpaca_orders.json")
	tradesBytes, err := os.ReadFile(stagedPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No staged_alpaca_orders.json found. Nothing to submit.")
			return
		}
		fmt.Println("Error reading staged orders:", err)
		os.Exit(1)
	}

	var payload StagedOrdersPayload
	if err := json.Unmarshal(tradesBytes, &payload); err != nil {
		fmt.Println("Failed to parse staged orders:", err)
		os.Exit(1)
	}

	if len(payload.Orders) == 0 {
		fmt.Println("Staged orders file is empty. Nothing to submit.")
		return
	}

	fmt.Printf("Found %d staged orders generated at %s:\n\n", len(payload.Orders), payload.GeneratedAt)
	for i, o := range payload.Orders {
		fmt.Printf("  %d) %s %d %s\n", i+1, strings.ToUpper(o.Side), o.Qty, o.Symbol)
		fmt.Printf("     Type: %s, TimeInForce: %s\n", o.Type, o.TimeInForce)
		fmt.Printf("     Reason: %s (Confidence: %.2f)\n\n", o.Reason, o.Confidence)
	}

	fmt.Print("Execute these orders on Alpaca Paper? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("Aborting submission.")
		return
	}

	fmt.Println("Submitting orders to Alpaca...")
	successCount := 0

	for _, o := range payload.Orders {
		apiPayload := map[string]interface{}{
			"symbol":        o.Symbol,
			"qty":           o.Qty,
			"side":          o.Side,
			"type":          o.Type,
			"time_in_force": o.TimeInForce,
		}
		
		body, _ := json.Marshal(apiPayload)
		req, _ := http.NewRequest("POST", cfg.AlpacaBaseURL+"/v2/orders", bytes.NewBuffer(body))
		req.Header.Set("APCA-API-KEY-ID", cfg.AlpacaAPIKeyID)
		req.Header.Set("APCA-API-SECRET-KEY", cfg.AlpacaAPISecret)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Failed to execute trade for %s: %v\n", o.Symbol, err)
			continue
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 400 {
			fmt.Printf("❌ Alpaca API Error (%d) for %s: %s\n", resp.StatusCode, o.Symbol, string(respBody))
		} else {
			fmt.Printf("✅ Success submitting %s: %s\n", o.Symbol, string(respBody))
			successCount++
		}
	}

	// Rename the file if any orders were processed, to avoid duplicate manual submission
	if successCount > 0 || len(payload.Orders) > 0 {
		executedPath := filepath.Join("output", "staged_alpaca_orders.executed.json")
		if err := os.Rename(stagedPath, executedPath); err != nil {
			fmt.Println("Warning: Failed to rename staged file to prevent duplicates:", err)
		} else {
			fmt.Println("Moved staged orders file to", executedPath)
		}
	}
}
