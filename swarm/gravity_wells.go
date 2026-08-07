package main

type GravityWell struct {
	CenterAgentID string
	Strength      float64
	Members       []string
}

const GRAVITY_THRESHOLD = 0.3

func ComputeGravityWells(analyses []AgentAnalysis) []GravityWell {
	wells := make([]GravityWell, 0)

	for _, a := range analyses {
		strength := a.OpportunityScore * a.Confidence
		if strength > GRAVITY_THRESHOLD {
			wells = append(wells, GravityWell{
				CenterAgentID: a.AgentID,
				Strength:      strength,
				Members:       []string{a.AgentID},
			})
		}
	}

	return wells
}
