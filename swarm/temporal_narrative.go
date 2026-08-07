package main

import (
	"fmt"
	"time"
)

type NarrativeID string
type NarrativeType string

const (
	NarrativeContract    NarrativeType = "CONTRACT"
	NarrativeArbitration NarrativeType = "ARBITRATION"
	NarrativeEnforcement NarrativeType = "ENFORCEMENT"
	NarrativeSanction    NarrativeType = "SANCTION"
	NarrativeRemedy      NarrativeType = "REMEDY"
	NarrativeAppeal      NarrativeType = "APPEAL"
)

type NarrativeEntry struct {
	ID         NarrativeID   `json:"id"`
	Type       NarrativeType `json:"type"`
	Actor      ActorID       `json:"actor"`
	Domain     Domain        `json:"domain"`
	Timestamp  time.Time     `json:"timestamp"`
	Title      string        `json:"title"`
	Summary    string        `json:"summary"`
	Details    string        `json:"details"`
	Context    map[string]any
	RelatedIDs []string `json:"related_ids"`
}

type TemporalNarrativeEngine struct {
	entries map[NarrativeID]*NarrativeEntry
}

func NewTemporalNarrativeEngine() *TemporalNarrativeEngine {
	return &TemporalNarrativeEngine{
		entries: make(map[NarrativeID]*NarrativeEntry),
	}
}

func (e *TemporalNarrativeEngine) RecordNarrative(entry *NarrativeEntry) {
	e.entries[entry.ID] = entry
}

func (e *TemporalNarrativeEngine) GenerateFromEvent(eventType string, actor string, domain string, details string) *NarrativeEntry {
	entry := &NarrativeEntry{
		ID:        NarrativeID(fmt.Sprintf("narr-%d", time.Now().UnixNano())),
		Type:      NarrativeType(eventType),
		Actor:     ActorID(actor),
		Domain:    Domain(domain),
		Timestamp: time.Now(),
		Title:     fmt.Sprintf("Narrative explanation for %s event", eventType),
		Summary:   details,
		Details:   details,
	}

	e.RecordNarrative(entry)
	return entry
}

func (e *TemporalNarrativeEngine) GenerateThread(actor ActorID, domain Domain, since time.Time) []*NarrativeEntry {
	var thread []*NarrativeEntry
	for _, entry := range e.entries {
		if entry.Actor == actor && entry.Domain == domain && entry.Timestamp.After(since) {
			thread = append(thread, entry)
		}
	}
	return thread
}
