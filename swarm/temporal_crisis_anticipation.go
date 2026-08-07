package main

import (
	"fmt"
	"time"
)

type CrisisID string

type CrisisVector struct {
	Domain     Domain    `json:"domain"`
	Legitimacy float64   `json:"legitimacy"`
	Stability  float64   `json:"stability"`
	Drift      float64   `json:"drift"`
	ShockRate  float64   `json:"shock_rate"`
	Sentiment  float64   `json:"sentiment"`
	Timestamp  time.Time `json:"timestamp"`
}

type CrisisForecast struct {
	ID          CrisisID      `json:"id"`
	Domain      Domain        `json:"domain"`
	Probability float64       `json:"probability"` // 0.0–1.0
	Window      time.Duration `json:"window"`
	Timestamp   time.Time     `json:"timestamp"`
	Context     map[string]any
}

type CascadeWarning struct {
	SourceDomain Domain    `json:"source_domain"`
	TargetDomain Domain    `json:"target_domain"`
	Severity     float64   `json:"severity"`
	Timestamp    time.Time `json:"timestamp"`
}

type TemporalCrisisAnticipationEngine struct {
	forecasts map[CrisisID]*CrisisForecast
	warnings  []CascadeWarning
}

func NewTemporalCrisisAnticipationEngine() *TemporalCrisisAnticipationEngine {
	return &TemporalCrisisAnticipationEngine{
		forecasts: make(map[CrisisID]*CrisisForecast),
		warnings:  []CascadeWarning{},
	}
}

func (e *TemporalCrisisAnticipationEngine) ComputeVector(domain Domain, sig map[string]any) CrisisVector {
	leg, _ := sig["legitimacy"].(float64)
	stab, _ := sig["stability"].(float64)
	dr, _ := sig["drift"].(float64)
	sr, _ := sig["shock_rate"].(float64)
	sent, _ := sig["sentiment"].(float64)

	return CrisisVector{
		Domain:     domain,
		Legitimacy: leg,
		Stability:  stab,
		Drift:      dr,
		ShockRate:  sr,
		Sentiment:  sent,
		Timestamp:  time.Now(),
	}
}

func (e *TemporalCrisisAnticipationEngine) Forecast(domain Domain, v CrisisVector) *CrisisForecast {
	prob := (1.0 - v.Stability) + v.Drift + v.ShockRate
	if prob < 0.0 {
		prob = 0.0
	} else if prob > 1.0 {
		prob = 1.0
	}

	f := &CrisisForecast{
		ID:          CrisisID(fmt.Sprintf("fc-%s-%d", domain, time.Now().UnixNano())),
		Domain:      domain,
		Probability: prob,
		Window:      24 * time.Hour,
		Timestamp:   time.Now(),
		Context:     map[string]any{"vector": v},
	}

	e.forecasts[f.ID] = f
	return f
}

func (e *TemporalCrisisAnticipationEngine) DetectCascade(v CrisisVector) []CascadeWarning {
	var warnings []CascadeWarning

	if v.Drift > 0.3 && v.ShockRate > 0.2 {
		w := CascadeWarning{
			SourceDomain: v.Domain,
			TargetDomain: Domain("GLOBAL"),
			Severity:     v.Drift + v.ShockRate,
			Timestamp:    time.Now(),
		}
		e.warnings = append(e.warnings, w)
		warnings = append(warnings, w)
	}

	return warnings
}

func (e *TemporalCrisisAnticipationEngine) Advise(domain Domain, f *CrisisForecast) map[string]any {
	advice := map[string]any{
		"domain":      domain,
		"probability": f.Probability,
	}

	if f.Probability > 0.7 {
		advice["action"] = "INITIATE_GOVERNANCE_REVIEW"
	} else if f.Probability > 0.4 {
		advice["action"] = "INCREASE_NARRATIVE_TRANSPARENCY"
	} else {
		advice["action"] = "MONITOR"
	}

	return advice
}
