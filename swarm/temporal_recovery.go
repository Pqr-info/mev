package main

import (
	"fmt"
	"time"
)

type RecoveryID string

type RecoveryVector struct {
	Domain      Domain    `json:"domain"`
	Legitimacy  float64   `json:"legitimacy"`
	Stability   float64   `json:"stability"`
	Drift       float64   `json:"drift"`
	ShockImpact float64   `json:"shock_impact"`
	Sentiment   float64   `json:"sentiment"`
	Timestamp   time.Time `json:"timestamp"`
}

type RegenerationPlan struct {
	ID        RecoveryID    `json:"id"`
	Domain    Domain        `json:"domain"`
	Actions   []string      `json:"actions"`
	Window    time.Duration `json:"window"`
	Timestamp time.Time     `json:"timestamp"`
	Context   map[string]any
}

type RecoveryActionResult struct {
	Action    string    `json:"action"`
	Domain    Domain    `json:"domain"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	Context   map[string]any
}

type TemporalRecoveryEngine struct {
	plans   map[RecoveryID]*RegenerationPlan
	results []RecoveryActionResult
}

func NewTemporalRecoveryEngine() *TemporalRecoveryEngine {
	return &TemporalRecoveryEngine{
		plans:   make(map[RecoveryID]*RegenerationPlan),
		results: []RecoveryActionResult{},
	}
}

func (e *TemporalRecoveryEngine) ComputeVector(domain Domain, sig map[string]any) RecoveryVector {
	leg, _ := sig["legitimacy"].(float64)
	stab, _ := sig["stability"].(float64)
	dr, _ := sig["drift"].(float64)
	si, _ := sig["shock_impact"].(float64)
	sent, _ := sig["sentiment"].(float64)

	return RecoveryVector{
		Domain:      domain,
		Legitimacy:  leg,
		Stability:   stab,
		Drift:       dr,
		ShockImpact: si,
		Sentiment:   sent,
		Timestamp:   time.Now(),
	}
}

func (e *TemporalRecoveryEngine) GeneratePlan(domain Domain, v RecoveryVector) *RegenerationPlan {
	actions := []string{}

	if v.Legitimacy < 0.4 {
		actions = append(actions, "INCREASE_NARRATIVE_TRANSPARENCY")
	}
	if v.Stability < 0.5 {
		actions = append(actions, "INITIATE_GOVERNANCE_REVIEW")
	}
	if v.Drift > 0.3 {
		actions = append(actions, "TEMPORAL_COOLDOWN")
	}
	if v.ShockImpact > 0.2 {
		actions = append(actions, "SANCTION_REEVALUATION")
	}

	plan := &RegenerationPlan{
		ID:        RecoveryID(fmt.Sprintf("rec-%s-%d", domain, time.Now().UnixNano())),
		Domain:    domain,
		Actions:   actions,
		Window:    24 * time.Hour,
		Timestamp: time.Now(),
		Context:   map[string]any{"vector": v},
	}

	e.plans[plan.ID] = plan
	return plan
}

func (e *TemporalRecoveryEngine) ExecutePlan(plan *RegenerationPlan) []RecoveryActionResult {
	var results []RecoveryActionResult

	for _, action := range plan.Actions {
		r := RecoveryActionResult{
			Action:    action,
			Domain:    plan.Domain,
			Success:   true,
			Timestamp: time.Now(),
			Context:   plan.Context,
		}
		e.results = append(e.results, r)
		results = append(results, r)
	}

	return results
}

func (e *TemporalRecoveryEngine) CoolDown(domain Domain, at time.Time) error {
	// CoolDown dampens volatile vectors
	return nil
}
