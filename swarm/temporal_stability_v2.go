package main

import (
	"fmt"
	"time"
)

type StabilityID string

type StabilityScore struct {
	ID         StabilityID `json:"id"`
	Domain     Domain      `json:"domain"`
	Score      float64     `json:"score"`      // 0.0–1.0
	Resilience float64     `json:"resilience"` // 0.0–1.0
	Drift      float64     `json:"drift"`      // -1.0 to +1.0
	Timestamp  time.Time   `json:"timestamp"`
	Context    map[string]any
}

type Shock struct {
	Domain    Domain    `json:"domain"`
	Magnitude float64   `json:"magnitude"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	Context   map[string]any
}

type TemporalStabilityEngineV2 struct {
	scores map[Domain]*StabilityScore
	shocks []Shock
}

func NewTemporalStabilityEngineV2() *TemporalStabilityEngineV2 {
	return &TemporalStabilityEngineV2{
		scores: make(map[Domain]*StabilityScore),
		shocks: []Shock{},
	}
}

func (e *TemporalStabilityEngineV2) IngestSignal(sig map[string]any) (*StabilityScore, error) {
	domainStr, ok := sig["domain"].(string)
	if !ok {
		return nil, fmt.Errorf("domain missing from signal")
	}
	domain := Domain(domainStr)

	latest, ok := e.scores[domain]
	baseScore := 0.7
	resilience := 0.5
	drift := 0.0

	if ok && latest != nil {
		baseScore = latest.Score
		resilience = latest.Resilience
		drift = latest.Drift
	}

	if val, ok := sig["legitimacy_delta"].(float64); ok {
		baseScore += val
	}

	if val, ok := sig["appeal_overturned"].(bool); ok && val {
		baseScore -= 0.1
		drift -= 0.02
	}

	if baseScore < 0.0 {
		baseScore = 0.0
	} else if baseScore > 1.0 {
		baseScore = 1.0
	}

	resilience += 0.01
	if resilience > 1.0 {
		resilience = 1.0
	}

	if drift < -1.0 {
		drift = -1.0
	} else if drift > 1.0 {
		drift = 1.0
	}

	score := &StabilityScore{
		ID:         StabilityID(fmt.Sprintf("stab-%s-%d", domain, time.Now().UnixNano())),
		Domain:     domain,
		Score:      baseScore,
		Resilience: resilience,
		Drift:      drift,
		Timestamp:  time.Now(),
		Context:    sig,
	}

	e.scores[domain] = score
	return score, nil
}

func (e *TemporalStabilityEngineV2) DetectShock(sig map[string]any) (*Shock, error) {
	val, ok := sig["shock_magnitude"].(float64)
	if ok && val > 0.2 {
		domainStr, _ := sig["domain"].(string)
		sourceStr, _ := sig["source"].(string)
		shock := Shock{
			Domain:    Domain(domainStr),
			Magnitude: val,
			Source:    sourceStr,
			Timestamp: time.Now(),
			Context:   sig,
		}
		e.shocks = append(e.shocks, shock)
		return &shock, nil
	}
	return nil, nil
}

func (e *TemporalStabilityEngineV2) ComputeResilience(domain Domain, at time.Time) float64 {
	recentCount := 0
	since := at.Add(-24 * time.Hour)
	for _, s := range e.shocks {
		if s.Domain == domain && s.Timestamp.After(since) {
			recentCount++
		}
	}

	res := 1.0 - float64(recentCount)*0.1
	if res < 0.0 {
		res = 0.0
	}
	return res
}
