package hooks

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type HealerHookManager struct {
	mu           sync.Mutex
	errorCounter map[string]int
	lastRestarts map[string]time.Time
}

func NewHealerHookManager() *HealerHookManager {
	return &HealerHookManager{
		errorCounter: make(map[string]int),
		lastRestarts: make(map[string]time.Time),
	}
}

// TrackAnomaly checks metrics for latency or error rate anomalies
func (h *HealerHookManager) TrackAnomaly(service string, latencyMs float64, errorRate float64) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Mode B threshold checks
	if latencyMs > 50.0 {
		log.Warn().Msgf("[Healer Hook] Anomaly detected: Latency for %s is %.2f ms (Threshold: 50ms)", service, latencyMs)
		h.errorCounter[service]++
	}

	if errorRate > 0.05 {
		log.Warn().Msgf("[Healer Hook] Anomaly detected: Error rate for %s is %.2f%% (Threshold: 5%%)", service, errorRate*100)
		h.errorCounter[service]++
	}

	// Trigger recovery if threshold is breached multiple times
	if h.errorCounter[service] >= 3 {
		h.TriggerRecovery(service)
		h.errorCounter[service] = 0
		return true
	}

	return false
}

// TriggerRecovery launches recovery tasks
func (h *HealerHookManager) TriggerRecovery(service string) {
	log.Info().Msgf("[Healer Mode B] Initiating automated self-healing pipeline for service: %s", service)

	h.lastRestarts[service] = time.Now()
	
	// Execute mock recovery procedures (flushing state, re-subscribing to firehose websocket)
	fmt.Printf("[Healer Mode B] Flushing mempool cache and resetting websocket connections for %s\n", service)
}

func (h *HealerHookManager) GetRecoveryMetrics() map[string]time.Time {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	metrics := make(map[string]time.Time)
	for k, v := range h.lastRestarts {
		metrics[k] = v
	}
	return metrics
}
