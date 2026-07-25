package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config represents the organ settings
type Config struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Version string `json:"version"`
	Mode    string `json:"mode"`
}

// HealthStatus represents telemetry/health info
type HealthStatus struct {
	Status    string    `json:"status"`
	Uptime    string    `json:"uptime"`
	BuildTime string    `json:"build_time"`
	Anomaly   bool      `json:"anomaly"`
	Metrics   Metrics   `json:"metrics"`
	Timestamp time.Time `json:"timestamp"`
}

// Metrics captures latency and processing rates
type Metrics struct {
	LatencyMs          float64 `json:"latency_ms"`
	Tps                float64 `json:"tps"`
	GasPriceGwei       float64 `json:"gas_price_gwei"`
	SimulatedArbsCount int64   `json:"simulated_arbs_count"`
}

type RTTicket struct {
	ID           string                 `json:"id"`
	Queue        string                 `json:"queue"`
	Subject      string                 `json:"subject"`
	Text         string                 `json:"text"`
	Status       string                 `json:"status"`
	Created      time.Time              `json:"created"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	MemoryType   string                 `json:"memory_type,omitempty"`
	PhysicalTier string                 `json:"physical_tier,omitempty"`
}

func classifyMemory(createdAt time.Time, now time.Time) (string, string) {
	ageDays := now.Sub(createdAt).Hours() / 24.0

	if ageDays < 1 {
		return "ACTIVE", "RAM"
	}
	if ageDays < 7 {
		return "RECENT", "PG"
	}
	if ageDays < 30 {
		return "ARCHIVAL", "SSD"
	}
	if ageDays < 365 {
		return "HERITAGE", "SSD_LT"
	}
	if ageDays < 730 {
		return "FOSSIL", "LCL"
	}
	if ageDays < 1825 {
		return "FOSSILX", "LCL_X"
	}
	return "NAS", "NAS"
}

type Adapter struct {
	mu           sync.RWMutex
	startTime    time.Time
	config       Config
	status       HealthStatus
	tickets      map[string]RTTicket
	lastID       int
	manifest     map[string]interface{}
	manifestPath string
	db           *sql.DB
}

func NewAdapter(cfg Config) *Adapter {
	manifestPath := os.Getenv("SOS_MANIFEST_PATH")
	if manifestPath == "" {
		manifestPath = filepath.Join("..", "manifest.json")
	}
	adapter := &Adapter{
		startTime:    time.Now(),
		config:       cfg,
		tickets:      make(map[string]RTTicket),
		lastID:       1000,
		manifestPath: manifestPath,
		status: HealthStatus{
			Status:    "HEALTHY",
			BuildTime: time.Now().Format(time.RFC3339),
			Anomaly:   false,
			Metrics: Metrics{
				LatencyMs:          1.2,
				Tps:                150.0,
				GasPriceGwei:       25.5,
				SimulatedArbsCount: 0,
			},
		},
	}
	adapter.loadManifest()
	adapter.refreshManifest()

	// Initialize CockroachDB connection
	dbURL := os.Getenv("COCKROACH_DB_URL")
	if dbURL == "" {
		dbURL = "postgresql://root@46.224.219.174:5196/antigravity?sslmode=disable"
	}
	db, err := sql.Open("pgx", dbURL)
	if err == nil {
		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(2)
		adapter.db = db
		log.Info().Msg("Connected to CockroachDB successfully")
		
		// Run schema migrations for memory_type and physical_tier
		_, _ = db.Exec("ALTER TABLE tickets ADD COLUMN IF NOT EXISTS memory_type TEXT")
		_, _ = db.Exec("ALTER TABLE tickets ADD COLUMN IF NOT EXISTS physical_tier TEXT")
	} else {
		log.Warn().Msgf("Failed to open CockroachDB connection pool: %v", err)
	}

	return adapter
}

func (a *Adapter) StartServer(port string) error {
	http.HandleFunc("/health", a.handleHealth)
	http.HandleFunc("/config", a.handleConfig)
	http.HandleFunc("/manifest.json", a.handleManifest)
	http.HandleFunc("/healer/trigger-recovery", a.handleRecovery)
	http.HandleFunc("/timeslip-generator", a.handleTimeslipGenerator)
	http.HandleFunc("/rt/REST/2.0/tickets", a.handleRTTickets)
	http.HandleFunc("/rt/REST/2.0/tickets/", a.handleRTTicket)
	http.HandleFunc("/agents/gemini", a.handleGemini)
	http.HandleFunc("/help", a.handleHelp)
	http.HandleFunc("/ping", a.handlePing)
	http.HandleFunc("/consensus/proposal", a.handleConsensusProposal)
	http.HandleFunc("/consensus/vote", a.handleConsensusVote)
	http.HandleFunc("/consensus/proposals", a.handleConsensusProposals)
	http.HandleFunc("/governance/lineage", a.handleGovernanceLineage)
	http.HandleFunc("/consensus/rollback", a.handleConsensusRollback)
	http.HandleFunc("/governance/simulate", a.handleGovernanceSimulate)
	http.HandleFunc("/governance/metrics", a.handleGovernanceMetrics)
	http.HandleFunc("/governance/forecast", a.handleGovernanceForecast)

	log.Info().Msgf("Mesh Citizen Shell adapter starting on port %s", port)
	return http.ListenAndServe(":"+port, nil)
}

func (a *Adapter) loadManifest() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.manifest = map[string]interface{}{
		"service":     "sos",
		"lastUpdated": time.Now().UTC().Format(time.RFC3339),
		"endpoints":   map[string]interface{}{},
		"hosts":       map[string]interface{}{},
		"auth":        map[string]interface{}{},
		"hints":       []string{"Use /rt/REST/2.0/tickets for RT-style SOS operations", "Ouroboros should own restarts and health reconciliation"},
	}

	if data, err := os.ReadFile(a.manifestPath); err == nil {
		if err := json.Unmarshal(data, &a.manifest); err != nil {
			log.Warn().Err(err).Str("path", a.manifestPath).Msg("Failed to parse manifest; using defaults")
		}
	}
}

func (a *Adapter) refreshManifest() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.manifest == nil {
		a.manifest = map[string]interface{}{}
	}
	if a.manifest["service"] == nil {
		a.manifest["service"] = "sos"
	}
	a.manifest["lastUpdated"] = time.Now().UTC().Format(time.RFC3339)
	a.manifest["endpoints"] = map[string]interface{}{
		"health":             "http://localhost:8100/health",
		"metrics":            "http://localhost:9091/metrics",
		"prometheus":         "http://localhost:9090",
		"grafana":            "http://localhost:3000",
		"ticketApi":          "http://localhost:8100/rt/REST/2.0/tickets",
		"timeslipGen":        "http://localhost:8100/timeslip-generator",
		"recovery":           "http://localhost:8100/healer/trigger-recovery",
		"agentsGemini":       "http://localhost:8100/agents/gemini",
		"help":               "http://localhost:8100/help",
		"consensusProposal":  "http://localhost:8100/consensus/proposal",
		"consensusVote":      "http://localhost:8100/consensus/vote",
		"consensusProposals": "http://localhost:8100/consensus/proposals",
		"governanceLineage":  "http://localhost:8100/governance/lineage",
		"consensusRollback":  "http://localhost:8100/consensus/rollback",
		"governanceSimulate": "http://localhost:8100/governance/simulate",
		"governanceMetrics":  "http://localhost:8100/governance/metrics",
		"governanceForecast": "http://localhost:8100/governance/forecast",
	}
	a.manifest["hosts"] = map[string]interface{}{
		"primary": os.Getenv("SOS_HOSTNAME"),
		"alias":   "sos-flagship",
	}
	a.manifest["auth"] = map[string]interface{}{
		"grafana": map[string]interface{}{
			"username": os.Getenv("GRAFANA_USER"),
			"password": os.Getenv("GRAFANA_PASSWORD"),
		},
	}
	if _, ok := a.manifest["hints"]; !ok {
		a.manifest["hints"] = []string{"Use /rt/REST/2.0/tickets for RT-style SOS operations", "Ouroboros should own restarts and health reconciliation"}
	}
	data, err := json.MarshalIndent(a.manifest, "", "  ")
	if err == nil {
		_ = os.WriteFile(a.manifestPath, data, 0o644)
	}
}

func (a *Adapter) handleManifest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":"error","message":"invalid json"}`))
			return
		}

		a.mu.Lock()
		a.manifest = payload
		if a.manifest == nil {
			a.manifest = map[string]interface{}{}
		}
		a.manifest["lastUpdated"] = time.Now().UTC().Format(time.RFC3339)
		data, err := json.MarshalIndent(a.manifest, "", "  ")
		a.mu.Unlock()
		if err == nil {
			_ = os.WriteFile(a.manifestPath, data, 0o644)
		}
	}

	a.mu.RLock()
	defer a.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a.manifest)
}

