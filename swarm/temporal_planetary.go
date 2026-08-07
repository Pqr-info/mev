package main

import (
	"time"
)

type PlanetID string

type PlanetaryVector struct {
	Planet              PlanetID         `json:"planet"`
	Civilizations       []CivilizationID `json:"civilizations"`
	Federations         []FederationID   `json:"federations"`
	Members             []Organism       `json:"members"`
	EcologicalStability float64          `json:"ecological_stability"`
	ResourceFlow        float64          `json:"resource_flow"`
	CivilizationalLoad  float64          `json:"civilizational_load"`
	SupraDynamics       float64          `json:"supra_dynamics"`
	PlanetaryDrift      float64          `json:"planetary_drift"`
	CollapseRisk        float64          `json:"collapse_risk"`
	Timestamp           time.Time        `json:"timestamp"`
}

type PlanetaryEvolutionPlan struct {
	ID        PlanetID      `json:"id"`
	Planet    PlanetID      `json:"planet"`
	Actions   []string      `json:"actions"`
	Window    time.Duration `json:"window"`
	Timestamp time.Time     `json:"timestamp"`
	Context   map[string]any
}

type PlanetaryActionResult struct {
	Action    string    `json:"action"`
	Planet    PlanetID  `json:"planet"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	Context   map[string]any
}

type TemporalPlanetaryEngine struct {
	plans   map[PlanetID]*PlanetaryEvolutionPlan
	results []PlanetaryActionResult
}

func NewTemporalPlanetaryEngine() *TemporalPlanetaryEngine {
	return &TemporalPlanetaryEngine{
		plans:   make(map[PlanetID]*PlanetaryEvolutionPlan),
		results: []PlanetaryActionResult{},
	}
}

func (e *TemporalPlanetaryEngine) ComputePlanetary(planet PlanetID, sig map[string]any) PlanetaryVector {
	civs, _ := sig["civilizations"].([]CivilizationID)
	feds, _ := sig["federations"].([]FederationID)
	mem, _ := sig["members"].([]Organism)
	es, _ := sig["ecological_stability"].(float64)
	rf, _ := sig["resource_flow"].(float64)
	cl, _ := sig["civilizational_load"].(float64)
	sd, _ := sig["supra_dynamics"].(float64)
	pd, _ := sig["planetary_drift"].(float64)
	cr, _ := sig["collapse_risk"].(float64)

	return PlanetaryVector{
		Planet:              planet,
		Civilizations:       civs,
		Federations:         feds,
		Members:             mem,
		EcologicalStability: es,
		ResourceFlow:        rf,
		CivilizationalLoad:  cl,
		SupraDynamics:       sd,
		PlanetaryDrift:      pd,
		CollapseRisk:        cr,
		Timestamp:           time.Now(),
	}
}

func (e *TemporalPlanetaryEngine) GeneratePlanetaryPlan(planet PlanetID, v PlanetaryVector) *PlanetaryEvolutionPlan {
	actions := []string{}

	if v.EcologicalStability < 0.6 {
		actions = append(actions, "STABILIZE_ECOLOGICAL_SYSTEMS")
	}
	if v.ResourceFlow < 0.5 {
		actions = append(actions, "OPTIMIZE_RESOURCE_DISTRIBUTION")
	}
	if v.CivilizationalLoad > 0.7 {
		actions = append(actions, "REDUCE_PLANETARY_LOAD")
	}
	if v.SupraDynamics < 0.5 {
		actions = append(actions, "IMPROVE_SUPRA_CIVILIZATIONAL_DYNAMICS")
	}
	if v.PlanetaryDrift > 0.3 {
		actions = append(actions, "CORRECT_PLANETARY_DRIFT")
	}
	if v.CollapseRisk > 0.2 {
		actions = append(actions, "MITIGATE_PLANETARY_COLLAPSE_RISK")
	}

	plan := &PlanetaryEvolutionPlan{
		ID:        planet,
		Planet:    planet,
		Actions:   actions,
		Window:    360 * time.Hour, // 15 days
		Timestamp: time.Now(),
		Context:   map[string]any{"planetary_vector": v},
	}

	e.plans[plan.ID] = plan
	return plan
}

func (e *TemporalPlanetaryEngine) ExecutePlanetaryPlan(plan *PlanetaryEvolutionPlan) []PlanetaryActionResult {
	var results []PlanetaryActionResult

	for _, action := range plan.Actions {
		r := PlanetaryActionResult{
			Action:    action,
			Planet:    plan.Planet,
			Success:   true,
			Timestamp: time.Now(),
			Context:   plan.Context,
		}
		e.results = append(e.results, r)
		results = append(results, r)
	}

	return results
}

func (e *TemporalPlanetaryEngine) ReinforcePlanetary(planet PlanetID, at time.Time) error {
	return nil
}
