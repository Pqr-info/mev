package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type AtlasMetricsResponse struct {
	Inference InferenceMetrics `json:"inference"`
	Temporal  TemporalMetrics  `json:"temporal"`
	Health    HealthMetrics    `json:"health"`
	Consensus ConsensusMetrics `json:"consensus"`
	Atlas     AtlasMetrics     `json:"atlas"`
	Timestamp time.Time        `json:"timestamp"`
}

func (api *AtlasAPI) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	// The scaffold shows (c *Coordinator) HandleAtlasMetrics, but we already have an AtlasAPI struct!
	// I'll attach this to AtlasAPI so we can easily add it to the Serve mux.
	
	// Assuming api.atlas has a metrics field (which we will add). 
	// For now we'll just mock it if it's not wired up yet, but the user will provide hooks later!
	if api.metrics == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	api.metrics.mu.RLock()
	defer api.metrics.mu.RUnlock()

	resp := AtlasMetricsResponse{
		Inference: api.metrics.Inference,
		Temporal:  api.metrics.Temporal,
		Health:    api.metrics.Health,
		Consensus: api.metrics.Consensus,
		Atlas:     api.metrics.Atlas,
		Timestamp: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(resp)
}
