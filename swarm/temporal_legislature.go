package main

import "time"

type EpochTreaty struct {
	TreatyID      string
	MutationLimit float64
}

type TemporalBill struct {
	BillID        string
	Title         string
	Epoch         string
	Proposer      string
	Amendments    []ConstitutionAmendment
	TreatyChanges []TreatyAmendment
	PolicyChanges []PolicyChange
	Timestamp     time.Time
}

type ConstitutionAmendment struct {
	Field    string
	NewValue interface{}
}

type TreatyAmendment struct {
	TreatyID string
	Field    string
	NewValue interface{}
}

type PolicyChange struct {
	Field    string
	NewValue interface{}
}

type TemporalLegislatureEngine struct {
	constitution *TemporalConstitution
	treaties     map[string]EpochTreaty
	policy       *MeshPolicyConfig
}

func NewTemporalLegislatureEngine(
	c *TemporalConstitution,
	t map[string]EpochTreaty,
	p *MeshPolicyConfig,
) *TemporalLegislatureEngine {
	return &TemporalLegislatureEngine{
		constitution: c,
		treaties:     t,
		policy:       p,
	}
}

func (l *TemporalLegislatureEngine) Enact(bill TemporalBill) {
	for _, a := range bill.Amendments {
		applyConstitutionAmendment(l.constitution, a)
	}
	for _, ta := range bill.TreatyChanges {
		applyTreatyAmendment(l.treaties, ta)
	}
	for _, pc := range bill.PolicyChanges {
		applyPolicyChange(l.policy, pc)
	}
}

func applyConstitutionAmendment(c *TemporalConstitution, a ConstitutionAmendment) {
	switch a.Field {
	case "MaxMutationRate":
		c.MutationBoundaries.MaxMutationRate = a.NewValue.(float64)
	case "AllowTimeslips":
		c.TimeslipRules.AllowTimeslips = a.NewValue.(bool)
	}
}

func applyTreatyAmendment(treaties map[string]EpochTreaty, a TreatyAmendment) {
	treaty, ok := treaties[a.TreatyID]
	if !ok {
		return
	}
	switch a.Field {
	case "MutationLimit":
		treaty.MutationLimit = a.NewValue.(float64)
	}
	treaties[a.TreatyID] = treaty
}

func applyPolicyChange(p *MeshPolicyConfig, a PolicyChange) {
	switch a.Field {
	case "RiskThreshold":
		p.RiskThreshold = a.NewValue.(float64)
	}
}
