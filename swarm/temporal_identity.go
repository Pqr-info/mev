package main

import "time"

type TemporalIdentity struct {
	IdentityID  string
	Type        string   // "citizen", "node", "role"
	Roles       []string // "market_actor", "bank", "regulator", etc.
	Permissions []string // "trade", "lend", "govern", "validate", etc.
	CreatedAt   time.Time
}

type IdentityRegistry struct {
	Identities map[string]*TemporalIdentity
}

type TemporalIdentityEngine struct {
	constitution *TemporalConstitutionEngineV2
	registry     *IdentityRegistry
}

func NewTemporalIdentityEngine(
	c *TemporalConstitutionEngineV2,
) *TemporalIdentityEngine {
	return &TemporalIdentityEngine{
		constitution: c,
		registry: &IdentityRegistry{
			Identities: make(map[string]*TemporalIdentity),
		},
	}
}

func (ie *TemporalIdentityEngine) CreateIdentity(
	id string,
	idType string,
	roles []string,
	permissions []string,
) *TemporalIdentity {
	ident := &TemporalIdentity{
		IdentityID:  id,
		Type:        idType,
		Roles:       roles,
		Permissions: permissions,
		CreatedAt:   time.Now(),
	}

	ie.registry.Identities[id] = ident
	return ident
}

func (ie *TemporalIdentityEngine) AssignPermission(
	identityID string,
	permission string,
) bool {
	ident, ok := ie.registry.Identities[identityID]
	if !ok {
		return false
	}

	ident.Permissions = append(ident.Permissions, permission)
	return true
}

func (ie *TemporalIdentityEngine) ValidateIdentity(
	identityID string,
) bool {
	ident, ok := ie.registry.Identities[identityID]
	if !ok {
		return false
	}

	// Validate against constitutional rule constraints limiting max_permissions
	for _, rule := range ie.constitution.constitution.Rules {
		if maxPerms, ok := rule.Constraints["max_permissions"]; ok {
			if float64(len(ident.Permissions)) > maxPerms {
				return false
			}
		}
	}

	return true
}
