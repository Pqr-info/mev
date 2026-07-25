package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"
)

// In-Memory state matching substrate.portfolio
type Position struct {
	Symbol    string  `json:"symbol"`
	Qty       float64 `json:"qty"`
	CostBasis float64 `json:"cost_basis"`
}

type Portfolio struct {
	Cash      float64             `json:"cash"`
	Positions map[string]Position `json:"positions"`
}

// Biographer Lineage Entry definition
type TradeMutation struct {
	Kind      string  `json:"kind"`
	Symbol    string  `json:"symbol"`
	Qty       float64 `json:"qty"`
	CostBasis float64 `json:"cost_basis"`
}

type TradeVerdict struct {
	Approved  bool   `json:"approved"`
	Rationale string `json:"rationale"`
}

type TradeEntry struct {
	LineageID string        `json:"lineage_id"`
	Timestamp string        `json:"timestamp"`
	Mutation  TradeMutation `json:"mutation"`
	Verdict   TradeVerdict  `json:"verdict"`
}

type EventBroadcaster struct {
	mu          sync.Mutex
	subscribers map[chan string]bool
}

func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{
		subscribers: make(map[chan string]bool),
	}
}

func (b *EventBroadcaster) Subscribe() chan string {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan string, 10)
	b.subscribers[ch] = true
	return ch
}

func (b *EventBroadcaster) Unsubscribe(ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.subscribers, ch)
	close(ch)
}

func (b *EventBroadcaster) Broadcast(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subscribers {
		select {
		case ch <- msg:
		default:
			// Buffer full, drop msg
		}
	}
}

func (b *EventBroadcaster) HandleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "data: {\"type\":\"system\",\"status\":\"connected\"}\n\n")
	flusher.Flush()

	notify := r.Context().Done()

	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-notify:
			return
		}
	}
}

type RiskProfile struct {
	MaxLeverage     float64            `json:"max_leverage"`
	CurrentLeverage float64            `json:"current_leverage"`
	SymbolExposure  map[string]float64 `json:"symbol_exposure"`
}

type PortfolioResponse struct {
	Cash        float64     `json:"cash"`
	Positions   []Position  `json:"positions"`
	RiskProfile RiskProfile `json:"risk_profile"`
}

type PortfolioSnapshot struct {
	Timestamp   time.Time         `json:"timestamp"`
	LineageID   string            `json:"lineage_id"`
	Portfolio   PortfolioResponse `json:"portfolio"`
}

var (
	portfolioMu sync.Mutex
	portfolio   = Portfolio{
		Cash:      114995.70, // Matches initial UI Net Liq
		Positions: make(map[string]Position),
	}

	tradeHistoryMu sync.Mutex
	tradeHistory   = []TradeEntry{}

	broadcaster = NewEventBroadcaster()

	snapshotsMu sync.Mutex
	snapshots   = []PortfolioSnapshot{}
)

func addSnapshot(lineageID string, port PortfolioResponse) {
	snapshotsMu.Lock()
	defer snapshotsMu.Unlock()
	snapshots = append(snapshots, PortfolioSnapshot{
		Timestamp: time.Now(),
		LineageID: lineageID,
		Portfolio: port,
	})
}

type ExecuteRequest struct {
	Symbol    string  `json:"symbol"`
	Qty       float64 `json:"qty"`
	CostBasis float64 `json:"cost_basis"`
	Kind      string  `json:"kind"` // "portfolio_add" or "portfolio_remove"
}

type ExecuteResponse struct {
	Approved  bool   `json:"approved"`
	Rationale string `json:"rationale"`
	LineageID string `json:"lineage_id"`
}

func handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	portfolioMu.Lock()
	defer portfolioMu.Unlock()

	// Helper to generate lineage ID
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s-%f-%d", req.Symbol, req.Qty, time.Now().UnixNano())))
	lineageID := hex.EncodeToString(hasher.Sum(nil))[:16]

	// 1. Article II Validation Checks
	// Forbidden instrument check
	if req.Symbol == "FORBIDDEN_COIN" || req.Symbol == "SCAM_TOKEN" {
		rationale := fmt.Sprintf("Article II violation: symbol %s is forbidden", req.Symbol)
		
		recordHistory(TradeEntry{
			LineageID: "rejected-" + lineageID[:8],
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Mutation: TradeMutation{
				Kind:      req.Kind,
				Symbol:    req.Symbol,
				Qty:       req.Qty,
				CostBasis: req.CostBasis,
			},
			Verdict: TradeVerdict{
				Approved:  false,
				Rationale: rationale,
			},
		})

		verdictJSON, _ := json.Marshal(map[string]interface{}{
			"type":       "verdict",
			"approved":   false,
			"rationale":  rationale,
			"lineage_id": "rejected-" + lineageID[:8],
		})
		broadcaster.Broadcast(string(verdictJSON))

		json.NewEncoder(w).Encode(ExecuteResponse{
			Approved:  false,
			Rationale: rationale,
		})
		return
	}

	// Calculate trade cost
	tradeCost := req.Qty * req.CostBasis

	if req.Kind == "portfolio_add" {
		// Cash check
		if tradeCost > portfolio.Cash {
			rationale := fmt.Sprintf("Article II violation: insufficient cash (cost: %f, held: %f)", tradeCost, portfolio.Cash)
			
			recordHistory(TradeEntry{
				LineageID: "rejected-" + lineageID[:8],
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Mutation: TradeMutation{
					Kind:      req.Kind,
					Symbol:    req.Symbol,
					Qty:       req.Qty,
					CostBasis: req.CostBasis,
				},
				Verdict: TradeVerdict{
					Approved:  false,
					Rationale: rationale,
				},
			})

			verdictJSON, _ := json.Marshal(map[string]interface{}{
				"type":       "verdict",
				"approved":   false,
				"rationale":  rationale,
				"lineage_id": "rejected-" + lineageID[:8],
			})
			broadcaster.Broadcast(string(verdictJSON))

			json.NewEncoder(w).Encode(ExecuteResponse{
				Approved:  false,
				Rationale: rationale,
			})
			return
		}

		// Leverage check (Max 3x leverage)
		totalEquity := 0.0
		for _, pos := range portfolio.Positions {
			totalEquity += pos.Qty * pos.CostBasis
		}
		totalEquity += tradeCost
		totalValue := portfolio.Cash + totalEquity - tradeCost
		if totalValue > 0 && (totalEquity/totalValue) > 3.0 {
			rationale := "Article II violation: projected leverage would exceed max limit (3.0x)"
			
			recordHistory(TradeEntry{
				LineageID: "rejected-" + lineageID[:8],
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Mutation: TradeMutation{
					Kind:      req.Kind,
					Symbol:    req.Symbol,
					Qty:       req.Qty,
					CostBasis: req.CostBasis,
				},
				Verdict: TradeVerdict{
					Approved:  false,
					Rationale: rationale,
				},
			})

			verdictJSON, _ := json.Marshal(map[string]interface{}{
				"type":       "verdict",
				"approved":   false,
				"rationale":  rationale,
				"lineage_id": "rejected-" + lineageID[:8],
			})
			broadcaster.Broadcast(string(verdictJSON))

			json.NewEncoder(w).Encode(ExecuteResponse{
				Approved:  false,
				Rationale: rationale,
			})
			return
		}

		// Mutate state
		portfolio.Cash -= tradeCost
		pos, exists := portfolio.Positions[req.Symbol]
		if exists {
			totalCost := (pos.Qty * pos.CostBasis) + tradeCost
			pos.Qty += req.Qty
			pos.CostBasis = totalCost / pos.Qty
			portfolio.Positions[req.Symbol] = pos
		} else {
			portfolio.Positions[req.Symbol] = Position{
				Symbol:    req.Symbol,
				Qty:       req.Qty,
				CostBasis: req.CostBasis,
			}
		}
	} else if req.Kind == "portfolio_remove" {
		pos, exists := portfolio.Positions[req.Symbol]
		if !exists || pos.Qty < req.Qty {
			rationale := "Article II violation: insufficient quantity held for removal"
			
			recordHistory(TradeEntry{
				LineageID: "rejected-" + lineageID[:8],
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Mutation: TradeMutation{
					Kind:      req.Kind,
					Symbol:    req.Symbol,
					Qty:       req.Qty,
					CostBasis: req.CostBasis,
				},
				Verdict: TradeVerdict{
					Approved:  false,
					Rationale: rationale,
				},
			})

			verdictJSON, _ := json.Marshal(map[string]interface{}{
				"type":       "verdict",
				"approved":   false,
				"rationale":  rationale,
				"lineage_id": "rejected-" + lineageID[:8],
			})
			broadcaster.Broadcast(string(verdictJSON))

			json.NewEncoder(w).Encode(ExecuteResponse{
				Approved:  false,
				Rationale: rationale,
			})
			return
		}

		// Mutate state
		portfolio.Cash += tradeCost
		pos.Qty -= req.Qty
		if pos.Qty <= 0 {
			delete(portfolio.Positions, req.Symbol)
		} else {
			portfolio.Positions[req.Symbol] = pos
		}
	} else {
		json.NewEncoder(w).Encode(ExecuteResponse{
			Approved:  false,
			Rationale: "Invalid mutation kind",
		})
		return
	}

	// Successful trade record
	recordHistory(TradeEntry{
		LineageID: lineageID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Mutation: TradeMutation{
			Kind:      req.Kind,
			Symbol:    req.Symbol,
			Qty:       req.Qty,
			CostBasis: req.CostBasis,
		},
		Verdict: TradeVerdict{
			Approved:  true,
			Rationale: "Consensus approved (TRADER lane)",
		},
	})

	// Broadcast verdict
	verdictJSON, _ := json.Marshal(map[string]interface{}{
		"type":       "verdict",
		"approved":   true,
		"rationale":  "Consensus approved (TRADER lane)",
		"lineage_id": lineageID,
	})
	broadcaster.Broadcast(string(verdictJSON))

	// Broadcast updated portfolio
	sendPortfolioBroadcast()

	// Capture temporal snapshot
	addSnapshot(lineageID, getPortfolioState())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ExecuteResponse{
		Approved:  true,
		Rationale: "Consensus approved (TRADER lane)",
		LineageID: lineageID,
	})
}

