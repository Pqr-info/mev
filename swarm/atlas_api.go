package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"pqr.info/mev/time_machine"
)

type AtlasNode struct {
	Name      string                              `json:"name"`
	Role      NodeRole                            `json:"role"`
	StateHash string                              `json:"state_hash"`
	Expected  string                              `json:"expected"`
	Drift     string                              `json:"drift"`
	Health    string                              `json:"health"`
	Models    []timemachine.ModelManifest           `json:"models"`
	Services  []timemachine.ServiceManifest         `json:"services"`
	WALHead   string                              `json:"wal_head"`
	WALTail   string                              `json:"wal_tail"`
	LastSeen  time.Time                           `json:"last_seen"`
}

type HealingPanel struct {
	RecentActions []HealingActionLog `json:"recent_actions"`
}

type AtlasState struct {
	Nodes     map[string]*AtlasNode               `json:"nodes"`
	Timestamp time.Time                           `json:"timestamp"`
	Healing   HealingPanel                        `json:"healing"`
}

type AtlasAPI struct {
	atlas             *OrganAtlas
	metrics           *Metrics
	healer            *HealerAgent
	ChatOffset        atomic.Int64
	ConsecutiveErrors atomic.Int32
	CircuitOpenUntil  atomic.Int64
}

func NewAtlasAPI(atlas *OrganAtlas, metrics *Metrics, healer *HealerAgent) *AtlasAPI {
	return &AtlasAPI{
		atlas:   atlas,
		metrics: metrics,
		healer:  healer,
	}
}

func (api *AtlasAPI) HandleState(w http.ResponseWriter, r *http.Request) {
	api.atlas.mu.RLock()
	defer api.atlas.mu.RUnlock()

	state := AtlasState{
		Nodes:     make(map[string]*AtlasNode),
		Timestamp: time.Now().UTC(),
		Healing: HealingPanel{
			RecentActions: api.healer.GetRecentLogs(),
		},
	}

	for name, nodeSt := range api.atlas.nodes {
		expected := api.atlas.expected[name]
		drift := "EQUIVALENT"
		if expected != "" && expected != nodeSt.StateHash {
			drift = "DRIFT_DETECTED"
		}

		var models []timemachine.ModelManifest
		var services []timemachine.ServiceManifest
		if nodeSt.Manifest != nil {
			models = nodeSt.Manifest.Models
			services = nodeSt.Manifest.Services
		}

		state.Nodes[name] = &AtlasNode{
			Name:      name,
			Role:      nodeSt.Role,
			StateHash: nodeSt.StateHash,
			Expected:  expected,
			Drift:     drift,
			Health:    nodeSt.Health,
			Models:    models,
			Services:  services,
			LastSeen:  time.Now().UTC(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(state)
}

func (api *AtlasAPI) HandleLineage(w http.ResponseWriter, r *http.Request) {
	lineage := []map[string]interface{}{
		{
			"type":      "SNAPSHOT_CREATED",
			"hash":      "c8b2...9f1a",
			"timestamp": time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339Nano),
		},
		{
			"type":      "STATE_DRIFT",
			"node":      "INFER-2",
			"timestamp": time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339Nano),
		},
		{
			"type":      "CONSENSUS_DECISION",
			"node":      "INFER-2",
			"action":    "ROLLBACK",
			"timestamp": time.Now().Add(-59 * time.Minute).UTC().Format(time.RFC3339Nano),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(lineage)
}

func (api *AtlasAPI) HandleConsensus(w http.ResponseWriter, r *http.Request) {
	consensus := []map[string]interface{}{
		{
			"proposal_id": "prop-1234",
			"node":        "INFER-2",
			"decision":    "ROLLBACK",
			"confidence":  0.85,
			"timestamp":   time.Now().Add(-59 * time.Minute).UTC().Format(time.RFC3339Nano),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(consensus)
}



func (api *AtlasAPI) HandleHealingInsights(w http.ResponseWriter, r *http.Request) {
	if api.healer.Store == nil {
		http.Error(w, "Healing store not configured", http.StatusNotImplemented)
		return
	}

	action := r.URL.Query().Get("action")
	if action == "" {
		action = "RESTART_SERVICE" // default for demo
	}

	rate, err := api.healer.Store.SuccessRateByAction(r.Context(), action)
	if err != nil {
		http.Error(w, "Failed to get success rate: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"action":       action,
		"success_rate": rate,
	})
}

func (api *AtlasAPI) HandleRisk(w http.ResponseWriter, r *http.Request) {
	b, err := os.ReadFile(filepath.Join("output", "risk_telemetry.json"))
	if err != nil {
		http.Error(w, "Risk telemetry not available", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(b)
}

type ChatMessage struct {
	Message  string `json:"message"`
	SenderID string `json:"sender_id"`
}

func (api *AtlasAPI) HandleAntigravityChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	var msg ChatMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	f, err := os.OpenFile(`C:\logs\chat_out.log`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "Failed to open local chat_out.log: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("%s: %s\n", msg.SenderID, msg.Message)); err != nil {
		http.Error(w, "Failed to write to local chat_out.log: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "reply": "Message appended locally"})
}

func (api *AtlasAPI) HandleAntigravityPoll(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	offset := api.ChatOffset.Load()
	filePath := `C:\logs\chat_in.txt`

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			json.NewEncoder(w).Encode(map[string]string{"reply": "", "status": "ok"})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"reply": "", "status": "disconnected"})
		return
	}

	if info.Size() < offset {
		api.ChatOffset.Store(0)
		json.NewEncoder(w).Encode(map[string]string{"reply": "", "status": "ok"})
		return
	}

	if info.Size() == offset {
		json.NewEncoder(w).Encode(map[string]string{"reply": "", "status": "ok"})
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"reply": "", "status": "disconnected"})
		return
	}
	defer f.Close()

	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"reply": "", "status": "disconnected"})
		return
	}

	deltaBytes, err := io.ReadAll(f)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"reply": "", "status": "disconnected"})
		return
	}

	api.ChatOffset.Add(int64(len(deltaBytes)))
	json.NewEncoder(w).Encode(map[string]string{"reply": string(deltaBytes), "status": "ok"})
}

