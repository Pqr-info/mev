package main

import "time"

type LedgerEntry struct {
	EntryID     string
	Epoch       string
	Type        string // "settlement", "clearing", "event"
	FromAccount string
	ToAccount   string
	Amount      float64
	Instrument  string
	ReferenceID string
	Timestamp   time.Time
}

type LedgerBlock struct {
	BlockID   string
	Epoch     string
	Entries   []LedgerEntry
	CreatedAt time.Time
}

type TemporalLedgerEngine struct {
	blocks map[string]*LedgerBlock
}

func NewTemporalLedgerEngine() *TemporalLedgerEngine {
	return &TemporalLedgerEngine{
		blocks: make(map[string]*LedgerBlock),
	}
}

func (le *TemporalLedgerEngine) Append(entry LedgerEntry) {
	block := le.blocks[entry.Epoch]

	if block == nil {
		block = &LedgerBlock{
			BlockID:   "block-" + entry.Epoch,
			Epoch:     entry.Epoch,
			Entries:   []LedgerEntry{},
			CreatedAt: time.Now(),
		}
		le.blocks[entry.Epoch] = block
	}

	block.Entries = append(block.Entries, entry)
}

func (le *TemporalLedgerEngine) Block(epoch string) *LedgerBlock {
	return le.blocks[epoch]
}