func recordHistory(entry TradeEntry) {
	tradeHistoryMu.Lock()
	tradeHistory = append([]TradeEntry{entry}, tradeHistory...)
	tradeHistoryMu.Unlock()

	entryJSON, _ := json.Marshal(entry)
	broadcaster.Broadcast(fmt.Sprintf(`{"type":"lineage","entry":%s}`, string(entryJSON)))
}

func handleGetTradeHistory(w http.ResponseWriter, r *http.Request) {
	tradeHistoryMu.Lock()
	defer tradeHistoryMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tradeHistory)
}

func handleGetPortfolio(w http.ResponseWriter, r *http.Request) {
	portfolioMu.Lock()
	defer portfolioMu.Unlock()

	resp := getPortfolioState()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getPortfolioState() PortfolioResponse {
	var positions []Position
	symbolExposure := make(map[string]float64)
	totalEquity := 0.0

	for _, pos := range portfolio.Positions {
		positions = append(positions, pos)
		exposure := pos.Qty * pos.CostBasis
		symbolExposure[pos.Symbol] = exposure
		totalEquity += exposure
	}

	netLiq := portfolio.Cash + totalEquity
	currentLeverage := 0.0
	if netLiq > 0 {
		currentLeverage = totalEquity / netLiq
	}

	return PortfolioResponse{
		Cash:      portfolio.Cash,
		Positions: positions,
		RiskProfile: RiskProfile{
			MaxLeverage:     3.0,
			CurrentLeverage: currentLeverage,
			SymbolExposure:  symbolExposure,
		},
	}
}

func sendPortfolioBroadcast() {
	resp := getPortfolioState()
	respJSON, _ := json.Marshal(resp)
	broadcaster.Broadcast(fmt.Sprintf(`{"type":"portfolio","state":%s}`, string(respJSON)))
}

func handleGetTimeline(w http.ResponseWriter, r *http.Request) {
	snapshotsMu.Lock()
	defer snapshotsMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshots)
}

