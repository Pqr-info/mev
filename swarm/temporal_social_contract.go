package main

import "time"

type ContractID string
type ClauseID string
type ActorID string
type Domain string

type ContractState string

const (
	StateDraft     ContractState = "DRAFT"
	StateProposed  ContractState = "PROPOSED"
	StateRatified  ContractState = "RATIFIED"
	StateActive    ContractState = "ACTIVE"
	StateSuspended ContractState = "SUSPENDED"
	StateExpired   ContractState = "EXPIRED"
)

type Recurrence struct {
	Pattern string        `json:"pattern"` // e.g. "DAILY", "WEEKLY", cron-like
	Offset  time.Duration `json:"offset"`
}

type TemporalScope struct {
	Start       time.Time      `json:"start"`
	End         *time.Time     `json:"end,omitempty"`
	Recurrence  *Recurrence    `json:"recurrence,omitempty"`
	GracePeriod *time.Duration `json:"grace_period,omitempty"`
	TimeZone    string         `json:"time_zone"`
}

type ActorSet struct {
	Actors []ActorID `json:"actors"`
	Roles  []string  `json:"roles"`
	Groups []string  `json:"groups"`
}

type Clause struct {
	ID           ClauseID      `json:"id"`
	ContractID   ContractID    `json:"contract_id"`
	Domain       Domain        `json:"domain"`
	Expression   string        `json:"expression"` // DSL or CEL-like
	Temporal     TemporalScope `json:"temporal"`
	Priority     int           `json:"priority"`
	Dependencies []ClauseID    `json:"dependencies"`
}

type Contract struct {
	ID        ContractID    `json:"id"`
	Title     string        `json:"title"`
	Summary   string        `json:"summary"`
	Actors    ActorSet      `json:"actors"`
	Clauses   []Clause      `json:"clauses"`
	State     ContractState `json:"state"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Version   int           `json:"version"`
	ParentID  *ContractID   `json:"parent_id,omitempty"`
}

type TemporalSocialContractEngine struct {
	contracts map[ContractID]*Contract
}

func NewTemporalSocialContractEngine() *TemporalSocialContractEngine {
	return &TemporalSocialContractEngine{
		contracts: make(map[ContractID]*Contract),
	}
}

func (tsce *TemporalSocialContractEngine) RegisterContract(c *Contract) {
	tsce.contracts[c.ID] = c
}

func (tsce *TemporalSocialContractEngine) GetContract(id ContractID) *Contract {
	return tsce.contracts[id]
}
