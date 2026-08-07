package main

import (
	"fmt"
	"time"
)

type LegitimacyID string
type EntityID string
type EntityType string

const (
	EntityContract   EntityType = "CONTRACT"
	EntityClause     EntityType = "CLAUSE"
	EntityGovernance EntityType = "GOVERNANCE_BODY"
	EntitySanction   EntityType = "SANCTION"
	EntityRemedy     EntityType = "REMEDY"
)

type LegitimacyScore struct {
	ID         LegitimacyID `json:"id"`
	EntityID   EntityID     `json:"entity_id"`
	EntityType EntityType   `json:"entity_type"`
	Score      float64      `json:"score"`      // 0.0–1.0
	Confidence float64      `json:"confidence"` // 0.0–1.0
	Timestamp  time.Time    `json:"timestamp"`
	Context    map[string]any
}

type SignalType string

const (
	SignalNarrativePositive SignalType = "NARRATIVE_POSITIVE"
	SignalNarrativeNegative SignalType = "NARRATIVE_NEGATIVE"
	SignalAppealUpheld      SignalType = "APPEAL_UPHELD"
	SignalAppealOverturned  SignalType = "APPEAL_OVERTURNED"
	SignalSanctionAccepted  SignalType = "SANCTION_ACCEPTED"
	SignalSanctionContested SignalType = "SANCTION_CONTESTED"
)

type Signal struct {
	EntityID   EntityID   `json:"entity_id"`
	EntityType EntityType `json:"entity_type"`
	Type       SignalType `json:"type"`
	Weight     float64    `json:"weight"`
	Timestamp  time.Time  `json:"timestamp"`
	Context    map[string]any
}

type TemporalLegitimacyEngine struct {
	scores map[EntityID]*LegitimacyScore
}

func NewTemporalLegitimacyEngine() *TemporalLegitimacyEngine {
	return &TemporalLegitimacyEngine{
		scores: make(map[EntityID]*LegitimacyScore),
	}
}

func (e *TemporalLegitimacyEngine) IngestSignal(sig Signal) (*LegitimacyScore, error) {
	current, ok := e.scores[sig.EntityID]
	if !ok {
		current = &LegitimacyScore{
			ID:         LegitimacyID(fmt.Sprintf("leg-%s", sig.EntityID)),
			EntityID:   sig.EntityID,
			EntityType: sig.EntityType,
			Score:      0.8, // default starting score
			Confidence: 0.5,
			Timestamp:  time.Now(),
		}
	}

	// Adjust score based on signal weight
	current.Score += sig.Weight
	if current.Score > 1.0 {
		current.Score = 1.0
	} else if current.Score < 0.0 {
		current.Score = 0.0
	}

	current.Confidence += 0.05
	if current.Confidence > 1.0 {
		current.Confidence = 1.0
	}

	current.Timestamp = time.Now()
	e.scores[sig.EntityID] = current

	return current, nil
}

func (e *TemporalLegitimacyEngine) GetLegitimacy(entityID EntityID) (*LegitimacyScore, error) {
	current, ok := e.scores[entityID]
	if !ok {
		return nil, fmt.Errorf("no score found for entity %s", entityID)
	}
	return current, nil
}