func (a *Adapter) handleHealth(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	a.status.Uptime = time.Since(a.startTime).String()
	a.status.Timestamp = time.Now()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a.status)
}

func (a *Adapter) handleConfig(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a.config)
}

func (a *Adapter) handleRecovery(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	log.Warn().Msg("HealerAgentV2 triggered recovery pipeline! Re-initializing memory states...")

	nodeURL := os.Getenv("MEV_NODE_URL")
	if nodeURL == "" {
		nodeURL = "http://localhost:9091"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(nodeURL+"/control/recover", "application/json", nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to forward recovery trigger to mev-node")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"status":"error","message":"failed to connect to mev-node"}`))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Msgf("mev-node returned non-200 status: %d", resp.StatusCode)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"status":"error","message":"mev-node recovery failed"}`))
		return
	}

	a.status.Anomaly = false
	a.status.Status = "RECOVERED"

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"recovered"}`))
}

func (a *Adapter) handleTimeslipGenerator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":"error","message":"invalid json"}`))
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	subject, _ := payload["subject"].(string)
	if subject == "" {
		subject = fmt.Sprintf("SOS event: %v", payload["event"])
	}
	text, _ := payload["text"].(string)
	if text == "" {
		text = fmt.Sprintf("Event %v recorded via timeslip generator", payload["event"])
	}
	queue, _ := payload["queue"].(string)
	if queue == "" {
		queue = "SOS"
	}

	ticket := RTTicket{
		ID:       strconv.Itoa(a.lastID),
		Queue:    queue,
		Subject:  subject,
		Text:     text,
		Status:   "new",
		Created:  time.Now().UTC(),
		Metadata: payload,
	}
	a.lastID++
	a.tickets[ticket.ID] = ticket

	if endpoint := os.Getenv("SOS_TIMESLIP_ENDPOINT"); endpoint != "" {
		body, _ := json.Marshal(ticket)
		client := &http.Client{Timeout: 5 * time.Second}
		_, _ = client.Post(endpoint, "application/json", bytes.NewReader(body))
	}

	// Real-time CockroachDB insertion
	if a.db != nil {
		u := make([]byte, 16)
		_, _ = rand.Read(u)
		u[8] = (u[8] & 0x3f) | 0x80
		u[6] = (u[6] & 0x0f) | 0x40
		uuidStr := fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])

		memType, physTier := classifyMemory(time.Now().UTC(), time.Now().UTC())

		_, err := a.db.Exec(`
			INSERT INTO tickets (ticket_id, layer_id, creator_agent_id, status, iteration, escalation_level, created_at, updated_at, memory_type, physical_tier)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, uuidStr, 1, "github_poller", "new", 1, 0, time.Now().UTC(), time.Now().UTC(), memType, physTier)
		if err != nil {
			log.Warn().Msgf("Failed to insert timeslip into tickets table: %v", err)
		} else {
			metaBytes, _ := json.Marshal(payload)
			_, err = a.db.Exec(`
				INSERT INTO ticket_content (ticket_id, intent_blob, consensus_score, raw_content, created_at)
				VALUES ($1, $2, $3, $4, $5)
			`, uuidStr, metaBytes, 1.0, []byte(text), time.Now().UTC())
			if err != nil {
				log.Warn().Msgf("Failed to insert timeslip content: %v", err)
			} else {
				log.Info().Msgf("Successfully saved timeslip %s into CockroachDB in real-time.", uuidStr)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "accepted", "id": ticket.ID, "ticket": ticket})
}

func (a *Adapter) handleRTTickets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var payload struct {
			Queue   string `json:"Queue"`
			Subject string `json:"Subject"`
			Text    string `json:"Text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":"error","message":"invalid json"}`))
			return
		}
		if payload.Subject == "" {
			payload.Subject = "Untitled SOS ticket"
		}
		if payload.Queue == "" {
			payload.Queue = "SOS"
		}

		a.mu.Lock()
		defer a.mu.Unlock()
		a.lastID++
		ticket := RTTicket{
			ID:      strconv.Itoa(a.lastID),
			Queue:   payload.Queue,
			Subject: payload.Subject,
			Text:    payload.Text,
			Status:  "new",
			Created: time.Now().UTC(),
		}
		a.tickets[ticket.ID] = ticket
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "created", "id": ticket.ID, "ticket": ticket})
	case http.MethodGet:
		a.mu.RLock()
		defer a.mu.RUnlock()
		list := make([]RTTicket, 0, len(a.tickets))
		for _, ticket := range a.tickets {
			list = append(list, ticket)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"tickets": list})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *Adapter) handleRTTicket(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/rt/REST/2.0/tickets/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	id := parts[0]
	a.mu.RLock()
	defer a.mu.RUnlock()

	ticket, ok := a.tickets[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ticket": ticket})
}

type GeminiTaskEnvelope struct {
	TaskID   string                 `json:"task_id"`
	Kind     string                 `json:"kind"`
	Mode     string                 `json:"mode"`
	Prompt   string                 `json:"prompt"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (a *Adapter) handleGemini(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var task GeminiTaskEnvelope
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_json","message":"failed to parse body"}`))
		return
	}

	bcpdURL := os.Getenv("BCPD_URL")
	if bcpdURL == "" {
		bcpdURL = "http://localhost:8100"
	}

	// Create intent payload
	intentPayload := map[string]interface{}{
		"agent": "gemini",
		"kind":  task.Kind,
		"payload": map[string]interface{}{
			"prompt":   task.Prompt,
			"task_id":  task.TaskID,
			"metadata": task.Metadata,
		},
	}

	bodyBytes, err := json.Marshal(intentPayload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"marshal_error"}`))
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(bcpdURL+"/intent", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"bcpd_unreachable","message":"%s"}`, err.Error())))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"bcpd_error","status":%d}`, resp.StatusCode)))
		return
	}

	var intentResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&intentResp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if task.Mode == "async" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(intentResp)
		return
	}

	// Mode is "sync" (or default). Poll for completion.
	pollStart := time.Now()
	pollTimeout := 35 * time.Second
	pollInterval := 500 * time.Millisecond

	for time.Since(pollStart) < pollTimeout {
		time.Sleep(pollInterval)

		stateResp, err := client.Get(bcpdURL + "/state")
		if err != nil {
			continue
		}
		defer stateResp.Body.Close()

		var stateData struct {
			Intents []struct {
				ID     string                 `json:"id"`
				Status string                 `json:"status"`
				Error  map[string]interface{} `json:"error"`
			} `json:"intents"`
			Agents map[string]struct {
				LastResponse map[string]interface{} `json:"last_response"`
			} `json:"agents"`
		}

		if err := json.NewDecoder(stateResp.Body).Decode(&stateData); err != nil {
			continue
		}

		found := false
		var status string
		var intentErr map[string]interface{}
		for _, intent := range stateData.Intents {
			if intent.ID == intentResp.ID {
				found = true
				status = intent.Status
				intentErr = intent.Error
				break
			}
		}

		if !found {
			continue
		}

		if status == "done" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Return last response from agents.gemini.last_response
			geminiAgentState, exists := stateData.Agents["gemini"]
			if exists && geminiAgentState.LastResponse != nil {
				_ = json.NewEncoder(w).Encode(geminiAgentState.LastResponse)
			} else {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "done", "id": intentResp.ID})
			}
			return
		}

		if status == "error" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "intent_failed",
				"details": intentErr,
			})
			return
		}
	}

	w.WriteHeader(http.StatusGatewayTimeout)
	_, _ = w.Write([]byte(`{"error":"sync_timeout","message":"exceeded polling limit"}`))
}

func (a *Adapter) handleHelp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Prompt   string                 `json:"prompt"`
		Metadata map[string]interface{} `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_json","message":"failed to parse body"}`))
		return
	}

	bcpdURL := os.Getenv("BCPD_URL")
	if bcpdURL == "" {
		bcpdURL = "http://localhost:8100"
	}

	intentPayload := map[string]interface{}{
		"agent": "gemini", // bcpd handles default routing
		"kind":  "help",
		"payload": map[string]interface{}{
			"prompt":   payload.Prompt,
			"metadata": payload.Metadata,
		},
	}

	bodyBytes, err := json.Marshal(intentPayload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(bcpdURL+"/intent", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"bcpd_unreachable","message":"%s"}`, err.Error())))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	var intentResp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&intentResp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(intentResp)
}

func (a *Adapter) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok","service":"mesh-adapter"}`))
}

