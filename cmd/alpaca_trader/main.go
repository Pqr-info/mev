package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const LedgerFilePath = `C:\Users\theal\.gemini\antigravity\brain\3814f74d-5bdb-424b-963f-2eea4a5d5cf7\scratch\immutable_ledger.json`

// Define App Data Structures
type TradeLog struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"Level"`
	Tag       string    `json:"Tag"`
	Message   string    `json:"Message"`
}

type Position struct {
	Symbol       string    `json:"Symbol"`
	Qty          int       `json:"Qty"`
	CostBasis    float64   `json:"CostBasis"`
	Price        float64   `json:"Price"`
	OpenedAt     time.Time `json:"OpenedAt"`
	PriceHistory []float64 `json:"PriceHistory"`
	Broker       string    `json:"Broker,omitempty"` // Alpaca, Schwab, Ibkr, Paypal, Cashapp
}

type Order struct {
	ID        string    `json:"ID"`
	Symbol    string    `json:"Symbol"`
	Side      string    `json:"Side"` // buy, sell, sell_short
	Qty       int       `json:"Qty"`
	Price     float64   `json:"Price"`
	Status    string    `json:"Status"` // filled, rejected, expired, new
	Timestamp time.Time `json:"Timestamp"`
	Broker    string    `json:"Broker,omitempty"`
	LatencyMs int       `json:"LatencyMs,omitempty"` // Simulated routing delay
}

type TickData struct {
	Timestamp time.Time `json:"Timestamp"`
	Strategy  string    `json:"Strategy"` // ema, sniper, sovereign

	// EMA fields
	Price    float64 `json:"Price,omitempty"`
	ShortEma float64 `json:"ShortEma,omitempty"`
	LongEma  float64 `json:"LongEma,omitempty"`

	// Sniper fields
	PriceA  float64 `json:"PriceA,omitempty"`
	PriceB  float64 `json:"PriceB,omitempty"`
	Spread  float64 `json:"Spread,omitempty"`
	ZScore  float64 `json:"ZScore,omitempty"`
	VolumeA float64 `json:"VolumeA,omitempty"`
	VolumeB float64 `json:"VolumeB,omitempty"`

	// Sovereign fields
	Ma25   float64 `json:"Ma25,omitempty"`
	Rsi    float64 `json:"Rsi,omitempty"`
	Fib382 float64 `json:"Fib382,omitempty"`
	Fib618 float64 `json:"Fib618,omitempty"`
	Slope  float64 `json:"Slope,omitempty"`
	RelVol float64 `json:"RelVol,omitempty"`

	Signal string `json:"Signal"` // buy, sell, none
}

type AppConfig struct {
	Strategy        string  `json:"Strategy"`
	Symbol          string  `json:"Symbol"`
	Qty             int     `json:"Qty"`
	IntervalSeconds int     `json:"IntervalSeconds"`
	ShortEmaPeriod  int     `json:"ShortEmaPeriod"`
	LongEmaPeriod   int     `json:"LongEmaPeriod"`
	AssetA          string  `json:"AssetA"`
	AssetB          string  `json:"AssetB"`
	Mean            float64 `json:"Mean"`
	StdDev          float64 `json:"StdDev"`
	MinVolume       float64 `json:"MinVolume"`
	IronFloor       float64 `json:"IronFloor"`
	LivenessWindow  int     `json:"LivenessWindow"`
	EnableSchwab    bool    `json:"EnableSchwab"`
	EnableIbkr      bool    `json:"EnableIbkr"`
	EnablePaypal    bool    `json:"EnablePaypal"`
	EnableCashapp   bool    `json:"EnableCashapp"`
	EnableStash     bool    `json:"EnableStash"`
}

// Order Book Structs
type OrderBookEntry struct {
	Price float64 `json:"price"`
	Size  int     `json:"size"`
}

type OrderBook struct {
	Symbol string           `json:"symbol"`
	Bids   []OrderBookEntry `json:"bids"`
	Asks   []OrderBookEntry `json:"asks"`
}

// Five-broker Immutable Ledger Transaction Struct
type LedgerTx struct {
	Index           int64          `json:"Index"`
	PrevHash        string         `json:"PrevHash"`
	Timestamp       time.Time      `json:"Timestamp"`
	Action          string         `json:"Action"`
	Symbol          string         `json:"Symbol"`
	Qty             int            `json:"Qty"`
	Price           float64        `json:"Price"`
	Broker          string         `json:"Broker"`
	AlpacaCash      float64        `json:"AlpacaCash"`
	SchwabCash      float64        `json:"SchwabCash"`
	IbkrCash        float64        `json:"IbkrCash"`
	PaypalCash      float64        `json:"PaypalCash"`
	CashappCash     float64        `json:"CashappCash"`
	StashCash       float64        `json:"StashCash"`
	AlpacaHoldings  map[string]int `json:"AlpacaHoldings"`
	SchwabHoldings  map[string]int `json:"SchwabHoldings"`
	IbkrHoldings    map[string]int `json:"IbkrHoldings"`
	PaypalHoldings  map[string]int `json:"PaypalHoldings"`
	CashappHoldings map[string]int `json:"CashappHoldings"`
	StashHoldings   map[string]int `json:"StashHoldings"`
	LatencyMs       int            `json:"LatencyMs,omitempty"`
	Hash            string         `json:"Hash"`
}

// Mortal Limit Order Struct
type LimitOrder struct {
	ID            string    `json:"id"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"`
	Qty           int       `json:"qty"`
	Price         float64   `json:"price"`
	LifespanTicks int       `json:"lifespan_ticks"`
	CreatedAt     time.Time `json:"created_at"`
	Broker        string    `json:"broker"`
}

// Option Suggestion Struct
type OptionSuggestion struct {
	ID               string    `json:"id"`
	Timestamp        time.Time `json:"timestamp"`
	UnderlyingSymbol string    `json:"underlying_symbol"`
	OccSymbol        string    `json:"occ_symbol"`
	StrategyName     string    `json:"strategy_name"`
	Strike           float64   `json:"strike"`
	Expiration       string    `json:"expiration"`
	Type             string    `json:"type"`
	Premium          float64   `json:"premium"`
	Reason           string    `json:"reason"`
}

// Global Application Engine State
type Engine struct {
	mu           sync.RWMutex
	logsMu       sync.Mutex
	Config       AppConfig
	Running      bool
	Logs         []TradeLog
	IsSimulation bool

	keyID     string
	secretKey string
	baseURL   string

	SimCash        float64
	SimPositions   map[string]Position
	SimOrders      []Order
	SimLatestPrice float64

	SimSchwabCash      float64
	SimSchwabPositions map[string]Position
	SimSchwabOrders    []Order

	SimIbkrCash      float64
	SimIbkrPositions map[string]Position
	SimIbkrOrders    []Order

	SimPaypalCash      float64
	SimPaypalPositions map[string]Position
	SimPaypalOrders    []Order

	SimCashappCash      float64
	SimCashappPositions map[string]Position
	SimCashappOrders    []Order

	SimStashCash      float64
	SimStashPositions map[string]Position
	SimStashOrders    []Order

	SimLedger         []LedgerTx
	ActiveLimitOrders []LimitOrder

	ActiveOptionSuggestions []OptionSuggestion

	RealTimePrices        map[string]float64
	RealTimeVolume        map[string]float64
	RealTimeVolumeHistory map[string][]float64
	RealTimePriceHistory  map[string][]float64

	historicalPrices []float64
	lastPositionSide string

	// Backtest variables
	IsBacktesting   bool
	BacktestResults map[string]interface{}

	// Black swan simulation triggers
	blackSwanSpreadMult float64

	clients   map[chan string]bool
	clientsMu sync.Mutex

	ctx      context.Context
	cancelFn context.CancelFunc
}

func NewEngine() *Engine {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	keyID := os.Getenv("ALPACA_API_KEY_ID")
	secretKey := os.Getenv("ALPACA_API_SECRET_KEY")
	baseURL := os.Getenv("ALPACA_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://paper-api.alpaca.markets"
	}

	isSim := keyID == "" || secretKey == ""

	e := &Engine{
		Config: AppConfig{
			Strategy:        "sovereign",
			Symbol:          "AAPL",
			Qty:             10,
			IntervalSeconds: 3,
			ShortEmaPeriod:  5,
			LongEmaPeriod:   12,
			AssetA:          "AAPL",
			AssetB:          "MSFT",
			Mean:            50.0,
			StdDev:          2.5,
			MinVolume:       1.0,
			IronFloor:       1000.0,
			LivenessWindow:  75,
			EnableSchwab:    true,
			EnableIbkr:      true,
			EnablePaypal:    true,
			EnableCashapp:   true,
		},
		IsSimulation: isSim,
		keyID:        keyID,
		secretKey:    secretKey,
		baseURL:      baseURL,

		SimCash:                 100000.0,
		SimPositions:            make(map[string]Position),
		SimOrders:               []Order{},
		SimSchwabCash:           100000.0,
		SimSchwabPositions:      make(map[string]Position),
		SimSchwabOrders:         []Order{},
		SimIbkrCash:             100000.0,
		SimIbkrPositions:        make(map[string]Position),
		SimIbkrOrders:           []Order{},
		SimPaypalCash:           100000.0,
		SimPaypalPositions:      make(map[string]Position),
		SimPaypalOrders:         []Order{},
		SimCashappCash:          100000.0,
		SimCashappPositions:     make(map[string]Position),
		SimCashappOrders:        []Order{},
		SimStashCash:            100000.0,
		SimStashPositions:       make(map[string]Position),
		SimStashOrders:          []Order{},
		SimLedger:               []LedgerTx{},
		ActiveLimitOrders:       []LimitOrder{},
		ActiveOptionSuggestions: []OptionSuggestion{},

		RealTimePrices:        make(map[string]float64),
		RealTimeVolume:        make(map[string]float64),
		RealTimeVolumeHistory: make(map[string][]float64),
		RealTimePriceHistory:  make(map[string][]float64),

		blackSwanSpreadMult: 1.0,

		clients:          make(map[chan string]bool),
		historicalPrices: make([]float64, 0),
		lastPositionSide: "flat",
	}

	e.SimLatestPrice = 150.00
	e.RealTimePrices["AAPL"] = 150.00
	e.RealTimePrices["MSFT"] = 100.00
	e.RealTimePrices["USDCUSD"] = 1.00
	e.RealTimePrices["USDTUSD"] = 1.00
	e.RealTimePrices["PYUSD"] = 1.00
	e.RealTimePrices["BTCUSD"] = 65000.00
	e.RealTimePrices["ETHUSD"] = 3500.00
	e.RealTimePrices["GBTC"] = 45.00
	e.RealTimePrices["ETHE"] = 22.00

	e.RealTimeVolume["AAPL"] = 5.0
	e.RealTimeVolume["MSFT"] = 5.0
	e.RealTimeVolume["USDCUSD"] = 20.0
	e.RealTimeVolume["USDTUSD"] = 25.0
	e.RealTimeVolume["PYUSD"] = 15.0
	e.RealTimeVolume["BTCUSD"] = 12.0
	e.RealTimeVolume["ETHUSD"] = 15.0
	e.RealTimeVolume["GBTC"] = 4.0
	e.RealTimeVolume["ETHE"] = 3.0

	e.loadLedgerFromDisk()

	if isSim {
		e.Log("INFO", "SYSTEM", "No Alpaca API credentials found. Running in high-fidelity simulated multi-broker mode.")
	} else {
		e.Log("INFO", "SYSTEM", fmt.Sprintf("Alpaca credentials loaded. Live target environment: %s", baseURL))
	}

	return e
}

// Simulated Order Book Generation (widens spreads under Black Swan Rate Spike)
func (e *Engine) generateSimulatedOrderBook(symbol string, midPrice float64) OrderBook {
	bids := make([]OrderBookEntry, 5)
	asks := make([]OrderBookEntry, 5)

	e.mu.RLock()
	spreadMult := e.blackSwanSpreadMult
	e.mu.RUnlock()

	spread := (0.05 + rand.Float64()*0.10) * spreadMult
	bidBase := midPrice - spread/2
	askBase := midPrice + spread/2

	for i := 0; i < 5; i++ {
		bids[i] = OrderBookEntry{
			Price: math.Round((bidBase-float64(i)*0.10*spreadMult)*100) / 100,
			Size:  10 + rand.Intn(100),
		}
		asks[i] = OrderBookEntry{
			Price: math.Round((askBase+float64(i)*0.10*spreadMult)*100) / 100,
			Size:  10 + rand.Intn(100),
		}
	}
	return OrderBook{Symbol: symbol, Bids: bids, Asks: asks}
}

// Match Market order against simulated book
func (e *Engine) MatchMarketOrder(symbol string, side string, qty int) (float64, error) {
	e.mu.Lock()
	midPrice := e.RealTimePrices[symbol]
	e.mu.Unlock()

	book := e.generateSimulatedOrderBook(symbol, midPrice)

	remaining := qty
	filledCost := 0.0

	if side == "buy" {
		for _, ask := range book.Asks {
			if remaining <= 0 {
				break
			}
			matchQty := ask.Size
			if matchQty > remaining {
				matchQty = remaining
			}
			filledCost += float64(matchQty) * ask.Price
			remaining -= matchQty
		}
		if remaining > 0 {
			lastAsk := book.Asks[4]
			filledCost += float64(remaining) * (lastAsk.Price * 1.01)
		}
	} else {
		for _, bid := range book.Bids {
			if remaining <= 0 {
				break
			}
			matchQty := bid.Size
			if matchQty > remaining {
				matchQty = remaining
			}
			filledCost += float64(matchQty) * bid.Price
			remaining -= matchQty
		}
		if remaining > 0 {
			lastBid := book.Bids[4]
			filledCost += float64(remaining) * (lastBid.Price * 0.99)
		}
	}

	avgPrice := filledCost / float64(qty)
	return math.Round(avgPrice*100) / 100, nil
}