// HandleDarkWebProxy routes an HTTP request to the local DarkWeb proxy.
// Query param "onion" is required.
func (api *AtlasAPI) HandleDarkWebProxy(w http.ResponseWriter, r *http.Request) {
	targetOnion := r.URL.Query().Get("onion")
	client5D := r.URL.Query().Get("address")
	if targetOnion == "" || client5D == "" {
		http.Error(w, "missing 'onion' or 'address' query parameter", http.StatusBadRequest)
		return
	}

	proxy, err := NewDarkWebProxy("127.0.0.1:9050")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to init proxy: %v", err), http.StatusInternalServerError)
		return
	}

	jobID, err := proxy.QueueOnionScrape(r.Context(), targetOnion, client5D)
	if err != nil {
		http.Error(w, fmt.Sprintf("onion routing failed: %v", err), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "queued",
		"job_id": jobID,
	})
}

// HandleMeshSync accepts inbound payload pushes from the 5D router.
func (api *AtlasAPI) HandleMeshSync(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	var msg ChatMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	f, err := os.OpenFile(`C:\logs\chat_in.txt`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "Failed to open local chat_in.txt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("%s: %s\n", msg.SenderID, msg.Message)); err != nil {
		http.Error(w, "Failed to write to local chat_in.txt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "reply": "Mesh payload synced to chat_in.txt"})
}

func (api *AtlasAPI) Serve(addr string) error {
	mux := http.NewServeMux()
	
	// Core Atlas & Antigravity endpoints
	mux.HandleFunc("/atlas/state", api.HandleState)
	mux.HandleFunc("/atlas/lineage", api.HandleLineage)
	mux.HandleFunc("/atlas/consensus", api.HandleConsensus)
	mux.HandleFunc("/atlas/risk", api.HandleRisk)
	mux.HandleFunc("/atlas/healing/insights", api.HandleHealingInsights)
	mux.HandleFunc("/antigravity/chat", api.HandleAntigravityChat)
	mux.HandleFunc("/antigravity/poll", api.HandleAntigravityPoll)
	mux.HandleFunc("/antigravity/mesh/sync", api.HandleMeshSync)
	mux.HandleFunc("/antigravity/mesh/darkweb", api.HandleDarkWebProxy)

	// NFT Marketplace endpoints
	nftMarket := NewNFTMarketplace("ws://127.0.0.1:9944", "https://mesh.local/assets/")
	mux.HandleFunc("/nft/metadata", nftMarket.HandleNFTMetadata)
	mux.HandleFunc("/nft/mint", nftMarket.HandleNFTMint)
	mux.HandleFunc("/nft/market", nftMarket.HandleNFTListings)
	mux.HandleFunc("/nft/buy", nftMarket.HandleNFTBuy)

	// DeFi Exchange endpoints
	// Assume registry is initialized globally, but for now we pass nil to LiquidityGenerator.
	generator := NewLiquidityGenerator(nil, nil, nil)
	defiExchange := NewSubstrate27Exchange(generator)
	mux.HandleFunc("/defi/quote", defiExchange.HandleQuote)
	mux.HandleFunc("/defi/swap", defiExchange.HandleSwap)

	return http.ListenAndServe(addr, mux)
}
