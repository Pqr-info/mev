package router

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// TransactionIntent represents classified traits of an incoming tx
type TransactionIntent struct {
	Hash       common.Hash
	MethodName string
	IsArbTarget bool
	TargetPool  common.Address
	EstimatedEV *big.Int
	RoutePath   string
}

type GemmaSemanticRouter struct {
	ThresholdGwei *big.Int
}

func NewSemanticRouter(thresholdGwei int64) *GemmaSemanticRouter {
	return &GemmaSemanticRouter{
		ThresholdGwei: big.NewInt(thresholdGwei),
	}
}

// ClassifyTransaction parses an incoming raw transaction and semanticizes it
func (g *GemmaSemanticRouter) ClassifyTransaction(ctx context.Context, txHash common.Hash, to common.Address, input []byte, gasPrice *big.Int) (*TransactionIntent, error) {
	intent := &TransactionIntent{
		Hash:       txHash,
		TargetPool: to,
		EstimatedEV: big.NewInt(0),
	}

	// Mock parsing based on function signature selectors
	if len(input) >= 4 {
		selector := fmt.Sprintf("0x%x", input[:4])
		switch selector {
		case "0xa9059cbb": // transfer
			intent.MethodName = "transfer"
		case "0x38ed1739": // swapExactTokensForTokens
			intent.MethodName = "swapExactTokensForTokens"
			intent.IsArbTarget = true
			intent.RoutePath = "UniswapV3->Camelot"
			intent.EstimatedEV = big.NewInt(50000000000000000) // 0.05 ETH estimated EV
		case "0x415565b0": // swap
			intent.MethodName = "swap"
			intent.IsArbTarget = true
			intent.RoutePath = "Camelot->UniswapV3"
			intent.EstimatedEV = big.NewInt(20000000000000000) // 0.02 ETH estimated EV
		default:
			intent.MethodName = "unknown"
		}
	} else {
		intent.MethodName = "native-transfer"
	}

	// Route dynamically based on gas price and EV
	if gasPrice.Cmp(g.ThresholdGwei) > 0 && intent.IsArbTarget {
		// Degrade EV estimate due to higher congestion costs
		intent.EstimatedEV = new(big.Int).Sub(intent.EstimatedEV, big.NewInt(5000000000000000))
	}

	return intent, nil
}

// RouteIntent selects the optimal execution queue / swarm agent based on classified intent
func (g *GemmaSemanticRouter) RouteIntent(intent *TransactionIntent) string {
	if !intent.IsArbTarget {
		return "discard"
	}

	if intent.EstimatedEV.Cmp(big.NewInt(30000000000000000)) > 0 {
		return "high-frequency-hot-path"
	}

	if strings.Contains(intent.RoutePath, "Camelot") {
		return "arbitrum-camelot-agent"
	}

	return "default-arb-worker"
}