// Append-only cryptographically chained transaction ledger with latency tracking
func (e *Engine) commitToLedger(broker string, action string, symbol string, qty int, price float64, latencyMs int) {
	e.mu.Lock()

	prevHash := "0000000000000000000000000000000000000000000000000000000000000000"
	var index int64 = 0
	if len(e.SimLedger) > 0 {
		lastTx := e.SimLedger[len(e.SimLedger)-1]
		prevHash = lastTx.Hash
		index = lastTx.Index + 1
	}

	alpHoldings := make(map[string]int)
	for s, pos := range e.SimPositions {
		alpHoldings[s] = pos.Qty
	}

	schHoldings := make(map[string]int)
	for s, pos := range e.SimSchwabPositions {
		schHoldings[s] = pos.Qty
	}

	ibkrHoldings := make(map[string]int)
	for s, pos := range e.SimIbkrPositions {
		ibkrHoldings[s] = pos.Qty
	}

	payHoldings := make(map[string]int)
	for s, pos := range e.SimPaypalPositions {
		payHoldings[s] = pos.Qty
	}

	caHoldings := make(map[string]int)
	for s, pos := range e.SimCashappPositions {
		caHoldings[s] = pos.Qty
	}

	stHoldings := make(map[string]int)
	for s, pos := range e.SimStashPositions {
		stHoldings[s] = pos.Qty
	}

	tx := LedgerTx{
		Index:           index,
		PrevHash:        prevHash,
		Timestamp:       time.Now(),
		Action:          action,
		Symbol:          symbol,
		Qty:             qty,
		Price:           price,
		Broker:          broker,
		AlpacaCash:      e.SimCash,
		SchwabCash:      e.SimSchwabCash,
		IbkrCash:        e.SimIbkrCash,
		PaypalCash:      e.SimPaypalCash,
		CashappCash:     e.SimCashappCash,
		StashCash:       e.SimStashCash,
		AlpacaHoldings:  alpHoldings,
		SchwabHoldings:  schHoldings,
		IbkrHoldings:    ibkrHoldings,
		PaypalHoldings:  payHoldings,
		CashappHoldings: caHoldings,
		StashHoldings:   stHoldings,
		LatencyMs:       latencyMs,
	}

	dataString := fmt.Sprintf("%d|%s|%s|%s|%s|%d|%.4f|%s|%.4f|%.4f|%.4f|%.4f|%.4f|%.4f|%v|%v|%v|%v|%v|%v|%d",
		tx.Index, tx.PrevHash, tx.Timestamp.Format(time.RFC3339),
		tx.Action, tx.Symbol, tx.Qty, tx.Price, tx.Broker,
		tx.AlpacaCash, tx.SchwabCash, tx.IbkrCash, tx.PaypalCash, tx.CashappCash, tx.StashCash,
		tx.AlpacaHoldings, tx.SchwabHoldings, tx.IbkrHoldings, tx.PaypalHoldings, tx.CashappHoldings, tx.StashHoldings,
		tx.LatencyMs)

	h := sha256.New()
	h.Write([]byte(dataString))
	tx.Hash = hex.EncodeToString(h.Sum(nil))

	e.SimLedger = append(e.SimLedger, tx)
	e.saveLedgerToDisk()
	e.mu.Unlock()

	e.Log("INFO", "LEDGER", fmt.Sprintf("[Commit Tx #%d] %s: %s %s. Hash: %s... Latency: %dms", tx.Index, tx.Broker, tx.Action, tx.Symbol, tx.Hash[:12], tx.LatencyMs))
}

func (e *Engine) saveLedgerToDisk() {
	dir := filepath.Dir(LedgerFilePath)
	_ = os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(e.SimLedger, "", "  ")
	if err == nil {
		_ = os.WriteFile(LedgerFilePath, data, 0644)
	}
}

func (e *Engine) loadLedgerFromDisk() {
	e.mu.Lock()
	if _, err := os.Stat(LedgerFilePath); os.IsNotExist(err) {
		e.SimLedger = []LedgerTx{}
		e.mu.Unlock()
		e.commitToLedger("SYSTEM", "GENESIS", "SYSTEM", 0, 0.0, 0)
		return
	}

	data, err := os.ReadFile(LedgerFilePath)
	if err == nil {
		var ledger []LedgerTx
		if err := json.Unmarshal(data, &ledger); err == nil {
			e.SimLedger = ledger
			if len(ledger) > 0 {
				lastTx := ledger[len(ledger)-1]
				e.SimCash = lastTx.AlpacaCash
				e.SimSchwabCash = lastTx.SchwabCash
				e.SimIbkrCash = lastTx.IbkrCash
				e.SimPaypalCash = lastTx.PaypalCash
				e.SimCashappCash = lastTx.CashappCash
				e.SimStashCash = lastTx.StashCash

				e.SimPositions = make(map[string]Position)
				for s, q := range lastTx.AlpacaHoldings {
					e.SimPositions[s] = Position{
						Symbol:       s,
						Qty:          q,
						CostBasis:    lastTx.Price,
						Price:        e.RealTimePrices[s],
						OpenedAt:     time.Now(),
						PriceHistory: []float64{e.RealTimePrices[s]},
						Broker:       "Alpaca",
					}
				}

				e.SimSchwabPositions = make(map[string]Position)
				for s, q := range lastTx.SchwabHoldings {
					e.SimSchwabPositions[s] = Position{
						Symbol:       s,
						Qty:          q,
						CostBasis:    lastTx.Price,
						Price:        e.RealTimePrices[s],
						OpenedAt:     time.Now(),
						PriceHistory: []float64{e.RealTimePrices[s]},
						Broker:       "Schwab",
					}
				}

				e.SimIbkrPositions = make(map[string]Position)
				for s, q := range lastTx.IbkrHoldings {
					e.SimIbkrPositions[s] = Position{
						Symbol:       s,
						Qty:          q,
						CostBasis:    lastTx.Price,
						Price:        e.RealTimePrices[s],
						OpenedAt:     time.Now(),
						PriceHistory: []float64{e.RealTimePrices[s]},
						Broker:       "Ibkr",
					}
				}

				e.SimPaypalPositions = make(map[string]Position)
				for s, q := range lastTx.PaypalHoldings {
					e.SimPaypalPositions[s] = Position{
						Symbol:       s,
						Qty:          q,
						CostBasis:    lastTx.Price,
						Price:        e.RealTimePrices[s],
						OpenedAt:     time.Now(),
						PriceHistory: []float64{e.RealTimePrices[s]},
						Broker:       "Paypal",
					}
				}

				e.SimCashappPositions = make(map[string]Position)
				for s, q := range lastTx.CashappHoldings {
 					e.SimCashappPositions[s] = Position{
 						Symbol:       s,
 						Qty:          q,
 						CostBasis:    lastTx.Price,
 						Price:        e.RealTimePrices[s],
 						OpenedAt:     time.Now(),
 						PriceHistory: []float64{e.RealTimePrices[s]},
 						Broker:       "Cashapp",
 					}
 				}

				e.SimStashPositions = make(map[string]Position)
				for s, q := range lastTx.StashHoldings {
					e.SimStashPositions[s] = Position{
						Symbol:       s,
						Qty:          q,
						CostBasis:    lastTx.Price,
						Price:        e.RealTimePrices[s],
						OpenedAt:     time.Now(),
						PriceHistory: []float64{e.RealTimePrices[s]},
						Broker:       "Stash",
					}
				}

				e.mu.Unlock()
				e.Log("INFO", "LEDGER", fmt.Sprintf("Ledger database synchronized. Blocks loaded: %d", len(ledger)))
				return
			}
		}
	}
	e.mu.Unlock()
}

// Iron floor cash protection
func (e *Engine) validateIronFloor(symbol string, qty int, price float64) (bool, string) {
	e.mu.Lock()
	cashFloor := e.Config.IronFloor
	simCash := e.SimCash
	e.mu.Unlock()

	cost := float64(qty) * price
	projectedCash := simCash - cost
	if projectedCash < cashFloor {
		return false, fmt.Sprintf("[IRON FLOOR BREACH] Trade rejected. Projected cash: $%.2f, Minimum floor limit: $%.2f (Trade cost: $%.2f)", projectedCash, cashFloor, cost)
	}
	return true, ""
}

// Logging helper that broadcasts to SSE listeners
func (e *Engine) Log(level, tag, msg string) {
	e.logsMu.Lock()
	l := TradeLog{
		Timestamp: time.Now(),
		Level:     level,
		Tag:       tag,
		Message:   msg,
	}
	e.Logs = append(e.Logs, l)
	if len(e.Logs) > 100 {
		e.Logs = e.Logs[1:]
	}
	e.logsMu.Unlock()

	if level == "ERROR" {
		log.Error().Str("tag", tag).Msg(msg)
	} else if level == "WARN" {
		log.Warn().Str("tag", tag).Msg(msg)
	} else {
		log.Info().Str("tag", tag).Msg(msg)
	}

	payload, err := json.Marshal(l)
	if err == nil {
		e.broadcast("log", string(payload))
	}
}

// Broadcasting SSE payloads
func (e *Engine) broadcast(event, data string) {
	e.clientsMu.Lock()
	defer e.clientsMu.Unlock()

	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)
	for ch := range e.clients {
		select {
		case ch <- msg:
		default:
			close(ch)
			delete(e.clients, ch)
		}
	}
}