func (a *Adapter) handleConsensusProposal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bcpdURL := os.Getenv("BCPD_URL")
	if bcpdURL == "" {
		bcpdURL = "http://localhost:8100"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(bcpdURL+"/proposal", "application/json", r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"bcpd_unreachable","message":"%s"}`, err.Error())))
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	var body bytes.Buffer
	_, _ = body.ReadFrom(resp.Body)
	_, _ = w.Write(body.Bytes())
}

func (a *Adapter) handleConsensusVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bcpdURL := os.Getenv("BCPD_URL")
	if bcpdURL == "" {
		bcpdURL = "http://localhost:8100"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(bcpdURL+"/proposal/vote", "application/json", r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"bcpd_unreachable","message":"%s"}`, err.Error())))
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	var body bytes.Buffer
	_, _ = body.ReadFrom(resp.Body)
	_, _ = w.Write(body.Bytes())
}

func (a *Adapter) handleConsensusProposals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bcpdURL := os.Getenv("BCPD_URL")
	if bcpdURL == "" {
		bcpdURL = "http://localhost:8100"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(bcpdURL + "/state")
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"bcpd_unreachable","message":"%s"}`, err.Error())))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	var stateData struct {
		Intents []struct {
			ID      string                 `json:"id"`
			Agent   string                 `json:"agent"`
			Kind    string                 `json:"kind"`
			Status  string                 `json:"status"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"intents"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&stateData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	proposals := make([]interface{}, 0)
	for _, intent := range stateData.Intents {
		if intent.Agent == "mesh-consensus" && intent.Kind == "proposal" {
			proposals = append(proposals, map[string]interface{}{
				"intent_id": intent.ID,
				"status":    intent.Status,
				"payload":   intent.Payload,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"proposals": proposals,
	})
}

func (a *Adapter) handleGovernanceLineage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bcpdURL := os.Getenv("BCPD_URL")
	if bcpdURL == "" {
		bcpdURL = "http://localhost:8100"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(bcpdURL + "/lineage")
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"bcpd_unreachable","message":"%s"}`, err.Error())))
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	var body bytes.Buffer
	_, _ = body.ReadFrom(resp.Body)
	_, _ = w.Write(body.Bytes())
}

func (a *Adapter) handleConsensusRollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bcpdURL := os.Getenv("BCPD_URL")
	if bcpdURL == "" {
		bcpdURL = "http://localhost:8100"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(bcpdURL+"/proposal/rollback", "application/json", r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"bcpd_unreachable","message":"%s"}`, err.Error())))
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	var body bytes.Buffer
	_, _ = body.ReadFrom(resp.Body)
	_, _ = w.Write(body.Bytes())
}

