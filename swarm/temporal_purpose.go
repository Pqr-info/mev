package main

import (
	"fmt"
	"time"
)

type PurposeID string

type PurposeVector struct {
	Domain         Domain    `json:"domain"`
	Mission        string    `json:"mission"`
	LongRangeGoals []string  `json:"long_range_goals"`
	IdentityFit    float64   `json:"identity_fit"`
	EvolutionFit   float64   `json:"evolution_fit"`
	StabilityFit   float64   `json:"stability_fit"`
	Drift          float64   `json:"drift"`
	Timestamp      time.Time `json:"timestamp"`
}

type TrajectoryPlan struct {
	ID          PurposeID     `json:"id"`
	Domain      Domain        `json:"domain"`
	Adjustments []string      `json:"adjustments"`
	Window      time.Duration `json:"window"`
	Timestamp   time.Time     `json:"timestamp"`
	Context     map[string]any
}

type DirectionActionResult struct {
	Adjustment string    `json:"adjustment"`
	Domain     Domain    `json:"domain"`
	Success    bool      `json:"success"`
	Timestamp  time.Time `json:"timestamp"`
	Context    map[string]any
}

type TemporalPurposeEngine struct {
	plans   map[PurposeID]*TrajectoryPlan
	results []DirectionActionResult
}

func NewTemporalPurposeEngine() *TemporalPurposeEngine {
	return &TemporalPurposeEngine{
		plans:   make(map[PurposeID]*TrajectoryPlan),
		results: []DirectionActionResult{},
	}
}

func (e *TemporalPurposeEngine) ComputePurpose(domain Domain, sig map[string]any) PurposeVector {
	mission, _ := sig["mission"].(string)
	goals, _ := sig["long_range_goals"].([]string)
	idFit, _ := sig["identity_fit"].(float64)
	evoFit, _ := sig["evolution_fit"].(float64)
	stabFit, _ := sig["stability_fit"].(float64)
	dr, _ := sig["drift"].(float64)

	return PurposeVector{
		Domain:         domain,
		Mission:        mission,
		LongRangeGoals: goals,
		IdentityFit:    idFit,
		EvolutionFit:   evoFit,
		StabilityFit:   stabFit,
		Drift:          dr,
		Timestamp:      time.Now(),
	}
}

func (e *TemporalPurposeEngine) GenerateTrajectoryPlan(domain Domain, v PurposeVector) *TrajectoryPlan {
	adjustments := []string{}

	if v.IdentityFit < 0.6 {
		adjustments = append(adjustments, "ALIGN_IDENTITY_WITH_MISSION")
	}
	if v.EvolutionFit < 0.6 {
		adjustments = append(adjustments, "ADJUST_EVOLUTION_TRAJECTORY")
	}
	if v.StabilityFit < 0.5 {
		adjustments = append(adjustments, "REINFORCE_STABILITY_FOR_LONG_RANGE_GOALS")
	}
	if v.Drift > 0.3 {
		adjustments = append(adjustments, "CORRECT_PURPOSE_DRIFT")
	}

	plan := &TrajectoryPlan{
		ID:          PurposeID(fmt.Sprintf("pur-%s-%d", domain, time.Now().UnixNano())),
		Domain:      domain,
		Adjustments: adjustments,
		Window:      96 * time.Hour,
		Timestamp:   time.Now(),
		Context:     map[string]any{"purpose_vector": v},
	}

	e.plans[plan.ID] = plan
	return plan
}

func (e *TemporalPurposeEngine) ExecuteTrajectoryPlan(plan *TrajectoryPlan) []DirectionActionResult {
	var results []DirectionActionResult

	for _, adj := range plan.Adjustments {
		r := DirectionActionResult{
			Adjustment: adj,
			Domain:     plan.Domain,
			Success:    true,
			Timestamp:  time.Now(),
			Context:    plan.Context,
		}
		e.results = append(e.results, r)
		results = append(results, r)
	}

	return results
}

func (e *TemporalPurposeEngine) ReinforcePurpose(domain Domain, at time.Time) error {
	return nil
}
