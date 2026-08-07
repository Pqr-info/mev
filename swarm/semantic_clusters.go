package main

type SemanticCluster struct {
	Label   string
	Members []string
}

func ComputeSemanticClusters(analyses []AgentAnalysis) []SemanticCluster {
	high := SemanticCluster{Label: "high-confidence", Members: []string{}}
	low := SemanticCluster{Label: "low-confidence", Members: []string{}}

	for _, a := range analyses {
		if a.Confidence > 0.7 {
			high.Members = append(high.Members, a.AgentID)
		} else {
			low.Members = append(low.Members, a.AgentID)
		}
	}

	return []SemanticCluster{high, low}
}
