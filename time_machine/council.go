package timemachine

import (
	"errors"
	"math"
)

// QuorumState represents the multi-sig approvals from the Council of Five.
type QuorumState struct {
	StabilityApproved   bool
	DeterminismApproved bool
	SafetyApproved      bool
	AdaptationApproved  bool
	MutationApproved    bool
	CopilotSignature    []byte
}

var ErrQuorumNotAchieved = errors.New("council of five quorum not achieved")

// VerifyQuorum checks if the quorum requirements are met for a JetWeb rollback.
// Requires: 3 Rust seats, 2 Go seats, and a valid Copilot human signature.
func VerifyQuorum(state QuorumState) error {
	rustApprovals := 0
	if state.StabilityApproved {
		rustApprovals++
	}
	if state.DeterminismApproved {
		rustApprovals++
	}
	if state.SafetyApproved {
		rustApprovals++
	}

	goApprovals := 0
	if state.AdaptationApproved {
		goApprovals++
	}
	if state.MutationApproved {
		goApprovals++
	}

	if rustApprovals != 3 || goApprovals != 2 || len(state.CopilotSignature) == 0 {
		return ErrQuorumNotAchieved
	}

	return nil
}

// CalculateConsensusThreshold dynamically scales the network consensus threshold (Tc)
// based on the temporal gradient to prevent 51% Causal Sabotage.
// Formula: Tc = 0.51 + 0.40 * (1 - e^(-lambda * gradient))
func CalculateConsensusThreshold(gradient float64) float64 {
	lambda := 2.0 // Decay constant
	return 0.51 + 0.40*(1.0-math.Exp(-lambda*gradient))
}

// CalculateNodeWeight computes the voting weight (W) of a node.
// Formula: W = Compute * TrustScore * e^(-k * driftPenalty)
func CalculateNodeWeight(compute float64, trustScore float64, driftPenalty float64) float64 {
	k := 1.5 // Sensitivity constant
	return compute * trustScore * math.Exp(-k*driftPenalty)
}

type TemporalStability string

const (
    Stable     TemporalStability = "stable"
    Chaotic    TemporalStability = "chaotic"
    Metastable TemporalStability = "metastable"
)

type TemporalMarker struct {
    ID        string            `json:"id"`
    TrackID   uint8             `json:"track_id"`
    Tick      uint64            `json:"tick"`
    CreatedAt int64             `json:"created_at"`
    EchoCycle uint32            `json:"echo_cycle"`
    Stability TemporalStability `json:"stability"`
}

// Protobuf bindings (mocked for now since protoc generation is skipped locally)
type PbTemporalMarker struct {
    Id        string
    TrackId   uint32
    Tick      uint64
    CreatedAt int64
    EchoCycle uint32
    Stability int32
}

func ToPbMarker(m TemporalMarker) *PbTemporalMarker {
    return &PbTemporalMarker{
        Id:        m.ID,
        TrackId:   uint32(m.TrackID),
        Tick:      m.Tick,
        CreatedAt: m.CreatedAt,
        EchoCycle: m.EchoCycle,
        Stability: toPbStability(m.Stability),
    }
}

func toPbStability(s TemporalStability) int32 {
    switch s {
    case Stable:
        return 1 // STABLE
    case Chaotic:
        return 2 // CHAOTIC
    case Metastable:
        return 3 // METASTABLE
    default:
        return 0 // STABILITY_UNSPECIFIED
    }
}

