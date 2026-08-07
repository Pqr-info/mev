package governance

import (
	"fmt"
	"strings"
)

type PolicyDiff struct {
	AddedLineage          map[string][]string `json:"added_lineage"`
	RemovedLineage        map[string][]string `json:"removed_lineage"`
	IdentityClassChanges  map[string]string   `json:"identity_class_changes"`
	AddedCapabilities     map[string][]string `json:"added_capabilities"`
	RemovedCapabilities   map[string][]string `json:"removed_capabilities"`
}

func ComputeRiskScore(diff PolicyDiff) (float64, []string) {
	score := 0.0
	factors := []string{}

	// Identity class changes
	for role := range diff.IdentityClassChanges {
		score += 0.3
		factors = append(factors, fmt.Sprintf("identity class change for role: %s", role))
	}

	// Lineage additions
	for role, lineages := range diff.AddedLineage {
		for _, ln := range lineages {
			score += 0.2
			factors = append(factors, fmt.Sprintf("added lineage authorization '%s' for role: %s", ln, role))
		}
	}

	// Lineage removals
	for role, lineages := range diff.RemovedLineage {
		for _, ln := range lineages {
			score += 0.15
			factors = append(factors, fmt.Sprintf("removed lineage restriction '%s' for role: %s", ln, role))
		}
	}

	// Capability additions
	for role, caps := range diff.AddedCapabilities {
		for _, c := range caps {
			if c == "governance.mutate.policy" || c == "governance.rollback.policy" {
				score += 0.5
				factors = append(factors, fmt.Sprintf("elevated administrative capability '%s' granted to role: %s", c, role))
			} else if strings.Contains(c, "mutate") || strings.Contains(c, "rotate") || strings.Contains(c, "admin") {
				score += 0.35
				factors = append(factors, fmt.Sprintf("high-impact mutation capability '%s' granted to role: %s", c, role))
			} else {
				score += 0.1
				factors = append(factors, fmt.Sprintf("added capability '%s' for role: %s", c, role))
			}
		}
	}

	// Clamp score between 0.0 and 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score, factors
}
