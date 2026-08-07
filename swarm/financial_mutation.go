package main

import "encoding/json"

type FinancialMutationKind string

const (
	PortfolioAdd      FinancialMutationKind = "portfolio_add"
	PortfolioRemove   FinancialMutationKind = "portfolio_remove"
	PortfolioAdjust   FinancialMutationKind = "portfolio_adjust"
	CashAdjust        FinancialMutationKind = "cash_adjust"
	RiskProfileUpdate FinancialMutationKind = "risk_profile_update"

	// Integrated custom operations
	GetSymbolPrice    FinancialMutationKind = "get_symbol_price"
	ViewOrderbook     FinancialMutationKind = "view_orderbook"
	PlaceTrade        FinancialMutationKind = "place_trade"
	ShortSell         FinancialMutationKind = "short_sell"
	Pull              FinancialMutationKind = "pull"
	Put               FinancialMutationKind = "put"
	CoveredCall       FinancialMutationKind = "covered_call"
	Moneypress        FinancialMutationKind = "moneypress"
	Long              FinancialMutationKind = "long"
	Short             FinancialMutationKind = "short"
	OverUnder         FinancialMutationKind = "over_under"
	CornerSlice       FinancialMutationKind = "corner_slice"

	// Additional Integrated Operations
	BidPrice          FinancialMutationKind = "bid_price"
	AskPrice          FinancialMutationKind = "ask_price"
	Spread            FinancialMutationKind = "spread"
	Squeeze           FinancialMutationKind = "squeeze"
	BidUnder          FinancialMutationKind = "bidunder"
	BidOver           FinancialMutationKind = "bidover"
	MarketBuy         FinancialMutationKind = "market_buy"
	MarketSell        FinancialMutationKind = "market_sell"
)

type FinancialMutation struct {
	Kind      FinancialMutationKind `json:"kind"`
	Symbol    string                `json:"symbol,omitempty"`
	Qty       float64               `json:"qty,omitempty"`
	CostBasis float64               `json:"cost_basis,omitempty"`
	Amount    float64               `json:"amount,omitempty"`
	Delta     float64               `json:"delta,omitempty"`
}

// Marshal helper for Proposal Engine transport
func (m *FinancialMutation) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal helper for HealerAgentV2 Mode B mutation loop
func DeserializeFinancialMutation(data []byte) (*FinancialMutation, error) {
	var m FinancialMutation
	err := json.Unmarshal(data, &m)
	return &m, err
}