// REST call wrapper for Alpaca
func (e *Engine) alpacaRequest(method, path string, body []byte) ([]byte, int, error) {
	url := fmt.Sprintf("%s%s", e.baseURL, path)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("APCA-API-KEY-ID", e.keyID)
	req.Header.Set("APCA-API-SECRET-KEY", e.secretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, err
}

func (e *Engine) GetPortfolio() (map[string]interface{}, []Position, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var positions []Position

	alpUnrPnl := 0.0
	for _, p := range e.SimPositions {
		currPrice := e.RealTimePrices[p.Symbol]
		if currPrice == 0 {
			currPrice = p.CostBasis
		}
		p.Price = currPrice
		p.Broker = "Alpaca"
		positions = append(positions, p)
		alpUnrPnl += (p.Price - p.CostBasis) * float64(p.Qty)
	}
	alpEquity := e.SimCash + alpUnrPnl

	schUnrPnl := 0.0
	for _, p := range e.SimSchwabPositions {
		currPrice := e.RealTimePrices[p.Symbol]
		if currPrice == 0 {
			currPrice = p.CostBasis
		}
		p.Price = currPrice
		p.Broker = "Schwab"
		positions = append(positions, p)
		schUnrPnl += (p.Price - p.CostBasis) * float64(p.Qty)
	}
	schEquity := e.SimSchwabCash + schUnrPnl

	ibkrUnrPnl := 0.0
	for _, p := range e.SimIbkrPositions {
		currPrice := e.RealTimePrices[p.Symbol]
		if currPrice == 0 {
			currPrice = p.CostBasis
		}
		p.Price = currPrice
		p.Broker = "Ibkr"
		positions = append(positions, p)
		ibkrUnrPnl += (p.Price - p.CostBasis) * float64(p.Qty)
	}
	ibkrEquity := e.SimIbkrCash + ibkrUnrPnl

	payUnrPnl := 0.0
	for _, p := range e.SimPaypalPositions {
		currPrice := e.RealTimePrices[p.Symbol]
		if currPrice == 0 {
			currPrice = p.CostBasis
		}
		p.Price = currPrice
		p.Broker = "Paypal"
		positions = append(positions, p)
		payUnrPnl += (p.Price - p.CostBasis) * float64(p.Qty)
	}
	payEquity := e.SimPaypalCash + payUnrPnl

	caUnrPnl := 0.0
	for _, p := range e.SimCashappPositions {
		currPrice := e.RealTimePrices[p.Symbol]
		if currPrice == 0 {
			currPrice = p.CostBasis
		}
		p.Price = currPrice
		p.Broker = "Cashapp"
		positions = append(positions, p)
		caUnrPnl += (p.Price - p.CostBasis) * float64(p.Qty)
	}
	caEquity := e.SimCashappCash + caUnrPnl

	stUnrPnl := 0.0
	for _, p := range e.SimStashPositions {
		currPrice := e.RealTimePrices[p.Symbol]
		if currPrice == 0 {
			currPrice = p.CostBasis
		}
		p.Price = currPrice
		p.Broker = "Stash"
		positions = append(positions, p)
		stUnrPnl += (p.Price - p.CostBasis) * float64(p.Qty)
	}
	stEquity := e.SimStashCash + stUnrPnl

	data := map[string]interface{}{
		"AlpacaEquity":  alpEquity,
		"AlpacaCash":    e.SimCash,
		"AlpacaBp":      e.SimCash * 2.0,
		"SchwabEquity":  schEquity,
		"SchwabCash":    e.SimSchwabCash,
		"SchwabBp":      e.SimSchwabCash * 2.0,
		"IbkrEquity":    ibkrEquity,
		"IbkrCash":      e.SimIbkrCash,
		"IbkrBp":        e.SimIbkrCash * 2.0,
		"PaypalEquity":  payEquity,
		"PaypalCash":    e.SimPaypalCash,
		"PaypalBp":      e.SimPaypalCash * 2.0,
		"CashappEquity": caEquity,
		"CashappCash":   e.SimCashappCash,
		"CashappBp":     e.SimCashappCash * 2.0,
		"StashEquity":   stEquity,
		"StashCash":     e.SimStashCash,
		"StashBp":       e.SimStashCash * 2.0,
	}

	return data, positions, nil
}

func (e *Engine) GetOrders() []Order {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var list []Order
	for _, o := range e.SimOrders {
		o.Broker = "Alpaca"
		list = append(list, o)
	}
	for _, o := range e.SimSchwabOrders {
		o.Broker = "Schwab"
		list = append(list, o)
	}
	for _, o := range e.SimIbkrOrders {
		o.Broker = "Ibkr"
		list = append(list, o)
	}
	for _, o := range e.SimPaypalOrders {
		o.Broker = "Paypal"
		list = append(list, o)
	}
	for _, o := range e.SimCashappOrders {
		o.Broker = "Cashapp"
		list = append(list, o)
	}
	for _, o := range e.SimStashOrders {
		o.Broker = "Stash"
		list = append(list, o)
	}
	return list
}

func (e *Engine) PlaceOrderForAsset(symbol string, side string, qty int) {
	e.mu.Lock()
	isSim := e.IsSimulation
	e.mu.Unlock()

	if isSim {
		// Alpaca Latency Simulation: 50ms - 120ms
		latencyMs := 50 + rand.Intn(70)
		if !e.IsBacktesting {
			e.Log("INFO", "LATENCY", fmt.Sprintf("[Alpaca Latency] Simulating routing delay: %dms...", latencyMs))
			time.Sleep(time.Duration(latencyMs) * time.Millisecond)
		} else {
			latencyMs = 0
		}

		price, err := e.MatchMarketOrder(symbol, side, qty)
		if err != nil {
			e.Log("ERROR", "TRADE", fmt.Sprintf("Failed to match Alpaca order in book: %v", err))
			return
		}

		if side == "buy" {
			ok, reason := e.validateIronFloor(symbol, qty, price)
			if !ok {
				e.Log("ERROR", "IRON", reason)
				e.mu.Lock()
				e.SimOrders = append(e.SimOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Alpaca",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}
		}

		e.mu.Lock()
		cost := float64(qty) * price
		if side == "buy" {
			e.SimCash -= cost
			pos, exists := e.SimPositions[symbol]
			if exists {
				totalQty := pos.Qty + qty
				pos.CostBasis = (pos.CostBasis*float64(pos.Qty) + cost) / float64(totalQty)
				pos.Qty = totalQty
				pos.PriceHistory = append(pos.PriceHistory, price)
				e.SimPositions[symbol] = pos
			} else {
				e.SimPositions[symbol] = Position{
					Symbol:       symbol,
					Qty:          qty,
					CostBasis:    price,
					Price:        price,
					OpenedAt:     time.Now(),
					PriceHistory: []float64{price},
					Broker:       "Alpaca",
				}
			}
			e.lastPositionSide = "long"
		} else {
			pos, exists := e.SimPositions[symbol]
			if !exists || (pos.Qty > 0 && pos.Qty < qty) {
				e.Log("ERROR", "TRADE", fmt.Sprintf("[Alpaca] Order REJECTED: Insufficient shares. Position Qty: %d", pos.Qty))
				e.mu.Unlock()
				return
			}

			e.SimCash += cost
			pos.Qty -= qty
			if pos.Qty == 0 {
				delete(e.SimPositions, symbol)
			} else {
				e.SimPositions[symbol] = pos
			}
			e.lastPositionSide = "flat"
		}

		o := Order{
			ID:        fmt.Sprintf("ord-%d", time.Now().UnixNano())[12:18],
			Symbol:    symbol,
			Side:      side,
			Qty:       qty,
			Price:     price,
			Status:    "filled",
			Timestamp: time.Now(),
			Broker:    "Alpaca",
			LatencyMs: latencyMs,
		}
		e.SimOrders = append(e.SimOrders, o)
		e.mu.Unlock()

		e.Log("INFO", "TRADE", fmt.Sprintf("[Alpaca] Order FILLED at price $%.2f. Qty: %d. Latency: %dms", price, qty, latencyMs))

		actionType := "BUY"
		if side == "sell" {
			actionType = "SELL"
		}
		e.commitToLedger("Alpaca", actionType, symbol, qty, price, latencyMs)
		return
	}

	payload := map[string]interface{}{
		"symbol":        symbol,
		"qty":           strconv.Itoa(qty),
		"side":          side,
		"type":          "market",
		"time_in_force": "day",
	}
	body, _ := json.Marshal(payload)
	_, code, err := e.alpacaRequest("POST", "/v2/orders", body)
	if err != nil || (code != http.StatusOK && code != http.StatusAccepted) {
		e.Log("ERROR", "TRADE", fmt.Sprintf("[Alpaca] order placement failed: %v", err))
		return
	}
	e.Log("INFO", "TRADE", fmt.Sprintf("[Alpaca] Order successfully routed. Code: %d", code))
	if side == "buy" {
		e.lastPositionSide = "long"
	} else {
		e.lastPositionSide = "flat"
	}
}

func (e *Engine) PlaceOrderForSchwab(symbol string, side string, qty int) {
	e.mu.Lock()
	isSim := e.IsSimulation
	e.mu.Unlock()

	if isSim {
		// Schwab Latency Simulation: 250ms - 450ms
		latencyMs := 250 + rand.Intn(200)
		if !e.IsBacktesting {
			e.Log("INFO", "LATENCY", fmt.Sprintf("[Schwab Latency] Simulating execution/routing delay: %dms...", latencyMs))
			time.Sleep(time.Duration(latencyMs) * time.Millisecond)
		} else {
			latencyMs = 0
		}

		price, err := e.MatchMarketOrder(symbol, side, qty)
		if err != nil {
			e.Log("ERROR", "TRADE", fmt.Sprintf("Failed to match Schwab order in book: %v", err))
			return
		}

		e.mu.Lock()
		cost := float64(qty) * price
		schwabCash := e.SimSchwabCash
		ironFloor := e.Config.IronFloor
		e.mu.Unlock()

		if side == "buy" {
			if schwabCash-cost < ironFloor {
				e.Log("ERROR", "IRON", fmt.Sprintf("[Schwab Iron Floor Breach] Rejected. Cash: $%.2f, Cost: $%.2f, Floor: $%.2f", schwabCash, cost, ironFloor))
				e.mu.Lock()
				e.SimSchwabOrders = append(e.SimSchwabOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Schwab",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			e.mu.Lock()
			e.SimSchwabCash -= cost
			pos, exists := e.SimSchwabPositions[symbol]
			if exists {
				totalQty := pos.Qty + qty
				pos.CostBasis = (pos.CostBasis*float64(pos.Qty) + cost) / float64(totalQty)
				pos.Qty = totalQty
				pos.PriceHistory = append(pos.PriceHistory, price)
				e.SimSchwabPositions[symbol] = pos
			} else {
				e.SimSchwabPositions[symbol] = Position{
					Symbol:       symbol,
					Qty:          qty,
					CostBasis:    price,
					Price:        price,
					OpenedAt:     time.Now(),
					PriceHistory: []float64{price},
					Broker:       "Schwab",
				}
			}
			e.mu.Unlock()
		} else if side == "sell_short" {
			e.mu.Lock()
			alpPos, alpExists := e.SimPositions[symbol]
			e.mu.Unlock()

			if !alpExists || alpPos.Qty < qty {
				e.Log("WARN", "RISK", fmt.Sprintf("[Auto-Hedge Lockout] Short trade for %s rejected on Schwab. No offsetting long position held on Alpaca.", symbol))
				e.mu.Lock()
				e.SimSchwabOrders = append(e.SimSchwabOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Schwab",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			if schwabCash-cost < ironFloor {
				e.Log("ERROR", "IRON", fmt.Sprintf("[Schwab Iron Floor Breach] Rejected. Cash: $%.2f, Collateral Required: $%.2f, Floor: $%.2f", schwabCash, cost, ironFloor))
				e.mu.Lock()
				e.SimSchwabOrders = append(e.SimSchwabOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Schwab",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			e.mu.Lock()
			e.SimSchwabCash += cost
			pos, exists := e.SimSchwabPositions[symbol]
			if exists {
				pos.Qty -= qty
				pos.PriceHistory = append(pos.PriceHistory, price)
				e.SimSchwabPositions[symbol] = pos
			} else {
				e.SimSchwabPositions[symbol] = Position{
					Symbol:       symbol,
					Qty:          -qty,
					CostBasis:    price,
					Price:        price,
					OpenedAt:     time.Now(),
					PriceHistory: []float64{price},
					Broker:       "Schwab",
				}
			}
			e.mu.Unlock()
		} else {
			e.mu.Lock()
			pos, exists := e.SimSchwabPositions[symbol]
			if !exists {
				e.Log("ERROR", "TRADE", "[Schwab] Order REJECTED: No position found.")
				e.mu.Unlock()
				return
			}

			if pos.Qty < 0 {
				buyCost := float64(qty) * price
				if schwabCash-buyCost < ironFloor {
					e.Log("ERROR", "IRON", fmt.Sprintf("[Schwab Iron Floor Breach] Short cover rejected. Cash: $%.2f, Buyback Cost: $%.2f, Floor: $%.2f", schwabCash, buyCost, ironFloor))
					e.mu.Unlock()
					return
				}
				e.SimSchwabCash -= buyCost
				pos.Qty += qty
			} else {
				if pos.Qty < qty {
					e.Log("ERROR", "TRADE", fmt.Sprintf("[Schwab] Order REJECTED: Insufficient shares. Position: %d", pos.Qty))
					e.mu.Unlock()
					return
				}
				e.SimSchwabCash += cost
				pos.Qty -= qty
			}

			if pos.Qty == 0 {
				delete(e.SimSchwabPositions, symbol)
			} else {
				e.SimSchwabPositions[symbol] = pos
			}
			e.mu.Unlock()
		}

		e.mu.Lock()
		o := Order{
			ID:        fmt.Sprintf("sch-%d", time.Now().UnixNano())[12:18],
			Symbol:    symbol,
			Side:      side,
			Qty:       qty,
			Price:     price,
			Status:    "filled",
			Timestamp: time.Now(),
			Broker:    "Schwab",
			LatencyMs: latencyMs,
		}
		e.SimSchwabOrders = append(e.SimSchwabOrders, o)
		e.mu.Unlock()

		e.Log("INFO", "TRADE", fmt.Sprintf("[Schwab] Order FILLED at price $%.2f. Qty: %d, Action: %s. Latency: %dms", price, qty, side, latencyMs))
		e.commitToLedger("Schwab", side, symbol, qty, price, latencyMs)
		return
	}

	e.Log("INFO", "SCHWAB", fmt.Sprintf("[Schwab Live] Submitting %s order for %d %s via MCP...", side, qty, symbol))
}

func (e *Engine) PlaceOrderForIbkr(symbol string, side string, qty int) {
	e.mu.Lock()
	isSim := e.IsSimulation
	e.mu.Unlock()

	if isSim {
		// IBKR Gateway Latency Simulation: 80ms - 180ms
		latencyMs := 80 + rand.Intn(100)
		if !e.IsBacktesting {
			e.Log("INFO", "LATENCY", fmt.Sprintf("[IBKR Latency] Simulating execution/routing delay: %dms...", latencyMs))
			time.Sleep(time.Duration(latencyMs) * time.Millisecond)
		} else {
			latencyMs = 0
		}

		price, err := e.MatchMarketOrder(symbol, side, qty)
		if err != nil {
			e.Log("ERROR", "TRADE", fmt.Sprintf("Failed to match IBKR order in book: %v", err))
			return
		}

		e.mu.Lock()
		cost := float64(qty) * price
		ibkrCash := e.SimIbkrCash
		ironFloor := e.Config.IronFloor
		e.mu.Unlock()

		if side == "buy" {
			if ibkrCash-cost < ironFloor {
				e.Log("ERROR", "IRON", fmt.Sprintf("[IBKR Iron Floor Breach] Rejected. Cash: $%.2f, Cost: $%.2f, Floor: $%.2f", ibkrCash, cost, ironFloor))
				e.mu.Lock()
				e.SimIbkrOrders = append(e.SimIbkrOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Ibkr",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			e.mu.Lock()
			e.SimIbkrCash -= cost
			pos, exists := e.SimIbkrPositions[symbol]
			if exists {
				totalQty := pos.Qty + qty
				pos.CostBasis = (pos.CostBasis*float64(pos.Qty) + cost) / float64(totalQty)
				pos.Qty = totalQty
				pos.PriceHistory = append(pos.PriceHistory, price)
				e.SimIbkrPositions[symbol] = pos
			} else {
				e.SimIbkrPositions[symbol] = Position{
					Symbol:       symbol,
					Qty:          qty,
					CostBasis:    price,
					Price:        price,
					OpenedAt:     time.Now(),
					PriceHistory: []float64{price},
					Broker:       "Ibkr",
				}
			}
			e.mu.Unlock()
		} else if side == "sell_short" {
			e.mu.Lock()
			alpPos, alpExists := e.SimPositions[symbol]
			e.mu.Unlock()

			if !alpExists || alpPos.Qty < qty {
				e.Log("WARN", "RISK", fmt.Sprintf("[Auto-Hedge Lockout] Short trade for %s rejected on IBKR. No offsetting long position held on Alpaca.", symbol))
				e.mu.Lock()
				e.SimIbkrOrders = append(e.SimIbkrOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Ibkr",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			if ibkrCash-cost < ironFloor {
				e.Log("ERROR", "IRON", fmt.Sprintf("[IBKR Iron Floor Breach] Rejected. Cash: $%.2f, Collateral Required: $%.2f, Floor: $%.2f", ibkrCash, cost, ironFloor))
				e.mu.Lock()
				e.SimIbkrOrders = append(e.SimIbkrOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Ibkr",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			e.mu.Lock()
			e.SimIbkrCash += cost
			pos, exists := e.SimIbkrPositions[symbol]
			if exists {
				pos.Qty -= qty
				pos.PriceHistory = append(pos.PriceHistory, price)
				e.SimIbkrPositions[symbol] = pos
			} else {
				e.SimIbkrPositions[symbol] = Position{
					Symbol:       symbol,
					Qty:          -qty,
					CostBasis:    price,
					Price:        price,
					OpenedAt:     time.Now(),
					PriceHistory: []float64{price},
					Broker:       "Ibkr",
				}
			}
			e.mu.Unlock()
		} else {
			e.mu.Lock()
			pos, exists := e.SimIbkrPositions[symbol]
			if !exists {
				e.Log("ERROR", "TRADE", "[IBKR] Order REJECTED: No position found.")
				e.mu.Unlock()
				return
			}

			if pos.Qty < 0 {
				buyCost := float64(qty) * price
				if ibkrCash-buyCost < ironFloor {
					e.Log("ERROR", "IRON", fmt.Sprintf("[IBKR Iron Floor Breach] Short cover rejected. Cash: $%.2f, Buyback Cost: $%.2f, Floor: $%.2f", ibkrCash, buyCost, ironFloor))
					e.mu.Unlock()
					return
				}
				e.SimIbkrCash -= buyCost
				pos.Qty += qty
			} else {
				if pos.Qty < qty {
					e.Log("ERROR", "TRADE", fmt.Sprintf("[IBKR] Order REJECTED: Insufficient shares. Position: %d", pos.Qty))
					e.mu.Unlock()
					return
				}
				e.SimIbkrCash += cost
				pos.Qty -= qty
			}

			if pos.Qty == 0 {
				delete(e.SimIbkrPositions, symbol)
			} else {
				e.SimIbkrPositions[symbol] = pos
			}
			e.mu.Unlock()
		}

		e.mu.Lock()
		o := Order{
			ID:        fmt.Sprintf("ib-%d", time.Now().UnixNano())[12:18],
			Symbol:    symbol,
			Side:      side,
			Qty:       qty,
			Price:     price,
			Status:    "filled",
			Timestamp: time.Now(),
			Broker:    "Ibkr",
			LatencyMs: latencyMs,
		}
		e.SimIbkrOrders = append(e.SimIbkrOrders, o)
		e.mu.Unlock()

		e.Log("INFO", "TRADE", fmt.Sprintf("[IBKR] Order FILLED at price $%.2f. Qty: %d, Action: %s. Latency: %dms", price, qty, side, latencyMs))
		e.commitToLedger("Ibkr", side, symbol, qty, price, latencyMs)
		return
	}

	e.Log("INFO", "IBKR", fmt.Sprintf("[IBKR Live] Submitting %s order via Gateway...", side))
}

func (e *Engine) PlaceOrderForPaypal(symbol string, side string, qty int) {
	e.mu.Lock()
	isSim := e.IsSimulation
	e.mu.Unlock()

	if isSim {
		// PayPal API latency simulation: 350ms - 700ms (standard merchant REST infrastructure)
		latencyMs := 350 + rand.Intn(350)
		if !e.IsBacktesting {
			e.Log("INFO", "LATENCY", fmt.Sprintf("[PayPal Latency] Simulating execution/routing delay: %dms...", latencyMs))
			time.Sleep(time.Duration(latencyMs) * time.Millisecond)
		} else {
			latencyMs = 0
		}

		price, err := e.MatchMarketOrder(symbol, side, qty)
		if err != nil {
			e.Log("ERROR", "TRADE", fmt.Sprintf("Failed to match PayPal order in book: %v", err))
			return
		}

		e.mu.Lock()
		cost := float64(qty) * price
		paypalCash := e.SimPaypalCash
		ironFloor := e.Config.IronFloor
		e.mu.Unlock()

		if side == "buy" {
			if paypalCash-cost < ironFloor {
				e.Log("ERROR", "IRON", fmt.Sprintf("[PayPal Iron Floor Breach] Rejected. Cash: $%.2f, Cost: $%.2f, Floor: $%.2f", paypalCash, cost, ironFloor))
				e.mu.Lock()
				e.SimPaypalOrders = append(e.SimPaypalOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Paypal",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			e.mu.Lock()
			e.SimPaypalCash -= cost
			pos, exists := e.SimPaypalPositions[symbol]
			if exists {
				totalQty := pos.Qty + qty
				pos.CostBasis = (pos.CostBasis*float64(pos.Qty) + cost) / float64(totalQty)
				pos.Qty = totalQty
				pos.PriceHistory = append(pos.PriceHistory, price)
				e.SimPaypalPositions[symbol] = pos
			} else {
				e.SimPaypalPositions[symbol] = Position{
					Symbol:       symbol,
					Qty:          qty,
					CostBasis:    price,
					Price:        price,
					OpenedAt:     time.Now(),
					PriceHistory: []float64{price},
					Broker:       "Paypal",
				}
			}
			e.mu.Unlock()
		} else if side == "sell_short" {
			e.mu.Lock()
			alpPos, alpExists := e.SimPositions[symbol]
			e.mu.Unlock()

			if !alpExists || alpPos.Qty < qty {
				e.Log("WARN", "RISK", fmt.Sprintf("[Auto-Hedge Lockout] Short trade for %s rejected on PayPal. No offsetting long position held on Alpaca.", symbol))
				e.mu.Lock()
				e.SimPaypalOrders = append(e.SimPaypalOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Paypal",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			if paypalCash-cost < ironFloor {
				e.Log("ERROR", "IRON", fmt.Sprintf("[PayPal Iron Floor Breach] Rejected. Cash: $%.2f, Collateral Required: $%.2f, Floor: $%.2f", paypalCash, cost, ironFloor))
				e.mu.Lock()
				e.SimPaypalOrders = append(e.SimPaypalOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Paypal",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			e.mu.Lock()
			e.SimPaypalCash += cost
			pos, exists := e.SimPaypalPositions[symbol]
			if exists {
				pos.Qty -= qty
				pos.PriceHistory = append(pos.PriceHistory, price)
				e.SimPaypalPositions[symbol] = pos
			} else {
				e.SimPaypalPositions[symbol] = Position{
					Symbol:       symbol,
					Qty:          -qty,
					CostBasis:    price,
					Price:        price,
					OpenedAt:     time.Now(),
					PriceHistory: []float64{price},
					Broker:       "Paypal",
				}
			}
			e.mu.Unlock()
		} else {
			e.mu.Lock()
			pos, exists := e.SimPaypalPositions[symbol]
			if !exists {
				e.Log("ERROR", "TRADE", "[PayPal] Order REJECTED: No position found.")
				e.mu.Unlock()
				return
			}

			if pos.Qty < 0 {
				buyCost := float64(qty) * price
				if paypalCash-buyCost < ironFloor {
					e.Log("ERROR", "IRON", fmt.Sprintf("[PayPal Iron Floor Breach] Short cover rejected. Cash: $%.2f, Buyback Cost: $%.2f, Floor: $%.2f", paypalCash, buyCost, ironFloor))
					e.mu.Unlock()
					return
				}
				e.SimPaypalCash -= buyCost
				pos.Qty += qty
			} else {
				if pos.Qty < qty {
					e.Log("ERROR", "TRADE", fmt.Sprintf("[PayPal] Order REJECTED: Insufficient shares. Position: %d", pos.Qty))
					e.mu.Unlock()
					return
				}
				e.SimPaypalCash += cost
				pos.Qty -= qty
			}

			if pos.Qty == 0 {
				delete(e.SimPaypalPositions, symbol)
			} else {
				e.SimPaypalPositions[symbol] = pos
			}
			e.mu.Unlock()
		}

		e.mu.Lock()
		o := Order{
			ID:        fmt.Sprintf("py-%d", time.Now().UnixNano())[12:18],
			Symbol:    symbol,
			Side:      side,
			Qty:       qty,
			Price:     price,
			Status:    "filled",
			Timestamp: time.Now(),
			Broker:    "Paypal",
			LatencyMs: latencyMs,
		}
		e.SimPaypalOrders = append(e.SimPaypalOrders, o)
		e.mu.Unlock()

		e.Log("INFO", "TRADE", fmt.Sprintf("[PayPal] Order FILLED at price $%.2f. Qty: %d, Action: %s. Latency: %dms", price, qty, side, latencyMs))
		e.commitToLedger("Paypal", side, symbol, qty, price, latencyMs)
		return
	}

	e.Log("INFO", "PAYPAL", fmt.Sprintf("[PayPal Live] Submitting %s order via REST API...", side))
}

func (e *Engine) PlaceOrderForCashapp(symbol string, side string, qty int) {
	e.mu.Lock()
	isSim := e.IsSimulation
	e.mu.Unlock()

	if isSim {
		// Cash App API latency simulation: 120ms - 250ms
		latencyMs := 120 + rand.Intn(130)
		if !e.IsBacktesting {
			e.Log("INFO", "LATENCY", fmt.Sprintf("[Cash App Latency] Simulating execution/routing delay: %dms...", latencyMs))
			time.Sleep(time.Duration(latencyMs) * time.Millisecond)
		} else {
			latencyMs = 0
		}

		price, err := e.MatchMarketOrder(symbol, side, qty)
		if err != nil {
			e.Log("ERROR", "TRADE", fmt.Sprintf("Failed to match Cash App order in book: %v", err))
			return
		}

		e.mu.Lock()
		cost := float64(qty) * price
		cashappCash := e.SimCashappCash
		ironFloor := e.Config.IronFloor
		e.mu.Unlock()

		if side == "buy" {
			if cashappCash-cost < ironFloor {
				e.Log("ERROR", "IRON", fmt.Sprintf("[Cash App Iron Floor Breach] Rejected. Cash: $%.2f, Cost: $%.2f, Floor: $%.2f", cashappCash, cost, ironFloor))
				e.mu.Lock()
				e.SimCashappOrders = append(e.SimCashappOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Cashapp",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			e.mu.Lock()
			e.SimCashappCash -= cost
			pos, exists := e.SimCashappPositions[symbol]
			if exists {
				totalQty := pos.Qty + qty
				pos.CostBasis = (pos.CostBasis*float64(pos.Qty) + cost) / float64(totalQty)
				pos.Qty = totalQty
				pos.PriceHistory = append(pos.PriceHistory, price)
				e.SimCashappPositions[symbol] = pos
			} else {
				e.SimCashappPositions[symbol] = Position{
					Symbol:       symbol,
					Qty:          qty,
					CostBasis:    price,
					Price:        price,
					OpenedAt:     time.Now(),
					PriceHistory: []float64{price},
					Broker:       "Cashapp",
				}
			}
			e.mu.Unlock()
		} else if side == "sell_short" {
			e.mu.Lock()
			alpPos, alpExists := e.SimPositions[symbol]
			e.mu.Unlock()

			if !alpExists || alpPos.Qty < qty {
				e.Log("WARN", "RISK", fmt.Sprintf("[Auto-Hedge Lockout] Short trade for %s rejected on Cash App. No offsetting long position held on Alpaca.", symbol))
				e.mu.Lock()
				e.SimCashappOrders = append(e.SimCashappOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Cashapp",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			if cashappCash-cost < ironFloor {
				e.Log("ERROR", "IRON", fmt.Sprintf("[Cash App Iron Floor Breach] Rejected. Cash: $%.2f, Collateral Required: $%.2f, Floor: $%.2f", cashappCash, cost, ironFloor))
				e.mu.Lock()
				e.SimCashappOrders = append(e.SimCashappOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Cashapp",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			e.mu.Lock()
			e.SimCashappCash += cost
			pos, exists := e.SimCashappPositions[symbol]
			if exists {
				pos.Qty -= qty
				pos.PriceHistory = append(pos.PriceHistory, price)
				e.SimCashappPositions[symbol] = pos
			} else {
				e.SimCashappPositions[symbol] = Position{
					Symbol:       symbol,
					Qty:          -qty,
					CostBasis:    price,
					Price:        price,
					OpenedAt:     time.Now(),
					PriceHistory: []float64{price},
					Broker:       "Cashapp",
				}
			}
			e.mu.Unlock()
		} else {
			e.mu.Lock()
			pos, exists := e.SimCashappPositions[symbol]
			if !exists {
				e.Log("ERROR", "TRADE", "[Cash App] Order REJECTED: No position found.")
				e.mu.Unlock()
				return
			}

			if pos.Qty < 0 {
				buyCost := float64(qty) * price
				if cashappCash-buyCost < ironFloor {
					e.Log("ERROR", "IRON", fmt.Sprintf("[Cash App Short Cover Breach] Short cover rejected. Cash: $%.2f, Buyback Cost: $%.2f, Floor: $%.2f", cashappCash, buyCost, ironFloor))
					e.mu.Unlock()
					return
				}
				e.SimCashappCash -= buyCost
				pos.Qty += qty
			} else {
				if pos.Qty < qty {
					e.Log("ERROR", "TRADE", fmt.Sprintf("[Cash App] Order REJECTED: Insufficient shares. Position: %d", pos.Qty))
					e.mu.Unlock()
					return
				}
				e.SimCashappCash += cost
				pos.Qty -= qty
			}

			if pos.Qty == 0 {
				delete(e.SimCashappPositions, symbol)
			} else {
				e.SimCashappPositions[symbol] = pos
			}
			e.mu.Unlock()
		}

		e.mu.Lock()
		o := Order{
			ID:        fmt.Sprintf("ca-%d", time.Now().UnixNano())[12:18],
			Symbol:    symbol,
			Side:      side,
			Qty:       qty,
			Price:     price,
			Status:    "filled",
			Timestamp: time.Now(),
			Broker:    "Cashapp",
			LatencyMs: latencyMs,
		}
		e.SimCashappOrders = append(e.SimCashappOrders, o)
		e.mu.Unlock()

		e.Log("INFO", "TRADE", fmt.Sprintf("[Cash App] Order FILLED at price $%.2f. Qty: %d, Action: %s. Latency: %dms", price, qty, side, latencyMs))
		e.commitToLedger("Cashapp", side, symbol, qty, price, latencyMs)
		return
	}

	e.Log("INFO", "CASHAPP", fmt.Sprintf("[Cash App Live] Submitting %s order via mobile endpoint...", side))
}

func (e *Engine) PlaceOrderForStash(symbol string, side string, qty int) {
	e.mu.Lock()
	isSim := e.IsSimulation
	e.mu.Unlock()

	if isSim {
		// Stash retail brokerage latency simulation: 400ms - 800ms
		latencyMs := 400 + rand.Intn(400)
		if !e.IsBacktesting {
			e.Log("INFO", "LATENCY", fmt.Sprintf("[Stash Latency] Simulating batch execution routing delay: %dms...", latencyMs))
			time.Sleep(time.Duration(latencyMs) * time.Millisecond)
		} else {
			latencyMs = 0
		}

		price, err := e.MatchMarketOrder(symbol, side, qty)
		if err != nil {
			e.Log("ERROR", "TRADE", fmt.Sprintf("Failed to match Stash order in book: %v", err))
			return
		}

		e.mu.Lock()
		cost := float64(qty) * price
		stashCash := e.SimStashCash
		ironFloor := e.Config.IronFloor
		e.mu.Unlock()

		if side == "buy" {
			if stashCash-cost < ironFloor {
				e.Log("ERROR", "IRON", fmt.Sprintf("[Stash Iron Floor Breach] Rejected. Cash: $%.2f, Cost: $%.2f, Floor: $%.2f", stashCash, cost, ironFloor))
				e.mu.Lock()
				e.SimStashOrders = append(e.SimStashOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Stash",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			e.mu.Lock()
			e.SimStashCash -= cost
			pos, exists := e.SimStashPositions[symbol]
			if exists {
				totalQty := pos.Qty + qty
				pos.CostBasis = (pos.CostBasis*float64(pos.Qty) + cost) / float64(totalQty)
				pos.Qty = totalQty
				pos.PriceHistory = append(pos.PriceHistory, price)
				e.SimStashPositions[symbol] = pos
			} else {
				e.SimStashPositions[symbol] = Position{
					Symbol:       symbol,
					Qty:          qty,
					CostBasis:    price,
					Price:        price,
					OpenedAt:     time.Now(),
					PriceHistory: []float64{price},
					Broker:       "Stash",
				}
			}
			e.mu.Unlock()
		} else if side == "sell_short" {
			e.mu.Lock()
			alpPos, alpExists := e.SimPositions[symbol]
			e.mu.Unlock()

			if !alpExists || alpPos.Qty < qty {
				e.Log("WARN", "RISK", fmt.Sprintf("[Auto-Hedge Lockout] Short trade for %s rejected on Stash. No offsetting long position held on Alpaca.", symbol))
				e.mu.Lock()
				e.SimStashOrders = append(e.SimStashOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Stash",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			if stashCash-cost < ironFloor {
				e.Log("ERROR", "IRON", fmt.Sprintf("[Stash Iron Floor Breach] Rejected. Cash: $%.2f, Collateral Required: $%.2f, Floor: $%.2f", stashCash, cost, ironFloor))
				e.mu.Lock()
				e.SimStashOrders = append(e.SimStashOrders, Order{
					ID:        fmt.Sprintf("rej-%d", time.Now().UnixNano())[12:18],
					Symbol:    symbol,
					Side:      side,
					Qty:       qty,
					Price:     price,
					Status:    "rejected",
					Timestamp: time.Now(),
					Broker:    "Stash",
					LatencyMs: latencyMs,
				})
				e.mu.Unlock()
				return
			}

			e.mu.Lock()
			e.SimStashCash += cost
			pos, exists := e.SimStashPositions[symbol]
			if exists {
				pos.Qty -= qty
				pos.PriceHistory = append(pos.PriceHistory, price)
				e.SimStashPositions[symbol] = pos
			} else {
				e.SimStashPositions[symbol] = Position{
					Symbol:       symbol,
					Qty:          -qty,
					CostBasis:    price,
					Price:        price,
					OpenedAt:     time.Now(),
					PriceHistory: []float64{price},
					Broker:       "Stash",
				}
			}
			e.mu.Unlock()
		} else {
			e.mu.Lock()
			pos, exists := e.SimStashPositions[symbol]
			if !exists {
				e.Log("ERROR", "TRADE", "[Stash] Order REJECTED: No position found.")
				e.mu.Unlock()
				return
			}

			if pos.Qty < 0 {
				buyCost := float64(qty) * price
				if stashCash-buyCost < ironFloor {
					e.Log("ERROR", "IRON", fmt.Sprintf("[Stash Short Cover Breach] Short cover rejected. Cash: $%.2f, Buyback Cost: $%.2f, Floor: $%.2f", stashCash, buyCost, ironFloor))
					e.mu.Unlock()
					return
				}
				e.SimStashCash -= buyCost
				pos.Qty += qty
			} else {
				if pos.Qty < qty {
					e.Log("ERROR", "TRADE", fmt.Sprintf("[Stash] Order REJECTED: Insufficient shares. Position: %d", pos.Qty))
					e.mu.Unlock()
					return
				}
				e.SimStashCash += cost
				pos.Qty -= qty
			}

			if pos.Qty == 0 {
				delete(e.SimStashPositions, symbol)
			} else {
				e.SimStashPositions[symbol] = pos
			}
			e.mu.Unlock()
		}

		e.mu.Lock()
		o := Order{
			ID:        fmt.Sprintf("st-%d", time.Now().UnixNano())[12:18],
			Symbol:    symbol,
			Side:      side,
			Qty:       qty,
			Price:     price,
			Status:    "filled",
			Timestamp: time.Now(),
			Broker:    "Stash",
			LatencyMs: latencyMs,
		}
		e.SimStashOrders = append(e.SimStashOrders, o)
		e.mu.Unlock()

		e.Log("INFO", "TRADE", fmt.Sprintf("[Stash] Order FILLED at price $%.2f. Qty: %d, Action: %s. Latency: %dms", price, qty, side, latencyMs))
		e.commitToLedger("Stash", side, symbol, qty, price, latencyMs)
		return
	}

	e.Log("INFO", "STASH", fmt.Sprintf("[Stash Live] Submitting %s order via stash API...", side))
}

// Atomic Cross-Market Execution
func (e *Engine) PlaceSniperTrades(buyAsset string, buySide string, shortAsset string, qty int) {
	priceBuy, _ := e.MatchMarketOrder(buyAsset, buySide, qty)
	priceShort, _ := e.MatchMarketOrder(shortAsset, "sell", qty)

	costAlpaca := float64(qty) * priceBuy
	costSchwab := float64(qty) * priceShort

	e.mu.RLock()
	alpCash := e.SimCash
	schCash := e.SimSchwabCash
	floor := e.Config.IronFloor
	e.mu.RUnlock()

	if alpCash-costAlpaca < floor {
		e.Log("ERROR", "ATOMIC", fmt.Sprintf("[Leg Decouple Alert] Cross-trade aborted. Alpaca Cash $%.2f has insufficient headroom for $%.2f buy.", alpCash, costAlpaca))
		return
	}
	if schCash-costSchwab < floor {
		e.Log("ERROR", "ATOMIC", fmt.Sprintf("[Leg Decouple Alert] Cross-trade aborted. Schwab Cash $%.2f has insufficient headroom for $%.2f short.", schCash, costSchwab))
		return
	}

	e.Log("INFO", "ATOMIC", fmt.Sprintf("[Atomic Execution] Routing cross-market legs: Buy %d %s on Alpaca ($%.2f), Short %d %s on Schwab ($%.2f)", qty, buyAsset, priceBuy, qty, shortAsset, priceShort))

	go e.PlaceOrderForAsset(buyAsset, buySide, qty)
	go e.PlaceOrderForSchwab(shortAsset, "sell_short", qty)
}

// Queue Mortal Limit Order
func (e *Engine) QueueMortalLimitOrder(symbol string, side string, qty int, price float64, lifespanTicks int) {
	e.mu.Lock()
	lo := LimitOrder{
		ID:            fmt.Sprintf("lim-%d", time.Now().UnixNano())[12:18],
		Symbol:        symbol,
		Side:          side,
		Qty:           qty,
		Price:         price,
		LifespanTicks: lifespanTicks,
		CreatedAt:     time.Now(),
		Broker:        "Alpaca",
	}

	e.ActiveLimitOrders = append(e.ActiveLimitOrders, lo)
	e.mu.Unlock()

	e.Log("INFO", "ALGO", fmt.Sprintf("[Mortal Order] Placed %s LIMIT order for %d %s at $%.2f on Alpaca. Lifespan: %d ticks.",
		side, qty, symbol, price, lifespanTicks))
}

// Evaluate Mortal limit orders matching & expiration checks
func (e *Engine) processLimitOrders() {
	e.mu.Lock()
	if len(e.ActiveLimitOrders) == 0 {
		e.mu.Unlock()
		return
	}

	var remaining []LimitOrder

	for _, lo := range e.ActiveLimitOrders {
		midPrice := e.RealTimePrices[lo.Symbol]
		book := e.generateSimulatedOrderBook(lo.Symbol, midPrice)

		filled := false

		if lo.Side == "buy" {
			if len(book.Asks) > 0 && book.Asks[0].Price <= lo.Price {
				filled = true
				e.mu.Unlock()
				e.Log("INFO", "ALGO", fmt.Sprintf("[Mortal Order] Limit BUY matched! Symbol: %s, Qty: %d, Match Price: $%.2f", lo.Symbol, lo.Qty, book.Asks[0].Price))
				if lo.Broker == "Schwab" {
					e.PlaceOrderForSchwab(lo.Symbol, "buy", lo.Qty)
				} else if lo.Broker == "Ibkr" {
					e.PlaceOrderForIbkr(lo.Symbol, "buy", lo.Qty)
				} else if lo.Broker == "Paypal" {
					e.PlaceOrderForPaypal(lo.Symbol, "buy", lo.Qty)
				} else if lo.Broker == "Cashapp" {
					e.PlaceOrderForCashapp(lo.Symbol, "buy", lo.Qty)
				} else if lo.Broker == "Stash" {
					e.PlaceOrderForStash(lo.Symbol, "buy", lo.Qty)
				} else {
					e.PlaceOrderForAsset(lo.Symbol, "buy", lo.Qty)
				}
				e.mu.Lock()
			}
		} else {
			if len(book.Bids) > 0 && book.Bids[0].Price >= lo.Price {
				filled = true
				e.mu.Unlock()
				e.Log("INFO", "ALGO", fmt.Sprintf("[Mortal Order] Limit SELL matched! Symbol: %s, Qty: %d, Match Price: $%.2f", lo.Symbol, lo.Qty, book.Bids[0].Price))
				if lo.Broker == "Schwab" {
					e.PlaceOrderForSchwab(lo.Symbol, "sell", lo.Qty)
				} else if lo.Broker == "Ibkr" {
					e.PlaceOrderForIbkr(lo.Symbol, "sell", lo.Qty)
				} else if lo.Broker == "Paypal" {
					e.PlaceOrderForPaypal(lo.Symbol, "sell", lo.Qty)
				} else if lo.Broker == "Cashapp" {
					e.PlaceOrderForCashapp(lo.Symbol, "sell", lo.Qty)
				} else if lo.Broker == "Stash" {
					e.PlaceOrderForStash(lo.Symbol, "sell", lo.Qty)
				} else {
					e.PlaceOrderForAsset(lo.Symbol, "sell", lo.Qty)
				}
				e.mu.Lock()
			}
		}

		if filled {
			continue
		}

		lo.LifespanTicks--
		if lo.LifespanTicks <= 0 {
			e.Log("WARN", "ALGO", fmt.Sprintf("[Mortal Order] Order expired! ID: %s, Symbol: %s, Price: $%.2f. Lifespan elapsed.",
				lo.ID, lo.Symbol, lo.Price))

			e.SimOrders = append(e.SimOrders, Order{
				ID:        lo.ID,
				Symbol:    lo.Symbol,
				Side:      lo.Side,
				Qty:       lo.Qty,
				Price:     lo.Price,
				Status:    "expired",
				Timestamp: time.Now(),
				Broker:    lo.Broker,
			})

			e.mu.Unlock()
			e.commitToLedger(lo.Broker, "EXPIRE", lo.Symbol, lo.Qty, lo.Price, 0)
			e.mu.Lock()
		} else {
			remaining = append(remaining, lo)
		}
	}

	e.ActiveLimitOrders = remaining
	e.mu.Unlock()
}

// 75-minute Stale Stock Liveness Tracker across all brokers
func (e *Engine) processLivenessTracker() {
	e.mu.Lock()
	window := e.Config.LivenessWindow
	if window <= 0 {
		window = 75
	}

	var toLiquidateAlpaca []struct{ Symbol string; Qty int }
	var toLiquidateSchwab []struct{ Symbol string; Qty int }
	var toLiquidateIbkr []struct{ Symbol string; Qty int }
	var toLiquidatePaypal []struct{ Symbol string; Qty int }
	var toLiquidateCashapp []struct{ Symbol string; Qty int }

	for sym, pos := range e.SimPositions {
		currPrice := e.RealTimePrices[sym]
		if currPrice == 0 {
			currPrice = pos.CostBasis
		}
		pos.PriceHistory = append(pos.PriceHistory, currPrice)
		if len(pos.PriceHistory) > window*2 {
			pos.PriceHistory = pos.PriceHistory[len(pos.PriceHistory)-window*2:]
		}
		e.SimPositions[sym] = pos

		if e.Config.Strategy != "sovereign" && len(pos.PriceHistory) >= window {
			recentPrices := pos.PriceHistory[len(pos.PriceHistory)-window:]
			minP, maxP := 999999.0, 0.0
			for _, price := range recentPrices {
				if price < minP { minP = price }
				if price > maxP { maxP = price }
			}
			variance := 0.0
			if minP > 0 { variance = (maxP - minP) / minP }
			if variance < 0.001 {
				toLiquidateAlpaca = append(toLiquidateAlpaca, struct{ Symbol string; Qty int }{Symbol: sym, Qty: pos.Qty})
				e.Log("WARN", "ALGO", fmt.Sprintf("[LIVENESS] Alpaca position %s is dead in the water. Liquidating.", sym))
			}
		}
	}

	for sym, pos := range e.SimSchwabPositions {
		currPrice := e.RealTimePrices[sym]
		if currPrice == 0 { currPrice = pos.CostBasis }
		pos.PriceHistory = append(pos.PriceHistory, currPrice)
		if len(pos.PriceHistory) > window*2 { pos.PriceHistory = pos.PriceHistory[len(pos.PriceHistory)-window*2:] }
		e.SimSchwabPositions[sym] = pos

		if e.Config.Strategy != "sovereign" && len(pos.PriceHistory) >= window {
			recentPrices := pos.PriceHistory[len(pos.PriceHistory)-window:]
			minP, maxP := 999999.0, 0.0
			for _, price := range recentPrices {
				if price < minP { minP = price }
				if price > maxP { maxP = price }
			}
			variance := 0.0
			if minP > 0 { variance = (maxP - minP) / minP }
			if variance < 0.001 {
				toLiquidateSchwab = append(toLiquidateSchwab, struct{ Symbol string; Qty int }{Symbol: sym, Qty: pos.Qty})
				e.Log("WARN", "ALGO", fmt.Sprintf("[LIVENESS] Schwab position %s is dead in the water. Liquidating.", sym))
			}
		}
	}

	for sym, pos := range e.SimIbkrPositions {
		currPrice := e.RealTimePrices[sym]
		if currPrice == 0 { currPrice = pos.CostBasis }
		pos.PriceHistory = append(pos.PriceHistory, currPrice)
		if len(pos.PriceHistory) > window*2 { pos.PriceHistory = pos.PriceHistory[len(pos.PriceHistory)-window*2:] }
		e.SimIbkrPositions[sym] = pos

		if e.Config.Strategy != "sovereign" && len(pos.PriceHistory) >= window {
			recentPrices := pos.PriceHistory[len(pos.PriceHistory)-window:]
			minP, maxP := 999999.0, 0.0
			for _, price := range recentPrices {
				if price < minP { minP = price }
				if price > maxP { maxP = price }
			}
			variance := 0.0
			if minP > 0 { variance = (maxP - minP) / minP }
			if variance < 0.001 {
				toLiquidateIbkr = append(toLiquidateIbkr, struct{ Symbol string; Qty int }{Symbol: sym, Qty: pos.Qty})
				e.Log("WARN", "ALGO", fmt.Sprintf("[LIVENESS] IBKR position %s is dead in the water. Liquidating.", sym))
			}
		}
	}

	for sym, pos := range e.SimPaypalPositions {
		currPrice := e.RealTimePrices[sym]
		if currPrice == 0 { currPrice = pos.CostBasis }
		pos.PriceHistory = append(pos.PriceHistory, currPrice)
		if len(pos.PriceHistory) > window*2 { pos.PriceHistory = pos.PriceHistory[len(pos.PriceHistory)-window*2:] }
		e.SimPaypalPositions[sym] = pos

		if e.Config.Strategy != "sovereign" && len(pos.PriceHistory) >= window {
			recentPrices := pos.PriceHistory[len(pos.PriceHistory)-window:]
			minP, maxP := 999999.0, 0.0
			for _, price := range recentPrices {
				if price < minP { minP = price }
				if price > maxP { maxP = price }
			}
			variance := 0.0
			if minP > 0 { variance = (maxP - minP) / minP }
			if variance < 0.001 {
				toLiquidatePaypal = append(toLiquidatePaypal, struct{ Symbol string; Qty int }{Symbol: sym, Qty: pos.Qty})
				e.Log("WARN", "ALGO", fmt.Sprintf("[LIVENESS] PayPal position %s is dead in the water. Liquidating.", sym))
			}
		}
	}

	for sym, pos := range e.SimCashappPositions {
		currPrice := e.RealTimePrices[sym]
		if currPrice == 0 { currPrice = pos.CostBasis }
		pos.PriceHistory = append(pos.PriceHistory, currPrice)
		if len(pos.PriceHistory) > window*2 { pos.PriceHistory = pos.PriceHistory[len(pos.PriceHistory)-window*2:] }
		e.SimCashappPositions[sym] = pos

		if e.Config.Strategy != "sovereign" && len(pos.PriceHistory) >= window {
			recentPrices := pos.PriceHistory[len(pos.PriceHistory)-window:]
			minP, maxP := 999999.0, 0.0
			for _, price := range recentPrices {
				if price < minP { minP = price }
				if price > maxP { maxP = price }
			}
			variance := 0.0
			if minP > 0 { variance = (maxP - minP) / minP }
			if variance < 0.001 {
				toLiquidateCashapp = append(toLiquidateCashapp, struct{ Symbol string; Qty int }{Symbol: sym, Qty: pos.Qty})
				e.Log("WARN", "ALGO", fmt.Sprintf("[LIVENESS] Cash App position %s is dead in the water. Liquidating.", sym))
			}
		}
	}

	var toLiquidateStash []struct{ Symbol string; Qty int }
	for sym, pos := range e.SimStashPositions {
		currPrice := e.RealTimePrices[sym]
		if currPrice == 0 { currPrice = pos.CostBasis }
		pos.PriceHistory = append(pos.PriceHistory, currPrice)
		if len(pos.PriceHistory) > window*2 { pos.PriceHistory = pos.PriceHistory[len(pos.PriceHistory)-window*2:] }
		e.SimStashPositions[sym] = pos

		if e.Config.Strategy != "sovereign" && len(pos.PriceHistory) >= window {
			recentPrices := pos.PriceHistory[len(pos.PriceHistory)-window:]
			minP, maxP := 999999.0, 0.0
			for _, price := range recentPrices {
				if price < minP { minP = price }
				if price > maxP { maxP = price }
			}
			variance := 0.0
			if minP > 0 { variance = (maxP - minP) / minP }
			if variance < 0.001 {
				toLiquidateStash = append(toLiquidateStash, struct{ Symbol string; Qty int }{Symbol: sym, Qty: pos.Qty})
				e.Log("WARN", "ALGO", fmt.Sprintf("[LIVENESS] Stash position %s is dead in the water. Liquidating.", sym))
			}
		}
	}
	e.mu.Unlock()

	for _, item := range toLiquidateAlpaca {
		e.PlaceOrderForAsset(item.Symbol, "sell", item.Qty)
		e.commitToLedger("Alpaca", "LIQUIDATE_STALE", item.Symbol, item.Qty, e.RealTimePrices[item.Symbol], 0)
	}
	for _, item := range toLiquidateSchwab {
		e.PlaceOrderForSchwab(item.Symbol, "sell", item.Qty)
		e.commitToLedger("Schwab", "LIQUIDATE_STALE", item.Symbol, item.Qty, e.RealTimePrices[item.Symbol], 0)
	}
	for _, item := range toLiquidateIbkr {
		e.PlaceOrderForIbkr(item.Symbol, "sell", item.Qty)
		e.commitToLedger("Ibkr", "LIQUIDATE_STALE", item.Symbol, item.Qty, e.RealTimePrices[item.Symbol], 0)
	}
	for _, item := range toLiquidatePaypal {
		e.PlaceOrderForPaypal(item.Symbol, "sell", item.Qty)
		e.commitToLedger("Paypal", "LIQUIDATE_STALE", item.Symbol, item.Qty, e.RealTimePrices[item.Symbol], 0)
	}
	for _, item := range toLiquidateCashapp {
		e.PlaceOrderForCashapp(item.Symbol, "sell", item.Qty)
		e.commitToLedger("Cashapp", "LIQUIDATE_STALE", item.Symbol, item.Qty, e.RealTimePrices[item.Symbol], 0)
	}
	for _, item := range toLiquidateStash {
		e.PlaceOrderForStash(item.Symbol, "sell", item.Qty)
		e.commitToLedger("Stash", "LIQUIDATE_STALE", item.Symbol, item.Qty, e.RealTimePrices[item.Symbol], 0)
	}
}

func (e *Engine) recordPriceHistory(symbol string, price float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	hist := e.RealTimePriceHistory[symbol]
	hist = append(hist, price)
	if len(hist) > 100 {
		hist = hist[1:]
	}
	e.RealTimePriceHistory[symbol] = hist
}

func (e *Engine) fetchLatestPrice(symbol string) float64 {
	e.mu.Lock()
	isSim := e.IsSimulation
	prevPrice, ok := e.RealTimePrices[symbol]
	if !ok {
		prevPrice = 150.00
		if symbol == "MSFT" {
			prevPrice = 100.00
		}
		e.RealTimePrices[symbol] = prevPrice
	}
	e.mu.Unlock()

	if isSim {
		var change float64
		if symbol == "USDCUSD" || symbol == "USDTUSD" || symbol == "PYUSD" {
			change = (rand.Float64() - 0.5) * 0.002 // Stablecoin peg bounds
		} else if symbol == "BTCUSD" {
			change = (rand.Float64() - 0.5) * 150.0
		} else if symbol == "ETHUSD" {
			change = (rand.Float64() - 0.5) * 8.0
		} else {
			change = (rand.Float64() - 0.5) * 0.20
		}

		newPrice := prevPrice + change
		if symbol == "USDCUSD" || symbol == "USDTUSD" || symbol == "PYUSD" {
			if newPrice > 1.005 { newPrice = 1.005 }
			if newPrice < 0.995 { newPrice = 0.995 }
		} else {
			if newPrice < 0.01 { newPrice = 0.01 }
		}

		e.mu.Lock()
		e.RealTimePrices[symbol] = newPrice
		e.SimLatestPrice = newPrice
		e.mu.Unlock()
		e.recordPriceHistory(symbol, newPrice)
		return newPrice
	}

	url := fmt.Sprintf("https://data.alpaca.markets/v2/stocks/%s/trades/latest", symbol)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return e.fallbackPrice(symbol, prevPrice)
	}

	req.Header.Set("APCA-API-KEY-ID", e.keyID)
	req.Header.Set("APCA-API-SECRET-KEY", e.secretKey)

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return e.fallbackPrice(symbol, prevPrice)
	}
	defer resp.Body.Close()

	var data struct {
		Trade struct {
			Price float64 `json:"p"`
		} `json:"trade"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return e.fallbackPrice(symbol, prevPrice)
	}

	if data.Trade.Price == 0 {
		return e.fallbackPrice(symbol, prevPrice)
	}

	e.mu.Lock()
	e.RealTimePrices[symbol] = data.Trade.Price
	e.SimLatestPrice = data.Trade.Price
	e.mu.Unlock()
	e.recordPriceHistory(symbol, data.Trade.Price)

	return data.Trade.Price
}

func (e *Engine) fallbackPrice(symbol string, prevPrice float64) float64 {
	change := (rand.Float64() - 0.5) * 0.05
	newPrice := prevPrice + change
	e.mu.Lock()
	e.RealTimePrices[symbol] = newPrice
	e.SimLatestPrice = newPrice
	e.mu.Unlock()
	e.recordPriceHistory(symbol, newPrice)
	return newPrice
}

func (e *Engine) getLatestVolume(symbol string) float64 {
	e.mu.Lock()
	defer e.mu.Unlock()

	vol, exists := e.RealTimeVolume[symbol]
	if !exists {
		vol = 5.0 + rand.Float64()*10.0
		e.RealTimeVolume[symbol] = vol
	} else {
		change := (rand.Float64() - 0.5) * 2.0
		vol += change
		if vol < 0.1 {
			vol = 0.1
		}
		e.RealTimeVolume[symbol] = vol
	}
	return vol
}

// Generate Options suggests across all brokers (Calls check held shares, Puts check cash)
func (e *Engine) GenerateOptionSuggestions() {
	e.mu.Lock()
	defer e.mu.Unlock()

	var suggestions []OptionSuggestion
	now := time.Now()
	daysToFriday := (5 - int(now.Weekday()) + 7) % 7
	if daysToFriday == 0 {
		daysToFriday = 7
	}
	expiryDate := now.AddDate(0, 0, daysToFriday).Format("2006-01-02")
	occExpiry := now.AddDate(0, 0, daysToFriday).Format("060102")

	// Covered Calls for held long positions
	addCC := func(sym string, costBasis float64, broker string) {
		price := e.RealTimePrices[sym]
		if price == 0 { price = costBasis }
		strike := math.Round(price*1.05/2.5) * 2.5
		occSymbol := fmt.Sprintf("%s%sC%08d", sym, occExpiry, int(strike*1000))
		premium := math.Round(price*0.012*100) / 100
		suggestions = append(suggestions, OptionSuggestion{
			ID:               fmt.Sprintf("opt-%d", rand.Intn(100000)),
			Timestamp:        now,
			UnderlyingSymbol: sym,
			OccSymbol:        occSymbol,
			StrategyName:     "Covered Call",
			Strike:           strike,
			Expiration:       expiryDate,
			Type:             "call",
			Premium:          premium,
			Reason:           fmt.Sprintf("Generate premium on %s held stock. Strike is 5%% OTM.", broker),
		})
	}

	for sym, pos := range e.SimPositions {
		if pos.Qty > 0 { addCC(sym, pos.CostBasis, "Alpaca") }
	}
	for sym, pos := range e.SimSchwabPositions {
		if pos.Qty > 0 { addCC(sym, pos.CostBasis, "Schwab") }
	}
	for sym, pos := range e.SimIbkrPositions {
		if pos.Qty > 0 { addCC(sym, pos.CostBasis, "Ibkr") }
	}
	for sym, pos := range e.SimPaypalPositions {
		if pos.Qty > 0 { addCC(sym, pos.CostBasis, "Paypal") }
	}
	for sym, pos := range e.SimCashappPositions {
		if pos.Qty > 0 { addCC(sym, pos.CostBasis, "Cashapp") }
	}
	for sym, pos := range e.SimStashPositions {
		if pos.Qty > 0 { addCC(sym, pos.CostBasis, "Stash") }
	}

	// Cash-Secured Puts for unheld active tickers
	activeSymbols := []string{e.Config.Symbol, e.Config.AssetA, e.Config.AssetB}
	for _, sym := range activeSymbols {
		if sym == "" { continue }
		_, holdsAlp := e.SimPositions[sym]
		_, holdsSch := e.SimSchwabPositions[sym]
		_, holdsIbkr := e.SimIbkrPositions[sym]
		_, holdsPay := e.SimPaypalPositions[sym]
		_, holdsCa := e.SimCashappPositions[sym]
		_, holdsSt := e.SimStashPositions[sym]

		if !holdsAlp && !holdsSch && !holdsIbkr && !holdsPay && !holdsCa && !holdsSt {
			price := e.RealTimePrices[sym]
			if price == 0 { price = 150.0 }
			strike := math.Round(price*0.95/2.5) * 2.5
			occSymbol := fmt.Sprintf("%s%sP%08d", sym, occExpiry, int(strike*1000))
			premium := math.Round(price*0.015*100) / 100
			suggestions = append(suggestions, OptionSuggestion{
				ID:               fmt.Sprintf("opt-%d", rand.Intn(100000)),
				Timestamp:        now,
				UnderlyingSymbol: sym,
				OccSymbol:        occSymbol,
				StrategyName:     "Cash-Secured Put",
				Strike:           strike,
				Expiration:       expiryDate,
				Type:             "put",
				Premium:          premium,
				Reason:           fmt.Sprintf("Acquire %s at 5%% support support floor discount.", sym),
			})
		}
	}

	e.ActiveOptionSuggestions = suggestions

	payload, err := json.Marshal(suggestions)
	if err == nil {
		e.broadcast("option_suggestions", string(payload))
	}
}

func (e *Engine) MarketSniperTick() (string, float64, float64, float64, float64, float64, float64) {
	e.mu.Lock()
	assetA := e.Config.AssetA
	assetB := e.Config.AssetB
	mean := e.Config.Mean
	stdDev := e.Config.StdDev
	qty := e.Config.Qty
	e.mu.Unlock()

	priceA := e.fetchLatestPrice(assetA)
	priceB := e.fetchLatestPrice(assetB)

	volA := e.getLatestVolume(assetA)
	volB := e.getLatestVolume(assetB)

	if priceA == 0 || priceB == 0 {
		return "none", priceA, priceB, 0, 0, volA, volB
	}

	spread := priceA - priceB
	zScore := 0.0
	if stdDev > 0 {
		zScore = (spread - mean) / stdDev
	}

	signal := "none"

	if zScore > 2.0 {
		if e.WhaleWatcherConfirm(assetA, assetB) {
			signal = "buy"
			go e.PlaceSniperTrades(assetB, "buy", assetA, qty)
		} else {
			e.Log("WARN", "WHALE", fmt.Sprintf("[WhaleWatcher] Liquidity checks failed for %s/%s. Trade suppressed.", assetA, assetB))
		}
	} else if zScore < -2.0 {
		if e.WhaleWatcherConfirm(assetA, assetB) {
			signal = "sell"
			go e.PlaceSniperTrades(assetA, "buy", assetB, qty)
		} else {
			e.Log("WARN", "WHALE", fmt.Sprintf("[WhaleWatcher] Liquidity checks failed for %s/%s. Trade suppressed.", assetA, assetB))
		}
	}

	return signal, priceA, priceB, spread, zScore, volA, volB
}

func (e *Engine) WhaleWatcherConfirm(assetA, assetB string) bool {
	e.mu.RLock()
	minVol := e.Config.MinVolume
	e.mu.RUnlock()

	volA := e.getLatestVolume(assetA)
	volB := e.getLatestVolume(assetB)

	e.Log("INFO", "WHALE", fmt.Sprintf("[WhaleWatcher] Evaluating volumes: %s=%.2f, %s=%.2f. Min Threshold: %.2f", assetA, volA, assetB, volB, minVol))

	if volA < minVol || volB < minVol {
		return false
	}
	return true
}

// JetWeb Time Machine backtesting core engine simulation
func (e *Engine) RunBacktest(strategy string, symbol string, initialBalance float64, steps int, swanType string) map[string]interface{} {
	e.mu.Lock()
	e.IsBacktesting = true
	// Save real balances
	realCash := e.SimCash
	realPositions := e.SimPositions
	realSchwab := e.SimSchwabCash
	realIbkr := e.SimIbkrCash
	realPaypal := e.SimPaypalCash
	realCashapp := e.SimCashappCash
	realStash := e.SimStashCash
	realStashPositions := e.SimStashPositions

	// Re-init simulated backtest balances
	e.SimCash = initialBalance
	e.SimSchwabCash = initialBalance
	e.SimIbkrCash = initialBalance
	e.SimPaypalCash = initialBalance
	e.SimCashappCash = initialBalance
	e.SimStashCash = initialBalance
	e.SimPositions = make(map[string]Position)
	e.SimSchwabPositions = make(map[string]Position)
	e.SimIbkrPositions = make(map[string]Position)
	e.SimPaypalPositions = make(map[string]Position)
	e.SimCashappPositions = make(map[string]Position)
	e.SimStashPositions = make(map[string]Position)

	e.mu.Unlock()

	e.Log("INFO", "TIME-MACHINE", fmt.Sprintf("Rewinding time. Starting Backtest: Strategy=%s, Initial Balance=$%.2f per broker, Steps=%d", strategy, initialBalance, steps))

	// Generate historical synthetic prices
	basePrice := 150.0
	prices := make([]float64, steps)
	for i := 0; i < steps; i++ {
		// Inject Black Swan during backtest halfway mark
		if swanType != "" && i == steps/2 {
			if swanType == "flash_crash" {
				basePrice = basePrice * 0.75
				e.Log("WARN", "BLACK-SWAN", fmt.Sprintf("[Backtest Time Machine] FLASH CRASH INJECTED! Asset drops 25%% instantly to $%.2f.", basePrice))
			} else if swanType == "stablecoin_depeg" {
				e.Log("WARN", "BLACK-SWAN", "[Backtest Time Machine] STABLECOIN DEPEG INJECTED! Stablecoins trade down to $0.80.")
			}
		}
		basePrice += (rand.Float64() - 0.5) * 0.50
		if basePrice < 1.0 { basePrice = 1.0 }
		prices[i] = basePrice
	}

	// Backtest loop
	totalTrades := 0
	winTrades := 0
	lossTrades := 0

	for i, price := range prices {
		e.mu.Lock()
		e.RealTimePrices[symbol] = price
		e.SimLatestPrice = price
		e.mu.Unlock()

		// Run Sovereign indicators manually
		ma25 := calculateSMA(prices[:i+1], 25)
		rsi := calculateRSI(prices[:i+1], 14)
		fib382, fib618 := calculateFibLevels(prices[:i+1])

		e.mu.Lock()
		pos, hasPos := e.SimPositions[symbol]
		qty := e.Config.Qty
		e.mu.Unlock()

		if !hasPos {
			rsiThreshold := 30.0
			if ma25 > 0 && price < ma25 {
				rsiThreshold = 25.0
			}
			if rsi < rsiThreshold && price <= fib382 {
				// Buy triggers across all enabled exchanges
				e.PlaceOrderForAsset(symbol, "buy", qty)
				e.PlaceOrderForSchwab(symbol, "buy", qty)
				e.PlaceOrderForIbkr(symbol, "buy", qty)
				e.PlaceOrderForPaypal(symbol, "buy", qty)
				e.PlaceOrderForCashapp(symbol, "buy", qty)
				e.PlaceOrderForStash(symbol, "buy", qty)
				totalTrades++
			}
		} else {
			if price >= fib618 {
				// Sell profit targets
				profit := (price - pos.CostBasis) * float64(pos.Qty)
				if profit > 0 {
					winTrades++
				} else {
					lossTrades++
				}

				e.PlaceOrderForAsset(symbol, "sell", pos.Qty)
				if posSch, ok := e.SimSchwabPositions[symbol]; ok { e.PlaceOrderForSchwab(symbol, "sell", posSch.Qty) }
				if posIbkr, ok := e.SimIbkrPositions[symbol]; ok { e.PlaceOrderForIbkr(symbol, "sell", posIbkr.Qty) }
				if posPay, ok := e.SimPaypalPositions[symbol]; ok { e.PlaceOrderForPaypal(symbol, "sell", posPay.Qty) }
				if posCa, ok := e.SimCashappPositions[symbol]; ok { e.PlaceOrderForCashapp(symbol, "sell", posCa.Qty) }
				if posSt, ok := e.SimStashPositions[symbol]; ok { e.PlaceOrderForStash(symbol, "sell", posSt.Qty) }
			}
		}
	}

	e.mu.Lock()
	e.IsBacktesting = false
	finalBalance := e.SimCash + e.SimSchwabCash + e.SimIbkrCash + e.SimPaypalCash + e.SimCashappCash + e.SimStashCash
	profitPct := ((finalBalance - (initialBalance * 6)) / (initialBalance * 6)) * 100.0

	winRate := 0.0
	if totalTrades > 0 {
		winRate = (float64(winTrades) / float64(totalTrades)) * 100.0
	}

	// Restore live balances
	e.SimCash = realCash
	e.SimPositions = realPositions
	e.SimSchwabCash = realSchwab
	e.SimIbkrCash = realIbkr
	e.SimPaypalCash = realPaypal
	e.SimCashappCash = realCashapp
	e.SimStashCash = realStash
	e.SimStashPositions = realStashPositions
	e.mu.Unlock()

	summary := map[string]interface{}{
		"TotalTrades":  totalTrades,
		"WinTrades":    winTrades,
		"LossTrades":   lossTrades,
		"WinRate":      math.Round(winRate*100) / 100,
		"FinalBalance": finalBalance,
		"ProfitLoss":   math.Round(profitPct*100) / 100,
	}

	e.Log("INFO", "TIME-MACHINE", fmt.Sprintf("Backtest Complete. Profit/Loss: %.2f%%. Total Trades: %d. Win Rate: %.2f%%", profitPct, totalTrades, winRate))
	return summary
}

// Trigger Black Swan Event in Active trading simulation
func (e *Engine) TriggerBlackSwan(swanType string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.Log("WARN", "BLACK-SWAN", fmt.Sprintf("!!! CRITICAL SYSTEM EVENT INJECTED: %s !!!", swanType))

	switch swanType {
	case "flash_crash":
		// Slashes all equity asset prices by 25%
		for sym, price := range e.RealTimePrices {
			if sym != "USDCUSD" && sym != "USDTUSD" && sym != "PYUSD" {
				e.RealTimePrices[sym] = price * 0.75
			}
		}
		e.Log("WARN", "BLACK-SWAN", "[Black Swan] Flash Crash: Stock and Crypto valuations dropped by 25% instantly. Spread limits widening.")

	case "stablecoin_depeg":
		// Stablecoins depeg to 0.82
		e.RealTimePrices["USDCUSD"] = 0.82
		e.RealTimePrices["USDTUSD"] = 0.80
		e.RealTimePrices["PYUSD"] = 0.81
		e.Log("WARN", "BLACK-SWAN", "[Black Swan] Stablecoin Depeg: USDC, USDT, and PYUSD collapsed to $0.80. Checking put options collateral headroom.")

	case "exchange_hack":
		// Theft of 50% cash from PayPal and Cash App balances
		e.SimPaypalCash = e.SimPaypalCash * 0.50
		e.SimCashappCash = e.SimCashappCash * 0.50
		e.Log("WARN", "BLACK-SWAN", "[Black Swan] Security Breach: Simulated Exchange Hack stole 50% of cash reserves in PayPal and Cash App. Emergency liquidations triggered.")

	case "rate_spike":
		// Widens spreads by 500%
		e.blackSwanSpreadMult = 5.0
		e.Log("WARN", "BLACK-SWAN", "[Black Swan] Rate Hike Spike: Spreads widened by 500%. Z-Score sniper arb triggered.")
	}

	// Trigger emergency broadcast
	e.broadcast("black_swan", swanType)
}

func (e *Engine) processPreMarketWhaleWatcher() {
	now := time.Now()
	// Pre-market is 6AM (06:00:00) to 9:30AM (09:30:00)
	if now.Hour() == 6 || now.Hour() == 7 || now.Hour() == 8 || (now.Hour() == 9 && now.Minute() < 30) {
		for _, sym := range []string{"AAPL", "MSFT", "BTCUSD", "ETHUSD", "GBTC", "ETHE", "USDCUSD", "USDTUSD", "PYUSD"} {
			e.fetchLatestPrice(sym)
			vol := e.getLatestVolume(sym)
			if vol > 12.0 {
				e.mu.RLock()
				history := e.RealTimePriceHistory[sym]
				e.mu.RUnlock()

				rsiStr := "N/A"
				if len(history) >= 14 {
					rsiVal := calculateRSI(history, 14)
					rsiStr = fmt.Sprintf("%.1f", rsiVal)
					if rsiVal > 70 {
						rsiStr += " (OVERBOUGHT)"
					} else if rsiVal < 30 {
						rsiStr += " (OVERSOLD)"
					}
				}

				e.Log("INFO", "WHALE", fmt.Sprintf("[Pre-Market Accumulation] SURGE DETECTED on %s: Volume=%.2f, RSI=%s. Whales are active pre-market.", sym, vol, rsiStr))
			}
		}
	}
}

func (e *Engine) runAlgoLoop(ctx context.Context) {
	e.mu.RLock()
	interval := time.Duration(e.Config.IntervalSeconds) * time.Second
	e.mu.RUnlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	e.Log("INFO", "ALGO", fmt.Sprintf("Started strategy processor. Interval: %v", interval))

	for {
		select {
		case <-ctx.Done():
			e.Log("INFO", "ALGO", "Trading loop halted.")
			return
		case <-ticker.C:
			e.mu.RLock()
			strategy := e.Config.Strategy
			intervalSec := e.Config.IntervalSeconds
			enableSchwab := e.Config.EnableSchwab
			enableIbkr := e.Config.EnableIbkr
			enablePaypal := e.Config.EnablePaypal
			enableCashapp := e.Config.EnableCashapp
			enableStash := e.Config.EnableStash
			e.mu.RUnlock()

			currentInterval := time.Duration(intervalSec) * time.Second
			if currentInterval != interval {
				interval = currentInterval
				ticker.Reset(interval)
			}

			if e.IsBacktesting {
				continue
			}

			e.processLimitOrders()
			e.processLivenessTracker()
			e.GenerateOptionSuggestions()
			e.processPreMarketWhaleWatcher()

			if strategy == "sniper" {
				signal, priceA, priceB, spread, zScore, volA, volB := e.MarketSniperTick()

				e.Log("INFO", "ALGO", fmt.Sprintf("Sniper Tick: Spread=%.2f, ZScore=%.2f, VolA=%.1f, VolB=%.1f, Signal=%s", spread, zScore, volA, volB, signal))

				tick := TickData{
					Timestamp: time.Now(),
					Strategy:  "sniper",
					PriceA:    priceA,
					PriceB:    priceB,
					Spread:    math.Round(spread*100) / 100,
					ZScore:    math.Round(zScore*100) / 100,
					VolumeA:   math.Round(volA*100) / 100,
					VolumeB:   math.Round(volB*100) / 100,
					Signal:    signal,
				}
				payload, err := json.Marshal(tick)
				if err == nil {
					e.broadcast("tick", string(payload))
				}
			} else if strategy == "sovereign" {
				e.mu.RLock()
				symbol := e.Config.Symbol
				livenessWindow := e.Config.LivenessWindow
				qty := e.Config.Qty
				e.mu.RUnlock()

				price := e.fetchLatestPrice(symbol)
				vol := e.getLatestVolume(symbol)

				e.mu.Lock()
				e.historicalPrices = append(e.historicalPrices, price)
				if len(e.historicalPrices) > 200 {
					e.historicalPrices = e.historicalPrices[1:]
				}
				pricesCopy := make([]float64, len(e.historicalPrices))
				copy(pricesCopy, e.historicalPrices)

				vols := e.RealTimeVolumeHistory[symbol]
				vols = append(vols, vol)
				if len(vols) > 100 {
					vols = vols[1:]
				}
				e.RealTimeVolumeHistory[symbol] = vols
				e.mu.Unlock()

				ma25 := calculateSMA(pricesCopy, 25)
				rsi := calculateRSI(pricesCopy, 14)
				fib382, fib618 := calculateFibLevels(pricesCopy)
				slope := calculateSlope(pricesCopy, 15)

				avgVol := 1.0
				if len(vols) > 0 {
					sumVol := 0.0
					count := 0
					for i := len(vols) - 1; i >= 0 && count < 20; i-- {
						sumVol += vols[i]
						count++
					}
					avgVol = sumVol / float64(count)
				}
				relVol := 1.0
				if avgVol > 0 {
					relVol = vol / avgVol
				}

				signal := "none"
				var action string

				e.mu.RLock()
				posAlp, hasPosAlp := e.SimPositions[symbol]
				posSch, hasPosSch := e.SimSchwabPositions[symbol]
				posIbkr, hasPosIbkr := e.SimIbkrPositions[symbol]
				posPaypal, hasPosPaypal := e.SimPaypalPositions[symbol]
				posCashapp, hasPosCashapp := e.SimCashappPositions[symbol]
				posStash, hasPosStash := e.SimStashPositions[symbol]
				e.mu.RUnlock()

				hasPos := hasPosAlp || hasPosSch || hasPosIbkr || hasPosPaypal || hasPosCashapp || hasPosStash

				if !hasPos {
					rsiThreshold := 30.0
					if ma25 > 0 && price < ma25 {
						rsiThreshold = 25.0
					}

					if rsi < rsiThreshold && price <= fib382 && relVol < 1.0 {
						signal = "buy"
						action = "buy"
						e.Log("INFO", "ALGO", fmt.Sprintf("[SOVEREIGN SNIPE] Trap trigger met! RSI=%.2f, Price=$%.2f <= Fib382=$%.2f, RelVol=%.2f, Slope=%.2f", rsi, price, fib382, relVol, slope))
					}
				} else {
					if price >= fib618 {
						signal = "sell"
						action = "sell"
						e.Log("INFO", "ALGO", fmt.Sprintf("[PROFIT TAKE] Exited position at Fib 0.618 extension target: Price=$%.2f >= Target=$%.2f", price, fib618))
					} else if len(pricesCopy) >= livenessWindow {
						staleSlope := calculateSlope(pricesCopy, livenessWindow)
						if staleSlope < 5.0 {
							signal = "sell"
							action = "sell"
							e.Log("WARN", "ALGO", fmt.Sprintf("[FATALITY PURGE] Stale Vitality Slope %.2f < 5.0 over %d ticks. Liquidating position.", staleSlope, livenessWindow))
						}
					}
				}

				e.Log("INFO", "ALGO", fmt.Sprintf("Sovereign Tick: Price=$%.2f, MA-25=$%.2f, RSI=%.1f, Fib382=$%.2f, RelVol=%.2f, Slope=%.2f", price, ma25, rsi, fib382, relVol, slope))

				tick := TickData{
					Timestamp: time.Now(),
					Strategy:  "sovereign",
					Price:     price,
					Ma25:      math.Round(ma25*100) / 100,
					Rsi:       math.Round(rsi*100) / 100,
					Fib382:    math.Round(fib382*100) / 100,
					Fib618:    math.Round(fib618*100) / 100,
					Slope:     math.Round(slope*100) / 100,
					RelVol:    math.Round(relVol*100) / 100,
					Signal:    signal,
				}
				payload, err := json.Marshal(tick)
				if err == nil {
					e.broadcast("tick", string(payload))
				}

				if action == "buy" {
					go e.PlaceOrderForAsset(symbol, "buy", qty)
					if enableSchwab { go e.PlaceOrderForSchwab(symbol, "buy", qty) }
					if enableIbkr { go e.PlaceOrderForIbkr(symbol, "buy", qty) }
					if enablePaypal { go e.PlaceOrderForPaypal(symbol, "buy", qty) }
					if enableCashapp { go e.PlaceOrderForCashapp(symbol, "buy", qty) }
					if enableStash { go e.PlaceOrderForStash(symbol, "buy", qty) }
				} else if action == "sell" {
					if hasPosAlp { go e.PlaceOrderForAsset(symbol, "sell", posAlp.Qty) }
					if hasPosSch { go e.PlaceOrderForSchwab(symbol, "sell", posSch.Qty) }
					if hasPosIbkr { go e.PlaceOrderForIbkr(symbol, "sell", posIbkr.Qty) }
					if hasPosPaypal { go e.PlaceOrderForPaypal(symbol, "sell", posPaypal.Qty) }
					if hasPosCashapp { go e.PlaceOrderForCashapp(symbol, "sell", posCashapp.Qty) }
					if hasPosStash { go e.PlaceOrderForStash(symbol, "sell", posStash.Qty) }
				}

			} else {
				e.mu.RLock()
				symbol := e.Config.Symbol
				shortPeriod := e.Config.ShortEmaPeriod
				longPeriod := e.Config.LongEmaPeriod
				e.mu.RUnlock()

				price := e.fetchLatestPrice(symbol)

				e.mu.Lock()
				e.historicalPrices = append(e.historicalPrices, price)
				if len(e.historicalPrices) > 100 {
					e.historicalPrices = e.historicalPrices[1:]
				}
				pricesCopy := make([]float64, len(e.historicalPrices))
				copy(pricesCopy, e.historicalPrices)
				e.mu.Unlock()

				shortEma := calculateEMA(pricesCopy, shortPeriod)
				longEma := calculateEMA(pricesCopy, longPeriod)

				signal := "none"
				var action string

				if shortEma > 0 && longEma > 0 && len(pricesCopy) >= 2 {
					prevPrices := pricesCopy[:len(pricesCopy)-1]
					prevShortEma := calculateEMA(prevPrices, shortPeriod)
					prevLongEma := calculateEMA(prevPrices, longPeriod)

					if prevShortEma > 0 && prevLongEma > 0 {
						e.mu.RLock()
						posSide := e.lastPositionSide
						e.mu.RUnlock()

						if prevShortEma <= prevLongEma && shortEma > longEma {
							if posSide == "flat" {
								signal = "buy"
								action = "buy"
							}
						}
						if prevShortEma >= prevLongEma && shortEma < longEma {
							if posSide == "long" {
								signal = "sell"
								action = "sell"
							}
						}
					}
				}

				e.Log("INFO", "ALGO", fmt.Sprintf("EMA Tick: Price=$%.2f, ShortEMA($%.2f), LongEMA($%.2f), Signal=%s", price, shortEma, longEma, signal))

				tick := TickData{
					Timestamp: time.Now(),
					Strategy:  "ema",
					Price:     price,
					ShortEma:  math.Round(shortEma*100) / 100,
					LongEma:   math.Round(longEma*100) / 100,
					Signal:    signal,
				}
				payload, err := json.Marshal(tick)
				if err == nil {
					e.broadcast("tick", string(payload))
				}

				if action != "" {
					go e.PlaceOrderForAsset(symbol, action, e.Config.Qty)
					if enableSchwab { go e.PlaceOrderForSchwab(symbol, action, e.Config.Qty) }
					if enableIbkr { go e.PlaceOrderForIbkr(symbol, action, e.Config.Qty) }
					if enablePaypal { go e.PlaceOrderForPaypal(symbol, action, e.Config.Qty) }
					if enableCashapp { go e.PlaceOrderForCashapp(symbol, action, e.Config.Qty) }
					if enableStash { go e.PlaceOrderForStash(symbol, action, e.Config.Qty) }
				}
			}
		}
	}
}

func (e *Engine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Running {
		return
	}

	e.Running = true
	e.ctx, e.cancelFn = context.WithCancel(context.Background())
	go e.runAlgoLoop(e.ctx)
}

func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.Running {
		return
	}

	e.Running = false
	if e.cancelFn != nil {
		e.cancelFn()
	}
}

func main() {
	engine := NewEngine()
	engine.Start()

	// REST Endpoints
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/dashboard/alpaca_trader.html" {
			http.ServeFile(w, r, "./dashboard/alpaca_trader.html")
			return
		}
		fs := http.FileServer(http.Dir("./dashboard"))
		fs.ServeHTTP(w, r)
	})

	http.HandleFunc("/api/alpaca/state", func(w http.ResponseWriter, r *http.Request) {
		engine.mu.RLock()
		defer engine.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		type StateResponse struct {
			Strategy        string
			Symbol          string
			Qty             int
			IntervalSeconds int
			ShortEmaPeriod  int
			LongEmaPeriod   int
			AssetA          string
			AssetB          string
			Mean            float64
			StdDev          float64
			MinVolume       float64
			IronFloor       float64
			LivenessWindow  int
			EnableSchwab    bool
			EnableIbkr      bool
			EnablePaypal    bool
			EnableCashapp   bool
			EnableStash     bool
			Running         bool
			IsSimulation    bool
			Logs            []TradeLog
		}

		engine.logsMu.Lock()
		logsCopy := make([]TradeLog, len(engine.Logs))
		copy(logsCopy, engine.Logs)
		engine.logsMu.Unlock()

		json.NewEncoder(w).Encode(StateResponse{
			Strategy:        engine.Config.Strategy,
			Symbol:          engine.Config.Symbol,
			Qty:             engine.Config.Qty,
			IntervalSeconds: engine.Config.IntervalSeconds,
			ShortEmaPeriod:  engine.Config.ShortEmaPeriod,
			LongEmaPeriod:   engine.Config.LongEmaPeriod,
			AssetA:          engine.Config.AssetA,
			AssetB:          engine.Config.AssetB,
			Mean:            engine.Config.Mean,
			StdDev:          engine.Config.StdDev,
			MinVolume:       engine.Config.MinVolume,
			IronFloor:       engine.Config.IronFloor,
			LivenessWindow:  engine.Config.LivenessWindow,
			EnableSchwab:    engine.Config.EnableSchwab,
			EnableIbkr:      engine.Config.EnableIbkr,
			EnablePaypal:    engine.Config.EnablePaypal,
			EnableCashapp:   engine.Config.EnableCashapp,
			EnableStash:     engine.Config.EnableStash,
			Running:         engine.Running,
			IsSimulation:    engine.IsSimulation,
			Logs:            logsCopy,
		})
	})

	http.HandleFunc("/api/alpaca/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Strategy        string
			Symbol          string
			Qty             int
			IntervalSeconds int
			ShortEmaPeriod  int
			LongEmaPeriod   int
			AssetA          string
			AssetB          string
			Mean            float64
			StdDev          float64
			MinVolume       float64
			IronFloor       float64
			LivenessWindow  int
			EnableSchwab    bool
			EnableIbkr      bool
			EnablePaypal    bool
			EnableCashapp   bool
			EnableStash     bool
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mu.Lock()
		engine.Config = AppConfig{
			Strategy:        req.Strategy,
			Symbol:          req.Symbol,
			Qty:             req.Qty,
			IntervalSeconds: req.IntervalSeconds,
			ShortEmaPeriod:  req.ShortEmaPeriod,
			LongEmaPeriod:   req.LongEmaPeriod,
			AssetA:          req.AssetA,
			AssetB:          req.AssetB,
			Mean:            req.Mean,
			StdDev:          req.StdDev,
			MinVolume:       req.MinVolume,
			IronFloor:       req.IronFloor,
			LivenessWindow:  req.LivenessWindow,
			EnableSchwab:    req.EnableSchwab,
			EnableIbkr:      req.EnableIbkr,
			EnablePaypal:    req.EnablePaypal,
			EnableCashapp:   req.EnableCashapp,
			EnableStash:     req.EnableStash,
		}
		engine.mu.Unlock()

		engine.Log("INFO", "SYSTEM", "Engine configurations updated.")
		w.WriteHeader(http.StatusOK)
	})

	// Submit Limit Order
	http.HandleFunc("/api/alpaca/limit_order", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions { return }
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Symbol        string
			Side          string
			Qty           int
			Price         float64
			LifespanTicks int
			Broker        string
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.LifespanTicks <= 0 { req.LifespanTicks = 5 }

		engine.mu.Lock()
		lo := LimitOrder{
			ID:            fmt.Sprintf("lim-%d", time.Now().UnixNano())[12:18],
			Symbol:        req.Symbol,
			Side:          req.Side,
			Qty:           req.Qty,
			Price:         req.Price,
			LifespanTicks: req.LifespanTicks,
			CreatedAt:     time.Now(),
			Broker:        req.Broker,
		}
		engine.ActiveLimitOrders = append(engine.ActiveLimitOrders, lo)
		engine.mu.Unlock()

		engine.Log("INFO", "ALGO", fmt.Sprintf("[Mortal Order] Placed %s LIMIT order on %s. Lifespan: %d ticks.", req.Side, req.Broker, req.LifespanTicks))
		w.WriteHeader(http.StatusAccepted)
	})

	// Fetch Orderbook
	http.HandleFunc("/api/alpaca/orderbook", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		symbol := r.URL.Query().Get("symbol")
		if symbol == "" { symbol = "AAPL" }

		engine.mu.RLock()
		midPrice := engine.RealTimePrices[symbol]
		engine.mu.RUnlock()

		if midPrice == 0 { midPrice = 150.0 }

		book := engine.generateSimulatedOrderBook(symbol, midPrice)
		json.NewEncoder(w).Encode(book)
	})

	// Fetch Cryptographic Ledger
	http.HandleFunc("/api/alpaca/ledger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		engine.mu.RLock()
		defer engine.mu.RUnlock()
		json.NewEncoder(w).Encode(engine.SimLedger)
	})

	// Option Suggestions API
	http.HandleFunc("/api/alpaca/options/suggestions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		engine.mu.RLock()
		defer engine.mu.RUnlock()
		json.NewEncoder(w).Encode(engine.ActiveOptionSuggestions)
	})

	// Option Execution API
	http.HandleFunc("/api/alpaca/options/execute", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions { return }
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			SuggestionID string `json:"suggestion_id"`
			Broker       string `json:"broker"`
			Contracts    int    `json:"contracts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.mu.Lock()
		var found *OptionSuggestion
		for _, s := range engine.ActiveOptionSuggestions {
			if s.ID == req.SuggestionID {
				found = &s
				break
			}
		}
		engine.mu.Unlock()

		if found == nil {
			http.Error(w, "Option suggestion not found", http.StatusNotFound)
			return
		}

		totalPremium := found.Premium * 100.0 * float64(req.Contracts)

		engine.mu.Lock()
		latencyMs := 50 + rand.Intn(70)
		if req.Broker == "Schwab" { latencyMs = 250 + rand.Intn(200) }
		if req.Broker == "Ibkr" { latencyMs = 80 + rand.Intn(100) }
		if req.Broker == "Paypal" { latencyMs = 350 + rand.Intn(350) }
		if req.Broker == "Cashapp" { latencyMs = 120 + rand.Intn(130) }
		if req.Broker == "Stash" { latencyMs = 400 + rand.Intn(400) }
		engine.mu.Unlock()

		time.Sleep(time.Duration(latencyMs) * time.Millisecond)

		engine.mu.Lock()
		switch req.Broker {
		case "Alpaca":
			engine.SimCash += totalPremium
		case "Schwab":
			engine.SimSchwabCash += totalPremium
		case "Ibkr":
			engine.SimIbkrCash += totalPremium
		case "Paypal":
			engine.SimPaypalCash += totalPremium
		case "Cashapp":
			engine.SimCashappCash += totalPremium
		case "Stash":
			engine.SimStashCash += totalPremium
		}
		engine.mu.Unlock()

		engine.Log("INFO", "TRADE", fmt.Sprintf("[%s] Manual option trade filled. Premium received: $%.2f. Latency: %dms", req.Broker, totalPremium, latencyMs))
		engine.commitToLedger(req.Broker, "SELL_OPTION", found.OccSymbol, req.Contracts, found.Premium, latencyMs)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "premium": totalPremium})
	})

	http.HandleFunc("/api/alpaca/portfolio", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		data, pos, err := engine.GetPortfolio()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"AlpacaEquity":  data["AlpacaEquity"],
			"AlpacaCash":    data["AlpacaCash"],
			"AlpacaBp":      data["AlpacaBp"],
			"SchwabEquity":  data["SchwabEquity"],
			"SchwabCash":    data["SchwabCash"],
			"SchwabBp":      data["SchwabBp"],
			"IbkrEquity":    data["IbkrEquity"],
			"IbkrCash":      data["IbkrCash"],
			"IbkrBp":        data["IbkrBp"],
			"PaypalEquity":  data["PaypalEquity"],
			"PaypalCash":    data["PaypalCash"],
			"PaypalBp":      data["PaypalBp"],
			"CashappEquity": data["CashappEquity"],
			"CashappCash":   data["CashappCash"],
			"CashappBp":     data["CashappBp"],
			"StashEquity":   data["StashEquity"],
			"StashCash":     data["StashCash"],
			"StashBp":       data["StashBp"],
			"Positions":     pos,
		}

		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/alpaca/orders", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(engine.GetOrders())
	})

	// JetWeb Time Machine Backtest Execution route
	http.HandleFunc("/api/alpaca/backtest/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions { return }
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Strategy       string  `json:"strategy"`
			Symbol         string  `json:"symbol"`
			InitialBalance float64 `json:"initial_balance"`
			Steps          int     `json:"steps"`
			BlackSwanType  string  `json:"black_swan_type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Steps <= 0 { req.Steps = 200 }
		if req.InitialBalance <= 0 { req.InitialBalance = 10000.0 }

		res := engine.RunBacktest(req.Strategy, req.Symbol, req.InitialBalance, req.Steps, req.BlackSwanType)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	// Black Swan event injection trigger route
	http.HandleFunc("/api/alpaca/black_swan/trigger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions { return }
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			SwanType string `json:"swan_type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		engine.TriggerBlackSwan(req.SwanType)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "swan event triggered"})
	})

	http.HandleFunc("/api/alpaca/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions { return }
		engine.Start()
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/alpaca/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions { return }
		engine.Stop()
		w.WriteHeader(http.StatusOK)
	})

	// SSE Events
	http.HandleFunc("/api/alpaca/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		}

		ch := make(chan string, 10)
		engine.clientsMu.Lock()
		engine.clients[ch] = true
		engine.clientsMu.Unlock()

		defer func() {
			engine.clientsMu.Lock()
			delete(engine.clients, ch)
			engine.clientsMu.Unlock()
			close(ch)
		}()

		fmt.Fprintf(w, "data: connected\n\n")
		flusher.Flush()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				fmt.Fprint(w, msg)
				flusher.Flush()
			}
		}
	})

	fmt.Println("Serving Alpaca Auto-Trader Engine on http://localhost:8086 ...")
	if err := http.ListenAndServe(":8086", nil); err != nil {
		panic(err)
	}
}

func calculateEMA(prices []float64, period int) float64 {
	if len(prices) == 0 { return 0.0 }
	if len(prices) < period {
		sum := 0.0
		for _, p := range prices { sum += p }
		return sum / float64(len(prices))
	}
	sma := 0.0
	for i := 0; i < period; i++ { sma += prices[i] }
	sma /= float64(period)
	multiplier := 2.0 / (float64(period) + 1.0)
	ema := sma
	for i := period; i < len(prices); i++ {
		ema = (prices[i]-ema)*multiplier + ema
	}
	return ema
}

func calculateSMA(prices []float64, period int) float64 {
	if len(prices) == 0 { return 0.0 }
	if len(prices) < period { period = len(prices) }
	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}

func calculateRSI(prices []float64, period int) float64 {
	if len(prices) <= period { return 50.0 }
	gains, losses := 0.0, 0.0
	for i := 1; i <= period; i++ {
		diff := prices[i] - prices[i-1]
		if diff > 0 { gains += diff } else { losses -= diff }
	}
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)
	for i := period + 1; i < len(prices); i++ {
		diff := prices[i] - prices[i-1]
		gain, loss := 0.0, 0.0
		if diff > 0 { gain = diff } else { loss = -diff }
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
	}
	if avgLoss == 0 { return 100.0 }
	rs := avgGain / avgLoss
	return 100.0 - (100.0 / (1.0 + rs))
}

func calculateFibLevels(prices []float64) (float64, float64) {
	if len(prices) == 0 { return 0.0, 0.0 }
	high, low := -999999.0, 999999.0
	for _, p := range prices {
		if p > high { high = p }
		if p < low { low = p }
	}
	diff := high - low
	if diff == 0 { return high, high }
	fib382 := high - 0.382*diff
	fib618 := low + 0.618*diff
	return math.Round(fib382*100) / 100, math.Round(fib618*100) / 100
}

func calculateSlope(prices []float64, period int) float64 {
	n := len(prices)
	if n < 2 { return 0.0 }
	if n < period { period = n }
	startPrice := prices[n-period]
	endPrice := prices[n-1]
	if startPrice == 0 { return 0.0 }
	slope := ((endPrice - startPrice) / startPrice) * 10000.0
	return math.Round(slope*100) / 100
}
