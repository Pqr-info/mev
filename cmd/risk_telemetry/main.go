package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type PaperConfig struct {
	AlpacaBaseURL   string `yaml:"alpaca_base_url"`
	AlpacaAPIKeyID  string `yaml:"alpaca_api_key_id"`
	AlpacaAPISecret string `yaml:"alpaca_api_secret"`
}

type AlpacaAccount struct {
	Equity          string `json:"equity"`
	LastEquity      string `json:"last_equity"`
	BuyingPower     string `json:"buying_power"`
	InitialMargin   string `json:"initial_margin"`
	PortfolioValue  string `json:"portfolio_value"`
	Cash            string `json:"cash"`
}

type RiskTelemetry struct {
	Timestamp      string  `json:"timestamp"`
	Equity         float64 `json:"equity"`
	DailyPnL       float64 `json:"daily_pnl"`
	DailyPnLPct    float64 `json:"daily_pnl_pct"`
	BuyingPower    float64 `json:"buying_power"`
	InitialMargin  float64 `json:"initial_margin"`
	Cash           float64 `json:"cash"`
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func main() {
	paperBytes, err := os.ReadFile(filepath.Join("config", "paper.yaml"))
	if err != nil {
		log.Fatalf("Failed to read paper config: %v", err)
	}
	var cfg PaperConfig
	yaml.Unmarshal(paperBytes, &cfg)

	outPath := filepath.Join("output", "risk_telemetry.json")

	for {
		req, _ := http.NewRequest("GET", cfg.AlpacaBaseURL+"/v2/account", nil)
		req.Header.Set("APCA-API-KEY-ID", cfg.AlpacaAPIKeyID)
		req.Header.Set("APCA-API-SECRET-KEY", cfg.AlpacaAPISecret)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Alpaca fetch error: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var acc AlpacaAccount
		if err := json.Unmarshal(body, &acc); err != nil {
			log.Printf("Alpaca JSON parse error: %v, body: %s", err, string(body))
			time.Sleep(10 * time.Second)
			continue
		}

		equity := parseFloat(acc.Equity)
		lastEquity := parseFloat(acc.LastEquity)
		pnl := equity - lastEquity
		pnlPct := 0.0
		if lastEquity > 0 {
			pnlPct = (pnl / lastEquity) * 100.0
		}

		rt := RiskTelemetry{
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			Equity:        equity,
			DailyPnL:      pnl,
			DailyPnLPct:   pnlPct,
			BuyingPower:   parseFloat(acc.BuyingPower),
			InitialMargin: parseFloat(acc.InitialMargin),
			Cash:          parseFloat(acc.Cash),
		}

		outBytes, _ := json.MarshalIndent(rt, "", "  ")
		
		// Write to temp file then rename for atomic write
		tmpPath := outPath + ".tmp"
		os.WriteFile(tmpPath, outBytes, 0644)
		os.Rename(tmpPath, outPath)

		time.Sleep(15 * time.Second) // poll every 15 seconds
	}
}
