package main

import (
	"time"
)

type WillID string

type WillVector struct {
	Will                WillID         `json:"will"`
	Phases              []PhaseID      `json:"phases"`
	Scales              []ScaleID      `json:"scales"`
	WillStrength        float64        `json:"will_strength"`
	ChoiceClarity       float64        `json:"choice_clarity"`
	InternalConflict    float64        `json:"internal_conflict"`
	ExternalPressure    float64        `json:"external_pressure"`
	DecisionStability   float64        `json:"decision_stability"`
	PreferredTrajectory float64        `json:"preferred_trajectory"`
	Context             map[string]any `json:"context"`
	Timestamp           time.Time      `json:"timestamp"`
}

type DecisionPlan struct {
	ID        WillID         `json:"id"`
	Actions   []string       `json:"actions"`
	Window    time.Duration  `json:"window"`
	Timestamp time.Time      `json:"timestamp"`
	Context   map[string]any `json:"context"`
}

type DecisionResult struct {
	Action    string         `json:"action"`
	Will      WillID         `json:"will"`
	Success   bool           `json:"success"`
	Timestamp time.Time      `json:"timestamp"`
	Context   map[string]any `json:"context"`
}

type TemporalWillChoiceEngine struct {
	plans   map[WillID]*DecisionPlan
	results []DecisionResult
}

func NewTemporalWillChoiceEngine() *TemporalWillChoiceEngine {
	return &TemporalWillChoiceEngine{
		plans:   make(map[WillID]*DecisionPlan),
		results: []DecisionResult{},
	}
}

func (e *TemporalWillChoiceEngine) ComputeWill(w WillID, sig map[string]any) WillVector {
	phases, _ := sig["phases"].([]PhaseID)
	scales, _ := sig["scales"].([]ScaleID)
	ws, _ := sig["will_strength"].(float64)
	cc, _ := sig["choice_clarity"].(float64)
	ic, _ := sig["internal_conflict"].(float64)
	ep, _ := sig["external_pressure"].(float64)
	ds, _ := sig["decision_stability"].(float64)
	pt, _ := sig["preferred_trajectory"].(float64)
	ctx, _ := sig["context"].(map[string]any)

	return WillVector{
		Will:                w,
		Phases:              phases,
		Scales:              scales,
		WillStrength:        ws,
		ChoiceClarity:       cc,
		InternalConflict:    ic,
		ExternalPressure:    ep,
		DecisionStability:   ds,
		PreferredTrajectory: pt,
		Context:             ctx,
		Timestamp:           time.Now(),
	}
}

func (e *TemporalWillChoiceEngine) GenerateDecisionPlan(w WillID, v WillVector) *DecisionPlan {
	actions := []string{}

	if v.WillStrength < 0.6 {
		actions = append(actions, "AMPLIFY_WILL_STRENGTH")
	}
	if v.ChoiceClarity < 0.6 {
		actions = append(actions, "INCREASE_CHOICE_CLARITY")
	}
	if v.InternalConflict > 0.3 {
		actions = append(actions, "RESOLVE_INTERNAL_CONFLICT")
	}
	if v.ExternalPressure > 0.3 {
		actions = append(actions, "REDUCE_EXTERNAL_PRESSURE")
	}
	if v.DecisionStability < 0.5 {
		actions = append(actions, "STABILIZE_DECISION_PATH")
	}
	if v.PreferredTrajectory < 0.5 {
		actions = append(actions, "SELECT_PREFERRED_TRAJECTORY")
	}

	plan := &DecisionPlan{
		ID:        w,
		Actions:   actions,
		Window:    480 * time.Hour, // 20 days
		Timestamp: time.Now(),
		Context:   map[string]any{"will_vector": v},
	}

	e.plans[plan.ID] = plan
	return plan
}

func (e *TemporalWillChoiceEngine) ExecuteDecisionPlan(plan *DecisionPlan) []DecisionResult {
	var results []DecisionResult

	for _, action := range plan.Actions {
		r := DecisionResult{
			Action:    action,
			Will:      plan.ID,
			Success:   true,
			Timestamp: time.Now(),
			Context:   plan.Context,
		}
		e.results = append(e.results, r)
		results = append(results, r)
	}

	return results
}

func (e *TemporalWillChoiceEngine) ReinforceWill(w WillID, at time.Time) error {
	return nil
}
