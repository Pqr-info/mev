package main

import "time"

type SettlementRecord struct {
	RecordID    string
	Epoch       string
	FromAccount string
	ToAccount   string
	Amount      float64
	Instrument  string // "spot", "derivative", "loan", "insurance"
	ReferenceID string // orderID, facilityID, claimID, etc.
	Timestamp   time.Time
}

type EpochSettlementEngine struct {
	bank    *EpochBank
	records []SettlementRecord
}

func NewEpochSettlementEngine(b *EpochBank) *EpochSettlementEngine {
	return &EpochSettlementEngine{
		bank:    b,
		records: []SettlementRecord{},
	}
}

func (se *EpochSettlementEngine) SettleTransfer(
	epoch string,
	fromAcctID string,
	toAcctID string,
	amount float64,
	instrument string,
	refID string,
) {
	from, ok := se.bank.Accounts[fromAcctID]
	if !ok {
		return
	}
	to, ok := se.bank.Accounts[toAcctID]
	if !ok {
		return
	}

	if from.CashBalance < amount {
		return
	}

	from.CashBalance -= amount
	to.CashBalance += amount

	rec := SettlementRecord{
		RecordID:    "set-" + epoch + "-" + refID,
		Epoch:       epoch,
		FromAccount: fromAcctID,
		ToAccount:   toAcctID,
		Amount:      amount,
		Instrument:  instrument,
		ReferenceID: refID,
		Timestamp:   time.Now(),
	}

	se.records = append(se.records, rec)
}

func (se *EpochSettlementEngine) RecordsForEpoch(epoch string) []SettlementRecord {
	out := []SettlementRecord{}
	for _, r := range se.records {
		if r.Epoch == epoch {
			out = append(out, r)
		}
	}
	return out
}
