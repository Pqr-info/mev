package main

import "time"

type BankAccount struct {
	AccountID      string
	OwnerID        string
	Epoch          string
	CashBalance    float64
	Commodities    map[string]float64 // commodityID -> quantity
	Derivatives    map[string]float64 // contractID -> quantity
	CollateralLock float64
	Timestamp      time.Time
}

type CreditFacility struct {
	FacilityID   string
	LenderBankID string
	BorrowerID   string
	Epoch        string
	Principal    float64
	InterestRate float64
	Collateral   float64
	Maturity     time.Time
	Active       bool
}

type EpochBank struct {
	BankID     string
	Accounts   map[string]*BankAccount
	Facilities map[string]*CreditFacility
	Market     *EpochMarketEngine
	Economics  *TemporalEconomicEngine
	Monetary   *TemporalMonetaryPolicyEngine
}

func NewEpochBank(
	id string,
	m *EpochMarketEngine,
	e *TemporalEconomicEngine,
	mp *TemporalMonetaryPolicyEngine,
) *EpochBank {
	return &EpochBank{
		BankID:     id,
		Accounts:   make(map[string]*BankAccount),
		Facilities: make(map[string]*CreditFacility),
		Market:     m,
		Economics:  e,
		Monetary:   mp,
	}
}

func (b *EpochBank) OpenAccount(ownerID, epoch string) *BankAccount {
	acc := &BankAccount{
		AccountID:   "acct-" + ownerID + "-" + epoch,
		OwnerID:     ownerID,
		Epoch:       epoch,
		CashBalance: 0,
		Commodities: make(map[string]float64),
		Derivatives: make(map[string]float64),
		Timestamp:   time.Now(),
	}
	b.Accounts[acc.AccountID] = acc
	return acc
}

func (b *EpochBank) DepositCash(accountID string, amount float64) {
	acc, ok := b.Accounts[accountID]
	if !ok {
		return
	}
	acc.CashBalance += amount
}

func (b *EpochBank) IssueLoan(
	borrowerID string,
	epoch string,
	amount float64,
	collateral float64,
	duration time.Duration,
) *CreditFacility {
	policy := b.Monetary.GeneratePolicy(epoch)

	facility := &CreditFacility{
		FacilityID:   "loan-" + borrowerID + "-" + epoch + "-" + time.Now().Format("150405"),
		LenderBankID: b.BankID,
		BorrowerID:   borrowerID,
		Epoch:        epoch,
		Principal:    amount,
		InterestRate: policy.InterestRate,
		Collateral:   collateral,
		Maturity:     time.Now().Add(duration),
		Active:       true,
	}

	b.Facilities[facility.FacilityID] = facility

	accID := "acct-" + borrowerID + "-" + epoch
	acc, ok := b.Accounts[accID]
	if ok {
		acc.CashBalance += amount
		acc.CollateralLock += collateral
	}

	return facility
}

func (b *EpochBank) EvaluateCollateral(
	accountID string,
) float64 {
	acc, ok := b.Accounts[accountID]
	if !ok {
		return 0
	}

	total := 0.0
	for commodityID, qty := range acc.Commodities {
		c, ok := b.Market.commodities[commodityID]
		if !ok {
			continue
		}
		total += qty * c.Price
	}

	acc.CollateralLock = total
	return total
}

