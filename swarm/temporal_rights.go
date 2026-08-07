package main

import "time"

type TemporalRight struct {
	RightID     string
	Description string
	Category    string   // "economic", "governance", "privacy", "participation"
	Guarantees  []string // specific protections
	Timestamp   time.Time
}

type RightsRegistry struct {
	Rights map[string]*TemporalRight
}

type TemporalRightsEngine struct {
	constitution *TemporalConstitutionEngineV2
	citizenship  *TemporalCitizenshipEngine
	registry     *RightsRegistry
}

func NewTemporalRightsEngine(
	ce *TemporalConstitutionEngineV2,
	tce *TemporalCitizenshipEngine,
) *TemporalRightsEngine {
	return &TemporalRightsEngine{
		constitution: ce,
		citizenship:  tce,
		registry: &RightsRegistry{
			Rights: make(map[string]*TemporalRight),
		},
	}
}

func (tre *TemporalRightsEngine) AddRight(
	id string,
	desc string,
	category string,
	guarantees []string,
) *TemporalRight {
	r := &TemporalRight{
		RightID:     id,
		Description: desc,
		Category:    category,
		Guarantees:  guarantees,
		Timestamp:   time.Now(),
	}

	tre.registry.Rights[id] = r
	return r
}

func (tre *TemporalRightsEngine) ValidateCitizenRights(
	citizenID string,
) bool {
	citizen, ok := tre.citizenship.registry.Citizens[citizenID]
	if !ok {
		return false
	}

	for _, right := range tre.registry.Rights {
		for _, guarantee := range right.Guarantees {
			if guarantee == "no_suspension_without_due_process" &&
				citizen.Status == "suspended" {
				return false
			}
		}
	}

	return true
}

func (tre *TemporalRightsEngine) ValidateGovernanceAction(
	actionID string,
) bool {
	action, ok := tre.constitution.governance.actions[actionID]
	if !ok {
		return false
	}

	for _, right := range tre.registry.Rights {
		for _, guarantee := range right.Guarantees {
			if guarantee == "no_arbitrary_policy_change" &&
				action.Directive == "adjust_policy" {
				return false
			}
		}
	}

	return true
}
