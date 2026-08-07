package main

import "time"

type MarketCommodity struct {
	CommodityID string
	Epoch       string
	Type        string // "replay-right", "timeslip-right", "mutation-bandwidth", etc.
	Quantity    float64
	Price       float64
	Timestamp   time.Time
}

type MarketOrder struct {
	OrderID     string
	CommodityID string
	Side        string // "BUY" or "SELL"
	Quantity    float64
	LimitPrice  float64
	Timestamp   time.Time
}

type EpochMarketEngine struct {
	commodities map[string]MarketCommodity
	orderBook   []MarketOrder
}

func NewEpochMarketEngine() *EpochMarketEngine {
	return &EpochMarketEngine{
		commodities: make(map[string]MarketCommodity),
		orderBook:   []MarketOrder{},
	}
}

func (m *EpochMarketEngine) ListCommodity(c MarketCommodity) {
	m.commodities[c.CommodityID] = c
}

func (m *EpochMarketEngine) SubmitOrder(o MarketOrder) {
	o.OrderID = "mkt-" + o.CommodityID + "-" + time.Now().Format("150405")
	m.orderBook = append(m.orderBook, o)
}

func (m *EpochMarketEngine) Match() []MarketOrder {
	fills := []MarketOrder{}

	for _, o := range m.orderBook {
		commodity, ok := m.commodities[o.CommodityID]
		if !ok {
			continue
		}

		if o.Side == "BUY" && o.LimitPrice >= commodity.Price {
			fills = append(fills, o)
		}
		if o.Side == "SELL" && o.LimitPrice <= commodity.Price {
			fills = append(fills, o)
		}
	}

	return fills
}

func (m *EpochMarketEngine) UpdatePrice(commodityID string, demand float64, supply float64) {
	c, ok := m.commodities[commodityID]
	if !ok {
		return
	}

	if supply == 0 {
		supply = 0.0001
	}

	c.Price = demand / supply
	m.commodities[commodityID] = c
}
