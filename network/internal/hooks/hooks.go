package hooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type RecoverFunc func()

type HealerHookManager struct {
	mu           sync.Mutex
	errorCounter map[string]int
	lastRestarts map[string]time.Time
	recoverFns   []RecoverFunc
}

func NewHealerHookManager() *HealerHookManager {
	return &HealerHookManager{
		errorCounter: make(map[string]int),
		lastRestarts: make(map[string]time.Time),
		recoverFns:   []RecoverFunc{},
	}
}

func (h *HealerHookManager) RegisterRecover(fn RecoverFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.recoverFns = append(h.recoverFns, fn)
}

func (h *HealerHookManager) TriggerRecover(service string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Warn().Msgf("[Healer] Triggering recovery pipeline for service: %s", service)
	h.lastRestarts[service] = time.Now()

	for _, fn := range h.recoverFns {
		go fn()
	}

	if timeslipEndpoint := os.Getenv("SOS_TIMESLIP_ENDPOINT"); timeslipEndpoint != "" {
		go func() {
			payload := map[string]interface{}{
				"service": service,
				"event":   "recovery_triggered",
				"ts":      time.Now().UTC().Format(time.RFC3339),
			}
			body, _ := json.Marshal(payload)
			client := &http.Client{Timeout: 5 * time.Second}
			_, _ = client.Post(timeslipEndpoint, "application/json", bytes.NewReader(body))
		}()
	}
}

// TrackAnomaly checks metrics for latency or error rate anomalies
func (h *HealerHookManager) TrackAnomaly(service string, latencyMs float64, errorRate float64) bool {
	h.mu.Lock()
	// Check latency threshold (e.g. 50ms)
	if latencyMs > 50.0 {
		log.Warn().Msgf("[Healer Hook] Anomaly detected: Latency for %s is %.2f ms (Threshold: 50ms)", service, latencyMs)
		h.errorCounter[service]++
	}

	// Check error rate threshold (e.g. 5%)
	if errorRate > 0.05 {
		log.Warn().Msgf("[Healer Hook] Anomaly detected: Error rate for %s is %.2f%% (Threshold: 5%%)", service, errorRate*100)
		h.errorCounter[service]++
	}

	triggered := false
	if h.errorCounter[service] >= 3 {
		h.mu.Unlock() // unlock before trigger to prevent deadlock
		h.TriggerRecover(service)
		h.mu.Lock()
		h.errorCounter[service] = 0
		triggered = true
	}
	h.mu.Unlock()

	return triggered
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
