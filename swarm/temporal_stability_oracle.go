package main

import "time"

type TemporalStabilityReport struct {
	DriftAverage      float64
	VolatilityAverage float64
	DiscontinuityRisk float64
	StabilityClass    string
	Timestamp         time.Time
}

type TemporalStabilityOracle struct {
	tme *TemporalMemoryEngine
}

func NewTemporalStabilityOracle(t *TemporalMemoryEngine) *TemporalStabilityOracle {
	return &TemporalStabilityOracle{
		tme: t,
	}
}

func (o *TemporalStabilityOracle) Evaluate(epoch string) TemporalStabilityReport {
	events := o.tme.GetEventsByEpoch(epoch)

	driftAvg := computeDriftAverage(events)
	volAvg := computeVolatilityAverage(events)
	discRisk := computeDiscontinuityRisk(events)

	class := classifyStability(driftAvg, volAvg, discRisk)

	return TemporalStabilityReport{
		DriftAverage:      driftAvg,
		VolatilityAverage: volAvg,
		DiscontinuityRisk: discRisk,
		StabilityClass:    class,
		Timestamp:         time.Now(),
	}
}

func computeDriftAverage(events []TemporalEvent) float64 {
	if len(events) == 0 {
		return 0
	}
	sum := 0.0
	for _, ev := range events {
		sum += ev.Drift
	}
	return sum / float64(len(events))
}

func computeVolatilityAverage(events []TemporalEvent) float64 {
	if len(events) == 0 {
		return 0
	}
	sum := 0.0
	for _, ev := range events {
		sum += ev.Volatility
	}
	return sum / float64(len(events))
}

func computeDiscontinuityRisk(events []TemporalEvent) float64 {
	if len(events) < 2 {
		return 0
	}
	risk := 0.0
	for i := 1; i < len(events); i++ {
		delta := events[i].Timestamp.Sub(events[i-1].Timestamp).Seconds()
		if delta > 30 { // arbitrary discontinuity threshold
			risk += 0.1
		}
	}
	return risk
}

func classifyStability(drift, vol, disc float64) string {
	if disc > 0.5 {
		return "fractured"
	}
	if drift > 0.7 || vol > 0.8 {
		return "unstable"
	}
	if drift > 0.4 || vol > 0.5 {
		return "recovering"
	}
	return "stable"
}
