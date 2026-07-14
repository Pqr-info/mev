package main

import (
	"context"
	"fmt"
	"math"
)

// Address5D defines coordinates in 5-dimensional spacetime topology: [X, Y, Z, Epoch(T), Volatility(W)]
type Address5D struct {
	X float64
	Y float64
	Z float64
	T float64
	W float64
}

// Distance computes Euclidean distance in 5D coordinate space.
func (a Address5D) Distance(b Address5D) float64 {
	return math.Sqrt(
		math.Pow(a.X-b.X, 2) +
			math.Pow(a.Y-b.Y, 2) +
			math.Pow(a.Z-b.Z, 2) +
			math.Pow(a.T-b.T, 2) +
			math.Pow(a.W-b.W, 2),
	)
}

// PredictiveModelNode represents a node in the mesh network offering model access.
type PredictiveModelNode struct {
	NodeID   string
	Address  Address5D
	IsActive bool
	Model    string // e.g., "gemini-3.5-flash"
}

// MeshPredictiveRouter5D coordinates query routing to model providers.
type MeshPredictiveRouter5D struct {
	Nodes  []PredictiveModelNode
	replay *TimeMachineReplay
}

// NewMeshPredictiveRouter5D instantiates a new 5D routing manager.
func NewMeshPredictiveRouter5D(replay *TimeMachineReplay) *MeshPredictiveRouter5D {
	return &MeshPredictiveRouter5D{
		Nodes:  []PredictiveModelNode{},
		replay: replay,
	}
}

// RegisterNode adds a model provider node to the 5D grid map.
func (r *MeshPredictiveRouter5D) RegisterNode(node PredictiveModelNode) {
	r.Nodes = append(r.Nodes, node)
}

// FindClosestNode looks up the nearest active model provider node using 5D addressing.
func (r *MeshPredictiveRouter5D) FindClosestNode(sender Address5D) (*PredictiveModelNode, error) {
	var closest *PredictiveModelNode
	minDist := math.MaxFloat64

	for i := range r.Nodes {
		node := &r.Nodes[i]
		if !node.IsActive {
			continue
		}
		dist := sender.Distance(node.Address)
		if dist < minDist {
			minDist = dist
			closest = node
		}
	}

	if closest == nil {
		return nil, fmt.Errorf("no active model provider nodes found in 5D mesh space")
	}
	return closest, nil
}

// RouteQuery routes an LLM request from a 5D mesh address to the closest active model provider node.
func (r *MeshPredictiveRouter5D) RouteQuery(ctx context.Context, sender Address5D, query string) (string, error) {
	node, err := r.FindClosestNode(sender)
	if err != nil {
		return "", err
	}

	fmt.Printf("[5D-ROUTER] Routing query from address (%.2f,%.2f,%.2f,%.2f,%.2f) to node %s (%s) at distance %.4f\n",
		sender.X, sender.Y, sender.Z, sender.T, sender.W, node.NodeID, node.Model, sender.Distance(node.Address),
	)

	// Delegate LLM generation to the mothership replay engine (Gemini 3.5 Flash brain)
	return r.replay.ConsultMothership(ctx, query)
}
