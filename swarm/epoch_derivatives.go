package main

import "time"

type DerivativeContract struct {
	ContractID   string
	Epoch        string
	UnderlyingID string
	Type         string // "future", "option", "swap"
	StrikePrice  float64
	Quantity     float64
	Expiration   time.Time
	Premium      float64
	Timestamp    time.Time
}

type DerivativeOrder struct {
	OrderID     string
	ContractID  string
	Side        string // BUY or SELL
	Quantity    float64
	LimitPrice  float64
	Timestamp   time.Time
}

type EpochDerivativesEngine struct {
	contracts map[string]DerivativeContract
	orderBook []DerivativeOrder
}

func NewEpochDerivativesEngine() *EpochDerivativesEngine {
	return &EpochDerivativesEngine{
		contracts: make(map[string]DerivativeContract),
		orderBook: []DerivativeOrder{},
	}
}

func (e *EpochDerivativesEngine) ListContract(c DerivativeContract) {
	e.contracts[c.ContractID] = c
}

func (e *EpochDerivativesEngine) SubmitOrder(o DerivativeOrder) {
	o.OrderID = "deriv-" + o.ContractID + "-" + time.Now().Format("150405")
	e.orderBook = append(e.orderBook, o)
}

func (e *EpochDerivativesEngine) Match() []DerivativeOrder {
	fills := []DerivativeOrder{}

	for _, o := range e.orderBook {
		contract, ok := e.contracts[o.ContractID]
		if !ok {
			continue
		}

		if o.Side == "BUY" && o.LimitPrice >= contract.Premium {
			fills = append(fills, o)
		}
		if o.Side == "SELL" && o.LimitPrice <= contract.Premium {
			fills = append(fills, o)
		}
	}

	return fills
}

func (e *EpochDerivativesEngine) UpdatePremium(
	contractID string,
	report EconomicReport,
) {
	c, ok := e.contracts[contractID]
	if !ok {
		return
	}

	risk := report.StabilityRisk + report.DivergenceRisk
	scarcity := report.ReplayScarcity + report.TimeslipScarcity

	c.Premium = (risk * 0.5) + (scarcity * 0.3) + (report.InflationRate * 0.2)

	e.contracts[contractID] = c
}
