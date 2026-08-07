package governance

import (
	"testing"
)

func TestComputeRiskScoreElevatedAdmin(t *testing.T) {
	diff := PolicyDiff{
		AddedCapabilities: map[string][]string{
			"Analyst": {"governance.mutate.policy"},
		},
		AddedLineage: map[string][]string{
			"Sentinel": {"corridor.new_horizon"},
		},
	}

	score, factors := ComputeRiskScore(diff)

	// Elevated capability (0.5) + Lineage addition (0.2) = 0.7 score
	if score != 0.7 {
		t.Errorf("expected risk score of 0.7, got %.4f", score)
	}

	if len(factors) != 2 {
		t.Errorf("expected 2 risk factors, got %d", len(factors))
	}
}
