package agents

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
)

type MarketState struct {
	GasPriceGwei int64
	BlockNumber  uint64
	UniV3Price   *big.Int // base-18 precision simulated price
	CamelotPrice *big.Int // base-18 precision simulated price
}

type BundleProposal struct {
	AgentID       string
	ExpectedValue *big.Int
	GasCost       *big.Int
	LatencyMs     float64
	Confidence    float64
	RiskScore     float64
	ExecutionPath string
}

type Agent interface {
	Analyze(ctx context.Context, state MarketState) (BundleProposal, error)
	ID() string
}

type ArbitrumArbitrageAgent struct {
	agentID      string
	uniV3Pool    common.Address
	camelotPool  common.Address
	maxSlippage  float64
	maxExposure  *big.Int
}

func NewArbitrumArbitrageAgent(id string, uniV3 common.Address, camelot common.Address, maxSlippage float64, maxExposureWei *big.Int) *ArbitrumArbitrageAgent {
	return &ArbitrumArbitrageAgent{
		agentID:      id,
		uniV3Pool:    uniV3,
		camelotPool:  camelot,
		maxSlippage:  maxSlippage,
		maxExposure:  maxExposureWei,
	}
}

func (a *ArbitrumArbitrageAgent) ID() string {
	return a.agentID
}

func (a *ArbitrumArbitrageAgent) Analyze(ctx context.Context, state MarketState) (BundleProposal, error) {
	// Calculate execution discrepancy: e.g. Arbitrum UniswapV3 ↔ Camelot
	priceDiff := new(big.Int).Sub(state.CamelotPrice, state.UniV3Price)
	absDiff := new(big.Int).Abs(priceDiff)

	// Simulated size of 10 ETH swap
	tradeSize := big.NewInt(10)
	rawEvWei := new(big.Int).Mul(absDiff, tradeSize)

	// Gas Cost in Wei (180,000 gas units)
	gasCostWei := new(big.Int).Mul(big.NewInt(180000), big.NewInt(state.GasPriceGwei*1e9))

	netEvWei := new(big.Int).Sub(rawEvWei, gasCostWei)

	// Confidence parameters
	confidence := 0.98
	if state.GasPriceGwei > 50 {
		confidence = 0.80 // reduced confidence during network congestion
	}

	// Calculate a dynamic risk score based on volatility/latency indicators
	riskScore := 0.15
	if priceDiff.Sign() < 0 {
		// Opposite arbitrage path: Camelot -> UniswapV3
		riskScore = 0.20
	}

	proposal := BundleProposal{
		AgentID:       a.agentID,
		ExpectedValue: netEvWei,
		GasCost:       gasCostWei,
		LatencyMs:     0.45, // sub-millisecond execution pipeline estimate
		Confidence:    confidence,
		RiskScore:     riskScore,
		ExecutionPath: "UniswapV3->Camelot",
	}

	log.Debug().
		Str("agent", a.agentID).
		Str("net_ev", netEvWei.String()).
		Msg("Arbitrum Arbitrage Agent analysis completed")

	return proposal, nil
}
