package main

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type TeleportRoute struct {
	Target     string
	Priority   int
	Suppressed bool
	Reason     string
	Timestamp  time.Time
}

type FirehoseTeleporter struct {
	tso              *TemporalStabilityOracle
	predictiveEngine PredictiveEngine
	evolutionCouncil *EvolutionCouncil
	bicameralEngine  *BicameralEngine
	router           TeleporterRouter
	alpacaFirehose   *AlpacaFirehose
}

func NewFirehoseTeleporter(
	tso *TemporalStabilityOracle,
	pe PredictiveEngine,
	ec *EvolutionCouncil,
	be *BicameralEngine,
	router TeleporterRouter,
	alpaca *AlpacaFirehose,
) *FirehoseTeleporter {
	return &FirehoseTeleporter{
		tso:              tso,
		predictiveEngine: pe,
		evolutionCouncil: ec,
		bicameralEngine:  be,
		router:           router,
		alpacaFirehose:   alpaca,
	}
}

func (f *FirehoseTeleporter) Route(
	ctx context.Context,
	env *FirehoseEnvelope,
	payload MeshEventPayload,
	epoch string,
) TeleportRoute {
	report := f.tso.Evaluate(epoch)

	suppressed := false
	target := "predictive-engine"
	reason := "normal"

	if report.StabilityClass == "fractured" {
		suppressed = true
		reason = "timeline-fractured"
	} else if payload.RiskScore > 0.8 {
		target = "evolution-council"
		reason = "high-risk"
	} else if report.StabilityClass == "unstable" {
		target = "bicameral-engine"
		reason = "unstable-epoch"
	}

	route := TeleportRoute{
		Target:     target,
		Priority:   computePriority(payload, report),
		Suppressed: suppressed,
		Reason:     reason,
		Timestamp:  time.Now(),
	}

	if f.alpacaFirehose != nil {
		_ = f.alpacaFirehose.Emit(AlpacaRecord{
			EventID:     payload.Agent,
			PayloadType: env.PayloadType,
			Risk:        payload.RiskScore,
			Stability:   report.StabilityClass,
			Epoch:       epoch,
			Timestamp:   time.Now(),
		})
	}

	if !route.Suppressed {
		f.dispatch(ctx, route, env, payload)
	}

	return route
}

func (f *FirehoseTeleporter) dispatch(
	ctx context.Context,
	route TeleportRoute,
	env *FirehoseEnvelope,
	payload MeshEventPayload,
) {
	switch route.Target {
	case "predictive-engine":
		_ = f.predictiveEngine.ObserveMeshEvent(payload)
	case "evolution-council":
		log.Info().Str("agent", payload.Agent).Msg("Dispatched telemetry target to Evolution Council")
	case "bicameral-engine":
		log.Info().Str("agent", payload.Agent).Msg("Dispatched telemetry target to Bicameral Engine")
	default:
		_ = f.router.Route(ctx, env, payload)
	}
}

func computePriority(
	payload MeshEventPayload,
	report TemporalStabilityReport,
) int {
	base := 1
	if payload.RiskScore > 0.7 {
		base += 2
	}
	if report.StabilityClass == "unstable" || report.StabilityClass == "fractured" {
		base += 2
	}
	return base
}
