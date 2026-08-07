package main

import "time"

type MonetaryPolicy struct {
	Epoch              string
	TargetInflation    float64
	InterestRate       float64
	LiquidityInjection float64
	StabilityBuffer    float64
	DivergenceBuffer   float64
	Timestamp          time.Time
}

type TemporalMonetaryPolicyEngine struct {
	market    *EpochMarketEngine
	economics *TemporalEconomicEngine
}

func NewTemporalMonetaryPolicyEngine(
	m *EpochMarketEngine,
	e *TemporalEconomicEngine,
) *TemporalMonetaryPolicyEngine {
	return &TemporalMonetaryPolicyEngine{
		market:    m,
		economics: e,
	}
}

func (mp *TemporalMonetaryPolicyEngine) GeneratePolicy(
	epoch string,
) MonetaryPolicy {
	report := mp.economics.Analyze(epoch)

	return MonetaryPolicy{
		Epoch:              epoch,
		TargetInflation:    1.0,
		InterestRate:       report.InflationRate * 0.5,
		LiquidityInjection: report.MarketEquilibrium * 0.1,
		StabilityBuffer:    report.StabilityRisk * 0.2,
		DivergenceBuffer:   report.DivergenceRisk * 0.2,
		Timestamp:          time.Now(),
	}
}

func (mp *TemporalMonetaryPolicyEngine) ApplyPolicy(
	policy MonetaryPolicy,
) {
	for id, c := range mp.market.commodities {
		if c.Epoch != policy.Epoch {
			continue
		}

		c.Price *= (1.0 + policy.LiquidityInjection*0.01)
		c.Price *= (1.0 - policy.InterestRate*0.01)

		mp.market.commodities[id] = c
	}
}
