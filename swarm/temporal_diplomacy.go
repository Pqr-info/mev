package main

import (
	"fmt"
	"time"
)

type DiplomacyID string
type Organism string

type DiplomaticVector struct {
	SelfOrganism     Organism  `json:"self"`
	ExternalOrganism Organism  `json:"external"`
	Trust            float64   `json:"trust"`
	Alignment        float64   `json:"alignment"`
	Tension          float64   `json:"tension"`
	EthicalFit       float64   `json:"ethical_fit"`
	PurposeFit       float64   `json:"purpose_fit"`
	Drift            float64   `json:"drift"`
	Timestamp        time.Time `json:"timestamp"`
}

type DiplomaticAlignmentPlan struct {
	ID        DiplomacyID   `json:"id"`
	Self      Organism      `json:"self"`
	External  Organism      `json:"external"`
	Actions   []string      `json:"actions"`
	Window    time.Duration `json:"window"`
	Timestamp time.Time     `json:"timestamp"`
	Context   map[string]any
}

type DiplomaticActionResult struct {
	Action    string    `json:"action"`
	Self      Organism  `json:"self"`
	External  Organism  `json:"external"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	Context   map[string]any
}

type TemporalDiplomacyEngine struct {
	plans   map[DiplomacyID]*DiplomaticAlignmentPlan
	results []DiplomaticActionResult
}

func NewTemporalDiplomacyEngine() *TemporalDiplomacyEngine {
	return &TemporalDiplomacyEngine{
		plans:   make(map[DiplomacyID]*DiplomaticAlignmentPlan),
		results: []DiplomaticActionResult{},
	}
}

func (e *TemporalDiplomacyEngine) ComputeDiplomacy(self Organism, external Organism, sig map[string]any) DiplomaticVector {
	trust, _ := sig["trust"].(float64)
	align, _ := sig["alignment"].(float64)
	tension, _ := sig["tension"].(float64)
	ethFit, _ := sig["ethical_fit"].(float64)
	purFit, _ := sig["purpose_fit"].(float64)
	dr, _ := sig["drift"].(float64)

	return DiplomaticVector{
		SelfOrganism:     self,
		ExternalOrganism: external,
		Trust:            trust,
		Alignment:        align,
		Tension:          tension,
		EthicalFit:       ethFit,
		PurposeFit:       purFit,
		Drift:            dr,
		Timestamp:        time.Now(),
	}
}

func (e *TemporalDiplomacyEngine) GenerateAlignmentPlan(self Organism, external Organism, v DiplomaticVector) *DiplomaticAlignmentPlan {
	actions := []string{}

	if v.Trust < 0.5 {
		actions = append(actions, "INITIATE_DIPLOMATIC_TRANSPARENCY")
	}
	if v.Alignment < 0.5 {
		actions = append(actions, "NEGOTIATE_PURPOSE_ALIGNMENT")
	}
	if v.Tension > 0.3 {
		actions = append(actions, "OPEN_CONFLICT_MEDIATION_CHANNEL")
	}
	if v.Drift > 0.3 {
		actions = append(actions, "CORRECT_DIPLOMATIC_DRIFT")
	}

	plan := &DiplomaticAlignmentPlan{
		ID:        DiplomacyID(fmt.Sprintf("dip-%s-%s-%d", self, external, time.Now().UnixNano())),
		Self:      self,
		External:  external,
		Actions:   actions,
		Window:    96 * time.Hour,
		Timestamp: time.Now(),
		Context:   map[string]any{"diplomatic_vector": v},
	}

	e.plans[plan.ID] = plan
	return plan
}

func (e *TemporalDiplomacyEngine) ExecuteAlignmentPlan(plan *DiplomaticAlignmentPlan) []DiplomaticActionResult {
	var results []DiplomaticActionResult

	for _, action := range plan.Actions {
		r := DiplomaticActionResult{
			Action:    action,
			Self:      plan.Self,
			External:  plan.External,
			Success:   true,
			Timestamp: time.Now(),
			Context:   plan.Context,
		}
		e.results = append(e.results, r)
		results = append(results, r)
	}

	return results
}

func (e *TemporalDiplomacyEngine) ReinforceDiplomacy(self Organism, external Organism, at time.Time) error {
	return nil
}
