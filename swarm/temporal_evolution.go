package main

import (
	"fmt"
	"time"
)

type EvolutionID string

type EvolutionVector struct {
	Domain        Domain    `json:"domain"`
	Legitimacy    float64   `json:"legitimacy"`
	Stability     float64   `json:"stability"`
	Drift         float64   `json:"drift"`
	CrisisFreq    float64   `json:"crisis_freq"`
	RecoveryEff   float64   `json:"recovery_eff"`
	NarrativeSent float64   `json:"narrative_sent"`
	Timestamp     time.Time `json:"timestamp"`
}

type AdaptationPlan struct {
	ID        EvolutionID   `json:"id"`
	Domain    Domain        `json:"domain"`
	Mutations []string      `json:"mutations"`
	Window    time.Duration `json:"window"`
	Timestamp time.Time     `json:"timestamp"`
	Context   map[string]any
}

type AdaptationResult struct {
	Mutation  string    `json:"mutation"`
	Domain    Domain    `json:"domain"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	Context   map[string]any
}

type TemporalEvolutionEngine struct {
	plans   map[EvolutionID]*AdaptationPlan
	results []AdaptationResult
}

func NewTemporalEvolutionEngine() *TemporalEvolutionEngine {
	return &TemporalEvolutionEngine{
		plans:   make(map[EvolutionID]*AdaptationPlan),
		results: []AdaptationResult{},
	}
}

func (e *TemporalEvolutionEngine) ComputeVector(domain Domain, sig map[string]any) EvolutionVector {
	leg, _ := sig["legitimacy"].(float64)
	stab, _ := sig["stability"].(float64)
	dr, _ := sig["drift"].(float64)
	cf, _ := sig["crisis_freq"].(float64)
	re, _ := sig["recovery_eff"].(float64)
	sent, _ := sig["narrative_sent"].(float64)

	return EvolutionVector{
		Domain:        domain,
		Legitimacy:    leg,
		Stability:     stab,
		Drift:         dr,
		CrisisFreq:    cf,
		RecoveryEff:   re,
		NarrativeSent: sent,
		Timestamp:     time.Now(),
	}
}

func (e *TemporalEvolutionEngine) GeneratePlan(domain Domain, v EvolutionVector) *AdaptationPlan {
	mutations := []string{}

	if v.Legitimacy < 0.5 {
		mutations = append(mutations, "ADJUST_GOVERNANCE_WEIGHTS")
	}
	if v.Stability < 0.6 {
		mutations = append(mutations, "TUNE_ARBITRATION_PRIORITIES")
	}
	if v.Drift > 0.3 {
		mutations = append(mutations, "REVISE_CONTRACT_CLAUSES")
	}
	if v.CrisisFreq > 0.2 {
		mutations = append(mutations, "EVOLVE_SANCTION_POLICIES")
	}
	if v.NarrativeSent < 0.4 {
		mutations = append(mutations, "OPTIMIZE_NARRATIVE_STRATEGY")
	}

	plan := &AdaptationPlan{
		ID:        EvolutionID(fmt.Sprintf("evo-%s-%d", domain, time.Now().UnixNano())),
		Domain:    domain,
		Mutations: mutations,
		Window:    48 * time.Hour,
		Timestamp: time.Now(),
		Context:   map[string]any{"vector": v},
	}

	e.plans[plan.ID] = plan
	return plan
}

func (e *TemporalEvolutionEngine) ExecutePlan(plan *AdaptationPlan) []AdaptationResult {
	var results []AdaptationResult

	for _, mut := range plan.Mutations {
		r := AdaptationResult{
			Mutation:  mut,
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

func (e *TemporalEvolutionEngine) ApplyAdaptiveWindow(domain Domain, at time.Time) error {
	return nil
}
