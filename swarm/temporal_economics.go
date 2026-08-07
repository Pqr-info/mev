package main

import "time"

type EconomicReport struct {
	Epoch             string
	InflationRate     float64
	StabilityRisk     float64
	DivergenceRisk    float64
	ReplayScarcity    float64
	TimeslipScarcity  float64
	MutationDemand    float64
	MarketEquilibrium float64
	Timestamp         time.Time
}

type TemporalEconomicEngine struct {
	market *EpochMarketEngine
}

func NewTemporalEconomicEngine(m *EpochMarketEngine) *TemporalEconomicEngine {
	return &TemporalEconomicEngine{
		market: m,
	}
}

func (e *TemporalEconomicEngine) Analyze(epoch string) EconomicReport {
	commodities := e.market.commodities

	var replayPrice, timeslipPrice, mutationPrice float64

	for _, c := range commodities {
		if c.Epoch != epoch {
			continue
		}
		switch c.Type {
		case "replay-right":
			replayPrice = c.Price
		case "timeslip-right":
			timeslipPrice = c.Price
		case "mutation-bandwidth":
			mutationPrice = c.Price
		}
	}

	inflation := (replayPrice + timeslipPrice + mutationPrice) / 3.0
	stabilityRisk := replayPrice * 0.1
	divergenceRisk := timeslipPrice * 0.2
	scarcityReplay := 1.0 / (replayPrice + 0.0001)
	scarcityTimeslip := 1.0 / (timeslipPrice + 0.0001)
	mutationDemand := mutationPrice * 1.5
	equilibrium := (replayPrice + timeslipPrice + mutationPrice) / 3.0

	return EconomicReport{
		Epoch:             epoch,
		InflationRate:     inflation,
		StabilityRisk:     stabilityRisk,
		DivergenceRisk:    divergenceRisk,
		ReplayScarcity:    scarcityReplay,
		TimeslipScarcity:  scarcityTimeslip,
		MutationDemand:    mutationDemand,
		MarketEquilibrium: equilibrium,
		Timestamp:         time.Now(),
	}
}

func (e *TemporalEconomicEngine) AdjustMarket(epoch string, report EconomicReport) {
	for id, c := range e.market.commodities {
		if c.Epoch != epoch {
			continue
		}

		c.Price *= (1.0 + report.InflationRate*0.01)
		e.market.commodities[id] = c
	}
}
