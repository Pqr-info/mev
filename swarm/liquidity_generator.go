package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pqr.info/shared/task"
)

type LiquidityGenerator struct {
	Registry ArbitrageRegistry
	PredictiveEngine *DefaultPredictiveEngine
	Substrate *SubstrateClient
}

func NewLiquidityGenerator(registry ArbitrageRegistry, pe *DefaultPredictiveEngine, client *SubstrateClient) *LiquidityGenerator {
	if pe == nil {
		pe = NewDefaultPredictiveEngine(client)
	}
	return &LiquidityGenerator{
		Registry: registry,
		PredictiveEngine: pe,
		Substrate: client,
	}
}

func (lg *LiquidityGenerator) Name() string {
	return "LiquidityGenerator"
}

func (lg *LiquidityGenerator) Execute(res task.TaskResult) error {
	fmt.Printf("[%s] Executing task %x from Mesh Supervisor\n", lg.Name(), res.Canonical.Packed)
	lg.OnSwapEvent("USDC", "SUB27", 1000.0, "mesh_supervisor_task")
	
	if lg.Substrate != nil {
		txHash, err := lg.Substrate.SubmitTransaction(context.Background(), []byte{})
		if err == nil {
			fmt.Printf("[%s] Transaction routed via Nuremberg peer: %s\n", lg.Name(), txHash)
		}
	}
	return nil
}

// OnSwapEvent is called by the DeFi exchange whenever a user swaps.
func (lg *LiquidityGenerator) OnSwapEvent(fromToken string, toToken string, amount float64, user string) {
	fmt.Printf("[LIQUIDITY-GEN] Intercepted Swap: %f %s -> %s by %s\n", amount, fromToken, toToken, user)

	if lg.Registry != nil {
		// Mock pool state
		poolState := AMMPoolState{
			LiquidityIn:   1000000.0,
			LiquidityOut:  2500000.0,
			Utilization:   0.45,
			DepthEstimate: 50000.0,
		}

		ctx := TemporalContext{
			Swap: SwapEvent{
				AssetIn:        1000, // mock USDC
				AssetOut:       27,   // mock Sub27
				AmountIn:       amount,
				AmountOut:      amount * 2.5, // generic rate
				SpotPrice:      2.5,
				SlippageActual: 0.05,
			},
			PoolState: poolState,
		}

		sig := lg.PredictiveEngine.Enrich(ctx)
		sig.SignalID = fmt.Sprintf("sig_%d", time.Now().UnixNano())
		sig.EmittedAt = time.Now().UTC()
		sig.ChainID = "Substrate27"
		sig.SourceModule = "liquidity_generator"
		sig.EventType = "USDC_SUB27_SWAP"
		
		sig.AssetIn = ctx.Swap.AssetIn
		sig.AssetOut = ctx.Swap.AssetOut
		sig.AmountIn = fmt.Sprintf("%f", ctx.Swap.AmountIn)
		sig.AmountOut = fmt.Sprintf("%f", ctx.Swap.AmountOut)
		sig.SpotPrice = fmt.Sprintf("%f", ctx.Swap.SpotPrice)
		sig.SlippageActual = fmt.Sprintf("%f", ctx.Swap.SlippageActual)

		sig.PoolLiquidityIn = fmt.Sprintf("%f", poolState.LiquidityIn)
		sig.PoolLiquidityOut = fmt.Sprintf("%f", poolState.LiquidityOut)
		sig.PoolUtilization = fmt.Sprintf("%f", poolState.Utilization)
		sig.DepthEstimate = fmt.Sprintf("%f", poolState.DepthEstimate)
		
		lg.Registry.StoreSignal(context.Background(), sig)
		fmt.Printf("[LIQUIDITY-GEN] Broadcasted Enriched TemporalArbitrageSignal to Engine: %s\n", sig.SignalID)

		payload, _ := json.Marshal(sig)
		LogStateTransition(FiveDAddress{}, payload)
	}
}

// OnNFTBuy is called whenever an NFT is purchased, potentially signaling liquidity shifts.
func (lg *LiquidityGenerator) OnNFTBuy(nftID uint32, currency string, price float64) {
	fmt.Printf("[LIQUIDITY-GEN] NFT Purchase detected: Item %d for %f %s\n", nftID, price, currency)
	
	if lg.Registry != nil && currency == "USDC" {
		ctx := TemporalContext{
			NFT: NFTBuyEvent{
				NFTID:    nftID,
				Currency: currency,
				Price:    price,
			},
		}

		sig := lg.PredictiveEngine.Enrich(ctx)
		sig.SignalID = fmt.Sprintf("sig_nft_%d", time.Now().UnixNano())
		sig.EmittedAt = time.Now().UTC()
		sig.ChainID = "Substrate27"
		sig.SourceModule = "liquidity_generator"
		sig.EventType = "NFT_BUY_USDC_INFLUX"
		
		lg.Registry.StoreSignal(context.Background(), sig)

		payload, _ := json.Marshal(sig)
		LogStateTransition(FiveDAddress{}, payload)
	}
}