func handleGetPortfolioAt(w http.ResponseWriter, r *http.Request) {
	ts := r.URL.Query().Get("timestamp")
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		// fallback to index if numeric
		var idx int
		if _, err := fmt.Sscanf(ts, "%d", &idx); err == nil {
			snapshotsMu.Lock()
			defer snapshotsMu.Unlock()
			if idx >= 0 && idx < len(snapshots) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(snapshots[idx])
				return
			}
		}
		http.Error(w, "Invalid timestamp format", http.StatusBadRequest)
		return
	}

	snapshotsMu.Lock()
	defer snapshotsMu.Unlock()
	
	if len(snapshots) == 0 {
		http.Error(w, "No snapshots recorded", http.StatusNotFound)
		return
	}
	nearest := snapshots[0]
	minDiff := absDuration(t.Sub(nearest.Timestamp))
	for _, snap := range snapshots {
		diff := absDuration(t.Sub(snap.Timestamp))
		if diff < minDiff {
			minDiff = diff
			nearest = snap
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nearest)
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

type PositionDiff struct {
	Symbol   string  `json:"symbol"`
	QtyDiff  float64 `json:"qty_diff"`
	CostDiff float64 `json:"cost_diff"`
}

type PortfolioDiff struct {
	CashDiff      float64        `json:"cash_diff"`
	PositionDiffs []PositionDiff `json:"position_diffs"`
}

func handleGetDiff(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	snapshotsMu.Lock()
	defer snapshotsMu.Unlock()

	var fromSnap, toSnap *PortfolioSnapshot
	for i := range snapshots {
		if snapshots[i].LineageID == fromStr {
			fromSnap = &snapshots[i]
		}
		if snapshots[i].LineageID == toStr {
			toSnap = &snapshots[i]
		}
	}

	if fromSnap == nil || toSnap == nil {
		http.Error(w, "Snapshot not found", http.StatusNotFound)
		return
	}

	cashDiff := toSnap.Portfolio.Cash - fromSnap.Portfolio.Cash
	posDiffMap := make(map[string]*PositionDiff)

	for _, pos := range fromSnap.Portfolio.Positions {
		posDiffMap[pos.Symbol] = &PositionDiff{
			Symbol:   pos.Symbol,
			QtyDiff:  -pos.Qty,
			CostDiff: -pos.CostBasis,
		}
	}

	for _, pos := range toSnap.Portfolio.Positions {
		diff, exists := posDiffMap[pos.Symbol]
		if exists {
			diff.QtyDiff += pos.Qty
			diff.CostDiff += pos.CostBasis
		} else {
			posDiffMap[pos.Symbol] = &PositionDiff{
				Symbol:   pos.Symbol,
				QtyDiff:  pos.Qty,
				CostDiff: pos.CostBasis,
			}
		}
	}

	var diffs []PositionDiff
	for _, diff := range posDiffMap {
		diffs = append(diffs, *diff)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PortfolioDiff{
		CashDiff:      cashDiff,
		PositionDiffs: diffs,
	})
}

func main() {
	fs := http.FileServer(http.Dir("./dashboard"))
	http.Handle("/", fs)

	http.HandleFunc("/api/trader/execute", handleExecute)
	http.HandleFunc("/api/portfolio/current", handleGetPortfolio)
	http.HandleFunc("/api/lineage/trades", handleGetTradeHistory)
	http.HandleFunc("/api/events", broadcaster.HandleSSE)
	
	http.HandleFunc("/api/portfolio/timeline", handleGetTimeline)
	http.HandleFunc("/api/portfolio/at", handleGetPortfolioAt)
	http.HandleFunc("/api/portfolio/diff", handleGetDiff)

	// Keep old endpoint for backwards compatibility
	http.HandleFunc("/api/substrate/portfolio", func(w http.ResponseWriter, r *http.Request) {
		portfolioMu.Lock()
		defer portfolioMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(portfolio)
	})

	// Add genesis snapshot
	addSnapshot("genesis", getPortfolioState())

	// Background Arbitrage Bot Simulator
	go func() {
		ticker := time.NewTicker(6 * time.Second)
		for range ticker.C {
			// Simulate disparity fluctuation
			uniPrice := 3000.0 + float64(time.Now().Unix()%10)
			camelotPrice := 3015.0 + float64((time.Now().Unix()*3)%20)
			disparity := math.Abs(camelotPrice - uniPrice)
			
			// Simulated EV in Wei (approx 0.001 ETH per dollar difference)
			evWei := disparity * 1e15
			gasPriceGwei := int64(28 + (time.Now().Unix()%15))
			gasCostWei := 180000.0 * float64(gasPriceGwei) * 1e9
			
			netEvWei := evWei - gasCostWei
			
			if netEvWei > 0 {
				lineageID := fmt.Sprintf("bot-arb-%x", time.Now().UnixNano())[0:16]
				
				// Broadcast verdict to UI
				verdictJSON, _ := json.Marshal(map[string]interface{}{
					"type":       "verdict",
					"approved":   true,
					"rationale":  fmt.Sprintf("[Arb Bot] UniV3 ($%.2f) ↔ Camelot ($%.2f). EV: %.4f ETH. Gas: %d Gwei", uniPrice, camelotPrice, netEvWei/1e18, gasPriceGwei),
					"lineage_id": lineageID,
				})
				broadcaster.Broadcast(string(verdictJSON))
				
				// Record bot event in lineage history
				recordHistory(TradeEntry{
					LineageID: lineageID,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Mutation: TradeMutation{
						Kind:      "arbitrage_swap",
						Symbol:    "ETH/USDC",
						Qty:       10.0,
						CostBasis: uniPrice,
					},
					Verdict: TradeVerdict{
						Approved:  true,
						Rationale: fmt.Sprintf("Disparity: $%.2f. Net EV: %.4f ETH", disparity, netEvWei/1e18),
					},
				})
			}
		}
	}()

	fmt.Println("Serving MEV HUD on http://localhost:2027 ...")
	if err := http.ListenAndServe(":2027", nil); err != nil {
		panic(err)
	}
}
