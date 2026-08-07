package main

import "time"

type AppealID string
type AppealTargetType string

const (
	TargetContract    AppealTargetType = "CONTRACT"
	TargetClause      AppealTargetType = "CLAUSE"
	TargetArbitration AppealTargetType = "ARBITRATION"
	TargetSanction    AppealTargetType = "SANCTION"
	TargetRemedy      AppealTargetType = "REMEDY"
)

type AppealState string

const (
	StateSubmitted   AppealState = "SUBMITTED"
	StateUnderReview AppealState = "UNDER_REVIEW"
	StateAccepted    AppealState = "ACCEPTED"
	StateRejected    AppealState = "REJECTED"
	StateImplemented AppealState = "IMPLEMENTED"
)

type Appeal struct {
	ID          AppealID         `json:"id"`
	Actor       ActorID          `json:"actor"`
	Domain      Domain           `json:"domain"`
	TargetType  AppealTargetType `json:"target_type"`
	TargetID    string           `json:"target_id"`
	Reason      string           `json:"reason"`
	State       AppealState      `json:"state"`
	SubmittedAt time.Time        `json:"submitted_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Context     map[string]any   `json:"context"`
}

type DecisionOutcome string

const (
	OutcomeUphold   DecisionOutcome = "UPHOLD"
	OutcomeOverturn DecisionOutcome = "OVERTURN"
	OutcomeModify   DecisionOutcome = "MODIFY"
)

type AppealDecision struct {
	AppealID  AppealID        `json:"appeal_id"`
	Outcome   DecisionOutcome `json:"outcome"`
	Reason    string          `json:"reason"`
	Timestamp time.Time       `json:"timestamp"`
	Context   map[string]any  `json:"context"`
}

type TemporalAppealsEngine struct {
	appeals   map[AppealID]*Appeal
	decisions map[AppealID]*AppealDecision
}

func NewTemporalAppealsEngine() *TemporalAppealsEngine {
	return &TemporalAppealsEngine{
		appeals:   make(map[AppealID]*Appeal),
		decisions: make(map[AppealID]*AppealDecision),
	}
}

func (e *TemporalAppealsEngine) SubmitAppeal(a *Appeal) {
	a.State = StateSubmitted
	now := time.Now()
	a.SubmittedAt = now
	a.UpdatedAt = now
	e.appeals[a.ID] = a
}

func (e *TemporalAppealsEngine) OnDecision(d *AppealDecision) {
	e.decisions[d.AppealID] = d
	a, ok := e.appeals[d.AppealID]
	if !ok {
		return
	}

	if d.Outcome == OutcomeOverturn {
		a.State = StateAccepted
	} else {
		a.State = StateRejected
	}
	a.UpdatedAt = time.Now()
}
