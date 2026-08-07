package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type TemporalEvent struct {
	EventID     string    `json:"event_id"`
	PayloadType string    `json:"payload_type"`
	Drift       float64   `json:"drift"`
	Volatility  float64   `json:"volatility"`
	Recurrence  float64   `json:"recurrence"`
	Timestamp   time.Time `json:"timestamp"`
	SigmaID     string    `json:"sigma_id"`
	Agent       string    `json:"agent"`
	Files       []string  `json:"files"`
	Risk        float64   `json:"risk"`
	Confidence  float64   `json:"confidence"`
	Epoch       string    `json:"epoch"`
	Anchor      string    `json:"anchor"`
}

type TimeMachine interface {
	CreateCheckpoint(summary string) error
}

type MockTimeMachine struct{}

func (m *MockTimeMachine) CreateCheckpoint(summary string) error {
	log.Info().Str("summary", summary).Msg("MockTimeMachine created state checkpoint")
	return nil
}

type TeleporterRouter interface {
	Route(ctx context.Context, env *FirehoseEnvelope, payload interface{}) error
}

type DefaultRouter struct {
	MinRiskThreshold float64
	predictiveEngine PredictiveEngine
	timeMachine      TimeMachine
}

func NewDefaultRouter(minRisk float64, pe PredictiveEngine, tm TimeMachine) *DefaultRouter {
	return &DefaultRouter{
		MinRiskThreshold: minRisk,
		predictiveEngine: pe,
		timeMachine:      tm,
	}
}

func (r *DefaultRouter) Route(ctx context.Context, env *FirehoseEnvelope, payload interface{}) error {
	switch p := payload.(type) {
	case MeshEventPayload:
		if err := r.predictiveEngine.ObserveMeshEvent(p); err != nil {
			return err
		}
		r.routeTimeMachine(env, p)
	case MetricPayload:
		if err := r.predictiveEngine.ObserveMetric(p); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown payload type: %T", payload)
	}
	return nil
}

func (r *DefaultRouter) routeTimeMachine(env *FirehoseEnvelope, p MeshEventPayload) {
	features := r.predictiveEngine.ComputeFeatures()
	if p.RiskScore > r.MinRiskThreshold || features.Volatility > 0.5 {
		log.Warn().
			Float64("event_risk", p.RiskScore).
			Float64("volatility", features.Volatility).
			Msg("🚨 Volatility or risk threshold breached. Generating Time Machine checkpoint...")
		
		_ = r.timeMachine.CreateCheckpoint("Auto-checkpoint: threshold breach")
	}
}
