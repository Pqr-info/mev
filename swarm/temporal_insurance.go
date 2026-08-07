package main

import "time"

type InsurancePolicy struct {
	PolicyID        string
	InsuredID       string
	Epoch           string
	CoverageAmount  float64
	Premium         float64
	StabilityRisk   float64
	DivergenceRisk  float64
	CollateralValue float64
	Active          bool
	Timestamp       time.Time
}

type InsuranceClaim struct {
	ClaimID    string
	PolicyID   string
	LossAmount float64
	Approved   bool
	Timestamp  time.Time
}

type TemporalInsuranceEngine struct {
	economics *TemporalEconomicEngine
	credit    *TemporalCreditEngine
	bank      *EpochBank

	policies map[string]*InsurancePolicy
	claims   map[string]*InsuranceClaim
}

func NewTemporalInsuranceEngine(
	e *TemporalEconomicEngine,
	c *TemporalCreditEngine,
	b *EpochBank,
) *TemporalInsuranceEngine {
	return &TemporalInsuranceEngine{
		economics: e,
		credit:    c,
		bank:      b,
		policies:  make(map[string]*InsurancePolicy),
		claims:    make(map[string]*InsuranceClaim),
	}
}

func (ie *TemporalInsuranceEngine) CreatePolicy(
	insuredID string,
	epoch string,
	coverage float64,
) *InsurancePolicy {
	report := ie.economics.Analyze(epoch)
	score := ie.credit.Score(insuredID, epoch)

	premium := (report.StabilityRisk * 0.4) +
		(report.DivergenceRisk * 0.4) +
		(1.0/(score.CollateralValue+0.0001))*0.2

	policy := &InsurancePolicy{
		PolicyID:        "pol-" + insuredID + "-" + epoch,
		InsuredID:       insuredID,
		Epoch:           epoch,
		CoverageAmount:  coverage,
		Premium:         premium,
		StabilityRisk:   report.StabilityRisk,
		DivergenceRisk:  report.DivergenceRisk,
		CollateralValue: score.CollateralValue,
		Active:          true,
		Timestamp:       time.Now(),
	}

	ie.policies[policy.PolicyID] = policy
	return policy
}

func (ie *TemporalInsuranceEngine) SubmitClaim(
	policyID string,
	loss float64,
) *InsuranceClaim {
	claim := &InsuranceClaim{
		ClaimID:    "claim-" + policyID,
		PolicyID:   policyID,
		LossAmount: loss,
		Approved:   false,
		Timestamp:  time.Now(),
	}

	ie.claims[claim.ClaimID] = claim
	return claim
}

func (ie *TemporalInsuranceEngine) EvaluateClaim(
	claimID string,
) bool {
	claim, ok := ie.claims[claimID]
	if !ok {
		return false
	}
	policy, ok := ie.policies[claim.PolicyID]
	if !ok {
		return false
	}

	if !policy.Active {
		return false
	}

	if claim.LossAmount <= policy.CoverageAmount {
		claim.Approved = true
		return true
	}

	return false
}

func (ie *TemporalInsuranceEngine) PayoutClaim(
	claimID string,
) bool {
	claim, ok := ie.claims[claimID]
	if !ok || !claim.Approved {
		return false
	}

	policy, ok := ie.policies[claim.PolicyID]
	if !ok {
		return false
	}

	acc, ok := ie.bank.Accounts["acct-"+policy.InsuredID+"-"+policy.Epoch]
	if !ok {
		return false
	}

	acc.CashBalance += claim.LossAmount
	return true
}

