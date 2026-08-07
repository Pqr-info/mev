package main

import "time"

type ViolationID string
type SanctionID string
type RemedyID string

type Violation struct {
	ID         ViolationID    `json:"id"`
	Actor      ActorID        `json:"actor"`
	Domain     Domain         `json:"domain"`
	ContractID string         `json:"contract_id"`
	ClauseID   string         `json:"clause_id"`
	Timestamp  time.Time      `json:"timestamp"`
	Context    map[string]any `json:"context"`
	Severity   int            `json:"severity"` // derived
	Count      int            `json:"count"`    // repeated violations
}

type SanctionType string

const (
	SanctionSuspend    SanctionType = "SUSPEND"
	SanctionRateLimit  SanctionType = "RATE_LIMIT"
	SanctionQuarantine SanctionType = "QUARANTINE"
	SanctionDemote     SanctionType = "DEMOTE_ROLE"
)

type Sanction struct {
	ID        SanctionID     `json:"id"`
	Actor     ActorID        `json:"actor"`
	Domain    Domain         `json:"domain"`
	Type      SanctionType   `json:"type"`
	Reason    string         `json:"reason"`
	AppliedAt time.Time      `json:"applied_at"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
	Context   map[string]any `json:"context"`
}

type Remedy struct {
	ID        RemedyID       `json:"id"`
	Actor     ActorID        `json:"actor"`
	Domain    Domain         `json:"domain"`
	Reason    string         `json:"reason"`
	AppliedAt time.Time      `json:"applied_at"`
	Context   map[string]any `json:"context"`
}

type TemporalSanctionsEngine struct {
	violations []Violation
	sanctions  map[SanctionID]*Sanction
	remedies   map[RemedyID]*Remedy
}

func NewTemporalSanctionsEngine() *TemporalSanctionsEngine {
	return &TemporalSanctionsEngine{
		violations: []Violation{},
		sanctions:  make(map[SanctionID]*Sanction),
		remedies:   make(map[RemedyID]*Remedy),
	}
}

func (e *TemporalSanctionsEngine) OnEnforcementPacket(packet EnforcementPacket) {
	if packet.Allowed {
		return
	}

	v := Violation{
		ID:         ViolationID("viol-" + packet.ContractID + "-" + packet.ClauseID),
		Actor:      ActorID(packet.Actor),
		Domain:     Domain(packet.Domain),
		ContractID: packet.ContractID,
		ClauseID:   packet.ClauseID,
		Timestamp:  time.Now(),
		Context:    packet.Context,
		Severity:   packet.Priority,
		Count:      1,
	}

	e.violations = append(e.violations, v)

	// Issue suspension if severity is high
	if v.Severity > 5 {
		expires := time.Now().Add(10 * time.Minute)
		s := &Sanction{
			ID:        SanctionID("sanc-" + string(v.ID)),
			Actor:     v.Actor,
			Domain:    v.Domain,
			Type:      SanctionSuspend,
			Reason:    "high severity violation of " + v.ContractID,
			AppliedAt: time.Now(),
			ExpiresAt: &expires,
		}
		e.sanctions[s.ID] = s
	}
}

func (e *TemporalSanctionsEngine) ResolveRemedy(r *Remedy) {
	e.remedies[r.ID] = r
}
