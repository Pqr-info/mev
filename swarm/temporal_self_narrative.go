package main

import (
	"fmt"
	"time"
)

type PhaseID string
type ScaleID string

type SelfNarrativeVector struct {
	Coherence      float64        `json:"coherence"`
	Integrity      float64        `json:"integrity"`
	ValueAlignment float64        `json:"value_alignment"`
	Regret         float64        `json:"regret"`
	Aspiration     float64        `json:"aspiration"`
	Themes         []string       `json:"themes"`
	Archetypes     []string       `json:"archetypes"`
	StoryFragments []string       `json:"story_fragments"`
	Context        map[string]any `json:"context"`
	Timestamp      time.Time      `json:"timestamp"`
}

type SelfNarrativePlan struct {
	ID        string         `json:"id"`
	Actions   []string       `json:"actions"`
	Window    time.Duration  `json:"window"`
	Timestamp time.Time      `json:"timestamp"`
	Context   map[string]any `json:"context"`
}

type SelfNarrativeActionResult struct {
	Action    string         `json:"action"`
	Success   bool           `json:"success"`
	Timestamp time.Time      `json:"timestamp"`
	Context   map[string]any `json:"context"`
}

type TemporalSelfNarrativeEngine struct {
	plans   map[string]*SelfNarrativePlan
	results []SelfNarrativeActionResult
}

func NewTemporalSelfNarrativeEngine() *TemporalSelfNarrativeEngine {
	return &TemporalSelfNarrativeEngine{
		plans:   make(map[string]*SelfNarrativePlan),
		results: []SelfNarrativeActionResult{},
	}
}

func (e *TemporalSelfNarrativeEngine) ComputeSelfNarrative(sig map[string]any) SelfNarrativeVector {
	coherence, _ := sig["coherence"].(float64)
	integrity, _ := sig["integrity"].(float64)
	va, _ := sig["value_alignment"].(float64)
	regret, _ := sig["regret"].(float64)
	asp, _ := sig["aspiration"].(float64)
	themes, _ := sig["themes"].([]string)
	arch, _ := sig["archetypes"].([]string)
	frag, _ := sig["story_fragments"].([]string)
	ctx, _ := sig["context"].(map[string]any)

	return SelfNarrativeVector{
		Coherence:      coherence,
		Integrity:      integrity,
		ValueAlignment: va,
		Regret:         regret,
		Aspiration:     asp,
		Themes:         themes,
		Archetypes:     arch,
		StoryFragments: frag,
		Context:        ctx,
		Timestamp:      time.Now(),
	}
}

func (e *TemporalSelfNarrativeEngine) GenerateSelfNarrativePlan(v SelfNarrativeVector) *SelfNarrativePlan {
	actions := []string{}

	if v.Coherence < 0.6 {
		actions = append(actions, "REWRITE_STORY_ARC_FOR_COHERENCE")
	}
	if v.Integrity < 0.6 {
		actions = append(actions, "ALIGN_ACTIONS_WITH_STATED_VALUES")
	}
	if v.ValueAlignment < 0.6 {
		actions = append(actions, "REINFORCE_CORE_VALUES")
	}
	if v.Regret > 0.3 {
		actions = append(actions, "PROCESS_AND_INTEGRATE_REGRET")
	}
	if v.Aspiration < 0.5 {
		actions = append(actions, "AMPLIFY_PREFERRED_FUTURE_TRAJECTORIES")
	}

	plan := &SelfNarrativePlan{
		ID:        fmt.Sprintf("SELF_NARRATIVE_PLAN_%d", time.Now().UnixNano()),
		Actions:   actions,
		Window:    240 * time.Hour, // ~10 days
		Timestamp: time.Now(),
		Context:   map[string]any{"self_narrative_vector": v},
	}

	e.plans[plan.ID] = plan
	return plan
}

func (e *TemporalSelfNarrativeEngine) ExecuteSelfNarrativePlan(plan *SelfNarrativePlan) []SelfNarrativeActionResult {
	var results []SelfNarrativeActionResult

	for _, action := range plan.Actions {
		r := SelfNarrativeActionResult{
			Action:    action,
			Success:   true,
			Timestamp: time.Now(),
			Context:   plan.Context,
		}
		e.results = append(e.results, r)
		results = append(results, r)
	}

	return results
}

func (e *TemporalSelfNarrativeEngine) ReinforceMeaning(at time.Time) error {
	return nil
}
