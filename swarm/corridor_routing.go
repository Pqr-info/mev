package main

type CorridorTier int

const (
	TierMYPN CorridorTier = iota // Short/local tethers
	TierNEB                      // Long geometric rulers
	TierACTN                     // Cross-link stabilizers
	TierNBL                      // High-fidelity long-range HT
	TierDES                      // Mesh-wide reinforcement backbone
	TierTPM                      // Regulatory / Troponin gated
)

type Corridor struct {
	ID         string
	From       string
	To         string
	Tier       CorridorTier
	BaseWeight float64 // baseline routing weight (lower = preferred)
	Stress     float64 // S_ij
	SMax       float64
	IsActive   bool
}

type RouteEdge struct {
	CorridorID string
	From       string
	To         string
	Weight     float64
}

type CrisisType int

const (
	CrisisDimensionalDrift CrisisType = iota
	CrisisCrossMeshCollapse
	CrisisEntropyWave
)

type CrisisEvent struct {
	Type      CrisisType
	Magnitude float64
}

// crisisWeightFactor computes the global crisis modifier factor for routing
func crisisWeightFactor(c Corridor, events []CrisisEvent) float64 {
	factor := 1.0

	for _, e := range events {
		switch e.Type {
		case CrisisDimensionalDrift:
			// Drift: penalize low tiers, favor DES/NBL backbone
			switch c.Tier {
			case TierDES:
				factor *= 0.7
			case TierNBL:
				factor *= 0.8
			case TierACTN:
				factor *= 1.0
			case TierNEB:
				factor *= 1.2
			case TierMYPN:
				factor *= 1.4
			}
		case CrisisCrossMeshCollapse:
			// Collapse: overall stress compression; backbone relative advantage increases
			switch c.Tier {
			case TierDES:
				factor *= 0.9
			case TierNBL:
				factor *= 1.0
			case TierACTN:
				factor *= 1.2
			case TierNEB:
				factor *= 1.5
			case TierMYPN:
				factor *= 1.8
			}
		case CrisisEntropyWave:
			// Entropy: decay/leak of signal across all tiers
			switch c.Tier {
			case TierDES:
				factor *= 0.85
			case TierNBL:
				factor *= 0.95
			case TierACTN:
				factor *= 1.1
			case TierNEB:
				factor *= 1.3
			case TierMYPN:
				factor *= 1.5
			}
		}
	}

	return factor
}

// AdjustRoutingWeights dynamically contracts weights based on the localized stress tensor and crisis events
func AdjustRoutingWeights(corridors []Corridor, events []CrisisEvent) []RouteEdge {
	edges := make([]RouteEdge, 0, len(corridors))

	for _, c := range corridors {
		if !c.IsActive {
			continue
		}

		stressRatio := c.Stress / c.SMax
		if c.SMax == 0 {
			stressRatio = 0
		}
		weight := c.BaseWeight

		switch {
		case stressRatio < 0.5:
			weight = c.BaseWeight

		case stressRatio >= 0.5 && stressRatio < 1.0:
			// Medium stress: soft contraction towards primary backbone
			switch c.Tier {
			case TierDES:
				weight *= 0.4
			case TierNBL:
				weight *= 0.6
			case TierACTN:
				weight *= 0.8
			case TierNEB:
				weight *= 1.2
			case TierMYPN:
				weight *= 1.4
			}

		default:
			// stressRatio >= 1.0: hard contraction / overload routing shift
			switch c.Tier {
			case TierDES:
				weight *= 0.2
			case TierNBL:
				weight *= 0.3
			case TierACTN:
				weight *= 0.6
			case TierNEB:
				weight *= 2.0
			case TierMYPN:
				weight *= 3.0
			}
		}

		// Apply crisis factors
		weight *= crisisWeightFactor(c, events)

		edges = append(edges, RouteEdge{
			CorridorID: c.ID,
			From:       c.From,
			To:         c.To,
			Weight:     weight,
		})
	}

	return edges
}

// EffectiveHopCost counts higher-tier stable corridors as fractional hops to contract routing
func EffectiveHopCost(c Corridor, stressRatio float64, events []CrisisEvent) float64 {
	baseHop := 1.0
	var hopCost float64

	if stressRatio < 0.5 {
		hopCost = baseHop
	} else if stressRatio < 1.0 {
		// Medium stress: mild hop compression for stabilizer/backbone paths
		switch c.Tier {
		case TierDES:
			hopCost = 0.7
		case TierNBL:
			hopCost = 0.8
		case TierACTN:
			hopCost = 1.0
		case TierNEB:
			hopCost = 1.3
		case TierMYPN:
			hopCost = 1.5
		default:
			hopCost = baseHop
		}
	} else {
		// High stress: aggressive contraction (sliding filament logic)
		switch c.Tier {
		case TierDES:
			hopCost = 0.4
		case TierNBL:
			hopCost = 0.5
		case TierACTN:
			hopCost = 0.9
		case TierNEB:
			hopCost = 2.0
		case TierMYPN:
			hopCost = 3.0
		default:
			hopCost = baseHop
		}
	}

	return hopCost * crisisWeightFactor(c, events)
}

// BuildRoutingGraph compiles dynamic edge weights based on hop costs, base weights, and crisis modifiers
func BuildRoutingGraph(corridors []Corridor, events []CrisisEvent) map[string][]RouteEdge {
	graph := make(map[string][]RouteEdge)

	for _, c := range corridors {
		if !c.IsActive {
			continue
		}
		
		stressRatio := 0.0
		if c.SMax > 0 {
			stressRatio = c.Stress / c.SMax
		}
		
		hopCost := EffectiveHopCost(c, stressRatio, events)
		weight := c.BaseWeight * hopCost

		graph[c.From] = append(graph[c.From], RouteEdge{
			CorridorID: c.ID,
			From:       c.From,
			To:         c.To,
			Weight:     weight,
		})
	}

	return graph
}
