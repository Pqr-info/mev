package safety

import (
	"sync"
)

var (
	safetyMutex        sync.RWMutex
	globalStressGauge  float64 = 0.0 // 0.0 (safe) to 1.0 (critical stress)
	globalStability    float64 = 1.0 // 1.0 (total stability) to 0.0 (fractured consensus)
)

// SetTemporalMetrics is a hook for the Consensus Engine to update the safety gate.
func SetTemporalMetrics(stress float64, stability float64) {
	safetyMutex.Lock()
	defer safetyMutex.Unlock()
	
	// Clamp values
	if stress < 0.0 { stress = 0.0 }
	if stress > 1.0 { stress = 1.0 }
	
	if stability < 0.0 { stability = 0.0 }
	if stability > 1.0 { stability = 1.0 }

	globalStressGauge = stress
	globalStability = stability
}

// GetTemporalMetrics returns the current organism temporal stress and consensus stability.
func GetTemporalMetrics() (float64, float64) {
	safetyMutex.RLock()
	defer safetyMutex.RUnlock()
	return globalStressGauge, globalStability
}

// IsTemporallySafe acts as the organism's survival instinct.
// If stress > 0.2 OR stability < 0.8, it returns false (unsafe).
func IsTemporallySafe() bool {
	safetyMutex.RLock()
	defer safetyMutex.RUnlock()
	
	if globalStressGauge > 0.2 {
		return false
	}
	if globalStability < 0.8 {
		return false
	}
	return true
}
