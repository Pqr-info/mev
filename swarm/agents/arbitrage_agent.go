package agents

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
)

type PoolState struct {
	Address  common.Address
	Price    *big.Int
	Reserve0 *big.Int
	Reserve1 *big.Int
}

type ArbitrageAgent struct {
	ID        string
	UniV3Pool PoolState
	CamelotPool PoolState
}

func NewArbitrageAgent(id string, uniV3 common.Address, camelot common.Address) *ArbitrageAgent {
	// Parse large float strings to avoid untyped float overflows in big.NewInt
	res0Uni, _ := new(big.Int).SetString("100000000000000000000", 10)
	res0Cam, _ := new(big.Int).SetString("80000000000000000000", 10)

	return &ArbitrageAgent{
		ID: id,
		UniV3Pool: PoolState{
			Address:  uniV3,
			Price:    big.NewInt(3000 * 1e6), // Mock $3000 per ETH in stablecoin
			Reserve0: res0Uni,
			Reserve1: big.NewInt(300000 * 1e6),
		},
		CamelotPool: PoolState{
			Address:  camelot,
			Price:    big.NewInt(3015 * 1e6), // Mock $3015 per ETH (0.5% disparity)
			Reserve0: res0Cam,
			Reserve1: big.NewInt(241200 * 1e6),
		},
	}
}

// CalculateEV determines if there's sufficient EV between pools to propose a trade
func (a *ArbitrageAgent) CalculateEV(ctx context.Context, gasPriceGwei int64) (bool, *big.Int, float64) {
	// Simple disparity calculation
	diff := new(big.Int).Sub(a.CamelotPool.Price, a.UniV3Pool.Price)
	log.Debug().Msgf("Price disparity: %s", diff.String())
	
	// Convert disparity value to simulated EV in Wei
	// Disparity of 15 USD on 1 ETH size trade is approx 0.005 ETH under normal market dynamics
	evWei := big.NewInt(5000000000000000) // Base 0.005 ETH
	
	// Calculate cost
	gasLimit := int64(180000)
	gasCostWei := gasLimit * gasPriceGwei * 1e9
	
	netEvWei := new(big.Int).Sub(evWei, big.NewInt(int64(gasCostWei)))
	
	confidence := 0.95
	if gasPriceGwei > 40 {
		confidence = 0.70 // High gas price volatility drops execution confidence
	}

	if netEvWei.Cmp(big.NewInt(0)) > 0 {
		log.Info().Str("agent", a.ID).Msgf("Positive EV arbitrage opportunity found: %s Wei", netEvWei.String())
		return true, netEvWei, confidence
	}

	return false, big.NewInt(0), 0.0
}
