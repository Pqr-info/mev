package main

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
)

type AgentPosition struct {
	AgentID string
	X       float64
	Y       float64
}

type Neighborhood struct {
	CenterAgentID string
	Neighbors     []string
}

type TopologySnapshot struct {
	Timestamp     time.Time
	Positions     []AgentPosition
	Neighborhoods []Neighborhood
}

const K_NEIGHBORS = 4 // Changed K to 4 since we have fewer test agents in mock swarm

func ComputeNeighborhoods(positions []AgentPosition) []Neighborhood {
	neighborhoods := make([]Neighborhood, 0)

	for _, center := range positions {
		type dist struct {
			id    string
			value float64
		}
		dists := make([]dist, 0)

		for _, other := range positions {
			if other.AgentID == center.AgentID {
				continue
			}
			dx := center.X - other.X
			dy := center.Y - other.Y
			d := math.Sqrt(dx*dx + dy*dy)
			dists = append(dists, dist{other.AgentID, d})
		}

		sort.Slice(dists, func(i, j int) bool {
			return dists[i].value < dists[j].value
		})

		neighbors := make([]string, 0)
		for i := 0; i < K_NEIGHBORS && i < len(dists); i++ {
			neighbors = append(neighbors, dists[i].id)
		}

		neighborhoods = append(neighborhoods, Neighborhood{
			CenterAgentID: center.AgentID,
			Neighbors:     neighbors,
		})
	}

	return neighborhoods
}

type TopologyDrift struct {
	CenterAgentID string
	DriftScore    float64
}

func ComputeTopologyDrift(prev, curr []Neighborhood) []TopologyDrift {
	drift := make([]TopologyDrift, 0)

	prevMap := make(map[string][]string)
	for _, n := range prev {
		prevMap[n.CenterAgentID] = n.Neighbors
	}

	for _, n := range curr {
		prevNeighbors, ok := prevMap[n.CenterAgentID]
		if !ok {
			continue
		}
		score := 1.0 - jaccard(prevNeighbors, n.Neighbors)
		drift = append(drift, TopologyDrift{
			CenterAgentID: n.CenterAgentID,
			DriftScore:    score,
		})
	}

	return drift
}

func jaccard(a, b []string) float64 {
	setA := make(map[string]struct{})
	setB := make(map[string]struct{})
	for _, x := range a {
		setA[x] = struct{}{}
	}
	for _, x := range b {
		setB[x] = struct{}{}
	}

	inter, union := 0, 0
	for k := range setA {
		union++
		if _, ok := setB[k]; ok {
			inter++
		}
	}
	for k := range setB {
		if _, ok := setA[k]; !ok {
			union++
		}
	}
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

type TopologyEngine struct {
	prevNeighborhoods []Neighborhood
	router            TeleporterRouter
}

func NewTopologyEngine(router TeleporterRouter) *TopologyEngine {
	return &TopologyEngine{
		router: router,
	}
}

func (t *TopologyEngine) Tick(ctx context.Context, positions []AgentPosition) {
	curr := ComputeNeighborhoods(positions)
	drift := ComputeTopologyDrift(t.prevNeighborhoods, curr)

	t.EmitTopologyTelemetry(ctx, curr, drift)

	t.prevNeighborhoods = curr
}

func (t *TopologyEngine) EmitTopologyTelemetry(ctx context.Context, neighborhoods []Neighborhood, drift []TopologyDrift) {
	avgDrift := 0.0
	if len(drift) > 0 {
		for _, d := range drift {
			avgDrift += d.DriftScore
		}
		avgDrift /= float64(len(drift))
	}

	log.Info().
		Float64("average_drift", avgDrift).
		Int("neighborhood_count", len(neighborhoods)).
		Msg("Topology engine tick completed")

	env := &FirehoseEnvelope{
		Source:      "topology-engine",
		StreamID:    "topology",
		Timestamp:   time.Now(),
		PayloadType: "mesh_event",
	}

	payload := MeshEventPayload{
		SigmaID:    "topology-engine",
		Agent:      "topology",
		Files:      []string{},
		RiskScore:  avgDrift,
		Confidence: 1.0,
	}

	_ = t.router.Route(ctx, env, payload)
}