func (a *Adapter) handleGovernanceSimulate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bcpdURL := os.Getenv("BCPD_URL")
	if bcpdURL == "" {
		bcpdURL = "http://localhost:8100"
	}

	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Post(bcpdURL+"/simulation", "application/json", r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"bcpd_unreachable","message":"%s"}`, err.Error())))
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	var body bytes.Buffer
	_, _ = body.ReadFrom(resp.Body)
	_, _ = w.Write(body.Bytes())
}

func (a *Adapter) handleGovernanceMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bcpdURL := os.Getenv("BCPD_URL")
	if bcpdURL == "" {
		bcpdURL = "http://localhost:8100"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(bcpdURL + "/metrics")
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"bcpd_unreachable","message":"%s"}`, err.Error())))
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	var body bytes.Buffer
	_, _ = body.ReadFrom(resp.Body)
	_, _ = w.Write(body.Bytes())
}

func (a *Adapter) handleGovernanceForecast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bcpdURL := os.Getenv("BCPD_URL")
	if bcpdURL == "" {
		bcpdURL = "http://localhost:8100"
	}

	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Post(bcpdURL+"/forecast", "application/json", r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"bcpd_unreachable","message":"%s"}`, err.Error())))
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	var body bytes.Buffer
	_, _ = body.ReadFrom(resp.Body)
	_, _ = w.Write(body.Bytes())
}

func main() {
	port := flag.String("port", "2026", "Port to run the MCSH adapter on")
	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg := Config{
		Name:    "MEV Engine",
		ID:      "mev",
		Version: "1.0.0",
		Mode:    "simulation",
	}

	adapter := NewAdapter(cfg)
	if err := adapter.StartServer(*port); err != nil {
		log.Fatal().Err(err).Msg("Failed to start MCSH adapter")
	}
}
