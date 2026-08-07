package main

import (
	"context"
	"time"
)

type QuorumVote struct {
	AgentID string
	Approve bool
	Weight  float64
}

type MeshQuorum struct {
	QuorumID   string
	Members    []string
	Approvals  int
	Rejections int
	Threshold  int
	Formed     bool
	Timestamp  time.Time
}

const QUORUM_SIZE = 4 // Align with our neighborhood size of 4
const SUPERMAJORITY = 3

func FormQuorumFromNeighborhood(n Neighborhood, votes []QuorumVote) MeshQuorum {
	quorum := MeshQuorum{
		QuorumID:  "quorum-" + n.CenterAgentID,
		Members:   n.Neighbors,
		Threshold: SUPERMAJORITY,
		Timestamp: time.Now(),
	}

	approveCount := 0
	rejectCount := 0

	voteMap := make(map[string]QuorumVote)
	for _, v := range votes {
		voteMap[v.AgentID] = v
	}

	for _, member := range n.Neighbors {
		v, ok := voteMap[member]
		if !ok {
			continue
		}
		if v.Approve {
			approveCount++
		} else {
			rejectCount++
		}
	}

	quorum.Approvals = approveCount
	quorum.Rejections = rejectCount
	quorum.Formed = approveCount >= SUPERMAJORITY

	return quorum
}

type MeshQuorumEngine struct {
	router TeleporterRouter
}

func NewMeshQuorumEngine(router TeleporterRouter) *MeshQuorumEngine {
	return &MeshQuorumEngine{
		router: router,
	}
}

func (e *MeshQuorumEngine) Tick(ctx context.Context, neighborhoods []Neighborhood, votes []QuorumVote) []MeshQuorum {
	quorums := make([]MeshQuorum, 0)
	for _, n := range neighborhoods {
		if len(n.Neighbors) < QUORUM_SIZE {
			continue
		}
		q := FormQuorumFromNeighborhood(n, votes)
		quorums = append(quorums, q)

		e.EmitQuorumTelemetry(ctx, q)
	}
	return quorums
}

func (e *MeshQuorumEngine) EmitQuorumTelemetry(ctx context.Context, q MeshQuorum) {
	env := &FirehoseEnvelope{
		Source:      "mesh-quorum",
		StreamID:    "quorum",
		Timestamp:   q.Timestamp,
		PayloadType: "mesh_event",
	}

	payload := MeshEventPayload{
		SigmaID:    q.QuorumID,
		Agent:      "mesh-quorum",
		Files:      q.Members,
		RiskScore:  float64(q.Rejections),
		Confidence: float64(q.Approvals) / float64(len(q.Members)),
	}

	_ = e.router.Route(ctx, env, payload)
}
