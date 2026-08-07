package main

import "time"

type ConsensusVote struct {
	VoteID    string
	NodeID    string
	Epoch     string
	BlockID   string
	Approved  bool
	Timestamp time.Time
}

type ConsensusResult struct {
	Epoch     string
	BlockID   string
	Approved  bool
	Votes     []ConsensusVote
	Finalized time.Time
}

type TemporalConsensusEngine struct {
	ledger  *TemporalLedgerEngine
	votes   map[string][]ConsensusVote
	results map[string]*ConsensusResult
}

func NewTemporalConsensusEngine(
	l *TemporalLedgerEngine,
) *TemporalConsensusEngine {
	return &TemporalConsensusEngine{
		ledger:  l,
		votes:   make(map[string][]ConsensusVote),
		results: make(map[string]*ConsensusResult),
	}
}

func (ce *TemporalConsensusEngine) SubmitVote(
	nodeID string,
	epoch string,
	approved bool,
) ConsensusVote {
	block := ce.ledger.Block(epoch)
	blockID := ""
	if block != nil {
		blockID = block.BlockID
	}

	vote := ConsensusVote{
		VoteID:    "vote-" + nodeID + "-" + epoch,
		NodeID:    nodeID,
		Epoch:     epoch,
		BlockID:   blockID,
		Approved:  approved,
		Timestamp: time.Now(),
	}

	ce.votes[epoch] = append(ce.votes[epoch], vote)
	return vote
}

func (ce *TemporalConsensusEngine) Finalize(epoch string) *ConsensusResult {
	votes := ce.votes[epoch]

	approvedCount := 0
	for _, v := range votes {
		if v.Approved {
			approvedCount++
		}
	}

	approved := false
	if len(votes) > 0 {
		approved = approvedCount > len(votes)/2
	}

	block := ce.ledger.Block(epoch)
	blockID := ""
	if block != nil {
		blockID = block.BlockID
	}

	result := &ConsensusResult{
		Epoch:     epoch,
		BlockID:   blockID,
		Approved:  approved,
		Votes:     votes,
		Finalized: time.Now(),
	}

	ce.results[epoch] = result
	return result
}
