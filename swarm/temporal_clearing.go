package main

import "time"

type ClearingInstruction struct {
	InstructionID string
	Epoch         string
	FromAccount   string
	ToAccount     string
	NetAmount     float64
	Instruments   []string
	ReferenceIDs  []string
	Timestamp     time.Time
}

type TemporalClearinghouseEngine struct {
	settlement   *EpochSettlementEngine
	instructions []ClearingInstruction
}

func NewTemporalClearinghouseEngine(
	s *EpochSettlementEngine,
) *TemporalClearinghouseEngine {
	return &TemporalClearinghouseEngine{
		settlement:   s,
		instructions: []ClearingInstruction{},
	}
}

func (ce *TemporalClearinghouseEngine) NetEpoch(epoch string) []ClearingInstruction {
	records := ce.settlement.RecordsForEpoch(epoch)

	netMap := map[string]map[string]*ClearingInstruction{}

	for _, r := range records {
		if netMap[r.FromAccount] == nil {
			netMap[r.FromAccount] = map[string]*ClearingInstruction{}
		}

		inst := netMap[r.FromAccount][r.ToAccount]
		if inst == nil {
			inst = &ClearingInstruction{
				InstructionID: "clr-" + epoch + "-" + r.ReferenceID,
				Epoch:         epoch,
				FromAccount:   r.FromAccount,
				ToAccount:     r.ToAccount,
				NetAmount:     0,
				Instruments:   []string{},
				ReferenceIDs:  []string{},
				Timestamp:     time.Now(),
			}
			netMap[r.FromAccount][r.ToAccount] = inst
		}

		inst.NetAmount += r.Amount
		inst.Instruments = append(inst.Instruments, r.Instrument)
		inst.ReferenceIDs = append(inst.ReferenceIDs, r.ReferenceID)
	}

	out := []ClearingInstruction{}
	for _, inner := range netMap {
		for _, inst := range inner {
			out = append(out, *inst)
		}
	}

	ce.instructions = append(ce.instructions, out...)
	return out
}

func (ce *TemporalClearinghouseEngine) Execute(instruction ClearingInstruction) {
	ce.settlement.SettleTransfer(
		instruction.Epoch,
		instruction.FromAccount,
		instruction.ToAccount,
		instruction.NetAmount,
		"clearing",
		instruction.InstructionID,
	)
}
