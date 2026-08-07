package main

import "time"

type CreditScore struct {
	BorrowerID      string
	Epoch           string
	StabilityRisk   float64
	DivergenceRisk  float64
	CollateralValue float64
	LiquidityFactor float64
	DerivativeRisk  float64
	Score           float64
	Timestamp       time.Time
}

type TemporalCreditEngine struct {
	economics *TemporalEconomicEngine
	market    *EpochMarketEngine
	bank      *EpochBank
}

func NewTemporalCreditEngine(
	e *TemporalEconomicEngine,
	m *EpochMarketEngine,
	b *EpochBank,
) *TemporalCreditEngine {
	return &TemporalCreditEngine{
		economics: e,
		market:    m,
		bank:      b,
	}
}

func (ce *TemporalCreditEngine) Score(
	borrowerID string,
	epoch string,
) CreditScore {
	report := ce.economics.Analyze(epoch)

	acc, ok := ce.bank.Accounts["acct-"+borrowerID+"-"+epoch]
	collateral := 0.0
	derivativeRisk := 0.0

	if ok {
		collateral = ce.bank.EvaluateCollateral(acc.AccountID)
		for _, qty := range acc.Derivatives {
			derivativeRisk += qty * 0.1
		}
	}

	liquidity := report.MarketEquilibrium * 0.2

	score := 1000.0 -
		(report.StabilityRisk * 50.0) -
		(report.DivergenceRisk * 50.0) +
		(collateral * 0.1) +
		(liquidity * 10.0) -
		(derivativeRisk * 20.0)

	return CreditScore{
		BorrowerID:      borrowerID,
		Epoch:           epoch,
		StabilityRisk:   report.StabilityRisk,
		DivergenceRisk:  report.DivergenceRisk,
		CollateralValue: collateral,
		LiquidityFactor: liquidity,
		DerivativeRisk:  derivativeRisk,
		Score:           score,
		Timestamp:       time.Now(),
	}
}

func (ce *TemporalCreditEngine) CreditLimit(score CreditScore) float64 {
	if score.Score < 300 {
		return 0
	}
	if score.Score < 600 {
		return score.CollateralValue * 0.2
	}
	if score.Score < 800 {
		return score.CollateralValue * 0.5
	}
	return score.CollateralValue * 1.0
}
