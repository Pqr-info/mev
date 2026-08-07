package main

import "time"

type ConstitutionalRuleV2 struct {
	RuleID      string
	Description string
	Constraints map[string]float64 // e.g. "max_margin", "min_liquidity_buffer"
	Timestamp   time.Time
}

type TemporalConstitutionV2 struct {
	Rules map[string]*ConstitutionalRuleV2
}

type TemporalConstitutionEngineV2 struct {
	governance   *TemporalGovernanceEngineV2
	constitution *TemporalConstitutionV2
}

func NewTemporalConstitutionEngineV2(
	g *TemporalGovernanceEngineV2,
) *TemporalConstitutionEngineV2 {
	return &TemporalConstitutionEngineV2{
		governance: g,
		constitution: &TemporalConstitutionV2{
			Rules: make(map[string]*ConstitutionalRuleV2),
		},
	}
}

func (ce *TemporalConstitutionEngineV2) AddRule(
	ruleID string,
	desc string,
	constraints map[string]float64,
) *ConstitutionalRuleV2 {
	rule := &ConstitutionalRuleV2{
		RuleID:      ruleID,
		Description: desc,
		Constraints: constraints,
		Timestamp:   time.Now(),
	}

	ce.constitution.Rules[ruleID] = rule
	return rule
}

func (ce *TemporalConstitutionEngineV2) ValidateAction(
	actionID string,
) bool {
	action, ok := ce.governance.actions[actionID]
	if !ok {
		return false
	}

	for _, rule := range ce.constitution.Rules {
		for k, max := range rule.Constraints {
			if val, ok := action.Parameters[k]; ok && val > max {
				return false // violates constitutional constraint
			}
		}
	}

	return true
}

func (ce *TemporalConstitutionEngineV2) Enforce(
	actionID string,
) bool {
	if !ce.ValidateAction(actionID) {
		return false
	}

	action, ok := ce.governance.actions[actionID]
	if !ok {
		return false
	}

	ce.governance.MutatePolicy(action.Directive, action.Parameters)
	return true
}
