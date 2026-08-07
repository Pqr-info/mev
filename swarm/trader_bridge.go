package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// LineageBlock represents the recorded history of a constitutional financial mutation (Biographer format)
type LineageBlock struct {
	LineageID string            `json:"lineage_id"`
	Mutation  FinancialMutation `json:"mutation"`
	AccountID string            `json:"account_id"`
	Decision  string            `json:"decision"` // Approved / Rejected
	Reason    string            `json:"reason"`
	Organ     string            `json:"organ"` // "TRADER"
	Timestamp time.Time         `json:"timestamp"`
}

// TraderBridge coordinates MEV execution, Consensus evaluation, Ledger mutations, and Biographer logging
type TraderBridge struct {
	mu              sync.RWMutex
	portfolioEngine *TemporalPortfolioEngine
	consensusEngine *ConsensusEngine
	lineageStore    map[string]LineageBlock
	proposalChan    chan FinancialMutation
}

func NewTraderBridge(pe *TemporalPortfolioEngine, ce *ConsensusEngine) *TraderBridge {
	return &TraderBridge{
		portfolioEngine: pe,
		consensusEngine: ce,
		lineageStore:    make(map[string]LineageBlock),
		proposalChan:    make(chan FinancialMutation, 100),
	}
}

// ProposeTrade converts trading execution signals into a governed mutation proposal
func (tb *TraderBridge) ProposeTrade(
	accountID string,
	mut FinancialMutation,
	votes []MutationVote,
	config ValidatorConfig,
) (bool, string, string) {
	// 1. Fetch current portfolio state
	p := tb.portfolioEngine.GetOrCreate(accountID)

	// 2. Queue mutation to proposal channel (simulation of proposal_engine.mutation stream)
	select {
	case tb.proposalChan <- mut:
	default:
		// Channel full, continue inline
	}

	// 3. Evaluate mutation through the Consensus Lane (which includes Article II validation)
	decision := tb.consensusEngine.DecideFinancial(p, mut, config, votes)

	var lineageID string
	if decision.Approved {
		// 4. Apply approved mutation to portfolio state
		if err := tb.portfolioEngine.ApplyMutation(accountID, mut); err != nil {
			return false, fmt.Sprintf("Ledger state mutation failed: %v", err), ""
		}

		// 5. Generate unique lineage transaction ID
		hasher := sha256.New()
		hasher.Write([]byte(fmt.Sprintf("%s-%s-%f-%d", accountID, mut.Kind, mut.Amount+mut.Qty, time.Now().UnixNano())))
		lineageID = hex.EncodeToString(hasher.Sum(nil))

		// 6. Record event under Biographer Lineage
		tb.mu.Lock()
		tb.lineageStore[lineageID] = LineageBlock{
			LineageID: lineageID,
			Mutation:  mut,
			AccountID: accountID,
			Decision:  "Approved",
			Reason:    decision.Reason,
			Organ:     "TRADER",
			Timestamp: decision.Timestamp,
		}
		tb.mu.Unlock()
	} else {
		// Log rejection lineage block
		hasher := sha256.New()
		hasher.Write([]byte(fmt.Sprintf("REJECTED-%s-%s-%d", accountID, mut.Kind, time.Now().UnixNano())))
		lineageID = "rejected-" + hex.EncodeToString(hasher.Sum(nil))[:16]

		tb.mu.Lock()
		tb.lineageStore[lineageID] = LineageBlock{
			LineageID: lineageID,
			Mutation:  mut,
			AccountID: accountID,
			Decision:  "Rejected",
			Reason:    decision.Reason,
			Organ:     "TRADER",
			Timestamp: decision.Timestamp,
		}
		tb.mu.Unlock()
	}

	return decision.Approved, decision.Reason, lineageID
}

// GetLineage returns a recorded lineage block by its unique transaction ID
func (tb *TraderBridge) GetLineage(lineageID string) (LineageBlock, bool) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	block, exists := tb.lineageStore[lineageID]
	return block, exists
}

// TraderIntent represents the sensory/motor intent to mutate trading state
type TraderIntent struct {
	AccountID string                `json:"account_id"`
	Symbol    string                `json:"symbol"`
	Qty       float64               `json:"qty"`
	CostBasis float64               `json:"cost_basis"`
	Amount    float64               `json:"amount"`
	Delta     float64               `json:"delta"`
	Kind      FinancialMutationKind `json:"kind"`
}

// TraderResult represents the final consensus outcome
type TraderResult struct {
	Approved  bool   `json:"approved"`
	Rationale string `json:"rationale"`
	LineageID string `json:"lineage_id"`
}

// IntentToMutation converts raw trader signals to governed FinancialMutation structs
func (tb *TraderBridge) IntentToMutation(intent TraderIntent) FinancialMutation {
	return FinancialMutation{
		Kind:      intent.Kind,
		Symbol:    intent.Symbol,
		Qty:       intent.Qty,
		CostBasis: intent.CostBasis,
		Amount:    intent.Amount,
		Delta:     intent.Delta,
	}
}

// Execute submits an intent, evaluates consensus, and emits the lineage block
func (tb *TraderBridge) Execute(
	intent TraderIntent,
	votes []MutationVote,
	config ValidatorConfig,
) (TraderResult, error) {
	mut := tb.IntentToMutation(intent)
	approved, reason, lineageID := tb.ProposeTrade(intent.AccountID, mut, votes, config)

	return TraderResult{
		Approved:  approved,
		Rationale: reason,
		LineageID: lineageID,
	}, nil
}

