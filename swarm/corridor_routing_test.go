package main

import (
	"testing"
)

func TestAdjustRoutingWeights(t *testing.T) {
	corridors := []Corridor{
		{
			ID:         "c1",
			From:       "node-a",
			To:         "node-b",
			Tier:       TierDES,
			BaseWeight: 100.0,
			Stress:     8.0,
			SMax:       10.0, // stressRatio = 0.8 (Medium)
			IsActive:   true,
		},
		{
			ID:         "c2",
			From:       "node-b",
			To:         "node-c",
			Tier:       TierMYPN,
			BaseWeight: 10.0,
			Stress:     12.0,
			SMax:       10.0, // stressRatio = 1.2 (High)
			IsActive:   true,
		},
	}

	edges := AdjustRoutingWeights(corridors, nil)
	if len(edges) != 2 {
		t.Fatalf("Expected 2 edges, got %d", len(edges))
	}

	// c1 (TierDES) at 0.8 stress should have weight = 100.0 * 0.4 = 40.0
	if edges[0].Weight != 40.0 {
		t.Errorf("Expected weight 40.0, got %f", edges[0].Weight)
	}

	// c2 (TierMYPN) at 1.2 stress should have weight = 10.0 * 3.0 = 30.0
	if edges[1].Weight != 30.0 {
		t.Errorf("Expected weight 30.0, got %f", edges[1].Weight)
	}
}

func TestBuildRoutingGraph(t *testing.T) {
	corridors := []Corridor{
		{
			ID:         "c1",
			From:       "node-a",
			To:         "node-b",
			Tier:       TierDES,
			BaseWeight: 100.0,
			Stress:     8.0,
			SMax:       10.0,
			IsActive:   true,
		},
	}

	graph := BuildRoutingGraph(corridors, nil)
	nodeAEdges, ok := graph["node-a"]
	if !ok || len(nodeAEdges) != 1 {
		t.Fatalf("Expected 1 edge from node-a")
	}

	// c1 (TierDES, stressRatio = 0.8) -> hopCost = 0.7 -> weight = 100.0 * 0.7 = 70.0
	if nodeAEdges[0].Weight != 70.0 {
		t.Errorf("Expected weight 70.0, got %f", nodeAEdges[0].Weight)
	}
}

func TestCrisisRoutingModifiers(t *testing.T) {
	corridors := []Corridor{
		{
			ID:         "c1",
			From:       "node-a",
			To:         "node-b",
			Tier:       TierDES,
			BaseWeight: 100.0,
			Stress:     0.0,
			SMax:       10.0,
			IsActive:   true,
		},
		{
			ID:         "c2",
			From:       "node-b",
			To:         "node-c",
			Tier:       TierMYPN,
			BaseWeight: 100.0,
			Stress:     0.0,
			SMax:       10.0,
			IsActive:   true,
		},
	}

	events := []CrisisEvent{
		{
			Type:      CrisisDimensionalDrift,
			Magnitude: 1.0,
		},
	}

	edges := AdjustRoutingWeights(corridors, events)

	// Under DimensionalDrift:
	// c1 (TierDES) weight factor should be 0.7 -> weight should decrease to 70.0
	// c2 (TierMYPN) weight factor should be 1.4 -> weight should increase to 140.0
	if edges[0].Weight != 70.0 {
		t.Errorf("Expected c1 weight 70.0 under drift, got %f", edges[0].Weight)
	}

	if edges[1].Weight != 140.0 {
		t.Errorf("Expected c2 weight 140.0 under drift, got %f", edges[1].Weight)
	}
}
