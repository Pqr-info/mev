package main

import (
	"fmt"
	"time"
)

type AllianceID string

type AllianceVector struct {
	Members       []Organism `json:"members"`
	Trust         float64    `json:"trust"`
	SharedPurpose float64    `json:"shared_purpose"`
	EthicalFit    float64    `json:"ethical_fit"`
	Stability     float64    `json:"stability"`
	Drift         float64    `json:"drift"`
	Timestamp     time.Time  `json:"timestamp"`
}

type CoalitionPlan struct {
	ID        AllianceID     `json:"id"`
	Members   []Organism     `json:"members"`
	Actions   []string       `json:"actions"`
	Window    time.Duration  `json:"window"`
	Timestamp time.Time      `json:"timestamp"`
	Context   map[string]any `json:"context"`
}

type CoalitionActionResult struct {
	Action    string     `json:"action"`
	Members   []Organism `json:"members"`
	Success   bool       `json:"success"`
	Timestamp time.Time  `json:"timestamp"`
	Context   map[string]any
}

type TemporalAllianceEngine struct {
	plans   map[AllianceID]*CoalitionPlan
	results []CoalitionActionResult
}

func NewTemporalAllianceEngine() *TemporalAllianceEngine {
	return &TemporalAllianceEngine{
		plans:   make(map[AllianceID]*CoalitionPlan),
		results: []CoalitionActionResult{},
	}
}

func (e *TemporalAllianceEngine) ComputeAlliance(members []Organism, sig map[string]any) AllianceVector {
	trust, _ := sig["trust"].(float64)
	sp, _ := sig["shared_purpose"].(float64)
	ef, _ := sig["ethical_fit"].(float64)
	stab, _ := sig["stability"].(float64)
	dr, _ := sig["drift"].(float64)

	return AllianceVector{
		Members:       members,
		Trust:         trust,
		SharedPurpose: sp,
		EthicalFit:    ef,
		Stability:     stab,
		Drift:         dr,
		Timestamp:     time.Now(),
	}
}

func (e *TemporalAllianceEngine) GenerateCoalitionPlan(members []Organism, v AllianceVector) *CoalitionPlan {
	actions := []string{}

	if v.Trust < 0.5 {
		actions = append(actions, "INCREASE_ALLIANCE_TRANSPARENCY")
	}
	if v.SharedPurpose < 0.6 {
		actions = append(actions, "NEGOTIATE_SHARED_PURPOSE")
	}
	if v.EthicalFit < 0.6 {
		actions = append(actions, "ALIGN_COALITION_ETHICS")
	}
	if v.Stability < 0.5 {
		actions = append(actions, "STRENGTHEN_COALITION_GOVERNANCE")
	}
	if v.Drift > 0.3 {
		actions = append(actions, "CORRECT_ALLIANCE_DRIFT")
	}

	plan := &CoalitionPlan{
		ID:        AllianceID(fmt.Sprintf("all-%d", time.Now().UnixNano())),
		Members:   members,
		Actions:   actions,
		Window:    120 * time.Hour,
		Timestamp: time.Now(),
		Context:   map[string]any{"alliance_vector": v},
	}

	e.plans[plan.ID] = plan
	return plan
}

func (e *TemporalAllianceEngine) ExecuteCoalitionPlan(plan *CoalitionPlan) []CoalitionActionResult {
	var results []CoalitionActionResult

	for _, action := range plan.Actions {
		r := CoalitionActionResult{
			Action:    action,
			Members:   plan.Members,
			Success:   true,
			Timestamp: time.Now(),
			Context:   plan.Context,
		}
		e.results = append(e.results, r)
		results = append(results, r)
	}

	return results
}

func (e *TemporalAllianceEngine) ReinforceAlliance(members []Organism, at time.Time) error {
	return nil
}
